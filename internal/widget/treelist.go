package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// TreeNode represents a node in a hierarchical tree structure.
type TreeNode struct {
	Data     any
	Children []*TreeNode
}

// ReactiveTreeNode is a tree node whose children are a reactive Array.
// Mutations to any Children array anywhere in the subtree are automatically
// reflected in a TreeList bound via BindRoots.
type ReactiveTreeNode struct {
	Data     any
	Children *Array[*ReactiveTreeNode]
}

// NewReactiveTreeNode creates a ReactiveTreeNode with an empty Children array.
func NewReactiveTreeNode(data any) *ReactiveTreeNode {
	return &ReactiveTreeNode{
		Data:     data,
		Children: NewArray[*ReactiveTreeNode](),
	}
}

// flatEntry is an internal representation of a visible tree node with its
// depth level, used after flattening the tree.
type flatEntry struct {
	node  *TreeNode
	depth int
}

// TreeList is a hierarchical list with expand/collapse support. Visible nodes
// are flattened into a list and rendered using the same virtualization as List.
type TreeList struct {
	Component
	viewport   *sg.Node
	content    *sg.Node
	scrollBar  *ScrollBar
	roots      []*TreeNode
	itemHeight float64
	renderItem func(node *TreeNode, depth int) *Component
	expanded   map[*TreeNode]bool
	flatList   []*flatEntry
	scrollPos  *Ref[float64]

	selected          *Ref[*TreeNode]
	selectable        bool
	leafOnlySelection bool
	selHighlight      *sg.Node
	selWatch          WatchHandle
	onChange          func(*TreeNode)

	// Reactive tree binding — populated by BindRoots.
	reactiveRoots *Array[*ReactiveTreeNode]
	tnMap         map[*ReactiveTreeNode]*TreeNode // stable pointer mapping for selection
	stopArray     func()                          // stops all reactive Children watches

	visibleStart int
	visibleEnd   int
	pool         []*Component
	poolIndex    map[int]int

	expandIcon   engine.Image // optional custom expand glyph
	collapseIcon engine.Image // optional custom collapse glyph
}

// NewTreeList creates a new tree list with the given item height.
func NewTreeList(name string, itemHeight float64) *TreeList {
	tl := &TreeList{
		itemHeight: itemHeight,
		expanded:   make(map[*TreeNode]bool),
		scrollPos:  NewRef(0.0),
		selected:   NewRef[*TreeNode](nil),
		poolIndex:  make(map[int]int),
	}
	initComponent(&tl.Component, name)

	tl.initBackground(name)
	tl.bgNode.SetColor(tl.EffectiveTheme().TreeList.Group(tl.Variant()).Background.Resolve(StateDefault).Color)

	// Viewport.
	tl.viewport = sg.NewContainer(name + "-viewport")
	tl.viewport.Interactable = true
	tl.node.AddChild(tl.viewport)

	// Content.
	tl.content = sg.NewContainer(name + "-content")
	tl.content.Interactable = true
	tl.viewport.AddChild(tl.content)

	// ScrollBar.
	tl.scrollBar = NewScrollBar(name + "-scrollbar")
	tl.scrollBar.SetOnChange(func(pos float64) {
		old := tl.scrollPos.Peek()
		tl.scrollPos.Set(pos)
		DefaultScheduler.Flush()
		if pos != old {
			tl.updateVisible()
		}
	})
	tl.scrollBar.AddToNode(tl.node)

	// Auto-update: mouse wheel scrolling.
	tl.node.OnUpdate = func(_ float64) {
		tl.Update()
	}

	// Click on tree list focuses it.
	tl.node.OnPointerDown(func(_ sg.PointerContext) {
		if tl.enabled {
			DefaultFocusManager.SetFocus(&tl.Component)
		}
	})

	tl.scrollBar.parent = &tl.Component
	tl.onThemeChange = func() {
		tl.applyThemeColors()
		tl.scrollBar.applyThemeColors()
		// Re-render pool items so they pick up the new theme (e.g. tree
		// toggle buttons created before the tree list joined the hierarchy).
		tl.clearPool()
		tl.updateVisible()
	}

	// Focus: tree lists participate in tab and spatial nav, intercept arrows.
	tl.enableFocusNavigation()
	tl.InterceptArrows = true
	tl.ConsumeHandledKeys = false // widget's own Update handles the key; only block spatial nav

	tl.onFocusChange = func(focused bool) { tl.applyThemeColors() }
	tl.SetHandleKey(func(key engine.Key) bool {
		n := len(tl.flatList)
		if n == 0 {
			return false
		}
		sel := tl.selected.Peek()
		idx := tl.flatIndexOf(sel)
		switch key {
		case engine.KeyUp:
			return idx > 0
		case engine.KeyDown:
			return idx < n-1
		}
		return false
	})

	// Default size.
	tl.SetSize(200, 300)

	return tl
}

