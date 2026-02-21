// Macos demonstrates a frameless OS window with custom macOS-style chrome
// built entirely from WillowUI widgets. The title bar, traffic-light controls,
// and all content use a rounded dark theme that mimics macOS System Settings.
// The layout reflows when the window is resized or maximized.
package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"

	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

const (
	initW    = 900
	initH    = 620
	titleH   = 52 // taller macOS title area for traffic lights
	sidebarW = 240
	cornerR  = 24                    // window corner radius
	tlSize   = 12                    // traffic-light button diameter
	tlY      = (titleH - tlSize) / 2 // vertically centred in title bar
	tlRed    = 14
	tlYellow = tlRed + tlSize + 8
	tlGreen  = tlYellow + tlSize + 8
)

type macosController struct {
	titleBar     *ui.Panel
	appTitle     *ui.Label
	sidebar      *ui.Panel
	content      *ui.Panel
	card1        *ui.Panel
	card2        *ui.Panel
	card3        *ui.Panel
	dispSep      *ui.Panel
	vertSep      *ui.Panel
	brightSlider *ui.Slider
	nightToggle  *ui.Toggle
	toneToggle   *ui.Toggle
	windowBorder *ui.Panel
	curW         int
	lastW        int
	lastH        int
	dragging     bool
	dragOffsetX  int
	dragOffsetY  int
}

func (c *macosController) OnCreate(s *ui.Screen) {
	w, h := ebiten.WindowSize()
	c.lastW = w
	c.lastH = h
	c.relayout(w, h)
}

func (c *macosController) OnUpdate(dt float64) {
	w, h := ebiten.WindowSize()
	if w != c.lastW || h != c.lastH {
		c.lastW = w
		c.lastH = h
		c.relayout(w, h)
	}

	if !ebiten.IsWindowMaximized() {
		mx, my := ebiten.CursorPosition()
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			if !c.dragging {
				if my >= 0 && my < titleH && mx < c.curW-tlSize*3-32 {
					c.dragging = true
					c.dragOffsetX = mx
					c.dragOffsetY = my
				}
			} else {
				wx, wy := ebiten.WindowPosition()
				ebiten.SetWindowPosition(wx+mx-c.dragOffsetX, wy+my-c.dragOffsetY)
			}
		} else {
			c.dragging = false
		}
	} else {
		c.dragging = false
	}
}

func (c *macosController) OnDestroy() {}

func (c *macosController) relayout(w, h int) {
	c.curW = w
	cw := float64(w - sidebarW)

	c.titleBar.SetSize(float64(w), titleH)
	c.appTitle.SetPosition((float64(w)-c.appTitle.Width)/2, (float64(titleH)-c.appTitle.Height)/2)

	c.sidebar.SetSize(sidebarW, float64(h-titleH))

	c.content.SetSize(cw, float64(h-titleH))

	c.card1.SetSize(cw-48, 108)
	c.card2.SetSize(cw-48, 92)
	c.card3.SetSize(cw-48, 140)

	c.brightSlider.SetSize(cw-48-28, 6)

	cardW := cw - 48
	c.nightToggle.SetPosition(cardW-64, 18)
	c.toneToggle.SetPosition(cardW-64, 82)
	c.dispSep.SetSize(cardW-28, 1)

	c.vertSep.SetSize(1, float64(h-titleH))

	c.windowBorder.SetSize(float64(w), float64(h))
}

// makeCircleImage returns a white circle on a transparent background for use
// as nav icon backgrounds. The 1-pixel soft edge gives clean AA rounding.
func makeCircleImage(size int) *ebiten.Image {
	half := float64(size) * 0.5
	r := half - 0.5
	rgba := image.NewRGBA(image.Rect(0, 0, size, size))
	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			dx := float64(px) + 0.5 - half
			dy := float64(py) + 0.5 - half
			dist := math.Sqrt(dx*dx + dy*dy)
			alpha := math.Max(0, math.Min(1, r-dist+1.0))
			rgba.SetRGBA(px, py, color.RGBA{255, 255, 255, uint8(alpha*255 + 0.5)})
		}
	}
	return ebiten.NewImageFromImage(rgba)
}

