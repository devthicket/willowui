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

	input := ui.NewTextInput("input", font, 14)
	input.SetSize(200, 28)
	input.SetPlaceholder("Type here...")
	input.SetPosition((screenW-200)/2, (screenH-28)/2)
	screen.Add(input)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "TextInput",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
