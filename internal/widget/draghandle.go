package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// DragAxis specifies which axis a DragHandle operates on.
type DragAxis int

const (
	DragAxisX        DragAxis = iota // horizontal drag — resizes width
	DragAxisY                        // vertical drag — resizes height
	DragAxisDiagonal                 // diagonal drag — resizes both
)

// DragGripStyle specifies the visual indicator rendered on a DragHandle.
type DragGripStyle int

const (
	DragGripDots  DragGripStyle = iota // dot grid
	DragGripLines                      // parallel short lines
	DragGripNone                       // invisible hit target, no visual
)

// sizable is implemented by widgets that have a SetSize method.
type sizable interface {
	SetSize(w, h float64)
}

// DragHandle is a visible grip primitive that emits drag delta events and can
// optionally resize a target component directly.
type DragHandle struct {
	Component

	axis      DragAxis
	gripStyle DragGripStyle
	target    *Component
	targetSz  sizable // cached typed SetSize, resolved at SetTarget time
	min, max  float64

	onDragStart func()
	onDrag      func(delta float64)
	onDragEnd   func(value float64)

	// Drag state.
	dragging      bool
	initialWidth  float64
	initialHeight float64

	// Grip visual nodes — small WhitePixel sprites arranged as dots or lines.
	gripNodes []*sg.Node
}

// NewDragHandle creates a DragHandle with default dot grip style.
func NewDragHandle(name string) *DragHandle {
	dh := &DragHandle{
		gripStyle: DragGripDots,
	}
	initComponent(&dh.Component, name)
	dh.initBackground(name)
	dh.initBorder(name)

	dh.SetSize(60, 8)

	// Drag handling.
	dh.node.OnDragStart(func(ctx sg.DragContext) {
		if !dh.enabled {
			return
		}
		dh.dragging = true
		if dh.target != nil {
			dh.initialWidth = dh.target.Width
			dh.initialHeight = dh.target.Height
		}
		if dh.onDragStart != nil {
			dh.onDragStart()
		}
	})

	dh.node.OnDrag(func(ctx sg.DragContext) {
		if !dh.enabled || !dh.dragging {
			return
		}
		dh.handleDrag(ctx)
	})

	dh.node.OnDragEnd(func(ctx sg.DragContext) {
		if !dh.dragging {
			return
		}
		dh.dragging = false
		dh.pressed = false
		if dh.onVisualStateChange != nil {
			dh.onVisualStateChange()
		}

		if dh.onDragEnd != nil {
			var value float64
			if dh.target != nil {
				switch dh.axis {
				case DragAxisX:
					value = dh.target.Width
				case DragAxisY:
					value = dh.target.Height
				default:
					value = dh.target.Width // report width for diagonal
				}
			} else {
				value = dh.axisDelta(ctx)
			}
			dh.onDragEnd(value)
		}
	})

	dh.onVisualStateChange = func() {
		if dh.dragging {
			dh.hovered = true
			dh.pressed = true
		}
		dh.UpdateVisuals()
	}
	dh.onThemeChange = func() { dh.UpdateVisuals() }
	dh.SetCursorShape(engine.CursorShapeNSResize)

	dh.UpdateVisuals()
	return dh
}

// SetAxis sets the drag axis.
func (dh *DragHandle) SetAxis(axis DragAxis) {
	dh.axis = axis
	switch axis {
	case DragAxisX:
		dh.SetCursorShape(engine.CursorShapeEWResize)
	case DragAxisY:
		dh.SetCursorShape(engine.CursorShapeNSResize)
	case DragAxisDiagonal:
		dh.SetCursorShape(engine.CursorShapeNWSEResize)
	}
	dh.rebuildGrip()
}

// Axis returns the current drag axis.
func (dh *DragHandle) Axis() DragAxis { return dh.axis }

// SetTarget sets the component to resize during drag.
func (dh *DragHandle) SetTarget(comp *Component) {
	dh.target = comp
	dh.targetSz = nil
	if comp != nil {
		if ud := comp.UserData(); ud != nil {
			if s, ok := ud.(sizable); ok {
				dh.targetSz = s
			}
		}
	}
}

// ClearTarget removes the resize target, switching to delegate mode.
func (dh *DragHandle) ClearTarget() {
	dh.target = nil
	dh.targetSz = nil
}

// SetMin sets the minimum constraint for resize mode.
func (dh *DragHandle) SetMin(v float64) { dh.min = v }

// SetMax sets the maximum constraint for resize mode.
func (dh *DragHandle) SetMax(v float64) { dh.max = v }

