// TreeList - reactive demo.
// A project file browser: a TreeList shows a fictional project hierarchy with
// expand/collapse support. TreeList.BindSelected binds the selection to a
// Ref[*TreeNode]. A detail panel on the right is driven entirely by WatchEffect
// from that Ref. Expand All / Collapse All buttons demonstrate programmatic
// tree control.
package main

import (
	"fmt"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 860
	screenH = 600
	treeX   = 20.0
	treeY   = 80.0
	treeW   = 320.0
	treeH   = 400.0
	detailX = 360.0
	detailY = 80.0
	detailW = 480.0
)

// fileEntry holds metadata for a file or directory node.
type fileEntry struct {
	Name string
	Kind string // "dir" or "file"
	Size string // human-readable size (files only)
	Desc string
}

func dir(name, desc string) *fileEntry {
	return &fileEntry{Name: name, Kind: "dir", Desc: desc}
}

func file(name, size, desc string) *fileEntry {
	return &fileEntry{Name: name, Kind: "file", Size: size, Desc: desc}
}

func buildTree() []*ui.TreeNode {
	// /project
	//   /cmd
	//     main.go
	//   /internal
	//     /core
	//       scene.go
	//       layout.go
	//     /widget
	//       button.go
	//       label.go
	//       list.go
	//   /docs
	//     index.html
	//     style.css
	//     README.md
	//   go.mod
	//   go.sum
	//   LICENSE

	n := func(entry *fileEntry, children ...*ui.TreeNode) *ui.TreeNode {
		return &ui.TreeNode{Data: entry, Children: children}
	}

	return []*ui.TreeNode{
		n(dir("project", "Root of the project"),
			n(dir("cmd", "CLI entry points"),
				n(file("main.go", "1.2 KB", "Application entry point. Initialises the scene and runs the game loop.")),
			),
			n(dir("internal", "Private implementation packages"),
				n(dir("core", "Scene graph and layout engine"),
					n(file("scene.go", "8.4 KB", "Manages the root node hierarchy, update/draw lifecycle, and camera.")),
					n(file("layout.go", "5.1 KB", "Constraint-based layout pass. Resolves widths, heights, and positions.")),
				),
				n(dir("widget", "UI widget implementations"),
					n(file("button.go", "3.7 KB", "Clickable button with hover, active, and disabled visual states.")),
					n(file("label.go", "2.1 KB", "Single-line and wrapped text label with color and font size control.")),
					n(file("list.go", "9.8 KB", "Virtualized vertical list with scroll, selection, and reactive bind.")),
				),
			),
			n(dir("docs", "Documentation and GitHub Pages site"),
				n(file("index.html", "4.3 KB", "Shell page for the docs site. Loads nav.js and viewer.html.")),
				n(file("style.css", "6.0 KB", "Global stylesheet. Accent color is #a78bfa (violet).")),
				n(file("README.md", "2.8 KB", "Project overview, installation steps, and quick-start example.")),
			),
			n(file("go.mod", "312 B", "Module declaration and dependency versions.")),
			n(file("go.sum", "18.2 KB", "Cryptographic checksums for all transitive dependencies.")),
			n(file("LICENSE", "1.1 KB", "MIT licence -- do whatever you want, just keep the notice.")),
		),
	}
}

