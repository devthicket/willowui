package widget

import (
	"time"

	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
	"github.com/devthicket/willowui/internal/theme"
)

// toastOverlayZIndex sits above menu popups (500_000) but below tooltips (1_000_000).
const toastOverlayZIndex = 750_000

// ---------------------------------------------------------------------------
// ToastAnchor
// ---------------------------------------------------------------------------

// ToastAnchor specifies which screen corner toasts stack at.
type ToastAnchor int

const (
	ToastBottomRight ToastAnchor = iota // default
	ToastBottomLeft
	ToastTopRight
	ToastTopLeft
)

// ---------------------------------------------------------------------------
// ToastOption
// ---------------------------------------------------------------------------

// ToastOption is a functional option for configuring a toast.
type ToastOption func(*toastConfig)

// toastConfig holds per-toast configuration.
type toastConfig struct {
	duration       time.Duration
	dismissOnClick bool
	showProgress   bool
	onDismiss      func()
}

func defaultToastConfig() toastConfig {
	return toastConfig{
		duration:       3 * time.Second,
		dismissOnClick: true,
	}
}

// WithDuration sets the auto-dismiss duration.
func WithDuration(d time.Duration) ToastOption {
	return func(c *toastConfig) { c.duration = d }
}

// WithDismissOnClick enables or disables click-to-dismiss (default true).
func WithDismissOnClick(v bool) ToastOption {
	return func(c *toastConfig) { c.dismissOnClick = v }
}

// WithProgress shows a shrinking remaining-time bar at the bottom of the toast.
func WithProgress(v bool) ToastOption {
	return func(c *toastConfig) { c.showProgress = v }
}

// WithOnDismiss sets a callback invoked when the toast is dismissed.
func WithOnDismiss(fn func()) ToastOption {
	return func(c *toastConfig) { c.onDismiss = fn }
}

// ---------------------------------------------------------------------------
// Internal entry
// ---------------------------------------------------------------------------

type toastState int

const (
	toastEntering toastState = iota
	toastActive
	toastDismissing
	toastDone
)

type toastEntry struct {
	node         *sg.Node
	progressNode *sg.Node // nil when showProgress is false
	progressMaxW float64  // cached max width of the progress bar (constant after build)
	width        float64
	height       float64
	elapsed      float64 // seconds accumulated during toastActive
	duration     float64 // seconds until auto-dismiss (0 = already dismissed)
	cfg          toastConfig
	state        toastState
	tween        *sg.TweenGroup
	targetX      float64
	targetY      float64
}

// ---------------------------------------------------------------------------
// ToastManager
// ---------------------------------------------------------------------------

// ToastManager manages a stack of transient toast notifications.
// Use DefaultToastManager; do not construct your own.
type ToastManager struct {
	font     *sg.FontFamily
	fontSize float64

	anchor   ToastAnchor
	maxStack int
	marginX  float64
	marginY  float64

	scene       *sg.Scene
	tickerNode  *sg.Node
	overlayNode *sg.Node

	entries []*toastEntry
}

// DefaultToastManager is the package-level singleton used by ShowToast.
var DefaultToastManager = &ToastManager{
	maxStack: 4,
	marginX:  16,
	marginY:  16,
	fontSize: 14,
}

// ---------------------------------------------------------------------------
// Package-level helpers
// ---------------------------------------------------------------------------

// ShowToast shows a toast via DefaultToastManager with the given variant.
// The variant selects the semantic color (Info, Success, Warning, Danger).
// Pass theme.Primary or omit for the default appearance.
func ShowToast(message string, variant Variant, opts ...ToastOption) {
	DefaultToastManager.Show(message, variant, opts...)
}

// ---------------------------------------------------------------------------
// Manager public API
// ---------------------------------------------------------------------------

// SetFont configures the font source used to render toast messages.
// Call this once at startup before showing toasts. Without a font, messages
// are not rendered but the toast still appears with icon and progress bar.
func (m *ToastManager) SetFont(source *sg.FontFamily, size float64) {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	m.font = font
	m.fontSize = size
}

