package widget

import (
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// GradientMode and Gradient are re-exported from render for widget-layer use.
type GradientMode = render.GradientMode

const (
	GradientModeH          = render.GradientModeH
	GradientModeV          = render.GradientModeV
	GradientModeFourCorner = render.GradientModeFourCorner
)

// Gradient is the value type for GradientEditor.
type Gradient = render.Gradient

// ---------------------------------------------------------------------------
// GradientEditor
// ---------------------------------------------------------------------------

// GradientEditor edits horizontal, vertical, or 4-corner gradients.
type GradientEditor struct {
	Component

	value    render.Gradient
	valueRef *Ref[render.Gradient]
	watch    WatchHandle
	onChange func(render.Gradient)

	source      *sg.FontFamily
	font        *sg.FontFamily
	displaySize float64

	allowedModes     []render.GradientMode
	showModeSelector bool

	// Sub-widgets
	modeBar     *ToggleButtonBar
	previewMesh *sg.Node // willow.Mesh

	// H/V mode pickers and labels
	startPicker *ColorPicker
	endPicker   *ColorPicker
	startLabel  *Label
	endLabel    *Label

	// 4-corner mode pickers
	tlPicker *ColorPicker
	trPicker *ColorPicker
	brPicker *ColorPicker
	blPicker *ColorPicker

	// guard against re-entrant picker callbacks
	updatingPickers bool
}

// NewGradientEditor creates a GradientEditor with the given name, font source, and display size.
func NewGradientEditor(name string, source *sg.FontFamily, displaySize float64) *GradientEditor {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	ge := &GradientEditor{
		source:           source,
		font:             font,
		displaySize:      displaySize,
		showModeSelector: true,
		allowedModes:     []render.GradientMode{GradientModeH, GradientModeV, GradientModeFourCorner},
		value: render.Gradient{
			Mode:   GradientModeH,
			Colors: geDefaultColors(),
		},
	}
	initComponent(&ge.Component, name)
	ge.initBackground(name)
	ge.initBorder(name)

	ge.buildModeBar()

	// H/V pickers
	ge.startPicker = NewColorPicker(name+"-start", source, displaySize)
	ge.startPicker.SetShowAlpha(false)
	ge.startPicker.SetOnChange(func(c sg.Color) { ge.onStartEndChange(c, true) })
	ge.node.AddChild(ge.startPicker.node)

	ge.endPicker = NewColorPicker(name+"-end", source, displaySize)
	ge.endPicker.SetShowAlpha(false)
	ge.endPicker.SetOnChange(func(c sg.Color) { ge.onStartEndChange(c, false) })
	ge.node.AddChild(ge.endPicker.node)

	ge.startLabel = NewLabel(name+"-start-lbl", "Start", source, displaySize)
	ge.node.AddChild(ge.startLabel.node)

	ge.endLabel = NewLabel(name+"-end-lbl", "End", source, displaySize)
	ge.node.AddChild(ge.endLabel.node)

	// 4-corner pickers
	ge.tlPicker = NewColorPicker(name+"-tl", source, displaySize)
	ge.tlPicker.SetShowAlpha(false)
	ge.tlPicker.SetOnChange(func(c sg.Color) { ge.onCornerChange(c, 0) })
	ge.node.AddChild(ge.tlPicker.node)

	ge.trPicker = NewColorPicker(name+"-tr", source, displaySize)
	ge.trPicker.SetShowAlpha(false)
	ge.trPicker.SetOnChange(func(c sg.Color) { ge.onCornerChange(c, 1) })
	ge.node.AddChild(ge.trPicker.node)

	ge.brPicker = NewColorPicker(name+"-br", source, displaySize)
	ge.brPicker.SetShowAlpha(false)
	ge.brPicker.SetOnChange(func(c sg.Color) { ge.onCornerChange(c, 2) })
	ge.node.AddChild(ge.brPicker.node)

	ge.blPicker = NewColorPicker(name+"-bl", source, displaySize)
	ge.blPicker.SetShowAlpha(false)
	ge.blPicker.SetOnChange(func(c sg.Color) { ge.onCornerChange(c, 3) })
	ge.node.AddChild(ge.blPicker.node)

	ge.onThemeChange = func() { ge.UpdateVisuals() }

	ge.updatePickerValues()
	ge.SetSize(300, ge.naturalHeight())
	ge.UpdateVisuals()
	return ge
}

func geDefaultColors() render.GradientColors {
	return render.GradientColors{
		TopLeft:     sg.RGBA(0, 0, 0, 1),
		TopRight:    sg.RGBA(1, 1, 1, 1),
		BottomRight: sg.RGBA(1, 1, 1, 1),
		BottomLeft:  sg.RGBA(0, 0, 0, 1),
	}
}

// buildModeBar creates and attaches the mode selector ToggleButtonBar.
func (ge *GradientEditor) buildModeBar() {
	if ge.modeBar != nil {
		ge.node.RemoveChild(ge.modeBar.node)
	}
	ge.modeBar = NewToggleButtonBar(ge.node.Name+"-mode", ge.source, ge.displaySize)
	for _, m := range ge.allowedModes {
		switch m {
		case GradientModeH:
			ge.modeBar.AddButton("H")
		case GradientModeV:
			ge.modeBar.AddButton("V")
		case GradientModeFourCorner:
			ge.modeBar.AddButton("4-Corner")
		}
	}
	ge.modeBar.SetOnChange(func(idx int) {
		if ge.updatingPickers {
			return
		}
		if idx >= 0 && idx < len(ge.allowedModes) {
			ge.applyModeSwitch(ge.allowedModes[idx])
		}
	})
	ge.node.AddChild(ge.modeBar.node)
}

// SetValue sets the current gradient without firing OnChange.
func (ge *GradientEditor) SetValue(g render.Gradient) {
	ge.value = g
	if ge.valueRef != nil {
		ge.valueRef.Set(g)
	}
	ge.syncModeBar()
	ge.updatePickerValues()
	ge.layout()
	ge.MarkDrawDirty()
}

// Value returns the current gradient.
func (ge *GradientEditor) Value() render.Gradient {
	return ge.value
}

// BindValue binds a *Ref[Gradient] so the editor writes back on every change.
func (ge *GradientEditor) BindValue(ref *Ref[render.Gradient]) {
	ge.watch.Stop()
	ge.valueRef = ref
	ge.value = ref.Peek()
	ge.syncModeBar()
	ge.updatePickerValues()
	ge.layout()
	ge.watch = WatchValue(ref, func(_, newVal render.Gradient) {
		ge.value = newVal
		ge.syncModeBar()
		ge.updatePickerValues()
		ge.layout()
		ge.MarkDrawDirty()
	})
}

// SetOnChange sets a callback invoked on every edit (color change or mode switch).
func (ge *GradientEditor) SetOnChange(fn func(render.Gradient)) {
	ge.onChange = fn
}

// SetAllowedModes restricts which modes appear in the mode selector.
// If the current mode is not in the list, the widget switches to the first allowed mode.
func (ge *GradientEditor) SetAllowedModes(modes ...render.GradientMode) {
	ge.allowedModes = modes
	ge.buildModeBar()
	// Switch mode if current is not in allowed list.
	inList := false
	for _, m := range modes {
		if m == ge.value.Mode {
			inList = true
			break
		}
	}
	if !inList && len(modes) > 0 {
		ge.value.Mode = modes[0]
		ge.updatePickerValues()
	}
	ge.syncModeBar()
	ge.layout()
}

// SetShowModeSelector shows or hides the H | V | 4-corner tab row.
func (ge *GradientEditor) SetShowModeSelector(show bool) {
	ge.showModeSelector = show
	ge.modeBar.node.SetVisible(show)
	ge.layout()
}

// SetSize sets the widget dimensions and re-layouts children.
func (ge *GradientEditor) SetSize(w, h float64) {
	ge.Width = w
	ge.Height = h
	ge.resizeBackground(w, h)
	ge.resizeBorder(w, h)
	ge.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	ge.layout()
	ge.MarkLayoutDirty()
}

// UpdateVisuals updates the visual appearance from the theme.
func (ge *GradientEditor) UpdateVisuals() {
	ge.state = computeState(ge.enabled, ge.focused, ge.hovered, ge.pressed)
	th := ge.EffectiveTheme()
	group := th.GradientEditor.Group(ge.Variant())

	cr := resolveCornerRadius(group.CornerRadius, ge.Height)
	ge.applyCornerRadius(cr)

	bg := group.Background.Resolve(ge.state)
	ge.applyBackground(bg)
	ge.applyBorder(group.BorderColor.Resolve(ge.state), group.BorderWidth, bg)

	ge.layout()
	ge.MarkDrawDirty()
}

// ---------------------------------------------------------------------------
// Internal layout and mode helpers
// ---------------------------------------------------------------------------

type geThemeVals struct {
	previewHeight float64
	previewSize   float64
	padding       render.Insets
}

func (ge *GradientEditor) themeVals() geThemeVals {
	th := ge.EffectiveTheme()
	group := th.GradientEditor.Group(ge.Variant())
	ph := group.PreviewHeight
	if ph <= 0 {
		ph = 40
	}
	ps := group.PreviewSize
	if ps <= 0 {
		ps = 140
	}
	pad := group.Padding
	if pad.IsAuto() {
		pad = render.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8}
	}
	return geThemeVals{previewHeight: ph, previewSize: ps, padding: pad}
}

