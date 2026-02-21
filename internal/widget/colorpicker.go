package widget

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/devthicket/willowui/internal/colorutil"
	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// ColorMode selects which input mode the picker popup displays.
type ColorMode int

const (
	ColorModeHex ColorMode = iota
	ColorModeRGB           // 0–255
	ColorModeHSV           // internal gradient representation
	ColorModeHSL
	ColorModeFloat // sg.Color 0–1
)

var colorModeLabels = [...]string{"Hex", "RGB", "HSV", "HSL", "Tint"}

// colorPickerOverlayZIndex is above menus but below tooltips.
const colorPickerOverlayZIndex = 550_000

// ---------------------------------------------------------------------------
// Shared singleton images
// ---------------------------------------------------------------------------

// hueBarImage is a 360×1 horizontal hue gradient shared across all pickers.
var hueBarImage engine.Image

func ensureHueBarImage() {
	if hueBarImage != nil {
		return
	}
	const w = 360
	hueBarImage = engine.NewImage(w, 1)
	for x := 0; x < w; x++ {
		h := float64(x) / float64(w)
		c := sg.ColorFromHSV(h, 1, 1)
		hueBarImage.Set(x, 0, color.NRGBA{
			R: uint8(math.Round(c.R() * 255)),
			G: uint8(math.Round(c.G() * 255)),
			B: uint8(math.Round(c.B() * 255)),
			A: 255,
		})
	}
}

// checkerImage is an 8×8 checkerboard for alpha transparency indication.
var checkerImage engine.Image

func ensureCheckerImage() {
	if checkerImage != nil {
		return
	}
	checkerImage = engine.NewImage(8, 8)
	light := color.NRGBA{R: 180, G: 180, B: 180, A: 255}
	dark := color.NRGBA{R: 120, G: 120, B: 120, A: 255}
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if (x/4+y/4)%2 == 0 {
				checkerImage.Set(x, y, light)
			} else {
				checkerImage.Set(x, y, dark)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// ColorPicker (trigger swatch)
// ---------------------------------------------------------------------------

// ColorPicker is a swatch + label trigger that opens a floating picker popup.
type ColorPicker struct {
	Component

	value       sg.Color
	valueRef    *Ref[sg.Color]
	watch       WatchHandle
	showAlpha   bool
	defaultMode ColorMode
	onChange    func(sg.Color)
	onCommit    func(sg.Color)

	source      *sg.FontFamily
	font        *sg.FontFamily
	displaySize float64

	// swatch child nodes
	swatchChecker *sg.Node
	swatchColor   *sg.Node
	hexLabel      *Label
}

// NewColorPicker creates a new ColorPicker trigger.
func NewColorPicker(name string, source *sg.FontFamily, displaySize float64) *ColorPicker {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	cp := &ColorPicker{
		value:       sg.RGBA(1, 1, 1, 1),
		showAlpha:   true,
		defaultMode: ColorModeHex,
		source:      source,
		font:        font,
		displaySize: displaySize,
	}
	initComponent(&cp.Component, name)
	cp.initBackground(name)
	cp.initBorder(name)

	// Checkerboard behind swatch for alpha indication
	ensureCheckerImage()
	cp.swatchChecker = sg.NewContainer(name + "-checker")
	cp.swatchChecker.SetZIndex(2)
	cp.node.AddChild(cp.swatchChecker)

	// Color swatch overlay
	cp.swatchColor = sg.NewSprite(name+"-swatch", sg.TextureRegion{})
	cp.swatchColor.SetZIndex(3)
	cp.node.AddChild(cp.swatchColor)

	// Hex label
	cp.hexLabel = NewLabel(name+"-hex", "", source, displaySize)
	cp.hexLabel.node.SetZIndex(4)
	cp.node.AddChild(cp.hexLabel.node)

	cp.node.OnClick(func(_ sg.ClickContext) {
		if !cp.enabled {
			return
		}
		cp.Open()
	})

	cp.onVisualStateChange = func() { cp.UpdateVisuals() }
	cp.onThemeChange = func() { cp.UpdateVisuals() }
	cp.SetCursorShape(engine.CursorShapePointer)

	cp.SetSize(120, 28)
	cp.UpdateVisuals()
	return cp
}

// SetShowAlpha controls whether the alpha channel is editable in the popup.
func (cp *ColorPicker) SetShowAlpha(show bool) {
	cp.showAlpha = show
	cp.updateSwatchDisplay()
}

// SetDefaultMode sets the color mode shown first when the popup opens.
func (cp *ColorPicker) SetDefaultMode(mode ColorMode) {
	cp.defaultMode = mode
}

// SetValue sets the current color without opening the popup.
func (cp *ColorPicker) SetValue(c sg.Color) {
	cp.value = c
	if cp.valueRef != nil {
		cp.valueRef.Set(c)
	}
	cp.updateSwatchDisplay()
}

// Value returns the current color.
func (cp *ColorPicker) Value() sg.Color {
	return cp.value
}

// BindValue binds a *Ref[sg.Color] so the picker writes back to it.
func (cp *ColorPicker) BindValue(ref *Ref[sg.Color]) {
	cp.watch.Stop()
	cp.valueRef = ref
	cp.value = ref.Peek()
	cp.updateSwatchDisplay()
	cp.watch = WatchValue(ref, func(_, newVal sg.Color) {
		cp.value = newVal
		cp.updateSwatchDisplay()
	})
}

// SetOnChange sets a callback invoked whenever the color changes (each drag step).
func (cp *ColorPicker) SetOnChange(fn func(sg.Color)) {
	cp.onChange = fn
}

// SetOnCommit sets a callback invoked when the popup closes with a committed value.
func (cp *ColorPicker) SetOnCommit(fn func(sg.Color)) {
	cp.onCommit = fn
}

// Open programmatically opens the picker popup.
func (cp *ColorPicker) Open() {
	DefaultColorPickerManager.Show(cp)
}

// Close programmatically commits and closes the picker popup.
func (cp *ColorPicker) Close() {
	DefaultColorPickerManager.commitAndClose()
}

// Cancel programmatically cancels the picker, restoring the original color.
func (cp *ColorPicker) Cancel() {
	DefaultColorPickerManager.cancelAndClose()
}

// SetSize sets the trigger dimensions.
func (cp *ColorPicker) SetSize(w, h float64) {
	cp.Width = w
	cp.Height = h
	cp.resizeBackground(w, h)
	cp.resizeBorder(w, h)
	cp.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	cp.MarkLayoutDirty()
	cp.layoutSwatch()
}

func (cp *ColorPicker) layoutSwatch() {
	th := cp.EffectiveTheme()
	group := th.ColorPicker.Group(cp.Variant())

	swatchH := cp.Height - 4
	swatchW := swatchH
	if swatchW > cp.Width*0.4 {
		swatchW = cp.Width * 0.4
	}

	sx := 2.0
	sy := 2.0

	// Position checkerboard — rebuild tiles
	cp.swatchChecker.SetPosition(sx, sy)
	cp.swatchChecker.RemoveChildren()
	tileCheckerboard(cp.swatchChecker, swatchW, swatchH)

	// Position color swatch
	cp.swatchColor.SetPosition(sx, sy)
	cp.swatchColor.SetScale(swatchW, swatchH)

	// Position hex label after swatch
	labelX := sx + swatchW + 4
	labelY := (cp.Height - cp.displaySize) / 2
	cp.hexLabel.SetPosition(labelX, labelY)

	_ = group
}

func (cp *ColorPicker) updateSwatchDisplay() {
	c := cp.value
	cp.swatchColor.SetColor(c)

	// Show checkerboard only when alpha < 1
	cp.swatchChecker.SetVisible(c.A() < 1)

	// Update hex label
	if cp.showAlpha && c.A() < 1 {
		cp.hexLabel.SetText(colorutil.FormatHexA(c))
	} else {
		cp.hexLabel.SetText(colorutil.FormatHex(c))
	}
}

// UpdateVisuals updates the trigger's visual appearance based on state and theme.
func (cp *ColorPicker) UpdateVisuals() {
	cp.state = computeState(cp.enabled, cp.focused, cp.hovered, cp.pressed)
	th := cp.EffectiveTheme()
	group := th.ColorPicker.Group(cp.Variant())

	cr := resolveCornerRadius(group.CornerRadius, cp.Height)
	cp.applyCornerRadius(cr)

	bg := group.Background.Resolve(cp.state)
	cp.applyBackground(bg)
	cp.applyBorder(group.BorderColor.Resolve(cp.state), group.BorderWidth, bg)

	cp.layoutSwatch()
	cp.updateSwatchDisplay()
	cp.MarkDrawDirty()
}

// tileCheckerboard fills a container with 8×8 checker tiles using SetCustomImage.
func tileCheckerboard(parent *sg.Node, w, h float64) {
	ensureCheckerImage()
	cols := int(math.Ceil(w / 8))
	rows := int(math.Ceil(h / 8))
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			tile := sg.NewSprite("checker-tile", sg.TextureRegion{})
			tile.SetCustomImage(checkerImage)
			tile.SetPosition(float64(col*8), float64(row*8))
			parent.AddChild(tile)
		}
	}
}

