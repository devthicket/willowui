package widget

import "github.com/devthicket/willowui/internal/core"

// LayoutMode controls how a Component arranges its children.
type LayoutMode = core.LayoutMode

const (
	// LayoutNone uses manual positioning; children keep their own X/Y.
	LayoutNone = core.LayoutNone
	// LayoutVBox stacks children vertically with spacing between them.
	LayoutVBox = core.LayoutVBox
	// LayoutHBox stacks children horizontally with spacing between them.
	LayoutHBox = core.LayoutHBox
	// LayoutGrid arranges children in a fixed-column grid.
	LayoutGrid = core.LayoutGrid
	// LayoutFlow arranges children left-to-right, wrapping to new rows when
	// the available width is exceeded.
	LayoutFlow = core.LayoutFlow
	// LayoutAnchor pins each child to a corner, edge, or center of the parent
	// using per-child anchor metadata. Use Component.AddAnchoredChild to add
	// children with explicit anchor positions.
	LayoutAnchor = core.LayoutAnchor
)

// Alignment controls child positioning. Used for both cross-axis (Align)
// and main-axis (Justify) in VBox/HBox layouts.
type Alignment = core.Alignment

const (
	AlignStart        = core.AlignStart        // left for VBox, top for HBox (default)
	AlignCenter       = core.AlignCenter       // center on cross-axis
	AlignEnd          = core.AlignEnd          // right for VBox, bottom for HBox
	AlignSpaceBetween = core.AlignSpaceBetween // distribute children evenly across main axis
)

// FillMode controls how a component stretches to fill its parent's content area.
type FillMode int

const (
	FillNone   FillMode = 0
	FillWidth  FillMode = 1
	FillHeight FillMode = 2
	FillBoth   FillMode = FillWidth | FillHeight
)

// Orientation represents horizontal or vertical direction.
type Orientation = core.Orientation

const (
	Horizontal = core.Horizontal
	Vertical   = core.Vertical
)

// Anchor identifies a position within a parent container.
type Anchor = core.Anchor

const (
	AnchorTopLeft      = core.AnchorTopLeft
	AnchorTopCenter    = core.AnchorTopCenter
	AnchorTopRight     = core.AnchorTopRight
	AnchorMiddleLeft   = core.AnchorMiddleLeft
	AnchorCenter       = core.AnchorCenter
	AnchorMiddleRight  = core.AnchorMiddleRight
	AnchorBottomLeft   = core.AnchorBottomLeft
	AnchorBottomCenter = core.AnchorBottomCenter
	AnchorBottomRight  = core.AnchorBottomRight
)

// NewHBox creates a Component with LayoutHBox pre-configured.
func NewHBox(name string) *Component {
	c := NewComponent(name)
	c.Layout = LayoutHBox
	return c
}

// NewVBox creates a Component with LayoutVBox pre-configured.
func NewVBox(name string) *Component {
	c := NewComponent(name)
	c.Layout = LayoutVBox
	return c
}

// NewFlow creates a Component with LayoutFlow pre-configured.
func NewFlow(name string) *Component {
	c := NewComponent(name)
	c.Layout = LayoutFlow
	return c
}
