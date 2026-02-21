package integration

import (
	"strings"
	"testing"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

var defaultHeadingScale = [3]float64{2.0, 1.5, 1.2}

func TestParseMarkup_PlainText(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("hello world", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Text != "hello world" {
		t.Errorf("text = %q, want %q", spans[0].Text, "hello world")
	}
}

func TestParseMarkup_Bold(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<b>bold</b>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Bold {
		t.Error("expected Bold=true")
	}
}

func TestParseMarkup_Italic(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<i>italic</i>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Italic {
		t.Error("expected Italic=true")
	}
}

func TestParseMarkup_BoldItalic(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<b><i>both</i></b>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Bold || !spans[0].Italic {
		t.Error("expected Bold=true and Italic=true")
	}
}

func TestParseMarkup_Underline(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<u>underline</u>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Underline {
		t.Error("expected underline=true")
	}
}

func TestParseMarkup_Strikethrough(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<strike>struck</strike>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].Strikethrough {
		t.Error("expected strikethrough=true")
	}
}

func TestParseMarkup_Color(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup(`<color value="#ff0000">red</color>`, font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].ColorSet {
		t.Error("expected ColorSet=true")
	}
	if spans[0].Color.R() != 1 || spans[0].Color.G() != 0 || spans[0].Color.B() != 0 {
		t.Errorf("color = %+v, want red", spans[0].Color)
	}
}

func TestParseMarkup_Size(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup(`<size value="24">big</size>`, font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].SizeOverride != 24 {
		t.Errorf("sizeOverride = %f, want 24", spans[0].SizeOverride)
	}
}

func TestParseMarkup_Outline(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup(`<outline thickness="3" color="white">outlined</outline>`, font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Outline == nil {
		t.Fatal("expected non-nil outline")
	}
	if spans[0].Outline.Thickness != 3 {
		t.Errorf("thickness = %f, want 3", spans[0].Outline.Thickness)
	}
	if spans[0].Outline.Color.R() != 1 || spans[0].Outline.Color.G() != 1 || spans[0].Outline.Color.B() != 1 {
		t.Errorf("outline color = %+v, want white", spans[0].Outline.Color)
	}
}

func TestParseMarkup_Span(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup(`<span color="#00ff00" size="20">styled</span>`, font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !spans[0].ColorSet {
		t.Error("expected ColorSet=true")
	}
	if spans[0].Color.G() != 1 {
		t.Errorf("color.G = %f, want 1", spans[0].Color.G())
	}
	if spans[0].SizeOverride != 20 {
		t.Errorf("sizeOverride = %f, want 20", spans[0].SizeOverride)
	}
}

func TestParseMarkup_Link(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup(`<link url="https://example.com">click</link>`, font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].LinkURL != "https://example.com" {
		t.Errorf("linkURL = %q", spans[0].LinkURL)
	}
	if !spans[0].Underline {
		t.Error("links should be underlined by default")
	}
}

func TestParseMarkup_Br(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("line1<br/>line2", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	// Should have: "line1", "\n", "line2"
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(spans))
	}
	if spans[1].Text != "\n" {
		t.Errorf("span[1].Text = %q, want newline", spans[1].Text)
	}
}

func TestParseMarkup_Heading(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<h1>Title</h1>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	// Should emit: newline, "Title", newline
	found := false
	for _, s := range spans {
		if s.Text == "Title" {
			found = true
			if s.SizeOverride != 32 { // 16 * 2.0
				t.Errorf("h1 sizeOverride = %f, want 32", s.SizeOverride)
			}
			if !s.Bold {
				t.Error("h1 should have Bold=true")
			}
		}
	}
	if !found {
		t.Error("did not find Title span")
	}
}

func TestParseMarkup_UnorderedList(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<ul><li>one</li><li>two</li></ul>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	// Should contain bullet characters
	hasBullet := false
	for _, s := range spans {
		if s.Text == "\u2022 " {
			hasBullet = true
		}
	}
	if !hasBullet {
		t.Error("expected bullet prefix in unordered list")
	}
}

func TestParseMarkup_OrderedList(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<ol><li>first</li><li>second</li></ol>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	has1 := false
	has2 := false
	for _, s := range spans {
		if s.Text == "1. " {
			has1 = true
		}
		if s.Text == "2. " {
			has2 = true
		}
	}
	if !has1 || !has2 {
		t.Error("expected numbered prefixes in ordered list")
	}
}

func TestParseMarkup_Nesting(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<b>bold <i>and italic</i> just bold</b>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(spans))
	}
	if !spans[0].Bold || spans[0].Italic {
		t.Error("span[0] should be bold only")
	}
	if !spans[1].Bold || !spans[1].Italic {
		t.Error("span[1] should be bold-italic")
	}
	if !spans[2].Bold || spans[2].Italic {
		t.Error("span[2] should be bold only (restored)")
	}
}