func (tl *TreeList) applyThemeColors() {
	group := tl.EffectiveTheme().TreeList.Group(tl.Variant())
	tl.bgNode.SetColor(group.Background.Resolve(StateDefault).Color)
	tl.state = computeState(tl.enabled, tl.focused, tl.hovered, false)
	tl.applyFocusRing(group.FocusColor.Resolve(tl.state), group.FocusRingWidth)
	tl.updateHighlight()
	tl.MarkDrawDirty()
}

// SetSelectable enables or disables built-in selection highlighting.
// When enabled, clicking a row selects it and a highlight bar is shown.
func (tl *TreeList) SetSelectable(enabled bool) {
	tl.selectable = enabled
	if enabled && tl.selHighlight == nil {
		tl.selHighlight = sg.NewSprite(tl.node.Name+"-sel-hl", sg.TextureRegion{})
		tl.selHighlight.SetVisible(false)
		tl.selHighlight.SetZIndex(999)
		tl.content.AddChild(tl.selHighlight)
	}
	tl.updateHighlight()
}

// Selectable reports whether built-in selection highlighting is enabled.
func (tl *TreeList) Selectable() bool { return tl.selectable }

// Selected returns the currently selected TreeNode, or nil if none.
func (tl *TreeList) Selected() *TreeNode { return tl.selected.Peek() }

// SelectedRef returns the reactive Ref backing the selection.
func (tl *TreeList) SelectedRef() *Ref[*TreeNode] { return tl.selected }

// SetSelected sets the selected node.
func (tl *TreeList) SetSelected(node *TreeNode) {
	old := tl.selected.Peek()
	tl.selected.Set(node)
	DefaultScheduler.Flush()
	if node != old && tl.onChange != nil {
		tl.onChange(node)
	}
	tl.updateHighlight()
	tl.MarkDrawDirty()
}

// ClearSelection deselects the current node.
func (tl *TreeList) ClearSelection() { tl.SetSelected(nil) }

// BindSelected binds the selection to a reactive Ref[*TreeNode].
func (tl *TreeList) BindSelected(ref *Ref[*TreeNode]) {
	tl.selected = ref
	bindRef(&tl.selWatch, ref, tl.SetSelected)
}

// SetOnChange sets the callback invoked when the selection changes.
func (tl *TreeList) SetOnChange(fn func(*TreeNode)) { tl.onChange = fn }

// SetLeafOnlySelection restricts selection to leaf nodes (nodes with no
// children). Parent nodes remain clickable for expand/collapse but cannot
// be selected. Has no effect when selectable is false.
func (tl *TreeList) SetLeafOnlySelection(enabled bool) {
	tl.leafOnlySelection = enabled
	// If the currently selected node is now disallowed, clear it.
	if enabled {
		sel := tl.selected.Peek()
		if sel != nil && len(sel.Children) > 0 {
			tl.SetSelected(nil)
		}
	}
	// Pool items were wired with the old flag — rebuild so click handlers
	// reflect the new mode.
	tl.clearPool()
	tl.updateVisible()
}

// LeafOnlySelection reports whether only leaf nodes can be selected.
func (tl *TreeList) LeafOnlySelection() bool { return tl.leafOnlySelection }

