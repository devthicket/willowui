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

	mi := ui.NewMaskedInput("phone", font, 14)
	mi.SetMask("(999) 999-9999")
	mi.SetMaskPlaceholder('_')
	mi.SetPlaceholder("Phone number")
	mi.SetWidth(200)
	mi.SetPosition((screenW-200)/2, (screenH-28)/2)
	screen.Add(mi)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "MaskedInput",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
