package markup

import (
	"math"
	"strings"
	"testing"

	"github.com/devthicket/willowui/internal/sg"
)

func colorApproxEq(a, b sg.Color, eps float64) bool {
	return math.Abs(a.R()-b.R()) < eps &&
		math.Abs(a.G()-b.G()) < eps &&
		math.Abs(a.B()-b.B()) < eps &&
		math.Abs(a.A()-b.A()) < eps
}

// ── ParseColor ──────────────────────────────────────────────────────────────

func TestParseColor_Named(t *testing.T) {
	tests := []struct {
		input string
		want  sg.Color
	}{
		{"white", sg.RGBA(1, 1, 1, 1)},
		{"WHITE", sg.RGBA(1, 1, 1, 1)},
		{"black", sg.RGBA(0, 0, 0, 1)},
		{"transparent", sg.RGBA(0, 0, 0, 0)},
	}
	for _, tt := range tests {
		c, err := ParseColor(tt.input)
		if err != nil {
			t.Errorf("ParseColor(%q) error: %v", tt.input, err)
			continue
		}
		if !colorApproxEq(c, tt.want, 0.01) {
			t.Errorf("ParseColor(%q) = %v, want %v", tt.input, c, tt.want)
		}
	}
}

func TestParseColor_Hex6(t *testing.T) {
	c, err := ParseColor("#FF8000")
	if err != nil {
		t.Fatal(err)
	}
	want := sg.RGBA(1.0, 128.0/255, 0, 1)
	if !colorApproxEq(c, want, 0.01) {
		t.Errorf("got %v, want %v", c, want)
	}
}

func TestParseColor_Hex8(t *testing.T) {
	c, err := ParseColor("#FF800080")
	if err != nil {
		t.Fatal(err)
	}
	want := sg.RGBA(1.0, 128.0/255, 0, 128.0/255)
	if !colorApproxEq(c, want, 0.01) {
		t.Errorf("got %v, want %v", c, want)
	}
}

func TestParseColor_Hex3(t *testing.T) {
	c, err := ParseColor("#F80")
	if err != nil {
		t.Fatal(err)
	}
	want := sg.RGBA(1.0, 136.0/255, 0, 1)
	if !colorApproxEq(c, want, 0.01) {
		t.Errorf("got %v, want %v", c, want)
	}
}

func TestParseColor_RGB(t *testing.T) {
	c, err := ParseColor("rgb(255, 128, 0)")
	if err != nil {
		t.Fatal(err)
	}
	want := sg.RGBA(1.0, 128.0/255, 0, 1)
	if !colorApproxEq(c, want, 0.01) {
		t.Errorf("got %v, want %v", c, want)
	}
}

func TestParseColor_RGBA(t *testing.T) {
	c, err := ParseColor("rgba(255, 128, 0, 0.5)")
	if err != nil {
		t.Fatal(err)
	}
	want := sg.RGBA(1.0, 128.0/255, 0, 0.5)
	if !colorApproxEq(c, want, 0.01) {
		t.Errorf("got %v, want %v", c, want)
	}
}

func TestParseColor_Errors(t *testing.T) {
	bad := []string{
		"",
		"notacolor",
		"#GG0000",
		"#12",
		"rgb(300, 0, 0)",
		"rgba(0, 0, 0, 2.0)",
		"rgb(0, 0)",
		"rgba(0, 0, 0)",
	}
	for _, s := range bad {
		_, err := ParseColor(s)
		if err == nil {
			t.Errorf("ParseColor(%q) expected error, got nil", s)
		}
	}
}

// ── ParseMarkup ─────────────────────────────────────────────────────────────

func TestParseMarkup_PlainText(t *testing.T) {
	spans, err := ParseMarkup("hello world", nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].Text != "hello world" {
		t.Errorf("got %q, want %q", spans[0].Text, "hello world")
	}
}

