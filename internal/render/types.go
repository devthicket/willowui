package render

import "github.com/devthicket/willowui/internal/sg"

// Insets represents spacing on four sides (top, right, bottom, left).
type Insets struct {
	Top, Right, Bottom, Left float64
}

// Horizontal returns Left + Right.
func (i Insets) Horizontal() float64 {
	return i.Left + i.Right
}

// Vertical returns Top + Bottom.
func (i Insets) Vertical() float64 {
	return i.Top + i.Bottom
}

// IsAuto reports whether i is the AutoPadding sentinel (any field is -1).
func (i Insets) IsAuto() bool {
	return i.Top < 0 || i.Right < 0 || i.Bottom < 0 || i.Left < 0
}

// NineSlice describes a nine-slice image for use as a component background.
// Region is the base texture region; Insets define the non-stretched border
// widths (top, right, bottom, left) in pixels.
// InnerRegion defines the content area within the grid, in local coordinates
// relative to Region. Components use this for padding reference point.
// CenterFill, when non-nil, replaces the center cell's texture with a
// gradient fill.
type NineSlice struct {
	Region      sg.TextureRegion
	Insets      Insets
	InnerRegion Rect
	CenterFill  *GradientColors
}

// GradientColors defines per-corner colors for gradient backgrounds.
// Colors are interpolated bilinearly across the rectangle.
type GradientColors struct {
	TopLeft, TopRight, BottomRight, BottomLeft sg.Color
}

// Rect describes a rectangle with position and dimensions.
type Rect struct {
	X, Y, Width, Height float64
}

// GradientMode selects which corners are independently editable.
type GradientMode int

const (
	GradientModeH          GradientMode = iota // horizontal: TL=BL, TR=BR
	GradientModeV                              // vertical:   TL=TR, BL=BR
	GradientModeFourCorner                     // all 4 corners independent
)

// Gradient is the value type produced by GradientEditor.
// Colors always carries all four corners; H and V modes keep linked corners in sync.
type Gradient struct {
	Mode   GradientMode
	Colors GradientColors
}