// Min returns the minimum constraint.
func (dh *DragHandle) Min() float64 { return dh.min }

// Max returns the maximum constraint.
func (dh *DragHandle) Max() float64 { return dh.max }

// SetOnDragStart sets the callback fired when a drag begins.
func (dh *DragHandle) SetOnDragStart(fn func()) { dh.onDragStart = fn }

// SetOnDrag sets the callback fired on each drag move with the total delta.
func (dh *DragHandle) SetOnDrag(fn func(delta float64)) { dh.onDrag = fn }

// SetOnDragEnd sets the callback fired on drag release with the final value.
func (dh *DragHandle) SetOnDragEnd(fn func(value float64)) { dh.onDragEnd = fn }

// SetGripStyle sets the visual grip indicator style.
func (dh *DragHandle) SetGripStyle(style DragGripStyle) {
	dh.gripStyle = style
	dh.rebuildGrip()
}

// GripStyle returns the current grip style.
func (dh *DragHandle) GripStyle() DragGripStyle { return dh.gripStyle }

// SetSize sets the handle dimensions (the entire area is the hit target).
func (dh *DragHandle) SetSize(w, h float64) {
	dh.Width = w
	dh.Height = h
	dh.resizeBackground(w, h)
	dh.resizeBorder(w, h)
	dh.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	dh.rebuildGrip()
	dh.MarkLayoutDirty()
}

// SetEnabled overrides Component.SetEnabled to also update visuals.
func (dh *DragHandle) SetEnabled(v bool) {
	dh.Component.SetEnabled(v)
	dh.UpdateVisuals()
}

// UpdateVisuals applies theme colors based on current state.
func (dh *DragHandle) UpdateVisuals() {
	dh.state = computeState(dh.enabled, dh.focused, dh.hovered, dh.pressed)
	group := dh.EffectiveTheme().DragHandle.Group(dh.Variant())

	bg := group.Background.Resolve(dh.state)
	dh.applyBackground(bg)

	// Resolve grip color based on state.
	var gripColor sg.Color
	if dh.pressed || dh.dragging {
		gripColor = group.GripActiveColor.Resolve(dh.state)
	} else if dh.hovered {
		gripColor = group.GripHoverColor.Resolve(dh.state)
	} else {
		gripColor = group.GripColor.Resolve(dh.state)
	}

	for _, n := range dh.gripNodes {
		n.SetColor(gripColor)
	}

	dh.MarkDrawDirty()
}

// handleDrag processes a drag move event.
func (dh *DragHandle) handleDrag(ctx sg.DragContext) {
	deltaX := ctx.GlobalX - ctx.StartX
	deltaY := ctx.GlobalY - ctx.StartY

	if dh.target != nil {
		switch dh.axis {
		case DragAxisX:
			newW := dh.clamp(dh.initialWidth + deltaX)
			dh.resizeTarget(newW, dh.target.Height)
			if dh.onDrag != nil {
				dh.onDrag(newW - dh.initialWidth)
			}
		case DragAxisY:
			newH := dh.clamp(dh.initialHeight + deltaY)
			dh.resizeTarget(dh.target.Width, newH)
			if dh.onDrag != nil {
				dh.onDrag(newH - dh.initialHeight)
			}
		case DragAxisDiagonal:
			newW := dh.clamp(dh.initialWidth + deltaX)
			newH := dh.clamp(dh.initialHeight + deltaY)
			dh.resizeTarget(newW, newH)
			if dh.onDrag != nil {
				dh.onDrag(deltaX) // report X delta for diagonal
			}
		}
	} else {
		if dh.onDrag != nil {
			dh.onDrag(dh.axisDelta(ctx))
		}
	}
}

// axisDelta returns the total displacement along the handle's active axis.
func (dh *DragHandle) axisDelta(ctx sg.DragContext) float64 {
	switch dh.axis {
	case DragAxisX:
		return ctx.GlobalX - ctx.StartX
	case DragAxisY:
		return ctx.GlobalY - ctx.StartY
	default:
		return ctx.GlobalX - ctx.StartX
	}
}

// resizeTarget sets the target's size using the cached sizable interface
// (resolved at SetTarget time), or falls back to direct field mutation.
func (dh *DragHandle) resizeTarget(w, h float64) {
	if dh.targetSz != nil {
		dh.targetSz.SetSize(w, h)
		return
	}
	// Fallback: set fields directly and invalidate layout.
	dh.target.Width = w
	dh.target.Height = h
	dh.target.resizeBackground(w, h)
	dh.target.resizeBorder(w, h)
	dh.target.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	dh.target.MarkLayoutDirty()
}

