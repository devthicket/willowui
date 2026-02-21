package widget

import "github.com/devthicket/willowui/internal/sg"

// fontScaleStyled returns the scale factor to apply to a text node so that the
// native atlas font renders at the requested displaySize.
func fontScaleStyled(font *sg.FontFamily, displaySize float64, bold, italic bool) float64 {
	if font == nil || displaySize <= 0 {
		return 1.0
	}
	native := font.LineHeight(displaySize, bold, italic)
	if native <= 0 {
		return 1.0
	}
	return displaySize / native
}

// fontScale is a convenience wrapper with no bold/italic.
func fontScale(font *sg.FontFamily, displaySize float64) float64 {
	return fontScaleStyled(font, displaySize, false, false)
}

// displayLineHeight returns the effective line height in display pixels.
func displayLineHeight(font *sg.FontFamily, displaySize float64) float64 {
	if displaySize > 0 {
		return displaySize
	}
	if font == nil {
		return 0
	}
	return font.LineHeight(displaySize, false, false)
}

// measureDisplayStyled measures text in display pixels using the specified style.
func measureDisplayStyled(font *sg.FontFamily, text string, displaySize float64, bold, italic bool) (float64, float64) {
	if font == nil {
		return 0, 0
	}
	w, h := font.MeasureString(text, displaySize, bold, italic)
	s := fontScaleStyled(font, displaySize, bold, italic)
	return w * s, h * s
}

// measureDisplay measures text in display pixels (regular style).
func measureDisplay(font *sg.FontFamily, text string, displaySize float64) (float64, float64) {
	return measureDisplayStyled(font, text, displaySize, false, false)
}

// Label is a text display component. It wraps a willow text node and
// supports reactive text binding via BindText.
type Label struct {
	Component
	textNode    *sg.Node       // NodeTypeText child
	source      *sg.FontFamily // font source (for bold/italic resolution)
	font        *sg.FontFamily // current resolved font (kept for measurement)
	bold        bool
	italic      bool
	displaySize float64      // display font size (0 = native atlas size)
	text        *Ref[string] // reactive text binding (optional)
	watch       WatchHandle  // for reactive binding cleanup
}

// NewLabel creates a Label with the given name, initial text, font source, and
// display size. If displaySize is 0, the native atlas size is used.
func NewLabel(name string, text string, source *sg.FontFamily, displaySize float64) *Label {
	l := &Label{
		source:      source,
		displaySize: displaySize,
	}
	if source != nil {
		l.font = source
	}
	initComponent(&l.Component, name)

	l.textNode = sg.NewText(name+"-text", text, l.font)
	l.textNode.TextBlock.FontSize = displaySize
	l.textNode.TextBlock.Color = l.EffectiveTheme().Label.Group(l.Variant()).TextColor.Resolve(StateDefault)
	l.node.AddChild(l.textNode)

	l.measureAndResize()

	l.onThemeChange = func() { l.applyThemeColors() }
	return l
}

func (l *Label) applyThemeColors() {
	l.textNode.SetTextColor(l.EffectiveTheme().Label.Group(l.Variant()).TextColor.Resolve(StateDefault))
	l.MarkDrawDirty()
}

// Text returns the current text content.
func (l *Label) Text() string {
	return l.textNode.TextBlock.Content
}

// SetText updates the displayed text and re-measures the label dimensions.
func (l *Label) SetText(t string) {
	if l.textNode.TextBlock.Content == t {
		return
	}
	l.textNode.SetContent(t)
	l.measureAndResize()
	l.MarkDrawDirty()
}

// Font returns the current resolved font.
func (l *Label) Font() *sg.FontFamily {
	return l.font
}

// SetFont replaces the font source used for rendering and measurement.
func (l *Label) SetFont(source *sg.FontFamily) {
	l.source = source
	if source != nil {
		l.font = source
	} else {
		l.font = nil
	}
	l.textNode.SetFont(l.font)
	l.measureAndResize()
	l.MarkDrawDirty()
}

// SetBold sets bold rendering and updates the TextBlock style.
func (l *Label) SetBold(bold bool) {
	l.bold = bold
	l.textNode.TextBlock.Bold = bold
	l.textNode.TextBlock.Invalidate()
	l.measureAndResize()
	l.MarkDrawDirty()
}

// SetItalic sets italic rendering and updates the TextBlock style.
func (l *Label) SetItalic(italic bool) {
	l.italic = italic
	l.textNode.TextBlock.Italic = italic
	l.textNode.TextBlock.Invalidate()
	l.measureAndResize()
	l.MarkDrawDirty()
}

// SetFontSize sets the display size and re-measures.
func (l *Label) SetFontSize(size float64) {
	l.displaySize = size
	l.textNode.SetFontSize(size)
	if l.source != nil {
		l.font = l.source
		l.textNode.SetFont(l.font)
	}
	l.measureAndResize()
	l.MarkDrawDirty()
}

// SetColor sets the text color.
func (l *Label) SetColor(c sg.Color) {
	l.textNode.SetTextColor(c)
	l.MarkDrawDirty()
}

// SetAlign sets the text alignment.
func (l *Label) SetAlign(a sg.TextAlign) {
	l.textNode.SetAlign(a)
	l.MarkDrawDirty()
}

// SetSharpness tightens the SDF edge rendering. 0.0 = default, 1.0 = sharpest.
// Values around 0.5–0.7 give crisp edges without aliasing on most displays.
func (l *Label) SetSharpness(s float64) {
	l.textNode.TextBlock.Sharpness = s
	l.textNode.TextBlock.Invalidate()
	l.MarkDrawDirty()
}

// SetWrapWidth sets the maximum line width for word wrapping.
// Zero means no wrapping.
func (l *Label) SetWrapWidth(w float64) {
	l.textNode.SetWrapWidth(w)
	l.measureAndResize()
	l.MarkDrawDirty()
}

// BindText binds the label to a reactive Ref[string]. The label text
// updates automatically whenever the ref changes. Any previous binding
// is stopped first. The label is immediately set to the ref's current value.
func (l *Label) BindText(ref *Ref[string]) {
	l.text = ref
	bindRef(&l.watch, ref, l.SetText)
}

// Dispose stops any reactive watch and disposes the underlying component.
func (l *Label) Dispose() {
	l.watch.Stop()
	l.Component.Dispose()
}

// TextNode returns the underlying willow text node. Used for testing.
func (l *Label) TextNode() *sg.Node { return l.textNode }

// measureAndResize updates Width and Height based on text measurement.
func (l *Label) measureAndResize() {
	if l.font == nil {
		return
	}
	w, h := measureDisplay(l.font, l.textNode.TextBlock.Content, l.displaySize)
	l.Width = w
	l.Height = h
	l.MarkLayoutDirty()
}
