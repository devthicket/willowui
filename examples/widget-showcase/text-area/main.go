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

	ta := ui.NewTextArea("ta", font, 13)
	ta.SetSize(240, 120)
	ta.SetPosition((screenW-240)/2, (screenH-120)/2)
	screen.Add(ta)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "TextArea",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