// ---------------------------------------------------------------------------
// ColorPickerManager
// ---------------------------------------------------------------------------

// ColorPickerManager manages the single active floating color picker popup.
type ColorPickerManager struct {
	scene       *sg.Scene
	tickerNode  *sg.Node
	overlayNode *sg.Node
	dismissNode *sg.Node
	active      *colorPickerPopup
}

// DefaultColorPickerManager is the singleton used by ColorPicker widgets.
var DefaultColorPickerManager = &ColorPickerManager{}

func (m *ColorPickerManager) setScene(s *sg.Scene) {
	if m.active != nil {
		m.commitAndClose()
	}
	m.scene = s
	if s == nil || s.Root == nil {
		return
	}
	m.ensureNodes(s)
}

func (m *ColorPickerManager) ensureNodes(s *sg.Scene) {
	if s == nil || s.Root == nil {
		return
	}
	if m.tickerNode == nil {
		m.tickerNode = sg.NewContainer("colorpicker-ticker")
		m.tickerNode.Interactable = false
		m.tickerNode.SetZIndex(colorPickerOverlayZIndex)
		m.tickerNode.OnUpdate = func(_ float64) {
			DefaultColorPickerManager.tick()
		}
	}
	if m.overlayNode == nil {
		m.overlayNode = sg.NewContainer("colorpicker-overlay")
		m.overlayNode.Interactable = true
		m.overlayNode.SetVisible(false)
		m.overlayNode.SetZIndex(colorPickerOverlayZIndex)
	}
	if m.dismissNode == nil {
		vw, vh := viewportSize()
		m.dismissNode = sg.NewContainer("colorpicker-dismiss")
		m.dismissNode.Interactable = true
		m.dismissNode.HitShape = sg.HitRect{X: 0, Y: 0, Width: vw, Height: vh}
		m.dismissNode.SetZIndex(colorPickerOverlayZIndex - 1)
		m.dismissNode.SetVisible(false)
		m.dismissNode.OnPointerDown(func(_ sg.PointerContext) {
			DefaultColorPickerManager.commitAndClose()
		})
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
	if m.dismissNode.Parent != s.Root {
		if m.dismissNode.Parent != nil {
			m.dismissNode.Parent.RemoveChild(m.dismissNode)
		}
		s.Root.AddChild(m.dismissNode)
	}
}

// Show opens the color picker popup for the given trigger.
func (m *ColorPickerManager) Show(trigger *ColorPicker) {
	if sc := currentScene(); sc != nil {
		m.ensureNodes(sc)
	}
	if m.active != nil {
		m.commitAndClose()
	}

	popup := newColorPickerPopup(trigger)
	popup.build()

	// Position below trigger, screen-edge clamped
	var x, y float64
	wx, wy := trigger.node.LocalToWorld(0, 0)
	x = wx
	y = wy + trigger.Height + 2

	vw, vh := viewportSize()
	if hr, ok := popup.root.HitShape.(sg.HitRect); ok {
		if y+hr.Height > vh-8 {
			y = wy - hr.Height - 2
		}
		if x+hr.Width > vw-8 {
			x = vw - 8 - hr.Width
		}
	}
	if x < 8 {
		x = 8
	}
	if y < 8 {
		y = 8
	}

	popup.root.SetPosition(x, y)
	m.overlayNode.AddChild(popup.root)
	m.overlayNode.SetVisible(true)

	m.dismissNode.HitShape = sg.HitRect{X: 0, Y: 0, Width: vw, Height: vh}
	m.dismissNode.Invalidate()
	m.dismissNode.SetVisible(true)

	m.active = popup
}

func (m *ColorPickerManager) commitAndClose() {
	if m.active == nil {
		return
	}
	popup := m.active
	trigger := popup.trigger
	trigger.value = popup.current
	if trigger.valueRef != nil {
		trigger.valueRef.Set(popup.current)
	}
	trigger.updateSwatchDisplay()
	if trigger.onCommit != nil {
		trigger.onCommit(popup.current)
	}
	m.hideActive()
}

func (m *ColorPickerManager) cancelAndClose() {
	if m.active == nil {
		return
	}
	popup := m.active
	trigger := popup.trigger
	trigger.value = popup.original
	if trigger.valueRef != nil {
		trigger.valueRef.Set(popup.original)
	}
	trigger.updateSwatchDisplay()
	if trigger.onChange != nil {
		trigger.onChange(popup.original)
	}
	m.hideActive()
}

func (m *ColorPickerManager) hideActive() {
	if m.active == nil {
		return
	}
	if m.overlayNode != nil && m.active.root != nil {
		m.overlayNode.RemoveChild(m.active.root)
		m.overlayNode.SetVisible(false)
	}
	if m.dismissNode != nil {
		m.dismissNode.SetVisible(false)
	}
	m.active = nil
}

func (m *ColorPickerManager) tick() {
	if m.active == nil {
		return
	}
	if core.IsKeyJustPressed(engine.KeyEscape) {
		m.cancelAndClose()
	} else if core.IsKeyJustPressed(engine.KeyEnter) {
		m.commitAndClose()
	}
}

// IsOpen returns true if a picker popup is currently visible.
func (m *ColorPickerManager) IsOpen() bool {
	return m.active != nil
}

// ---------------------------------------------------------------------------
// colorPickerPopup — internal popup state
// ---------------------------------------------------------------------------

type colorPickerPopup struct {
	original sg.Color
	current  sg.Color
	mode     ColorMode
	trigger  *ColorPicker

	root *sg.Node

	// SV field
	svImage  engine.Image
	svSprite *sg.Node
	svHue    float64 // [0,1] hue that the SV field is generated for
	svMarker *sg.Node

	// Hue bar
	hueSprite *sg.Node
	hueMarker *sg.Node

	// Alpha bar
	alphaSprite *sg.Node
	alphaMarker *sg.Node

	// Preview
	previewNew *sg.Node
	previewOld *sg.Node

	// Mode tabs + input area
	modeTabs       []*Button
	inputArea      *sg.Node
	hexInput       *TextInput
	fieldInputs    []*NumberStepper
	fieldLabels    []*Label
	updatingInputs bool // guard: true while programmatically setting input values

	// Theme dimensions
	popupW      float64
	svFieldSize float64
	hueBarH     float64
	alphaBarH   float64
	padding     render.Insets
}

func newColorPickerPopup(trigger *ColorPicker) *colorPickerPopup {
	return &colorPickerPopup{
		original: trigger.value,
		current:  trigger.value,
		mode:     trigger.defaultMode,
		trigger:  trigger,
	}
}

func (p *colorPickerPopup) build() {
	th := p.trigger.EffectiveTheme()
	group := th.ColorPicker.Group(p.trigger.Variant())

	p.popupW = group.PopupWidth
	if p.popupW <= 0 {
		p.popupW = 440
	}
	p.svFieldSize = group.SVFieldSize
	if p.svFieldSize <= 0 {
		p.svFieldSize = 200
	}
	p.hueBarH = group.HueBarHeight
	if p.hueBarH <= 0 {
		p.hueBarH = 14
	}
	p.alphaBarH = group.AlphaBarHeight
	if p.alphaBarH <= 0 {
		p.alphaBarH = 14
	}
	p.padding = group.Padding
	if p.padding.IsAuto() {
		p.padding = render.Insets{Top: 10, Right: 10, Bottom: 10, Left: 10}
	}

	// Compute initial HSV (normalize for overbright tint values).
	h, s, v, _ := colorutil.ToHSV(colorutil.NormalizeRGB(p.current))
	p.svHue = h

	// Calculate layout
	innerW := p.popupW - p.padding.Left - p.padding.Right
	previewH := 40.0
	previewW := 60.0
	svSize := p.svFieldSize
	if svSize > innerW-previewW-8 {
		svSize = innerW - previewW - 8
	}

	tabH := 28.0
	inputH := 28.0
	inputRowGap := 4.0
	gap := 6.0

	// Channel inputs use a 2-column grid: ceil(count/2) rows.
	channelCount := 4
	if !p.trigger.showAlpha {
		channelCount = 3
	}
	inputRows := (channelCount + 1) / 2
	inputAreaH := float64(inputRows)*inputH + float64(inputRows-1)*inputRowGap

	btnRowH := 28.0
	totalH := p.padding.Top + svSize + gap + p.hueBarH
	if p.trigger.showAlpha {
		totalH += gap + p.alphaBarH
	}
	totalH += gap + tabH + gap + inputAreaH + gap + btnRowH + p.padding.Bottom

	bgColor := group.Background.Resolve(core.StateDefault).Color

	// Root container
	root := sg.NewContainer("colorpicker-popup")
	root.Interactable = true
	root.HitShape = sg.HitRect{X: 0, Y: 0, Width: p.popupW, Height: totalH}

	// Background
	bg := sg.NewSprite("cp-bg", sg.TextureRegion{})
	bg.SetScale(p.popupW, totalH)
	bg.SetColor(bgColor)
	root.AddChild(bg)

	// Border
	bw := group.BorderWidth
	borderCol := group.BorderColor.Resolve(core.StateDefault)
	if bw > 0 {
		buildBorderSprites(root, "cp-border", p.popupW, totalH, bw, borderCol)
	}

	curY := p.padding.Top
	curX := p.padding.Left

	// ── SV Field ──────────────────────────────────────────────────────────
	svSizeInt := int(svSize)
	if svSizeInt < 2 {
		svSizeInt = 2
	}
	p.svImage = p.generateSVImage(h, svSizeInt)
	p.svSprite = sg.NewSprite("cp-sv", sg.TextureRegion{})
	p.svSprite.SetCustomImage(p.svImage)
	p.svSprite.SetPosition(curX, curY)
	p.svSprite.SetZIndex(2)
	root.AddChild(p.svSprite)

	// SV hit node for drag — positioned at the field origin so LocalX/LocalY
	// in pointer callbacks are directly in field-local coordinates [0, svSize].
	svHit := sg.NewContainer("cp-sv-hit")
	svHit.Interactable = true
	svHit.SetPosition(curX, curY)
	svHit.HitShape = sg.HitRect{X: 0, Y: 0, Width: svSize, Height: svSize}
	svHit.SetZIndex(5)
	svFieldX := curX
	svFieldY := curY
	svHit.OnPointerDown(func(ctx sg.PointerContext) {
		p.handleSVLocal(ctx.LocalX, ctx.LocalY, svFieldX, svFieldY, svSize)
	})
	svHit.OnDragStart(func(ctx sg.DragContext) {
		p.handleSVLocal(ctx.LocalX, ctx.LocalY, svFieldX, svFieldY, svSize)
	})
	svHit.OnDrag(func(ctx sg.DragContext) {
		p.handleSVLocal(ctx.LocalX, ctx.LocalY, svFieldX, svFieldY, svSize)
	})
	root.AddChild(svHit)

	// SV marker (small square indicator)
	p.svMarker = sg.NewSprite("cp-sv-marker", sg.TextureRegion{})
	p.svMarker.SetScale(6, 6)
	p.svMarker.SetColor(sg.RGBA(1, 1, 1, 1))
	p.svMarker.SetZIndex(6)
	root.AddChild(p.svMarker)
	p.updateSVMarker(s, v, curX, curY, svSize)

	// ── Preview ───────────────────────────────────────────────────────────
	pvX := curX + svSize + 8
	pvY := curY

	// New color preview
	p.previewNew = sg.NewSprite("cp-preview-new", sg.TextureRegion{})
	p.previewNew.SetPosition(pvX, pvY)
	p.previewNew.SetScale(previewW, previewH/2)
	p.previewNew.SetColor(p.current)
	p.previewNew.SetZIndex(3)
	root.AddChild(p.previewNew)

	// Old color preview
	p.previewOld = sg.NewSprite("cp-preview-old", sg.TextureRegion{})
	p.previewOld.SetPosition(pvX, pvY+previewH/2)
	p.previewOld.SetScale(previewW, previewH/2)
	p.previewOld.SetColor(p.original)
	p.previewOld.SetZIndex(3)
	root.AddChild(p.previewOld)

	curY += svSize + gap

	// ── Hue bar ───────────────────────────────────────────────────────────
	ensureHueBarImage()
	p.hueSprite = sg.NewSprite("cp-hue", sg.TextureRegion{})
	p.hueSprite.SetCustomImage(hueBarImage)
	p.hueSprite.SetPosition(curX, curY)
	p.hueSprite.SetScale(innerW/360.0, p.hueBarH)
	p.hueSprite.SetZIndex(2)
	root.AddChild(p.hueSprite)

	hueHit := sg.NewContainer("cp-hue-hit")
	hueHit.Interactable = true
	hueHit.SetPosition(curX, curY)
	hueHit.HitShape = sg.HitRect{X: 0, Y: 0, Width: innerW, Height: p.hueBarH}
	hueHit.SetZIndex(5)
	hueHit.OnPointerDown(func(ctx sg.PointerContext) {
		p.handleHueLocal(ctx.LocalX, innerW, svSize, svFieldX, svFieldY)
	})
	hueHit.OnDragStart(func(ctx sg.DragContext) {
		p.handleHueLocal(ctx.LocalX, innerW, svSize, svFieldX, svFieldY)
	})
	hueHit.OnDrag(func(ctx sg.DragContext) {
		p.handleHueLocal(ctx.LocalX, innerW, svSize, svFieldX, svFieldY)
	})
	root.AddChild(hueHit)

	// Hue marker
	p.hueMarker = sg.NewSprite("cp-hue-marker", sg.TextureRegion{})
	p.hueMarker.SetScale(2, p.hueBarH)
	p.hueMarker.SetColor(sg.RGBA(1, 1, 1, 1))
	p.hueMarker.SetZIndex(6)
	root.AddChild(p.hueMarker)
	p.updateHueMarker(h, curX, curY, innerW)

	curY += p.hueBarH + gap

	// ── Alpha bar ─────────────────────────────────────────────────────────
	if p.trigger.showAlpha {
		// Alpha gradient — pre-composited over 2D checkerboard pattern.
		p.alphaSprite = sg.NewSprite("cp-alpha", sg.TextureRegion{})
		p.alphaSprite.SetPosition(curX, curY)
		p.alphaSprite.SetZIndex(3)
		root.AddChild(p.alphaSprite)
		p.updateAlphaBar()

		alphaHit := sg.NewContainer("cp-alpha-hit")
		alphaHit.Interactable = true
		alphaHit.SetPosition(curX, curY)
		alphaHit.HitShape = sg.HitRect{X: 0, Y: 0, Width: innerW, Height: p.alphaBarH}
		alphaHit.SetZIndex(5)
		alphaBarX := curX
		alphaBarY := curY
		alphaHit.OnPointerDown(func(ctx sg.PointerContext) {
			p.handleAlphaLocal(ctx.LocalX, innerW, alphaBarX, alphaBarY)
		})
		alphaHit.OnDragStart(func(ctx sg.DragContext) {
			p.handleAlphaLocal(ctx.LocalX, innerW, alphaBarX, alphaBarY)
		})
		alphaHit.OnDrag(func(ctx sg.DragContext) {
			p.handleAlphaLocal(ctx.LocalX, innerW, alphaBarX, alphaBarY)
		})
		root.AddChild(alphaHit)

		// Alpha marker
		p.alphaMarker = sg.NewSprite("cp-alpha-marker", sg.TextureRegion{})
		p.alphaMarker.SetScale(2, p.alphaBarH)
		p.alphaMarker.SetColor(sg.RGBA(1, 1, 1, 1))
		p.alphaMarker.SetZIndex(6)
		root.AddChild(p.alphaMarker)
		p.updateAlphaMarker(p.current.A(), curX, curY, innerW)

		curY += p.alphaBarH + gap
	}

	// ── Mode tabs ─────────────────────────────────────────────────────────
	tabW := innerW / float64(len(colorModeLabels))
	for i, label := range colorModeLabels {
		modeIdx := ColorMode(i)
		btn := NewButton(fmt.Sprintf("cp-tab-%d", i), label, p.trigger.source, p.trigger.displaySize)
		btn.SetSize(tabW, tabH)
		btn.SetPosition(curX+float64(i)*tabW, curY)
		btn.node.SetZIndex(5)
		btn.SetOnClick(func() {
			p.setMode(modeIdx)
		})
		root.AddChild(btn.node)
		p.modeTabs = append(p.modeTabs, btn)
	}
	p.styleModeTabs()

	curY += tabH + gap

	// ── Input area ────────────────────────────────────────────────────────
	p.inputArea = sg.NewContainer("cp-input-area")
	p.inputArea.Interactable = true
	p.inputArea.SetPosition(curX, curY)
	p.inputArea.SetZIndex(5)
	root.AddChild(p.inputArea)

	p.buildModeInputs(innerW, inputH)

	curY += inputAreaH + gap

	// ── OK / Cancel buttons ──────────────────────────────────────────────
	btnW := 80.0
	btnGap := 8.0

	okBtn := NewButton("cp-ok", "OK", p.trigger.source, p.trigger.displaySize)
	okBtn.SetSize(btnW, btnRowH)
	okBtn.SetPosition(curX+innerW-btnW, curY)
	okBtn.SetVariant(Accent)
	okBtn.node.SetZIndex(5)
	okBtn.SetOnClick(func() {
		DefaultColorPickerManager.commitAndClose()
	})
	root.AddChild(okBtn.node)

	cancelBtn := NewButton("cp-cancel", "Cancel", p.trigger.source, p.trigger.displaySize)
	cancelBtn.SetSize(btnW, btnRowH)
	cancelBtn.SetPosition(curX+innerW-btnW*2-btnGap, curY)
	cancelBtn.SetVariant(Neutral)
	cancelBtn.node.SetZIndex(5)
	cancelBtn.SetOnClick(func() {
		DefaultColorPickerManager.cancelAndClose()
	})
	root.AddChild(cancelBtn.node)

	p.root = root
}

// handleSVLocal processes a click/drag on the SV field using node-local coordinates.
func (p *colorPickerPopup) handleSVLocal(localX, localY, fieldX, fieldY, fieldSize float64) {
	s := clamp01(localX / fieldSize)
	v := 1.0 - clamp01(localY/fieldSize)

	p.current = colorutil.FromHSV(p.svHue, s, v, p.current.A())

	p.updateSVMarker(s, v, fieldX, fieldY, fieldSize)
	p.updatePreview()
	p.updateModeInputValues()
	p.notifyChange()
}

// handleHueLocal processes a click/drag on the hue bar using node-local X.
func (p *colorPickerPopup) handleHueLocal(localX, barW, svSize, svFieldX, svFieldY float64) {
	newHue := clamp01(localX / barW)

	_, s, v, _ := colorutil.ToHSV(colorutil.NormalizeRGB(p.current))
	p.svHue = newHue
	p.current = colorutil.FromHSV(newHue, s, v, p.current.A())

	// Regenerate SV field
	p.svImage = p.generateSVImage(newHue, p.svImage.Bounds().Dx())
	p.svSprite.SetCustomImage(p.svImage)

	gap := 6.0
	innerW := p.popupW - p.padding.Left - p.padding.Right
	p.updateHueMarker(newHue, svFieldX, svFieldY+svSize+gap, innerW)
	p.updateSVMarker(s, v, svFieldX, svFieldY, svSize)
	p.updatePreview()
	p.updateAlphaBar()
	p.updateModeInputValues()
	p.notifyChange()
}

// handleAlphaLocal processes a click/drag on the alpha bar using node-local X.
func (p *colorPickerPopup) handleAlphaLocal(localX, barW, barX, barY float64) {
	newAlpha := clamp01(localX / barW)

	p.current = sg.RGBA(p.current.R(), p.current.G(), p.current.B(), newAlpha)

	p.updateAlphaMarker(newAlpha, barX, barY, barW)
	p.updatePreview()
	p.updateModeInputValues()
	p.notifyChange()
}

func (p *colorPickerPopup) generateSVImage(hue float64, size int) engine.Image {
	img := engine.NewImage(size, size)
	for y := 0; y < size; y++ {
		v := 1.0 - float64(y)/float64(size)
		for x := 0; x < size; x++ {
			s := float64(x) / float64(size)
			c := sg.ColorFromHSV(hue, s, v)
			img.Set(x, y, color.NRGBA{
				R: uint8(math.Round(c.R() * 255)),
				G: uint8(math.Round(c.G() * 255)),
				B: uint8(math.Round(c.B() * 255)),
				A: 255,
			})
		}
	}
	return img
}

func (p *colorPickerPopup) updateSVMarker(s, v, fieldX, fieldY, fieldSize float64) {
	mx := fieldX + s*fieldSize - 3
	my := fieldY + (1-v)*fieldSize - 3
	p.svMarker.SetPosition(mx, my)

	if v > 0.5 && s < 0.5 {
		p.svMarker.SetColor(sg.RGBA(0, 0, 0, 1))
	} else {
		p.svMarker.SetColor(sg.RGBA(1, 1, 1, 1))
	}
}

func (p *colorPickerPopup) updateHueMarker(h, barX, barY, barW float64) {
	mx := barX + h*barW - 1
	p.hueMarker.SetPosition(mx, barY)
}

func (p *colorPickerPopup) updateAlphaMarker(a, barX, barY, barW float64) {
	if p.alphaMarker == nil {
		return
	}
	mx := barX + a*barW - 1
	p.alphaMarker.SetPosition(mx, barY)
}

func (p *colorPickerPopup) updateAlphaBar() {
	if p.alphaSprite == nil {
		return
	}
	// Generate a 2D image: checkerboard pattern composited with a horizontal
	// alpha gradient of the current color, so it doesn't rely on GPU alpha blending.
	norm := colorutil.NormalizeRGB(p.current)
	r, g, b := norm.R(), norm.G(), norm.B()
	innerW := p.popupW - p.padding.Left - p.padding.Right
	w := int(innerW)
	h := int(p.alphaBarH)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	img := engine.NewImage(w, h)
	const cL, cD = 180.0 / 255.0, 120.0 / 255.0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			a := float64(x) / float64(w)
			// 4-pixel checkerboard tiles.
			base := cL
			if (x/4+y/4)%2 != 0 {
				base = cD
			}
			cr := r*a + base*(1-a)
			cg := g*a + base*(1-a)
			cb := b*a + base*(1-a)
			img.Set(x, y, color.NRGBA{
				R: uint8(math.Round(cr * 255)),
				G: uint8(math.Round(cg * 255)),
				B: uint8(math.Round(cb * 255)),
				A: 255,
			})
		}
	}
	p.alphaSprite.SetCustomImage(img)
	p.alphaSprite.SetColor(sg.RGBA(1, 1, 1, 1))
}

