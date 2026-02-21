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

type row struct {
	Name  string
	Role  string
	Level int
}

func main() {
	font := ui.MustLoadDefaultFont()
	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	screen := ui.NewScreen()

	dt := ui.NewDataTable("table", 24)
	dt.SetSize(280, 190)
	dt.SetFont(font, 12)
	dt.SetPosition((screenW-280)/2, (screenH-190)/2)
	dt.SetSelectionMode(ui.SelectionModeSingle)
	dt.SetZebraStriping(true)

	dt.AddColumn(ui.LabelColumn("name", "Name", func(d any) string {
		return d.(row).Name
	}))
	dt.AddColumn(ui.LabelColumn("role", "Role", func(d any) string {
		return d.(row).Role
	}))
	dt.AddColumn(ui.LabelColumn("level", "Lv", func(d any) string {
		return fmt.Sprintf("%d", d.(row).Level)
	}))

	data := ui.NewArray[any]()
	data.Push(row{"Alice", "Mage", 12})
	data.Push(row{"Bob", "Knight", 8})
	data.Push(row{"Carol", "Healer", 15})
	data.Push(row{"Dave", "Rogue", 10})
	data.Push(row{"Eve", "Archer", 7})
	dt.BindItems(data)

	screen.Add(dt)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "DataTable",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
