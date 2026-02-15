package internal

import (
	"fmt"
	"time"
)

// AddToShelf navigates to a book page and adds it to the specified shelf.
func AddToShelf(b *Browser, bookID string, shelfName string) error {
	url := fmt.Sprintf("https://www.goodreads.com/book/show/%s", bookID)
	b.Page.MustNavigate(url)
	b.Page.MustWaitStable()

	// Click the "Want to Read" button or shelf selector to open the dropdown
	shelfButton, err := b.Page.Timeout(10 * time.Second).Element(`button[aria-label*="Shelve"], .wtrButtonContainer button, [data-testid="shelfButton"]`)
	if err != nil {
		return fmt.Errorf("could not find shelf button on book page: %w", err)
	}

	// Check if we need to open a dropdown for non-default shelves
	if shelfName == "want-to-read" {
		// Just click the main button
		shelfButton.MustClick()
		b.Page.MustWaitStable()
	} else {
		// Click the dropdown caret/arrow to get shelf options
		caret, err := b.Page.Timeout(5 * time.Second).Element(`button[aria-label*="Choose a shelf"], .wtrShelfButton, [data-testid="shelfDropdown"]`)
		if err != nil {
			// Try clicking the main button first, then look for dropdown
			shelfButton.MustClick()
			time.Sleep(500 * time.Millisecond)
			caret, err = b.Page.Timeout(5 * time.Second).Element(`button[aria-label*="Choose a shelf"], .wtrShelfButton, [data-testid="shelfDropdown"]`)
			if err != nil {
				return fmt.Errorf("could not find shelf dropdown: %w", err)
			}
		}
		caret.MustClick()
		time.Sleep(500 * time.Millisecond)

		// Select the desired shelf from the dropdown
		shelfOption, err := b.Page.Timeout(5*time.Second).ElementR(`button, a, [role="option"], [role="menuitem"]`, shelfName)
		if err != nil {
			return fmt.Errorf("could not find shelf option '%s': %w", shelfName, err)
		}
		shelfOption.MustClick()
		b.Page.MustWaitStable()
	}

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
