package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

func main() {
	font := ui.MustLoadDefaultFont()

	title := ui.NewLabel("title", "TreeTable - Entity Inspector", font, 16)
	title.SetColor(willow.RGBA(0.85, 0.88, 0.92, 1))
	title.SetPosition(20, 12)

	table := ui.NewTreeTable("inspector", font, 13)
	table.SetColumns([]ui.TableColumn{
		{Key: "name", Label: "Name", Width: 200, Sortable: true},
		{Key: "type", Label: "Type", Width: 100, Sortable: true},
		{Key: "value", Label: "Value", Width: 150, Sortable: false},
	})
	table.SetRows([]ui.TreeTableRow{
		{
			ID:    "entity",
			Cells: map[string]string{"name": "Entity", "type": "Group"},
			Children: []ui.TreeTableRow{
				{
					ID:    "transform",
					Cells: map[string]string{"name": "Transform", "type": "Object"},
					Children: []ui.TreeTableRow{
						{ID: "pos-x", Cells: map[string]string{"name": "Position X", "type": "float", "value": "10.5"}},
						{ID: "pos-y", Cells: map[string]string{"name": "Position Y", "type": "float", "value": "20.3"}},
						{ID: "pos-z", Cells: map[string]string{"name": "Position Z", "type": "float", "value": "0.0"}},
					},
				},
				{
					ID:    "renderer",
					Cells: map[string]string{"name": "Renderer", "type": "Object"},
					Children: []ui.TreeTableRow{
						{ID: "material", Cells: map[string]string{"name": "Material", "type": "string", "value": "metal"}},
						{ID: "cast-shadow", Cells: map[string]string{"name": "CastShadow", "type": "bool", "value": "true"}},
					},
				},
			},
		},
		{
			ID:    "player",
			Cells: map[string]string{"name": "Player", "type": "Group"},
			Children: []ui.TreeTableRow{
				{ID: "health", Cells: map[string]string{"name": "Health", "type": "int", "value": "100"}},
				{ID: "speed", Cells: map[string]string{"name": "Speed", "type": "float", "value": "5.5"}},
			},
		},
		{
			ID:    "camera",
			Cells: map[string]string{"name": "Camera", "type": "Object"},
			Children: []ui.TreeTableRow{
				{ID: "fov", Cells: map[string]string{"name": "FOV", "type": "float", "value": "60.0"}},
				{ID: "near", Cells: map[string]string{"name": "Near", "type": "float", "value": "0.1"}},
				{ID: "far", Cells: map[string]string{"name": "Far", "type": "float", "value": "1000.0"}},
			},
		},
	})

	table.SetOnRowClick(func(id string) {
		fmt.Printf("Clicked row: %s\n", id)
	})
	table.SetOnRowExpand(func(id string, expanded bool) {
		fmt.Printf("Row %s expanded=%v\n", id, expanded)
	})

	table.SetPosition(20, 40)
	table.SetSize(500, 350)
	table.SetExpanded("entity", true)

	expandAllBtn := ui.NewButton("expand-all", "Expand All", font, 13)
	expandAllBtn.SetSize(100, 30)
	expandAllBtn.SetPosition(20, 400)
	expandAllBtn.OnClick(func(_ willow.ClickContext) { table.ExpandAll() })

	collapseAllBtn := ui.NewButton("collapse-all", "Collapse All", font, 13)
	collapseAllBtn.SetSize(100, 30)
	collapseAllBtn.SetPosition(130, 400)
	collapseAllBtn.OnClick(func(_ willow.ClickContext) { table.CollapseAll() })

	sortNameBtn := ui.NewButton("sort-name", "Sort by Name", font, 13)
	sortNameBtn.SetSize(120, 30)
	sortNameBtn.SetPosition(240, 400)
	sortNameBtn.OnClick(func(_ willow.ClickContext) { table.SetSortColumn("name", ui.SortDirAsc) })

	sortTypeBtn := ui.NewButton("sort-type", "Sort by Type", font, 13)
	sortTypeBtn.SetSize(120, 30)
	sortTypeBtn.SetPosition(370, 400)
	sortTypeBtn.OnClick(func(_ willow.ClickContext) { table.SetSortColumn("type", ui.SortDirAsc) })

	screen := ui.NewScreen()
	screen.Add(title)
	screen.AddNode(table.Node())
	screen.Add(expandAllBtn)
	screen.Add(collapseAllBtn)
	screen.Add(sortNameBtn)
	screen.Add(sortTypeBtn)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "TreeTable Example",
		Width:      560,
		Height:     450,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
