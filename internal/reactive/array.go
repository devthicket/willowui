package reactive

import (
	"math/rand"
	"slices"
)

// ---------------------------------------------------------------------------
// Event types
// ---------------------------------------------------------------------------

type addedEvt[T any] struct {
	index int
	item  T
}

type removedEvt[T any] struct {
	index int
	item  T
}

type movedEvt struct {
	from, to int
}

// ---------------------------------------------------------------------------
// Callback wrapper types (used for O(n) pointer-based removal)
// ---------------------------------------------------------------------------

type onChangedCB struct{ fn func() }
type onReplacedCB struct{ fn func() }
type onAddedCB[T any] struct{ fn func(int, T) }
type onRemovedCB[T any] struct{ fn func(int, T) }
type onMovedCB struct{ fn func(int, int) }
type onItemChangedCB[T any] struct{ fn func(int, T) }

// ---------------------------------------------------------------------------
// Array[T]
// ---------------------------------------------------------------------------

// Array[T] is a reactive ordered collection. T is unconstrained (any).
// Every mutation bumps an internal version counter, notifying all
// WatchEffect and Computed subscribers through the normal reactive graph.
// Granular callbacks (OnAdded, OnRemoved, OnMoved, OnReplaced) carry
// operation detail for incremental or animated widget updates.
//
// Array implements reactiveSource directly — no intermediate *Ref allocation.
// bumpVersion is a no-op when there are no reactive subscribers (common case).
type Array[T any] struct {
	items       []T
	version     uint64
	subscribers []subscriber // reactive graph — populated only when tracked

	// versionRef is allocated lazily only if Version() is ever called.
	versionRef *Ref[uint64]

	// lenRef is allocated lazily only if LenRef() is ever called.
	// Updated in bumpVersion; Ref.Set no-ops when length is unchanged,
	// so length-preserving mutations (SetAt, Swap, Sort, …) cost nothing.
	lenRef *Ref[int]

	onAdded       []*onAddedCB[T]
	onRemoved     []*onRemovedCB[T]
	onMoved       []*onMovedCB
	onReplaced    []*onReplacedCB
	onChanged     []*onChangedCB
	onItemChanged []*onItemChangedCB[T]

	batching   bool
	batchDirty bool
}

// Array implements reactiveSource so track(a) works inside WatchEffect/Computed.
func (a *Array[T]) addSubscriber(sub subscriber) {
	for _, s := range a.subscribers {
		if s == sub {
			return
		}
	}
	a.subscribers = append(a.subscribers, sub)
}

func (a *Array[T]) removeSubscriber(sub subscriber) {
	for i, s := range a.subscribers {
		if s == sub {
			a.subscribers[i] = a.subscribers[len(a.subscribers)-1]
			a.subscribers[len(a.subscribers)-1] = nil
			a.subscribers = a.subscribers[:len(a.subscribers)-1]
			return
		}
	}
}

// NewArray creates an empty reactive array.
func NewArray[T any]() *Array[T] {
	return &Array[T]{}
}

// NewArrayFrom creates a reactive array from an existing slice.
// The input slice is copied — the caller retains ownership of the original.
func NewArrayFrom[T any](items []T) *Array[T] {
	return &Array[T]{items: append(make([]T, 0, len(items)), items...)}
}

// NewArrayFromAny creates a reactive Array[any] from a typed slice, boxing each
// element. This avoids the need for callers to manually convert []T to []any.
func NewArrayFromAny[T any](items []T) *Array[any] {
	boxed := make([]any, len(items))
	for i, v := range items {
		boxed[i] = v
	}
	return &Array[any]{items: boxed}
}

// NewArrayWithCap creates an empty reactive array with the given initial capacity.
func NewArrayWithCap[T any](cap int) *Array[T] {
	return &Array[T]{items: make([]T, 0, cap)}
}

// ---------------------------------------------------------------------------
// Internal notify helpers — zero-alloc fast paths for the common single-item ops
// ---------------------------------------------------------------------------

