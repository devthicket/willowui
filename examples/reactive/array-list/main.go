// array-list demonstrates reactive Array bindings on list widgets.
//
// Patterns shown:
//   - List.BindItems: a live task list — add, remove selected, sort, shuffle, clear
//   - TileList.BindItems: a live tile grid — add, remove last, shuffle
//   - TreeList.BindRoots with ReactiveTreeNode: add/remove nodes at any depth
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 900
	screenH  = 620
	fontSize = 14.0
)

func main() {
	font := ui.MustLoadDefaultFont()

	screen := ui.NewScreen()

	// Title.
	titleNode := willow.NewText("title", "Reactive: Array Bindings: List Widgets", font)
	titleNode.TextBlock.FontSize = 18
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetPosition(24, 12)
	screen.AddNode(titleNode)
	screen.AddNode(divider("top-div", screenW-48, 40))

	// ── Column 1: List.BindItems ─────────────────────────────────────────────
	const (
		c1x   = 24.0
		c2x   = 316.0
		c3x   = 610.0
		listY = 80.0
		listH = 360.0
		btnY  = listY + listH + 8
	)

	sectionLabel(screen, font, "List.BindItems  (tasks)", c1x, 56)

	tasks := ui.NewArrayFrom([]ui.ListItem{
		{Data: "Write unit tests"},
		{Data: "Fix scrollbar flicker"},
		{Data: "Review PR #42"},
		{Data: "Update docs"},
		{Data: "Refactor theme loader"},
	})

	taskList := ui.NewList("task-list", 28)
	taskList.SetSize(272, listH)
	taskList.SetPosition(c1x, listY)
	taskList.SetSelectable(true)
	taskList.SetRenderItem(func(index int, data any) *ui.Component {
		lbl := ui.NewLabel("task-lbl", data.(string), font, fontSize)
		return &lbl.Component
	})
	taskList.BindItems(tasks)
	screen.Add(taskList)

	// Counter label.
	taskCount := ui.NewLabel("task-count", "", font, fontSize-1)
	taskCount.SetColor(willow.RGBA(0.5, 0.6, 0.7, 1))
	taskCount.SetPosition(c1x, btnY-18)
	screen.Add(taskCount)
	taskCounter := func() {
		taskCount.SetText(fmt.Sprintf("%d tasks", tasks.Len()))
	}
	taskCounter()

	// Task name input for adding new items.
	taskInput := ui.NewTextInput("task-input", font, fontSize)
	taskInput.SetPlaceholder("New task name...")
	taskInput.SetWidth(272)
	taskInput.SetPosition(c1x, btnY)
	screen.Add(taskInput)

	// Buttons row 1.
	addTaskBtn := ui.NewButton("add-task", "+ Add", font, fontSize)
	addTaskBtn.SetSize(80, 26)
	addTaskBtn.SetPosition(c1x, btnY+32)
	addTaskBtn.SetOnClick(func() {
		name := taskInput.Value()
		if name == "" {
			name = fmt.Sprintf("Task %d", tasks.Len()+1)
		}
		tasks.Push(ui.ListItem{Data: name})
		taskInput.SetValue("")
		taskCounter()
	})
	screen.Add(addTaskBtn)

	removeTaskBtn := ui.NewButton("remove-task", "Remove", font, fontSize)
	removeTaskBtn.SetSize(90, 26)
	removeTaskBtn.SetPosition(c1x+88, btnY+32)
	removeTaskBtn.SetOnClick(func() {
		idx := taskList.Selected()
		if idx >= 0 && idx < tasks.Len() {
			tasks.Remove(idx)
			taskCounter()
		}
	})
	screen.Add(removeTaskBtn)

	// Buttons row 2.
	sortTasksBtn := ui.NewButton("sort-tasks", "Sort A-Z", font, fontSize)
	sortTasksBtn.SetSize(90, 26)
	sortTasksBtn.SetPosition(c1x, btnY+64)
	sortTasksBtn.SetOnClick(func() {
		ui.ArraySortFold(tasks, func(item ui.ListItem) string { return item.Data.(string) })
	})
	screen.Add(sortTasksBtn)

	shuffleTasksBtn := ui.NewButton("shuffle-tasks", "Shuffle", font, fontSize)
	shuffleTasksBtn.SetSize(80, 26)
	shuffleTasksBtn.SetPosition(c1x+98, btnY+64)
	shuffleTasksBtn.SetOnClick(func() { tasks.Shuffle() })
	screen.Add(shuffleTasksBtn)

	clearTasksBtn := ui.NewButton("clear-tasks", "Clear", font, fontSize)
	clearTasksBtn.SetSize(68, 26)
	clearTasksBtn.SetPosition(c1x+186, btnY+64)
	clearTasksBtn.SetOnClick(func() {
		tasks.Clear()
		taskCounter()
	})
	screen.Add(clearTasksBtn)

	// ── Column 2: TileList.BindItems ─────────────────────────────────────────
	sectionLabel(screen, font, "TileList.BindItems  (tiles)", c2x, 56)

	tileColors := []willow.Color{
		willow.RGBA(0.8, 0.3, 0.3, 1),
		willow.RGBA(0.3, 0.7, 0.4, 1),
		willow.RGBA(0.3, 0.5, 0.9, 1),
		willow.RGBA(0.8, 0.6, 0.2, 1),
		willow.RGBA(0.6, 0.3, 0.8, 1),
		willow.RGBA(0.2, 0.7, 0.8, 1),
	}
	nextColor := 0

	type tileData struct {
		n     int
		color willow.Color
	}
	tiles := ui.NewArray[ui.ListItem]()
	for i := 0; i < 12; i++ {
		tiles.Push(ui.ListItem{Data: tileData{n: i + 1, color: tileColors[i%len(tileColors)]}})
		nextColor = (i + 1) % len(tileColors)
	}
	nextColor = 12 % len(tileColors)

	tileList := ui.NewTileList("tile-list", 72, 72)
	tileList.SetSize(272, listH)
	tileList.SetColumns(0)
	tileList.SetPosition(c2x, listY)
	tileList.SetRenderItem(func(index int, data any) *ui.Component {
		d := data.(tileData)
		c := ui.NewComponent(fmt.Sprintf("tile-%d", d.n))
		c.Width = 68
		c.Height = 68

		bg := willow.NewSprite("tile-bg", willow.TextureRegion{})
		bg.SetColor(d.color)
		bg.SetScale(68, 68)
		c.AddRawChild(bg)

		lbl := ui.NewLabel("tile-n", fmt.Sprintf("%d", d.n), font, fontSize)
		lbl.SetColor(willow.RGBA(1, 1, 1, 1))
		lbl.SetPosition(28, 26)
		c.AddChild(lbl)

		return c
	})
	tileList.BindItems(tiles)
	screen.Add(tileList)

	tileCount := ui.NewLabel("tile-count", "", font, fontSize-1)
	tileCount.SetColor(willow.RGBA(0.5, 0.6, 0.7, 1))
	tileCount.SetPosition(c2x, btnY-18)
	screen.Add(tileCount)
	tileCounter := func() { tileCount.SetText(fmt.Sprintf("%d tiles", tiles.Len())) }
	tileCounter()

	addTileBtn := ui.NewButton("add-tile", "+ Add Tile", font, fontSize)
	addTileBtn.SetSize(110, 26)
	addTileBtn.SetPosition(c2x, btnY+8)
	addTileBtn.SetOnClick(func() {
		n := tiles.Len() + 1
		tiles.Push(ui.ListItem{Data: tileData{n: n, color: tileColors[nextColor]}})
		nextColor = (nextColor + 1) % len(tileColors)
		tileCounter()
	})
	screen.Add(addTileBtn)

	removeLastTileBtn := ui.NewButton("remove-last-tile", "Remove Last", font, fontSize)
	removeLastTileBtn.SetSize(110, 26)
	removeLastTileBtn.SetPosition(c2x+118, btnY+8)
	removeLastTileBtn.SetOnClick(func() {
		if tiles.Len() > 0 {
			tiles.Pop()
			tileCounter()
		}
	})
	screen.Add(removeLastTileBtn)

	shuffleTilesBtn := ui.NewButton("shuffle-tiles", "Shuffle", font, fontSize)
	shuffleTilesBtn.SetSize(110, 26)
	shuffleTilesBtn.SetPosition(c2x, btnY+40)
	shuffleTilesBtn.SetOnClick(func() { tiles.Shuffle() })
	screen.Add(shuffleTilesBtn)

	clearTilesBtn := ui.NewButton("clear-tiles", "Clear", font, fontSize)
	clearTilesBtn.SetSize(110, 26)
	clearTilesBtn.SetPosition(c2x+118, btnY+40)
	clearTilesBtn.SetOnClick(func() {
		tiles.Clear()
		tileCounter()
	})
	screen.Add(clearTilesBtn)

	// ── Column 3: TreeList.BindRoots ─────────────────────────────────────────
	sectionLabel(screen, font, "TreeList.BindRoots  (ReactiveTreeNode)", c3x, 56)

	// Build initial reactive tree.
	src := ui.NewReactiveTreeNode("src/")
	src.Children.Push(ui.NewReactiveTreeNode("main.go"))
	src.Children.Push(ui.NewReactiveTreeNode("config.go"))

	internal := ui.NewReactiveTreeNode("internal/")
	widget := ui.NewReactiveTreeNode("widget/")
	widget.Children.Push(ui.NewReactiveTreeNode("button.go"))
	widget.Children.Push(ui.NewReactiveTreeNode("label.go"))
	internal.Children.Push(widget)
	internal.Children.Push(ui.NewReactiveTreeNode("theme/"))

	roots := ui.NewArrayFrom([]*ui.ReactiveTreeNode{src, internal})

	treeList := ui.NewTreeList("tree-list", 24)
	treeList.SetSize(266, listH)
	treeList.SetPosition(c3x, listY)
	treeList.SetSelectable(true)
	treeList.SetDefaultTextRenderer(font, fontSize, 260, 24)
	treeList.BindRoots(roots)
	treeList.ExpandAll()
	screen.Add(treeList)

	fileCounter := func() string {
		var count func(arr *ui.Array[*ui.ReactiveTreeNode]) int
		count = func(arr *ui.Array[*ui.ReactiveTreeNode]) int {
			n := arr.Len()
			arr.ForEach(func(_ int, rn *ui.ReactiveTreeNode) {
				n += count(rn.Children)
			})
			return n
		}
		return fmt.Sprintf("%d nodes", count(roots))
	}

	treeCount := ui.NewLabel("tree-count", "", font, fontSize-1)
	treeCount.SetColor(willow.RGBA(0.5, 0.6, 0.7, 1))
	treeCount.SetPosition(c3x, btnY-18)
	screen.Add(treeCount)
	updateTreeCount := func() { treeCount.SetText(fileCounter()) }
	updateTreeCount()

	addRootBtn := ui.NewButton("add-root", "+ Add Root", font, fontSize)
	addRootBtn.SetSize(120, 26)
	addRootBtn.SetPosition(c3x, btnY+8)
	addRootBtn.SetOnClick(func() {
		n := fmt.Sprintf("pkg%d/", roots.Len()+1)
		roots.Push(ui.NewReactiveTreeNode(n))
		treeList.ExpandAll()
		updateTreeCount()
	})
	screen.Add(addRootBtn)

	addChildBtn := ui.NewButton("add-child", "+ Add Child", font, fontSize)
	addChildBtn.SetSize(120, 26)
	addChildBtn.SetPosition(c3x+128, btnY+8)
	addChildBtn.SetOnClick(func() {
		sel := treeList.Selected()
		if sel == nil {
			return
		}
		// Find the ReactiveTreeNode matching the selected TreeNode by Data value.
		var find func(arr *ui.Array[*ui.ReactiveTreeNode], name string) *ui.ReactiveTreeNode
		find = func(arr *ui.Array[*ui.ReactiveTreeNode], name string) *ui.ReactiveTreeNode {
			var result *ui.ReactiveTreeNode
			arr.ForEach(func(_ int, rn *ui.ReactiveTreeNode) {
				if result != nil {
					return
				}
				if rn.Data.(string) == name {
					result = rn
					return
				}
				result = find(rn.Children, name)
			})
			return result
		}
		rn := find(roots, sel.Data.(string))
		if rn != nil {
			childName := fmt.Sprintf("file%d.go", rn.Children.Len()+1)
			rn.Children.Push(ui.NewReactiveTreeNode(childName))
			treeList.ExpandAll()
			updateTreeCount()
		}
	})
	screen.Add(addChildBtn)

	removeSelectedBtn := ui.NewButton("remove-node", "Remove", font, fontSize)
	removeSelectedBtn.SetSize(120, 26)
	removeSelectedBtn.SetPosition(c3x, btnY+40)
	removeSelectedBtn.SetOnClick(func() {
		sel := treeList.Selected()
		if sel == nil {
			return
		}
		name := sel.Data.(string)
		// Remove from roots or from any child array.
		var remove func(arr *ui.Array[*ui.ReactiveTreeNode]) bool
		remove = func(arr *ui.Array[*ui.ReactiveTreeNode]) bool {
			idx := arr.FindIndex(func(rn *ui.ReactiveTreeNode) bool {
				return rn.Data.(string) == name
			})
			if idx >= 0 {
				arr.Remove(idx)
				return true
			}
			var found bool
			arr.ForEach(func(_ int, rn *ui.ReactiveTreeNode) {
				if !found {
					found = remove(rn.Children)
				}
			})
			return found
		}
		if remove(roots) {
			treeList.ClearSelection()
			updateTreeCount()
		}
	})
	screen.Add(removeSelectedBtn)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Reactive Array: List Widgets",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.07, 0.08, 0.10, 1),
	})
}

func sectionLabel(screen *ui.Screen, font *willow.FontFamily, text string, x, y float64) {
	n := willow.NewText("sec-lbl", text, font)
	n.TextBlock.FontSize = 12
	n.TextBlock.Color = willow.RGBA(0.45, 0.55, 0.68, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}

func divider(name string, w, y float64) *willow.Node {
	d := willow.NewSprite(name, willow.TextureRegion{})
	d.SetPosition(24, y)
	d.SetScale(w, 1)
	d.SetColor(willow.RGBA(0.18, 0.23, 0.28, 1))
	return d
}
