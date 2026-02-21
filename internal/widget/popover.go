package widget

import (
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// newRoundedRect creates a standalone polygon node with a rounded-rect shape.
func newRoundedRect(name string, w, h, r float64, col sg.Color) *sg.Node {
	pts := render.RoundedRectPoints(w, h, r, defaultCornerSegments)
	n := sg.NewPolygon(name, pts)
	n.SetColor(col)
	return n
}

// popoverOverlayZIndex is above menus but below tooltips.
const popoverOverlayZIndex = 525_000

// popoverTitleBarHeight is the fixed height of the optional title bar.
const popoverTitleBarHeight = 32.0

// PopoverSide controls which side of the trigger the popover prefers to appear on.
type PopoverSide int

const (
	PopoverBelow PopoverSide = iota // default
	PopoverAbove
	PopoverRight
	PopoverLeft
)

// Popover is a floating rich-content panel anchored to a trigger component.
// It is similar to Tooltip but dismissable, interactive, and designed for
// heavier content: mini-inspectors, inline documentation, quick-pick panels.
//
// Popovers are managed by DefaultPopoverManager — only one is open at a time.
type Popover struct {
	Component

	preferredSide PopoverSide
	contentW      float64
	contentH      float64
	showClose     bool
	titleText     string
	titleFont     *sg.FontFamily
	titleFontSize float64
	content       UIElement

	open    bool
	onOpen  func()
	onClose func()
}

// NewPopover creates a new Popover with sensible defaults.
func NewPopover(name string) *Popover {
	p := &Popover{}
	initComponent(&p.Component, name)
	p.initBackground(name)
	p.initBorder(name)
	p.onThemeChange = func() {}
	return p
}

// SetPreferredSide sets which side of the trigger the popover prefers to appear on.
func (p *Popover) SetPreferredSide(side PopoverSide) {
	p.preferredSide = side
}

// SetContentSize sets the size of the content area (excluding title bar).
func (p *Popover) SetContentSize(w, h float64) {
	p.contentW = w
	p.contentH = h
}

// SetTitle sets an optional title displayed in the popover header.
func (p *Popover) SetTitle(text string, font *sg.FontFamily, size float64) {
	p.titleText = text
	p.titleFont = font
	p.titleFontSize = size
}

// SetShowCloseButton controls whether an X button appears in the title bar.
func (p *Popover) SetShowCloseButton(v bool) {
	p.showClose = v
}

// SetContent sets the component displayed in the popover body.
func (p *Popover) SetContent(comp UIElement) {
	p.content = comp
}

// SetOnOpen sets a callback fired when the popover opens.
func (p *Popover) SetOnOpen(fn func()) {
	p.onOpen = fn
}

// SetOnClose sets a callback fired when the popover closes.
func (p *Popover) SetOnClose(fn func()) {
	p.onClose = fn
}

// Open opens the popover anchored to the given trigger component.
// If another popover is open it will be closed first.
func (p *Popover) Open(trigger *Component) {
	DefaultPopoverManager.open(p, trigger)
}

// Close closes the popover if it is currently open.
func (p *Popover) Close() {
	DefaultPopoverManager.close(p)
}

// IsOpen returns true when the popover is currently visible.
func (p *Popover) IsOpen() bool {
	return p.open
}

// ---------------------------------------------------------------------------
// PopoverManager
// ---------------------------------------------------------------------------

// PopoverManager manages the single active floating popover.
// Use DefaultPopoverManager; do not construct your own.
type PopoverManager struct {
	scene       *sg.Scene
	overlayNode *sg.Node // active popover node lives here
	dismissNode *sg.Node // full-screen transparent click-catcher

	active     *Popover
	activeRoot *sg.Node // the root node added to overlayNode for active
}

// DefaultPopoverManager is the singleton used by all Popover instances.
var DefaultPopoverManager = &PopoverManager{}

// setScene is called from widget.SetScene.
func (m *PopoverManager) setScene(s *sg.Scene) {
	if m.active != nil {
		m.hideActive(false)
	}
	m.scene = s
	if s == nil || s.Root == nil {
		return
	}
	m.ensureNodes(s)
}

func (m *PopoverManager) ensureNodes(s *sg.Scene) {
	if s == nil || s.Root == nil {
		return
	}
	if m.overlayNode == nil {
		m.overlayNode = sg.NewContainer("popover-overlay")
		m.overlayNode.Interactable = true
		m.overlayNode.SetVisible(false)
		m.overlayNode.SetZIndex(popoverOverlayZIndex)
	}
	if m.dismissNode == nil {
		vw, vh := viewportSize()
		m.dismissNode = sg.NewContainer("popover-dismiss")
		m.dismissNode.Interactable = true
		m.dismissNode.HitShape = sg.HitRect{X: 0, Y: 0, Width: vw, Height: vh}
		m.dismissNode.SetZIndex(popoverOverlayZIndex - 1)
		m.dismissNode.SetVisible(false)
		m.dismissNode.OnPointerDown(func(_ sg.PointerContext) {
			DefaultPopoverManager.dismiss()
		})
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

// open displays the popover anchored to trigger.
func (m *PopoverManager) open(p *Popover, trigger *Component) {
	if sc := currentScene(); sc != nil {
		m.ensureNodes(sc)
	}
	// Close any existing popover first.
	if m.active != nil {
		m.hideActive(true)
	}

	// Build the node tree for this popover.
	theme := getDefaultTheme()
	group := theme.Popover.Group(p.Variant())
	pad := group.Padding

	titleH := 0.0
	if p.titleText != "" || p.showClose {
		titleH = popoverTitleBarHeight
	}
	totalW := p.contentW + pad.Left + pad.Right
	totalH := p.contentH + titleH + pad.Top + pad.Bottom

	root := m.buildRoot(p, totalW, totalH, titleH)

	// Position relative to trigger.
	x, y := m.computePosition(trigger, p.preferredSide, totalW, totalH)
	root.SetPosition(x, y)

	if m.overlayNode == nil {
		// No scene — mark open and fire callback but don't add to overlay.
		p.open = true
		m.active = p
		m.activeRoot = root
		if p.onOpen != nil {
			p.onOpen()
		}
		return
	}

	m.overlayNode.AddChild(root)
	m.overlayNode.SetVisible(true)

	if m.dismissNode != nil {
		vw, vh := viewportSize()
		m.dismissNode.HitShape = sg.HitRect{X: 0, Y: 0, Width: vw, Height: vh}
		m.dismissNode.Invalidate()
		m.dismissNode.SetVisible(true)
	}

	p.open = true
	m.active = p
	m.activeRoot = root

	if p.onOpen != nil {
		p.onOpen()
	}
}

// buildRoot constructs the willow node tree for the popover and returns the root node.
func (m *PopoverManager) buildRoot(p *Popover, totalW, totalH, titleH float64) *sg.Node {
	theme := getDefaultTheme()
	group := theme.Popover.Group(p.Variant())

	// Root container — interactable so clicks inside don't reach dismiss node.
	root := sg.NewContainer("popover-root")
	root.Interactable = true
	root.HitShape = sg.HitRect{X: 0, Y: 0, Width: totalW, Height: totalH}

	// Background.
	bgColor := group.Background.Resolve(StateDefault).Color
	cr := resolveCornerRadius(group.CornerRadius, totalH)
	if cr > 0 {
		bgNode := newRoundedRect("popover-bg", totalW, totalH, cr, bgColor)
		root.AddChild(bgNode)
	} else {
		bg := sg.NewSprite("popover-bg", sg.TextureRegion{})
		bg.SetScale(totalW, totalH)
		bg.SetColor(bgColor)
		root.AddChild(bg)
	}

	// Border.
	bw := group.BorderWidth
	if bw > 0 && cr <= 0 {
		borderCol := group.BorderColor.Resolve(StateDefault)
		top := sg.NewSprite("popover-border-t", sg.TextureRegion{})
		top.SetScale(totalW, bw)
		top.SetColor(borderCol)
		top.SetZIndex(1)

		bot := sg.NewSprite("popover-border-b", sg.TextureRegion{})
		bot.SetScale(totalW, bw)
		bot.SetPosition(0, totalH-bw)
		bot.SetColor(borderCol)
		bot.SetZIndex(1)

		left := sg.NewSprite("popover-border-l", sg.TextureRegion{})
		left.SetScale(bw, totalH-bw*2)
		left.SetPosition(0, bw)
		left.SetColor(borderCol)
		left.SetZIndex(1)

		right := sg.NewSprite("popover-border-r", sg.TextureRegion{})
		right.SetScale(bw, totalH-bw*2)
		right.SetPosition(totalW-bw, bw)
		right.SetColor(borderCol)
		right.SetZIndex(1)

		for _, bd := range []*sg.Node{top, bot, left, right} {
			root.AddChild(bd)
		}
	}

	// Title bar.
	if titleH > 0 {
		// Separator line at bottom of title bar.
		sepColor := group.BorderColor.Resolve(StateDefault)
		if sepColor.A() == 0 {
			sepColor = sg.RGBA(1, 1, 1, 0.15)
		}
		sep := sg.NewSprite("popover-title-sep", sg.TextureRegion{})
		sep.SetScale(totalW, 1)
		sep.SetPosition(0, titleH-1)
		sep.SetColor(sepColor)
		sep.SetZIndex(2)
		root.AddChild(sep)

		// Title label.
		if p.titleText != "" && p.titleFont != nil {
			lbl := sg.NewText("popover-title", p.titleText, p.titleFont)
			lbl.TextBlock.FontSize = p.titleFontSize
			titleColor := group.TitleColor.Resolve(StateDefault)
			if titleColor.A() == 0 {
				titleColor = sg.RGBA(1, 1, 1, 1)
			}
			lbl.TextBlock.Color = titleColor
			lbl.SetPosition(8, (titleH-p.titleFontSize)/2)
			lbl.SetZIndex(3)
			root.AddChild(lbl)
		}

		// Close button: transparent hit area + "×" text sibling in root.
		if p.showClose {
			const closeBtnW = 32.0

			closeBg := sg.NewContainer("popover-close-hit")
			closeBg.Interactable = true
			closeBg.HitShape = sg.HitRect{X: 0, Y: 0, Width: closeBtnW, Height: titleH}
			closeBg.SetPosition(totalW-closeBtnW, 0)
			closeBg.SetZIndex(5)
			closeBg.OnPointerDown(func(_ sg.PointerContext) {
				DefaultPopoverManager.dismiss()
			})
			root.AddChild(closeBg)

			if p.titleFont != nil {
				titleColor := group.TitleColor.Resolve(StateDefault)
				if titleColor.A() == 0 {
					titleColor = sg.RGBA(1, 1, 1, 0.6)
				} else {
					titleColor = sg.RGBA(titleColor.R(), titleColor.G(), titleColor.B(), 0.6)
				}
				closeLabel := sg.NewText("popover-close-x", "x", p.titleFont)
				closeLabel.TextBlock.FontSize = p.titleFontSize
				closeLabel.TextBlock.Color = titleColor
				closeLabel.SetPosition(totalW-closeBtnW+10, (titleH-p.titleFontSize)/2)
				closeLabel.SetZIndex(3)
				root.AddChild(closeLabel)
			}
		}
	}

	// Content area: apply padding so content doesn't press against the edges.
	if p.content != nil {
		pad := group.Padding
		contentComp := p.content.base()
		contentComp.X = pad.Left
		contentComp.Y = titleH + pad.Top
		contentComp.MarkLayoutDirty()
		contentComp.UpdateLayout()
		contentComp.node.SetZIndex(2)
		root.AddChild(contentComp.node)
	}

	return root
}

// computePosition returns the world-space top-left for the popover given a trigger and preferred side.
func (m *PopoverManager) computePosition(trigger *Component, side PopoverSide, w, h float64) (float64, float64) {
	if trigger == nil {
		vw, vh := viewportSize()
		return (vw - w) / 2, (vh - h) / 2
	}

	wx, wy := trigger.node.LocalToWorld(0, 0)
	tw := trigger.Width
	th := trigger.Height
	vw, vh := viewportSize()
	const margin = 8.0

	// Try preferred side, then opposite, then below, then above.
	tryOrder := [4]PopoverSide{side, popoverOpposite(side), PopoverBelow, PopoverAbove}

	for _, s := range tryOrder {
		x, y := popoverSidePos(s, wx, wy, tw, th, w, h)
		if x >= margin && y >= margin && x+w <= vw-margin && y+h <= vh-margin {
			return x, y
		}
	}

	// Last resort: clamp preferred side.
	x, y := popoverSidePos(side, wx, wy, tw, th, w, h)
	return popoverClamp(x, y, w, h, vw, vh, margin)
}

func popoverSidePos(side PopoverSide, wx, wy, tw, th, w, h float64) (float64, float64) {
	switch side {
	case PopoverAbove:
		return wx + tw/2 - w/2, wy - h - 4
	case PopoverRight:
		return wx + tw + 4, wy + th/2 - h/2
	case PopoverLeft:
		return wx - w - 4, wy + th/2 - h/2
	default: // PopoverBelow
		return wx + tw/2 - w/2, wy + th + 4
	}
}

func popoverOpposite(side PopoverSide) PopoverSide {
	switch side {
	case PopoverAbove:
		return PopoverBelow
	case PopoverBelow:
		return PopoverAbove
	case PopoverLeft:
		return PopoverRight
	case PopoverRight:
		return PopoverLeft
	}
	return PopoverBelow
}

func popoverClamp(x, y, w, h, vw, vh, margin float64) (float64, float64) {
	if x < margin {
		x = margin
	}
	if y < margin {
		y = margin
	}
	if x+w > vw-margin {
		x = vw - margin - w
	}
	if y+h > vh-margin {
		y = vh - margin - h
	}
	if x < margin {
		x = margin
	}
	if y < margin {
		y = margin
	}
	return x, y
}

// Open displays popover p anchored to trigger. If another popover is open it
// is closed first. Callers may use this on a custom PopoverManager; otherwise
// prefer p.Open(trigger) which routes through DefaultPopoverManager.
func (m *PopoverManager) Open(p *Popover, trigger *Component) {
	m.open(p, trigger)
}

// Close closes popover p if it is currently open on this manager.
func (m *PopoverManager) Close(p *Popover) {
	m.close(p)
}

// close closes popover p if it is the active one.
func (m *PopoverManager) close(p *Popover) {
	if m.active == p {
		m.hideActive(true)
	}
}

// dismiss closes the active popover (called by dismiss node click).
func (m *PopoverManager) dismiss() {
	m.hideActive(true)
}

// hideActive removes the active popover's node tree from the overlay.
func (m *PopoverManager) hideActive(fireCallback bool) {
	if m.active == nil {
		return
	}
	p := m.active
	root := m.activeRoot
	m.active = nil
	m.activeRoot = nil

	// Detach content node before removing root so the content component
	// remains alive for the next open call.
	if p.content != nil {
		contentNode := p.content.base().node
		if contentNode.Parent != nil {
			contentNode.Parent.RemoveChild(contentNode)
		}
	}

	if m.overlayNode != nil && root != nil {
		m.overlayNode.RemoveChild(root)
	}
	if m.overlayNode != nil {
		m.overlayNode.SetVisible(false)
	}
	if m.dismissNode != nil {
		m.dismissNode.SetVisible(false)
	}

	p.open = false

	if fireCallback && p.onClose != nil {
		p.onClose()
	}
}
