package internal

import (
	"fmt"
	"strings"
	"time"
)

// shelfAriaLabels maps shelf names to the exact aria-label text in the Goodreads
// shelf dialog (verified by DOM inspection 2026-04).
var shelfAriaLabels = map[string]string{
	"want-to-read":      "Want to Read",
	"currently-reading": "Currently Reading",
	"read":              "Read",
}

// shelfSelectorFor builds a CSS selector for a shelf option button using
// exact matching (=) so "Read" never accidentally matches "Currently Reading".
func shelfSelectorFor(label string) string {
	return fmt.Sprintf(`button[aria-label="%s"]`, label)
}

// shelfClickJS returns a JavaScript snippet that finds a shelf option button
// by aria-label or text content and clicks it, returning true on success.
// This is used as a fallback when the CSS selector fails.
func shelfClickJS(label string) string {
	return fmt.Sprintf(`() => {
		const label = %q;
		const lower = label.toLowerCase();
		// Broad selector: any interactive-looking element inside a dialog/modal or the page
		const candidates = document.querySelectorAll(
			'button, [role="radio"], [role="option"], [role="menuitem"], ' +
			'[role="listbox"] > *, [data-testid], label, div[class*="shelf"], ' +
			'div[class*="Shelf"], span[class*="shelf"], span[class*="Shelf"]'
		);
		for (const el of candidates) {
			const ariaLabel = (el.getAttribute('aria-label') || '').toLowerCase();
			const textContent = el.textContent.trim().toLowerCase();
			const testId = (el.getAttribute('data-testid') || '').toLowerCase();
			if (ariaLabel.includes(lower) || textContent === lower ||
				textContent.includes(lower) || testId.includes(lower.replace(/ /g, ''))) {
				el.click();
				return true;
			}
		}
		return false;
	}`, label)
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
		time.Sleep(2 * time.Second)

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

	// Click the edit button to open the shelf dialog, then wait for it to render
	editBtn.MustClick()
	b.Page.MustWaitStable()
	time.Sleep(2 * time.Second)

	// Select the target shelf from the dialog.
	label, ok := shelfAriaLabels[shelfName]
	if !ok {
		label = shelfName
	}

	option, err := b.Page.Timeout(10 * time.Second).Element(shelfSelectorFor(label))
	if err != nil {
		// CSS selector failed — try JS text-content fallback before giving up
		res, jsErr := b.Page.Eval(shelfClickJS(label))
		if jsErr != nil || res == nil || !res.Value.Bool() {
			saveDebugScreenshot(b)
			return fmt.Errorf("could not find shelf option '%s' in dialog: %w", shelfName, err)
		}
	} else {
		option.MustClick()
	}
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
