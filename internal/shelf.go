package internal

import (
	"fmt"
	"regexp"
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
	b.Log.Record("navigate", map[string]any{"url": url, "bookID": bookID, "shelf": shelfName}, nil)
	b.Page.MustNavigate(url)
	b.Page.MustWaitStable()

	// Check if the book is already shelved
	editBtn, err := b.Page.Timeout(10 * time.Second).Element(
		`button[aria-label*="Tap to edit shelf"], button.Button--wtr`,
	)
	b.Log.Record("find_shelf_button", map[string]any{"selector": `button[aria-label*="Tap to edit shelf"], button.Button--wtr`}, err)
	if err != nil {
		saveDebugArtifacts(b)
		return fmt.Errorf("could not find shelf button on book page: %w", err)
	}

	ariaLabel, _ := editBtn.Attribute("aria-label")
	alreadyShelved := ariaLabel != nil && strings.Contains(*ariaLabel, "Tap to edit shelf")
	b.Log.Record("shelf_button_state", map[string]any{
		"ariaLabel":      derefString(ariaLabel),
		"alreadyShelved": alreadyShelved,
	}, nil)

	// The BookActions ButtonGroup on the current Goodreads book page is a
	// pair: the main WTR/edit button plus a chevron dropdown next to it
	// (aria-label="Tap to choose a shelf for this book" when unshelved,
	// "Edit shelf choice"-style when shelved) that opens the shelf-picker
	// dialog directly. Going through the dropdown is a one-click path to
	// the dialog and works identically whether the book is already shelved
	// or not — sidestepping the old two-step "click WTR → wait for the SPA
	// to rerender the button into the edit variant → click again" flow,
	// which flaked when the rerender never fired and left the click either
	// silently ignored (issue #230: nothing shelved, edit button never
	// appeared) or landing on a stale button.
	//
	// Both the desktop and mobile layouts render this ButtonGroup, so the
	// selector matches multiple elements — one is hidden by CSS at any
	// given viewport and rod's rich click loses on it with "context
	// deadline exceeded". Use a JS-scoped click via querySelector +
	// .click(), which fires the React handler regardless of visibility
	// and picks whichever instance the DOM lists first.
	dialogClicked, jsErr := b.Page.Eval(`() => {
		const selectors = [
			'button[aria-label="Tap to choose a shelf for this book"]',
			'button[aria-label*="edit shelf choice" i]',
		];
		for (const sel of selectors) {
			const el = document.querySelector(sel);
			if (el) { el.click(); return sel; }
		}
		return '';
	}`)
	dialogClickedSel := ""
	if jsErr == nil && dialogClicked != nil {
		dialogClickedSel = dialogClicked.Value.Str()
	}
	b.Log.Record("click_dialog_opener_js", map[string]any{
		"matchedSelector": dialogClickedSel,
		"alreadyShelved":  alreadyShelved,
	}, jsErr)

	if jsErr != nil || dialogClickedSel == "" {
		// Fall back to the main button — on already-shelved books that
		// button itself opens the dialog. On unshelved books it only
		// shelves as WTR, so this path won't move the book to a non-WTR
		// shelf, but it keeps the "want-to-read" fast-path working when
		// Goodreads reshuffles their DOM.
		b.Log.Record("dialog_opener_fallback", map[string]any{"reason": "chevron not clickable, using main button"}, nil)
		editBtn.MustClick()
		b.Log.Record("click_shelf_button", map[string]any{"variant": "open_shelf_dialog_fallback"}, nil)
	}
	b.Page.MustWaitStable()
	time.Sleep(2 * time.Second)

	// Select the target shelf from the dialog.
	label, ok := shelfAriaLabels[shelfName]
	if !ok {
		label = shelfName
	}

	option, err := b.Page.Timeout(10 * time.Second).Element(shelfSelectorFor(label))
	b.Log.Record("find_shelf_option", map[string]any{"selector": shelfSelectorFor(label), "label": label}, err)
	if err != nil {
		// CSS selector failed — try JS text-content fallback before giving up
		res, jsErr := b.Page.Eval(shelfClickJS(label))
		jsFound := jsErr == nil && res != nil && res.Value.Bool()
		b.Log.Record("shelf_option_js_fallback", map[string]any{"label": label, "found": jsFound}, jsErr)
		if !jsFound {
			saveDebugArtifacts(b)
			return fmt.Errorf("could not find shelf option '%s' in dialog: %w", shelfName, err)
		}
	} else {
		option.MustClick()
		b.Log.Record("click_shelf_option", map[string]any{"label": label}, nil)
	}
	b.Page.MustWaitStable()

	// Post-action verification: the page button's aria-label flips to
	// "Shelved as '<shelf>'. Tap to edit shelf for this book" once Goodreads
	// commits the change server-side. Without this check the function used
	// to report success even when the click happened on a slow / WAF-walled
	// page and never reached the backend (issues #217, #218).
	if err := verifyShelf(b, label); err != nil {
		saveDebugArtifacts(b)
		return err
	}

	return b.SaveCookies()
}

