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

	btn := ui.NewButton("btn", "Click Me", font, 14)
	btnW, btnH := 120.0, 40.0
	btn.SetSize(btnW, btnH)
	btn.SetPosition((screenW-btnW)/2, (screenH-btnH)/2)
	btn.SetOnClick(func() {})
	screen.Add(btn)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Button",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
