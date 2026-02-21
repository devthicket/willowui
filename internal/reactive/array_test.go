package reactive

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

func TestArrayNewArray(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()
	if a.Len() != 0 {
		t.Fatalf("expected empty array, got len %d", a.Len())
	}
}

func TestArrayNewArrayFrom(t *testing.T) {
	resetScheduler()
	src := []int{1, 2, 3}
	a := NewArrayFrom(src)
	if a.Len() != 3 {
		t.Fatalf("expected len 3, got %d", a.Len())
	}
	// Verify copy independence.
	src[0] = 99
	if a.At(0) != 1 {
		t.Fatalf("expected 1 (copy independence), got %d", a.At(0))
	}
}

func TestArrayNewArrayWithCap(t *testing.T) {
	resetScheduler()
	a := NewArrayWithCap[string](16)
	if a.Len() != 0 {
		t.Fatalf("expected empty array, got len %d", a.Len())
	}
}

// ---------------------------------------------------------------------------
// Mutation: Push / Pop
// ---------------------------------------------------------------------------

func TestArrayPushAndPop(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()

	var addedIdx []int
	var addedItems []int
	a.OnAdded(func(i int, v int) {
		addedIdx = append(addedIdx, i)
		addedItems = append(addedItems, v)
	})

	a.Push(10, 20)
	if a.Len() != 2 || a.At(0) != 10 || a.At(1) != 20 {
		t.Fatalf("push failed: %v", a.Slice(0, a.Len()))
	}
	if len(addedIdx) != 2 || addedIdx[0] != 0 || addedIdx[1] != 1 {
		t.Fatalf("OnAdded indices wrong: %v", addedIdx)
	}

	var removedIdx []int
	var removedItems []int
	a.OnRemoved(func(i int, v int) {
		removedIdx = append(removedIdx, i)
		removedItems = append(removedItems, v)
	})

	got := a.Pop()
	if got != 20 || a.Len() != 1 {
		t.Fatalf("pop failed: got=%d len=%d", got, a.Len())
	}
	if len(removedIdx) != 1 || removedItems[0] != 20 {
		t.Fatalf("OnRemoved wrong: idx=%v items=%v", removedIdx, removedItems)
	}
}

// ---------------------------------------------------------------------------
// Mutation: Unshift / Shift
// ---------------------------------------------------------------------------

func TestArrayUnshiftAndShift(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{3, 4})

	var added []int
	a.OnAdded(func(i int, v int) { added = append(added, v) })

	a.Unshift(1, 2)
	want := []int{1, 2, 3, 4}
	for i, v := range want {
		if a.At(i) != v {
			t.Fatalf("at %d: want %d got %d", i, v, a.At(i))
		}
	}
	if len(added) != 2 || added[0] != 1 || added[1] != 2 {
		t.Fatalf("OnAdded wrong: %v", added)
	}

	got := a.Shift()
	if got != 1 || a.Len() != 3 {
		t.Fatalf("shift: got=%d len=%d", got, a.Len())
	}
}

// ---------------------------------------------------------------------------
// Mutation: Insert / Remove
// ---------------------------------------------------------------------------

func TestArrayInsertAndRemove(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 3})
	a.Insert(1, 2)
	if a.Len() != 3 || a.At(1) != 2 {
		t.Fatalf("insert failed: %v", a.Slice(0, 3))
	}

	a.Remove(0)
	if a.Len() != 2 || a.At(0) != 2 {
		t.Fatalf("remove failed: %v", a.Slice(0, 2))
	}
}

// ---------------------------------------------------------------------------
// Mutation: Splice callback routing
// ---------------------------------------------------------------------------

func TestArraySpliceRemoveOnly(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3, 4})

	removedFired := 0
	addedFired := 0
	replacedFired := 0
	a.OnRemoved(func(int, int) { removedFired++ })
	a.OnAdded(func(int, int) { addedFired++ })
	a.OnReplaced(func() { replacedFired++ })

	a.Splice(1, 2) // remove [2,3], insert nothing
	if a.Len() != 2 {
		t.Fatalf("expected len 2, got %d", a.Len())
	}
	if removedFired != 2 {
		t.Fatalf("expected 2 OnRemoved, got %d", removedFired)
	}
	if addedFired != 0 {
		t.Fatalf("expected 0 OnAdded, got %d", addedFired)
	}
	if replacedFired != 0 {
		t.Fatalf("expected 0 OnReplaced, got %d", replacedFired)
	}
}

