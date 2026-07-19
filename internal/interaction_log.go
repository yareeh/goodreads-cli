package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// InteractionLogVersion identifies the on-disk JSON schema. Bump when the
// event shape changes so bug-report consumers can tell old dumps apart.
const InteractionLogVersion = "goodreads-cli/interaction-log/v1"

// InteractionEvent captures one step in a Goodreads session — a navigation,
// a click, an element lookup, a JS evaluation, an HTTP request, or an
// assertion. Every step records what was attempted, whether it succeeded,
// and the error message if it did not.
type InteractionEvent struct {
	Time   time.Time      `json:"time"`
	Kind   string         `json:"kind"`
	Detail map[string]any `json:"detail,omitempty"`
	OK     bool           `json:"ok"`
	Err    string         `json:"error,omitempty"`
}

// InteractionLog accumulates events during a single CLI invocation so the
// error path can dump them as JSON for the user to attach to a bug report.
// The zero value is unusable — construct with NewInteractionLog. A nil
// *InteractionLog is a valid no-op receiver, so instrumentation sites need
// no nil guards.
type InteractionLog struct {
	mu     sync.Mutex
	events []InteractionEvent
	start  time.Time
}

// NewInteractionLog returns an empty log stamped with the current UTC time.
func NewInteractionLog() *InteractionLog {
	return &InteractionLog{start: time.Now().UTC()}
}

// Record appends one event. Safe for concurrent callers. If err is non-nil
// the event is marked !OK and Err is populated. A nil receiver is a no-op
// so callers can instrument freely without checking whether the log exists.
func (l *InteractionLog) Record(kind string, detail map[string]any, err error) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	ev := InteractionEvent{
		Time:   time.Now().UTC(),
		Kind:   kind,
		Detail: detail,
		OK:     err == nil,
	}
	if err != nil {
		ev.Err = err.Error()
	}
	l.events = append(l.events, ev)
}

// Events returns a defensive copy of the events recorded so far. Nil
// receiver returns nil.
func (l *InteractionLog) Events() []InteractionEvent {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]InteractionEvent, len(l.events))
	copy(out, l.events)
	return out
}

// MarshalJSON produces a stable JSON envelope: version, start timestamp,
// event array. Nil receiver marshals to null so it can be embedded freely.
func (l *InteractionLog) MarshalJSON() ([]byte, error) {
	if l == nil {
		return []byte("null"), nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	payload := struct {
		Version string             `json:"version"`
		Start   time.Time          `json:"start"`
		Events  []InteractionEvent `json:"events"`
	}{
		Version: InteractionLogVersion,
		Start:   l.start,
		Events:  l.events,
	}
	return json.MarshalIndent(payload, "", "  ")
}

// Dump writes the log as indented JSON to path. A nil receiver is a no-op
// (returns nil, writes nothing) so error-path code can `l.Dump(path)`
// without a guard even when no browser session was ever established.
func (l *InteractionLog) Dump(path string) error {
	if l == nil {
		return nil
	}
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// DebugLogPath is the default location for the JSON interaction log
// written alongside the debug screenshot and HTML on error.
func DebugLogPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "goodreads-cli-debug.log.json")
}