// SetAnchor sets the screen corner where toasts appear.
func (m *ToastManager) SetAnchor(corner ToastAnchor) {
	m.anchor = corner
}

// SetMaxStack sets the maximum number of toasts visible simultaneously (default 4).
// When the stack is full the oldest toast is dropped immediately.
func (m *ToastManager) SetMaxStack(n int) {
	if n < 1 {
		n = 1
	}
	m.maxStack = n
}

// SetMargin sets the pixel gap between the toast stack and the screen edge (default 16×16).
func (m *ToastManager) SetMargin(x, y float64) {
	m.marginX = x
	m.marginY = y
}

// Show displays a toast with the given message and variant.
func (m *ToastManager) Show(message string, variant Variant, opts ...ToastOption) {
	cfg := defaultToastConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if sc := currentScene(); sc != nil {
		m.ensureNodes(sc)
	}
	if m.overlayNode == nil {
		return
	}

	// Drop the oldest entry if we are at capacity.
	if m.countActive() >= m.maxStack {
		// activeEntries is newest-first; oldest is the last element of m.entries
		// that is still entering/active.
		for i := 0; i < len(m.entries); i++ {
			if e := m.entries[i]; e.state == toastEntering || e.state == toastActive {
				m.finishDismiss(e)
				break
			}
		}
	}

	// Build the new entry.
	entry := m.buildEntry(message, variant, &cfg)
	entry.state = toastEntering
	m.entries = append(m.entries, entry)

	// Compute target positions for all entries (including the new one).
	// This sets entry.targetX / entry.targetY.
	m.restackEntries()

	// Place new entry off-screen at the slide origin, invisible.
	slideX, slideY := m.slideOffset()
	entry.node.SetPosition(entry.targetX+slideX, entry.targetY+slideY)
	entry.node.SetAlpha(0)
	m.overlayNode.AddChild(entry.node)

	// Animate entry: fade in + slide to target.
	g := m.toastGroup()
	dur := float32(g.AnimDuration)
	if dur <= 0 {
		dur = 0.2
	}
	entry.tween = sg.TweenAlpha(entry.node, 1.0, sg.TweenConfig{Duration: dur})
	sg.TweenPosition(entry.node, entry.targetX, entry.targetY, sg.TweenConfig{Duration: dur})
}

// DismissAll removes all active toasts immediately.
func (m *ToastManager) DismissAll() {
	for _, e := range m.entries {
		if e.state == toastEntering || e.state == toastActive || e.state == toastDismissing {
			if e.tween != nil {
				e.tween.Cancel()
				e.tween = nil
			}
			e.state = toastDone
			if m.overlayNode != nil && e.node != nil && e.node.Parent != nil {
				m.overlayNode.RemoveChild(e.node)
			}
			if e.cfg.onDismiss != nil {
				e.cfg.onDismiss()
			}
		}
	}
	m.entries = m.entries[:0]
}

// ---------------------------------------------------------------------------
// Manager internal
// ---------------------------------------------------------------------------

// setScene is called from widget.SetScene when the scene changes.
func (m *ToastManager) setScene(s *sg.Scene) {
	// Remove all entries from the old overlay immediately.
	for _, e := range m.entries {
		if e.tween != nil {
			e.tween.Cancel()
			e.tween = nil
		}
		if m.overlayNode != nil && e.node != nil && e.node.Parent != nil {
			m.overlayNode.RemoveChild(e.node)
		}
	}
	m.entries = m.entries[:0]

	m.scene = s
	if s == nil || s.Root == nil {
		return
	}
	m.ensureNodes(s)
}