func TestArraySpliceInsertOnly(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 4})

	removedFired := 0
	addedFired := 0
	replacedFired := 0
	a.OnRemoved(func(int, int) { removedFired++ })
	a.OnAdded(func(int, int) { addedFired++ })
	a.OnReplaced(func() { replacedFired++ })

	a.Splice(1, 0, 2, 3) // insert [2,3] at index 1, remove nothing
	if a.Len() != 4 {
		t.Fatalf("expected len 4, got %d", a.Len())
	}
	if addedFired != 2 {
		t.Fatalf("expected 2 OnAdded, got %d", addedFired)
	}
	if removedFired != 0 {
		t.Fatalf("expected 0 OnRemoved, got %d", removedFired)
	}
	if replacedFired != 0 {
		t.Fatalf("expected 0 OnReplaced, got %d", replacedFired)
	}
}

func TestArraySpliceMixedFiresOnReplaced(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})

	removedFired := 0
	addedFired := 0
	replacedFired := 0
	a.OnRemoved(func(int, int) { removedFired++ })
	a.OnAdded(func(int, int) { addedFired++ })
	a.OnReplaced(func() { replacedFired++ })

	a.Splice(0, 2, 9, 8, 7) // remove 2, insert 3
	if replacedFired != 1 {
		t.Fatalf("expected 1 OnReplaced, got %d", replacedFired)
	}
	if removedFired != 0 || addedFired != 0 {
		t.Fatalf("expected 0 OnRemoved/OnAdded on mixed splice, got r=%d a=%d", removedFired, addedFired)
	}
}

// ---------------------------------------------------------------------------
// Mutation: Swap / Move
// ---------------------------------------------------------------------------

func TestArraySwap(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3, 4})

	movedCalls := 0
	a.OnMoved(func(from, to int) { movedCalls++ })

	a.Swap(0, 3)
	if a.At(0) != 4 || a.At(3) != 1 {
		t.Fatalf("swap wrong: got %v", a.Slice(0, 4))
	}
	if movedCalls != 1 {
		t.Fatalf("expected 1 OnMoved call for swap, got %d", movedCalls)
	}
}

func TestArrayMoveForward(t *testing.T) {
	resetScheduler()
	// [A=0, B=1, C=2, D=3]  Move(0, 2) → [B, C, A, D]
	a := NewArrayFrom([]int{0, 1, 2, 3})

	var movedFrom, movedTo int
	a.OnMoved(func(f, t int) { movedFrom = f; movedTo = t })

	a.Move(0, 2)
	want := []int{1, 2, 0, 3}
	for i, v := range want {
		if a.At(i) != v {
			t.Fatalf("at %d: want %d got %d", i, v, a.At(i))
		}
	}
	if movedFrom != 0 || movedTo != 2 {
		t.Fatalf("OnMoved params wrong: from=%d to=%d", movedFrom, movedTo)
	}
}

func TestArrayMoveBackward(t *testing.T) {
	resetScheduler()
	// [A=0, B=1, C=2, D=3]  Move(2, 0) → [C, A, B, D]
	a := NewArrayFrom([]int{0, 1, 2, 3})
	a.Move(2, 0)
	want := []int{2, 0, 1, 3}
	for i, v := range want {
		if a.At(i) != v {
			t.Fatalf("at %d: want %d got %d", i, v, a.At(i))
		}
	}
}

// ---------------------------------------------------------------------------
// Mutation: Reverse / Sort / Shuffle / Set / Fill / Truncate / Clear
// ---------------------------------------------------------------------------

func TestArrayReverse(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})
	replacedFired := 0
	a.OnReplaced(func() { replacedFired++ })
	a.Reverse()
	if a.At(0) != 3 || a.At(2) != 1 {
		t.Fatalf("reverse failed: %v", a.Slice(0, 3))
	}
	if replacedFired != 1 {
		t.Fatalf("expected 1 OnReplaced, got %d", replacedFired)
	}
}

