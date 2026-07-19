package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Login authenticates with Goodreads via Amazon's OpenID flow using rod.
//
// Selector timeouts are generous (30s) because Goodreads' sign-in page can be
// slow under load, occasionally rate-limits repeated logins, and runs a
// Cloudflare challenge whose splash holds the SSO buttons off-DOM until it
// clears. A 10s timeout (the original value) flaked frequently on consecutive
// test runs.
func Login(b *Browser, cfg *Config) error {
	// Navigate to Goodreads sign-in
	signInURL := "https://www.goodreads.com/user/sign_in"
	b.Log.Record("navigate", map[string]any{"url": signInURL, "purpose": "login"}, nil)
	b.Page.MustNavigate(signInURL)
	b.Page.MustWaitStable()

	// Click the "Sign in with email" button to go to Amazon's login form
	signInBtn, err := b.Page.Timeout(30 * time.Second).Element(`.authPortalSignInButton`)
	b.Log.Record("find_signin_button", map[string]any{"selector": ".authPortalSignInButton"}, err)
	if err != nil {
		saveDebugArtifacts(b)
		return fmt.Errorf("could not find 'Sign in with email' button: %w", err)
	}
	signInBtn.MustClick()
	b.Page.MustWaitStable()

	// Wait for the Amazon login form (ap_ prefixed IDs are Amazon's)
	emailField, err := b.Page.Timeout(30 * time.Second).Element(`#ap_email, input[name="email"], input[type="email"]`)
	if err != nil {
		saveDebugArtifacts(b)
		return fmt.Errorf("could not find email field — run with --no-headless to debug: %w", err)
	}

	emailField.MustSelectAllText().MustInput(cfg.Email)

	passwordField, err := b.Page.Timeout(5 * time.Second).Element(`#ap_password, input[name="password"], input[type="password"]`)
	if err != nil {
		saveDebugArtifacts(b)
		return fmt.Errorf("could not find password field: %w", err)
	}
	passwordField.MustSelectAllText().MustInput(cfg.Password)

	// Submit the form
	submitBtn, err := b.Page.Timeout(5 * time.Second).Element(`#signInSubmit, input[type="submit"], button[type="submit"]`)
	if err != nil {
		saveDebugArtifacts(b)
		return fmt.Errorf("could not find submit button: %w", err)
	}
	submitBtn.MustClick()

	// Wait for redirect back to Goodreads
	b.Page.MustWaitStable()
	time.Sleep(3 * time.Second)

	// Verify login succeeded
	if !b.IsLoggedIn() {
		saveDebugArtifacts(b)
		return fmt.Errorf("login failed — check your credentials in %s, or run with --no-headless to check for CAPTCHA/2FA", ConfigPath())
	}

	return b.SaveCookies()
}

// saveDebugArtifacts writes as much of the failure context to disk as it
// can — screenshot, current HTML, and the in-memory interaction log —
// treating each artifact independently. Previously the screenshot early-
// returned on any error, so a Cloudflare interstitial that blocked
// screenshots also silently dropped the HTML and log; the user opening a
// bug report saw an empty ~/goodreads-cli-debug.* set. Each artifact now
// prints its own outcome so the user knows exactly which files landed.
func saveDebugArtifacts(b *Browser) {
	home, _ := os.UserHomeDir()

	pngPath := filepath.Join(home, "goodreads-cli-debug.png")
	if data, err := b.Page.Screenshot(true, nil); err == nil {
		if werr := os.WriteFile(pngPath, data, 0600); werr == nil {
			fmt.Fprintf(os.Stderr, "Debug screenshot saved to %s\n", pngPath)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", pngPath, werr)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Debug screenshot unavailable: %v\n", err)
	}

	htmlPath := filepath.Join(home, "goodreads-cli-debug.html")
	if html, err := b.Page.HTML(); err == nil {
		if werr := os.WriteFile(htmlPath, []byte(html), 0600); werr == nil {
			fmt.Fprintf(os.Stderr, "Debug HTML saved to %s\n", htmlPath)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", htmlPath, werr)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Debug HTML unavailable: %v\n", err)
	}

	logPath := DebugLogPath()
	if err := b.Log.Dump(logPath); err == nil {
		fmt.Fprintf(os.Stderr, "Interaction log saved to %s — attach to bug report\n", logPath)
	} else {
		fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", logPath, err)
	}
}