func (tl *TreeList) updateHighlight() {
	if tl.selHighlight == nil {
		return
	}
	sel := tl.selected.Peek()
	if !tl.selectable || sel == nil {
		tl.selHighlight.SetVisible(false)
		return
	}
	idx := -1
	for i, e := range tl.flatList {
		if e.node == sel {
			idx = i
			break
		}
	}
	if idx < 0 {
		tl.selHighlight.SetVisible(false)
		return
	}
	itemW := tl.Width - float64(DefaultScrollBarWidth)
	y := float64(idx) * tl.itemHeight
	tl.selHighlight.SetPosition(0, y)
	tl.selHighlight.SetScale(itemW, tl.itemHeight)
	bg := tl.EffectiveTheme().TreeList.Group(tl.Variant()).ItemBackground.Resolve(StateDefault)
	tl.selHighlight.SetColor(bg.Color)
	tl.selHighlight.SetVisible(true)
	tl.MarkDrawDirty()
}

// SetRoots sets the root nodes of the tree and refreshes the view.
func (tl *TreeList) SetRoots(roots []*TreeNode) {
	tl.roots = roots
	tl.scrollPos.Set(0)
	DefaultScheduler.Flush()
	tl.rebuildFlatList()
	tl.updateScrollBar()
	tl.updateVisible()
}

// BindRoots binds the tree list to a reactive Array[*ReactiveTreeNode].
// Changes to any Children array anywhere in the subtree — at any depth —
// are automatically reflected in the tree view without resetting the scroll
// position or expand/collapse state.
//
// The binding uses a stable internal map so that the *TreeNode pointers
// backing the selection ref remain valid across tree mutations.
//
// Pass nil to detach the current binding.
func (tl *TreeList) BindRoots(arr *Array[*ReactiveTreeNode]) {
	if tl.stopArray != nil {
		tl.stopArray()
		tl.stopArray = nil
	}
	tl.reactiveRoots = arr
	tl.tnMap = make(map[*ReactiveTreeNode]*TreeNode)

	if arr == nil {
		tl.roots = nil
		tl.scrollPos.Set(0)
		DefaultScheduler.Flush()
		tl.rebuildFlatList()
		tl.updateScrollBar()
		tl.updateVisible()
		return
	}

	// Initial snapshot.
	tl.syncFromReactive()
	tl.scrollPos.Set(0)
	DefaultScheduler.Flush()
	tl.rebuildFlatList()
	tl.updateScrollBar()
	tl.updateVisible()

	// Wire watches on every Children array in the current tree topology.
	// On any change: re-snapshot, rebuild, then re-wire (to pick up new nodes).
	var watchHandles []WatchHandle

	stopAll := func() {
		for _, h := range watchHandles {
			h.Stop()
		}
		watchHandles = watchHandles[:0]
	}

	var wireAll func()
	onChange := func() {
		tl.syncFromReactive()
		tl.rebuildFlatList()
		tl.updateScrollBar()
		tl.updateVisible()
		wireAll()
	}
	wireAll = func() {
		stopAll()
		watchHandles = append(watchHandles, arr.OnChange(func() { onChange() }))
		var watchNode func(rn *ReactiveTreeNode)
		watchNode = func(rn *ReactiveTreeNode) {
			watchHandles = append(watchHandles, rn.Children.OnChange(func() { onChange() }))
			rn.Children.ForEach(func(_ int, child *ReactiveTreeNode) {
				watchNode(child)
			})
		}
		arr.ForEach(func(_ int, rn *ReactiveTreeNode) {
			watchNode(rn)
		})
	}
	wireAll()

	tl.stopArray = stopAll
}

// syncFromReactive rebuilds tl.roots from tl.reactiveRoots, reusing existing
// *TreeNode pointers via tnMap for selection stability.
func (tl *TreeList) syncFromReactive() {
	if tl.reactiveRoots == nil {
		tl.roots = nil
		return
	}
	seen := make(map[*ReactiveTreeNode]bool)
	tl.roots = tl.buildTreeNodes(tl.reactiveRoots, seen)
	// Prune stale entries so removed nodes can be garbage collected.
	for rn := range tl.tnMap {
		if !seen[rn] {
			delete(tl.tnMap, rn)
		}
	}
}