func TestParseMarkup_UnknownTag(t *testing.T) {
	font := newTestFont()
	_, err := ui.ParseMarkup("<unknown>text</unknown>", font, 16, defaultHeadingScale)
	if err == nil {
		t.Error("expected error for unknown tag")
	}
}

func TestParseMarkup_MixedContent(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("Hello <b>world</b>!", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(spans))
	}
	if spans[0].Text != "Hello " {
		t.Errorf("span[0] = %q", spans[0].Text)
	}
	if spans[1].Text != "world" {
		t.Errorf("span[1] = %q", spans[1].Text)
	}
	if spans[2].Text != "!" {
		t.Errorf("span[2] = %q", spans[2].Text)
	}
}

func TestParseMarkup_EmptyInput(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 0 {
		t.Errorf("expected 0 spans for empty input, got %d", len(spans))
	}
}

// --- RichText.SetMarkup ---

func TestRichText_SetMarkup(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 16)

	err := rt.SetMarkup("Hello <b>bold</b> world")
	if err != nil {
		t.Fatal(err)
	}
	if len(rt.Spans()) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(rt.Spans()))
	}
	// Middle span should have Bold=true, Source=nil (inherits from rt source).
	if !rt.Spans()[1].Bold {
		t.Error("middle span should have Bold=true")
	}
}

func TestRichText_SetMarkupError(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 16)
	err := rt.SetMarkup("<badtag>oops</badtag>")
	if err == nil {
		t.Error("expected error for unknown tag")
	}
}

// --- TextSpan new fields ---

func TestTextSpan_SizeOverrideLayout(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 16)
	rt.SetSpans([]ui.TextSpan{
		{Text: "normal "},
		{Text: "big", SizeOverride: 32},
	})
	lines := rt.LayoutLinesForTest()
	if len(lines) == 0 {
		t.Fatal("expected at least 1 line")
	}
	// The big text fragment should have sizeOverride set
	found := false
	for _, line := range lines {
		for _, frag := range line.Fragments {
			if frag.SizeOverride == 32 {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected fragment with sizeOverride=32")
	}
}

func TestTextSpan_UnderlineLayout(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 16)
	rt.SetSpans([]ui.TextSpan{
		{Text: "underlined", Underline: true},
	})
	lines := rt.LayoutLinesForTest()
	if len(lines) == 0 {
		t.Fatal("expected at least 1 line")
	}
	if !lines[0].Fragments[0].Underline {
		t.Error("expected underline=true on fragment")
	}
}

func TestTextSpan_LinkURLLayout(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 16)
	rt.SetSpans([]ui.TextSpan{
		{Text: "click me", LinkURL: "https://example.com"},
	})
	lines := rt.LayoutLinesForTest()
	if len(lines) == 0 {
		t.Fatal("expected at least 1 line")
	}
	if lines[0].Fragments[0].LinkURL != "https://example.com" {
		t.Error("expected linkURL on fragment")
	}
}

// --- HeadingScale ---

func TestRichText_HeadingScale(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 16)
	if rt.HeadingScale() != [3]float64{2.0, 1.5, 1.2} {
		t.Errorf("default headingScale = %v", rt.HeadingScale())
	}
	rt.SetHeadingScale(3.0, 2.0, 1.5)
	if rt.HeadingScale() != [3]float64{3.0, 2.0, 1.5} {
		t.Errorf("after SetHeadingScale = %v", rt.HeadingScale())
	}
}

// --- H2/H3 ---

func TestParseMarkup_H2(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<h2>Sub</h2>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range spans {
		if s.Text == "Sub" {
			found = true
			if s.SizeOverride != 24 { // 16 * 1.5
				t.Errorf("h2 sizeOverride = %f, want 24", s.SizeOverride)
			}
		}
	}
	if !found {
		t.Error("did not find Sub span")
	}
}

func TestParseMarkup_H3(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<h3>Minor</h3>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range spans {
		if s.Text == "Minor" {
			found = true
			expected := 16 * 1.2
			if spans[1].SizeOverride != expected {
				t.Errorf("h3 sizeOverride = %f, want %f", s.SizeOverride, expected)
			}
		}
	}
	if !found {
		t.Error("did not find Minor span")
	}
}

// --- Outline default thickness ---

