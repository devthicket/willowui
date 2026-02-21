package reactive

import (
	"sort"
	"testing"
)

// ---------------------------------------------------------------------------
// Record.Keys
// ---------------------------------------------------------------------------

func TestRecordKeys(t *testing.T) {
	resetScheduler()
	r := NewRecordFrom(map[string]any{"b": 2, "a": 1, "c": 3})
	keys := r.Keys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	sort.Strings(keys)
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Fatalf("unexpected keys: %v", keys)
	}
}

func TestRecordKeysEmpty(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	keys := r.Keys()
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys, got %d", len(keys))
	}
}

// ---------------------------------------------------------------------------
// Record.Ref — creates field if not exists
// ---------------------------------------------------------------------------

func TestRecordRefCreatesField(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	ref := r.Ref("newkey")
	if ref == nil {
		t.Fatal("Ref should not return nil")
	}
	if ref.Get() != nil {
		t.Fatalf("expected nil initial value, got %v", ref.Get())
	}
	// The field should now exist.
	if !r.Has("newkey") {
		t.Fatal("Ref should create the field")
	}
}

func TestRecordRefReturnsExisting(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	r.Set("score", 42)
	ref := r.Ref("score")
	if ref.Get() != 42 {
		t.Fatalf("expected 42, got %v", ref.Get())
	}
}

func TestRecordRefReactive(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	ref := r.Ref("hp")

	fires := 0
	WatchEffect(func() {
		_ = ref.Get()
		fires++
	})
	if fires != 1 {
		t.Fatalf("expected 1 initial fire, got %d", fires)
	}

	r.Set("hp", 100)
	DefaultScheduler.Flush()
	if fires != 2 {
		t.Fatalf("expected 2 fires after set, got %d", fires)
	}
}

// ---------------------------------------------------------------------------
// Record.Set — no-op for same value does not fire onChange
// ---------------------------------------------------------------------------

func TestRecordSetNoOpSameValue(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	r.Set("x", 10)

	calls := 0
	r.OnChange(func(string, any) { calls++ })

	r.Set("x", 10) // same value
	if calls != 0 {
		t.Fatalf("expected 0 OnChange for same value, got %d", calls)
	}
}

// ---------------------------------------------------------------------------
// Record.Delete — no-op for missing key
// ---------------------------------------------------------------------------

func TestRecordDeleteMissingKey(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	calls := 0
	r.OnChange(func(string, any) { calls++ })
	r.Delete("nonexistent") // should be a no-op
	if calls != 0 {
		t.Fatalf("expected 0 OnChange for deleting missing key, got %d", calls)
	}
}
