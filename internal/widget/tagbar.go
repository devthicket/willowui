package widget

import (
	"strings"

	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
	"github.com/devthicket/willowui/internal/theme"
)

// tagEntry holds a single tag and its label text within a TagBar.
type tagEntry struct {
	tag  *Tag
	text string
}

// TagBar is a tag-input widget where the user types text and presses Space
// (or Enter) to create a Tag chip. Each tag shows a × to delete it. The
// TagBar owns the list of tag values and renders them using Tag internally.
type TagBar struct {
	Component
	input       *TextInput
	tags        []*tagEntry
	source      *sg.FontFamily
	font        *sg.FontFamily
	displaySize float64

	// prevInputValue tracks the input value from the previous frame so that
	// Backspace-to-delete-tag only fires when the input was already empty.
	prevInputValue string

	// Callbacks.
	onChange    func(tags []string)
	onAddTag    func(text string)
	onRemoveTag func(text string)
}

// NewTagBar creates a TagBar with the given name, font source, and display size.
func NewTagBar(name string, source *sg.FontFamily, displaySize float64) *TagBar {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	tb := &TagBar{
		source:      source,
		font:        font,
		displaySize: displaySize,
	}
	initComponent(&tb.Component, name)
	tb.initBackground(name)
	tb.initBorder(name)

	// Create inner TextInput.
	tb.input = NewTextInput(name+"-input", source, displaySize)
	tb.input.embedded = true
	tb.input.hideBackground()
	tb.input.hideBorder()
	tb.input.hideFocusRing()
	tb.node.AddChild(tb.input.Node())

	// Reject space characters — TagBar intercepts them to create tags.
	tb.input.SetCharFilter(func(ch rune) bool {
		return ch != ' '
	})

	// Submit on Enter also creates a tag.
	tb.input.SetOnSubmit(func(v string) {
		text := strings.TrimSpace(v)
		if text != "" {
			tb.addTag(text)
			tb.input.SetValue("")
		}
	})

	// Focus forwarding: clicking the TagBar container focuses the input.
	tb.node.OnPointerDown(func(ctx sg.PointerContext) {
		if !tb.enabled {
			return
		}
		tb.pressed = true
		tb.bubbleActivation()
		DefaultFocusManager.SetFocus(&tb.input.Component)
	})

	tb.input.onFocusChange = func(focused bool) {
		tb.UpdateVisuals()
	}

	tb.onVisualStateChange = func() { tb.UpdateVisuals() }
	tb.onThemeChange = func() { tb.UpdateVisuals() }

	tb.SetCursorShape(engine.CursorShapeText)

	defaultW := 300.0
	defaultH := displayLineHeight(font, displaySize) + 12
	tb.SetSize(defaultW, defaultH)
	tb.UpdateVisuals()

	tb.node.OnUpdate = func(dt float64) {
		tb.update()
	}

	return tb
}

// ---------------------------------------------------------------------------
// Tag management API
// ---------------------------------------------------------------------------

// Tags returns the current list of tag strings.
func (tb *TagBar) Tags() []string {
	out := make([]string, len(tb.tags))
	for i, e := range tb.tags {
		out[i] = e.text
	}
	return out
}

// SetTags replaces the entire tag list.
func (tb *TagBar) SetTags(tags []string) {
	// Remove all existing tags.
	for len(tb.tags) > 0 {
		tb.removeTagAt(len(tb.tags) - 1)
	}
	for _, t := range tags {
		tb.addTag(t)
	}
}

// AddTag adds a tag with the given text.
func (tb *TagBar) AddTag(text string) {
	tb.addTag(text)
}

// RemoveTagAt removes the tag at the given index.
func (tb *TagBar) RemoveTagAt(idx int) {
	if idx < 0 || idx >= len(tb.tags) {
		return
	}
	tb.removeTagAt(idx)
}

// ---------------------------------------------------------------------------
// Callbacks
// ---------------------------------------------------------------------------

// SetOnChange sets a callback fired whenever the tag list changes.
func (tb *TagBar) SetOnChange(fn func(tags []string)) { tb.onChange = fn }

// SetOnAddTag sets a callback fired when a new tag is added.
func (tb *TagBar) SetOnAddTag(fn func(text string)) { tb.onAddTag = fn }

// SetOnRemoveTag sets a callback fired when a tag is removed.
func (tb *TagBar) SetOnRemoveTag(fn func(text string)) { tb.onRemoveTag = fn }

// ---------------------------------------------------------------------------
// Input access
// ---------------------------------------------------------------------------

// SetPlaceholder sets the placeholder text shown when the input is empty.
func (tb *TagBar) SetPlaceholder(text string) {
	tb.input.SetPlaceholder(text)
}

// ---------------------------------------------------------------------------
// Layout and visuals
// ---------------------------------------------------------------------------

// SetSize sets the outer dimensions of the TagBar.
func (tb *TagBar) SetSize(w, h float64) {
	tb.Width = w
	tb.Height = h
	tb.resizeBackground(w, h)
	tb.resizeBorder(w, h)
	tb.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	tb.updateLayout()
	tb.MarkLayoutDirty()
}

