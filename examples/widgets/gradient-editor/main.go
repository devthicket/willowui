// GradientEditor demonstrates the WillowUI GradientEditor widget: an interactive
// editor for horizontal, vertical, and 4-corner bilinear gradients.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 760
	screenH = 720
)

func main() {
	font := ui.MustLoadDefaultFont()

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	titleNode := willow.NewText("title", "WillowUI: GradientEditor", font)
	titleNode.TextBlock.FontSize = 18
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetPosition(24, 16)
	screen.AddNode(titleNode)

	const (
		colLeft  = 24.0
		colRight = 370.0
		colW1    = 330.0 // H/V column width
		colW2    = 360.0 // 4-corner column width
		topY     = 50.0
	)

	// ── Column 1: H/V mode editor ─────────────────────────────────────────────
	sectionLabel(screen, font, "HORIZONTAL / VERTICAL", colLeft, topY)

	hvEditor := ui.NewGradientEditor("hv-editor", font, 12)
	hvEditor.SetAllowedModes(ui.GradientModeH, ui.GradientModeV)
	hvEditor.SetSize(colW1, 150)
	hvEditor.SetPosition(colLeft, topY+18)
	hvEditor.SetValue(ui.Gradient{
		Mode: ui.GradientModeH,
		Colors: ui.GradientColors{
			TopLeft:     willow.RGBA(0.10, 0.20, 0.50, 1),
			TopRight:    willow.RGBA(0.60, 0.80, 1.00, 1),
			BottomRight: willow.RGBA(0.60, 0.80, 1.00, 1),
			BottomLeft:  willow.RGBA(0.10, 0.20, 0.50, 1),
		},
	})

	hvStatusLbl := ui.NewLabel("hv-status", ui.FormatGradientString(hvEditor.Value()), font, 10)
	hvStatusLbl.SetColor(willow.RGBA(0.5, 0.7, 0.5, 1))
	hvStatusLbl.SetPosition(colLeft, topY+18+150+6)
	screen.Add(hvStatusLbl)

	hvEditor.SetOnChange(func(g ui.Gradient) {
		hvStatusLbl.SetText(ui.FormatGradientString(g))
	})
	screen.Add(hvEditor)

	// ── Column 2: 4-corner mode editor ───────────────────────────────────────
	sectionLabel(screen, font, "4-CORNER BILINEAR", colRight, topY)

	fcEditor := ui.NewGradientEditor("fc-editor", font, 12)
	fcEditor.SetAllowedModes(ui.GradientModeFourCorner)
	fcEditor.SetShowModeSelector(false)
	fcEditor.SetSize(colW2, 180)
	fcEditor.SetPosition(colRight, topY+18)
	fcEditor.SetValue(ui.Gradient{
		Mode: ui.GradientModeFourCorner,
		Colors: ui.GradientColors{
			TopLeft:     willow.RGBA(0.8, 0.2, 0.2, 1),
			TopRight:    willow.RGBA(0.2, 0.8, 0.2, 1),
			BottomRight: willow.RGBA(0.2, 0.2, 0.8, 1),
			BottomLeft:  willow.RGBA(0.8, 0.8, 0.2, 1),
		},
	})

	fcStatusLbl := ui.NewLabel("fc-status", ui.FormatGradientString(fcEditor.Value()), font, 10)
	fcStatusLbl.SetColor(willow.RGBA(0.5, 0.7, 0.5, 1))
	fcStatusLbl.SetPosition(colRight, topY+18+180+6)
	screen.Add(fcStatusLbl)

	fcEditor.SetOnChange(func(g ui.Gradient) {
		fcStatusLbl.SetText(ui.FormatGradientString(g))
	})
	screen.Add(fcEditor)

	// ── Full-width: all modes editor ──────────────────────────────────────────
	const allY = 300.0
	sectionLabel(screen, font, "ALL MODES  (H | V | 4-CORNER)", colLeft, allY)

	allEditor := ui.NewGradientEditor("all-editor", font, 12)
	allEditor.SetSize(screenW-48, 150)
	allEditor.SetPosition(colLeft, allY+18)

	allStatusLbl := ui.NewLabel("all-status", ui.FormatGradientString(allEditor.Value()), font, 10)
	allStatusLbl.SetColor(willow.RGBA(0.5, 0.7, 0.5, 1))
	allStatusLbl.SetPosition(colLeft, allY+18+150+6)
	screen.Add(allStatusLbl)

	allEditor.SetOnChange(func(g ui.Gradient) {
		allStatusLbl.SetText(ui.FormatGradientString(g))
	})
	screen.Add(allEditor)

	// ── BindValue demo ────────────────────────────────────────────────────────
	const bindY = 497.0
	sectionLabel(screen, font, "REACTIVE BINDING (BindValue)", colLeft, bindY)

	gradRef := ui.NewRef(ui.DefaultGradient())

	boundEditor := ui.NewGradientEditor("bound-editor", font, 12)
	boundEditor.SetAllowedModes(ui.GradientModeH, ui.GradientModeV)
	boundEditor.SetSize(colW1, 150)
	boundEditor.SetPosition(colLeft, bindY+18)
	boundEditor.BindValue(gradRef)
	screen.Add(boundEditor)

	bindStatusLbl := ui.NewLabel("bind-status", ui.FormatGradientString(gradRef.Peek()), font, 10)
	bindStatusLbl.SetColor(willow.RGBA(0.5, 0.7, 0.5, 1))
	bindStatusLbl.SetPosition(colLeft, bindY+18+150+6)
	screen.Add(bindStatusLbl)

	boundEditor.SetOnChange(func(g ui.Gradient) {
		bindStatusLbl.SetText(ui.FormatGradientString(g))
	})

	// SampleBilinear demo
	g := ui.DefaultGradient()
	center := ui.SampleBilinear(g.Colors, 0.5, 0.5)
	fmt.Printf("DefaultGradient center: R=%.2f G=%.2f B=%.2f\n",
		center.R(), center.G(), center.B())

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "GradientEditor — WillowUI",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.10, 0.10, 0.12, 1),
	})
}

func sectionLabel(screen *ui.Screen, font *willow.FontFamily, text string, x, y float64) {
	node := willow.NewText("sec-"+text, text, font)
	node.TextBlock.FontSize = 11
	node.TextBlock.Color = willow.RGBA(0.45, 0.50, 0.70, 1)
	node.SetPosition(x, y)
	screen.AddNode(node)
}