// clamp constrains a value to [min, max] if set.
func (dh *DragHandle) clamp(v float64) float64 {
	if dh.min != 0 && v < dh.min {
		v = dh.min
	}
	if dh.max != 0 && v > dh.max {
		v = dh.max
	}
	return v
}

// rebuildGrip removes existing grip nodes and creates new ones based on
// the current axis, grip style, and component dimensions.
func (dh *DragHandle) rebuildGrip() {
	// Remove old grip nodes.
	for _, n := range dh.gripNodes {
		dh.node.RemoveChild(n)
	}
	dh.gripNodes = nil

	if dh.gripStyle == DragGripNone {
		return
	}

	group := dh.EffectiveTheme().DragHandle.Group(dh.Variant())
	dotSize := group.GripDotSize
	if dotSize <= 0 {
		dotSize = 3
	}
	spacing := group.GripSpacing
	if spacing <= 0 {
		spacing = 4
	}
	count := group.GripCount
	if count <= 0 {
		count = 3
	}

	gripColor := group.GripColor.Resolve(dh.state)

	switch dh.gripStyle {
	case DragGripDots:
		dh.buildDotGrip(dotSize, spacing, count, gripColor)
	case DragGripLines:
		dh.buildLineGrip(dotSize, spacing, count, gripColor)
	}
}

// buildDotGrip creates a grid of small square sprites centered in the handle.
func (dh *DragHandle) buildDotGrip(dotSize, spacing float64, count int, color sg.Color) {
	var cols, rows int
	switch dh.axis {
	case DragAxisX:
		cols = 2
		rows = count
	case DragAxisY:
		cols = count
		rows = 2
	case DragAxisDiagonal:
		cols = count
		rows = count
	}

	totalW := float64(cols)*dotSize + float64(cols-1)*spacing
	totalH := float64(rows)*dotSize + float64(rows-1)*spacing
	startX := (dh.Width - totalW) / 2
	startY := (dh.Height - totalH) / 2

	dh.gripNodes = make([]*sg.Node, 0, rows*cols)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			dot := sg.NewSprite(dh.node.Name+"-grip-dot", sg.TextureRegion{})
			dot.SetScale(dotSize, dotSize)
			x := startX + float64(c)*(dotSize+spacing)
			y := startY + float64(r)*(dotSize+spacing)
			dot.SetPosition(x, y)
			dot.SetColor(color)
			dh.node.AddChild(dot)
			dh.gripNodes = append(dh.gripNodes, dot)
		}
	}
}

// buildLineGrip creates parallel line sprites centered in the handle.
func (dh *DragHandle) buildLineGrip(dotSize, spacing float64, count int, color sg.Color) {
	lineThickness := math.Max(dotSize/2, 1)
	dh.gripNodes = make([]*sg.Node, 0, count)

	switch dh.axis {
	case DragAxisX:
		// Vertical lines for horizontal drag.
		lineLen := dh.Height * 0.6
		totalW := float64(count)*lineThickness + float64(count-1)*spacing
		startX := (dh.Width - totalW) / 2
		startY := (dh.Height - lineLen) / 2
		for i := 0; i < count; i++ {
			line := sg.NewSprite(dh.node.Name+"-grip-line", sg.TextureRegion{})
			line.SetScale(lineThickness, lineLen)
			line.SetPosition(startX+float64(i)*(lineThickness+spacing), startY)
			line.SetColor(color)
			dh.node.AddChild(line)
			dh.gripNodes = append(dh.gripNodes, line)
		}
	case DragAxisY, DragAxisDiagonal:
		// Horizontal lines for vertical/diagonal drag.
		lineLen := dh.Width * 0.6
		totalH := float64(count)*lineThickness + float64(count-1)*spacing
		startX := (dh.Width - lineLen) / 2
		startY := (dh.Height - totalH) / 2
		for i := 0; i < count; i++ {
			line := sg.NewSprite(dh.node.Name+"-grip-line", sg.TextureRegion{})
			line.SetScale(lineLen, lineThickness)
			line.SetPosition(startX, startY+float64(i)*(lineThickness+spacing))
			line.SetColor(color)
			dh.node.AddChild(line)
			dh.gripNodes = append(dh.gripNodes, line)
		}
	}
}

// GripNodes returns the grip child nodes. Used for testing.
func (dh *DragHandle) GripNodes() []*sg.Node { return dh.gripNodes }

// Target returns the current resize target, or nil. Used for testing.
func (dh *DragHandle) Target() *Component { return dh.target }
