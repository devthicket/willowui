package theme

import (
	"encoding/json"
	"image"
	"image/color"
	"testing"
)

// ---------------------------------------------------------------------------
// nextPow2
// ---------------------------------------------------------------------------

func TestNextPow2(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{0, 64},
		{1, 64},
		{2, 64},
		{63, 64},
		{64, 64},
		{65, 128},
		{100, 128},
		{127, 128},
		{128, 128},
		{129, 256},
		{256, 256},
		{257, 512},
		{512, 512},
		{1000, 1024},
		{1024, 1024},
		{1025, 2048},
	}
	for _, tt := range tests {
		got := nextPow2(tt.input)
		if got != tt.want {
			t.Errorf("nextPow2(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// rectsOverlap
// ---------------------------------------------------------------------------

func TestRectsOverlap(t *testing.T) {
	tests := []struct {
		name string
		a, b packFreeRect
		want bool
	}{
		{
			name: "fully overlapping",
			a:    packFreeRect{0, 0, 10, 10},
			b:    packFreeRect{0, 0, 10, 10},
			want: true,
		},
		{
			name: "partial overlap",
			a:    packFreeRect{0, 0, 10, 10},
			b:    packFreeRect{5, 5, 10, 10},
			want: true,
		},
		{
			name: "a contains b",
			a:    packFreeRect{0, 0, 100, 100},
			b:    packFreeRect{10, 10, 5, 5},
			want: true,
		},
		{
			name: "touching right edge - no overlap",
			a:    packFreeRect{0, 0, 10, 10},
			b:    packFreeRect{10, 0, 10, 10},
			want: false,
		},
		{
			name: "touching bottom edge - no overlap",
			a:    packFreeRect{0, 0, 10, 10},
			b:    packFreeRect{0, 10, 10, 10},
			want: false,
		},
		{
			name: "far apart horizontally",
			a:    packFreeRect{0, 0, 10, 10},
			b:    packFreeRect{50, 0, 10, 10},
			want: false,
		},
		{
			name: "far apart vertically",
			a:    packFreeRect{0, 0, 10, 10},
			b:    packFreeRect{0, 50, 10, 10},
			want: false,
		},
		{
			name: "overlap by one pixel",
			a:    packFreeRect{0, 0, 10, 10},
			b:    packFreeRect{9, 9, 10, 10},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rectsOverlap(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("rectsOverlap(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
			// Overlap is symmetric.
			got2 := rectsOverlap(tt.b, tt.a)
			if got2 != tt.want {
				t.Errorf("rectsOverlap(%v, %v) symmetric = %v, want %v", tt.b, tt.a, got2, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// rectContains
// ---------------------------------------------------------------------------

func TestRectContains(t *testing.T) {
	tests := []struct {
		name        string
		outer, inner packFreeRect
		want        bool
	}{
		{
			name:  "same rect",
			outer: packFreeRect{0, 0, 10, 10},
			inner: packFreeRect{0, 0, 10, 10},
			want:  true,
		},
		{
			name:  "inner fully inside",
			outer: packFreeRect{0, 0, 100, 100},
			inner: packFreeRect{10, 10, 5, 5},
			want:  true,
		},
		{
			name:  "inner at top-left corner",
			outer: packFreeRect{0, 0, 100, 100},
			inner: packFreeRect{0, 0, 50, 50},
			want:  true,
		},
		{
			name:  "inner at bottom-right edge",
			outer: packFreeRect{0, 0, 100, 100},
			inner: packFreeRect{50, 50, 50, 50},
			want:  true,
		},
		{
			name:  "inner exceeds right",
			outer: packFreeRect{0, 0, 10, 10},
			inner: packFreeRect{5, 0, 10, 10},
			want:  false,
		},
		{
			name:  "inner exceeds bottom",
			outer: packFreeRect{0, 0, 10, 10},
			inner: packFreeRect{0, 5, 10, 10},
			want:  false,
		},
		{
			name:  "inner starts before outer",
			outer: packFreeRect{10, 10, 10, 10},
			inner: packFreeRect{5, 5, 10, 10},
			want:  false,
		},
		{
			name:  "completely disjoint",
			outer: packFreeRect{0, 0, 10, 10},
			inner: packFreeRect{50, 50, 10, 10},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rectContains(tt.outer, tt.inner)
			if got != tt.want {
				t.Errorf("rectContains(%v, %v) = %v, want %v", tt.outer, tt.inner, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// packAtlas - empty input
// ---------------------------------------------------------------------------

func TestPackAtlas_Empty(t *testing.T) {
	results, w, h, err := packAtlas(nil, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results, got %v", results)
	}
	if w != 0 || h != 0 {
		t.Errorf("expected 0x0 dimensions, got %dx%d", w, h)
	}
}

func TestPackAtlas_EmptyMap(t *testing.T) {
	results, w, h, err := packAtlas(map[string]image.Image{}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results, got %v", results)
	}
	if w != 0 || h != 0 {
		t.Errorf("expected 0x0 dimensions, got %dx%d", w, h)
	}
}

// ---------------------------------------------------------------------------
// packAtlas - single image
// ---------------------------------------------------------------------------

func makeTestImage(w, h int, c color.Color) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestPackAtlas_SingleImage(t *testing.T) {
	images := map[string]image.Image{
		"test": makeTestImage(32, 32, color.NRGBA{255, 0, 0, 255}),
	}

	results, w, h, err := packAtlas(images, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].name != "test" {
		t.Errorf("result name = %q, want %q", results[0].name, "test")
	}
	if results[0].width != 32 || results[0].height != 32 {
		t.Errorf("result size = %dx%d, want 32x32", results[0].width, results[0].height)
	}
	if w < 32 || h < 32 {
		t.Errorf("atlas dimensions %dx%d too small for 32x32 image", w, h)
	}
	// Dimensions should be powers of 2.
	if w&(w-1) != 0 {
		t.Errorf("atlas width %d is not a power of 2", w)
	}
	if h&(h-1) != 0 {
		t.Errorf("atlas height %d is not a power of 2", h)
	}
}

// ---------------------------------------------------------------------------
// packAtlas - multiple images, non-overlapping
// ---------------------------------------------------------------------------

func TestPackAtlas_MultipleImages_NonOverlapping(t *testing.T) {
	images := map[string]image.Image{
		"small":  makeTestImage(16, 16, color.NRGBA{255, 0, 0, 255}),
		"medium": makeTestImage(32, 32, color.NRGBA{0, 255, 0, 255}),
		"large":  makeTestImage(48, 48, color.NRGBA{0, 0, 255, 255}),
	}

	results, w, h, err := packAtlas(images, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Verify all names are present.
	names := make(map[string]bool)
	for _, r := range results {
		names[r.name] = true
	}
	for _, want := range []string{"small", "medium", "large"} {
		if !names[want] {
			t.Errorf("missing result %q", want)
		}
	}

	// Verify no two results overlap.
	for i := 0; i < len(results); i++ {
		ri := packFreeRect{results[i].x, results[i].y, results[i].width, results[i].height}
		for j := i + 1; j < len(results); j++ {
			rj := packFreeRect{results[j].x, results[j].y, results[j].width, results[j].height}
			if rectsOverlap(ri, rj) {
				t.Errorf("results %q and %q overlap: %v vs %v", results[i].name, results[j].name, ri, rj)
			}
		}
	}

	// All results must fit within atlas bounds.
	for _, r := range results {
		if r.x+r.width > w || r.y+r.height > h {
			t.Errorf("result %q at (%d,%d) size %dx%d exceeds atlas %dx%d",
				r.name, r.x, r.y, r.width, r.height, w, h)
		}
	}
}

// ---------------------------------------------------------------------------
// packAtlas - varying sizes
// ---------------------------------------------------------------------------

func TestPackAtlas_VaryingSizes(t *testing.T) {
	images := map[string]image.Image{
		"tiny":   makeTestImage(4, 4, color.NRGBA{255, 0, 0, 255}),
		"wide":   makeTestImage(60, 10, color.NRGBA{0, 255, 0, 255}),
		"tall":   makeTestImage(10, 60, color.NRGBA{0, 0, 255, 255}),
		"square": makeTestImage(30, 30, color.NRGBA{255, 255, 0, 255}),
	}

	results, w, h, err := packAtlas(images, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}

	// All must fit within bounds.
	for _, r := range results {
		if r.x < 0 || r.y < 0 {
			t.Errorf("result %q has negative position: (%d,%d)", r.name, r.x, r.y)
		}
		if r.x+r.width > w || r.y+r.height > h {
			t.Errorf("result %q exceeds atlas bounds", r.name)
		}
	}
}

// ---------------------------------------------------------------------------
// ComposeAtlasImage
// ---------------------------------------------------------------------------

func TestComposeAtlasImage(t *testing.T) {
	red := color.NRGBA{255, 0, 0, 255}
	results := []packResult{
		{name: "a", img: makeTestImage(10, 10, red), x: 0, y: 0, width: 10, height: 10},
		{name: "b", img: makeTestImage(10, 10, red), x: 20, y: 0, width: 10, height: 10},
	}

	img := ComposeAtlasImage(results, 64, 64)
	if img == nil {
		t.Fatal("ComposeAtlasImage returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("atlas image size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
	}

	// Check that pixels at placed positions are red.
	r, g, b, a := img.At(5, 5).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
		t.Errorf("pixel at (5,5) = (%d,%d,%d,%d), want red", r>>8, g>>8, b>>8, a>>8)
	}

	// Check that a pixel outside placed areas is transparent.
	r2, g2, b2, a2 := img.At(15, 15).RGBA()
	if a2 != 0 {
		t.Errorf("pixel at (15,15) = (%d,%d,%d,%d), want transparent", r2>>8, g2>>8, b2>>8, a2>>8)
	}
}

func TestComposeAtlasImage_Empty(t *testing.T) {
	img := ComposeAtlasImage(nil, 64, 64)
	if img == nil {
		t.Fatal("ComposeAtlasImage returned nil for empty results")
	}
	// Should be a blank 64x64 image.
	if img.Bounds().Dx() != 64 || img.Bounds().Dy() != 64 {
		t.Errorf("size = %dx%d, want 64x64", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

// ---------------------------------------------------------------------------
// EncodeAtlasPNG
// ---------------------------------------------------------------------------

func TestEncodeAtlasPNG(t *testing.T) {
	img := makeTestImage(16, 16, color.NRGBA{128, 128, 128, 255})
	data, err := EncodeAtlasPNG(img)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("encoded PNG is empty")
	}
	// PNG signature: 0x89 P N G
	if data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
		t.Errorf("PNG header = %x, want PNG signature", data[:4])
	}
}

// ---------------------------------------------------------------------------
// GenerateAtlasJSON
// ---------------------------------------------------------------------------

func TestGenerateAtlasJSON(t *testing.T) {
	results := []packResult{
		{name: "icon-a", x: 0, y: 0, width: 48, height: 48},
		{name: "icon-b", x: 48, y: 0, width: 32, height: 32},
	}

	data, err := GenerateAtlasJSON(results, 128, 64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var root atlasJSONRoot
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Check meta.
	if root.Meta.Image != "atlas.png" {
		t.Errorf("meta.image = %q, want %q", root.Meta.Image, "atlas.png")
	}
	if root.Meta.Size.W != 128 || root.Meta.Size.H != 64 {
		t.Errorf("meta.size = %dx%d, want 128x64", root.Meta.Size.W, root.Meta.Size.H)
	}

	// Check frames.
	if len(root.Frames) != 2 {
		t.Fatalf("frames count = %d, want 2", len(root.Frames))
	}

	frameA, ok := root.Frames["icon-a"]
	if !ok {
		t.Fatal("missing frame icon-a")
	}
	if frameA.Frame.X != 0 || frameA.Frame.Y != 0 || frameA.Frame.W != 48 || frameA.Frame.H != 48 {
		t.Errorf("icon-a frame = %+v, want {0,0,48,48}", frameA.Frame)
	}
	if frameA.SourceSize.W != 48 || frameA.SourceSize.H != 48 {
		t.Errorf("icon-a sourceSize = %+v, want {48,48}", frameA.SourceSize)
	}
	if frameA.Rotated || frameA.Trimmed {
		t.Error("icon-a should not be rotated or trimmed")
	}

	frameB, ok := root.Frames["icon-b"]
	if !ok {
		t.Fatal("missing frame icon-b")
	}
	if frameB.Frame.X != 48 || frameB.Frame.W != 32 {
		t.Errorf("icon-b frame = %+v, want x=48, w=32", frameB.Frame)
	}
}

func TestGenerateAtlasJSON_Empty(t *testing.T) {
	data, err := GenerateAtlasJSON(nil, 64, 64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var root atlasJSONRoot
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(root.Frames) != 0 {
		t.Errorf("expected 0 frames, got %d", len(root.Frames))
	}
}
