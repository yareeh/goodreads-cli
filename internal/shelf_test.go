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
