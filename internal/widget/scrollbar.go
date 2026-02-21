package widget

import "github.com/devthicket/willowui/internal/sg"

// ScrollBar is a scrollbar with a draggable thumb whose size reflects
// the visible portion relative to total content.
type ScrollBar struct {
	Component
	thumbComp Component // sub-component: draggable thumb

	orientation Orientation
	totalSize   float64
	viewSize    float64
	scrollPos   *Ref[float64]
	watch       WatchHandle
	onChange    func(float64)

	dragging    bool
	grabOffset  float64 // where on the thumb the user grabbed (local coords)
	dragOriginW float64 // scrollbar node's world position at drag start

	thumbHovered bool
	thumbPressed bool
}

// Default scrollbar dimensions.
const (
	DefaultScrollBarWidth  = 16
	DefaultScrollBarLength = 200
	MinScrollThumbSize     = 20
)

// NewScrollBar creates a vertical scrollbar.
func NewScrollBar(name string) *ScrollBar {
	sb := &ScrollBar{
		orientation: Vertical,
		totalSize:   100,
		viewSize:    100,
		scrollPos:   NewRef(0.0),
	}
	initComponent(&sb.Component, name)

	sb.initBackground(name)
	sb.initBorder(name)

	// Thumb sub-component.
	initComponent(&sb.thumbComp, name+"-thumb")
	sb.thumbComp.initBackground(name + "-thumb")
	sb.thumbComp.initBorder(name + "-thumb")
	sb.thumbComp.node.Interactable = true
	sb.node.AddChild(sb.thumbComp.node)

	sb.SetSize(DefaultScrollBarWidth, DefaultScrollBarLength)

	// Forward thumb hover/press to both the thumb state and the scrollbar's
	// own hovered/pressed so the track also stays highlighted when the pointer
	// is directly over the thumb (willow fires leave on the parent when a
	// child captures the pointer).
	sb.thumbComp.node.OnPointerEnter(func(_ sg.PointerContext) {
		sb.thumbHovered = true
		sb.hovered = true
		sb.applyThumbState()
		sb.MarkDrawDirty()
	})
	sb.thumbComp.node.OnPointerLeave(func(_ sg.PointerContext) {
		if sb.dragging {
			return // Keep pressed state during drag.
		}
		sb.thumbHovered = false
		sb.thumbPressed = false
		sb.hovered = false
		sb.pressed = false
		sb.applyThumbState()
		sb.MarkDrawDirty()
	})
	sb.thumbComp.node.OnPointerDown(func(_ sg.PointerContext) {
		if sb.enabled {
			sb.thumbPressed = true
			sb.pressed = true
			sb.applyThumbState()
			sb.MarkDrawDirty()
		}
	})
	sb.thumbComp.node.OnPointerUp(func(_ sg.PointerContext) {
		if sb.dragging {
			return // DragEnd will handle cleanup.
		}
		sb.thumbPressed = false
		sb.pressed = false
		sb.applyThumbState()
		sb.MarkDrawDirty()
	})

	// Drag handling on thumb.
	sb.thumbComp.node.OnDragStart(func(ctx sg.DragContext) {
		if !sb.enabled {
			return
		}
		sb.dragging = true
		if sb.orientation == Vertical {
			sb.grabOffset = ctx.LocalY
			sb.dragOriginW = ctx.GlobalY - ctx.LocalY - sb.thumbComp.node.Y()
		} else {
			sb.grabOffset = ctx.LocalX
			sb.dragOriginW = ctx.GlobalX - ctx.LocalX - sb.thumbComp.node.X()
		}
	})
	sb.thumbComp.node.OnDrag(func(ctx sg.DragContext) {
		if !sb.enabled || !sb.dragging {
			return
		}
		sb.handleDrag(ctx)
	})
	sb.thumbComp.node.OnDragEnd(func(ctx sg.DragContext) {
		sb.dragging = false
		sb.thumbPressed = false
		sb.pressed = false
		sb.applyThumbState()
		sb.MarkDrawDirty()
	})

	// Click on track jumps to position.
	sb.node.OnClick(func(ctx sg.ClickContext) {
		if !sb.enabled {
			return
		}
		sb.handleTrackClick(ctx)
	})

	sb.onVisualStateChange = func() {
		// While dragging, force pressed state — the thumb moves under the
		// cursor causing enter/leave oscillation between thumb and track.
		if sb.dragging {
			sb.thumbHovered = true
			sb.thumbPressed = true
			sb.hovered = true
			sb.pressed = true
			sb.applyThumbState()
		}
	}
	sb.onThemeChange = func() { sb.applyThemeColors() }
	sb.applyThemeColors()
	sb.updateLayout()
	return sb
}

