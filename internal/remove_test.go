package internal

import "testing"

func TestIsShelvedLabel(t *testing.T) {
	cases := []struct {
		label string
		want  bool
	}{
		{"Shelved as 'Want to Read'. Tap to edit shelf for this book", true},
		{"Shelved as 'Currently Reading'. Tap to edit shelf for this book", true},
		{"Tap to shelve book as want to read", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isShelvedLabel(c.label); got != c.want {
			t.Errorf("isShelvedLabel(%q) = %v, want %v", c.label, got, c.want)
		}
	}
}

func TestRemoveButtonSelectorIsExact(t *testing.T) {
	// The remove option is found by its exact aria-label, never by text that a
	// shelf name could share.
	if removeFromShelfSelector != `button[aria-label="Tap to remove from my shelf"]` {
		t.Fatalf("unexpected selector %q", removeFromShelfSelector)
	}
}
