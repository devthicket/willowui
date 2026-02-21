package render

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// RoundedRectPoints generates the outline of a rounded rectangle as a closed
// polygon. Each corner is an arc of segments line segments. The radius r is
// clamped to min(w/2, h/2) to prevent overlap. Corners are generated
// clockwise: top-right, bottom-right, bottom-left, top-left.
func RoundedRectPoints(w, h, r float64, segments int) []sg.Vec2 {
	if segments < 1 {
		segments = 1
	}

	// Clamp radius.
	maxR := math.Min(w/2, h/2)
	if r > maxR {
		r = maxR
	}
	if r < 0 {
		r = 0
	}

	// Degenerate: no rounding → simple rectangle.
	if r == 0 {
		return []sg.Vec2{
			{X: 0, Y: 0},
			{X: w, Y: 0},
			{X: w, Y: h},
			{X: 0, Y: h},
		}
	}

	// 4 corners, each with (segments+1) points, but adjacent corners share
	// their endpoint/startpoint, so total = 4*(segments+1) - 4 = 4*segments.
	// Actually we emit segments+1 points per corner and they naturally
	// connect since the last point of one corner is the first of the next.
	// Total unique points: 4 * (segments + 1). The polygon is implicitly
	// closed by the renderer.
	points := make([]sg.Vec2, 0, 4*(segments+1))

	// Corner centers and start angles (clockwise):
	// Top-right:    center (w-r, r),   angles from -π/2 to 0
	// Bottom-right: center (w-r, h-r), angles from 0 to π/2
	// Bottom-left:  center (r, h-r),   angles from π/2 to π
	// Top-left:     center (r, r),     angles from π to 3π/2
	corners := [4]struct {
		cx, cy     float64
		startAngle float64
	}{
		{w - r, r, -math.Pi / 2},
		{w - r, h - r, 0},
		{r, h - r, math.Pi / 2},
		{r, r, math.Pi},
	}

	step := (math.Pi / 2) / float64(segments)
	for _, c := range corners {
		for i := 0; i <= segments; i++ {
			angle := c.startAngle + step*float64(i)
			px := c.cx + r*math.Cos(angle)
			py := c.cy + r*math.Sin(angle)
			points = append(points, sg.Vec2{X: px, Y: py})
		}
	}

	return points
}

// RoundedRectPointsPerCorner generates a rounded rectangle outline with
// independent radii for each corner: top-left, top-right, bottom-right,
// bottom-left. A radius of 0 produces a sharp corner.
func RoundedRectPointsPerCorner(w, h float64, rTL, rTR, rBR, rBL float64, segments int) []sg.Vec2 {
	if segments < 1 {
		segments = 1
	}

	maxR := math.Min(w/2, h/2)
	clamp := func(r float64) float64 {
		if r > maxR {
			return maxR
		}
		if r < 0 {
			return 0
		}
		return r
	}
	rTL = clamp(rTL)
	rTR = clamp(rTR)
	rBR = clamp(rBR)
	rBL = clamp(rBL)

	points := make([]sg.Vec2, 0, 4*(segments+1))
	step := (math.Pi / 2) / float64(segments)

	// Top-right corner.
	if rTR > 0 {
		cx, cy := w-rTR, rTR
		for i := 0; i <= segments; i++ {
			a := -math.Pi/2 + step*float64(i)
			points = append(points, sg.Vec2{X: cx + rTR*math.Cos(a), Y: cy + rTR*math.Sin(a)})
		}
	} else {
		points = append(points, sg.Vec2{X: w, Y: 0})
	}

	// Bottom-right corner.
	if rBR > 0 {
		cx, cy := w-rBR, h-rBR
		for i := 0; i <= segments; i++ {
			a := step * float64(i)
			points = append(points, sg.Vec2{X: cx + rBR*math.Cos(a), Y: cy + rBR*math.Sin(a)})
		}
	} else {
		points = append(points, sg.Vec2{X: w, Y: h})
	}

	// Bottom-left corner.
	if rBL > 0 {
		cx, cy := rBL, h-rBL
		for i := 0; i <= segments; i++ {
			a := math.Pi/2 + step*float64(i)
			points = append(points, sg.Vec2{X: cx + rBL*math.Cos(a), Y: cy + rBL*math.Sin(a)})
		}
	} else {
		points = append(points, sg.Vec2{X: 0, Y: h})
	}

	// Top-left corner.
	if rTL > 0 {
		cx, cy := rTL, rTL
		for i := 0; i <= segments; i++ {
			a := math.Pi + step*float64(i)
			points = append(points, sg.Vec2{X: cx + rTL*math.Cos(a), Y: cy + rTL*math.Sin(a)})
		}
	} else {
		points = append(points, sg.Vec2{X: 0, Y: 0})
	}

	return points
}

