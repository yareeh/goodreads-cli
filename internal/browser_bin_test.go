package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveBrowserBin_EnvOverride(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "chrome")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOODREADS_BROWSER_BIN", bin)
	if got := resolveBrowserBin(); got != bin {
		t.Fatalf("resolveBrowserBin() = %q, want %q", got, bin)
	}
}

func TestResolveBrowserBin_MissingOverrideIsIgnored(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	t.Setenv("GOODREADS_BROWSER_BIN", missing)
	if got := resolveBrowserBin(); got == missing {
		t.Fatalf("resolveBrowserBin() returned the missing override %q", got)
	}
}

func TestResolveBrowserBin_NeverReturnsMissingPath(t *testing.T) {
	t.Setenv("GOODREADS_BROWSER_BIN", "")
	got := resolveBrowserBin()
	if got == "" {
		return // nothing cached: rod downloads one
	}
	if _, err := os.Stat(got); err != nil {
		t.Fatalf("resolveBrowserBin() = %q which does not exist: %v", got, err)
	}
}
