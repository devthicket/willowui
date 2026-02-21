package reactive

import (
	"testing"
)

// --- Ref tests ---

func TestRefGetSet(t *testing.T) {
	resetScheduler()
	r := NewRef(42)
	if got := r.Get(); got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
	r.Set(99)
	if got := r.Get(); got != 99 {
		t.Fatalf("expected 99, got %d", got)
	}
}

func TestRefSetNoOp(t *testing.T) {
	resetScheduler()
	r := NewRef(10)

	calls := 0
	WatchEffect(func() {
		_ = r.Get()
		calls++
	})
	// WatchEffect runs once on setup.
	if calls != 1 {
		t.Fatalf("expected 1 initial call, got %d", calls)
	}

	// Set to same value — should NOT enqueue anything.
	r.Set(10)
	DefaultScheduler.Flush()
	if calls != 1 {
		t.Fatalf("expected no additional call on no-op set, got %d", calls)
	}

	// Set to different value — should fire.
	r.Set(20)
	DefaultScheduler.Flush()
	if calls != 2 {
		t.Fatalf("expected 2 calls after real set, got %d", calls)
	}
}

func TestRefPeek(t *testing.T) {
	resetScheduler()
	r := NewRef(5)

	calls := 0
	c := NewComputed(func() int {
		calls++
		// Use Peek — should NOT create a dependency.
		return r.Peek() * 2
	})

	if got := c.Get(); got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
	if calls != 1 {
		t.Fatalf("expected 1 eval, got %d", calls)
	}

	// Change ref — computed should NOT be dirty because Peek was used.
	r.Set(100)
	DefaultScheduler.Flush()

	if got := c.Get(); got != 10 {
		t.Fatalf("expected 10 (no re-eval), got %d", got)
	}
	if calls != 1 {
		t.Fatalf("expected still 1 eval (Peek doesn't track), got %d", calls)
	}
}

// --- Computed tests ---

func TestComputedLazyEval(t *testing.T) {
	resetScheduler()
	a := NewRef(3)
	evalCount := 0

	c := NewComputed(func() int {
		evalCount++
		return a.Get() * 2
	})

	// Initial eval happens in NewComputed.
	if evalCount != 1 {
		t.Fatalf("expected 1 initial eval, got %d", evalCount)
	}
	if got := c.Get(); got != 6 {
		t.Fatalf("expected 6, got %d", got)
	}
	// Get() should not re-eval if not dirty.
	if evalCount != 1 {
		t.Fatalf("expected still 1 eval, got %d", evalCount)
	}

	// Mark dirty by changing ref.
	a.Set(5)
	// Computed is lazy — should not have re-eval'd yet.
	if evalCount != 1 {
		t.Fatalf("expected still 1 eval before Get, got %d", evalCount)
	}

	DefaultScheduler.Flush()

	if got := c.Get(); got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
	if evalCount != 2 {
		t.Fatalf("expected 2 evals, got %d", evalCount)
	}
}

func TestComputedReEvalOnDepChange(t *testing.T) {
	resetScheduler()
	a := NewRef(1)
	b := NewRef(2)

	c := NewComputed(func() int {
		return a.Get() + b.Get()
	})

	if got := c.Get(); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}

	a.Set(10)
	DefaultScheduler.Flush()

	if got := c.Get(); got != 12 {
		t.Fatalf("expected 12, got %d", got)
	}

	b.Set(20)
	DefaultScheduler.Flush()

	if got := c.Get(); got != 30 {
		t.Fatalf("expected 30, got %d", got)
	}
}

// --- Diamond dependency test ---

func TestDiamondDependency(t *testing.T) {
	resetScheduler()

	// A (ref)
	// / \
	// B   C (computed)
	// \ /
	//  D   (computed)

	a := NewRef(1)

	bEvals := 0
	b := NewComputed(func() int {
		bEvals++
		return a.Get() * 2
	})

	cEvals := 0
	c := NewComputed(func() int {
		cEvals++
		return a.Get() * 3
	})

	dEvals := 0
	d := NewComputed(func() int {
		dEvals++
		return b.Get() + c.Get()
	})

	// Initial evaluation.
	if got := d.Get(); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
	if bEvals != 1 || cEvals != 1 || dEvals != 1 {
		t.Fatalf("expected 1,1,1 evals, got %d,%d,%d", bEvals, cEvals, dEvals)
	}

	// Change A — B, C, and D all get marked dirty.
	a.Set(2)
	DefaultScheduler.Flush()

	// D should only eval ONCE despite two dirty paths.
	if got := d.Get(); got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
	if dEvals != 2 {
		t.Fatalf("expected D to eval exactly 2 times total, got %d", dEvals)
	}
}

