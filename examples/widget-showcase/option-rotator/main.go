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

	rot := ui.NewOptionRotator("rot", []string{"Easy", "Normal", "Hard"}, font, 14)
	rot.SetSize(180, 32)
	rot.SetPosition((screenW-180)/2, (screenH-32)/2)
	screen.Add(rot)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "OptionRotator",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
