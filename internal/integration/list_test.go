package integration

import (
	"fmt"
	"testing"

	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/widget"
)

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestNewListDefaults(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	if l.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if l.ListScrollBar() == nil {
		t.Fatal("scrollBar should not be nil")
	}
	if l.ItemCount() != 0 {
		t.Errorf("ItemCount() = %d, want 0", l.ItemCount())
	}
	if l.Selected() != -1 {
		t.Errorf("Selected() = %d, want -1", l.Selected())
	}
}

func TestListSetItems(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	items := make([]ui.ListItem, 100)
	for i := range items {
		items[i] = ui.ListItem{Data: fmt.Sprintf("item-%d", i)}
	}
	l.SetItems(items)

	if l.ItemCount() != 100 {
		t.Errorf("ItemCount() = %d, want 100", l.ItemCount())
	}
}

func TestListRenderVisibleItems(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	l.SetSize(200, 100) // viewport = 100px, itemHeight = 30 -> ~4 items visible

	renderCount := 0
	l.SetRenderItem(func(index int, data any) *ui.Component {
		renderCount++
		c := ui.NewComponent(fmt.Sprintf("item-%d", index))
		return c
	})

	items := make([]ui.ListItem, 100)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	// Should only render items visible in viewport, not all 100.
	if renderCount > 10 {
		t.Errorf("renderCount = %d, want <= 10 (virtualized)", renderCount)
	}
	if renderCount == 0 {
		t.Error("renderCount = 0, expected some items rendered")
	}
}

func TestListItemRecycling(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	l.SetSize(200, 90) // ~3 items visible + 1 buffer

	l.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := make([]ui.ListItem, 1000)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	poolSize := l.PoolSize()
	if poolSize > 10 {
		t.Errorf("PoolSize() = %d, want <= 10 for 1000 items in small viewport", poolSize)
	}
	if poolSize == 0 {
		t.Error("PoolSize() = 0, expected some active components")
	}
}

func TestListSetSelected(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	items := make([]ui.ListItem, 10)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	l.SetSelected(5)
	if l.Selected() != 5 {
		t.Errorf("Selected() = %d, want 5", l.Selected())
	}

	// Out of range should be ignored.
	l.SetSelected(100)
	if l.Selected() != 5 {
		t.Errorf("Selected() = %d, want 5 (out of range ignored)", l.Selected())
	}
}

func TestListBindSelected(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	items := make([]ui.ListItem, 10)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	ref := ui.NewRef(3)
	l.BindSelected(ref)

	if l.Selected() != 3 {
		t.Errorf("binding should sync initial value, got %d", l.Selected())
	}

	ref.Set(7)
	ui.DefaultScheduler.Flush()
	if l.Selected() != 7 {
		t.Errorf("reactive update should set selection, got %d", l.Selected())
	}
}

func TestListOnChange(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	items := make([]ui.ListItem, 10)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	var got int
	l.SetOnChange(func(idx int) { got = idx })

	l.SetSelected(4)
	if got != 4 {
		t.Errorf("onChange got %d, want 4", got)
	}
}

func TestListScrollToIndex(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	l.SetSize(200, 90) // 3 items visible

	l.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := make([]ui.ListItem, 100)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	l.ScrollToIndex(50)
	// After scrolling, scrollPos should have changed.
	pos := l.ListScrollPos().Peek()
	if pos <= 0 {
		t.Errorf("scrollPos = %f, want > 0 after ScrollToIndex(50)", pos)
	}
}

func TestListEmptyState(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	l.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent("item")
	})

	// Set empty items.
	l.SetItems([]ui.ListItem{})
	if l.ItemCount() != 0 {
		t.Errorf("ItemCount() = %d, want 0", l.ItemCount())
	}
	if l.PoolSize() != 0 {
		t.Errorf("PoolSize() = %d, want 0 for empty list", l.PoolSize())
	}
}

// ---------------------------------------------------------------------------
// List — helper methods
// ---------------------------------------------------------------------------