// buildTreeNodes recursively converts a reactive subtree into *TreeNode
// pointers, reusing existing entries from tnMap to keep pointers stable.
func (tl *TreeList) buildTreeNodes(arr *Array[*ReactiveTreeNode], seen map[*ReactiveTreeNode]bool) []*TreeNode {
	var result []*TreeNode
	arr.ForEach(func(_ int, rn *ReactiveTreeNode) {
		seen[rn] = true
		tn, ok := tl.tnMap[rn]
		if !ok {
			tn = &TreeNode{Data: rn.Data}
			tl.tnMap[rn] = tn
		} else {
			tn.Data = rn.Data // propagate data changes
		}
		tn.Children = tl.buildTreeNodes(rn.Children, seen)
		result = append(result, tn)
	})
	return result
}

// SetToggleIcons sets custom expand and collapse icon images for tree toggle
// buttons created by NewTreeToggle. When set, these are used instead of the
// default procedural glyphs.
func (tl *TreeList) SetToggleIcons(expand, collapse engine.Image) {
	tl.expandIcon = expand
	tl.collapseIcon = collapse
}

// SetRenderItem sets the factory function for rendering tree nodes.
func (tl *TreeList) SetRenderItem(fn func(*TreeNode, int) *Component) {
	tl.renderItem = fn
	tl.updateVisible()
}

// SetDefaultTextRenderer configures a standard text rendering function for the
// tree list. Each node's Data is expected to be a string. Rows include
// depth-based indentation, a toggle button (or spacer for leaves), and a label.
func (tl *TreeList) SetDefaultTextRenderer(source *sg.FontFamily, displaySize float64, rowW, rowH float64) {
	tl.SetRenderItem(func(node *TreeNode, depth int) *Component {
		text, _ := node.Data.(string)
		row := NewHBox("tree-row")
		row.Spacing = 4
		row.Align = AlignCenter
		row.Padding = Insets{Left: float64(depth) * TreeToggleSize}

		toggle := NewTreeToggle("toggle", tl, node)
		if toggle != nil {
			row.AddChild(toggle)
		} else {
			spacer := NewComponent("spacer")
			spacer.Width = TreeToggleSize
			spacer.Height = TreeToggleSize
			row.AddChild(spacer)
		}

		lbl := NewLabel("lbl", text, source, displaySize)
		lbl.SetInteractable(false)
		row.AddChild(lbl)

		if len(node.Children) > 0 {
			n := node
			row.OnClick(func(_ sg.ClickContext) {
				tl.Toggle(n)
			})
		}

		row.Width = rowW
		row.Height = rowH
		row.UpdateLayout()
		return row
	})
}

// Expand marks a tree node as expanded and refreshes the view.
func (tl *TreeList) Expand(node *TreeNode) {
	if node == nil || len(node.Children) == 0 {
		return
	}
	tl.expanded[node] = true
	tl.rebuildFlatList()
	tl.updateScrollBar()
	tl.updateVisible()
}

// Collapse marks a tree node as collapsed and refreshes the view.
func (tl *TreeList) Collapse(node *TreeNode) {
	if node == nil {
		return
	}
	delete(tl.expanded, node)
	tl.rebuildFlatList()
	tl.updateScrollBar()
	tl.updateVisible()
}

// Toggle toggles the expanded state of a tree node.
func (tl *TreeList) Toggle(node *TreeNode) {
	if tl.expanded[node] {
		tl.Collapse(node)
	} else {
		tl.Expand(node)
	}
}

// ExpandAll expands all tree nodes that have children.
func (tl *TreeList) ExpandAll() {
	tl.expandAllRecursive(tl.roots)
	tl.rebuildFlatList()
	tl.updateScrollBar()
	tl.updateVisible()
}

// CollapseAll collapses all tree nodes.
func (tl *TreeList) CollapseAll() {
	tl.expanded = make(map[*TreeNode]bool)
	tl.rebuildFlatList()
	tl.updateScrollBar()
	tl.updateVisible()
}

// IsExpanded reports whether a tree node is expanded.
func (tl *TreeList) IsExpanded(node *TreeNode) bool {
	return tl.expanded[node]
}

