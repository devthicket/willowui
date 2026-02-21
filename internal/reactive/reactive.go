package reactive

import "fmt"

// ---------------------------------------------------------------------------
// Numeric constraint for Increment helper
// ---------------------------------------------------------------------------

// Numeric is a constraint for types that support arithmetic operations.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// ---------------------------------------------------------------------------
// Closure helpers — return func() suitable for SetOnClick, etc.
// ---------------------------------------------------------------------------

// Increment returns a func() that adds delta to the ref's value.
func Increment[T Numeric](r *Ref[T], delta T) func() {
	return func() { r.Update(func(v T) T { return v + delta }) }
}

// Toggle returns a func() that flips a boolean ref.
func Toggle(r *Ref[bool]) func() {
	return func() { r.Set(!r.Peek()) }
}

// Set returns a func() that sets the ref to the given value.
func Set[T comparable](r *Ref[T], val T) func() {
	return func() { r.Set(val) }
}

// ---------------------------------------------------------------------------
// BindFormatter — convert any Ref[T] to a *Ref[string] for BindText
// ---------------------------------------------------------------------------

// BindFormatter returns a *Ref[string] that stays in sync with source,
// converting values via fmt.Sprint.
//
// The returned WatchHandle must be stopped when the binding is no longer
// needed (e.g. via screen.TrackRef or an explicit h.Stop() call). Dropping
// the handle leaks a watcher on source for the lifetime of source.
func BindFormatter[T comparable](source *Ref[T]) (*Ref[string], WatchHandle) {
	out := NewRef(fmt.Sprint(source.Peek()))
	h := WatchValue(source, func(_, v T) {
		out.Set(fmt.Sprint(v))
	})
	return out, h
}

// BindFormatterf returns a *Ref[string] that stays in sync with source,
// converting values via fmt.Sprintf with the given format string.
//
// The returned WatchHandle must be stopped when the binding is no longer
// needed (e.g. via screen.TrackRef or an explicit h.Stop() call). Dropping
// the handle leaks a watcher on source for the lifetime of source.
func BindFormatterf[T comparable](source *Ref[T], format string) (*Ref[string], WatchHandle) {
	out := NewRef(fmt.Sprintf(format, source.Peek()))
	h := WatchValue(source, func(_, v T) {
		out.Set(fmt.Sprintf(format, v))
	})
	return out, h
}

// ---------------------------------------------------------------------------
// Dependency tracking context (single-threaded, stack-based)
// ---------------------------------------------------------------------------

// subscriber is any node that can subscribe to a reactive source.
type subscriber interface {
	flushable
	// addSource records that this subscriber depends on the given source.
	addSource(src reactiveSource)
}

// reactiveSource is any reactive node that provides a value and tracks subscribers.
type reactiveSource interface {
	addSubscriber(sub subscriber)
	removeSubscriber(sub subscriber)
}

// trackingFrame collects dependencies during a Computed or Watch evaluation.
type trackingFrame struct {
	deps []reactiveSource
}

// TrackingStack is the package-level dependency tracking stack.
var TrackingStack []*trackingFrame

// pushTrackingFrame starts collecting dependencies.
func pushTrackingFrame() *trackingFrame {
	f := &trackingFrame{}
	TrackingStack = append(TrackingStack, f)
	return f
}

// popTrackingFrame stops collecting and returns the frame.
func popTrackingFrame() *trackingFrame {
	n := len(TrackingStack)
	f := TrackingStack[n-1]
	TrackingStack = TrackingStack[:n-1]
	return f
}

// track registers a dependency on the current tracking frame (if any).
func track(src reactiveSource) {
	if len(TrackingStack) == 0 {
		return
	}
	frame := TrackingStack[len(TrackingStack)-1]
	// Deduplicate within the frame.
	for _, d := range frame.deps {
		if d == src {
			return
		}
	}
	frame.deps = append(frame.deps, src)
}

// ---------------------------------------------------------------------------
// Ref[T]
// ---------------------------------------------------------------------------