func TestListSelectedRef(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	items := make([]ui.ListItem, 5)
	for i := range items {
		items[i] = ui.ListItem{Data: fmt.Sprintf("item-%d", i)}
	}
	l.SetItems(items)

	ref := l.SelectedRef()
	if ref == nil {
		t.Fatal("SelectedRef() should not be nil")
	}

	// Watch for reactive changes.
	var watchedVal int
	w := ui.WatchValue(ref, func(_, newVal int) { watchedVal = newVal })
	defer w.Stop()

	l.SetSelected(3)
	ui.DefaultScheduler.Flush()
	if watchedVal != 3 {
		t.Errorf("reactive watch got %d, want 3", watchedVal)
	}
}

func TestListSelectedItem(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	items := []ui.ListItem{
		{Data: "apple"}, {Data: "banana"}, {Data: "cherry"},
	}
	l.SetItems(items)

	if l.SelectedItem() != nil {
		t.Error("SelectedItem() should be nil when nothing selected")
	}

	l.SetSelected(1)
	if l.SelectedItem() != "banana" {
		t.Errorf("SelectedItem() = %v, want banana", l.SelectedItem())
	}
}

func TestListClearSelection(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	items := make([]ui.ListItem, 5)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	l.SetSelected(2)
	l.ClearSelection()
	if l.Selected() != -1 {
		t.Errorf("Selected() = %d, want -1 after ClearSelection", l.Selected())
	}
}

func TestListSelectNext(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	l.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})
	items := make([]ui.ListItem, 5)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	// From no selection, selects first.
	l.SelectNext()
	if l.Selected() != 0 {
		t.Errorf("SelectNext from -1: got %d, want 0", l.Selected())
	}

	l.SelectNext()
	if l.Selected() != 1 {
		t.Errorf("SelectNext from 0: got %d, want 1", l.Selected())
	}

	// At end, stays at last.
	l.SetSelected(4)
	l.SelectNext()
	if l.Selected() != 4 {
		t.Errorf("SelectNext at end: got %d, want 4", l.Selected())
	}
}

func TestListSelectPrevious(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	l.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})
	items := make([]ui.ListItem, 5)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	// From no selection, selects last.
	l.SelectPrevious()
	if l.Selected() != 4 {
		t.Errorf("SelectPrevious from -1: got %d, want 4", l.Selected())
	}

	l.SetSelected(2)
	l.SelectPrevious()
	if l.Selected() != 1 {
		t.Errorf("SelectPrevious from 2: got %d, want 1", l.Selected())
	}

	// At start, stays at first.
	l.SetSelected(0)
	l.SelectPrevious()
	if l.Selected() != 0 {
		t.Errorf("SelectPrevious at start: got %d, want 0", l.Selected())
	}
}

func TestListSelectFirstLast(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	l.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})
	items := make([]ui.ListItem, 10)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	l.SelectLast()
	if l.Selected() != 9 {
		t.Errorf("SelectLast: got %d, want 9", l.Selected())
	}

	l.SelectFirst()
	if l.Selected() != 0 {
		t.Errorf("SelectFirst: got %d, want 0", l.Selected())
	}
}

func TestListScrollToSelection(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	l.SetSize(200, 90)
	l.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})
	items := make([]ui.ListItem, 100)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	l.SetSelected(50)
	l.ScrollToSelection()
	pos := l.ListScrollPos().Peek()
	if pos <= 0 {
		t.Errorf("scrollPos = %f, want > 0 after ScrollToSelection", pos)
	}
}

func TestListItems(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	items := []ui.ListItem{{Data: "a"}, {Data: "b"}}
	l.SetItems(items)

	got := l.Items()
	if len(got) != 2 {
		t.Fatalf("Items() len = %d, want 2", len(got))
	}
	if got[0].Data != "a" || got[1].Data != "b" {
		t.Errorf("Items() = %v, want [a, b]", got)
	}
}

func TestListSelectNextEmptyList(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	// Should not panic on empty list.
	l.SelectNext()
	l.SelectPrevious()
	l.SelectFirst()
	l.SelectLast()
	if l.Selected() != -1 {
		t.Errorf("Selected() = %d, want -1 on empty list", l.Selected())
	}
}

// ---------------------------------------------------------------------------
// List — selectable mode
// ---------------------------------------------------------------------------

