package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/go-rod/rod/lib/proto"
)

const BaseURL = "https://www.goodreads.com"

// Client is a plain HTTP client for operations that don't need a browser (e.g. search).
type Client struct {
	HTTP *http.Client
}

// NewClient creates a new HTTP client, loading cookies from the rod session file if available.
func NewClient() (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		HTTP: &http.Client{Jar: jar},
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
			httpCookies = append(httpCookies, &http.Cookie{
				Name:     rc.Name,
				Value:    rc.Value,
				Domain:   rc.Domain,
				Path:     rc.Path,
				Secure:   rc.Secure,
				HttpOnly: rc.HTTPOnly,
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
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

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
