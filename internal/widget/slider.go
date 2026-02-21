package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// Slider is a draggable range control for selecting a numeric value.
type Slider struct {
	Component
	thumbComp Component // sub-component: draggable thumb

	orientation Orientation
	min, max    float64
	step        float64
	value       *Ref[float64]
	watch       WatchHandle
	onChange    func(float64)

	// Internal state for dragging.
	dragging    bool
	grabOffset  float64 // where on the thumb the user grabbed (local coords)
	dragOriginW float64 // slider node's world position at drag start
}

// Default slider dimensions.
const (
	DefaultSliderWidth     = 200
	DefaultSliderHeight    = 20
	DefaultSliderThumbSize = 16
)

// NewSlider creates a horizontal slider with range [0, 1].
func NewSlider(name string) *Slider {
	s := &Slider{
		min:   0,
		max:   1,
		value: NewRef(0.0),
	}
	initComponent(&s.Component, name)

	s.initBackground(name)
	s.initBorder(name)

	// Thumb sub-component.
	initComponent(&s.thumbComp, name+"-thumb")
	s.thumbComp.initBackground(name + "-thumb")
	s.thumbComp.initBorder(name + "-thumb")
	s.node.AddChild(s.thumbComp.node)

	s.SetSize(DefaultSliderWidth, DefaultSliderHeight)

	// Drag handling on thumb.
	s.thumbComp.node.Interactable = true
	s.thumbComp.node.OnDragStart(func(ctx sg.DragContext) {
		if !s.enabled {
			return
		}
		s.dragging = true
		// Capture fixed reference values at drag start.
		if s.orientation == Vertical {
			s.grabOffset = ctx.LocalY
			s.dragOriginW = ctx.GlobalY - ctx.LocalY - s.thumbComp.node.Y()
		} else {
			s.grabOffset = ctx.LocalX
			s.dragOriginW = ctx.GlobalX - ctx.LocalX - s.thumbComp.node.X()
		}
	})
	s.thumbComp.node.OnDrag(func(ctx sg.DragContext) {
		if !s.enabled || !s.dragging {
			return
		}
		s.handleDrag(ctx)
	})
	s.thumbComp.node.OnDragEnd(func(ctx sg.DragContext) {
		s.dragging = false
		s.pressed = false
		s.MarkDrawDirty()
		if s.onVisualStateChange != nil {
			s.onVisualStateChange()
		}
	})

	// Forward thumb hover/press to slider so the thumb stays highlighted
	// when the pointer is directly over it (willow fires leave on the parent
	// when a child captures the pointer).
	s.thumbComp.node.OnPointerEnter(func(_ sg.PointerContext) {
		s.hovered = true
		s.MarkDrawDirty()
		if s.onVisualStateChange != nil {
			s.onVisualStateChange()
		}
	})
	s.thumbComp.node.OnPointerLeave(func(_ sg.PointerContext) {
		if s.dragging {
			return // Keep pressed state during drag.
		}
		s.hovered = false
		s.pressed = false
		s.MarkDrawDirty()
		if s.onVisualStateChange != nil {
			s.onVisualStateChange()
		}
	})
	s.thumbComp.node.OnPointerDown(func(_ sg.PointerContext) {
		if s.enabled {
			s.pressed = true
			s.MarkDrawDirty()
			if s.onVisualStateChange != nil {
				s.onVisualStateChange()
			}
		}
	})
	s.thumbComp.node.OnPointerUp(func(_ sg.PointerContext) {
		if s.dragging {
			return // DragEnd will handle cleanup.
		}
		s.pressed = false
		s.MarkDrawDirty()
		if s.onVisualStateChange != nil {
			s.onVisualStateChange()
		}
	})

	// Click on track jumps to position.
	s.node.OnClick(func(ctx sg.ClickContext) {
		if !s.enabled {
			return
		}
		s.handleTrackClick(ctx)
	})

	s.onVisualStateChange = func() {
		// While dragging, force pressed state regardless of which node
		// the pointer is over — the thumb moves under the cursor causing
		// enter/leave oscillation between thumb and track.
		if s.dragging {
			s.hovered = true
			s.pressed = true
		}
		s.UpdateVisuals()
	}
	s.onThemeChange = func() { s.UpdateVisuals() }

	// Focus: sliders participate in tab and spatial nav. Value-axis arrows
	// (Left/Right for horizontal, Up/Down for vertical) adjust the value;
	// cross-axis arrows navigate to the prev/next widget like Shift+Tab/Tab.
	s.enableFocusNavigation()
	s.InterceptArrows = true
	s.ConsumeHandledKeys = false // slider's OnUpdate handles the key; only block spatial nav

	s.onFocusChange = func(focused bool) { s.UpdateVisuals() }
	// Intercept the value-axis arrows so spatial nav doesn't steal them.
	s.SetHandleKey(func(key engine.Key) bool {
		if s.orientation == Vertical {
			return key == engine.KeyUp || key == engine.KeyDown
		}
		return key == engine.KeyLeft || key == engine.KeyRight
	})

	// Cross-axis arrows navigate like Shift+Tab / Tab.
	DefaultFocusManager.BindScoped(&s.Component, Key(engine.KeyUp, ModNone), func() bool {
		if s.orientation == Vertical {
			return false // handled by OnUpdate
		}
		DefaultFocusManager.TabPrev()
		return true
	})
	DefaultFocusManager.BindScoped(&s.Component, Key(engine.KeyDown, ModNone), func() bool {
		if s.orientation == Vertical {
			return false // handled by OnUpdate
		}
		DefaultFocusManager.TabNext()
		return true
	})
	DefaultFocusManager.BindScoped(&s.Component, Key(engine.KeyLeft, ModNone), func() bool {
		if s.orientation != Vertical {
			return false // handled by OnUpdate
		}
		DefaultFocusManager.TabPrev()
		return true
	})
	DefaultFocusManager.BindScoped(&s.Component, Key(engine.KeyRight, ModNone), func() bool {
		if s.orientation != Vertical {
			return false // handled by OnUpdate
		}
		DefaultFocusManager.TabNext()
		return true
	})

	// Keyboard activation: value-axis arrows adjust value.
	s.node.OnUpdate = func(_ float64) {
		if !s.focused || !s.enabled {
			return
		}
		im := DefaultInputManager
		step := s.step
		if step == 0 {
			step = (s.max - s.min) / 20
		}
		if s.orientation == Vertical {
			if im.IsKeyJustAvailable(engine.KeyUp) {
				s.SetValue(s.value.Peek() + step)
				im.Consume(engine.KeyUp)
			}
			if im.IsKeyJustAvailable(engine.KeyDown) {
				s.SetValue(s.value.Peek() - step)
				im.Consume(engine.KeyDown)
			}
		} else {
			if im.IsKeyJustAvailable(engine.KeyRight) {
				s.SetValue(s.value.Peek() + step)
				im.Consume(engine.KeyRight)
			}
			if im.IsKeyJustAvailable(engine.KeyLeft) {
				s.SetValue(s.value.Peek() - step)
				im.Consume(engine.KeyLeft)
			}
		}
	}

	s.updateLayout()
	s.UpdateVisuals()
	return s
}