func TestArraySort(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{3, 1, 2})
	replacedFired := 0
	a.OnReplaced(func() { replacedFired++ })
	a.Sort(func(x, y int) int { return x - y })
	if a.At(0) != 1 || a.At(1) != 2 || a.At(2) != 3 {
		t.Fatalf("sort failed: %v", a.Slice(0, 3))
	}
	if replacedFired != 1 {
		t.Fatalf("expected 1 OnReplaced, got %d", replacedFired)
	}
}

func TestArrayShuffle(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3, 4, 5})
	replacedFired := 0
	a.OnReplaced(func() { replacedFired++ })
	a.Shuffle()
	if a.Len() != 5 {
		t.Fatalf("shuffle changed len: %d", a.Len())
	}
	if replacedFired != 1 {
		t.Fatalf("expected 1 OnReplaced, got %d", replacedFired)
	}
}

func TestArraySet(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})
	a.Set([]int{9, 8})
	if a.Len() != 2 || a.At(0) != 9 {
		t.Fatalf("set failed: %v", a.Slice(0, 2))
	}
}

func TestArrayFill(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3, 4})
	a.Fill(0, 1, 3)
	if a.At(0) != 1 || a.At(1) != 0 || a.At(2) != 0 || a.At(3) != 4 {
		t.Fatalf("fill failed: %v", a.Slice(0, 4))
	}
}

func TestArrayTruncate(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3, 4})
	replacedFired := 0
	a.OnReplaced(func() { replacedFired++ })
	a.Truncate(2)
	if a.Len() != 2 || a.At(1) != 2 {
		t.Fatalf("truncate failed: len=%d", a.Len())
	}
	if replacedFired != 1 {
		t.Fatalf("expected 1 OnReplaced, got %d", replacedFired)
	}
}

func TestArrayClear(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})
	replacedFired := 0
	a.OnReplaced(func() { replacedFired++ })
	a.Clear()
	if a.Len() != 0 {
		t.Fatalf("clear failed: len=%d", a.Len())
	}
	if replacedFired != 1 {
		t.Fatalf("expected 1 OnReplaced, got %d", replacedFired)
	}
}

// ---------------------------------------------------------------------------
// Negative indexing
// ---------------------------------------------------------------------------

func TestArrayNegativeIndex(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{10, 20, 30})
	if a.At(-1) != 30 {
		t.Fatalf("At(-1): expected 30, got %d", a.At(-1))
	}
	if a.At(-2) != 20 {
		t.Fatalf("At(-2): expected 20, got %d", a.At(-2))
	}
	if a.At(-3) != 10 {
		t.Fatalf("At(-3): expected 10, got %d", a.At(-3))
	}
}

// ---------------------------------------------------------------------------
// Batch
// ---------------------------------------------------------------------------

func TestArrayBatch(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})

	replacedCount := 0
	changedCount := 0
	a.OnReplaced(func() { replacedCount++ })
	a.OnChange(func() { changedCount++ })

	a.Batch(func() {
		a.Push(4)
		a.Push(5)
		a.Reverse()
	})

	if replacedCount != 1 {
		t.Fatalf("expected 1 OnReplaced from batch, got %d", replacedCount)
	}
	if changedCount != 1 {
		t.Fatalf("expected 1 OnChange from batch, got %d", changedCount)
	}
}

func TestArrayBatchNoop(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()

	changed := 0
	a.OnChange(func() { changed++ })

	// Batch with no mutations should not fire.
	a.Batch(func() {})
	if changed != 0 {
		t.Fatalf("expected no OnChange for empty batch, got %d", changed)
	}
}

// ---------------------------------------------------------------------------
// Reactive integration
// ---------------------------------------------------------------------------

func TestArrayWatchEffectRerunsOnMutation(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2})

	runs := 0
	WatchEffect(func() {
		_ = a.Len()
		runs++
	})
	if runs != 1 {
		t.Fatalf("expected 1 initial run, got %d", runs)
	}

	a.Push(3)
	DefaultScheduler.Flush()
	if runs != 2 {
		t.Fatalf("expected 2 runs after push, got %d", runs)
	}
}

func TestArrayComputedUpdatesOnLen(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})

	c := NewComputed(func() int {
		return a.Len() * 10
	})
	if c.Get() != 30 {
		t.Fatalf("expected 30, got %d", c.Get())
	}

	a.Push(4)
	DefaultScheduler.Flush()
	if c.Get() != 40 {
		t.Fatalf("expected 40 after push, got %d", c.Get())
	}
}

