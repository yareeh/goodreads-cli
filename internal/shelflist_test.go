package internal

import (
	"errors"
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

// TestIsAWSWAFChallengeBody documents the AWS WAF landing-page signature we
// key off of. Goodreads started serving these (status 202) on the
// /review/list/… shelf endpoint in July 2026, so the pure-HTTP fetch path
// needs to detect them and surface ErrAWSWAFChallenge to callers instead
// of leaking through as a generic "status 202".
func TestIsAWSWAFChallengeBody(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "aws waf challenge with gokuProps",
			body: `<!DOCTYPE html><html><head><script>window.awsWafCookieDomainList = []; window.gokuProps = {"key":"..."};</script></head></html>`,
			want: true,
		},
		{
			name: "aws waf challenge with only awsWafCookieDomainList marker",
			body: `<script>window.awsWafCookieDomainList = [".goodreads.com"];</script>`,
			want: true,
		},
		{
			name: "regular goodreads shelf HTML",
			body: `<html><body><tr id="review_1" class="bookalike review">…</tr></body></html>`,
			want: false,
		},
		{
			name: "empty body",
			body: "",
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isAWSWAFChallengeBody(c.body); got != c.want {
				t.Errorf("isAWSWAFChallengeBody(...) = %v, want %v", got, c.want)
			}
		})
	}
}

func shelfRow(id string) string {
	return `<tr id="review_` + id + `" class="bookalike review"><td class="field title"><a title="T` + id +
		`" href="/book/show/` + id + `"></a></td><td class="field author"><a href="/author/show/1">A</a></td>` +
		`<div data-resource-id="` + id + `"></div></tr>`
}

func TestCollectShelfPages_FollowsPagesUntilEmpty(t *testing.T) {
	pages := map[int]string{
		1: shelfRow("1") + shelfRow("2"),
		2: shelfRow("3") + shelfRow("4"),
		3: shelfRow("5"),
	}
	var asked []int
	books, err := collectShelfPages(func(page int) (string, error) {
		asked = append(asked, page)
		return pages[page], nil // page 4 is empty
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 5 {
		t.Fatalf("got %d books, want 5", len(books))
	}
	if books[0].ID != "1" || books[4].ID != "5" {
		t.Errorf("order wrong: %v", books)
	}
	if len(asked) != 4 {
		t.Errorf("asked pages %v, want 1..4", asked)
	}
}

func TestCollectShelfPages_StopsWhenPageRepeats(t *testing.T) {
	// Goodreads serves the last page again for an out-of-range page number.
	books, err := collectShelfPages(func(page int) (string, error) {
		if page == 1 {
			return shelfRow("1") + shelfRow("2"), nil
		}
		return shelfRow("2"), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 2 {
		t.Fatalf("got %d books, want 2 (no duplicates)", len(books))
	}
}

func TestCollectShelfPages_ErrorOnLaterPageIsReturned(t *testing.T) {
	_, err := collectShelfPages(func(page int) (string, error) {
		if page == 2 {
			return "", errors.New("boom")
		}
		return shelfRow("1"), nil
	})
	if err == nil {
		t.Fatal("want error, a truncated shelf must not look complete")
	}
}

func TestShelfURL_HasPage(t *testing.T) {
	got := shelfURL("42", "read", 3)
	want := BaseURL + "/review/list/42?shelf=read&per_page=100&page=3"
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestShelfURL_EscapesShelfName(t *testing.T) {
	got := shelfURL("42", "audio books & essays", 1)
	want := BaseURL + "/review/list/42?shelf=audio+books+%26+essays&per_page=100&page=1"
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

// Goodreads' slug for the Want to Read shelf is "to-read"; the CLI name
// "want-to-read" must list that shelf.
func TestShelfURL_WantToReadUsesGoodreadsSlug(t *testing.T) {
	got := shelfURL("42", "want-to-read", 1)
	want := BaseURL + "/review/list/42?shelf=to-read&per_page=100&page=1"
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}
