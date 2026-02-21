package widget

import (
	"fmt"
	"strings"

	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// KeyBinding represents a keyboard or gamepad binding.
type KeyBinding struct {
	Key           engine.Key
	GamepadButton engine.GamepadButton
	IsGamepad     bool
	IsUnset       bool // true when no binding is assigned

	// Modifier keys for combo bindings.
	Ctrl  bool
	Shift bool
	Alt   bool
}

// DisplayName returns a human-readable string for the binding (e.g. "SPACE", "Ctrl+A", "GP:0").
func (b KeyBinding) DisplayName() string {
	if b.IsUnset {
		return ""
	}
	if b.IsGamepad {
		return fmt.Sprintf("GP:%d", int(b.GamepadButton))
	}
	var parts []string
	if b.Ctrl {
		parts = append(parts, "Ctrl")
	}
	if b.Shift {
		parts = append(parts, "Shift")
	}
	if b.Alt {
		parts = append(parts, "Alt")
	}
	parts = append(parts, keyDisplayName(b.Key))
	return strings.Join(parts, "+")
}

// KeybindInput is a settings control that captures a keyboard or gamepad
// binding. It displays the current binding as a styled key cap label and
// enters listening mode on click to capture a new binding.
type KeybindInput struct {
	Component

	font        *sg.FontFamily
	displaySize float64

	// Visual nodes.
	bg       *sg.Node // background sprite
	capBg    *sg.Node // key cap background
	capLabel *Label   // displays current binding text
	clearBtn *Label   // "x" clear button

	// State.
	binding   KeyBinding
	listening bool

	combosEnabled bool // when true, modifier+key combos are supported

	// Callbacks.
	onBindingChanged func(binding KeyBinding)
	onConflict       func(existing KeyBinding) bool

	// Reactive binding.
	bindingRef   *Ref[KeyBinding]
	bindingWatch WatchHandle
}

// NewKeybindInput creates a KeybindInput with the given name, font source, and display size.
func NewKeybindInput(name string, source *sg.FontFamily, displaySize float64) *KeybindInput {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	k := &KeybindInput{
		font:          font,
		displaySize:   displaySize,
		binding:       KeyBinding{IsUnset: true},
		combosEnabled: true,
	}
	initComponent(&k.Component, name)

	// Background.
	k.bg = sg.NewSprite(name+"-bg", sg.TextureRegion{})
	k.node.AddChild(k.bg)

	// Key cap background.
	k.capBg = sg.NewSprite(name+"-cap-bg", sg.TextureRegion{})
	k.node.AddChild(k.capBg)

	// Key cap label.
	k.capLabel = NewLabel(name+"-cap-label", "---", source, displaySize)
	k.capLabel.AddToNode(k.node)

	// Clear button.
	k.clearBtn = NewLabel(name+"-clear", "x", source, displaySize)
	k.clearBtn.AddToNode(k.node)
	k.clearBtn.Node().OnClick(func(ctx sg.ClickContext) {
		if !k.enabled {
			return
		}
		k.ClearBinding()
	})
	k.clearBtn.SetCursorShape(engine.CursorShapePointer)

	k.Width = 200
	k.Height = 36
	k.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: k.Width, Height: k.Height}

	// Click enters listening mode (unless click is on the clear button area).
	k.node.OnClick(func(ctx sg.ClickContext) {
		if !k.enabled {
			return
		}
		// If the clear button is visible and the click is in its area, clear instead.
		if k.clearBtn.Node().Visible() {
			cbX := k.clearBtn.Node().X()
			cbW := k.clearBtn.Width
			if ctx.LocalX >= cbX && ctx.LocalX <= cbX+cbW {
				k.ClearBinding()
				return
			}
		}
		k.SetListening(true)
	})

	// OnUpdate: capture input when listening.
	k.node.OnUpdate = func(_ float64) {
		if !k.listening {
			return
		}

		// Escape cancels listening.
		if core.IsKeyJustPressed(engine.KeyEscape) {
			k.SetListening(false)
			return
		}

		// Check for gamepad button presses on all connected gamepads.
		gpIDs := engine.AppendGamepadIDs(nil)
		for _, gpID := range gpIDs {
			for btn := engine.GamepadButton(0); btn <= engine.GamepadButton(engine.GamepadButtonMax); btn++ {
				if engine.IsGamepadButtonJustPressed(gpID, btn) {
					newBinding := KeyBinding{
						GamepadButton: btn,
						IsGamepad:     true,
					}
					k.tryAcceptBinding(newBinding)
					return
				}
			}
		}

		// Check for keyboard key presses.
		pressed := engine.AppendJustPressedKeys(nil)
		for _, key := range pressed {
			// Skip modifier keys themselves when combos are enabled —
			// they will be captured as modifiers on the next non-modifier key.
			if k.combosEnabled && isModifierKey(key) {
				continue
			}

			newBinding := KeyBinding{Key: key}
			if k.combosEnabled {
				newBinding.Ctrl = engine.IsKeyPressed(engine.KeyControl) || engine.IsKeyPressed(engine.KeyControlLeft) || engine.IsKeyPressed(engine.KeyControlRight)
				newBinding.Shift = engine.IsKeyPressed(engine.KeyShift) || engine.IsKeyPressed(engine.KeyShiftLeft) || engine.IsKeyPressed(engine.KeyShiftRight)
				newBinding.Alt = engine.IsKeyPressed(engine.KeyAlt) || engine.IsKeyPressed(engine.KeyAltLeft) || engine.IsKeyPressed(engine.KeyAltRight)
			}
			k.tryAcceptBinding(newBinding)
			return
		}
	}

	k.onVisualStateChange = func() { k.UpdateVisuals() }
	k.onThemeChange = func() { k.UpdateVisuals() }
	k.SetCursorShape(engine.CursorShapePointer)

	// Focus.
	k.enableFocusNavigation()
	k.onFocusChange = func(focused bool) {
		if !focused && k.listening {
			k.SetListening(false)
		}
		k.UpdateVisuals()
	}

	k.UpdateVisuals()
	return k
}

