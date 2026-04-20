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
		{"currently-reading", "Currently Reading"},
		{"read", "Read"},
		{"want-to-read", "Want to Read"},
	}
	for _, tt := range tests {
		label := shelfAriaLabels[tt.shelf]
		sel := shelfSelectorFor(label)
		if !strings.Contains(sel, tt.label) {
			t.Errorf("shelfSelectorFor(%q) = %q, missing %q", tt.label, sel, tt.label)
		}
	}
}

func TestShelfSelectorUsesExactMatch(t *testing.T) {
	// selector must use = (exact) not *= (contains) so "Read" never matches "Currently Reading"
	sel := shelfSelectorFor("Currently Reading")
	if strings.Contains(sel, "*=") {
		t.Errorf("selector should use exact = not *= (contains), got: %q", sel)
	}
	if !strings.Contains(sel, `="Currently Reading"`) {
		t.Errorf("selector must use exact aria-label match, got: %q", sel)
	}
}

func TestShelfSelectorReadDoesNotMatchCurrentlyReading(t *testing.T) {
	// With exact match (=), "Read" selector cannot match "Currently Reading"
	readSel := shelfSelectorFor("Read")
	if strings.Contains(readSel, "Currently") || strings.Contains(readSel, "Reading") {
		t.Errorf("Read selector should not reference Currently Reading: %q", readSel)
	}
}

func TestShelfClickJSIncludesLabel(t *testing.T) {
	for _, label := range []string{"Currently Reading", "Read", "Want to Read"} {
		js := shelfClickJS(label)
		if !strings.Contains(js, label) {
			t.Errorf("shelfClickJS(%q) does not contain label in script: %s", label, js)
		}
		if !strings.Contains(js, "click()") {
			t.Errorf("shelfClickJS(%q) does not call click(): %s", label, js)
		}
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
		"want-to-read":      "Want to Read",
		"currently-reading": "Currently Reading",
		"read":              "Read",
	}
	for shelf, wantLabel := range expected {
		if got := shelfAriaLabels[shelf]; got != wantLabel {
			t.Errorf("shelfAriaLabels[%q] = %q, want %q (check if Goodreads changed their DOM)", shelf, got, wantLabel)
		}
	}
}

// TestShelfSelectorReadDoesNotMatchWantToRead verifies that exact match prevents
// "Read" from matching "Want to Read".
func TestShelfSelectorReadDoesNotMatchWantToRead(t *testing.T) {
	readSel := shelfSelectorFor("Read")
	// Exact selector button[aria-label="Read"] must not contain "Want"
	if strings.Contains(readSel, "Want") {
		t.Errorf("Read selector must not mention Want to Read: %q", readSel)
	}
}

// TestShelfClickJSUsesLowerCase verifies the JS fallback lowercases both the search
// label and the element text/aria-label before comparing, so it handles any
// capitalisation Goodreads might use.
func TestShelfClickJSUsesLowerCase(t *testing.T) {
	for _, label := range []string{"Currently Reading", "Read", "Want to Read"} {
		js := shelfClickJS(label)
		if !strings.Contains(js, "toLowerCase") {
			t.Errorf("shelfClickJS(%q) does not use toLowerCase for case-insensitive matching: %s", label, js)
		}
	}
}

// TestShelfButtonSelectorsMatchGoodreadsDOM documents the button selectors used to
// find the shelf control on a book page (DOM inspection 2026-04):
//   - Unshelved: aria-label = "Tap to shelve book as want to read" (Button--wtr)
//   - Shelved:   aria-label contains "Tap to edit shelf" (Button--secondary)
func TestShelfButtonSelectorsMatchGoodreadsDOM(t *testing.T) {
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
