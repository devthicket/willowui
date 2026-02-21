// Win10 demonstrates a frameless OS window with custom Windows 10-style chrome
// built entirely from WillowUI widgets. The title bar, window controls, and all
// content use a flat dark theme that mimics Windows 10 Settings in dark mode.
// The layout reflows when the window is resized or maximized.
package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	initW    = 900
	initH    = 600
	titleH   = 32
	sidebarW = 220
	winBtnW  = 46 // width of each window control button (Win10 standard)
)

type win10Controller struct {
	titleBar     *ui.Panel
	closeBtn     *ui.Button
	maxBtn       *ui.Button
	minBtn       *ui.Button
	sidebar      *ui.Panel
	content      *ui.Panel
	sep1         *ui.Panel
	sep2         *ui.Panel
	sep3         *ui.Panel
	sep4         *ui.Panel
	brightSlider *ui.Slider
	nightToggle  *ui.Toggle
	hdrToggle    *ui.Toggle
	curW         int
	lastW        int
	lastH        int
	dragging     bool
	dragOffsetX  int
	dragOffsetY  int
}

func (c *win10Controller) OnCreate(s *ui.Screen) {
	w, h := ebiten.WindowSize()
	c.lastW = w
	c.lastH = h
	c.relayout(w, h)
}

func (c *win10Controller) OnUpdate(dt float64) {
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
				if my >= 0 && my < titleH && mx < c.curW-winBtnW*3 {
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

func (c *win10Controller) OnDestroy() {}

func (c *win10Controller) relayout(w, h int) {
	c.curW = w
	cw := float64(w - sidebarW)

	c.titleBar.SetSize(float64(w), titleH)
	c.closeBtn.SetPosition(float64(w-winBtnW), 0)
	c.maxBtn.SetPosition(float64(w-winBtnW*2), 0)
	c.minBtn.SetPosition(float64(w-winBtnW*3), 0)

	c.sidebar.SetSize(sidebarW, float64(h-titleH))

	c.content.SetSize(cw, float64(h-titleH))

	c.sep1.SetSize(cw-64, 1)
	c.sep2.SetSize(cw-64, 1)
	c.sep3.SetSize(cw-64, 1)
	c.sep4.SetSize(cw-64, 1)

	c.brightSlider.SetSize(cw-64, 20)
	c.nightToggle.SetPosition(cw-80, 268)
	c.hdrToggle.SetPosition(cw-80, 342)
}

func main() {
	font := ui.MustLoadDefaultFont()
	fontLg := font

	theme, err := ui.LoadThemeRelative("../../_themes/windows.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	ctrl := &win10Controller{
		curW: initW,
	}
	screen := ui.NewScreen(ui.WithController(ctrl))

	// ── Title bar ────────────────────────────────────────────────────────────
	titleBar := ui.NewPanel("titlebar")
	titleBar.SetVariant(ui.Custom4)
	titleBar.SetSize(initW, titleH)
	screen.Add(titleBar)
	ctrl.titleBar = titleBar

	appTitle := ui.NewLabel("app-title", "Settings", font, 17.0)
	appTitle.SetSharpness(0.15)
	appTitle.SetPosition(14, 9)
	titleBar.AddChild(appTitle)

	closeBtn := ui.NewButton("close", "", font, 16.0)
	closeBtn.SetSize(winBtnW, titleH)
	closeBtn.SetPosition(float64(initW-winBtnW), 0)
	closeBtn.SetVariant(ui.Danger)
	closeBtn.SetOnClick(func() { os.Exit(0) })
	titleBar.AddChild(closeBtn)
	addCloseIcon(closeBtn, winBtnW, titleH)
	ctrl.closeBtn = closeBtn

	maxBtn := ui.NewButton("max", "", font, 16.0)
	maxBtn.SetSize(winBtnW, titleH)
	maxBtn.SetPosition(float64(initW-winBtnW*2), 0)
	maxBtn.SetVariant(ui.Neutral)
	maxBtn.SetOnClick(func() {
		if ebiten.IsWindowMaximized() {
			ebiten.RestoreWindow()
		} else {
			ebiten.MaximizeWindow()
		}
	})
	titleBar.AddChild(maxBtn)
	addMaxIcon(maxBtn, winBtnW, titleH)
	ctrl.maxBtn = maxBtn

	minBtn := ui.NewButton("min", "", font, 16.0)
	minBtn.SetSize(winBtnW, titleH)
	minBtn.SetPosition(float64(initW-winBtnW*3), 0)
	minBtn.SetVariant(ui.Neutral)
	minBtn.SetOnClick(func() { ebiten.MinimizeWindow() })
	titleBar.AddChild(minBtn)
	addMinIcon(minBtn, winBtnW, titleH)
	ctrl.minBtn = minBtn

	// ── Left sidebar ─────────────────────────────────────────────────────────
	sidebar := ui.NewPanel("sidebar")
	sidebar.SetVariant(ui.Custom5)
	sidebar.SetSize(sidebarW, initH-titleH)
	sidebar.SetPosition(0, titleH)
	screen.Add(sidebar)
	ctrl.sidebar = sidebar

	navItems := []string{
		"System",
		"Devices",
		"Phone",
		"Network & Internet",
		"Personalization",
		"Apps",
		"Accounts",
		"Time & Language",
		"Gaming",
		"Ease of Access",
		"Privacy",
		"Update & Security",
	}

	selectedNav := ui.NewRef(0)

	type navRow struct {
		highlight *ui.Panel
		accentBar *ui.Panel
		label     *ui.Label
	}
	rows := make([]navRow, len(navItems))

	for i, name := range navItems {
		itemY := float64(8 + i*44)
		idx := i

		hl := ui.NewPanel(fmt.Sprintf("nav-hl-%d", i))
		hl.SetSize(sidebarW, 40)
		hl.SetPosition(0, itemY)
		if i == 0 {
			hl.SetVariant(ui.Custom8)
		} else {
			hl.SetVariant(ui.Custom1)
		}
		hl.OnClick(func(_ willow.ClickContext) { selectedNav.Set(idx) })
		sidebar.AddChild(hl)

		bar := ui.NewPanel(fmt.Sprintf("nav-bar-%d", i))
		bar.SetVariant(ui.Accent)
		bar.SetSize(3, 28)
		bar.SetPosition(0, itemY+6)
		bar.SetVisible(i == 0)
		sidebar.AddChild(bar)

		lbl := ui.NewLabel(fmt.Sprintf("nav-%d", i), name, font, 16.0)
		if i == 0 {
			lbl.SetVariant(ui.Primary)
		} else {
			lbl.SetVariant(ui.Custom1)
		}
		lbl.SetPosition(18, itemY+13)
		sidebar.AddChild(lbl)

		rows[i] = navRow{hl, bar, lbl}
	}

	ui.WatchValue(selectedNav, func(_, sel int) {
		for i, r := range rows {
			active := i == sel
			if active {
				r.highlight.SetVariant(ui.Custom8)
				r.label.SetVariant(ui.Primary)
			} else {
				r.highlight.SetVariant(ui.Custom1)
				r.label.SetVariant(ui.Custom1)
			}
			r.accentBar.SetVisible(active)
		}
	})

	// ── Content area ─────────────────────────────────────────────────────────
	initCW := float64(initW - sidebarW)

	content := ui.NewPanel("content")
	content.SetVariant(ui.Custom6)
	content.SetSize(initCW, float64(initH-titleH))
	content.SetPosition(sidebarW, titleH)
	screen.Add(content)
	ctrl.content = content

	headingRef := ui.NewRef(navItems[0])
	ui.WatchValue(selectedNav, func(_, sel int) {
		headingRef.Set(navItems[sel])
	})
	heading := ui.NewLabel("heading", "", fontLg, 22.0)
	heading.SetSharpness(0.15)
	heading.BindText(headingRef)
	heading.SetPosition(32, 20)
	content.AddChild(heading)

	sectionLabel(content, "scale-hdr", "Scale and layout", font, 32, 68)
	rowLabel(content, "scale-lbl", "Change the size of text, apps, and other items", font, 32, 88)

	selectedScale := ui.NewRef(1)

	var scaleBtns []*ui.Button
	for i, s := range []string{"100%", "125%", "150%", "175%"} {
		b := ui.NewButton(fmt.Sprintf("scale-%d", i), s, font, 16.0)
		b.SetSize(72, 28)
		b.SetPosition(32+float64(i)*80, 114)
		b.SetVariant(ui.Secondary)
		idx := i
		b.SetOnClick(func() { selectedScale.Set(idx) })
		content.AddChild(b)
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

	sep1 := separator(content, "sep1", initCW, 156)
	ctrl.sep1 = sep1

	sectionLabel(content, "bright-hdr", "Brightness and color", font, 32, 170)
	rowLabel(content, "bright-lbl", "Adjust your screen brightness", font, 32, 190)

	brightRef := ui.NewRef(0.65)
	brightSlider := ui.NewSlider("brightness")
	brightSlider.SetSize(initCW-64, 20)
	brightSlider.SetPosition(32, 214)
	brightSlider.BindValue(brightRef)
	content.AddChild(brightSlider)
	ctrl.brightSlider = brightSlider

	sep2 := separator(content, "sep2", initCW, 250)
	ctrl.sep2 = sep2

	rowLabel(content, "night-lbl", "Night light", font, 32, 268)
	dimLabel(content, "night-desc", "Show warmer colors to help you sleep", font, 32, 288)

	nightToggle := ui.NewToggle("night-toggle")
	nightToggle.SetPosition(initCW-80, 268)
	content.AddChild(nightToggle)
	ctrl.nightToggle = nightToggle

	sep3 := separator(content, "sep3", initCW, 324)
	ctrl.sep3 = sep3

	rowLabel(content, "hdr-lbl", "Windows HD Color", font, 32, 342)
	dimLabel(content, "hdr-desc", "Stream HDR video and play HDR games", font, 32, 362)

	hdrToggle := ui.NewToggle("hdr-toggle")
	hdrToggle.SetPosition(initCW-80, 342)
	content.AddChild(hdrToggle)
	ctrl.hdrToggle = hdrToggle

	sep4 := separator(content, "sep4", initCW, 400)
	ctrl.sep4 = sep4

	advBtn := ui.NewButton("adv-btn", "Advanced display settings", font, 16.0)
	advBtn.SetSize(220, 32)
	advBtn.SetPosition(32, 420)
	advBtn.SetVariant(ui.Secondary)
	content.AddChild(advBtn)

	gfxBtn := ui.NewButton("gfx-btn", "Graphics settings", font, 16.0)
	gfxBtn.SetSize(160, 32)
	gfxBtn.SetPosition(260, 420)
	gfxBtn.SetVariant(ui.Secondary)
	content.AddChild(gfxBtn)

	ui.Stage.Add(screen)

	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	ui.Setup(ui.StageConfig{
		Title:      "Settings",
		Width:      initW,
		Height:     initH,
		ClearColor: willow.RGBA(0.118, 0.118, 0.118, 1),
	})
}

// separator adds a 1 px horizontal rule and returns the panel for later resizing.
func separator(parent *ui.Panel, name string, cw, y float64) *ui.Panel {
	p := ui.NewPanel(name)
	p.SetVariant(ui.Custom7)
	p.SetSize(cw-64, 1)
	p.SetPosition(32, y)
	parent.AddChild(p)
	return p
}

func sectionLabel(parent *ui.Panel, name, text string, font *willow.FontFamily, x, y float64) {
	lbl := ui.NewLabel(name, text, font, 15.0)
	lbl.SetVariant(ui.Secondary)
	lbl.SetSharpness(0.15)
	lbl.SetPosition(x, y)
	parent.AddChild(lbl)
}

func rowLabel(parent *ui.Panel, name, text string, font *willow.FontFamily, x, y float64) {
	lbl := ui.NewLabel(name, text, font, 16.0)
	lbl.SetSharpness(0.15)
	lbl.SetPosition(x, y)
	parent.AddChild(lbl)
}

func dimLabel(parent *ui.Panel, name, text string, font *willow.FontFamily, x, y float64) {
	lbl := ui.NewLabel(name, text, font, 15.0)
	lbl.SetVariant(ui.Secondary)
	lbl.SetSharpness(0.15)
	lbl.SetPosition(x, y)
	parent.AddChild(lbl)
}

// ---------------------------------------------------------------------------
// Window control icons (sprite-based, no font glyphs)
// ---------------------------------------------------------------------------

func iconBar(name string, x, y, w, h float64) *willow.Node {
	bar := willow.NewSprite(name, willow.TextureRegion{})
	bar.SetScale(w, h)
	bar.SetPosition(x, y)
	bar.SetColor(willow.RGBA(1, 1, 1, 1))
	return bar
}

func iconBarRotated(name string, cx, cy, w, h, angle float64) *willow.Node {
	pivot := willow.NewContainer(name + "-p")
	pivot.SetPosition(cx, cy)
	pivot.SetRotation(angle)
	bar := willow.NewSprite(name, willow.TextureRegion{})
	bar.SetScale(w, h)
	bar.SetPosition(-w/2, -h/2)
	bar.SetColor(willow.RGBA(1, 1, 1, 1))
	pivot.AddChild(bar)
	return pivot
}

func addMinIcon(btn *ui.Button, bw, bh float64) {
	const barW, barH = 10.0, 1.5
	cx, cy := bw/2, bh/2
	btn.AddRawChild(iconBar("min-bar", cx-barW/2, cy-barH/2, barW, barH))
}

func addMaxIcon(btn *ui.Button, bw, bh float64) {
	const sz, thick = 10.0, 1.5
	cx, cy := bw/2-sz/2, bh/2-sz/2
	btn.AddRawChild(iconBar("max-top", cx, cy, sz, thick))
	btn.AddRawChild(iconBar("max-bot", cx, cy+sz-thick, sz, thick))
	btn.AddRawChild(iconBar("max-lft", cx, cy+thick, thick, sz-thick*2))
	btn.AddRawChild(iconBar("max-rgt", cx+sz-thick, cy+thick, thick, sz-thick*2))
}

func addCloseIcon(btn *ui.Button, bw, bh float64) {
	const barW, barH = 11.0, 1.5
	cx, cy := bw/2, bh/2
	btn.AddRawChild(iconBarRotated("close-b1", cx, cy, barW, barH, math.Pi/4))
	btn.AddRawChild(iconBarRotated("close-b2", cx, cy, barW, barH, -math.Pi/4))
}