// RoundedRectBorderMesh generates a triangle-strip ring between an outer
// rounded rect and an inner rounded rect (inset by bw on all sides).
// Returns raw vertices and indices for use with willow.NewMesh.
// All vertices use white pixel UV (SrcX=0.5, SrcY=0.5) and full white
// color (tinted by node.Color at render time).
func RoundedRectBorderMesh(w, h, r, bw float64, segments int) ([]engine.Vertex, []uint16) {
	outer := RoundedRectPoints(w, h, r, segments)

	innerR := r - bw
	if innerR < 0 {
		innerR = 0
	}
	innerW := w - 2*bw
	innerH := h - 2*bw
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}

	inner := RoundedRectPoints(innerW, innerH, innerR, segments)
	// Offset inner points by (bw, bw).
	for i := range inner {
		inner[i].X += bw
		inner[i].Y += bw
	}

	n := len(outer)
	if len(inner) != n {
		// Mismatch shouldn't happen with same segments, but guard.
		if len(inner) < n {
			n = len(inner)
		}
	}

	// Build triangle strip ring: 2*n vertices, 2*n triangles (quads).
	verts := make([]engine.Vertex, 2*n)
	for i := 0; i < n; i++ {
		verts[2*i] = engine.Vertex{
			DstX:   float32(outer[i].X),
			DstY:   float32(outer[i].Y),
			SrcX:   0.5,
			SrcY:   0.5,
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
		}
		verts[2*i+1] = engine.Vertex{
			DstX:   float32(inner[i].X),
			DstY:   float32(inner[i].Y),
			SrcX:   0.5,
			SrcY:   0.5,
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
		}
	}

	// Build indices: for each pair of adjacent edge segments, form 2 triangles.
	inds := make([]uint16, 0, 6*n)
	for i := 0; i < n; i++ {
		next := (i + 1) % n
		o0 := uint16(2 * i)
		i0 := uint16(2*i + 1)
		o1 := uint16(2 * next)
		i1 := uint16(2*next + 1)

		// Triangle 1: outer[i], inner[i], outer[next]
		inds = append(inds, o0, i0, o1)
		// Triangle 2: inner[i], inner[next], outer[next]
		inds = append(inds, i0, i1, o1)
	}

	return verts, inds
}

// LerpColor linearly interpolates between two colors by t in [0,1].
func LerpColor(a, b sg.Color, t float64) sg.Color {
	return sg.RGBA(
		a.R()+(b.R()-a.R())*t,
		a.G()+(b.G()-a.G())*t,
		a.B()+(b.B()-a.B())*t,
		a.A()+(b.A()-a.A())*t,
	)
}

// BilinearColor computes the bilinearly interpolated color at normalized
// coordinates (u, v) given the four corner colors of a GradientColors.
func BilinearColor(g *GradientColors, u, v float64) sg.Color {
	top := LerpColor(g.TopLeft, g.TopRight, u)
	bot := LerpColor(g.BottomLeft, g.BottomRight, u)
	return LerpColor(top, bot, v)
}

