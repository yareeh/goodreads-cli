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
type Browser struct {
	Rod  *rod.Browser
	Page *rod.Page
}

// NewBrowser launches a Chrome instance and navigates to goodreads.com.
// Set headless to false to see the browser for debugging.
func NewBrowser(headless bool) (*Browser, error) {
	u, err := launcher.New().
		Headless(headless).
		Launch()
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

	b := &Browser{Rod: browser, Page: page}

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
