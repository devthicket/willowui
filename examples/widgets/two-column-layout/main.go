// TwoColumnLayout demonstrates the full feature set:
//   - LeftWidth (fixed pixel left column)
//   - ColumnRatio (proportional split, e.g. 60/40)
//   - LeftAlign / RightAlign (AlignStart, AlignCenter, AlignEnd per column)
//   - Span rows via AddRow(child, nil) — full-width / clear:both
//   - Any UIElement as children: toggles, sliders, text inputs, checkboxes, labels
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 920
	screenH = 680
)

var (
	colorBg      = willow.RGBA(0.10, 0.10, 0.12, 1)
	colorPanel   = willow.RGBA(0.16, 0.18, 0.22, 1)
	colorBorder  = willow.RGBA(0.28, 0.30, 0.35, 1)
	colorText    = willow.RGBA(0.93, 0.93, 0.93, 1)
	colorDim     = willow.RGBA(0.50, 0.55, 0.60, 1)
	colorAccent  = willow.RGBA(0.40, 0.65, 1.00, 1)
	colorGold    = willow.RGBA(0.95, 0.85, 0.50, 1)
	colorDivider = willow.RGBA(0.22, 0.25, 0.30, 1)
	colorGreen   = willow.RGBA(0.40, 0.78, 0.55, 1)
)

type controller struct{ font *willow.FontFamily }

func (c *controller) OnCreate(s *ui.Screen) {
	font := c.font

	addText := func(text string, size float64, col willow.Color, x, y float64) {
		n := willow.NewText(text, text, font)
		n.TextBlock.FontSize = size
		n.TextBlock.Color = col
		n.SetPosition(x, y)
		s.AddNode(n)
	}

	addText("TwoColumnLayout", 22, colorText, 24, 14)
	addText("LeftWidth · ColumnRatio · LeftAlign/RightAlign · Span rows · Any child type", 12, colorDim, 24, 42)

	// ── Panel 1: Settings (LeftWidth, AlignEnd/AlignStart, span rows, mixed children) ──
	p1 := makeSettingsPanel(font)
	p1.SetPosition(24, 68)
	s.Add(p1)

	// ── Panel 2: Stat sheet (ColumnRatio=0.60, AlignStart/AlignEnd, span rows) ──
	p2 := makeStatPanel(font)
	p2.SetPosition(400, 68)
	s.Add(p2)

	// ── Panel 3: Alignment showcase (all three AlignX modes demonstrated) ──
	p3 := makeAlignPanel(font)
	p3.SetPosition(400, 68+p2.Height+16)
	s.Add(p3)
}

func (c *controller) OnUpdate(_ float64) {}
func (c *controller) OnDestroy()         {}

// ── Panel 1: Settings ───────────────────────────────────────────────────────