// RoundedRectGradientMesh creates a mesh of vertices and indices for a
// gradient-filled rounded rectangle. cornerRadius can be 0 for sharp corners.
// The gradient colors are bilinearly interpolated across the rectangle using
// per-vertex coloring. The mesh uses fan triangulation from the centroid.
func RoundedRectGradientMesh(w, h, cornerRadius float64, segments int, g *GradientColors) ([]engine.Vertex, []uint16) {
	outline := RoundedRectPoints(w, h, cornerRadius, segments)
	n := len(outline)
	if n < 3 {
		return nil, nil
	}

	// Compute centroid of the polygon.
	var cx, cy float64
	for _, p := range outline {
		cx += p.X
		cy += p.Y
	}
	cx /= float64(n)
	cy /= float64(n)

	// Guard against zero-size rectangles.
	if w <= 0 || h <= 0 {
		return nil, nil
	}

	// Build vertices: index 0 = centroid, indices 1..n = outline points.
	verts := make([]engine.Vertex, n+1)

	// Centroid vertex with bilinearly interpolated color.
	cu, cv := cx/w, cy/h
	cc := BilinearColor(g, cu, cv)
	verts[0] = engine.Vertex{
		DstX:   float32(cx),
		DstY:   float32(cy),
		SrcX:   0.5,
		SrcY:   0.5,
		ColorR: float32(cc.R()),
		ColorG: float32(cc.G()),
		ColorB: float32(cc.B()),
		ColorA: float32(cc.A()),
	}

	// Outline vertices with per-vertex gradient colors.
	for i, p := range outline {
		u := p.X / w
		v := p.Y / h
		c := BilinearColor(g, u, v)
		verts[i+1] = engine.Vertex{
			DstX:   float32(p.X),
			DstY:   float32(p.Y),
			SrcX:   0.5,
			SrcY:   0.5,
			ColorR: float32(c.R()),
			ColorG: float32(c.G()),
			ColorB: float32(c.B()),
			ColorA: float32(c.A()),
		}
	}

	// Build fan indices: for each pair of adjacent outline points,
	// create a triangle: centroid(0), outline[i+1], outline[i+2].
	inds := make([]uint16, 0, 3*n)
	for i := 0; i < n; i++ {
		next := (i + 1) % n
		inds = append(inds, 0, uint16(i+1), uint16(next+1))
	}

	return verts, inds
}

// RoundedRectGradientMeshPerCorner creates a gradient-filled mesh with
// independent corner radii (top-left, top-right, bottom-right, bottom-left).
func RoundedRectGradientMeshPerCorner(w, h float64, rTL, rTR, rBR, rBL float64, segments int, g *GradientColors) ([]engine.Vertex, []uint16) {
	outline := RoundedRectPointsPerCorner(w, h, rTL, rTR, rBR, rBL, segments)
	n := len(outline)
	if n < 3 || w <= 0 || h <= 0 {
		return nil, nil
	}

	var cx, cy float64
	for _, p := range outline {
		cx += p.X
		cy += p.Y
	}
	cx /= float64(n)
	cy /= float64(n)

	verts := make([]engine.Vertex, n+1)
	cu, cv := cx/w, cy/h
	cc := BilinearColor(g, cu, cv)
	verts[0] = engine.Vertex{
		DstX: float32(cx), DstY: float32(cy),
		SrcX: 0.5, SrcY: 0.5,
		ColorR: float32(cc.R()), ColorG: float32(cc.G()), ColorB: float32(cc.B()), ColorA: float32(cc.A()),
	}
	for i, p := range outline {
		u := p.X / w
		v := p.Y / h
		c := BilinearColor(g, u, v)
		verts[i+1] = engine.Vertex{
			DstX: float32(p.X), DstY: float32(p.Y),
			SrcX: 0.5, SrcY: 0.5,
			ColorR: float32(c.R()), ColorG: float32(c.G()), ColorB: float32(c.B()), ColorA: float32(c.A()),
		}
	}

	inds := make([]uint16, 0, 3*n)
	for i := 0; i < n; i++ {
		next := (i + 1) % n
		inds = append(inds, 0, uint16(i+1), uint16(next+1))
	}
	return verts, inds
}