func TestListSelectable(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	l.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})
	items := make([]ui.ListItem, 10)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	l.SetSelectable(true)
	if !l.Selectable() {
		t.Fatal("Selectable() should be true after SetSelectable(true)")
	}
	if l.ListSelHighlight() == nil {
		t.Fatal("selHighlight should be created when selectable is enabled")
	}

	l.SetSelected(3)
	if !l.ListSelHighlight().Visible() {
		t.Error("highlight should be visible when an item is selected")
	}
}

func TestListSelectableOff(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	items := make([]ui.ListItem, 5)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	// Not selectable by default.
	if l.Selectable() {
		t.Error("Selectable() should be false by default")
	}
	l.SetSelected(2)
	// selHighlight should be nil since SetSelectable was never called.
	if l.ListSelHighlight() != nil {
		t.Error("selHighlight should be nil when selectable is not enabled")
	}
}

func TestListSelectableThemeColor(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	l.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})
	items := make([]ui.ListItem, 5)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	l.SetSelectable(true)
	l.SetSelected(1)

	expected := l.EffectiveTheme().List.Group(l.Variant()).ItemBackground.Resolve(ui.StateDefault).Color
	if l.ListSelHighlight().Color() != expected {
		t.Errorf("highlight color = %v, want %v", l.ListSelHighlight().Color(), expected)
	}
}

func TestListSelectableHidesOnDeselect(t *testing.T) {
	resetScheduler()
	l := ui.NewList("list", 30)
	defer l.Dispose()

	l.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})
	items := make([]ui.ListItem, 5)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	l.SetItems(items)

	l.SetSelectable(true)
	l.SetSelected(2)
	if !l.ListSelHighlight().Visible() {
		t.Fatal("highlight should be visible")
	}

	l.SetSelected(-1)
	if l.ListSelHighlight().Visible() {
		t.Error("highlight should be hidden when selection is -1")
	}
}

// ---------------------------------------------------------------------------
// TileList
// ---------------------------------------------------------------------------

func TestNewTileListDefaults(t *testing.T) {
	resetScheduler()
	tl := ui.NewTileList("tiles", 64, 64)
	defer tl.Dispose()

	if tl.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if tl.TileScrollBar() == nil {
		t.Fatal("scrollBar should not be nil")
	}
}

func TestTileListAutoColumns(t *testing.T) {
	resetScheduler()
	tl := ui.NewTileList("tiles", 50, 50)
	defer tl.Dispose()

	// Width 300 - scrollbar 16 = 284, / 50 = 5 columns.
	tl.SetSize(300, 300)
	tl.SetColumns(0) // auto

	items := make([]ui.ListItem, 20)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	tl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("tile-%d", index))
	})
	tl.SetItems(items)

	cols := tl.EffectiveColumns()
	if cols != 5 {
		t.Errorf("EffectiveColumns() = %d, want 5", cols)
	}
}

func TestTileListSetColumns(t *testing.T) {
	resetScheduler()
	tl := ui.NewTileList("tiles", 50, 50)
	defer tl.Dispose()

	tl.SetSize(300, 300)
	tl.SetColumns(3)

	if tl.EffectiveColumns() != 3 {
		t.Errorf("EffectiveColumns() = %d, want 3", tl.EffectiveColumns())
	}
}

func TestTileListVirtualization(t *testing.T) {
	resetScheduler()
	tl := ui.NewTileList("tiles", 50, 50)
	defer tl.Dispose()

	tl.SetSize(200, 100) // small viewport
	tl.SetColumns(3)

	tl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("tile-%d", index))
	})

	items := make([]ui.ListItem, 100)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	tl.SetItems(items)

	poolSize := tl.PoolSize()
	if poolSize > 30 {
		t.Errorf("PoolSize() = %d, expected virtualized rendering", poolSize)
	}
	if poolSize == 0 {
		t.Error("PoolSize() = 0, expected some rendered tiles")
	}
}

func TestTileListEmptyState(t *testing.T) {
	resetScheduler()
	tl := ui.NewTileList("tiles", 50, 50)
	defer tl.Dispose()

	tl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent("tile")
	})
	tl.SetItems([]ui.ListItem{})

	if tl.PoolSize() != 0 {
		t.Errorf("PoolSize() = %d, want 0 for empty tile list", tl.PoolSize())
	}
}

