package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/go-rod/rod/lib/proto"
)

const BaseURL = "https://www.goodreads.com"

func rodSameSite(ss proto.NetworkCookieSameSite) http.SameSite {
	switch ss {
	case proto.NetworkCookieSameSiteLax:
		return http.SameSiteLaxMode
	case proto.NetworkCookieSameSiteStrict:
		return http.SameSiteStrictMode
	case proto.NetworkCookieSameSiteNone:
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// Client is a plain HTTP client for operations that don't need a browser (e.g. search).
//
// Log mirrors the Browser field of the same name so both interaction paths
// end up in the same JSON tail when the error handler dumps it.
type Client struct {
	HTTP *http.Client
	Log  *InteractionLog
}

// NewClient creates a new HTTP client, loading cookies from the rod session file if available.
func NewClient() (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		HTTP: &http.Client{Jar: jar},
		Log:  NewInteractionLog(),
	}

	// Load cookies saved by the rod browser session
	client.loadRodSession()

	return client, nil
}

// loadRodSession loads rod-format cookies from the session file into the HTTP cookie jar.
func (c *Client) loadRodSession() {
	data, err := os.ReadFile(SessionPath())
	if err != nil {
		return
	}

	// Try rod cookie format first (proto.NetworkCookie)
	var rodCookies []*proto.NetworkCookie
	if err := json.Unmarshal(data, &rodCookies); err == nil && len(rodCookies) > 0 {
		u, _ := url.Parse(BaseURL)
		var httpCookies []*http.Cookie
		for _, rc := range rodCookies {
			httpCookies = append(httpCookies, &http.Cookie{ // #nosec G124 -- preserving original browser cookie attributes
				Name:     rc.Name,
				Value:    rc.Value,
				Domain:   rc.Domain,
				Path:     rc.Path,
				Secure:   rc.Secure,
				HttpOnly: rc.HTTPOnly,
				SameSite: rodSameSite(rc.SameSite),
			})
		}
		c.HTTP.Jar.SetCookies(u, httpCookies)
		return
	}

	// Fallback: try plain http.Cookie format
	var httpCookies []*http.Cookie
	if err := json.Unmarshal(data, &httpCookies); err == nil {
		u, _ := url.Parse(BaseURL)
		c.HTTP.Jar.SetCookies(u, httpCookies)
	}
}

// Search calls the Goodreads autocomplete endpoint and returns matching books.
func (c *Client) Search(query string) ([]Book, error) {
	reqURL := fmt.Sprintf("%s/book/auto_complete?format=json&q=%s", BaseURL, url.QueryEscape(query))

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating search request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		c.Log.Record("http_search", map[string]any{"url": reqURL}, err)
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()
	c.Log.Record("http_search", map[string]any{"url": reqURL, "status": resp.StatusCode}, nil)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	var results []autoCompleteResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("parsing search results: %w", err)
	}

	var books []Book
	for _, r := range results {
		books = append(books, Book{
			ID:       r.BookID,
			Title:    r.Title,
			Author:   r.Author.Name,
			ImageURL: r.ImageURL,
		})
	}
	return books, nil
}

type autoCompleteResult struct {
	BookID   string           `json:"bookId"`
	Title    string           `json:"title"`
	Author   autoCompleteAuth `json:"author"`
	ImageURL string           `json:"imageUrl"`
}

type autoCompleteAuth struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// FetchBookDetails downloads a Goodreads book page by legacy ID and extracts
// the full bibliographic record (ISBN, publisher, edition year, original
// title, language, page count, …) from the embedded structured data.
func (c *Client) FetchBookDetails(bookID string) (Book, error) {
	reqURL := fmt.Sprintf("%s/book/show/%s", BaseURL, bookID)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return Book{}, fmt.Errorf("creating book request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		c.Log.Record("http_book_details", map[string]any{"url": reqURL}, err)
		return Book{}, fmt.Errorf("book request failed: %w", err)
	}
	defer resp.Body.Close()
	c.Log.Record("http_book_details", map[string]any{"url": reqURL, "status": resp.StatusCode}, nil)

	if resp.StatusCode != http.StatusOK {
		return Book{}, fmt.Errorf("book page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Book{}, fmt.Errorf("reading book page: %w", err)
	}
	return ParseBookDetailsFromHTML(string(body), bookID)
}