// Value returns the current slider value.
func (s *Slider) Value() float64 {
	return s.value.Peek()
}

// SetValue sets the slider value, clamping and snapping as needed.
func (s *Slider) SetValue(v float64) {
	v = s.clampAndSnap(v)
	old := s.value.Peek()
	s.value.Set(v)
	DefaultScheduler.Flush()
	s.updateLayout()
	if v != old && s.onChange != nil {
		s.onChange(v)
	}
}

// SetRange sets the min and max values.
func (s *Slider) SetRange(min, max float64) {
	s.min = min
	s.max = max
	s.SetValue(s.value.Peek()) // re-clamp
}

// SetStep sets the step increment (0 = continuous).
func (s *Slider) SetStep(step float64) {
	s.step = step
}

// SetOrientation sets horizontal or vertical orientation.
func (s *Slider) SetOrientation(o Orientation) {
	s.orientation = o
	s.updateLayout()
}

// SetOnChange sets the callback for value changes.
func (s *Slider) SetOnChange(fn func(float64)) {
	s.onChange = fn
}

// SetSize sets the slider dimensions.
func (s *Slider) SetSize(w, h float64) {
	s.Width = w
	s.Height = h
	s.resizeBackground(w, h)
	s.resizeBorder(w, h)
	s.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	s.updateLayout()
	s.MarkLayoutDirty()
}

