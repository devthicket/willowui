package widget

import (
	"strings"
	"time"
	"unicode"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// ---------------------------------------------------------------------------
// Mask token types
// ---------------------------------------------------------------------------

// maskTokenKind identifies the kind of a mask position.
type maskTokenKind int

const (
	maskLiteral     maskTokenKind = iota // fixed character in the formatted string
	maskDigit                            // '9' — digit 0-9
	maskLetter                           // 'a' — letter A-Z or a-z
	maskUpperLetter                      // 'A' — uppercase letter; lowercase auto-uppercased
	maskAny                              // '*' — any visible ASCII character
	maskUpperAlnum                       // 'X' — uppercase alphanumeric; lowercase auto-uppercased
)

// maskToken is one position in a parsed mask.
type maskToken struct {
	Kind    maskTokenKind
	Literal rune // only when Kind == maskLiteral
}

// parseMask parses a mask string into a slice of maskTokens.
func parseMask(mask string) []maskToken {
	tokens := make([]maskToken, 0, len(mask))
	for _, ch := range mask {
		switch ch {
		case '9':
			tokens = append(tokens, maskToken{Kind: maskDigit})
		case 'a':
			tokens = append(tokens, maskToken{Kind: maskLetter})
		case 'A':
			tokens = append(tokens, maskToken{Kind: maskUpperLetter})
		case '*':
			tokens = append(tokens, maskToken{Kind: maskAny})
		case 'X':
			tokens = append(tokens, maskToken{Kind: maskUpperAlnum})
		default:
			tokens = append(tokens, maskToken{Kind: maskLiteral, Literal: ch})
		}
	}
	return tokens
}

// validateSlotChar normalizes and validates a character for a given slot kind.
// Returns the (possibly normalized) character and whether it is valid.
func validateSlotChar(kind maskTokenKind, ch rune) (rune, bool) {
	switch kind {
	case maskDigit:
		return ch, ch >= '0' && ch <= '9'
	case maskLetter:
		return ch, unicode.IsLetter(ch)
	case maskUpperLetter:
		if unicode.IsLetter(ch) {
			return unicode.ToUpper(ch), true
		}
		return ch, false
	case maskAny:
		return ch, ch > ' ' && ch <= '~'
	case maskUpperAlnum:
		if unicode.IsLetter(ch) {
			return unicode.ToUpper(ch), true
		}
		return ch, ch >= '0' && ch <= '9'
	}
	return ch, false
}

// ---------------------------------------------------------------------------
// maskCell — one rendered glyph position
// ---------------------------------------------------------------------------

// maskCell represents a single character position in the slot-cell display.
// Slot cells are interactive; literal cells are static.
type maskCell struct {
	node    *sg.Node // text node for the glyph
	dbgRect *sg.Node // debug highlight sprite; nil for literals
	slotIdx int      // index into mi.slots[]; -1 for literals
	x       float64  // left edge of the full cell (including slot padding)
	w       float64  // total cell width (including slot padding on both sides)
}

// maxWidthChar returns a representative "widest" character string for a slot
// kind, used to compute a stable cell width before any character is typed.
func maxWidthChar(kind maskTokenKind) string {
	switch kind {
	case maskDigit:
		return "0"
	case maskLetter, maskUpperLetter, maskUpperAlnum, maskAny:
		return "W"
	}
	return "0"
}

// ---------------------------------------------------------------------------
// MaskedInput widget
// ---------------------------------------------------------------------------

// MaskedInput is a single-line text entry field constrained by a mask pattern.
// The mask describes editable slot positions and literal separator characters.
// Supported slot characters: '9' (digit), 'a' (letter), 'A' (upper letter),
// 'X' (upper alphanumeric), '*' (any visible ASCII). All other characters
// are treated as literals.
//
// Each mask position is rendered as its own glyph node. Clicking a filled
// slot positions the cursor there; clicking an empty slot or any non-slot
// area snaps the cursor to the first empty slot. Typing a valid character
// fills the current slot and automatically advances to the next slot.
type MaskedInput struct {
	Component
	content         *sg.Node // clipped container
	placeholderNode *sg.Node // field-level placeholder (no-mask case)
	cursor          *sg.Node // blinking caret
	selRect         *sg.Node // selection highlight (multi-slot range)
	activeSlotRect  *sg.Node // single-slot highlight on the current cursor slot

	// Slot-cell rendering state. Rebuilt whenever SetMask is called.
	cells           []maskCell
	tokenIdxForSlot []int // tokenIdxForSlot[slotIdx] → index in cells[]

	font        *sg.FontFamily
	displaySize float64

	// AutoHeight, when true, causes SetSize to ignore the height argument
	// and instead compute it automatically from the font size and theme padding.
	AutoHeight bool

	// Mask state.
	mask   string
	tokens []maskToken
	slots  []rune // one entry per editable slot; 0 = empty

	// Display options.
	maskPlaceholder rune   // char to show for empty slots (0 = blank)
	placeholder     string // field-level placeholder shown when no mask is set

	// Extra length cap (0 = no limit beyond mask capacity).
	maxLength int

	// Reactive bindings (external refs set by BindValue / BindRawValue).
	valueRef   *Ref[string]
	rawRef     *Ref[string]
	valueWatch WatchHandle
	rawWatch   WatchHandle

	// Cursor / selection in raw-slot space (0 .. len(slots)).
	cursorPos int
	selStart  int
	selEnd    int

	blinkCounter     int
	blinkVisible     bool
	lastClickTime    time.Time
	focusFromPointer bool // true while OnPointerDown is calling SetFocus
	debugSlots       bool // when true, each slot cell shows a colored border

	// Completion tracking.
	wasComplete bool

	// Callbacks.
	onChange     func(string)
	onRawChange  func(string)
	onSubmit     func(string)
	onBlur       func()
	onComplete   func(raw, formatted string)
	onIncomplete func(raw, formatted string)
}

// NewMaskedInput creates a single-line masked input with the given font source and display size.
func NewMaskedInput(name string, source *sg.FontFamily, displaySize float64) *MaskedInput {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	mi := &MaskedInput{
		font:         font,
		displaySize:  displaySize,
		blinkVisible: true,
	}
	initComponent(&mi.Component, name)

	mi.initBackground(name)
	mi.initBorder(name)

	group := mi.EffectiveTheme().MaskedInput.Group(mi.Variant())
	pad := resolveAutoInsets(group.Padding, defaultTextInputPadding)

	// Clipped content container.
	mi.content = sg.NewContainer(name + "-content")
	mi.content.SetPosition(pad.Left, pad.Top)
	mi.node.AddChild(mi.content)

	// Selection highlight — rendered below glyphs.
	mi.selRect = sg.NewSprite(name+"-sel", sg.TextureRegion{})
	mi.selRect.SetVisible(false)
	mi.content.AddChild(mi.selRect)

	// Active-slot highlight — shows which slot will receive the next keystroke.
	mi.activeSlotRect = sg.NewSprite(name+"-active-slot", sg.TextureRegion{})
	mi.activeSlotRect.SetVisible(false)
	mi.content.AddChild(mi.activeSlotRect)

	// Field-level placeholder text (shown when no mask is set or field is empty
	// without a maskPlaceholder).
	mi.placeholderNode = sg.NewText(name+"-ph", "", font)
	mi.placeholderNode.TextBlock.FontSize = displaySize
	mi.content.AddChild(mi.placeholderNode)

	// Cursor — vertical blinking bar.
	mi.cursor = sg.NewSprite(name+"-cursor", sg.TextureRegion{})
	mi.cursor.SetScale(1, displayLineHeight(font, displaySize))
	mi.cursor.SetVisible(false)
	mi.content.AddChild(mi.cursor)

	// Default size.
	defaultW := 200.0
	defaultH := displayLineHeight(font, displaySize) + pad.Top + pad.Bottom
	mi.SetSize(defaultW, defaultH)

	// Pointer-down: focus the field, position cursor at the clicked cell.
	// Double-click selects all.
	mi.node.OnPointerDown(func(ctx sg.PointerContext) {
		if !mi.enabled {
			return
		}
		mi.pressed = true
		mi.bubbleActivation()
		mi.focusFromPointer = true
		DefaultFocusManager.SetFocus(&mi.Component)
		mi.focusFromPointer = false

		now := time.Now()
		if now.Sub(mi.lastClickTime) < doubleClickThreshold {
			mi.SelectAll()
			mi.lastClickTime = time.Time{}
		} else {
			shift := ctx.Modifiers&sg.ModShift != 0
			mi.setCursorFromCell(ctx.LocalX, shift)
			mi.lastClickTime = now
		}
		mi.UpdateVisuals()
	})

	// Drag-to-select: extend selection as pointer moves.
	mi.node.OnDrag(func(ctx sg.DragContext) {
		if !mi.enabled {
			return
		}
		mi.setCursorFromCell(ctx.LocalX, true)
	})

	mi.onFocusChange = func(focused bool) {
		if focused {
			if !mi.focusFromPointer {
				// Keyboard focus (Tab): snap to first empty slot so typing starts
				// at the right position immediately.
				mi.snapCursorToFirstEmpty()
			}
		} else {
			mi.clearSelection()
			mi.updateCursorPosition()
			if mi.onBlur != nil {
				mi.onBlur()
			}
		}
		mi.UpdateVisuals()
	}
	mi.onVisualStateChange = func() { mi.UpdateVisuals() }
	mi.onThemeChange = func() { mi.UpdateVisuals() }

	mi.SetCursorShape(engine.CursorShapeText)

	// Focus navigation — same flags as TextInput.
	mi.enableFocusNavigation()
	mi.InterceptArrows = true
	mi.ConsumeHandledKeys = false

	// HandleKey: boundary-aware arrow key query.
	mi.SetHandleKey(func(key engine.Key) bool {
		switch key {
		case engine.KeyLeft:
			return mi.cursorPos > 0
		case engine.KeyRight:
			return mi.cursorPos < len(mi.slots)
		case engine.KeyUp, engine.KeyDown:
			return false
		}
		return false
	})

	mi.UpdateVisuals()

	// Auto-update hook.
	mi.node.OnUpdate = func(_ float64) {
		mi.Update()
	}

	return mi
}

// ---------------------------------------------------------------------------
// Mask API
// ---------------------------------------------------------------------------

// SetMask parses and applies a new mask string, resetting all slot values.
// Mask characters: '9'=digit, 'a'=letter, 'A'=upper letter, 'X'=upper alphanumeric,
// '*'=any visible ASCII. All other characters become literals.
func (mi *MaskedInput) SetMask(mask string) {
	mi.mask = mask
	mi.tokens = parseMask(mask)
	cap := mi.countSlots()
	mi.slots = make([]rune, cap)
	mi.cursorPos = 0
	mi.clearSelection()
	mi.wasComplete = false
	mi.rebuildCells()
	mi.updateTextDisplay()
}

// Mask returns the current mask pattern string.
func (mi *MaskedInput) Mask() string { return mi.mask }

// countSlots returns the number of editable slots in the current token list.
func (mi *MaskedInput) countSlots() int {
	n := 0
	for _, tok := range mi.tokens {
		if tok.Kind != maskLiteral {
			n++
		}
	}
	return n
}

// slotKindAt returns the maskTokenKind for raw slot index rawIdx.
func (mi *MaskedInput) slotKindAt(rawIdx int) maskTokenKind {
	count := 0
	for _, tok := range mi.tokens {
		if tok.Kind != maskLiteral {
			if count == rawIdx {
				return tok.Kind
			}
			count++
		}
	}
	return maskDigit // unreachable for valid rawIdx
}

// RawToDisplayIndex converts a raw slot index (0..capacity) to a display
// character offset. rawIdx=0 is before any char (offset 0);
// rawIdx=n is after the n-th slot character in the display string.
// This is exported for testing.
func (mi *MaskedInput) RawToDisplayIndex(rawIdx int) int {
	return mi.rawToDisplayIndex(rawIdx)
}

// rawToDisplayIndex is the internal implementation.
func (mi *MaskedInput) rawToDisplayIndex(rawIdx int) int {
	if len(mi.tokens) == 0 || rawIdx == 0 {
		return 0
	}
	count := 0
	for i, tok := range mi.tokens {
		if tok.Kind != maskLiteral {
			count++
			if count == rawIdx {
				return i + 1
			}
		}
	}
	return len(mi.tokens)
}

// ---------------------------------------------------------------------------
// Value API
// ---------------------------------------------------------------------------

// Value returns the formatted display string (filled slots + literals).
func (mi *MaskedInput) Value() string {
	return mi.buildDisplayString()
}

// SetValue fills slots from a formatted string. Characters that match slot
// types are accepted; literal separator characters in the input are ignored.
func (mi *MaskedInput) SetValue(v string) {
	mi.clearAllSlots()
	if len(mi.tokens) == 0 {
		mi.cursorPos = 0
		mi.clearSelection()
		mi.updateTextDisplay()
		return
	}
	slotIdx := 0
	for _, ch := range v {
		if slotIdx >= len(mi.slots) {
			break
		}
		kind := mi.slotKindAt(slotIdx)
		if normalized, ok := validateSlotChar(kind, ch); ok {
			mi.slots[slotIdx] = normalized
			slotIdx++
		}
		// Skip literal characters and invalid chars.
	}
	mi.cursorPos = slotIdx
	mi.clearSelection()
	mi.updateTextDisplay()
	mi.fireValueCallbacks()
	mi.checkCompleteTransition()
}

// RawValue returns the filled slot characters without any literal separators.
func (mi *MaskedInput) RawValue() string {
	var b strings.Builder
	for _, ch := range mi.slots {
		if ch != 0 {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// SetRawValue fills slots sequentially from a raw string (no literals expected).
func (mi *MaskedInput) SetRawValue(v string) {
	mi.clearAllSlots()
	slotIdx := 0
	for _, ch := range v {
		if slotIdx >= len(mi.slots) {
			break
		}
		kind := mi.slotKindAt(slotIdx)
		if normalized, ok := validateSlotChar(kind, ch); ok {
			mi.slots[slotIdx] = normalized
			slotIdx++
		}
	}
	mi.cursorPos = slotIdx
	mi.clearSelection()
	mi.updateTextDisplay()
	mi.fireValueCallbacks()
	mi.checkCompleteTransition()
}

// SetPlaceholder sets the field-level placeholder text shown when the mask
// is not set or the field is completely empty and mask placeholders are disabled.
func (mi *MaskedInput) SetPlaceholder(p string) {
	mi.placeholder = p
	mi.updateTextDisplay()
}

// Placeholder returns the current field-level placeholder string.
func (mi *MaskedInput) Placeholder() string { return mi.placeholder }

// SetMaskPlaceholder sets the character shown for unfilled editable slots.
// Use 0 (default) to show blank space for unfilled slots.
// Note: changing this after SetMask is called rebuilds the cell layout since
// cell widths are derived from the placeholder character.
func (mi *MaskedInput) SetMaskPlaceholder(ch rune) {
	mi.maskPlaceholder = ch
	if len(mi.tokens) > 0 {
		mi.rebuildCells()
	}
	mi.updateTextDisplay()
}

// MaskPlaceholder returns the current mask placeholder character.
func (mi *MaskedInput) MaskPlaceholder() rune { return mi.maskPlaceholder }

// SetMaxLength sets an optional character cap (0 = no limit beyond mask capacity).
func (mi *MaskedInput) SetMaxLength(n int) { mi.maxLength = n }

// IsComplete reports whether all editable slots are filled.
func (mi *MaskedInput) IsComplete() bool {
	if len(mi.slots) == 0 {
		return false
	}
	for _, ch := range mi.slots {
		if ch == 0 {
			return false
		}
	}
	return true
}

// IsEmpty reports whether no editable slot has a value.
func (mi *MaskedInput) IsEmpty() bool {
	for _, ch := range mi.slots {
		if ch != 0 {
			return false
		}
	}
	return true
}

// Clear clears all slot values and resets the cursor to the start.
func (mi *MaskedInput) Clear() {
	mi.clearAllSlots()
	mi.cursorPos = 0
	mi.clearSelection()
	mi.updateTextDisplay()
	mi.fireValueCallbacks()
	mi.checkCompleteTransition()
}

// ---------------------------------------------------------------------------
// Selection API
// ---------------------------------------------------------------------------

// SelectAll selects all editable slots.
func (mi *MaskedInput) SelectAll() {
	mi.selStart = 0
	mi.selEnd = len(mi.slots)
	mi.cursorPos = len(mi.slots)
	mi.resetBlink()
	mi.updateCursorPosition()
}

// HasSelection reports whether any slots are selected.
func (mi *MaskedInput) HasSelection() bool { return mi.selStart != mi.selEnd }

// SelectedText returns the display text of the currently selected range,
// including any literal separator characters within the range.
func (mi *MaskedInput) SelectedText() string {
	if !mi.HasSelection() {
		return ""
	}
	lo, hi := mi.selStart, mi.selEnd
	if lo > hi {
		lo, hi = hi, lo
	}
	displayStr := mi.buildDisplayString()
	dispRunes := []rune(displayStr)
	loIdx := mi.rawToDisplayIndex(lo)
	hiIdx := mi.rawToDisplayIndex(hi)
	if loIdx < 0 {
		loIdx = 0
	}
	if hiIdx > len(dispRunes) {
		hiIdx = len(dispRunes)
	}
	return string(dispRunes[loIdx:hiIdx])
}

// ---------------------------------------------------------------------------
// Callbacks
// ---------------------------------------------------------------------------

// SetOnChange sets the callback invoked with the formatted value on any change.
func (mi *MaskedInput) SetOnChange(fn func(string)) { mi.onChange = fn }

// SetOnRawChange sets the callback invoked with the raw value on any change.
func (mi *MaskedInput) SetOnRawChange(fn func(string)) { mi.onRawChange = fn }

// SetOnSubmit sets the callback invoked when Enter is pressed.
func (mi *MaskedInput) SetOnSubmit(fn func(string)) { mi.onSubmit = fn }

// SetOnBlur sets the callback invoked when the field loses focus.
func (mi *MaskedInput) SetOnBlur(fn func()) { mi.onBlur = fn }

// SetOnComplete sets the callback invoked when all slots become filled.
func (mi *MaskedInput) SetOnComplete(fn func(raw, formatted string)) { mi.onComplete = fn }

// SetOnIncomplete sets the callback invoked when the field transitions from
// complete to incomplete.
func (mi *MaskedInput) SetOnIncomplete(fn func(raw, formatted string)) { mi.onIncomplete = fn }

// Submit fires the OnSubmit callback with the current formatted value.
func (mi *MaskedInput) Submit() {
	if mi.onSubmit != nil {
		mi.onSubmit(mi.Value())
	}
}

// ---------------------------------------------------------------------------
// Reactive bindings
// ---------------------------------------------------------------------------

// BindValue binds the formatted value to an external Ref[string].
// The ref is updated whenever the field value changes, and the field is
// updated whenever the ref changes externally.
func (mi *MaskedInput) BindValue(ref *Ref[string]) {
	mi.valueWatch.Stop()
	mi.valueRef = ref
	mi.SetValue(ref.Peek())
	mi.valueWatch = WatchValue(ref, func(_, newVal string) {
		if newVal != mi.Value() {
			mi.SetValue(newVal)
		}
	})
}

// BindRawValue binds the raw value (no literals) to an external Ref[string].
func (mi *MaskedInput) BindRawValue(ref *Ref[string]) {
	mi.rawWatch.Stop()
	mi.rawRef = ref
	mi.SetRawValue(ref.Peek())
	mi.rawWatch = WatchValue(ref, func(_, newVal string) {
		if newVal != mi.RawValue() {
			mi.SetRawValue(newVal)
		}
	})
}

// ---------------------------------------------------------------------------
// Sizing
// ---------------------------------------------------------------------------

// SetWidth sets only the width, computing height automatically from font and padding.
func (mi *MaskedInput) SetWidth(w float64) {
	pad := resolveAutoInsets(mi.EffectiveTheme().MaskedInput.Group(mi.Variant()).Padding, defaultTextInputPadding)
	h := displayLineHeight(mi.font, mi.displaySize) + pad.Top + pad.Bottom
	mi.SetSize(w, h)
}

// SetSize sets the widget dimensions.
func (mi *MaskedInput) SetSize(w, h float64) {
	if mi.AutoHeight {
		pad := resolveAutoInsets(mi.EffectiveTheme().MaskedInput.Group(mi.Variant()).Padding, defaultTextInputPadding)
		h = displayLineHeight(mi.font, mi.displaySize) + pad.Top + pad.Bottom
	}
	mi.Width = w
	mi.Height = h
	mi.resizeBackground(w, h)
	mi.resizeBorder(w, h)
	mi.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	pad := resolveAutoInsets(mi.EffectiveTheme().MaskedInput.Group(mi.Variant()).Padding, defaultTextInputPadding)
	innerW := w - pad.Left - pad.Right
	innerH := h - pad.Top - pad.Bottom
	mi.updateContentMask(innerW, innerH)
	mi.MarkLayoutDirty()
}

// updateContentMask creates or updates the clipping mask for the content area.
func (mi *MaskedInput) updateContentMask(innerW, innerH float64) {
	maskRoot := sg.NewContainer(mi.node.Name + "-mask")
	maskSprite := sg.NewSprite(mi.node.Name+"-mask-rect", sg.TextureRegion{})
	maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	maskSprite.SetScale(innerW, innerH)
	maskRoot.AddChild(maskSprite)
	mi.content.SetMask(maskRoot)
}

// SetEnabled overrides Component.SetEnabled to also update visuals.
func (mi *MaskedInput) SetEnabled(v bool) {
	mi.Component.SetEnabled(v)
	mi.UpdateVisuals()
}

// Dispose stops reactive watches and disposes the component tree.
func (mi *MaskedInput) Dispose() {
	mi.valueWatch.Stop()
	mi.rawWatch.Stop()
	mi.Component.Dispose()
}

// ---------------------------------------------------------------------------
// Update and visuals
// ---------------------------------------------------------------------------

// UpdateVisuals applies theme colors based on current state.
func (mi *MaskedInput) UpdateVisuals() {
	mi.state = computeState(mi.enabled, mi.focused, mi.hovered, mi.pressed)
	group := mi.EffectiveTheme().MaskedInput.Group(mi.Variant())

	cr := resolveCornerRadius(group.CornerRadius, mi.Height)
	mi.applyCornerRadius(cr)

	bg := group.Background.Resolve(mi.state)
	mi.applyBackground(bg)
	mi.applyBorder(group.Border.Resolve(mi.state), group.BorderWidth, bg)
	mi.cursor.SetColor(group.CursorColor.Resolve(mi.state))
	mi.selRect.SetColor(group.SelectionColor.Resolve(mi.state))
	mi.activeSlotRect.SetColor(group.SelectionColor.Resolve(mi.state))
	mi.cursor.SetVisible(mi.focused && mi.blinkVisible)

	pad := resolveAutoInsets(group.Padding, defaultTextInputPadding)
	if mi.content.X() != pad.Left || mi.content.Y() != pad.Top {
		mi.content.SetPosition(pad.Left, pad.Top)
	}
	mi.applyFocusRing(group.FocusColor.Resolve(mi.state), group.FocusRingWidth)
	mi.updateTextDisplay()
}

// Update handles keyboard input, cursor blink, and visual state.
func (mi *MaskedInput) Update() {
	if !mi.focused {
		mi.UpdateVisuals()
		return
	}

	im := DefaultInputManager
	shift := engine.IsKeyPressed(engine.KeyShift)
	ctrl := engine.IsKeyPressed(engine.KeyControl) || engine.IsKeyPressed(engine.KeyMeta)

	// Ctrl+A: select all.
	if ctrl && im.IsKeyJustAvailable(engine.KeyA) {
		im.Consume(engine.KeyA)
		mi.SelectAll()
		return
	}

	// Ctrl+C: copy selection.
	if ctrl && im.IsKeyJustAvailable(engine.KeyC) {
		im.Consume(engine.KeyC)
		if sel := mi.SelectedText(); sel != "" {
			clipboardWrite(sel)
		}
		return
	}

	// Ctrl+X: cut selection.
	if ctrl && im.IsKeyJustAvailable(engine.KeyX) {
		im.Consume(engine.KeyX)
		if mi.HasSelection() {
			clipboardWrite(mi.SelectedText())
			lo, hi := mi.selStart, mi.selEnd
			if lo > hi {
				lo, hi = hi, lo
			}
			mi.clearSlotRange(lo, hi)
			mi.cursorPos = lo
			mi.clearSelection()
			mi.resetBlink()
			mi.updateTextDisplay()
			mi.fireValueCallbacks()
			mi.checkCompleteTransition()
		}
		return
	}

	// Ctrl+V: paste.
	if ctrl && im.IsKeyJustAvailable(engine.KeyV) {
		im.Consume(engine.KeyV)
		if text, err := clipboardRead(); err == nil && text != "" {
			mi.InsertText(text)
		}
		return
	}

	// Ctrl+Left / Ctrl+Right: jump to start/end.
	if ctrl && im.IsKeyJustAvailable(engine.KeyLeft) {
		im.Consume(engine.KeyLeft)
		mi.moveCursorHomeShift(shift)
		return
	}
	if ctrl && im.IsKeyJustAvailable(engine.KeyRight) {
		im.Consume(engine.KeyRight)
		mi.moveCursorEndShift(shift)
		return
	}

	// Typed characters.
	if !ctrl {
		chars := engine.AppendInputChars(nil)
		if scene := currentScene(); scene != nil {
			chars = scene.AppendInjectedChars(chars)
		}
		for _, ch := range chars {
			mi.typeChar(ch)
		}
	}

	// Special keys.
	if im.IsKeyJustAvailable(engine.KeyEscape) {
		im.Consume(engine.KeyEscape)
		DefaultFocusManager.ClearFocus()
		mi.UpdateVisuals()
		return
	}
	if im.IsKeyJustAvailable(engine.KeyBackspace) {
		im.Consume(engine.KeyBackspace)
		mi.DeleteBack()
	}
	if im.IsKeyJustAvailable(engine.KeyDelete) {
		im.Consume(engine.KeyDelete)
		mi.DeleteForward()
	}
	if im.IsKeyJustAvailable(engine.KeyLeft) {
		im.Consume(engine.KeyLeft)
		mi.moveCursorLeftShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyRight) {
		im.Consume(engine.KeyRight)
		mi.moveCursorRightShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyHome) {
		im.Consume(engine.KeyHome)
		mi.moveCursorHomeShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyEnd) {
		im.Consume(engine.KeyEnd)
		mi.moveCursorEndShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyEnter) {
		im.Consume(engine.KeyEnter)
		mi.Submit()
	}

	// Cursor blink.
	mi.blinkCounter++
	if mi.blinkCounter >= 30 {
		mi.blinkCounter = 0
		mi.blinkVisible = !mi.blinkVisible
		mi.cursor.SetVisible(mi.blinkVisible)
	}

	mi.UpdateVisuals()
}

// ---------------------------------------------------------------------------
// Text editing internals
// ---------------------------------------------------------------------------

// typeChar handles a single typed character, inserting it into the current slot.
func (mi *MaskedInput) typeChar(ch rune) {
	cap := len(mi.slots)
	if cap == 0 {
		return
	}
	// If there's a selection, clear it first.
	if mi.HasSelection() {
		lo, hi := mi.selStart, mi.selEnd
		if lo > hi {
			lo, hi = hi, lo
		}
		mi.clearSlotRange(lo, hi)
		mi.cursorPos = lo
		mi.clearSelection()
	}
	slotIdx := mi.cursorPos
	if slotIdx >= cap {
		return
	}
	// Respect optional maxLength cap.
	if mi.maxLength > 0 && slotIdx >= mi.maxLength {
		return
	}
	kind := mi.slotKindAt(slotIdx)
	normalized, ok := validateSlotChar(kind, ch)
	if !ok {
		return
	}
	mi.slots[slotIdx] = normalized
	mi.cursorPos = slotIdx + 1 // auto-advance to next slot
	mi.clearSelection()
	mi.resetBlink()
	mi.updateTextDisplay()
	mi.fireValueCallbacks()
	mi.checkCompleteTransition()
}

// InsertText distributes pasted characters into slots starting at cursorPos.
// Invalid characters are skipped; literal separators in the pasted string that
// match the mask are also skipped.
func (mi *MaskedInput) InsertText(s string) {
	if len(mi.tokens) == 0 {
		return
	}
	// Clear any existing selection.
	if mi.HasSelection() {
		lo, hi := mi.selStart, mi.selEnd
		if lo > hi {
			lo, hi = hi, lo
		}
		mi.clearSlotRange(lo, hi)
		mi.cursorPos = lo
		mi.clearSelection()
	}
	slotIdx := mi.cursorPos
	for _, ch := range s {
		if slotIdx >= len(mi.slots) {
			break
		}
		if mi.maxLength > 0 && slotIdx >= mi.maxLength {
			break
		}
		kind := mi.slotKindAt(slotIdx)
		if normalized, ok := validateSlotChar(kind, ch); ok {
			mi.slots[slotIdx] = normalized
			slotIdx++
		}
		// Invalid and literal-matching chars are silently skipped.
	}
	mi.cursorPos = slotIdx
	mi.clearSelection()
	mi.resetBlink()
	mi.updateTextDisplay()
	mi.fireValueCallbacks()
	mi.checkCompleteTransition()
}

// DeleteBack handles backspace: clears the previous slot and moves cursor back.
func (mi *MaskedInput) DeleteBack() {
	if mi.HasSelection() {
		lo, hi := mi.selStart, mi.selEnd
		if lo > hi {
			lo, hi = hi, lo
		}
		mi.clearSlotRange(lo, hi)
		mi.cursorPos = lo
		mi.clearSelection()
		mi.resetBlink()
		mi.updateTextDisplay()
		mi.fireValueCallbacks()
		mi.checkCompleteTransition()
		return
	}
	if mi.cursorPos <= 0 {
		return
	}
	mi.cursorPos--
	mi.slots[mi.cursorPos] = 0
	mi.clearSelection()
	mi.resetBlink()
	mi.updateTextDisplay()
	mi.fireValueCallbacks()
	mi.checkCompleteTransition()
}

// DeleteForward handles delete: clears the current slot without moving cursor.
func (mi *MaskedInput) DeleteForward() {
	if mi.HasSelection() {
		lo, hi := mi.selStart, mi.selEnd
		if lo > hi {
			lo, hi = hi, lo
		}
		mi.clearSlotRange(lo, hi)
		mi.cursorPos = lo
		mi.clearSelection()
		mi.resetBlink()
		mi.updateTextDisplay()
		mi.fireValueCallbacks()
		mi.checkCompleteTransition()
		return
	}
	if mi.cursorPos >= len(mi.slots) {
		return
	}
	mi.slots[mi.cursorPos] = 0
	mi.clearSelection()
	mi.resetBlink()
	mi.updateTextDisplay()
	mi.fireValueCallbacks()
	mi.checkCompleteTransition()
}

// ---------------------------------------------------------------------------
// Cursor movement
// ---------------------------------------------------------------------------

// MoveCursorLeft moves the cursor one slot to the left.
func (mi *MaskedInput) MoveCursorLeft() { mi.moveCursorLeftShift(false) }

func (mi *MaskedInput) moveCursorLeftShift(shift bool) {
	if !shift && mi.HasSelection() {
		lo := mi.selStart
		if mi.selEnd < lo {
			lo = mi.selEnd
		}
		mi.cursorPos = lo
		mi.clearSelection()
		mi.updateCursorPosition()
		return
	}
	if mi.cursorPos > 0 {
		mi.cursorPos--
		if shift {
			mi.selEnd = mi.cursorPos
		} else {
			mi.clearSelection()
		}
		mi.updateCursorPosition()
	}
}

// MoveCursorRight moves the cursor one slot to the right.
func (mi *MaskedInput) MoveCursorRight() { mi.moveCursorRightShift(false) }

func (mi *MaskedInput) moveCursorRightShift(shift bool) {
	if !shift && mi.HasSelection() {
		hi := mi.selStart
		if mi.selEnd > hi {
			hi = mi.selEnd
		}
		mi.cursorPos = hi
		mi.clearSelection()
		mi.updateCursorPosition()
		return
	}
	if mi.cursorPos < len(mi.slots) {
		mi.cursorPos++
		if shift {
			mi.selEnd = mi.cursorPos
		} else {
			mi.clearSelection()
		}
		mi.updateCursorPosition()
	}
}

func (mi *MaskedInput) moveCursorHomeShift(shift bool) {
	mi.cursorPos = 0
	if shift {
		mi.selEnd = mi.cursorPos
	} else {
		mi.clearSelection()
	}
	mi.resetBlink()
	mi.updateCursorPosition()
}

func (mi *MaskedInput) moveCursorEndShift(shift bool) {
	mi.cursorPos = len(mi.slots)
	if shift {
		mi.selEnd = mi.cursorPos
	} else {
		mi.clearSelection()
	}
	mi.resetBlink()
	mi.updateCursorPosition()
}

// ---------------------------------------------------------------------------
// Slot-cell construction and layout
// ---------------------------------------------------------------------------

// rebuildCells destroys and recreates all per-token glyph nodes.
// Called whenever the mask, maskPlaceholder, or theme changes.
func (mi *MaskedInput) rebuildCells() {
	// Remove old cell nodes from the content container.
	for _, c := range mi.cells {
		mi.content.RemoveChild(c.node)
		if c.dbgRect != nil {
			mi.content.RemoveChild(c.dbgRect)
		}
	}
	mi.cells = nil
	mi.tokenIdxForSlot = nil

	if len(mi.tokens) == 0 {
		return
	}

	group := mi.EffectiveTheme().MaskedInput.Group(mi.Variant())
	slotPad := group.SlotPadding
	if slotPad < 0 {
		slotPad = 0
	}

	lineH := displayLineHeight(mi.font, mi.displaySize)
	mi.cells = make([]maskCell, len(mi.tokens))
	mi.tokenIdxForSlot = make([]int, 0, len(mi.slots))

	x := 0.0
	slotCount := 0
	for i, tok := range mi.tokens {
		var initChar string
		var glyphW float64 // width of the character glyph itself

		if tok.Kind == maskLiteral {
			initChar = string(tok.Literal)
			glyphW, _ = measureDisplay(mi.font, initChar, mi.displaySize)
		} else {
			// Width reference: maskPlaceholder char if set, otherwise the
			// widest representative character for this slot type.
			var widthRef string
			if mi.maskPlaceholder != 0 {
				widthRef = string(mi.maskPlaceholder)
			} else {
				widthRef = maxWidthChar(tok.Kind)
			}
			glyphW, _ = measureDisplay(mi.font, widthRef, mi.displaySize)
			mi.tokenIdxForSlot = append(mi.tokenIdxForSlot, i)
		}

		isSlot := tok.Kind != maskLiteral
		cellW := glyphW
		glyphOffX := x
		if isSlot {
			cellW = glyphW + 2*slotPad
			glyphOffX = x + slotPad
		}

		// Debug highlight rect (slot cells only).
		var dbg *sg.Node
		if isSlot {
			dbg = sg.NewSprite(mi.node.Name+"-cell-dbg", sg.TextureRegion{})
			dbg.SetScale(cellW, lineH)
			dbg.SetPosition(x, 0)
			dbg.SetColor(sg.RGBA(1, 0.5, 0, 0.25))
			dbg.SetVisible(mi.debugSlots)
			mi.content.AddChild(dbg)
		}

		n := sg.NewText(mi.node.Name+"-cell", initChar, mi.font)
		n.TextBlock.FontSize = mi.displaySize
		n.SetPosition(glyphOffX, 0)
		mi.content.AddChild(n)

		si := -1
		if isSlot {
			n.Interactable = true
			n.HitShape = sg.HitRect{X: 0, Y: 0, Width: glyphW, Height: lineH}
			si = slotCount
			slotCount++
		}

		mi.cells[i] = maskCell{node: n, dbgRect: dbg, slotIdx: si, x: x, w: cellW}
		x += cellW
	}

	// Z-order: selRect and activeSlotRect below glyphs, cursor on top.
	mi.cursor.SetZIndex(2)
	mi.selRect.SetZIndex(-2)
	mi.activeSlotRect.SetZIndex(-1)
}

// SetDebugSlots toggles a debug overlay that outlines each slot cell with a
// colored highlight, making the cell boundaries and padding visible.
func (mi *MaskedInput) SetDebugSlots(v bool) {
	mi.debugSlots = v
	for _, c := range mi.cells {
		if c.dbgRect != nil {
			c.dbgRect.SetVisible(v)
		}
	}
	mi.MarkDrawDirty()
}

// cellX returns the x position of the blinking cursor for a given slot index.
// The cursor sits at the left glyph edge (inside slot padding) so it appears
// adjacent to the character, not at the outer cell boundary.
// slotIdx == len(slots) places the cursor after the last glyph.
func (mi *MaskedInput) cellX(slotIdx int) float64 {
	if len(mi.tokenIdxForSlot) == 0 {
		return 0
	}
	group := mi.EffectiveTheme().MaskedInput.Group(mi.Variant())
	pad := group.SlotPadding
	if pad < 0 {
		pad = 0
	}
	if slotIdx >= len(mi.tokenIdxForSlot) {
		// After last slot: right glyph edge = cell.x + cell.w - pad.
		last := mi.tokenIdxForSlot[len(mi.tokenIdxForSlot)-1]
		c := mi.cells[last]
		return c.x + c.w - pad
	}
	// Before slot slotIdx: left glyph edge = cell.x + pad.
	return mi.cells[mi.tokenIdxForSlot[slotIdx]].x + pad
}

// setCursorFromCell positions the cursor based on a click at a local x
// coordinate (in the field's own coordinate space, including padding).
//
//   - Clicking on a filled slot → cursor goes to that slot.
//   - Clicking on an empty slot → snap to the first empty slot.
//   - Clicking between/past all slot cells → snap to first empty slot.
func (mi *MaskedInput) setCursorFromCell(localX float64, shift bool) {
	if len(mi.cells) == 0 {
		return
	}
	group := mi.EffectiveTheme().MaskedInput.Group(mi.Variant())
	pad := resolveAutoInsets(group.Padding, defaultTextInputPadding)
	x := localX - pad.Left

	slotIdx := -1
	for _, c := range mi.cells {
		if c.slotIdx < 0 {
			continue // literal — not directly clickable
		}
		if x >= c.x && x < c.x+c.w {
			slotIdx = c.slotIdx
			break
		}
	}

	if slotIdx >= 0 && mi.slots[slotIdx] != 0 {
		// Filled slot: position cursor here.
		if shift {
			mi.selEnd = slotIdx
			mi.cursorPos = slotIdx
		} else {
			mi.cursorPos = slotIdx
			mi.clearSelection()
		}
	} else {
		// Empty slot or non-slot area: snap to first empty.
		mi.snapCursorToFirstEmpty()
	}
	mi.resetBlink()
	mi.updateCursorPosition()
}

// ---------------------------------------------------------------------------
// Display helpers
// ---------------------------------------------------------------------------

// buildDisplayString returns the full display string: filled slots + literals +
// mask placeholder characters for empty slots (or spaces if no placeholder set).
func (mi *MaskedInput) buildDisplayString() string {
	if len(mi.tokens) == 0 {
		return ""
	}
	var b strings.Builder
	slotIdx := 0
	for _, tok := range mi.tokens {
		if tok.Kind == maskLiteral {
			b.WriteRune(tok.Literal)
		} else {
			if slotIdx < len(mi.slots) && mi.slots[slotIdx] != 0 {
				b.WriteRune(mi.slots[slotIdx])
			} else if mi.maskPlaceholder != 0 {
				b.WriteRune(mi.maskPlaceholder)
			} else {
				b.WriteRune(' ')
			}
			slotIdx++
		}
	}
	return b.String()
}

// updateTextDisplay updates each cell's glyph content and color to reflect
// the current slot values and theme state.
func (mi *MaskedInput) updateTextDisplay() {
	group := mi.EffectiveTheme().MaskedInput.Group(mi.Variant())
	textColor := group.TextColor.Resolve(mi.state)
	literalColor := group.LiteralColor.Resolve(mi.state)
	maskPHColor := group.MaskPlaceholderColor.Resolve(mi.state)

	if len(mi.cells) == 0 {
		// No mask — show field-level placeholder.
		mi.placeholderNode.SetVisible(true)
		if mi.placeholder != "" {
			mi.placeholderNode.SetContent(mi.placeholder)
			mi.placeholderNode.SetTextColor(sg.RGBA(
				textColor.R(), textColor.G(), textColor.B(), group.PlaceholderAlpha))
		} else {
			mi.placeholderNode.SetContent("")
			mi.placeholderNode.SetTextColor(textColor)
		}
		mi.updateCursorPosition()
		mi.MarkDrawDirty()
		return
	}

	mi.placeholderNode.SetVisible(false)

	for _, c := range mi.cells {
		if c.slotIdx < 0 {
			// Literal cell: content never changes, only color.
			c.node.SetTextColor(literalColor)
		} else {
			// Slot cell: update glyph to reflect current value.
			if mi.slots[c.slotIdx] != 0 {
				c.node.SetContent(string(mi.slots[c.slotIdx]))
				c.node.SetTextColor(textColor)
			} else if mi.maskPlaceholder != 0 {
				c.node.SetContent(string(mi.maskPlaceholder))
				c.node.SetTextColor(maskPHColor)
			} else {
				c.node.SetContent(" ")
				c.node.SetTextColor(textColor)
			}
		}
	}

	mi.updateCursorPosition()
	mi.MarkDrawDirty()
}

// updateCursorPosition places the cursor at the correct cell x offset,
// updates the active-slot highlight, and refreshes the selection rect.
func (mi *MaskedInput) updateCursorPosition() {
	cursorX := mi.cellX(mi.cursorPos)
	mi.cursor.SetPosition(cursorX, 0)

	// Active-slot highlight: covers the cell the cursor is sitting in.
	// Hidden when there is a multi-slot selection (selRect takes over),
	// or when the cursor is past the last slot (all slots filled), or unfocused.
	showActive := mi.focused &&
		!mi.HasSelection() &&
		mi.cursorPos < len(mi.tokenIdxForSlot)
	if showActive {
		tokIdx := mi.tokenIdxForSlot[mi.cursorPos]
		c := mi.cells[tokIdx]
		mi.activeSlotRect.SetPosition(c.x, 0)
		mi.activeSlotRect.SetScale(c.w, displayLineHeight(mi.font, mi.displaySize))
		mi.activeSlotRect.SetVisible(true)
	} else {
		mi.activeSlotRect.SetVisible(false)
	}

	mi.updateSelectionRect()
}

// updateSelectionRect sizes and positions the selection highlight based on
// the current selection range's cell x boundaries.
func (mi *MaskedInput) updateSelectionRect() {
	if !mi.HasSelection() {
		mi.selRect.SetVisible(false)
		return
	}
	lo, hi := mi.selStart, mi.selEnd
	if lo > hi {
		lo, hi = hi, lo
	}
	if hi > len(mi.slots) {
		hi = len(mi.slots)
	}
	xStart := mi.cellX(lo)
	xEnd := mi.cellX(hi)
	mi.selRect.SetPosition(xStart, 0)
	mi.selRect.SetScale(xEnd-xStart, displayLineHeight(mi.font, mi.displaySize))
	mi.selRect.SetVisible(true)
}

// ---------------------------------------------------------------------------
// visualDisplayIdx / rawToDisplayIndex — kept for test compatibility
// ---------------------------------------------------------------------------

// visualDisplayIdx converts a raw slot index to a display character offset,
// advancing past any consecutive literals that immediately follow the position.
func (mi *MaskedInput) visualDisplayIdx(rawIdx int) int {
	idx := mi.rawToDisplayIndex(rawIdx)
	for idx < len(mi.tokens) && mi.tokens[idx].Kind == maskLiteral {
		idx++
	}
	return idx
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// snapCursorToFirstEmpty moves the cursor to the first slot that has no value.
// If all slots are filled, the cursor goes to the end. Called on Tab focus so
// keyboard navigation lands ready to type without extra arrow-key presses.
func (mi *MaskedInput) snapCursorToFirstEmpty() {
	for i, ch := range mi.slots {
		if ch == 0 {
			mi.cursorPos = i
			mi.clearSelection()
			return
		}
	}
	// All filled — go to end.
	mi.cursorPos = len(mi.slots)
	mi.clearSelection()
}

func (mi *MaskedInput) clearSelection() {
	mi.selStart = mi.cursorPos
	mi.selEnd = mi.cursorPos
}

func (mi *MaskedInput) clearAllSlots() {
	for i := range mi.slots {
		mi.slots[i] = 0
	}
}

func (mi *MaskedInput) clearSlotRange(lo, hi int) {
	if lo > hi {
		lo, hi = hi, lo
	}
	for i := lo; i < hi && i < len(mi.slots); i++ {
		mi.slots[i] = 0
	}
}

func (mi *MaskedInput) resetBlink() {
	mi.blinkCounter = 0
	mi.blinkVisible = true
	mi.cursor.SetVisible(mi.focused)
}

func (mi *MaskedInput) fireValueCallbacks() {
	formatted := mi.Value()
	raw := mi.RawValue()
	if mi.onChange != nil {
		mi.onChange(formatted)
	}
	if mi.onRawChange != nil {
		mi.onRawChange(raw)
	}
	// Push to bound external refs without re-triggering our own watches.
	if mi.valueRef != nil {
		mi.valueRef.Set(formatted)
		DefaultScheduler.Flush()
	}
	if mi.rawRef != nil {
		mi.rawRef.Set(raw)
		DefaultScheduler.Flush()
	}
}

func (mi *MaskedInput) checkCompleteTransition() {
	complete := mi.IsComplete()
	if complete == mi.wasComplete {
		return
	}
	mi.wasComplete = complete
	if complete && mi.onComplete != nil {
		mi.onComplete(mi.RawValue(), mi.Value())
	} else if !complete && mi.onIncomplete != nil {
		mi.onIncomplete(mi.RawValue(), mi.Value())
	}
}

// ---------------------------------------------------------------------------
// Testing accessors
// ---------------------------------------------------------------------------

// GetCursorPos returns the current raw-slot cursor position. Used for testing.
func (mi *MaskedInput) GetCursorPos() int { return mi.cursorPos }

// SetCursorPos sets the raw-slot cursor position directly. Used for testing.
func (mi *MaskedInput) SetCursorPos(pos int) { mi.cursorPos = pos }

// GetSelStart returns the selection start raw-slot index. Used for testing.
func (mi *MaskedInput) GetSelStart() int { return mi.selStart }

// GetSelEnd returns the selection end raw-slot index. Used for testing.
func (mi *MaskedInput) GetSelEnd() int { return mi.selEnd }

// CellCount returns the number of rendered cells (one per mask token). Used for testing.
func (mi *MaskedInput) CellCount() int { return len(mi.cells) }

// CellNodeForSlot returns the glyph node for the given slot index. Used for testing.
func (mi *MaskedInput) CellNodeForSlot(slotIdx int) *sg.Node {
	if slotIdx < 0 || slotIdx >= len(mi.tokenIdxForSlot) {
		return nil
	}
	return mi.cells[mi.tokenIdxForSlot[slotIdx]].node
}

// SnapCursorToFirstEmptyForTest calls snapCursorToFirstEmpty. Used for testing.
func (mi *MaskedInput) SnapCursorToFirstEmptyForTest() { mi.snapCursorToFirstEmpty() }

// VisualDisplayIdxForTest calls visualDisplayIdx. Used for testing.
func (mi *MaskedInput) VisualDisplayIdxForTest(rawIdx int) int {
	return mi.visualDisplayIdx(rawIdx)
}
