package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// SortableTreeItem represents a node in a sortable tree hierarchy.
type SortableTreeItem struct {
	ID       string
	ParentID string // empty = root
	Label    string
	Icon     sg.TextureRegion
	Depth    int // computed from ParentID if 0
}

// sortableTreeEntry is a flattened visible entry with computed depth.
type sortableTreeEntry struct {
	item     SortableTreeItem
	depth    int
	hasKids  bool
	expanded bool
}

// SortableTreeList is a hierarchical list where nodes can be reordered by drag
// within their level and optionally reparented by dragging onto another node.
type SortableTreeList struct {
	Component
	viewport  *sg.Node
	content   *sg.Node
	scrollBar *ScrollBar

	items    []SortableTreeItem
	flatList []sortableTreeEntry
	expanded map[string]bool // item ID -> expanded

	itemHeight float64
	scrollPos  *Ref[float64]

	// Selection.
	selected     *Ref[int] // flat list index, -1 = none
	selHighlight *sg.Node
	selWatch     WatchHandle
	onChange     func(int)

	// Row management (non-virtualized).
	rows []*sortableTreeRow

	// Feature flags.
	allowReparent   bool
	allowCrossLevel bool

	// Drag state.
	dragging       bool
	dragFromIndex  int
	dragStartY     float64
	dragCurrentY   float64
	insertTarget   int // flat index for between-sibling drop
	reparentTarget int // flat index for reparent highlight (-1 = none)

	// Visual indicators.
	insertIndicator *sg.Node // horizontal line between items
	reparentBg      *sg.Node // highlight on reparent target

	// Callbacks.
	onReorder func(itemID, newParentID string, newIndex int)

	// Font for default rendering.
	font        *sg.FontFamily
	displaySize float64
}

// sortableTreeRow tracks a rendered row.
type sortableTreeRow struct {
	container  *sg.Node
	bgNode     *sg.Node
	toggleNode *sg.Node
	labelNode  *sg.Node
	iconNode   *sg.Node
}

// NewSortableTreeList creates a new sortable tree list.
func NewSortableTreeList(name string, source *sg.FontFamily, displaySize float64) *SortableTreeList {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	st := &SortableTreeList{
		itemHeight:     30,
		expanded:       make(map[string]bool),
		selected:       NewRef(-1),
		scrollPos:      NewRef(0.0),
		insertTarget:   -1,
		reparentTarget: -1,
		font:           font,
		displaySize:    displaySize,
	}
	initComponent(&st.Component, name)
	st.initBackground(name)
	st.bgNode.SetColor(st.EffectiveTheme().SortableTreeList.Group(st.Variant()).Background.Resolve(StateDefault).Color)

	// Viewport container (clips content).
	st.viewport = sg.NewContainer(name + "-viewport")
	st.viewport.Interactable = true
	st.node.AddChild(st.viewport)

	// Content container inside viewport.
	st.content = sg.NewContainer(name + "-content")
	st.content.Interactable = true
	st.viewport.AddChild(st.content)

	// ScrollBar.
	st.scrollBar = NewScrollBar(name + "-scrollbar")
	st.scrollBar.SetOnChange(func(pos float64) {
		old := st.scrollPos.Peek()
		st.scrollPos.Set(pos)
		DefaultScheduler.Flush()
		if pos != old {
			st.layoutRows()
		}
	})
	st.scrollBar.AddToNode(st.node)

	// Selection highlight.
	st.selHighlight = sg.NewSprite(name+"-sel-hl", sg.TextureRegion{})
	st.selHighlight.SetVisible(false)
	st.selHighlight.SetZIndex(999)
	st.content.AddChild(st.selHighlight)

	// Insert indicator (hidden until drag).
	st.insertIndicator = sg.NewSprite(name+"-insert-ind", sg.TextureRegion{})
	st.insertIndicator.SetVisible(false)
	st.insertIndicator.SetZIndex(10)
	st.content.AddChild(st.insertIndicator)

	// Reparent highlight (hidden until drag over target).
	st.reparentBg = sg.NewSprite(name+"-reparent-bg", sg.TextureRegion{})
	st.reparentBg.SetVisible(false)
	st.reparentBg.SetZIndex(9)
	st.content.AddChild(st.reparentBg)

	// Auto-update.
	st.node.OnUpdate = func(_ float64) {
		st.Update()
	}

	// Click on list focuses it.
	st.node.OnPointerDown(func(_ sg.PointerContext) {
		if st.enabled {
			DefaultFocusManager.SetFocus(&st.Component)
		}
	})

	st.scrollBar.parent = &st.Component
	st.onThemeChange = func() {
		st.applyThemeColors()
		st.scrollBar.applyThemeColors()
	}

	// Focus registration.
	st.enableFocusNavigation()
	st.InterceptArrows = true
	st.ConsumeHandledKeys = false

	st.onFocusChange = func(focused bool) { st.applyThemeColors() }
	st.SetHandleKey(func(key engine.Key) bool {
		n := len(st.flatList)
		if n == 0 {
			return false
		}
		sel := st.selected.Peek()
		switch key {
		case engine.KeyUp:
			return sel > 0
		case engine.KeyDown:
			return sel < n-1
		case engine.KeyLeft, engine.KeyRight:
			return sel >= 0 && sel < n
		}
		return false
	})

	st.SetSize(240, 400)
	return st
}