func (ge *GradientEditor) naturalHeight() float64 {
	tv := ge.themeVals()
	const modeBarH = 32
	const pickerH = 28
	const gap = 6
	h := tv.padding.Top + tv.padding.Bottom + tv.previewHeight + gap + pickerH + ge.displaySize + 4
	if ge.showModeSelector {
		h += modeBarH + gap
	}
	return h
}

func (ge *GradientEditor) syncModeBar() {
	ge.updatingPickers = true
	defer func() { ge.updatingPickers = false }()
	for i, m := range ge.allowedModes {
		if m == ge.value.Mode {
			ge.modeBar.SetSelected(i)
			return
		}
	}
	if len(ge.allowedModes) > 0 {
		ge.modeBar.SetSelected(0)
	}
}

func (ge *GradientEditor) applyModeSwitch(newMode render.GradientMode) {
	old := ge.value.Mode
	if old == newMode {
		return
	}
	ge.value.Mode = newMode
	switch newMode {
	case GradientModeH:
		if old == GradientModeFourCorner {
			// Collapse 4-corner → H: use top row, copy down.
			ge.value.Colors.BottomLeft = ge.value.Colors.TopLeft
			ge.value.Colors.BottomRight = ge.value.Colors.TopRight
		} else if old == GradientModeV {
			// Rotate V → H: old top becomes left, old bottom becomes right.
			prevBL := ge.value.Colors.BottomLeft
			ge.value.Colors.BottomLeft = ge.value.Colors.TopLeft
			ge.value.Colors.TopRight = prevBL
			ge.value.Colors.BottomRight = prevBL
		}
	case GradientModeV:
		if old == GradientModeFourCorner {
			// Collapse 4-corner → V: use left column, copy right.
			ge.value.Colors.TopRight = ge.value.Colors.TopLeft
			ge.value.Colors.BottomRight = ge.value.Colors.BottomLeft
		} else if old == GradientModeH {
			// Rotate H → V: old left becomes top, old right becomes bottom.
			prevTR := ge.value.Colors.TopRight
			ge.value.Colors.TopRight = ge.value.Colors.TopLeft
			ge.value.Colors.BottomLeft = prevTR
			ge.value.Colors.BottomRight = prevTR
		}
	}
	ge.updatePickerValues()
	ge.layout()
	ge.notifyChange()
	ge.MarkDrawDirty()
}

