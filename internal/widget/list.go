package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// ListItem wraps arbitrary data for use in a List.
type ListItem struct {
	Data any
}

// scrollWheelSpeedList is the pixel distance per wheel tick for lists.
const scrollWheelSpeedList = 40

// List is a virtualized vertical scrollable list. Only items visible in the
// viewport are rendered, and components are recycled as the user scrolls.
type List struct {
	Component
	viewport   *sg.Node // clipping container
	content    *sg.Node // positioned content container
	scrollBar  *ScrollBar
	items      []ListItem
	itemHeight float64
	renderItem func(index int, data any) *Component
	selected   *Ref[int]
	scrollPos  *Ref[float64]
	watch      WatchHandle
	selWatch   WatchHandle
	stopArray  func() // stops reactive Array binding callbacks
	onChange   func(int)

	selectable   bool
	selHighlight *sg.Node

	// ItemAlign controls horizontal alignment of each rendered item within
	// its cell. Default is AlignStart (left-aligned).
	ItemAlign Alignment
	// ItemVAlign controls vertical alignment of each rendered item within
	// its cell. Default is AlignCenter.
	ItemVAlign Alignment

	visibleStart int
	visibleEnd   int
	pool         []*Component
	poolIndex    map[int]int // maps item index -> pool slot
	maskSprite   *sg.Node    // reused across SetSize calls to avoid node accumulation
}

// NewList creates a new virtualized list with fixed item height.
func NewList(name string, itemHeight float64) *List {
	l := &List{
		itemHeight: itemHeight,
		selected:   NewRef(-1),
		scrollPos:  NewRef(0.0),
		poolIndex:  make(map[int]int),
		ItemAlign:  AlignStart,
		ItemVAlign: AlignCenter,
	}
	initComponent(&l.Component, name)

	l.initBackground(name)
	l.bgNode.SetColor(l.EffectiveTheme().List.Group(l.Variant()).Background.Resolve(StateDefault).Color)

	// Viewport container (clips content).
	l.viewport = sg.NewContainer(name + "-viewport")
	l.viewport.Interactable = true
	l.node.AddChild(l.viewport)

	// Content container inside viewport.
	l.content = sg.NewContainer(name + "-content")
	l.content.Interactable = true
	l.viewport.AddChild(l.content)

	// ScrollBar.
	l.scrollBar = NewScrollBar(name + "-scrollbar")
	l.scrollBar.SetOnChange(func(pos float64) {
		old := l.scrollPos.Peek()
		l.scrollPos.Set(pos)
		DefaultScheduler.Flush()
		if pos != old {
			l.updateVisible()
		}
	})
	l.scrollBar.AddToNode(l.node)

	// Auto-update: mouse wheel scrolling and keyboard activation.
	l.node.OnUpdate = func(_ float64) {
		l.Update()

		// Keyboard activation: Enter confirms selection.
		if l.focused && l.enabled && l.selectable {
			if DefaultInputManager.IsKeyJustAvailable(engine.KeyEnter) {
				if l.onChange != nil && l.selected.Peek() >= 0 {
					l.onChange(l.selected.Peek())
				}
				DefaultInputManager.Consume(engine.KeyEnter)
			}
		}
	}

	// Click on list focuses it.
	l.node.OnPointerDown(func(_ sg.PointerContext) {
		if l.enabled {
			DefaultFocusManager.SetFocus(&l.Component)
		}
	})

	l.scrollBar.parent = &l.Component
	l.onThemeChange = func() {
		l.applyThemeColors()
		l.scrollBar.applyThemeColors()
	}

	// Focus: lists participate in tab and spatial nav, intercept arrows.
	l.enableFocusNavigation()
	l.InterceptArrows = true
	l.ConsumeHandledKeys = false // widget's own Update handles the key; only block spatial nav

	l.onFocusChange = func(focused bool) { l.applyThemeColors() }
	l.SetHandleKey(func(key engine.Key) bool {
		n := len(l.items)
		if n == 0 {
			return false
		}
		sel := l.selected.Peek()
		switch key {
		case engine.KeyUp:
			return sel > 0
		case engine.KeyDown:
			return sel < n-1
		}
		return false
	})

	// Default size.
	l.SetSize(200, 300)

	return l
}