func (p *colorPickerPopup) updatePreview() {
	p.previewNew.SetColor(p.current)
}

func (p *colorPickerPopup) notifyChange() {
	trigger := p.trigger
	trigger.value = p.current
	if trigger.valueRef != nil {
		trigger.valueRef.Set(p.current)
	}
	trigger.updateSwatchDisplay()
	if trigger.onChange != nil {
		trigger.onChange(p.current)
	}
}

func (p *colorPickerPopup) setMode(mode ColorMode) {
	p.mode = mode
	p.styleModeTabs()
	p.buildModeInputs(p.popupW-p.padding.Left-p.padding.Right, 28)
}

func (p *colorPickerPopup) styleModeTabs() {
	for i, btn := range p.modeTabs {
		if ColorMode(i) == p.mode {
			btn.SetVariant(Accent)
		} else {
			btn.SetVariant(Neutral)
		}
	}
}

// channelDef describes a single NumberStepper channel field.
type channelDef struct {
	label    string
	min, max float64
	step     float64
	decimals int
	maxLen   int // max characters in the text input (0 = no limit)
}

func (p *colorPickerPopup) buildModeInputs(innerW, inputH float64) {
	// Clear existing input area children
	p.inputArea.RemoveChildren()
	p.fieldInputs = nil
	p.fieldLabels = nil
	p.hexInput = nil

	switch p.mode {
	case ColorModeHex:
		p.buildHexInput(innerW, inputH)
	case ColorModeRGB:
		defs := []channelDef{
			{"R", 0, 255, 1, 0, 3}, {"G", 0, 255, 1, 0, 3}, {"B", 0, 255, 1, 0, 3}, {"A", 0, 255, 1, 0, 3},
		}
		p.buildChannelInputs(innerW, inputH, defs, p.rgbValues, p.onRGBChange)
	case ColorModeHSV:
		defs := []channelDef{
			{"H", 0, 360, 1, 0, 3}, {"S", 0, 100, 1, 0, 3}, {"V", 0, 100, 1, 0, 3}, {"A", 0, 100, 1, 0, 3},
		}
		p.buildChannelInputs(innerW, inputH, defs, p.hsvValues, p.onHSVChange)
	case ColorModeHSL:
		defs := []channelDef{
			{"H", 0, 360, 1, 0, 3}, {"S", 0, 100, 1, 0, 3}, {"L", 0, 100, 1, 0, 3}, {"A", 0, 100, 1, 0, 3},
		}
		p.buildChannelInputs(innerW, inputH, defs, p.hslValues, p.onHSLChange)
	case ColorModeFloat:
		defs := []channelDef{
			{"R", 0, 5, 0.05, 2, 4}, {"G", 0, 5, 0.05, 2, 4}, {"B", 0, 5, 0.05, 2, 4}, {"A", 0, 5, 0.05, 2, 4},
		}
		p.buildChannelInputs(innerW, inputH, defs, p.floatValues, p.onFloatChange)
	}
}

