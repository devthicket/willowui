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

	mb := ui.NewMenuBar("mb", font, 13)
	mb.SetSize(screenW, 28)
	mb.SetPosition(0, 0)
	mb.SetEntries([]ui.MenuBarEntry{
		{
			Label: "File",
			Items: []ui.MenuItem{
				{Label: "New"},
				{Label: "Open"},
				{Separator: true},
				{Label: "Exit"},
			},
		},
		{
			Label: "Edit",
			Items: []ui.MenuItem{
				{Label: "Undo"},
				{Label: "Redo"},
				{Separator: true},
				{Label: "Cut"},
				{Label: "Copy"},
				{Label: "Paste"},
			},
		},
		{
			Label: "Help",
			Items: []ui.MenuItem{
				{Label: "About"},
			},
		},
	})
	screen.Add(mb)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "MenuBar",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