// derefString returns the value of s, or "" if s is nil. Used only to keep
// interaction-log details JSON-serialisable (a nil *string marshals to
// "null", which is fine but harder to eyeball in a bug report).
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// verifyShelf polls the page button's aria-label until it confirms the book
// is on `wantLabel`. Polls every 1s for up to 8s. Returns an error if the
// shelf state never matches — the caller should treat this as a failed
// shelf operation rather than a silent success.
func verifyShelf(b *Browser, wantLabel string) error {
	deadline := time.Now().Add(8 * time.Second)
	var last string
	for {
		el, err := b.Page.Timeout(2 * time.Second).Element(`button[aria-label*="Tap to edit shelf"]`)
		if err == nil {
			al, _ := el.Attribute("aria-label")
			if al != nil {
				last = parseShelvedAriaLabel(*al)
				b.Log.Record("verify_shelf_poll", map[string]any{
					"ariaLabel": *al, "parsed": last, "want": wantLabel,
				}, nil)
				if shelfLabelsMatch(last, wantLabel) {
					return nil
				}
			}
		} else {
			b.Log.Record("verify_shelf_poll", map[string]any{"want": wantLabel}, err)
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if last == "" {
		return fmt.Errorf("shelf operation could not be verified — button aria-label never read 'Shelved as ...'")
	}
	return fmt.Errorf("shelf operation appeared to land on %q, not %q", last, wantLabel)
}

// shelfLabelsMatch compares two shelf display names in a way that survives
// Goodreads shifting the aria-label's capitalisation or padding it with
// whitespace. The exact string match used previously flagged
// `Currently Reading` vs `currently reading` as a mismatch and made
// verifyShelf falsely report failure even when the shelf had actually
// changed (bug: "failed to add book to currently reading shelf").
func shelfLabelsMatch(got, want string) bool {
	return strings.EqualFold(strings.TrimSpace(got), strings.TrimSpace(want))
}

// _shelvedAriaLabelRE matches the post-action button's aria-label —
// "Shelved as 'Want to Read'. Tap to edit shelf for this book" — and
// captures the current shelf's display name.
var _shelvedAriaLabelRE = regexp.MustCompile(`Shelved as '([^']+)'`)

// parseShelvedAriaLabel returns the current shelf's display name from a
// shelf-button aria-label, or "" if the label doesn't indicate a shelved
// state. Issues #217 / #218.
func parseShelvedAriaLabel(label string) string {
	m := _shelvedAriaLabelRE.FindStringSubmatch(label)
	if m == nil {
		return ""
	}
	return m[1]
}

// MarkCurrentlyReading adds a book to the "currently-reading" shelf.
func MarkCurrentlyReading(b *Browser, bookID string) error {
	return AddToShelf(b, bookID, "currently-reading")
}

// MarkRead adds a book to the "read" shelf.
func MarkRead(b *Browser, bookID string) error {
	return AddToShelf(b, bookID, "read")
}
