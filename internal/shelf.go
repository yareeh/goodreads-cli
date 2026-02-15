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

	// Check if the book is already shelved
	editBtn, err := b.Page.Timeout(10 * time.Second).Element(
		`button[aria-label*="Tap to edit shelf"], button.Button--wtr`,
	)
	if err != nil {
		saveDebugScreenshot(b)
		return fmt.Errorf("could not find shelf button on book page: %w", err)
	}

	ariaLabel, _ := editBtn.Attribute("aria-label")
	alreadyShelved := ariaLabel != nil && strings.Contains(*ariaLabel, "Tap to edit shelf")

	if !alreadyShelved {
		// Book is unshelved — click "Want to Read" first to shelve it
		editBtn.MustClick()
		b.Page.MustWaitStable()
		time.Sleep(1 * time.Second)

		if shelfName == "want-to-read" {
			return b.SaveCookies()
		}

		// Now the button should change to the edit-shelf variant, re-find it
		editBtn, err = b.Page.Timeout(10 * time.Second).Element(
			`button[aria-label*="Tap to edit shelf"]`,
		)
		if err != nil {
			saveDebugScreenshot(b)
			return fmt.Errorf("book was shelved but edit button did not appear: %w", err)
		}
	}

	// Click the edit button to open the shelf dialog
	editBtn.MustClick()
	time.Sleep(1 * time.Second)

	// Select the target shelf from the dialog.
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