func (p *colorPickerPopup) buildHexInput(innerW, inputH float64) {
	lbl := NewLabel("cp-hex-lbl", "#", p.trigger.source, p.trigger.displaySize)
	lbl.SetPosition(0, (inputH-p.trigger.displaySize)/2)
	p.inputArea.AddChild(lbl.node)

	ti := NewTextInput("cp-hex-input", p.trigger.source, p.trigger.displaySize)
	ti.SetSize(innerW-20, inputH)
	ti.SetPosition(16, 0)
	ti.node.SetZIndex(2)
	ti.SetMaxLength(6)

	hexStr := colorutil.FormatHex(p.current)
	if p.trigger.showAlpha && p.current.A() < 1 {
		hexStr = colorutil.FormatHexA(p.current)
	}
	ti.SetValue(strings.TrimPrefix(hexStr, "#"))

	applyHex := func(text string) {
		text = strings.TrimSpace(text)
		if !strings.HasPrefix(text, "#") {
			text = "#" + text
		}
		if c, ok := colorutil.ParseHex(text); ok {
			if !p.trigger.showAlpha {
				c = sg.RGBA(c.R(), c.G(), c.B(), p.current.A())
			}
			p.current = c
			p.syncFromColor()
		}
	}

	ti.SetOnChange(func(text string) {
		if p.updatingInputs {
			return
		}
		applyHex(text)
	})

	ti.SetOnSubmit(func(text string) {
		applyHex(text)
		// Reformat to canonical form on submit
		hexStr := colorutil.FormatHex(p.current)
		if p.trigger.showAlpha && p.current.A() < 1 {
			hexStr = colorutil.FormatHexA(p.current)
		}
		ti.SetValue(strings.TrimPrefix(hexStr, "#"))
	})

	p.inputArea.AddChild(ti.node)
	p.hexInput = ti
}

