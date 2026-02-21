package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// SortHandleSide specifies which side of each row the drag handle appears on.
type SortHandleSide int

const (
	SortHandleLeft SortHandleSide = iota
	SortHandleRight
)

// SortableList is a vertical list widget specialized for ordered collections.
// It supports drag-handle-based reordering, keyboard reorder commands, and
// reactive array binding.
type SortableList struct {
	Component
	viewport   *sg.Node
	content    *sg.Node
	scrollBar  *ScrollBar
	itemHeight float64

	// Data binding — set via BindSortableListItems.
	itemCount func() int
	getItem   func(int) any
	moveItem  func(from, to int)

	renderItem func(index int, data any) *Component
	updateItem func(index int, data any, comp *Component)
	selected   *Ref[int]
	scrollPos  *Ref[float64]
	selWatch   WatchHandle
	onChange   func(int)

	selHighlight *sg.Node

	// Row management (non-virtualized).
	rows []*sortableRow

	// Feature flags.
	dragEnabled            bool
	keyboardReorderEnabled bool
	showHandles            bool
	handleSide             SortHandleSide

	// Drag state.
	dragging        bool
	dragFromIndex   int
	dragStartY      float64
	dragCurrentY    float64
	insertTarget    int
	insertIndicator *sg.Node

	// Callbacks.
	onReorder    func(from, to int)
	onMoveDenied func(from, to int)

	// Cleanup handles for reactive array subscriptions.
	arrayWatches []WatchHandle
}

// sortableRow tracks a rendered row and its handle.
type sortableRow struct {
	container   *sg.Node
	handle      *sg.Node
	comp        *Component
	gripNodes   []*sg.Node
	bgNode      *sg.Node   // per-item background (nil if transparent)
	borderNodes []*sg.Node // per-item border edges (nil if no border)
}

// NewSortableList creates a new sortable list with fixed item height.
func NewSortableList(name string, itemHeight float64) *SortableList {
	sl := &SortableList{
		itemHeight:             itemHeight,
		selected:               NewRef(-1),
		scrollPos:              NewRef(0.0),
		dragEnabled:            true,
		keyboardReorderEnabled: true,
		showHandles:            true,
		handleSide:             SortHandleLeft,
		insertTarget:           -1,
	}
	initComponent(&sl.Component, name)
	sl.initBackground(name)
	sl.bgNode.SetColor(sl.EffectiveTheme().SortableList.Group(sl.Variant()).Background.Resolve(StateDefault).Color)

	// Viewport container (clips content).
	sl.viewport = sg.NewContainer(name + "-viewport")
	sl.viewport.Interactable = true
	sl.node.AddChild(sl.viewport)

	// Content container inside viewport.
	sl.content = sg.NewContainer(name + "-content")
	sl.content.Interactable = true
	sl.viewport.AddChild(sl.content)

	// ScrollBar.
	sl.scrollBar = NewScrollBar(name + "-scrollbar")
	sl.scrollBar.SetOnChange(func(pos float64) {
		old := sl.scrollPos.Peek()
		sl.scrollPos.Set(pos)
		DefaultScheduler.Flush()
		if pos != old {
			sl.layoutRows()
		}
	})
	sl.scrollBar.AddToNode(sl.node)

	// Selection highlight.
	sl.selHighlight = sg.NewSprite(name+"-sel-hl", sg.TextureRegion{})
	sl.selHighlight.SetVisible(false)
	sl.selHighlight.SetZIndex(999)
	sl.content.AddChild(sl.selHighlight)

	// Insert indicator (hidden until drag).
	sl.insertIndicator = sg.NewSprite(name+"-insert-ind", sg.TextureRegion{})
	sl.insertIndicator.SetVisible(false)
	sl.insertIndicator.SetZIndex(10)
	sl.content.AddChild(sl.insertIndicator)

	// Auto-update: wheel scrolling, keyboard nav, keyboard reorder.
	sl.node.OnUpdate = func(_ float64) {
		sl.Update()
	}

	// Click on list focuses it.
	sl.node.OnPointerDown(func(_ sg.PointerContext) {
		if sl.enabled {
			DefaultFocusManager.SetFocus(&sl.Component)
		}
	})

	sl.scrollBar.parent = &sl.Component
	sl.onThemeChange = func() {
		sl.applyThemeColors()
		sl.scrollBar.applyThemeColors()
	}

	// Focus registration.
	sl.enableFocusNavigation()
	sl.InterceptArrows = true
	sl.ConsumeHandledKeys = false

	sl.onFocusChange = func(focused bool) { sl.applyThemeColors() }
	sl.SetHandleKey(func(key engine.Key) bool {
		n := sl.itemCountSafe()
		if n == 0 {
			return false
		}
		sel := sl.selected.Peek()
		switch key {
		case engine.KeyUp:
			return sel > 0
		case engine.KeyDown:
			return sel < n-1
		}
		return false
	})

	sl.SetSize(200, 300)
	return sl
}

