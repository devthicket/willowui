// genicons generates the default icon spritesheet for WillowUI.
//
// Usage:
//
//	go run ./scripts/genicons/
//
// Output: assets/icons/default-glyphs.png
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// Glyph definitions. Each glyph has a name, size, and draw function.
type glyph struct {
	name string
	size int
	draw func(img *image.NRGBA, ox, oy, size int)
}

var glyphs = []glyph{
	// Chevrons (directional navigation)
	{"chevron-right", 48, drawChevronRight},
	{"chevron-down", 48, drawChevronDown},
	{"chevron-left", 48, drawChevronLeft},
	{"chevron-up", 48, drawChevronUp},

	// Arrows (filled triangles for sort indicators, dropdowns)
	{"arrow-up", 48, drawArrowUp},
	{"arrow-down", 48, drawArrowDown},

	// Actions
	{"close-x", 48, drawCloseX},
	{"plus", 48, drawPlus},
	{"minus", 48, drawMinus},
	{"checkmark", 48, drawCheckmark},

	// Search
	{"search", 48, drawSearch},

	// Menu
	{"hamburger", 48, drawHamburger},

	// Filter
	{"filter", 48, drawFilter},

	// Radio dot
	{"radio-dot", 48, drawRadioDot},

	// Drag handle grips
	{"grip-dots-v", 48, drawGripDotsV},
	{"grip-dots-h", 48, drawGripDotsH},
	{"grip-dots-square", 48, drawGripDotsSquare},
	{"grip-lines-v", 48, drawGripLinesV},
	{"grip-lines-h", 48, drawGripLinesH},

	// Password masking (larger for quality at small display sizes)
	{"password-dot", 48, drawPasswordDot},
}

func main() {
	// Compute spritesheet dimensions. Lay out left-to-right.
	totalW := 0
	maxH := 0
	for _, g := range glyphs {
		totalW += g.size
		if g.size > maxH {
			maxH = g.size
		}
	}

	img := image.NewNRGBA(image.Rect(0, 0, totalW, maxH))
	// Starts fully transparent (zero value).

	ox := 0
	fmt.Println("Generating default-glyphs.png")
	fmt.Printf("  Spritesheet: %dx%d\n", totalW, maxH)
	for _, g := range glyphs {
		fmt.Printf("  %-20s %3dx%-3d  at (%d, 0)\n", g.name, g.size, g.size, ox)
		g.draw(img, ox, 0, g.size)
		ox += g.size
	}

	outPath := "assets/icons/default-glyphs.png"
	f, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating %s: %v\n", outPath, err)
		os.Exit(1)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding PNG: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Written to %s\n", outPath)
}

// --- Drawing helpers ---

// setAA sets a pixel with anti-aliased alpha based on distance to edge.
// dist < 0 means inside the shape, dist > 0 means outside.
func setAA(img *image.NRGBA, x, y int, dist float64) {
	if dist > 1.0 {
		return
	}
	a := 1.0 - clamp(dist, -1.0, 1.0)*0.5 - 0.5
	if a <= 0 {
		return
	}
	existing := img.NRGBAAt(x, y)
	newA := uint8(clamp(a*255, 0, 255))
	if newA > existing.A {
		img.SetNRGBA(x, y, color.NRGBA{255, 255, 255, newA})
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// distToSegment returns the distance from point (px,py) to line segment (ax,ay)-(bx,by).
func distToSegment(px, py, ax, ay, bx, by float64) float64 {
	dx, dy := bx-ax, by-ay
	if dx == 0 && dy == 0 {
		return math.Hypot(px-ax, py-ay)
	}
	t := ((px-ax)*dx + (py-ay)*dy) / (dx*dx + dy*dy)
	t = clamp(t, 0, 1)
	return math.Hypot(px-(ax+t*dx), py-(ay+t*dy))
}

// drawLine draws an anti-aliased line with given stroke half-width.
func drawLine(img *image.NRGBA, ox, oy int, ax, ay, bx, by, halfW float64, bounds int) {
	for y := 0; y < bounds; y++ {
		for x := 0; x < bounds; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			d := distToSegment(px, py, ax, ay, bx, by)
			setAA(img, ox+x, oy+y, d-halfW)
		}
	}
}

// drawCircle draws an anti-aliased filled circle.
func drawCircle(img *image.NRGBA, ox, oy int, cx, cy, radius float64, bounds int) {
	for y := 0; y < bounds; y++ {
		for x := 0; x < bounds; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			d := math.Hypot(px-cx, py-cy) - radius
			setAA(img, ox+x, oy+y, d)
		}
	}
}

