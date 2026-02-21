package widget

import (
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// Button is an interactive component with a colored background and centered
// text label. Visual state (normal, hovered, pressed, disabled) is driven
// by theme colors and updated via UpdateVisuals.
type Button struct {
	Component
	label       *Label  // centered text
	displaySize float64 // display font size
	onClick     func()
	textOX      float64     // current text offset X from theme
	textOY      float64     // current text offset Y from theme
	textWatch   WatchHandle // reactive label binding
	autoSize    bool        // when true, SetText auto-fits to content
}

// NewButton creates a Button with the given name, text label, font source, and
// display size. If displaySize is 0, the native atlas size is used.
func NewButton(name string, text string, source *sg.FontFamily, displaySize float64) *Button {
	b := &Button{displaySize: displaySize}
	initComponent(&b.Component, name)

	b.initBackground(name)
	b.initBorder(name)

	// Label: centered text child.
	b.label = NewLabel(name+"-label", text, source, displaySize)
	b.node.AddChild(b.label.Node())

	// Default size: text measurement + padding from theme config.
	// autoSize stays true so SetText will re-fit automatically.
	b.autoSize = true
	b.fitToText()

	// Wire OnClick on the component's node.
	b.node.OnClick(func(ctx sg.ClickContext) {
		if !b.enabled {
			return
		}
		DefaultFocusManager.SetFocus(&b.Component)
		if b.onClick != nil {
			b.onClick()
		}
	})

	b.onVisualStateChange = func() { b.UpdateVisuals() }
	b.onThemeChange = func() { b.UpdateVisuals() }
	b.SetCursorShape(engine.CursorShapePointer)

	// Focus: buttons participate in tab and spatial nav.
	b.enableFocusNavigation()

	// Keyboard activation: Space or Enter triggers onClick.
	b.onFocusChange = func(focused bool) { b.UpdateVisuals() }
	b.node.OnUpdate = func(_ float64) {
		if !b.focused || !b.enabled {
			return
		}
		im := DefaultInputManager
		if im.IsKeyJustAvailable(engine.KeySpace) || im.IsKeyJustAvailable(engine.KeyEnter) {
			if b.onClick != nil {
				b.onClick()
			}
			im.Consume(engine.KeySpace)
			im.Consume(engine.KeyEnter)
		}
	}

	b.UpdateVisuals()
	return b
}

// LabelWidth returns the width of the button's text label.
func (b *Button) LabelWidth() float64 {
	if b.label == nil {
		return 0
	}
	return b.label.Width
}

// LabelHeight returns the height of the button's text label.
func (b *Button) LabelHeight() float64 {
	if b.label == nil {
		return 0
	}
	return b.label.Height
}

// LabelLabel returns the button's label widget, or nil if none.
func (b *Button) LabelLabel() *Label {
	return b.label
}

// TextOY returns the current text vertical offset from the theme. Used for testing.
func (b *Button) TextOY() float64 { return b.textOY }

// LabelText returns the current text of the button's label, or "" if none.
func (b *Button) LabelText() string {
	if b.label == nil {
		return ""
	}
	return b.label.Text()
}

// SetText updates the button's label text. If auto-size is enabled (the
// default), the button resizes to fit the new text content.
func (b *Button) SetText(t string) {
	b.label.SetText(t)
	if b.autoSize {
		b.fitToText()
	} else {
		b.centerLabel()
	}
}

// BindText binds the button's label to a reactive Ref[string].
// The label updates automatically whenever the ref changes.
// Any previous binding is stopped first.
func (b *Button) BindText(ref *Ref[string]) {
	b.textWatch.Stop()
	b.textWatch = WatchValue(ref, func(_, newVal string) {
		b.SetText(newVal)
	})
}

// SetOnClick sets the callback invoked when the button is clicked.
func (b *Button) SetOnClick(fn func()) {
	b.onClick = fn
}

// SetSize sets the button dimensions and updates the background, label position,
// and hit shape so the button is immediately clickable. Calling SetSize disables
// auto-size — the button will no longer resize when SetText is called.
func (b *Button) SetSize(w, h float64) {
	b.autoSize = false
	b.applySize(w, h)
}

// SetAutoSize enables or disables automatic sizing to fit text content.
// When enabled, the button resizes on every SetText call. When disabled
// (the default after any explicit SetSize call), the button keeps its
// current dimensions.
func (b *Button) SetAutoSize(v bool) {
	b.autoSize = v
	if v {
		b.fitToText()
	}
}

// AutoSize reports whether the button auto-sizes to fit its text.
func (b *Button) AutoSize() bool {
	return b.autoSize
}

// fitToText computes the button size from its label + theme padding.
func (b *Button) fitToText() {
	th := b.EffectiveTheme()
	group := th.Button.Group(b.Variant())
	pad := resolveAutoInsets(group.Padding, defaultButtonPadding)
	b.applySize(b.label.Width+pad.Horizontal(), b.label.Height+pad.Vertical())
}

// applySize is the internal size setter that doesn't touch autoSize.
func (b *Button) applySize(w, h float64) {
	b.Width = w
	b.Height = h
	b.resizeBackground(w, h)
	b.resizeBorder(w, h)
	b.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	b.centerLabel()
	b.UpdateVisuals()
	b.MarkLayoutDirty()
}

// SetEnabled overrides Component.SetEnabled to also update visuals.
func (b *Button) SetEnabled(v bool) {
	b.Component.SetEnabled(v)
	b.UpdateVisuals()
}

// UpdateVisuals applies theme colors based on the current interaction state.
func (b *Button) UpdateVisuals() {
	b.state = computeState(b.enabled, b.focused, b.hovered, b.pressed)
	th := b.EffectiveTheme()
	group := th.Button.Group(b.Variant())
	b.applyCornerRadius(group.CornerRadius)
	bg := group.Background.Resolve(b.state)
	b.applyBackground(bg)
	b.applyBorder(group.Border.Resolve(b.state), group.BorderWidth, bg)
	if b.label != nil {
		b.label.SetColor(group.TextColor.Resolve(b.state))
	}

	newOX := group.OffsetX.Resolve(b.state)
	newOY := group.OffsetY.Resolve(b.state)
	if newOX != b.OffsetX || newOY != b.OffsetY {
		b.OffsetX = newOX
		b.OffsetY = newOY
		b.node.SetPosition(b.X+b.OffsetX, b.Y+b.OffsetY)
		// Compensate the hit shape so it stays at the original layout
		// position. Without this the hit area shifts with the visual
		// offset, causing hover jitter at the edges.
		b.node.HitShape = sg.HitRect{
			X: -b.OffsetX, Y: -b.OffsetY,
			Width: b.Width, Height: b.Height,
		}
	}

	newTOX := group.TextOffsetX.Resolve(b.state)
	newTOY := group.TextOffsetY.Resolve(b.state)
	if newTOX != b.textOX || newTOY != b.textOY {
		b.textOX = newTOX
		b.textOY = newTOY
		b.centerLabel()
	}

	b.applyFocusRing(group.FocusColor.Resolve(b.state), group.FocusRingWidth)
}

// Dispose cleans up the button and its label.
func (b *Button) Dispose() {
	b.textWatch.Stop()
	if b.label != nil {
		b.label.Dispose()
	}
	b.Component.Dispose()
}

// centerLabel positions the label node at the center of the button,
// applying any per-state text offsets from the theme.
func (b *Button) centerLabel() {
	if b.label == nil {
		return
	}
	lx := (b.Width-b.label.Width)/2 + b.textOX
	ly := (b.Height-b.label.Height)/2 + b.textOY
	b.label.SetPosition(lx, ly)
}

// ---------------------------------------------------------------------------
// IconButton
// ---------------------------------------------------------------------------

// IconLabelPosition controls where the text label appears relative to the icon.
type IconLabelPosition int

const (
	// IconLabelBelow places the label beneath the icon (default).
	IconLabelBelow IconLabelPosition = iota
	// IconLabelRight places the label to the right of the icon.
	IconLabelRight
)

// IconButton is an icon-first button that renders a sprite as its primary
// content, with an optional text label beneath or beside it. It participates
// in the same focus, hover, pressed, and disabled state system as Button.
type IconButton struct {
	Component
	icon         *sg.Node     // sprite child
	iconW, iconH float64      // explicit icon dimensions (0 = from theme)
	iconImg      engine.Image // direct image override (nil = use key/theme)
	iconKey      string       // theme icon key (not yet implemented)
	label        *Label       // optional text label
	labelPos     IconLabelPosition
	embedded     bool // suppresses background/border/focus ring when inside a composite
	active       bool
	activeRef    *Ref[bool]
	activeWatch  WatchHandle
	onClick      func()
}

// NewIconButton creates an icon-only button. The icon is initially blank;
// call SetIconImage or SetIconKey to provide an icon source.
func NewIconButton(name string) *IconButton {
	ib := &IconButton{}
	initComponent(&ib.Component, name)

	ib.initBackground(name)
	ib.initBorder(name)

	// Icon sprite (white pixel, tinted via node.Color).
	ib.icon = sg.NewSprite(name+"-icon", sg.TextureRegion{})
	ib.node.AddChild(ib.icon)

	// Default size.
	ib.applySize(32, 32)

	// Wire OnClick. Icon buttons do not claim keyboard focus on mouse click —
	// they fire their action and leave focus wherever it was. Focus moves to
	// an icon button only via keyboard Tab navigation.
	ib.node.OnClick(func(ctx sg.ClickContext) {
		if !ib.enabled {
			return
		}
		if ib.onClick != nil {
			ib.onClick()
		}
	})

	ib.onVisualStateChange = func() { ib.UpdateVisuals() }
	ib.onThemeChange = func() { ib.UpdateVisuals() }
	ib.SetCursorShape(engine.CursorShapePointer)

	// Focus: icon buttons participate in tab and spatial nav.
	ib.enableFocusNavigation()

	// Keyboard activation: Space or Enter triggers onClick.
	ib.onFocusChange = func(focused bool) { ib.UpdateVisuals() }
	ib.node.OnUpdate = func(_ float64) {
		if !ib.focused || !ib.enabled {
			return
		}
		im := DefaultInputManager
		if im.IsKeyJustAvailable(engine.KeySpace) || im.IsKeyJustAvailable(engine.KeyEnter) {
			if ib.onClick != nil {
				ib.onClick()
			}
			im.Consume(engine.KeySpace)
			im.Consume(engine.KeyEnter)
		}
	}

	ib.UpdateVisuals()
	return ib
}

// SetIconKey sets the icon using a sprite key from the theme's sprites map.
// If the key is found in the effective theme, the resolved sprite image is
// applied; otherwise the icon is cleared to the white pixel fallback.
func (ib *IconButton) SetIconKey(key string) {
	ib.iconKey = key
	if t := ib.EffectiveTheme(); t != nil {
		sr := t.GetSprite(key)
		if sr.Set {
			ib.SetIconImage(sr.Image)
			return
		}
	}
	ib.icon.SetTextureRegion(sg.TextureRegion{})
	ib.UpdateVisuals()
}

// SetIconImage sets the icon from a direct engine.Image override.
func (ib *IconButton) SetIconImage(img engine.Image) {
	ib.iconImg = img
	ib.icon.SetCustomImage(img)
	ib.layoutChildren()
	ib.UpdateVisuals()
}

// SetIconSize overrides the icon display dimensions. Pass 0,0 to restore
// the size from the theme (IconSize field, square by default).
func (ib *IconButton) SetIconSize(w, h float64) {
	ib.iconW = w
	ib.iconH = h
	ib.layoutChildren()
}

// SetLabel attaches an optional text label to the button.
func (ib *IconButton) SetLabel(text string, source *sg.FontFamily, size float64) {
	if ib.label == nil {
		ib.label = NewLabel(ib.node.Name+"-label", text, source, size)
		ib.node.AddChild(ib.label.Node())
	} else {
		ib.label.SetText(text)
	}
	ib.layoutChildren()
	ib.UpdateVisuals()
}

// SetLabelPosition sets whether the label appears below or to the right of
// the icon. Defaults to IconLabelBelow.
func (ib *IconButton) SetLabelPosition(pos IconLabelPosition) {
	ib.labelPos = pos
	ib.layoutChildren()
}

// SetEnabled overrides Component.SetEnabled to also update visuals.
func (ib *IconButton) SetEnabled(v bool) {
	ib.Component.SetEnabled(v)
	ib.UpdateVisuals()
}

// SetActive sets the toggle-style active highlight on this button.
func (ib *IconButton) SetActive(v bool) {
	ib.active = v
	if ib.activeRef != nil {
		ib.activeRef.Set(v)
		DefaultScheduler.Flush()
	}
	ib.UpdateVisuals()
}

// BindActive binds the button's active state to an external Ref[bool].
// The button's active highlight reflects the ref value; clicking the button
// does NOT automatically toggle the ref — that is the caller's responsibility.
func (ib *IconButton) BindActive(ref *Ref[bool]) {
	ib.activeWatch.Stop()
	ib.activeRef = ref
	ib.active = ref.Peek()
	ib.UpdateVisuals()
	ib.activeWatch = WatchValue(ref, func(_, newVal bool) {
		ib.active = newVal
		ib.UpdateVisuals()
	})
}

// SetOnClick sets the callback invoked when the button is clicked.
func (ib *IconButton) SetOnClick(fn func()) {
	ib.onClick = fn
}

// SetSize sets the button dimensions.
func (ib *IconButton) SetSize(w, h float64) {
	ib.applySize(w, h)
}

// applySize is the internal setter used by SetSize and the constructor.
func (ib *IconButton) applySize(w, h float64) {
	ib.Width = w
	ib.Height = h
	ib.resizeBackground(w, h)
	ib.resizeBorder(w, h)
	ib.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	ib.layoutChildren()
	ib.MarkLayoutDirty()
}

// UpdateVisuals applies theme colors based on the current interaction state.
func (ib *IconButton) UpdateVisuals() {
	ib.state = computeState(ib.enabled, ib.focused, ib.hovered, ib.pressed)
	th := ib.EffectiveTheme()
	group := th.IconButton.Group(ib.Variant())

	if !ib.embedded {
		cr := resolveCornerRadius(group.CornerRadius, ib.Height)
		ib.applyCornerRadius(cr)

		// Apply active background or normal background.
		if ib.active && group.ActiveColor[0].Type != BgNone {
			ib.applyBackground(group.ActiveColor.Resolve(ib.state))
		} else {
			ib.applyBackground(group.Background.Resolve(ib.state))
		}
		ib.applyBorder(group.BorderColor.Resolve(ib.state), group.BorderWidth, group.Background.Resolve(ib.state))
	}

	// Tint the icon sprite.
	iconColor := group.IconColor.Resolve(ib.state)
	ib.icon.SetColor(iconColor)

	// Label color.
	if ib.label != nil {
		ib.label.SetColor(group.LabelColor.Resolve(ib.state))
	}

	if !ib.embedded {
		ib.applyFocusRing(group.FocusColor.Resolve(ib.state), group.FocusRingWidth)
	}
}

// layoutChildren positions the icon (and optional label) within the button.
func (ib *IconButton) layoutChildren() {
	th := ib.EffectiveTheme()
	group := th.IconButton.Group(ib.Variant())

	iW := ib.iconW
	iH := ib.iconH
	if iW <= 0 {
		iW = group.IconSize
		if iW <= 0 {
			iW = 20
		}
	}
	if iH <= 0 {
		iH = iW
	}
	ib.icon.SetSize(iW, iH)

	gap := group.LabelGap
	if gap <= 0 {
		gap = 4
	}

	if ib.label == nil {
		// Icon only: center in button.
		ib.icon.SetPosition((ib.Width-iW)/2, (ib.Height-iH)/2)
		return
	}

	lW := ib.label.Width
	lH := ib.label.Height

	switch ib.labelPos {
	case IconLabelRight:
		// Icon + label side-by-side: center the pair vertically.
		totalW := iW + gap + lW
		startX := (ib.Width - totalW) / 2
		ib.icon.SetPosition(startX, (ib.Height-iH)/2)
		ib.label.SetPosition(startX+iW+gap, (ib.Height-lH)/2)
	default: // IconLabelBelow
		totalH := iH + gap + lH
		startY := (ib.Height - totalH) / 2
		ib.icon.SetPosition((ib.Width-iW)/2, startY)
		ib.label.SetPosition((ib.Width-lW)/2, startY+iH+gap)
	}
}

// Dispose cleans up the button and its resources.
func (ib *IconButton) Dispose() {
	ib.activeWatch.Stop()
	if ib.label != nil {
		ib.label.Dispose()
	}
	ib.Component.Dispose()
}

// IconNode returns the button's icon willow node.
func (ib *IconButton) IconNode() *sg.Node {
	return ib.icon
}

// IsActive reports whether the button is in the active/toggled-on state.
func (ib *IconButton) IsActive() bool {
	return ib.active
}

// SimulateOnClick invokes the onClick callback directly if one is set.
// Intended for unit tests.
func (ib *IconButton) SimulateOnClick() {
	if ib.onClick != nil {
		ib.onClick()
	}
}

// SimulateOnClick invokes the onClick callback directly if one is set.
// Intended for unit tests.
func (b *Button) SimulateOnClick() {
	if b.onClick != nil {
		b.onClick()
	}
}

// HasOnClickCallback reports whether an onClick callback is registered.
func (b *Button) HasOnClickCallback() bool {
	return b.onClick != nil
}
