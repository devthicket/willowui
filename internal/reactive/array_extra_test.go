package reactive

import (
	"testing"
)

// ---------------------------------------------------------------------------
// NewArrayFromAny
// ---------------------------------------------------------------------------

func TestNewArrayFromAny(t *testing.T) {
	resetScheduler()
	src := []int{10, 20, 30}
	a := NewArrayFromAny(src)
	if a.Len() != 3 {
		t.Fatalf("expected len 3, got %d", a.Len())
	}
	if a.At(0) != 10 || a.At(1) != 20 || a.At(2) != 30 {
		t.Fatalf("unexpected values: %v %v %v", a.At(0), a.At(1), a.At(2))
	}
	// Verify independence from original slice.
	src[0] = 999
	if a.At(0) != 10 {
		t.Fatal("NewArrayFromAny should copy elements")
	}
}

func TestNewArrayFromAnyEmpty(t *testing.T) {
	resetScheduler()
	a := NewArrayFromAny([]string{})
	if a.Len() != 0 {
		t.Fatalf("expected empty array, got len %d", a.Len())
	}
}

// ---------------------------------------------------------------------------
// Slice
// ---------------------------------------------------------------------------

func TestArraySlice(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3, 4, 5})
	s := a.Slice(1, 4)
	if len(s) != 3 || s[0] != 2 || s[1] != 3 || s[2] != 4 {
		t.Fatalf("Slice(1,4) wrong: %v", s)
	}
	// Verify returned slice is a copy.
	s[0] = 99
	if a.At(1) != 2 {
		t.Fatal("Slice should return a copy")
	}
}

func TestArraySliceEmpty(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})
	s := a.Slice(1, 1)
	if len(s) != 0 {
		t.Fatalf("expected empty slice, got %v", s)
	}
}

// ---------------------------------------------------------------------------
// Clone
// ---------------------------------------------------------------------------

func TestArrayClone(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{10, 20, 30})
	b := a.Clone()
	if b.Len() != 3 || b.At(0) != 10 || b.At(1) != 20 || b.At(2) != 30 {
		t.Fatalf("Clone values wrong")
	}
	// Verify independence.
	b.Push(40)
	if a.Len() != 3 {
		t.Fatal("Clone should be independent from original")
	}
	a.SetAt(0, 99)
	if b.At(0) != 10 {
		t.Fatal("Clone should be independent from original")
	}
}

// ---------------------------------------------------------------------------
// Version
// ---------------------------------------------------------------------------

func TestArrayVersion(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})
	vr := a.Version()
	v0 := vr.Get()

	a.Push(4)
	DefaultScheduler.Flush()
	v1 := vr.Get()
	if v1 <= v0 {
		t.Fatalf("expected version to increase after push: %d -> %d", v0, v1)
	}

	// Same instance returned on repeated calls.
	if a.Version() != vr {
		t.Fatal("Version() should return the same Ref instance")
	}
}

func TestArrayVersionReactiveWatch(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()
	vr := a.Version()

	fires := 0
	WatchEffect(func() {
		_ = vr.Get()
		fires++
	})
	if fires != 1 {
		t.Fatalf("expected 1 initial fire, got %d", fires)
	}

	a.Push(1)
	DefaultScheduler.Flush()
	if fires != 2 {
		t.Fatalf("expected 2 fires after push, got %d", fires)
	}
}

// ---------------------------------------------------------------------------
// LenRef — lazy allocation and length-preserving no-op
// ---------------------------------------------------------------------------

func TestArrayLenRefLazy(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2})
	lr := a.LenRef()
	if lr.Get() != 2 {
		t.Fatalf("expected 2, got %d", lr.Get())
	}
	// SetAt does not change length.
	a.SetAt(0, 99)
	DefaultScheduler.Flush()
	if lr.Get() != 2 {
		t.Fatalf("expected 2 after SetAt, got %d", lr.Get())
	}
}

// ---------------------------------------------------------------------------
// Find / FindIndex
// ---------------------------------------------------------------------------

func TestArrayFind(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3, 4, 5})

	v, ok := a.Find(func(x int) bool { return x > 3 })
	if !ok || v != 4 {
		t.Fatalf("expected (4, true), got (%d, %v)", v, ok)
	}

	_, ok = a.Find(func(x int) bool { return x > 100 })
	if ok {
		t.Fatal("expected false for not found")
	}
}