func (p *colorPickerPopup) buildChannelInputs(innerW, inputH float64, defs []channelDef, valuesFn func() []string, onChangeFn func([]string)) {
	count := len(defs)
	if !p.trigger.showAlpha && defs[len(defs)-1].label == "A" {
		count = len(defs) - 1
	}

	// 2-column grid layout.
	cols := 2
	colGap := 8.0
	rowGap := 4.0
	labelW := 14.0
	colW := (innerW - colGap) / float64(cols)
	stepperW := colW - labelW

	vals := valuesFn()

	for i := 0; i < count; i++ {
		def := defs[i]
		col := i % cols
		row := i / cols

		x := float64(col) * (colW + colGap)
		y := float64(row) * (inputH + rowGap)

		lbl := NewLabel(fmt.Sprintf("cp-ch-lbl-%d", i), def.label, p.trigger.source, p.trigger.displaySize)
		lbl.SetPosition(x, y+(inputH-p.trigger.displaySize)/2)
		p.inputArea.AddChild(lbl.node)
		p.fieldLabels = append(p.fieldLabels, lbl)

		ns := NewNumberStepper(fmt.Sprintf("cp-ch-ns-%d", i), p.trigger.source, p.trigger.displaySize)
		ns.SetMin(def.min)
		ns.SetMax(def.max)
		ns.SetStep(def.step)
		ns.SetDecimals(def.decimals)
		ns.SetSize(stepperW, inputH)
		ns.SetPosition(x+labelW, y)
		ns.node.SetZIndex(2)
		if def.maxLen > 0 {
			ns.InputField().SetMaxLength(def.maxLen)
		}
		if i < len(vals) {
			if v, err := strconv.ParseFloat(vals[i], 64); err == nil {
				ns.SetValue(v)
			}
		}

		ns.SetOnChange(func(_ float64) {
			if p.updatingInputs {
				return
			}
			newVals := make([]string, count)
			for j := 0; j < count; j++ {
				newVals[j] = fmt.Sprintf("%.*f", defs[j].decimals, p.fieldInputs[j].Value())
			}
			onChangeFn(newVals)
		})

		p.inputArea.AddChild(ns.node)
		p.fieldInputs = append(p.fieldInputs, ns)
	}
}

