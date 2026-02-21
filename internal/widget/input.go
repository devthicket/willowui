package widget

import (
	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/engine"
)

// DefaultInputManager is the package-level InputManager singleton.
// Game logic and widgets read key state through this instead of ebiten directly.
var DefaultInputManager = NewInputManager()

// ListenerHandle identifies a registered key listener for later removal.
type ListenerHandle uint64

// listenerKind distinguishes the three listener types.
type listenerKind int

const (
	listenerDown listenerKind = iota
	listenerHeld
	listenerUp
)

// keyListener is a registered callback for a specific key.
type keyListener struct {
	key    engine.Key
	kind   listenerKind
	fn     func()
	handle ListenerHandle
}

// InputManager reads all keyboard state from ebiten once per frame, tracks
// which keys have been consumed by UI, and exposes availability queries and
// event-style listeners for game logic.
type InputManager struct {
	consumed map[engine.Key]bool

	// pressAvailable tracks whether a key's initial press was available (not
	// consumed). OnKeyUp only fires for keys that were available on press.
	pressAvailable map[engine.Key]bool

	listeners  []keyListener
	nextHandle ListenerHandle
}

// NewInputManager creates a new InputManager with empty state.
func NewInputManager() *InputManager {
	return &InputManager{
		consumed:       make(map[engine.Key]bool),
		pressAvailable: make(map[engine.Key]bool),
	}
}

// Update reads all ebiten keyboard state for the current frame. Must be
// called once per frame before FocusManager.Update().
func (im *InputManager) Update() {
	// Clear consumed set from the previous frame.
	for k := range im.consumed {
		delete(im.consumed, k)
	}
}

// FireListeners fires registered passthrough listeners for keys that were not
// consumed this frame. Must be called after FocusManager.Update() completes.
func (im *InputManager) FireListeners() {
	for i := range im.listeners {
		l := &im.listeners[i]
		switch l.kind {
		case listenerDown:
			if im.IsKeyJustAvailable(l.key) {
				l.fn()
			}
		case listenerHeld:
			if im.IsKeyAvailable(l.key) {
				l.fn()
			}
		case listenerUp:
			if engine.IsKeyJustReleased(l.key) {
				if im.pressAvailable[l.key] {
					l.fn()
				}
			}
		}
	}

	// Track press-availability for OnKeyUp filtering.
	// A key that was just pressed and available marks the start of an
	// available press lifetime. A key that was just released ends it.
	for k := range im.pressAvailable {
		if !engine.IsKeyPressed(k) {
			delete(im.pressAvailable, k)
		}
	}
	// Check all just-pressed keys this frame.
	pressed := engine.AppendJustPressedKeys(nil)
	for _, k := range pressed {
		if !im.consumed[k] {
			im.pressAvailable[k] = true
		}
	}
	// Also check injected keys via the core wrapper.
	// core.IsKeyJustPressed checks both real and injected; if a key is
	// just-pressed according to core but not in the real pressed list,
	// it came from injection.
}

// Consume marks a key as claimed by UI for this frame. Subsequent calls to
// IsKeyAvailable / IsKeyJustAvailable for this key will return false.
func (im *InputManager) Consume(key engine.Key) {
	im.consumed[key] = true
}

// IsKeyAvailable returns true if the key is currently held AND was not
// consumed by UI this frame.
func (im *InputManager) IsKeyAvailable(key engine.Key) bool {
	if im.consumed[key] {
		return false
	}
	return engine.IsKeyPressed(key)
}

// IsKeyJustAvailable returns true if the key was just pressed this frame
// AND was not consumed by UI. Uses the injection-aware wrapper so
// synthetically injected keys are included.
func (im *InputManager) IsKeyJustAvailable(key engine.Key) bool {
	if im.consumed[key] {
		return false
	}
	return core.IsKeyJustPressed(key)
}

// IsKeyJustReleased returns true if the key was just released this frame.
// Release detection is not affected by consumption — it reflects raw state.
func (im *InputManager) IsKeyJustReleased(key engine.Key) bool {
	return engine.IsKeyJustReleased(key)
}

// OnKeyDown registers a callback that fires once on the frame a key is first
// pressed and not consumed by UI. Returns a handle for later removal.
func (im *InputManager) OnKeyDown(key engine.Key, fn func()) ListenerHandle {
	h := im.nextHandle
	im.nextHandle++
	im.listeners = append(im.listeners, keyListener{
		key:    key,
		kind:   listenerDown,
		fn:     fn,
		handle: h,
	})
	return h
}

// OnKeyHeld registers a callback that fires every frame a key is held and
// not consumed by UI. Returns a handle for later removal.
func (im *InputManager) OnKeyHeld(key engine.Key, fn func()) ListenerHandle {
	h := im.nextHandle
	im.nextHandle++
	im.listeners = append(im.listeners, keyListener{
		key:    key,
		kind:   listenerHeld,
		fn:     fn,
		handle: h,
	})
	return h
}

// OnKeyUp registers a callback that fires once on the frame a key is released,
// but only if the key was available (not consumed) when it was first pressed.
// This prevents phantom release events for keys entirely consumed by UI.
// Returns a handle for later removal.
func (im *InputManager) OnKeyUp(key engine.Key, fn func()) ListenerHandle {
	h := im.nextHandle
	im.nextHandle++
	im.listeners = append(im.listeners, keyListener{
		key:    key,
		kind:   listenerUp,
		fn:     fn,
		handle: h,
	})
	return h
}

// RemoveListener removes a previously registered listener by handle.
func (im *InputManager) RemoveListener(handle ListenerHandle) {
	for i, l := range im.listeners {
		if l.handle == handle {
			im.listeners = append(im.listeners[:i], im.listeners[i+1:]...)
			return
		}
	}
}