func TestArrayFindIndex(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]string{"apple", "banana", "cherry"})

	idx := a.FindIndex(func(s string) bool { return s == "banana" })
	if idx != 1 {
		t.Fatalf("expected 1, got %d", idx)
	}

	idx = a.FindIndex(func(s string) bool { return s == "grape" })
	if idx != -1 {
		t.Fatalf("expected -1, got %d", idx)
	}
}

// ---------------------------------------------------------------------------
// Some / Every
// ---------------------------------------------------------------------------

func TestArraySome(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})

	if !a.Some(func(x int) bool { return x == 2 }) {
		t.Fatal("expected true for Some(==2)")
	}
	if a.Some(func(x int) bool { return x > 10 }) {
		t.Fatal("expected false for Some(>10)")
	}
}

func TestArraySomeEmpty(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()
	if a.Some(func(int) bool { return true }) {
		t.Fatal("expected false for empty array")
	}
}

func TestArrayEvery(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{2, 4, 6})

	if !a.Every(func(x int) bool { return x%2 == 0 }) {
		t.Fatal("expected true for Every(even)")
	}
	if a.Every(func(x int) bool { return x > 3 }) {
		t.Fatal("expected false for Every(>3)")
	}
}

func TestArrayEveryEmpty(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()
	if !a.Every(func(int) bool { return false }) {
		t.Fatal("expected true for empty array (vacuous truth)")
	}
}

// ---------------------------------------------------------------------------
// ForEach
// ---------------------------------------------------------------------------

func TestArrayForEach(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{10, 20, 30})

	var indices []int
	var values []int
	a.ForEach(func(i int, v int) {
		indices = append(indices, i)
		values = append(values, v)
	})
	if len(indices) != 3 || indices[0] != 0 || indices[2] != 2 {
		t.Fatalf("indices wrong: %v", indices)
	}
	if values[0] != 10 || values[1] != 20 || values[2] != 30 {
		t.Fatalf("values wrong: %v", values)
	}
}

func TestArrayForEachEmpty(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()
	calls := 0
	a.ForEach(func(int, int) { calls++ })
	if calls != 0 {
		t.Fatalf("expected 0 calls for empty array, got %d", calls)
	}
}

// ---------------------------------------------------------------------------
// RandomItem
// ---------------------------------------------------------------------------

func TestArrayRandomItemNonEmpty(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{42})
	v, ok := a.RandomItem()
	if !ok || v != 42 {
		t.Fatalf("expected (42, true), got (%d, %v)", v, ok)
	}
}

func TestArrayRandomItemEmpty(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()
	v, ok := a.RandomItem()
	if ok {
		t.Fatal("expected false for empty array")
	}
	if v != 0 {
		t.Fatalf("expected zero value, got %d", v)
	}
}

// ---------------------------------------------------------------------------
// Filter
// ---------------------------------------------------------------------------

func TestArrayFilter(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3, 4, 5, 6})
	evens := a.Filter(func(x int) bool { return x%2 == 0 })
	if evens.Len() != 3 || evens.At(0) != 2 || evens.At(1) != 4 || evens.At(2) != 6 {
		t.Fatalf("Filter wrong: len=%d", evens.Len())
	}
}

func TestArrayFilterNoMatch(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 3, 5})
	result := a.Filter(func(x int) bool { return x%2 == 0 })
	if result.Len() != 0 {
		t.Fatalf("expected empty result, got len %d", result.Len())
	}
}

// ---------------------------------------------------------------------------
// Concat
// ---------------------------------------------------------------------------

func TestArrayConcat(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2})
	b := NewArrayFrom([]int{3, 4})
	c := NewArrayFrom([]int{5})
	result := a.Concat(b, c)
	if result.Len() != 5 {
		t.Fatalf("expected len 5, got %d", result.Len())
	}
	for i := 0; i < 5; i++ {
		if result.At(i) != i+1 {
			t.Fatalf("at %d: expected %d, got %d", i, i+1, result.At(i))
		}
	}
	// Verify independence.
	a.Push(99)
	if result.Len() != 5 {
		t.Fatal("Concat result should be independent")
	}
}

func TestArrayConcatEmpty(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2})
	result := a.Concat()
	if result.Len() != 2 {
		t.Fatalf("expected len 2 with no others, got %d", result.Len())
	}
}

// ---------------------------------------------------------------------------
// Reserve / Shrink
// ---------------------------------------------------------------------------

