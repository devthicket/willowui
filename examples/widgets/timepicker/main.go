// TimePicker - demonstrates 24h and 12h time pickers with optional seconds.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 600
	screenH = 400
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 14.0
	)

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "TimePicker Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	y := 62.0

	// -- 24h mode (no seconds) --
	addHeader(screen, font, sizeSmall, "24h mode (default)", 40, y)
	y += 22

	tp24 := ui.NewTimePicker("time-24h", font, sizeMedium)
	tp24.SetTime(14, 30, 0)
	tp24.SetSize(160, 90)
	tp24.SetPosition(40, y)
	screen.Add(tp24)

	status24 := addStatus(screen, font, sizeSmall, 220, y+35)
	status24.SetText("14:30")
	tp24.SetOnTimeChanged(func(h, m, s int) {
		status24.SetText(fmt.Sprintf("%02d:%02d", h, m))
	})

	y += 110

	// -- 12h mode with seconds --
	addHeader(screen, font, sizeSmall, "12h mode with seconds", 40, y)
	y += 22

	tp12 := ui.NewTimePicker("time-12h", font, sizeMedium)
	tp12.SetFormat(ui.TimeFormat12h)
	tp12.SetShowSeconds(true)
	tp12.SetTime(8, 30, 0)
	tp12.SetSize(300, 90)
	tp12.SetPosition(40, y)
	screen.Add(tp12)

	status12 := addStatus(screen, font, sizeSmall, 360, y+35)
	status12.SetText("08:30:00 AM")
	tp12.SetOnTimeChanged(func(h, m, s int) {
		period := "AM"
		dispH := h % 12
		if dispH == 0 {
			dispH = 12
		}
		if h >= 12 {
			period = "PM"
		}
		status12.SetText(fmt.Sprintf("%02d:%02d:%02d %s", dispH, m, s, period))
	})

	y += 110

	// -- Programmatic control --
	addHeader(screen, font, sizeSmall, "Programmatic control", 40, y)
	y += 22

	setBtn := ui.NewButton("set-noon", "Set Noon", font, sizeMedium)
	setBtn.SetSize(110, 36)
	setBtn.SetPosition(40, y)
	screen.Add(setBtn)
	setBtn.SetOnClick(func() {
		tp24.SetTime(12, 0, 0)
		tp12.SetTime(12, 0, 0)
		status24.SetText("12:00")
		status12.SetText("12:00:00 PM")
	})

	setMidBtn := ui.NewButton("set-midnight", "Set Midnight", font, sizeMedium)
	setMidBtn.SetSize(130, 36)
	setMidBtn.SetPosition(160, y)
	screen.Add(setMidBtn)
	setMidBtn.SetOnClick(func() {
		tp24.SetTime(0, 0, 0)
		tp12.SetTime(0, 0, 0)
		status24.SetText("00:00")
		status12.SetText("12:00:00 AM")
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "TimePicker Demo",
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