// tryAcceptBinding attempts to accept a new binding, checking the conflict callback.
func (k *KeybindInput) tryAcceptBinding(newBinding KeyBinding) {
	if k.onConflict != nil {
		if !k.onConflict(newBinding) {
			// Conflict rejected — stay in listening mode.
			return
		}
	}
	k.binding = newBinding
	k.listening = false
	k.UpdateVisuals()
	if k.bindingRef != nil {
		k.bindingRef.Set(k.binding)
	}
	if k.onBindingChanged != nil {
		k.onBindingChanged(k.binding)
	}
}

// Binding returns the current key binding.
func (k *KeybindInput) Binding() KeyBinding {
	return k.binding
}

// SetBinding sets the binding programmatically without entering listening mode.
func (k *KeybindInput) SetBinding(binding KeyBinding) {
	k.binding = binding
	k.UpdateVisuals()
}

// ClearBinding unsets the current binding.
func (k *KeybindInput) ClearBinding() {
	k.binding = KeyBinding{IsUnset: true}
	k.listening = false
	k.UpdateVisuals()
	if k.bindingRef != nil {
		k.bindingRef.Set(k.binding)
	}
	if k.onBindingChanged != nil {
		k.onBindingChanged(k.binding)
	}
}

// SetListening enters or exits listening mode.
func (k *KeybindInput) SetListening(v bool) {
	k.listening = v
	k.UpdateVisuals()
}

// IsListening returns true if the widget is in listening mode.
func (k *KeybindInput) IsListening() bool {
	return k.listening
}

// SetCombosEnabled sets whether modifier+key combos are supported.
// Enabled by default. Disable for games that only need single-key bindings.
func (k *KeybindInput) SetCombosEnabled(v bool) {
	k.combosEnabled = v
}

// CombosEnabled returns whether modifier+key combos are supported.
func (k *KeybindInput) CombosEnabled() bool {
	return k.combosEnabled
}

// SetOnBindingChanged sets the callback invoked after a new binding is captured.
func (k *KeybindInput) SetOnBindingChanged(fn func(binding KeyBinding)) {
	k.onBindingChanged = fn
}

