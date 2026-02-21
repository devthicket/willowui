package widget

import (
	"fmt"

	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// ToolBarOverflowMode controls how items behave when they exceed the toolbar's bounds.
type ToolBarOverflowMode int

const (
	ToolBarClip   ToolBarOverflowMode = iota // items clip at edge (default)
	ToolBarScroll                            // reserved for future scroll arrows
	ToolBarWrap                              // items wrap to new row/column
)

// toolBarItemKind distinguishes item types within the toolbar.
type toolBarItemKind int

const (
	toolBarItemComponent toolBarItemKind = iota
	toolBarItemSeparator
	toolBarItemSpacer
)

// toolBarItem is an internal entry in the toolbar.
type toolBarItem struct {
	kind toolBarItemKind
	comp interface{ Node() *sg.Node } // nil for spacers
	sep  *sg.Node                     // separator sprite (for separator items)
}

// ToolBar is a horizontal or vertical command strip for housing actions,
// toggle groups, separators, and compact controls.
type ToolBar struct {
	Component
	items        []*toolBarItem
	orientation  Orientation
	overflowMode ToolBarOverflowMode
	separators   []*sg.Node // cached separator sprites for cleanup
}

// NewToolBar creates a new ToolBar with horizontal orientation and clip overflow.
func NewToolBar(name string) *ToolBar {
	tb := &ToolBar{
		orientation: Horizontal,
	}
	initComponent(&tb.Component, name)

	tb.initBackground(name)
	tb.initBorder(name)

	tb.onThemeChange = func() { tb.updateVisuals() }
	tb.updateVisuals()
	tb.SetSize(400, 40)

	return tb
}

// SetOrientation sets the toolbar orientation (Horizontal or Vertical).
func (tb *ToolBar) SetOrientation(o Orientation) {
	if tb.orientation == o {
		return
	}
	tb.orientation = o
	tb.updateLayout()
	tb.updateVisuals()
}

// Orientation returns the current toolbar orientation.
func (tb *ToolBar) Orientation() Orientation {
	return tb.orientation
}

// AddItem adds a component as a toolbar item.
func (tb *ToolBar) AddItem(comp interface{ Node() *sg.Node }) {
	item := &toolBarItem{kind: toolBarItemComponent, comp: comp}
	tb.items = append(tb.items, item)
	tb.node.AddChild(comp.Node())
	tb.updateLayout()
}

// AddSeparator adds a visual divider between items.
func (tb *ToolBar) AddSeparator() {
	name := fmt.Sprintf("%s-sep-%d", tb.node.Name, len(tb.items))
	sep := sg.NewSprite(name, sg.TextureRegion{})
	item := &toolBarItem{kind: toolBarItemSeparator, sep: sep}
	tb.items = append(tb.items, item)
	tb.separators = append(tb.separators, sep)
	tb.node.AddChild(sep)
	tb.updateLayout()
	tb.updateVisuals()
}

// AddSpacer adds a flexible space that pushes remaining items to the opposite end.
func (tb *ToolBar) AddSpacer() {
	item := &toolBarItem{kind: toolBarItemSpacer}
	tb.items = append(tb.items, item)
	tb.updateLayout()
}

// RemoveItem removes an item by name. Only component items are matched.
func (tb *ToolBar) RemoveItem(name string) {
	for i, item := range tb.items {
		if item.kind != toolBarItemComponent {
			continue
		}
		if item.comp.Node().Name != name {
			continue
		}
		tb.node.RemoveChild(item.comp.Node())
		copy(tb.items[i:], tb.items[i+1:])
		tb.items[len(tb.items)-1] = nil
		tb.items = tb.items[:len(tb.items)-1]
		tb.updateLayout()
		return
	}
}

// Clear removes all items from the toolbar.
func (tb *ToolBar) Clear() {
	for _, item := range tb.items {
		switch item.kind {
		case toolBarItemComponent:
			tb.node.RemoveChild(item.comp.Node())
		case toolBarItemSeparator:
			tb.node.RemoveChild(item.sep)
		}
	}
	tb.items = tb.items[:0]
	tb.separators = tb.separators[:0]
	tb.updateLayout()
}

// SetOverflowMode sets how overflow is handled.
func (tb *ToolBar) SetOverflowMode(mode ToolBarOverflowMode) {
	if tb.overflowMode == mode {
		return
	}
	tb.overflowMode = mode
	tb.updateLayout()
}

// SetWrap is a shorthand for setting ToolBarWrap overflow mode.
func (tb *ToolBar) SetWrap(v bool) {
	if v {
		tb.SetOverflowMode(ToolBarWrap)
	} else {
		tb.SetOverflowMode(ToolBarClip)
	}
}

// SetSize sets the toolbar dimensions and updates layout.
func (tb *ToolBar) SetSize(w, h float64) {
	tb.Width = w
	tb.Height = h
	tb.resizeBackground(w, h)
	tb.resizeBorder(w, h)
	tb.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	tb.updateLayout()
	tb.MarkLayoutDirty()
}

// Dispose cleans up the toolbar and all its items.
func (tb *ToolBar) Dispose() {
	tb.items = nil
	tb.separators = nil
	tb.Component.Dispose()
}

// defaultToolBarPadding is the default padding for toolbars.
var defaultToolBarPadding = render.Insets{Top: 4, Right: 8, Bottom: 4, Left: 8}

// toolBarItemSize returns the (width, height) of an item's node by reading
// its HitShape. All WillowUI widgets set HitShape in SetSize, so this is the
// most reliable way to get dimensions from an opaque interface{Node()}.
func toolBarItemSize(n *sg.Node) (float64, float64) {
	if hs, ok := n.HitShape.(sg.HitRect); ok {
		return hs.Width, hs.Height
	}
	// Fallback for sprites/text.
	return n.ScaleX(), n.ScaleY()
}

// updateLayout positions items within the toolbar.
func (tb *ToolBar) updateLayout() {
	if len(tb.items) == 0 {
		return
	}

	group := tb.EffectiveTheme().ToolBar.Group(tb.Variant())
	pad := resolveAutoInsets(group.Padding, defaultToolBarPadding)
	spacing := group.Spacing
	if spacing == 0 {
		spacing = 4
	}
	sepThickness := group.SeparatorThickness
	if sepThickness == 0 {
		sepThickness = 1
	}
	sepFraction := group.SeparatorHeight
	if sepFraction == 0 {
		sepFraction = 0.6
	}

	horiz := tb.orientation == Horizontal

	var availMain, availCross float64
	if horiz {
		availMain = tb.Width - pad.Left - pad.Right
		availCross = tb.Height - pad.Top - pad.Bottom
	} else {
		availMain = tb.Height - pad.Top - pad.Bottom
		availCross = tb.Width - pad.Left - pad.Right
	}

	// First pass: measure total fixed size and count spacers.
	var totalFixed float64
	var spacerCount int
	var gapCount int

	for i, item := range tb.items {
		switch item.kind {
		case toolBarItemComponent:
			w, h := toolBarItemSize(item.comp.Node())
			if horiz {
				totalFixed += w
			} else {
				totalFixed += h
			}
		case toolBarItemSeparator:
			totalFixed += sepThickness
		case toolBarItemSpacer:
			spacerCount++
		}
		if i > 0 {
			gapCount++
		}
	}

	totalFixed += float64(gapCount) * spacing

	// Calculate spacer size.
	var spacerSize float64
	if spacerCount > 0 {
		remaining := availMain - totalFixed
		if remaining > 0 {
			spacerSize = remaining / float64(spacerCount)
		}
	}

	// Second pass: position items.
	var mainPos float64
	if horiz {
		mainPos = pad.Left
	} else {
		mainPos = pad.Top
	}

	sepHeight := availCross * sepFraction

	for i, item := range tb.items {
		if i > 0 {
			mainPos += spacing
		}

		switch item.kind {
		case toolBarItemComponent:
			iw, ih := toolBarItemSize(item.comp.Node())

			if horiz {
				crossPos := pad.Top + (availCross-ih)/2
				item.comp.Node().SetPosition(mainPos, crossPos)
				mainPos += iw
			} else {
				crossPos := pad.Left + (availCross-iw)/2
				item.comp.Node().SetPosition(crossPos, mainPos)
				mainPos += ih
			}

		case toolBarItemSeparator:
			if horiz {
				sepY := pad.Top + (availCross-sepHeight)/2
				item.sep.SetScale(sepThickness, sepHeight)
				item.sep.SetPosition(mainPos, sepY)
				mainPos += sepThickness
			} else {
				sepX := pad.Left + (availCross-sepHeight)/2
				item.sep.SetScale(sepHeight, sepThickness)
				item.sep.SetPosition(sepX, mainPos)
				mainPos += sepThickness
			}

		case toolBarItemSpacer:
			mainPos += spacerSize
		}
	}

	// Handle wrap mode: if items overflow, wrap to new rows.
	if tb.overflowMode == ToolBarWrap && horiz {
		// Calculate row height from tallest item.
		var maxItemH float64
		for _, item := range tb.items {
			if item.kind == toolBarItemComponent {
				_, ih := toolBarItemSize(item.comp.Node())
				if ih > maxItemH {
					maxItemH = ih
				}
			}
		}
		if maxItemH == 0 {
			maxItemH = 32
		}
		tb.applyWrapLayout(pad, spacing, sepThickness, sepFraction, maxItemH)
	}
}

// applyWrapLayout re-positions items with wrapping when they exceed available width.
// It also resizes the toolbar background/border to fit all wrapped rows.
func (tb *ToolBar) applyWrapLayout(pad render.Insets, spacing, sepThickness, sepFraction, rowHeight float64) {
	availW := tb.Width - pad.Left - pad.Right

	x := pad.Left
	y := pad.Top
	sepH := rowHeight * sepFraction

	for i, item := range tb.items {
		if i > 0 {
			x += spacing
		}

		var itemW, itemH float64
		switch item.kind {
		case toolBarItemComponent:
			itemW, itemH = toolBarItemSize(item.comp.Node())
		case toolBarItemSeparator:
			itemW = sepThickness
		case toolBarItemSpacer:
			continue
		}

		// Wrap to next row if needed.
		if x+itemW > pad.Left+availW && i > 0 {
			x = pad.Left
			y += rowHeight + spacing
		}

		switch item.kind {
		case toolBarItemComponent:
			crossPos := y + (rowHeight-itemH)/2
			item.comp.Node().SetPosition(x, crossPos)
			x += itemW
		case toolBarItemSeparator:
			sepY := y + (rowHeight-sepH)/2
			item.sep.SetScale(sepThickness, sepH)
			item.sep.SetPosition(x, sepY)
			x += sepThickness
		}
	}

	// Resize background/border to fit all rows.
	totalH := y + rowHeight + pad.Bottom
	if totalH > tb.Height {
		tb.Height = totalH
		tb.resizeBackground(tb.Width, totalH)
		tb.resizeBorder(tb.Width, totalH)
		tb.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: tb.Width, Height: totalH}
	}
}

// updateVisuals applies theme colors to the toolbar and separators.
func (tb *ToolBar) updateVisuals() {
	tb.state = computeState(tb.enabled, tb.focused, tb.hovered, false)
	group := tb.EffectiveTheme().ToolBar.Group(tb.Variant())

	// Bar background and border.
	fallbackDim := tb.Height
	if tb.orientation != Horizontal {
		fallbackDim = tb.Width
	}
	cr := resolveCornerRadius(group.CornerRadius, fallbackDim)
	tb.applyCornerRadius(cr)
	bg := group.Background.Resolve(StateDefault)
	tb.applyBackground(bg)
	tb.applyBorder(group.BorderColor.Resolve(tb.state), group.BorderWidth, bg)

	// Update separator colors.
	sepColor := group.SeparatorColor.Resolve(StateDefault)
	for _, sep := range tb.separators {
		sep.SetColor(sepColor)
	}

	tb.MarkDrawDirty()
}