func (st *SortableTreeList) applyThemeColors() {
	group := st.EffectiveTheme().SortableTreeList.Group(st.Variant())
	st.bgNode.SetColor(group.Background.Resolve(StateDefault).Color)
	st.state = computeState(st.enabled, st.focused, st.hovered, false)
	st.applyFocusRing(group.FocusColor.Resolve(st.state), group.FocusRingWidth)
	st.updateHighlight()
	st.updateIndicatorColors()
	st.MarkDrawDirty()
}

// SetItems sets the tree items and refreshes the view.
func (st *SortableTreeList) SetItems(items []SortableTreeItem) {
	st.items = make([]SortableTreeItem, len(items))
	copy(st.items, items)
	st.computeDepths()
	st.rebuildFlatList()
	st.updateScrollBar()
	st.rebuild()
}

// Items returns the current items.
func (st *SortableTreeList) Items() []SortableTreeItem {
	result := make([]SortableTreeItem, len(st.items))
	copy(result, st.items)
	return result
}

// SetAllowReparent enables or disables reparenting by dragging onto a node.
func (st *SortableTreeList) SetAllowReparent(v bool) {
	st.allowReparent = v
}

// SetAllowCrossLevel enables or disables moving items to different depth levels.
func (st *SortableTreeList) SetAllowCrossLevel(v bool) {
	st.allowCrossLevel = v
}

// SetExpanded sets the expansion state for a given item ID.
func (st *SortableTreeList) SetExpanded(id string, v bool) {
	if v {
		st.expanded[id] = true
	} else {
		delete(st.expanded, id)
	}
	st.rebuildFlatList()
	st.updateScrollBar()
	st.rebuild()
}

// IsExpanded reports whether the given item ID is expanded.
func (st *SortableTreeList) IsExpanded(id string) bool {
	return st.expanded[id]
}

// ExpandAll expands all items that have children.
func (st *SortableTreeList) ExpandAll() {
	for _, item := range st.items {
		if st.hasChildren(item.ID) {
			st.expanded[item.ID] = true
		}
	}
	st.rebuildFlatList()
	st.updateScrollBar()
	st.rebuild()
}

// CollapseAll collapses all items.
func (st *SortableTreeList) CollapseAll() {
	st.expanded = make(map[string]bool)
	st.rebuildFlatList()
	st.updateScrollBar()
	st.rebuild()
}

// SetOnReorder sets the callback for reorder/reparent completion.
func (st *SortableTreeList) SetOnReorder(fn func(itemID, newParentID string, newIndex int)) {
	st.onReorder = fn
}

// SetSize sets the list dimensions.
func (st *SortableTreeList) SetSize(w, h float64) {
	st.Width = w
	st.Height = h

	sbWidth := float64(DefaultScrollBarWidth)

	st.resizeBackground(w, h)
	st.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}

	// ScrollBar.
	st.scrollBar.SetSize(sbWidth, h)
	st.scrollBar.SetPosition(w-sbWidth, 0)

	// Clipping mask.
	maskRoot := sg.NewContainer(st.node.Name + "-mask")
	maskSprite := sg.NewSprite(st.node.Name+"-mask-rect", sg.TextureRegion{})
	maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	maskSprite.SetScale(w-sbWidth, h)
	maskRoot.AddChild(maskSprite)
	st.viewport.SetMask(maskRoot)

	st.updateScrollBar()
	st.rebuild()
	st.MarkLayoutDirty()
}

// Selected returns the currently selected flat list index, or -1.
func (st *SortableTreeList) Selected() int {
	return st.selected.Peek()
}