func (p *colorPickerPopup) rgbValues() []string {
	r, g, b, a := colorutil.ToRGB255(p.current)
	return []string{strconv.Itoa(r), strconv.Itoa(g), strconv.Itoa(b), strconv.Itoa(a)}
}

func (p *colorPickerPopup) hsvValues() []string {
	_, s, v, a := colorutil.ToHSV(colorutil.NormalizeRGB(p.current))
	return []string{
		strconv.Itoa(int(math.Round(p.svHue * 360))),
		strconv.Itoa(int(math.Round(s * 100))),
		strconv.Itoa(int(math.Round(v * 100))),
		strconv.Itoa(int(math.Round(a * 100))),
	}
}

func (p *colorPickerPopup) hslValues() []string {
	_, s, l, a := colorutil.ToHSL(colorutil.NormalizeRGB(p.current))
	return []string{
		strconv.Itoa(int(math.Round(p.svHue * 360))),
		strconv.Itoa(int(math.Round(s * 100))),
		strconv.Itoa(int(math.Round(l * 100))),
		strconv.Itoa(int(math.Round(a * 100))),
	}
}

func (p *colorPickerPopup) floatValues() []string {
	return []string{
		fmt.Sprintf("%.2f", p.current.R()),
		fmt.Sprintf("%.2f", p.current.G()),
		fmt.Sprintf("%.2f", p.current.B()),
		fmt.Sprintf("%.2f", p.current.A()),
	}
}

