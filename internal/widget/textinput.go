package widget

import (
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// passwordDotGlyph returns the password dot glyph from the default spritesheet.
func passwordDotGlyph() engine.Image { return IconPasswordDot() }

// PasswordDotGlyph returns the procedural dot glyph used for password masking.
func PasswordDotGlyph() engine.Image { return passwordDotGlyph() }

// doubleClickThreshold is the maximum time between clicks for a double-click.
const doubleClickThreshold = 400 * time.Millisecond

// TextInput is a single-line text entry field.
type TextInput struct {
	Component
	content  *sg.Node // clipped container for text/cursor/sel
	textNode *sg.Node // text display
	cursor   *sg.Node // WhitePixel sprite: blinking cursor
	selRect  *sg.Node // WhitePixel sprite: selection highlight

	font        *sg.FontFamily
	displaySize float64

	// AutoHeight, when true, causes SetSize to ignore the height argument and
	// instead compute it automatically from the font size and theme padding.
	// This is equivalent to calling SetWidth instead of SetSize.
	AutoHeight  bool
	value       *Ref[string]
	watch       WatchHandle
	placeholder string
	maxLength   int
	cursorPos   int
	selStart    int // selection anchor (rune index)
	selEnd      int // selection moving end (rune index); cursor sits here
	scrollX     float64

	// Cursor blink state.
	blinkCounter int
	blinkVisible bool

	// Double-click detection.
	lastClickTime time.Time

	// Password mode fields.
	passwordMode      bool
	passwordDots      []*sg.Node
	passwordModeWatch WatchHandle

	// embedded suppresses background, border, and focus ring rendering.
	// Set by composite widgets (SearchBox, InputField) that provide their own chrome.
	embedded bool

	// Callbacks.
	onChange   func(string)
	onSubmit   func(string)
	onBlur     func()
	keyFilter  func(engine.Key) bool // optional; returns true to consume a key
	charFilter func(rune) bool       // optional; returns true to accept a character
}

// NewTextInput creates a single-line text input with the given font source and display size.
func NewTextInput(name string, source *sg.FontFamily, displaySize float64) *TextInput {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	ti := &TextInput{
		font:         font,
		displaySize:  displaySize,
		value:        NewRef(""),
		blinkVisible: true,
	}
	initComponent(&ti.Component, name)

	ti.initBackground(name)
	ti.initBorder(name)

	group := ti.EffectiveTheme().TextInput.Group(ti.Variant())
	pad := resolveAutoInsets(group.Padding, defaultTextInputPadding)

	// Clipped content container.
	ti.content = sg.NewContainer(name + "-content")
	ti.content.SetPosition(pad.Left, pad.Top)
	ti.node.AddChild(ti.content)

	// Selection highlight — between background and text.
	ti.selRect = sg.NewSprite(name+"-sel", sg.TextureRegion{})
	ti.selRect.SetVisible(false)
	ti.content.AddChild(ti.selRect)

	// Text display node (child of content container).
	ti.textNode = sg.NewText(name+"-text", "", font)
	ti.textNode.TextBlock.FontSize = displaySize
	ti.content.AddChild(ti.textNode)

	// Cursor (child of content container).
	ti.cursor = sg.NewSprite(name+"-cursor", sg.TextureRegion{})

	ti.cursor.SetScale(1, displayLineHeight(font, displaySize))
	ti.cursor.SetVisible(false)
	ti.content.AddChild(ti.cursor)

	// Default size.
	defaultW := 200.0
	defaultH := displayLineHeight(font, displaySize) + pad.Top + pad.Bottom
	ti.SetSize(defaultW, defaultH)

	// Pointer-down sets the selection anchor at press time (before
	// any drag dead-zone). Shift+click extends from the existing anchor.
	ti.node.OnPointerDown(func(ctx sg.PointerContext) {
		if !ti.enabled {
			return
		}
		ti.pressed = true
		ti.bubbleActivation()
		DefaultFocusManager.SetFocus(&ti.Component)

		now := time.Now()
		if now.Sub(ti.lastClickTime) < doubleClickThreshold {
			ti.selectWordAtCursor()
			ti.lastClickTime = time.Time{} // reset to prevent triple-click
		} else {
			shift := ctx.Modifiers&sg.ModShift != 0
			ti.setCursorFromX(ctx.LocalX, shift)
			ti.lastClickTime = now
		}
		ti.UpdateVisuals()
	})

	// Click fires on press+release without drag — already handled by
	// OnPointerDown above, so nothing extra needed here.

	// Drag-to-select: each drag frame extends selection from the anchor.
	ti.node.OnDrag(func(ctx sg.DragContext) {
		if !ti.enabled {
			return
		}
		ti.setCursorFromX(ctx.LocalX, true)
	})

	ti.onFocusChange = func(focused bool) {
		if !focused {
			ti.clearSelection()
			ti.updateCursorPosition()
			if ti.onBlur != nil {
				ti.onBlur()
			}
		}
		ti.UpdateVisuals()
	}
	ti.onVisualStateChange = func() { ti.UpdateVisuals() }
	ti.onThemeChange = func() { ti.UpdateVisuals() }

	ti.SetCursorShape(engine.CursorShapeText)

	// Focus: text inputs participate in tab and spatial nav, intercept arrows.
	ti.enableFocusNavigation()
	ti.InterceptArrows = true
	ti.ConsumeHandledKeys = false // widget's own Update handles the key; only block spatial nav

	// HandleKey: boundary-aware arrow key query (pure, no side effects).
	ti.SetHandleKey(func(key engine.Key) bool {
		switch key {
		case engine.KeyLeft:
			return ti.cursorPos > 0
		case engine.KeyRight:
			runes := []rune(ti.value.Peek())
			return ti.cursorPos < len(runes)
		case engine.KeyUp, engine.KeyDown:
			return false // single-line: vertical nav always escapes
		}
		return false
	})

	ti.UpdateVisuals()

	// Auto-update: keyboard input via willow's per-frame hook.
	ti.node.OnUpdate = func(_ float64) {
		ti.Update()
	}

	return ti
}

// HasSelection returns true when text is selected.
func (ti *TextInput) HasSelection() bool {
	return ti.selStart != ti.selEnd
}

// SelectedText returns the currently selected text.
func (ti *TextInput) SelectedText() string {
	if !ti.HasSelection() {
		return ""
	}
	runes := []rune(ti.value.Peek())
	lo, hi := ti.selStart, ti.selEnd
	if lo > hi {
		lo, hi = hi, lo
	}
	if hi > len(runes) {
		hi = len(runes)
	}
	return string(runes[lo:hi])
}

// SelectAll selects the entire text content.
func (ti *TextInput) SelectAll() {
	runes := []rune(ti.value.Peek())
	ti.selStart = 0
	ti.selEnd = len(runes)
	ti.cursorPos = len(runes)
	ti.resetBlink()
	ti.updateCursorPosition()
}

// clearSelection collapses selection to the current cursor position.
func (ti *TextInput) clearSelection() {
	ti.selStart = ti.cursorPos
	ti.selEnd = ti.cursorPos
}

// deleteSelection deletes the selected text, moving cursor to the start edge.
func (ti *TextInput) deleteSelection() {
	if !ti.HasSelection() {
		return
	}
	lo, hi := ti.selStart, ti.selEnd
	if lo > hi {
		lo, hi = hi, lo
	}
	runes := []rune(ti.value.Peek())
	if hi > len(runes) {
		hi = len(runes)
	}
	newRunes := append(runes[:lo], runes[hi:]...)
	ti.cursorPos = lo
	ti.selStart = lo
	ti.selEnd = lo
	ti.value.Set(string(newRunes))
	DefaultScheduler.Flush()
}

// Value returns the current text.
func (ti *TextInput) Value() string {
	return ti.value.Peek()
}

// ValueRef returns the reactive Ref[string] backing this input's text value.
func (ti *TextInput) ValueRef() *Ref[string] {
	return ti.value
}

// SetValue sets the text content.
func (ti *TextInput) SetValue(v string) {
	if ti.maxLength > 0 && len([]rune(v)) > ti.maxLength {
		v = string([]rune(v)[:ti.maxLength])
	}
	ti.value.Set(v)
	DefaultScheduler.Flush()
	ti.cursorPos = len([]rune(v))
	ti.clearSelection()
	ti.updateTextDisplay()
}

// SetPlaceholder sets the placeholder text shown when empty.
func (ti *TextInput) SetPlaceholder(p string) {
	ti.placeholder = p
	ti.updateTextDisplay()
}

// SetMaxLength limits the number of characters (0 = no limit).
func (ti *TextInput) SetMaxLength(n int) {
	ti.maxLength = n
}

// SetCharFilter sets a function called for each typed or pasted character.
// Return true to accept the character, false to reject it.
// Passing nil clears any existing filter.
func (ti *TextInput) SetCharFilter(fn func(rune) bool) {
	ti.charFilter = fn
}

// SetNumericOnly restricts input to digit characters (0–9).
func (ti *TextInput) SetNumericOnly() {
	ti.charFilter = func(ch rune) bool { return ch >= '0' && ch <= '9' }
}

// SetAlphanumericOnly restricts input to ASCII letters and digits.
func (ti *TextInput) SetAlphanumericOnly() {
	ti.charFilter = func(ch rune) bool {
		return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
	}
}

// SetAllowedChars restricts input to characters present in the given string.
func (ti *TextInput) SetAllowedChars(chars string) {
	allowed := []rune(chars)
	ti.charFilter = func(ch rune) bool {
		for _, r := range allowed {
			if r == ch {
				return true
			}
		}
		return false
	}
}

// SetOnChange sets the callback for text changes.
func (ti *TextInput) SetOnChange(fn func(string)) {
	ti.onChange = fn
}

// SetOnSubmit sets the callback for enter/submit.
func (ti *TextInput) SetOnSubmit(fn func(string)) {
	ti.onSubmit = fn
}

// SetKeyFilter sets an optional function called before special-key processing.
// If it returns true for a key, TextInput skips its default handling of that key.
// This lets composite widgets (e.g. NumberStepper) intercept keys like Home/End.
func (ti *TextInput) SetKeyFilter(fn func(engine.Key) bool) {
	ti.keyFilter = fn
}

// SetOnBlur sets the callback invoked when the text input loses focus.
func (ti *TextInput) SetOnBlur(fn func()) {
	ti.onBlur = fn
}

// SetPasswordMode enables or disables password masking.
func (ti *TextInput) SetPasswordMode(v bool) {
	if ti.passwordMode == v {
		return
	}
	ti.passwordMode = v
	ti.updateTextDisplay()
}

// IsPasswordMode returns true when password masking is active.
func (ti *TextInput) IsPasswordMode() bool {
	return ti.passwordMode
}

// BindPasswordMode binds password mode to a reactive Ref[bool].
func (ti *TextInput) BindPasswordMode(ref *Ref[bool]) {
	ti.passwordModeWatch.Stop()
	ti.passwordMode = ref.Peek()
	ti.passwordModeWatch = WatchValue(ref, func(_, newVal bool) {
		ti.passwordMode = newVal
		ti.updateTextDisplay()
	})
	ti.updateTextDisplay()
}

// passwordDotPitch returns the center-to-center spacing between dots.
func (ti *TextInput) passwordDotPitch() float64 {
	return ti.displaySize * 0.65
}

// passwordTotalWidth returns the total width occupied by n dots.
func (ti *TextInput) passwordTotalWidth(n int) float64 {
	return float64(n) * ti.passwordDotPitch()
}

// PasswordDots returns the password dot sprite pool. Used for testing.
func (ti *TextInput) PasswordDots() []*sg.Node { return ti.passwordDots }

// SetWidth sets only the width, computing the height automatically from the
// font size and theme padding. This is the preferred way to size a TextInput
// when you don't need a custom height.
func (ti *TextInput) SetWidth(w float64) {
	pad := resolveAutoInsets(ti.EffectiveTheme().TextInput.Group(ti.Variant()).Padding, defaultTextInputPadding)
	h := displayLineHeight(ti.font, ti.displaySize) + pad.Top + pad.Bottom
	ti.SetSize(w, h)
}

// SetSize sets the input dimensions. If AutoHeight is true, the height
// argument is ignored and computed automatically from the font size and padding.
func (ti *TextInput) SetSize(w, h float64) {
	if ti.AutoHeight {
		pad := resolveAutoInsets(ti.EffectiveTheme().TextInput.Group(ti.Variant()).Padding, defaultTextInputPadding)
		h = displayLineHeight(ti.font, ti.displaySize) + pad.Top + pad.Bottom
	}
	ti.Width = w
	ti.Height = h
	ti.resizeBackground(w, h)
	ti.resizeBorder(w, h)
	ti.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	pad := resolveAutoInsets(ti.EffectiveTheme().TextInput.Group(ti.Variant()).Padding, defaultTextInputPadding)
	innerW := w - pad.Left - pad.Right
	innerH := h - pad.Top - pad.Bottom
	// Update mask for clipping.
	ti.updateContentMask(innerW, innerH)
	ti.MarkLayoutDirty()
}

// updateContentMask creates or updates the mask that clips content to the
// inner area of the text input. The mask root must be a container because
// willow ignores the root mask node's own transform — only children's
// transforms are applied.
func (ti *TextInput) updateContentMask(innerW, innerH float64) {
	maskRoot := sg.NewContainer(ti.node.Name + "-mask")
	maskSprite := sg.NewSprite(ti.node.Name+"-mask-rect", sg.TextureRegion{})
	maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	maskSprite.SetScale(innerW, innerH)
	maskRoot.AddChild(maskSprite)
	ti.content.SetMask(maskRoot)
}

// BindValue binds the text input to a reactive Ref[string].
func (ti *TextInput) BindValue(ref *Ref[string]) {
	ti.watch.Stop()
	ti.value = ref
	ti.SetValue(ref.Peek())
	ti.watch = WatchValue(ref, func(_, newVal string) {
		ti.cursorPos = len([]rune(newVal))
		ti.clearSelection()
		ti.updateTextDisplay()
	})
}

// InsertText inserts text at the current cursor position.
// If there is a selection, it replaces the selected text.
// Newlines and carriage returns are stripped since TextInput is single-line.
func (ti *TextInput) InsertText(s string) {
	// Strip newlines/carriage returns — single-line field.
	s = strings.ReplaceAll(s, "\r\n", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")

	// Apply character filter.
	if ti.charFilter != nil {
		filtered := make([]rune, 0, len([]rune(s)))
		for _, ch := range s {
			if ti.charFilter(ch) {
				filtered = append(filtered, ch)
			}
		}
		s = string(filtered)
	}

	if ti.HasSelection() {
		ti.deleteSelection()
	}

	runes := []rune(ti.value.Peek())
	insert := []rune(s)

	if ti.maxLength > 0 && len(runes)+len(insert) > ti.maxLength {
		insert = insert[:ti.maxLength-len(runes)]
	}
	if len(insert) == 0 {
		return
	}

	newRunes := make([]rune, 0, len(runes)+len(insert))
	newRunes = append(newRunes, runes[:ti.cursorPos]...)
	newRunes = append(newRunes, insert...)
	newRunes = append(newRunes, runes[ti.cursorPos:]...)

	ti.cursorPos += len(insert)
	ti.clearSelection()
	ti.value.Set(string(newRunes))
	DefaultScheduler.Flush()
	ti.resetBlink()
	ti.updateTextDisplay()

	if ti.onChange != nil {
		ti.onChange(ti.value.Peek())
	}
}

// DeleteBack deletes the character before the cursor (backspace).
// If there is a selection, it deletes the selected text instead.
func (ti *TextInput) DeleteBack() {
	if ti.HasSelection() {
		ti.deleteSelection()
		ti.resetBlink()
		ti.updateTextDisplay()
		if ti.onChange != nil {
			ti.onChange(ti.value.Peek())
		}
		return
	}
	if ti.cursorPos <= 0 {
		return
	}
	runes := []rune(ti.value.Peek())
	newRunes := append(runes[:ti.cursorPos-1], runes[ti.cursorPos:]...)
	ti.cursorPos--
	ti.clearSelection()
	ti.value.Set(string(newRunes))
	DefaultScheduler.Flush()
	ti.resetBlink()
	ti.updateTextDisplay()

	if ti.onChange != nil {
		ti.onChange(ti.value.Peek())
	}
}

// DeleteForward deletes the character after the cursor (delete key).
// If there is a selection, it deletes the selected text instead.
func (ti *TextInput) DeleteForward() {
	if ti.HasSelection() {
		ti.deleteSelection()
		ti.resetBlink()
		ti.updateTextDisplay()
		if ti.onChange != nil {
			ti.onChange(ti.value.Peek())
		}
		return
	}
	runes := []rune(ti.value.Peek())
	if ti.cursorPos >= len(runes) {
		return
	}
	newRunes := append(runes[:ti.cursorPos], runes[ti.cursorPos+1:]...)
	ti.clearSelection()
	ti.value.Set(string(newRunes))
	DefaultScheduler.Flush()
	ti.resetBlink()
	ti.updateTextDisplay()

	if ti.onChange != nil {
		ti.onChange(ti.value.Peek())
	}
}

// Submit triggers the onSubmit callback.
func (ti *TextInput) Submit() {
	if ti.onSubmit != nil {
		ti.onSubmit(ti.value.Peek())
	}
}

// MoveCursorLeft moves the cursor one position to the left.
// If shift is true, the selection is extended; otherwise it collapses.
func (ti *TextInput) MoveCursorLeft() {
	ti.moveCursorLeftShift(false)
}

func (ti *TextInput) moveCursorLeftShift(shift bool) {
	if !shift && ti.HasSelection() {
		// Collapse to the left edge of the selection.
		lo := ti.selStart
		if ti.selEnd < lo {
			lo = ti.selEnd
		}
		ti.cursorPos = lo
		ti.clearSelection()
		ti.updateCursorPosition()
		return
	}
	if ti.cursorPos > 0 {
		ti.cursorPos--
		if shift {
			ti.selEnd = ti.cursorPos
		} else {
			ti.clearSelection()
		}
		ti.updateCursorPosition()
	}
}

// MoveCursorRight moves the cursor one position to the right.
// If shift is true, the selection is extended; otherwise it collapses.
func (ti *TextInput) MoveCursorRight() {
	ti.moveCursorRightShift(false)
}

func (ti *TextInput) moveCursorRightShift(shift bool) {
	if !shift && ti.HasSelection() {
		// Collapse to the right edge of the selection.
		hi := ti.selStart
		if ti.selEnd > hi {
			hi = ti.selEnd
		}
		ti.cursorPos = hi
		ti.clearSelection()
		ti.updateCursorPosition()
		return
	}
	runes := []rune(ti.value.Peek())
	if ti.cursorPos < len(runes) {
		ti.cursorPos++
		if shift {
			ti.selEnd = ti.cursorPos
		} else {
			ti.clearSelection()
		}
		ti.updateCursorPosition()
	}
}

func (ti *TextInput) moveCursorHomeShift(shift bool) {
	ti.cursorPos = 0
	if shift {
		ti.selEnd = ti.cursorPos
	} else {
		ti.clearSelection()
	}
	ti.resetBlink()
	ti.updateCursorPosition()
}

func (ti *TextInput) moveCursorEndShift(shift bool) {
	ti.cursorPos = len([]rune(ti.value.Peek()))
	if shift {
		ti.selEnd = ti.cursorPos
	} else {
		ti.clearSelection()
	}
	ti.resetBlink()
	ti.updateCursorPosition()
}

// selectWordAtCursor selects the word under the current cursor position.
func (ti *TextInput) selectWordAtCursor() {
	runes := []rune(ti.value.Peek())
	lo, hi := wordBoundaries(runes, ti.cursorPos)
	ti.selStart = lo
	ti.selEnd = hi
	ti.cursorPos = hi
	ti.resetBlink()
	ti.updateCursorPosition()
}

// SetEnabled overrides Component.SetEnabled to also update visuals.
func (ti *TextInput) SetEnabled(v bool) {
	ti.Component.SetEnabled(v)
	ti.UpdateVisuals()
}

// applyBorderColor overrides the border color without changing other visuals.
// Used by InputField to apply validation state colors.
func (ti *TextInput) applyBorderColor(c sg.Color) {
	group := ti.EffectiveTheme().TextInput.Group(ti.Variant())
	bg := group.Background.Resolve(ti.state)
	ti.applyBorder(c, group.BorderWidth, bg)
}

// UpdateVisuals applies theme colors based on current state.
func (ti *TextInput) UpdateVisuals() {
	ti.state = computeState(ti.enabled, ti.focused, ti.hovered, ti.pressed)
	group := ti.EffectiveTheme().TextInput.Group(ti.Variant())

	if !ti.embedded {
		ti.applyCornerRadius(group.CornerRadius)
		bg := group.Background.Resolve(ti.state)
		ti.applyBackground(bg)
		ti.applyBorder(group.Border.Resolve(ti.state), group.BorderWidth, bg)
		ti.applyFocusRing(group.FocusColor.Resolve(ti.state), group.FocusRingWidth)
	}

	ti.cursor.SetColor(group.CursorColor.Resolve(ti.state))
	ti.selRect.SetColor(group.SelectionColor.Resolve(ti.state))
	ti.cursor.SetVisible(ti.focused && ti.blinkVisible)

	if !ti.embedded {
		// Keep content position in sync with the current theme's padding so
		// click-to-cursor math (which reads padding from EffectiveTheme) matches
		// where the text is actually rendered.
		pad := resolveAutoInsets(group.Padding, defaultTextInputPadding)
		if ti.content.X() != pad.Left || ti.content.Y() != pad.Top {
			ti.content.SetPosition(pad.Left, pad.Top)
		}
	}

	ti.updateTextDisplay()
}

// Update handles keyboard input, cursor blink, and visual state.
func (ti *TextInput) Update() {
	if !ti.focused {
		ti.UpdateVisuals()
		return
	}

	im := DefaultInputManager
	shift := engine.IsKeyPressed(engine.KeyShift)
	ctrl := engine.IsKeyPressed(engine.KeyControl) || engine.IsKeyPressed(engine.KeyMeta)

	// Ctrl+A: select all.
	if ctrl && im.IsKeyJustAvailable(engine.KeyA) {
		im.Consume(engine.KeyA)
		ti.SelectAll()
		return
	}

	// Ctrl+C: copy selection to clipboard (disabled in password mode).
	if ctrl && im.IsKeyJustAvailable(engine.KeyC) {
		im.Consume(engine.KeyC)
		if !ti.passwordMode {
			if sel := ti.SelectedText(); sel != "" {
				clipboardWrite(sel)
			}
		}
		return
	}

	// Ctrl+X: cut selection to clipboard (disabled in password mode).
	if ctrl && im.IsKeyJustAvailable(engine.KeyX) {
		im.Consume(engine.KeyX)
		if !ti.passwordMode {
			if sel := ti.SelectedText(); sel != "" {
				clipboardWrite(sel)
				ti.deleteSelection()
				ti.resetBlink()
				ti.updateTextDisplay()
				if ti.onChange != nil {
					ti.onChange(ti.value.Peek())
				}
			}
		}
		return
	}

	// Ctrl+V: paste from clipboard.
	if ctrl && im.IsKeyJustAvailable(engine.KeyV) {
		im.Consume(engine.KeyV)
		if text, err := clipboardRead(); err == nil && text != "" {
			ti.InsertText(text)
		}
		return
	}

	// Cmd/Ctrl+Left: move to start of line (same as Home).
	if ctrl && im.IsKeyJustAvailable(engine.KeyLeft) {
		im.Consume(engine.KeyLeft)
		ti.moveCursorHomeShift(shift)
		return
	}
	// Cmd/Ctrl+Right: move to end of line (same as End).
	if ctrl && im.IsKeyJustAvailable(engine.KeyRight) {
		im.Consume(engine.KeyRight)
		ti.moveCursorEndShift(shift)
		return
	}

	// Read typed characters (skip if ctrl is held to avoid control chars).
	if !ctrl {
		chars := engine.AppendInputChars(nil)
		if scene := currentScene(); scene != nil {
			chars = scene.AppendInjectedChars(chars)
		}
		if len(chars) > 0 {
			ti.InsertText(string(chars))
		}
	}

	// Handle special keys.
	if im.IsKeyJustAvailable(engine.KeyEscape) {
		im.Consume(engine.KeyEscape)
		DefaultFocusManager.ClearFocus()
		ti.UpdateVisuals()
		return
	}
	if im.IsKeyJustAvailable(engine.KeyBackspace) {
		im.Consume(engine.KeyBackspace)
		ti.DeleteBack()
	}
	if im.IsKeyJustAvailable(engine.KeyDelete) {
		im.Consume(engine.KeyDelete)
		ti.DeleteForward()
	}
	if im.IsKeyJustAvailable(engine.KeyLeft) {
		im.Consume(engine.KeyLeft)
		ti.moveCursorLeftShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyRight) {
		im.Consume(engine.KeyRight)
		ti.moveCursorRightShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyHome) && (ti.keyFilter == nil || !ti.keyFilter(engine.KeyHome)) {
		im.Consume(engine.KeyHome)
		ti.moveCursorHomeShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyEnd) && (ti.keyFilter == nil || !ti.keyFilter(engine.KeyEnd)) {
		im.Consume(engine.KeyEnd)
		ti.moveCursorEndShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyEnter) {
		im.Consume(engine.KeyEnter)
		ti.Submit()
	}

	// Cursor blink.
	ti.blinkCounter++
	if ti.blinkCounter >= 30 { // ~0.5s at 60fps
		ti.blinkCounter = 0
		ti.blinkVisible = !ti.blinkVisible
		ti.cursor.SetVisible(ti.blinkVisible)
	}

	ti.UpdateVisuals()
}

// Dispose stops reactive watches and disposes the component tree.
func (ti *TextInput) Dispose() {
	ti.watch.Stop()
	ti.passwordModeWatch.Stop()
	ti.Component.Dispose()
}

// resetBlink makes the cursor visible and resets the blink timer,
// so the cursor is always shown immediately after a text change.
func (ti *TextInput) resetBlink() {
	ti.blinkCounter = 0
	ti.blinkVisible = true
	ti.cursor.SetVisible(ti.focused)
}

// setCursorFromX positions the cursor at the character closest to the
// given local x coordinate within the text input. If shift is true,
// the selection is extended from the current anchor.
func (ti *TextInput) setCursorFromX(localX float64, shift bool) {
	if ti.font == nil && !ti.passwordMode {
		return
	}
	pad := resolveAutoInsets(ti.EffectiveTheme().TextInput.Group(ti.Variant()).Padding, defaultTextInputPadding)
	// Convert screen x to text-space x (account for scroll offset).
	x := localX - pad.Left + ti.scrollX
	runes := []rune(ti.value.Peek())

	var best int
	if ti.passwordMode {
		pitch := ti.passwordDotPitch()
		if pitch > 0 {
			best = int(math.Round(x / pitch))
		}
		if best < 0 {
			best = 0
		}
		if best > len(runes) {
			best = len(runes)
		}
	} else {
		// Walk characters, find the position where the click falls.
		for i := 1; i <= len(runes); i++ {
			w, _ := measureDisplay(ti.font, string(runes[:i]), ti.displaySize)
			if w <= x {
				best = i
			} else {
				prevW, _ := measureDisplay(ti.font, string(runes[:i-1]), ti.displaySize)
				if x-prevW > w-x {
					best = i
				}
				break
			}
		}
	}

	if shift {
		// Extend selection from anchor.
		ti.selEnd = best
		ti.cursorPos = best
	} else {
		ti.cursorPos = best
		ti.clearSelection()
	}
	ti.resetBlink()
	ti.updateCursorPosition()
}

// updateTextDisplay syncs the text node with the current value.
func (ti *TextInput) updateTextDisplay() {
	v := ti.value.Peek()
	group := ti.EffectiveTheme().TextInput.Group(ti.Variant())
	textColor := group.TextColor.Resolve(ti.state)

	if ti.passwordMode && v != "" {
		// Hide text, show dots.
		ti.textNode.SetVisible(false)
		runeCount := len([]rune(v))
		ti.syncPasswordDots(runeCount, group)
	} else {
		// Normal mode or empty value (show placeholder).
		ti.textNode.SetVisible(true)
		ti.hideAllPasswordDots()
		if v == "" && ti.placeholder != "" {
			ti.textNode.SetContent(ti.placeholder)
			ti.textNode.SetTextColor(sg.RGBA(textColor.R(), textColor.G(), textColor.B(), group.PlaceholderAlpha))
		} else {
			ti.textNode.SetContent(v)
			ti.textNode.SetTextColor(textColor)
		}
	}
	ti.updateCursorPosition()
	ti.MarkDrawDirty()
}

// syncPasswordDots ensures the correct number of dot sprites are visible and positioned.
func (ti *TextInput) syncPasswordDots(count int, group *TextInputGroup) {
	dotDiameter := ti.displaySize * 0.45
	dotScale := GlyphScale(passwordDotGlyph(), dotDiameter)
	lineHeight := displayLineHeight(ti.font, ti.displaySize)
	dotY := (lineHeight - dotDiameter) / 2
	pitch := ti.passwordDotPitch()

	// Resolve dot color: PasswordDotColor if set, else TextColor.
	dotColor := group.PasswordDotColor.Resolve(ti.state)
	if dotColor == (sg.Color{}) {
		dotColor = group.TextColor.Resolve(ti.state)
	}

	// Grow pool if needed.
	for len(ti.passwordDots) < count {
		dot := sg.NewSprite(ti.node.Name+"-pwdot", sg.TextureRegion{})
		dot.SetCustomImage(passwordDotGlyph())
		ti.content.AddChild(dot)
		ti.passwordDots = append(ti.passwordDots, dot)
	}

	// Position visible dots, hide extras.
	for i, dot := range ti.passwordDots {
		if i < count {
			dot.SetVisible(true)
			dot.SetPosition(float64(i)*pitch, dotY)
			dot.SetScale(dotScale, dotScale)
			dot.SetColor(dotColor)
		} else {
			dot.SetVisible(false)
		}
	}
}

// hideAllPasswordDots hides all password dot sprites.
func (ti *TextInput) hideAllPasswordDots() {
	for _, dot := range ti.passwordDots {
		dot.SetVisible(false)
	}
}

// updateCursorPosition places the cursor at the correct x offset,
// adjusts scrollX to keep the cursor visible, and updates the
// selection highlight rectangle.
func (ti *TextInput) updateCursorPosition() {
	if ti.font == nil && !ti.passwordMode {
		return
	}
	var cursorX float64
	if ti.passwordMode {
		cursorX = float64(ti.cursorPos) * ti.passwordDotPitch()
	} else {
		textBefore := string([]rune(ti.value.Peek())[:ti.cursorPos])
		cursorX, _ = measureDisplay(ti.font, textBefore, ti.displaySize)
	}

	// Ensure cursor is visible within the inner area.
	ti.ensureCursorVisible(cursorX)

	// Position text and cursor relative to the content container,
	// offset by scrollX.
	ti.textNode.SetX(-ti.scrollX)
	ti.cursor.SetPosition(cursorX-ti.scrollX, 0)
	ti.updateSelectionRect()
}

// ensureCursorVisible adjusts scrollX so the cursor pixel position
// is within the visible inner width and the text doesn't leave a gap.
func (ti *TextInput) ensureCursorVisible(cursorX float64) {
	pad := resolveAutoInsets(ti.EffectiveTheme().TextInput.Group(ti.Variant()).Padding, defaultTextInputPadding)
	innerW := ti.Width - pad.Left - pad.Right

	// Clamp scrollX so the text end doesn't leave a gap on the right.
	// maxScroll is the furthest we should scroll: total text width minus
	// the visible area, but never negative.
	var totalW float64
	if ti.passwordMode {
		totalW = ti.passwordTotalWidth(len([]rune(ti.value.Peek())))
	} else {
		totalW, _ = measureDisplay(ti.font, ti.value.Peek(), ti.displaySize)
	}
	maxScroll := totalW - innerW
	if maxScroll < 0 {
		maxScroll = 0
	}
	if ti.scrollX > maxScroll {
		ti.scrollX = maxScroll
	}

	// Then ensure the cursor itself is in view.
	if cursorX-ti.scrollX > innerW {
		ti.scrollX = cursorX - innerW
	}
	if cursorX-ti.scrollX < 0 {
		ti.scrollX = cursorX
	}
	if ti.scrollX < 0 {
		ti.scrollX = 0
	}
}

// updateSelectionRect sizes and positions the selection highlight.
func (ti *TextInput) updateSelectionRect() {
	if !ti.HasSelection() || (ti.font == nil && !ti.passwordMode) {
		ti.selRect.SetVisible(false)
		return
	}
	runes := []rune(ti.value.Peek())
	lo, hi := ti.selStart, ti.selEnd
	if lo > hi {
		lo, hi = hi, lo
	}
	if hi > len(runes) {
		hi = len(runes)
	}

	var xStart, xEnd float64
	if ti.passwordMode {
		pitch := ti.passwordDotPitch()
		xStart = float64(lo) * pitch
		xEnd = float64(hi) * pitch
	} else {
		xStart, _ = measureDisplay(ti.font, string(runes[:lo]), ti.displaySize)
		xEnd, _ = measureDisplay(ti.font, string(runes[:hi]), ti.displaySize)
	}

	// Position relative to content container, offset by scrollX.
	ti.selRect.SetPosition(xStart-ti.scrollX, 0)
	ti.selRect.SetScale(xEnd-xStart, displayLineHeight(ti.font, ti.displaySize))
	ti.selRect.SetVisible(true)
}

// TextNode returns the willow text node used for displaying input content.
// Used for testing text input internals.
func (ti *TextInput) TextNode() *sg.Node { return ti.textNode }

// CursorNode returns the cursor sprite node. Used for testing.
func (ti *TextInput) CursorNode() *sg.Node { return ti.cursor }

// Placeholder returns the current placeholder string. Used for testing.
func (ti *TextInput) GetPlaceholder() string { return ti.placeholder }

// GetCursorPos returns the current cursor rune position. Used for testing.
func (ti *TextInput) GetCursorPos() int { return ti.cursorPos }

// SetCursorPos sets the cursor rune position directly. Used for testing.
func (ti *TextInput) SetCursorPos(pos int) { ti.cursorPos = pos }

// GetSelStart returns the selection start rune index. Used for testing.
func (ti *TextInput) GetSelStart() int { return ti.selStart }

// SetSelStart sets the selection start directly. Used for testing.
func (ti *TextInput) SetSelStart(v int) { ti.selStart = v }

// GetSelEnd returns the selection end rune index. Used for testing.
func (ti *TextInput) GetSelEnd() int { return ti.selEnd }

// SetSelEnd sets the selection end directly. Used for testing.
func (ti *TextInput) SetSelEnd(v int) { ti.selEnd = v }

// GetScrollX returns the current horizontal scroll offset. Used for testing.
func (ti *TextInput) GetScrollX() float64 { return ti.scrollX }

// SelRectVisible reports whether the selection rectangle is visible. Used for testing.
func (ti *TextInput) SelRectVisible() bool {
	if ti.selRect == nil {
		return false
	}
	return ti.selRect.Visible()
}

// SelRectNode returns the selection rectangle node, or nil if not created. Used for testing.
func (ti *TextInput) SelRectNode() *sg.Node { return ti.selRect }

// ClearSelectionForTest calls the internal clearSelection method. Used for testing.
func (ti *TextInput) ClearSelectionForTest() { ti.clearSelection() }

// DeleteSelectionForTest calls the internal deleteSelection method. Used for testing.
func (ti *TextInput) DeleteSelectionForTest() { ti.deleteSelection() }

// UpdateSelectionRectForTest calls the internal updateSelectionRect method. Used for testing.
func (ti *TextInput) UpdateSelectionRectForTest() { ti.updateSelectionRect() }

// MoveCursorLeftShiftForTest calls moveCursorLeftShift. Used for testing.
func (ti *TextInput) MoveCursorLeftShiftForTest(shift bool) { ti.moveCursorLeftShift(shift) }

// MoveCursorRightShiftForTest calls moveCursorRightShift. Used for testing.
func (ti *TextInput) MoveCursorRightShiftForTest(shift bool) { ti.moveCursorRightShift(shift) }

// MoveCursorHomeShiftForTest calls moveCursorHomeShift. Used for testing.
func (ti *TextInput) MoveCursorHomeShiftForTest(shift bool) { ti.moveCursorHomeShift(shift) }

// MoveCursorEndShiftForTest calls moveCursorEndShift. Used for testing.
func (ti *TextInput) MoveCursorEndShiftForTest(shift bool) { ti.moveCursorEndShift(shift) }

// SelectWordAtCursorForTest calls selectWordAtCursor. Used for testing.
func (ti *TextInput) SelectWordAtCursorForTest() { ti.selectWordAtCursor() }

// wordBoundaries returns the start (inclusive) and end (exclusive) rune
// indices of the word surrounding pos. A "word" is a contiguous run of
// letters/digits; everything else is a contiguous run of non-word chars.
// WordBoundaries is the exported equivalent of wordBoundaries, for use by
// the root package tests.
func WordBoundaries(runes []rune, pos int) (lo, hi int) { return wordBoundaries(runes, pos) }

func wordBoundaries(runes []rune, pos int) (lo, hi int) {
	if len(runes) == 0 {
		return 0, 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}
	// Pick the character to classify. If pos is at the end or on a boundary,
	// use the character before pos.
	idx := pos
	if idx >= len(runes) {
		idx = len(runes) - 1
	}
	isWord := unicode.IsLetter(runes[idx]) || unicode.IsDigit(runes[idx])

	// Scan left.
	lo = idx
	for lo > 0 {
		r := runes[lo-1]
		if (unicode.IsLetter(r) || unicode.IsDigit(r)) != isWord {
			break
		}
		lo--
	}

	// Scan right.
	hi = idx
	for hi < len(runes) {
		if (unicode.IsLetter(runes[hi]) || unicode.IsDigit(runes[hi])) != isWord {
			break
		}
		hi++
	}
	return lo, hi
}
