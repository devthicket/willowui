//go:build ignore

// Prebuild generates jrpg.png containing three nine-grid frames side by side:
// a thick-border window frame, a thin-border button frame, and a thin-border
// accent frame.
// Run via: go run ./examples/_themes/prebuild.go
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	_, src, _, _ := runtime.Caller(0)
	dir := filepath.Dir(src)
	generateJRPGPNG(filepath.Join(dir, "jrpg.png"))
	log.Println("generated jrpg.png")
}

// generateJRPGPNG creates a PNG containing four 36x36 nine-grid frames
// side by side, plus three 11x11 glyphs:
//
//	[0]    thick-border window frame  (3px border, deep navy fill)
//	[1]    thin-border button frame   (1px border, mid blue fill)
//	[2]    thin-border accent frame   (1px border, dark indigo fill)
//	[3]    title frame                (outer border only, transparent fill)
//	[144]  11x11 close "X" glyph      (2px thick white pixel art)
//	[155]  11x11 expand "▸" glyph     (right-pointing filled triangle)
//	[166]  11x11 collapse "▾" glyph   (down-pointing filled triangle)
//	[177]  20x20 checkmark "✓" glyph  (3px thick white pixel art)
func generateJRPGPNG(path string) {
	const size = 36
	const glyphSize = 11
	const checkSize = 20
	img := image.NewRGBA(image.Rect(0, 0, size*4+glyphSize*3+checkSize, size))

	outerBorder := color.RGBA{R: 200, G: 208, B: 224, A: 255} // silver-white
	innerBorder := color.RGBA{R: 90, G: 110, B: 180, A: 255}  // lighter blue
	fillNavy := color.RGBA{R: 12, G: 20, B: 69, A: 255}       // deep navy (#0C1445)
	fillMid := color.RGBA{R: 26, G: 45, B: 107, A: 255}       // mid blue (#1A2D6B)
	fillDark := color.RGBA{R: 16, G: 28, B: 82, A: 255}       // dark indigo (#101C52)

	// [0] Window frame: thick 3px border, 6px radius, deep navy fill.
	drawFrame(img, 0, 0, size, 6.0, 3, outerBorder, innerBorder, fillNavy)

	// [1] Button frame: thin 1px border, 4px radius, mid blue fill.
	drawFrame(img, size, 0, size, 4.0, 1, outerBorder, color.RGBA{}, fillMid)

	// [2] Accent frame: thin 1px border, 4px radius, dark indigo fill.
	drawFrame(img, size*2, 0, size, 4.0, 1, outerBorder, color.RGBA{}, fillDark)

	// [3] Title frame: outer border only, transparent fill. The gradient
	//     centerFill replaces the center cell at runtime.
	drawFrame(img, size*3, 0, size, 6.0, 3, outerBorder, color.RGBA{}, color.RGBA{})

	// [4] Close button X glyph: 11x11 thick pixel art X.
	drawCloseGlyph(img, size*4, 0, glyphSize, outerBorder)

	// [5] Expand glyph: 11x11 right-pointing filled triangle.
	drawExpandGlyph(img, size*4+glyphSize, 0, glyphSize, outerBorder)

	// [6] Collapse glyph: 11x11 down-pointing filled triangle.
	drawCollapseGlyph(img, size*4+glyphSize*2, 0, glyphSize, outerBorder)

	// [7] Checkmark glyph: 20x20 thick pixel art checkmark.
	drawCheckGlyph(img, size*4+glyphSize*3, 0, checkSize, outerBorder)

	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatalf("encode %s: %v", path, err)
	}
}

// drawFrame draws a rounded-rectangle frame at (ox, oy) with the given size,
// corner radius, and border thickness into img.
func drawFrame(img *image.RGBA, ox, oy, size int, radius float64, borderPx int, outer, inner color.RGBA, fill color.RGBA) {
	innerSize := size - 2*borderPx
	innerR := radius - float64(borderPx)
	if innerR < 0 {
		innerR = 0
	}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !insideRoundedRect(x, y, size, size, radius) {
				continue
			}
			if !insideRoundedRect(x-borderPx, y-borderPx, innerSize, innerSize, innerR) {
				img.Set(ox+x, oy+y, outer)
			} else if inner.A > 0 && !insideRoundedRect(x-borderPx-1, y-borderPx-1, innerSize-2, innerSize-2, max(innerR-1, 0)) {
				img.Set(ox+x, oy+y, inner)
			} else {
				img.Set(ox+x, oy+y, fill)
			}
		}
	}
}

