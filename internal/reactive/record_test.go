package reactive

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

func TestRecordNewRecord(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	if r.Has("x") {
		t.Fatal("expected empty record")
	}
}

func TestRecordNewRecordFrom(t *testing.T) {
	resetScheduler()
	src := map[string]any{"name": "Alice", "level": 10}
	r := NewRecordFrom(src)
	if r.Get("name") != "Alice" {
		t.Fatalf("expected Alice, got %v", r.Get("name"))
	}
	// Verify map copy independence.
	src["name"] = "Bob"
	if r.Get("name") != "Alice" {
		t.Fatal("NewRecordFrom should copy the map")
	}
}

// ---------------------------------------------------------------------------
// Set / Get / Has / Delete
// ---------------------------------------------------------------------------

func TestRecordSetGet(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	r.Set("score", 42)
	if r.Get("score") != 42 {
		t.Fatalf("expected 42, got %v", r.Get("score"))
	}
}

func TestRecordGetMissingKey(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	if r.Get("missing") != nil {
		t.Fatal("expected nil for missing key")
	}
}

func TestRecordHas(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	if r.Has("x") {
		t.Fatal("expected false before set")
	}
	r.Set("x", 1)
	if !r.Has("x") {
		t.Fatal("expected true after set")
	}
}

func TestRecordDelete(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	r.Set("key", "val")
	r.Delete("key")
	if r.Has("key") {
		t.Fatal("expected key removed")
	}
}

func TestRecordDeleteFiresOnChange(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	r.Set("key", "val")

	var gotKey string
	var gotVal any
	r.OnChange(func(k string, v any) {
		gotKey = k
		gotVal = v
	})

	r.Delete("key")
	if gotKey != "key" || gotVal != nil {
		t.Fatalf("Delete should fire OnChange with nil: key=%q val=%v", gotKey, gotVal)
	}
}

// ---------------------------------------------------------------------------
// SetMany
// ---------------------------------------------------------------------------

func TestRecordSetManyFiresOnChangePerField(t *testing.T) {
	resetScheduler()
	r := NewRecord()

	fired := map[string]any{}
	r.OnChange(func(k string, v any) {
		fired[k] = v
	})

	r.SetMany(map[string]any{"a": 1, "b": 2, "c": 3})
	if len(fired) != 3 {
		t.Fatalf("expected 3 OnChange calls, got %d: %v", len(fired), fired)
	}
	if fired["a"] != 1 || fired["b"] != 2 || fired["c"] != 3 {
		t.Fatalf("wrong values: %v", fired)
	}
}

// ---------------------------------------------------------------------------
// OnFieldChange
// ---------------------------------------------------------------------------

func TestRecordOnFieldChange(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	r.Set("score", 0)

	var calls []any
	r.OnFieldChange("score", func(v any) {
		calls = append(calls, v)
	})

	// Must not fire on registration.
	if len(calls) != 0 {
		t.Fatalf("expected 0 calls on registration, got %d", len(calls))
	}

	r.Set("score", 100)
	DefaultScheduler.Flush()

	if len(calls) != 1 || calls[0] != 100 {
		t.Fatalf("expected [100], got %v", calls)
	}
}

func TestRecordOnFieldChangeDoesNotFireForOtherFields(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	r.Set("score", 0)
	r.Set("name", "Alice")

	scoreCalls := 0
	r.OnFieldChange("score", func(any) { scoreCalls++ })

	r.Set("name", "Bob")
	DefaultScheduler.Flush()

	if scoreCalls != 0 {
		t.Fatalf("expected 0 score callbacks on name change, got %d", scoreCalls)
	}
}

// ---------------------------------------------------------------------------
// OnChange
// ---------------------------------------------------------------------------

func TestRecordOnChange(t *testing.T) {
	resetScheduler()
	r := NewRecord()

	var keys []string
	r.OnChange(func(k string, v any) {
		keys = append(keys, k)
	})

	r.Set("x", 1)
	r.Set("y", 2)
	if len(keys) != 2 || keys[0] != "x" || keys[1] != "y" {
		t.Fatalf("expected [x,y], got %v", keys)
	}
}

// ---------------------------------------------------------------------------
// WatchHandle.Stop
// ---------------------------------------------------------------------------

func TestRecordOnChangeStop(t *testing.T) {
	resetScheduler()
	r := NewRecord()

	calls := 0
	h := r.OnChange(func(string, any) { calls++ })

	r.Set("a", 1)
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	h.Stop()
	r.Set("b", 2)
	if calls != 1 {
		t.Fatalf("expected still 1 call after stop, got %d", calls)
	}
}

func TestRecordOnFieldChangeStop(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	r.Set("score", 0)

	calls := 0
	h := r.OnFieldChange("score", func(any) { calls++ })
	h.Stop()

	r.Set("score", 99)
	DefaultScheduler.Flush()

	if calls != 0 {
		t.Fatalf("expected 0 calls after stop, got %d", calls)
	}
}

// ---------------------------------------------------------------------------
// Ref() integration with WatchEffect
// ---------------------------------------------------------------------------

func TestRecordRefWatchEffect(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	r.Set("hp", 100)

	var observed []any
	WatchEffect(func() {
		observed = append(observed, r.Get("hp"))
	})
	if len(observed) != 1 || observed[0] != 100 {
		t.Fatalf("expected [100], got %v", observed)
	}

	r.Set("hp", 80)
	DefaultScheduler.Flush()
	if len(observed) != 2 || observed[1] != 80 {
		t.Fatalf("expected [100,80], got %v", observed)
	}
}

// ---------------------------------------------------------------------------
// ToMap
// ---------------------------------------------------------------------------

func TestRecordToMapCopy(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	r.Set("a", 1)
	r.Set("b", 2)

	m := r.ToMap()
	if m["a"] != 1 || m["b"] != 2 {
		t.Fatalf("ToMap wrong: %v", m)
	}

	// Mutation of returned map does not affect record.
	m["a"] = 99
	if r.Get("a") != 1 {
		t.Fatal("ToMap should return an independent copy")
	}
}
