package widget

import (
	"math"
	"strconv"

	"github.com/devthicket/willowui/internal/sg"
)

// MeterBar displays a horizontal bar that fills according to a value range.
// The default range is [0, 1]; call SetRange to use a custom range.
// It is non-interactive (no pointer callbacks needed).
type MeterBar struct {
	Component
	fillComp  Component // sub-component: filled portion
	label     *Label    // optional percentage label
	value     *Ref[float64]
	watch     WatchHandle
	showLabel bool
	min, max  float64

	fillOverride    bool     // true when SetFillColor has been called
	fillOverrideClr sg.Color // color to use instead of theme fill
}

// ProgressBar is an alias for MeterBar.
type ProgressBar = MeterBar

// Default meter bar dimensions.
const (
	DefaultMeterWidth  = 200
	DefaultMeterHeight = 20
)

// NewMeterBar creates a meter bar with range [0, 1] and value 0.
func NewMeterBar(name string) *MeterBar {
	mb := &MeterBar{
		value: NewRef(0.0),
		min:   0,
		max:   1,
	}
	initComponent(&mb.Component, name)

	mb.initBackground(name)
	mb.initBorder(name)

	// Fill sub-component.
	initComponent(&mb.fillComp, name+"-fill")
	mb.fillComp.initBackground(name + "-fill")
	mb.fillComp.initBorder(name + "-fill")
	mb.node.AddChild(mb.fillComp.node)

	mb.SetSize(DefaultMeterWidth, DefaultMeterHeight)
	mb.node.Interactable = false

	mb.onThemeChange = func() { mb.applyThemeColors() }
	mb.applyThemeColors()
	return mb
}

// NewProgressBar creates a MeterBar with range [0, 1]. Alias for NewMeterBar.
func NewProgressBar(name string) *MeterBar {
	return NewMeterBar(name)
}

func (mb *MeterBar) applyThemeColors() {
	group := mb.EffectiveTheme().MeterBar.Group(mb.Variant())

	// Track.
	mb.applyCornerRadius(group.CornerRadius)
	bg := group.Background.Resolve(StateDefault)
	mb.applyBackground(bg)
	mb.applyBorder(group.Border.Resolve(StateDefault), group.BorderWidth, bg)

	// Fill sub-component styling (corner radius set here; size/position in updateFill).
	mb.fillComp.applyCornerRadius(group.FillCornerRadius)

	if mb.label != nil && mb.showLabel {
		mb.label.SetColor(group.TextColor.Resolve(StateDefault))
	}

	// Recompute fill dimensions and apply fill background/border with
	// correct inset accounting for the track's border width.
	mb.updateFill()
	mb.MarkDrawDirty()
}

// Value returns the current normalized fill value (0-1), regardless of range.
func (mb *MeterBar) Value() float64 {
	return mb.value.Peek()
}

// SetRange sets the value range. Default is [0, 1].
func (mb *MeterBar) SetRange(min, max float64) {
	mb.min = min
	mb.max = max
}

// SetValue sets the meter value in the [min, max] range, mapping to 0-1 internally.
// With the default range [0, 1] this behaves identically to SetProgress.
func (mb *MeterBar) SetValue(v float64) {
	if v < mb.min {
		v = mb.min
	}
	if v > mb.max {
		v = mb.max
	}
	var normalized float64
	if mb.max > mb.min {
		normalized = (v - mb.min) / (mb.max - mb.min)
	}
	mb.value.Set(normalized)
	DefaultScheduler.Flush()
	mb.updateFill()
}

// SetProgress sets the normalized 0-1 fill directly, bypassing range mapping.
// Used internally by BindValue.
func (mb *MeterBar) SetProgress(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	mb.value.Set(v)
	DefaultScheduler.Flush()
	mb.updateFill()
}

// SetSize sets the meter bar dimensions.
func (mb *MeterBar) SetSize(w, h float64) {
	mb.Width = w
	mb.Height = h
	mb.resizeBackground(w, h)
	mb.resizeBorder(w, h)
	mb.updateFill()
	mb.MarkLayoutDirty()
}

// SetShowLabel enables or disables a percentage label overlay.
func (mb *MeterBar) SetShowLabel(show bool, source *sg.FontFamily, displaySize float64) {
	mb.showLabel = show
	if show && mb.label == nil && source != nil {
		mb.label = NewLabel(mb.Name()+"-label", "0%", source, displaySize)
		mb.label.SetColor(mb.EffectiveTheme().MeterBar.Group(mb.Variant()).TextColor.Resolve(StateDefault))
		mb.label.AddToNode(mb.node)
	}
	if mb.label != nil {
		mb.label.SetVisible(show)
	}
	mb.updateFill()
}

// BindValue binds the meter bar to a reactive Ref[float64] (0-1 normalized).
func (mb *MeterBar) BindValue(ref *Ref[float64]) {
	mb.value = ref
	bindRef(&mb.watch, ref, mb.SetProgress)
}

// Dispose stops reactive watches and disposes children.
func (mb *MeterBar) Dispose() {
	mb.watch.Stop()
	if mb.label != nil {
		mb.label.Dispose()
	}
	mb.fillComp.Dispose()
	mb.Component.Dispose()
}

// SetFillColor overrides the theme fill color with a custom color.
func (mb *MeterBar) SetFillColor(c sg.Color) {
	mb.fillOverride = true
	mb.fillOverrideClr = c
	mb.updateFill()
}

// ClearFillColor removes the fill color override, reverting to the theme color.
func (mb *MeterBar) ClearFillColor() {
	mb.fillOverride = false
	mb.updateFill()
}

// FillNode returns the fill component's willow node. Used for testing.
func (mb *MeterBar) FillNode() *sg.Node { return mb.fillComp.node }

// FillWidth returns the fill component's Width. Used for testing.
func (mb *MeterBar) FillWidth() float64 { return mb.fillComp.Width }

// LabelComp returns the progress label, or nil if not set. Used for testing.
func (mb *MeterBar) LabelComp() *Label { return mb.label }

func (mb *MeterBar) updateFill() {
	v := mb.value.Peek()
	group := mb.EffectiveTheme().MeterBar.Group(mb.Variant())
	bw := group.BorderWidth
	// Add 1px gap between track border and fill for visual separation.
	inset := bw + 1

	// Inset fill so it doesn't overlap the track border.
	innerW := mb.Width - 2*inset
	innerH := mb.Height - 2*inset
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}

	fillW := v * innerW
	mb.fillComp.Width = fillW
	mb.fillComp.Height = innerH
	mb.fillComp.node.SetPosition(inset, inset)
	mb.fillComp.resizeBackground(fillW, innerH)
	mb.fillComp.resizeBorder(fillW, innerH)

	// Apply fill background and border with current dimensions.
	var fillBg Background
	if mb.fillOverride {
		fillBg = Background{Type: BgSolid, Color: mb.fillOverrideClr}
	} else {
		fillBg = group.FillBackground.Resolve(StateDefault)
	}
	mb.fillComp.applyBackground(fillBg)
	mb.fillComp.applyBorder(group.FillBorder.Resolve(StateDefault), group.FillBorderWidth, fillBg)

	if mb.label != nil && mb.showLabel {
		pct := int(math.Round(v * 100))
		mb.label.SetText(strconv.Itoa(pct) + "%")
		// Center label within the full bar.
		mb.label.SetPosition((mb.Width-mb.label.Width)/2, (mb.Height-mb.label.Height)/2)
	}
	mb.MarkDrawDirty()
}