func (l *List) applyThemeColors() {
	group := l.EffectiveTheme().List.Group(l.Variant())
	l.bgNode.SetColor(group.Background.Resolve(StateDefault).Color)
	l.state = computeState(l.enabled, l.focused, l.hovered, false)
	l.applyFocusRing(group.FocusColor.Resolve(l.state), group.FocusRingWidth)
	l.updateHighlight()
	l.MarkDrawDirty()
}

// SetItems replaces the list data and refreshes the visible items.
func (l *List) SetItems(items []ListItem) {
	l.items = items
	l.scrollPos.Set(0)
	l.clearPool()
	DefaultScheduler.Flush()
	l.updateScrollBar()
	l.updateVisible()
}

// BindItems binds the list to a reactive Array[ListItem]. Any mutation to the
// array (Push, Remove, Sort, etc.) is automatically reflected in the list
// without resetting the scroll position.
//
// Pass nil to detach the current binding.
func (l *List) BindItems(arr *Array[ListItem]) {
	if l.stopArray != nil {
		l.stopArray()
		l.stopArray = nil
	}
	if arr == nil {
		l.items = nil
		l.clearPool()
		l.updateScrollBar()
		l.updateVisible()
		return
	}

	// Snapshot current array contents.
	l.items = l.items[:0]
	arr.ForEach(func(_ int, item ListItem) {
		l.items = append(l.items, item)
	})
	l.clearPool()
	l.updateScrollBar()
	l.updateVisible()

	// OnAdded: insert item then invalidate pool entries whose click handlers
	// closed over old indices (all entries at index >= insertion point).
	h1 := arr.OnAdded(func(idx int, item ListItem) {
		l.items = append(l.items, ListItem{})
		copy(l.items[idx+1:], l.items[idx:])
		l.items[idx] = item
		for k := range l.poolIndex {
			if k >= idx {
				l.removePoolItem(k)
			}
		}
		l.updateScrollBar()
		l.updateVisible()
	})

	// OnRemoved: invalidate pool entries at and above the removed index before
	// splicing l.items, so updateVisible recreates them at correct positions.
	h2 := arr.OnRemoved(func(idx int, _ ListItem) {
		for k := range l.poolIndex {
			if k >= idx {
				l.removePoolItem(k)
			}
		}
		l.items = append(l.items[:idx], l.items[idx+1:]...)
		l.updateScrollBar()
		l.updateVisible()
	})

	// OnReplaced/OnMoved: bulk operations require a full pool rebuild.
	syncAll := func() {
		l.items = l.items[:0]
		arr.ForEach(func(_ int, item ListItem) {
			l.items = append(l.items, item)
		})
		l.clearPool()
		l.updateScrollBar()
		l.updateVisible()
	}
	h3 := arr.OnReplaced(syncAll)
	h4 := arr.OnMoved(func(_, _ int) { syncAll() })

	l.stopArray = func() {
		h1.Stop()
		h2.Stop()
		h3.Stop()
		h4.Stop()
	}
}

// SetRenderItem sets the factory function that creates a Component for a
// given item index and data value.
func (l *List) SetRenderItem(fn func(int, any) *Component) {
	l.renderItem = fn
	l.updateVisible()
}

// Selected returns the currently selected item index, or -1 if none.
func (l *List) Selected() int {
	return l.selected.Peek()
}

// SetSelected sets the selected item index.
func (l *List) SetSelected(idx int) {
	if idx < -1 || idx >= len(l.items) {
		return
	}
	old := l.selected.Peek()
	l.selected.Set(idx)
	DefaultScheduler.Flush()
	if idx != old && l.onChange != nil {
		l.onChange(idx)
	}
	l.updateHighlight()
	l.MarkDrawDirty()
	if idx >= 0 {
		l.ScrollToSelection()
	}
}

// BindSelected binds the selection to a reactive Ref[int].
func (l *List) BindSelected(ref *Ref[int]) {
	l.selected = ref
	bindRef(&l.selWatch, ref, l.SetSelected)
}

// SelectedRef returns the reactive Ref backing the selection index.
// Use this to subscribe to selection changes via WatchValue or bind to
// other reactive state.
func (l *List) SelectedRef() *Ref[int] {
	return l.selected
}

// SelectedItem returns the data of the currently selected item, or nil if
// no item is selected.
func (l *List) SelectedItem() any {
	idx := l.selected.Peek()
	if idx < 0 || idx >= len(l.items) {
		return nil
	}
	return l.items[idx].Data
}

// ClearSelection deselects the current item.
func (l *List) ClearSelection() {
	l.SetSelected(-1)
}

