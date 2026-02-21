package widget

import (
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// Panel is a static container with optional background color, border, and
// automatic child layout (VBox, HBox, Grid, or manual).
type Panel struct {
	Component
	bgOverride           bool // true when SetBackground was called manually
	borderOverride       bool // true when SetBorder was called manually
	cornerRadiusOverride bool // true when SetCornerRadii was called manually
	paddingOverride      bool // true when SetPadding was called manually
}

// NewPanel creates a Panel with no background and no border.
func NewPanel(name string) *Panel {
	p := &Panel{}
	initComponent(&p.Component, name)

	p.initBackground(name)
	p.initBorder(name)

	// Wire theme change handler.
	p.onThemeChange = func() { p.applyThemeColors() }
	p.applyThemeColors()

	return p
}

// applyThemeColors reads the PanelGroup from the effective theme and applies
// background and border unless manually overridden.
func (p *Panel) applyThemeColors() {
	group := p.EffectiveTheme().Panel.Group(p.Variant())
	if !p.cornerRadiusOverride {
		p.applyCornerRadius(group.CornerRadius)
	}
	if !p.bgOverride {
		p.applyBackground(group.Background.Resolve(p.state))
	}
	if !p.borderOverride {
		bg := group.Background.Resolve(p.state)
		p.applyBorder(group.Border.Resolve(p.state), group.BorderWidth, bg)
	}
	if !p.paddingOverride {
		p.Padding = group.Padding
	}
	p.MarkDrawDirty()
}

// SetBackground sets the panel's background color as a manual override.
// This prevents the theme from overwriting the background.
func (p *Panel) SetBackground(c sg.Color) {
	p.bgOverride = true
	p.applyBackground(Background{Type: BgSolid, Color: c})
	p.MarkDrawDirty()
}

// SetCornerRadii sets independent radii for each corner (TL, TR, BR, BL).
func (p *Panel) SetCornerRadii(tl, tr, br, bl float64) {
	p.cornerRadiusOverride = true
	p.applyCornerRadiiPerCorner(tl, tr, br, bl)
	p.applyThemeColors()
}

// SetBorder sets the border color and width as a manual override.
// This prevents the theme from overwriting the border.
func (p *Panel) SetBorder(c sg.Color, width float64) {
	p.borderOverride = true
	p.applyBorder(c, width, Background{Type: BgSolid})
	p.MarkDrawDirty()
}

// SetPadding sets the panel's content padding, overriding the theme default.
func (p *Panel) SetPadding(top, right, bottom, left float64) {
	p.paddingOverride = true
	p.Padding = render.Insets{Top: top, Right: right, Bottom: bottom, Left: left}
	p.MarkLayoutDirty()
}

// SetLayout sets the child layout mode (LayoutNone, LayoutVBox, LayoutHBox, LayoutGrid).
func (p *Panel) SetLayout(mode LayoutMode) {
	p.Layout = mode
	p.MarkLayoutDirty()
}

// SetSpacing sets the spacing between children for VBox/HBox/Grid layouts.
func (p *Panel) SetSpacing(s float64) {
	p.Spacing = s
	p.MarkLayoutDirty()
}

// SetAlignment sets the cross-axis alignment for VBox/HBox layouts.
func (p *Panel) SetAlignment(a Alignment) {
	p.Align = a
	p.MarkLayoutDirty()
}

// SetJustify sets the main-axis alignment for VBox/HBox layouts.
func (p *Panel) SetJustify(j Alignment) {
	p.Justify = j
	p.MarkLayoutDirty()
}

// SetSize sets the panel dimensions and updates background, border, and hit shape.
// If a layout mode is active, layout is applied immediately so callers see the
// final child positions without waiting for the next controller update cycle.
func (p *Panel) SetSize(w, h float64) {
	p.Width = w
	p.Height = h
	p.resizeBackground(w, h)
	p.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	p.resizeBorder(w, h)
	p.MarkLayoutDirty()
	if p.Layout != LayoutNone {
		p.UpdateLayout()
	}
}

// AddChild adds a child component. The child's willow node is added as a
// child of the panel's root node, and layout is scheduled.
func (p *Panel) AddChild(child UIElement) {
	p.Component.AddChild(child)
}
