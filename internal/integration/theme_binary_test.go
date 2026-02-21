package integration

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	ui "github.com/devthicket/willowui"
	interntheme "github.com/devthicket/willowui/internal/theme"
)

// ---------------------------------------------------------------------------
// WUIT binary encode/decode round-trip
// ---------------------------------------------------------------------------

func TestEncodeDecodeThemeBinary_RoundTrip(t *testing.T) {
	themeJSON := []byte(`{"colors":{"bg":"#1a1a1a"},"components":{"button":{"primary":{"backgroundColor":"$bg"}}}}`)
	atlasJSON := []byte(`{"frames":{},"meta":{"image":"atlas.png","size":{"w":64,"h":64}}}`)
	atlasPNG := makeTinyPNG(t, 64, 64)

	encoded, err := interntheme.EncodeThemeBinary(themeJSON, atlasJSON, atlasPNG)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	decoded, err := interntheme.DecodeThemeBinary(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if string(decoded.ThemeJSON) != string(themeJSON) {
		t.Errorf("theme JSON mismatch: got %q, want %q", decoded.ThemeJSON, themeJSON)
	}
	if string(decoded.AtlasJSON) != string(atlasJSON) {
		t.Errorf("atlas JSON mismatch: got %q, want %q", decoded.AtlasJSON, atlasJSON)
	}
	if len(decoded.AtlasPNG) != len(atlasPNG) {
		t.Errorf("atlas PNG length mismatch: got %d, want %d", len(decoded.AtlasPNG), len(atlasPNG))
	}
}

func TestEncodeDecodeThemeBinary_NoAtlas(t *testing.T) {
	themeJSON := []byte(`{"colors":{}}`)

	encoded, err := interntheme.EncodeThemeBinary(themeJSON, nil, nil)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	decoded, err := interntheme.DecodeThemeBinary(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if string(decoded.ThemeJSON) != string(themeJSON) {
		t.Errorf("theme JSON mismatch")
	}
	if decoded.AtlasJSON != nil {
		t.Errorf("expected nil atlas JSON, got %d bytes", len(decoded.AtlasJSON))
	}
	if decoded.AtlasPNG != nil {
		t.Errorf("expected nil atlas PNG, got %d bytes", len(decoded.AtlasPNG))
	}
}

// ---------------------------------------------------------------------------
// Error cases
// ---------------------------------------------------------------------------

func TestDecodeThemeBinary_TooShort(t *testing.T) {
	_, err := interntheme.DecodeThemeBinary([]byte{0, 1, 2})
	if err == nil {
		t.Fatal("expected error for short data")
	}
}

func TestDecodeThemeBinary_BadMagic(t *testing.T) {
	data := make([]byte, 18)
	copy(data, "XXXX")
	_, err := interntheme.DecodeThemeBinary(data)
	if err == nil {
		t.Fatal("expected error for bad magic")
	}
}

func TestDecodeThemeBinary_BadVersion(t *testing.T) {
	data := make([]byte, 18)
	copy(data, "WUIT")
	data[4] = 99 // bad version
	_, err := interntheme.DecodeThemeBinary(data)
	if err == nil {
		t.Fatal("expected error for bad version")
	}
}

func TestDecodeThemeBinary_Truncated(t *testing.T) {
	themeJSON := []byte(`{}`)
	encoded, _ := interntheme.EncodeThemeBinary(themeJSON, nil, nil)
	// Truncate the data.
	_, err := interntheme.DecodeThemeBinary(encoded[:len(encoded)-1])
	if err == nil {
		t.Fatal("expected error for truncated data")
	}
}

func TestEncodeThemeBinary_NilJSON(t *testing.T) {
	_, err := interntheme.EncodeThemeBinary(nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil theme JSON")
	}
}

// ---------------------------------------------------------------------------
// Default glyph splitting
// ---------------------------------------------------------------------------

func TestSplitDefaultGlyphs(t *testing.T) {
	glyphsPNG, err := os.ReadFile("../../assets/icons/default-glyphs.png")
	if err != nil {
		t.Skipf("default-glyphs.png not found: %v", err)
	}

	glyphs, err := interntheme.SplitDefaultGlyphs(glyphsPNG)
	if err != nil {
		t.Fatalf("split: %v", err)
	}

	if len(glyphs) != len(interntheme.DefaultGlyphNames) {
		t.Errorf("glyph count: got %d, want %d", len(glyphs), len(interntheme.DefaultGlyphNames))
	}

	for _, name := range interntheme.DefaultGlyphNames {
		img, ok := glyphs[name]
		if !ok {
			t.Errorf("missing glyph %q", name)
			continue
		}
		b := img.Bounds()
		if b.Dx() != 48 || b.Dy() != 48 {
			t.Errorf("glyph %q: got %dx%d, want 48x48", name, b.Dx(), b.Dy())
		}
	}
}

// ---------------------------------------------------------------------------
// Atlas packing
// ---------------------------------------------------------------------------

func TestCompileThemeAtlas_DefaultGlyphs(t *testing.T) {
	glyphsPNG, err := os.ReadFile("../../assets/icons/default-glyphs.png")
	if err != nil {
		t.Skipf("default-glyphs.png not found: %v", err)
	}

	input := &interntheme.ThemeAtlasInput{
		DefaultGlyphsPNG: glyphsPNG,
	}

	out, err := interntheme.CompileThemeAtlas(input)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if len(out.AtlasJSON) == 0 {
		t.Error("expected non-empty atlas JSON")
	}
	if len(out.AtlasPNG) == 0 {
		t.Error("expected non-empty atlas PNG")
	}
}

func TestCompileThemeAtlas_Empty(t *testing.T) {
	input := &interntheme.ThemeAtlasInput{}
	out, err := interntheme.CompileThemeAtlas(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Error("expected nil output for empty input")
	}
}

// ---------------------------------------------------------------------------
// Full round-trip: JSON → binary → load
// ---------------------------------------------------------------------------

func TestLoadThemeBinary_RoundTrip(t *testing.T) {
	themeJSON := []byte(`{
		"colors": {
			"primary": "#4488FF",
			"bg": "#1a1a2e"
		},
		"fonts": {
			"body": "gofont",
			"heading": "gofont-bold"
		},
		"components": {
			"button": {
				"primary": {
					"backgroundColor": {"default": "$primary"},
					"textColor": {"default": "#ffffff"}
				}
			}
		}
	}`)

	// Encode with no atlas.
	encoded, err := interntheme.EncodeThemeBinary(themeJSON, nil, nil)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Load from binary.
	th, err := ui.LoadThemeBinary(encoded)
	if err != nil {
		t.Fatalf("load binary: %v", err)
	}

	// Verify the button background was compiled correctly.
	bg := th.Button.Primary.Background[0] // StateDefault
	if bg.Type == 0 {
		t.Error("expected button background to be set")
	}

	// Verify fonts.
	if th.FontName("body") != "gofont" {
		t.Errorf("font body: got %q, want %q", th.FontName("body"), "gofont")
	}
	if th.FontName("heading") != "gofont-bold" {
		t.Errorf("font heading: got %q, want %q", th.FontName("heading"), "gofont-bold")
	}
	if th.FontName("missing") != "" {
		t.Errorf("font missing: got %q, want empty", th.FontName("missing"))
	}
}

// ---------------------------------------------------------------------------
// Full round-trip with atlas
// ---------------------------------------------------------------------------

func TestLoadThemeBinary_WithAtlas(t *testing.T) {
	glyphsPNG, err := os.ReadFile("../../assets/icons/default-glyphs.png")
	if err != nil {
		t.Skipf("default-glyphs.png not found: %v", err)
	}

	// Build atlas from default glyphs.
	input := &interntheme.ThemeAtlasInput{
		DefaultGlyphsPNG: glyphsPNG,
	}
	atlasOut, err := interntheme.CompileThemeAtlas(input)
	if err != nil {
		t.Fatalf("compile atlas: %v", err)
	}

	themeJSON := []byte(`{
		"colors": {"bg": "#1a1a2e"},
		"components": {
			"button": {
				"primary": {
					"backgroundColor": {"default": "$bg"}
				}
			}
		}
	}`)

	// Encode binary with atlas.
	encoded, err := interntheme.EncodeThemeBinary(themeJSON, atlasOut.AtlasJSON, atlasOut.AtlasPNG)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	// Verify it's a valid WUIT file.
	if len(encoded) < 18 {
		t.Fatal("encoded data too short")
	}
	if string(encoded[:4]) != "WUIT" {
		t.Errorf("magic: got %q, want WUIT", encoded[:4])
	}

	// Load from binary.
	th, err := ui.LoadThemeBinary(encoded)
	if err != nil {
		t.Fatalf("load binary: %v", err)
	}

	// The theme should have an atlas with the default icon regions.
	if th.Atlas == nil {
		t.Fatal("expected atlas to be set")
	}

	// Check that known glyph regions exist.
	for _, name := range []string{"chevron-right", "close-x", "password-dot"} {
		region := th.Atlas.Region(name)
		if region.Width == 0 || region.Height == 0 {
			t.Errorf("atlas region %q: got zero dimensions", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makeTinyPNG creates a minimal valid PNG of the given size.
func makeTinyPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := makeTestImageBinary(w, h)
	return encodePNGBinary(t, img)
}

func makeTestImageBinary(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(x & 0xFF), G: uint8(y & 0xFF), B: 128, A: 255})
		}
	}
	return img
}

func encodePNGBinary(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	return buf.Bytes()
}