// BindValue binds the slider to a reactive Ref[float64].
func (s *Slider) BindValue(ref *Ref[float64]) {
	s.value = ref
	bindRef(&s.watch, ref, s.SetValue)
}

// SetEnabled overrides Component.SetEnabled to also update visuals.
func (s *Slider) SetEnabled(v bool) {
	s.Component.SetEnabled(v)
	s.UpdateVisuals()
}

// UpdateVisuals applies theme colors based on current state.
func (s *Slider) UpdateVisuals() {
	s.state = computeState(s.enabled, s.focused, s.hovered, s.pressed)
	group := s.EffectiveTheme().Slider.Group(s.Variant())

	// Track. -1 means full pill radius.
	trackR := group.CornerRadius
	if trackR < 0 {
		trackR = s.Height / 2
	}
	s.applyCornerRadius(trackR)
	bg := group.Background.Resolve(s.state)
	s.applyBackground(bg)
	s.applyBorder(group.Border.Resolve(s.state), group.BorderWidth, bg)

	// Recompute thumb dimensions for the current theme's border width.
	s.updateLayout()

	// Thumb. -1 means maximum rounding: use half the shorter dimension so a
	// non-square thumb becomes a pill rather than a circle.
	_, cross := s.effectiveThumbDims()
	thumbR := group.ThumbCornerRadius
	if thumbR < 0 {
		thumbR = cross / 2
	}
	s.thumbComp.applyCornerRadius(thumbR)
	thumbBg := group.ThumbBackground.Resolve(s.state)
	s.thumbComp.applyBackground(thumbBg)
	s.thumbComp.applyBorder(group.ThumbBorder.Resolve(s.state), group.ThumbBorderWidth, thumbBg)
	s.applyFocusRing(group.FocusColor.Resolve(s.state), group.FocusRingWidth)

	s.MarkDrawDirty()
}

// Dispose stops reactive watches and disposes the component tree.
func (s *Slider) Dispose() {
	s.watch.Stop()
	s.thumbComp.Dispose()
	s.Component.Dispose()
}

// ThumbNode returns the thumb component's willow node. Used for testing.
func (s *Slider) ThumbNode() *sg.Node { return s.thumbComp.node }

// GetOrientation returns the current orientation. Used for testing.
func (s *Slider) GetOrientation() Orientation { return s.orientation }

// effectiveThumbSize returns the cross-axis thumb size (height for horizontal,
// width for vertical). Falls back to DefaultSliderThumbSize when unset.
func (s *Slider) effectiveThumbSize() float64 {
	if ts := s.EffectiveTheme().Slider.Group(s.Variant()).ThumbSize; ts > 0 {
		return ts
	}
	return DefaultSliderThumbSize
}

// effectiveThumbDims returns (alongTrack, crossTrack) thumb dimensions.
// alongTrack is the dimension in the direction of travel (width for horizontal,
// height for vertical). crossTrack is perpendicular (height for horizontal,
// width for vertical).
func (s *Slider) effectiveThumbDims() (alongTrack, crossTrack float64) {
	group := s.EffectiveTheme().Slider.Group(s.Variant())
	crossTrack = group.ThumbSize
	if crossTrack <= 0 {
		crossTrack = DefaultSliderThumbSize
	}
	alongTrack = group.ThumbLength
	if alongTrack <= 0 {
		alongTrack = crossTrack
	}
	return alongTrack, crossTrack
}

func (s *Slider) clampAndSnap(v float64) float64 {
	if v < s.min {
		v = s.min
	}
	if v > s.max {
		v = s.max
	}
	if s.step > 0 {
		v = math.Round((v-s.min)/s.step)*s.step + s.min
		if v > s.max {
			v = s.max
		}
	}
	return v
}