func (sl *SortableList) applyThemeColors() {
	group := sl.EffectiveTheme().SortableList.Group(sl.Variant())
	sl.bgNode.SetColor(group.Background.Resolve(StateDefault).Color)
	sl.state = computeState(sl.enabled, sl.focused, sl.hovered, false)
	sl.applyFocusRing(group.FocusColor.Resolve(sl.state), group.FocusRingWidth)
	sl.updateHighlight()
	sl.updateInsertIndicatorColor()
	sl.updateHandleColors()
	sl.MarkDrawDirty()
}

func (sl *SortableList) itemCountSafe() int {
	if sl.itemCount == nil {
		return 0
	}
	return sl.itemCount()
}

// SetRenderItem sets the factory function that creates a Component for a
// given item index and data value.
func (sl *SortableList) SetRenderItem(fn func(int, any) *Component) {
	sl.renderItem = fn
	sl.rebuild()
}

// SetUpdateItem sets a function that updates an existing Component for a
// given item index and data value without recreating it.
func (sl *SortableList) SetUpdateItem(fn func(int, any, *Component)) {
	sl.updateItem = fn
}

// SetSize sets the list dimensions and updates internal layout.
func (sl *SortableList) SetSize(w, h float64) {
	sl.Width = w
	sl.Height = h

	sbWidth := float64(DefaultScrollBarWidth)

	sl.resizeBackground(w, h)
	sl.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}

	// ScrollBar.
	sl.scrollBar.SetSize(sbWidth, h)
	sl.scrollBar.SetPosition(w-sbWidth, 0)

	// Viewport.
	sl.viewport.SetPosition(0, 0)

	// Clipping mask.
	maskRoot := sg.NewContainer(sl.node.Name + "-mask")
	maskSprite := sg.NewSprite(sl.node.Name+"-mask-rect", sg.TextureRegion{})
	maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	maskSprite.SetScale(w-sbWidth, h)
	maskRoot.AddChild(maskSprite)
	sl.viewport.SetMask(maskRoot)

	sl.updateScrollBar()
	sl.layoutRows()
	sl.MarkLayoutDirty()
}

// Selected returns the currently selected item index, or -1 if none.
func (sl *SortableList) Selected() int {
	return sl.selected.Peek()
}

// SetSelected sets the selected item index.
func (sl *SortableList) SetSelected(idx int) {
	n := sl.itemCountSafe()
	if idx < -1 || idx >= n {
		return
	}
	old := sl.selected.Peek()
	sl.selected.Set(idx)
	DefaultScheduler.Flush()
	if idx != old && sl.onChange != nil {
		sl.onChange(idx)
	}
	sl.updateHighlight()
	sl.MarkDrawDirty()
	if idx >= 0 {
		sl.ScrollToIndex(idx)
	}
}

// SelectedItem returns the data of the currently selected item, or nil.
func (sl *SortableList) SelectedItem() any {
	idx := sl.selected.Peek()
	if sl.getItem == nil || idx < 0 || idx >= sl.itemCountSafe() {
		return nil
	}
	return sl.getItem(idx)
}

// BindSelected binds the selection to a reactive Ref[int].
func (sl *SortableList) BindSelected(ref *Ref[int]) {
	sl.selected = ref
	bindRef(&sl.selWatch, ref, sl.SetSelected)
}

// SetOnChange sets the callback for selection changes.
func (sl *SortableList) SetOnChange(fn func(int)) {
	sl.onChange = fn
}

// SetDragEnabled enables or disables pointer-based drag reordering.
func (sl *SortableList) SetDragEnabled(v bool) {
	sl.dragEnabled = v
}

// SetKeyboardReorderEnabled enables or disables Alt+Up/Down reordering.
func (sl *SortableList) SetKeyboardReorderEnabled(v bool) {
	sl.keyboardReorderEnabled = v
}

// SetShowHandles shows or hides drag handles.
func (sl *SortableList) SetShowHandles(v bool) {
	sl.showHandles = v
	sl.rebuild()
}

