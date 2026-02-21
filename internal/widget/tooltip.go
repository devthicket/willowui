package widget

import (
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// tooltipOverlayZIndex is the z-index for the tooltip overlay node.
// It renders above windows and all other UI elements.
const tooltipOverlayZIndex = 1_000_000

// TooltipAnchor controls where a tooltip is placed relative to its trigger.
type TooltipAnchor int

const (
	// Trigger-relative anchors.
	TooltipAuto  TooltipAnchor = iota // pick direction with most available space
	TooltipAbove                      // centered horizontally, flush above trigger
	TooltipBelow                      // centered horizontally, flush below trigger
	TooltipLeft                       // centered vertically, flush left of trigger
	TooltipRight                      // centered vertically, flush right of trigger

	// Viewport-corner anchors.
	TooltipCornerTopLeft
	TooltipCornerTopRight
	TooltipCornerBottomLeft
	TooltipCornerBottomRight

	// Mouse-tracking anchor.
	TooltipFollowMouse // repositions every frame at cursor + OffsetX/OffsetY
)

// Tooltip is a floating overlay that appears after a hover delay, anchored
// relative to the triggering widget. It embeds Component, giving it full
// layout and child-management capabilities.
//
// Tooltips are never added directly to the scene — the DefaultTooltipManager
// manages their lifecycle.
type Tooltip struct {
	Component

	// Configuration — set before first show; take effect on next show.
	ShowDelay       int           // frames of hover before appearing (default 30)
	HideDelay       int           // frames after cursor leaves before hiding (default 0)
	Anchor          TooltipAnchor // placement strategy (default TooltipAuto)
	OffsetX         float64       // horizontal nudge after placement (+right)
	OffsetY         float64       // vertical nudge after placement (+down) (default 4)
	FadeInDuration  float32       // fade-in duration in seconds (0 = instant)
	FadeOutDuration float32       // fade-out duration in seconds (0 = instant)
	FadeInEase      sg.EaseFunc   // easing for fade-in (nil = linear)
	FadeOutEase     sg.EaseFunc   // easing for fade-out (nil = linear)
	ClampToScreen   bool          // keep fully inside viewport (default true)
	ClampMargin     float64       // minimum gap from each viewport edge (default 4)

	// showing tracks whether the manager has added this tooltip to the overlay.
	showing bool
	// explicitSize is true when SetSize was called; suppresses auto-sizing on show.
	explicitSize bool
}

// NewTooltip creates a Tooltip with sensible defaults.
func NewTooltip(name string) *Tooltip {
	tt := &Tooltip{}
	initComponent(&tt.Component, name)
	tt.initBackground(name)
	tt.initBorder(name)

	// Defaults from spec.
	tt.ShowDelay = 30
	tt.OffsetY = 4
	tt.ClampToScreen = true
	tt.ClampMargin = 4
	tt.Padding = AutoPadding
	tt.Layout = LayoutVBox

	// Wire theme changes.
	tt.onThemeChange = func() { tt.applyThemeColors() }

	return tt
}

// applyThemeColors reads the TooltipGroup from the effective theme and applies
// background, border, and padding defaults.
func (tt *Tooltip) applyThemeColors() {
	group := tt.EffectiveTheme().Tooltip.Group(tt.Variant())
	tt.applyCornerRadius(group.CornerRadius)
	bg := group.Background.Resolve(tt.state)
	tt.applyBackground(bg)
	tt.applyBorder(group.BorderColor.Resolve(tt.state), group.BorderWidth, bg)
	// Resolve auto-padding from theme on first apply; leave explicit padding alone.
	if tt.Padding.IsAuto() {
		tt.Padding = group.Padding
	}
	tt.MarkDrawDirty()
}

// SetSize sets explicit dimensions, bypassing auto-sizing.
func (tt *Tooltip) SetSize(w, h float64) {
	tt.explicitSize = true
	tt.Width = w
	tt.Height = h
	tt.resizeBackground(w, h)
	tt.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	tt.resizeBorder(w, h)
	tt.MarkLayoutDirty()
}

// SetText replaces all children with a single Label displaying text.
func (tt *Tooltip) SetText(text string, source *sg.FontFamily, size float64) {
	children := tt.Component.Children()
	for len(children) > 0 {
		tt.Component.RemoveChild(children[0])
		children = tt.Component.Children()
	}
	lbl := NewLabel(tt.node.Name+"-label", text, source, size)
	tt.Component.AddChild(lbl)
	tt.MarkLayoutDirty()
}

// SetShowDelay sets the ShowDelay.
func (tt *Tooltip) SetShowDelay(frames int) { tt.ShowDelay = frames }

// SetHideDelay sets the HideDelay.
func (tt *Tooltip) SetHideDelay(frames int) { tt.HideDelay = frames }

// SetAnchor sets the Anchor placement strategy.
func (tt *Tooltip) SetAnchor(a TooltipAnchor) { tt.Anchor = a }

// SetOffset sets OffsetX and OffsetY simultaneously.
func (tt *Tooltip) SetOffset(x, y float64) { tt.OffsetX = x; tt.OffsetY = y }

// SetFadeIn sets the fade-in duration in seconds.
func (tt *Tooltip) SetFadeIn(seconds float32) { tt.FadeInDuration = seconds }

// SetFadeOut sets the fade-out duration in seconds.
func (tt *Tooltip) SetFadeOut(seconds float32) { tt.FadeOutDuration = seconds }

// SetClampToScreen enables or disables viewport clamping.
func (tt *Tooltip) SetClampToScreen(v bool) { tt.ClampToScreen = v }

// SetClampMargin sets the minimum pixel gap from each viewport edge.
func (tt *Tooltip) SetClampMargin(px float64) { tt.ClampMargin = px }

// Show displays the tooltip at the given screen position, bypassing the hover
// delay. The tooltip is not associated with any trigger; hide it with Hide().
func (tt *Tooltip) Show(x, y float64) {
	DefaultTooltipManager.showAtPosition(tt, x, y)
}

// Hide hides the tooltip immediately, bypassing HideDelay and FadeOut.
func (tt *Tooltip) Hide() {
	DefaultTooltipManager.hideTooltip(tt)
}

// IsShowing returns true when the tooltip is currently visible.
func (tt *Tooltip) IsShowing() bool {
	return tt.showing
}

// Dispose hides the tooltip and releases resources.
func (tt *Tooltip) Dispose() {
	DefaultTooltipManager.onTooltipDisposed(tt)
	tt.Component.Dispose()
}

// ---------------------------------------------------------------------------
// TooltipManager
// ---------------------------------------------------------------------------

type tooltipManagerState int

const (
	tipIdle tooltipManagerState = iota
	tipShowPending
	tipFadingIn
	tipShowing
	tipHidePending
	tipFadingOut
)

// TooltipManager manages tooltip visibility for the scene. It is driven by
// a ticker node added to the scene root when SetScene is called.
//
// The package-level DefaultTooltipManager is used by all components.
type TooltipManager struct {
	// Enabled controls whether tooltips are shown at all (default true).
	Enabled bool

	scene       *sg.Scene
	tickerNode  *sg.Node // hidden node in scene root; OnUpdate drives tick
	overlayNode *sg.Node // tooltip content is added here when showing

	state         tooltipManagerState
	activeTrigger *Component // current hover trigger (may be nil for Show())
	activeTooltip *Tooltip
	counter       int            // countdown frames for show/hide delay
	activeTween   *sg.TweenGroup // non-nil while a fade tween is running
}

// DefaultTooltipManager is the package-level singleton used by all components.
var DefaultTooltipManager = &TooltipManager{Enabled: true}

// setScene is called from widget.SetScene to hook the manager into the scene.
func (m *TooltipManager) setScene(s *sg.Scene) {
	// Hide active tooltip when scene changes.
	if m.state != tipIdle {
		m.hideImmediate(false)
	}
	m.scene = s
	if s == nil || s.Root == nil {
		return
	}
	m.ensureNodes(s)
}

// ensureNodes lazily creates and attaches the ticker and overlay nodes.
func (m *TooltipManager) ensureNodes(s *sg.Scene) {
	if s == nil || s.Root == nil {
		return
	}

	if m.tickerNode == nil {
		m.tickerNode = sg.NewContainer("tooltip-ticker")
		m.tickerNode.Interactable = false
		m.tickerNode.SetZIndex(tooltipOverlayZIndex)
		m.tickerNode.OnUpdate = func(_ float64) {
			DefaultTooltipManager.tick()
		}
	}
	if m.overlayNode == nil {
		m.overlayNode = sg.NewContainer("tooltip-overlay")
		m.overlayNode.Interactable = false
		m.overlayNode.SetVisible(false)
		m.overlayNode.SetZIndex(tooltipOverlayZIndex)
	}

	// Re-attach to scene root if needed.
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
}

// onTriggerEnter is called when the cursor enters a tooltip-enabled component.
func (m *TooltipManager) onTriggerEnter(c *Component) {
	if !m.Enabled || c.tooltip == nil || !c.enabled {
		return
	}
	// Lazily ensure nodes in case setScene was called before any tooltip existed.
	if sc := currentScene(); sc != nil {
		m.ensureNodes(sc)
	}

	tt := c.tooltip
	switch m.state {
	case tipIdle, tipFadingOut:
		m.startShowCountdown(c, tt)
	case tipHidePending:
		if m.activeTrigger == c && m.activeTooltip == tt {
			// Re-enter during grace period: stay visible.
			m.state = tipShowing
		} else {
			m.hideImmediate(true)
			m.startShowCountdown(c, tt)
		}
	case tipShowing, tipFadingIn:
		if m.activeTrigger != c {
			m.hideImmediate(true)
			m.startShowCountdown(c, tt)
		}
	case tipShowPending:
		if m.activeTrigger != c {
			m.activeTrigger = c
			m.activeTooltip = tt
			m.counter = tt.ShowDelay
		}
	}
}

// onTriggerLeave is called when the cursor leaves a tooltip-enabled component.
func (m *TooltipManager) onTriggerLeave(c *Component) {
	if m.activeTrigger != c {
		return
	}
	switch m.state {
	case tipShowPending:
		m.transitionToIdle()
	case tipShowing, tipFadingIn:
		if m.activeTooltip.HideDelay > 0 {
			m.state = tipHidePending
			m.counter = m.activeTooltip.HideDelay
		} else {
			m.startFadeOut()
		}
	}
}

// onTriggerDisposed is called when a trigger component is disposed.
func (m *TooltipManager) onTriggerDisposed(c *Component) {
	if m.activeTrigger == c {
		m.hideImmediate(false)
	}
}

// onTriggerCleared is called when ClearTooltip is called on a trigger.
func (m *TooltipManager) onTriggerCleared(c *Component) {
	if m.activeTrigger == c {
		m.hideImmediate(false)
	}
}

// onTooltipDisposed is called when a Tooltip is disposed directly.
func (m *TooltipManager) onTooltipDisposed(tt *Tooltip) {
	if m.activeTooltip == tt {
		m.hideImmediate(false)
	}
}

// showAtPosition shows the tooltip at an explicit screen position, bypassing hover.
// For viewport-corner and follow-mouse anchors the passed (x, y) is ignored and
// computePosition is used instead, matching the behaviour of the hover-driven path.
func (m *TooltipManager) showAtPosition(tt *Tooltip, x, y float64) {
	if sc := currentScene(); sc != nil {
		m.ensureNodes(sc)
	}
	m.hideImmediate(false)
	m.activeTrigger = nil
	m.activeTooltip = tt

	// Apply theme and auto-size (same as showTooltip).
	tt.applyThemeColors()
	if !tt.explicitSize {
		tt.SizeToContent()
		tt.resizeBackground(tt.Width, tt.Height)
		tt.resizeBorder(tt.Width, tt.Height)
	}

	// For anchors that don't depend on a trigger position, compute placement
	// from the anchor so the caller doesn't need to supply meaningful coords.
	switch tt.Anchor {
	case TooltipCornerTopLeft, TooltipCornerTopRight,
		TooltipCornerBottomLeft, TooltipCornerBottomRight,
		TooltipFollowMouse:
		x, y = m.computePosition(nil, tt)
	}

	if tt.ClampToScreen {
		x, y = m.clampToScreen(x, y, tt.Width, tt.Height, tt.ClampMargin)
	}
	m.doShowTooltip(x, y)
}

// hideTooltip hides the given tooltip immediately if it is active.
func (m *TooltipManager) hideTooltip(tt *Tooltip) {
	if m.activeTooltip == tt {
		m.hideImmediate(false)
	}
}

// startShowCountdown begins counting toward showing the tooltip.
func (m *TooltipManager) startShowCountdown(c *Component, tt *Tooltip) {
	m.activeTrigger = c
	m.activeTooltip = tt
	m.counter = tt.ShowDelay
	if m.counter <= 0 {
		m.showTooltip()
	} else {
		m.state = tipShowPending
	}
}

// tick is called once per frame via the ticker node's OnUpdate.
func (m *TooltipManager) tick() {
	if !m.Enabled {
		if m.state != tipIdle {
			m.hideImmediate(false)
		}
		return
	}

	switch m.state {
	case tipIdle:
		// nothing

	case tipShowPending:
		if m.activeTrigger != nil && !m.activeTrigger.hovered {
			m.transitionToIdle()
			return
		}
		m.counter--
		if m.counter <= 0 {
			m.showTooltip()
		}

	case tipFadingIn:
		if m.activeTween != nil && m.activeTween.Done {
			m.activeTween = nil
			m.state = tipShowing
		}
		m.maybeUpdateFollowPosition()

	case tipShowing:
		m.maybeUpdateFollowPosition()

	case tipHidePending:
		if m.activeTrigger != nil && m.activeTrigger.hovered {
			m.state = tipShowing
			return
		}
		m.counter--
		if m.counter <= 0 {
			m.startFadeOut()
		}

	case tipFadingOut:
		if m.activeTween != nil && m.activeTween.Done {
			m.activeTween = nil
			m.finishHide()
		}
	}
}

// showTooltip sizes, positions, and adds the active tooltip to the overlay.
func (m *TooltipManager) showTooltip() {
	if m.activeTooltip == nil || m.overlayNode == nil {
		m.transitionToIdle()
		return
	}
	tt := m.activeTooltip
	trigger := m.activeTrigger

	// Fire callback before layout so content can be updated.
	if trigger != nil && trigger.onTooltipShow != nil {
		trigger.onTooltipShow()
	}

	// Apply theme and resolve auto-padding before sizing.
	tt.applyThemeColors()

	// Auto-size unless the user set an explicit size via SetSize.
	if !tt.explicitSize {
		tt.SizeToContent()
		tt.resizeBackground(tt.Width, tt.Height)
		tt.resizeBorder(tt.Width, tt.Height)
	}

	// Compute placement position.
	x, y := m.computePosition(trigger, tt)
	if tt.ClampToScreen {
		x, y = m.clampToScreen(x, y, tt.Width, tt.Height, tt.ClampMargin)
	}

	m.doShowTooltip(x, y)
}

// doShowTooltip attaches the tooltip node to the overlay at (x, y) and starts fade.
func (m *TooltipManager) doShowTooltip(x, y float64) {
	tt := m.activeTooltip
	if tt == nil || m.overlayNode == nil {
		return
	}
	tt.Component.X = x
	tt.Component.Y = y
	tt.Component.node.SetPosition(x, y)
	tt.Component.UpdateLayout()

	m.overlayNode.AddChild(tt.Component.node)
	m.overlayNode.SetVisible(true)
	tt.showing = true

	if tt.FadeInDuration > 0 {
		tt.Component.node.SetAlpha(0)
		m.activeTween = sg.TweenAlpha(tt.Component.node, 1.0, sg.TweenConfig{
			Duration: tt.FadeInDuration,
			Ease:     tt.FadeInEase,
		})
		m.state = tipFadingIn
	} else {
		tt.Component.node.SetAlpha(1.0)
		m.state = tipShowing
	}
}

// startFadeOut begins the fade-out sequence.
func (m *TooltipManager) startFadeOut() {
	tt := m.activeTooltip
	if tt != nil && tt.FadeOutDuration > 0 {
		m.activeTween = sg.TweenAlpha(tt.Component.node, 0, sg.TweenConfig{
			Duration: tt.FadeOutDuration,
			Ease:     tt.FadeOutEase,
		})
		m.state = tipFadingOut
	} else {
		m.finishHide()
	}
}

// finishHide removes the tooltip from the overlay and fires the hide callback.
func (m *TooltipManager) finishHide() {
	trigger := m.activeTrigger
	tt := m.activeTooltip

	if tt != nil && !tt.Component.IsDisposed() {
		m.overlayNode.RemoveChild(tt.Component.node)
		tt.showing = false
	}
	if m.overlayNode != nil {
		m.overlayNode.SetVisible(false)
	}

	if trigger != nil && trigger.onTooltipHide != nil {
		trigger.onTooltipHide()
	}

	m.activeTrigger = nil
	m.activeTooltip = nil
	m.state = tipIdle
}

// hideImmediate hides immediately with no fade. If fireCallback is true,
// fires the hide callback on the active trigger.
func (m *TooltipManager) hideImmediate(fireCallback bool) {
	if !fireCallback {
		if m.activeTrigger != nil {
			m.activeTrigger = nil
		}
	}
	m.counter = 0
	if m.activeTween != nil {
		m.activeTween.Cancel()
		m.activeTween = nil
	}
	if m.activeTooltip != nil && !m.activeTooltip.Component.IsDisposed() {
		m.activeTooltip.Component.node.SetAlpha(1.0)
	}
	m.finishHide()
}

// transitionToIdle resets state without firing any callbacks.
func (m *TooltipManager) transitionToIdle() {
	m.activeTrigger = nil
	m.activeTooltip = nil
	m.state = tipIdle
	m.counter = 0
}

// maybeUpdateFollowPosition updates position when Anchor == TooltipFollowMouse.
func (m *TooltipManager) maybeUpdateFollowPosition() {
	if m.activeTooltip == nil || m.activeTooltip.Anchor != TooltipFollowMouse {
		return
	}
	tt := m.activeTooltip
	cx, cy := engine.CursorPosition()
	x := float64(cx) + tt.OffsetX
	y := float64(cy) + tt.OffsetY
	if tt.ClampToScreen {
		x, y = m.clampToScreen(x, y, tt.Width, tt.Height, tt.ClampMargin)
	}
	tt.Component.X = x
	tt.Component.Y = y
	tt.Component.node.SetPosition(x, y)
}

// computePosition returns the world-space position for the tooltip.
func (m *TooltipManager) computePosition(trigger *Component, tt *Tooltip) (float64, float64) {
	tw := tt.Width
	th := tt.Height

	switch tt.Anchor {
	case TooltipAuto:
		return m.computeAutoPosition(trigger, tw, th, tt)

	case TooltipAbove:
		wx, wy := triggerWorldOrigin(trigger)
		return wx + trigger.Width/2 - tw/2 + tt.OffsetX, wy - th + tt.OffsetY

	case TooltipBelow:
		wx, wy := triggerWorldOrigin(trigger)
		return wx + trigger.Width/2 - tw/2 + tt.OffsetX, wy + trigger.Height + tt.OffsetY

	case TooltipLeft:
		wx, wy := triggerWorldOrigin(trigger)
		return wx - tw + tt.OffsetX, wy + trigger.Height/2 - th/2 + tt.OffsetY

	case TooltipRight:
		wx, wy := triggerWorldOrigin(trigger)
		return wx + trigger.Width + tt.OffsetX, wy + trigger.Height/2 - th/2 + tt.OffsetY

	case TooltipCornerTopLeft:
		mg := tt.ClampMargin
		return mg + tt.OffsetX, mg + tt.OffsetY

	case TooltipCornerTopRight:
		vw, _ := viewportSize()
		return vw - tw + tt.OffsetX, tt.ClampMargin + tt.OffsetY

	case TooltipCornerBottomLeft:
		_, vh := viewportSize()
		return tt.ClampMargin + tt.OffsetX, vh - th + tt.OffsetY

	case TooltipCornerBottomRight:
		vw, vh := viewportSize()
		return vw - tw + tt.OffsetX, vh - th + tt.OffsetY

	case TooltipFollowMouse:
		cx, cy := engine.CursorPosition()
		return float64(cx) + tt.OffsetX, float64(cy) + tt.OffsetY
	}
	return 0, 0
}

// computeAutoPosition picks the direction with the most available room.
func (m *TooltipManager) computeAutoPosition(trigger *Component, tw, th float64, tt *Tooltip) (float64, float64) {
	vw, vh := viewportSize()
	var wx, wy, trigW, trigH float64
	if trigger != nil {
		wx, wy = triggerWorldOrigin(trigger)
		trigW = trigger.Width
		trigH = trigger.Height
	}

	spaces := [4]float64{
		vh - (wy + trigH), // below
		wy,                // above
		vw - (wx + trigW), // right
		wx,                // left
	}
	best := 0
	for i := 1; i < 4; i++ {
		if spaces[i] > spaces[best] {
			best = i
		}
	}

	switch best {
	case 1: // above
		return wx + trigW/2 - tw/2 + tt.OffsetX, wy - th + tt.OffsetY
	case 2: // right
		return wx + trigW + tt.OffsetX, wy + trigH/2 - th/2 + tt.OffsetY
	case 3: // left
		return wx - tw + tt.OffsetX, wy + trigH/2 - th/2 + tt.OffsetY
	default: // 0 = below
		return wx + trigW/2 - tw/2 + tt.OffsetX, wy + trigH + tt.OffsetY
	}
}

// clampToScreen shifts (x,y) so the tooltip of size (w,h) stays within the
// viewport minus margin pixels on each edge.
func (m *TooltipManager) clampToScreen(x, y, w, h, margin float64) (float64, float64) {
	vw, vh := viewportSize()
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
	// Last resort: snap to safe area top-left.
	if x < margin {
		x = margin
	}
	if y < margin {
		y = margin
	}
	return x, y
}

// triggerWorldOrigin returns the top-left world-space corner of the trigger's node.
func triggerWorldOrigin(c *Component) (wx, wy float64) {
	if c == nil {
		return 0, 0
	}
	return c.node.LocalToWorld(0, 0)
}

// viewportSize returns the current window dimensions as the viewport reference.
func viewportSize() (float64, float64) {
	w, h := engine.WindowSize()
	if w <= 0 {
		w = 640
	}
	if h <= 0 {
		h = 480
	}
	return float64(w), float64(h)
}
