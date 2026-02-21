package widget

import (
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// defaultCornerSegments is the number of line segments per corner arc.
// 8 segments per 90-degree arc yields smooth curves at UI scale.
const defaultCornerSegments = 8

// RoundedRectPoints is an exported wrapper for render.RoundedRectPoints. Used for testing.
func RoundedRectPoints(w, h, r float64, segments int) []sg.Vec2 {
	return render.RoundedRectPoints(w, h, r, segments)
}

// RoundedRectBorderMesh is an exported wrapper for render.RoundedRectBorderMesh. Used for testing.
func RoundedRectBorderMesh(w, h, r, bw float64, segments int) ([]engine.Vertex, []uint16) {
	return render.RoundedRectBorderMesh(w, h, r, bw, segments)
}

// LerpColor is an exported wrapper for render.LerpColor. Used for testing.
func LerpColor(a, b sg.Color, t float64) sg.Color {
	return render.LerpColor(a, b, t)
}

// RoundedRectGradientMesh is an exported wrapper for render.RoundedRectGradientMesh. Used for testing.
func RoundedRectGradientMesh(w, h, cornerRadius float64, segments int, g *GradientColors) ([]engine.Vertex, []uint16) {
	return render.RoundedRectGradientMesh(w, h, cornerRadius, segments, g)
}
