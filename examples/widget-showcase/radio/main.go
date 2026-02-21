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

	rg := ui.NewRadio("radio")
	rg.AddOption("Option A", font, 14)
	rg.AddOption("Option B", font, 14)
	rg.AddOption("Option C", font, 14)
	rg.SetPosition(100, 75)
	screen.Add(rg)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Radio",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