func (ge *GradientEditor) onStartEndChange(c sg.Color, isStart bool) {
	if ge.updatingPickers {
		return
	}
	switch ge.value.Mode {
	case GradientModeH:
		if isStart {
			ge.value.Colors.TopLeft = c
			ge.value.Colors.BottomLeft = c
		} else {
			ge.value.Colors.TopRight = c
			ge.value.Colors.BottomRight = c
		}
	case GradientModeV:
		if isStart {
			ge.value.Colors.TopLeft = c
			ge.value.Colors.TopRight = c
		} else {
			ge.value.Colors.BottomLeft = c
			ge.value.Colors.BottomRight = c
		}
	}
	ge.refreshPreviewMesh()
	ge.notifyChange()
}

func (ge *GradientEditor) onCornerChange(c sg.Color, corner int) {
	if ge.updatingPickers {
		return
	}
	switch corner {
	case 0:
		ge.value.Colors.TopLeft = c
	case 1:
		ge.value.Colors.TopRight = c
	case 2:
		ge.value.Colors.BottomRight = c
	case 3:
		ge.value.Colors.BottomLeft = c
	}
	ge.refreshPreviewMesh()
	ge.notifyChange()
}

func (ge *GradientEditor) updatePickerValues() {
	ge.updatingPickers = true
	defer func() { ge.updatingPickers = false }()
	c := ge.value.Colors
	switch ge.value.Mode {
	case GradientModeH:
		ge.startPicker.SetValue(c.TopLeft)
		ge.endPicker.SetValue(c.TopRight)
	case GradientModeV:
		ge.startPicker.SetValue(c.TopLeft)
		ge.endPicker.SetValue(c.BottomLeft)
	case GradientModeFourCorner:
		ge.tlPicker.SetValue(c.TopLeft)
		ge.trPicker.SetValue(c.TopRight)
		ge.brPicker.SetValue(c.BottomRight)
		ge.blPicker.SetValue(c.BottomLeft)
	}
}