func (m *ToastManager) ensureNodes(s *sg.Scene) {
	if s == nil || s.Root == nil {
		return
	}
	if m.tickerNode == nil {
		m.tickerNode = sg.NewContainer("toast-ticker")
		m.tickerNode.Interactable = false
		m.tickerNode.SetZIndex(toastOverlayZIndex)
		m.tickerNode.OnUpdate = func(dt float64) {
			DefaultToastManager.tick(dt)
		}
	}
	if m.overlayNode == nil {
		m.overlayNode = sg.NewContainer("toast-overlay")
		m.overlayNode.Interactable = true
		m.overlayNode.SetZIndex(toastOverlayZIndex)
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
}

// tick is called once per frame via the ticker node's OnUpdate.
func (m *ToastManager) tick(dt float64) {
	if len(m.entries) == 0 {
		return
	}
	changed := false
	for _, e := range m.entries {
		switch e.state {
		case toastEntering:
			if e.tween == nil || e.tween.Done {
				e.tween = nil
				e.state = toastActive
			}

		case toastActive:
			e.elapsed += dt
			if e.progressNode != nil {
				fraction := 1.0 - (e.elapsed / e.duration)
				if fraction < 0 {
					fraction = 0
				}
				e.progressNode.SetScale(e.progressMaxW*fraction, toastProgressBarH)
				e.progressNode.Invalidate()
			}
			if e.elapsed >= e.duration {
				m.startDismiss(e)
			}

		case toastDismissing:
			if e.tween == nil || e.tween.Done {
				e.tween = nil
				e.state = toastDone
				if m.overlayNode != nil && e.node != nil && e.node.Parent != nil {
					m.overlayNode.RemoveChild(e.node)
				}
				if e.cfg.onDismiss != nil {
					e.cfg.onDismiss()
				}
				changed = true
			}
		}
	}

	if changed {
		n := 0
		for _, e := range m.entries {
			if e.state != toastDone {
				m.entries[n] = e
				n++
			}
		}
		m.entries = m.entries[:n]
		m.restackEntries()
	}
}

// dismissEntry triggers the dismiss animation for e. Called by click-to-dismiss.
func (m *ToastManager) dismissEntry(e *toastEntry) {
	if e.state == toastDismissing || e.state == toastDone {
		return
	}
	m.startDismiss(e)
}

// startDismiss begins the fade-out animation.
func (m *ToastManager) startDismiss(e *toastEntry) {
	if e.tween != nil {
		e.tween.Cancel()
		e.tween = nil
	}
	g := m.toastGroup()
	dur := float32(g.AnimDuration)
	if dur <= 0 {
		dur = 0.2
	}
	slideX, slideY := m.slideOffset()
	e.tween = sg.TweenAlpha(e.node, 0, sg.TweenConfig{Duration: dur})
	sg.TweenPosition(e.node, e.targetX+slideX, e.targetY+slideY, sg.TweenConfig{Duration: dur})
	e.state = toastDismissing
}

// finishDismiss removes the entry instantly without animation.
func (m *ToastManager) finishDismiss(e *toastEntry) {
	if e.state == toastDone {
		return
	}
	if e.tween != nil {
		e.tween.Cancel()
		e.tween = nil
	}
	e.state = toastDone
	if m.overlayNode != nil && e.node != nil && e.node.Parent != nil {
		m.overlayNode.RemoveChild(e.node)
	}
	// Compact the entries slice.
	n := 0
	for _, en := range m.entries {
		if en.state != toastDone {
			m.entries[n] = en
			n++
		}
	}
	m.entries = m.entries[:n]
}

// activeEntries returns non-done entries in newest-first order.
func (m *ToastManager) activeEntries() []*toastEntry {
	var out []*toastEntry
	for i := len(m.entries) - 1; i >= 0; i-- {
		e := m.entries[i]
		if e.state == toastEntering || e.state == toastActive {
			out = append(out, e)
		}
	}
	return out
}

// countActive returns the number of entering/active entries without allocating.
func (m *ToastManager) countActive() int {
	n := 0
	for _, e := range m.entries {
		if e.state == toastEntering || e.state == toastActive {
			n++
		}
	}
	return n
}

// restackEntries computes and applies target positions for all active entries.
// Iterates m.entries in reverse (newest first = highest index) to avoid
// allocating a temporary slice.
func (m *ToastManager) restackEntries() {
	vw, vh := viewportSize()
	g := m.toastGroup()
	spacing := g.ItemSpacing
	if spacing <= 0 {
		spacing = 6
	}
	dur := float32(g.AnimDuration)
	if dur <= 0 {
		dur = 0.2
	}

	cumH := 0.0
	for i := len(m.entries) - 1; i >= 0; i-- {
		e := m.entries[i]
		if e.state != toastEntering && e.state != toastActive {
			continue
		}
		x, y := m.computePosition(e, cumH, vw, vh)
		e.targetX = x
		e.targetY = y
		if e.state == toastActive || e.state == toastEntering {
			sg.TweenPosition(e.node, x, y, sg.TweenConfig{Duration: dur})
		}
		cumH += e.height + spacing
	}
}

// computePosition returns the target (x, y) for a toast given its cumulative stack offset.
func (m *ToastManager) computePosition(e *toastEntry, cumH, vw, vh float64) (float64, float64) {
	switch m.anchor {
	case ToastBottomRight:
		return vw - e.width - m.marginX, vh - e.height - m.marginY - cumH
	case ToastBottomLeft:
		return m.marginX, vh - e.height - m.marginY - cumH
	case ToastTopRight:
		return vw - e.width - m.marginX, m.marginY + cumH
	case ToastTopLeft:
		return m.marginX, m.marginY + cumH
	}
	return vw - e.width - m.marginX, vh - e.height - m.marginY - cumH
}

// slideOffset returns the (dx, dy) applied when a toast enters or exits.
// The toast slides in from outside the anchor edge.
func (m *ToastManager) slideOffset() (float64, float64) {
	const amt = 24.0
	switch m.anchor {
	case ToastBottomRight, ToastTopRight:
		return amt, 0
	case ToastBottomLeft, ToastTopLeft:
		return -amt, 0
	}
	return amt, 0
}

func (m *ToastManager) toastGroup(variant ...Variant) *theme.ToastGroup {
	v := theme.Primary
	if len(variant) > 0 {
		v = variant[0]
	}
	return getDefaultTheme().Toast.Group(v)
}

func (m *ToastManager) resolvedPadding(g *theme.ToastGroup) render.Insets {
	if !g.Padding.IsAuto() {
		return g.Padding
	}
	return render.Insets{Top: 10, Right: 14, Bottom: 10, Left: 14}
}

// ---------------------------------------------------------------------------
// Toast node construction
// ---------------------------------------------------------------------------

const (
	toastIconSize     = 12.0
	toastIconPad      = 8.0
	toastProgressBarH = 3.0
	toastProgressPad  = 4.0
)

func (m *ToastManager) buildEntry(message string, variant Variant, cfg *toastConfig) *toastEntry {
	g := m.toastGroup(variant)
	pad := m.resolvedPadding(g)

	// Background from variant group.
	bgColor := g.Background.Resolve(core.StateDefault).Color

	// Border.
	borderColor := g.BorderColor.Resolve(core.StateDefault)

	// Text measurement.
	textLineH := m.fontSize * 1.2
	if m.font != nil {
		textLineH = displayLineHeight(m.font, m.fontSize)
	}
	textW := 0.0
	if m.font != nil && message != "" {
		textW, _ = measureDisplay(m.font, message, m.fontSize)
	}

	// Width.
	minW, maxW := g.MinWidth, g.MaxWidth
	if minW <= 0 {
		minW = 200
	}
	if maxW <= 0 {
		maxW = 360
	}
	contentW := toastIconSize + toastIconPad + textW
	w := contentW + pad.Left + pad.Right
	if w < minW {
		w = minW
	}
	if w > maxW {
		w = maxW
	}

	// Height.
	contentH := textLineH
	if toastIconSize > contentH {
		contentH = toastIconSize
	}
	progressH := 0.0
	if cfg.showProgress {
		progressH = toastProgressBarH + toastProgressPad
	}
	h := pad.Top + contentH + pad.Bottom + progressH

	// Root container.
	root := sg.NewContainer("toast-root")
	root.Interactable = cfg.dismissOnClick
	root.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}

	// Background sprite.
	bgNode := sg.NewSprite("toast-bg", sg.TextureRegion{})
	bgNode.SetScale(w, h)
	bgNode.SetColor(bgColor)
	root.AddChild(bgNode)

	// Border (4 edge sprites at z-index 1).
	bw := g.BorderWidth
	if bw > 0 {
		bc := borderColor
		top := sg.NewSprite("toast-border-t", sg.TextureRegion{})
		top.SetScale(w, bw)
		top.SetColor(bc)
		top.SetZIndex(1)

		bot := sg.NewSprite("toast-border-b", sg.TextureRegion{})
		bot.SetScale(w, bw)
		bot.SetPosition(0, h-bw)
		bot.SetColor(bc)
		bot.SetZIndex(1)

		lft := sg.NewSprite("toast-border-l", sg.TextureRegion{})
		lft.SetScale(bw, h-bw*2)
		lft.SetPosition(0, bw)
		lft.SetColor(bc)
		lft.SetZIndex(1)

		rgt := sg.NewSprite("toast-border-r", sg.TextureRegion{})
		rgt.SetScale(bw, h-bw*2)
		rgt.SetPosition(w-bw, bw)
		rgt.SetColor(bc)
		rgt.SetZIndex(1)

		root.AddChild(top)
		root.AddChild(bot)
		root.AddChild(lft)
		root.AddChild(rgt)
	}

	// Icon sprite (small colored square at z-index 2).
	iconColor := g.IconColor.Resolve(core.StateDefault)
	if iconColor.A() == 0 {
		iconColor = g.TextColor.Resolve(core.StateDefault)
	}
	iconNode := sg.NewSprite("toast-icon", sg.TextureRegion{})
	iconNode.SetScale(toastIconSize, toastIconSize)
	iconNode.SetPosition(pad.Left, pad.Top+(contentH-toastIconSize)/2)
	iconNode.SetColor(iconColor)
	iconNode.SetZIndex(2)
	root.AddChild(iconNode)

	// Text node (z-index 2). Use TextBlock.FontSize only — no SetScale.
	if m.font != nil {
		textNode := sg.NewText("toast-text", message, m.font)
		textNode.TextBlock.FontSize = m.fontSize
		textNode.TextBlock.Color = g.TextColor.Resolve(core.StateDefault)
		textNode.SetPosition(pad.Left+toastIconSize+toastIconPad, pad.Top+(contentH-textLineH)/2)
		textNode.SetZIndex(2)
		root.AddChild(textNode)
	}

	// Progress bar sprite (z-index 2).
	var progressNode *sg.Node
	progressMaxW := w - pad.Left - pad.Right
	if cfg.showProgress {
		progColor := g.ProgressBarColor.Resolve(core.StateDefault)
		progressNode = sg.NewSprite("toast-progress", sg.TextureRegion{})
		progressNode.SetScale(progressMaxW, toastProgressBarH)
		progressNode.SetPosition(pad.Left, h-pad.Bottom-progressH+toastProgressPad)
		progressNode.SetColor(progColor)
		progressNode.SetZIndex(2)
		root.AddChild(progressNode)
	}

	entry := &toastEntry{
		node:         root,
		progressNode: progressNode,
		progressMaxW: progressMaxW,
		width:        w,
		height:       h,
		duration:     cfg.duration.Seconds(),
		cfg:          *cfg,
	}

	if cfg.dismissOnClick {
		e := entry
		root.OnPointerDown(func(_ sg.PointerContext) {
			DefaultToastManager.dismissEntry(e)
		})
	}

	return entry
}

