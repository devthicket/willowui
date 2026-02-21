package integration

import (
	"testing"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

// --- Constructor ---

func TestNewRichText(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("test-rt", font, 0)

	if rt.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if rt.Name() != "test-rt" {
		t.Errorf("Name() = %q, want %q", rt.Name(), "test-rt")
	}
	if len(rt.Spans()) != 0 {
		t.Errorf("spans should be empty, got %d", len(rt.Spans()))
	}
	if !rt.Dirty() {
		t.Error("new RichText should start dirty")
	}
}

// --- Single span ---

func TestSingleSpanPreservesText(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.AddSpan("hello world")

	if len(rt.Spans()) != 1 {
		t.Fatalf("expected 1 span, got %d", len(rt.Spans()))
	}
	if rt.Spans()[0].Text != "hello world" {
		t.Errorf("span text = %q, want %q", rt.Spans()[0].Text, "hello world")
	}
}

func TestSingleSpanInheritsFont(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.AddSpan("test")

	if rt.Spans()[0].Source != nil {
		t.Error("span source should be nil (inherited)")
	}
}

// --- Multiple spans with different colors ---

func TestMultipleSpansColors(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)

	red := willow.RGBA(1, 0, 0, 1)
	blue := willow.RGBA(0, 0, 1, 1)

	rt.AddStyledSpan("red text", nil, red, nil)
	rt.AddStyledSpan("blue text", nil, blue, nil)

	if len(rt.Spans()) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(rt.Spans()))
	}
	if !rt.Spans()[0].ColorSet {
		t.Error("first span should have ColorSet=true")
	}
	if rt.Spans()[0].Color != red {
		t.Errorf("first span color = %v, want red", rt.Spans()[0].Color)
	}
	if rt.Spans()[1].Color != blue {
		t.Errorf("second span color = %v, want blue", rt.Spans()[1].Color)
	}
}

// --- Word wrapping ---

func TestWordWrapping(t *testing.T) {
	font := newTestFont()
	ds := float64(16) // display size
	rt := ui.NewRichText("rt", font, ds)
	// Measure "hello " at display size to find a wrap width that forces wrapping.
	helloW, _ := font.MeasureString("hello ", ds, false, false)
	scale := ds / font.LineHeight(ds, false, false)
	wrapW := helloW*scale + 1 // fits "hello " but not "hello world"
	rt.SetWrapWidth(wrapW)
	rt.AddSpan("hello world")

	lines := rt.LayoutLinesForTest()
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines with wrapping, got %d", len(lines))
	}
}

func TestWordWrappingAcrossSpans(t *testing.T) {
	font := newTestFont()
	ds := float64(16) // display size
	rt := ui.NewRichText("rt", font, ds)
	// "hello " + "world" — word boundary crosses span boundary.
	helloW, _ := font.MeasureString("hello ", ds, false, false)
	scale := ds / font.LineHeight(ds, false, false)
	wrapW := helloW*scale + 1 // fits "hello " but not "hello world"
	rt.SetWrapWidth(wrapW)
	rt.AddSpan("hello ")
	rt.AddSpan("world")

	lines := rt.LayoutLinesForTest()
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}
}

// --- Mixed font sizes ---

func TestMixedFontSizesLineHeight(t *testing.T) {
	font := newTestFont()

	// Use a base displaySize of 16 and add a span with a larger size via markup.
	rt := ui.NewRichText("rt", font, 16)
	rt.SetMarkup("small <size value=\"32\">large</size>")

	lines := rt.LayoutLinesForTest()
	if len(lines) == 0 {
		t.Fatal("expected at least 1 line")
	}
	// The line height should be at least 32 (the taller span's display size).
	if lines[0].Height < 32 {
		t.Errorf("line height = %f, want >= 32 (tallest span)", lines[0].Height)
	}
}

// --- Outline inheritance ---

func TestOutlineInheritance(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	defaultOutline := &ui.Outline{Color: willow.RGBA(0, 0, 0, 1), Thickness: 2}
	rt.SetOutline(defaultOutline)
	rt.AddSpan("inherited")

	resolved := rt.ResolveOutlineForTest(rt.Spans()[0])
	if resolved != defaultOutline {
		t.Error("span without outline should inherit from RichText")
	}
}

func TestOutlinePerSpan(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	defaultOutline := &ui.Outline{Color: willow.RGBA(0, 0, 0, 1), Thickness: 2}
	rt.SetOutline(defaultOutline)

	spanOutline := &ui.Outline{Color: willow.RGBA(1, 0, 0, 1), Thickness: 4}
	rt.AddStyledSpan("custom", nil, willow.Color{}, spanOutline)

	resolved := rt.ResolveOutlineForTest(rt.Spans()[0])
	if resolved != spanOutline {
		t.Error("span with explicit outline should use its own")
	}
}

func TestOutlineExplicitlyNone(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	defaultOutline := &ui.Outline{Color: willow.RGBA(0, 0, 0, 1), Thickness: 2}
	rt.SetOutline(defaultOutline)

	noOutline := &ui.Outline{}
	rt.AddStyledSpan("no-outline", nil, willow.Color{}, noOutline)

	resolved := rt.ResolveOutlineForTest(rt.Spans()[0])
	if resolved != noOutline {
		t.Error("span with explicit zero outline should not inherit default")
	}
}

// --- Empty spans ---