// drawCloseGlyph draws a thick "X" glyph at (ox, oy).
// The X uses 3px-wide diagonal strokes for a chunky pixel-art look.
func drawCloseGlyph(img *image.RGBA, ox, oy, size int, c color.RGBA) {
	for i := 0; i < size; i++ {
		for d := -1; d <= 1; d++ {
			// Main diagonal (\)
			y := i + d
			if y >= 0 && y < size {
				img.Set(ox+i, oy+y, c)
			}
			// Anti-diagonal (/)
			y2 := i + d
			x2 := size - 1 - i
			if y2 >= 0 && y2 < size {
				img.Set(ox+x2, oy+y2, c)
			}
		}
	}
}

// drawExpandGlyph draws a right-pointing chevron ">" at (ox, oy).
// Uses 3px-wide diagonal strokes matching the close button X style.
func drawExpandGlyph(img *image.RGBA, ox, oy, size int, c color.RGBA) {
	half := size / 2
	for y := 0; y < size; y++ {
		dist := y
		if y > half {
			dist = size - 1 - y
		}
		cx := 2 + dist
		for d := -1; d <= 1; d++ {
			x := cx + d
			if x >= 0 && x < size {
				img.Set(ox+x, oy+y, c)
			}
		}
	}
}

// drawCollapseGlyph draws a down-pointing chevron "v" at (ox, oy).
// Uses 3px-wide diagonal strokes matching the close button X style.
func drawCollapseGlyph(img *image.RGBA, ox, oy, size int, c color.RGBA) {
	half := size / 2
	for x := 0; x < size; x++ {
		dist := x
		if x > half {
			dist = size - 1 - x
		}
		cy := 2 + dist
		for d := -1; d <= 1; d++ {
			y := cy + d
			if y >= 0 && y < size {
				img.Set(ox+x, oy+y, c)
			}
		}
	}
}

// drawCheckGlyph draws a checkmark "✓" at (ox, oy).
// Short stroke goes down-right from top-left, long stroke goes up-right to top-right.
// Uses 3px-wide strokes matching the other glyphs.
func drawCheckGlyph(img *image.RGBA, ox, oy, size int, c color.RGBA) {
	// Win95-style checkmark: short left leg, long right leg, bold strokes.
	// Vertex at bottom-center-left, right leg rises steeply.
	//   short leg: (3,10) -> (7,16)
	//   long leg:  (7,16) -> (17,3)
	drawThickLine(img, ox, oy, 3, 10, 7, 16, 2, c)
	drawThickLine(img, ox, oy, 7, 16, 17, 3, 2, c)
}

// drawThickLine draws a line from (x0,y0) to (x1,y1) using Bresenham with a
// filled circle of the given radius at each point for smooth, round strokes.
func drawThickLine(img *image.RGBA, ox, oy, x0, y0, x1, y1, r int, c color.RGBA) {
	dx := intAbs(x1 - x0)
	dy := intAbs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	rr := r * r
	for {
		// Stamp a filled circle at each point.
		for dy2 := -r; dy2 <= r; dy2++ {
			for dx2 := -r; dx2 <= r; dx2++ {
				if dx2*dx2+dy2*dy2 <= rr {
					img.Set(ox+x0+dx2, oy+y0+dy2, c)
				}
			}
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func insideRoundedRect(px, py, w, h int, r float64) bool {
	if px < 0 || px >= w || py < 0 || py >= h {
		return false
	}
	// Pixel center.
	fx := float64(px) + 0.5
	fy := float64(py) + 0.5
	fw := float64(w)
	fh := float64(h)
	// Check each corner arc.
	if fx < r && fy < r {
		dx, dy := r-fx, r-fy
		return dx*dx+dy*dy <= r*r
	}
	if fx > fw-r && fy < r {
		dx, dy := fx-(fw-r), r-fy
		return dx*dx+dy*dy <= r*r
	}
	if fx < r && fy > fh-r {
		dx, dy := r-fx, fy-(fh-r)
		return dx*dx+dy*dy <= r*r
	}
	if fx > fw-r && fy > fh-r {
		dx, dy := fx-(fw-r), fy-(fh-r)
		return dx*dx+dy*dy <= r*r
	}
	return true
}