func (p *colorPickerPopup) onRGBChange(vals []string) {
	r := parseIntClamped(vals[0], 0, 255)
	g := parseIntClamped(vals[1], 0, 255)
	b := parseIntClamped(vals[2], 0, 255)
	a := int(math.Round(p.current.A() * 255))
	if len(vals) > 3 {
		a = parseIntClamped(vals[3], 0, 255)
	}
	p.current = colorutil.FromRGB255(r, g, b, a)
	p.syncFromColor()
}

func (p *colorPickerPopup) onHSVChange(vals []string) {
	h := float64(parseIntClamped(vals[0], 0, 360)) / 360.0
	s := float64(parseIntClamped(vals[1], 0, 100)) / 100.0
	v := float64(parseIntClamped(vals[2], 0, 100)) / 100.0
	a := p.current.A()
	if len(vals) > 3 {
		a = float64(parseIntClamped(vals[3], 0, 100)) / 100.0
	}
	p.svHue = h
	p.current = colorutil.FromHSV(h, s, v, a)
	p.syncFromColor()
}

func (p *colorPickerPopup) onHSLChange(vals []string) {
	h := float64(parseIntClamped(vals[0], 0, 360)) / 360.0
	s := float64(parseIntClamped(vals[1], 0, 100)) / 100.0
	l := float64(parseIntClamped(vals[2], 0, 100)) / 100.0
	a := p.current.A()
	if len(vals) > 3 {
		a = float64(parseIntClamped(vals[3], 0, 100)) / 100.0
	}
	p.current = colorutil.FromHSL(h, s, l, a)
	p.syncFromColor()
}