// Ref holds a single reactive value of type T. When the value changes,
// all subscribers (Computed nodes and Watches) are marked dirty and
// enqueued for the next scheduler flush.
type Ref[T comparable] struct {
	value       T
	subscribers []subscriber
}

// NewRef creates a new Ref with the given initial value.
func NewRef[T comparable](initial T) *Ref[T] {
	return &Ref[T]{value: initial}
}

// Get returns the current value and registers a dependency if called
// inside a Computed or Watch evaluation.
func (r *Ref[T]) Get() T {
	track(r)
	return r.value
}

// Set updates the value. If the new value equals the old value, this is
// a no-op. Otherwise, all subscribers are marked dirty and enqueued on
// the default scheduler.
func (r *Ref[T]) Set(v T) {
	if r.value == v {
		return
	}
	r.value = v
	for _, sub := range r.subscribers {
		sub.markDirty()
		DefaultScheduler.Enqueue(sub)
	}
}

// Peek returns the current value without registering a dependency.
func (r *Ref[T]) Peek() T {
	return r.value
}

// Update applies fn to the current value and sets the result.
// This is shorthand for r.Set(fn(r.Peek())).
func (r *Ref[T]) Update(fn func(T) T) {
	r.Set(fn(r.value))
}

func (r *Ref[T]) addSubscriber(sub subscriber) {
	// Avoid duplicates.
	for _, s := range r.subscribers {
		if s == sub {
			return
		}
	}
	r.subscribers = append(r.subscribers, sub)
}

