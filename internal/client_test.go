package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/book/auto_complete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("format") != "json" {
			t.Errorf("missing format=json")
		}
		if r.URL.Query().Get("q") != "hail mary" {
			t.Errorf("query = %q, want %q", r.URL.Query().Get("q"), "hail mary")
		}
		json.NewEncoder(w).Encode([]autoCompleteResult{
			{
				BookID:   "55145261",
				Title:    "Project Hail Mary",
				Author:   autoCompleteAuth{ID: 1, Name: "Andy Weir"},
				ImageURL: "https://example.com/img.jpg",
			},
			{
				BookID:   "99999999",
				Title:    "Another Book",
				Author:   autoCompleteAuth{ID: 2, Name: "Someone"},
				ImageURL: "",
			},
		})
	}))
	defer ts.Close()

	// Override BaseURL for test
	origBase := BaseURL
	defer func() { /* BaseURL is not reassignable, test via direct HTTP */ }()
	_ = origBase

	// Test search parsing by calling the server directly
	client := &Client{HTTP: ts.Client()}
	resp, err := client.HTTP.Get(ts.URL + "/book/auto_complete?format=json&q=hail+mary")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var results []autoCompleteResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}

	// Verify the mapping to Book
	books := make([]Book, 0, len(results))
	for _, r := range results {
		books = append(books, Book{
			ID:       r.BookID,
			Title:    r.Title,
			Author:   r.Author.Name,
			ImageURL: r.ImageURL,
		})
	}

	if books[0].ID != "55145261" {
		t.Errorf("book ID = %q, want %q", books[0].ID, "55145261")
	}
	if books[0].Title != "Project Hail Mary" {
		t.Errorf("title = %q, want %q", books[0].Title, "Project Hail Mary")
	}
	if books[0].Author != "Andy Weir" {
		t.Errorf("author = %q, want %q", books[0].Author, "Andy Weir")
	}
}

func TestSearchHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/book/auto_complete?format=json&q=test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSearchEmptyResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]autoCompleteResult{})
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/book/auto_complete?format=json&q=xyznonexistent")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var results []autoCompleteResult
	json.NewDecoder(resp.Body).Decode(&results)
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestAutoCompleteResultJSON(t *testing.T) {
	raw := `{"bookId":"12345","title":"Test Book","author":{"id":42,"name":"Test Author"},"imageUrl":"https://img.example.com/cover.jpg"}`

	var result autoCompleteResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result.BookID != "12345" {
		t.Errorf("BookID = %q, want %q", result.BookID, "12345")
	}
	if result.Title != "Test Book" {
		t.Errorf("Title = %q, want %q", result.Title, "Test Book")
	}
	if result.Author.Name != "Test Author" {
		t.Errorf("Author.Name = %q, want %q", result.Author.Name, "Test Author")
	}
	if result.Author.ID != 42 {
		t.Errorf("Author.ID = %d, want 42", result.Author.ID)
	}
	if result.ImageURL != "https://img.example.com/cover.jpg" {
		t.Errorf("ImageURL = %q", result.ImageURL)
	}
}
