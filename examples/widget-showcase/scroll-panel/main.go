package main

import (
	"fmt"
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

	sp := ui.NewScrollPanel("scroll")
	sp.SetSize(240, 180)
	sp.SetContentSize(220, 400)
	sp.SetPosition((screenW-240)/2, (screenH-180)/2)

	for i := 0; i < 15; i++ {
		lbl := ui.NewLabel(fmt.Sprintf("item-%d", i), fmt.Sprintf("Item %d", i+1), font, 13)
		lbl.SetPosition(10, float64(i)*26+4)
		sp.AddChild(lbl)
	}

	screen.Add(sp)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "ScrollPanel",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
