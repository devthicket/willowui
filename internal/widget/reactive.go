package widget

import "github.com/devthicket/willowui/internal/reactive"

// Type aliases — re-export internal types as willowui types.
type Ref[T comparable] = reactive.Ref[T]
type Computed[T comparable] = reactive.Computed[T]
type Array[T any] = reactive.Array[T]
type WatchHandle = reactive.WatchHandle
type Scheduler = reactive.Scheduler

// DefaultScheduler is the package-level scheduler. It points into the
// internal reactive package so that Ref.Set() and Computed.markDirty()
// enqueue to the same scheduler that external callers flush.
var DefaultScheduler = &reactive.DefaultScheduler

// NewArray creates a new reactive array.
func NewArray[T any]() *Array[T] {
	return reactive.NewArray[T]()
}

// NewArrayFrom creates a reactive array from an existing slice.
func NewArrayFrom[T any](items []T) *Array[T] {
	return reactive.NewArrayFrom(items)
}

// NewRef creates a new reactive reference.
func NewRef[T comparable](initial T) *Ref[T] {
	return reactive.NewRef(initial)
}

// NewComputed creates a new computed reactive value.
func NewComputed[T comparable](fn func() T) *Computed[T] {
	return reactive.NewComputed(fn)
}

// WatchEffect creates a reactive effect that re-runs when dependencies change.
func WatchEffect(fn func()) WatchHandle {
	return reactive.WatchEffect(fn)
}

// WatchValue watches a specific Ref and calls fn with old and new values.
func WatchValue[T comparable](ref *Ref[T], fn func(old, new T)) WatchHandle {
	return reactive.WatchValue(ref, fn)
}

// bindRef is a helper for the common Bind pattern: stop old watch, sync
// current value, start new watch. apply is called immediately with the
// ref's current value and again whenever the ref changes.
func bindRef[T comparable](watch *WatchHandle, ref *Ref[T], apply func(T)) {
	watch.Stop()
	apply(ref.Peek())
	*watch = WatchValue(ref, func(_, newVal T) { apply(newVal) })
}
