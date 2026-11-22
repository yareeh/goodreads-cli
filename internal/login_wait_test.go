package internal

import (
	"testing"
	"time"
)

func TestWaitUntilSucceedsWhenCheckTurnsTrue(t *testing.T) {
	calls := 0
	ok := waitUntil(func() bool { calls++; return calls >= 3 }, time.Second, time.Millisecond)
	if !ok {
		t.Fatal("waitUntil returned false, want true once check succeeds")
	}
	if calls != 3 {
		t.Errorf("check called %d times, want 3", calls)
	}
}

func TestWaitUntilTimesOut(t *testing.T) {
	start := time.Now()
	ok := waitUntil(func() bool { return false }, 20*time.Millisecond, time.Millisecond)
	if ok {
		t.Fatal("waitUntil returned true, want false on timeout")
	}
	if elapsed := time.Since(start); elapsed < 20*time.Millisecond {
		t.Errorf("returned after %v, before the 20ms timeout", elapsed)
	}
}
