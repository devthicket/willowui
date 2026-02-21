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

	win := ui.NewWindow("win", "My Window", font, 14)
	win.SetSize(200, 140)
	win.SetPosition(60, 50)

	lbl := ui.NewLabel("win-content", "Drag me around!", font, 12)
	lbl.SetPosition(12, 8)
	win.Body().AddChild(lbl)

	screen.Add(win)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Window",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
