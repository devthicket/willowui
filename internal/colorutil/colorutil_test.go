package colorutil

import (
	"math"
	"testing"

	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

func approxEq(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}

func colorApproxEq(a, b sg.Color, eps float64) bool {
	return approxEq(a.R(), b.R(), eps) &&
		approxEq(a.G(), b.G(), eps) &&
		approxEq(a.B(), b.B(), eps) &&
		approxEq(a.A(), b.A(), eps)
}

// ── Hex ──────────────────────────────────────────────────────────────────────

func TestParseHex6(t *testing.T) {
	c, ok := ParseHex("#FF8800")
	if !ok {
		t.Fatal("expected ok")
	}
	r, g, b, a := ToRGB255(c)
	if r != 255 || g != 136 || b != 0 || a != 255 {
		t.Fatalf("got (%d,%d,%d,%d), want (255,136,0,255)", r, g, b, a)
	}
}

func TestParseHex8(t *testing.T) {
	c, ok := ParseHex("#FF880080")
	if !ok {
		t.Fatal("expected ok")
	}
	r, g, b, a := ToRGB255(c)
	if r != 255 || g != 136 || b != 0 || a != 128 {
		t.Fatalf("got (%d,%d,%d,%d), want (255,136,0,128)", r, g, b, a)
	}
}

func TestParseHexNoHash(t *testing.T) {
	c, ok := ParseHex("FF8800")
	if !ok {
		t.Fatal("expected ok")
	}
	r, g, b, _ := ToRGB255(c)
	if r != 255 || g != 136 || b != 0 {
		t.Fatalf("got (%d,%d,%d), want (255,136,0)", r, g, b)
	}
}

func TestParseHexCaseInsensitive(t *testing.T) {
	c1, ok1 := ParseHex("#ff8800")
	c2, ok2 := ParseHex("#FF8800")
	if !ok1 || !ok2 {
		t.Fatal("expected ok")
	}
	if !colorApproxEq(c1, c2, 0.001) {
		t.Fatal("case-insensitive parse should produce same color")
	}
}

func TestParseHexInvalid(t *testing.T) {
	cases := []string{"", "#", "#FFF", "#GGGGGG", "#12345", "not-a-color"}
	for _, s := range cases {
		if _, ok := ParseHex(s); ok {
			t.Fatalf("expected ParseHex(%q) to fail", s)
		}
	}
}

func TestFormatHex(t *testing.T) {
	c := sg.RGBA(1, 0.533, 0, 1) // approx #FF8800
	hex := FormatHex(c)
	if hex != "#FF8800" {
		t.Fatalf("got %q, want #FF8800", hex)
	}
}

func TestFormatHexA(t *testing.T) {
	c := sg.RGBA(1, 0, 0, 0.502) // approx #FF000080
	hexA := FormatHexA(c)
	if hexA != "#FF000080" {
		t.Fatalf("got %q, want #FF000080", hexA)
	}
}

func TestHexRoundTrip(t *testing.T) {
	original := "#3A7BF2"
	c, ok := ParseHex(original)
	if !ok {
		t.Fatal("parse failed")
	}
	result := FormatHex(c)
	if result != original {
		t.Fatalf("round-trip: got %q, want %q", result, original)
	}
}

// ── RGB 0–255 ─────────────────────────────────────────────────────────────────

func TestToFromRGB255(t *testing.T) {
	c := sg.RGBA(0.5, 0.25, 0.75, 1)
	r, g, b, a := ToRGB255(c)
	if r != 128 || g != 64 || b != 191 || a != 255 {
		t.Fatalf("got (%d,%d,%d,%d), want (128,64,191,255)", r, g, b, a)
	}

	back := FromRGB255(r, g, b, a)
	if !colorApproxEq(c, back, 0.005) {
		t.Fatalf("round-trip failed: got (%f,%f,%f,%f)", back.R(), back.G(), back.B(), back.A())
	}
}

func TestFromRGB255Boundaries(t *testing.T) {
	black := FromRGB255(0, 0, 0, 255)
	if black.R() != 0 || black.G() != 0 || black.B() != 0 || black.A() != 1 {
		t.Fatalf("black: got (%f,%f,%f,%f)", black.R(), black.G(), black.B(), black.A())
	}
	white := FromRGB255(255, 255, 255, 255)
	if white.R() != 1 || white.G() != 1 || white.B() != 1 || white.A() != 1 {
		t.Fatalf("white: got (%f,%f,%f,%f)", white.R(), white.G(), white.B(), white.A())
	}
}

// ── HSV ───────────────────────────────────────────────────────────────────────

func TestToHSVPureRed(t *testing.T) {
	c := sg.RGBA(1, 0, 0, 1)
	h, s, v, a := ToHSV(c)
	if !approxEq(h, 0, 0.001) || !approxEq(s, 1, 0.001) || !approxEq(v, 1, 0.001) || !approxEq(a, 1, 0.001) {
		t.Fatalf("red HSV: got (%f,%f,%f,%f), want (0,1,1,1)", h, s, v, a)
	}
}

func TestToHSVPureGreen(t *testing.T) {
	c := sg.RGBA(0, 1, 0, 1)
	h, s, v, a := ToHSV(c)
	if !approxEq(h, 1.0/3.0, 0.001) || !approxEq(s, 1, 0.001) || !approxEq(v, 1, 0.001) {
		t.Fatalf("green HSV: got (%f,%f,%f,%f), want (0.333,1,1,1)", h, s, v, a)
	}
	_ = a
}

func TestToHSVPureBlue(t *testing.T) {
	c := sg.RGBA(0, 0, 1, 1)
	h, s, v, _ := ToHSV(c)
	if !approxEq(h, 2.0/3.0, 0.001) || !approxEq(s, 1, 0.001) || !approxEq(v, 1, 0.001) {
		t.Fatalf("blue HSV: got (%f,%f,%f)", h, s, v)
	}
}

func TestToHSVWhite(t *testing.T) {
	c := sg.RGBA(1, 1, 1, 1)
	h, s, v, _ := ToHSV(c)
	if !approxEq(s, 0, 0.001) || !approxEq(v, 1, 0.001) {
		t.Fatalf("white HSV: got (%f,%f,%f)", h, s, v)
	}
}

func TestToHSVBlack(t *testing.T) {
	c := sg.RGBA(0, 0, 0, 1)
	_, s, v, _ := ToHSV(c)
	if !approxEq(s, 0, 0.001) || !approxEq(v, 0, 0.001) {
		t.Fatalf("black HSV: got s=%f v=%f", s, v)
	}
}

func TestHSVRoundTrip(t *testing.T) {
	colors := []sg.Color{
		sg.RGBA(1, 0, 0, 1),
		sg.RGBA(0, 1, 0, 1),
		sg.RGBA(0, 0, 1, 1),
		sg.RGBA(0.5, 0.3, 0.8, 0.7),
		sg.RGBA(1, 0.5, 0, 1),
		sg.RGBA(0, 0, 0, 1),
		sg.RGBA(1, 1, 1, 1),
		sg.RGBA(0.5, 0.5, 0.5, 1),
	}
	for i, c := range colors {
		h, s, v, a := ToHSV(c)
		back := FromHSV(h, s, v, a)
		if !colorApproxEq(c, back, 0.01) {
			t.Errorf("HSV round-trip %d: (%f,%f,%f,%f) -> hsv(%f,%f,%f) -> (%f,%f,%f,%f)",
				i, c.R(), c.G(), c.B(), c.A(), h, s, v, back.R(), back.G(), back.B(), back.A())
		}
	}
}

func TestFromHSVMatchesWillow(t *testing.T) {
	// Verify our FromHSV wrapper matches sg.ColorFromHSV for the RGB channels
	for i := 0; i < 36; i++ {
		h := float64(i) / 36.0
		expected := sg.ColorFromHSV(h, 0.8, 0.9)
		got := FromHSV(h, 0.8, 0.9, 0.5)
		if !approxEq(got.R(), expected.R(), 0.001) ||
			!approxEq(got.G(), expected.G(), 0.001) ||
			!approxEq(got.B(), expected.B(), 0.001) {
			t.Errorf("hue %f: RGB mismatch", h)
		}
		if !approxEq(got.A(), 0.5, 0.001) {
			t.Errorf("hue %f: alpha should be 0.5, got %f", h, got.A())
		}
	}
}

// ── HSL ───────────────────────────────────────────────────────────────────────

func TestToHSLPureRed(t *testing.T) {
	c := sg.RGBA(1, 0, 0, 1)
	h, s, l, _ := ToHSL(c)
	if !approxEq(h, 0, 0.001) || !approxEq(s, 1, 0.001) || !approxEq(l, 0.5, 0.001) {
		t.Fatalf("red HSL: got (%f,%f,%f), want (0,1,0.5)", h, s, l)
	}
}

func TestToHSLWhite(t *testing.T) {
	c := sg.RGBA(1, 1, 1, 1)
	_, s, l, _ := ToHSL(c)
	if !approxEq(s, 0, 0.001) || !approxEq(l, 1, 0.001) {
		t.Fatalf("white HSL: got s=%f l=%f", s, l)
	}
}

func TestToHSLBlack(t *testing.T) {
	c := sg.RGBA(0, 0, 0, 1)
	_, s, l, _ := ToHSL(c)
	if !approxEq(s, 0, 0.001) || !approxEq(l, 0, 0.001) {
		t.Fatalf("black HSL: got s=%f l=%f", s, l)
	}
}

func TestHSLRoundTrip(t *testing.T) {
	colors := []sg.Color{
		sg.RGBA(1, 0, 0, 1),
		sg.RGBA(0, 1, 0, 1),
		sg.RGBA(0, 0, 1, 1),
		sg.RGBA(0.5, 0.3, 0.8, 0.7),
		sg.RGBA(1, 0.5, 0, 1),
		sg.RGBA(0, 0, 0, 1),
		sg.RGBA(1, 1, 1, 1),
		sg.RGBA(0.5, 0.5, 0.5, 1),
		sg.RGBA(0.2, 0.6, 0.9, 1),
	}
	for i, c := range colors {
		h, s, l, a := ToHSL(c)
		back := FromHSL(h, s, l, a)
		if !colorApproxEq(c, back, 0.01) {
			t.Errorf("HSL round-trip %d: (%f,%f,%f,%f) -> hsl(%f,%f,%f) -> (%f,%f,%f,%f)",
				i, c.R(), c.G(), c.B(), c.A(), h, s, l, back.R(), back.G(), back.B(), back.A())
		}
	}
}

func TestFromHSLGray(t *testing.T) {
	// s=0 should produce gray
	c := FromHSL(0.5, 0, 0.5, 1)
	if !approxEq(c.R(), 0.5, 0.001) || !approxEq(c.G(), 0.5, 0.001) || !approxEq(c.B(), 0.5, 0.001) {
		t.Fatalf("gray: got (%f,%f,%f)", c.R(), c.G(), c.B())
	}
}

// ── Cross-space consistency ──────────────────────────────────────────────────

func TestHSVToHSLConsistency(t *testing.T) {
	// Pure red should be the same through both spaces
	c := sg.RGBA(1, 0, 0, 1)

	hv, sv, vv, _ := ToHSV(c)
	hl, sl, ll, _ := ToHSL(c)

	// Hue should match
	if !approxEq(hv, hl, 0.001) {
		t.Fatalf("hue mismatch: HSV=%f HSL=%f", hv, hl)
	}
	_ = sv
	_ = vv
	_ = sl
	_ = ll
}

// ── SV gradient alignment verification ──────────────────────────────────────

func TestSVGradientCorners(t *testing.T) {
	// The SV gradient for a given hue should have these corners:
	// Top-left (s=0, v=1): white
	// Top-right (s=1, v=1): pure hue color
	// Bottom-left (s=0, v=0): black
	// Bottom-right (s=1, v=0): black

	hue := 0.0                           // red
	topLeft := FromHSV(hue, 0, 1, 1)     // white
	topRight := FromHSV(hue, 1, 1, 1)    // red
	bottomLeft := FromHSV(hue, 0, 0, 1)  // black
	bottomRight := FromHSV(hue, 1, 0, 1) // black

	if !colorApproxEq(topLeft, sg.RGBA(1, 1, 1, 1), 0.01) {
		t.Errorf("top-left (s=0,v=1) should be white, got (%f,%f,%f)", topLeft.R(), topLeft.G(), topLeft.B())
	}
	if !colorApproxEq(topRight, sg.RGBA(1, 0, 0, 1), 0.01) {
		t.Errorf("top-right (s=1,v=1) should be red, got (%f,%f,%f)", topRight.R(), topRight.G(), topRight.B())
	}
	if !colorApproxEq(bottomLeft, sg.RGBA(0, 0, 0, 1), 0.01) {
		t.Errorf("bottom-left (s=0,v=0) should be black, got (%f,%f,%f)", bottomLeft.R(), bottomLeft.G(), bottomLeft.B())
	}
	if !colorApproxEq(bottomRight, sg.RGBA(0, 0, 0, 1), 0.01) {
		t.Errorf("bottom-right (s=1,v=0) should be black, got (%f,%f,%f)", bottomRight.R(), bottomRight.G(), bottomRight.B())
	}
}

func TestSVGradientPickPosition(t *testing.T) {
	// Verify that picking at a specific position in the SV field produces
	// the expected color. This tests the alignment between the gradient image
	// generation and the coordinate-to-HSV mapping.
	hue := 0.0 // red

	// Center of field: s=0.5, v=0.5
	center := FromHSV(hue, 0.5, 0.5, 1)
	ch, cs, cv, _ := ToHSV(center)
	if !approxEq(ch, hue, 0.01) || !approxEq(cs, 0.5, 0.01) || !approxEq(cv, 0.5, 0.01) {
		t.Errorf("center round-trip: got hsv(%f,%f,%f), want (0, 0.5, 0.5)", ch, cs, cv)
	}

	// Quarter point: s=0.25, v=0.75
	quarter := FromHSV(hue, 0.25, 0.75, 1)
	qh, qs, qv, _ := ToHSV(quarter)
	if !approxEq(qh, hue, 0.01) || !approxEq(qs, 0.25, 0.01) || !approxEq(qv, 0.75, 0.01) {
		t.Errorf("quarter round-trip: got hsv(%f,%f,%f), want (0, 0.25, 0.75)", qh, qs, qv)
	}
}

func TestSVGradientMultipleHues(t *testing.T) {
	hues := []float64{0, 0.1, 0.25, 0.333, 0.5, 0.667, 0.75, 0.9}
	for _, hue := range hues {
		// Pick at s=0.7, v=0.8
		c := FromHSV(hue, 0.7, 0.8, 1)
		h2, s2, v2, _ := ToHSV(c)
		if !approxEq(h2, hue, 0.01) || !approxEq(s2, 0.7, 0.01) || !approxEq(v2, 0.8, 0.01) {
			t.Errorf("hue %f: got hsv(%f,%f,%f), want (%f, 0.7, 0.8)", hue, h2, s2, v2, hue)
		}
	}
}

// ── Alpha preservation ──────────────────────────────────────────────────────

func TestAlphaPreservedThroughHSV(t *testing.T) {
	c := sg.RGBA(0.5, 0.3, 0.8, 0.42)
	h, s, v, a := ToHSV(c)
	if !approxEq(a, 0.42, 0.001) {
		t.Fatalf("ToHSV alpha: got %f, want 0.42", a)
	}
	back := FromHSV(h, s, v, a)
	if !approxEq(back.A(), 0.42, 0.01) {
		t.Fatalf("FromHSV alpha: got %f, want 0.42", back.A())
	}
}

func TestAlphaPreservedThroughHSL(t *testing.T) {
	c := sg.RGBA(0.5, 0.3, 0.8, 0.42)
	h, s, l, a := ToHSL(c)
	if !approxEq(a, 0.42, 0.001) {
		t.Fatalf("ToHSL alpha: got %f, want 0.42", a)
	}
	back := FromHSL(h, s, l, a)
	if !approxEq(back.A(), 0.42, 0.01) {
		t.Fatalf("FromHSL alpha: got %f, want 0.42", back.A())
	}
}

// ── NormalizeRGB ─────────────────────────────────────────────────────────────

func TestNormalizeRGBNormalRange(t *testing.T) {
	// All components already in [0,1]: should pass through unchanged.
	c := sg.RGBA(0.2, 0.5, 0.8, 0.9)
	n := NormalizeRGB(c)
	if !colorApproxEq(n, c, 0.001) {
		t.Fatalf("normal range: got (%f,%f,%f,%f), want (%f,%f,%f,%f)",
			n.R(), n.G(), n.B(), n.A(), c.R(), c.G(), c.B(), c.A())
	}
}

func TestNormalizeRGBOverbright(t *testing.T) {
	// Overbright: max component is 2.0, so all are scaled by 1/2.
	c := sg.RGBA(2.0, 1.0, 0.5, 0.7)
	n := NormalizeRGB(c)
	if !approxEq(n.R(), 1.0, 0.001) {
		t.Errorf("overbright R: got %f, want 1.0", n.R())
	}
	if !approxEq(n.G(), 0.5, 0.001) {
		t.Errorf("overbright G: got %f, want 0.5", n.G())
	}
	if !approxEq(n.B(), 0.25, 0.001) {
		t.Errorf("overbright B: got %f, want 0.25", n.B())
	}
	if !approxEq(n.A(), 0.7, 0.001) {
		t.Errorf("overbright A: got %f, want 0.7 (alpha unchanged)", n.A())
	}
}

func TestNormalizeRGBNegativeComponents(t *testing.T) {
	// Negative components should be clamped to 0 when maxC <= 1.
	c := sg.RGBA(-0.5, 0.3, 0.8, 1.0)
	n := NormalizeRGB(c)
	if !approxEq(n.R(), 0.0, 0.001) {
		t.Errorf("negative R: got %f, want 0.0", n.R())
	}
	if !approxEq(n.G(), 0.3, 0.001) {
		t.Errorf("G: got %f, want 0.3", n.G())
	}
	if !approxEq(n.B(), 0.8, 0.001) {
		t.Errorf("B: got %f, want 0.8", n.B())
	}
}

func TestNormalizeRGBAllZero(t *testing.T) {
	c := sg.RGBA(0, 0, 0, 1)
	n := NormalizeRGB(c)
	if !colorApproxEq(n, c, 0.001) {
		t.Fatalf("all-zero RGB: got (%f,%f,%f,%f)", n.R(), n.G(), n.B(), n.A())
	}
}

func TestNormalizeRGBExactlyOne(t *testing.T) {
	// maxC == 1 is the boundary; should take the normal (non-overbright) path.
	c := sg.RGBA(1.0, 0.5, 0.0, 1.0)
	n := NormalizeRGB(c)
	if !colorApproxEq(n, c, 0.001) {
		t.Fatalf("maxC==1: got (%f,%f,%f,%f)", n.R(), n.G(), n.B(), n.A())
	}
}

// ── clampInt edge cases ──────────────────────────────────────────────────────

func TestClampIntBelowLo(t *testing.T) {
	// Trigger clamping via ToRGB255 with a color that has negative components.
	c := sg.RGBA(-1.0, 0, 0, 1)
	r, _, _, _ := ToRGB255(c)
	if r != 0 {
		t.Fatalf("below lo: got r=%d, want 0", r)
	}
}

func TestClampIntAboveHi(t *testing.T) {
	// Trigger clamping via ToRGB255 with overbright component.
	c := sg.RGBA(2.0, 0, 0, 1)
	r, _, _, _ := ToRGB255(c)
	if r != 255 {
		t.Fatalf("above hi: got r=%d, want 255", r)
	}
}

func TestClampIntInRange(t *testing.T) {
	// Value exactly at boundaries should not be clamped.
	c := sg.RGBA(0, 1, 0, 0)
	r, g, _, a := ToRGB255(c)
	if r != 0 {
		t.Fatalf("lo boundary: got r=%d, want 0", r)
	}
	if g != 255 {
		t.Fatalf("hi boundary: got g=%d, want 255", g)
	}
	if a != 0 {
		t.Fatalf("zero alpha: got a=%d, want 0", a)
	}
}

// ── SampleBilinear ──────────────────────────────────────────────────────────

func TestSampleBilinearCorners(t *testing.T) {
	g := render.GradientColors{
		TopLeft:     sg.RGBA(1, 0, 0, 1), // red
		TopRight:    sg.RGBA(0, 1, 0, 1), // green
		BottomLeft:  sg.RGBA(0, 0, 1, 1), // blue
		BottomRight: sg.RGBA(1, 1, 1, 1), // white
	}

	tl := SampleBilinear(g, 0, 0)
	if !colorApproxEq(tl, g.TopLeft, 0.01) {
		t.Errorf("(0,0) should be TopLeft red, got (%f,%f,%f,%f)", tl.R(), tl.G(), tl.B(), tl.A())
	}

	tr := SampleBilinear(g, 1, 0)
	if !colorApproxEq(tr, g.TopRight, 0.01) {
		t.Errorf("(1,0) should be TopRight green, got (%f,%f,%f,%f)", tr.R(), tr.G(), tr.B(), tr.A())
	}

	bl := SampleBilinear(g, 0, 1)
	if !colorApproxEq(bl, g.BottomLeft, 0.01) {
		t.Errorf("(0,1) should be BottomLeft blue, got (%f,%f,%f,%f)", bl.R(), bl.G(), bl.B(), bl.A())
	}

	br := SampleBilinear(g, 1, 1)
	if !colorApproxEq(br, g.BottomRight, 0.01) {
		t.Errorf("(1,1) should be BottomRight white, got (%f,%f,%f,%f)", br.R(), br.G(), br.B(), br.A())
	}
}

func TestSampleBilinearCenter(t *testing.T) {
	// Uniform color: center should equal any corner.
	g := render.GradientColors{
		TopLeft:     sg.RGBA(0.4, 0.4, 0.4, 1),
		TopRight:    sg.RGBA(0.4, 0.4, 0.4, 1),
		BottomLeft:  sg.RGBA(0.4, 0.4, 0.4, 1),
		BottomRight: sg.RGBA(0.4, 0.4, 0.4, 1),
	}
	c := SampleBilinear(g, 0.5, 0.5)
	if !colorApproxEq(c, sg.RGBA(0.4, 0.4, 0.4, 1), 0.01) {
		t.Errorf("uniform center: got (%f,%f,%f,%f)", c.R(), c.G(), c.B(), c.A())
	}
}

func TestSampleBilinearMidpoint(t *testing.T) {
	// Horizontal black-to-white: midpoint should be gray.
	g := render.GradientColors{
		TopLeft:     sg.RGBA(0, 0, 0, 1),
		TopRight:    sg.RGBA(1, 1, 1, 1),
		BottomLeft:  sg.RGBA(0, 0, 0, 1),
		BottomRight: sg.RGBA(1, 1, 1, 1),
	}
	c := SampleBilinear(g, 0.5, 0.5)
	if !approxEq(c.R(), 0.5, 0.01) || !approxEq(c.G(), 0.5, 0.01) || !approxEq(c.B(), 0.5, 0.01) {
		t.Errorf("midpoint should be gray, got (%f,%f,%f)", c.R(), c.G(), c.B())
	}
}

// ── FormatGradientString ────────────────────────────────────────────────────

func TestFormatGradientStringHorizontal(t *testing.T) {
	g := render.Gradient{
		Mode: render.GradientModeH,
		Colors: render.GradientColors{
			TopLeft:  sg.RGBA(0, 0, 0, 1),
			TopRight: sg.RGBA(1, 1, 1, 1),
		},
	}
	got := FormatGradientString(g)
	want := "gradientH(#000000, #FFFFFF)"
	if got != want {
		t.Fatalf("horizontal: got %q, want %q", got, want)
	}
}

func TestFormatGradientStringVertical(t *testing.T) {
	g := render.Gradient{
		Mode: render.GradientModeV,
		Colors: render.GradientColors{
			TopLeft:    sg.RGBA(1, 0, 0, 1),
			BottomLeft: sg.RGBA(0, 0, 1, 1),
		},
	}
	got := FormatGradientString(g)
	want := "gradientV(#FF0000, #0000FF)"
	if got != want {
		t.Fatalf("vertical: got %q, want %q", got, want)
	}
}

func TestFormatGradientStringFourCorner(t *testing.T) {
	g := render.Gradient{
		Mode: render.GradientModeFourCorner,
		Colors: render.GradientColors{
			TopLeft:     sg.RGBA(1, 0, 0, 1),
			TopRight:    sg.RGBA(0, 1, 0, 1),
			BottomRight: sg.RGBA(0, 0, 1, 1),
			BottomLeft:  sg.RGBA(1, 1, 0, 1),
		},
	}
	got := FormatGradientString(g)
	want := "gradient(#FF0000, #00FF00, #0000FF, #FFFF00)"
	if got != want {
		t.Fatalf("four-corner: got %q, want %q", got, want)
	}
}

// ── DefaultGradient ─────────────────────────────────────────────────────────

func TestDefaultGradient(t *testing.T) {
	g := DefaultGradient()
	if g.Mode != render.GradientModeH {
		t.Fatalf("mode: got %d, want GradientModeH (%d)", g.Mode, render.GradientModeH)
	}

	black := sg.RGBA(0, 0, 0, 1)
	white := sg.RGBA(1, 1, 1, 1)

	if !colorApproxEq(g.Colors.TopLeft, black, 0.001) {
		t.Errorf("TopLeft should be black")
	}
	if !colorApproxEq(g.Colors.BottomLeft, black, 0.001) {
		t.Errorf("BottomLeft should be black")
	}
	if !colorApproxEq(g.Colors.TopRight, white, 0.001) {
		t.Errorf("TopRight should be white")
	}
	if !colorApproxEq(g.Colors.BottomRight, white, 0.001) {
		t.Errorf("BottomRight should be white")
	}
}

// ── HSL additional branches ─────────────────────────────────────────────────

func TestToHSLHighLightness(t *testing.T) {
	// l >= 0.5 triggers the else branch for saturation calculation.
	c := sg.RGBA(0.9, 0.7, 0.8, 1)
	h, s, l, _ := ToHSL(c)
	back := FromHSL(h, s, l, 1)
	if !colorApproxEq(c, back, 0.01) {
		t.Fatalf("high-lightness round-trip: got (%f,%f,%f), want (%f,%f,%f)",
			back.R(), back.G(), back.B(), c.R(), c.G(), c.B())
	}
	if l < 0.5 {
		t.Fatalf("expected l >= 0.5, got %f", l)
	}
}

func TestToHSLGreenHue(t *testing.T) {
	// Green-dominant color triggers the maxC==g branch in ToHSL.
	c := sg.RGBA(0.2, 0.8, 0.3, 1)
	h, s, l, _ := ToHSL(c)
	back := FromHSL(h, s, l, 1)
	if !colorApproxEq(c, back, 0.01) {
		t.Fatalf("green-hue round-trip failed: got (%f,%f,%f)", back.R(), back.G(), back.B())
	}
}

func TestToHSLBlueHue(t *testing.T) {
	// Blue-dominant color triggers the maxC==b branch in ToHSL.
	c := sg.RGBA(0.1, 0.2, 0.9, 1)
	h, s, l, _ := ToHSL(c)
	back := FromHSL(h, s, l, 1)
	if !colorApproxEq(c, back, 0.01) {
		t.Fatalf("blue-hue round-trip failed: got (%f,%f,%f)", back.R(), back.G(), back.B())
	}
}

func TestFromHSLHighLightness(t *testing.T) {
	// l >= 0.5 triggers else branch in FromHSL.
	c := FromHSL(0.0, 0.5, 0.75, 1)
	h, s, l, _ := ToHSL(c)
	back := FromHSL(h, s, l, 1)
	if !colorApproxEq(c, back, 0.01) {
		t.Fatalf("FromHSL high-lightness round-trip failed")
	}
}

// ── HSV additional branches ─────────────────────────────────────────────────

func TestToHSVRedWrapAround(t *testing.T) {
	// Color where maxC==R but (g-b)/delta < 0, triggering h += 6 branch.
	// Magenta-ish: R=0.8, G=0.1, B=0.7 -> maxC=R, g-b = 0.1-0.7 = -0.6 < 0
	c := sg.RGBA(0.8, 0.1, 0.7, 1)
	h, s, v, _ := ToHSV(c)
	back := FromHSV(h, s, v, 1)
	if !colorApproxEq(c, back, 0.01) {
		t.Fatalf("red-wrap round-trip failed: got (%f,%f,%f)", back.R(), back.G(), back.B())
	}
	// Hue should be near 1.0 (wrapping around from red toward magenta).
	if h < 0.8 {
		t.Fatalf("expected high hue for magenta-ish, got %f", h)
	}
}

// ── ParseHex 8-char invalid ─────────────────────────────────────────────────

func TestParseHex8Invalid(t *testing.T) {
	// 8-char hex with invalid characters should fail.
	if _, ok := ParseHex("#GGGGGGGG"); ok {
		t.Fatal("expected ParseHex(#GGGGGGGG) to fail")
	}
}

// ── ToHSL red hue wrap-around ───────────────────────────────────────────────

func TestToHSLRedHueWrapAround(t *testing.T) {
	// When maxC==R and g < b, hue goes negative and needs += 6.
	// Magenta: R high, G low, B medium-high.
	c := sg.RGBA(0.9, 0.1, 0.8, 1)
	h, s, l, _ := ToHSL(c)
	back := FromHSL(h, s, l, 1)
	if !colorApproxEq(c, back, 0.01) {
		t.Fatalf("HSL red-wrap round-trip failed: got (%f,%f,%f)", back.R(), back.G(), back.B())
	}
	// Hue should be high (near 1.0), indicating magenta region.
	if h < 0.8 {
		t.Fatalf("expected high hue for magenta-ish color, got %f", h)
	}
}

// ── FromHSL low lightness branch ────────────────────────────────────────────

func TestFromHSLLowLightness(t *testing.T) {
	// l < 0.5 with s > 0 triggers the q = l*(1+s) branch.
	c := FromHSL(0.6, 0.8, 0.3, 1)
	h, s, l, _ := ToHSL(c)
	if l >= 0.5 {
		t.Fatalf("expected l < 0.5, got %f", l)
	}
	back := FromHSL(h, s, l, 1)
	if !colorApproxEq(c, back, 0.01) {
		t.Fatalf("FromHSL low-lightness round-trip failed: got (%f,%f,%f)", back.R(), back.G(), back.B())
	}
}
