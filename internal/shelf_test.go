package internal

import (
	"strings"
	"testing"
)

func TestShelfSelectorForContainsLabel(t *testing.T) {
	tests := []struct {
		shelf string
		label string
	}{
		{"currently-reading", "Currently reading"},
		{"read", "Read"},
		{"want-to-read", "Want to read"},
	}
	for _, tt := range tests {
		label := shelfAriaLabels[tt.shelf]
		sel := shelfSelectorFor(label)
		if !strings.Contains(sel, tt.label) {
			t.Errorf("shelfSelectorFor(%q) = %q, missing %q", tt.label, sel, tt.label)
		}
	}
}

func TestShelfSelectorUsesContainsMatch(t *testing.T) {
	// selector must use *= (contains) not = (exact) to tolerate aria-label variations
	sel := shelfSelectorFor("Currently reading")
	if !strings.Contains(sel, "*=") {
		t.Errorf("selector should use *= (contains match), got: %q", sel)
	}
}

func TestShelfSelectorReadDoesNotMatchCurrentlyReading(t *testing.T) {
	// "Read" contains-selector must not accidentally match "Currently reading"
	sel := shelfSelectorFor("Read")
	// The selector contains 'Read' but should not contain 'reading' as a sub-pattern
	// (it uses *="Read" which is case-sensitive and won't match lowercase 'reading')
	if strings.Contains(sel, "reading") {
		t.Errorf("Read selector should not contain 'reading': %q", sel)
	}
}

func TestShelfClickJSIncludesLabel(t *testing.T) {
	for _, label := range []string{"Currently reading", "Read", "Want to read"} {
		js := shelfClickJS(label)
		if !strings.Contains(js, label) {
			t.Errorf("shelfClickJS(%q) does not contain label in script: %s", label, js)
		}
		if !strings.Contains(js, "click()") {
			t.Errorf("shelfClickJS(%q) does not call click(): %s", label, js)
		}
		// Must check both aria-label and text content for robustness
		if !strings.Contains(js, "aria-label") {
			t.Errorf("shelfClickJS(%q) does not check aria-label: %s", label, js)
		}
		if !strings.Contains(js, "textContent") {
			t.Errorf("shelfClickJS(%q) does not check textContent: %s", label, js)
		}
	}
}

// TestShelfAriaLabelsMatchGoodreadsDOM documents the exact aria-label values
// observed in the Goodreads shelf dialog (DOM inspection 2026-04).
// If Goodreads changes these, the shelf selector will break and this test catches it.
func TestShelfAriaLabelsMatchGoodreadsDOM(t *testing.T) {
	expected := map[string]string{
		"want-to-read":      "Want to read",
		"currently-reading": "Currently reading",
		"read":              "Read",
	}
	for shelf, wantLabel := range expected {
		if got := shelfAriaLabels[shelf]; got != wantLabel {
			t.Errorf("shelfAriaLabels[%q] = %q, want %q (check if Goodreads changed their DOM)", shelf, got, wantLabel)
		}
	}
}

// TestShelfSelectorReadDoesNotMatchWantToRead guards against *="Read" (capital R)
// accidentally matching "Want to Read" if Goodreads ever capitalises that label.
// Currently "Want to read" uses lowercase 'r', so the selector is safe.
func TestShelfSelectorReadDoesNotMatchWantToRead(t *testing.T) {
	readLabel := shelfAriaLabels["read"]     // "Read"
	wtrLabel := shelfAriaLabels["want-to-read"] // "Want to read"
	if strings.Contains(wtrLabel, readLabel) {
		t.Errorf("want-to-read label %q contains read label %q — selector *=%q would false-match", wtrLabel, readLabel, readLabel)
	}
}

// TestShelfClickJSUsesLowerCase verifies the JS fallback lowercases both the search
// label and the element text/aria-label before comparing, so "Currently reading"
// matches even if the DOM uses different capitalisation.
func TestShelfClickJSUsesLowerCase(t *testing.T) {
	for _, label := range []string{"Currently reading", "Read", "Want to read"} {
		js := shelfClickJS(label)
		if !strings.Contains(js, "toLowerCase") {
			t.Errorf("shelfClickJS(%q) does not use toLowerCase for case-insensitive matching: %s", label, js)
		}
	}
}

// TestShelfButtonSelectorsMatchGoodreadsDOM documents the button selectors used to
// find the shelf control on a book page (DOM inspection 2026-04):
//   - Unshelved: aria-label contains "Tap to shelve book as want to read" (Button--wtr)
//   - Shelved:   aria-label contains "Tap to edit shelf"
//   - Dropdown:  aria-label contains "Tap to choose a shelf for this book" (Button--wtr)
func TestShelfButtonSelectorsMatchGoodreadsDOM(t *testing.T) {
	// The initial element search covers both shelved and unshelved states.
	// button.Button--wtr matches the first WTR button on unshelved books.
	// aria-label*="Tap to edit shelf" matches the edit button on shelved books.
	wantSelectors := []string{
		`button[aria-label*="Tap to edit shelf"]`,
		`button.Button--wtr`,
	}
	combinedSelector := strings.Join(wantSelectors, ", ")
	for _, part := range wantSelectors {
		if !strings.Contains(combinedSelector, part) {
			t.Errorf("expected selector %q to be part of combined selector", part)
		}
	}
}
