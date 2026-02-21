// ColorPicker - demonstrates all color input modes.
// Shows standalone picker, ref-bound picker, and utility conversions.
package main

import (
	"fmt"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 600
)

func main() {
	font := ui.MustLoadDefaultFont()
	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 14.0
	)

	screen := ui.NewScreen()

	title := willow.NewText("title", "ColorPicker", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	y := 62.0
	x := 40.0

	// ── 1. Standalone picker ─────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Standalone picker with OnCommit callback", x, y)
	y += 22

	picker1 := ui.NewColorPicker("bg-color", font, sizeSmall)
	picker1.SetSize(160, 28)
	picker1.SetValue(willow.RGBA(0.2, 0.4, 0.8, 1))
	picker1.SetShowAlpha(true)
	picker1.SetDefaultMode(ui.ColorModeHex)
	picker1.SetPosition(x, y)
	screen.Add(picker1)

	commitLbl := ui.NewLabel("commit-lbl", "committed: #3366CC", font, sizeSmall)
	commitLbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	commitLbl.SetPosition(x+180, y+4)
	screen.Add(commitLbl)

	picker1.SetOnCommit(func(c willow.Color) {
		commitLbl.SetText(fmt.Sprintf("committed: %s", ui.FormatHexA(c)))
	})

	y += 48

	// ── 2. Ref-bound pickers ─────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Two pickers bound to the same Ref[willow.Color]", x, y)
	y += 22

	colorRef := ui.NewRef(willow.RGBA(1, 0.5, 0, 1))

	picker2 := ui.NewColorPicker("picker-a", font, sizeSmall)
	picker2.SetSize(140, 28)
	picker2.BindValue(colorRef)
	picker2.SetDefaultMode(ui.ColorModeRGB)
	picker2.SetPosition(x, y)
	screen.Add(picker2)

	picker3 := ui.NewColorPicker("picker-b", font, sizeSmall)
	picker3.SetSize(140, 28)
	picker3.BindValue(colorRef)
	picker3.SetDefaultMode(ui.ColorModeHSV)
	picker3.SetPosition(x+160, y)
	screen.Add(picker3)

	refLbl := ui.NewLabel("ref-lbl", "", font, sizeSmall)
	refLbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	refLbl.SetPosition(x+320, y+4)
	screen.Add(refLbl)

	ui.WatchValue(colorRef, func(_, c willow.Color) {
		r, g, b, _ := ui.ToRGB255(c)
		refLbl.SetText(fmt.Sprintf("RGB(%d, %d, %d)", r, g, b))
	})
	// Trigger initial display
	c := colorRef.Peek()
	r, g, b, _ := ui.ToRGB255(c)
	refLbl.SetText(fmt.Sprintf("RGB(%d, %d, %d)", r, g, b))

	y += 48

	// ── 3. No-alpha picker ───────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Alpha disabled (showAlpha=false)", x, y)
	y += 22

	picker4 := ui.NewColorPicker("no-alpha", font, sizeSmall)
	picker4.SetSize(140, 28)
	picker4.SetValue(willow.RGBA(0.9, 0.2, 0.3, 1))
	picker4.SetShowAlpha(false)
	picker4.SetDefaultMode(ui.ColorModeHSL)
	picker4.SetPosition(x, y)
	screen.Add(picker4)

	y += 48

	// ── 4. Float mode picker ─────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Default mode: Float (0-1)", x, y)
	y += 22

	picker5 := ui.NewColorPicker("float-picker", font, sizeSmall)
	picker5.SetSize(140, 28)
	picker5.SetValue(willow.RGBA(0.5, 0.8, 0.3, 0.75))
	picker5.SetDefaultMode(ui.ColorModeFloat)
	picker5.SetPosition(x, y)
	screen.Add(picker5)

	changeLbl := ui.NewLabel("change-lbl", "", font, sizeSmall)
	changeLbl.SetColor(willow.RGBA(0.5, 0.7, 0.9, 1))
	changeLbl.SetPosition(x+160, y+4)
	screen.Add(changeLbl)

	picker5.SetOnChange(func(c willow.Color) {
		h, s, l, _ := ui.ToHSL(c)
		changeLbl.SetText(fmt.Sprintf("HSL(%.0f, %.0f%%, %.0f%%)", h*360, s*100, l*100))
	})

	y += 48

	// ── 5. Utility conversions ───────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Utility conversions (no picker UI)", x, y)
	y += 22

	parsed, ok := ui.ParseHex("#FF8800")
	if ok {
		hex := ui.FormatHex(parsed)
		hexA := ui.FormatHexA(parsed)
		pr, pg, pb, pa := ui.ToRGB255(parsed)
		h, s, v, _ := ui.ToHSV(parsed)

		addInfo(screen, font, sizeSmall, fmt.Sprintf("ParseHex(\"#FF8800\"): %s / %s", hex, hexA), x, y)
		y += 18
		addInfo(screen, font, sizeSmall, fmt.Sprintf("  ToRGB255: (%d, %d, %d, %d)", pr, pg, pb, pa), x, y)
		y += 18
		addInfo(screen, font, sizeSmall, fmt.Sprintf("  ToHSV: (%.0f, %.0f%%, %.0f%%)", h*360, s*100, v*100), x, y)
	}

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "ColorPicker Example",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addHeader(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("hdr", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.45, 0.55, 0.65, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}

func addInfo(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("info", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.7, 0.75, 0.8, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