// drawRing draws an anti-aliased ring (circle outline).
func drawRing(img *image.NRGBA, ox, oy int, cx, cy, radius, halfW float64, bounds int) {
	for y := 0; y < bounds; y++ {
		for x := 0; x < bounds; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			d := math.Abs(math.Hypot(px-cx, py-cy)-radius) - halfW
			setAA(img, ox+x, oy+y, d)
		}
	}
}

// drawFilledTriangle draws an anti-aliased filled triangle.
func drawFilledTriangle(img *image.NRGBA, ox, oy int, x1, y1, x2, y2, x3, y3 float64, bounds int) {
	for y := 0; y < bounds; y++ {
		for x := 0; x < bounds; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			d := distToTriangle(px, py, x1, y1, x2, y2, x3, y3)
			setAA(img, ox+x, oy+y, d)
		}
	}
}

// distToTriangle returns signed distance to a filled triangle (negative inside).
func distToTriangle(px, py, x1, y1, x2, y2, x3, y3 float64) float64 {
	// Edge distances (positive = outside that edge)
	d1 := edgeDist(px, py, x1, y1, x2, y2)
	d2 := edgeDist(px, py, x2, y2, x3, y3)
	d3 := edgeDist(px, py, x3, y3, x1, y1)

	// If all same sign, point is inside
	if d1 <= 0 && d2 <= 0 && d3 <= 0 {
		return math.Max(d1, math.Max(d2, d3)) // most negative = deepest inside
	}
	if d1 >= 0 && d2 >= 0 && d3 >= 0 {
		return -math.Min(d1, math.Min(d2, d3))
	}

	// Outside: distance to nearest edge segment
	s1 := distToSegment(px, py, x1, y1, x2, y2)
	s2 := distToSegment(px, py, x2, y2, x3, y3)
	s3 := distToSegment(px, py, x3, y3, x1, y1)
	return math.Min(s1, math.Min(s2, s3))
}

// edgeDist returns signed distance from point to edge (negative on left side).
func edgeDist(px, py, ax, ay, bx, by float64) float64 {
	return (bx-ax)*(py-ay) - (by-ay)*(px-ax)
}

// --- Chevrons (stroke-based, open shapes) ---

func drawChevronRight(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	hw := s * 0.10
	padY := s * 0.12
	leftX := s * 0.33
	tipX := s * 0.60

	drawLine(img, ox, oy, leftX, padY, tipX, s/2, hw, size)
	drawLine(img, ox, oy, tipX, s/2, leftX, s-padY, hw, size)
}

func drawChevronDown(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	hw := s * 0.10
	padX := s * 0.12
	topY := s * 0.33
	tipY := s * 0.60

	drawLine(img, ox, oy, padX, topY, s/2, tipY, hw, size)
	drawLine(img, ox, oy, s/2, tipY, s-padX, topY, hw, size)
}

func drawChevronLeft(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	hw := s * 0.10
	padY := s * 0.12
	rightX := s * 0.67
	tipX := s * 0.40

	drawLine(img, ox, oy, rightX, padY, tipX, s/2, hw, size)
	drawLine(img, ox, oy, tipX, s/2, rightX, s-padY, hw, size)
}

func drawChevronUp(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	hw := s * 0.10
	padX := s * 0.12
	botY := s * 0.67
	tipY := s * 0.40

	drawLine(img, ox, oy, padX, botY, s/2, tipY, hw, size)
	drawLine(img, ox, oy, s/2, tipY, s-padX, botY, hw, size)
}

// --- Arrows (filled triangles for sort indicators, dropdowns) ---

func drawArrowUp(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	pad := s * 0.25
	// Upward-pointing filled triangle
	drawFilledTriangle(img, ox, oy,
		s/2, pad, // top center
		pad, s-pad, // bottom left
		s-pad, s-pad, // bottom right
		size)
}

func drawArrowDown(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	pad := s * 0.25
	// Downward-pointing filled triangle
	drawFilledTriangle(img, ox, oy,
		pad, pad, // top left
		s-pad, pad, // top right
		s/2, s-pad, // bottom center
		size)
}

// --- Actions ---

func drawCloseX(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	pad := s * 0.25
	hw := s * 0.06

	drawLine(img, ox, oy, pad, pad, s-pad, s-pad, hw, size)
	drawLine(img, ox, oy, s-pad, pad, pad, s-pad, hw, size)
}

func drawPlus(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	pad := s * 0.22
	hw := s * 0.06

	// Horizontal bar
	drawLine(img, ox, oy, pad, s/2, s-pad, s/2, hw, size)
	// Vertical bar
	drawLine(img, ox, oy, s/2, pad, s/2, s-pad, hw, size)
}

func drawMinus(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	pad := s * 0.22
	hw := s * 0.06

	// Horizontal bar
	drawLine(img, ox, oy, pad, s/2, s-pad, s/2, hw, size)
}