// SetOnConflict sets the conflict check callback. Return false to reject a binding.
func (k *KeybindInput) SetOnConflict(fn func(existing KeyBinding) bool) {
	k.onConflict = fn
}

// BindValue binds the current key binding to a reactive Ref. Changes to the
// Ref update the widget, and user captures update the Ref.
func (k *KeybindInput) BindValue(ref *Ref[KeyBinding]) {
	k.bindingRef = ref
	bindRef(&k.bindingWatch, ref, k.SetBinding)
}

// SetSize sets the widget dimensions and repositions children.
func (k *KeybindInput) SetSize(w, h float64) {
	k.Width = w
	k.Height = h
	k.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	k.UpdateVisuals()
}

// SetEnabled overrides Component.SetEnabled to update visuals.
func (k *KeybindInput) SetEnabled(v bool) {
	k.Component.SetEnabled(v)
	if !v && k.listening {
		k.listening = false
	}
	k.UpdateVisuals()
}

// defaultPadding returns the given insets or falls back to the provided defaults
// when all fields are zero.
func defaultPadding(p render.Insets, fallback render.Insets) render.Insets {
	if p.Left == 0 && p.Right == 0 && p.Top == 0 && p.Bottom == 0 {
		return fallback
	}
	return p
}

// Hardcoded fallback colors for when no theme is set.
var (
	kbDefaultBg          = sg.RGBA(0.15, 0.17, 0.20, 1)
	kbDefaultBgHover     = sg.RGBA(0.18, 0.20, 0.24, 1)
	kbDefaultBgDisabled  = sg.RGBA(0.12, 0.13, 0.15, 1)
	kbDefaultCapBg       = sg.RGBA(0.25, 0.28, 0.33, 1)
	kbDefaultCapBgHover  = sg.RGBA(0.30, 0.33, 0.38, 1)
	kbDefaultCapText     = sg.RGBA(0.95, 0.95, 0.95, 1)
	kbDefaultCapTextDis  = sg.RGBA(0.45, 0.45, 0.50, 1)
	kbDefaultListeningBg = sg.RGBA(0.12, 0.18, 0.28, 1)
	kbDefaultListenText  = sg.RGBA(0.5, 0.7, 1.0, 1)
	kbDefaultUnsetText   = sg.RGBA(0.4, 0.4, 0.45, 1)
	kbDefaultClearColor  = sg.RGBA(0.45, 0.45, 0.50, 1)
	kbDefaultClearHover  = sg.RGBA(0.8, 0.3, 0.3, 1)
)

// colorOr returns c if it is non-zero, otherwise returns the fallback.
func colorOr(c, fallback sg.Color) sg.Color {
	if c == (sg.Color{}) {
		return fallback
	}
	return c
}

