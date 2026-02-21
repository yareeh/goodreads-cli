package main

import (
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/jari/goodreads-cli/internal"
)

// Test data: well-known Goodreads books
const (
	testBookID    = "55145261"  // Project Hail Mary
	testBookTitle = "Project Hail Mary"
	testBookQuery = "project hail mary andy weir"
	// A safe book to add/remove from shelf during testing
	testShelfBookID = "55145261" // Project Hail Mary
)

// --- Search tests (no auth needed) ---

func TestIntegrationSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	client, err := internal.NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	books, err := client.Search(testBookQuery)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(books) == 0 {
		t.Fatal("Search returned no results")
	}

	found := false
	for _, b := range books {
		if b.ID == testBookID {
			found = true
			if b.Title != testBookTitle {
				t.Errorf("expected title %q, got %q", testBookTitle, b.Title)
			}
			if b.Author == "" {
				t.Error("expected non-empty author")
			}
			break
		}
	}
	if !found {
		t.Logf("Book ID %s not found in results, but got %d results (first: %s by %s)",
			testBookID, len(books), books[0].Title, books[0].Author)
	}
}

func TestIntegrationSearchMultipleQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	client, err := internal.NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	queries := []struct {
		query    string
		minBooks int
	}{
		{"the martian", 1},
		{"harry potter", 1},
		{"neil gaiman", 1},
	}

	for _, q := range queries {
		t.Run(q.query, func(t *testing.T) {
			books, err := client.Search(q.query)
			if err != nil {
				t.Fatalf("Search(%q): %v", q.query, err)
			}
			if len(books) < q.minBooks {
				t.Errorf("Search(%q): got %d results, want at least %d", q.query, len(books), q.minBooks)
			}
		})
	}
}

func TestIntegrationSearchNoResults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	client, err := internal.NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	books, err := client.Search("xyznonexistentbook99999zzz")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	t.Logf("Search for nonsense returned %d results", len(books))
}

func TestIntegrationSearchResultFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	client, err := internal.NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	books, err := client.Search("project hail mary")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(books) == 0 {
		t.Fatal("Search returned no results")
	}

	b := books[0]
	if b.ID == "" {
		t.Error("expected non-empty book ID")
	}
	if b.Title == "" {
		t.Error("expected non-empty title")
	}
	if b.Author == "" {
		t.Error("expected non-empty author")
	}
}

// --- Config tests ---

func TestIntegrationLoadConfigFromEnv(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Set env vars
	origEmail := os.Getenv("GOODREADS_EMAIL")
	origPass := os.Getenv("GOODREADS_PASSWORD")
	os.Setenv("GOODREADS_EMAIL", "test@example.com")
	os.Setenv("GOODREADS_PASSWORD", "secret123")
	defer func() {
		if origEmail != "" {
			os.Setenv("GOODREADS_EMAIL", origEmail)
		} else {
			os.Unsetenv("GOODREADS_EMAIL")
		}
		if origPass != "" {
			os.Setenv("GOODREADS_PASSWORD", origPass)
		} else {
			os.Unsetenv("GOODREADS_PASSWORD")
		}
	}()

	cfg, err := internal.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", cfg.Email, "test@example.com")
	}
	if cfg.Password != "secret123" {
		t.Errorf("Password = %q, want %q", cfg.Password, "secret123")
	}
}

func TestIntegrationLogout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a temp session file
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	sessionPath := internal.SessionPath()
	os.WriteFile(sessionPath, []byte("fake-cookies"), 0600)

	if err := internal.Logout(); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	if _, err := os.Stat(sessionPath); !os.IsNotExist(err) {
		t.Error("session file still exists after logout")
	}
}

// --- Browser-based tests ---
// Require GOODREADS_EMAIL + GOODREADS_PASSWORD env vars, or
// GOODREADS_SESSION_COOKIES (base64-encoded rod cookies).
// Also need headless Chrome available.

func setupBrowserWithLogin(t *testing.T) *internal.Browser {
	t.Helper()

	email := os.Getenv("GOODREADS_EMAIL")
	password := os.Getenv("GOODREADS_PASSWORD")
	cookies := os.Getenv("GOODREADS_SESSION_COOKIES")

	if email == "" && cookies == "" {
		t.Skip("GOODREADS_EMAIL or GOODREADS_SESSION_COOKIES not set, skipping browser test")
	}

	browser, err := internal.NewBrowser(true) // headless
	if err != nil {
		t.Fatalf("NewBrowser: %v", err)
	}
	t.Cleanup(func() { browser.Close() })

	// Try loading existing cookies first
	if cookies != "" {
		tmpDir := t.TempDir()
		data, err := base64.StdEncoding.DecodeString(cookies)
		if err != nil {
			t.Fatalf("decoding session cookies: %v", err)
		}
		origHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		t.Cleanup(func() { os.Setenv("HOME", origHome) })
		os.WriteFile(internal.SessionPath(), data, 0600)
		if err := browser.LoadCookies(); err != nil {
			t.Fatalf("LoadCookies: %v", err)
		}
		if browser.IsLoggedIn() {
			return browser
		}
		t.Log("Stored cookies expired, falling back to login")
		os.Setenv("HOME", origHome)
	}

	// Login with credentials
	if email == "" || password == "" {
		t.Skip("cookies expired and GOODREADS_EMAIL/GOODREADS_PASSWORD not set")
	}

	cfg := &internal.Config{Email: email, Password: password}
	if err := internal.Login(browser, cfg); err != nil {
		t.Fatalf("Login: %v", err)
	}

	return browser
}

func TestIntegrationLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	email := os.Getenv("GOODREADS_EMAIL")
	password := os.Getenv("GOODREADS_PASSWORD")
	if email == "" || password == "" {
		t.Skip("GOODREADS_EMAIL and GOODREADS_PASSWORD not set, skipping login test")
	}

	browser, err := internal.NewBrowser(true)
	if err != nil {
		t.Fatalf("NewBrowser: %v", err)
	}
	defer browser.Close()

	cfg := &internal.Config{Email: email, Password: password}
	if err := internal.Login(browser, cfg); err != nil {
		t.Fatalf("Login: %v", err)
	}

	if !browser.IsLoggedIn() {
		t.Error("expected to be logged in after Login()")
	}
	t.Log("Login successful")
}

func TestIntegrationShelfAddAndRemove(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	browser := setupBrowserWithLogin(t)

	// Add book to want-to-read shelf
	err := internal.AddToShelf(browser, testShelfBookID, "want-to-read")
	if err != nil {
		t.Fatalf("AddToShelf(want-to-read): %v", err)
	}
	t.Log("Added book to want-to-read shelf")

	time.Sleep(2 * time.Second)

	// Move to read shelf (also tests changing shelves)
	err = internal.AddToShelf(browser, testShelfBookID, "read")
	if err != nil {
		t.Fatalf("AddToShelf(read): %v", err)
	}
	t.Log("Moved book to read shelf")
}

func TestIntegrationMarkCurrentlyReading(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	browser := setupBrowserWithLogin(t)

	err := internal.MarkCurrentlyReading(browser, testShelfBookID)
	if err != nil {
		t.Fatalf("MarkCurrentlyReading: %v", err)
	}
	t.Log("Marked book as currently reading")
}