// SetHandleSide sets which side of each row the drag handle appears on.
func (sl *SortableList) SetHandleSide(side SortHandleSide) {
	sl.handleSide = side
	sl.rebuild()
}

// SetOnReorder sets the callback fired after a successful reorder.
func (sl *SortableList) SetOnReorder(fn func(from, to int)) {
	sl.onReorder = fn
}

// SetOnMoveDenied sets the callback fired when a reorder is denied.
func (sl *SortableList) SetOnMoveDenied(fn func(from, to int)) {
	sl.onMoveDenied = fn
}

// MoveItem moves an item from one index to another.
func (sl *SortableList) MoveItem(from, to int) {
	n := sl.itemCountSafe()
	if from < 0 || from >= n || to < 0 || to >= n || from == to {
		return
	}
	if sl.moveItem == nil {
		return
	}

	// Track selection to follow the moved item.
	sel := sl.selected.Peek()
	sl.moveItem(from, to)

	// Update selection to follow the moved item.
	if sel == from {
		sl.selected.Set(to)
		DefaultScheduler.Flush()
	} else if from < to && sel > from && sel <= to {
		sl.selected.Set(sel - 1)
		DefaultScheduler.Flush()
	} else if from > to && sel >= to && sel < from {
		sl.selected.Set(sel + 1)
		DefaultScheduler.Flush()
	}

	if sl.onReorder != nil {
		sl.onReorder(from, to)
	}
}

// MoveSelectedUp moves the selected item up by one position.
func (sl *SortableList) MoveSelectedUp() {
	sel := sl.selected.Peek()
	if sel <= 0 {
		if sl.onMoveDenied != nil && sel == 0 {
			sl.onMoveDenied(0, -1)
		}
		return
	}
	sl.MoveItem(sel, sel-1)
}

// MoveSelectedDown moves the selected item down by one position.
func (sl *SortableList) MoveSelectedDown() {
	sel := sl.selected.Peek()
	n := sl.itemCountSafe()
	if sel < 0 || sel >= n-1 {
		if sl.onMoveDenied != nil && sel == n-1 {
			sl.onMoveDenied(sel, sel+1)
		}
		return
	}
	sl.MoveItem(sel, sel+1)
}

// ScrollToIndex scrolls so that the given item index is visible.
func (sl *SortableList) ScrollToIndex(idx int) {
	n := sl.itemCountSafe()
	if idx < 0 || idx >= n {
		return
	}
	itemTop := float64(idx) * sl.itemHeight
	itemBottom := itemTop + sl.itemHeight
	viewH := sl.Height

	pos := sl.scrollPos.Peek()
	if itemTop < pos {
		pos = itemTop
	} else if itemBottom > pos+viewH {
		pos = itemBottom - viewH
	}
	sl.scrollBar.SetScrollPos(pos)
}

// Update processes mouse wheel input and keyboard navigation/reorder.
func (sl *SortableList) Update() {
	if sl.containsCursor() {
		_, wy := engine.Wheel()
		if wy != 0 {
			newPos := sl.scrollPos.Peek() - wy*scrollWheelSpeedList
			sl.scrollBar.SetScrollPos(newPos)
		}
	}

	if !sl.focused || !sl.enabled {
		return
	}

	im := DefaultInputManager
	sel := sl.selected.Peek()
	n := sl.itemCountSafe()

	// Keyboard navigation: Up/Down move selection.
	if im.IsKeyJustAvailable(engine.KeyUp) && sel > 0 {
		// Check for Alt+Up (reorder).
		if sl.keyboardReorderEnabled && engine.IsKeyPressed(engine.KeyAlt) {
			sl.MoveSelectedUp()
			sl.ScrollToIndex(sl.selected.Peek())
			im.Consume(engine.KeyUp)
			return
		}
		sl.SetSelected(sel - 1)
		sl.ScrollToIndex(sel - 1)
		im.Consume(engine.KeyUp)
	} else if im.IsKeyJustAvailable(engine.KeyDown) && sel < n-1 {
		// Check for Alt+Down (reorder).
		if sl.keyboardReorderEnabled && engine.IsKeyPressed(engine.KeyAlt) {
			sl.MoveSelectedDown()
			sl.ScrollToIndex(sl.selected.Peek())
			im.Consume(engine.KeyDown)
			return
		}
		sl.SetSelected(sel + 1)
		sl.ScrollToIndex(sel + 1)
		im.Consume(engine.KeyDown)
	}
}

