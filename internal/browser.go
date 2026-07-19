package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Browser wraps a rod browser and page for Goodreads interactions.
//
// Log accumulates every server interaction attempted during the session so
// error paths can dump the JSON tail into a bug report — see
// InteractionLog.Dump and saveDebugArtifacts.
type Browser struct {
	Rod  *rod.Browser
	Page *rod.Page
	Log  *InteractionLog
}

// NewBrowser launches a Chrome instance and navigates to goodreads.com.
// Set headless to false to see the browser for debugging.
//
// Chromium's setuid sandbox depends on either kernel.unprivileged_userns_clone
// being enabled or running as root with the helper binary. Many Linux
// environments (CI runners, the skyeclaw VM, ASUS kernels with the
// "Copy Fail" mitigation enabled) satisfy neither — so the browser dies
// with "No usable sandbox". goodreads-cli is an automation tool, not a
// user-facing browser; the sandbox guarantees aren't load-bearing here.
// Disable it on Linux unconditionally and let GOODREADS_BROWSER_SANDBOX=1
// force it back on for the cases where it actually works.
func NewBrowser(headless bool) (*Browser, error) {
	l := launcher.New().
		Headless(headless)
	if os.Getenv("GOODREADS_BROWSER_SANDBOX") != "1" {
		l = l.NoSandbox(true)
	}
	u, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w\n\nOn Linux, install required dependencies:\n  sudo apt install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libxdamage1 libxrandr2 libgbm1 libpango-1.0-0 libcairo2 libasound2 libxcomposite1 libxfixes3 libxkbcommon0 libdrm2 libatspi2.0-0", err)
	}

	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: "https://www.goodreads.com"})
	if err != nil {
		browser.MustClose()
		return nil, fmt.Errorf("failed to open page: %w", err)
	}
	page.MustWaitStable()

	b := &Browser{Rod: browser, Page: page, Log: NewInteractionLog()}
	b.Log.Record("browser_launch", map[string]any{"headless": headless}, nil)

	if err := b.LoadCookies(); err == nil {
		// Reload page with cookies applied
		page.MustNavigate("https://www.goodreads.com")
		page.MustWaitStable()
	}

	return b, nil
}

// Close cleans up the browser.
func (b *Browser) Close() {
	b.Rod.MustClose()
}

// FetchRenderedHTML navigates the browser to url, waits for the page to
// settle — including any AWS WAF JavaScript challenge — and returns the
// resulting HTML. This is how goodreads-cli reaches page endpoints that
// the plain HTTP client can't touch anymore: Chrome runs the WAF JS,
// picks up the aws-waf-token cookie, reloads to the real page, and we
// read back the fully rendered DOM.
//
// If after MustWaitStable the DOM still carries the WAF challenge
// markers (gokuProps / awsWafCookieDomainList), FetchRenderedHTML polls
// for up to ~15 s giving Chrome time to complete the challenge and
// auto-reload before returning ErrAWSWAFChallenge.
func (b *Browser) FetchRenderedHTML(url string) (string, error) {
	b.Log.Record("navigate", map[string]any{"url": url, "via": "browser"}, nil)
	if err := b.Page.Navigate(url); err != nil {
		b.Log.Record("navigate_error", map[string]any{"url": url}, err)
		return "", fmt.Errorf("navigating %s: %w", url, err)
	}
	if err := b.Page.WaitLoad(); err != nil {
		b.Log.Record("wait_load", map[string]any{"url": url}, err)
	}
	// MustWaitStable waits for the network to go idle — during a WAF
	// challenge Chrome issues the follow-up request itself, so waiting
	// for stability lands us on the post-challenge page in the common
	// case. Cap it because a page with continuous polling (analytics,
	// live-update widgets) never settles.
	b.Page.Timeout(15 * time.Second).MustWaitStable()

	html, err := b.Page.HTML()
	if err != nil {
		b.Log.Record("read_html", map[string]any{"url": url}, err)
		return "", fmt.Errorf("reading rendered HTML: %w", err)
	}
	// One more chance: if we still see the WAF landing page in the DOM,
	// poll a few more times so Chrome's follow-up request can complete.
	if isAWSWAFChallengeBody(html) {
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			time.Sleep(1 * time.Second)
			html, err = b.Page.HTML()
			if err == nil && !isAWSWAFChallengeBody(html) {
				break
			}
		}
	}
	if isAWSWAFChallengeBody(html) {
		b.Log.Record("waf_still_blocking", map[string]any{"url": url, "bytes": len(html)}, ErrAWSWAFChallenge)
		return "", fmt.Errorf("%w (url=%s)", ErrAWSWAFChallenge, url)
	}
	b.Log.Record("read_html", map[string]any{"url": url, "bytes": len(html)}, nil)
	return html, nil
}

// FetchBookDetails navigates to /book/show/<id> in the browser and returns
// the parsed bibliographic record. Prefer this over Client.FetchBookDetails
// — the book-show endpoint has been walled behind AWS WAF since July 2026
// and the plain HTTP client can no longer reach it.
func (b *Browser) FetchBookDetails(bookID string) (Book, error) {
	u := fmt.Sprintf("%s/book/show/%s", BaseURL, bookID)
	html, err := b.FetchRenderedHTML(u)
	if err != nil {
		return Book{}, fmt.Errorf("fetching book %s: %w", bookID, err)
	}
	return ParseBookDetailsFromHTML(html, bookID)
}

// ListShelf navigates through the browser to the logged-in user's shelf
// page and returns the parsed books. Same WAF motivation as
// FetchBookDetails.
func (b *Browser) ListShelf(shelfName string) ([]Book, error) {
	homeHTML, err := b.FetchRenderedHTML(BaseURL + "/")
	if err != nil {
		return nil, fmt.Errorf("fetching home page: %w", err)
	}
	userID, err := ExtractUserIDFromHomeHTML(homeHTML)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s/review/list/%s?shelf=%s&per_page=100", BaseURL, userID, shelfName)
	html, err := b.FetchRenderedHTML(u)
	if err != nil {
		return nil, fmt.Errorf("fetching shelf %q: %w", shelfName, err)
	}
	return ParseShelfHTML(html)
}

// IsLoggedIn checks if the user is logged in by looking for user-specific elements.
func (b *Browser) IsLoggedIn() bool {
	// Look for the user nav dropdown that appears when logged in
	el, err := b.Page.Timeout(3 * time.Second).Element(`a[href*="/user/show/"], .dropdown--profileMenu, .siteHeader__personal a[href*="/review/list"]`)
	return err == nil && el != nil
}

// SaveCookies persists browser cookies to the session file.
func (b *Browser) SaveCookies() error {
	cookies, err := b.Page.Cookies([]string{"https://www.goodreads.com"})
	if err != nil {
		return fmt.Errorf("getting cookies: %w", err)
	}

	data, err := json.Marshal(cookies)
	if err != nil {
		return fmt.Errorf("marshaling cookies: %w", err)
	}

	return os.WriteFile(SessionPath(), data, 0600)
}

// LoadCookies loads cookies from the session file into the browser.
func (b *Browser) LoadCookies() error {
	data, err := os.ReadFile(SessionPath())
	if err != nil {
		return err
	}

	var cookies []*proto.NetworkCookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return fmt.Errorf("unmarshaling cookies: %w", err)
	}

	for _, c := range cookies {
		err := b.Page.SetCookies([]*proto.NetworkCookieParam{{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
			Expires:  proto.TimeSinceEpoch(c.Expires),
		}})
		if err != nil {
			return fmt.Errorf("setting cookie %s: %w", c.Name, err)
		}
	}

	return nil
}