// ---------------------------------------------------------------------------
// TileList — helper methods
// ---------------------------------------------------------------------------

func TestTileListSelectedRef(t *testing.T) {
	resetScheduler()
	tl := ui.NewTileList("tiles", 50, 50)
	defer tl.Dispose()

	items := make([]ui.ListItem, 5)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	tl.SetItems(items)

	ref := tl.SelectedRef()
	if ref == nil {
		t.Fatal("SelectedRef() should not be nil")
	}

	var watchedVal int
	w := ui.WatchValue(ref, func(_, newVal int) { watchedVal = newVal })
	defer w.Stop()

	tl.SetSelected(2)
	ui.DefaultScheduler.Flush()
	if watchedVal != 2 {
		t.Errorf("reactive watch got %d, want 2", watchedVal)
	}
}

func TestTileListSelectedItem(t *testing.T) {
	resetScheduler()
	tl := ui.NewTileList("tiles", 50, 50)
	defer tl.Dispose()

	items := []ui.ListItem{{Data: "x"}, {Data: "y"}, {Data: "z"}}
	tl.SetItems(items)

	if tl.SelectedItem() != nil {
		t.Error("SelectedItem() should be nil when nothing selected")
	}
	tl.SetSelected(2)
	if tl.SelectedItem() != "z" {
		t.Errorf("SelectedItem() = %v, want z", tl.SelectedItem())
	}
}

func TestTileListSelectNextPrevious(t *testing.T) {
	resetScheduler()
	tl := ui.NewTileList("tiles", 50, 50)
	defer tl.Dispose()

	tl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("tile-%d", index))
	})
	items := make([]ui.ListItem, 6)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	tl.SetItems(items)

	tl.SelectNext()
	if tl.Selected() != 0 {
		t.Errorf("SelectNext from -1: got %d, want 0", tl.Selected())
	}
	tl.SelectNext()
	if tl.Selected() != 1 {
		t.Errorf("SelectNext from 0: got %d, want 1", tl.Selected())
	}

	tl.SetSelected(5)
	tl.SelectNext()
	if tl.Selected() != 5 {
		t.Errorf("SelectNext at end: got %d, want 5", tl.Selected())
	}

	tl.SelectPrevious()
	if tl.Selected() != 4 {
		t.Errorf("SelectPrevious from 5: got %d, want 4", tl.Selected())
	}
}

func TestTileListScrollToSelection(t *testing.T) {
	resetScheduler()
	tl := ui.NewTileList("tiles", 50, 50)
	defer tl.Dispose()

	tl.SetSize(200, 100)
	tl.SetColumns(3)
	tl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("tile-%d", index))
	})
	items := make([]ui.ListItem, 100)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	tl.SetItems(items)

	tl.SetSelected(90)
	tl.ScrollToSelection()
	pos := tl.TileScrollPos().Peek()
	if pos <= 0 {
		t.Errorf("scrollPos = %f, want > 0 after ScrollToSelection", pos)
	}
}

func TestTileListClearSelection(t *testing.T) {
	resetScheduler()
	tl := ui.NewTileList("tiles", 50, 50)
	defer tl.Dispose()

	items := make([]ui.ListItem, 5)
	for i := range items {
		items[i] = ui.ListItem{Data: i}
	}
	tl.SetItems(items)

	tl.SetSelected(3)
	tl.ClearSelection()
	if tl.Selected() != -1 {
		t.Errorf("Selected() = %d, want -1 after ClearSelection", tl.Selected())
	}
}

// ---------------------------------------------------------------------------
// TreeList
// ---------------------------------------------------------------------------

func TestNewTreeListDefaults(t *testing.T) {
	resetScheduler()
	tl := ui.NewTreeList("tree", 24)
	defer tl.Dispose()

	if tl.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if tl.FlatCount() != 0 {
		t.Errorf("FlatCount() = %d, want 0", tl.FlatCount())
	}
}

