package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
)

// ---------------------------------------------------------------------------
// Modifier mask & KeyCombo
// ---------------------------------------------------------------------------

// ModifierMask is a bitmask of modifier keys.
type ModifierMask uint8

const (
	ModNone  ModifierMask = 0
	ModCtrl  ModifierMask = 1 << iota // Ctrl or Cmd
	ModShift                          // Shift
	ModAlt                            // Alt / Option
)

// KeyCombo pairs an ebiten key with a modifier mask.
type KeyCombo struct {
	Key  engine.Key
	Mods ModifierMask
}

// Key creates a KeyCombo from a key and modifier mask.
func Key(k engine.Key, mods ModifierMask) KeyCombo {
	return KeyCombo{Key: k, Mods: mods}
}

// modifiersMatch returns true when the currently held modifier keys match
// the given mask exactly.
func modifiersMatch(mods ModifierMask) bool {
	ctrl := engine.IsKeyPressed(engine.KeyControl) || engine.IsKeyPressed(engine.KeyMeta)
	shift := engine.IsKeyPressed(engine.KeyShift)
	alt := engine.IsKeyPressed(engine.KeyAlt)

	wantCtrl := mods&ModCtrl != 0
	wantShift := mods&ModShift != 0
	wantAlt := mods&ModAlt != 0

	return ctrl == wantCtrl && shift == wantShift && alt == wantAlt
}

// ---------------------------------------------------------------------------
// BindHandle & keybind storage
// ---------------------------------------------------------------------------

// BindHandle identifies a registered keybind for later removal.
type BindHandle uint64

type keybind struct {
	combo  KeyCombo
	fn     func() bool
	handle BindHandle
	comp   *Component // nil for global binds
}

// ---------------------------------------------------------------------------
// FocusManager
// ---------------------------------------------------------------------------

// FocusManager tracks which component has keyboard focus. Owns UI keyboard
// dispatch: widget interception, scoped and global hotkeys/keybinds, Tab
// cycling, and spatial navigation.
type FocusManager struct {
	focused  *Component
	tabOrder []*Component

	// focusClaimed is set by SetFocus when any widget claims focus.
	// Update checks this to decide whether a click missed all widgets.
	focusClaimed bool

	// handledClick prevents double-processing when Update is called
	// multiple times in the same frame.
	handledClick bool

	// input is the InputManager used for key state queries and consumption.
	input *InputManager

	// keybinds stores scoped (comp != nil) and global (comp == nil) binds.
	keybinds   []keybind
	nextHandle BindHandle
}

// DefaultFocusManager is the package-level focus manager used by form
// controls. Clicking a focusable component routes through this manager
// so that at most one component holds focus at a time.
var DefaultFocusManager = NewFocusManager()

// NewFocusManager creates an empty focus manager wired to the
// DefaultInputManager.
func NewFocusManager() *FocusManager {
	return &FocusManager{
		input: DefaultInputManager,
	}
}

// ---------------------------------------------------------------------------
// Focus state
// ---------------------------------------------------------------------------

// SetFocus gives focus to c, removing it from the previously focused component.
// Passing nil clears focus.
func (fm *FocusManager) SetFocus(c *Component) {
	fm.focusClaimed = true
	if fm.focused == c {
		return
	}
	if fm.focused != nil {
		fm.focused.SetFocused(false)
	}
	fm.focused = c
	if c != nil {
		c.SetFocused(true)
	}
}

// ClearFocus removes focus from the currently focused component.
func (fm *FocusManager) ClearFocus() {
	fm.SetFocus(nil)
}

