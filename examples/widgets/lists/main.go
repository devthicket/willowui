// Lists demonstrates WillowUI's list components: flat list with 1000 items
// showing virtualization, tile list grid, and tree list with expand/collapse.
package main

import (
	"fmt"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 600
)

func main() {
	font := ui.MustLoadDefaultFont()

	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	const (
		sizeLarge  = 24.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI: Lists Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	// Layout constants.
	const (
		col1X  = 24.0
		col2X  = 290.0
		row1Y  = 48.0  // section labels
		listY  = 64.0  // list tops
		listH  = 230.0 // list heights (top row)
		row2Y  = 316.0 // second row section labels
		list2Y = 332.0 // second row list tops
		list2H = 250.0 // second row list heights
	)

	// ── Flat List (1000 items) ──────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Flat List (1000 items, virtualized)", col1X, row1Y)

	selLabel := ui.NewLabel("sel-label", "Selected: none", font, sizeSmall)
	selLabel.SetPosition(col1X, listY+listH+4)
	screen.Add(selLabel)

	list := ui.NewList("flat-list", 24)
	list.SetSize(240, listH)
	list.SetPosition(col1X, listY)

	items := make([]ui.ListItem, 1000)
	for i := range items {
		items[i] = ui.ListItem{Data: fmt.Sprintf("Item %d", i)}
	}

	list.SetRenderItem(func(index int, data any) *ui.Component {
		text := data.(string)
		lbl := ui.NewLabel("item-label", text, font, sizeMedium)
		// Label inherits color from theme.
		return &lbl.Component
	})
	list.SetItems(items)

	list.SetOnChange(func(idx int) {
		if idx >= 0 {
			selLabel.SetText(fmt.Sprintf("Selected: %d", idx))
		}
	})
	screen.Add(list)

	// ── Tile List ───────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Tile List (grid layout)", col2X, row1Y)

	tileList := ui.NewTileList("tile-list", 64, 64)
	tileList.SetSize(260, listH)
	tileList.SetColumns(0) // auto-fit
	tileList.SetPosition(col2X, listY)

	tileItems := make([]ui.ListItem, 50)
	for i := range tileItems {
		tileItems[i] = ui.ListItem{Data: fmt.Sprintf("%d", i)}
	}

	tileList.SetRenderItem(func(index int, data any) *ui.Component {
		text := data.(string)
		c := ui.NewComponent("tile-" + text)

		bg := willow.NewSprite("tile-bg", willow.TextureRegion{})
		bg.SetColor(willow.RGBA(0.2, 0.3, 0.5, 1))
		bg.SetScale(60, 60)
		bg.SetPosition(2, 2)
		c.AddRawChild(bg)

		lbl := ui.NewLabel("tile-label", text, font, sizeSmall)
		lbl.SetColor(willow.RGBA(1, 1, 1, 1))
		lbl.SetPosition(24, 24)
		c.AddChild(lbl)

		return c
	})
	tileList.SetItems(tileItems)
	screen.Add(tileList)

	// ── Tree List ───────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Tree List (expand/collapse)", col1X, row2Y)

	treeRoots := []*ui.TreeNode{
		{Data: "Documents", Children: []*ui.TreeNode{
			{Data: "Reports", Children: []*ui.TreeNode{
				{Data: "Q1 Report.pdf"},
				{Data: "Q2 Report.pdf"},
			}},
			{Data: "Letters", Children: []*ui.TreeNode{
				{Data: "Cover Letter.docx"},
			}},
		}},
		{Data: "Pictures", Children: []*ui.TreeNode{
			{Data: "vacation.jpg"},
			{Data: "family.png"},
		}},
		{Data: "Music", Children: []*ui.TreeNode{
			{Data: "song1.mp3"},
			{Data: "song2.mp3"},
			{Data: "song3.mp3"},
		}},
	}

	treeList := ui.NewTreeList("tree-list", 24)
	treeList.SetSize(240, list2H)
	treeList.SetPosition(col1X, list2Y)

	treeList.SetRenderItem(func(node *ui.TreeNode, depth int) *ui.Component {
		text := node.Data.(string)
		prefix := ""
		if len(node.Children) > 0 {
			if treeList.IsExpanded(node) {
				prefix = "[-] "
			} else {
				prefix = "[+] "
			}
		} else {
			prefix = "    "
		}
		for i := 0; i < depth; i++ {
			prefix = "  " + prefix
		}

		c := ui.NewComponent("tree-entry")
		lbl := ui.NewLabel("tree-label", prefix+text, font, sizeMedium)
		// Label inherits color from theme.
		c.AddChild(lbl)

		// Toggle on click for parent nodes.
		if len(node.Children) > 0 {
			n := node
			c.SetInteractable(true)
			c.OnClick(func(ctx willow.ClickContext) {
				treeList.Toggle(n)
			})
		}

		return c
	})
	treeList.SetRoots(treeRoots)
	treeList.ExpandAll()
	screen.Add(treeList)

	// ── TabBar ──────────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Tab Bar", col2X, row2Y)

	tabBar := ui.NewTabBar("tabs", font, sizeMedium)
	tabBar.SetSize(360, list2H)
	tabBar.X = col2X
	tabBar.Y = list2Y

	// Tab pages.
	tabPad := ui.Insets{Top: 16, Left: 16, Right: 16, Bottom: 16}

	p1, _ := tabBar.AddTabPage("General", ui.LayoutVBox, 8, tabPad)
	lbl1 := ui.NewLabel("p1-text", "Content of Tab 1", font, sizeMedium)
	p1.AddChild(lbl1)
	p1.UpdateLayout()

	p2, _ := tabBar.AddTabPage("Settings", ui.LayoutVBox, 8, tabPad)
	lbl2 := ui.NewLabel("p2-text", "Content of Tab 2", font, sizeMedium)
	p2.AddChild(lbl2)
	p2.UpdateLayout()

	p3, _ := tabBar.AddTabPage("About", ui.LayoutVBox, 8, tabPad)
	lbl3 := ui.NewLabel("p3-text", "Content of Tab 3", font, sizeMedium)
	p3.AddChild(lbl3)
	p3.UpdateLayout()
	screen.Add(tabBar)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Lists Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addSectionLabel(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("section", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
