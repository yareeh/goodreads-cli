package internal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestInteractionLog_RecordCapturesKindAndDetail verifies the log stores the
// kind, detail payload and success flag for a single event so a downstream
// consumer (bug report attachment) can reconstruct what the CLI attempted.
func TestInteractionLog_RecordCapturesKindAndDetail(t *testing.T) {
	l := NewInteractionLog()
	l.Record("navigate", map[string]any{"url": "https://www.goodreads.com/book/show/1"}, nil)

	events := l.Events()
	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d", len(events))
	}
	got := events[0]
	if got.Kind != "navigate" {
		t.Errorf("Kind = %q, want %q", got.Kind, "navigate")
	}
	if got.Detail["url"] != "https://www.goodreads.com/book/show/1" {
		t.Errorf("Detail[url] = %v", got.Detail["url"])
	}
	if !got.OK {
		t.Errorf("OK = false, want true (no error passed)")
	}
	if got.Err != "" {
		t.Errorf("Err = %q, want empty", got.Err)
	}
	if got.Time.IsZero() {
		t.Errorf("Time zero, want set")
	}
}

// TestInteractionLog_RecordWithErrorMarksNotOK ensures a step recorded with
// an error is marked !OK and carries the message so bug reports can surface
// the failure without needing separate stderr capture.
func TestInteractionLog_RecordWithErrorMarksNotOK(t *testing.T) {
	l := NewInteractionLog()
	l.Record("element", map[string]any{"selector": "#nope"}, errors.New("not found"))

	events := l.Events()
	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d", len(events))
	}
	if events[0].OK {
		t.Errorf("OK = true, want false (error was passed)")
	}
	if events[0].Err != "not found" {
		t.Errorf("Err = %q, want %q", events[0].Err, "not found")
	}
}

// TestInteractionLog_NilReceiverIsNoOp lets callers instrument freely without
// worrying about whether the log exists — a nil *InteractionLog is a valid
// zero-effort recorder (matches the pattern used in stdlib log).
func TestInteractionLog_NilReceiverIsNoOp(t *testing.T) {
	var l *InteractionLog
	// Must not panic
	l.Record("navigate", map[string]any{"url": "x"}, nil)
	if events := l.Events(); events != nil {
		t.Errorf("nil log Events() = %v, want nil", events)
	}
}

// TestInteractionLog_ConcurrentRecordSafe checks that N goroutines can
// record without racing — shelf verification loops and browser callbacks
// can fire from different rod goroutines, and losing events (or panicking)
// would defeat the point of the log.
func TestInteractionLog_ConcurrentRecordSafe(t *testing.T) {
	l := NewInteractionLog()
	var wg sync.WaitGroup
	const n = 50
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			l.Record("navigate", map[string]any{"i": i}, nil)
		}(i)
	}
	wg.Wait()
	if got := len(l.Events()); got != n {
		t.Errorf("Events() count = %d, want %d — likely a race", got, n)
	}
}

// TestInteractionLog_MarshalJSON_HasVersionAndEvents documents the JSON
// shape that ships in bug reports so future consumers can rely on it.
func TestInteractionLog_MarshalJSON_HasVersionAndEvents(t *testing.T) {
	l := NewInteractionLog()
	l.Record("click", map[string]any{"selector": "button.wtr"}, nil)

	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out["version"] == nil || out["version"] == "" {
		t.Errorf("missing version field: %s", string(data))
	}
	events, ok := out["events"].([]any)
	if !ok {
		t.Fatalf("events not an array: %T", out["events"])
	}
	if len(events) != 1 {
		t.Errorf("events len = %d, want 1", len(events))
	}
}

// TestInteractionLog_DumpWritesFile confirms the log-to-file path the CLI
// prints on error is real, machine-readable JSON.
func TestInteractionLog_DumpWritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log.json")

	l := NewInteractionLog()
	l.Record("navigate", map[string]any{"url": "https://www.goodreads.com"}, nil)
	if err := l.Dump(path); err != nil {
		t.Fatalf("Dump: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("dumped JSON invalid: %v — content: %s", err, string(data))
	}
	if events, _ := out["events"].([]any); len(events) != 1 {
		t.Errorf("dumped events len = %d, want 1", len(events))
	}
}

// TestInteractionLog_DumpNilLogWritesNothing keeps the CLI's error-path code
// simple: it can always call log.Dump(path) without a nil guard.
func TestInteractionLog_DumpNilLogWritesNothing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log.json")
	var l *InteractionLog
	if err := l.Dump(path); err != nil {
		t.Errorf("nil.Dump: %v, want nil", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("nil.Dump created file %s (err=%v), want no file", path, err)
	}
}