// SelectNext moves the selection to the next item. If nothing is selected,
// selects the first item. Scrolls to keep the selection visible.
func (l *List) SelectNext() {
	if len(l.items) == 0 {
		return
	}
	idx := l.selected.Peek() + 1
	if idx >= len(l.items) {
		idx = len(l.items) - 1
	}
	l.SetSelected(idx)
	l.ScrollToSelection()
}

// SelectPrevious moves the selection to the previous item. If nothing is
// selected, selects the last item. Scrolls to keep the selection visible.
func (l *List) SelectPrevious() {
	if len(l.items) == 0 {
		return
	}
	idx := l.selected.Peek()
	if idx < 0 {
		idx = len(l.items) - 1
	} else if idx > 0 {
		idx--
	}
	l.SetSelected(idx)
	l.ScrollToSelection()
}

// SelectFirst selects the first item and scrolls to it.
func (l *List) SelectFirst() {
	if len(l.items) == 0 {
		return
	}
	l.SetSelected(0)
	l.ScrollToSelection()
}

// SelectLast selects the last item and scrolls to it.
func (l *List) SelectLast() {
	if len(l.items) == 0 {
		return
	}
	l.SetSelected(len(l.items) - 1)
	l.ScrollToSelection()
}

// ScrollToSelection scrolls so that the currently selected item is visible.
func (l *List) ScrollToSelection() {
	l.ScrollToIndex(l.selected.Peek())
}

// Items returns the current list items.
func (l *List) Items() []ListItem {
	return l.items
}

// SetOnChange sets the callback for selection changes.
func (l *List) SetOnChange(fn func(int)) {
	l.onChange = fn
}

// SetSelectable enables or disables built-in selection highlighting.
// When enabled, the list renders a background highlight behind the selected item.
func (l *List) SetSelectable(enabled bool) {
	l.selectable = enabled
	if enabled && l.selHighlight == nil {
		l.selHighlight = sg.NewSprite(l.node.Name+"-sel-hl", sg.TextureRegion{})
		l.selHighlight.SetVisible(false)
		l.selHighlight.SetZIndex(999)
		l.content.AddChild(l.selHighlight)
	}
	l.updateHighlight()
}

// Selectable returns whether built-in selection highlighting is enabled.
func (l *List) Selectable() bool {
	return l.selectable
}

func (l *List) updateHighlight() {
	if l.selHighlight == nil {
		return
	}
	idx := l.selected.Peek()
	if !l.selectable || idx < 0 || idx >= len(l.items) {
		l.selHighlight.SetVisible(false)
		return
	}
	itemW := l.Width - float64(DefaultScrollBarWidth)
	y := float64(idx) * l.itemHeight
	l.selHighlight.SetPosition(0, y)
	l.selHighlight.SetScale(itemW, l.itemHeight)
	bg := l.EffectiveTheme().List.Group(l.Variant()).ItemBackground.Resolve(StateDefault)
	l.selHighlight.SetColor(bg.Color)
	l.selHighlight.SetVisible(true)
	l.MarkDrawDirty()
}

// ScrollToIndex scrolls so that the given item index is visible.
func (l *List) ScrollToIndex(idx int) {
	if idx < 0 || idx >= len(l.items) {
		return
	}
	itemTop := float64(idx) * l.itemHeight
	itemBottom := itemTop + l.itemHeight
	viewH := l.Height

	pos := l.scrollPos.Peek()
	if itemTop < pos {
		pos = itemTop
	} else if itemBottom > pos+viewH {
		pos = itemBottom - viewH
	}
	l.scrollBar.SetScrollPos(pos)
}

// Update processes mouse wheel input and keyboard navigation for scrolling.
// This is called automatically via the willow node's OnUpdate hook; no
// manual call needed.
func (l *List) Update() {
	if l.containsCursor() {
		_, wy := engine.Wheel()
		if wy != 0 {
			newPos := l.scrollPos.Peek() - wy*scrollWheelSpeedList
			l.scrollBar.SetScrollPos(newPos)
		}
	}

	// Keyboard navigation: Up/Down move selection when focused.
	if l.focused && l.enabled && l.selectable {
		im := DefaultInputManager
		sel := l.selected.Peek()
		n := len(l.items)
		if im.IsKeyJustAvailable(engine.KeyUp) && sel > 0 {
			l.SetSelected(sel - 1)
			l.ScrollToIndex(sel - 1)
			im.Consume(engine.KeyUp)
		} else if im.IsKeyJustAvailable(engine.KeyDown) && sel < n-1 {
			l.SetSelected(sel + 1)
			l.ScrollToIndex(sel + 1)
			im.Consume(engine.KeyDown)
		}
	}
}