func (p *colorPickerPopup) onFloatChange(vals []string) {
	r := parseFloatClamped(vals[0], 0, 5)
	g := parseFloatClamped(vals[1], 0, 5)
	b := parseFloatClamped(vals[2], 0, 5)
	a := p.current.A()
	if len(vals) > 3 {
		a = parseFloatClamped(vals[3], 0, 5)
	}
	p.current = sg.RGBA(r, g, b, a)
	p.syncFromColor()
}

// syncFromColor regenerates the SV field and updates all UI elements from p.current.
func (p *colorPickerPopup) syncFromColor() {
	// Normalize overbright tint values to 0–1 for HSV conversion.
	norm := colorutil.NormalizeRGB(p.current)
	h, s, v, _ := colorutil.ToHSV(norm)
	// Preserve the remembered hue when the color is achromatic (S≈0),
	// because ToHSV returns H=0 for grays, which would snap the gradient to red.
	if s > 0.001 {
		p.svHue = h
	}
	h = p.svHue

	innerW := p.popupW - p.padding.Left - p.padding.Right
	svSize := p.svFieldSize
	previewW := 60.0
	if svSize > innerW-previewW-8 {
		svSize = innerW - previewW - 8
	}

	// Regenerate SV field
	p.svImage = p.generateSVImage(h, p.svImage.Bounds().Dx())
	p.svSprite.SetCustomImage(p.svImage)

	p.updateSVMarker(s, v, p.padding.Left, p.padding.Top, svSize)

	gap := 6.0
	hueBarY := p.padding.Top + svSize + gap
	p.updateHueMarker(h, p.padding.Left, hueBarY, innerW)

	if p.trigger.showAlpha {
		alphaBarY := hueBarY + p.hueBarH + gap
		p.updateAlphaMarker(p.current.A(), p.padding.Left, alphaBarY, innerW)
		p.updateAlphaBar()
	}

	p.updatePreview()
	p.notifyChange()
}

func (p *colorPickerPopup) updateModeInputValues() {
	switch p.mode {
	case ColorModeHex:
		if p.hexInput != nil {
			hexStr := colorutil.FormatHex(p.current)
			if p.trigger.showAlpha && p.current.A() < 1 {
				hexStr = colorutil.FormatHexA(p.current)
			}
			p.updatingInputs = true
			p.hexInput.SetValue(strings.TrimPrefix(hexStr, "#"))
			p.updatingInputs = false
		}
	case ColorModeRGB:
		p.updateFieldValues(p.rgbValues())
	case ColorModeHSV:
		p.updateFieldValues(p.hsvValues())
	case ColorModeHSL:
		p.updateFieldValues(p.hslValues())
	case ColorModeFloat:
		p.updateFieldValues(p.floatValues())
	}
}

func (p *colorPickerPopup) updateFieldValues(vals []string) {
	p.updatingInputs = true
	for i, ns := range p.fieldInputs {
		if i < len(vals) {
			if v, err := strconv.ParseFloat(vals[i], 64); err == nil {
				ns.SetValue(v)
			}
		}
	}
	p.updatingInputs = false
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func parseIntClamped(s string, min, max int) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func parseFloatClamped(s string, min, max float64) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// buildBorderSprites adds 4 edge sprites to a container.
func buildBorderSprites(parent *sg.Node, prefix string, w, h, bw float64, col sg.Color) {
	top := sg.NewSprite(prefix+"-t", sg.TextureRegion{})
	top.SetScale(w, bw)
	top.SetColor(col)
	top.SetZIndex(1)

	bot := sg.NewSprite(prefix+"-b", sg.TextureRegion{})
	bot.SetScale(w, bw)
	bot.SetPosition(0, h-bw)
	bot.SetColor(col)
	bot.SetZIndex(1)

	left := sg.NewSprite(prefix+"-l", sg.TextureRegion{})
	left.SetScale(bw, h-bw*2)
	left.SetPosition(0, bw)
	left.SetColor(col)
	left.SetZIndex(1)

	right := sg.NewSprite(prefix+"-r", sg.TextureRegion{})
	right.SetScale(bw, h-bw*2)
	right.SetPosition(w-bw, bw)
	right.SetColor(col)
	right.SetZIndex(1)

	parent.AddChild(top)
	parent.AddChild(bot)
	parent.AddChild(left)
	parent.AddChild(right)
}
