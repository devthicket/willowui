package widget

import (
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

// Toggle is a binary on/off switch with an animated sliding thumb.
type Toggle struct {
	Component
	track     *sg.Node // WhitePixel sprite: track background
	thumb     *sg.Node // WhitePixel sprite: sliding thumb
	trackPoly *sg.Node // polygon for rounded track (nil when sharp)
	thumbPoly *sg.Node // polygon for rounded thumb (nil when sharp)
	value     *Ref[bool]
	watch     WatchHandle
	onChange  func(bool)

	// Animation state (updated via Update method).
	thumbTween *gween.Tween
}

// DefaultToggleWidth is the default track width.
const DefaultToggleWidth = 48

// DefaultToggleHeight is the default track height.
const DefaultToggleHeight = 24

// DefaultToggleThumbSize is the default thumb diameter.
const DefaultToggleThumbSize = 20

// toggleCornerSegments is the number of line segments per corner arc for
// the toggle's small rounded shapes.
const toggleCornerSegments = 8

// NewToggle creates a Toggle switch with default dimensions.
func NewToggle(name string) *Toggle {
	t := &Toggle{
		value: NewRef(false),
	}
	initComponent(&t.Component, name)

	// Track.
	t.track = sg.NewSprite(name+"-track", sg.TextureRegion{})

	t.track.SetScale(DefaultToggleWidth, DefaultToggleHeight)
	t.node.AddChild(t.track)

	// Thumb.
	t.thumb = sg.NewSprite(name+"-thumb", sg.TextureRegion{})

	t.thumb.SetScale(DefaultToggleThumbSize, DefaultToggleThumbSize)
	thumbY := (DefaultToggleHeight - DefaultToggleThumbSize) / 2.0
	t.thumb.SetPosition(2, thumbY)
	t.node.AddChild(t.thumb)

	t.Width = DefaultToggleWidth
	t.Height = DefaultToggleHeight
	t.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: DefaultToggleWidth, Height: DefaultToggleHeight}

	// Click toggles value.
	t.node.OnClick(func(ctx sg.ClickContext) {
		if !t.enabled {
			return
		}
		newVal := !t.value.Peek()
		t.value.Set(newVal)
		DefaultScheduler.Flush()
		t.updateThumbPosition()
		t.UpdateVisuals()
		if t.onChange != nil {
			t.onChange(newVal)
		}
	})

	// Auto-update: advance thumb animation and keyboard activation.
	t.node.OnUpdate = func(dt float64) {
		t.Update(float32(dt))

		// Keyboard activation: Space toggles.
		if t.focused && t.enabled {
			if DefaultInputManager.IsKeyJustAvailable(engine.KeySpace) {
				newVal := !t.value.Peek()
				t.value.Set(newVal)
				DefaultScheduler.Flush()
				t.updateThumbPosition()
				t.UpdateVisuals()
				if t.onChange != nil {
					t.onChange(newVal)
				}
				DefaultInputManager.Consume(engine.KeySpace)
			}
		}
	}

	t.wireVisualCallbacks(t.UpdateVisuals)
	t.SetCursorShape(engine.CursorShapePointer)

	// Focus: toggles participate in tab and spatial nav.
	t.enableFocusNavigation()

	t.UpdateVisuals()
	return t
}

// Value returns the current toggle state.
func (t *Toggle) Value() bool {
	return t.value.Peek()
}

// SetValue sets the toggle state.
func (t *Toggle) SetValue(v bool) {
	t.value.Set(v)
	DefaultScheduler.Flush()
	t.updateThumbPosition()
	t.UpdateVisuals()
}

// SetOnChange sets the callback invoked when the toggle changes.
func (t *Toggle) SetOnChange(fn func(bool)) {
	t.onChange = fn
}

// BindValue binds the toggle to a reactive Ref[bool] for two-way binding.
func (t *Toggle) BindValue(ref *Ref[bool]) {
	t.watch.Stop()
	t.value = ref
	t.SetValue(ref.Peek()) // sync immediately
	t.watch = WatchValue(ref, func(_, newVal bool) {
		t.updateThumbPosition()
		t.UpdateVisuals()
	})
}

// SetEnabled overrides Component.SetEnabled to also update visuals.
func (t *Toggle) SetEnabled(v bool) {
	t.Component.SetEnabled(v)
	t.UpdateVisuals()
}

