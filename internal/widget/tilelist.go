package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// TileList is a grid-layout variant of List. Items are arranged in rows and
// columns, and only visible rows are rendered (virtualized).
type TileList struct {
	Component
	viewport   *sg.Node
	content    *sg.Node
	scrollBar  *ScrollBar
	items      []ListItem
	tileWidth  float64
	tileHeight float64
	columns    int // 0 = auto-fit to width
	renderItem func(int, any) *Component
	updateItem func(int, any, *Component) // optional in-place update callback
	selected   *Ref[int]
	scrollPos  *Ref[float64]
	selWatch   WatchHandle
	stopArray  func() // stops reactive Array binding callbacks
	onChange   func(int)

	selectable   bool
	selHighlight *sg.Node

	// ItemAlign controls horizontal alignment of each rendered item within
	// its tile cell. Default is AlignCenter.
	ItemAlign Alignment
	// ItemVAlign controls vertical alignment of each rendered item within
	// its tile cell. Default is AlignCenter.
	ItemVAlign Alignment

	effectiveCols int
	visibleStart  int
	visibleEnd    int
	pool          []*Component
	poolIndex     map[int]int
	freeSlots     []int // free-list for O(1) pool slot reuse
}

// NewTileList creates a new tile list with the given tile dimensions.
func NewTileList(name string, tileW, tileH float64) *TileList {
	tl := &TileList{
		tileWidth:  tileW,
		tileHeight: tileH,
		selected:   NewRef(-1),
		scrollPos:  NewRef(0.0),
		poolIndex:  make(map[int]int),
		ItemAlign:  AlignCenter,
		ItemVAlign: AlignCenter,
	}
	initComponent(&tl.Component, name)

	tl.initBackground(name)
	tl.bgNode.SetColor(tl.EffectiveTheme().TileList.Group(tl.Variant()).Background.Resolve(StateDefault).Color)

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

	// Auto-update: mouse wheel scrolling and keyboard activation.
	tl.node.OnUpdate = func(_ float64) {
		tl.Update()

		// Keyboard activation: Enter confirms selection.
		if tl.focused && tl.enabled {
			if DefaultInputManager.IsKeyJustAvailable(engine.KeyEnter) {
				if tl.onChange != nil && tl.selected.Peek() >= 0 {
					tl.onChange(tl.selected.Peek())
				}
				DefaultInputManager.Consume(engine.KeyEnter)
			}
		}
	}

	// Click on tile list focuses it.
	tl.node.OnPointerDown(func(_ sg.PointerContext) {
		if tl.enabled {
			DefaultFocusManager.SetFocus(&tl.Component)
		}
	})

	tl.scrollBar.parent = &tl.Component
	tl.onThemeChange = func() {
		tl.applyThemeColors()
		tl.scrollBar.applyThemeColors()
	}

	// Focus: tile lists participate in tab and spatial nav, intercept arrows.
	tl.enableFocusNavigation()
	tl.InterceptArrows = true
	tl.ConsumeHandledKeys = false // widget's own Update handles the key; only block spatial nav

	tl.onFocusChange = func(focused bool) { tl.applyThemeColors() }
	tl.SetHandleKey(func(key engine.Key) bool {
		n := len(tl.items)
		if n == 0 {
			return false
		}
		sel := tl.selected.Peek()
		switch key {
		case engine.KeyUp:
			return sel > 0
		case engine.KeyDown:
			return sel < n-1
		}
		return false
	})

	// Default size.
	tl.SetSize(300, 300)

	return tl
}

func (tl *TileList) applyThemeColors() {
	group := tl.EffectiveTheme().TileList.Group(tl.Variant())
	tl.bgNode.SetColor(group.Background.Resolve(StateDefault).Color)
	tl.state = computeState(tl.enabled, tl.focused, tl.hovered, false)
	tl.applyFocusRing(group.FocusColor.Resolve(tl.state), group.FocusRingWidth)
	tl.updateHighlight()
	tl.MarkDrawDirty()
}

// SetItems replaces the tile list data and refreshes visible tiles.
// If SetUpdateItem has been configured, existing pool components are updated
// in place rather than destroyed and recreated, saving node allocation cost.
func (tl *TileList) SetItems(items []ListItem) {
	tl.items = items
	tl.scrollPos.Set(0)
	DefaultScheduler.Flush()
	tl.recalcColumns()
	tl.updateScrollBar()
	if tl.updateItem != nil && len(tl.poolIndex) > 0 {
		tl.refreshPool()
	} else {
		tl.clearPool()
	}
	tl.updateVisible()
}

