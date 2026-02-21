package widget

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"sync"

	"github.com/devthicket/willowui/internal/engine"
)

// ---------------------------------------------------------------------------
// Glyph spritesheet — injected by the root willowui package at init time
// ---------------------------------------------------------------------------

// GlyphSize is the pixel width and height of each glyph cell in the
// default spritesheet (default-glyphs.png).
const GlyphSize = 48

// Glyph index constants. Order must match scripts/genicons/main.go.
const (
	idxChevronRight   = 0
	idxChevronDown    = 1
	idxChevronLeft    = 2
	idxChevronUp      = 3
	idxArrowUp        = 4
	idxArrowDown      = 5
	idxCloseX         = 6
	idxPlus           = 7
	idxMinus          = 8
	idxCheckmark      = 9
	idxSearch         = 10
	idxHamburger      = 11
	idxFilter         = 12
	idxRadioDot       = 13
	idxGripDotsV      = 14
	idxGripDotsH      = 15
	idxGripDotsSquare = 16
	idxGripLinesV     = 17
	idxGripLinesH     = 18
	idxPasswordDot    = 19
	glyphCount        = 20
)

var (
	glyphSheetData []byte       // raw PNG bytes, set by SetGlyphSheet
	glyphSheet     engine.Image // decoded full spritesheet
	glyphImages    [glyphCount]engine.Image
	glyphOnce      [glyphCount]sync.Once
	sheetOnce      sync.Once
	fallbackOnce   sync.Once
	fallbackImg    engine.Image // 1×1 fuchsia pixel for error fallback
)

// SetGlyphSheet stores the raw PNG bytes of the default glyph spritesheet.
// Called by the root willowui package at init time.
func SetGlyphSheet(pngData []byte) {
	glyphSheetData = pngData
}

// decodeSheet lazily decodes the PNG spritesheet into an engine.Image.
func decodeSheet() engine.Image {
	sheetOnce.Do(func() {
		if len(glyphSheetData) == 0 {
			return
		}
		img, err := png.Decode(bytes.NewReader(glyphSheetData))
		if err != nil {
			return
		}
		glyphSheet = engine.NewImageFromImage(img)
	})
	return glyphSheet
}

// glyphFallback returns a 1×1 fuchsia image used when the spritesheet is
// missing or fails to decode.
func glyphFallback() engine.Image {
	fallbackOnce.Do(func() {
		img := engine.NewImage(1, 1)
		img.Fill(color.NRGBA{R: 255, G: 0, B: 255, A: 255})
		fallbackImg = img
	})
	return fallbackImg
}

// glyphAt extracts the sub-image at the given index from the spritesheet.
// The result is cached via sync.Once so each glyph is decoded at most once.
func glyphAt(index int) engine.Image {
	if index < 0 || index >= glyphCount {
		return glyphFallback()
	}
	glyphOnce[index].Do(func() {
		sheet := decodeSheet()
		if sheet == nil {
			glyphImages[index] = glyphFallback()
			return
		}
		x := index * GlyphSize
		glyphImages[index] = sheet.SubImage(image.Rect(x, 0, x+GlyphSize, GlyphSize)).(engine.Image)
	})
	return glyphImages[index]
}

// GlyphScale returns the uniform scale factor to display img at desiredPx
// width. For a 48×48 spritesheet glyph displayed at 9px: GlyphScale(img, 9) ≈ 0.1875.
func GlyphScale(img engine.Image, desiredPx float64) float64 {
	w := img.Bounds().Dx()
	if w <= 0 {
		return 1
	}
	return desiredPx / float64(w)
}

// ---------------------------------------------------------------------------
// Named icon accessors
// ---------------------------------------------------------------------------

func IconChevronRight() engine.Image   { return glyphAt(idxChevronRight) }
func IconChevronDown() engine.Image    { return glyphAt(idxChevronDown) }
func IconChevronLeft() engine.Image    { return glyphAt(idxChevronLeft) }
func IconChevronUp() engine.Image      { return glyphAt(idxChevronUp) }
func IconArrowUp() engine.Image        { return glyphAt(idxArrowUp) }
func IconArrowDown() engine.Image      { return glyphAt(idxArrowDown) }
func IconCloseX() engine.Image         { return glyphAt(idxCloseX) }
func IconPlus() engine.Image           { return glyphAt(idxPlus) }
func IconMinus() engine.Image          { return glyphAt(idxMinus) }
func IconCheckmark() engine.Image      { return glyphAt(idxCheckmark) }
func IconSearch() engine.Image         { return glyphAt(idxSearch) }
func IconHamburger() engine.Image      { return glyphAt(idxHamburger) }
func IconFilter() engine.Image         { return glyphAt(idxFilter) }
func IconRadioDot() engine.Image       { return glyphAt(idxRadioDot) }
func IconGripDotsV() engine.Image      { return glyphAt(idxGripDotsV) }
func IconGripDotsH() engine.Image      { return glyphAt(idxGripDotsH) }
func IconGripDotsSquare() engine.Image { return glyphAt(idxGripDotsSquare) }
func IconGripLinesV() engine.Image     { return glyphAt(idxGripLinesV) }
func IconGripLinesH() engine.Image     { return glyphAt(idxGripLinesH) }
func IconPasswordDot() engine.Image    { return glyphAt(idxPasswordDot) }