func main() {
	font := ui.MustLoadDefaultFont()
	fontLg := font
	fontSm := font

	theme, err := ui.LoadThemeRelative("../../_themes/macos.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	ctrl := &macosController{
		curW: initW,
	}
	screen := ui.NewScreen(ui.WithController(ctrl))

	// ── Title bar ─────────────────────────────────────────────────────────────
	titleBar := ui.NewPanel("titlebar")
	titleBar.SetVariant(ui.Custom4)
	titleBar.SetSize(initW, titleH)
	titleBar.SetCornerRadii(cornerR, cornerR, 0, 0)
	screen.Add(titleBar)
	ctrl.titleBar = titleBar

	appTitle := ui.NewLabel("app-title", "System Settings", font, 19.0)
	appTitle.SetSharpness(0.15)
	appTitle.SetPosition((initW-appTitle.Width)/2, (float64(titleH)-appTitle.Height)/2)
	titleBar.AddChild(appTitle)
	ctrl.appTitle = appTitle

	closeBtn := ui.NewButton("tl-close", "", font, 16.0)
	closeBtn.SetSize(tlSize, tlSize)
	closeBtn.SetPosition(tlRed, tlY)
	closeBtn.SetVariant(ui.Custom1)
	closeBtn.SetOnClick(func() { os.Exit(0) })
	titleBar.AddChild(closeBtn)

	minBtn := ui.NewButton("tl-min", "", font, 16.0)
	minBtn.SetSize(tlSize, tlSize)
	minBtn.SetPosition(tlYellow, tlY)
	minBtn.SetVariant(ui.Custom2)
	minBtn.SetOnClick(func() { ebiten.MinimizeWindow() })
	titleBar.AddChild(minBtn)

	maxBtn := ui.NewButton("tl-max", "", font, 16.0)
	maxBtn.SetSize(tlSize, tlSize)
	maxBtn.SetPosition(tlGreen, tlY)
	maxBtn.SetVariant(ui.Custom3)
	maxBtn.SetOnClick(func() {
		if ebiten.IsWindowMaximized() {
			ebiten.RestoreWindow()
		} else {
			ebiten.MaximizeWindow()
		}
	})
	titleBar.AddChild(maxBtn)

	// ── Left sidebar ──────────────────────────────────────────────────────────
	sidebar := ui.NewPanel("sidebar")
	sidebar.SetVariant(ui.Custom5)
	sidebar.SetSize(sidebarW, initH-titleH)
	sidebar.SetPosition(0, titleH)
	sidebar.SetCornerRadii(0, 0, 0, cornerR)
	screen.Add(sidebar)
	ctrl.sidebar = sidebar

	type navItem struct {
		label string
		color willow.Color
	}
	navItems := []navItem{
		{"Wi-Fi", willow.RGBA(0.0, 0.478, 1.0, 1)},
		{"Bluetooth", willow.RGBA(0.0, 0.478, 1.0, 1)},
		{"Network", willow.RGBA(0.5, 0.5, 0.55, 1)},
		{"Sound", willow.RGBA(1.0, 0.18, 0.33, 1)},
		{"Notifications", willow.RGBA(1.0, 0.40, 0.0, 1)},
		{"General", willow.RGBA(0.5, 0.5, 0.55, 1)},
		{"Appearance", willow.RGBA(0.5, 0.5, 0.55, 1)},
		{"Accessibility", willow.RGBA(0.0, 0.478, 1.0, 1)},
		{"Privacy & Security", willow.RGBA(0.5, 0.5, 0.55, 1)},
		{"Desktop & Dock", willow.RGBA(0.5, 0.5, 0.55, 1)},
		{"Displays", willow.RGBA(0.0, 0.478, 1.0, 1)},
		{"Battery", willow.RGBA(0.157, 0.784, 0.251, 1)},
	}

	selectedNav := ui.NewRef(0)

	circImg := makeCircleImage(26)

	type navRow struct {
		highlight *ui.Panel
		label     *ui.Label
	}
	rows := make([]navRow, len(navItems))

	for i, item := range navItems {
		itemY := float64(8 + i*44)
		idx := i

		hl := ui.NewPanel(fmt.Sprintf("nav-hl-%d", i))
		hl.SetSize(sidebarW-16, 42)
		hl.SetPosition(8, itemY)
		hl.SetVariant(ui.Custom1)
		hl.OnClick(func(_ willow.ClickContext) { selectedNav.Set(idx) })
		sidebar.AddChild(hl)

		icon := willow.NewSprite(fmt.Sprintf("nav-icon-%d", i), willow.TextureRegion{})
		icon.CustomImage_ = circImg
		icon.SetScale(1, 1)
		icon.SetPosition(14, itemY+8)
		icon.SetColor(item.color)
		sidebar.AddRawChild(icon)

		lbl := ui.NewLabel(fmt.Sprintf("nav-lbl-%d", i), item.label, font, 18.0)
		lbl.SetVariant(ui.Custom1)
		lbl.SetPosition(48, itemY+12)
		lbl.SetSharpness(0.15)
		sidebar.AddChild(lbl)

		rows[i] = navRow{hl, lbl}
	}

	ui.WatchValue(selectedNav, func(_, sel int) {
		for i, r := range rows {
			if i == sel {
				r.highlight.SetVariant(ui.Custom8)
				r.label.SetVariant(ui.Primary)
			} else {
				r.highlight.SetVariant(ui.Custom1)
				r.label.SetVariant(ui.Custom1)
			}
		}
	})

	// ── Content area ──────────────────────────────────────────────────────────
	initCW := float64(initW - sidebarW)

	content := ui.NewPanel("content")
	content.SetVariant(ui.Custom6)
	content.SetSize(initCW, float64(initH-titleH))
	content.SetPosition(sidebarW, titleH)
	content.SetCornerRadii(0, 0, cornerR, 0)
	screen.Add(content)
	ctrl.content = content

	headingRef := ui.NewRef(navItems[0].label)
	ui.WatchValue(selectedNav, func(_, sel int) {
		headingRef.Set(navItems[sel].label)
	})
	heading := ui.NewLabel("heading", "", font, 26.0)
	heading.SetSharpness(0.15)
	heading.BindText(headingRef)
	heading.SetPosition(32, 20)
	content.AddChild(heading)

	// ── Card 1: Resolution ────────────────────────────────────────────────────
	cardW := initCW - 48
	card1 := ui.NewPanel("card-resolution")
	card1.SetVariant(ui.Primary)
	card1.SetSize(cardW, 108)
	card1.SetPosition(24, 58)
	card1.SetCornerRadii(12, 12, 12, 12)
	content.AddChild(card1)
	ctrl.card1 = card1

	sectionLabel(card1, "res-hdr", "Resolution", fontSm, 14, 14)
	rowLabel(card1, "res-lbl", "Choose how content appears on your display", fontLg, 14, 32)

	selectedScale := ui.NewRef(1)

	var scaleBtns []*ui.Button
	for i, s := range []string{"More Space", "Default", "Larger Text"} {
		b := ui.NewButton(fmt.Sprintf("scale-%d", i), s, fontSm, 17.0)
		b.SetSize(88, 28)
		b.SetPosition(14+float64(i)*98, 68)
		b.SetVariant(ui.Secondary)
		idx := i
		b.SetOnClick(func() { selectedScale.Set(idx) })
		card1.AddChild(b)
		scaleBtns = append(scaleBtns, b)
	}

	ui.WatchValue(selectedScale, func(_, sel int) {
		for i, b := range scaleBtns {
			if i == sel {
				b.SetVariant(ui.Accent)
			} else {
				b.SetVariant(ui.Secondary)
			}
			b.UpdateVisuals()
		}
	})

	// ── Card 2: Brightness ────────────────────────────────────────────────────
	card2 := ui.NewPanel("card-brightness")
	card2.SetVariant(ui.Primary)
	card2.SetSize(cardW, 92)
	card2.SetPosition(24, 176)
	card2.SetCornerRadii(12, 12, 12, 12)
	content.AddChild(card2)
	ctrl.card2 = card2

	sectionLabel(card2, "bright-hdr", "Brightness & Color", fontSm, 14, 14)
	rowLabel(card2, "bright-lbl", "Adjust your display brightness", fontLg, 14, 32)

	brightRef := ui.NewRef(0.65)
	brightSlider := ui.NewSlider("brightness")
	brightSlider.SetSize(cardW-28, 6)
	brightSlider.SetPosition(14, 62)
	brightSlider.BindValue(brightRef)
	card2.AddChild(brightSlider)
	ctrl.brightSlider = brightSlider

	// ── Card 3: Display settings ──────────────────────────────────────────────
	card3 := ui.NewPanel("card-display")
	card3.SetVariant(ui.Primary)
	card3.SetSize(cardW, 140)
	card3.SetPosition(24, 278)
	card3.SetCornerRadii(12, 12, 12, 12)
	content.AddChild(card3)
	ctrl.card3 = card3

	rowLabel(card3, "night-lbl", "Night Shift", fontLg, 14, 18)
	dimLabel(card3, "night-desc", "Automatically warm your display's colors after sunset", fontSm, 14, 38)

	nightToggle := ui.NewToggle("night-toggle")
	nightToggle.SetPosition(cardW-64, 18)
	card3.AddChild(nightToggle)
	ctrl.nightToggle = nightToggle

	// 1px separator inside card3
	dispSep := ui.NewPanel("disp-sep")
	dispSep.SetVariant(ui.Custom7)
	dispSep.SetSize(cardW-28, 1)
	dispSep.SetPosition(14, 70)
	card3.AddChild(dispSep)
	ctrl.dispSep = dispSep

	rowLabel(card3, "tone-lbl", "True Tone", fontLg, 14, 82)
	dimLabel(card3, "tone-desc", "Automatically adapt display to make colors appear consistent", fontSm, 14, 102)

	toneToggle := ui.NewToggle("tone-toggle")
	toneToggle.SetPosition(cardW-64, 82)
	card3.AddChild(toneToggle)
	ctrl.toneToggle = toneToggle

	// ── Action buttons ────────────────────────────────────────────────────────
	colorBtn := ui.NewButton("color-btn", "Color Profile...", fontSm, 17.0)
	colorBtn.SetSize(160, 32)
	colorBtn.SetPosition(28, 432)
	colorBtn.SetVariant(ui.Secondary)
	content.AddChild(colorBtn)

	nightBtn := ui.NewButton("night-btn", "Night Shift...", fontSm, 17.0)
	nightBtn.SetSize(148, 32)
	nightBtn.SetPosition(196, 432)
	nightBtn.SetVariant(ui.Secondary)
	content.AddChild(nightBtn)

	// ── Vertical separator between sidebar and content ─────────────────────────
	vertSep := ui.NewPanel("vsep")
	vertSep.SetVariant(ui.Custom7)
	vertSep.SetSize(1, float64(initH-titleH))
	vertSep.SetPosition(sidebarW, titleH)
	vertSep.SetZIndex(10)
	screen.Add(vertSep)
	ctrl.vertSep = vertSep

	windowBorder := ui.NewPanel("window-border")
	windowBorder.SetVariant(ui.Custom2)
	windowBorder.SetSize(initW, initH)
	windowBorder.SetInteractable(false)
	windowBorder.SetZIndex(200)
	screen.Add(windowBorder)
	ctrl.windowBorder = windowBorder

	ui.Stage.Add(screen)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	ui.Setup(ui.StageConfig{
		Title:      "System Settings",
		Width:      initW,
		Height:     initH,
		ClearColor: willow.RGBA(0, 0, 0, 1),
	})
}

func sectionLabel(parent *ui.Panel, name, text string, font *willow.FontFamily, x, y float64) {
	lbl := ui.NewLabel(name, text, font, 15.0)
	lbl.SetVariant(ui.Secondary)
	lbl.SetSharpness(0.15)
	lbl.SetPosition(x, y)
	parent.AddChild(lbl)
}

func rowLabel(parent *ui.Panel, name, text string, font *willow.FontFamily, x, y float64) {
	lbl := ui.NewLabel(name, text, font, 18.0)
	lbl.SetSharpness(0.15)
	lbl.SetPosition(x, y)
	parent.AddChild(lbl)
}

func dimLabel(parent *ui.Panel, name, text string, font *willow.FontFamily, x, y float64) {
	lbl := ui.NewLabel(name, text, font, 16.0)
	lbl.SetVariant(ui.Secondary)
	lbl.SetSharpness(0.15)
	lbl.SetPosition(x, y)
	parent.AddChild(lbl)
}
