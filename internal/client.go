package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

const BaseURL = "https://www.goodreads.com"

type Client struct {
	HTTP *http.Client
}

func NewClient() (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		HTTP: &http.Client{Jar: jar},
	}

	if err := client.loadSession(); err == nil {
		// Session loaded successfully
	}

	return client, nil
}

func (c *Client) HasSession() bool {
	u, _ := url.Parse(BaseURL)
	return len(c.HTTP.Jar.Cookies(u)) > 0
}

func (c *Client) SaveSession() error {
	u, _ := url.Parse(BaseURL)
	cookies := c.HTTP.Jar.Cookies(u)

	data, err := json.Marshal(cookies)
	if err != nil {
		return err
	}

	return os.WriteFile(SessionPath(), data, 0600)
}

func (c *Client) loadSession() error {
	data, err := os.ReadFile(SessionPath())
	if err != nil {
		return err
	}

	var cookies []*http.Cookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return err
	}

	u, _ := url.Parse(BaseURL)
	c.HTTP.Jar.SetCookies(u, cookies)
	return nil
}

func (c *Client) NewRequest(method, path string) (*http.Request, error) {
	req, err := http.NewRequest(method, BaseURL+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}

	if err := c.SaveSession(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to save session: %v\n", err)
	}

	return resp, nil
}
