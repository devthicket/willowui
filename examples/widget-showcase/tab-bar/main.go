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

	tabs := ui.NewTabBar("tabs", font, 14)
	tabs.SetSize(280, 180)
	tabs.SetPosition((screenW-280)/2, (screenH-180)/2)

	for _, name := range []string{"Tab A", "Tab B", "Tab C"} {
		p := ui.NewComponent("content-" + name)
		p.Width = 280
		p.Height = 140
		lbl := ui.NewLabel("lbl-"+name, name+" content", font, 14)
		lbl.SetPosition(10, 10)
		p.AddChild(lbl)
		tabs.AddTab(name, p)
	}

	screen.Add(tabs)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "TabBar",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