func TestParseMarkup_OutlineDefaultThickness(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup(`<outline>text</outline>`, font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Outline == nil {
		t.Fatal("expected outline")
	}
	if spans[0].Outline.Thickness != 2 {
		t.Errorf("default thickness = %f, want 2", spans[0].Outline.Thickness)
	}
}

// --- Link callback ---

func TestRichText_OnLinkClick(t *testing.T) {
	font := newTestFont()
	rt := ui.NewRichText("rt", font, 16)

	var clickedURL string
	rt.SetOnLinkClick(func(url string) {
		clickedURL = url
	})
	if rt.OnLinkClickForTest() == nil {
		t.Error("onLinkClick should be set")
	}
	rt.OnLinkClickForTest()("https://test.com")
	if clickedURL != "https://test.com" {
		t.Errorf("clickedURL = %q", clickedURL)
	}
}

// --- Compiler integration: RichText with markup children ---

func TestCompileXML_RichTextMarkup(t *testing.T) {
	xmlData := `<RichText><b>bold</b> text</RichText>`
	node, err := ui.CompileXML([]byte(xmlData))
	if err != nil {
		t.Fatal(err)
	}
	if node.ComponentType != "RichText" {
		t.Errorf("type = %q, want RichText", node.ComponentType)
	}
	// Should have no children (markup tags are not compiled as components)
	if len(node.Children) != 0 {
		t.Errorf("expected 0 children, got %d", len(node.Children))
	}
	// Should have a markup attribute
	found := false
	for _, attr := range node.Attributes {
		if attr.Name == "markup" {
			found = true
			if attr.Static == "" {
				t.Error("markup attribute should not be empty")
			}
		}
	}
	if !found {
		t.Error("expected markup attribute")
	}
}

func TestCompileXML_RichTextWithComponent(t *testing.T) {
	// RichText with a real component child should compile normally
	xmlData := `<RichText><Label text="hi" /></RichText>`
	node, err := ui.CompileXML([]byte(xmlData))
	if err != nil {
		t.Fatal(err)
	}
	if len(node.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(node.Children))
	}
	if node.Children[0].ComponentType != "Label" {
		t.Errorf("child type = %q, want Label", node.Children[0].ComponentType)
	}
}

func TestCompileXML_RichTextMarkupSerialization(t *testing.T) {
	xmlData := `<RichText><b>bold</b> and <i>italic</i></RichText>`
	node, err := ui.CompileXML([]byte(xmlData))
	if err != nil {
		t.Fatal(err)
	}
	var markup string
	for _, attr := range node.Attributes {
		if attr.Name == "markup" {
			markup = attr.Static
		}
	}
	if markup == "" {
		t.Fatal("expected non-empty markup attribute")
	}
	// Verify the markup contains the expected tags
	if !containsAll(markup, "<b>", "</b>", "<i>", "</i>", "bold", "italic") {
		t.Errorf("markup = %q, expected bold/italic tags", markup)
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !containsStr(s, sub) {
			return false
		}
	}
	return true
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// --- List spacing ---

func TestParseMarkup_ListNoLeadingNewline(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<ul><li>one</li><li>two</li></ul>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	// First span should be the bullet, not a newline.
	if len(spans) == 0 {
		t.Fatal("expected spans")
	}
	if spans[0].Text == "\n" {
		t.Error("first span should not be a newline — no leading blank line in lists")
	}
}

func TestParseMarkup_ListWhitespaceSuppressed(t *testing.T) {
	// Formatted with newlines between tags (common in XML templates).
	input := "<ul>\n  <li>one</li>\n  <li>two</li>\n</ul>"
	font := newTestFont()
	spans, err := ui.ParseMarkup(input, font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	// Should not contain any whitespace-only spans from between tags.
	for i, s := range spans {
		trimmed := strings.TrimSpace(s.Text)
		if s.Text != "\n" && trimmed == "" {
			t.Errorf("span[%d] = %q — whitespace between list tags should be suppressed", i, s.Text)
		}
	}
}

func TestParseMarkup_TextThenListNoDoubleNewline(t *testing.T) {
	// "Here are the reasons:\n<ul><li>hi</li></ul>"
	// The \n at the end of the text already breaks the line,
	// so the first <li> should NOT insert another newline.
	input := "Here are the reasons:\n<ul><li>hi</li><li>bye</li></ul>"
	font := newTestFont()
	spans, err := ui.ParseMarkup(input, font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	// Count consecutive newlines — should never have two in a row.
	prevNewline := false
	for _, s := range spans {
		isNL := s.Text == "\n"
		if isNL && prevNewline {
			t.Error("found double newline between text and list — should be suppressed")
		}
		prevNewline = isNL || (len(s.Text) > 0 && s.Text[len(s.Text)-1] == '\n')
	}
}

func TestParseMarkup_ListIndent(t *testing.T) {
	font := newTestFont()
	spans, err := ui.ParseMarkup("<ul><li>item</li></ul>", font, 16, defaultHeadingScale)
	if err != nil {
		t.Fatal(err)
	}
	// The bullet prefix span should have Indent > 0.
	found := false
	for _, s := range spans {
		if s.Indent > 0 {
			found = true
		}
	}
	if !found {
		t.Error("expected at least one span with Indent > 0 for list items")
	}
}

// Verify willow.Color is not used incorrectly
var _ willow.Color = willow.RGBA(1, 0, 0, 1)