func TestEmptySpansSkipped(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.AddSpan("")
	rt.AddSpan("visible")
	rt.AddSpan("")

	lines := rt.LayoutLinesForTest()
	totalFragments := 0
	for _, line := range lines {
		totalFragments += len(line.Fragments)
	}
	if totalFragments != 1 {
		t.Errorf("expected 1 non-empty fragment, got %d", totalFragments)
	}
}

// --- ClearSpans + re-add ---

func TestClearSpansAndReAdd(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.AddSpan("first")
	rt.AddSpan("second")

	if len(rt.Spans()) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(rt.Spans()))
	}

	rt.ClearSpans()
	if len(rt.Spans()) != 0 {
		t.Errorf("after ClearSpans, expected 0 spans, got %d", len(rt.Spans()))
	}
	if !rt.Dirty() {
		t.Error("ClearSpans should mark dirty")
	}

	rt.AddSpan("third")
	if len(rt.Spans()) != 1 {
		t.Errorf("after re-add, expected 1 span, got %d", len(rt.Spans()))
	}
	if rt.Spans()[0].Text != "third" {
		t.Errorf("span text = %q, want %q", rt.Spans()[0].Text, "third")
	}
}

// --- SetSpans replaces all ---

func TestSetSpansReplacesAll(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.AddSpan("original")

	newSpans := []ui.TextSpan{
		{Text: "replaced one"},
		{Text: "replaced two"},
	}
	rt.SetSpans(newSpans)

	if len(rt.Spans()) != 2 {
		t.Fatalf("expected 2 spans after SetSpans, got %d", len(rt.Spans()))
	}
	if rt.Spans()[0].Text != "replaced one" {
		t.Errorf("span[0] text = %q, want %q", rt.Spans()[0].Text, "replaced one")
	}
	if rt.Spans()[1].Text != "replaced two" {
		t.Errorf("span[1] text = %q, want %q", rt.Spans()[1].Text, "replaced two")
	}
	if !rt.Dirty() {
		t.Error("SetSpans should mark dirty")
	}
}

// --- Dispose ---

func TestRichTextDisposeReleasesImage(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.AddSpan("test")
	rt.Render()

	rt.Dispose()
	if rt.ImageForTest() != nil {
		t.Error("image should be nil after Dispose")
	}
}

// --- Dirty tracking ---

func TestAddSpanMarksDirty(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.SetDirtyForTest(false)

	rt.AddSpan("new")
	if !rt.Dirty() {
		t.Error("AddSpan should mark dirty")
	}
}

func TestSetWrapWidthMarksDirty(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.SetDirtyForTest(false)

	rt.SetWrapWidth(200)
	if !rt.Dirty() {
		t.Error("SetWrapWidth should mark dirty")
	}
}

func TestSetAlignMarksDirty(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.SetDirtyForTest(false)

	rt.SetAlign(willow.TextAlignCenter)
	if !rt.Dirty() {
		t.Error("SetAlign should mark dirty")
	}
}

func TestSetColorMarksDirty(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.SetDirtyForTest(false)

	rt.SetColor(willow.RGBA(1, 0, 0, 1))
	if !rt.Dirty() {
		t.Error("SetColor should mark dirty")
	}
}

func TestSetOutlineMarksDirty(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.SetDirtyForTest(false)

	rt.SetOutline(&ui.Outline{Thickness: 1})
	if !rt.Dirty() {
		t.Error("SetOutline should mark dirty")
	}
}

// --- Chaining ---

func TestMethodChaining(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)

	result := rt.AddSpan("a").AddSpan("b").ClearSpans().AddSpan("c")
	if result != rt {
		t.Error("chained methods should return the same *RichText")
	}
	if len(rt.Spans()) != 1 {
		t.Errorf("expected 1 span after chain, got %d", len(rt.Spans()))
	}
}

// --- Color resolution ---

func TestColorResolution(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.SetColor(willow.RGBA(0.5, 0.5, 0.5, 1))

	rt.AddSpan("inherit")
	resolved := rt.ResolveColorForTest(rt.Spans()[0])
	if resolved != rt.Color() {
		t.Errorf("inherited color = %v, want %v", resolved, rt.Color())
	}

	red := willow.RGBA(1, 0, 0, 1)
	rt.AddStyledSpan("red", nil, red, nil)
	resolved = rt.ResolveColorForTest(rt.Spans()[1])
	if resolved != red {
		t.Errorf("explicit color = %v, want %v", resolved, red)
	}
}

// --- Font resolution ---

func TestFontResolution(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.AddSpan("default font")

	// Span without source: resolves via the RichText source (returns font itself).
	resolved := rt.ResolveFontForTest(rt.Spans()[0])
	if resolved == nil {
		t.Error("span without source should resolve to RichText default font")
	}

	largeFont := newLargeTestFont()
	rt.AddStyledSpan("large", largeFont, willow.Color{}, nil)
	// Span with explicit source: resolves via span source.
	resolved = rt.ResolveFontForTest(rt.Spans()[1])
	if resolved == nil {
		t.Error("span with explicit source should resolve to a font")
	}
}

// --- Sprite node creation ---

func TestSpriteNodeCreated(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)

	if rt.SpriteNode() == nil {
		t.Fatal("sprite node should be created")
	}
	if rt.Node().NumChildren() < 1 {
		t.Error("container node should have the sprite child")
	}
}

// --- Newlines in spans ---

func TestNewlineBreaksLine(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 0)
	rt.AddSpan("line one\nline two")

	lines := rt.LayoutLinesForTest()
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines from newline, got %d", len(lines))
	}
}
