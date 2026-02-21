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

	web := ui.NewStatWeb("stats", font, 11)
	web.SetAxes([]ui.StatAxis{
		{Name: "STR", Min: 0, Max: 100, Value: 75},
		{Name: "AGI", Min: 0, Max: 100, Value: 60},
		{Name: "VIT", Min: 0, Max: 100, Value: 85},
		{Name: "INT", Min: 0, Max: 100, Value: 40},
		{Name: "LCK", Min: 0, Max: 100, Value: 55},
	})
	web.SetEditable(true)
	web.SetSize(200, 200)
	web.SetPosition((screenW-200)/2, (screenH-200)/2)
	screen.Add(web)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "StatWeb",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
