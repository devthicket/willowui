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

	tiles := ui.NewTileList("tiles", 80, 60)
	tiles.SetSize(280, 200)
	tiles.SetColumns(3)
	tiles.SetSelectable(true)
	tiles.SetPosition((screenW-280)/2, (screenH-200)/2)

	tiles.SetRenderItem(func(index int, data any) *ui.Component {
		panel := ui.NewPanel("tile-wrap")
		panel.SetSize(80, 60)
		lbl := ui.NewLabel("tile", data.(string), font, 12)
		lbl.SetPosition(8, 22)
		panel.AddChild(lbl)
		return &panel.Component
	})

	items := make([]ui.ListItem, 9)
	for i := range items {
		items[i] = ui.ListItem{Data: fmt.Sprintf("Tile %d", i+1)}
	}
	tiles.SetItems(items)

	screen.Add(tiles)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "TileList",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