// SetSize sets the list dimensions and updates internal layout.
func (l *List) SetSize(w, h float64) {
	l.Width = w
	l.Height = h

	sbWidth := float64(DefaultScrollBarWidth)

	l.resizeBackground(w, h)

	l.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}

	// ScrollBar.
	l.scrollBar.SetSize(sbWidth, h)
	l.scrollBar.SetPosition(w-sbWidth, 0)

	// Viewport.
	l.viewport.SetPosition(0, 0)

	// Clipping mask so items don't render outside bounds.
	// Reuse the existing mask sprite on subsequent calls to avoid accumulating
	// orphaned nodes every time the list is resized.
	if l.maskSprite == nil {
		maskRoot := sg.NewContainer(l.node.Name + "-mask")
		l.maskSprite = sg.NewSprite(l.node.Name+"-mask-rect", sg.TextureRegion{})
		l.maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
		maskRoot.AddChild(l.maskSprite)
		l.viewport.SetMask(maskRoot)
	}
	l.maskSprite.SetScale(w-sbWidth, h)

	l.updateScrollBar()
	l.updateVisible()
	l.MarkLayoutDirty()
}

// Dispose cleans up watches and child components.
func (l *List) Dispose() {
	if l.stopArray != nil {
		l.stopArray()
	}
	l.watch.Stop()
	l.selWatch.Stop()
	l.clearPool()
	l.scrollBar.Dispose()
	l.Component.Dispose()
}

// ItemCount returns the number of items in the list.
func (l *List) ItemCount() int {
	return len(l.items)
}

// ListScrollBar returns the internal scrollbar. Used for testing.
func (l *List) ListScrollBar() *ScrollBar { return l.scrollBar }

// ListScrollPos returns the reactive scroll position ref. Used for testing.
func (l *List) ListScrollPos() *Ref[float64] { return l.scrollPos }

// ListSelHighlight returns the selection highlight node, or nil if not created.
// Used for testing.
func (l *List) ListSelHighlight() *sg.Node { return l.selHighlight }

func (l *List) updateScrollBar() {
	totalH := float64(len(l.items)) * l.itemHeight
	l.scrollBar.SetContentSize(totalH, l.Height)

	// Hide scrollbar when content fits.
	l.scrollBar.SetVisible(totalH > l.Height)
}

func (l *List) updateVisible() {
	if l.renderItem == nil || len(l.items) == 0 {
		l.clearPool()
		return
	}

	pos := l.scrollPos.Peek()
	viewH := l.Height

	// Move the content container to reflect the scroll offset. Items are
	// placed at fixed absolute positions (index * itemHeight) so only
	// the container moves each frame — no per-item repositioning needed.
	l.content.SetPosition(0, -pos)

	// Two-zone virtualization: items are created when they enter the
	// render zone and only destroyed when they leave the wider keep zone.
	// This prevents flicker during fast scrolling — items that were just
	// visible persist in the tree long enough for newly created items
	// to be picked up by the renderer.
	const (
		renderBuffer = 2  // create items this far ahead of viewport
		keepBuffer   = 12 // keep items alive this far past viewport
	)

	renderStart := int(math.Floor(pos/l.itemHeight)) - renderBuffer
	if renderStart < 0 {
		renderStart = 0
	}
	renderEnd := renderStart + int(math.Ceil(viewH/l.itemHeight)) + 1 + 2*renderBuffer
	if renderEnd > len(l.items) {
		renderEnd = len(l.items)
	}

	keepStart := int(math.Floor(pos/l.itemHeight)) - keepBuffer
	if keepStart < 0 {
		keepStart = 0
	}
	keepEnd := keepStart + int(math.Ceil(viewH/l.itemHeight)) + 1 + 2*keepBuffer
	if keepEnd > len(l.items) {
		keepEnd = len(l.items)
	}

	// Phase 1: Add newly visible items FIRST (before any removals).
	for i := renderStart; i < renderEnd; i++ {
		if _, exists := l.poolIndex[i]; !exists {
			l.addPoolItem(i)
		}
	}

	// Phase 2: Remove items that are outside the wider keep zone.
	for idx := range l.poolIndex {
		if idx < keepStart || idx >= keepEnd {
			l.removePoolItem(idx)
		}
	}

	l.visibleStart = renderStart
	l.visibleEnd = renderEnd

	// Clip hit shapes to the visible portion of each item so that items
	// partially scrolled above/below the viewport don't intercept clicks
	// on elements outside the list (e.g. a tab bar above).
	itemW := l.Width - float64(DefaultScrollBarWidth)
	for idx, slot := range l.poolIndex {
		if slot < len(l.pool) && l.pool[slot] != nil {
			comp := l.pool[slot]
			cellY := float64(idx) * l.itemHeight

			// Compute the node's alignment offset within the cell.
			offX, offY := l.itemOffset(comp)

			topVisible := cellY+l.itemHeight > pos
			bottomVisible := cellY < pos+viewH
			if topVisible && bottomVisible {
				comp.node.Interactable = true
				// Hit rect relative to the node's position — cover the
				// full cell area, clamped to the visible viewport.
				visTop := math.Max(cellY, pos)
				visBot := math.Min(cellY+l.itemHeight, pos+viewH)
				comp.node.HitShape = sg.HitRect{
					X:      -offX,
					Y:      visTop - (cellY + offY),
					Width:  itemW,
					Height: visBot - visTop,
				}
			} else {
				comp.node.Interactable = false
			}
		}
	}

	l.updateHighlight()
}

