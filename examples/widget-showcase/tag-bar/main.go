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

	bar := ui.NewTagBar("tags", font, 13)
	bar.SetSize(260, 36)
	bar.SetPlaceholder("Add tags...")
	bar.SetTags([]string{"Go", "Rust"})
	bar.SetPosition((screenW-260)/2, (screenH-36)/2)
	screen.Add(bar)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "TagBar",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
