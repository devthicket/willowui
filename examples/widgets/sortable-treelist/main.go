// SortableTreeList example — demonstrates a hierarchical list with drag
// reordering within levels and optional reparenting.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 700
	screenH = 560
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 15.0
		sizeSmall  = 13.0
	)

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "SortableTreeList", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	// -- Scene Tree (reparent enabled) ----------------------------------------

	sceneHeader := willow.NewText("scene-header", "Scene Tree (reparent on)", font)
	sceneHeader.TextBlock.FontSize = sizeMedium
	sceneHeader.TextBlock.Color = willow.RGBA(0.75, 0.85, 1.0, 1)
	sceneHeader.SetPosition(28, 56)
	screen.AddNode(sceneHeader)

	tree := ui.NewSortableTreeList("scene", font, sizeSmall)
	tree.SetAllowReparent(true)
	tree.SetAllowCrossLevel(true)
	tree.SetItems([]ui.SortableTreeItem{
		{ID: "bg", Label: "Background", ParentID: ""},
		{ID: "chars", Label: "Characters", ParentID: ""},
		{ID: "hero", Label: "Hero", ParentID: "chars"},
		{ID: "villain", Label: "Villain", ParentID: "chars"},
		{ID: "npc1", Label: "NPC Guard", ParentID: "chars"},
		{ID: "env", Label: "Environment", ParentID: ""},
		{ID: "tree1", Label: "Oak Tree", ParentID: "env"},
		{ID: "rock1", Label: "Boulder", ParentID: "env"},
		{ID: "ui", Label: "UI Layer", ParentID: ""},
		{ID: "hpbar", Label: "Health Bar", ParentID: "ui"},
		{ID: "minimap", Label: "Minimap", ParentID: "ui"},
	})
	tree.ExpandAll()
	tree.SetSize(260, 380)
	tree.SetPosition(20, 78)
	screen.Add(tree)

	// Status label for scene tree.
	sceneStatus := ui.NewLabel("scene-status", "Drag items to reorder or reparent", font, sizeSmall)
	sceneStatus.SetPosition(20, 466)
	screen.Add(sceneStatus)

	tree.SetOnReorder(func(itemID, newParentID string, newIndex int) {
		parentDesc := "root"
		if newParentID != "" {
			parentDesc = newParentID
		}
		sceneStatus.SetText(fmt.Sprintf("Moved %s -> %s [%d]", itemID, parentDesc, newIndex))
	})

	// -- Layer Panel (reorder only) -------------------------------------------

	layerHeader := willow.NewText("layer-header", "Layer Stack (reorder only)", font)
	layerHeader.TextBlock.FontSize = sizeMedium
	layerHeader.TextBlock.Color = willow.RGBA(1.0, 0.85, 0.55, 1)
	layerHeader.SetPosition(308, 56)
	screen.AddNode(layerHeader)

	layers := ui.NewSortableTreeList("layers", font, sizeSmall)
	layers.SetAllowReparent(false)
	layers.SetAllowCrossLevel(false)
	layers.SetItems([]ui.SortableTreeItem{
		{ID: "fg", Label: "Foreground", ParentID: ""},
		{ID: "fx", Label: "Effects", ParentID: "fg"},
		{ID: "sprites", Label: "Sprites", ParentID: "fg"},
		{ID: "mg", Label: "Midground", ParentID: ""},
		{ID: "tiles", Label: "Tiles", ParentID: "mg"},
		{ID: "decor", Label: "Decorations", ParentID: "mg"},
		{ID: "bg-layer", Label: "Background", ParentID: ""},
		{ID: "sky", Label: "Sky", ParentID: "bg-layer"},
		{ID: "clouds", Label: "Clouds", ParentID: "bg-layer"},
	})
	layers.ExpandAll()
	layers.SetSize(220, 300)
	layers.SetPosition(300, 78)
	screen.Add(layers)

	layerStatus := ui.NewLabel("layer-status", "Ctrl+Up/Down to reorder", font, sizeSmall)
	layerStatus.SetPosition(300, 386)
	screen.Add(layerStatus)

	layers.SetOnReorder(func(itemID, newParentID string, newIndex int) {
		layerStatus.SetText(fmt.Sprintf("Moved %s to position %d", itemID, newIndex))
	})

	// -- Instructions ---------------------------------------------------------

	instrLabel := ui.NewLabel("instructions",
		"Controls:\n"+
			"  Up/Down       Navigate\n"+
			"  Left/Right    Collapse/Expand\n"+
			"  Ctrl+Up/Down  Reorder\n"+
			"  Ctrl+L/R      Indent/Outdent\n"+
			"  Drag row      Reorder/Reparent\n\n"+
			"Left tree: full reparenting\n"+
			"Right tree: reorder only",
		font, sizeSmall)
	instrLabel.SetPosition(548, 78)
	screen.Add(instrLabel)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "SortableTreeList Example",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.1, 0.1, 0.12, 1),
	})
}