func (s *Slider) fraction() float64 {
	if s.max <= s.min {
		return 0
	}
	return (s.value.Peek() - s.min) / (s.max - s.min)
}

func (s *Slider) updateLayout() {
	frac := s.fraction()
	along, cross := s.effectiveThumbDims()
	bw := s.EffectiveTheme().Slider.Group(s.Variant()).BorderWidth
	inset := bw + 1

	if s.orientation == Vertical {
		// along = height (along track), cross = width (perpendicular)
		tw := cross
		th := along
		cx := (s.Width - tw) / 2
		innerH := s.Height - 2*inset
		if innerH < 0 {
			innerH = 0
		}
		s.thumbComp.Width = tw
		s.thumbComp.Height = th
		s.thumbComp.resizeBackground(tw, th)
		s.thumbComp.resizeBorder(tw, th)
		thumbRange := innerH - th
		if thumbRange < 0 {
			thumbRange = 0
		}
		s.thumbComp.node.SetPosition(cx, inset+(1-frac)*thumbRange)
		s.thumbComp.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: tw, Height: th}
		// Expand track hit area to at least the thumb cross-size so thin tracks
		// are easy to click.
		hitW := math.Max(s.Width, cross)
		hitX := (s.Width - hitW) / 2
		s.node.HitShape = sg.HitRect{X: hitX, Y: 0, Width: hitW, Height: s.Height}
	} else {
		// along = width (along track), cross = height (perpendicular)
		tw := along
		th := cross
		cy := (s.Height - th) / 2
		innerW := s.Width - 2*inset
		if innerW < 0 {
			innerW = 0
		}
		s.thumbComp.Width = tw
		s.thumbComp.Height = th
		s.thumbComp.resizeBackground(tw, th)
		s.thumbComp.resizeBorder(tw, th)
		thumbRange := innerW - tw
		if thumbRange < 0 {
			thumbRange = 0
		}
		s.thumbComp.node.SetPosition(inset+frac*thumbRange, cy)
		s.thumbComp.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: tw, Height: th}
		// Expand track hit area to at least the thumb cross-size so thin tracks
		// are easy to click.
		hitH := math.Max(s.Height, cross)
		hitY := (s.Height - hitH) / 2
		s.node.HitShape = sg.HitRect{X: 0, Y: hitY, Width: s.Width, Height: hitH}
	}
}

func (s *Slider) handleDrag(ctx sg.DragContext) {
	var frac float64
	along, _ := s.effectiveThumbDims()
	bw := s.EffectiveTheme().Slider.Group(s.Variant()).BorderWidth
	inset := bw + 1

	if s.orientation == Vertical {
		innerH := s.Height - 2*inset
		trackRange := innerH - along
		if trackRange <= 0 {
			return
		}
		posInTrack := ctx.GlobalY - s.dragOriginW - s.grabOffset
		frac = 1 - posInTrack/trackRange
	} else {
		innerW := s.Width - 2*inset
		trackRange := innerW - along
		if trackRange <= 0 {
			return
		}
		posInTrack := ctx.GlobalX - s.dragOriginW - s.grabOffset
		frac = posInTrack / trackRange
	}

	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}

	newVal := s.min + frac*(s.max-s.min)
	s.SetValue(newVal)
}

func (s *Slider) handleTrackClick(ctx sg.ClickContext) {
	var frac float64
	along, _ := s.effectiveThumbDims()
	bw := s.EffectiveTheme().Slider.Group(s.Variant()).BorderWidth
	inset := bw + 1

	if s.orientation == Vertical {
		innerH := s.Height - 2*inset
		trackRange := innerH - along
		if trackRange <= 0 {
			return
		}
		posInTrack := ctx.LocalY - inset - along/2
		frac = 1 - posInTrack/trackRange
	} else {
		innerW := s.Width - 2*inset
		trackRange := innerW - along
		if trackRange <= 0 {
			return
		}
		posInTrack := ctx.LocalX - inset - along/2
		frac = posInTrack / trackRange
	}

	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}

	newVal := s.min + frac*(s.max-s.min)
	s.SetValue(newVal)
}
