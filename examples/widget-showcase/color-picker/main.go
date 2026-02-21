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

	picker := ui.NewColorPicker("picker", font, 14)
	picker.SetSize(160, 28)
	picker.SetValue(willow.RGBA(0.2, 0.5, 0.8, 1))
	picker.SetPosition((screenW-160)/2, (screenH-28)/2)
	screen.Add(picker)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "ColorPicker",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
