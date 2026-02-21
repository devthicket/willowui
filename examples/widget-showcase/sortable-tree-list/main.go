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

	tree := ui.NewSortableTreeList("scene", font, 12)
	tree.SetAllowReparent(true)
	tree.SetItems([]ui.SortableTreeItem{
		{ID: "bg", Label: "Background", ParentID: ""},
		{ID: "chars", Label: "Characters", ParentID: ""},
		{ID: "hero", Label: "Hero", ParentID: "chars"},
		{ID: "ally", Label: "Ally", ParentID: "chars"},
		{ID: "ui", Label: "UI Layer", ParentID: ""},
		{ID: "hud", Label: "HUD", ParentID: "ui"},
	})
	tree.ExpandAll()
	tree.SetSize(240, 190)
	tree.SetPosition((screenW-240)/2, (screenH-190)/2)
	screen.Add(tree)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "SortableTreeList",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
