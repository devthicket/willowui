package integration

import (
	"math"
	"testing"

	"github.com/devthicket/willow"
	"github.com/devthicket/willowui/internal/widget"
)

func TestRoundedRectPoints_ZeroRadius(t *testing.T) {
	pts := widget.RoundedRectPoints(100, 50, 0, 8)
	if len(pts) != 4 {
		t.Fatalf("expected 4 points for zero radius, got %d", len(pts))
	}
	// Should be a simple rectangle.
	expect := [][2]float64{{0, 0}, {100, 0}, {100, 50}, {0, 50}}
	for i, e := range expect {
		if math.Abs(pts[i].X-e[0]) > 1e-9 || math.Abs(pts[i].Y-e[1]) > 1e-9 {
			t.Errorf("point %d: got (%.2f, %.2f), want (%.2f, %.2f)", i, pts[i].X, pts[i].Y, e[0], e[1])
		}
	}
}

func TestRoundedRectPoints_Clamped(t *testing.T) {
	// Radius larger than half the smaller dimension should be clamped.
	pts := widget.RoundedRectPoints(40, 20, 50, 8)
	// Clamped radius = 10 (min(20, 10)).
	// All points must be within bounds.
	for i, p := range pts {
		if p.X < -1e-9 || p.X > 40+1e-9 || p.Y < -1e-9 || p.Y > 20+1e-9 {
			t.Errorf("point %d out of bounds: (%.4f, %.4f)", i, p.X, p.Y)
		}
	}
}

func TestRoundedRectPoints_Normal(t *testing.T) {
	pts := widget.RoundedRectPoints(100, 60, 10, 8)
	// Expected: 4 corners * (8+1) = 36 points.
	expected := 4 * (8 + 1)
	if len(pts) != expected {
		t.Fatalf("expected %d points, got %d", expected, len(pts))
	}
	// All points within bounds.
	for i, p := range pts {
		if p.X < -1e-9 || p.X > 100+1e-9 || p.Y < -1e-9 || p.Y > 60+1e-9 {
			t.Errorf("point %d out of bounds: (%.4f, %.4f)", i, p.X, p.Y)
		}
	}
}