func (sb *ScrollBar) applyThemeColors() {
	group := sb.EffectiveTheme().ScrollBar.Group(sb.Variant())

	// Track.
	sb.applyCornerRadius(group.CornerRadius)
	bg := group.Background.Resolve(StateDefault)
	sb.applyBackground(bg)
	sb.applyBorder(group.Border.Resolve(StateDefault), group.BorderWidth, bg)

	// Thumb sub-component.
	sb.thumbComp.applyCornerRadius(group.ThumbCornerRadius)
	sb.applyThumbState()

	// Recalculate thumb sizing/position since trackInset depends on border width.
	sb.updateLayout()
	sb.MarkDrawDirty()
}

func (sb *ScrollBar) applyThumbState() {
	group := sb.EffectiveTheme().ScrollBar.Group(sb.Variant())
	state := computeState(sb.enabled, false, sb.thumbHovered, sb.thumbPressed)
	thumbBg := group.ThumbBackground.Resolve(state)
	sb.thumbComp.applyBackground(thumbBg)
	sb.thumbComp.applyBorder(group.ThumbBorder.Resolve(state), group.ThumbBorderWidth, thumbBg)
	sb.thumbComp.MarkDrawDirty()
}

// ScrollPos returns the current scroll position.
func (sb *ScrollBar) ScrollPos() float64 {
	return sb.scrollPos.Peek()
}

// SetScrollPos sets the scroll position, clamping to valid range.
func (sb *ScrollBar) SetScrollPos(v float64) {
	maxScroll := sb.maxScroll()
	if v < 0 {
		v = 0
	}
	if v > maxScroll {
		v = maxScroll
	}
	old := sb.scrollPos.Peek()
	sb.scrollPos.Set(v)
	DefaultScheduler.Flush()
	sb.updateLayout()
	if v != old && sb.onChange != nil {
		sb.onChange(v)
	}
}

// SetContentSize sets the total content size and visible viewport size.
func (sb *ScrollBar) SetContentSize(total, view float64) {
	sb.totalSize = total
	sb.viewSize = view
	sb.updateLayout()
	sb.SetScrollPos(sb.scrollPos.Peek()) // re-clamp
}

// SetOrientation sets horizontal or vertical orientation.
func (sb *ScrollBar) SetOrientation(o Orientation) {
	sb.orientation = o
	sb.updateLayout()
}

// SetOnChange sets the callback for scroll position changes.
func (sb *ScrollBar) SetOnChange(fn func(float64)) {
	sb.onChange = fn
}

// SetSize sets the scrollbar dimensions.
func (sb *ScrollBar) SetSize(w, h float64) {
	sb.Width = w
	sb.Height = h
	sb.resizeBackground(w, h)
	sb.resizeBorder(w, h)
	sb.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	sb.updateLayout()
	sb.MarkLayoutDirty()
}

// BindScrollPos binds the scrollbar to a reactive Ref[float64].
func (sb *ScrollBar) BindScrollPos(ref *Ref[float64]) {
	sb.scrollPos = ref
	bindRef(&sb.watch, ref, sb.SetScrollPos)
}

// Dispose stops reactive watches and disposes the component tree.
func (sb *ScrollBar) Dispose() {
	sb.watch.Stop()
	sb.thumbComp.Dispose()
	sb.Component.Dispose()
}

// ThumbNode returns the thumb component's willow node. Used for testing.
func (sb *ScrollBar) ThumbNode() *sg.Node { return sb.thumbComp.node }

// ThumbHeight returns the thumb component's Height. Used for testing.
func (sb *ScrollBar) ThumbHeight() float64 { return sb.thumbComp.Height }

func (sb *ScrollBar) maxScroll() float64 {
	max := sb.totalSize - sb.viewSize
	if max < 0 {
		return 0
	}
	return max
}

func (sb *ScrollBar) trackInset() float64 {
	bw := sb.EffectiveTheme().ScrollBar.Group(sb.Variant()).BorderWidth
	return bw + 1
}