// UpdateVisuals applies theme colors based on the current state.
func (t *Toggle) UpdateVisuals() {
	t.state = computeState(t.enabled, t.focused, t.hovered, t.value.Peek())
	group := t.EffectiveTheme().Toggle.Group(t.Variant())
	cr := resolveCornerRadius(group.CornerRadius, t.Height)

	trackColor := group.TrackColor.Resolve(t.state)
	thumbColor := group.ThumbColor.Resolve(t.state)

	if cr > 0 {
		// Rounded track.
		t.ensureTrackPoly()
		pts := render.RoundedRectPoints(t.Width, t.Height, cr, toggleCornerSegments)
		sg.SetPolygonPoints(t.trackPoly, pts)
		t.trackPoly.SetColor(trackColor)
		t.trackPoly.SetVisible(true)
		t.track.SetVisible(false)

		// Rounded thumb.
		thumbSize := float64(DefaultToggleThumbSize)
		thumbR := thumbSize / 2
		t.ensureThumbPoly()
		tpts := render.RoundedRectPoints(thumbSize, thumbSize, thumbR, toggleCornerSegments)
		sg.SetPolygonPoints(t.thumbPoly, tpts)
		t.thumbPoly.SetColor(thumbColor)
		t.thumbPoly.SetVisible(true)
		if t.thumbTween == nil {
			t.thumbPoly.SetPosition(t.thumb.X(), t.thumb.Y())
		}
		t.thumb.SetVisible(false)
	} else {
		// Sharp: use flat sprites.
		t.track.SetColor(trackColor)
		t.track.SetVisible(true)
		t.thumb.SetColor(thumbColor)
		t.thumb.SetVisible(true)
		if t.trackPoly != nil {
			t.trackPoly.SetVisible(false)
		}
		if t.thumbPoly != nil {
			t.thumbPoly.SetVisible(false)
		}
	}
	t.applyFocusRingEx(group.FocusColor.Resolve(t.state), group.FocusRingWidth, t.Width, t.Height, cr)
	t.MarkDrawDirty()
}

// Update advances any active thumb animation. Call from your scene's UpdateFunc.
func (t *Toggle) Update(dt float32) {
	if t.thumbTween != nil {
		x, done := t.thumbTween.Update(dt)
		t.setThumbX(float64(x))
		if done {
			t.thumbTween = nil
		}
	}

	t.UpdateVisuals()
}

// Dispose stops reactive watches and disposes the component tree.
func (t *Toggle) Dispose() {
	t.watch.Stop()
	t.Component.Dispose()
}

// thumbTargetX returns the target X position of the thumb for the current value.
func (t *Toggle) thumbTargetX() float64 {
	thumbPad := 2.0
	if t.value.Peek() {
		return t.Width - DefaultToggleThumbSize - thumbPad
	}
	return thumbPad
}

// setThumbX moves both the thumb sprite and rounded polygon to the given X.
func (t *Toggle) setThumbX(x float64) {
	t.thumb.SetPosition(x, t.thumb.Y())
	if t.thumbPoly != nil {
		t.thumbPoly.SetPosition(x, t.thumb.Y())
	}
}

// updateThumbPosition animates the thumb to the correct position.
func (t *Toggle) updateThumbPosition() {
	t.thumbTween = gween.New(float32(t.thumb.X()), float32(t.thumbTargetX()), 0.12, ease.OutCubic)
}

// snapThumbPosition sets the thumb position instantly (no animation).
func (t *Toggle) snapThumbPosition() {
	t.thumbTween = nil
	t.setThumbX(t.thumbTargetX())
}

// ensureTrackPoly lazily creates the polygon node for the rounded track.
func (t *Toggle) ensureTrackPoly() {
	if t.trackPoly != nil {
		return
	}
	pts := render.RoundedRectPoints(1, 1, 0, toggleCornerSegments)
	t.trackPoly = sg.NewPolygon(t.node.Name+"-track-poly", pts)
	t.trackPoly.SetVisible(false)
	// Insert before the flat track sprite so z-order is correct.
	t.node.AddChildAt(t.trackPoly, 0)
}

// ensureThumbPoly lazily creates the polygon node for the rounded thumb.
func (t *Toggle) ensureThumbPoly() {
	if t.thumbPoly != nil {
		return
	}
	pts := render.RoundedRectPoints(1, 1, 0, toggleCornerSegments)
	t.thumbPoly = sg.NewPolygon(t.node.Name+"-thumb-poly", pts)
	t.thumbPoly.SetVisible(false)
	t.node.AddChild(t.thumbPoly)
}

// TrackNode returns the flat track sprite node. Used for testing.
func (t *Toggle) TrackNode() *sg.Node { return t.track }

// ThumbNode returns the flat thumb sprite node. Used for testing.
func (t *Toggle) ThumbNode() *sg.Node { return t.thumb }

// TrackPoly returns the rounded track polygon node, or nil if not yet created.
// Used for testing.
func (t *Toggle) TrackPoly() *sg.Node { return t.trackPoly }

// ThumbPoly returns the rounded thumb polygon node, or nil if not yet created.
// Used for testing.
func (t *Toggle) ThumbPoly() *sg.Node { return t.thumbPoly }
