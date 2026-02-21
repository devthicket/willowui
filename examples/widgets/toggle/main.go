// Toggle - reactive demo.
// Shows Toggle.BindValue(Ref[bool]).
// Two toggles share a Ref so they stay in sync; a third is bound to
// a Computed that inverts the shared value (read-only mirror).
package main

import (
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 800
	screenH  = 360
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

	title := willow.NewText("title", "Reactive - Toggle", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	y := 62.0

	// ── 1. Toggle A and Toggle B share a Ref[bool] ───────────────────────────
	addHeader(screen, font, sizeSmall, "Two toggles share Ref[bool] - flipping one updates the other", colLeft, y)
	y += 22

	sharedRef := ui.NewRef(false)

	// Toggle A
	togA := ui.NewToggle("tog-a")
	togA.BindValue(sharedRef)
	togA.SetPosition(colLeft, y)
	screen.Add(togA)

	lblA := ui.NewLabel("lbl-a", "Toggle A", font, sizeMedium)
	lblA.SetColor(willow.RGBA(0.8, 0.8, 0.8, 1))
	lblA.SetPosition(colLeft+60, y+2)
	screen.Add(lblA)

	y += 36

	// Toggle B - same ref
	togB := ui.NewToggle("tog-b")
	togB.BindValue(sharedRef)
	togB.SetPosition(colLeft, y)
	screen.Add(togB)

	lblB := ui.NewLabel("lbl-b", "Toggle B (same Ref)", font, sizeMedium)
	lblB.SetColor(willow.RGBA(0.8, 0.8, 0.8, 1))
	lblB.SetPosition(colLeft+60, y+2)
	screen.Add(lblB)

	st1 := addStatus(screen, font, sizeSmall, colRight, y-18)
	ui.WatchValue(sharedRef, func(_, v bool) {
		if v {
			st1.SetText("value: on")
		} else {
			st1.SetText("value: off")
		}
	})

	y += 48 + 24

	// ── 2. Programmatic control ───────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Button drives the same Ref - all bound toggles stay in sync", colLeft, y)
	y += 22

	onBtn := ui.NewButton("on", "Force On", font, sizeMedium)
	onBtn.SetSize(110, 36)
	onBtn.SetPosition(colLeft, y)
	screen.Add(onBtn)
	onBtn.SetOnClick(func() {
		sharedRef.Set(true)
	})

	offBtn := ui.NewButton("off", "Force Off", font, sizeMedium)
	offBtn.SetSize(110, 36)
	offBtn.SetPosition(colLeft+120, y)
	screen.Add(offBtn)
	offBtn.SetOnClick(func() {
		sharedRef.Set(false)
	})

	flipBtn := ui.NewButton("flip", "Flip", font, sizeMedium)
	flipBtn.SetSize(80, 36)
	flipBtn.SetPosition(colLeft+240, y)
	screen.Add(flipBtn)
	flipBtn.SetOnClick(func() {
		sharedRef.Set(!sharedRef.Peek())
	})

	y += 36 + 24

	// ── 3. WatchValue - inverted label ───────────────────────────────────────
	addHeader(screen, font, sizeSmall, "WatchValue - label reflects the inverse of sharedRef", colLeft, y)
	y += 22

	invertedLbl := ui.NewLabel("inv-lbl", "inverted: on", font, sizeMedium)
	invertedLbl.SetColor(willow.RGBA(1, 0.7, 0.3, 1))
	invertedLbl.SetPosition(colLeft, y)
	screen.Add(invertedLbl)
	ui.WatchValue(sharedRef, func(_, v bool) {
		if !v {
			invertedLbl.SetText("inverted: on")
		} else {
			invertedLbl.SetText("inverted: off")
		}
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive - Toggle",
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