func (a *Array[T]) bumpVersion() {
	a.version++
	// Propagate to reactive graph only when something is actually watching.
	if len(a.subscribers) > 0 {
		for _, sub := range a.subscribers {
			sub.markDirty()
			DefaultScheduler.Enqueue(sub)
		}
	}
	// Keep the lazy versionRef in sync if it was ever exposed.
	if a.versionRef != nil {
		a.versionRef.Set(a.version)
	}
	// Keep the lazy lenRef in sync. Ref.Set is a no-op when unchanged,
	// so length-preserving mutations (SetAt, Swap, Sort, …) skip propagation.
	if a.lenRef != nil {
		a.lenRef.Set(len(a.items))
	}
}

// notifyReplaced fires OnReplaced + OnChange. No per-element detail.
func (a *Array[T]) notifyReplaced() {
	if a.batching {
		a.batchDirty = true
		return
	}
	a.bumpVersion()
	for _, c := range a.onReplaced {
		c.fn()
	}
	for _, c := range a.onChanged {
		c.fn()
	}
}

// notifyAdded fires OnAdded + OnChange for a single element. Zero allocation.
func (a *Array[T]) notifyAdded(index int, item T) {
	if a.batching {
		a.batchDirty = true
		return
	}
	a.bumpVersion()
	for _, c := range a.onAdded {
		c.fn(index, item)
	}
	for _, c := range a.onChanged {
		c.fn()
	}
}

// notifyRemoved fires OnRemoved + OnChange for a single element. Zero allocation.
func (a *Array[T]) notifyRemoved(index int, item T) {
	if a.batching {
		a.batchDirty = true
		return
	}
	a.bumpVersion()
	for _, c := range a.onRemoved {
		c.fn(index, item)
	}
	for _, c := range a.onChanged {
		c.fn()
	}
}

// notifyMoved fires OnMoved + OnChange. Zero allocation.
func (a *Array[T]) notifyMoved(from, to int) {
	if a.batching {
		a.batchDirty = true
		return
	}
	a.bumpVersion()
	for _, c := range a.onMoved {
		c.fn(from, to)
	}
	for _, c := range a.onChanged {
		c.fn()
	}
}

// notify is the general multi-element path used by Push (multi), Unshift (multi), Splice.
func (a *Array[T]) notify(added []addedEvt[T], removed []removedEvt[T], moved *movedEvt, replaced bool) {
	if a.batching {
		a.batchDirty = true
		return
	}
	a.bumpVersion()

	if replaced || (len(added) > 0 && len(removed) > 0) {
		for _, c := range a.onReplaced {
			c.fn()
		}
	} else {
		for _, e := range removed {
			for _, c := range a.onRemoved {
				c.fn(e.index, e.item)
			}
		}
		for _, e := range added {
			for _, c := range a.onAdded {
				c.fn(e.index, e.item)
			}
		}
		if moved != nil {
			for _, c := range a.onMoved {
				c.fn(moved.from, moved.to)
			}
		}
	}
	for _, c := range a.onChanged {
		c.fn()
	}
}

// ---------------------------------------------------------------------------
// Reading methods
// ---------------------------------------------------------------------------

// Len returns the number of elements and registers a reactive dependency.
func (a *Array[T]) Len() int {
	track(a)
	return len(a.items)
}

// At returns the element at index i. Supports negative indexing (At(-1) = last).
// Panics if out of bounds.
func (a *Array[T]) At(i int) T {
	track(a)
	if i < 0 {
		i = len(a.items) + i
	}
	return a.items[i]
}

// Slice returns a copy of elements [start, end) and registers a reactive dependency.
func (a *Array[T]) Slice(start, end int) []T {
	track(a)
	return append([]T{}, a.items[start:end]...)
}

// Clone returns a new Array with the same contents.
func (a *Array[T]) Clone() *Array[T] {
	track(a)
	return NewArrayFrom(a.items)
}

// Version returns a *Ref[uint64] that mirrors the array's version counter,
// for use with WatchValue. The Ref is allocated lazily on first call.
func (a *Array[T]) Version() *Ref[uint64] {
	if a.versionRef == nil {
		a.versionRef = NewRef[uint64](a.version)
	}
	return a.versionRef
}

