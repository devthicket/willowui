package widget

import (
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// Radio manages a group of mutually exclusive radio buttons.
type Radio struct {
	Component
	buttons       []*RadioButton
	selected      *Ref[int]
	watch         WatchHandle
	onChange      func(int)
	columns       int
	verticalFirst bool
}

// NewRadio creates a new empty Radio widget.
func NewRadio(name string) *Radio {
	rg := &Radio{
		selected: NewRef(-1),
		columns:  1,
	}
	initComponent(&rg.Component, name)
	rg.Layout = LayoutVBox
	rg.Spacing = DefaultRadioGap
	return rg
}

// SetColumns sets the number of columns (default 1 = single vertical stack).
// Call before or after adding options; triggers a layout update.
func (rg *Radio) SetColumns(n int) {
	if n < 1 {
		n = 1
	}
	rg.columns = n
	rg.applyLayout()
	rg.recomputeSize()
}

// SetVerticalFirst controls fill order when columns > 1.
// false (default) = fill left-to-right then wrap (horizontal-first).
// true = fill top-to-bottom per column (vertical-first).
func (rg *Radio) SetVerticalFirst(v bool) {
	rg.verticalFirst = v
	if rg.columns > 1 {
		rg.recomputeSize()
		rg.MarkLayoutDirty()
	}
}

// applyLayout switches the underlying layout mode based on column count.
func (rg *Radio) applyLayout() {
	if rg.columns <= 1 {
		rg.Layout = LayoutVBox
		rg.onLayout = nil
	} else {
		rg.Layout = LayoutNone
		rg.onLayout = func() { rg.layoutMultiColumn() }
	}
	rg.MarkLayoutDirty()
}

// AddOption adds a radio button with the given label to the group.
func (rg *Radio) AddOption(text string, source *sg.FontFamily, displaySize float64) *RadioButton {
	idx := len(rg.buttons)
	rb := newRadioButton(rg, idx, text, source, displaySize)
	rg.buttons = append(rg.buttons, rb)
	rg.AddChild(rb)
	rg.recomputeSize()
	return rb
}

// recomputeSize updates the group's Width/Height from its buttons.
func (rg *Radio) recomputeSize() {
	if len(rg.buttons) == 0 {
		rg.Width, rg.Height = 0, 0
		return
	}
	if rg.columns <= 1 {
		var maxW, totalH float64
		for i, btn := range rg.buttons {
			if btn.Width > maxW {
				maxW = btn.Width
			}
			if i > 0 {
				totalH += rg.Spacing
			}
			totalH += btn.Height
		}
		rg.Width = maxW
		rg.Height = totalH
		return
	}

	// Multi-column: find the widest and tallest cell.
	var cellW, cellH float64
	for _, btn := range rg.buttons {
		if btn.Width > cellW {
			cellW = btn.Width
		}
		if btn.Height > cellH {
			cellH = btn.Height
		}
	}
	cols := rg.columns
	rows := (len(rg.buttons) + cols - 1) / cols
	rg.Width = float64(cols)*cellW + float64(cols-1)*rg.Spacing
	rg.Height = float64(rows)*cellH + float64(rows-1)*rg.Spacing
}

// layoutMultiColumn positions buttons in a grid (called via onLayout hook).
func (rg *Radio) layoutMultiColumn() {
	if len(rg.buttons) == 0 {
		return
	}
	var cellW, cellH float64
	for _, btn := range rg.buttons {
		if btn.Width > cellW {
			cellW = btn.Width
		}
		if btn.Height > cellH {
			cellH = btn.Height
		}
	}
	cols := rg.columns
	rows := (len(rg.buttons) + cols - 1) / cols
	for i, btn := range rg.buttons {
		var col, row int
		if rg.verticalFirst {
			// Fill down each column before moving right.
			col = i / rows
			row = i % rows
		} else {
			// Fill across each row before moving down.
			col = i % cols
			row = i / cols
		}
		x := float64(col) * (cellW + rg.Spacing)
		y := float64(row) * (cellH + rg.Spacing)
		btn.X = x
		btn.Y = y
		btn.node.SetPosition(x, y)
	}
}

// Selected returns the index of the selected button, or -1 if none.
func (rg *Radio) Selected() int {
	return rg.selected.Peek()
}

// SetSelected programmatically selects a button by index.
func (rg *Radio) SetSelected(idx int) {
	old := rg.selected.Peek()
	if old == idx {
		return
	}
	rg.selected.Set(idx)
	DefaultScheduler.Flush()
	rg.updateButtons()
}

// SetOnChange sets the callback invoked when selection changes.
func (rg *Radio) SetOnChange(fn func(int)) {
	rg.onChange = fn
}

// BindSelected binds the selection to a reactive Ref[int].
func (rg *Radio) BindSelected(ref *Ref[int]) {
	rg.selected = ref
	bindRef(&rg.watch, ref, rg.updateButtonsForIndex)
}

// Dispose stops watches and disposes all buttons.
func (rg *Radio) Dispose() {
	rg.watch.Stop()
	rg.Component.Dispose()
}

func (rg *Radio) selectIndex(idx int) {
	old := rg.selected.Peek()
	if old == idx {
		return
	}
	rg.selected.Set(idx)
	DefaultScheduler.Flush()
	rg.updateButtons()
	if rg.onChange != nil {
		rg.onChange(idx)
	}
}

func (rg *Radio) updateButtons() {
	rg.updateButtonsForIndex(rg.selected.Peek())
}

func (rg *Radio) updateButtonsForIndex(idx int) {
	for i, btn := range rg.buttons {
		btn.setSelected(i == idx)
	}
}

// Buttons returns the slice of RadioButton widgets in this group.
// Used for testing radio group internals.
func (rg *Radio) Buttons() []*RadioButton { return rg.buttons }

// ---------------------------------------------------------------------------
// RadioButton
// ---------------------------------------------------------------------------

// RadioButton is a single option within a Radio widget.
type RadioButton struct {
	Component
	group          *Radio
	index          int
	selected       bool     // canonical selection state
	circle         *sg.Node // outer circle (WhitePixel flat)
	circlePoly     *sg.Node // outer circle (polygon, rounded)
	dot            *sg.Node // inner dot (WhitePixel flat)
	dotPoly        *sg.Node // inner dot (polygon, rounded)
	label          *Label
	appliedDotIcon engine.Image // tracks last applied theme dot icon
}

// DefaultRadioSize is the outer circle dimension.
const DefaultRadioSize = 20

// DefaultRadioDotSize is the inner dot dimension.
const DefaultRadioDotSize = 10

// DefaultRadioGap is the spacing between the circle and label.
const DefaultRadioGap = 8

func newRadioButton(group *Radio, index int, text string, source *sg.FontFamily, displaySize float64) *RadioButton {
	name := group.Name() + "-opt-" + text
	rb := &RadioButton{
		group: group,
		index: index,
	}
	initComponent(&rb.Component, name)

	// Outer circle: flat sprite.
	rb.circle = sg.NewSprite(name+"-circle", sg.TextureRegion{})

	rb.circle.SetScale(DefaultRadioSize, DefaultRadioSize)
	rb.node.AddChild(rb.circle)

	// Outer circle: polygon (for rounded/circular).
	circleR := float64(DefaultRadioSize) / 2
	cpts := render.RoundedRectPoints(DefaultRadioSize, DefaultRadioSize, circleR, defaultCornerSegments)
	rb.circlePoly = sg.NewPolygon(name+"-circle-poly", cpts)
	rb.circlePoly.SetVisible(false)
	rb.node.AddChild(rb.circlePoly)

	// Inner dot: flat sprite.
	rb.dot = sg.NewSprite(name+"-dot", sg.TextureRegion{})
	dotOffset := (DefaultRadioSize - DefaultRadioDotSize) / 2.0
	rb.dot.SetPosition(dotOffset, dotOffset)

	rb.dot.SetScale(DefaultRadioDotSize, DefaultRadioDotSize)
	rb.dot.SetVisible(false)
	rb.node.AddChild(rb.dot)

	// Inner dot: polygon (for rounded/circular).
	dotR := float64(DefaultRadioDotSize) / 2
	dpts := render.RoundedRectPoints(DefaultRadioDotSize, DefaultRadioDotSize, dotR, defaultCornerSegments)
	rb.dotPoly = sg.NewPolygon(name+"-dot-poly", dpts)
	rb.dotPoly.SetPosition(dotOffset, dotOffset)
	rb.dotPoly.SetVisible(false)
	rb.node.AddChild(rb.dotPoly)

	// Label.
	rb.label = NewLabel(name+"-label", text, source, displaySize)
	rb.label.SetPosition(DefaultRadioSize+DefaultRadioGap, (DefaultRadioSize-rb.label.Height)/2)
	rb.label.AddToNode(rb.node)

	rb.Width = DefaultRadioSize + DefaultRadioGap + rb.label.Width
	rb.Height = DefaultRadioSize
	if rb.label.Height > rb.Height {
		rb.Height = rb.label.Height
	}
	rb.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: rb.Width, Height: rb.Height}

	// Click selects this option.
	rb.node.OnClick(func(ctx sg.ClickContext) {
		if !rb.enabled {
			return
		}
		rb.group.selectIndex(rb.index)
	})

	rb.onThemeChange = func() { rb.applyThemeColors() }
	rb.SetCursorShape(engine.CursorShapePointer)

	// Focus: radio buttons participate in tab and spatial nav.
	rb.enableFocusNavigation()

	rb.onFocusChange = func(focused bool) { rb.applyThemeColors() }
	rb.node.OnUpdate = func(_ float64) {
		if !rb.focused || !rb.enabled {
			return
		}
		if DefaultInputManager.IsKeyJustAvailable(engine.KeySpace) {
			rb.group.selectIndex(rb.index)
			DefaultInputManager.Consume(engine.KeySpace)
		}
	}

	rb.applyThemeColors()
	return rb
}