func drawCheckmark(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	hw := s * 0.07

	// Short leg: top-left area down to bottom-center
	ax, ay := s*0.18, s*0.50
	bx, by := s*0.40, s*0.75
	// Long leg: bottom-center up to top-right
	cx, cy := s*0.82, s*0.22

	drawLine(img, ox, oy, ax, ay, bx, by, hw, size)
	drawLine(img, ox, oy, bx, by, cx, cy, hw, size)
}

// --- Search ---

func drawSearch(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	hw := s * 0.055

	// Magnifier circle: centered in upper-left area
	circCX := s * 0.40
	circCY := s * 0.40
	circR := s * 0.22

	drawRing(img, ox, oy, circCX, circCY, circR, hw, size)

	// Handle: from circle edge down-right
	handleStart := math.Sqrt2 * circR * 0.5 // offset along 45-degree diagonal
	hx1 := circCX + handleStart*1.05
	hy1 := circCY + handleStart*1.05
	hx2 := s * 0.78
	hy2 := s * 0.78

	drawLine(img, ox, oy, hx1, hy1, hx2, hy2, hw, size)
}

// --- Menu ---

func drawHamburger(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	padX := s * 0.20
	hw := s * 0.055

	// Three horizontal lines, evenly spaced
	y1 := s * 0.28
	y2 := s * 0.50
	y3 := s * 0.72

	drawLine(img, ox, oy, padX, y1, s-padX, y1, hw, size)
	drawLine(img, ox, oy, padX, y2, s-padX, y2, hw, size)
	drawLine(img, ox, oy, padX, y3, s-padX, y3, hw, size)
}

// --- Filter ---

func drawFilter(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	hw := s * 0.055

	// Funnel shape: wide top, narrow middle, stem to bottom
	// Top bar
	topL := s * 0.15
	topR := s * 0.85
	topY := s * 0.22

	// Funnel narrows to center
	midL := s * 0.38
	midR := s * 0.62
	midY := s * 0.55

	// Stem down
	stemX := s * 0.50
	stemY := s * 0.78

	// Left side of funnel
	drawLine(img, ox, oy, topL, topY, midL, midY, hw, size)
	// Right side of funnel
	drawLine(img, ox, oy, topR, topY, midR, midY, hw, size)
	// Top bar
	drawLine(img, ox, oy, topL, topY, topR, topY, hw, size)
	// Left to stem
	drawLine(img, ox, oy, midL, midY, stemX, stemY, hw, size)
	// Right to stem
	drawLine(img, ox, oy, midR, midY, stemX, stemY, hw, size)
}

// --- Radio dot ---

func drawRadioDot(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	radius := s * 0.28
	drawCircle(img, ox, oy, s/2, s/2, radius, size)
}

// --- Drag handle grips ---

func drawGripDotsV(img *image.NRGBA, ox, oy, size int) {
	drawDotGrid(img, ox, oy, size, 2, 3)
}

func drawGripDotsH(img *image.NRGBA, ox, oy, size int) {
	drawDotGrid(img, ox, oy, size, 3, 2)
}

func drawGripDotsSquare(img *image.NRGBA, ox, oy, size int) {
	drawDotGrid(img, ox, oy, size, 3, 3)
}

func drawDotGrid(img *image.NRGBA, ox, oy, size, cols, rows int) {
	s := float64(size)
	dotR := s * 0.07
	spacing := s * 0.20

	gridW := float64(cols-1) * spacing
	gridH := float64(rows-1) * spacing
	startX := (s - gridW) / 2
	startY := (s - gridH) / 2

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			cx := startX + float64(c)*spacing
			cy := startY + float64(r)*spacing
			drawCircle(img, ox, oy, cx, cy, dotR, size)
		}
	}
}

func drawGripLinesV(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	hw := s * 0.035
	lineLen := s * 0.55
	spacing := s * 0.20
	n := 3

	gridW := float64(n-1) * spacing
	startX := (s - gridW) / 2
	topY := (s - lineLen) / 2
	botY := topY + lineLen

	for i := 0; i < n; i++ {
		cx := startX + float64(i)*spacing
		drawLine(img, ox, oy, cx, topY, cx, botY, hw, size)
	}
}

func drawGripLinesH(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	hw := s * 0.035
	lineLen := s * 0.55
	spacing := s * 0.20
	n := 3

	gridH := float64(n-1) * spacing
	startY := (s - gridH) / 2
	leftX := (s - lineLen) / 2
	rightX := leftX + lineLen

	for i := 0; i < n; i++ {
		cy := startY + float64(i)*spacing
		drawLine(img, ox, oy, leftX, cy, rightX, cy, hw, size)
	}
}

// --- Password masking ---

func drawPasswordDot(img *image.NRGBA, ox, oy, size int) {
	s := float64(size)
	radius := s * 0.42
	drawCircle(img, ox, oy, s/2, s/2, radius, size)
}