// Dispose cleans up watches and child components.
func (sl *SortableList) Dispose() {
	sl.selWatch.Stop()
	for _, w := range sl.arrayWatches {
		w.Stop()
	}
	sl.arrayWatches = nil
	sl.clearRows()
	sl.scrollBar.Dispose()
	sl.Component.Dispose()
}

// ItemCount returns the number of items.
func (sl *SortableList) ItemCount() int {
	return sl.itemCountSafe()
}

// SortableListScrollBar returns the internal scrollbar. Used for testing.
func (sl *SortableList) SortableListScrollBar() *ScrollBar { return sl.scrollBar }

// rebuild clears and recreates all row components.
func (sl *SortableList) rebuild() {
	sl.clearRows()
	if sl.renderItem == nil || sl.itemCount == nil {
		return
	}

	n := sl.itemCount()
	group := sl.EffectiveTheme().SortableList.Group(sl.Variant())
	handleW := group.HandleWidth
	if handleW <= 0 {
		handleW = 24
	}
	handleGap := group.HandleGap
	sbWidth := float64(DefaultScrollBarWidth)
	contentW := sl.Width - sbWidth

	for i := 0; i < n; i++ {
		row := sl.createRow(i, contentW, handleW, handleGap, group)
		sl.rows = append(sl.rows, row)
	}

	sl.updateScrollBar()
	sl.layoutRows()
	sl.updateHighlight()
}

func (sl *SortableList) createRow(index int, contentW, handleW, handleGap float64, group *SortableListGroup) *sortableRow {
	row := &sortableRow{}
	pad := group.ItemPadding

	// Container for the whole row.
	row.container = sg.NewContainer(sl.node.Name + "-row")
	row.container.Interactable = true
	row.container.HitShape = sg.HitRect{X: 0, Y: 0, Width: contentW, Height: sl.itemHeight}

	// Wire click for selection.
	idx := index
	row.container.OnClick(func(_ sg.ClickContext) {
		sl.SetSelected(idx)
	})

	// Per-item background tile.
	itemBg := group.ItemBackground.Resolve(StateDefault)
	if itemBg.Type != BgNone && itemBg.Color != (sg.Color{}) {
		bgSprite := sg.NewSprite(sl.node.Name+"-item-bg", sg.TextureRegion{})
		bgSprite.SetScale(contentW, sl.itemHeight)
		bgSprite.SetColor(itemBg.Color)
		bgSprite.SetZIndex(-2)
		row.bgNode = bgSprite
		row.container.AddChild(bgSprite)
	}

	// Per-item border.
	borderW := group.ItemBorderWidth
	borderColor := group.ItemBorderColor.Resolve(StateDefault)
	zero := sg.Color{}
	if borderW > 0 && borderColor != zero {
		row.borderNodes = sl.buildItemBorder(row.container, contentW, sl.itemHeight, borderW, borderColor)
	}

	// Create the user content.
	var data any
	if sl.getItem != nil {
		data = sl.getItem(index)
	}
	comp := sl.renderItem(index, data)
	row.comp = comp

	// Content area offset by padding.
	contentLeft := pad.Left
	contentTop := pad.Top
	contentAreaH := sl.itemHeight - pad.Top - pad.Bottom

	if sl.showHandles {
		// Create handle.
		handle := sg.NewContainer(sl.node.Name + "-handle")
		handle.Interactable = true
		handle.HitShape = sg.HitRect{X: 0, Y: 0, Width: handleW, Height: sl.itemHeight}
		row.handle = handle

		// Build grip dots on handle.
		row.gripNodes = sl.buildHandleGrip(handle, handleW, sl.itemHeight, group)

		// Wire drag on handle.
		if sl.dragEnabled {
			sl.wireHandleDrag(handle, idx)
		}

		// Hover state for handle.
		handle.OnPointerEnter(func(_ sg.PointerContext) {
			sl.updateGripColor(row, group, true, false)
			if sl.dragEnabled {
				engine.SetCursorShape(engine.CursorShapeNSResize)
			}
		})
		handle.OnPointerLeave(func(_ sg.PointerContext) {
			sl.updateGripColor(row, group, false, false)
			engine.SetCursorShape(engine.CursorShapeDefault)
		})

		// Vertical centering for content within padded area.
		compY := contentTop
		if comp != nil && comp.Height > 0 && comp.Height < contentAreaH {
			compY = contentTop + (contentAreaH-comp.Height)/2
		}

		// Position handle and content based on side.
		if sl.handleSide == SortHandleLeft {
			handle.SetPosition(0, 0)
			if comp != nil {
				comp.SetPosition(handleW+handleGap+contentLeft, compY)
			}
		} else {
			if comp != nil {
				comp.SetPosition(contentLeft, compY)
			}
			handle.SetPosition(contentW-handleW, 0)
		}
		row.container.AddChild(handle)
	} else {
		// No handle — content fills the row.
		if comp != nil {
			compY := contentTop
			if comp.Height > 0 && comp.Height < contentAreaH {
				compY = contentTop + (contentAreaH-comp.Height)/2
			}
			comp.SetPosition(contentLeft, compY)
		}
	}

	if comp != nil {
		row.container.AddChild(comp.Node())
	}

	sl.content.AddChild(row.container)
	return row
}

