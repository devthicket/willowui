package main

import (
	"log"
	"time"

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

	cal := ui.NewCalendarSelector("cal", font, 11)
	cal.SetDate(time.Now())
	cal.SetSize(280, 210)
	cal.SetPosition((screenW-280)/2, (screenH-210)/2)
	screen.Add(cal)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Calendar",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