// LenRef returns a *Ref[int] that reactively tracks the array's length.
// The Ref is allocated lazily on first call and kept in sync by bumpVersion.
// Because Ref.Set is a no-op when the value is unchanged, length-preserving
// mutations (SetAt, Swap, Sort, Reverse, …) never propagate through it.
func (a *Array[T]) LenRef() *Ref[int] {
	if a.lenRef == nil {
		a.lenRef = NewRef(len(a.items))
	}
	return a.lenRef
}

// ---------------------------------------------------------------------------
// Mutation methods
// ---------------------------------------------------------------------------

// Push appends one or more items to the end.
func (a *Array[T]) Push(items ...T) {
	if len(items) == 1 {
		// Fast path: single item — no allocation.
		idx := len(a.items)
		a.items = append(a.items, items[0])
		a.notifyAdded(idx, items[0])
		return
	}
	start := len(a.items)
	a.items = append(a.items, items...)
	// Skip building event structs when no granular OnAdded subscribers.
	if len(a.onAdded) == 0 {
		a.notifyReplaced()
		return
	}
	evts := make([]addedEvt[T], len(items))
	for i, item := range items {
		evts[i] = addedEvt[T]{index: start + i, item: item}
	}
	a.notify(evts, nil, nil, false)
}

// Pop removes and returns the last element. Panics if empty.
func (a *Array[T]) Pop() T {
	last := len(a.items) - 1
	item := a.items[last]
	var zero T
	a.items[last] = zero
	a.items = a.items[:last]
	a.notifyRemoved(last, item)
	return item
}

// Unshift prepends one or more items to the front.
func (a *Array[T]) Unshift(items ...T) {
	if len(items) == 1 {
		// Fast path: single item — no allocation.
		a.items = append(a.items, items[0]) // grow by one
		copy(a.items[1:], a.items)
		a.items[0] = items[0]
		a.notifyAdded(0, items[0])
		return
	}
	a.items = append(items, a.items...)
	if len(a.onAdded) == 0 {
		a.notifyReplaced()
		return
	}
	evts := make([]addedEvt[T], len(items))
	for i, item := range items {
		evts[i] = addedEvt[T]{index: i, item: item}
	}
	a.notify(evts, nil, nil, false)
}

// Shift removes and returns the first element. Panics if empty.
func (a *Array[T]) Shift() T {
	item := a.items[0]
	a.items = a.items[1:]
	a.notifyRemoved(0, item)
	return item
}

// Insert inserts item at index i, shifting existing elements right.
func (a *Array[T]) Insert(i int, item T) {
	a.items = append(a.items, item) // grow by one
	copy(a.items[i+1:], a.items[i:])
	a.items[i] = item
	a.notifyAdded(i, item)
}

// Remove removes the element at index i.
func (a *Array[T]) Remove(i int) {
	item := a.items[i]
	a.items = append(a.items[:i], a.items[i+1:]...)
	a.notifyRemoved(i, item)
}

// Splice removes del elements starting at start, then inserts items.
// Fires OnRemoved (remove-only), OnAdded (insert-only), or OnReplaced (both).
func (a *Array[T]) Splice(start, del int, items ...T) {
	removed := make([]removedEvt[T], del)
	for k := 0; k < del; k++ {
		removed[k] = removedEvt[T]{index: start + k, item: a.items[start+k]}
	}
	tail := append([]T{}, a.items[start+del:]...)
	a.items = append(a.items[:start], items...)
	a.items = append(a.items, tail...)
	added := make([]addedEvt[T], len(items))
	for k, item := range items {
		added[k] = addedEvt[T]{index: start + k, item: item}
	}
	a.notify(added, removed, nil, false)
}

// Swap swaps the elements at indices i and j.
// Fires OnMoved once with (i, j) to signal the swap.
func (a *Array[T]) Swap(i, j int) {
	a.items[i], a.items[j] = a.items[j], a.items[i]
	a.notifyMoved(i, j)
}

// Move moves the element at from to index to.
// Elements between from and to shift one position to fill the gap.
//
//	[A, B, C, D]  Move(0, 2) → [B, C, A, D]
//	[A, B, C, D]  Move(2, 0) → [C, A, B, D]
func (a *Array[T]) Move(from, to int) {
	if from == to {
		return
	}
	item := a.items[from]
	if from < to {
		copy(a.items[from:], a.items[from+1:to+1])
	} else {
		copy(a.items[to+1:], a.items[to:from])
	}
	a.items[to] = item
	a.notifyMoved(from, to)
}

