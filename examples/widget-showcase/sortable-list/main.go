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

	items := ui.NewArrayFrom([]string{
		"Defeat the Dragon",
		"Find the Artifact",
		"Rescue the Prince",
		"Explore the Dungeon",
		"Craft a Sword",
	})

	sl := ui.NewSortableList("quests", 28)
	sl.SetSize(240, 180)
	sl.SetPosition((screenW-240)/2, (screenH-180)/2)

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		lbl := ui.NewLabel("item", fmt.Sprintf("%d. %s", index+1, data.(string)), font, 12)
		return &lbl.Component
	})

	ui.BindSortableListItems(sl, items)
	screen.Add(sl)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "SortableList",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