func TestArrayLenRef(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})

	lr := a.LenRef()
	if lr.Get() != 3 {
		t.Fatalf("expected 3, got %d", lr.Get())
	}

	a.Push(4)
	DefaultScheduler.Flush()
	if lr.Get() != 4 {
		t.Fatalf("expected 4 after push, got %d", lr.Get())
	}

	a.Remove(0)
	DefaultScheduler.Flush()
	if lr.Get() != 3 {
		t.Fatalf("expected 3 after remove, got %d", lr.Get())
	}

	// Verify same instance is returned on repeated calls.
	if a.LenRef() != lr {
		t.Fatal("LenRef() should return the same Ref on repeated calls")
	}

	// Length-preserving mutations should not change the ref value.
	a.SetAt(0, 99)
	if lr.Get() != 3 {
		t.Fatalf("expected 3 after SetAt (no length change), got %d", lr.Get())
	}

	// WatchEffect should not fire for length-preserving mutations.
	fires := 0
	WatchEffect(func() {
		_ = lr.Get()
		fires++
	})
	DefaultScheduler.Flush()
	if fires != 1 {
		t.Fatalf("expected 1 initial fire, got %d", fires)
	}
	a.SetAt(1, 42) // no length change
	DefaultScheduler.Flush()
	if fires != 1 {
		t.Fatalf("expected still 1 fire after SetAt, got %d", fires)
	}
	a.Push(5) // length changes
	DefaultScheduler.Flush()
	if fires != 2 {
		t.Fatalf("expected 2 fires after Push, got %d", fires)
	}
}

// ---------------------------------------------------------------------------
// WatchHandle.Stop on granular callbacks
// ---------------------------------------------------------------------------

func TestArrayOnChangedStop(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()

	calls := 0
	h := a.OnChange(func() { calls++ })

	a.Push(1)
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	h.Stop()
	a.Push(2)
	if calls != 1 {
		t.Fatalf("expected still 1 call after stop, got %d", calls)
	}
}

func TestArrayOnAddedStop(t *testing.T) {
	resetScheduler()
	a := NewArray[int]()

	calls := 0
	h := a.OnAdded(func(int, int) { calls++ })
	a.Push(1)
	h.Stop()
	a.Push(2)
	if calls != 1 {
		t.Fatalf("expected 1 call (before stop), got %d", calls)
	}
}

// ---------------------------------------------------------------------------
// ArrayMap / ArrayReduce / ArraySort / ArraySortDesc / IndexOf / Includes
// ---------------------------------------------------------------------------

func TestArrayMap(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3})
	out := ArrayMap(a, func(v int) string {
		return string(rune('A' - 1 + v))
	})
	if len(out) != 3 || out[0] != "A" || out[1] != "B" || out[2] != "C" {
		t.Fatalf("ArrayMap wrong: %v", out)
	}
}

func TestArrayReduce(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{1, 2, 3, 4})
	sum := ArrayReduce(a, func(acc, v int) int { return acc + v }, 0)
	if sum != 10 {
		t.Fatalf("ArrayReduce: expected 10, got %d", sum)
	}
}

func TestArraySort_Util(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{5, 3, 1, 4, 2})
	ArraySort(a)
	for i := 0; i < a.Len(); i++ {
		if a.At(i) != i+1 {
			t.Fatalf("ArraySort at %d: expected %d got %d", i, i+1, a.At(i))
		}
	}
}

func TestArraySortDesc(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{2, 5, 1, 3, 4})
	ArraySortDesc(a)
	if a.At(0) != 5 || a.At(4) != 1 {
		t.Fatalf("ArraySortDesc failed: %v", a.Slice(0, 5))
	}
}

func TestIndexOf(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]string{"x", "y", "z"})
	if IndexOf(a, "y") != 1 {
		t.Fatalf("IndexOf: expected 1")
	}
	if IndexOf(a, "w") != -1 {
		t.Fatalf("IndexOf: expected -1 for missing")
	}
}

func TestIncludes(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]int{10, 20, 30})
	if !Includes(a, 20) {
		t.Fatalf("Includes: expected true for 20")
	}
	if Includes(a, 99) {
		t.Fatalf("Includes: expected false for 99")
	}
}