// UpdateVisuals applies theme colors and repositions child nodes.
func (k *KeybindInput) UpdateVisuals() {
	k.state = computeState(k.enabled, k.focused, k.hovered, k.pressed)
	group := k.EffectiveTheme().KeybindInput.Group(k.Variant())

	padding := defaultPadding(group.Padding, render.Insets{Left: 8, Right: 8, Top: 4, Bottom: 4})

	// Background.
	bgFallback := kbDefaultBg
	if !k.enabled {
		bgFallback = kbDefaultBgDisabled
	} else if k.hovered {
		bgFallback = kbDefaultBgHover
	}
	if k.listening {
		bgFallback = kbDefaultListeningBg
		bg := group.ListeningBackground.Resolve(k.state)
		k.bg.SetColor(colorOr(bg.Color, bgFallback))
	} else {
		bg := group.Background.Resolve(k.state)
		k.bg.SetColor(colorOr(bg.Color, bgFallback))
	}
	k.bg.SetScale(k.Width, k.Height)

	// Clear button.
	clearFallback := kbDefaultClearColor
	if k.hovered {
		clearFallback = kbDefaultClearHover
	}
	k.clearBtn.SetColor(colorOr(group.ClearButtonColor.Resolve(k.state), clearFallback))
	clearBtnW := k.clearBtn.Width
	clearBtnH := k.clearBtn.Height
	clearX := k.Width - padding.Right - clearBtnW
	clearY := (k.Height - clearBtnH) / 2
	k.clearBtn.SetPosition(clearX, clearY)
	k.clearBtn.Node().SetVisible(k.enabled && !k.listening && !k.binding.IsUnset)

	// Key cap area: between left padding and clear button.
	capAreaX := padding.Left
	capAreaW := clearX - capAreaX - 4
	if k.listening || k.binding.IsUnset {
		capAreaW = k.Width - padding.Left - padding.Right
	}

	if k.listening {
		// Listening mode: show prompt text.
		listeningText := group.ListeningText
		if listeningText == "" {
			listeningText = "Press any key..."
		}
		k.capLabel.SetText(listeningText)
		k.capLabel.SetColor(colorOr(group.ListeningTextColor.Resolve(k.state), kbDefaultListenText))
		k.capBg.SetVisible(false)
	} else if k.binding.IsUnset {
		// Unset: show placeholder.
		unsetText := group.UnsetText
		if unsetText == "" {
			unsetText = "---"
		}
		k.capLabel.SetText(unsetText)
		k.capLabel.SetColor(colorOr(group.UnsetTextColor.Resolve(k.state), kbDefaultUnsetText))
		k.capBg.SetVisible(false)
	} else {
		// Normal: show key cap.
		k.capLabel.SetText(k.binding.DisplayName())
		capTextFallback := kbDefaultCapText
		if !k.enabled {
			capTextFallback = kbDefaultCapTextDis
		}
		k.capLabel.SetColor(colorOr(group.KeyCapTextColor.Resolve(k.state), capTextFallback))

		// Key cap background.
		capPad := defaultPadding(group.KeyCapPadding, render.Insets{Left: 8, Right: 8, Top: 2, Bottom: 2})
		capW := k.capLabel.Width + capPad.Left + capPad.Right
		capH := k.capLabel.Height + capPad.Top + capPad.Bottom
		capX := capAreaX
		capY := (k.Height - capH) / 2
		k.capBg.SetPosition(capX, capY)
		k.capBg.SetScale(capW, capH)
		capBgFallback := kbDefaultCapBg
		if !k.enabled {
			capBgFallback = kbDefaultBgDisabled
		} else if k.hovered {
			capBgFallback = kbDefaultCapBgHover
		}
		capBg := group.KeyCapBackground.Resolve(k.state)
		k.capBg.SetColor(colorOr(capBg.Color, capBgFallback))
		k.capBg.SetVisible(true)

		// Position label inside key cap.
		k.capLabel.SetPosition(capX+capPad.Left, capY+capPad.Top)
		k.MarkDrawDirty()
		return
	}

	// Center label in cap area for listening/unset states.
	labelX := capAreaX + (capAreaW-k.capLabel.Width)/2
	labelY := (k.Height - k.capLabel.Height) / 2
	k.capLabel.SetPosition(labelX, labelY)

	k.MarkDrawDirty()
}

// Dispose cleans up the widget.
func (k *KeybindInput) Dispose() {
	k.bindingWatch.Stop()
	if k.capLabel != nil {
		k.capLabel.Dispose()
	}
	if k.clearBtn != nil {
		k.clearBtn.Dispose()
	}
	k.Component.Dispose()
}

// isModifierKey returns true if the key is a modifier (Ctrl, Shift, Alt, Meta).
func isModifierKey(key engine.Key) bool {
	switch key {
	case engine.KeyControl, engine.KeyControlLeft, engine.KeyControlRight,
		engine.KeyShift, engine.KeyShiftLeft, engine.KeyShiftRight,
		engine.KeyAlt, engine.KeyAltLeft, engine.KeyAltRight,
		engine.KeyMeta, engine.KeyMetaLeft, engine.KeyMetaRight:
		return true
	}
	return false
}

// keyDisplayName returns a human-readable name for an ebiten key.
func keyDisplayName(key engine.Key) string {
	name := key.String()
	if name == "" {
		return fmt.Sprintf("Key%d", int(key))
	}
	return strings.ToUpper(name)
}