// SetUpdateItem registers an optional callback for in-place component updates.
// When set, SetItems will call fn for each currently pooled tile instead of
// disposing and recreating it. Use this to update only the data-driven parts of
// a tile component (text, colors, progress bars) while keeping the node tree
// structure intact.
//
//	tileList.SetUpdateItem(func(idx int, data any, comp *ui.Component) {
//	    spell := data.(Spell)
//	    // update name label, color bar, etc. on comp
//	})
func (tl *TileList) SetUpdateItem(fn func(int, any, *Component)) {
	tl.updateItem = fn
}

// BindItems binds the tile list to a reactive Array[ListItem]. Any mutation
// to the array is automatically reflected in the tile list without resetting
// the scroll position.
//
// Pass nil to detach the current binding.
func (tl *TileList) BindItems(arr *Array[ListItem]) {
	if tl.stopArray != nil {
		tl.stopArray()
		tl.stopArray = nil
	}
	if arr == nil {
		tl.items = nil
		tl.clearPool()
		tl.recalcColumns()
		tl.updateScrollBar()
		tl.updateVisible()
		return
	}

	// Snapshot current array contents.
	tl.items = tl.items[:0]
	arr.ForEach(func(_ int, item ListItem) {
		tl.items = append(tl.items, item)
	})
	tl.clearPool()
	tl.recalcColumns()
	tl.updateScrollBar()
	tl.updateVisible()

	// OnAdded: insert item then invalidate pool entries at indices >= insertion
	// point (their grid positions and click handlers are now stale).
	h1 := arr.OnAdded(func(idx int, item ListItem) {
		tl.items = append(tl.items, ListItem{})
		copy(tl.items[idx+1:], tl.items[idx:])
		tl.items[idx] = item
		for k := range tl.poolIndex {
			if k >= idx {
				tl.removePoolItem(k)
			}
		}
		tl.updateScrollBar()
		tl.updateVisible()
	})

	// OnRemoved: invalidate pool entries at and above the removed index.
	h2 := arr.OnRemoved(func(idx int, _ ListItem) {
		for k := range tl.poolIndex {
			if k >= idx {
				tl.removePoolItem(k)
			}
		}
		tl.items = append(tl.items[:idx], tl.items[idx+1:]...)
		tl.updateScrollBar()
		tl.updateVisible()
	})

	// OnReplaced/OnMoved: bulk operations require a full pool rebuild.
	syncAll := func() {
		tl.items = tl.items[:0]
		arr.ForEach(func(_ int, item ListItem) {
			tl.items = append(tl.items, item)
		})
		tl.clearPool()
		tl.recalcColumns()
		tl.updateScrollBar()
		tl.updateVisible()
	}
	h3 := arr.OnReplaced(syncAll)
	h4 := arr.OnMoved(func(_, _ int) { syncAll() })

	tl.stopArray = func() {
		h1.Stop()
		h2.Stop()
		h3.Stop()
		h4.Stop()
	}
}

// SetColumns sets the number of columns. 0 means auto-fit to width.
func (tl *TileList) SetColumns(n int) {
	tl.columns = n
	tl.recalcColumns()
	tl.updateScrollBar()
	tl.updateVisible()
}

// SetRenderItem sets the factory function for tile components.
func (tl *TileList) SetRenderItem(fn func(int, any) *Component) {
	tl.renderItem = fn
	tl.updateVisible()
}

// Selected returns the currently selected tile index.
func (tl *TileList) Selected() int {
	return tl.selected.Peek()
}

// SetSelected sets the selected tile index.
func (tl *TileList) SetSelected(idx int) {
	if idx < -1 || idx >= len(tl.items) {
		return
	}
	old := tl.selected.Peek()
	tl.selected.Set(idx)
	DefaultScheduler.Flush()
	if idx != old && tl.onChange != nil {
		tl.onChange(idx)
	}
	tl.updateHighlight()
	tl.MarkDrawDirty()
}

// BindSelected binds the selection to a reactive Ref[int].
func (tl *TileList) BindSelected(ref *Ref[int]) {
	tl.selected = ref
	bindRef(&tl.selWatch, ref, tl.SetSelected)
}

// SelectedRef returns the reactive Ref backing the selection index.
func (tl *TileList) SelectedRef() *Ref[int] {
	return tl.selected
}

// SelectedItem returns the data of the currently selected tile, or nil if
// no tile is selected.
func (tl *TileList) SelectedItem() any {
	idx := tl.selected.Peek()
	if idx < 0 || idx >= len(tl.items) {
		return nil
	}
	return tl.items[idx].Data
}