func main() {
	font := ui.MustLoadDefaultFont()
	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	title := willow.NewText("title", "Reactive: File Browser (TreeList)", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", float64(screenW))
	div.SetPosition(0, 50)
	screen.AddNode(div)

	// ── Selection Ref ──────────────────────────────────────────────────────────
	selectedRef := ui.NewRef[*ui.TreeNode](nil)

	// ── TreeList ───────────────────────────────────────────────────────────────
	tree := ui.NewTreeList("files", 28)
	tree.SetSize(treeW, treeH)
	tree.SetSelectable(true)
	tree.BindSelected(selectedRef)
	tree.SetPosition(treeX, treeY)
	screen.Add(tree)

	tree.SetRenderItem(func(node *ui.TreeNode, depth int) *ui.Component {
		e := node.Data.(*fileEntry)

		row := ui.NewHBox("tree-row")
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

		lbl := ui.NewLabel("lbl", e.Name, font, sizeSmall)
		lbl.SetColor(willow.RGBA(0.9, 0.9, 0.9, 1))
		lbl.SetInteractable(false)
		row.AddChild(lbl)

		if len(node.Children) > 0 {
			n := node
			row.OnClick(func(_ willow.ClickContext) {
				tree.Toggle(n)
			})
		}

		row.Width = treeW - 14 // leave room for scrollbar
		row.Height = 28
		row.UpdateLayout()
		return row
	})

	roots := buildTree()
	tree.SetRoots(roots)
	// Expand the root by default so the tree isn't completely collapsed.
	if len(roots) > 0 {
		tree.Expand(roots[0])
	}

	// ── Buttons row 1: tree control ───────────────────────────────────────────
	const btnY1 = treeY + treeH + 10

	expandBtn := ui.NewButton("expand-all", "Expand All", font, sizeSmall)
	expandBtn.SetSize(100, 26)
	expandBtn.SetPosition(treeX, btnY1)
	screen.Add(expandBtn)
	expandBtn.SetOnClick(func() { tree.ExpandAll() })

	collapseBtn := ui.NewButton("collapse-all", "Collapse All", font, sizeSmall)
	collapseBtn.SetSize(110, 26)
	collapseBtn.SetPosition(treeX+110, btnY1)
	screen.Add(collapseBtn)
	collapseBtn.SetOnClick(func() { tree.CollapseAll() })

	clearBtn := ui.NewButton("clear-sel", "Clear Selection", font, sizeSmall)
	clearBtn.SetSize(120, 26)
	clearBtn.SetPosition(treeX+230, btnY1)
	screen.Add(clearBtn)
	clearBtn.SetOnClick(func() {
		selectedRef.Set(nil)
	})

	// ── Row 2: leaf-only toggle ───────────────────────────────────────────────
	const btnY2 = btnY1 + 36

	leafToggle := ui.NewToggle("leaf-only")
	leafToggle.SetPosition(treeX, btnY2)
	screen.Add(leafToggle)
	leafToggle.SetOnChange(func(on bool) {
		tree.SetLeafOnlySelection(on)
	})

	leafLbl := ui.NewLabel("leaf-lbl", "Leaf-only selection", font, sizeSmall)
	leafLbl.SetColor(willow.RGBA(0.75, 0.75, 0.75, 1))
	leafLbl.SetPosition(treeX+48+8, btnY2+4)
	screen.Add(leafLbl)

	// ── Detail panel ──────────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Selected Node", detailX, detailY-22)

	detailNameLbl := ui.NewLabel("d-name", "(nothing selected)", font, sizeMedium)
	detailNameLbl.SetColor(willow.RGBA(1, 1, 1, 1))
	detailNameLbl.SetPosition(detailX, detailY)
	screen.Add(detailNameLbl)

	detailKindLbl := ui.NewLabel("d-kind", "", font, sizeSmall)
	detailKindLbl.SetPosition(detailX, detailY+28)
	screen.Add(detailKindLbl)

	addDivider(screen, detailX, detailY+50, detailW)

	addFieldLabel(screen, font, sizeSmall, "Size:", detailX, detailY+64)
	detailSizeLbl := ui.NewLabel("d-size", "--", font, sizeSmall)
	detailSizeLbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	detailSizeLbl.SetPosition(detailX+60, detailY+64)
	screen.Add(detailSizeLbl)

	addDivider(screen, detailX, detailY+88, detailW)

	detailDescLbl := ui.NewLabel("d-desc", "", font, sizeSmall)
	detailDescLbl.SetColor(willow.RGBA(0.72, 0.78, 0.84, 1))
	detailDescLbl.SetWrapWidth(detailW)
	detailDescLbl.SetPosition(detailX, detailY+102)
	screen.Add(detailDescLbl)

	// Flat-index display (Computed) to show reactive derivation.
	addDivider(screen, detailX, detailY+160, detailW)
	addHeader(screen, font, sizeSmall, "Reactive info", detailX, detailY+172)

	nodeInfo := ui.NewComputed(func() string {
		n := selectedRef.Get()
		if n == nil {
			return "selectedRef.Get() → nil"
		}
		e := n.Data.(*fileEntry)
		return fmt.Sprintf("selectedRef.Get() → %q (%s)\nChildren: %d",
			e.Name, e.Kind, len(n.Children))
	})

	infoLbl := ui.NewLabel("info", "", font, sizeSmall)
	infoLbl.SetColor(willow.RGBA(0.5, 0.6, 0.45, 1))
	infoLbl.SetWrapWidth(detailW)
	infoLbl.SetPosition(detailX, detailY+190)
	screen.Add(infoLbl)

	// ── Reactive wiring ───────────────────────────────────────────────────────
	ui.WatchEffect(func() {
		n := selectedRef.Get()
		if n == nil {
			detailNameLbl.SetText("(nothing selected)")
			detailKindLbl.SetText("")
			detailSizeLbl.SetText("--")
			detailDescLbl.SetText("")
			return
		}
		e := n.Data.(*fileEntry)
		detailNameLbl.SetText(e.Name)
		if e.Kind == "dir" {
			detailKindLbl.SetText("Directory")
			detailKindLbl.SetColor(willow.RGBA(0.55, 0.78, 1.0, 1))
			detailSizeLbl.SetText("--")
		} else {
			detailKindLbl.SetText("File")
			detailKindLbl.SetColor(willow.RGBA(0.75, 0.88, 0.65, 1))
			detailSizeLbl.SetText(e.Size)
		}
		detailDescLbl.SetText(e.Desc)
	})

	ui.WatchEffect(func() {
		infoLbl.SetText(nodeInfo.Get())
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive — File Browser (TreeList)",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addHeader(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("hdr", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.45, 0.55, 0.65, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}

func addDivider(screen *ui.Screen, x, y, width float64) {
	d := willow.NewSprite("div", willow.TextureRegion{})
	d.SetPosition(x, y)
	d.SetScale(width, 1)
	d.SetColor(willow.RGBA(0.22, 0.27, 0.32, 1))
	screen.AddNode(d)
}

func addFieldLabel(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("fl", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.46, 0.55, 0.65, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
