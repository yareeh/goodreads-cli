package internal

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestIsBuiltinShelf(t *testing.T) {
	for _, s := range []string{"want-to-read", "currently-reading", "read"} {
		if !isBuiltinShelf(s) {
			t.Errorf("isBuiltinShelf(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"ebook", "did-not-finish", "to-read", ""} {
		if isBuiltinShelf(s) {
			t.Errorf("isBuiltinShelf(%q) = true, want false", s)
		}
	}
}

// The shelf list page of the test account carries the shelf chooser config:
// every shelf of the user and which of them are exclusive.
func TestParseShelfChooser_Fixture(t *testing.T) {
	html, err := os.ReadFile("testdata/shelf_currently_reading.html")
	if err != nil {
		t.Fatal(err)
	}
	set, err := ParseShelfChooser(string(html))
	if err != nil {
		t.Fatal(err)
	}
	wantAll := []string{"to-read", "currently-reading", "read", "did-not-finish"}
	if strings.Join(set.All, ",") != strings.Join(wantAll, ",") {
		t.Errorf("All = %v, want %v", set.All, wantAll)
	}
	if !set.IsExclusive("did-not-finish") {
		t.Error("did-not-finish should be exclusive")
	}
}

func TestParseShelfChooser_CustomShelves(t *testing.T) {
	html := `<script>window.shelfChooser = new ShelfChooser(
    "shelfChooserInput",
    0,
    ["to-read", "currently-reading", "read", "ebook", "audio books"],
    {
      chosen: [],
      exclusive: ["to-read", "currently-reading", "read"],
    });</script>`
	set, err := ParseShelfChooser(html)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.All) != 5 || set.All[4] != "audio books" {
		t.Errorf("All = %v", set.All)
	}
	if set.IsExclusive("ebook") {
		t.Error("ebook should not be exclusive")
	}
}

func TestParseShelfChooser_Missing(t *testing.T) {
	if _, err := ParseShelfChooser("<html>no chooser</html>"); err == nil {
		t.Fatal("want error when the page has no shelf chooser")
	}
}

func TestResolveShelf(t *testing.T) {
	set := ShelfSet{All: []string{"to-read", "read", "ebook", "Audiobook"}}
	cases := []struct {
		in, want string
		wantErr  bool
	}{
		{"ebook", "ebook", false},
		{"EBook", "ebook", false},         // Goodreads shelf names are case-insensitive
		{"audiobook", "Audiobook", false}, // canonical spelling is returned
		{"ebok", "", true},                // a typo must not create a new shelf
	}
	for _, c := range cases {
		got, err := set.Resolve(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("Resolve(%q) err = %v, wantErr %v", c.in, err, c.wantErr)
			continue
		}
		if got != c.want {
			t.Errorf("Resolve(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	_, err := set.Resolve("ebok")
	if err == nil || !strings.Contains(err.Error(), "ebook") {
		t.Errorf("error should list the existing shelves, got %v", err)
	}
}

// The book ID and shelf name are embedded as JSON string literals, so no
// value can break out of the string into the script.
func TestCustomShelfAddJS_EscapesValues(t *testing.T) {
	evil := `x"; alert(1); //` + " </script>"
	js := customShelfAddJS("123", evil)
	m := regexp.MustCompile(`const shelfName = (.*);\n`).FindStringSubmatch(js)
	if m == nil {
		t.Fatalf("shelfName assignment not found in:\n%s", js)
	}
	var got string
	if err := json.Unmarshal([]byte(m[1]), &got); err != nil {
		t.Fatalf("shelf name literal %s is not a single JSON string: %v", m[1], err)
	}
	if got != evil {
		t.Errorf("round-trip = %q, want %q", got, evil)
	}
	for _, needle := range []string{`/shelf/add_to_shelf`, `authenticity_token`, `credentials: 'same-origin'`} {
		if !strings.Contains(js, needle) {
			t.Errorf("customShelfAddJS missing %q", needle)
		}
	}
}

func TestValidBookID(t *testing.T) {
	for _, id := range []string{"1", "57933306"} {
		if !validBookID(id) {
			t.Errorf("validBookID(%q) = false", id)
		}
	}
	for _, id := range []string{"", "12a", "../x", "1 2", "57933306.Title"} {
		if validBookID(id) {
			t.Errorf("validBookID(%q) = true", id)
		}
	}
}

func TestShelfHasBook_StopsWhenFound(t *testing.T) {
	asked := 0
	found, err := shelfHasBook("2", func(page int) (string, error) {
		asked++
		return shelfRow("1") + shelfRow("2"), nil
	})
	if err != nil || !found {
		t.Fatalf("found=%v err=%v, want found", found, err)
	}
	if asked != 1 {
		t.Errorf("fetched %d pages, want 1 (stop as soon as the book is seen)", asked)
	}
}

func TestShelfHasBook_ReadsLaterPages(t *testing.T) {
	pages := map[int]string{1: shelfRow("1"), 2: shelfRow("2"), 3: shelfRow("3")}
	found, err := shelfHasBook("3", func(page int) (string, error) { return pages[page], nil })
	if err != nil || !found {
		t.Fatalf("found=%v err=%v, want found on page 3", found, err)
	}
}

func TestShelfHasBook_NotFound(t *testing.T) {
	found, err := shelfHasBook("9", func(page int) (string, error) {
		if page == 1 {
			return shelfRow("1"), nil
		}
		return shelfRow("1"), nil // Goodreads repeats the last page
	})
	if err != nil || found {
		t.Fatalf("found=%v err=%v, want not found", found, err)
	}
}

func TestShelfHasBook_Error(t *testing.T) {
	_, err := shelfHasBook("9", func(page int) (string, error) { return "", errors.New("boom") })
	if err == nil {
		t.Fatal("want error")
	}
}
