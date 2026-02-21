package widget

import (
	"bytes"
	"image/color"
	"math"
	"strings"
	"sync"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// RichText renders multi-span styled text into a single offscreen image.
// Each span can have its own font, color, and outline; unset fields inherit
// from the RichText defaults. Word wrapping is supported across span boundaries.
type RichText struct {
	Component
	spans        []TextSpan
	source       *sg.FontFamily
	displaySize  float64
	color        sg.Color
	outline      *Outline
	align        sg.TextAlign
	wrapWidth    float64
	image        engine.Image
	sprite       *sg.Node
	dirty        bool
	onLinkClick  func(url string)
	headingScale [3]float64 // {h1, h2, h3} size multipliers
	linkRects    []linkRect // populated during Render
}

// linkRect stores a clickable region mapped during Render.
type linkRect struct {
	x, y, w, h float64
	url        string
}

// NewRichText creates a new RichText component with the given default font
// source and display size.
func NewRichText(name string, source *sg.FontFamily, displaySize float64) *RichText {
	rt := &RichText{
		source:       source,
		displaySize:  displaySize,
		dirty:        true,
		headingScale: [3]float64{2.0, 1.5, 1.2},
	}
	initComponent(&rt.Component, name)
	rt.color = rt.EffectiveTheme().RichText.Group(rt.Variant()).TextColor.Resolve(StateDefault)

	// Create the sprite node that will display the offscreen rendered text.
	rt.sprite = sg.NewSprite(name+"-sprite", sg.TextureRegion{})
	rt.node.AddChild(rt.sprite)

	rt.onThemeChange = func() { rt.applyThemeColors() }
	return rt
}

// SetOnLinkClick sets a callback invoked when a <link> region is clicked.
func (rt *RichText) SetOnLinkClick(fn func(url string)) {
	rt.onLinkClick = fn
}

// SetHeadingScale sets the size multipliers for h1, h2, and h3 headings.
func (rt *RichText) SetHeadingScale(h1, h2, h3 float64) {
	rt.headingScale = [3]float64{h1, h2, h3}
	rt.markContentDirty()
}

func (rt *RichText) applyThemeColors() {
	rt.color = rt.EffectiveTheme().RichText.Group(rt.Variant()).TextColor.Resolve(StateDefault)
	rt.markContentDirty()
}

// AddSpan appends a plain text span that inherits all styling from the
// RichText defaults. Returns rt for chaining.
func (rt *RichText) AddSpan(text string) *RichText {
	rt.spans = append(rt.spans, TextSpan{Text: text})
	rt.markContentDirty()
	return rt
}

// AddStyledSpan appends a span with explicit styling overrides.
// Pass nil for source or outline to inherit from the RichText defaults.
// Color is always set explicitly (ColorSet = true) so the caller's color is
// used even if it happens to be the zero value. Returns rt for chaining.
func (rt *RichText) AddStyledSpan(text string, source *sg.FontFamily, color sg.Color, outline *Outline) *RichText {
	rt.spans = append(rt.spans, TextSpan{
		Text:     text,
		Source:   source,
		Color:    color,
		ColorSet: true,
		Outline:  outline,
	})
	rt.markContentDirty()
	return rt
}

// AddTextSpan appends a fully configured TextSpan. Returns rt for chaining.
func (rt *RichText) AddTextSpan(span TextSpan) *RichText {
	rt.spans = append(rt.spans, span)
	rt.markContentDirty()
	return rt
}

// AddBoldSpan appends a bold text span with the given color. Returns rt for chaining.
func (rt *RichText) AddBoldSpan(text string, color sg.Color) *RichText {
	rt.spans = append(rt.spans, TextSpan{Text: text, Bold: true, Color: color, ColorSet: true})
	rt.markContentDirty()
	return rt
}

// AddItalicSpan appends an italic text span with the given color. Returns rt for chaining.
func (rt *RichText) AddItalicSpan(text string, color sg.Color) *RichText {
	rt.spans = append(rt.spans, TextSpan{Text: text, Italic: true, Color: color, ColorSet: true})
	rt.markContentDirty()
	return rt
}

// AddBoldItalicSpan appends a bold+italic text span with the given color. Returns rt for chaining.
func (rt *RichText) AddBoldItalicSpan(text string, color sg.Color) *RichText {
	rt.spans = append(rt.spans, TextSpan{Text: text, Bold: true, Italic: true, Color: color, ColorSet: true})
	rt.markContentDirty()
	return rt
}

// ClearSpans removes all spans. Returns rt for chaining.
func (rt *RichText) ClearSpans() *RichText {
	rt.spans = rt.spans[:0]
	rt.markContentDirty()
	return rt
}

// SetSpans replaces all spans at once. Returns rt for chaining.
func (rt *RichText) SetSpans(spans []TextSpan) *RichText {
	rt.spans = make([]TextSpan, len(spans))
	copy(rt.spans, spans)
	rt.markContentDirty()
	return rt
}

// SetWrapWidth sets the maximum pixel width before text wraps to the next line.
// A value of 0 disables wrapping.
func (rt *RichText) SetWrapWidth(w float64) {
	rt.wrapWidth = w
	rt.markContentDirty()
}

// SetAlign sets the text alignment mode.
func (rt *RichText) SetAlign(a sg.TextAlign) {
	rt.align = a
	rt.markContentDirty()
}

// SetColor sets the default text color for spans that do not override it.
func (rt *RichText) SetColor(c sg.Color) {
	rt.color = c
	rt.markContentDirty()
}

// SetOutline sets the default text outline for spans that do not override it.
func (rt *RichText) SetOutline(o *Outline) {
	rt.outline = o
	rt.markContentDirty()
}

// SetMarkup parses XML-like markup and replaces all spans.
func (rt *RichText) SetMarkup(markup string) error {
	spans, err := ParseMarkup(markup, rt.source, rt.displaySize, rt.headingScale)
	if err != nil {
		return err
	}
	rt.SetSpans(spans)
	return nil
}

// Render composites all spans into the offscreen image and updates the sprite
// node. This should be called once per frame before drawing, typically in the
// scene's update function. It is a no-op if nothing has changed.
func (rt *RichText) Render() {
	if !rt.dirty {
		return
	}
	rt.dirty = false

	lines := rt.layoutLines()
	if len(lines) == 0 {
		if rt.image != nil {
			rt.image.Clear()
		}
		rt.sprite.SetScale(0, 0)
		return
	}

	// Measure total dimensions.
	var totalW, totalH float64
	for _, line := range lines {
		if line.width > totalW {
			totalW = line.width
		}
		totalH += line.height
	}

	// When wrapWidth is set, use it as the image width so that center/right
	// alignment offsets are computed against the full container width.
	alignW := totalW
	if rt.wrapWidth > 0 && rt.wrapWidth > totalW {
		alignW = rt.wrapWidth
	}

	imgW := int(math.Ceil(alignW)) + 1
	imgH := int(math.Ceil(totalH)) + 1
	if imgW < 1 {
		imgW = 1
	}
	if imgH < 1 {
		imgH = 1
	}

	// Reuse or create offscreen image.
	if rt.image != nil {
		b := rt.image.Bounds()
		if b.Dx() < imgW || b.Dy() < imgH {
			rt.image.Deallocate()
			rt.image = engine.NewImage(imgW, imgH)
		} else {
			rt.image.Clear()
		}
	} else {
		rt.image = engine.NewImage(imgW, imgH)
	}

	// Draw each fragment and collect link rects.
	rt.linkRects = rt.linkRects[:0]
	y := 0.0
	for _, line := range lines {
		x := line.indent + rt.alignOffset(line.width-line.indent, alignW-line.indent)
		for _, frag := range line.fragments {
			rt.drawFragment(frag, x, y)
			if frag.linkURL != "" {
				rt.linkRects = append(rt.linkRects, linkRect{
					x: x, y: y, w: frag.width, h: frag.height, url: frag.linkURL,
				})
			}
			x += frag.width
		}
		y += line.height
	}

	// Wire up link click handling on the sprite node if we have links.
	if len(rt.linkRects) > 0 && rt.onLinkClick != nil {
		rects := rt.linkRects
		cb := rt.onLinkClick
		rt.sprite.OnClick(func(ctx sg.ClickContext) {
			px, py := ctx.LocalX, ctx.LocalY
			for _, lr := range rects {
				if px >= lr.x && px <= lr.x+lr.w && py >= lr.y && py <= lr.y+lr.h {
					cb(lr.url)
					return
				}
			}
		})
	}

	// Apply the rendered image to the sprite node.
	rt.sprite.SetCustomImage(rt.image)
	rt.sprite.SetScale(1, 1)

	// Update component dimensions.
	rt.Width = alignW
	rt.Height = totalH
}

// Dispose releases the offscreen image and disposes the component tree.
func (rt *RichText) Dispose() {
	if rt.image != nil {
		rt.image.Deallocate()
		rt.image = nil
	}
	rt.Component.Dispose()
}

// --- Internal types and helpers ---

// lineFragment is a piece of text on a line, from a single span, with resolved
// styling and measured dimensions.
type lineFragment struct {
	text          string
	font          *sg.FontFamily
	bold          bool
	italic        bool
	color         sg.Color
	outline       *Outline
	width         float64
	height        float64
	underline     bool
	strikethrough bool
	sizeOverride  float64
	linkURL       string
	ascent        float64
}

// layoutLine is a single line of laid-out text fragments.
type layoutLine struct {
	fragments []lineFragment
	width     float64
	height    float64
	indent    float64 // left indent for this line
}

// layoutLines performs word-wrap layout across all spans, producing lines of
// fragments. This is the core layout algorithm.
func (rt *RichText) layoutLines() []layoutLine {
	type token struct {
		text          string
		font          *sg.FontFamily
		bold          bool
		italic        bool
		color         sg.Color
		outline       *Outline
		width         float64
		height        float64
		ascent        float64
		newline       bool
		underline     bool
		strikethrough bool
		sizeOverride  float64
		linkURL       string
		indent        float64
	}

	var tokens []token

	for _, span := range rt.spans {
		if span.Text == "" {
			continue
		}

		font := rt.resolveFont(span)
		color := rt.resolveColor(span)
		outline := rt.resolveOutline(span)
		size := span.SizeOverride
		bold := span.Bold
		italic := span.Italic

		parts := splitWords(span.Text)
		for pi, part := range parts {
			if part == "\n" {
				_, h, asc := rt.measureGoTextAt(font, "M", size, bold, italic)
				tokens = append(tokens, token{
					font:         font,
					bold:         bold,
					italic:       italic,
					color:        color,
					outline:      outline,
					newline:      true,
					height:       h,
					ascent:       asc,
					sizeOverride: size,
				})
				continue
			}
			w, h, asc := rt.measureGoTextAt(font, part, size, bold, italic)
			tok := token{
				text:          part,
				font:          font,
				bold:          bold,
				italic:        italic,
				color:         color,
				outline:       outline,
				width:         w,
				height:        h,
				ascent:        asc,
				underline:     span.Underline,
				strikethrough: span.Strikethrough,
				sizeOverride:  size,
				linkURL:       span.LinkURL,
			}
			// Apply indent only on the first token of the span.
			if pi == 0 && span.Indent > 0 {
				tok.indent = span.Indent
			}
			tokens = append(tokens, tok)
		}
	}

	if len(tokens) == 0 {
		return nil
	}

	// Lay out tokens into lines with word wrapping.
	var lines []layoutLine
	var curLine layoutLine
	curX := 0.0
	curIndent := 0.0     // left margin for the current line group
	hangingIndent := 0.0 // wrap continuations start here (after bullet/prefix)

	for _, tok := range tokens {
		if tok.newline {
			if curLine.height == 0 {
				curLine.height = tok.height
			}
			lines = append(lines, curLine)
			curLine = layoutLine{}
			curX = 0
			curIndent = 0
			hangingIndent = 0
			continue
		}

		// If this token sets an indent, apply it to the line.
		if tok.indent > 0 {
			curIndent = tok.indent
			curLine.indent = curIndent
			curX = curIndent
			// The hanging indent will be set after this token is placed
			// (indent + this token's width), so wrapped text aligns with
			// the content after the bullet/prefix.
			hangingIndent = 0 // will be computed below
		}

		// Check if this token exceeds wrapWidth and we're not at line start.
		effectiveStart := hangingIndent
		if effectiveStart == 0 {
			effectiveStart = curIndent
		}
		if rt.wrapWidth > 0 && curX > effectiveStart && curX+tok.width > rt.wrapWidth {
			lines = append(lines, curLine)
			wrapX := hangingIndent
			if wrapX == 0 {
				wrapX = curIndent
			}
			curLine = layoutLine{indent: wrapX}
			curX = wrapX

			// Strip leading space from the token on the new line.
			trimmed := strings.TrimLeft(tok.text, " ")
			if trimmed == "" {
				continue
			}
			if trimmed != tok.text {
				tok.text = trimmed
				tok.width, _, _ = rt.measureGoTextAt(tok.font, trimmed, tok.sizeOverride, tok.bold, tok.italic)
			}
		}

		curLine.fragments = append(curLine.fragments, lineFragment{
			text:          tok.text,
			font:          tok.font,
			bold:          tok.bold,
			italic:        tok.italic,
			color:         tok.color,
			outline:       tok.outline,
			width:         tok.width,
			height:        tok.height,
			underline:     tok.underline,
			strikethrough: tok.strikethrough,
			sizeOverride:  tok.sizeOverride,
			linkURL:       tok.linkURL,
			ascent:        tok.ascent,
		})
		curX += tok.width
		curLine.width = curX
		if tok.height > curLine.height {
			curLine.height = tok.height
		}

		// After placing an indent token (the bullet/number prefix),
		// set the hanging indent so wraps align with text after it.
		if tok.indent > 0 && hangingIndent == 0 {
			hangingIndent = curX
		}
	}

	if len(curLine.fragments) > 0 {
		lines = append(lines, curLine)
	}

	return lines
}

// splitWords splits text into tokens where each token is either a word
// (possibly with trailing space) or a newline marker.
func splitWords(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			if i > start {
				result = append(result, s[start:i])
			}
			result = append(result, "\n")
			start = i + 1
		} else if s[i] == ' ' {
			j := i + 1
			for j < len(s) && s[j] == ' ' {
				j++
			}
			result = append(result, s[start:j])
			start = j
			i = j - 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

// resolveFont returns the effective font for a span.
func (rt *RichText) resolveFont(span TextSpan) *sg.FontFamily {
	src := span.Source
	if src == nil {
		src = rt.source
	}
	if src == nil {
		return nil
	}
	return src
}

// resolveColor returns the effective color for a span.
func (rt *RichText) resolveColor(span TextSpan) sg.Color {
	if span.ColorSet {
		return span.Color
	}
	return rt.color
}

// resolveOutline returns the effective outline for a span.
func (rt *RichText) resolveOutline(span TextSpan) *Outline {
	if span.Outline != nil {
		return span.Outline
	}
	return rt.outline
}

// alignOffset computes the x-offset for a line based on alignment.
func (rt *RichText) alignOffset(lineWidth, totalWidth float64) float64 {
	switch rt.align {
	case sg.TextAlignCenter:
		return (totalWidth - lineWidth) / 2
	case sg.TextAlignRight:
		return totalWidth - lineWidth
	default:
		return 0
	}
}

// goTextFaceCache caches GoTextFaceSource instances keyed by raw TTF data
// pointer to avoid re-parsing the same TTF bytes repeatedly.
var (
	goTextFaceCacheMu sync.Mutex
	goTextFaceCache   = map[*byte]*engine.GoTextFaceSource{}
)

// getGoTextFaceSource returns a cached GoTextFaceSource for the given TTF data.
func getGoTextFaceSource(ttfData []byte) *engine.GoTextFaceSource {
	if len(ttfData) == 0 {
		return nil
	}
	key := &ttfData[0]
	goTextFaceCacheMu.Lock()
	defer goTextFaceCacheMu.Unlock()
	if src, ok := goTextFaceCache[key]; ok {
		return src
	}
	src, err := engine.NewGoTextFaceSource(bytes.NewReader(ttfData))
	if err != nil {
		return nil
	}
	goTextFaceCache[key] = src
	return src
}

// goTextFaceAt returns a GoTextFace for the given FontFamily, style, and size,
// or nil if the font doesn't carry TTF data.
func (rt *RichText) goTextFaceAt(f *sg.FontFamily, sizeOverride float64, bold, italic bool) *engine.GoTextFace {
	if f == nil {
		return nil
	}
	ttfData := f.TTFData(bold, italic)
	if ttfData == nil {
		return nil
	}
	src := getGoTextFaceSource(ttfData)
	if src == nil {
		return nil
	}
	size := rt.displaySize
	if sizeOverride > 0 {
		size = sizeOverride
	}
	return &engine.GoTextFace{
		Source: src,
		Size:   size,
	}
}

// measureGoTextAt measures text at the given size override (0 = default),
// returning width, height, and ascent.
func (rt *RichText) measureGoTextAt(f *sg.FontFamily, s string, sizeOverride float64, bold, italic bool) (width, height, ascent float64) {
	face := rt.goTextFaceAt(f, sizeOverride, bold, italic)
	if face == nil {
		size := rt.displaySize
		if sizeOverride > 0 {
			size = sizeOverride
		}
		w, h := measureDisplayStyled(f, s, size, bold, italic)
		return w, h, h * 0.75 // approximate ascent
	}
	w := engine.TextAdvance(s, face)
	m := face.Metrics()
	h := m.HAscent + m.HDescent
	return w, h, m.HAscent
}

// drawFragment renders a single text fragment onto the offscreen image at (x, y).
func (rt *RichText) drawFragment(frag lineFragment, x, y float64) {
	if frag.text == "" {
		return
	}

	face := rt.goTextFaceAt(frag.font, frag.sizeOverride, frag.bold, frag.italic)
	if face == nil {
		return
	}

	// Outline: draw text at 8 cardinal offsets in outline color, then fill on top.
	outline := frag.outline
	if outline != nil && outline.Thickness > 0 {
		oc := outline.Color
		th := outline.Thickness
		offsets := [8][2]float64{
			{-th, 0}, {th, 0}, {0, -th}, {0, th},
			{-th, -th}, {th, -th}, {-th, th}, {th, th},
		}
		for _, off := range offsets {
			op := &engine.DrawOptions{}
			op.GeoM.Translate(x+off[0], y+off[1])
			op.ColorScale.Scale(float32(oc.R()), float32(oc.G()), float32(oc.B()), float32(oc.A()))
			engine.TextDraw(rt.image, frag.text, face, op)
		}
	}

	// Fill text.
	op := &engine.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.Scale(
		float32(frag.color.R()),
		float32(frag.color.G()),
		float32(frag.color.B()),
		float32(frag.color.A()),
	)
	engine.TextDraw(rt.image, frag.text, face, op)

	// Underline / strikethrough decorations.
	if frag.underline || frag.strikethrough {
		rt.drawDecorations(frag, x, y)
	}
}

// drawDecorations draws underline and/or strikethrough lines for a fragment.
func (rt *RichText) drawDecorations(frag lineFragment, x, y float64) {
	c := frag.color
	clr := colorToRGBA(c)

	if frag.underline {
		ly := y + frag.ascent + 2
		engine.StrokeLine(rt.image, float32(x), float32(ly), float32(x+frag.width), float32(ly), 1, clr, false)
	}
	if frag.strikethrough {
		ly := y + frag.ascent*0.65
		engine.StrokeLine(rt.image, float32(x), float32(ly), float32(x+frag.width), float32(ly), 1, clr, false)
	}
}

// colorToRGBA converts a sg.Color to a standard color.RGBA.
func colorToRGBA(c sg.Color) color.RGBA {
	return color.RGBA{
		R: uint8(c.R() * 255),
		G: uint8(c.G() * 255),
		B: uint8(c.B() * 255),
		A: uint8(c.A() * 255),
	}
}

// markContentDirty flags the rich text for re-layout and re-render.
func (rt *RichText) markContentDirty() {
	rt.dirty = true
	rt.MarkDrawDirty()
}

// LineFragmentForTest is a public view of a lineFragment for testing.
type LineFragmentForTest struct {
	SizeOverride float64
	Underline    bool
	LinkURL      string
}

// LayoutLineForTest is a public view of a layoutLine for testing.
type LayoutLineForTest struct {
	Fragments []LineFragmentForTest
	Height    float64
}

// LayoutLinesForTest calls layoutLines and returns public views. Used for testing.
func (rt *RichText) LayoutLinesForTest() []LayoutLineForTest {
	internal := rt.layoutLines()
	result := make([]LayoutLineForTest, len(internal))
	for i, line := range internal {
		frags := make([]LineFragmentForTest, len(line.fragments))
		for j, frag := range line.fragments {
			frags[j] = LineFragmentForTest{
				SizeOverride: frag.sizeOverride,
				Underline:    frag.underline,
				LinkURL:      frag.linkURL,
			}
		}
		result[i] = LayoutLineForTest{Fragments: frags, Height: line.height}
	}
	return result
}

// Spans returns the current spans. Used for testing.
func (rt *RichText) Spans() []TextSpan { return rt.spans }

// Color returns the default text color. Used for testing.
func (rt *RichText) Color() sg.Color { return rt.color }

// SpriteNode returns the internal sprite node. Used for testing.
func (rt *RichText) SpriteNode() *sg.Node { return rt.sprite }

// ImageForTest returns the internal offscreen image. Used for testing.
func (rt *RichText) ImageForTest() engine.Image { return rt.image }

// Dirty returns whether the rich text needs re-rendering. Used for testing.
func (rt *RichText) Dirty() bool { return rt.dirty }

// SetDirtyForTest sets the dirty flag. Used for testing.
func (rt *RichText) SetDirtyForTest(v bool) { rt.dirty = v }

// ResolveOutlineForTest calls resolveOutline on the given span. Used for testing.
func (rt *RichText) ResolveOutlineForTest(span TextSpan) *Outline { return rt.resolveOutline(span) }

// ResolveColorForTest calls resolveColor on the given span. Used for testing.
func (rt *RichText) ResolveColorForTest(span TextSpan) sg.Color { return rt.resolveColor(span) }

// ResolveFontForTest calls resolveFont on the given span. Used for testing.
func (rt *RichText) ResolveFontForTest(span TextSpan) *sg.FontFamily { return rt.resolveFont(span) }

// OnLinkClickForTest returns the onLinkClick callback. Used for testing.
func (rt *RichText) OnLinkClickForTest() func(string) { return rt.onLinkClick }

// HeadingScale returns the heading scale array. Used for testing.
func (rt *RichText) HeadingScale() [3]float64 { return rt.headingScale }