func TestTreeListSetRoots(t *testing.T) {
	resetScheduler()
	tl := ui.NewTreeList("tree", 24)
	defer tl.Dispose()

	tl.SetRenderItem(func(node *ui.TreeNode, depth int) *ui.Component {
		return ui.NewComponent("tree-item")
	})

	roots := []*ui.TreeNode{
		{Data: "A", Children: []*ui.TreeNode{
			{Data: "A1"},
			{Data: "A2"},
		}},
		{Data: "B"},
	}
	tl.SetRoots(roots)

	// Without expansion, only root nodes are visible.
	if tl.FlatCount() != 2 {
		t.Errorf("FlatCount() = %d, want 2 (roots only)", tl.FlatCount())
	}
}

func TestTreeListExpandCollapse(t *testing.T) {
	resetScheduler()
	tl := ui.NewTreeList("tree", 24)
	defer tl.Dispose()

	tl.SetRenderItem(func(node *ui.TreeNode, depth int) *ui.Component {
		return ui.NewComponent("tree-item")
	})

	root := &ui.TreeNode{Data: "A", Children: []*ui.TreeNode{
		{Data: "A1"},
		{Data: "A2"},
		{Data: "A3"},
	}}
	tl.SetRoots([]*ui.TreeNode{root})

	// Expand root.
	tl.Expand(root)
	if tl.FlatCount() != 4 {
		t.Errorf("FlatCount() = %d, want 4 after expand", tl.FlatCount())
	}
	if !tl.IsExpanded(root) {
		t.Error("root should be expanded")
	}

	// Collapse root.
	tl.Collapse(root)
	if tl.FlatCount() != 1 {
		t.Errorf("FlatCount() = %d, want 1 after collapse", tl.FlatCount())
	}
}

func TestTreeListToggle(t *testing.T) {
	resetScheduler()
	tl := ui.NewTreeList("tree", 24)
	defer tl.Dispose()

	tl.SetRenderItem(func(node *ui.TreeNode, depth int) *ui.Component {
		return ui.NewComponent("tree-item")
	})

	root := &ui.TreeNode{Data: "A", Children: []*ui.TreeNode{
		{Data: "A1"},
	}}
	tl.SetRoots([]*ui.TreeNode{root})

	tl.Toggle(root)
	if tl.FlatCount() != 2 {
		t.Errorf("FlatCount() = %d, want 2 after first toggle", tl.FlatCount())
	}

	tl.Toggle(root)
	if tl.FlatCount() != 1 {
		t.Errorf("FlatCount() = %d, want 1 after second toggle", tl.FlatCount())
	}
}

func TestTreeListExpandAll(t *testing.T) {
	resetScheduler()
	tl := ui.NewTreeList("tree", 24)
	defer tl.Dispose()

	tl.SetRenderItem(func(node *ui.TreeNode, depth int) *ui.Component {
		return ui.NewComponent("tree-item")
	})

	roots := []*ui.TreeNode{
		{Data: "A", Children: []*ui.TreeNode{
			{Data: "A1", Children: []*ui.TreeNode{
				{Data: "A1a"},
			}},
			{Data: "A2"},
		}},
		{Data: "B"},
	}
	tl.SetRoots(roots)

	tl.ExpandAll()
	// A, A1, A1a, A2, B = 5
	if tl.FlatCount() != 5 {
		t.Errorf("FlatCount() = %d, want 5 after ExpandAll", tl.FlatCount())
	}
}

func TestTreeListCollapseAll(t *testing.T) {
	resetScheduler()
	tl := ui.NewTreeList("tree", 24)
	defer tl.Dispose()

	tl.SetRenderItem(func(node *ui.TreeNode, depth int) *ui.Component {
		return ui.NewComponent("tree-item")
	})

	roots := []*ui.TreeNode{
		{Data: "A", Children: []*ui.TreeNode{
			{Data: "A1"},
		}},
		{Data: "B"},
	}
	tl.SetRoots(roots)

	tl.ExpandAll()
	tl.CollapseAll()
	if tl.FlatCount() != 2 {
		t.Errorf("FlatCount() = %d, want 2 after CollapseAll", tl.FlatCount())
	}
}