// ClearSelection deselects the current tile.
func (tl *TileList) ClearSelection() {
	tl.SetSelected(-1)
}

// SelectNext moves the selection to the next tile. If nothing is selected,
// selects the first tile. Scrolls to keep the selection visible.
func (tl *TileList) SelectNext() {
	if len(tl.items) == 0 {
		return
	}
	idx := tl.selected.Peek() + 1
	if idx >= len(tl.items) {
		idx = len(tl.items) - 1
	}
	tl.SetSelected(idx)
	tl.ScrollToSelection()
}

// SelectPrevious moves the selection to the previous tile. If nothing is
// selected, selects the last tile. Scrolls to keep the selection visible.
func (tl *TileList) SelectPrevious() {
	if len(tl.items) == 0 {
		return
	}
	idx := tl.selected.Peek()
	if idx < 0 {
		idx = len(tl.items) - 1
	} else if idx > 0 {
		idx--
	}
	tl.SetSelected(idx)
	tl.ScrollToSelection()
}

// SelectFirst selects the first tile and scrolls to it.
func (tl *TileList) SelectFirst() {
	if len(tl.items) == 0 {
		return
	}
	tl.SetSelected(0)
	tl.ScrollToSelection()
}

// SelectLast selects the last tile and scrolls to it.
func (tl *TileList) SelectLast() {
	if len(tl.items) == 0 {
		return
	}
	tl.SetSelected(len(tl.items) - 1)
	tl.ScrollToSelection()
}

// ScrollToIndex scrolls so that the given tile index is visible.
func (tl *TileList) ScrollToIndex(idx int) {
	if idx < 0 || idx >= len(tl.items) || tl.effectiveCols <= 0 {
		return
	}
	row := idx / tl.effectiveCols
	rowTop := float64(row) * tl.tileHeight
	rowBottom := rowTop + tl.tileHeight
	viewH := tl.Height

	pos := tl.scrollPos.Peek()
	if rowTop < pos {
		pos = rowTop
	} else if rowBottom > pos+viewH {
		pos = rowBottom - viewH
	}
	tl.scrollBar.SetScrollPos(pos)
}

// ScrollToSelection scrolls so that the currently selected tile is visible.
func (tl *TileList) ScrollToSelection() {
	tl.ScrollToIndex(tl.selected.Peek())
}

// Items returns the current tile list items.
func (tl *TileList) Items() []ListItem {
	return tl.items
}

// SetOnChange sets the callback for selection changes.
func (tl *TileList) SetOnChange(fn func(int)) {
	tl.onChange = fn
}

// SetSelectable enables or disables built-in selection highlighting.
func (tl *TileList) SetSelectable(enabled bool) {
	tl.selectable = enabled
	if enabled && tl.selHighlight == nil {
		tl.selHighlight = sg.NewSprite(tl.node.Name+"-sel-hl", sg.TextureRegion{})
		tl.selHighlight.SetVisible(false)
		tl.selHighlight.SetZIndex(999)
		tl.content.AddChild(tl.selHighlight)
	}
	tl.updateHighlight()
}

// Selectable returns whether built-in selection highlighting is enabled.
func (tl *TileList) Selectable() bool {
	return tl.selectable
}

func (tl *TileList) updateHighlight() {
	if tl.selHighlight == nil {
		return
	}
	idx := tl.selected.Peek()
	if !tl.selectable || idx < 0 || idx >= len(tl.items) || tl.effectiveCols <= 0 {
		tl.selHighlight.SetVisible(false)
		return
	}
	row := idx / tl.effectiveCols
	col := idx % tl.effectiveCols
	x := float64(col) * tl.tileWidth
	y := float64(row) * tl.tileHeight
	tl.selHighlight.SetPosition(x, y)
	tl.selHighlight.SetScale(tl.tileWidth, tl.tileHeight)
	bg := tl.EffectiveTheme().TileList.Group(tl.Variant()).ItemBackground.Resolve(StateDefault)
	tl.selHighlight.SetColor(bg.Color)
	tl.selHighlight.SetVisible(true)
	tl.MarkDrawDirty()
}