func (sl *SortableList) wireHandleDrag(handle *sg.Node, index int) {
	handle.OnDragStart(func(ctx sg.DragContext) {
		if !sl.enabled || !sl.dragEnabled {
			return
		}
		sl.dragging = true
		sl.dragFromIndex = index
		sl.dragStartY = ctx.GlobalY
		sl.dragCurrentY = ctx.GlobalY
		sl.SetSelected(index)
		sl.updateInsertTarget()
	})

	handle.OnDrag(func(ctx sg.DragContext) {
		if !sl.dragging {
			return
		}
		sl.dragCurrentY = ctx.GlobalY
		sl.updateInsertTarget()
	})

	handle.OnDragEnd(func(ctx sg.DragContext) {
		if !sl.dragging {
			return
		}
		sl.dragging = false
		target := sl.insertTarget
		sl.insertTarget = -1
		sl.insertIndicator.SetVisible(false)

		if target >= 0 && target != sl.dragFromIndex {
			sl.MoveItem(sl.dragFromIndex, target)
			sl.SetSelected(target)
		} else {
			// No move happened — still select the dragged item.
			sl.SetSelected(sl.dragFromIndex)
		}
	})
}

func (sl *SortableList) updateInsertTarget() {
	if !sl.dragging {
		return
	}

	// Convert global Y to widget-local Y.
	_, localY := sl.node.WorldToLocal(0, sl.dragCurrentY)
	localY += sl.scrollPos.Peek()

	n := sl.itemCountSafe()
	target := int(math.Floor(localY / sl.itemHeight))

	// Clamp to valid range.
	if target < 0 {
		target = 0
	}
	if target >= n {
		target = n - 1
	}

	sl.insertTarget = target

	// Position the insert indicator.
	if target != sl.dragFromIndex {
		sl.insertIndicator.SetVisible(true)
		sbWidth := float64(DefaultScrollBarWidth)
		indicatorY := float64(target) * sl.itemHeight
		if target > sl.dragFromIndex {
			indicatorY += sl.itemHeight
		}
		group := sl.EffectiveTheme().SortableList.Group(sl.Variant())
		indicatorH := group.InsertIndicatorWidth
		if indicatorH <= 0 {
			indicatorH = 2
		}
		sl.insertIndicator.SetPosition(0, indicatorY-indicatorH/2)
		sl.insertIndicator.SetScale(sl.Width-sbWidth, indicatorH)
	} else {
		sl.insertIndicator.SetVisible(false)
	}
}

func (sl *SortableList) buildHandleGrip(handle *sg.Node, handleW, handleH float64, group *SortableListGroup) []*sg.Node {
	dotSize := 3.0
	spacing := 4.0
	cols := 2
	rows := 3

	totalW := float64(cols)*dotSize + float64(cols-1)*spacing
	totalH := float64(rows)*dotSize + float64(rows-1)*spacing
	startX := (handleW - totalW) / 2
	startY := (handleH - totalH) / 2

	gripColor := group.HandleColor.Resolve(StateDefault)
	var nodes []*sg.Node
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			dot := sg.NewSprite(handle.Name+"-grip-dot", sg.TextureRegion{})
			dot.SetScale(dotSize, dotSize)
			x := startX + float64(c)*(dotSize+spacing)
			y := startY + float64(r)*(dotSize+spacing)
			dot.SetPosition(x, y)
			dot.SetColor(gripColor)
			handle.AddChild(dot)
			nodes = append(nodes, dot)
		}
	}
	return nodes
}