func TestTreeListDepthInRenderItem(t *testing.T) {
	resetScheduler()
	tl := ui.NewTreeList("tree", 24)
	defer tl.Dispose()

	depths := make(map[string]int)
	tl.SetRenderItem(func(node *ui.TreeNode, depth int) *ui.Component {
		depths[node.Data.(string)] = depth
		return ui.NewComponent("tree-item")
	})

	roots := []*ui.TreeNode{
		{Data: "A", Children: []*ui.TreeNode{
			{Data: "A1", Children: []*ui.TreeNode{
				{Data: "A1a"},
			}},
		}},
	}
	tl.SetRoots(roots)
	tl.ExpandAll()

	if depths["A"] != 0 {
		t.Errorf("depth of A = %d, want 0", depths["A"])
	}
	if depths["A1"] != 1 {
		t.Errorf("depth of A1 = %d, want 1", depths["A1"])
	}
	if depths["A1a"] != 2 {
		t.Errorf("depth of A1a = %d, want 2", depths["A1a"])
	}
}

func TestTreeListEmptyState(t *testing.T) {
	resetScheduler()
	tl := ui.NewTreeList("tree", 24)
	defer tl.Dispose()

	tl.SetRenderItem(func(node *ui.TreeNode, depth int) *ui.Component {
		return ui.NewComponent("tree-item")
	})
	tl.SetRoots([]*ui.TreeNode{})

	if tl.FlatCount() != 0 {
		t.Errorf("FlatCount() = %d, want 0", tl.FlatCount())
	}
	if tl.PoolSize() != 0 {
		t.Errorf("PoolSize() = %d, want 0", tl.PoolSize())
	}
}

func TestTreeListExpandLeafIsNoop(t *testing.T) {
	resetScheduler()
	tl := ui.NewTreeList("tree", 24)
	defer tl.Dispose()

	tl.SetRenderItem(func(node *ui.TreeNode, depth int) *ui.Component {
		return ui.NewComponent("tree-item")
	})

	leaf := &ui.TreeNode{Data: "leaf"}
	tl.SetRoots([]*ui.TreeNode{leaf})

	tl.Expand(leaf) // no children, should be noop
	if tl.FlatCount() != 1 {
		t.Errorf("FlatCount() = %d, want 1 after expanding leaf", tl.FlatCount())
	}
}

// ---------------------------------------------------------------------------
// Tree toggle glyphs & helper
// ---------------------------------------------------------------------------

func TestTreeExpandGlyph(t *testing.T) {
	img := widget.TreeExpandGlyph()
	if img == nil {
		t.Fatal("TreeExpandGlyph() returned nil")
	}
	b := img.Bounds()
	if b.Dx() != widget.GlyphSize || b.Dy() != widget.GlyphSize {
		t.Errorf("expand glyph size = %dx%d, want %dx%d", b.Dx(), b.Dy(), widget.GlyphSize, widget.GlyphSize)
	}
}

func TestTreeCollapseGlyph(t *testing.T) {
	img := widget.TreeCollapseGlyph()
	if img == nil {
		t.Fatal("TreeCollapseGlyph() returned nil")
	}
	b := img.Bounds()
	if b.Dx() != widget.GlyphSize || b.Dy() != widget.GlyphSize {
		t.Errorf("collapse glyph size = %dx%d, want %dx%d", b.Dx(), b.Dy(), widget.GlyphSize, widget.GlyphSize)
	}
}

func TestNewTreeToggleNilForLeaf(t *testing.T) {
	resetScheduler()
	tl := ui.NewTreeList("tree", 24)
	defer tl.Dispose()

	leaf := &ui.TreeNode{Data: "leaf"}
	btn := ui.NewTreeToggle("toggle", tl, leaf)
	if btn != nil {
		t.Error("NewTreeToggle should return nil for leaf node")
	}
}

func TestNewTreeToggleButtonForParent(t *testing.T) {
	resetScheduler()
	tl := ui.NewTreeList("tree", 24)
	defer tl.Dispose()

	parent := &ui.TreeNode{Data: "parent", Children: []*ui.TreeNode{
		{Data: "child"},
	}}
	btn := ui.NewTreeToggle("toggle", tl, parent)
	if btn == nil {
		t.Fatal("NewTreeToggle should return non-nil for parent node")
	}
	if btn.Width != ui.TreeToggleSize || btn.Height != ui.TreeToggleSize {
		t.Errorf("toggle size = %.0fx%.0f, want %dx%d", btn.Width, btn.Height, ui.TreeToggleSize, ui.TreeToggleSize)
	}
	btn.Dispose()
}