func (sb *ScrollBar) thumbSize() float64 {
	if sb.totalSize <= 0 {
		return 0
	}
	ratio := sb.viewSize / sb.totalSize
	if ratio > 1 {
		ratio = 1
	}
	inset := sb.trackInset()
	if sb.orientation == Vertical {
		h := ratio * (sb.Height - 2*inset)
		if h < MinScrollThumbSize {
			h = MinScrollThumbSize
		}
		return h
	}
	w := ratio * (sb.Width - 2*inset)
	if w < MinScrollThumbSize {
		w = MinScrollThumbSize
	}
	return w
}

func (sb *ScrollBar) updateLayout() {
	if sb.totalSize <= 0 {
		return
	}

	ts := sb.thumbSize()
	inset := sb.trackInset()

	if sb.orientation == Vertical {
		tw := sb.Width - 2*inset
		if tw < 0 {
			tw = 0
		}
		sb.thumbComp.Width = tw
		sb.thumbComp.Height = ts
		sb.thumbComp.resizeBackground(tw, ts)
		sb.thumbComp.resizeBorder(tw, ts)
		sb.thumbComp.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: tw, Height: ts}

		// Position thumb inside track border.
		maxScroll := sb.maxScroll()
		trackRange := sb.Height - 2*inset - ts
		var thumbY float64
		if maxScroll > 0 && trackRange > 0 {
			thumbY = (sb.scrollPos.Peek() / maxScroll) * trackRange
		}
		sb.thumbComp.node.SetPosition(inset, inset+thumbY)
	} else {
		th := sb.Height - 2*inset
		if th < 0 {
			th = 0
		}
		sb.thumbComp.Width = ts
		sb.thumbComp.Height = th
		sb.thumbComp.resizeBackground(ts, th)
		sb.thumbComp.resizeBorder(ts, th)
		sb.thumbComp.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: ts, Height: th}

		maxScroll := sb.maxScroll()
		trackRange := sb.Width - 2*inset - ts
		var thumbX float64
		if maxScroll > 0 && trackRange > 0 {
			thumbX = (sb.scrollPos.Peek() / maxScroll) * trackRange
		}
		sb.thumbComp.node.SetPosition(inset+thumbX, inset)
	}
}

func (sb *ScrollBar) handleDrag(ctx sg.DragContext) {
	maxScroll := sb.maxScroll()
	if maxScroll <= 0 {
		return
	}

	ts := sb.thumbSize()
	inset := sb.trackInset()

	if sb.orientation == Vertical {
		trackRange := sb.Height - 2*inset - ts
		if trackRange <= 0 {
			return
		}
		posInTrack := ctx.GlobalY - sb.dragOriginW - sb.grabOffset
		frac := posInTrack / trackRange
		if frac < 0 {
			frac = 0
		}
		if frac > 1 {
			frac = 1
		}
		sb.SetScrollPos(frac * maxScroll)
	} else {
		trackRange := sb.Width - 2*inset - ts
		if trackRange <= 0 {
			return
		}
		posInTrack := ctx.GlobalX - sb.dragOriginW - sb.grabOffset
		frac := posInTrack / trackRange
		if frac < 0 {
			frac = 0
		}
		if frac > 1 {
			frac = 1
		}
		sb.SetScrollPos(frac * maxScroll)
	}
}

func (sb *ScrollBar) handleTrackClick(ctx sg.ClickContext) {
	maxScroll := sb.maxScroll()
	if maxScroll <= 0 {
		return
	}

	ts := sb.thumbSize()
	inset := sb.trackInset()

	if sb.orientation == Vertical {
		trackRange := sb.Height - 2*inset - ts
		if trackRange <= 0 {
			return
		}
		posInTrack := ctx.LocalY - inset - ts/2
		frac := posInTrack / trackRange
		if frac < 0 {
			frac = 0
		}
		if frac > 1 {
			frac = 1
		}
		sb.SetScrollPos(frac * maxScroll)
	} else {
		trackRange := sb.Width - 2*inset - ts
		if trackRange <= 0 {
			return
		}
		posInTrack := ctx.LocalX - inset - ts/2
		frac := posInTrack / trackRange
		if frac < 0 {
			frac = 0
		}
		if frac > 1 {
			frac = 1
		}
		sb.SetScrollPos(frac * maxScroll)
	}
}
