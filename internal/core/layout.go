package core

// LayoutMode controls how a Component arranges its children.
type LayoutMode int

const (
	// LayoutNone uses manual positioning; children keep their own X/Y.
	LayoutNone LayoutMode = iota
	// LayoutVBox stacks children vertically with spacing between them.
	LayoutVBox
	// LayoutHBox stacks children horizontally with spacing between them.
	LayoutHBox
	// LayoutGrid arranges children in a fixed-column grid.
	LayoutGrid
	// LayoutFlow arranges children left-to-right, wrapping to new rows when
	// the available width is exceeded.
	LayoutFlow
	// LayoutAnchor pins each child to a corner, edge, or center of the parent
	// using per-child anchor metadata. Use Component.AddAnchoredChild to add
	// children with explicit anchor positions.
	LayoutAnchor
)

// Alignment controls child positioning. Used for both cross-axis (Align)
// and main-axis (Justify) in VBox/HBox layouts.
type Alignment int

const (
	AlignStart        Alignment = iota // left for VBox, top for HBox (default)
	AlignCenter                        // center on cross-axis
	AlignEnd                           // right for VBox, bottom for HBox
	AlignSpaceBetween                  // distribute children evenly across main axis
)

// Orientation represents horizontal or vertical direction.
type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

// Anchor identifies a position within a parent container. Used by AnchorLayout
// to pin children to edges or corners.
type Anchor int

const (
	AnchorTopLeft Anchor = iota
	AnchorTopCenter
	AnchorTopRight
	AnchorMiddleLeft
	AnchorCenter
	AnchorMiddleRight
	AnchorBottomLeft
	AnchorBottomCenter
	AnchorBottomRight
)
