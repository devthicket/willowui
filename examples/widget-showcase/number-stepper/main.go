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

	stepper := ui.NewNumberStepper("stepper", font, 14)
	stepper.SetSize(140, 32)
	stepper.SetMin(0)
	stepper.SetMax(100)
	stepper.SetStep(1)
	stepper.SetPosition((screenW-140)/2, (screenH-32)/2)
	screen.Add(stepper)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "NumberStepper",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