// Focused returns the currently focused component, or nil.
func (fm *FocusManager) Focused() *Component {
	return fm.focused
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

// Register adds a component to the tab order. Duplicates are ignored.
func (fm *FocusManager) Register(c *Component) {
	if c == nil {
		return
	}
	for _, existing := range fm.tabOrder {
		if existing == c {
			return
		}
	}
	fm.tabOrder = append(fm.tabOrder, c)
}

// Unregister removes a component from the tab order and cleans up any
// scoped keybinds. If it was focused, focus is cleared.
func (fm *FocusManager) Unregister(c *Component) {
	if c == nil {
		return
	}
	for i, existing := range fm.tabOrder {
		if existing == c {
			copy(fm.tabOrder[i:], fm.tabOrder[i+1:])
			fm.tabOrder[len(fm.tabOrder)-1] = nil
			fm.tabOrder = fm.tabOrder[:len(fm.tabOrder)-1]
			break
		}
	}
	// Remove scoped keybinds for this component.
	for i := len(fm.keybinds) - 1; i >= 0; i-- {
		if fm.keybinds[i].comp == c {
			fm.keybinds = append(fm.keybinds[:i], fm.keybinds[i+1:]...)
		}
	}
	if fm.focused == c {
		fm.ClearFocus()
	}
}

// ---------------------------------------------------------------------------
// Keybinds
// ---------------------------------------------------------------------------

// Bind registers a global keybind that fires regardless of focus. The
// callback returns true to consume the key from InputManager, false to
// let it pass through. Returns a handle for later removal via Unbind.
func (fm *FocusManager) Bind(combo KeyCombo, fn func() bool) BindHandle {
	h := fm.nextHandle
	fm.nextHandle++
	fm.keybinds = append(fm.keybinds, keybind{
		combo:  combo,
		fn:     fn,
		handle: h,
	})
	return h
}

// Unbind removes a global keybind by handle.
func (fm *FocusManager) Unbind(handle BindHandle) {
	for i, kb := range fm.keybinds {
		if kb.handle == handle {
			fm.keybinds = append(fm.keybinds[:i], fm.keybinds[i+1:]...)
			return
		}
	}
}

// BindScoped registers a keybind that fires only when comp (or a descendant)
// is focused. Automatically removed when the component is unregistered or
// disposed.
func (fm *FocusManager) BindScoped(comp *Component, combo KeyCombo, fn func() bool) {
	h := fm.nextHandle
	fm.nextHandle++
	fm.keybinds = append(fm.keybinds, keybind{
		combo:  combo,
		fn:     fn,
		handle: h,
		comp:   comp,
	})
}

// ---------------------------------------------------------------------------
// Update — main dispatch loop
// ---------------------------------------------------------------------------

// Update processes one frame of keyboard dispatch. Called once per frame by
// Screen.Update(). Must not be called by widgets directly.
//
// Dispatch order:
//  1. Click-outside detection (clear focus on click outside any focusable)
//  2. Focused widget interception (InterceptArrows → HandleKey for arrow keys)
//  3. Scoped hotkeys (focused component match)
//  4. Global hotkeys
//  5. Focus navigation (Tab/Shift+Tab, unhandled arrow keys → spatial nav)
func (fm *FocusManager) Update() {
	im := fm.input

	// --- Click outside detection ---
	clicked := engine.IsMouseButtonJustPressed(engine.MouseButtonLeft)
	if !clicked {
		fm.handledClick = false
		fm.focusClaimed = false
	} else if !fm.handledClick {
		fm.handledClick = true
		if fm.focused != nil && !fm.focusClaimed {
			fm.ClearFocus()
		}
		fm.focusClaimed = false
	}

	// Track which keys have been handled this frame so later steps skip them.
	type consumed = struct{}
	handled := make(map[engine.Key]consumed)

	// --- Step 1: Focused widget interception (arrow keys) ---
	// Check HandleKey first (pure, no side effects) so we only call
	// IsKeyJustAvailable — which consumes injected keys — when the widget
	// actually claims the direction.
	arrowKeys := [4]engine.Key{engine.KeyUp, engine.KeyDown, engine.KeyLeft, engine.KeyRight}
	if fm.focused != nil && fm.focused.InterceptArrows {
		for _, ak := range arrowKeys {
			if !fm.focused.HandleKey(ak) {
				continue // widget doesn't claim this direction; don't consume the key
			}
			if !im.IsKeyJustAvailable(ak) {
				continue
			}
			handled[ak] = consumed{}
			if fm.focused.ConsumeHandledKeys {
				im.Consume(ak)
			}
		}
	}

	// --- Step 2: Scoped hotkeys ---
	for i := range fm.keybinds {
		kb := &fm.keybinds[i]
		if kb.comp == nil {
			continue // global — handled in step 3
		}
		if !im.IsKeyJustAvailable(kb.combo.Key) {
			continue
		}
		if _, done := handled[kb.combo.Key]; done {
			continue
		}
		if !modifiersMatch(kb.combo.Mods) {
			continue
		}
		if !fm.isScopeActive(kb.comp) {
			continue
		}
		if kb.fn() {
			im.Consume(kb.combo.Key)
		}
		handled[kb.combo.Key] = consumed{}
	}

	// --- Step 3: Global hotkeys ---
	for i := range fm.keybinds {
		kb := &fm.keybinds[i]
		if kb.comp != nil {
			continue // scoped — already handled
		}
		if !im.IsKeyJustAvailable(kb.combo.Key) {
			continue
		}
		if _, done := handled[kb.combo.Key]; done {
			continue
		}
		if !modifiersMatch(kb.combo.Mods) {
			continue
		}
		if kb.fn() {
			im.Consume(kb.combo.Key)
		}
		handled[kb.combo.Key] = consumed{}
	}

	// --- Step 4: Focus navigation ---

	// Tab / Shift+Tab
	if _, done := handled[engine.KeyTab]; !done {
		if im.IsKeyJustAvailable(engine.KeyTab) {
			if engine.IsKeyPressed(engine.KeyShift) {
				fm.TabPrev()
			} else {
				fm.TabNext()
			}
			im.Consume(engine.KeyTab)
			handled[engine.KeyTab] = consumed{}
		}
	}

	// Arrow keys — spatial navigation for unhandled arrows.
	// Check for a spatial neighbor first (pure, no side effects) to avoid
	// consuming injected keys when there is nowhere to navigate.
	for _, ak := range arrowKeys {
		if _, done := handled[ak]; done {
			continue
		}
		if neighbor := fm.findSpatialNeighbor(ak); neighbor != nil {
			if im.IsKeyJustAvailable(ak) {
				fm.SetFocus(neighbor)
				im.Consume(ak)
			}
		}
	}
}

// isScopeActive returns true if comp or any descendant of comp is the
// currently focused component.
func (fm *FocusManager) isScopeActive(comp *Component) bool {
	if fm.focused == nil {
		return false
	}
	for c := fm.focused; c != nil; c = c.parent {
		if c == comp {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Tab navigation
// ---------------------------------------------------------------------------

// TabNext moves focus to the next component in tab order that has AllowTab,
// wrapping around. If nothing is focused, focuses the first eligible.
func (fm *FocusManager) TabNext() {
	n := len(fm.tabOrder)
	if n == 0 {
		return
	}
	start := 0
	if fm.focused != nil {
		idx := fm.indexOfFocused()
		if idx >= 0 {
			start = idx + 1
		}
	}
	for i := 0; i < n; i++ {
		candidate := fm.tabOrder[(start+i)%n]
		if fm.isTabEligible(candidate) {
			fm.SetFocus(candidate)
			return
		}
	}
}

// TabPrev moves focus to the previous component in tab order that has
// AllowTab, wrapping around. If nothing is focused, focuses the last eligible.
func (fm *FocusManager) TabPrev() {
	n := len(fm.tabOrder)
	if n == 0 {
		return
	}
	start := n - 1
	if fm.focused != nil {
		idx := fm.indexOfFocused()
		if idx >= 0 {
			start = idx - 1 + n
		}
	}
	for i := 0; i < n; i++ {
		candidate := fm.tabOrder[(start-i+n)%n]
		if fm.isTabEligible(candidate) {
			fm.SetFocus(candidate)
			return
		}
	}
}

// isTabEligible returns true if c can receive focus via Tab.
func (fm *FocusManager) isTabEligible(c *Component) bool {
	return c.Focusable && c.AllowTab && c.IsEnabled() && c.IsVisible() && !fm.hasDisabledAncestor(c)
}

// indexOfFocused returns the index of the currently focused component in
// the tab order, or -1 if not found.
func (fm *FocusManager) indexOfFocused() int {
	for i, c := range fm.tabOrder {
		if c == fm.focused {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Spatial navigation
// ---------------------------------------------------------------------------

// findSpatialNeighbor returns the best candidate to move focus to in the
// given direction, or nil if none found.
func (fm *FocusManager) findSpatialNeighbor(dir engine.Key) *Component {
	if fm.focused == nil {
		return nil
	}
	// Priority A: layout-aware sibling search.
	if c := fm.findLayoutNeighbor(dir); c != nil {
		return c
	}
	// Priority B: distance-based global search.
	return fm.findDistanceNeighbor(dir)
}

// findLayoutNeighbor scans the focused component's parent layout (VBox/HBox)
// for the nearest eligible sibling in the given direction.
func (fm *FocusManager) findLayoutNeighbor(dir engine.Key) *Component {
	parent := fm.focused.parent
	if parent == nil {
		return nil
	}

	layout := parent.Layout
	var delta int
	switch layout {
	case LayoutVBox:
		switch dir {
		case engine.KeyUp:
			delta = -1
		case engine.KeyDown:
			delta = 1
		default:
			return nil // Left/Right fall through to Priority B
		}
	case LayoutHBox:
		switch dir {
		case engine.KeyLeft:
			delta = -1
		case engine.KeyRight:
			delta = 1
		default:
			return nil // Up/Down fall through to Priority B
		}
	default:
		return nil // not a VBox/HBox parent
	}

	// Find the focused component's index among siblings.
	siblings := parent.children
	idx := -1
	for i, s := range siblings {
		if s == fm.focused {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}

	// Scan in the delta direction for an eligible sibling.
	for i := idx + delta; i >= 0 && i < len(siblings); i += delta {
		c := siblings[i]
		if fm.isSpatialEligible(c) {
			return c
		}
	}
	return nil
}

// findDistanceNeighbor performs a global search over all registered
// components using the directional score function.
func (fm *FocusManager) findDistanceNeighbor(dir engine.Key) *Component {
	fcx, fcy := fm.focused.WorldCenter()
	fx, fy, fw, fh := fm.focused.WorldBounds()

	var best *Component
	bestScore := math.MaxFloat64
	bestSecondary := math.MaxFloat64

	for _, c := range fm.tabOrder {
		if c == fm.focused {
			continue
		}
		if !fm.isSpatialEligible(c) {
			continue
		}

		cx, cy := c.WorldCenter()
		cx2, cy2, cw, ch := c.WorldBounds()

		// Directional filter + compute distances.
		// primary: edge-to-edge gap along the navigation axis (0 if overlapping).
		// secondary: center-to-center offset on the perpendicular axis.
		var primary, secondary float64
		switch dir {
		case engine.KeyRight:
			if cx <= fcx {
				continue
			}
			primary = math.Max(0, cx2-fx-fw)
			secondary = math.Abs(cy - fcy)
		case engine.KeyLeft:
			if cx >= fcx {
				continue
			}
			primary = math.Max(0, fx-cx2-cw)
			secondary = math.Abs(cy - fcy)
		case engine.KeyDown:
			if cy <= fcy {
				continue
			}
			primary = math.Max(0, cy2-fy-fh)
			secondary = math.Abs(cx - fcx)
		case engine.KeyUp:
			if cy >= fcy {
				continue
			}
			primary = math.Max(0, fy-cy2-ch)
			secondary = math.Abs(cx - fcx)
		default:
			continue
		}

		score := primary + secondary
		if score < bestScore || (score == bestScore && secondary < bestSecondary) {
			best = c
			bestScore = score
			bestSecondary = secondary
		}
	}
	return best
}

// isSpatialEligible returns true if c can receive focus via spatial nav.
func (fm *FocusManager) isSpatialEligible(c *Component) bool {
	return c.Focusable && c.AllowSpatial && c.IsEnabled() && c.IsVisible() && !fm.hasDisabledAncestor(c)
}

// hasDisabledAncestor walks up the parent chain and returns true if any
// ancestor is disabled or hidden.
func (fm *FocusManager) hasDisabledAncestor(c *Component) bool {
	for p := c.parent; p != nil; p = p.parent {
		if !p.IsEnabled() || !p.IsVisible() {
			return true
		}
	}
	return false
}