func TestArrayReserve(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()
	a.Reserve(100)
	// After reserve, should still have 0 elements but capacity >= 100.
	if a.Len() != 0 {
		t.Fatalf("expected len 0, got %d", a.Len())
	}
	// Push elements and verify no panic.
	for i := 0; i < 100; i++ {
		a.Push(i)
	}
	if a.Len() != 100 {
		t.Fatalf("expected len 100, got %d", a.Len())
	}
}

func TestArrayReserveNoOpWhenSufficient(t *testing.T) {
	resetScheduler()
	a := NewArrayWithCap[int](50)
	a.Push(1)
	a.Reserve(10) // already have cap >= 10, should be no-op
	if a.Len() != 1 || a.At(0) != 1 {
		t.Fatal("Reserve should not alter contents when capacity is sufficient")
	}
}

func TestArrayShrink(t *testing.T) {
	resetScheduler()
	a := NewArrayWithCap[int](100)
	a.Push(1)
	a.Push(2)
	a.Push(3)
	a.Shrink()
	if a.Len() != 3 || a.At(0) != 1 || a.At(1) != 2 || a.At(2) != 3 {
		t.Fatal("Shrink should not alter contents")
	}
}

// ---------------------------------------------------------------------------
// Batch — nested batch
// ---------------------------------------------------------------------------

func TestArrayBatchNested(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1})

	changed := 0
	a.OnChange(func() { changed++ })

	a.Batch(func() {
		a.Push(2)
		// Nested batch should just run fn immediately.
		a.Batch(func() {
			a.Push(3)
		})
		a.Push(4)
	})

	if changed != 1 {
		t.Fatalf("expected 1 OnChange from nested batch, got %d", changed)
	}
	if a.Len() != 4 {
		t.Fatalf("expected len 4, got %d", a.Len())
	}
}

// ---------------------------------------------------------------------------
// RemoveAt
// ---------------------------------------------------------------------------

func TestArrayRemoveAt(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{10, 20, 30})

	removedIdx := -1
	removedItem := -1
	a.OnRemoved(func(i int, v int) { removedIdx = i; removedItem = v })

	a.RemoveAt(1)
	if a.Len() != 2 || a.At(0) != 10 || a.At(1) != 30 {
		t.Fatalf("RemoveAt wrong: %v", a.Slice(0, a.Len()))
	}
	if removedIdx != 1 || removedItem != 20 {
		t.Fatalf("OnRemoved wrong: idx=%d item=%d", removedIdx, removedItem)
	}
}

// ---------------------------------------------------------------------------
// OnItemChanged + SetAt
// ---------------------------------------------------------------------------

func TestArrayOnItemChanged(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{10, 20, 30})

	var changedIdx []int
	var changedItems []int
	h := a.OnItemChanged(func(i int, v int) {
		changedIdx = append(changedIdx, i)
		changedItems = append(changedItems, v)
	})

	a.SetAt(1, 99)
	if len(changedIdx) != 1 || changedIdx[0] != 1 || changedItems[0] != 99 {
		t.Fatalf("OnItemChanged wrong: idx=%v items=%v", changedIdx, changedItems)
	}

	// Also fires OnChange.
	changeCalls := 0
	a.OnChange(func() { changeCalls++ })
	a.SetAt(2, 88)
	if changeCalls != 1 {
		t.Fatalf("expected 1 OnChange from SetAt, got %d", changeCalls)
	}

	// Stop the handle.
	h.Stop()
	a.SetAt(0, 77)
	if len(changedIdx) != 2 {
		t.Fatalf("expected 2 OnItemChanged calls (after stop), got %d", len(changedIdx))
	}
}

func TestArrayOnItemChangedDuringBatch(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})

	itemChangeCalls := 0
	a.OnItemChanged(func(int, int) { itemChangeCalls++ })

	a.Batch(func() {
		a.SetAt(0, 10)
		a.SetAt(1, 20)
	})

	// During batch, notifyItemChanged is suppressed; single notifyReplaced fires instead.
	if itemChangeCalls != 0 {
		t.Fatalf("expected 0 OnItemChanged during batch, got %d", itemChangeCalls)
	}
}

// ---------------------------------------------------------------------------
// OnRemoved / OnMoved / OnReplaced with subscribers
// ---------------------------------------------------------------------------