func (r *Ref[T]) removeSubscriber(sub subscriber) {
	for i, s := range r.subscribers {
		if s == sub {
			r.subscribers[i] = r.subscribers[len(r.subscribers)-1]
			r.subscribers[len(r.subscribers)-1] = nil
			r.subscribers = r.subscribers[:len(r.subscribers)-1]
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Computed[T]
// ---------------------------------------------------------------------------

// Computed derives a value from reactive dependencies. It re-evaluates
// lazily: the function is only called on the next Get() after the node
// has been marked dirty.
type Computed[T comparable] struct {
	fn          func() T
	value       T
	dirty       bool
	sources     []reactiveSource
	subscribers []subscriber
	lastGen     uint64
}

// NewComputed creates a new Computed that derives its value by calling fn.
// The function is evaluated eagerly once to establish initial dependencies.
func NewComputed[T comparable](fn func() T) *Computed[T] {
	c := &Computed[T]{fn: fn, dirty: true}
	// Eagerly evaluate to capture initial dependencies and value.
	c.eval()
	return c
}

// Get returns the current value, re-evaluating if dirty. It also registers
// a dependency if called inside another Computed or Watch evaluation.
func (c *Computed[T]) Get() T {
	if c.dirty {
		c.eval()
	}
	track(c)
	return c.value
}

// Peek returns the current cached value without registering a dependency
// and without re-evaluating even if dirty.
func (c *Computed[T]) Peek() T {
	return c.value
}

func (c *Computed[T]) eval() {
	// Clear old subscriptions.
	for _, src := range c.sources {
		src.removeSubscriber(c)
	}
	c.sources = c.sources[:0]

	// Push tracking frame, evaluate, pop frame.
	// Use defer so a panic inside fn cannot leave a dangling frame on the
	// global TrackingStack, which would corrupt all subsequent dependency
	// tracking for the remainder of the process lifetime.
	frame := pushTrackingFrame()
	defer popTrackingFrame()
	newVal := c.fn()

	c.value = newVal
	c.dirty = false

	// Record new dependencies.
	c.sources = append(c.sources, frame.deps...)
	for _, src := range c.sources {
		src.addSubscriber(c)
	}
}

func (c *Computed[T]) markDirty() {
	if c.dirty {
		return
	}
	c.dirty = true
	// Propagate dirtiness to downstream subscribers.
	for _, sub := range c.subscribers {
		sub.markDirty()
		DefaultScheduler.Enqueue(sub)
	}
}

func (c *Computed[T]) flush(gen uint64) {
	if c.lastGen == gen {
		return
	}
	c.lastGen = gen
	// Computeds re-eval lazily on Get(), but flush marks them as processed
	// for this generation so downstream watches see consistent values.
	// We do NOT eagerly eval here — only mark generation.
}

func (c *Computed[T]) priority() int { return 0 }

func (c *Computed[T]) addSubscriber(sub subscriber) {
	for _, s := range c.subscribers {
		if s == sub {
			return
		}
	}
	c.subscribers = append(c.subscribers, sub)
}

func (c *Computed[T]) removeSubscriber(sub subscriber) {
	for i, s := range c.subscribers {
		if s == sub {
			c.subscribers[i] = c.subscribers[len(c.subscribers)-1]
			c.subscribers[len(c.subscribers)-1] = nil
			c.subscribers = c.subscribers[:len(c.subscribers)-1]
			return
		}
	}
}

func (c *Computed[T]) addSource(src reactiveSource) {
	c.sources = append(c.sources, src)
}

// ---------------------------------------------------------------------------
// Watch
// ---------------------------------------------------------------------------

// watchNode is an internal subscriber that runs a callback when its
// dependencies change.
type watchNode struct {
	fn      func()
	sources []reactiveSource
	dirty   bool
	stopped bool
	lastGen uint64
}

func (w *watchNode) markDirty() {
	w.dirty = true
}

func (w *watchNode) flush(gen uint64) {
	if w.stopped || !w.dirty || w.lastGen == gen {
		return
	}
	w.lastGen = gen
	w.dirty = false
	w.run()
}

func (w *watchNode) priority() int { return 1 }

func (w *watchNode) run() {
	// Clear old subscriptions.
	for _, src := range w.sources {
		src.removeSubscriber(w)
	}
	w.sources = w.sources[:0]

	frame := pushTrackingFrame()
	defer popTrackingFrame()
	w.fn()

	w.sources = append(w.sources, frame.deps...)
	for _, src := range w.sources {
		src.addSubscriber(w)
	}
}

func (w *watchNode) stop() {
	w.stopped = true
	for _, src := range w.sources {
		src.removeSubscriber(w)
	}
	w.sources = w.sources[:0]
}

func (w *watchNode) addSource(src reactiveSource) {
	w.sources = append(w.sources, src)
}

// WatchHandle allows stopping a watch subscription.
type WatchHandle struct {
	node   *watchNode
	stopFn func()
}

// Stop removes all subscriptions and prevents the watch from firing again.
func (h WatchHandle) Stop() {
	if h.node != nil {
		h.node.stop()
	}
	if h.stopFn != nil {
		h.stopFn()
	}
}

// WatchEffect creates a watch that re-runs fn whenever any reactive value
// accessed inside fn changes. The function is run once immediately to
// establish initial dependencies.
func WatchEffect(fn func()) WatchHandle {
	w := &watchNode{fn: fn}
	// Run once to capture initial dependencies.
	w.run()
	return WatchHandle{node: w}
}

// watchValueNode wraps a typed callback for WatchValue.
type watchValueNode[T comparable] struct {
	watchNode
	ref  *Ref[T]
	last T
	cb   func(old, new T)
}

// WatchValue watches a specific Ref and calls fn with the old and new values.
// The callback fires immediately with (current, current) to prime the initial
// state, then with (old, new) on each subsequent change.
func WatchValue[T comparable](ref *Ref[T], fn func(old, new T)) WatchHandle {
	w := &watchNode{}
	last := ref.Peek()

	w.fn = func() {
		cur := ref.Get()
		if cur != last {
			old := last
			last = cur
			fn(old, cur)
		}
	}

	// Run once to establish dependency on the ref.
	w.run()

	// Fire immediately so callers don't need separate initialization code.
	// old == new on this first call signals "initial state, not a transition".
	fn(last, last)

	return WatchHandle{node: w}
}
