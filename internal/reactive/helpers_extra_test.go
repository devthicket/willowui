package reactive

import (
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// Increment
// ---------------------------------------------------------------------------

func TestIncrement(t *testing.T) {
	resetScheduler()
	r := NewRef(10)
	inc := Increment(r, 5)
	inc()
	if r.Get() != 15 {
		t.Fatalf("expected 15, got %d", r.Get())
	}
	inc()
	if r.Get() != 20 {
		t.Fatalf("expected 20, got %d", r.Get())
	}
}

func TestIncrementNegative(t *testing.T) {
	resetScheduler()
	r := NewRef(10)
	dec := Increment(r, -3)
	dec()
	if r.Get() != 7 {
		t.Fatalf("expected 7, got %d", r.Get())
	}
}

func TestIncrementFloat(t *testing.T) {
	resetScheduler()
	r := NewRef(1.5)
	inc := Increment(r, 0.5)
	inc()
	if r.Get() != 2.0 {
		t.Fatalf("expected 2.0, got %f", r.Get())
	}
}

// ---------------------------------------------------------------------------
// Toggle
// ---------------------------------------------------------------------------

func TestToggle(t *testing.T) {
	resetScheduler()
	r := NewRef(false)
	tog := Toggle(r)
	tog()
	if !r.Get() {
		t.Fatal("expected true after first toggle")
	}
	tog()
	if r.Get() {
		t.Fatal("expected false after second toggle")
	}
}

// ---------------------------------------------------------------------------
// Set
// ---------------------------------------------------------------------------

func TestSet(t *testing.T) {
	resetScheduler()
	r := NewRef("hello")
	setter := Set(r, "world")
	setter()
	if r.Get() != "world" {
		t.Fatalf("expected world, got %s", r.Get())
	}
}

func TestSetSameValue(t *testing.T) {
	resetScheduler()
	r := NewRef(42)
	calls := 0
	WatchEffect(func() {
		_ = r.Get()
		calls++
	})

	setter := Set(r, 42)
	setter() // no-op set
	DefaultScheduler.Flush()
	if calls != 1 {
		t.Fatalf("expected 1 call (no-op), got %d", calls)
	}
}

// ---------------------------------------------------------------------------
// BindFormatter
// ---------------------------------------------------------------------------

func TestBindFormatter(t *testing.T) {
	resetScheduler()
	src := NewRef(42)
	bound, h := BindFormatter(src)
	defer h.Stop()
	if bound.Get() != "42" {
		t.Fatalf("expected '42', got %q", bound.Get())
	}

	src.Set(100)
	DefaultScheduler.Flush()
	if bound.Get() != "100" {
		t.Fatalf("expected '100', got %q", bound.Get())
	}
}

func TestBindFormatterString(t *testing.T) {
	resetScheduler()
	src := NewRef("hello")
	bound, h := BindFormatter(src)
	defer h.Stop()
	if bound.Get() != "hello" {
		t.Fatalf("expected 'hello', got %q", bound.Get())
	}

	src.Set("world")
	DefaultScheduler.Flush()
	if bound.Get() != "world" {
		t.Fatalf("expected 'world', got %q", bound.Get())
	}
}

// ---------------------------------------------------------------------------
// BindFormatterf
// ---------------------------------------------------------------------------

func TestBindFormatterf(t *testing.T) {
	resetScheduler()
	src := NewRef(3)
	bound, h := BindFormatterf(src, "Count: %d")
	defer h.Stop()
	if bound.Get() != "Count: 3" {
		t.Fatalf("expected 'Count: 3', got %q", bound.Get())
	}

	src.Set(10)
	DefaultScheduler.Flush()
	if bound.Get() != "Count: 10" {
		t.Fatalf("expected 'Count: 10', got %q", bound.Get())
	}
}

func TestBindFormatterfFloat(t *testing.T) {
	resetScheduler()
	src := NewRef(3.14)
	bound, h := BindFormatterf(src, "Pi is %.2f")
	defer h.Stop()
	expected := fmt.Sprintf("Pi is %.2f", 3.14)
	if bound.Get() != expected {
		t.Fatalf("expected %q, got %q", expected, bound.Get())
	}
}

// ---------------------------------------------------------------------------
// Ref.Update
// ---------------------------------------------------------------------------

func TestRefUpdate(t *testing.T) {
	resetScheduler()
	r := NewRef(10)
	r.Update(func(v int) int { return v * 3 })
	if r.Get() != 30 {
		t.Fatalf("expected 30, got %d", r.Get())
	}
}

func TestRefUpdateNoOp(t *testing.T) {
	resetScheduler()
	r := NewRef(5)
	calls := 0
	WatchEffect(func() {
		_ = r.Get()
		calls++
	})

	// Update that returns same value should be a no-op.
	r.Update(func(v int) int { return v })
	DefaultScheduler.Flush()
	if calls != 1 {
		t.Fatalf("expected 1 call (no-op update), got %d", calls)
	}
}

func TestRefUpdateReactive(t *testing.T) {
	resetScheduler()
	r := NewRef(1)

	var observed []int
	WatchEffect(func() {
		observed = append(observed, r.Get())
	})

	r.Update(func(v int) int { return v + 10 })
	DefaultScheduler.Flush()

	if len(observed) != 2 || observed[1] != 11 {
		t.Fatalf("expected [1, 11], got %v", observed)
	}
}

// ---------------------------------------------------------------------------
// Computed.Peek
// ---------------------------------------------------------------------------

func TestComputedPeek(t *testing.T) {
	resetScheduler()
	r := NewRef(5)
	c := NewComputed(func() int { return r.Get() * 2 })

	if c.Peek() != 10 {
		t.Fatalf("expected 10, got %d", c.Peek())
	}

	// Change the ref and mark computed dirty.
	r.Set(100)
	// Peek should return the cached (stale) value without re-evaluating.
	if c.Peek() != 10 {
		t.Fatalf("expected 10 (cached), got %d", c.Peek())
	}

	// Get() should trigger re-eval.
	DefaultScheduler.Flush()
	if c.Get() != 200 {
		t.Fatalf("expected 200, got %d", c.Get())
	}
}

func TestComputedPeekDoesNotTrack(t *testing.T) {
	resetScheduler()
	src := NewRef(1)
	c := NewComputed(func() int { return src.Get() })

	calls := 0
	WatchEffect(func() {
		// Peek should not establish a dependency.
		_ = c.Peek()
		calls++
	})
	if calls != 1 {
		t.Fatalf("expected 1 initial, got %d", calls)
	}

	src.Set(2)
	DefaultScheduler.Flush()
	// The watch should NOT re-run because it used Peek, not Get.
	if calls != 1 {
		t.Fatalf("expected still 1 (Peek doesn't track), got %d", calls)
	}
}
