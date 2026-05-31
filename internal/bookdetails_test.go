package internal

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseBookDetailsFromHTML loads a real Goodreads book page (Finnish
// edition of André Brink's "An Instant in the Wind", legacyId 18690730)
// and asserts that the structured-data parser extracts every field the
// page is known to carry: ISBN-13, publisher, edition year, original
// title (Afrikaans), language, page count, format, and the canonical
// goodreads URL.
//
// The Finnish edition has ISBN 9789510085660 and was published in 1978
// by WSOY; the original Afrikaans work is "'n Oomblik in die wind"
// (1975). All of these values appear in JSON-LD or the embedded
// __NEXT_DATA__ Apollo state on the page.
func TestParseBookDetailsFromHTML(t *testing.T) {
	path := filepath.Join("testdata", "book_18690730_tuokio_tuulessa.html")
	html, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	got, err := ParseBookDetailsFromHTML(string(html), "18690730")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	want := Book{
		ID:            "18690730",
		Title:         "Tuokio tuulessa",
		Author:        "André Brink",
		URL:           "https://www.goodreads.com/book/show/18690730-tuokio-tuulessa",
		ISBN:          "9510085669",
		ISBN13:        "9789510085660",
		Publisher:     "WSOY",
		OriginalTitle: "’n Oomblik in die wind",
		Year:          "1978",
		Pages:         350,
		Language:      "Finnish",
		Format:        "Hardcover",
	}

	if got.ID != want.ID {
		t.Errorf("ID = %q, want %q", got.ID, want.ID)
	}
	if got.Title != want.Title {
		t.Errorf("Title = %q, want %q", got.Title, want.Title)
	}
	if got.Author != want.Author {
		t.Errorf("Author = %q, want %q", got.Author, want.Author)
	}
	if got.URL != want.URL {
		t.Errorf("URL = %q, want %q", got.URL, want.URL)
	}
	if got.ISBN != want.ISBN {
		t.Errorf("ISBN = %q, want %q", got.ISBN, want.ISBN)
	}
	if got.ISBN13 != want.ISBN13 {
		t.Errorf("ISBN13 = %q, want %q", got.ISBN13, want.ISBN13)
	}
	if got.Publisher != want.Publisher {
		t.Errorf("Publisher = %q, want %q", got.Publisher, want.Publisher)
	}
	if got.OriginalTitle != want.OriginalTitle {
		t.Errorf("OriginalTitle = %q, want %q", got.OriginalTitle, want.OriginalTitle)
	}
	if got.Year != want.Year {
		t.Errorf("Year = %q, want %q", got.Year, want.Year)
	}
	if got.Pages != want.Pages {
		t.Errorf("Pages = %d, want %d", got.Pages, want.Pages)
	}
	if got.Language != want.Language {
		t.Errorf("Language = %q, want %q", got.Language, want.Language)
	}
	if got.Format != want.Format {
		t.Errorf("Format = %q, want %q", got.Format, want.Format)
	}
}
