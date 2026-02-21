package widget

import (
	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// menuPopupOverlayZIndex is below the tooltip overlay but above all windows.
const menuPopupOverlayZIndex = 500_000

// defaultMenuMaxHeight is used when MenuPopupGroup.MaxHeight is 0.
const defaultMenuMaxHeight = 280.0

// MenuItem is a single entry in a MenuPopup.
type MenuItem struct {
	Label     string
	OnSelect  func()
	Disabled  bool
	Separator bool   // if true renders as a thin divider; Label/OnSelect ignored
	Shortcut  string // display-only shortcut hint (e.g. "Ctrl+S"), right-aligned
}

// MenuPopup is a floating list of items displayed by MenuPopupManager.
// It is never added to the scene directly — Show/ShowAt go through the manager.
type MenuPopup struct {
	items       []MenuItem
	font        *sg.FontFamily
	displaySize float64
	highlighted int // index of hovered/keyboard-highlighted item (-1 = none)
	onDismiss   func()
	variant     Variant
	minWidth    float64 // minimum popup width (e.g. trigger button width)

	selectedIdx int // index of the currently-selected item (-1 = none); set before Show

	// internal render nodes, rebuilt on each show
	node      *sg.Node     // root container
	itemNodes []*sg.Node   // background sprite per item (nil for separators)
	itemY     []float64    // y position of each item in content space (for scroll nav)
	scroll    *ScrollPanel // non-nil when list height exceeds MaxHeight
	itemH     float64      // item height used in this build (for scroll nav)
}

// NewMenuPopup creates a MenuPopup that will display items using source at displaySize.
func NewMenuPopup(source *sg.FontFamily, displaySize float64) *MenuPopup {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	return &MenuPopup{
		font:        font,
		displaySize: displaySize,
		highlighted: -1,
		selectedIdx: -1,
	}
}

// SetItems sets the items to display.
func (p *MenuPopup) SetItems(items []MenuItem) {
	p.items = items
}

// SetOnDismiss sets a callback invoked when the popup closes without a selection.
func (p *MenuPopup) SetOnDismiss(fn func()) {
	p.onDismiss = fn
}

// SetVariant sets the theme variant used for styling.
func (p *MenuPopup) SetVariant(v Variant) {
	p.variant = v
}

// group returns the theme group for this popup's variant.
func (p *MenuPopup) menuGroup() *MenuPopupGroup {
	return getDefaultTheme().MenuPopup.Group(p.variant)
}

// build constructs the node tree for the popup using the current theme.
func (p *MenuPopup) build() {
	g := p.menuGroup()

	pad := g.Padding
	if pad.IsAuto() {
		pad = render.Insets{Top: 4, Right: 0, Bottom: 4, Left: 0}
	}

	itemH := g.ItemHeight
	if itemH <= 0 {
		itemH = 28
	}
	p.itemH = itemH
	itemPad := g.ItemPadding

	maxH := g.MaxHeight
	if maxH <= 0 {
		maxH = defaultMenuMaxHeight
	}

	// Measure max label width and shortcut width.
	maxLabelW := 0.0
	maxShortcutW := 0.0
	hasShortcut := false
	for _, item := range p.items {
		if item.Separator {
			continue
		}
		w, _ := measureDisplay(p.font, item.Label, p.displaySize)
		if w > maxLabelW {
			maxLabelW = w
		}
		if item.Shortcut != "" {
			hasShortcut = true
			sw, _ := measureDisplay(p.font, item.Shortcut, p.displaySize)
			if sw > maxShortcutW {
				maxShortcutW = sw
			}
		}
	}

	shortcutGap := 0.0
	if hasShortcut {
		shortcutGap = 24 // min gap between label and shortcut
	}

	popupW := maxLabelW + shortcutGap + maxShortcutW + itemPad.Left + itemPad.Right
	if popupW < 120 {
		popupW = 120
	}
	if p.minWidth > popupW {
		popupW = p.minWidth
	}

	// Total content height (items + padding).
	const sepH = 9.0
	totalH := pad.Top + pad.Bottom
	for _, item := range p.items {
		if item.Separator {
			totalH += sepH
		} else {
			totalH += itemH
		}
	}

	needsScroll := totalH > maxH
	visibleH := totalH
	if needsScroll {
		// Snap to a whole number of items to avoid half-row cutoff.
		sbW := float64(DefaultScrollBarWidth)
		innerW := popupW - sbW
		if innerW < 80 {
			popupW = 80 + sbW
		}
		// Cap to the nearest full itemH boundary below maxH.
		usable := maxH - pad.Top - pad.Bottom
		rows := float64(int(usable / itemH))
		if rows < 1 {
			rows = 1
		}
		visibleH = pad.Top + pad.Bottom + rows*itemH
	}

	bgColor := g.Background.Resolve(core.StateDefault).Color

	// Root container.
	root := sg.NewContainer("menu-popup")
	root.Interactable = true
	root.HitShape = sg.HitRect{X: 0, Y: 0, Width: popupW, Height: visibleH}

	// Background.
	bg := sg.NewSprite("menu-popup-bg", sg.TextureRegion{})
	bg.SetScale(popupW, visibleH)
	bg.SetColor(bgColor)
	root.AddChild(bg)

	// Border (4 edge sprites at z-index 1).
	bw := g.BorderWidth
	borderCol := g.Border.Resolve(core.StateDefault)
	if bw > 0 {
		top := sg.NewSprite("menu-popup-border-t", sg.TextureRegion{})
		top.SetScale(popupW, bw)
		top.SetColor(borderCol)
		top.SetZIndex(1)

		bot := sg.NewSprite("menu-popup-border-b", sg.TextureRegion{})
		bot.SetScale(popupW, bw)
		bot.SetPosition(0, visibleH-bw)
		bot.SetColor(borderCol)
		bot.SetZIndex(1)

		left := sg.NewSprite("menu-popup-border-l", sg.TextureRegion{})
		left.SetScale(bw, visibleH-bw*2)
		left.SetPosition(0, bw)
		left.SetColor(borderCol)
		left.SetZIndex(1)

		right := sg.NewSprite("menu-popup-border-r", sg.TextureRegion{})
		right.SetScale(bw, visibleH-bw*2)
		right.SetPosition(popupW-bw, bw)
		right.SetColor(borderCol)
		right.SetZIndex(1)

		for _, bd := range []*sg.Node{top, bot, left, right} {
			root.AddChild(bd)
		}
	}

	// Decide where items are parented: directly in root, or in a scroll panel.
	var itemsParent *sg.Node
	itemW := popupW // item width in the content area
	p.scroll = nil

	if needsScroll {
		sp := NewScrollPanel("menu-scroll")
		sp.SetTheme(getDefaultTheme())
		sp.SetBackground(bgColor)
		sp.SetBorder(bgColor, 0)
		sp.ShowHScroll(false)
		sp.ShowVScroll(true)
		sp.SetSize(popupW, visibleH)
		sp.SetContentSize(popupW-float64(DefaultScrollBarWidth), totalH)
		sp.node.SetZIndex(2)
		root.AddChild(sp.node)
		itemsParent = sp.ContentNode()
		itemW = popupW - float64(DefaultScrollBarWidth)
		p.scroll = sp
	} else {
		itemsParent = root
	}

	// Build items into itemsParent.
	p.itemNodes = p.itemNodes[:0]
	p.itemY = p.itemY[:0]
	defaultItemBg := g.ItemBackground.Resolve(core.StateDefault).Color
	y := pad.Top
	for idx, item := range p.items {
		p.itemY = append(p.itemY, y)
		if item.Separator {
			sep := sg.NewSprite("menu-sep", sg.TextureRegion{})
			sep.SetScale(itemW-itemPad.Left-itemPad.Right, 1)
			sep.SetPosition(itemPad.Left, y+4)
			sep.SetColor(g.SeparatorColor.Resolve(core.StateDefault))
			sep.SetZIndex(2)
			itemsParent.AddChild(sep)
			p.itemNodes = append(p.itemNodes, nil)
			y += sepH
			continue
		}

		iIdx := idx

		// Item background sprite (used to show highlight).
		itemBg := sg.NewSprite("menu-item-bg", sg.TextureRegion{})
		itemBg.SetScale(itemW, itemH)
		itemBg.SetPosition(0, y)
		itemBg.SetZIndex(2)
		if defaultItemBg.A() > 0 {
			itemBg.SetColor(defaultItemBg)
		} else {
			itemBg.SetVisible(false)
		}

		// Item label.
		textColor := g.TextColor.Resolve(core.StateDefault)
		if item.Disabled {
			textColor = g.DisabledColor.Resolve(core.StateDefault)
		}
		lbl := sg.NewText("menu-item-lbl", item.Label, p.font)
		lbl.TextBlock.FontSize = p.displaySize
		lbl.TextBlock.Color = textColor
		lbl.SetPosition(itemPad.Left, y+itemPad.Top)
		lbl.SetZIndex(3)

		// Shortcut hint (right-aligned).
		var shortcutNode *sg.Node
		if item.Shortcut != "" {
			shortcutNode = sg.NewText("menu-item-shortcut", item.Shortcut, p.font)
			shortcutNode.TextBlock.FontSize = p.displaySize
			scColor := g.DisabledColor.Resolve(core.StateDefault)
			shortcutNode.TextBlock.Color = scColor
			scW, _ := measureDisplay(p.font, item.Shortcut, p.displaySize)
			shortcutNode.SetPosition(itemW-itemPad.Right-scW, y+itemPad.Top)
			shortcutNode.SetZIndex(3)
		}

		// Hit container for pointer events.
		hitNode := sg.NewContainer("menu-item-hit")
		hitNode.Interactable = true
		hitNode.HitShape = sg.HitRect{X: 0, Y: y, Width: itemW, Height: itemH}
		hitNode.SetZIndex(4)

		if !item.Disabled {
			hitNode.OnPointerEnter(func(_ sg.PointerContext) {
				DefaultMenuPopupManager.setHighlight(iIdx)
			})
			hitNode.OnPointerLeave(func(_ sg.PointerContext) {
				DefaultMenuPopupManager.clearHighlight(iIdx)
			})
			hitNode.OnPointerDown(func(_ sg.PointerContext) {
				DefaultMenuPopupManager.selectItem(iIdx)
			})
		}

		itemsParent.AddChild(itemBg)
		itemsParent.AddChild(lbl)
		if shortcutNode != nil {
			itemsParent.AddChild(shortcutNode)
		}
		itemsParent.AddChild(hitNode)
		p.itemNodes = append(p.itemNodes, itemBg)
		y += itemH
	}

	p.node = root
	p.highlighted = -1

	// Pre-color the currently-selected item.
	if p.selectedIdx >= 0 && p.selectedIdx < len(p.itemNodes) {
		if n := p.itemNodes[p.selectedIdx]; n != nil {
			selBg := g.SelectedColor.Resolve(core.StateDefault)
			if selBg.Color.A() > 0 {
				n.SetColor(selBg.Color)
				n.SetVisible(true)
			}
		}
	}
}

// updateHighlight applies highlight/default color to the item background nodes.
func (p *MenuPopup) updateHighlight(prev, next int) {
	g := p.menuGroup()
	defaultBg := g.ItemBackground.Resolve(core.StateDefault).Color
	hoverBg := g.ItemBackground.Resolve(core.StateHover).Color
	selBg := g.SelectedColor.Resolve(core.StateDefault).Color

	if prev >= 0 && prev < len(p.itemNodes) && p.itemNodes[prev] != nil {
		n := p.itemNodes[prev]
		// Restore to selected color if this was the selected item, else default.
		restoreColor := defaultBg
		if prev == p.selectedIdx && selBg.A() > 0 {
			restoreColor = selBg
		}
		if restoreColor.A() > 0 {
			n.SetColor(restoreColor)
			n.SetVisible(true)
		} else {
			n.SetVisible(false)
		}
	}
	if next >= 0 && next < len(p.itemNodes) && p.itemNodes[next] != nil {
		n := p.itemNodes[next]
		if hoverBg.A() > 0 {
			n.SetColor(hoverBg)
			n.SetVisible(true)
		} else {
			n.SetVisible(false)
		}
	}
}

// scrollToHighlighted ensures the highlighted item is visible when navigating
// by keyboard in a scrollable popup.
func (p *MenuPopup) scrollToHighlighted() {
	if p.scroll == nil || p.highlighted < 0 || p.highlighted >= len(p.itemY) {
		return
	}
	iy := p.itemY[p.highlighted]
	cur := p.scroll.ScrollY()
	viewH := p.scroll.Height
	if iy < cur {
		p.scroll.SetScrollY(iy)
	} else if iy+p.itemH > cur+viewH {
		p.scroll.SetScrollY(iy + p.itemH - viewH)
	}
}

// ---------------------------------------------------------------------------
// MenuPopupManager
// ---------------------------------------------------------------------------

// MenuPopupManager manages the single active floating menu popup.
// Use DefaultMenuPopupManager; do not construct your own.
type MenuPopupManager struct {
	scene       *sg.Scene
	tickerNode  *sg.Node // drives keyboard navigation via OnUpdate
	overlayNode *sg.Node // popup content lives here
	dismissNode *sg.Node // full-screen transparent overlay catches click-outside

	active *MenuPopup
}

// DefaultMenuPopupManager is the singleton used by Select and ContextMenu.
var DefaultMenuPopupManager = &MenuPopupManager{}

// setScene is called from widget.SetScene.
func (m *MenuPopupManager) setScene(s *sg.Scene) {
	if m.active != nil {
		m.hideActive()
	}
	m.scene = s
	if s == nil || s.Root == nil {
		return
	}
	m.ensureNodes(s)
}

func (m *MenuPopupManager) ensureNodes(s *sg.Scene) {
	if s == nil || s.Root == nil {
		return
	}
	if m.tickerNode == nil {
		m.tickerNode = sg.NewContainer("menupopup-ticker")
		m.tickerNode.Interactable = false
		m.tickerNode.SetZIndex(menuPopupOverlayZIndex)
		// tick() is called from Screen.Update before FocusManager.Update so
		// popup key handling (arrow nav, Enter, Escape) takes priority over
		// spatial focus navigation. No OnUpdate hook needed here.
	}
	if m.overlayNode == nil {
		m.overlayNode = sg.NewContainer("menupopup-overlay")
		m.overlayNode.Interactable = true
		m.overlayNode.SetVisible(false)
		m.overlayNode.SetZIndex(menuPopupOverlayZIndex)
	}
	if m.dismissNode == nil {
		vw, vh := viewportSize()
		m.dismissNode = sg.NewContainer("menupopup-dismiss")
		m.dismissNode.Interactable = true
		m.dismissNode.HitShape = sg.HitRect{X: 0, Y: 0, Width: vw, Height: vh}
		m.dismissNode.SetZIndex(menuPopupOverlayZIndex - 1)
		m.dismissNode.SetVisible(false)
		m.dismissNode.OnPointerDown(func(_ sg.PointerContext) {
			DefaultMenuPopupManager.dismiss()
		})
	}
	if m.tickerNode.Parent != s.Root {
		if m.tickerNode.Parent != nil {
			m.tickerNode.Parent.RemoveChild(m.tickerNode)
		}
		s.Root.AddChild(m.tickerNode)
	}
	if m.overlayNode.Parent != s.Root {
		if m.overlayNode.Parent != nil {
			m.overlayNode.Parent.RemoveChild(m.overlayNode)
		}
		s.Root.AddChild(m.overlayNode)
	}
	if m.dismissNode.Parent != s.Root {
		if m.dismissNode.Parent != nil {
			m.dismissNode.Parent.RemoveChild(m.dismissNode)
		}
		s.Root.AddChild(m.dismissNode)
	}
}

// Show displays popup anchored below (or above) trigger component.
func (m *MenuPopupManager) Show(popup *MenuPopup, trigger *Component) {
	if sc := currentScene(); sc != nil {
		m.ensureNodes(sc)
	}
	m.hideActive()

	popup.build()
	if popup.node == nil {
		return
	}

	var x, y float64
	if trigger != nil {
		wx, wy := trigger.node.LocalToWorld(0, 0)
		x = wx
		y = wy + trigger.Height + 2
		vw, vh := viewportSize()
		hr := popup.node.HitShape.(sg.HitRect)
		if y+hr.Height > vh-8 {
			y = wy - hr.Height - 2
		}
		if x+hr.Width > vw-8 {
			x = vw - 8 - hr.Width
		}
		if x < 8 {
			x = 8
		}
	}
	m.showAt(popup, x, y)
}

// ShowAt displays popup with top-left at (x, y).
func (m *MenuPopupManager) ShowAt(popup *MenuPopup, x, y float64) {
	if sc := currentScene(); sc != nil {
		m.ensureNodes(sc)
	}
	m.hideActive()
	popup.build()
	if popup.node == nil {
		return
	}
	m.showAt(popup, x, y)
}

func (m *MenuPopupManager) showAt(popup *MenuPopup, x, y float64) {
	vw, vh := viewportSize()
	if hr, ok := popup.node.HitShape.(sg.HitRect); ok {
		if x+hr.Width > vw-4 {
			x = vw - 4 - hr.Width
		}
		if y+hr.Height > vh-4 {
			y = vh - 4 - hr.Height
		}
	}
	if x < 4 {
		x = 4
	}
	if y < 4 {
		y = 4
	}

	popup.node.SetPosition(x, y)
	m.overlayNode.AddChild(popup.node)
	m.overlayNode.SetVisible(true)

	m.dismissNode.HitShape = sg.HitRect{X: 0, Y: 0, Width: vw, Height: vh}
	m.dismissNode.Invalidate()
	m.dismissNode.SetVisible(true)

	m.active = popup
}

// Hide closes the active popup without selection.
func (m *MenuPopupManager) Hide() {
	m.dismiss()
}

// IsOpen returns true if a popup is currently visible.
func (m *MenuPopupManager) IsOpen() bool {
	return m.active != nil
}

func (m *MenuPopupManager) dismiss() {
	var onDismiss func()
	if m.active != nil {
		onDismiss = m.active.onDismiss
	}
	m.hideActive()
	if onDismiss != nil {
		onDismiss()
	}
}

func (m *MenuPopupManager) hideActive() {
	if m.active == nil {
		return
	}
	if m.overlayNode != nil && m.active.node != nil {
		m.overlayNode.RemoveChild(m.active.node)
		m.overlayNode.SetVisible(false)
	}
	if m.dismissNode != nil {
		m.dismissNode.SetVisible(false)
	}
	m.active = nil
}

// setHighlight is called by item pointer enter.
func (m *MenuPopupManager) setHighlight(idx int) {
	if m.active == nil {
		return
	}
	prev := m.active.highlighted
	m.active.highlighted = idx
	m.active.updateHighlight(prev, idx)
}

// clearHighlight is called by item pointer leave.
func (m *MenuPopupManager) clearHighlight(idx int) {
	if m.active == nil || m.active.highlighted != idx {
		return
	}
	m.active.updateHighlight(idx, -1)
	m.active.highlighted = -1
}

// selectItem fires the item's OnSelect and closes the popup.
func (m *MenuPopupManager) selectItem(idx int) {
	if m.active == nil {
		return
	}
	items := m.active.items
	if idx < 0 || idx >= len(items) {
		return
	}
	item := items[idx]
	if item.Disabled || item.Separator {
		return
	}
	m.hideActive()
	if item.OnSelect != nil {
		item.OnSelect()
	}
}

// tick drives keyboard navigation (arrow keys, Enter, Escape).
func (m *MenuPopupManager) tick() {
	if m.active == nil {
		return
	}

	p := m.active
	items := p.items

	if core.IsKeyJustPressed(engine.KeyArrowDown) {
		next := p.highlighted + 1
		for next < len(items) && (items[next].Separator || items[next].Disabled) {
			next++
		}
		if next < len(items) {
			m.setHighlight(next)
			p.scrollToHighlighted()
		}
	} else if core.IsKeyJustPressed(engine.KeyArrowUp) {
		next := p.highlighted - 1
		for next >= 0 && (items[next].Separator || items[next].Disabled) {
			next--
		}
		if next >= 0 {
			m.setHighlight(next)
			p.scrollToHighlighted()
		}
	} else if core.IsKeyJustPressed(engine.KeyEnter) {
		if p.highlighted >= 0 {
			m.selectItem(p.highlighted)
		}
	} else if core.IsKeyJustPressed(engine.KeyEscape) {
		m.dismiss()
	}
}