func makeSettingsPanel(font *willow.FontFamily) *ui.Panel {
	const (
		fs    = 13.0
		formW = 340.0
	)

	statusRef := ui.NewRef("(interact with controls)")

	tl := ui.NewTwoColumnLayout("settings-form")
	tl.LeftWidth = 110 // fixed label column
	tl.Gap = 16
	tl.RowSpacing = 10
	tl.Padding = ui.Insets{Top: 6, Right: 14, Bottom: 6, Left: 14}
	// LeftAlign defaults to AlignEnd (right-align labels)
	// RightAlign defaults to AlignStart (left-align inputs)
	tl.Width = formW

	// ── Span row: section header "Audio" ────────────────────────────────
	tl.AddRow(makeSectionDivider("settings-audio", "Audio", font, fs, formW), nil)

	// Music toggle
	musicLbl := label("music-lbl", "Music", font, fs, colorDim)
	musicTgl := ui.NewToggle("music-tgl")
	musicTgl.SetValue(true)
	musicTgl.SetOnChange(func(v bool) { statusRef.Set(fmt.Sprintf("Music: %v", v)) })
	tl.AddRow(musicLbl, musicTgl)

	// Volume slider
	volLbl := label("vol-lbl", "Volume", font, fs, colorDim)
	volSlider := ui.NewSlider("vol-slider")
	volSlider.SetSize(174, 20)
	volSlider.SetValue(0.7)
	volSlider.SetOnChange(func(v float64) { statusRef.Set(fmt.Sprintf("Volume: %.0f%%", v*100)) })
	tl.AddRow(volLbl, volSlider)

	// Mute checkbox
	muteLbl := label("mute-lbl", "Mute SFX", font, fs, colorDim)
	muteCb := ui.NewCheckbox("mute-cb", "", font, fs)
	muteCb.SetOnChange(func(v bool) { statusRef.Set(fmt.Sprintf("Mute: %v", v)) })
	tl.AddRow(muteLbl, muteCb)

	// ── Span row: section header "Display" ──────────────────────────────
	tl.AddRow(makeSectionDivider("settings-display", "Display", font, fs, formW), nil)

	// Fullscreen toggle
	fsLbl := label("fs-lbl", "Fullscreen", font, fs, colorDim)
	fsTgl := ui.NewToggle("fs-tgl")
	fsTgl.SetOnChange(func(v bool) { statusRef.Set(fmt.Sprintf("Fullscreen: %v", v)) })
	tl.AddRow(fsLbl, fsTgl)

	// Username text input
	nameLbl := label("name-lbl", "Username", font, fs, colorDim)
	nameTi := ui.NewTextInput("name-ti", font, fs)
	nameTi.SetWidth(174)
	nameTi.SetPlaceholder("Enter name...")
	nameTi.SetOnChange(func(v string) { statusRef.Set("Username: " + v) })
	tl.AddRow(nameLbl, nameTi)

	// ── Span row: status bar spanning full width ─────────────────────────
	tl.AddRow(makeStatusBar("settings-status", statusRef, font, 11, formW), nil)

	tl.SizeToContent()
	return wrapPanel("settings-outer", "Settings  ·  LeftWidth=110", font, tl)
}

// ── Panel 2: Character stats ─────────────────────────────────────────────────

func makeStatPanel(font *willow.FontFamily) *ui.Panel {
	const (
		fs    = 13.0
		formW = 280.0
	)

	tl := ui.NewTwoColumnLayout("stats-form")
	tl.ColumnRatio = 0.60 // left col gets 60% of available width
	tl.Gap = 12
	tl.RowSpacing = 9
	tl.Padding = ui.Insets{Top: 6, Right: 14, Bottom: 6, Left: 14}
	tl.LeftAlign = ui.AlignStart // stat names left-aligned
	tl.RightAlign = ui.AlignEnd  // values right-aligned to column edge
	tl.Width = formW

	addStatGroup := func(heading string, stats []struct{ n, v string }) {
		tl.AddRow(makeSectionDivider("stat-"+heading, heading, font, fs, formW), nil)
		for _, s := range stats {
			nameLbl := label(s.n+"-lbl", s.n, font, fs, colorDim)
			valLbl := label(s.n+"-val", s.v, font, fs, colorGold)
			tl.AddRow(nameLbl, valLbl)
		}
	}

	addStatGroup("Core", []struct{ n, v string }{
		{"Strength", "18"}, {"Dexterity", "14"},
		{"Intelligence", "12"}, {"Vitality", "20"},
	})
	addStatGroup("Combat", []struct{ n, v string }{
		{"Attack", "47"}, {"Defense", "31"}, {"Speed", "22"},
	})

	// Span row: total score banner
	tl.AddRow(makeTotalBanner("stats-total", "Total Score: 142", font, fs, formW), nil)

	tl.SizeToContent()
	return wrapPanel("stats-outer", "Stats  ·  ColumnRatio=0.60", font, tl)
}

// ── Panel 3: Alignment modes showcase ────────────────────────────────────────

