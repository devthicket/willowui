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

	list := ui.NewList("list", 24)
	list.SetSize(200, 180)
	list.SetSelectable(true)
	list.SetPosition((screenW-200)/2, (screenH-180)/2)

	list.SetRenderItem(func(index int, data any) *ui.Component {
		lbl := ui.NewLabel("item", data.(string), font, 13)
		return &lbl.Component
	})

	items := make([]ui.ListItem, 10)
	for i := range items {
		items[i] = ui.ListItem{Data: fmt.Sprintf("Item %d", i+1)}
	}
	list.SetItems(items)

	screen.Add(list)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "List",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
