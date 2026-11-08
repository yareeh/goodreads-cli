package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

// confirmRemoveText is the label of the confirmation dialog's button.
const confirmRemoveText = "Remove"

// removeFromShelfSelector is the "Remove from my shelf" button of the shelf
// dialog that a shelved book's main button opens (verified by DOM inspection
// 2026-11).
const removeFromShelfSelector = `button[aria-label="Tap to remove from my shelf"]`

// isShelvedLabel reports whether a shelf button's aria-label says the book is
// on one of the user's shelves ("Shelved as 'Want to Read'. Tap to edit ...").
func isShelvedLabel(label string) bool {
	return strings.HasPrefix(label, "Shelved as ")
}

// RemoveFromShelf takes a book off the user's shelves (the "Remove from my
// shelf" option of the shelf dialog). It returns removed=false, nil when the
// book was not on any shelf, so a repeat is harmless. After the click the page
// is reloaded and must offer to shelve the book again; anything else is an
// error, never a silent success (same rule as AddToShelf, issues #217 / #218).
func RemoveFromShelf(b *Browser, bookID string) (removed bool, err error) {
	url := fmt.Sprintf("https://www.goodreads.com/book/show/%s", bookID)
	b.Log.Record("navigate", map[string]any{"url": url, "bookID": bookID, "action": "remove"}, nil)
	b.Page.MustNavigate(url)
	b.Page.MustWaitStable()

	btn, err := b.Page.Timeout(10 * time.Second).Element(
		`button[aria-label*="Tap to edit shelf"], button[aria-label*="Tap to shelve book"], button.Button--wtr`,
	)
	if err != nil {
		saveDebugArtifacts(b)
		return false, fmt.Errorf("could not find shelf button on book page: %w", err)
	}
	al, _ := btn.Attribute("aria-label")
	state := derefString(al)
	b.Log.Record("shelf_button_state", map[string]any{"ariaLabel": state, "shelved": isShelvedLabel(state)}, nil)
	if !isShelvedLabel(state) {
		return false, nil
	}

	// Open the dialog with the main button of the shelved state; its option
	// list ends with "Remove from my shelf".
	if _, err := clickFirstVisible(b, []string{`button[aria-label*="Tap to edit shelf"]`}, 5*time.Second); err != nil {
		saveDebugArtifacts(b)
		return false, fmt.Errorf("could not open the shelf dialog: %w", err)
	}
	if !pollAndClickOption(b, removeFromShelfSelector, 8*time.Second) {
		saveDebugArtifacts(b)
		return false, fmt.Errorf("the shelf dialog has no 'Remove from my shelf' option")
	}
	b.Log.Record("click_remove_from_shelf", map[string]any{"bookID": bookID}, nil)
	// Goodreads asks "Are you sure you want to remove this book from your shelves?"
	// (it clears the book's rating, review and reading activity); confirm it.
	if !confirmRemove(b, 8*time.Second) {
		saveDebugArtifacts(b)
		return false, fmt.Errorf("the remove confirmation dialog did not appear or could not be confirmed")
	}
	b.Log.Record("confirm_remove", map[string]any{"bookID": bookID}, nil)
	b.Page.MustWaitStable()

	if err := verifyRemoved(b, url); err != nil {
		saveDebugArtifacts(b)
		return false, err
	}
	return true, b.SaveCookies()
}

// confirmRemove clicks the "Remove" button of the confirmation dialog (it has
// no aria-label, so it is found by its exact text inside a dialog).
func confirmRemove(b *Browser, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if elems, err := b.Page.Elements(`[role="dialog"] button`); err == nil {
			for _, el := range elems {
				txt, terr := el.Text()
				if terr != nil || strings.TrimSpace(txt) != confirmRemoveText {
					continue
				}
				if visible, _ := el.Visible(); !visible {
					continue
				}
				if el.Click(proto.InputMouseButtonLeft, 1) == nil {
					return true
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

// verifyRemoved reloads the book page and waits for its main button to offer
// "Tap to shelve book ...", i.e. the book is on no shelf. Polls for up to 10 s.
func verifyRemoved(b *Browser, url string) error {
	b.Page.MustNavigate(url)
	b.Page.MustWaitStable()
	deadline := time.Now().Add(10 * time.Second)
	last := ""
	for {
		if el, err := b.Page.Timeout(2 * time.Second).Element(`button[aria-label*="Tap to edit shelf"], button[aria-label*="Tap to shelve book"]`); err == nil {
			if al, _ := el.Attribute("aria-label"); al != nil {
				last = *al
				b.Log.Record("verify_removed_poll", map[string]any{"ariaLabel": last}, nil)
				if !isShelvedLabel(last) {
					return nil
				}
			}
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("book is still shelved after removing it (button: %q)", last)
}