func makeAlignPanel(font *willow.FontFamily) *ui.Panel {
	const (
		fs    = 12.0
		formW = 280.0
	)

	tl := ui.NewTwoColumnLayout("align-form")
	tl.LeftWidth = 100
	tl.Gap = 12
	tl.RowSpacing = 8
	tl.Padding = ui.Insets{Top: 6, Right: 14, Bottom: 6, Left: 14}
	tl.LeftAlign = ui.AlignCenter  // labels centered within left column
	tl.RightAlign = ui.AlignCenter // values centered within right column
	tl.Width = formW

	tl.AddRow(makeSectionDivider("align-hdr", "Both columns: AlignCenter", font, fs, formW), nil)

	for _, row := range []struct{ l, r string }{
		{"Label A", "Value A"},
		{"Longer Label", "Short"},
		{"X", "A much longer value"},
	} {
		ll := label(row.l+"-l", row.l, font, fs, colorAccent)
		rl := label(row.r+"-r", row.r, font, fs, colorText)
		tl.AddRow(ll, rl)
	}

	tl.SizeToContent()
	return wrapPanel("align-outer", "Alignment  ·  Center / Center", font, tl)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// label creates a Label with a specific color.
func label(name, text string, font *willow.FontFamily, size float64, col willow.Color) *ui.Label {
	l := ui.NewLabel(name, text, font, size)
	l.SetColor(col)
	return l
}

// makeSectionDivider returns a full-width span panel with a label and thin rule.
func makeSectionDivider(name, text string, font *willow.FontFamily, fs, width float64) *ui.Panel {
	const h = 24.0
	p := ui.NewPanel(name + "-div")
	p.SetSize(width, h)
	p.SetBackground(willow.RGBA(0, 0, 0, 0))

	rule := willow.NewSprite(name+"-rule", willow.TextureRegion{})
	rule.SetScale(width, 1)
	rule.SetColor(colorDivider)
	rule.SetPosition(0, h/2)
	p.AddRawChild(rule)

	lbl := willow.NewText(name+"-lbl", text, font)
	lbl.TextBlock.FontSize = fs - 1
	lbl.TextBlock.Color = colorDim
	lbl.SetPosition(0, 4)
	p.AddRawChild(lbl)

	return p
}

// makeStatusBar returns a full-width span panel showing reactive status text.
func makeStatusBar(name string, ref *ui.Ref[string], font *willow.FontFamily, fs, width float64) *ui.Panel {
	const h = 22.0
	p := ui.NewPanel(name + "-bar")
	p.SetSize(width, h)
	p.SetBackground(willow.RGBA(0.12, 0.14, 0.18, 1))

	lbl := ui.NewLabel(name+"-lbl", "", font, fs)
	lbl.BindText(ref)
	lbl.SetColor(colorGreen)
	lbl.SetPosition(6, 4)
	p.AddChild(lbl)

	return p
}

// makeTotalBanner returns a full-width span panel with highlighted text.
func makeTotalBanner(name, text string, font *willow.FontFamily, fs, width float64) *ui.Panel {
	const h = 26.0
	p := ui.NewPanel(name + "-banner")
	p.SetSize(width, h)
	p.SetBackground(willow.RGBA(0.20, 0.22, 0.28, 1))
	p.SetBorder(colorDivider, 1)

	lbl := willow.NewText(name+"-lbl", text, font)
	lbl.TextBlock.FontSize = fs
	lbl.TextBlock.Color = colorGold
	lbl.SetPosition(8, 6)
	p.AddRawChild(lbl)

	return p
}

// wrapPanel wraps a TwoColumnLayout in a titled outer Panel.
func wrapPanel(name, title string, font *willow.FontFamily, tl *ui.TwoColumnLayout) *ui.Panel {
	outer := ui.NewPanel(name)
	outer.SetBackground(colorPanel)
	outer.SetBorder(colorBorder, 1)

	hdr := willow.NewText(name+"-hdr", title, font)
	hdr.TextBlock.FontSize = 15
	hdr.TextBlock.Color = colorDim
	hdr.SetPosition(14, 10)
	outer.AddRawChild(hdr)

	tl.SetPosition(0, 28)
	outer.AddChild(tl)
	outer.SetSize(tl.Width, tl.Height+28+10)
	return outer
}

func main() {
	font := ui.MustLoadDefaultFont()

	ui.Stage.Add(ui.NewScreen(ui.WithController(&controller{font: font})))
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — TwoColumnLayout",
		Width:      screenW,
		Height:     screenH,
		ClearColor: colorBg,
	})
}
