package widget

import (
	"github.com/devthicket/willowui/internal/sg"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

// NavDrawerAnchor specifies which edge the drawer slides from.
type NavDrawerAnchor int

const (
	NavDrawerLeft NavDrawerAnchor = iota
	NavDrawerRight
)

// Default NavDrawer dimensions and animation.
const (
	DefaultNavDrawerWidth    = 240
	DefaultNavDrawerDuration = 0.25 // seconds
)

// navDrawerBackdropSize is the half-extent used for the full-screen backdrop.
const navDrawerBackdropSize = 10000.0

// NavDrawer is a slide-out navigation panel anchored to the left or right edge.
// It supports overlay mode (dims backdrop) and pinned mode (always visible,
// no backdrop).
type NavDrawer struct {
	Component

	drawerPanel *Panel    // the sliding panel containing user content
	backdrop    *sg.Node  // semi-transparent overlay behind the drawer
	content     UIElement // the current content component

	anchor               NavDrawerAnchor
	drawerWidth          float64
	open                 bool
	pinned               bool
	closeOnBackdropClick bool
	animDuration         float32 // seconds

	slideTween *gween.Tween // animates the drawer panel X position
	slidePos   float64      // current slide position (0 = closed, 1 = open)

	onOpen  func()
	onClose func()

	// Reactive binding.
	openRef   *Ref[bool]
	openWatch WatchHandle
}

// NewNavDrawer creates a NavDrawer anchored to the left edge by default.
func NewNavDrawer(name string) *NavDrawer {
	d := &NavDrawer{
		anchor:               NavDrawerLeft,
		drawerWidth:          DefaultNavDrawerWidth,
		closeOnBackdropClick: true,
		animDuration:         DefaultNavDrawerDuration,
	}
	initComponent(&d.Component, name)

	// The NavDrawer container should not block input when closed.
	// We disable interactability and remove hit shape by default;
	// Open/Close toggle these.
	d.node.Interactable = false
	d.node.HitShape = nil

	// Backdrop: a full-screen semi-transparent overlay.
	d.backdrop = sg.NewSprite(name+"-backdrop", sg.TextureRegion{})
	d.backdrop.Interactable = true
	d.backdrop.HitShape = sg.HitRect{
		X: 0, Y: 0,
		Width:  navDrawerBackdropSize * 2,
		Height: navDrawerBackdropSize * 2,
	}
	d.backdrop.SetPosition(-navDrawerBackdropSize, -navDrawerBackdropSize)
	d.backdrop.SetScale(navDrawerBackdropSize*2, navDrawerBackdropSize*2)
	d.backdrop.SetVisible(false)
	d.backdrop.OnPointerDown(func(_ sg.PointerContext) {
		if d.closeOnBackdropClick && d.open && !d.pinned {
			d.Close()
		}
	})
	d.node.AddChild(d.backdrop)

	// Drawer panel.
	d.drawerPanel = NewPanel(name + "-drawer")
	d.drawerPanel.parent = &d.Component
	d.node.AddChild(d.drawerPanel.Node())

	// Ensure drawer is above backdrop.
	d.drawerPanel.Node().SetZIndex(1)

	// Auto-update: advance slide animation.
	d.node.OnUpdate = func(dt float64) {
		d.Update(float32(dt))
	}

	d.onThemeChange = func() { d.applyThemeColors() }
	d.applyThemeColors()

	// Start closed.
	d.slidePos = 0
	d.applySlidePosition()

	return d
}

// SetContent sets the drawer's content component.
func (d *NavDrawer) SetContent(comp UIElement) {
	if d.content != nil {
		d.drawerPanel.RemoveChild(d.content)
	}
	d.content = comp
	if comp != nil {
		d.drawerPanel.AddChild(comp)
	}
}

// Open slides the drawer into view.
func (d *NavDrawer) Open() {
	if d.open {
		return
	}
	d.open = true
	d.node.Interactable = true
	d.startSlideAnimation(1)
	if !d.pinned {
		d.backdrop.SetVisible(true)
	}
	if d.openRef != nil {
		d.openRef.Set(true)
	}
	if d.onOpen != nil {
		d.onOpen()
	}
}

// Close slides the drawer out of view.
func (d *NavDrawer) Close() {
	if !d.open {
		return
	}
	d.open = false
	d.startSlideAnimation(0)
	if d.openRef != nil {
		d.openRef.Set(false)
	}
	if d.onClose != nil {
		d.onClose()
	}
}

// Toggle opens the drawer if closed, or closes it if open.
func (d *NavDrawer) Toggle() {
	if d.open {
		d.Close()
	} else {
		d.Open()
	}
}

// IsOpen returns whether the drawer is currently open (or opening).
func (d *NavDrawer) IsOpen() bool {
	return d.open
}

// SetPinned sets whether the drawer is pinned open (always visible, no backdrop).
func (d *NavDrawer) SetPinned(v bool) {
	d.pinned = v
	if v {
		d.backdrop.SetVisible(false)
		// When pinning, force open.
		if !d.open {
			d.open = true
			d.node.Interactable = true
			d.slidePos = 1
			d.slideTween = nil
			d.applySlidePosition()
			if d.onOpen != nil {
				d.onOpen()
			}
		}
	} else {
		// When unpinning, show backdrop if open.
		if d.open {
			d.backdrop.SetVisible(true)
		}
	}
}

// IsPinned returns whether the drawer is pinned open.
func (d *NavDrawer) IsPinned() bool {
	return d.pinned
}

// SetAnchor sets which edge the drawer slides from.
func (d *NavDrawer) SetAnchor(anchor NavDrawerAnchor) {
	d.anchor = anchor
	d.applySlidePosition()
}

// SetWidth sets the drawer panel width.
func (d *NavDrawer) SetWidth(w float64) {
	d.drawerWidth = w
	d.applySlidePosition()
	d.updateDrawerSize()
}

// SetCloseOnBackdropClick sets whether clicking the backdrop closes the drawer.
func (d *NavDrawer) SetCloseOnBackdropClick(v bool) {
	d.closeOnBackdropClick = v
}

// SetOnOpen sets the callback invoked when the drawer opens.
func (d *NavDrawer) SetOnOpen(fn func()) {
	d.onOpen = fn
}

// SetOnClose sets the callback invoked when the drawer closes.
func (d *NavDrawer) SetOnClose(fn func()) {
	d.onClose = fn
}

// BindOpen binds the open/closed state to a reactive Ref. Changes to the Ref
// open or close the drawer, and user interactions update the Ref.
func (d *NavDrawer) BindOpen(ref *Ref[bool]) {
	d.openWatch.Stop()
	d.openRef = ref
	if ref.Peek() {
		d.Open()
	} else {
		d.Close()
	}
	d.openWatch = WatchValue(ref, func(_, newVal bool) {
		if newVal {
			d.Open()
		} else {
			d.Close()
		}
	})
}

// SetSize sets the overall NavDrawer container size (typically the full screen).
func (d *NavDrawer) SetSize(w, h float64) {
	d.Width = w
	d.Height = h
	d.updateDrawerSize()
	d.applySlidePosition()
	d.MarkLayoutDirty()
}

// SetAnimationDuration sets the slide animation duration in seconds.
func (d *NavDrawer) SetAnimationDuration(seconds float32) {
	d.animDuration = seconds
}

// DrawerPanel returns the inner drawer panel for direct access.
func (d *NavDrawer) DrawerPanel() *Panel {
	return d.drawerPanel
}

// Backdrop returns the backdrop node (for testing).
func (d *NavDrawer) Backdrop() *sg.Node {
	return d.backdrop
}

// Update advances the slide animation.
func (d *NavDrawer) Update(dt float32) {
	if d.slideTween != nil {
		val, done := d.slideTween.Update(dt)
		d.slidePos = float64(val)
		d.applySlidePosition()
		if done {
			d.slideTween = nil
			// Hide backdrop and disable input after close animation completes.
			if !d.open {
				d.backdrop.SetVisible(false)
				d.node.Interactable = false
			}
		}
	}
}

// Dispose cleans up the drawer and its children.
func (d *NavDrawer) Dispose() {
	d.openWatch.Stop()
	if d.drawerPanel != nil {
		d.drawerPanel.Dispose()
	}
	d.Component.Dispose()
}

// startSlideAnimation begins a tween from the current position to the target.
func (d *NavDrawer) startSlideAnimation(target float64) {
	dur := d.animDuration
	if dur <= 0 {
		// Instant.
		d.slidePos = target
		d.slideTween = nil
		d.applySlidePosition()
		if !d.open {
			d.backdrop.SetVisible(false)
			d.node.Interactable = false
		}
		return
	}
	d.slideTween = gween.New(float32(d.slidePos), float32(target), dur, ease.OutCubic)
}

// applySlidePosition updates the drawer panel X position based on slidePos.
// slidePos 0 = fully closed (off-screen), 1 = fully open.
func (d *NavDrawer) applySlidePosition() {
	var x float64
	switch d.anchor {
	case NavDrawerLeft:
		// Closed: -drawerWidth, Open: 0
		x = -d.drawerWidth + (d.slidePos * d.drawerWidth)
	case NavDrawerRight:
		// Closed: Width, Open: Width - drawerWidth
		x = d.Width - (d.slidePos * d.drawerWidth)
	}
	d.drawerPanel.SetPosition(x, 0)
}

// updateDrawerSize resizes the drawer panel to match the configured width
// and the container height.
func (d *NavDrawer) updateDrawerSize() {
	d.drawerPanel.SetSize(d.drawerWidth, d.Height)
}

// applyThemeColors reads the NavDrawerGroup from the effective theme and
// applies background, border, and backdrop colors.
func (d *NavDrawer) applyThemeColors() {
	group := d.EffectiveTheme().NavDrawer.Group(d.Variant())

	// Drawer panel background and border.
	bg := group.Background.Resolve(StateDefault)
	d.drawerPanel.SetBackground(bg.Color)

	if group.BorderWidth > 0 {
		d.drawerPanel.SetBorder(group.BorderColor.Resolve(StateDefault), group.BorderWidth)
	}

	// Padding.
	d.drawerPanel.SetPadding(group.Padding.Top, group.Padding.Right, group.Padding.Bottom, group.Padding.Left)

	// Backdrop color.
	d.backdrop.SetColor(group.BackdropColor.Resolve(StateDefault))

	d.MarkDrawDirty()
}