func (sl *SortableList) buildItemBorder(parent *sg.Node, w, h, bw float64, color sg.Color) []*sg.Node {
	var nodes []*sg.Node
	// Top edge.
	top := sg.NewSprite(parent.Name+"-border-t", sg.TextureRegion{})
	top.SetScale(w, bw)
	top.SetPosition(0, 0)
	top.SetColor(color)
	top.SetZIndex(-1)
	parent.AddChild(top)
	nodes = append(nodes, top)
	// Bottom edge.
	bot := sg.NewSprite(parent.Name+"-border-b", sg.TextureRegion{})
	bot.SetScale(w, bw)
	bot.SetPosition(0, h-bw)
	bot.SetColor(color)
	bot.SetZIndex(-1)
	parent.AddChild(bot)
	nodes = append(nodes, bot)
	// Left edge.
	left := sg.NewSprite(parent.Name+"-border-l", sg.TextureRegion{})
	left.SetScale(bw, h)
	left.SetPosition(0, 0)
	left.SetColor(color)
	left.SetZIndex(-1)
	parent.AddChild(left)
	nodes = append(nodes, left)
	// Right edge.
	right := sg.NewSprite(parent.Name+"-border-r", sg.TextureRegion{})
	right.SetScale(bw, h)
	right.SetPosition(w-bw, 0)
	right.SetColor(color)
	right.SetZIndex(-1)
	parent.AddChild(right)
	nodes = append(nodes, right)
	return nodes
}

func (sl *SortableList) updateGripColor(row *sortableRow, group *SortableListGroup, hovered, active bool) {
	var color sg.Color
	if active {
		color = group.HandleActiveColor.Resolve(StateDefault)
	} else if hovered {
		color = group.HandleHoverColor.Resolve(StateDefault)
	} else {
		color = group.HandleColor.Resolve(StateDefault)
	}
	for _, n := range row.gripNodes {
		n.SetColor(color)
	}
}

func (sl *SortableList) updateHandleColors() {
	group := sl.EffectiveTheme().SortableList.Group(sl.Variant())
	for _, row := range sl.rows {
		if row.gripNodes != nil {
			sl.updateGripColor(row, group, false, false)
		}
	}
}

func (sl *SortableList) updateHighlight() {
	idx := sl.selected.Peek()
	n := sl.itemCountSafe()
	if idx < 0 || idx >= n {
		sl.selHighlight.SetVisible(false)
		return
	}
	sbWidth := float64(DefaultScrollBarWidth)
	itemW := sl.Width - sbWidth
	y := float64(idx) * sl.itemHeight
	sl.selHighlight.SetPosition(0, y)
	sl.selHighlight.SetScale(itemW, sl.itemHeight)
	group := sl.EffectiveTheme().SortableList.Group(sl.Variant())
	sl.selHighlight.SetColor(group.SelectionColor.Resolve(StateDefault))
	sl.selHighlight.SetVisible(true)
	sl.MarkDrawDirty()
}

func (sl *SortableList) updateInsertIndicatorColor() {
	group := sl.EffectiveTheme().SortableList.Group(sl.Variant())
	sl.insertIndicator.SetColor(group.InsertIndicatorColor.Resolve(StateDefault))
}

func (sl *SortableList) updateScrollBar() {
	n := sl.itemCountSafe()
	totalH := float64(n) * sl.itemHeight
	sl.scrollBar.SetContentSize(totalH, sl.Height)
	sl.scrollBar.SetVisible(totalH > sl.Height)
}

func (sl *SortableList) layoutRows() {
	pos := sl.scrollPos.Peek()
	sl.content.SetPosition(0, -pos)

	for i, row := range sl.rows {
		y := float64(i) * sl.itemHeight
		row.container.SetPosition(0, y)
	}

	sl.updateHighlight()
}

func (sl *SortableList) clearRows() {
	for _, row := range sl.rows {
		sl.content.RemoveChild(row.container)
		if row.comp != nil {
			row.comp.Dispose()
		}
	}
	sl.rows = nil
}

// BindSortableListItems binds a reactive Array[T] to a SortableList.
// This is a package-level generic function because Go does not support
// generic methods.
func BindSortableListItems[T any](sl *SortableList, items *Array[T]) {
	// Stop previous watches.
	for _, w := range sl.arrayWatches {
		w.Stop()
	}
	sl.arrayWatches = nil

	sl.itemCount = func() int { return items.Len() }
	sl.getItem = func(i int) any { return items.At(i) }
	sl.moveItem = func(from, to int) { items.Move(from, to) }

	// Subscribe to array changes for full rebuild.
	w1 := items.OnChange(func() {
		sl.rebuild()
	})
	sl.arrayWatches = append(sl.arrayWatches, w1)

	sl.rebuild()
}
