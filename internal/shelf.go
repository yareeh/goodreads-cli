package internal

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/proto"
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
	// dialog. Going through the chevron is a one-click path to the dialog
	// and works identically whether the book is already shelved or not —
	// sidestepping the old two-step "click WTR → wait for SPA rerender →
	// click edit again" flow that flaked in issue #230.
	//
	// Both the desktop and mobile layouts render this ButtonGroup, so the
	// selector matches multiple elements — one is hidden by CSS at any
	// given viewport. The v1.8.1 attempt used a JS `.click()` to bypass
	// visibility checks; that turned out to fire without React actually
	// picking it up on this control, so no dialog opened (see #230's
	// re-open: chevron event landed but page state never changed). Use
	// clickFirstVisible: rod native events (dispatched from real mouse
	// coordinates) on the visible desktop-or-mobile instance, so React
	// sees the interaction and opens the shelf picker.
	chevronSelectors := []string{
		`button[aria-label="Tap to choose a shelf for this book"]`,
		`button[aria-label*="edit shelf choice" i]`,
	}
	dialogOpened, chevErr := clickFirstVisible(b, chevronSelectors, 10*time.Second)
	b.Log.Record("click_dialog_opener", map[string]any{
		"clicked":        dialogOpened,
		"alreadyShelved": alreadyShelved,
	}, chevErr)
	if chevErr != nil {
		// Fall back to the main button. On already-shelved books that
		// button itself opens the dialog; on unshelved books it only
		// shelves as WTR (which is still a useful degradation for the
		// want-to-read fast-path). Same trick as the chevron — click
		// the visible instance to fire React events reliably.
		b.Log.Record("dialog_opener_fallback", map[string]any{"reason": "no visible chevron, using main button"}, nil)
		if _, mainErr := clickFirstVisible(
			b,
			[]string{`button[aria-label*="Tap to edit shelf"]`, `button.Button--wtr`},
			5*time.Second,
		); mainErr != nil {
			saveDebugArtifacts(b)
			return fmt.Errorf("could not click any shelf-opener button: %w", mainErr)
		}
	}
	b.Page.MustWaitStable()
	time.Sleep(2 * time.Second)

	// Select the target shelf from the dialog.
	label, ok := shelfAriaLabels[shelfName]
	if !ok {
		label = shelfName
	}

	// Same visibility-safe click for the shelf option inside the dialog.
	optionSelectors := []string{shelfSelectorFor(label)}
	if _, err := clickFirstVisible(b, optionSelectors, 10*time.Second); err != nil {
		// CSS selector didn't hit a visible option — try the broader
		// JS text-content matcher as a last resort. Note: this
		// intentionally does NOT match the main page WTR button
		// (which trapped us in issue #230 re-open), because the JS
		// selector scopes to elements inside dialogs/menus.
		res, jsErr := b.Page.Eval(dialogShelfClickJS(label))
		jsFound := jsErr == nil && res != nil && res.Value.Bool()
		b.Log.Record("shelf_option_js_fallback", map[string]any{"label": label, "found": jsFound}, jsErr)
		if !jsFound {
			saveDebugArtifacts(b)
			return fmt.Errorf("could not find shelf option '%s' in dialog: %w", shelfName, err)
		}
	} else {
		b.Log.Record("click_shelf_option", map[string]any{"label": label}, nil)
	}
	b.Page.MustWaitStable()

	// Post-action verification: the page button's aria-label flips to
	// "Shelved as '<shelf>'. Tap to edit shelf for this book" once Goodreads
	// commits the change server-side. Without this check the function used
	// to report success even when the click happened on a slow / WAF-walled
	// page and never reached the backend (issues #217, #218). If in-place
	// polling times out we do one page reload and re-poll: sometimes the
	// backend committed but the SPA never rerendered the button.
	if err := verifyShelf(b, label); err != nil {
		b.Log.Record("verify_reload", map[string]any{"url": url, "reason": "in-place verify timed out"}, nil)
		b.Page.MustNavigate(url)
		b.Page.MustWaitStable()
		if err2 := verifyShelf(b, label); err2 != nil {
			saveDebugArtifacts(b)
			return err2
		}
	}

	return b.SaveCookies()
}

// clickFirstVisible finds elements matching any of the given selectors,
// waits up to `timeout` for them to render, and clicks the first one that
// is visible + interactable. Returns the selector that actually clicked,
// or an error if no matching element became visible in time.
//
// Motivation: Goodreads renders duplicate desktop + mobile ButtonGroups
// and CSS-hides one per viewport. rod's Element().MustClick() picks
// whichever comes first in the DOM regardless of visibility, so it can
// deadlock trying to click a display:none button. This helper iterates
// with .Visible()/.Interactable() so we click the layout the user's
// browser would actually interact with — and rod's real synthesised
// pointer events then fire React handlers reliably (unlike raw
// element.click() from JS, which some Goodreads controls ignore).
func clickFirstVisible(b *Browser, selectors []string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	joined := strings.Join(selectors, ", ")
	// First wait for any matching element to render at all, so we don't
	// race a still-loading page.
	if _, err := b.Page.Timeout(timeout).Element(joined); err != nil {
		return "", fmt.Errorf("no element matched %q within %s: %w", joined, timeout, err)
	}
	for time.Now().Before(deadline) {
		for _, sel := range selectors {
			elems, err := b.Page.Elements(sel)
			if err != nil {
				continue
			}
			for _, el := range elems {
				visible, verr := el.Visible()
				if verr != nil || !visible {
					continue
				}
				if err := el.Click(proto.InputMouseButtonLeft, 1); err == nil {
					return sel, nil
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return "", fmt.Errorf("no visible element among selectors %q became clickable within %s", joined, timeout)
}

// dialogShelfClickJS scopes the JS text-content click to elements INSIDE
// an open shelf-picker dialog/menu, so we never accidentally click the
// main page's WTR button (which was the mis-match in issue #230's
// re-open). The Goodreads shelf dialog uses [role="menu"] / [role="dialog"]
// wrappers plus data-testids like "shelvesMenu"; scope the query there.
func dialogShelfClickJS(label string) string {
	return fmt.Sprintf(`() => {
		const label = %q;
		const lower = label.toLowerCase();
		const scopes = document.querySelectorAll(
			'[role="dialog"], [role="menu"], [data-testid*="helf"], ' +
			'[class*="ShelfDialog"], [class*="ShelvesDialog"]'
		);
		for (const scope of scopes) {
			const candidates = scope.querySelectorAll(
				'button, [role="menuitem"], [role="option"], label'
			);
			for (const el of candidates) {
				const al = (el.getAttribute('aria-label') || '').toLowerCase();
				const tc = el.textContent.trim().toLowerCase();
				if (al.includes(lower) || tc === lower || tc.includes(lower)) {
					el.click();
					return true;
				}
			}
		}
		return false;
	}`, label)
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
