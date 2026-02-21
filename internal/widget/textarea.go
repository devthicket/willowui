package widget

import (
	"time"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// TextArea is a multi-line text entry field with word wrapping and scrolling.
type TextArea struct {
	Component
	content   *sg.Node   // clipped container for text/cursor/sel
	textNode  *sg.Node   // text display
	cursor    *sg.Node   // WhitePixel sprite: blinking cursor
	selRects  []*sg.Node // WhitePixel sprites: selection highlight per line
	scrollbar *ScrollBar // vertical scrollbar, shown when content overflows

	font        *sg.FontFamily
	displaySize float64
	value       *Ref[string]
	watch       WatchHandle
	maxLength   int
	cursorPos   int
	selStart    int // selection anchor (rune index)
	selEnd      int // selection moving end (rune index); cursor sits here
	scrollY     float64
	rows        int

	// Cursor blink state.
	blinkCounter int
	blinkVisible bool

	// Double-click detection.
	lastClickTime time.Time

	// Callbacks.
	onChange   func(string)
	charFilter func(rune) bool // optional; returns true to accept a character
}

// visualLine represents a single visual line in the text area, accounting
// for both hard line breaks (\n) and word-wrap breaks.
type visualLine struct {
	runeStart int // first rune index in original text (inclusive)
	runeEnd   int // end rune index in original text (exclusive)
}

// VisualLine is the exported representation of a visual line for testing.
type VisualLine struct {
	RuneStart int // first rune index (inclusive)
	RuneEnd   int // end rune index (exclusive)
}

// NewTextArea creates a multi-line text area with the given font source and display size.
func NewTextArea(name string, source *sg.FontFamily, displaySize float64) *TextArea {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	ta := &TextArea{
		font:         font,
		displaySize:  displaySize,
		value:        NewRef(""),
		rows:         3,
		blinkVisible: true,
	}
	initComponent(&ta.Component, name)

	// Background.
	ta.initBackground(name)
	ta.initBorder(name)

	group := ta.EffectiveTheme().TextArea.Group(ta.Variant())
	pad := resolveAutoInsets(group.Padding, defaultTextAreaPadding)

	// Clipped content container.
	ta.content = sg.NewContainer(name + "-content")
	ta.content.SetPosition(pad.Left, pad.Top)
	ta.node.AddChild(ta.content)

	// Text display node (child of content container).
	ta.textNode = sg.NewText(name+"-text", "", font)
	ta.textNode.TextBlock.FontSize = displaySize
	ta.content.AddChild(ta.textNode)

	// Cursor (child of content container).
	ta.cursor = sg.NewSprite(name+"-cursor", sg.TextureRegion{})

	ta.cursor.SetScale(1, displayLineHeight(font, displaySize))
	ta.cursor.SetVisible(false)
	ta.content.AddChild(ta.cursor)

	// Scrollbar — positioned on the right edge, hidden until content overflows.
	ta.scrollbar = NewScrollBar(name + "-scrollbar")
	ta.scrollbar.SetOnChange(func(pos float64) {
		ta.scrollY = pos
		ta.applyScroll()
	})
	ta.scrollbar.SetVisible(false)
	ta.scrollbar.AddToNode(ta.node)

	// Default size.
	defaultW := 250.0
	defaultH := displayLineHeight(font, displaySize)*float64(ta.rows) + pad.Top + pad.Bottom
	ta.SetSize(defaultW, defaultH)

	// Click focuses via the DefaultFocusManager so that at most one
	// component holds focus at a time, and repositions the cursor.
	// Pointer-down sets the selection anchor at press time (before
	// any drag dead-zone). Shift+click extends from the existing anchor.
	ta.node.OnPointerDown(func(ctx sg.PointerContext) {
		if !ta.enabled {
			return
		}
		ta.pressed = true
		ta.bubbleActivation()
		DefaultFocusManager.SetFocus(&ta.Component)

		now := time.Now()
		if now.Sub(ta.lastClickTime) < doubleClickThreshold {
			ta.setCursorFromClick(ctx.LocalX, ctx.LocalY, false)
			ta.selectWordAtCursor()
			ta.lastClickTime = time.Time{} // reset to prevent triple-click
		} else {
			shift := ctx.Modifiers&sg.ModShift != 0
			ta.setCursorFromClick(ctx.LocalX, ctx.LocalY, shift)
			ta.lastClickTime = now
		}
		ta.UpdateVisuals()
	})

	// Drag-to-select: each drag frame extends selection from the anchor.
	ta.node.OnDrag(func(ctx sg.DragContext) {
		if !ta.enabled {
			return
		}
		ta.setCursorFromClick(ctx.LocalX, ctx.LocalY, true)
	})

	ta.wireVisualCallbacks(ta.UpdateVisuals)

	ta.SetCursorShape(engine.CursorShapeText)

	// Focus: text areas participate in tab and spatial nav, intercept arrows.
	ta.enableFocusNavigation()
	ta.InterceptArrows = true
	ta.ConsumeHandledKeys = false // widget's own Update handles the key; only block spatial nav

	// HandleKey: boundary-aware arrow key query (pure, no side effects).
	ta.SetHandleKey(func(key engine.Key) bool {
		runes := []rune(ta.value.Peek())
		switch key {
		case engine.KeyLeft:
			return ta.cursorPos > 0
		case engine.KeyRight:
			return ta.cursorPos < len(runes)
		case engine.KeyUp:
			vlines := ta.getVisualLines()
			line, _ := ta.cursorVisualLineCol(ta.cursorPos, vlines)
			return line > 0
		case engine.KeyDown:
			vlines := ta.getVisualLines()
			line, _ := ta.cursorVisualLineCol(ta.cursorPos, vlines)
			return line < len(vlines)-1
		}
		return false
	})

	ta.UpdateVisuals()

	// Auto-update: keyboard input via willow's per-frame hook.
	ta.node.OnUpdate = func(_ float64) {
		ta.Update()
	}

	return ta
}

// HasSelection returns true when text is selected.
func (ta *TextArea) HasSelection() bool {
	return ta.selStart != ta.selEnd
}

// SelectedText returns the currently selected text.
func (ta *TextArea) SelectedText() string {
	if !ta.HasSelection() {
		return ""
	}
	runes := []rune(ta.value.Peek())
	lo, hi := ta.selStart, ta.selEnd
	if lo > hi {
		lo, hi = hi, lo
	}
	if hi > len(runes) {
		hi = len(runes)
	}
	return string(runes[lo:hi])
}

// SelectAll selects the entire text content.
func (ta *TextArea) SelectAll() {
	runes := []rune(ta.value.Peek())
	ta.selStart = 0
	ta.selEnd = len(runes)
	ta.cursorPos = len(runes)
	ta.resetBlink()
	ta.updateCursorPosition()
}

// clearSelection collapses selection to the current cursor position.
func (ta *TextArea) clearSelection() {
	ta.selStart = ta.cursorPos
	ta.selEnd = ta.cursorPos
}

// deleteSelection deletes the selected text, moving cursor to the start edge.
func (ta *TextArea) deleteSelection() {
	if !ta.HasSelection() {
		return
	}
	lo, hi := ta.selStart, ta.selEnd
	if lo > hi {
		lo, hi = hi, lo
	}
	runes := []rune(ta.value.Peek())
	if hi > len(runes) {
		hi = len(runes)
	}
	newRunes := append(runes[:lo], runes[hi:]...)
	ta.cursorPos = lo
	ta.selStart = lo
	ta.selEnd = lo
	ta.value.Set(string(newRunes))
	DefaultScheduler.Flush()
}

// Value returns the current text.
func (ta *TextArea) Value() string {
	return ta.value.Peek()
}

// SetValue sets the text content.
func (ta *TextArea) SetValue(v string) {
	if ta.maxLength > 0 && len([]rune(v)) > ta.maxLength {
		v = string([]rune(v)[:ta.maxLength])
	}
	ta.value.Set(v)
	DefaultScheduler.Flush()
	ta.cursorPos = len([]rune(v))
	ta.clearSelection()
	ta.updateTextDisplay()
}

// SetRows sets the number of visible rows.
func (ta *TextArea) SetRows(n int) {
	if n < 1 {
		n = 1
	}
	ta.rows = n
	pad := resolveAutoInsets(ta.EffectiveTheme().TextArea.Group(ta.Variant()).Padding, defaultTextAreaPadding)
	ta.SetSize(ta.Width, displayLineHeight(ta.font, ta.displaySize)*float64(n)+pad.Top+pad.Bottom)
}

// SetMaxLength limits the number of characters (0 = no limit).
func (ta *TextArea) SetMaxLength(n int) {
	ta.maxLength = n
}

// SetCharFilter sets a function called for each typed or pasted character.
// Return true to accept the character, false to reject it.
// Newlines are always allowed regardless of the filter.
// Passing nil clears any existing filter.
func (ta *TextArea) SetCharFilter(fn func(rune) bool) {
	ta.charFilter = fn
}

// SetNumericOnly restricts input to digit characters (0–9).
func (ta *TextArea) SetNumericOnly() {
	ta.charFilter = func(ch rune) bool { return ch >= '0' && ch <= '9' }
}

// SetAlphanumericOnly restricts input to ASCII letters and digits.
func (ta *TextArea) SetAlphanumericOnly() {
	ta.charFilter = func(ch rune) bool {
		return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
	}
}

// SetAllowedChars restricts input to characters present in the given string.
func (ta *TextArea) SetAllowedChars(chars string) {
	allowed := []rune(chars)
	ta.charFilter = func(ch rune) bool {
		for _, r := range allowed {
			if r == ch {
				return true
			}
		}
		return false
	}
}

// SetOnChange sets the callback for text changes.
func (ta *TextArea) SetOnChange(fn func(string)) {
	ta.onChange = fn
}

// SetSize sets the textarea dimensions.
func (ta *TextArea) SetSize(w, h float64) {
	ta.Width = w
	ta.Height = h
	ta.resizeBackground(w, h)
	ta.resizeBorder(w, h)
	ta.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	pad := resolveAutoInsets(ta.EffectiveTheme().TextArea.Group(ta.Variant()).Padding, defaultTextAreaPadding)
	innerH := h - pad.Top - pad.Bottom

	// Position scrollbar on the right edge.
	sbW := float64(DefaultScrollBarWidth)
	ta.scrollbar.SetSize(sbW, innerH)
	ta.scrollbar.SetPosition(w-sbW-pad.Right, pad.Top)

	// Content width excludes scrollbar when visible.
	innerW := w - pad.Left - pad.Right
	if ta.scrollbar.IsVisible() {
		innerW -= sbW
	}
	ta.textNode.SetWrapWidth(innerW)
	// Update mask for clipping.
	ta.updateContentMask(innerW, innerH)
	ta.MarkLayoutDirty()
}

// updateContentMask creates or updates the mask that clips content to the
// inner area of the text area. The mask root must be a container because
// willow ignores the root mask node's own transform — only children's
// transforms are applied.
func (ta *TextArea) updateContentMask(innerW, innerH float64) {
	maskRoot := sg.NewContainer(ta.node.Name + "-mask")
	maskSprite := sg.NewSprite(ta.node.Name+"-mask-rect", sg.TextureRegion{})
	maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	maskSprite.SetScale(innerW, innerH)
	maskRoot.AddChild(maskSprite)
	ta.content.SetMask(maskRoot)
}

// BindValue binds the text area to a reactive Ref[string].
func (ta *TextArea) BindValue(ref *Ref[string]) {
	ta.watch.Stop()
	ta.value = ref
	ta.SetValue(ref.Peek())
	ta.watch = WatchValue(ref, func(_, newVal string) {
		ta.cursorPos = len([]rune(newVal))
		ta.clearSelection()
		ta.updateTextDisplay()
	})
}

// InsertText inserts text at the current cursor position.
// If there is a selection, it replaces the selected text.
func (ta *TextArea) InsertText(s string) {
	// Apply character filter (newlines are always allowed in TextArea).
	if ta.charFilter != nil {
		filtered := make([]rune, 0, len([]rune(s)))
		for _, ch := range s {
			if ch == '\n' || ch == '\r' || ta.charFilter(ch) {
				filtered = append(filtered, ch)
			}
		}
		s = string(filtered)
	}

	if ta.HasSelection() {
		ta.deleteSelection()
	}

	runes := []rune(ta.value.Peek())
	insert := []rune(s)

	if ta.maxLength > 0 && len(runes)+len(insert) > ta.maxLength {
		insert = insert[:ta.maxLength-len(runes)]
	}
	if len(insert) == 0 {
		return
	}

	newRunes := make([]rune, 0, len(runes)+len(insert))
	newRunes = append(newRunes, runes[:ta.cursorPos]...)
	newRunes = append(newRunes, insert...)
	newRunes = append(newRunes, runes[ta.cursorPos:]...)

	ta.cursorPos += len(insert)
	ta.clearSelection()
	ta.value.Set(string(newRunes))
	DefaultScheduler.Flush()
	ta.resetBlink()
	ta.updateTextDisplay()

	if ta.onChange != nil {
		ta.onChange(ta.value.Peek())
	}
}

// DeleteBack deletes the character before the cursor (backspace).
// If there is a selection, it deletes the selected text instead.
func (ta *TextArea) DeleteBack() {
	if ta.HasSelection() {
		ta.deleteSelection()
		ta.resetBlink()
		ta.updateTextDisplay()
		if ta.onChange != nil {
			ta.onChange(ta.value.Peek())
		}
		return
	}
	if ta.cursorPos <= 0 {
		return
	}
	runes := []rune(ta.value.Peek())
	newRunes := append(runes[:ta.cursorPos-1], runes[ta.cursorPos:]...)
	ta.cursorPos--
	ta.clearSelection()
	ta.value.Set(string(newRunes))
	DefaultScheduler.Flush()
	ta.resetBlink()
	ta.updateTextDisplay()

	if ta.onChange != nil {
		ta.onChange(ta.value.Peek())
	}
}

// DeleteForward deletes the character after the cursor (delete key).
// If there is a selection, it deletes the selected text instead.
func (ta *TextArea) DeleteForward() {
	if ta.HasSelection() {
		ta.deleteSelection()
		ta.resetBlink()
		ta.updateTextDisplay()
		if ta.onChange != nil {
			ta.onChange(ta.value.Peek())
		}
		return
	}
	runes := []rune(ta.value.Peek())
	if ta.cursorPos >= len(runes) {
		return
	}
	newRunes := append(runes[:ta.cursorPos], runes[ta.cursorPos+1:]...)
	ta.clearSelection()
	ta.value.Set(string(newRunes))
	DefaultScheduler.Flush()
	ta.resetBlink()
	ta.updateTextDisplay()

	if ta.onChange != nil {
		ta.onChange(ta.value.Peek())
	}
}

// SetEnabled overrides Component.SetEnabled to also update visuals.
func (ta *TextArea) SetEnabled(v bool) {
	ta.Component.SetEnabled(v)
	ta.UpdateVisuals()
}

// UpdateVisuals applies theme colors based on current state.
func (ta *TextArea) UpdateVisuals() {
	ta.state = computeState(ta.enabled, ta.focused, ta.hovered, ta.pressed)
	group := ta.EffectiveTheme().TextArea.Group(ta.Variant())
	ta.applyCornerRadius(group.CornerRadius)
	bg := group.Background.Resolve(ta.state)
	ta.applyBackground(bg)
	ta.applyBorder(group.Border.Resolve(ta.state), group.BorderWidth, bg)
	ta.textNode.TextBlock.Color = group.TextColor.Resolve(ta.state)
	ta.cursor.SetColor(group.CursorColor.Resolve(ta.state))
	ta.cursor.SetVisible(ta.focused && ta.blinkVisible)
	// Keep content position in sync with the current theme's padding.
	pad := resolveAutoInsets(group.Padding, defaultTextAreaPadding)
	if ta.content.X() != pad.Left || ta.content.Y() != pad.Top {
		ta.content.SetPosition(pad.Left, pad.Top)
	}
	ta.applyFocusRing(group.FocusColor.Resolve(ta.state), group.FocusRingWidth)
	ta.MarkDrawDirty()
}

// MoveCursorLeft moves the cursor one position to the left.
func (ta *TextArea) MoveCursorLeft() {
	ta.moveCursorLeftShift(false)
}

func (ta *TextArea) moveCursorLeftShift(shift bool) {
	if !shift && ta.HasSelection() {
		lo := ta.selStart
		if ta.selEnd < lo {
			lo = ta.selEnd
		}
		ta.cursorPos = lo
		ta.clearSelection()
		ta.resetBlink()
		ta.updateCursorPosition()
		return
	}
	if ta.cursorPos > 0 {
		ta.cursorPos--
		if shift {
			ta.selEnd = ta.cursorPos
		} else {
			ta.clearSelection()
		}
		ta.resetBlink()
		ta.updateCursorPosition()
	}
}

// MoveCursorRight moves the cursor one position to the right.
func (ta *TextArea) MoveCursorRight() {
	ta.moveCursorRightShift(false)
}

func (ta *TextArea) moveCursorRightShift(shift bool) {
	if !shift && ta.HasSelection() {
		hi := ta.selStart
		if ta.selEnd > hi {
			hi = ta.selEnd
		}
		ta.cursorPos = hi
		ta.clearSelection()
		ta.resetBlink()
		ta.updateCursorPosition()
		return
	}
	runes := []rune(ta.value.Peek())
	if ta.cursorPos < len(runes) {
		ta.cursorPos++
		if shift {
			ta.selEnd = ta.cursorPos
		} else {
			ta.clearSelection()
		}
		ta.resetBlink()
		ta.updateCursorPosition()
	}
}

// MoveCursorUp moves the cursor to the same horizontal position on the
// previous visual line (or to the start of the text if already on the first line).
func (ta *TextArea) MoveCursorUp() {
	ta.moveCursorUpShift(false)
}

func (ta *TextArea) moveCursorUpShift(shift bool) {
	runes := []rune(ta.value.Peek())
	vlines := ta.getVisualLines()
	line, col := ta.cursorVisualLineCol(ta.cursorPos, vlines)

	if line == 0 {
		ta.cursorPos = 0
	} else {
		// Measure the x offset on the current line to find equivalent column above.
		curLineRunes := runes[vlines[line].runeStart:vlines[line].runeEnd]
		xTarget, _ := measureDisplay(ta.font, string(curLineRunes[:col]), ta.displaySize)
		prevLineRunes := runes[vlines[line-1].runeStart:vlines[line-1].runeEnd]
		newCol := ta.colFromX(prevLineRunes, xTarget)
		ta.cursorPos = vlines[line-1].runeStart + newCol
	}
	if shift {
		ta.selEnd = ta.cursorPos
	} else {
		ta.clearSelection()
	}
	ta.resetBlink()
	ta.updateCursorPosition()
}

// MoveCursorDown moves the cursor to the same horizontal position on the
// next visual line (or to the end of the text if already on the last line).
func (ta *TextArea) MoveCursorDown() {
	ta.moveCursorDownShift(false)
}

func (ta *TextArea) moveCursorDownShift(shift bool) {
	runes := []rune(ta.value.Peek())
	vlines := ta.getVisualLines()
	line, col := ta.cursorVisualLineCol(ta.cursorPos, vlines)

	if line >= len(vlines)-1 {
		ta.cursorPos = len(runes)
	} else {
		curLineRunes := runes[vlines[line].runeStart:vlines[line].runeEnd]
		xTarget, _ := measureDisplay(ta.font, string(curLineRunes[:col]), ta.displaySize)
		nextLineRunes := runes[vlines[line+1].runeStart:vlines[line+1].runeEnd]
		newCol := ta.colFromX(nextLineRunes, xTarget)
		ta.cursorPos = vlines[line+1].runeStart + newCol
	}
	if shift {
		ta.selEnd = ta.cursorPos
	} else {
		ta.clearSelection()
	}
	ta.resetBlink()
	ta.updateCursorPosition()
}

func (ta *TextArea) moveCursorHomeShift(shift bool) {
	vlines := ta.getVisualLines()
	line, _ := ta.cursorVisualLineCol(ta.cursorPos, vlines)
	ta.cursorPos = vlines[line].runeStart
	if shift {
		ta.selEnd = ta.cursorPos
	} else {
		ta.clearSelection()
	}
	ta.resetBlink()
	ta.updateCursorPosition()
}

func (ta *TextArea) moveCursorEndShift(shift bool) {
	vlines := ta.getVisualLines()
	line, _ := ta.cursorVisualLineCol(ta.cursorPos, vlines)
	ta.cursorPos = vlines[line].runeEnd
	if shift {
		ta.selEnd = ta.cursorPos
	} else {
		ta.clearSelection()
	}
	ta.resetBlink()
	ta.updateCursorPosition()
}

func (ta *TextArea) moveCursorPageUpShift(shift bool) {
	runes := []rune(ta.value.Peek())
	vlines := ta.getVisualLines()
	line, col := ta.cursorVisualLineCol(ta.cursorPos, vlines)

	targetLine := line - ta.rows
	if targetLine < 0 {
		targetLine = 0
	}

	if targetLine == 0 && line == 0 {
		ta.cursorPos = 0
	} else {
		curLineRunes := runes[vlines[line].runeStart:vlines[line].runeEnd]
		xTarget, _ := measureDisplay(ta.font, string(curLineRunes[:col]), ta.displaySize)
		tgtLineRunes := runes[vlines[targetLine].runeStart:vlines[targetLine].runeEnd]
		newCol := ta.colFromX(tgtLineRunes, xTarget)
		ta.cursorPos = vlines[targetLine].runeStart + newCol
	}
	if shift {
		ta.selEnd = ta.cursorPos
	} else {
		ta.clearSelection()
	}
	ta.resetBlink()
	ta.updateCursorPosition()
}

func (ta *TextArea) moveCursorPageDownShift(shift bool) {
	runes := []rune(ta.value.Peek())
	vlines := ta.getVisualLines()
	line, col := ta.cursorVisualLineCol(ta.cursorPos, vlines)

	targetLine := line + ta.rows
	lastLine := len(vlines) - 1
	if targetLine > lastLine {
		targetLine = lastLine
	}

	if targetLine == lastLine && line == lastLine {
		ta.cursorPos = len(runes)
	} else {
		curLineRunes := runes[vlines[line].runeStart:vlines[line].runeEnd]
		xTarget, _ := measureDisplay(ta.font, string(curLineRunes[:col]), ta.displaySize)
		tgtLineRunes := runes[vlines[targetLine].runeStart:vlines[targetLine].runeEnd]
		newCol := ta.colFromX(tgtLineRunes, xTarget)
		ta.cursorPos = vlines[targetLine].runeStart + newCol
	}
	if shift {
		ta.selEnd = ta.cursorPos
	} else {
		ta.clearSelection()
	}
	ta.resetBlink()
	ta.updateCursorPosition()
}

// selectWordAtCursor selects the word under the current cursor position.
func (ta *TextArea) selectWordAtCursor() {
	runes := []rune(ta.value.Peek())
	lo, hi := wordBoundaries(runes, ta.cursorPos)
	ta.selStart = lo
	ta.selEnd = hi
	ta.cursorPos = hi
	ta.resetBlink()
	ta.updateCursorPosition()
}

// cursorVisualLineCol returns the visual line index and column (rune offset
// within that visual line) for the given cursor position.
func (ta *TextArea) cursorVisualLineCol(pos int, vlines []visualLine) (line, col int) {
	for i, vl := range vlines {
		if pos < vl.runeEnd {
			return i, pos - vl.runeStart
		}
		if pos == vl.runeEnd {
			// Soft wrap: next line starts at the same rune index → cursor
			// belongs to the next visual line at column 0.
			if i+1 < len(vlines) && vlines[i+1].runeStart == pos {
				return i + 1, 0
			}
			return i, pos - vl.runeStart
		}
	}
	// Past end — last line.
	last := len(vlines) - 1
	return last, vlines[last].runeEnd - vlines[last].runeStart
}

// colFromX returns the rune column on lineRunes closest to the pixel offset x.
func (ta *TextArea) colFromX(lineRunes []rune, x float64) int {
	best := 0
	for i := 1; i <= len(lineRunes); i++ {
		w, _ := measureDisplay(ta.font, string(lineRunes[:i]), ta.displaySize)
		if w <= x {
			best = i
		} else {
			prevW, _ := measureDisplay(ta.font, string(lineRunes[:i-1]), ta.displaySize)
			if x-prevW > w-x {
				best = i
			}
			break
		}
	}
	return best
}

// getVisualLines returns the text split into visual lines, accounting for
// both hard line breaks (\n) and word-wrap breaks based on the current
// wrap width.
func (ta *TextArea) getVisualLines() []visualLine {
	text := ta.value.Peek()
	runes := []rune(text)
	wrapWidth := ta.textNode.TextBlock.WrapWidth

	if wrapWidth <= 0 || ta.font == nil {
		// No wrapping — split on \n only.
		var lines []visualLine
		start := 0
		for i, r := range runes {
			if r == '\n' {
				lines = append(lines, visualLine{runeStart: start, runeEnd: i})
				start = i + 1
			}
		}
		lines = append(lines, visualLine{runeStart: start, runeEnd: len(runes)})
		return lines
	}

	var lines []visualLine
	paraStart := 0
	for {
		// Find end of paragraph (next \n or end of text).
		paraEnd := len(runes)
		for i := paraStart; i < len(runes); i++ {
			if runes[i] == '\n' {
				paraEnd = i
				break
			}
		}

		// Word-wrap this paragraph into visual lines.
		ta.wrapParagraph(runes, paraStart, paraEnd, wrapWidth, &lines)

		if paraEnd >= len(runes) {
			break
		}
		paraStart = paraEnd + 1 // skip \n
	}

	// Ensure at least one line.
	if len(lines) == 0 {
		lines = append(lines, visualLine{runeStart: 0, runeEnd: 0})
	}
	return lines
}

// wrapParagraph splits the rune range [start, end) into visual lines that
// fit within wrapWidth, breaking at word boundaries (spaces).
func (ta *TextArea) wrapParagraph(runes []rune, start, end int, wrapWidth float64, lines *[]visualLine) {
	if start >= end {
		*lines = append(*lines, visualLine{runeStart: start, runeEnd: end})
		return
	}

	para := runes[start:end]

	// Find word boundaries: a word is a run of non-space characters.
	type wordSpan struct {
		start, end int // indices within para
	}
	var words []wordSpan
	i := 0
	for i < len(para) {
		// Skip spaces.
		for i < len(para) && para[i] == ' ' {
			i++
		}
		if i >= len(para) {
			break
		}
		ws := i
		for i < len(para) && para[i] != ' ' {
			i++
		}
		words = append(words, wordSpan{ws, i})
	}

	if len(words) == 0 {
		// All spaces or empty — single visual line.
		*lines = append(*lines, visualLine{runeStart: start, runeEnd: end})
		return
	}

	lineStartIdx := 0 // rune index within para where current line starts
	lineStartWord := 0

	for wi := 0; wi < len(words); wi++ {
		// Candidate line from lineStartIdx to end of this word.
		candidateEnd := words[wi].end
		candidate := string(para[lineStartIdx:candidateEnd])
		cw, _ := measureDisplay(ta.font, candidate, ta.displaySize)

		if cw > wrapWidth && wi > lineStartWord {
			// Break before this word: emit line up to end of previous word.
			prevWordEnd := words[wi-1].end
			*lines = append(*lines, visualLine{
				runeStart: start + lineStartIdx,
				runeEnd:   start + prevWordEnd,
			})
			// Next line starts at this word (skip leading spaces).
			lineStartIdx = words[wi].start
			lineStartWord = wi
		}
	}

	// Emit remaining runes (including trailing spaces).
	*lines = append(*lines, visualLine{
		runeStart: start + lineStartIdx,
		runeEnd:   end,
	})
}

// Update handles keyboard input, cursor blink, and visual state.
func (ta *TextArea) Update() {
	if !ta.focused {
		ta.UpdateVisuals()
		return
	}

	im := DefaultInputManager
	shift := engine.IsKeyPressed(engine.KeyShift)
	ctrl := engine.IsKeyPressed(engine.KeyControl) || engine.IsKeyPressed(engine.KeyMeta)

	// Ctrl+A: select all.
	if ctrl && im.IsKeyJustAvailable(engine.KeyA) {
		im.Consume(engine.KeyA)
		ta.SelectAll()
		return
	}

	// Ctrl+C: copy selection to clipboard.
	if ctrl && im.IsKeyJustAvailable(engine.KeyC) {
		im.Consume(engine.KeyC)
		if sel := ta.SelectedText(); sel != "" {
			clipboardWrite(sel)
		}
		return
	}

	// Ctrl+X: cut selection to clipboard.
	if ctrl && im.IsKeyJustAvailable(engine.KeyX) {
		im.Consume(engine.KeyX)
		if sel := ta.SelectedText(); sel != "" {
			clipboardWrite(sel)
			ta.deleteSelection()
			ta.resetBlink()
			ta.updateTextDisplay()
			if ta.onChange != nil {
				ta.onChange(ta.value.Peek())
			}
		}
		return
	}

	// Ctrl+V: paste from clipboard.
	if ctrl && im.IsKeyJustAvailable(engine.KeyV) {
		im.Consume(engine.KeyV)
		if text, err := clipboardRead(); err == nil && text != "" {
			ta.InsertText(text)
		}
		return
	}

	// Cmd/Ctrl+Left: move to start of visual line (same as Home).
	if ctrl && im.IsKeyJustAvailable(engine.KeyLeft) {
		im.Consume(engine.KeyLeft)
		ta.moveCursorHomeShift(shift)
		return
	}
	// Cmd/Ctrl+Right: move to end of visual line (same as End).
	if ctrl && im.IsKeyJustAvailable(engine.KeyRight) {
		im.Consume(engine.KeyRight)
		ta.moveCursorEndShift(shift)
		return
	}

	// Read typed characters (skip if ctrl is held to avoid control chars).
	if !ctrl {
		chars := engine.AppendInputChars(nil)
		if scene := currentScene(); scene != nil {
			chars = scene.AppendInjectedChars(chars)
		}
		if len(chars) > 0 {
			ta.InsertText(string(chars))
		}
	}

	// Handle special keys.
	if im.IsKeyJustAvailable(engine.KeyEscape) {
		im.Consume(engine.KeyEscape)
		DefaultFocusManager.ClearFocus()
		ta.UpdateVisuals()
		return
	}
	if im.IsKeyJustAvailable(engine.KeyBackspace) {
		im.Consume(engine.KeyBackspace)
		ta.DeleteBack()
	}
	if im.IsKeyJustAvailable(engine.KeyDelete) {
		im.Consume(engine.KeyDelete)
		ta.DeleteForward()
	}
	if im.IsKeyJustAvailable(engine.KeyLeft) {
		im.Consume(engine.KeyLeft)
		ta.moveCursorLeftShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyRight) {
		im.Consume(engine.KeyRight)
		ta.moveCursorRightShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyUp) {
		im.Consume(engine.KeyUp)
		ta.moveCursorUpShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyDown) {
		im.Consume(engine.KeyDown)
		ta.moveCursorDownShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyHome) {
		im.Consume(engine.KeyHome)
		ta.moveCursorHomeShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyEnd) {
		im.Consume(engine.KeyEnd)
		ta.moveCursorEndShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyPageUp) {
		im.Consume(engine.KeyPageUp)
		ta.moveCursorPageUpShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyPageDown) {
		im.Consume(engine.KeyPageDown)
		ta.moveCursorPageDownShift(shift)
	}
	if im.IsKeyJustAvailable(engine.KeyEnter) {
		im.Consume(engine.KeyEnter)
		ta.InsertText("\n")
	}

	// Cursor blink.
	ta.blinkCounter++
	if ta.blinkCounter >= 30 {
		ta.blinkCounter = 0
		ta.blinkVisible = !ta.blinkVisible
		ta.cursor.SetVisible(ta.blinkVisible)
	}

	ta.UpdateVisuals()
}

// Dispose stops reactive watches and disposes the component tree.
func (ta *TextArea) Dispose() {
	ta.watch.Stop()
	ta.scrollbar.Dispose()
	ta.Component.Dispose()
}

// resetBlink makes the cursor visible and resets the blink timer.
func (ta *TextArea) resetBlink() {
	ta.blinkCounter = 0
	ta.blinkVisible = true
	ta.cursor.SetVisible(ta.focused)
}

// setCursorFromClick positions the cursor at the character closest to the
// given local (x, y) coordinates within the text area. If shift is true,
// the selection is extended from the current anchor.
func (ta *TextArea) setCursorFromClick(localX, localY float64, shift bool) {
	if ta.font == nil {
		return
	}
	pad := resolveAutoInsets(ta.EffectiveTheme().TextArea.Group(ta.Variant()).Padding, defaultTextAreaPadding)
	x := localX - pad.Left
	y := localY - pad.Top + ta.scrollY

	runes := []rune(ta.value.Peek())
	lineHeight := displayLineHeight(ta.font, ta.displaySize)
	vlines := ta.getVisualLines()

	// Determine which visual line was clicked.
	clickedLine := int(y / lineHeight)
	if clickedLine < 0 {
		clickedLine = 0
	}
	if clickedLine >= len(vlines) {
		clickedLine = len(vlines) - 1
	}

	// Find the character position within the clicked visual line.
	vl := vlines[clickedLine]
	lineRunes := runes[vl.runeStart:vl.runeEnd]
	charPos := ta.colFromX(lineRunes, x)

	newPos := vl.runeStart + charPos
	if newPos > len(runes) {
		newPos = len(runes)
	}

	if shift {
		ta.selEnd = newPos
		ta.cursorPos = newPos
	} else {
		ta.cursorPos = newPos
		ta.clearSelection()
	}
	ta.resetBlink()
	ta.updateCursorPosition()
}

// splitLines splits text on newlines, always returning at least one element.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

// updateTextDisplay syncs the text node with the current value.
func (ta *TextArea) updateTextDisplay() {
	v := ta.value.Peek()
	ta.textNode.SetContent(v)
	ta.textNode.SetTextColor(ta.EffectiveTheme().TextArea.Group(ta.Variant()).TextColor.Resolve(ta.state))
	ta.updateScrollbar()
	ta.updateCursorPosition()
	ta.MarkDrawDirty()
}

// updateCursorPosition places the cursor based on the current cursor position
// and updates the selection highlight rectangles.
func (ta *TextArea) updateCursorPosition() {
	if ta.font == nil {
		return
	}
	runes := []rune(ta.value.Peek())
	if ta.cursorPos > len(runes) {
		ta.cursorPos = len(runes)
	}

	vlines := ta.getVisualLines()
	line, col := ta.cursorVisualLineCol(ta.cursorPos, vlines)

	// Measure text on this visual line up to the cursor column.
	vl := vlines[line]
	lineText := string(runes[vl.runeStart : vl.runeStart+col])
	w, _ := measureDisplay(ta.font, lineText, ta.displaySize)

	lineHeight := displayLineHeight(ta.font, ta.displaySize)
	cursorY := float64(line)*lineHeight - ta.scrollY
	ta.cursor.SetPosition(w, cursorY)

	// Auto-scroll to keep cursor visible.
	ta.ensureCursorVisible(line)

	ta.updateSelectionRects()
}

// ensureCursorVisible adjusts scrollY so the cursor's visual line is visible.
func (ta *TextArea) ensureCursorVisible(line int) {
	lineHeight := displayLineHeight(ta.font, ta.displaySize)
	pad := resolveAutoInsets(ta.EffectiveTheme().TextArea.Group(ta.Variant()).Padding, defaultTextAreaPadding)
	innerH := ta.Height - pad.Top - pad.Bottom

	// Clamp scrollY so content doesn't leave a gap at the bottom.
	vlines := ta.getVisualLines()
	totalH := float64(len(vlines)) * lineHeight
	maxScroll := totalH - innerH
	if maxScroll < 0 {
		maxScroll = 0
	}
	if ta.scrollY > maxScroll {
		ta.scrollY = maxScroll
	}

	cursorTop := float64(line) * lineHeight
	cursorBottom := cursorTop + lineHeight

	if cursorTop < ta.scrollY {
		ta.scrollY = cursorTop
	} else if cursorBottom > ta.scrollY+innerH {
		ta.scrollY = cursorBottom - innerH
	}
	if ta.scrollY < 0 {
		ta.scrollY = 0
	}

	ta.applyScroll()
	// Update cursor Y to reflect scroll.
	ta.cursor.SetY(float64(line)*lineHeight - ta.scrollY)
}

// applyScroll repositions the text node and selection rects based on the
// current scrollY, and syncs the scrollbar.
func (ta *TextArea) applyScroll() {
	ta.textNode.SetY(-ta.scrollY)
	ta.scrollbar.SetScrollPos(ta.scrollY)
	ta.updateSelectionRects()
	ta.MarkDrawDirty()
}

// updateScrollbar shows or hides the scrollbar and updates its content/view
// sizes based on the current text content. Recalculates wrap width when the
// scrollbar visibility changes.
func (ta *TextArea) updateScrollbar() {
	if ta.font == nil {
		return
	}
	vlines := ta.getVisualLines()
	lineHeight := displayLineHeight(ta.font, ta.displaySize)
	pad := resolveAutoInsets(ta.EffectiveTheme().TextArea.Group(ta.Variant()).Padding, defaultTextAreaPadding)
	totalH := float64(len(vlines)) * lineHeight
	innerH := ta.Height - pad.Top - pad.Bottom

	wasVisible := ta.scrollbar.IsVisible()
	needsScroll := totalH > innerH

	ta.scrollbar.SetVisible(needsScroll)
	if needsScroll {
		ta.scrollbar.SetContentSize(totalH, innerH)
	}

	// If visibility changed, recalculate wrap width.
	if wasVisible != needsScroll {
		sbW := float64(DefaultScrollBarWidth)
		innerW := ta.Width - pad.Left - pad.Right
		if needsScroll {
			innerW -= sbW
		}
		ta.textNode.SetWrapWidth(innerW)
		ta.updateContentMask(innerW, innerH)
	}
}

// updateSelectionRects sizes and positions per-line selection highlight rects.
func (ta *TextArea) updateSelectionRects() {
	if !ta.HasSelection() || ta.font == nil {
		for _, r := range ta.selRects {
			r.SetVisible(false)
		}
		return
	}

	lineHeight := displayLineHeight(ta.font, ta.displaySize)
	runes := []rune(ta.value.Peek())
	vlines := ta.getVisualLines()

	lo, hi := ta.selStart, ta.selEnd
	if lo > hi {
		lo, hi = hi, lo
	}
	if hi > len(runes) {
		hi = len(runes)
	}

	// Find which visual lines the selection spans and compute per-line ranges.
	type lineRange struct {
		line     int
		startCol int
		endCol   int
	}
	var ranges []lineRange

	for i, vl := range vlines {
		vlLen := vl.runeEnd - vl.runeStart
		// Does this visual line intersect [lo, hi)?
		if hi > vl.runeStart && lo < vl.runeEnd {
			startCol := 0
			if lo > vl.runeStart {
				startCol = lo - vl.runeStart
			}
			endCol := vlLen
			if hi-vl.runeStart < vlLen {
				endCol = hi - vl.runeStart
			}
			ranges = append(ranges, lineRange{line: i, startCol: startCol, endCol: endCol})
		}
	}

	// Ensure we have enough sprite nodes.
	for len(ta.selRects) < len(ranges) {
		r := sg.NewSprite(ta.node.Name+"-sel", sg.TextureRegion{})
		r.SetColor(ta.EffectiveTheme().TextArea.Group(ta.Variant()).SelectionColor.Resolve(ta.state))
		r.SetVisible(false)
		// Insert at start of content (before textNode).
		ta.content.AddChildAt(r, 0)
		ta.selRects = append(ta.selRects, r)
	}

	// Position active rects and update their color (theme may have changed).
	selColor := ta.EffectiveTheme().TextArea.Group(ta.Variant()).SelectionColor.Resolve(ta.state)
	for i, rng := range ranges {
		vl := vlines[rng.line]
		lineRunes := runes[vl.runeStart:vl.runeEnd]
		xStart, _ := measureDisplay(ta.font, string(lineRunes[:rng.startCol]), ta.displaySize)
		xEnd, _ := measureDisplay(ta.font, string(lineRunes[:rng.endCol]), ta.displaySize)

		rect := ta.selRects[i]
		rect.SetColor(selColor)
		rect.SetPosition(xStart, float64(rng.line)*lineHeight-ta.scrollY)
		rect.SetScale(xEnd-xStart, lineHeight)
		rect.SetVisible(true)
	}

	// Hide unused rects.
	for i := len(ranges); i < len(ta.selRects); i++ {
		ta.selRects[i].SetVisible(false)
	}
}

// GetCursorPos returns the current cursor rune position. Used for testing.
func (ta *TextArea) GetCursorPos() int { return ta.cursorPos }

// SetCursorPos sets the cursor rune position directly. Used for testing.
func (ta *TextArea) SetCursorPos(pos int) { ta.cursorPos = pos }

// GetSelStart returns the selection start rune index. Used for testing.
func (ta *TextArea) GetSelStart() int { return ta.selStart }

// SetSelStart sets the selection start directly. Used for testing.
func (ta *TextArea) SetSelStart(v int) { ta.selStart = v }

// GetSelEnd returns the selection end rune index. Used for testing.
func (ta *TextArea) GetSelEnd() int { return ta.selEnd }

// SetSelEnd sets the selection end directly. Used for testing.
func (ta *TextArea) SetSelEnd(v int) { ta.selEnd = v }

// GetScrollY returns the current vertical scroll offset. Used for testing.
func (ta *TextArea) GetScrollY() float64 { return ta.scrollY }

// ScrollBar returns the textarea's scrollbar widget. Used for testing.
func (ta *TextArea) ScrollBar() *ScrollBar { return ta.scrollbar }

// ClearSelectionForTest calls the internal clearSelection method. Used for testing.
func (ta *TextArea) ClearSelectionForTest() { ta.clearSelection() }

// DeleteSelectionForTest calls the internal deleteSelection method. Used for testing.
func (ta *TextArea) DeleteSelectionForTest() { ta.deleteSelection() }

// GetVisualLinesForTest returns the current visual lines as public VisualLine structs.
// Used for testing.
func (ta *TextArea) GetVisualLinesForTest() []VisualLine {
	vlines := ta.getVisualLines()
	result := make([]VisualLine, len(vlines))
	for i, v := range vlines {
		result[i] = VisualLine{RuneStart: v.runeStart, RuneEnd: v.runeEnd}
	}
	return result
}

// CursorVisualLineColForTest calls cursorVisualLineCol. Used for testing.
func (ta *TextArea) CursorVisualLineColForTest(pos int) (line, col int) {
	vlines := ta.getVisualLines()
	return ta.cursorVisualLineCol(pos, vlines)
}

// MoveCursorRightShiftForTest calls moveCursorRightShift. Used for testing.
func (ta *TextArea) MoveCursorRightShiftForTest(shift bool) { ta.moveCursorRightShift(shift) }

// MoveCursorUpShiftForTest calls moveCursorUpShift. Used for testing.
func (ta *TextArea) MoveCursorUpShiftForTest(shift bool) { ta.moveCursorUpShift(shift) }

// MoveCursorHomeShiftForTest calls moveCursorHomeShift. Used for testing.
func (ta *TextArea) MoveCursorHomeShiftForTest(shift bool) { ta.moveCursorHomeShift(shift) }

// MoveCursorEndShiftForTest calls moveCursorEndShift. Used for testing.
func (ta *TextArea) MoveCursorEndShiftForTest(shift bool) { ta.moveCursorEndShift(shift) }

// MoveCursorPageUpShiftForTest calls moveCursorPageUpShift. Used for testing.
func (ta *TextArea) MoveCursorPageUpShiftForTest(shift bool) { ta.moveCursorPageUpShift(shift) }

// MoveCursorPageDownShiftForTest calls moveCursorPageDownShift. Used for testing.
func (ta *TextArea) MoveCursorPageDownShiftForTest(shift bool) {
	ta.moveCursorPageDownShift(shift)
}

// SelectWordAtCursorForTest calls selectWordAtCursor. Used for testing.
func (ta *TextArea) SelectWordAtCursorForTest() { ta.selectWordAtCursor() }