// --- Watch tests ---

func TestWatchEffectFires(t *testing.T) {
	resetScheduler()
	r := NewRef(0)

	var observed []int
	WatchEffect(func() {
		observed = append(observed, r.Get())
	})

	// Initial run captures value 0.
	if len(observed) != 1 || observed[0] != 0 {
		t.Fatalf("expected [0], got %v", observed)
	}

	r.Set(1)
	DefaultScheduler.Flush()

	if len(observed) != 2 || observed[1] != 1 {
		t.Fatalf("expected [0,1], got %v", observed)
	}

	r.Set(2)
	r.Set(3) // multiple sets before flush — last value wins
	DefaultScheduler.Flush()

	if len(observed) != 3 || observed[2] != 3 {
		t.Fatalf("expected [0,1,3], got %v", observed)
	}
}

func TestWatchEffectNoFireOnNoOp(t *testing.T) {
	resetScheduler()
	r := NewRef(5)

	calls := 0
	WatchEffect(func() {
		_ = r.Get()
		calls++
	})

	r.Set(5) // no-op
	DefaultScheduler.Flush()

	if calls != 1 {
		t.Fatalf("expected 1 call (initial only), got %d", calls)
	}
}

func TestWatchValue(t *testing.T) {
	resetScheduler()
	r := NewRef(10)

	var changes [][2]int
	WatchValue(r, func(old, new int) {
		changes = append(changes, [2]int{old, new})
	})

	// Initial setup fires immediately with (current, current).
	if len(changes) != 1 || changes[0] != [2]int{10, 10} {
		t.Fatalf("expected [{10,10}] on setup, got %v", changes)
	}

	r.Set(20)
	DefaultScheduler.Flush()

	if len(changes) != 2 || changes[1] != [2]int{10, 20} {
		t.Fatalf("expected [{0,10},{10,20}], got %v", changes)
	}

	r.Set(30)
	DefaultScheduler.Flush()

	if len(changes) != 3 || changes[2] != [2]int{20, 30} {
		t.Fatalf("expected [{0,10},{10,20},{20,30}], got %v", changes)
	}
}

func TestWatchHandleStop(t *testing.T) {
	resetScheduler()
	r := NewRef(0)

	calls := 0
	h := WatchEffect(func() {
		_ = r.Get()
		calls++
	})

	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	h.Stop()

	r.Set(99)
	DefaultScheduler.Flush()

	if calls != 1 {
		t.Fatalf("expected still 1 call after stop, got %d", calls)
	}
}

// --- Scheduler ordering test ---

func TestSchedulerFlushOrdering(t *testing.T) {
	resetScheduler()
	r := NewRef(1)

	// Computed depends on r.
	c := NewComputed(func() int {
		return r.Get() + 10
	})

	// Watch depends on the computed.
	var watchSaw []int
	WatchEffect(func() {
		watchSaw = append(watchSaw, c.Get())
	})

	// Initial: watch sees 11.
	if len(watchSaw) != 1 || watchSaw[0] != 11 {
		t.Fatalf("expected [11], got %v", watchSaw)
	}

	r.Set(2)
	DefaultScheduler.Flush()

	// Watch should see updated computed value 12.
	if len(watchSaw) != 2 || watchSaw[1] != 12 {
		t.Fatalf("expected [11,12], got %v", watchSaw)
	}
}

// --- Benchmark ---

func BenchmarkFlush1000Refs100Computeds(b *testing.B) {
	resetScheduler()

	const numRefs = 1000
	const numComputeds = 100

	refs := make([]*Ref[int], numRefs)
	for i := range refs {
		refs[i] = NewRef(i)
	}

	// Each computed reads 10 consecutive refs.
	computeds := make([]*Computed[int], numComputeds)
	for i := range computeds {
		start := (i * 10) % numRefs
		computeds[i] = NewComputed(func() int {
			sum := 0
			for j := 0; j < 10; j++ {
				sum += refs[(start+j)%numRefs].Get()
			}
			return sum
		})
	}

	// Pre-read all computeds to establish dependencies.
	for _, c := range computeds {
		_ = c.Get()
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// Mutate one ref per iteration.
		refs[n%numRefs].Set(n)
		DefaultScheduler.Flush()
		// Read a computed to trigger lazy eval.
		_ = computeds[n%numComputeds].Get()
	}
}
