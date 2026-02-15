package internal

import (
	"fmt"
	"time"
)

// Login authenticates with Goodreads via Amazon's OpenID flow using rod.
func Login(b *Browser, cfg *Config) error {
	// Navigate to Goodreads sign-in, which redirects to Amazon
	b.Page.MustNavigate("https://www.goodreads.com/user/sign_in")
	b.Page.MustWaitStable()

	// Click "Sign in with email" if that option is presented
	signInLink, err := b.Page.Timeout(5 * time.Second).ElementR("a, button", "(?i)sign in with email|Sign in")
	if err == nil {
		signInLink.MustClick()
		b.Page.MustWaitStable()
	}

	// Wait for the Amazon login form
	emailField, err := b.Page.Timeout(10 * time.Second).Element(`input[name="email"], input[type="email"], #ap_email`)
	if err != nil {
		return fmt.Errorf("could not find email field — login page may have changed: %w", err)
	}

	// Fill in credentials
	emailField.MustSelectAllText().MustInput(cfg.Email)

	passwordField, err := b.Page.Timeout(5 * time.Second).Element(`input[name="password"], input[type="password"], #ap_password`)
	if err != nil {
		return fmt.Errorf("could not find password field: %w", err)
	}
	passwordField.MustSelectAllText().MustInput(cfg.Password)

	// Submit the form
	submitBtn, err := b.Page.Timeout(5 * time.Second).Element(`input[type="submit"], #signInSubmit, button[type="submit"]`)
	if err != nil {
		return fmt.Errorf("could not find submit button: %w", err)
	}
	submitBtn.MustClick()

	// Wait for redirect back to Goodreads
	b.Page.MustWaitStable()
	time.Sleep(2 * time.Second)

	// Verify login succeeded
	if !b.IsLoggedIn() {
		return fmt.Errorf("login failed — check your credentials in %s or look for CAPTCHA/2FA requirements", ConfigPath())
	}

	return b.SaveCookies()
}
