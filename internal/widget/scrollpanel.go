package widget

import (
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// ScrollPanel is a Panel with built-in scrolling. Content that exceeds the
// viewport is clipped via a willow mask, and optional horizontal/vertical
// scrollbars allow the user to navigate.
type ScrollPanel struct {
	Panel
	viewport   *sg.Node // masked container that clips content
	content    *sg.Node // movable container holding child nodes
	vScrollBar *ScrollBar
	hScrollBar *ScrollBar
	scrollX    *Ref[float64]
	scrollY    *Ref[float64]
	watchX     WatchHandle
	watchY     WatchHandle

	contentW float64 // total content width
	contentH float64 // total content height
	showVScr bool
	showHScr bool
}

// scrollWheelSpeed is the number of pixels per mouse wheel tick.
const scrollWheelSpeed = 40

// NewScrollPanel creates a scroll panel with a vertical scrollbar shown by
// default and a horizontal scrollbar hidden.
func NewScrollPanel(name string) *ScrollPanel {
	sp := &ScrollPanel{
		scrollX:  NewRef(0.0),
		scrollY:  NewRef(0.0),
		showVScr: true,
	}
	initComponent(&sp.Component, name)

	sp.initBackground(name)
	sp.initBorder(name)

	// Wire theme change handler (inherits Panel's pattern).
	sp.onThemeChange = func() { sp.applyThemeColors() }
	sp.applyThemeColors()

	// Wire focus-child scroll: when a descendant gains focus, scroll it into view.
	sp.ensureChildVisible = func(child *Component) {
		sp.EnsureVisible(child)
	}

	// Viewport: masked container that clips children.
	// Must be Interactable=true so willow traverses into content children.
	sp.viewport = sg.NewContainer(name + "-viewport")
	sp.viewport.Interactable = true
	sp.node.AddChild(sp.viewport)

	// Content: movable container inside the viewport.
	// Must be Interactable=true so willow traverses into item children.
	sp.content = sg.NewContainer(name + "-content")
	sp.content.Interactable = true
	sp.viewport.AddChild(sp.content)

	// Vertical scrollbar.
	sp.vScrollBar = NewScrollBar(name + "-vscroll")
	sp.vScrollBar.SetOrientation(Vertical)
	sp.vScrollBar.AddToNode(sp.node)
	sp.vScrollBar.SetOnChange(func(v float64) {
		sp.scrollY.Set(v)
		DefaultScheduler.Flush()
		sp.syncContentPosition()
	})

	// Horizontal scrollbar.
	sp.hScrollBar = NewScrollBar(name + "-hscroll")
	sp.hScrollBar.SetOrientation(Horizontal)
	sp.hScrollBar.SetVisible(false)
	sp.hScrollBar.AddToNode(sp.node)
	sp.hScrollBar.SetOnChange(func(v float64) {
		sp.scrollX.Set(v)
		DefaultScheduler.Flush()
		sp.syncContentPosition()
	})

	// Auto-update: mouse wheel scrolling via willow's per-frame hook.
	sp.node.OnUpdate = func(_ float64) {
		sp.Update()
	}

	// Default size.
	sp.SetSize(300, 200)

	return sp
}

// ContentNode returns the content container node. Add child willow nodes
// directly to this node for scrolled content that is not managed by the
// Component child system.
func (sp *ScrollPanel) ContentNode() *sg.Node {
	return sp.content
}

// AddContent adds a UIElement component to the scroll panel's content node.
func (sp *ScrollPanel) AddContent(child UIElement) {
	if child == nil {
		return
	}
	sp.content.AddChild(child.base().node)
}

// SetContentSize sets the total size of the scrollable content area.
func (sp *ScrollPanel) SetContentSize(w, h float64) {
	sp.contentW = w
	sp.contentH = h
	sp.updateScrollBars()
}

// SetScrollX sets the horizontal scroll position.
func (sp *ScrollPanel) SetScrollX(v float64) {
	sp.hScrollBar.SetScrollPos(v)
	sp.scrollX.Set(sp.hScrollBar.ScrollPos())
	DefaultScheduler.Flush()
	sp.syncContentPosition()
}

// SetScrollY sets the vertical scroll position.
func (sp *ScrollPanel) SetScrollY(v float64) {
	sp.vScrollBar.SetScrollPos(v)
	sp.scrollY.Set(sp.vScrollBar.ScrollPos())
	DefaultScheduler.Flush()
	sp.syncContentPosition()
}

// ScrollTo sets both scroll positions at once.
func (sp *ScrollPanel) ScrollTo(x, y float64) {
	sp.SetScrollX(x)
	sp.SetScrollY(y)
}

// ScrollX returns the current horizontal scroll position.
func (sp *ScrollPanel) ScrollX() float64 {
	return sp.scrollX.Peek()
}

// ScrollY returns the current vertical scroll position.
func (sp *ScrollPanel) ScrollY() float64 {
	return sp.scrollY.Peek()
}

// ShowVScroll shows or hides the vertical scrollbar.
func (sp *ScrollPanel) ShowVScroll(show bool) {
	sp.showVScr = show
	sp.vScrollBar.SetVisible(show)
	sp.updateScrollLayout()
}

// ShowHScroll shows or hides the horizontal scrollbar.
func (sp *ScrollPanel) ShowHScroll(show bool) {
	sp.showHScr = show
	sp.hScrollBar.SetVisible(show)
	sp.updateScrollLayout()
}

// SetSize sets the scroll panel dimensions.
func (sp *ScrollPanel) SetSize(w, h float64) {
	sp.Width = w
	sp.Height = h
	sp.resizeBackground(w, h)
	sp.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	sp.resizeBorder(w, h)
	sp.updateScrollLayout()
	sp.MarkLayoutDirty()
}

// SetBackground sets the panel's background color as a manual override.
func (sp *ScrollPanel) SetBackground(c sg.Color) {
	sp.bgOverride = true
	sp.bgNode.SetColor(c)
	sp.MarkDrawDirty()
}

// SetBorder sets the border color and width as a manual override.
func (sp *ScrollPanel) SetBorder(c sg.Color, width float64) {
	sp.borderOverride = true
	sp.applyBorder(c, width, Background{Type: BgSolid})
	sp.updateScrollLayout()
	sp.MarkDrawDirty()
}

// Update processes mouse wheel input for scrolling. Call from your scene's
// UpdateFunc.
func (sp *ScrollPanel) Update() {
	if !sp.containsCursor() {
		return
	}
	if !sp.isInActiveWindow() {
		return
	}
	_, wy := engine.Wheel()
	if wy != 0 {
		newY := sp.scrollY.Peek() - wy*scrollWheelSpeed
		sp.SetScrollY(newY)
	}
}

// isInActiveWindow returns true when this scroll panel is either not inside
// any managed window, or is inside the currently active (foremost) window.
// It walks the Component parent chain and compares against the registered
// windows in DefaultWindowManager.
func (sp *ScrollPanel) isInActiveWindow() bool {
	if len(DefaultWindowManager.windows) == 0 {
		return true
	}
	cur := sp.Component.parent
	for cur != nil {
		for _, w := range DefaultWindowManager.windows {
			if cur == &w.Component {
				return DefaultWindowManager.active == w
			}
		}
		cur = cur.parent
	}
	return true // not inside any managed window
}

// AddChild adds a child component to the scroll panel. The child's node is
// placed inside the scrollable content container.
func (sp *ScrollPanel) AddChild(child UIElement) {
	if child == nil {
		return
	}
	cc := child.base()
	if cc.parent != nil {
		cc.parent.RemoveChild(child)
	}
	cc.parent = &sp.Component
	sp.children = append(sp.children, cc)
	sp.content.AddChild(cc.node)
	sp.MarkLayoutDirty()

	if cc.theme == nil {
		cc.propagateThemeChange()
	}
}

// RemoveChild detaches a child component from the scroll panel's content.
func (sp *ScrollPanel) RemoveChild(child UIElement) {
	if child == nil {
		return
	}
	cc := child.base()
	if cc.parent != &sp.Component {
		return
	}
	for i, ch := range sp.children {
		if ch == cc {
			copy(sp.children[i:], sp.children[i+1:])
			sp.children[len(sp.children)-1] = nil
			sp.children = sp.children[:len(sp.children)-1]
			break
		}
	}
	cc.parent = nil
	sp.content.RemoveChild(cc.node)
	sp.MarkLayoutDirty()
}

// EnsureVisible scrolls so that the given child component is fully visible
// within the scroll panel's viewport. If the component is already fully
// visible, no scrolling occurs.
func (sp *ScrollPanel) EnsureVisible(child *Component) {
	if child == nil {
		return
	}
	// Compute child's position relative to the content container.
	// Walk the node parent chain from child up to sp.content to accumulate offset.
	var relX, relY float64
	for n := child.node; n != nil && n != sp.content; n = n.Parent {
		relX += n.X()
		relY += n.Y()
	}

	bw := sp.borderWidth_
	vpW := sp.Width - bw*2
	vpH := sp.Height - bw*2
	if sp.showVScr {
		vpW -= float64(DefaultScrollBarWidth)
	}
	if sp.showHScr {
		vpH -= float64(DefaultScrollBarWidth)
	}

	// Vertical.
	sy := sp.scrollY.Peek()
	if relY < sy {
		sp.SetScrollY(relY)
	} else if relY+child.Height > sy+vpH {
		sp.SetScrollY(relY + child.Height - vpH)
	}

	// Horizontal.
	sx := sp.scrollX.Peek()
	if relX < sx {
		sp.SetScrollX(relX)
	} else if relX+child.Width > sx+vpW {
		sp.SetScrollX(relX + child.Width - vpW)
	}
}

// Dispose cleans up scrollbars and watches.
func (sp *ScrollPanel) Dispose() {
	sp.watchX.Stop()
	sp.watchY.Stop()
	sp.vScrollBar.Dispose()
	sp.hScrollBar.Dispose()
	sp.Component.Dispose()
}

// updateScrollLayout recalculates viewport and scrollbar positions/sizes
// based on the current panel dimensions and border width.
func (sp *ScrollPanel) updateScrollLayout() {
	bw := sp.borderWidth_
	sbW := float64(DefaultScrollBarWidth)

	// Viewport area.
	vpX := bw
	vpY := bw
	vpW := sp.Width - bw*2
	vpH := sp.Height - bw*2

	// Shrink viewport for visible scrollbars.
	if sp.showVScr {
		vpW -= sbW
	}
	if sp.showHScr {
		vpH -= sbW
	}
	if vpW < 0 {
		vpW = 0
	}
	if vpH < 0 {
		vpH = 0
	}

	sp.viewport.SetPosition(vpX, vpY)

	// Apply mask for clipping. Root mask transform is ignored by willow,
	// so we wrap the mask shape in a container.
	maskRoot := sg.NewContainer(sp.node.Name + "-mask")
	maskSprite := sg.NewSprite(sp.node.Name+"-mask-rect", sg.TextureRegion{})
	maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	maskSprite.SetScale(vpW, vpH)
	maskRoot.AddChild(maskSprite)
	sp.viewport.SetMask(maskRoot)

	// Position scrollbars.
	if sp.showVScr {
		sp.vScrollBar.SetSize(sbW, vpH)
		sp.vScrollBar.SetPosition(sp.Width-bw-sbW, bw)
	}
	if sp.showHScr {
		hsbW := vpW
		sp.hScrollBar.SetSize(hsbW, sbW)
		sp.hScrollBar.SetPosition(bw, sp.Height-bw-sbW)
	}

	sp.updateScrollBars()
}

// updateScrollBars syncs scrollbar content sizes with the current
// content and viewport dimensions.
func (sp *ScrollPanel) updateScrollBars() {
	bw := sp.borderWidth_
	sbW := float64(DefaultScrollBarWidth)

	vpW := sp.Width - bw*2
	vpH := sp.Height - bw*2
	if sp.showVScr {
		vpW -= sbW
	}
	if sp.showHScr {
		vpH -= sbW
	}
	if vpW < 0 {
		vpW = 0
	}
	if vpH < 0 {
		vpH = 0
	}

	sp.vScrollBar.SetContentSize(sp.contentH, vpH)
	sp.hScrollBar.SetContentSize(sp.contentW, vpW)

	// Re-clamp current positions.
	sp.vScrollBar.SetScrollPos(sp.scrollY.Peek())
	sp.hScrollBar.SetScrollPos(sp.scrollX.Peek())

	sp.syncContentPosition()
}

// syncContentPosition moves the content container to reflect the current
// scroll offsets.
func (sp *ScrollPanel) syncContentPosition() {
	sp.content.SetPosition(-sp.scrollX.Peek(), -sp.scrollY.Peek())
}

// Viewport returns the masked container node that clips content.
// Used for testing scroll panel internals.
func (sp *ScrollPanel) Viewport() *sg.Node { return sp.viewport }

// VScrollBar returns the vertical scrollbar widget.
// Used for testing scroll panel internals.
func (sp *ScrollPanel) VScrollBar() *ScrollBar { return sp.vScrollBar }

// HScrollBar returns the horizontal scrollbar widget.
// Used for testing scroll panel internals.
func (sp *ScrollPanel) HScrollBar() *ScrollBar { return sp.hScrollBar }

// ContentH returns the total content height set via SetContentSize.
// Used for testing scroll panel internals.
func (sp *ScrollPanel) ContentH() float64 { return sp.contentH }
