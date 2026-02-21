package main

import (
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 320
	screenH = 240
)

func main() {
	font := ui.MustLoadDefaultFont()
	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	screen := ui.NewScreen()

	bar := ui.NewToggleButtonBar("tbb", font, 14)
	bar.AddButton("Day")
	bar.AddButton("Week")
	bar.AddButton("Month")
	bar.SetSize(240, 32)
	bar.SetPosition((screenW-240)/2, (screenH-32)/2)
	screen.Add(bar)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "ToggleButtonBar",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