// FlatCount returns the number of visible (flattened) entries.
func (tl *TreeList) FlatCount() int {
	return len(tl.flatList)
}

// Update processes mouse wheel input and keyboard navigation for scrolling.
// This is called automatically via the willow node's OnUpdate hook; no
// manual call needed.
func (tl *TreeList) Update() {
	if tl.containsCursor() {
		_, wy := engine.Wheel()
		if wy != 0 {
			newPos := tl.scrollPos.Peek() - wy*scrollWheelSpeedList
			tl.scrollBar.SetScrollPos(newPos)
		}
	}

	// Keyboard navigation: Up/Down move selection when focused.
	if tl.focused && tl.enabled && tl.selectable {
		im := DefaultInputManager
		sel := tl.selected.Peek()
		idx := tl.flatIndexOf(sel)
		n := len(tl.flatList)
		if im.IsKeyJustAvailable(engine.KeyUp) && idx > 0 {
			tl.SetSelected(tl.flatList[idx-1].node)
			im.Consume(engine.KeyUp)
		} else if im.IsKeyJustAvailable(engine.KeyDown) && idx < n-1 {
			tl.SetSelected(tl.flatList[idx+1].node)
			im.Consume(engine.KeyDown)
		}
	}
}

// flatIndexOf returns the index of the given node in the flat list, or -1.
func (tl *TreeList) flatIndexOf(node *TreeNode) int {
	if node == nil {
		return -1
	}
	for i, e := range tl.flatList {
		if e.node == node {
			return i
		}
	}
	return -1
}

// SetSize sets the tree list dimensions.
func (tl *TreeList) SetSize(w, h float64) {
	tl.Width = w
	tl.Height = h

	sbWidth := float64(DefaultScrollBarWidth)

	tl.resizeBackground(w, h)
	tl.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}

	tl.scrollBar.SetSize(sbWidth, h)
	tl.scrollBar.SetPosition(w-sbWidth, 0)

	// Clipping mask so tree items don't render outside bounds.
	maskRoot := sg.NewContainer(tl.node.Name + "-mask")
	maskSprite := sg.NewSprite(tl.node.Name+"-mask-rect", sg.TextureRegion{})
	maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	maskSprite.SetScale(w-sbWidth, h)
	maskRoot.AddChild(maskSprite)
	tl.viewport.SetMask(maskRoot)

	tl.updateScrollBar()
	tl.updateVisible()
	tl.MarkLayoutDirty()
}

// Dispose cleans up the tree list.
func (tl *TreeList) Dispose() {
	if tl.stopArray != nil {
		tl.stopArray()
	}
	tl.selWatch.Stop()
	tl.clearPool()
	tl.scrollBar.Dispose()
	tl.Component.Dispose()
}

// PoolSize returns active component count for testing virtualization.
func (tl *TreeList) PoolSize() int {
	count := 0
	for _, p := range tl.pool {
		if p != nil {
			count++
		}
	}
	return count
}

func (tl *TreeList) expandAllRecursive(nodes []*TreeNode) {
	for _, n := range nodes {
		if len(n.Children) > 0 {
			tl.expanded[n] = true
			tl.expandAllRecursive(n.Children)
		}
	}
}

func (tl *TreeList) rebuildFlatList() {
	tl.clearPool() // items must re-render with current expansion state
	tl.flatList = tl.flatList[:0]
	tl.flattenNodes(tl.roots, 0)
	tl.updateHighlight()
}

func (tl *TreeList) flattenNodes(nodes []*TreeNode, depth int) {
	for _, n := range nodes {
		tl.flatList = append(tl.flatList, &flatEntry{node: n, depth: depth})
		if tl.expanded[n] && len(n.Children) > 0 {
			tl.flattenNodes(n.Children, depth+1)
		}
	}
}

func (tl *TreeList) updateScrollBar() {
	totalH := float64(len(tl.flatList)) * tl.itemHeight
	tl.scrollBar.SetContentSize(totalH, tl.Height)
	tl.scrollBar.SetVisible(totalH > tl.Height)
}

