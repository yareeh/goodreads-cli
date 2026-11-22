package internal

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// isBuiltinShelf reports whether name is one of the three reading-status
// shelves the book page's shelf dialog offers. Every other name is a user
// shelf, added through addToCustomShelf.
func isBuiltinShelf(name string) bool {
	_, ok := shelfAriaLabels[name]
	return ok
}

// ShelfSet is the user's shelves, as listed by the shelf chooser of a
// /review/list page.
type ShelfSet struct {
	All       []string
	Exclusive []string
}

// IsExclusive reports whether shelf is exclusive (a book can be on only one
// exclusive shelf: to-read, currently-reading, read, and any the user made
// exclusive, such as did-not-finish).
func (s ShelfSet) IsExclusive(shelf string) bool {
	for _, e := range s.Exclusive {
		if e == shelf {
			return true
		}
	}
	return false
}

// Resolve returns the user's spelling of shelf, matched case-insensitively.
// A name that is not one of the user's shelves is an error, never a new
// shelf: Goodreads' add_to_shelf would silently create one for a typo.
func (s ShelfSet) Resolve(shelf string) (string, error) {
	for _, name := range s.All {
		if strings.EqualFold(name, shelf) {
			return name, nil
		}
	}
	return "", fmt.Errorf("you have no shelf named %q (your shelves: %s); create it on Goodreads first",
		shelf, strings.Join(s.All, ", "))
}

var (
	_chooserShelvesRE   = regexp.MustCompile(`new ShelfChooser\(\s*"shelfChooserInput",\s*\d+,\s*(\[[^\]]*\])`)
	_chooserExclusiveRE = regexp.MustCompile(`exclusive:\s*(\[[^\]]*\])`)
)

// ParseShelfChooser reads the user's shelves from the ShelfChooser script of
// a rendered /review/list page:
//
//	new ShelfChooser("shelfChooserInput", 0, ["to-read", ...], {exclusive: [...], ...})
func ParseShelfChooser(html string) (ShelfSet, error) {
	m := _chooserShelvesRE.FindStringSubmatch(html)
	if m == nil {
		return ShelfSet{}, fmt.Errorf("shelf list not found on the shelf page — Goodreads layout changed?")
	}
	var set ShelfSet
	if err := json.Unmarshal([]byte(m[1]), &set.All); err != nil {
		return ShelfSet{}, fmt.Errorf("parsing shelf list: %w", err)
	}
	if e := _chooserExclusiveRE.FindStringSubmatch(html[strings.Index(html, m[0]):]); e != nil {
		if err := json.Unmarshal([]byte(e[1]), &set.Exclusive); err != nil {
			return ShelfSet{}, fmt.Errorf("parsing exclusive shelves: %w", err)
		}
	}
	return set, nil
}

var _bookIDRE = regexp.MustCompile(`^[0-9]+$`)

// validBookID reports whether id is a plain numeric Goodreads book ID, the
// form the add_to_shelf endpoint takes.
func validBookID(id string) bool {
	return _bookIDRE.MatchString(id)
}

// jsString encodes s as a JavaScript string literal. JSON strings are valid
// JS literals, and encoding/json escapes <, >, &, U+2028 and U+2029 too.
func jsString(s string) string {
	b, _ := json.Marshal(s) // a string always marshals
	return string(b)
}

// customShelfAddJS posts to Goodreads' authenticated Rails shelf helper. The
// book page's shelf dialog only offers the three reading-status shelves, so
// user shelves (ebook, audiobook, did-not-finish, ...) cannot be chosen there.
// It must run on a signed-in goodreads.com page that carries a CSRF token
// (the /review/list page does). The token stays in the page; it never
// reaches Go logs or command output.
func customShelfAddJS(bookID, shelfName string) string {
	return fmt.Sprintf(`async () => {
		const bookID = %s;
		const shelfName = %s;
		const token = document.querySelector('input[name="authenticity_token"]')?.value ||
			document.querySelector('meta[name="csrf-token"]')?.content || '';
		if (!token) return {ok: false, status: 0, error: 'CSRF token not found'};
		const body = new URLSearchParams({authenticity_token: token, book_id: bookID, name: shelfName});
		const response = await fetch('/shelf/add_to_shelf', {
			method: 'POST',
			credentials: 'same-origin',
			headers: {
				'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
				'X-CSRF-Token': token,
				'X-Requested-With': 'XMLHttpRequest',
			},
			body: body.toString(),
		});
		return {ok: response.ok, status: response.status};
	}`, jsString(bookID), jsString(shelfName))
}

