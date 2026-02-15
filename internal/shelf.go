package internal

import (
	"fmt"
	"strings"
	"time"
)

// shelfAriaLabels maps shelf names to the aria-label text in the shelf dialog.
var shelfAriaLabels = map[string]string{
	"want-to-read":      "Want to read",
	"currently-reading":  "Currently reading",
	"read":              "Read",
}

// AddToShelf navigates to a book page and adds it to the specified shelf.
func AddToShelf(b *Browser, bookID string, shelfName string) error {
	url := fmt.Sprintf("https://www.goodreads.com/book/show/%s", bookID)
	b.Page.MustNavigate(url)
	b.Page.MustWaitStable()

	// The shelf button differs depending on whether the book is already shelved.
	// Unshelved: button.Button--wtr with "Tap to shelve book as want to read"
	// Shelved:   button.Button--secondary with "Shelved as '...'. Tap to edit shelf"
	// Try the shelved version first, then the unshelved "Want to Read" button.
	shelfBtn, err := b.Page.Timeout(10 * time.Second).Element(
		`button[aria-label*="Tap to edit shelf"], button.Button--wtr`,
	)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find shelf button on book page: %w", err)
	}
	shelfBtn.MustClick()
	time.Sleep(1 * time.Second)

	// If we clicked "Want to Read" on an unshelved book and the target is want-to-read, we're done.
	ariaLabel, _ := shelfBtn.Attribute("aria-label")
	if shelfName == "want-to-read" && ariaLabel != nil && !strings.Contains(*ariaLabel, "Tap to edit shelf") {
		b.Page.MustWaitStable()
		return b.SaveCookies()
	}

	// A dialog should now be open with shelf options. Find the target shelf by aria-label.
	label, ok := shelfAriaLabels[shelfName]
	if !ok {
		label = shelfName
	}

	selector := fmt.Sprintf(`button[aria-label="%s"], button[aria-label="%s, selected"]`, label, label)
	option, err := b.Page.Timeout(5 * time.Second).Element(selector)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find shelf option '%s' in dialog: %w", shelfName, err)
	}
	option.MustClick()
	b.Page.MustWaitStable()

	return b.SaveCookies()
}

// MarkCurrentlyReading adds a book to the "currently-reading" shelf.
func MarkCurrentlyReading(b *Browser, bookID string) error {
	return AddToShelf(b, bookID, "currently-reading")
}

// MarkRead adds a book to the "read" shelf.
func MarkRead(b *Browser, bookID string) error {
	return AddToShelf(b, bookID, "read")
}

