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

// Browser-based tests require GOODREADS_SESSION_COOKIES env var
// containing base64-encoded rod session cookies, plus headless Chrome.
// These are skipped if the env var is not set.

func setupBrowser(t *testing.T) *internal.Browser {
	t.Helper()
	cookies := os.Getenv("GOODREADS_SESSION_COOKIES")
	if cookies == "" {
		t.Skip("GOODREADS_SESSION_COOKIES not set, skipping browser test")
	}

	// Write cookies to temp session file
	tmpDir := t.TempDir()

	data, err := base64.StdEncoding.DecodeString(cookies)
	if err != nil {
		t.Fatalf("decoding session cookies: %v", err)
	}
	if err := os.WriteFile(tmpDir+"/.goodreads-cli-session", data, 0600); err != nil {
		t.Fatalf("writing session file: %v", err)
	}

	// Override HOME so Browser loads our temp cookies
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	browser, err := internal.NewBrowser(true) // headless
	if err != nil {
		t.Fatalf("NewBrowser: %v", err)
	}
	t.Cleanup(func() { browser.Close() })

	if err := browser.LoadCookies(); err != nil {
		t.Fatalf("LoadCookies: %v", err)
	}

	return browser
}

func TestIntegrationShelfAddAndRemove(t *testing.T) {
	browser := setupBrowser(t)

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

	// Note: Goodreads doesn't have a simple "remove from all shelves" API via browser.
	// The book stays on the "read" shelf. This is acceptable for testing.
}

func TestIntegrationMarkCurrentlyReading(t *testing.T) {
	browser := setupBrowser(t)

	err := internal.MarkCurrentlyReading(browser, testShelfBookID)
	if err != nil {
		t.Fatalf("MarkCurrentlyReading: %v", err)
	}
	t.Log("Marked book as currently reading")
}
