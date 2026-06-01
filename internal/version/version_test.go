package version

import "testing"

func TestConstantIsValidSemver(t *testing.T) {
	// The release Constant is the source of truth for what `goodreads --version`
	// reports when built without ldflags (the default for `go install @latest`).
	// It must be a valid semver MAJOR.MINOR.PATCH triple.
	if Constant == "" {
		t.Fatal("Constant is empty — set it to the next release version")
	}
	if v := parts(Constant); len(v) != 3 {
		t.Errorf("Constant = %q, want MAJOR.MINOR.PATCH (3 numeric segments), got %d segments", Constant, len(v))
	}
}

func TestCurrentFallsBackToConstant(t *testing.T) {
	// Under `go test` the build info reports "(devel)" or empty, so Current()
	// must yield Constant rather than a placeholder string.
	got := Current()
	if got != Constant {
		t.Errorf("Current() = %q, want %q (Constant) — build info should not be set during go test", got, Constant)
	}
}

func TestGTE(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.5.0", "1.4.5", true},
		{"1.4.5", "1.5.0", false},
		{"1.5.0", "1.5.0", true},
		{"v1.5.0", "1.5.0", true},
		{"2.0.0", "1.99.99", true},
		{"1.5.0-rc1", "1.5.0", true}, // pre-release suffix ignored
		{"", "0.0.0", true},
		{"1.5.0", "", true},
		{"", "1.5.0", false},
	}
	for _, c := range cases {
		if got := GTE(c.a, c.b); got != c.want {
			t.Errorf("GTE(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}