// addToCustomShelf adds bookID to one of the user's own shelves and checks
// that the book is listed on it afterwards.
func addToCustomShelf(b *Browser, bookID, shelfName string) error {
	if !validBookID(bookID) {
		return fmt.Errorf("invalid book ID %q: custom shelves need the numeric Goodreads ID", bookID)
	}
	homeHTML, err := b.FetchRenderedHTML(BaseURL + "/")
	if err != nil {
		return fmt.Errorf("fetching home page: %w", err)
	}
	userID, err := ExtractUserIDFromHomeHTML(homeHTML)
	if err != nil {
		return err
	}
	// The shelf page carries both the user's shelf list and a CSRF token,
	// and the POST below runs in it.
	listHTML, err := b.FetchRenderedHTML(fmt.Sprintf("%s/review/list/%s?per_page=20", BaseURL, userID))
	if err != nil {
		return fmt.Errorf("fetching your shelves: %w", err)
	}
	shelves, err := ParseShelfChooser(listHTML)
	if err != nil {
		saveDebugArtifacts(b)
		return err
	}
	shelf, err := shelves.Resolve(shelfName)
	if err != nil {
		return err
	}

	result, err := b.Page.Eval(customShelfAddJS(bookID, shelf))
	if err != nil {
		b.Log.Record("custom_shelf_post", map[string]any{"bookID": bookID, "shelf": shelf}, err)
		saveDebugArtifacts(b)
		return fmt.Errorf("adding book to shelf %q: %w", shelf, err)
	}
	ok := result != nil && result.Value.Get("ok").Bool()
	status, errText := 0, ""
	if result != nil {
		status = result.Value.Get("status").Int()
		errText = result.Value.Get("error").Str()
	}
	b.Log.Record("custom_shelf_post", map[string]any{
		"bookID": bookID, "shelf": shelf, "status": status, "exclusive": shelves.IsExclusive(shelf),
	}, nil)
	if !ok {
		saveDebugArtifacts(b)
		if errText != "" {
			return fmt.Errorf("adding book to shelf %q: %s", shelf, errText)
		}
		return fmt.Errorf("adding book to shelf %q: Goodreads returned HTTP %d", shelf, status)
	}

	// A 200 alone does not prove the write landed (issues #217, #218), so
	// look for the book on the shelf.
	found, err := shelfHasBook(bookID, func(page int) (string, error) {
		return b.FetchRenderedHTML(shelfURL(userID, shelf, page))
	})
	b.Log.Record("verify_custom_shelf", map[string]any{"bookID": bookID, "shelf": shelf, "found": found}, err)
	if err != nil {
		return fmt.Errorf("checking shelf %q: %w", shelf, err)
	}
	if !found {
		return fmt.Errorf("book %s is not on shelf %q after adding it", bookID, shelf)
	}
	return nil
}

// shelfHasBook reads shelf pages 1, 2, 3 … through fetch until it sees
// bookID (true), or a page brings no new books (false). Like
// collectShelfPages, but it stops early so a large shelf costs one page in
// the common case.
func shelfHasBook(bookID string, fetch func(page int) (string, error)) (bool, error) {
	seen := map[string]bool{}
	for page := 1; page <= maxShelfPages; page++ {
		html, err := fetch(page)
		if err != nil {
			return false, err
		}
		books, err := ParseShelfHTML(html)
		if err != nil {
			return false, err
		}
		added := 0
		for _, bk := range books {
			if bk.ID == bookID {
				return true, nil
			}
			if !seen[bk.ID] {
				seen[bk.ID] = true
				added++
			}
		}
		if added == 0 {
			return false, nil
		}
	}
	return false, nil
}
