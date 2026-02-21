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

	sel := ui.NewSelect("sel", []ui.SelectOption{
		{Label: "Apple"},
		{Label: "Banana"},
		{Label: "Cherry"},
		{Label: "Date"},
	}, font, 14)
	sel.SetSize(160, 28)
	sel.SetPosition((screenW-160)/2, 60)
	screen.Add(sel)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Select",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
