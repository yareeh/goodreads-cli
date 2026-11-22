// Command headed-login logs in to Goodreads in a visible browser window so a
// human can solve a CAPTCHA or 2FA prompt, then saves the session cookies to
// the integration tests' session store (.goodreads-cli-session in the repo
// root, gitignored — never the user's ~/.goodreads-cli-session). The tests
// reuse that session instead of logging in again, getting past the bot check.
//
// Credentials come from GOODREADS_EMAIL / GOODREADS_PASSWORD, or from a .env
// file in the current directory (existing env vars win). Run from the repo root:
//
//	go run ./cmd/headed-login            # wait up to 5 minutes
//	go run ./cmd/headed-login -wait 10m
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yareeh/goodreads-cli/internal"
)

func main() {
	wait := flag.Duration("wait", 5*time.Minute, "how long to wait for the login to complete")
	envFile := flag.String("env", ".env", "dotenv file to read credentials from")
	flag.Parse()

	loadDotEnv(*envFile)
	if os.Getenv("GOODREADS_SESSION_PATH") == "" {
		// Same store as TestMain in integration_test.go.
		p, err := filepath.Abs(".goodreads-cli-session")
		if err != nil {
			fmt.Fprintln(os.Stderr, "resolve session path:", err)
			os.Exit(1)
		}
		os.Setenv("GOODREADS_SESSION_PATH", p)
	}

	b, err := internal.NewBrowser(false) // headed
	if err != nil {
		fmt.Fprintln(os.Stderr, "NewBrowser:", err)
		os.Exit(1)
	}
	defer b.Close()

	// A headed Login fills the form, then waits for the human to finish any
	// CAPTCHA/2FA in the window.
	internal.ManualLoginTimeout = *wait
	cfg := &internal.Config{Email: os.Getenv("GOODREADS_EMAIL"), Password: os.Getenv("GOODREADS_PASSWORD")}
	if err := internal.Login(b, cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Logged in — cookies saved to", internal.SessionPath())
}

// loadDotEnv sets KEY=VALUE pairs from path as env vars, without overriding
// variables that are already set. A missing file is not an error.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok && os.Getenv(k) == "" {
			os.Setenv(strings.TrimSpace(k), strings.Trim(strings.TrimSpace(v), `"'`))
		}
	}
}