// UpdateVisuals applies theme styling to the TagBar and its children.
func (tb *TagBar) UpdateVisuals() {
	tb.state = computeState(tb.enabled, tb.input.IsFocused(), tb.hovered, tb.pressed)
	group := tb.effectiveGroup()

	cr := resolveCornerRadius(group.CornerRadius, tb.Height)
	tb.applyCornerRadius(cr)

	bg := group.Background.Resolve(tb.state)
	tb.applyBackground(bg)
	tb.applyBorder(group.Border.Resolve(tb.state), group.BorderWidth, bg)
	tb.applyFocusRing(group.FocusColor.Resolve(tb.state), group.FocusRingWidth)

	// Update tag visuals.
	for _, entry := range tb.tags {
		entry.tag.UpdateVisuals()
	}

	tb.MarkDrawDirty()
}

// Dispose cleans up all tags and the inner TextInput.
func (tb *TagBar) Dispose() {
	for _, entry := range tb.tags {
		entry.tag.Dispose()
	}
	tb.tags = nil
	tb.input.Dispose()
	tb.Component.Dispose()
}

// ---------------------------------------------------------------------------
// Internal
// ---------------------------------------------------------------------------

func (tb *TagBar) effectiveGroup() *theme.TagBarGroup {
	return tb.EffectiveTheme().TagBar.Group(tb.Variant())
}

func (tb *TagBar) addTag(text string) {
	tag := NewTag(tb.node.Name+"_tag_"+text, tb.source, tb.displaySize)
	tag.SetText(text)
	tag.SetRemovable(true)

	tag.SetOnRemove(func() {
		tb.removeTagByPtr(tag)
	})

	// Inherit theme from TagBar.
	if th := tb.theme; th != nil {
		tag.SetTheme(th)
	}

	tb.node.AddChild(tag.Node())
	tb.tags = append(tb.tags, &tagEntry{tag: tag, text: text})
	tag.SizeToContent()
	tb.updateLayout()

	if tb.onAddTag != nil {
		tb.onAddTag(text)
	}
	if tb.onChange != nil {
		tb.onChange(tb.Tags())
	}
}

func (tb *TagBar) removeTagAt(idx int) {
	entry := tb.tags[idx]
	tb.node.RemoveChild(entry.tag.Node())
	entry.tag.Dispose()

	copy(tb.tags[idx:], tb.tags[idx+1:])
	tb.tags = tb.tags[:len(tb.tags)-1]
	tb.updateLayout()

	if tb.onRemoveTag != nil {
		tb.onRemoveTag(entry.text)
	}
	if tb.onChange != nil {
		tb.onChange(tb.Tags())
	}
}

func (tb *TagBar) removeTagByPtr(tag *Tag) {
	for i, entry := range tb.tags {
		if entry.tag == tag {
			tb.removeTagAt(i)
			return
		}
	}
}

func (tb *TagBar) updateLayout() {
	group := tb.effectiveGroup()
	pad := group.Padding
	spacing := group.Spacing

	x := pad.Left
	y := pad.Top
	lineHeight := displayLineHeight(tb.font, tb.displaySize) + 4 // tag height
	availW := tb.Width - pad.Left - pad.Right

	for _, entry := range tb.tags {
		tw := entry.tag.Width
		if x > pad.Left && x+tw > tb.Width-pad.Right {
			// Wrap to next line.
			x = pad.Left
			y += lineHeight + spacing
		}
		entry.tag.SetPosition(x, y)
		x += tw + spacing
	}

	// Position the input after the last tag.
	minInputW := 60.0
	inputW := availW - (x - pad.Left)
	if inputW < minInputW && x > pad.Left {
		// Wrap input to next line.
		x = pad.Left
		y += lineHeight + spacing
		inputW = availW
	}
	inputH := displayLineHeight(tb.font, tb.displaySize) + 4
	tb.input.SetSize(inputW, inputH)
	tb.input.Node().SetPosition(x, y)

	// Override the embedded input's content position to remove its own
	// internal padding — the TagBar provides padding at its level.
	lineH := displayLineHeight(tb.font, tb.displaySize)
	vCenter := (inputH - lineH) / 2
	tb.input.content.SetPosition(0, vCenter)
	tb.input.updateContentMask(inputW, lineH)
}

func (tb *TagBar) update() {
	if !tb.input.IsFocused() {
		tb.prevInputValue = tb.input.Value()
		return
	}

	// Space key: create tag from current input.
	if core.IsKeyJustPressed(engine.KeySpace) {
		text := strings.TrimSpace(tb.input.Value())
		if text != "" {
			tb.addTag(text)
			tb.input.SetValue("")
		}
	}

	// Backspace on empty input: remove last tag.
	// Only remove if the input was already empty in the previous frame,
	// to avoid removing a tag when the user just deleted the last character.
	if core.IsKeyJustPressed(engine.KeyBackspace) &&
		tb.prevInputValue == "" &&
		tb.input.Value() == "" &&
		len(tb.tags) > 0 {
		tb.removeTagAt(len(tb.tags) - 1)
	}

	tb.prevInputValue = tb.input.Value()
}
