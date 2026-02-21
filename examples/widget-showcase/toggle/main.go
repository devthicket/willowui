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
	_ = font

	screen := ui.NewScreen()

	tog := ui.NewToggle("tog")
	tog.SetPosition((screenW-48)/2, (screenH-24)/2)
	screen.Add(tog)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Toggle",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