func TestArrayOnRemovedCallback(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})

	var removed []int
	h := a.OnRemoved(func(i int, v int) { removed = append(removed, v) })

	a.Remove(1) // remove element 2
	if len(removed) != 1 || removed[0] != 2 {
		t.Fatalf("OnRemoved wrong: %v", removed)
	}

	h.Stop()
	a.Remove(0)
	if len(removed) != 1 {
		t.Fatalf("expected still 1 after stop, got %d", len(removed))
	}
}

func TestArrayOnMovedCallback(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3, 4})

	var moves [][2]int
	h := a.OnMoved(func(from, to int) { moves = append(moves, [2]int{from, to}) })

	a.Move(0, 3)
	if len(moves) != 1 || moves[0] != [2]int{0, 3} {
		t.Fatalf("OnMoved wrong: %v", moves)
	}

	h.Stop()
	a.Move(1, 2)
	if len(moves) != 1 {
		t.Fatalf("expected still 1 after stop, got %d", len(moves))
	}
}

func TestArrayOnReplacedCallback(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{3, 1, 2})

	calls := 0
	h := a.OnReplaced(func() { calls++ })

	a.Sort(func(x, y int) int { return x - y })
	if calls != 1 {
		t.Fatalf("expected 1 OnReplaced, got %d", calls)
	}

	h.Stop()
	a.Reverse()
	if calls != 1 {
		t.Fatalf("expected still 1 after stop, got %d", calls)
	}
}

// ---------------------------------------------------------------------------
// Unshift with multiple items
// ---------------------------------------------------------------------------

func TestArrayUnshiftMultiple(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{4, 5})

	var added []int
	a.OnAdded(func(i int, v int) { added = append(added, v) })

	a.Unshift(1, 2, 3)
	if a.Len() != 5 {
		t.Fatalf("expected len 5, got %d", a.Len())
	}
	for i := 0; i < 5; i++ {
		if a.At(i) != i+1 {
			t.Fatalf("at %d: expected %d, got %d", i, i+1, a.At(i))
		}
	}
	if len(added) != 3 || added[0] != 1 || added[1] != 2 || added[2] != 3 {
		t.Fatalf("OnAdded wrong: %v", added)
	}
}

func TestArrayUnshiftMultipleNoOnAdded(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{3})

	replaced := 0
	a.OnReplaced(func() { replaced++ })

	// Multi-unshift with no OnAdded subscribers takes the notifyReplaced path.
	a.Unshift(1, 2)
	if a.Len() != 3 || a.At(0) != 1 {
		t.Fatalf("unshift failed: %v", a.Slice(0, a.Len()))
	}
	if replaced != 1 {
		t.Fatalf("expected 1 OnReplaced, got %d", replaced)
	}
}

// ---------------------------------------------------------------------------
// Push with multiple items and OnAdded subscribers
// ---------------------------------------------------------------------------

func TestArrayPushMultipleWithOnAdded(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1})

	var addedPairs [][2]int
	a.OnAdded(func(i int, v int) { addedPairs = append(addedPairs, [2]int{i, v}) })

	a.Push(2, 3, 4)
	if a.Len() != 4 {
		t.Fatalf("expected len 4, got %d", a.Len())
	}
	if len(addedPairs) != 3 {
		t.Fatalf("expected 3 OnAdded, got %d", len(addedPairs))
	}
	if addedPairs[0] != [2]int{1, 2} || addedPairs[1] != [2]int{2, 3} || addedPairs[2] != [2]int{3, 4} {
		t.Fatalf("OnAdded pairs wrong: %v", addedPairs)
	}
}

func TestArrayPushMultipleNoOnAdded(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()

	replaced := 0
	a.OnReplaced(func() { replaced++ })

	// Multi-push with no OnAdded subscriber -> notifyReplaced path.
	a.Push(1, 2, 3)
	if replaced != 1 {
		t.Fatalf("expected 1 OnReplaced, got %d", replaced)
	}
}

// ---------------------------------------------------------------------------
// Move no-op (from == to)
// ---------------------------------------------------------------------------

func TestArrayMoveNoOp(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})

	movedCalls := 0
	a.OnMoved(func(int, int) { movedCalls++ })

	a.Move(1, 1)
	if movedCalls != 0 {
		t.Fatalf("expected no OnMoved for same-index move, got %d", movedCalls)
	}
}

// ---------------------------------------------------------------------------
// Batching suppresses all notify types
// ---------------------------------------------------------------------------

