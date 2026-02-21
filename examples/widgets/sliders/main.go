// Sliders - reactive demo.
// Shows Slider.BindValue and ProgressBar.BindValue sharing the same
// Ref[float64], plus a MeterBar driven via WatchValue, and a second
// slider whose value is constrained by a Computed minimum.
package main

import (
	"fmt"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 800
	screenH  = 500
	colLeft  = 40.0
	colRight = 450.0
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
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	title := willow.NewText("title", "Reactive - Sliders & Progress Bars", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	y := 62.0

	// ── 1. Slider + ProgressBar share Ref[float64] ───────────────────────────
	addHeader(screen, font, sizeSmall, "Slider.BindValue + ProgressBar.BindValue - same Ref[float64]", colLeft, y)
	y += 20

	volRef := ui.NewRef(0.5)

	slider := ui.NewSlider("vol-slider")
	slider.SetSize(340, 20)
	slider.BindValue(volRef)
	slider.SetPosition(colLeft, y)
	screen.Add(slider)
	y += 28

	pb := ui.NewProgressBar("vol-pb")
	pb.SetSize(340, 16)
	pb.SetShowLabel(true, font, sizeSmall)
	pb.BindValue(volRef)
	pb.SetPosition(colLeft, y)
	screen.Add(pb)

	st1 := addStatus(screen, font, sizeSmall, colRight, y-14)
	ui.WatchValue(volRef, func(_, v float64) {
		st1.SetText(fmt.Sprintf("value: %.3f", v))
	})

	y += 24 + 30

	// ── 2. MeterBar driven via WatchValue ─────────────────────────────────────
	addHeader(screen, font, sizeSmall, "MeterBar - driven by WatchValue on the same Ref", colLeft, y)
	y += 20

	meter := ui.NewMeterBar("meter")
	meter.SetSize(340, 16)
	meter.SetRange(0, 1)
	meter.SetPosition(colLeft, y)
	screen.Add(meter)

	ui.WatchValue(volRef, func(_, v float64) {
		meter.SetValue(v)
	})

	noteLbl := ui.NewLabel("note", "Mirrors the slider above via WatchValue (no BindValue on MeterBar)", font, sizeSmall)
	noteLbl.SetColor(willow.RGBA(0.5, 0.55, 0.6, 1))
	noteLbl.SetPosition(colLeft, y+20)
	screen.Add(noteLbl)

	y += 44 + 30

	// ── 3. Two linked sliders - min/max constraint via Computed ──────────────
	addHeader(screen, font, sizeSmall, "Two sliders - min/max constrained to each other via WatchValue", colLeft, y)
	y += 20

	minRef := ui.NewRef(0.2)
	maxRef := ui.NewRef(0.8)

	// When minRef changes, ensure maxRef stays above it.
	ui.WatchValue(minRef, func(_, v float64) {
		if maxRef.Peek() < v {
			maxRef.Set(v)
		}
	})
	// When maxRef changes, ensure minRef stays below it.
	ui.WatchValue(maxRef, func(_, v float64) {
		if minRef.Peek() > v {
			minRef.Set(v)
		}
	})

	minSlider := ui.NewSlider("min-slider")
	minSlider.SetSize(340, 20)
	minSlider.BindValue(minRef)
	minSlider.SetPosition(colLeft, y)
	screen.Add(minSlider)

	minLbl := ui.NewLabel("min-lbl", "min", font, sizeSmall)
	minLbl.SetColor(willow.RGBA(0.5, 0.7, 1, 1))
	minLbl.SetPosition(colLeft+350, y)
	screen.Add(minLbl)

	y += 28

	maxSlider := ui.NewSlider("max-slider")
	maxSlider.SetSize(340, 20)
	maxSlider.BindValue(maxRef)
	maxSlider.SetPosition(colLeft, y)
	screen.Add(maxSlider)

	maxLbl := ui.NewLabel("max-lbl", "max", font, sizeSmall)
	maxLbl.SetColor(willow.RGBA(1, 0.7, 0.5, 1))
	maxLbl.SetPosition(colLeft+350, y)
	screen.Add(maxLbl)

	st3 := addStatus(screen, font, sizeSmall, colRight, y-14)
	rangeComputed := ui.NewComputed(func() string {
		return fmt.Sprintf("range: [%.2f, %.2f]", minRef.Get(), maxRef.Get())
	})
	ui.WatchEffect(func() {
		st3.SetText(rangeComputed.Get())
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive - Sliders & Progress Bars",
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

func addStatus(screen *ui.Screen, font *willow.FontFamily, fontSize, x, y float64) *ui.Label {
	lbl := ui.NewLabel("status", "...", font, fontSize)
	lbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	lbl.SetPosition(x, y)
	screen.Add(lbl)
	return lbl
}