func (rb *RadioButton) setSelected(sel bool) {
	rb.selected = sel
	rb.applyThemeColors()
}

func (rb *RadioButton) applyThemeColors() {
	rb.state = computeState(rb.enabled, rb.focused, rb.hovered, rb.selected)
	group := rb.EffectiveTheme().Radio.Group(rb.Variant())

	cr := resolveCornerRadius(group.CornerRadius, float64(DefaultRadioSize))

	circleColor := group.CircleColor.Resolve(rb.state)
	dotColor := group.DotColor.Resolve(rb.state)

	// When a theme dot icon is set, use the sprite node with the custom image
	// instead of the default flat/polygon rendering.
	if group.DotIcon.Set {
		rb.circlePoly.SetVisible(false)
		rb.dotPoly.SetVisible(false)
		rb.circle.SetColor(circleColor)
		rb.circle.SetVisible(true)
		if rb.appliedDotIcon != group.DotIcon.Image {
			rb.appliedDotIcon = group.DotIcon.Image
			rb.dot.SetCustomImage(group.DotIcon.Image)
			rb.dot.SetScale(1, 1)
			b := group.DotIcon.Image.Bounds()
			rb.dot.SetPosition((DefaultRadioSize-float64(b.Dx()))/2, (DefaultRadioSize-float64(b.Dy()))/2)
		}
		rb.dot.SetColor(dotColor)
		rb.dot.SetVisible(rb.selected)
	} else if cr > 0 {
		// Rounded: use polygon nodes.
		rb.circle.SetVisible(false)
		rb.dot.SetVisible(false)
		rb.circlePoly.SetColor(circleColor)
		rb.circlePoly.SetVisible(true)
		rb.dotPoly.SetColor(dotColor)
		rb.dotPoly.SetVisible(rb.selected)
	} else {
		// Sharp: use flat sprites.
		rb.circlePoly.SetVisible(false)
		rb.dotPoly.SetVisible(false)
		rb.circle.SetColor(circleColor)
		rb.circle.SetVisible(true)
		rb.dot.SetColor(dotColor)
		rb.dot.SetVisible(rb.selected)
	}
	rb.applyFocusRingEx(group.FocusColor.Resolve(rb.state), group.FocusRingWidth, DefaultRadioSize, DefaultRadioSize, cr)
	rb.MarkDrawDirty()
}

// Dispose disposes the radio button and its label.
func (rb *RadioButton) Dispose() {
	if rb.label != nil {
		rb.label.Dispose()
	}
	rb.Component.Dispose()
}
