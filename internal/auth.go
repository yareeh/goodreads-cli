package internal

import (
	"fmt"
	"io"
	"net/url"
	"strings"
)

func Login(client *Client, cfg *Config) error {
	// Step 1: GET the sign-in page to get CSRF token and cookies
	req, err := client.NewRequest("GET", "/user/sign_in")
	if err != nil {
		return fmt.Errorf("creating sign-in request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetching sign-in page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading sign-in page: %w", err)
	}

	// Extract CSRF token (authenticity_token) from the page
	token := extractCSRFToken(string(body))
	if token == "" {
		return fmt.Errorf("could not find CSRF token on sign-in page — Goodreads may have changed their login flow.\nUse the recorder tool to investigate: go run ./cmd/recorder")
	}

	// Step 2: POST login credentials
	form := url.Values{}
	form.Set("user[email]", cfg.Email)
	form.Set("user[password]", cfg.Password)
	form.Set("authenticity_token", token)

	loginReq, err := client.NewRequest("POST", "/user/sign_in")
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}

	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Body = io.NopCloser(strings.NewReader(form.Encode()))
	loginReq.ContentLength = int64(len(form.Encode()))

	loginResp, err := client.Do(loginReq)
	if err != nil {
		return fmt.Errorf("submitting login: %w", err)
	}
	defer loginResp.Body.Close()

	// Check if login succeeded by looking at redirect or session
	if loginResp.StatusCode >= 400 {
		return fmt.Errorf("login failed with status %d — check your credentials in %s", loginResp.StatusCode, ConfigPath())
	}

	if !client.HasSession() {
		return fmt.Errorf("login appeared to succeed but no session cookies were set")
	}

	return client.SaveSession()
}

func extractCSRFToken(html string) string {
	// Look for <input name="authenticity_token" ... value="...">
	needle := `name="authenticity_token"`
	idx := strings.Index(html, needle)
	if idx == -1 {
		return ""
	}

	// Search for value= near the token input
	sub := html[max(0, idx-200) : min(len(html), idx+200)]
	valIdx := strings.Index(sub, `value="`)
	if valIdx == -1 {
		return ""
	}

	start := valIdx + len(`value="`)
	end := strings.Index(sub[start:], `"`)
	if end == -1 {
		return ""
	}

	return sub[start : start+end]
}