func TestArrayBatchSuppressesRemoved(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})

	removedCalls := 0
	a.OnRemoved(func(int, int) { removedCalls++ })

	a.Batch(func() {
		a.Remove(0)
		a.Remove(0)
	})

	// During batch, individual removes are suppressed.
	if removedCalls != 0 {
		t.Fatalf("expected 0 OnRemoved during batch, got %d", removedCalls)
	}
}

func TestArrayBatchSuppressesMoved(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3, 4})

	movedCalls := 0
	a.OnMoved(func(int, int) { movedCalls++ })

	a.Batch(func() {
		a.Swap(0, 3)
		a.Move(1, 2)
	})

	if movedCalls != 0 {
		t.Fatalf("expected 0 OnMoved during batch, got %d", movedCalls)
	}
}

// ---------------------------------------------------------------------------
// Unshift single item
// ---------------------------------------------------------------------------

func TestArrayUnshiftSingle(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{2, 3})

	addedCalls := 0
	a.OnAdded(func(i int, v int) {
		if i != 0 || v != 1 {
			t.Fatalf("expected (0, 1), got (%d, %d)", i, v)
		}
		addedCalls++
	})

	a.Unshift(1)
	if a.Len() != 3 || a.At(0) != 1 || a.At(1) != 2 || a.At(2) != 3 {
		t.Fatalf("single unshift failed: %v", a.Slice(0, a.Len()))
	}
	if addedCalls != 1 {
		t.Fatalf("expected 1 OnAdded, got %d", addedCalls)
	}
}

// ---------------------------------------------------------------------------
// Reactive subscriber on Remove (exercises notifyRemoved -> bumpVersion -> subscriber path)
// ---------------------------------------------------------------------------

func TestArrayRemoveWithReactiveSubscriber(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})

	runs := 0
	WatchEffect(func() {
		_ = a.Len()
		runs++
	})
	if runs != 1 {
		t.Fatalf("expected 1, got %d", runs)
	}

	a.Remove(1)
	DefaultScheduler.Flush()
	if runs != 2 {
		t.Fatalf("expected 2 after Remove, got %d", runs)
	}
}

// ---------------------------------------------------------------------------
// Reactive subscriber on Move (exercises notifyMoved -> bumpVersion -> subscriber path)
// ---------------------------------------------------------------------------

func TestArrayMoveWithReactiveSubscriber(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})

	runs := 0
	WatchEffect(func() {
		_ = a.Len()
		runs++
	})
	if runs != 1 {
		t.Fatalf("expected 1, got %d", runs)
	}

	a.Move(0, 2)
	DefaultScheduler.Flush()
	if runs != 2 {
		t.Fatalf("expected 2 after Move, got %d", runs)
	}
}

// ---------------------------------------------------------------------------
// Array addSubscriber deduplication
// ---------------------------------------------------------------------------

func TestArrayAddSubscriberDedup(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()

	// Create two WatchEffects that both read from a.
	// The first one reads twice, which tests the track dedup within a frame.
	// But we also want to verify the addSubscriber dedup on re-run.
	runs := 0
	WatchEffect(func() {
		_ = a.Len()
		runs++
	})

	a.Push(1)
	DefaultScheduler.Flush()
	a.Push(2)
	DefaultScheduler.Flush()

	if runs != 3 {
		t.Fatalf("expected 3 runs, got %d", runs)
	}
}

// ---------------------------------------------------------------------------
// Record.Set fires onChange for existing key value change
// ---------------------------------------------------------------------------

func TestRecordSetExistingKeyFiresOnChange(t *testing.T) {
	resetScheduler()
	r := NewRecord()
	r.Set("x", 1)

	var changes []any
	r.OnChange(func(k string, v any) { changes = append(changes, v) })

	r.Set("x", 2)
	if len(changes) != 1 || changes[0] != 2 {
		t.Fatalf("expected [2], got %v", changes)
	}
}

// ---------------------------------------------------------------------------
// Reactive: duplicate subscriber de-duplication
// ---------------------------------------------------------------------------

func TestArrayDuplicateSubscriber(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()

	runs := 0
	WatchEffect(func() {
		// Access Len twice in the same effect -- should not create duplicate subscription.
		_ = a.Len()
		_ = a.Len()
		runs++
	})
	if runs != 1 {
		t.Fatalf("expected 1 initial, got %d", runs)
	}

	a.Push(1)
	DefaultScheduler.Flush()
	if runs != 2 {
		t.Fatalf("expected 2 runs, got %d", runs)
	}
}
