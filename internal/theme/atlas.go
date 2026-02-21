package theme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"sort"
)

// ---------------------------------------------------------------------------
// Default glyph names & layout
// ---------------------------------------------------------------------------

// DefaultGlyphSize is the pixel dimension of each icon cell in the default
// glyphs spritesheet (48x48 per cell, packed left-to-right).
const DefaultGlyphSize = 48

// DefaultGlyphNames lists every icon in the default spritesheet in order.
// Index i occupies pixels [i*48, 0] to [(i+1)*48, 48].
var DefaultGlyphNames = []string{
	"chevron-right",
	"chevron-down",
	"chevron-left",
	"chevron-up",
	"arrow-up",
	"arrow-down",
	"close-x",
	"plus",
	"minus",
	"checkmark",
	"search",
	"hamburger",
	"filter",
	"radio-dot",
	"grip-dots-v",
	"grip-dots-h",
	"grip-dots-square",
	"grip-lines-v",
	"grip-lines-h",
	"password-dot",
}

// SplitDefaultGlyphs decodes pngData (the default-glyphs.png spritesheet)
// and returns individual 48x48 sub-images keyed by name.
func SplitDefaultGlyphs(pngData []byte) (map[string]image.Image, error) {
	src, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("decode default glyphs: %w", err)
	}

	result := make(map[string]image.Image, len(DefaultGlyphNames))
	for i, name := range DefaultGlyphNames {
		x0 := i * DefaultGlyphSize
		rect := image.Rect(x0, 0, x0+DefaultGlyphSize, DefaultGlyphSize)
		dst := image.NewNRGBA(image.Rect(0, 0, DefaultGlyphSize, DefaultGlyphSize))
		draw.Draw(dst, dst.Bounds(), src, rect.Min, draw.Src)
		result[name] = dst
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// MaxRects bin packer (standalone, no ebiten)
// ---------------------------------------------------------------------------

type packInput struct {
	name   string
	img    image.Image
	width  int
	height int
}

type packResult struct {
	name   string
	img    image.Image
	x, y   int
	width  int
	height int
}

type packFreeRect struct {
	x, y, w, h int
}

// packAtlas packs images into a single atlas using MaxRects BSSF.
// Returns the packed results and the atlas dimensions.
func packAtlas(images map[string]image.Image, padding int) ([]packResult, int, int, error) {
	if len(images) == 0 {
		return nil, 0, 0, nil
	}

	// Build sorted input list (largest area first for better packing).
	inputs := make([]packInput, 0, len(images))
	for name, img := range images {
		b := img.Bounds()
		inputs = append(inputs, packInput{
			name:   name,
			img:    img,
			width:  b.Dx(),
			height: b.Dy(),
		})
	}
	sort.Slice(inputs, func(i, j int) bool {
		ai := inputs[i].width * inputs[i].height
		aj := inputs[j].width * inputs[j].height
		if ai != aj {
			return ai > aj
		}
		return inputs[i].name < inputs[j].name
	})

	// Estimate atlas size — start with smallest power-of-two that fits.
	totalArea := 0
	maxW, maxH := 0, 0
	for _, in := range inputs {
		totalArea += (in.width + padding) * (in.height + padding)
		if in.width > maxW {
			maxW = in.width
		}
		if in.height > maxH {
			maxH = in.height
		}
	}

	atlasW := nextPow2(maxW + padding)
	atlasH := nextPow2(maxH + padding)
	// Grow until area is sufficient.
	for atlasW*atlasH < totalArea {
		if atlasW <= atlasH {
			atlasW *= 2
		} else {
			atlasH *= 2
		}
	}

	// Try packing, grow if it fails.
	for attempt := 0; attempt < 10; attempt++ {
		results, err := tryPack(inputs, atlasW, atlasH, padding)
		if err == nil {
			return results, atlasW, atlasH, nil
		}
		// Grow the smaller dimension.
		if atlasW <= atlasH {
			atlasW *= 2
		} else {
			atlasH *= 2
		}
	}

	return nil, 0, 0, fmt.Errorf("atlas pack: images do not fit in reasonable atlas size")
}

func tryPack(inputs []packInput, atlasW, atlasH, padding int) ([]packResult, error) {
	freeRects := []packFreeRect{{0, 0, atlasW, atlasH}}
	results := make([]packResult, 0, len(inputs))

	for _, in := range inputs {
		pw := in.width + padding
		ph := in.height + padding

		bestIdx := -1
		bestShort := int(^uint(0) >> 1)
		bestLong := bestShort
		bestX, bestY := 0, 0

		for i, fr := range freeRects {
			if pw <= fr.w && ph <= fr.h {
				shortSide := min(fr.w-pw, fr.h-ph)
				longSide := max(fr.w-pw, fr.h-ph)
				if shortSide < bestShort || (shortSide == bestShort && longSide < bestLong) {
					bestIdx = i
					bestShort = shortSide
					bestLong = longSide
					bestX = fr.x
					bestY = fr.y
				}
			}
		}
		_ = bestIdx

		if bestShort == int(^uint(0)>>1) {
			return nil, fmt.Errorf("sprite %q (%dx%d) does not fit", in.name, in.width, in.height)
		}

		results = append(results, packResult{
			name:   in.name,
			img:    in.img,
			x:      bestX,
			y:      bestY,
			width:  in.width,
			height: in.height,
		})

		placed := packFreeRect{bestX, bestY, pw, ph}
		freeRects = splitFree(freeRects, placed)
		freeRects = pruneFree(freeRects)
	}

	return results, nil
}

func splitFree(rects []packFreeRect, placed packFreeRect) []packFreeRect {
	var newRects []packFreeRect

	i := 0
	for i < len(rects) {
		fr := rects[i]
		if !rectsOverlap(fr, placed) {
			i++
			continue
		}

		// Remove overlapping rect.
		rects[i] = rects[len(rects)-1]
		rects = rects[:len(rects)-1]

		if placed.x > fr.x {
			newRects = append(newRects, packFreeRect{fr.x, fr.y, placed.x - fr.x, fr.h})
		}
		if placed.x+placed.w < fr.x+fr.w {
			newRects = append(newRects, packFreeRect{placed.x + placed.w, fr.y, (fr.x + fr.w) - (placed.x + placed.w), fr.h})
		}
		if placed.y > fr.y {
			newRects = append(newRects, packFreeRect{fr.x, fr.y, fr.w, placed.y - fr.y})
		}
		if placed.y+placed.h < fr.y+fr.h {
			newRects = append(newRects, packFreeRect{fr.x, placed.y + placed.h, fr.w, (fr.y + fr.h) - (placed.y + placed.h)})
		}
	}

	return append(rects, newRects...)
}

func pruneFree(rects []packFreeRect) []packFreeRect {
	n := len(rects)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			if rectContains(rects[j], rects[i]) {
				rects[i] = rects[n-1]
				n--
				i--
				break
			}
		}
	}
	return rects[:n]
}