// itemOffset returns the (x, y) offset of a rendered component within its
// cell based on the list's ItemAlign/ItemVAlign settings and theme ItemPadding.
func (l *List) itemOffset(comp *Component) (float64, float64) {
	pad := l.EffectiveTheme().List.Group(l.Variant()).ItemPadding
	cellW := l.Width - float64(DefaultScrollBarWidth)
	var offX, offY float64
	switch l.ItemAlign {
	case AlignStart:
		offX = pad.Left
	case AlignCenter:
		offX = pad.Left + (cellW-pad.Left-pad.Right-comp.Width)/2
	case AlignEnd:
		offX = cellW - pad.Right - comp.Width
	}
	switch l.ItemVAlign {
	case AlignCenter:
		offY = (l.itemHeight - comp.Height) / 2
	case AlignEnd:
		offY = l.itemHeight - comp.Height
	}
	return offX, offY
}

func (l *List) addPoolItem(index int) {
	comp := l.renderItem(index, l.items[index].Data)
	if comp == nil {
		return
	}

	// Absolute position within content — the content container's own
	// position handles the scroll offset.
	cellY := float64(index) * l.itemHeight
	offX, offY := l.itemOffset(comp)
	comp.SetPosition(offX, cellY+offY)
	comp.SetHitShape(sg.HitRect{X: 0, Y: 0, Width: comp.Width, Height: comp.Height})

	// Wire click for selection.
	idx := index
	comp.SetInteractable(true)
	comp.OnClick(func(ctx sg.ClickContext) {
		l.SetSelected(idx)
	})

	l.content.AddChild(comp.Node())

	// Find a free pool slot or append.
	slot := -1
	for i, p := range l.pool {
		if p == nil {
			slot = i
			break
		}
	}
	if slot >= 0 {
		l.pool[slot] = comp
	} else {
		slot = len(l.pool)
		l.pool = append(l.pool, comp)
	}
	l.poolIndex[index] = slot
}

func (l *List) removePoolItem(index int) {
	slot, ok := l.poolIndex[index]
	if !ok {
		return
	}
	if slot < len(l.pool) && l.pool[slot] != nil {
		l.content.RemoveChild(l.pool[slot].Node())
		l.pool[slot].Dispose()
		l.pool[slot] = nil
	}
	delete(l.poolIndex, index)
}

func (l *List) clearPool() {
	for idx := range l.poolIndex {
		slot := l.poolIndex[idx]
		if slot < len(l.pool) && l.pool[slot] != nil {
			l.content.RemoveChild(l.pool[slot].Node())
			l.pool[slot].Dispose()
			l.pool[slot] = nil
		}
	}
	l.poolIndex = make(map[int]int)
	l.pool = l.pool[:0]
}

// PoolSize returns the number of active (non-nil) components in the pool.
// This is useful for verifying virtualization in tests.
func (l *List) PoolSize() int {
	count := 0
	for _, p := range l.pool {
		if p != nil {
			count++
		}
	}
	return count
}
