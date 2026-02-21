package widget

import (
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// optionRotatorLeftGlyph returns the left-pointing chevron glyph from the
// default spritesheet.
func optionRotatorLeftGlyph() engine.Image { return IconChevronLeft() }

// ---------------------------------------------------------------------------
// OptionRotator
// ---------------------------------------------------------------------------

// OptionRotator is a compact selection widget consisting of a left chevron,
// a centered value label, and a right chevron. Clicking either chevron — or
// pressing Left/Right arrow keys when focused — cycles through a fixed list
// of string options. Wraps by default.
type OptionRotator struct {
	Component
	options  []string
	selected *Ref[int]
	wrap     bool
	onChange func(int, string)

	ignoreWatch bool
	watch       WatchHandle
	stopOptions func()       // stops reactive options Array binding
	boundIdxRef *Ref[int]    // set when BindSelected is active
	boundValRef *Ref[string] // set when BindValue is active

	leftBtn    Component
	rightBtn   Component
	valueLabel *Label

	leftGlyph  *sg.Node // sprite for left chevron glyph
	rightGlyph *sg.Node // sprite for right chevron glyph

	leftIcon  engine.Image // nil = use procedural glyph
	rightIcon engine.Image // nil = use procedural glyph

	font        *sg.FontFamily
	displaySize float64
}

// NewOptionRotator creates an OptionRotator with the given name and initial
// options list. The selected index starts at 0 and wrapping is enabled.
// Panics if options is empty.
func NewOptionRotator(name string, options []string, source *sg.FontFamily, displaySize float64) *OptionRotator {
	if len(options) == 0 {
		panic("willowui: NewOptionRotator: options must not be empty")
	}
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	or := &OptionRotator{
		options:     make([]string, len(options)),
		selected:    NewRef(0),
		wrap:        true,
		font:        font,
		displaySize: displaySize,
	}
	copy(or.options, options)

	initComponent(&or.Component, name)
	or.initBackground(name)
	or.initBorder(name)

	// Left chevron button.
	initComponent(&or.leftBtn, name+"-left")
	or.leftBtn.initBackground(name + "-left")
	or.leftBtn.initBorder(name + "-left")
	or.leftBtn.node.Interactable = true
	or.leftBtn.SetCursorShape(engine.CursorShapePointer)
	or.leftBtn.node.OnClick(func(_ sg.ClickContext) {
		if !or.enabled {
			return
		}
		or.Prev()
	})
	or.leftBtn.onVisualStateChange = func() { or.UpdateVisuals() }
	or.node.AddChild(or.leftBtn.node)

	// Left glyph sprite.
	or.leftGlyph = sg.NewSprite(name+"-left-glyph", sg.TextureRegion{})
	or.leftGlyph.SetCustomImage(IconChevronLeft())
	or.leftBtn.node.AddChild(or.leftGlyph)

	// Value label.
	or.valueLabel = NewLabel(name+"-label", options[0], source, displaySize)
	or.node.AddChild(or.valueLabel.Node())

	// Right chevron button.
	initComponent(&or.rightBtn, name+"-right")
	or.rightBtn.initBackground(name + "-right")
	or.rightBtn.initBorder(name + "-right")
	or.rightBtn.node.Interactable = true
	or.rightBtn.SetCursorShape(engine.CursorShapePointer)
	or.rightBtn.node.OnClick(func(_ sg.ClickContext) {
		if !or.enabled {
			return
		}
		or.Next()
	})
	or.rightBtn.onVisualStateChange = func() { or.UpdateVisuals() }
	or.node.AddChild(or.rightBtn.node)

	// Right glyph sprite.
	or.rightGlyph = sg.NewSprite(name+"-right-glyph", sg.TextureRegion{})
	or.rightGlyph.SetCustomImage(IconChevronRight())
	or.rightBtn.node.AddChild(or.rightGlyph)

	// Visual state hooks.
	or.onVisualStateChange = func() { or.UpdateVisuals() }
	or.onFocusChange = func(_ bool) { or.UpdateVisuals() }
	or.onThemeChange = func() { or.UpdateVisuals() }

	// Focus: single tab stop, participates in spatial nav, intercepts arrows.
	or.enableFocusNavigation()
	or.InterceptArrows = true
	or.ConsumeHandledKeys = false // widget's own Update handles the key; only block spatial nav
	or.SetHandleKey(func(key engine.Key) bool {
		sel := or.selected.Peek()
		n := len(or.options)
		switch key {
		case engine.KeyRight, engine.KeyUp:
			if or.wrap {
				return true
			}
			return sel < n-1
		case engine.KeyLeft, engine.KeyDown:
			if or.wrap {
				return true
			}
			return sel > 0
		}
		return false
	})

	// Keyboard polling via OnUpdate hook.
	or.node.OnUpdate = func(_ float64) {
		if !or.focused || !or.enabled {
			return
		}
		im := DefaultInputManager
		switch {
		case im.IsKeyJustAvailable(engine.KeyRight) || im.IsKeyJustAvailable(engine.KeyUp):
			im.Consume(engine.KeyRight)
			im.Consume(engine.KeyUp)
			or.Next()
		case im.IsKeyJustAvailable(engine.KeyLeft) || im.IsKeyJustAvailable(engine.KeyDown):
			im.Consume(engine.KeyLeft)
			im.Consume(engine.KeyDown)
			or.Prev()
		}
	}

	or.SetSize(200, 32)
	return or
}

// ---------------------------------------------------------------------------
// Navigation
// ---------------------------------------------------------------------------

// Next advances the selection by one step. Wraps to index 0 if wrap is
// enabled; no-op on the last option when wrap is disabled.
func (or *OptionRotator) Next() {
	sel := or.selected.Peek()
	n := len(or.options)
	if sel < n-1 {
		or.SetSelected(sel + 1)
	} else if or.wrap {
		or.SetSelected(0)
	}
}

// Prev steps the selection back by one. Wraps to the last option if wrap is
// enabled; no-op at index 0 when wrap is disabled.
func (or *OptionRotator) Prev() {
	sel := or.selected.Peek()
	if sel > 0 {
		or.SetSelected(sel - 1)
	} else if or.wrap {
		or.SetSelected(len(or.options) - 1)
	}
}

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// SetOptions replaces the options list entirely. The selected index is clamped
// to the new length. If the index changes as a result, OnChange fires.
func (or *OptionRotator) SetOptions(opts []string) {
	or.options = make([]string, len(opts))
	copy(or.options, opts)

	idx := or.selected.Peek()
	clamped := idx
	if clamped >= len(opts) {
		clamped = len(opts) - 1
	}
	if clamped < 0 {
		clamped = 0
	}

	fire := clamped != idx
	or.setSelectedInternal(clamped, fire)
}

// SetSelected selects the option at index i. The index is clamped to the
// valid range. Fires OnChange only if the selection changes.
func (or *OptionRotator) SetSelected(i int) {
	if i < 0 {
		i = 0
	}
	if i >= len(or.options) {
		i = len(or.options) - 1
	}
	old := or.selected.Peek()
	if i == old {
		return
	}
	or.setSelectedInternal(i, true)
}

// SetWrap controls whether cycling wraps around at the ends (default: true).
// When false, the corresponding chevron shows the Disabled visual state at
// the boundaries.
func (or *OptionRotator) SetWrap(v bool) {
	or.wrap = v
	or.UpdateVisuals()
}

// SetOnChange registers a callback invoked after each selection change.
func (or *OptionRotator) SetOnChange(fn func(int, string)) {
	or.onChange = fn
}

// SetChevronIcons overrides the procedural chevron glyphs. Either argument
// may be nil to keep the procedural default for that side.
func (or *OptionRotator) SetChevronIcons(left, right engine.Image) {
	or.leftIcon = left
	or.rightIcon = right
	or.UpdateVisuals()
}

// ---------------------------------------------------------------------------
// Reactive bindings
// ---------------------------------------------------------------------------

// BindOptions binds the options list to a reactive Array[string]. When the
// array changes the widget re-syncs its options and clamps the selection.
// Pass nil to detach.
func (or *OptionRotator) BindOptions(arr *Array[string]) {
	if or.stopOptions != nil {
		or.stopOptions()
		or.stopOptions = nil
	}
	if arr == nil {
		return
	}
	sync := func() {
		or.options = or.options[:0]
		arr.ForEach(func(_ int, s string) {
			or.options = append(or.options, s)
		})
		idx := or.selected.Peek()
		if idx >= len(or.options) {
			idx = max(0, len(or.options)-1)
		}
		or.setSelectedInternal(idx, false)
	}
	sync()
	h := arr.OnChange(func() { sync() })
	or.stopOptions = func() { h.Stop() }
}

// BindSelected binds the widget to a *Ref[int] representing the selected
// index. External changes to the ref update the widget; user interaction
// updates the ref. Replaces any previous binding.
func (or *OptionRotator) BindSelected(ref *Ref[int]) {
	or.watch.Stop()
	or.boundIdxRef = ref
	or.boundValRef = nil
	or.setSelectedInternal(ref.Peek(), false)
	or.watch = WatchValue(ref, func(_, newIdx int) {
		if or.ignoreWatch {
			return
		}
		or.setSelectedInternal(newIdx, false)
	})
}

// BindValue binds the widget to a *Ref[string] representing the current value
// string. On bind, the index is resolved by scanning options for an exact
// match. A value not present in the options list is silently ignored (index
// stays at 0). Replaces any previous binding.
func (or *OptionRotator) BindValue(ref *Ref[string]) {
	or.watch.Stop()
	or.boundValRef = ref
	or.boundIdxRef = nil
	idx := or.findOption(ref.Peek())
	or.setSelectedInternal(idx, false)
	or.watch = WatchValue(ref, func(_, newVal string) {
		if or.ignoreWatch {
			return
		}
		idx := or.findOption(newVal)
		or.setSelectedInternal(idx, false)
	})
}

// ---------------------------------------------------------------------------
// Read methods
// ---------------------------------------------------------------------------

// Selected returns the current selected index.
func (or *OptionRotator) Selected() int {
	return or.selected.Peek()
}

// Value returns the current selected option string.
func (or *OptionRotator) Value() string {
	idx := or.selected.Peek()
	if idx < 0 || idx >= len(or.options) {
		return ""
	}
	return or.options[idx]
}

// Options returns a copy of the current options list.
func (or *OptionRotator) Options() []string {
	out := make([]string, len(or.options))
	copy(out, or.options)
	return out
}

// SelectedRef returns the internal index Ref.
func (or *OptionRotator) SelectedRef() *Ref[int] {
	return or.selected
}

// ---------------------------------------------------------------------------
// Layout and size
// ---------------------------------------------------------------------------

// SetSize resizes the widget and re-positions its sub-components.
func (or *OptionRotator) SetSize(w, h float64) {
	or.Width = w
	or.Height = h
	or.resizeBackground(w, h)
	or.resizeBorder(w, h)
	or.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	or.updateLayout()
	or.UpdateVisuals()
	or.MarkLayoutDirty()
}

// SetEnabled enables or disables the widget and its chevron sub-components.
// Component.SetEnabled already fires onVisualStateChange → UpdateVisuals.
func (or *OptionRotator) SetEnabled(v bool) {
	or.Component.SetEnabled(v)
	or.leftBtn.node.Interactable = v
	or.rightBtn.node.Interactable = v
}

// Dispose stops reactive watches and disposes the component tree.
func (or *OptionRotator) Dispose() {
	if or.stopOptions != nil {
		or.stopOptions()
	}
	or.watch.Stop()
	or.valueLabel.Dispose()
	or.leftBtn.Dispose()
	or.rightBtn.Dispose()
	or.Component.Dispose()
}

// ---------------------------------------------------------------------------
// Visuals
// ---------------------------------------------------------------------------

// UpdateVisuals applies theme colors and layout for the current state.
func (or *OptionRotator) UpdateVisuals() {
	group := or.EffectiveTheme().OptionRotator.Group(or.Variant())

	// Outer widget state.
	st := computeState(or.enabled, or.focused, or.hovered, or.pressed)

	// Outer corner radius.
	cr := resolveCornerRadius(group.CornerRadius, or.Height)
	or.applyCornerRadius(cr)
	bg := group.Background.Resolve(st)
	or.applyBackground(bg)
	or.applyBorder(group.Border.Resolve(st), group.BorderWidth, bg)

	// Label text color.
	or.valueLabel.SetColor(group.TextColor.Resolve(st))

	// Chevron enabled state.
	sel := or.selected.Peek()
	nOpts := len(or.options)
	leftEnabled := or.enabled && (or.wrap || sel > 0)
	rightEnabled := or.enabled && (or.wrap || sel < nOpts-1)

	leftSt := computeState(leftEnabled, false, or.leftBtn.hovered, or.leftBtn.pressed)
	rightSt := computeState(rightEnabled, false, or.rightBtn.hovered, or.rightBtn.pressed)

	// Chevron corner radius.
	chevCr := group.Chevron.CornerRadius
	if chevCr < 0 {
		chevCr = or.Height / 2
	}

	// Left chevron visuals.
	or.leftBtn.applyCornerRadius(chevCr)
	leftBg := group.Chevron.Background.Resolve(leftSt)
	or.leftBtn.applyBackground(leftBg)
	or.leftBtn.applyBorder(group.Chevron.Border.Resolve(leftSt), group.Chevron.BorderWidth, leftBg)

	// Right chevron visuals.
	or.rightBtn.applyCornerRadius(chevCr)
	rightBg := group.Chevron.Background.Resolve(rightSt)
	or.rightBtn.applyBackground(rightBg)
	or.rightBtn.applyBorder(group.Chevron.Border.Resolve(rightSt), group.Chevron.BorderWidth, rightBg)

	// Glyph icons.
	iconSize := group.Chevron.IconSize
	if iconSize <= 0 {
		iconSize = 1.0
	}

	leftImg := or.leftIcon
	if leftImg == nil {
		if group.ChevronLeftIcon.Set {
			leftImg = group.ChevronLeftIcon.Image
		} else {
			leftImg = IconChevronLeft()
		}
	}
	rightImg := or.rightIcon
	if rightImg == nil {
		if group.ChevronRightIcon.Set {
			rightImg = group.ChevronRightIcon.Image
		} else {
			rightImg = IconChevronRight()
		}
	}

	or.leftGlyph.SetCustomImage(leftImg)
	or.leftGlyph.SetColor(group.Chevron.IconColor.Resolve(leftSt))

	or.rightGlyph.SetCustomImage(rightImg)
	or.rightGlyph.SetColor(group.Chevron.IconColor.Resolve(rightSt))

	// Center glyphs within their chevron button areas.
	// Use SetSize which divides by image dimensions for custom images,
	// converting the desired display size into the correct scale.
	chevW := group.Chevron.Width
	if chevW <= 0 {
		chevW = 20
	}
	pad := group.Padding
	h := or.Height - pad.Top - pad.Bottom

	// Display size: iconSize is a theme multiplier. Normalize to a 9px
	// base so the visual size stays consistent regardless of source
	// image resolution.
	const baseDisplaySize = 9.0
	displayPx := baseDisplaySize * iconSize

	or.leftGlyph.SetSize(displayPx, displayPx)
	or.leftGlyph.SetPosition((chevW-displayPx)/2, (h-displayPx)/2)

	or.rightGlyph.SetSize(displayPx, displayPx)
	or.rightGlyph.SetPosition((chevW-displayPx)/2, (h-displayPx)/2)

	or.applyFocusRing(group.FocusColor.Resolve(st), group.FocusRingWidth)
	or.MarkDrawDirty()
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// updateLayout positions the chevron buttons and value label within the widget.
func (or *OptionRotator) updateLayout() {
	group := or.EffectiveTheme().OptionRotator.Group(or.Variant())
	pad := group.Padding
	chevW := group.Chevron.Width
	if chevW <= 0 {
		chevW = 20
	}

	h := or.Height - pad.Top - pad.Bottom

	// Left button.
	or.leftBtn.Width = chevW
	or.leftBtn.Height = h
	or.leftBtn.resizeBackground(chevW, h)
	or.leftBtn.resizeBorder(chevW, h)
	or.leftBtn.node.SetPosition(pad.Left, pad.Top)
	or.leftBtn.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: chevW, Height: h}

	// Value label: centered in the remaining space between the chevrons.
	labelAreaX := pad.Left + chevW
	labelAreaW := or.Width - pad.Left - pad.Right - 2*chevW
	lx := labelAreaX + (labelAreaW-or.valueLabel.Width)/2
	ly := pad.Top + (h-or.valueLabel.Height)/2
	or.valueLabel.SetPosition(lx, ly)

	// Right button.
	or.rightBtn.Width = chevW
	or.rightBtn.Height = h
	or.rightBtn.resizeBackground(chevW, h)
	or.rightBtn.resizeBorder(chevW, h)
	or.rightBtn.node.SetPosition(or.Width-pad.Right-chevW, pad.Top)
	or.rightBtn.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: chevW, Height: h}
}

// setSelectedInternal clamps, sets, and optionally fires the selection.
func (or *OptionRotator) setSelectedInternal(i int, fire bool) {
	n := len(or.options)
	if n == 0 {
		return
	}
	if i < 0 {
		i = 0
	}
	if i >= n {
		i = n - 1
	}
	or.selected.Set(i)
	DefaultScheduler.Flush()
	or.valueLabel.SetText(or.options[i])
	or.updateLayout()
	or.UpdateVisuals()
	if fire {
		or.fireChange(i)
	}
}

// fireChange notifies bound refs and the onChange callback.
func (or *OptionRotator) fireChange(idx int) {
	or.ignoreWatch = true
	if or.boundIdxRef != nil {
		or.boundIdxRef.Set(idx)
		DefaultScheduler.Flush()
	}
	if or.boundValRef != nil {
		or.boundValRef.Set(or.options[idx])
		DefaultScheduler.Flush()
	}
	or.ignoreWatch = false
	if or.onChange != nil {
		or.onChange(idx, or.options[idx])
	}
}

// findOption returns the index of the first option equal to val, or 0 if not found.
func (or *OptionRotator) findOption(val string) int {
	for i, o := range or.options {
		if o == val {
			return i
		}
	}
	return 0
}