// Reverse reverses the array in place.
func (a *Array[T]) Reverse() {
	for i, j := 0, len(a.items)-1; i < j; i, j = i+1, j-1 {
		a.items[i], a.items[j] = a.items[j], a.items[i]
	}
	a.notifyReplaced()
}

// Sort sorts the array using the provided less-than function.
func (a *Array[T]) Sort(less func(T, T) int) {
	slices.SortFunc(a.items, less)
	a.notifyReplaced()
}

// Shuffle randomizes the order of elements using Fisher-Yates.
func (a *Array[T]) Shuffle() {
	rand.Shuffle(len(a.items), func(i, j int) {
		a.items[i], a.items[j] = a.items[j], a.items[i]
	})
	a.notifyReplaced()
}

// Set replaces all contents with the provided slice (copied).
func (a *Array[T]) Set(items []T) {
	a.items = append(a.items[:0], items...)
	a.notifyReplaced()
}

// Fill sets elements in [start, end) to v.
func (a *Array[T]) Fill(v T, start, end int) {
	for i := start; i < end; i++ {
		a.items[i] = v
	}
	a.notifyReplaced()
}

// Truncate keeps only the first n elements.
func (a *Array[T]) Truncate(n int) {
	var zero T
	for i := n; i < len(a.items); i++ {
		a.items[i] = zero
	}
	a.items = a.items[:n]
	a.notifyReplaced()
}

// Clear removes all elements.
func (a *Array[T]) Clear() {
	var zero T
	for i := range a.items {
		a.items[i] = zero
	}
	a.items = a.items[:0]
	a.notifyReplaced()
}

// Reserve ensures the underlying slice has at least n capacity.
func (a *Array[T]) Reserve(n int) {
	if cap(a.items) < n {
		next := make([]T, len(a.items), n)
		copy(next, a.items)
		a.items = next
	}
}

// Shrink reallocates the underlying slice to exactly fit its contents.
func (a *Array[T]) Shrink() {
	next := make([]T, len(a.items))
	copy(next, a.items)
	a.items = next
}

// ---------------------------------------------------------------------------
// Batch
// ---------------------------------------------------------------------------

// Batch groups multiple mutations into a single notification.
// The outermost Batch controls the single notification; inner Batch calls
// while batching is already active run fn immediately without re-batching.
func (a *Array[T]) Batch(fn func()) {
	if a.batching {
		fn()
		return
	}
	a.batching = true
	a.batchDirty = false
	fn()
	a.batching = false
	if a.batchDirty {
		a.notifyReplaced()
	}
}

// ---------------------------------------------------------------------------
// Search and iteration (no reactive dependency tracking)
// ---------------------------------------------------------------------------