// Update processes mouse wheel input and keyboard navigation for scrolling.
// This is called automatically via the willow node's OnUpdate hook; no
// manual call needed.
func (tl *TileList) Update() {
	if tl.containsCursor() {
		_, wy := engine.Wheel()
		if wy != 0 {
			newPos := tl.scrollPos.Peek() - wy*scrollWheelSpeedList
			tl.scrollBar.SetScrollPos(newPos)
		}
	}

	// Keyboard navigation: Up/Down move selection when focused.
	if tl.focused && tl.enabled {
		im := DefaultInputManager
		sel := tl.selected.Peek()
		n := len(tl.items)
		if im.IsKeyJustAvailable(engine.KeyUp) && sel > 0 {
			tl.SetSelected(sel - 1)
			im.Consume(engine.KeyUp)
		} else if im.IsKeyJustAvailable(engine.KeyDown) && sel < n-1 {
			tl.SetSelected(sel + 1)
			im.Consume(engine.KeyDown)
		}
	}
}

// SetSize sets the tile list dimensions.
func (tl *TileList) SetSize(w, h float64) {
	tl.Width = w
	tl.Height = h

	sbWidth := float64(DefaultScrollBarWidth)

	tl.resizeBackground(w, h)
	tl.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}

	// ScrollBar.
	tl.scrollBar.SetSize(sbWidth, h)
	tl.scrollBar.SetPosition(w-sbWidth, 0)

	// Clipping mask so tiles don't render outside bounds.
	maskRoot := sg.NewContainer(tl.node.Name + "-mask")
	maskSprite := sg.NewSprite(tl.node.Name+"-mask-rect", sg.TextureRegion{})
	maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	maskSprite.SetScale(w-sbWidth, h)
	maskRoot.AddChild(maskSprite)
	tl.viewport.SetMask(maskRoot)

	tl.recalcColumns()
	tl.updateScrollBar()
	tl.updateVisible()
	tl.MarkLayoutDirty()
}

// Dispose cleans up the tile list.
func (tl *TileList) Dispose() {
	if tl.stopArray != nil {
		tl.stopArray()
	}
	tl.selWatch.Stop()
	tl.clearPool()
	tl.scrollBar.Dispose()
	tl.Component.Dispose()
}

// EffectiveColumns returns the computed number of columns.
func (tl *TileList) EffectiveColumns() int {
	return tl.effectiveCols
}

// TileScrollBar returns the internal scrollbar. Used for testing.
func (tl *TileList) TileScrollBar() *ScrollBar { return tl.scrollBar }

// TileScrollPos returns the reactive scroll position ref. Used for testing.
func (tl *TileList) TileScrollPos() *Ref[float64] { return tl.scrollPos }

// PoolSize returns active component count for testing virtualization.
func (tl *TileList) PoolSize() int {
	count := 0
	for _, p := range tl.pool {
		if p != nil {
			count++
		}
	}
	return count
}

func (tl *TileList) recalcColumns() {
	if tl.columns > 0 {
		tl.effectiveCols = tl.columns
	} else {
		contentW := tl.Width - float64(DefaultScrollBarWidth)
		if tl.tileWidth > 0 {
			tl.effectiveCols = int(math.Floor(contentW / tl.tileWidth))
		}
		if tl.effectiveCols < 1 {
			tl.effectiveCols = 1
		}
	}
}

func (tl *TileList) totalRows() int {
	if tl.effectiveCols <= 0 {
		return 0
	}
	return int(math.Ceil(float64(len(tl.items)) / float64(tl.effectiveCols)))
}

func (tl *TileList) updateScrollBar() {
	totalH := float64(tl.totalRows()) * tl.tileHeight
	tl.scrollBar.SetContentSize(totalH, tl.Height)
	tl.scrollBar.SetVisible(totalH > tl.Height)
}

