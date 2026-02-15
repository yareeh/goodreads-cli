package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Login authenticates with Goodreads via Amazon's OpenID flow using rod.
func Login(b *Browser, cfg *Config) error {
	// Navigate to Goodreads sign-in
	b.Page.MustNavigate("https://www.goodreads.com/user/sign_in")
	b.Page.MustWaitStable()

	// Click the "Sign in with email" button to go to Amazon's login form
	signInBtn, err := b.Page.Timeout(10 * time.Second).Element(`.authPortalSignInButton`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find 'Sign in with email' button: %w", err)
	}
	signInBtn.MustClick()
	b.Page.MustWaitStable()

	// Wait for the Amazon login form (ap_ prefixed IDs are Amazon's)
	emailField, err := b.Page.Timeout(15 * time.Second).Element(`#ap_email, input[name="email"], input[type="email"]`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find email field — run with --no-headless to debug: %w", err)
	}

	emailField.MustSelectAllText().MustInput(cfg.Email)

	passwordField, err := b.Page.Timeout(5 * time.Second).Element(`#ap_password, input[name="password"], input[type="password"]`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find password field: %w", err)
	}
	passwordField.MustSelectAllText().MustInput(cfg.Password)

	// Submit the form
	submitBtn, err := b.Page.Timeout(5 * time.Second).Element(`#signInSubmit, input[type="submit"], button[type="submit"]`)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find submit button: %w", err)
	}
	submitBtn.MustClick()

	// Wait for redirect back to Goodreads
	b.Page.MustWaitStable()
	time.Sleep(3 * time.Second)

	// Verify login succeeded
	if !b.IsLoggedIn() {
		saveDebugScreenshot(b)
		return fmt.Errorf("login failed — check your credentials in %s, or run with --no-headless to check for CAPTCHA/2FA", ConfigPath())
	}

	return b.SaveCookies()
}

func saveDebugScreenshot(b *Browser) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "goodreads-cli-debug.png")
	data, err := b.Page.Screenshot(true, nil)
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0644)
	fmt.Fprintf(os.Stderr, "Debug screenshot saved to %s\n", path)
}