func TestRoundedRectPoints_Symmetry(t *testing.T) {
	w, h := 80.0, 60.0
	pts := widget.RoundedRectPoints(w, h, 10, 8)

	// Verify horizontal symmetry: for each point (x, y), there should be a
	// corresponding point at (w-x, y).
	tol := 1e-9
	for _, p := range pts {
		mirrorX := w - p.X
		found := false
		for _, q := range pts {
			if math.Abs(q.X-mirrorX) < tol && math.Abs(q.Y-p.Y) < tol {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("no horizontal mirror for (%.4f, %.4f) → expected (%.4f, %.4f)", p.X, p.Y, mirrorX, p.Y)
		}
	}

	// Verify vertical symmetry.
	for _, p := range pts {
		mirrorY := h - p.Y
		found := false
		for _, q := range pts {
			if math.Abs(q.X-p.X) < tol && math.Abs(q.Y-mirrorY) < tol {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("no vertical mirror for (%.4f, %.4f) → expected (%.4f, %.4f)", p.X, p.Y, p.X, mirrorY)
		}
	}
}

func TestRoundedRectBorderMesh_VertexCount(t *testing.T) {
	verts, inds := widget.RoundedRectBorderMesh(100, 60, 10, 2, 8)
	outerPts := 4 * (8 + 1) // 36
	expectedVerts := 2 * outerPts
	expectedInds := 6 * outerPts
	if len(verts) != expectedVerts {
		t.Errorf("expected %d vertices, got %d", expectedVerts, len(verts))
	}
	if len(inds) != expectedInds {
		t.Errorf("expected %d indices, got %d", expectedInds, len(inds))
	}
}

func TestRoundedRectBorderMesh_ZeroBorderWidth(t *testing.T) {
	verts, inds := widget.RoundedRectBorderMesh(100, 60, 10, 0, 8)
	// With zero border width, outer and inner contours coincide →
	// mesh should still be valid (degenerate zero-area triangles).
	if len(verts) == 0 || len(inds) == 0 {
		t.Error("expected non-empty mesh even with zero border width")
	}

	// Verify outer and inner vertices overlap.
	for i := 0; i < len(verts); i += 2 {
		dx := verts[i].DstX - verts[i+1].DstX
		dy := verts[i].DstY - verts[i+1].DstY
		if math.Abs(float64(dx)) > 1e-4 || math.Abs(float64(dy)) > 1e-4 {
			t.Errorf("vertex pair %d: outer (%.2f, %.2f) != inner (%.2f, %.2f)",
				i/2, verts[i].DstX, verts[i].DstY, verts[i+1].DstX, verts[i+1].DstY)
		}
	}
}

func TestRoundedRectBorderMesh_ZeroRadius(t *testing.T) {
	verts, inds := widget.RoundedRectBorderMesh(100, 60, 0, 2, 8)
	// With zero radius, outer is 4 points, inner is 4 points.
	if len(verts) != 8 {
		t.Errorf("expected 8 vertices for zero radius, got %d", len(verts))
	}
	if len(inds) != 24 {
		t.Errorf("expected 24 indices for zero radius, got %d", len(inds))
	}
}

// ---------------------------------------------------------------------------
// Gradient tests
// ---------------------------------------------------------------------------

func TestLerpColor(t *testing.T) {
	a := willow.RGBA(0, 0, 0, 1)
	b := willow.RGBA(1, 1, 1, 1)

	// t=0 → a
	c0 := widget.LerpColor(a, b, 0)
	if math.Abs(c0.R()) > 1e-9 || math.Abs(c0.G()) > 1e-9 || math.Abs(c0.B()) > 1e-9 {
		t.Errorf("LerpColor(a, b, 0) = %v, want black", c0)
	}

	// t=1 → b
	c1 := widget.LerpColor(a, b, 1)
	if math.Abs(c1.R()-1) > 1e-9 || math.Abs(c1.G()-1) > 1e-9 || math.Abs(c1.B()-1) > 1e-9 {
		t.Errorf("LerpColor(a, b, 1) = %v, want white", c1)
	}

	// t=0.5 → midpoint
	c5 := widget.LerpColor(a, b, 0.5)
	if math.Abs(c5.R()-0.5) > 1e-9 || math.Abs(c5.G()-0.5) > 1e-9 || math.Abs(c5.B()-0.5) > 1e-9 {
		t.Errorf("LerpColor(a, b, 0.5) = %v, want {0.5, 0.5, 0.5, 1}", c5)
	}

	// Alpha interpolation.
	a2 := willow.RGBA(0, 0, 0, 0)
	b2 := willow.RGBA(0, 0, 0, 1)
	c25 := widget.LerpColor(a2, b2, 0.25)
	if math.Abs(c25.A()-0.25) > 1e-9 {
		t.Errorf("LerpColor alpha at 0.25 = %f, want 0.25", c25.A())
	}
}

func TestGradientMesh_SharpCorners(t *testing.T) {
	g := &widget.GradientColors{
		TopLeft:     willow.RGBA(1, 0, 0, 1),
		TopRight:    willow.RGBA(0, 1, 0, 1),
		BottomRight: willow.RGBA(0, 0, 1, 1),
		BottomLeft:  willow.RGBA(1, 1, 0, 1),
	}
	verts, inds := widget.RoundedRectGradientMesh(100, 60, 0, 8, g)

	// With radius=0, outline is 4 points. Total verts = 4 + 1 (centroid) = 5.
	if len(verts) != 5 {
		t.Errorf("expected 5 vertices, got %d", len(verts))
	}
	// Indices: 4 triangles * 3 = 12.
	if len(inds) != 12 {
		t.Errorf("expected 12 indices, got %d", len(inds))
	}

	// All vertices should be within bounds.
	for i, v := range verts {
		if v.DstX < -1e-4 || v.DstX > 100+1e-4 || v.DstY < -1e-4 || v.DstY > 60+1e-4 {
			t.Errorf("vertex %d out of bounds: (%.2f, %.2f)", i, v.DstX, v.DstY)
		}
	}
}

func TestGradientMesh_RoundedCorners(t *testing.T) {
	g := &widget.GradientColors{
		TopLeft:     willow.RGBA(1, 0, 0, 1),
		TopRight:    willow.RGBA(0, 1, 0, 1),
		BottomRight: willow.RGBA(0, 0, 1, 1),
		BottomLeft:  willow.RGBA(1, 1, 0, 1),
	}
	verts, inds := widget.RoundedRectGradientMesh(100, 60, 10, 8, g)

	// With radius=10, outline has 4*(8+1) = 36 points. Verts = 36 + 1 = 37.
	expectedVerts := 4*(8+1) + 1
	if len(verts) != expectedVerts {
		t.Errorf("expected %d vertices, got %d", expectedVerts, len(verts))
	}
	// Indices: 36 triangles * 3 = 108.
	expectedInds := 4 * (8 + 1) * 3
	if len(inds) != expectedInds {
		t.Errorf("expected %d indices, got %d", expectedInds, len(inds))
	}

	// All vertices within bounds.
	for i, v := range verts {
		if v.DstX < -1e-4 || v.DstX > 100+1e-4 || v.DstY < -1e-4 || v.DstY > 60+1e-4 {
			t.Errorf("vertex %d out of bounds: (%.2f, %.2f)", i, v.DstX, v.DstY)
		}
	}
}

func TestGradientMesh_VertexColors(t *testing.T) {
	red := willow.RGBA(1, 0, 0, 1)
	green := willow.RGBA(0, 1, 0, 1)
	blue := willow.RGBA(0, 0, 1, 1)
	yellow := willow.RGBA(1, 1, 0, 1)

	g := &widget.GradientColors{
		TopLeft:     red,
		TopRight:    green,
		BottomRight: blue,
		BottomLeft:  yellow,
	}
	verts, _ := widget.RoundedRectGradientMesh(100, 60, 0, 8, g)

	// Sharp corners: verts[0] = centroid, verts[1..4] = corners.
	// Corners: (0,0)=TL=red, (100,0)=TR=green, (100,60)=BR=blue, (0,60)=BL=yellow.
	const eps = 1e-4
	// Vert 1: top-left (0, 0) -> red
	if math.Abs(float64(verts[1].ColorR)-1) > eps || math.Abs(float64(verts[1].ColorG)) > eps {
		t.Errorf("top-left vertex color = (%.2f, %.2f, %.2f), want red",
			verts[1].ColorR, verts[1].ColorG, verts[1].ColorB)
	}
	// Vert 2: top-right (100, 0) -> green
	if math.Abs(float64(verts[2].ColorR)) > eps || math.Abs(float64(verts[2].ColorG)-1) > eps {
		t.Errorf("top-right vertex color = (%.2f, %.2f, %.2f), want green",
			verts[2].ColorR, verts[2].ColorG, verts[2].ColorB)
	}
	// Vert 3: bottom-right (100, 60) -> blue
	if math.Abs(float64(verts[3].ColorR)) > eps || math.Abs(float64(verts[3].ColorB)-1) > eps {
		t.Errorf("bottom-right vertex color = (%.2f, %.2f, %.2f), want blue",
			verts[3].ColorR, verts[3].ColorG, verts[3].ColorB)
	}
	// Vert 4: bottom-left (0, 60) -> yellow
	if math.Abs(float64(verts[4].ColorR)-1) > eps || math.Abs(float64(verts[4].ColorG)-1) > eps || math.Abs(float64(verts[4].ColorB)) > eps {
		t.Errorf("bottom-left vertex color = (%.2f, %.2f, %.2f), want yellow",
			verts[4].ColorR, verts[4].ColorG, verts[4].ColorB)
	}

	// Centroid (50, 30): bilinear interpolation at (0.5, 0.5).
	// Top row: lerp(red, green, 0.5) = (0.5, 0.5, 0, 1)
	// Bottom row: lerp(yellow, blue, 0.5) = (0.5, 0.5, 0.5, 1)
	// Final: lerp(top, bottom, 0.5) = (0.5, 0.5, 0.25, 1)
	if math.Abs(float64(verts[0].ColorR)-0.5) > eps ||
		math.Abs(float64(verts[0].ColorG)-0.5) > eps ||
		math.Abs(float64(verts[0].ColorB)-0.25) > eps {
		t.Errorf("centroid vertex color = (%.4f, %.4f, %.4f), want (0.5, 0.5, 0.25)",
			verts[0].ColorR, verts[0].ColorG, verts[0].ColorB)
	}
}