func (tl *TileList) updateVisible() {
	if tl.renderItem == nil || len(tl.items) == 0 || tl.effectiveCols <= 0 {
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
		keepBuffer   = 8
	)

	renderStartRow := int(math.Floor(pos/tl.tileHeight)) - renderBuffer
	if renderStartRow < 0 {
		renderStartRow = 0
	}
	renderEndRow := renderStartRow + int(math.Ceil(viewH/tl.tileHeight)) + 1 + 2*renderBuffer
	totalR := tl.totalRows()
	if renderEndRow > totalR {
		renderEndRow = totalR
	}

	keepStartRow := int(math.Floor(pos/tl.tileHeight)) - keepBuffer
	if keepStartRow < 0 {
		keepStartRow = 0
	}
	keepEndRow := keepStartRow + int(math.Ceil(viewH/tl.tileHeight)) + 1 + 2*keepBuffer
	if keepEndRow > totalR {
		keepEndRow = totalR
	}

	renderStart := renderStartRow * tl.effectiveCols
	renderEnd := renderEndRow * tl.effectiveCols
	if renderEnd > len(tl.items) {
		renderEnd = len(tl.items)
	}
	keepStart := keepStartRow * tl.effectiveCols
	keepEnd := keepEndRow * tl.effectiveCols
	if keepEnd > len(tl.items) {
		keepEnd = len(tl.items)
	}

	// Phase 1: Add newly visible items FIRST.
	for i := renderStart; i < renderEnd; i++ {
		if _, exists := tl.poolIndex[i]; !exists {
			row := i / tl.effectiveCols
			col := i % tl.effectiveCols
			x := float64(col) * tl.tileWidth
			y := float64(row) * tl.tileHeight
			tl.addPoolItem(i, x, y)
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

	// Clip hit shapes to the visible portion of each tile so that tiles
	// partially scrolled above/below the viewport don't intercept clicks
	// on elements outside the list.
	for idx, slot := range tl.poolIndex {
		if slot < len(tl.pool) && tl.pool[slot] != nil {
			comp := tl.pool[slot]
			row := idx / tl.effectiveCols
			cellY := float64(row) * tl.tileHeight
			offX, offY := tl.tileItemOffset(comp)
			topVisible := cellY+tl.tileHeight > pos
			bottomVisible := cellY < pos+viewH
			if topVisible && bottomVisible {
				comp.node.Interactable = true
				visTop := math.Max(cellY, pos)
				visBot := math.Min(cellY+tl.tileHeight, pos+viewH)
				comp.node.HitShape = sg.HitRect{
					X:      -offX,
					Y:      visTop - (cellY + offY),
					Width:  tl.tileWidth,
					Height: visBot - visTop,
				}
			} else {
				comp.node.Interactable = false
			}
		}
	}

	tl.updateHighlight()
}

// tileItemOffset returns the (x, y) offset of a rendered component within its
// tile cell based on the tile list's ItemAlign and ItemVAlign settings.
func (tl *TileList) tileItemOffset(comp *Component) (float64, float64) {
	var offX, offY float64
	switch tl.ItemAlign {
	case AlignCenter:
		offX = (tl.tileWidth - comp.Width) / 2
	case AlignEnd:
		offX = tl.tileWidth - comp.Width
	}
	switch tl.ItemVAlign {
	case AlignCenter:
		offY = (tl.tileHeight - comp.Height) / 2
	case AlignEnd:
		offY = tl.tileHeight - comp.Height
	}
	return offX, offY
}

func (tl *TileList) addPoolItem(index int, x, y float64) {
	comp := tl.renderItem(index, tl.items[index].Data)
	if comp == nil {
		return
	}

	offX, offY := tl.tileItemOffset(comp)
	comp.SetPosition(x+offX, y+offY)
	comp.SetHitShape(sg.HitRect{X: 0, Y: 0, Width: comp.Width, Height: comp.Height})

	idx := index
	comp.SetInteractable(true)
	comp.OnClick(func(ctx sg.ClickContext) {
		tl.SetSelected(idx)
	})

	tl.content.AddChild(comp.Node())

	var slot int
	if n := len(tl.freeSlots); n > 0 {
		slot = tl.freeSlots[n-1]
		tl.freeSlots = tl.freeSlots[:n-1]
		tl.pool[slot] = comp
	} else {
		slot = len(tl.pool)
		tl.pool = append(tl.pool, comp)
	}
	tl.poolIndex[index] = slot
}

func (tl *TileList) removePoolItem(index int) {
	slot, ok := tl.poolIndex[index]
	if !ok {
		return
	}
	if slot < len(tl.pool) && tl.pool[slot] != nil {
		tl.content.RemoveChild(tl.pool[slot].Node())
		tl.pool[slot].Dispose()
		tl.pool[slot] = nil
		tl.freeSlots = append(tl.freeSlots, slot)
	}
	delete(tl.poolIndex, index)
}

func (tl *TileList) clearPool() {
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
	tl.freeSlots = tl.freeSlots[:0]
}

// refreshPool updates all currently pooled components with new item data,
// removing any pooled indices that are now out of bounds. Called by SetItems
// when updateItem is set, to avoid full pool teardown.
func (tl *TileList) refreshPool() {
	// Remove pooled items whose index no longer exists in the new item set.
	for idx := range tl.poolIndex {
		if idx >= len(tl.items) {
			tl.removePoolItem(idx)
		}
	}
	// Update surviving pool entries with their new data.
	for idx, slot := range tl.poolIndex {
		if slot < len(tl.pool) && tl.pool[slot] != nil {
			tl.updateItem(idx, tl.items[idx].Data, tl.pool[slot])
		}
	}
}
