package internal

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// ErrAWSWAFChallenge is returned by ListShelf-style HTTP fetches when
// Goodreads responds with an AWS WAF JavaScript challenge (status 202 with
// a body containing `gokuProps` / `awsWafCookieDomainList`) instead of the
// requested HTML. The plain HTTP client cannot solve the challenge — a
// browser-based fallback is needed. Callers (and tests) can detect this
// with errors.Is so they can degrade gracefully instead of reporting the
// generic "status 202" that used to leak through.
var ErrAWSWAFChallenge = errors.New("goodreads returned AWS WAF challenge — browser session cookie needed")

// isAWSWAFChallengeBody reports whether the response body is the AWS WAF
// JS challenge landing page. AWS WAF injects `awsWafCookieDomainList` and a
// `gokuProps` block into the tiny HTML wrapper it serves before letting the
// real request through.
func isAWSWAFChallengeBody(body string) bool {
	return strings.Contains(body, "awsWafCookieDomainList") ||
		strings.Contains(body, "gokuProps")
}

// ParseShelfHTML extracts Book records from the HTML of a
// /review/list/<user_id>?shelf=<name> page. Each book appears as
// `<tr id="review_NNN" class="bookalike review">...</tr>` with the resource
// ID, title, and author nested inside known `<td class="field …">` cells.
func ParseShelfHTML(html string) ([]Book, error) {
	rows := _shelfRowRE.FindAllString(html, -1)
	if rows == nil {
		return []Book{}, nil
	}
	books := make([]Book, 0, len(rows))
	for _, row := range rows {
		b, ok := parseShelfRow(row)
		if !ok {
			continue
		}
		books = append(books, b)
	}
	return books, nil
}

// ExtractUserIDFromHomeHTML pulls the logged-in user's numeric ID from any
// `<a href="/user/show/<id>(-<slug>)?">` link on a Goodreads page rendered
// for an authenticated session. The first match wins because Goodreads
// always renders the signed-in user's profile link before any other
// `/user/show/` link on the page.
func ExtractUserIDFromHomeHTML(html string) (string, error) {
	m := _userIDRE.FindStringSubmatch(html)
	if m == nil {
		return "", fmt.Errorf("no /user/show/<id> link found — not logged in?")
	}
	return m[1], nil
}

// ListShelf fetches a Goodreads shelf for the logged-in user and returns the
// books on it. Requires the cookies loaded from a prior `goodreads login` —
// without them, Goodreads either redirects to login or shows an empty page.
func (c *Client) ListShelf(shelfName string) ([]Book, error) {
	userID, err := c.fetchUserID()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(
		"%s/review/list/%s?shelf=%s&per_page=100",
		BaseURL, userID, shelfName,
	)
	html, err := c.fetchHTML(url)
	if err != nil {
		return nil, fmt.Errorf("fetching shelf %q: %w", shelfName, err)
	}
	return ParseShelfHTML(html)
}

func (c *Client) fetchUserID() (string, error) {
	html, err := c.fetchHTML(BaseURL + "/")
	if err != nil {
		return "", fmt.Errorf("fetching home page: %w", err)
	}
	return ExtractUserIDFromHomeHTML(html)
}

func (c *Client) fetchHTML(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		c.Log.Record("http_fetch_html", map[string]any{"url": url}, err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Log.Record("http_fetch_html", map[string]any{"url": url, "status": resp.StatusCode}, err)
		return "", err
	}
	c.Log.Record("http_fetch_html", map[string]any{"url": url, "status": resp.StatusCode, "bytes": len(body)}, nil)
	if resp.StatusCode == http.StatusAccepted && isAWSWAFChallengeBody(string(body)) {
		return "", fmt.Errorf("%w (url=%s)", ErrAWSWAFChallenge, url)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	return string(body), nil
}

// ---------------------------------------------------------------------------
// HTML parsing internals
// ---------------------------------------------------------------------------

// _shelfRowRE matches the bookalike review row that Goodreads renders for
// each book on a shelf. Greedy match to `</tr>` is fine because the rows
// don't nest.
var _shelfRowRE = regexp.MustCompile(`(?s)<tr\s+id="review_\d+"\s+class="bookalike review">.*?</tr>`)

// _resourceIDRE finds the book's stable resource ID inside the cover cell
// (`data-resource-id="55145261"`). The ID lives on the
// `js-tooltipTrigger` div regardless of cover-vs-table view.
var _resourceIDRE = regexp.MustCompile(`data-resource-id="(\d+)"`)

// _titleRE extracts the title from `<a title="…" href="/book/show/…">` —
// the `title` attribute holds the full, un-truncated title even when the
// anchor text is truncated for display.
var _titleRE = regexp.MustCompile(`<td class="field title">[\s\S]*?<a\s+title="([^"]+)"`)

// _authorRE extracts the author name from the first
// `<a href="/author/show/…">Name</a>` inside the author cell.
var _authorRE = regexp.MustCompile(`<td class="field author">[\s\S]*?<a\s+href="/author/show/[^"]+">([^<]+)</a>`)

// _userIDRE matches the logged-in user's profile link in the header.
var _userIDRE = regexp.MustCompile(`<a[^>]+href="/user/show/(\d+)`)

func parseShelfRow(row string) (Book, bool) {
	idMatch := _resourceIDRE.FindStringSubmatch(row)
	titleMatch := _titleRE.FindStringSubmatch(row)
	authorMatch := _authorRE.FindStringSubmatch(row)
	if idMatch == nil || titleMatch == nil {
		return Book{}, false
	}
	author := ""
	if authorMatch != nil {
		author = strings.TrimSpace(authorMatch[1])
	}
	return Book{
		ID:     idMatch[1],
		Title:  decodeHTMLEntities(strings.TrimSpace(titleMatch[1])),
		Author: decodeHTMLEntities(author),
	}, true
}

// decodeHTMLEntities decodes the small set of HTML entities Goodreads
// emits in title and author attributes (apostrophes, ampersands, quotes).
// A full HTML parser is overkill for these well-formed attribute values.
func decodeHTMLEntities(s string) string {
	r := strings.NewReplacer(
		"&amp;", "&",
		"&#39;", "'",
		"&apos;", "'",
		"&quot;", `"`,
		"&lt;", "<",
		"&gt;", ">",
	)
	return r.Replace(s)
}