func (tl *TreeList) updateVisible() {
	if tl.renderItem == nil || len(tl.flatList) == 0 {
		tl.clearPool()
		return
	}

	pos := tl.scrollPos.Peek()
	viewH := tl.Height

	// Move the content container to reflect the scroll offset.
	tl.content.SetPosition(0, -pos)

	// Two-zone virtualization (see List.updateVisible for rationale).
	const (
		renderBuffer = 2
		keepBuffer   = 12
	)

	renderStart := int(math.Floor(pos/tl.itemHeight)) - renderBuffer
	if renderStart < 0 {
		renderStart = 0
	}
	renderEnd := renderStart + int(math.Ceil(viewH/tl.itemHeight)) + 1 + 2*renderBuffer
	if renderEnd > len(tl.flatList) {
		renderEnd = len(tl.flatList)
	}

	keepStart := int(math.Floor(pos/tl.itemHeight)) - keepBuffer
	if keepStart < 0 {
		keepStart = 0
	}
	keepEnd := keepStart + int(math.Ceil(viewH/tl.itemHeight)) + 1 + 2*keepBuffer
	if keepEnd > len(tl.flatList) {
		keepEnd = len(tl.flatList)
	}

	// Phase 1: Add newly visible items FIRST.
	for i := renderStart; i < renderEnd; i++ {
		if _, exists := tl.poolIndex[i]; !exists {
			y := float64(i) * tl.itemHeight
			tl.addPoolItem(i, y)
		}
	}

	// Phase 2: Remove items outside the wider keep zone.
	for idx := range tl.poolIndex {
		if idx < keepStart || idx >= keepEnd {
			tl.removePoolItem(idx)
		}
	}

	tl.visibleStart = renderStart
	tl.visibleEnd = renderEnd

	// Clip hit shapes to the visible portion of each item so that items
	// partially scrolled above/below the viewport don't intercept clicks
	// on elements outside the list.
	itemW := tl.Width - float64(DefaultScrollBarWidth)
	for idx, slot := range tl.poolIndex {
		if slot < len(tl.pool) && tl.pool[slot] != nil {
			itemY := float64(idx) * tl.itemHeight
			topVisible := itemY+tl.itemHeight > pos
			bottomVisible := itemY < pos+viewH
			if topVisible && bottomVisible {
				tl.pool[slot].node.Interactable = true
				hitY := math.Max(0, pos-itemY)
				hitH := math.Min(tl.itemHeight, pos+viewH-itemY) - hitY
				tl.pool[slot].node.HitShape = sg.HitRect{X: 0, Y: hitY, Width: itemW, Height: hitH}
			} else {
				tl.pool[slot].node.Interactable = false
			}
		}
	}
}

func (tl *TreeList) addPoolItem(index int, y float64) {
	entry := tl.flatList[index]
	comp := tl.renderItem(entry.node, entry.depth)
	if comp == nil {
		return
	}

	comp.SetPosition(0, y)
	comp.Width = tl.Width - float64(DefaultScrollBarWidth)
	comp.Height = tl.itemHeight
	comp.SetHitShape(sg.HitRect{X: 0, Y: 0, Width: comp.Width, Height: comp.Height})

	// Wire click for selection when selectable.
	if tl.selectable {
		isLeaf := len(entry.node.Children) == 0
		if !tl.leafOnlySelection || isLeaf {
			n := entry.node
			comp.OnClick(func(_ sg.ClickContext) {
				tl.SetSelected(n)
			})
		}
	}

	tl.content.AddChild(comp.Node())

	slot := -1
	for i, p := range tl.pool {
		if p == nil {
			slot = i
			break
		}
	}
	if slot >= 0 {
		tl.pool[slot] = comp
	} else {
		slot = len(tl.pool)
		tl.pool = append(tl.pool, comp)
	}
	tl.poolIndex[index] = slot
}

func (tl *TreeList) removePoolItem(index int) {
	slot, ok := tl.poolIndex[index]
	if !ok {
		return
	}
	if slot < len(tl.pool) && tl.pool[slot] != nil {
		tl.content.RemoveChild(tl.pool[slot].Node())
		tl.pool[slot].Dispose()
		tl.pool[slot] = nil
	}
	delete(tl.poolIndex, index)
}