func (ge *GradientEditor) ensurePreviewMesh(pw, ph float64) {
	if pw <= 0 || ph <= 0 {
		return
	}
	verts, inds := render.RoundedRectGradientMesh(pw, ph, 0, 8, &ge.value.Colors)
	if ge.previewMesh == nil {
		ge.previewMesh = sg.NewMesh(ge.node.Name+"-preview", sg.WhitePixel, verts, inds)
		ge.node.AddChild(ge.previewMesh)
	} else {
		ge.previewMesh.SetMeshVertices(verts)
		ge.previewMesh.InvalidateMeshAABB()
	}
}

func (ge *GradientEditor) refreshPreviewMesh() {
	if ge.previewMesh == nil {
		return
	}
	tv := ge.themeVals()
	var pw, ph float64
	if ge.value.Mode == GradientModeFourCorner {
		pw = tv.previewSize
		ph = tv.previewSize
	} else {
		pw = ge.Width - tv.padding.Left - tv.padding.Right
		ph = tv.previewHeight
	}
	ge.ensurePreviewMesh(pw, ph)
}

// layout positions all child widgets for the current mode and size.
func (ge *GradientEditor) layout() {
	tv := ge.themeVals()
	pad := tv.padding
	const modeBarH = 32
	const pickerH = 28
	const gap = 6
	innerW := ge.Width - pad.Left - pad.Right
	curY := pad.Top

	// Mode selector
	ge.modeBar.node.SetVisible(ge.showModeSelector)
	if ge.showModeSelector {
		ge.modeBar.SetSize(innerW, modeBarH)
		ge.modeBar.node.SetPosition(pad.Left, curY)
		curY += modeBarH + gap
	}

	mode := ge.value.Mode
	hvVisible := mode == GradientModeH || mode == GradientModeV
	fcVisible := mode == GradientModeFourCorner

	ge.startPicker.node.SetVisible(hvVisible)
	ge.endPicker.node.SetVisible(hvVisible)
	ge.startLabel.node.SetVisible(hvVisible)
	ge.endLabel.node.SetVisible(hvVisible)
	ge.tlPicker.node.SetVisible(fcVisible)
	ge.trPicker.node.SetVisible(fcVisible)
	ge.brPicker.node.SetVisible(fcVisible)
	ge.blPicker.node.SetVisible(fcVisible)

	if hvVisible {
		pw := innerW
		ph := tv.previewHeight
		ge.ensurePreviewMesh(pw, ph)
		if ge.previewMesh != nil {
			ge.previewMesh.SetPosition(pad.Left, curY)
			ge.previewMesh.SetVisible(true)
		}
		curY += ph + gap

		labelH := ge.displaySize + 2
		halfW := innerW / 2
		ge.startLabel.SetPosition(pad.Left, curY)
		ge.endLabel.SetPosition(pad.Left+halfW, curY)
		curY += labelH + 2

		ge.startPicker.SetSize(halfW-4, pickerH)
		ge.startPicker.node.SetPosition(pad.Left, curY)
		ge.endPicker.SetSize(halfW-4, pickerH)
		ge.endPicker.node.SetPosition(pad.Left+halfW, curY)

	} else if fcVisible {
		previewSz := tv.previewSize
		const pickerW = 100.0
		maxPreviewW := innerW - 2*(pickerW+4)
		if previewSz > maxPreviewW && maxPreviewW > 20 {
			previewSz = maxPreviewW
		}
		// Cap by available vertical space.
		availH := ge.Height - pad.Top - pad.Bottom
		if ge.showModeSelector {
			availH -= modeBarH + gap
		}
		if previewSz > availH && availH > 20 {
			previewSz = availH
		}

		px := pad.Left + pickerW + 4
		py := curY

		ge.ensurePreviewMesh(previewSz, previewSz)
		if ge.previewMesh != nil {
			ge.previewMesh.SetPosition(px, py)
			ge.previewMesh.SetVisible(true)
		}

		// TL top-left, TR top-right
		ge.tlPicker.SetSize(pickerW, pickerH)
		ge.tlPicker.node.SetPosition(pad.Left, py)
		ge.trPicker.SetSize(pickerW, pickerH)
		ge.trPicker.node.SetPosition(px+previewSz+4, py)

		// BL bottom-left, BR bottom-right
		ge.blPicker.SetSize(pickerW, pickerH)
		ge.blPicker.node.SetPosition(pad.Left, py+previewSz-pickerH)
		ge.brPicker.SetSize(pickerW, pickerH)
		ge.brPicker.node.SetPosition(px+previewSz+4, py+previewSz-pickerH)
	}
}

func (ge *GradientEditor) notifyChange() {
	if ge.valueRef != nil {
		ge.valueRef.Set(ge.value)
	}
	if ge.onChange != nil {
		ge.onChange(ge.value)
	}
}