// SetSelected sets the selected flat list index.
func (st *SortableTreeList) SetSelected(idx int) {
	n := len(st.flatList)
	if idx < -1 || idx >= n {
		return
	}
	old := st.selected.Peek()
	st.selected.Set(idx)
	DefaultScheduler.Flush()
	if idx != old && st.onChange != nil {
		st.onChange(idx)
	}
	st.updateHighlight()
	st.MarkDrawDirty()
	if idx >= 0 {
		st.scrollToIndex(idx)
	}
}

// BindSelected two-way binds the selected index to an external Ref.
func (st *SortableTreeList) BindSelected(ref *Ref[int]) {
	st.selected = ref
	bindRef(&st.selWatch, ref, st.SetSelected)
}

// SetOnChange sets the callback for selection changes.
func (st *SortableTreeList) SetOnChange(fn func(int)) {
	st.onChange = fn
}

// Update processes mouse wheel and keyboard input.
func (st *SortableTreeList) Update() {
	if st.containsCursor() {
		_, wy := engine.Wheel()
		if wy != 0 {
			newPos := st.scrollPos.Peek() - wy*scrollWheelSpeedList
			st.scrollBar.SetScrollPos(newPos)
		}
	}

	if !st.focused || !st.enabled {
		return
	}

	im := DefaultInputManager
	sel := st.selected.Peek()
	n := len(st.flatList)

	// Ctrl+Up/Down: reorder within level.
	if im.IsKeyJustAvailable(engine.KeyUp) && sel > 0 {
		if engine.IsKeyPressed(engine.KeyControl) {
			st.MoveSelectedUp()
			im.Consume(engine.KeyUp)
			return
		}
		st.SetSelected(sel - 1)
		im.Consume(engine.KeyUp)
	} else if im.IsKeyJustAvailable(engine.KeyDown) && sel < n-1 {
		if engine.IsKeyPressed(engine.KeyControl) {
			st.MoveSelectedDown()
			im.Consume(engine.KeyDown)
			return
		}
		st.SetSelected(sel + 1)
		im.Consume(engine.KeyDown)
	}

	// Left/Right: collapse/expand or Ctrl+Left/Right for indent/outdent.
	if im.IsKeyJustAvailable(engine.KeyLeft) && sel >= 0 && sel < n {
		if engine.IsKeyPressed(engine.KeyControl) && st.allowCrossLevel {
			st.OutdentSelected()
			im.Consume(engine.KeyLeft)
		} else {
			// Collapse if expanded parent.
			entry := st.flatList[sel]
			if entry.hasKids && entry.expanded {
				st.expanded[entry.item.ID] = false
				st.rebuildFlatList()
				st.updateScrollBar()
				st.rebuild()
			}
			im.Consume(engine.KeyLeft)
		}
	} else if im.IsKeyJustAvailable(engine.KeyRight) && sel >= 0 && sel < n {
		if engine.IsKeyPressed(engine.KeyControl) && st.allowCrossLevel {
			st.IndentSelected()
			im.Consume(engine.KeyRight)
		} else {
			// Expand if collapsed parent.
			entry := st.flatList[sel]
			if entry.hasKids && !entry.expanded {
				st.expanded[entry.item.ID] = true
				st.rebuildFlatList()
				st.updateScrollBar()
				st.rebuild()
			}
			im.Consume(engine.KeyRight)
		}
	}
}

// Dispose cleans up resources.
func (st *SortableTreeList) Dispose() {
	st.clearRows()
	st.scrollBar.Dispose()
	st.Component.Dispose()
}

// ---------------------------------------------------------------------------
// Internal: tree data management
// ---------------------------------------------------------------------------

func (st *SortableTreeList) computeDepths() {
	// Build a lookup from ID to depth based on parent chain.
	depthMap := make(map[string]int)
	idMap := make(map[string]*SortableTreeItem)
	for i := range st.items {
		idMap[st.items[i].ID] = &st.items[i]
	}
	var getDepth func(id string) int
	getDepth = func(id string) int {
		if d, ok := depthMap[id]; ok {
			return d
		}
		item, ok := idMap[id]
		if !ok || item.ParentID == "" {
			depthMap[id] = 0
			return 0
		}
		d := getDepth(item.ParentID) + 1
		depthMap[id] = d
		return d
	}
	for i := range st.items {
		st.items[i].Depth = getDepth(st.items[i].ID)
	}
}

func (st *SortableTreeList) hasChildren(parentID string) bool {
	for _, item := range st.items {
		if item.ParentID == parentID {
			return true
		}
	}
	return false
}

