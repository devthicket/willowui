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

	bar := ui.NewProgressBar("bar")
	bar.SetSize(200, 20)
	bar.SetPosition((screenW-200)/2, (screenH-20)/2)

	var elapsed float64
	bar.Node().OnUpdate = func(dt float64) {
		elapsed += dt
		t := elapsed / 3.0
		if t > 1.0 {
			t = 1.0
		}
		bar.SetValue(t)
	}

	screen.Add(bar)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "ProgressBar",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