// Find returns the first element satisfying fn, and true. Returns zero, false if not found.
func (a *Array[T]) Find(fn func(T) bool) (T, bool) {
	for _, v := range a.items {
		if fn(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex returns the index of the first element satisfying fn, or -1.
func (a *Array[T]) FindIndex(fn func(T) bool) int {
	for i, v := range a.items {
		if fn(v) {
			return i
		}
	}
	return -1
}

// Some returns true if any element satisfies fn.
func (a *Array[T]) Some(fn func(T) bool) bool {
	for _, v := range a.items {
		if fn(v) {
			return true
		}
	}
	return false
}

// Every returns true if all elements satisfy fn (vacuously true for empty).
func (a *Array[T]) Every(fn func(T) bool) bool {
	for _, v := range a.items {
		if !fn(v) {
			return false
		}
	}
	return true
}

// ForEach calls fn for each element with its index.
func (a *Array[T]) ForEach(fn func(i int, v T)) {
	for i, v := range a.items {
		fn(i, v)
	}
}

// RandomItem returns a random element, or zero/false if empty.
func (a *Array[T]) RandomItem() (T, bool) {
	if len(a.items) == 0 {
		var zero T
		return zero, false
	}
	return a.items[rand.Intn(len(a.items))], true
}

// Filter returns a new Array containing elements that satisfy fn.
func (a *Array[T]) Filter(fn func(T) bool) *Array[T] {
	result := NewArray[T]()
	for _, v := range a.items {
		if fn(v) {
			result.items = append(result.items, v)
		}
	}
	return result
}

// Concat returns a new Array with all elements from a followed by all elements from others.
func (a *Array[T]) Concat(others ...*Array[T]) *Array[T] {
	total := len(a.items)
	for _, o := range others {
		total += len(o.items)
	}
	result := NewArrayWithCap[T](total)
	result.items = append(result.items, a.items...)
	for _, o := range others {
		result.items = append(result.items, o.items...)
	}
	return result
}

// ---------------------------------------------------------------------------
// Change callbacks
// ---------------------------------------------------------------------------

func removeFromSlice[S ~[]*E, E any](s *S, target *E) {
	for i, c := range *s {
		if c == target {
			last := len(*s) - 1
			(*s)[i] = (*s)[last]
			(*s)[last] = nil
			*s = (*s)[:last]
			return
		}
	}
}

// OnChange registers a callback that fires after every mutation.
// Returns a WatchHandle — call Stop() to unsubscribe.
func (a *Array[T]) OnChange(fn func()) WatchHandle {
	cb := &onChangedCB{fn: fn}
	a.onChanged = append(a.onChanged, cb)
	return WatchHandle{stopFn: func() { removeFromSlice(&a.onChanged, cb) }}
}

// OnAdded registers a callback that fires when elements are added.
func (a *Array[T]) OnAdded(fn func(index int, item T)) WatchHandle {
	cb := &onAddedCB[T]{fn: fn}
	a.onAdded = append(a.onAdded, cb)
	return WatchHandle{stopFn: func() { removeFromSlice(&a.onAdded, cb) }}
}

// OnRemoved registers a callback that fires when elements are removed.
func (a *Array[T]) OnRemoved(fn func(index int, item T)) WatchHandle {
	cb := &onRemovedCB[T]{fn: fn}
	a.onRemoved = append(a.onRemoved, cb)
	return WatchHandle{stopFn: func() { removeFromSlice(&a.onRemoved, cb) }}
}

// OnMoved registers a callback that fires when an element is moved.
func (a *Array[T]) OnMoved(fn func(from, to int)) WatchHandle {
	cb := &onMovedCB{fn: fn}
	a.onMoved = append(a.onMoved, cb)
	return WatchHandle{stopFn: func() { removeFromSlice(&a.onMoved, cb) }}
}

// OnReplaced registers a callback that fires when the array is restructured.
func (a *Array[T]) OnReplaced(fn func()) WatchHandle {
	cb := &onReplacedCB{fn: fn}
	a.onReplaced = append(a.onReplaced, cb)
	return WatchHandle{stopFn: func() { removeFromSlice(&a.onReplaced, cb) }}
}

// notifyItemChanged fires OnItemChanged + OnChange for a single element. Zero allocation.
func (a *Array[T]) notifyItemChanged(index int, item T) {
	if a.batching {
		a.batchDirty = true
		return
	}
	a.bumpVersion()
	for _, c := range a.onItemChanged {
		c.fn(index, item)
	}
	for _, c := range a.onChanged {
		c.fn()
	}
}

// SetAt replaces the item at index in-place and fires OnItemChanged.
// Panics if index is out of range.
func (a *Array[T]) SetAt(index int, item T) {
	a.items[index] = item
	a.notifyItemChanged(index, item)
}

// RemoveAt removes the element at index i (alias for Remove).
func (a *Array[T]) RemoveAt(i int) {
	a.Remove(i)
}

// OnItemChanged registers a callback that fires when a single item is updated in-place via SetAt.
func (a *Array[T]) OnItemChanged(fn func(index int, item T)) WatchHandle {
	cb := &onItemChangedCB[T]{fn: fn}
	a.onItemChanged = append(a.onItemChanged, cb)
	return WatchHandle{stopFn: func() { removeFromSlice(&a.onItemChanged, cb) }}
}