func (st *SortableTreeList) childrenOf(parentID string) []SortableTreeItem {
	var result []SortableTreeItem
	for _, item := range st.items {
		if item.ParentID == parentID {
			result = append(result, item)
		}
	}
	return result
}

func (st *SortableTreeList) rebuildFlatList() {
	st.flatList = st.flatList[:0]
	st.flattenItems("", 0)
}

func (st *SortableTreeList) flattenItems(parentID string, depth int) {
	children := st.childrenOf(parentID)
	for _, child := range children {
		hasKids := st.hasChildren(child.ID)
		expanded := st.expanded[child.ID]
		st.flatList = append(st.flatList, sortableTreeEntry{
			item:     child,
			depth:    depth,
			hasKids:  hasKids,
			expanded: expanded,
		})
		if hasKids && expanded {
			st.flattenItems(child.ID, depth+1)
		}
	}
}

// siblingIndex returns the index of itemID among its siblings (children of the same parent).
func (st *SortableTreeList) siblingIndex(itemID, parentID string) int {
	idx := 0
	for _, item := range st.items {
		if item.ParentID == parentID {
			if item.ID == itemID {
				return idx
			}
			idx++
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Internal: keyboard reorder
// ---------------------------------------------------------------------------

// MoveSelectedUp moves the selected item one position up within its parent.
func (st *SortableTreeList) MoveSelectedUp() {
	sel := st.selected.Peek()
	if sel <= 0 || sel >= len(st.flatList) {
		return
	}
	entry := st.flatList[sel]
	siblings := st.childrenOf(entry.item.ParentID)
	sibIdx := st.siblingIndex(entry.item.ID, entry.item.ParentID)
	if sibIdx <= 0 {
		return
	}
	// Swap with previous sibling in the items array.
	prevSibling := siblings[sibIdx-1]
	st.swapSiblings(entry.item.ID, prevSibling.ID, entry.item.ParentID)
	st.rebuildFlatList()
	st.updateScrollBar()
	// Find new flat index of the moved item.
	newIdx := st.flatIndexOf(entry.item.ID)
	st.rebuild()
	st.SetSelected(newIdx)
	if st.onReorder != nil {
		st.onReorder(entry.item.ID, entry.item.ParentID, sibIdx-1)
	}
}

// MoveSelectedDown moves the selected item one position down within its parent.
func (st *SortableTreeList) MoveSelectedDown() {
	sel := st.selected.Peek()
	if sel < 0 || sel >= len(st.flatList) {
		return
	}
	entry := st.flatList[sel]
	siblings := st.childrenOf(entry.item.ParentID)
	sibIdx := st.siblingIndex(entry.item.ID, entry.item.ParentID)
	if sibIdx < 0 || sibIdx >= len(siblings)-1 {
		return
	}
	nextSibling := siblings[sibIdx+1]
	st.swapSiblings(entry.item.ID, nextSibling.ID, entry.item.ParentID)
	st.rebuildFlatList()
	st.updateScrollBar()
	newIdx := st.flatIndexOf(entry.item.ID)
	st.rebuild()
	st.SetSelected(newIdx)
	if st.onReorder != nil {
		st.onReorder(entry.item.ID, entry.item.ParentID, sibIdx+1)
	}
}

// IndentSelected moves the selected item under its preceding sibling.
func (st *SortableTreeList) IndentSelected() {
	sel := st.selected.Peek()
	if sel < 0 || sel >= len(st.flatList) {
		return
	}
	entry := st.flatList[sel]
	siblings := st.childrenOf(entry.item.ParentID)
	sibIdx := st.siblingIndex(entry.item.ID, entry.item.ParentID)
	if sibIdx <= 0 {
		return // No preceding sibling to become parent.
	}
	newParent := siblings[sibIdx-1]
	// Update the item's ParentID.
	for i := range st.items {
		if st.items[i].ID == entry.item.ID {
			st.items[i].ParentID = newParent.ID
			break
		}
	}
	st.computeDepths()
	// Expand new parent so the moved item is visible.
	st.expanded[newParent.ID] = true
	st.rebuildFlatList()
	st.updateScrollBar()
	newIdx := st.flatIndexOf(entry.item.ID)
	st.rebuild()
	st.SetSelected(newIdx)
	newSibIdx := st.siblingIndex(entry.item.ID, newParent.ID)
	if st.onReorder != nil {
		st.onReorder(entry.item.ID, newParent.ID, newSibIdx)
	}
}

// OutdentSelected moves the selected item to its grandparent level.
func (st *SortableTreeList) OutdentSelected() {
	sel := st.selected.Peek()
	if sel < 0 || sel >= len(st.flatList) {
		return
	}
	entry := st.flatList[sel]
	if entry.item.ParentID == "" {
		return // Already at root.
	}
	// Find grandparent.
	var grandparentID string
	for _, item := range st.items {
		if item.ID == entry.item.ParentID {
			grandparentID = item.ParentID
			break
		}
	}
	// Update the item's ParentID to grandparent.
	for i := range st.items {
		if st.items[i].ID == entry.item.ID {
			st.items[i].ParentID = grandparentID
			break
		}
	}
	st.computeDepths()
	st.rebuildFlatList()
	st.updateScrollBar()
	newIdx := st.flatIndexOf(entry.item.ID)
	st.rebuild()
	st.SetSelected(newIdx)
	newSibIdx := st.siblingIndex(entry.item.ID, grandparentID)
	if st.onReorder != nil {
		st.onReorder(entry.item.ID, grandparentID, newSibIdx)
	}
}

func (st *SortableTreeList) swapSiblings(idA, idB, parentID string) {
	// Find indices of A and B in the items slice.
	ai, bi := -1, -1
	for i, item := range st.items {
		if item.ID == idA {
			ai = i
		}
		if item.ID == idB {
			bi = i
		}
	}
	if ai >= 0 && bi >= 0 {
		st.items[ai], st.items[bi] = st.items[bi], st.items[ai]
	}
}

func (st *SortableTreeList) flatIndexOf(itemID string) int {
	for i, entry := range st.flatList {
		if entry.item.ID == itemID {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Internal: drag mechanics
// ---------------------------------------------------------------------------

func (st *SortableTreeList) wireDrag(container *sg.Node, flatIdx int) {
	container.OnDragStart(func(ctx sg.DragContext) {
		if !st.enabled {
			return
		}
		st.dragging = true
		st.dragFromIndex = flatIdx
		st.dragStartY = ctx.GlobalY
		st.dragCurrentY = ctx.GlobalY
		st.SetSelected(flatIdx)
		st.updateDropTarget()
	})

	container.OnDrag(func(ctx sg.DragContext) {
		if !st.dragging {
			return
		}
		st.dragCurrentY = ctx.GlobalY
		st.updateDropTarget()
	})

	container.OnDragEnd(func(ctx sg.DragContext) {
		if !st.dragging {
			return
		}
		st.dragging = false
		st.insertIndicator.SetVisible(false)
		st.reparentBg.SetVisible(false)

		if st.reparentTarget >= 0 && st.allowReparent {
			st.doReparent(st.dragFromIndex, st.reparentTarget)
		} else if st.insertTarget >= 0 && st.insertTarget != st.dragFromIndex {
			st.doReorder(st.dragFromIndex, st.insertTarget)
		}

		st.insertTarget = -1
		st.reparentTarget = -1
	})
}

func (st *SortableTreeList) updateDropTarget() {
	if !st.dragging {
		return
	}

	// Convert global Y to widget-local Y.
	_, localY := st.node.WorldToLocal(0, st.dragCurrentY)
	localY += st.scrollPos.Peek()

	n := len(st.flatList)
	if n == 0 {
		return
	}

	group := st.EffectiveTheme().SortableTreeList.Group(st.Variant())
	sbWidth := float64(DefaultScrollBarWidth)
	contentW := st.Width - sbWidth

	// Determine if we're in the middle of a row (reparent) or at an edge (reorder).
	rowIdx := int(math.Floor(localY / st.itemHeight))
	if rowIdx < 0 {
		rowIdx = 0
	}
	if rowIdx >= n {
		rowIdx = n - 1
	}

	rowY := float64(rowIdx) * st.itemHeight
	posInRow := localY - rowY
	edgeThreshold := st.itemHeight * 0.25

	dragEntry := st.flatList[st.dragFromIndex]

	if st.allowReparent && posInRow > edgeThreshold && posInRow < st.itemHeight-edgeThreshold && rowIdx != st.dragFromIndex {
		// Reparent zone: middle of the target row.
		targetEntry := st.flatList[rowIdx]

		// Don't allow reparenting onto self or own descendant.
		if !st.isDescendantOf(targetEntry.item.ID, dragEntry.item.ID) &&
			targetEntry.item.ID != dragEntry.item.ID {
			// Check cross-level constraint.
			if st.allowCrossLevel || targetEntry.depth == dragEntry.depth-1 {
				st.reparentTarget = rowIdx
				st.insertTarget = -1
				st.insertIndicator.SetVisible(false)

				// Show reparent highlight.
				st.reparentBg.SetPosition(0, rowY)
				st.reparentBg.SetScale(contentW, st.itemHeight)
				st.reparentBg.SetColor(group.DropTargetBg.Resolve(StateDefault).Color)
				st.reparentBg.SetVisible(true)
				return
			}
		}
	}

	// Reorder zone: between siblings.
	st.reparentTarget = -1
	st.reparentBg.SetVisible(false)

	target := rowIdx
	if posInRow > st.itemHeight/2 {
		target++ // Insert after this row.
	}
	if target > n {
		target = n
	}

	// Only allow reorder within same parent.
	if target < n && target >= 0 {
		// Determine what the target parent would be.
		var targetParentID string
		if target < n {
			targetParentID = st.flatList[target].item.ParentID
		}
		if target == n || (target > 0 && target == n) {
			// Dropping at the end — use the parent of the last item at drag source depth.
			targetParentID = dragEntry.item.ParentID
		}

		if !st.allowCrossLevel && targetParentID != dragEntry.item.ParentID {
			// Can't cross levels — find nearest valid position within same parent.
			st.insertTarget = -1
			st.insertIndicator.SetVisible(false)
			return
		}
	}

	// Clamp and skip no-op.
	if target < 0 {
		target = 0
	}
	if target > n {
		target = n
	}
	if target == st.dragFromIndex {
		st.insertTarget = -1
		st.insertIndicator.SetVisible(false)
		return
	}

	st.insertTarget = target

	// Position the insert indicator at the top edge of the target slot.
	// The item will be inserted BEFORE flatList[target], so the indicator
	// goes at the top of that row. When target == n, it goes at the bottom
	// of the last row (append).
	indicatorY := float64(target) * st.itemHeight
	indicatorH := group.DropLineWidth
	if indicatorH <= 0 {
		indicatorH = 2
	}
	st.insertIndicator.SetPosition(0, indicatorY-indicatorH/2)
	st.insertIndicator.SetScale(contentW, indicatorH)
	st.insertIndicator.SetVisible(true)
}

func (st *SortableTreeList) isDescendantOf(candidateID, ancestorID string) bool {
	// Walk up from candidate checking if we hit ancestor.
	visited := make(map[string]bool)
	current := candidateID
	for current != "" {
		if visited[current] {
			return false // cycle guard
		}
		visited[current] = true
		if current == ancestorID {
			return true
		}
		// Find parent.
		found := false
		for _, item := range st.items {
			if item.ID == current {
				current = item.ParentID
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return false
}

func (st *SortableTreeList) doReparent(fromFlatIdx, toFlatIdx int) {
	if fromFlatIdx < 0 || fromFlatIdx >= len(st.flatList) || toFlatIdx < 0 || toFlatIdx >= len(st.flatList) {
		return
	}
	dragItem := st.flatList[fromFlatIdx].item
	targetItem := st.flatList[toFlatIdx].item

	// Update item's parent.
	for i := range st.items {
		if st.items[i].ID == dragItem.ID {
			st.items[i].ParentID = targetItem.ID
			break
		}
	}
	st.computeDepths()
	// Expand target so new child is visible.
	st.expanded[targetItem.ID] = true
	st.rebuildFlatList()
	st.updateScrollBar()

	newSibIdx := st.siblingIndex(dragItem.ID, targetItem.ID)
	newFlatIdx := st.flatIndexOf(dragItem.ID)
	st.rebuild()
	st.SetSelected(newFlatIdx)

	if st.onReorder != nil {
		st.onReorder(dragItem.ID, targetItem.ID, newSibIdx)
	}
}

func (st *SortableTreeList) doReorder(fromFlatIdx, toFlatIdx int) {
	if fromFlatIdx < 0 || fromFlatIdx >= len(st.flatList) || toFlatIdx < 0 || toFlatIdx > len(st.flatList) {
		return
	}
	dragItem := st.flatList[fromFlatIdx].item

	// When dropping at the very end, use the drag item's own parent.
	var targetItem SortableTreeItem
	if toFlatIdx < len(st.flatList) {
		targetItem = st.flatList[toFlatIdx].item
	} else {
		targetItem = dragItem
	}

	// Only reorder within the same parent unless cross-level is allowed.
	newParentID := dragItem.ParentID
	if st.allowCrossLevel && toFlatIdx < len(st.flatList) {
		newParentID = targetItem.ParentID
	} else if dragItem.ParentID != targetItem.ParentID {
		return
	}

	if newParentID != dragItem.ParentID {
		// Cross-level move: update parent.
		for i := range st.items {
			if st.items[i].ID == dragItem.ID {
				st.items[i].ParentID = newParentID
				break
			}
		}
		st.computeDepths()
	}

	// Remove dragItem (and its collapsed children) from items, then reinsert
	// at the position indicated by the drop target.
	var dragItemCopy SortableTreeItem
	var dragChildren []SortableTreeItem
	var filtered []SortableTreeItem
	for _, item := range st.items {
		if item.ID == dragItem.ID {
			dragItemCopy = item
			dragItemCopy.ParentID = newParentID
		} else if st.isDescendantOf(item.ID, dragItem.ID) {
			dragChildren = append(dragChildren, item)
		} else {
			filtered = append(filtered, item)
		}
	}

	// Determine insert position in the filtered items list.
	// toFlatIdx points to the flat list slot where the indicator is drawn.
	// If dragging down, this is the item that should come AFTER the dragged
	// item (insert before it). If dragging up, it's the item that should
	// come BEFORE (insert before it as well, since we want to take its slot).
	insertPos := len(filtered) // default: append at end
	if toFlatIdx < len(st.flatList) {
		targetID := st.flatList[toFlatIdx].item.ID
		for i, item := range filtered {
			if item.ID == targetID {
				insertPos = i
				break
			}
		}
	}

	// Insert dragItem + its children at the computed position.
	newItems := make([]SortableTreeItem, 0, len(st.items))
	newItems = append(newItems, filtered[:insertPos]...)
	newItems = append(newItems, dragItemCopy)
	newItems = append(newItems, dragChildren...)
	newItems = append(newItems, filtered[insertPos:]...)
	st.items = newItems

	st.computeDepths()
	st.rebuildFlatList()
	st.updateScrollBar()

	newFlatIdx := st.flatIndexOf(dragItem.ID)
	st.rebuild()
	st.SetSelected(newFlatIdx)

	newSibIdx := st.siblingIndex(dragItem.ID, newParentID)
	if st.onReorder != nil {
		st.onReorder(dragItem.ID, newParentID, newSibIdx)
	}
}

// ---------------------------------------------------------------------------
// Internal: rendering
// ---------------------------------------------------------------------------

func (st *SortableTreeList) rebuild() {
	st.clearRows()
	if len(st.flatList) == 0 {
		return
	}

	group := st.EffectiveTheme().SortableTreeList.Group(st.Variant())
	sbWidth := float64(DefaultScrollBarWidth)
	contentW := st.Width - sbWidth

	for i, entry := range st.flatList {
		row := st.createRow(i, entry, contentW, group)
		st.rows = append(st.rows, row)
	}

	st.layoutRows()
	st.updateHighlight()
}

func (st *SortableTreeList) createRow(flatIdx int, entry sortableTreeEntry, contentW float64, group *SortableTreeListGroup) *sortableTreeRow {
	row := &sortableTreeRow{}

	// Container for the whole row.
	row.container = sg.NewContainer(st.node.Name + "-row")
	row.container.Interactable = true
	row.container.HitShape = sg.HitRect{X: 0, Y: 0, Width: contentW, Height: st.itemHeight}

	// Row background.
	rowBg := group.RowBackground.Resolve(StateDefault)
	if rowBg.Type != BgNone && rowBg.Color != (sg.Color{}) {
		bgSprite := sg.NewSprite(st.node.Name+"-row-bg", sg.TextureRegion{})
		bgSprite.SetScale(contentW, st.itemHeight)
		bgSprite.SetColor(rowBg.Color)
		bgSprite.SetZIndex(-2)
		row.bgNode = bgSprite
		row.container.AddChild(bgSprite)
	}

	// Indentation.
	indent := float64(entry.depth) * group.IndentWidth

	// Click for selection, or toggle expand/collapse when clicking the chevron area.
	idx := flatIdx
	clickItemID := entry.item.ID
	clickHasKids := entry.hasKids
	clickChevronMaxX := indent + group.ChevronSize + 8
	row.container.OnClick(func(ctx sg.ClickContext) {
		st.SetSelected(idx)
		if clickHasKids && ctx.LocalX < clickChevronMaxX {
			if st.expanded[clickItemID] {
				delete(st.expanded, clickItemID)
			} else {
				st.expanded[clickItemID] = true
			}
			st.rebuildFlatList()
			st.updateScrollBar()
			st.rebuild()
		}
	})

	// Wire drag on the whole row.
	st.wireDrag(row.container, idx)

	// Toggle chevron.
	chevronSize := group.ChevronSize
	if chevronSize <= 0 {
		chevronSize = 12
	}
	toggleX := indent + 4

	if entry.hasKids {
		var glyph engine.Image
		if entry.expanded {
			glyph = treeCollapseGlyph()
		} else {
			glyph = treeExpandGlyph()
		}
		toggleSprite := sg.NewSprite(st.node.Name+"-toggle", sg.TextureRegion{})
		toggleSprite.SetCustomImage(glyph)
		toggleSprite.SetSize(chevronSize, chevronSize)
		toggleSprite.SetPosition(toggleX, (st.itemHeight-chevronSize)/2)
		toggleSprite.SetColor(group.ChevronColor.Resolve(StateDefault))
		row.toggleNode = toggleSprite
		row.container.AddChild(toggleSprite)
	}

	// Content: icon + label.
	contentX := toggleX + chevronSize + 4

	// Icon.
	if entry.item.Icon.Width > 0 {
		iconSize := group.IconSize
		if iconSize <= 0 {
			iconSize = 16
		}
		iconSprite := sg.NewSprite(st.node.Name+"-icon", entry.item.Icon)
		iconSprite.SetScale(iconSize, iconSize)
		iconSprite.SetPosition(contentX, (st.itemHeight-iconSize)/2)
		row.iconNode = iconSprite
		row.container.AddChild(iconSprite)
		contentX += iconSize + group.IconGap
	}

	// Label.
	if entry.item.Label != "" {
		labelNode := sg.NewText(st.node.Name+"-label", entry.item.Label, st.font)
		labelNode.TextBlock.FontSize = st.displaySize
		labelNode.TextBlock.Color = group.LabelColor.Resolve(StateDefault)
		labelNode.SetPosition(contentX, (st.itemHeight-st.displaySize)/2)
		row.labelNode = labelNode
		row.container.AddChild(labelNode)
	}

	st.content.AddChild(row.container)
	return row
}

func (st *SortableTreeList) layoutRows() {
	pos := st.scrollPos.Peek()
	st.content.SetPosition(0, -pos)

	for i, row := range st.rows {
		y := float64(i) * st.itemHeight
		row.container.SetPosition(0, y)
	}

	st.updateHighlight()
}

func (st *SortableTreeList) clearRows() {
	for _, row := range st.rows {
		st.content.RemoveChild(row.container)
	}
	st.rows = nil
}

func (st *SortableTreeList) updateHighlight() {
	idx := st.selected.Peek()
	n := len(st.flatList)
	if idx < 0 || idx >= n {
		st.selHighlight.SetVisible(false)
		return
	}
	sbWidth := float64(DefaultScrollBarWidth)
	itemW := st.Width - sbWidth
	y := float64(idx) * st.itemHeight
	st.selHighlight.SetPosition(0, y)
	st.selHighlight.SetScale(itemW, st.itemHeight)
	group := st.EffectiveTheme().SortableTreeList.Group(st.Variant())
	st.selHighlight.SetColor(group.RowSelectedBg.Resolve(StateDefault).Color)
	st.selHighlight.SetVisible(true)
	st.MarkDrawDirty()
}

func (st *SortableTreeList) updateIndicatorColors() {
	group := st.EffectiveTheme().SortableTreeList.Group(st.Variant())
	st.insertIndicator.SetColor(group.DropLineColor.Resolve(StateDefault))
}

func (st *SortableTreeList) updateScrollBar() {
	n := len(st.flatList)
	totalH := float64(n) * st.itemHeight
	st.scrollBar.SetContentSize(totalH, st.Height)
	st.scrollBar.SetVisible(totalH > st.Height)
}

func (st *SortableTreeList) scrollToIndex(idx int) {
	n := len(st.flatList)
	if idx < 0 || idx >= n {
		return
	}
	itemTop := float64(idx) * st.itemHeight
	itemBottom := itemTop + st.itemHeight
	viewH := st.Height

	pos := st.scrollPos.Peek()
	if itemTop < pos {
		pos = itemTop
	} else if itemBottom > pos+viewH {
		pos = itemBottom - viewH
	}
	st.scrollBar.SetScrollPos(pos)
}