func (tl *TreeList) clearPool() {
	for idx := range tl.poolIndex {
		slot := tl.poolIndex[idx]
		if slot < len(tl.pool) && tl.pool[slot] != nil {
			tl.content.RemoveChild(tl.pool[slot].Node())
			tl.pool[slot].Dispose()
			tl.pool[slot] = nil
		}
	}
	tl.poolIndex = make(map[int]int)
	tl.pool = tl.pool[:0]
}

// ---------------------------------------------------------------------------
// Tree toggle glyphs
// ---------------------------------------------------------------------------

// TreeToggleSize is the default size (width and height) for tree toggle
// icon buttons returned by NewTreeToggle. Exported so render callbacks
// can use it for leaf-node spacer sizing.
const TreeToggleSize = 16

// treeExpandGlyph returns the right-pointing chevron glyph from the
// default spritesheet.
func treeExpandGlyph() engine.Image { return IconChevronRight() }

// treeCollapseGlyph returns the down-pointing chevron glyph from the
// default spritesheet.
func treeCollapseGlyph() engine.Image { return IconChevronDown() }

// TreeExpandGlyph returns the expand glyph image. Used for testing.
func TreeExpandGlyph() engine.Image { return treeExpandGlyph() }

// TreeCollapseGlyph returns the collapse glyph image. Used for testing.
func TreeCollapseGlyph() engine.Image { return treeCollapseGlyph() }

// NewTreeToggle creates a ready-to-use IconButton for expanding/collapsing
// the given TreeNode within a TreeList. Returns nil for leaf nodes (no
// children). The button uses the Custom1 variant and tints the glyph icon
// with hover/active colors from the theme, following the same pattern as
// the window close button.
func NewTreeToggle(name string, tl *TreeList, node *TreeNode) *IconButton {
	if len(node.Children) == 0 {
		return nil
	}

	btn := NewIconButton(name)
	btn.SetTheme(tl.EffectiveTheme()) // Inherit theme from tree list (node-level add skips Component parent).
	btn.SetVariant(Custom1)
	btn.SetSize(TreeToggleSize, TreeToggleSize)

	// Set the appropriate glyph based on current expansion state.
	// Resolution order: per-instance override > theme icon > procedural glyph.
	tlGroup := tl.EffectiveTheme().TreeList.Group(tl.Variant())
	var glyph engine.Image
	if tl.IsExpanded(node) {
		glyph = tl.collapseIcon
		if glyph == nil && tlGroup.CollapseIcon.Set {
			glyph = tlGroup.CollapseIcon.Image
		}
		if glyph == nil {
			glyph = treeCollapseGlyph()
		}
	} else {
		glyph = tl.expandIcon
		if glyph == nil && tlGroup.ExpandIcon.Set {
			glyph = tlGroup.ExpandIcon.Image
		}
		if glyph == nil {
			glyph = treeExpandGlyph()
		}
	}
	// Render the glyph centered in the button. SetIconImage must be called
	// before SetIconSize so layoutChildren uses the divide-by-dims path
	// (SetScale(iW/img.Dx(), iH/img.Dy())). The display size is capped to
	// fit within the button with padding.
	b := glyph.Bounds()
	displaySize := math.Min(float64(b.Dx()), float64(TreeToggleSize-4))
	btn.SetIconImage(glyph)
	btn.SetIconSize(displaySize, displaySize)

	// Wire click to toggle expansion.
	n := node
	btn.SetOnClick(func() {
		tl.Toggle(n)
	})

	// Tint the icon: white in default state, theme TextColor on hover/active.
	white := sg.RGBA(1, 1, 1, 1)
	origVisualChange := btn.onVisualStateChange
	tintFn := func() {
		if origVisualChange != nil {
			origVisualChange()
		}
		st := computeState(btn.enabled, btn.focused, btn.hovered, btn.pressed)
		if st == StateDefault {
			btn.icon.SetColor(white)
		} else {
			group := btn.EffectiveTheme().Button.Group(btn.Variant())
			btn.icon.SetColor(group.TextColor.Resolve(st))
		}
	}
	btn.onVisualStateChange = tintFn
	btn.onThemeChange = tintFn

	return btn
}
