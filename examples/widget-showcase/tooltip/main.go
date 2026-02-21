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

	btn := ui.NewButton("btn", "Hover me", font, 14)
	btn.SetSize(120, 36)
	btn.SetPosition((screenW-120)/2, (screenH-36)/2)

	tip := ui.NewTooltip("tip")
	tip.ShowDelay = 10
	tip.SetText("This is a tooltip!", font, 12)
	btn.SetTooltip(tip)

	screen.Add(btn)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Tooltip",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