func rectsOverlap(a, b packFreeRect) bool {
	return a.x < b.x+b.w && a.x+a.w > b.x && a.y < b.y+b.h && a.y+a.h > b.y
}

func rectContains(outer, inner packFreeRect) bool {
	return inner.x >= outer.x && inner.y >= outer.y &&
		inner.x+inner.w <= outer.x+outer.w &&
		inner.y+inner.h <= outer.y+outer.h
}

func nextPow2(v int) int {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	if v < 64 {
		return 64
	}
	return v
}

// ---------------------------------------------------------------------------
// Atlas composition & JSON generation
// ---------------------------------------------------------------------------

// ComposeAtlasImage draws all packed results onto a new NRGBA image.
func ComposeAtlasImage(results []packResult, w, h int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for _, r := range results {
		draw.Draw(dst, image.Rect(r.x, r.y, r.x+r.width, r.y+r.height),
			r.img, r.img.Bounds().Min, draw.Src)
	}
	return dst
}

// EncodeAtlasPNG encodes an atlas image as PNG bytes.
func EncodeAtlasPNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode atlas PNG: %w", err)
	}
	return buf.Bytes(), nil
}

// atlasJSONFrame matches the TexturePacker hash format used by willow.LoadAtlas.
type atlasJSONFrame struct {
	Frame            atlasJSONRect `json:"frame"`
	Rotated          bool          `json:"rotated"`
	Trimmed          bool          `json:"trimmed"`
	SpriteSourceSize atlasJSONRect `json:"spriteSourceSize"`
	SourceSize       atlasJSONSize `json:"sourceSize"`
}

type atlasJSONRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type atlasJSONSize struct {
	W int `json:"w"`
	H int `json:"h"`
}

type atlasJSONRoot struct {
	Frames map[string]atlasJSONFrame `json:"frames"`
	Meta   atlasJSONMeta             `json:"meta"`
}

type atlasJSONMeta struct {
	Image string        `json:"image"`
	Size  atlasJSONSize `json:"size"`
}

// GenerateAtlasJSON produces TexturePacker-compatible hash-format JSON
// from pack results. This is the format willow.LoadAtlas expects.
func GenerateAtlasJSON(results []packResult, w, h int) ([]byte, error) {
	frames := make(map[string]atlasJSONFrame, len(results))
	for _, r := range results {
		frames[r.name] = atlasJSONFrame{
			Frame:            atlasJSONRect{X: r.x, Y: r.y, W: r.width, H: r.height},
			Rotated:          false,
			Trimmed:          false,
			SpriteSourceSize: atlasJSONRect{X: 0, Y: 0, W: r.width, H: r.height},
			SourceSize:       atlasJSONSize{W: r.width, H: r.height},
		}
	}

	root := atlasJSONRoot{
		Frames: frames,
		Meta: atlasJSONMeta{
			Image: "atlas.png",
			Size:  atlasJSONSize{W: w, H: h},
		},
	}

	return json.Marshal(root)
}

// ---------------------------------------------------------------------------
// Theme atlas compilation
// ---------------------------------------------------------------------------

// ThemeAtlasInput holds the images to pack into a theme atlas.
type ThemeAtlasInput struct {
	// DefaultGlyphsPNG is the raw PNG bytes of the default icon spritesheet.
	// If nil, default icons are not included.
	DefaultGlyphsPNG []byte

	// NineGridImages maps nine-grid names to their source images.
	NineGridImages map[string]image.Image

	// SpriteImages maps sprite names to their source images.
	// Names matching default glyph names override the defaults.
	SpriteImages map[string]image.Image
}

// ThemeAtlasOutput holds the result of atlas compilation.
type ThemeAtlasOutput struct {
	AtlasJSON []byte
	AtlasPNG  []byte
}

// CompileThemeAtlas packs all theme images into a single atlas and returns
// the atlas JSON and PNG bytes. Returns nil output if there are no images.
func CompileThemeAtlas(input *ThemeAtlasInput) (*ThemeAtlasOutput, error) {
	images := make(map[string]image.Image)

	// Layer 1: default icons (base layer).
	if input.DefaultGlyphsPNG != nil {
		glyphs, err := SplitDefaultGlyphs(input.DefaultGlyphsPNG)
		if err != nil {
			return nil, err
		}
		for name, img := range glyphs {
			images[name] = img
		}
	}

	// Layer 2: nine-grid source images.
	for name, img := range input.NineGridImages {
		images[name] = img
	}

	// Layer 3: user sprites (can override default icons by name).
	for name, img := range input.SpriteImages {
		images[name] = img
	}

	if len(images) == 0 {
		return nil, nil
	}

	// Pack.
	results, w, h, err := packAtlas(images, 1)
	if err != nil {
		return nil, err
	}

	// Compose atlas image.
	atlasImg := ComposeAtlasImage(results, w, h)

	// Generate JSON.
	atlasJSON, err := GenerateAtlasJSON(results, w, h)
	if err != nil {
		return nil, err
	}

	// Encode PNG.
	atlasPNG, err := EncodeAtlasPNG(atlasImg)
	if err != nil {
		return nil, err
	}

	return &ThemeAtlasOutput{
		AtlasJSON: atlasJSON,
		AtlasPNG:  atlasPNG,
	}, nil
}
