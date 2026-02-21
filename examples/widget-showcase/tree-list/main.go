package main

import (
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 320
	screenH = 240
	treeW   = 240.0
)

func main() {
	font := ui.MustLoadDefaultFont()
	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	screen := ui.NewScreen()

	tree := ui.NewTreeList("tree", 24)
	tree.SetSize(treeW, 190)
	tree.SetSelectable(true)
	tree.SetPosition((screenW-treeW)/2, (screenH-190)/2)

	tree.SetRenderItem(func(node *ui.TreeNode, depth int) *ui.Component {
		row := ui.NewHBox("row")
		row.Spacing = 4
		row.Align = ui.AlignCenter
		row.Padding = ui.Insets{Left: float64(depth) * ui.TreeToggleSize}

		toggle := ui.NewTreeToggle("toggle", tree, node)
		if toggle != nil {
			row.AddChild(toggle)
		} else {
			spacer := ui.NewComponent("spacer")
			spacer.Width = ui.TreeToggleSize
			spacer.Height = ui.TreeToggleSize
			row.AddChild(spacer)
		}

		lbl := ui.NewLabel("lbl", node.Data.(string), font, 13)
		lbl.SetInteractable(false)
		row.AddChild(lbl)

		if len(node.Children) > 0 {
			n := node
			row.OnClick(func(_ willow.ClickContext) {
				tree.Toggle(n)
			})
		}

		row.Width = treeW - 14
		row.Height = 24
		row.UpdateLayout()
		return row
	})

	tree.SetRoots([]*ui.TreeNode{
		{Data: "Documents", Children: []*ui.TreeNode{
			{Data: "Resume.pdf"},
			{Data: "Notes.txt"},
		}},
		{Data: "Pictures", Children: []*ui.TreeNode{
			{Data: "Vacation", Children: []*ui.TreeNode{
				{Data: "beach.jpg"},
				{Data: "sunset.jpg"},
			}},
			{Data: "avatar.png"},
		}},
		{Data: "Music", Children: []*ui.TreeNode{
			{Data: "playlist.m3u"},
		}},
	})

	screen.Add(tree)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "TreeList",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