func TestParseMarkup_Bold(t *testing.T) {
	spans, err := ParseMarkup("normal <b>bold</b> normal", nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 3 {
		t.Fatalf("got %d spans, want 3", len(spans))
	}
	if spans[1].Text != "bold" || !spans[1].Bold {
		t.Errorf("span[1]: text=%q bold=%v", spans[1].Text, spans[1].Bold)
	}
	if spans[0].Bold || spans[2].Bold {
		t.Error("surrounding spans should not be bold")
	}
}

func TestParseMarkup_Nested(t *testing.T) {
	spans, err := ParseMarkup("<b><i>both</i></b>", nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if !spans[0].Bold || !spans[0].Italic {
		t.Errorf("expected bold+italic, got bold=%v italic=%v", spans[0].Bold, spans[0].Italic)
	}
}

func TestParseMarkup_Color(t *testing.T) {
	spans, err := ParseMarkup(`<color value="#FF0000">red</color>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if !spans[0].ColorSet {
		t.Error("expected ColorSet=true")
	}
	want := sg.RGBA(1, 0, 0, 1)
	if !colorApproxEq(spans[0].Color, want, 0.01) {
		t.Errorf("got color %v, want %v", spans[0].Color, want)
	}
}

func TestParseMarkup_Size(t *testing.T) {
	spans, err := ParseMarkup(`<size value="24">big</size>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].SizeOverride != 24 {
		t.Errorf("got size %v, want 24", spans[0].SizeOverride)
	}
}

func TestParseMarkup_Heading(t *testing.T) {
	scale := [3]float64{2, 1.5, 1.2}
	spans, err := ParseMarkup("<h1>Title</h1>", nil, 16, scale)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) < 1 {
		t.Fatal("expected at least 1 span")
	}
	if spans[0].SizeOverride != 32 { // 16 * 2.0
		t.Errorf("h1 size = %v, want 32", spans[0].SizeOverride)
	}
	if !spans[0].Bold {
		t.Error("h1 should be bold")
	}
}

func TestParseMarkup_LineBreak(t *testing.T) {
	spans, err := ParseMarkup("line1<br/>line2", nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range spans {
		if s.Text == "\n" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a newline span from <br/>")
	}
}

func TestParseMarkup_UnknownTag(t *testing.T) {
	_, err := ParseMarkup("<foo>text</foo>", nil, 16, [3]float64{2, 1.5, 1.2})
	if err == nil {
		t.Error("expected error for unknown tag")
	}
}

func TestParseMarkup_Empty(t *testing.T) {
	spans, err := ParseMarkup("", nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 0 {
		t.Errorf("expected 0 spans for empty input, got %d", len(spans))
	}
}

func TestParseMarkup_Link(t *testing.T) {
	spans, err := ParseMarkup(`<link url="https://example.com">click</link>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].LinkURL != "https://example.com" {
		t.Errorf("got link %q, want %q", spans[0].LinkURL, "https://example.com")
	}
	if !spans[0].Underline {
		t.Error("links should have underline")
	}
}

func TestParseMarkup_UnorderedList(t *testing.T) {
	spans, err := ParseMarkup("<ul><li>A</li><li>B</li></ul>", nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	// Should contain bullet characters and the text items
	var texts []string
	for _, s := range spans {
		if s.Text != "\n" && s.Text != "" {
			texts = append(texts, s.Text)
		}
	}
	if len(texts) < 2 {
		t.Errorf("expected at least 2 text spans, got %v", texts)
	}
}

func TestParseMarkup_Outline(t *testing.T) {
	spans, err := ParseMarkup(`<outline thickness="2" color="white">text</outline>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].Outline == nil {
		t.Fatal("expected outline to be set")
	}
	if spans[0].Outline.Thickness != 2 {
		t.Errorf("got thickness %v, want 2", spans[0].Outline.Thickness)
	}
}

// ── Additional coverage tests ───────────────────────────────────────────────

func TestParseColor_Hex3_Invalid(t *testing.T) {
	bad := []string{"#GGG", "#ZAB", "#1G2"}
	for _, s := range bad {
		_, err := ParseColor(s)
		if err == nil {
			t.Errorf("ParseColor(%q) expected error for invalid 3-digit hex", s)
		}
	}
}

func TestParseColor_Hex6_Invalid(t *testing.T) {
	bad := []string{"#GGGGGG", "#ZZ0000", "#00GG00", "#0000ZZ"}
	for _, s := range bad {
		_, err := ParseColor(s)
		if err == nil {
			t.Errorf("ParseColor(%q) expected error for invalid 6-digit hex", s)
		}
	}
}

func TestParseColor_Hex8_Invalid(t *testing.T) {
	bad := []string{"#FF0000GG", "#FFFFFFZZ"}
	for _, s := range bad {
		_, err := ParseColor(s)
		if err == nil {
			t.Errorf("ParseColor(%q) expected error for invalid 8-digit hex", s)
		}
	}
}

func TestParseColor_Hex_InvalidLength(t *testing.T) {
	bad := []string{"#1", "#12", "#1234", "#12345", "#1234567", "#123456789"}
	for _, s := range bad {
		_, err := ParseColor(s)
		if err == nil {
			t.Errorf("ParseColor(%q) expected error for unsupported hex length", s)
		}
	}
}

func TestParseColor_RGBA_WrongValueCount(t *testing.T) {
	bad := []string{
		"rgba(0, 0)",
		"rgba(0, 0, 0, 0.5, 1)",
		"rgba(0)",
	}
	for _, s := range bad {
		_, err := ParseColor(s)
		if err == nil {
			t.Errorf("ParseColor(%q) expected error for wrong number of rgba values", s)
		}
	}
}

func TestParseColor_RGBA_OutOfRange(t *testing.T) {
	bad := []string{
		"rgba(256, 0, 0, 1.0)",
		"rgba(0, -1, 0, 1.0)",
		"rgba(0, 0, 0, -0.1)",
		"rgba(0, 0, 0, 1.1)",
	}
	for _, s := range bad {
		_, err := ParseColor(s)
		if err == nil {
			t.Errorf("ParseColor(%q) expected error for out-of-range value", s)
		}
	}
}

func TestParseColor_RGB_Negative(t *testing.T) {
	bad := []string{
		"rgb(-1, 0, 0)",
		"rgb(0, -5, 0)",
		"rgb(0, 0, -10)",
	}
	for _, s := range bad {
		_, err := ParseColor(s)
		if err == nil {
			t.Errorf("ParseColor(%q) expected error for negative rgb value", s)
		}
	}
}

func TestParseColor_RGB_NonNumeric(t *testing.T) {
	bad := []string{
		"rgb(abc, 0, 0)",
		"rgb(0, xyz, 0)",
		"rgb(0, 0, !!!)",
	}
	for _, s := range bad {
		_, err := ParseColor(s)
		if err == nil {
			t.Errorf("ParseColor(%q) expected error for non-numeric rgb value", s)
		}
	}
}

func TestParseColor_RGB_WrongValueCount(t *testing.T) {
	bad := []string{
		"rgb(0, 0)",
		"rgb(0, 0, 0, 0)",
		"rgb(0)",
	}
	for _, s := range bad {
		_, err := ParseColor(s)
		if err == nil {
			t.Errorf("ParseColor(%q) expected error for wrong number of rgb values", s)
		}
	}
}

func TestParseMarkup_Strike(t *testing.T) {
	spans, err := ParseMarkup("normal <strike>deleted</strike> normal", nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 3 {
		t.Fatalf("got %d spans, want 3", len(spans))
	}
	if spans[1].Text != "deleted" || !spans[1].Strikethrough {
		t.Errorf("span[1]: text=%q strikethrough=%v", spans[1].Text, spans[1].Strikethrough)
	}
	if spans[0].Strikethrough || spans[2].Strikethrough {
		t.Error("surrounding spans should not be strikethrough")
	}
}

func TestParseMarkup_Span_ColorAndSize(t *testing.T) {
	spans, err := ParseMarkup(`<span color="#00FF00" size="20">styled</span>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if !spans[0].ColorSet {
		t.Error("expected ColorSet=true from span color attribute")
	}
	wantColor := sg.RGBA(0, 1, 0, 1)
	if !colorApproxEq(spans[0].Color, wantColor, 0.01) {
		t.Errorf("span color = %v, want %v", spans[0].Color, wantColor)
	}
	if spans[0].SizeOverride != 20 {
		t.Errorf("span size = %v, want 20", spans[0].SizeOverride)
	}
}

func TestParseMarkup_Span_FontBold(t *testing.T) {
	spans, err := ParseMarkup(`<span font="bold">text</span>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if !spans[0].Bold {
		t.Error("expected Bold=true for font=bold")
	}
}

func TestParseMarkup_Span_FontItalic(t *testing.T) {
	spans, err := ParseMarkup(`<span font="italic">text</span>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if !spans[0].Italic {
		t.Error("expected Italic=true for font=italic")
	}
}

func TestParseMarkup_Span_FontBoldItalic(t *testing.T) {
	spans, err := ParseMarkup(`<span font="bolditalic">text</span>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if !spans[0].Bold || !spans[0].Italic {
		t.Errorf("expected Bold+Italic for font=bolditalic, got bold=%v italic=%v", spans[0].Bold, spans[0].Italic)
	}
}

func TestParseMarkup_Span_FontRegular(t *testing.T) {
	// Start bold, then span font=regular should reset
	spans, err := ParseMarkup(`<b><span font="regular">text</span></b>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].Bold {
		t.Error("expected Bold=false for font=regular inside <b>")
	}
}

func TestParseMarkup_OrderedList(t *testing.T) {
	spans, err := ParseMarkup("<ol><li>A</li><li>B</li><li>C</li></ol>", nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	// Should contain numbered prefixes: "1. ", "2. ", "3. "
	var prefixes []string
	var items []string
	for _, s := range spans {
		if s.Text == "\n" || s.Text == "" {
			continue
		}
		if s.Indent > 0 {
			prefixes = append(prefixes, s.Text)
		} else {
			items = append(items, s.Text)
		}
	}
	if len(prefixes) != 3 {
		t.Fatalf("expected 3 numbered prefixes, got %d: %v", len(prefixes), prefixes)
	}
	if prefixes[0] != "1. " || prefixes[1] != "2. " || prefixes[2] != "3. " {
		t.Errorf("prefixes = %v, want [1.  2.  3. ]", prefixes)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d: %v", len(items), items)
	}
}

func TestParseMarkup_NestedLists(t *testing.T) {
	input := "<ul><li>outer<ul><li>inner1</li><li>inner2</li></ul></li><li>outer2</li></ul>"
	spans, err := ParseMarkup(input, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	// Verify we get spans with different indent levels
	var indents []float64
	for _, s := range spans {
		if s.Indent > 0 {
			indents = append(indents, s.Indent)
		}
	}
	if len(indents) < 3 {
		t.Fatalf("expected at least 3 indented spans for nested list, got %d", len(indents))
	}
	// Inner items should have deeper indent than outer
	hasDeeper := false
	for _, ind := range indents {
		if ind > listIndent {
			hasDeeper = true
			break
		}
	}
	if !hasDeeper {
		t.Error("expected nested list items to have deeper indent")
	}
}

func TestParseMarkup_H2(t *testing.T) {
	scale := [3]float64{2, 1.5, 1.2}
	spans, err := ParseMarkup("<h2>Subtitle</h2>", nil, 16, scale)
	if err != nil {
		t.Fatal(err)
	}
	// Find the text span
	var found *TextSpan
	for i := range spans {
		if spans[i].Text == "Subtitle" {
			found = &spans[i]
			break
		}
	}
	if found == nil {
		t.Fatal("could not find 'Subtitle' span")
	}
	if found.SizeOverride != 24 { // 16 * 1.5
		t.Errorf("h2 size = %v, want 24", found.SizeOverride)
	}
	if !found.Bold {
		t.Error("h2 should be bold")
	}
}

func TestParseMarkup_H3(t *testing.T) {
	scale := [3]float64{2, 1.5, 1.2}
	spans, err := ParseMarkup("<h3>Section</h3>", nil, 16, scale)
	if err != nil {
		t.Fatal(err)
	}
	var found *TextSpan
	for i := range spans {
		if spans[i].Text == "Section" {
			found = &spans[i]
			break
		}
	}
	if found == nil {
		t.Fatal("could not find 'Section' span")
	}
	if found.SizeOverride != 19.2 { // 16 * 1.2
		t.Errorf("h3 size = %v, want 19.2", found.SizeOverride)
	}
	if !found.Bold {
		t.Error("h3 should be bold")
	}
}

func TestParseMarkup_HeadingNewlines(t *testing.T) {
	spans, err := ParseMarkup("<h1>Title</h1>after", nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	// h1 produces: \n (before), "Title", \n (after), "after"
	newlineCount := 0
	for _, s := range spans {
		if s.Text == "\n" {
			newlineCount++
		}
	}
	if newlineCount < 2 {
		t.Errorf("expected at least 2 newlines around heading, got %d", newlineCount)
	}
}

func TestParseMarkup_ListWhitespaceSuppressed(t *testing.T) {
	// Whitespace between list items should be suppressed
	input := "<ul>\n  <li>A</li>\n  <li>B</li>\n</ul>"
	spans, err := ParseMarkup(input, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	// No span should contain only whitespace (spaces/tabs) from between tags
	for _, s := range spans {
		trimmed := strings.TrimSpace(s.Text)
		if trimmed == "" && s.Text != "\n" && s.Text != "" {
			t.Errorf("found whitespace-only span %q that should have been suppressed", s.Text)
		}
	}
}

func TestParseMarkup_ColorAttrMissing(t *testing.T) {
	// <color> tag without a value attribute - attrVal returns "", ParseColor fails, colorSet stays false
	spans, err := ParseMarkup(`<color>text</color>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].ColorSet {
		t.Error("expected ColorSet=false when color tag has no value attribute")
	}
}

func TestParseMarkup_SizeAttrMissing(t *testing.T) {
	// <size> tag without a value attribute - sizeOverride stays 0
	spans, err := ParseMarkup(`<size>text</size>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].SizeOverride != 0 {
		t.Errorf("expected SizeOverride=0 when size tag has no value, got %v", spans[0].SizeOverride)
	}
}

func TestParseMarkup_SizeAttrInvalid(t *testing.T) {
	// <size value="abc"> - non-numeric value, sizeOverride stays 0
	spans, err := ParseMarkup(`<size value="abc">text</size>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].SizeOverride != 0 {
		t.Errorf("expected SizeOverride=0 for invalid size value, got %v", spans[0].SizeOverride)
	}
}

func TestParseMarkup_SizeAttrNegative(t *testing.T) {
	// <size value="-5"> - negative value is not > 0, sizeOverride stays 0
	spans, err := ParseMarkup(`<size value="-5">text</size>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].SizeOverride != 0 {
		t.Errorf("expected SizeOverride=0 for negative size value, got %v", spans[0].SizeOverride)
	}
}

func TestParseMarkup_OutlineDefaults(t *testing.T) {
	// <outline> with no attributes should use defaults (thickness=2, color=black)
	spans, err := ParseMarkup(`<outline>text</outline>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].Outline == nil {
		t.Fatal("expected outline to be set")
	}
	if spans[0].Outline.Thickness != 2 {
		t.Errorf("default thickness = %v, want 2", spans[0].Outline.Thickness)
	}
	wantColor := sg.RGBA(0, 0, 0, 1)
	if !colorApproxEq(spans[0].Outline.Color, wantColor, 0.01) {
		t.Errorf("default outline color = %v, want black", spans[0].Outline.Color)
	}
}

func TestParseMarkup_OrderedListNumbering(t *testing.T) {
	// Verifies endsWithNewline indirectly: first <li> in an <ol> preceded by text
	// that does NOT end with newline should still get a newline inserted.
	spans, err := ParseMarkup("intro<ol><li>first</li><li>second</li></ol>", nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	// The "intro" text doesn't end with \n, so the first li should insert one.
	// Check that the first span after "intro" is "\n".
	if len(spans) < 2 {
		t.Fatalf("expected multiple spans, got %d", len(spans))
	}
	if spans[0].Text != "intro" {
		t.Errorf("first span = %q, want 'intro'", spans[0].Text)
	}
	if spans[1].Text != "\n" {
		t.Errorf("second span = %q, want newline before first list item", spans[1].Text)
	}
}

func TestParseMarkup_OrderedListAfterNewline(t *testing.T) {
	// If preceding text already ends with \n, no extra newline before first <li>.
	spans, err := ParseMarkup("intro\n<ol><li>first</li></ol>", nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	// Count consecutive newlines after "intro\n"
	newlineCount := 0
	pastIntro := false
	for _, s := range spans {
		if strings.Contains(s.Text, "intro") {
			pastIntro = true
			continue
		}
		if pastIntro && s.Text == "\n" {
			newlineCount++
		} else if pastIntro {
			break
		}
	}
	// Should NOT have an extra newline because "intro\n" already ends with \n
	if newlineCount > 0 {
		t.Errorf("expected 0 extra newlines after text ending with \\n, got %d", newlineCount)
	}
}

func TestParseMarkup_MixedOlUl(t *testing.T) {
	input := "<ol><li>numbered</li></ol><ul><li>bullet</li></ul>"
	spans, err := ParseMarkup(input, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	hasNumbered := false
	hasBullet := false
	for _, s := range spans {
		if strings.HasPrefix(s.Text, "1. ") {
			hasNumbered = true
		}
		if strings.Contains(s.Text, "\u2022") {
			hasBullet = true
		}
	}
	if !hasNumbered {
		t.Error("expected numbered prefix '1. ' from ol")
	}
	if !hasBullet {
		t.Error("expected bullet prefix from ul")
	}
}

func TestParseMarkup_Span_NoAttributes(t *testing.T) {
	// <span> with no attributes should produce the text unchanged
	spans, err := ParseMarkup(`<span>plain</span>`, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	if spans[0].Text != "plain" {
		t.Errorf("span text = %q, want 'plain'", spans[0].Text)
	}
	if spans[0].Bold || spans[0].Italic || spans[0].ColorSet || spans[0].SizeOverride != 0 {
		t.Error("expected no style overrides for bare <span>")
	}
}

func TestParseMarkup_MultipleInlineTags(t *testing.T) {
	input := "<b>bold</b> <i>italic</i> <u>under</u> <strike>struck</strike>"
	spans, err := ParseMarkup(input, nil, 16, [3]float64{2, 1.5, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	// Verify each styled span
	styleMap := map[string]func(TextSpan) bool{
		"bold":   func(s TextSpan) bool { return s.Bold },
		"italic": func(s TextSpan) bool { return s.Italic },
		"under":  func(s TextSpan) bool { return s.Underline },
		"struck": func(s TextSpan) bool { return s.Strikethrough },
	}
	for _, s := range spans {
		text := strings.TrimSpace(s.Text)
		if checker, ok := styleMap[text]; ok {
			if !checker(s) {
				t.Errorf("span %q missing expected style", text)
			}
		}
	}
}
