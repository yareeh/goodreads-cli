package internal

import (
	"os"
	"testing"
)

func TestParseShelfHTML_ExtractsBookFromCurrentlyReading(t *testing.T) {
	data, err := os.ReadFile("testdata/shelf_currently_reading.html")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	books, err := ParseShelfHTML(string(data))
	if err != nil {
		t.Fatalf("ParseShelfHTML: %v", err)
	}
	if len(books) == 0 {
		t.Fatal("expected at least one book in the currently-reading fixture")
	}

	want := Book{
		ID:     "55145261",
		Title:  "The Anthropocene Reviewed: Essays on a Human-Centered Planet",
		Author: "Green, John",
	}
	var found *Book
	for i := range books {
		if books[i].ID == want.ID {
			found = &books[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("book %q not in parsed result; got %d books", want.ID, len(books))
	}
	if found.Title != want.Title {
		t.Errorf("Title = %q, want %q", found.Title, want.Title)
	}
	if found.Author != want.Author {
		t.Errorf("Author = %q, want %q", found.Author, want.Author)
	}
}

func TestParseShelfHTML_EmptyShelfReturnsEmptySlice(t *testing.T) {
	html := `<html><body><div id="bookShelf">No books on this shelf.</div></body></html>`
	books, err := ParseShelfHTML(html)
	if err != nil {
		t.Fatalf("ParseShelfHTML on empty shelf: %v", err)
	}
	if len(books) != 0 {
		t.Errorf("want empty slice on empty shelf, got %d books", len(books))
	}
}

func TestExtractUserIDFromHomeHTML(t *testing.T) {
	// The signed-in home page links to /user/show/<id>-<slug> in the header
	// avatar / profile menu. The first such occurrence is the logged-in user.
	cases := []struct {
		name string
		html string
		want string
	}{
		{
			name: "user_id from profile link with slug",
			html: `<a href="/user/show/199003311-skye-claw">Profile</a>`,
			want: "199003311",
		},
		{
			name: "user_id from profile link without slug",
			html: `<a href="/user/show/12345" class="userPic">…</a>`,
			want: "12345",
		},
		{
			name: "skips earlier non-numeric matches",
			html: `<a href="/help/user/show/x">help</a><a href="/user/show/777-me">me</a>`,
			want: "777",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ExtractUserIDFromHomeHTML(c.html)
			if err != nil {
				t.Fatalf("ExtractUserIDFromHomeHTML: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestExtractUserIDFromHomeHTML_NoMatchReturnsError(t *testing.T) {
	_, err := ExtractUserIDFromHomeHTML(`<html><body>not logged in</body></html>`)
	if err == nil {
		t.Fatal("expected error when no /user/show/ link is present")
	}
}
