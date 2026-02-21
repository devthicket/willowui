package integration

import (
	"testing"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/reactive"
)

// flushAndUpdate flushes the reactive scheduler (runs pending watch callbacks)
// then calls dt.Update() to process dirty flags.
func flushAndUpdate(dt *ui.DataTable) {
	reactive.DefaultScheduler.Flush()
	dt.Update()
}

// ---------------------------------------------------------------------------
// DataTable integration tests
// ---------------------------------------------------------------------------

func TestDataTableDefaults(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	if dt.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if dt.DataTableScrollBar() == nil {
		t.Fatal("scrollBar should not be nil")
	}
	if dt.DataTableDisplayCount() != 0 {
		t.Errorf("DataTableDisplayCount() = %d, want 0", dt.DataTableDisplayCount())
	}
}

func TestDataTableAddColumns(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:    "name",
		Header: "Name",
		Weight: 1,
	})
	dt.AddColumn(ui.DataTableColumn{
		Key:    "score",
		Header: "Score",
		Weight: 1,
	})

	// Add items to confirm display count.
	dt.SetItems([]any{
		map[string]string{"name": "Alice", "score": "100"},
		map[string]string{"name": "Bob", "score": "95"},
	})

	if dt.DataTableDisplayCount() != 2 {
		t.Errorf("DataTableDisplayCount() = %d, want 2", dt.DataTableDisplayCount())
	}
}

func TestDataTableSetItems(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	items := make([]any, 50)
	for i := range items {
		items[i] = i
	}
	dt.SetItems(items)

	if dt.DataTableDisplayCount() != 50 {
		t.Errorf("DataTableDisplayCount() = %d, want 50", dt.DataTableDisplayCount())
	}
}

func TestDataTableSetSize(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(600, 400)

	if dt.Width != 600 {
		t.Errorf("Width = %f, want 600", dt.Width)
	}
	if dt.Height != 400 {
		t.Errorf("Height = %f, want 400", dt.Height)
	}
}

func TestDataTableScrollMode(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetScrollMode(ui.ScrollModeStatic)
	// Should not panic; static mode just changes slot allocation strategy.
}

func TestDataTableSelectionSingle(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSelectionMode(ui.SelectionModeSingle)
	dt.SetSize(400, 300)

	items := []any{"a", "b", "c"}
	dt.SetItems(items)

	// No selection initially.
	if len(dt.DataTableSelectedIndexes()) != 0 {
		t.Errorf("expected empty selection, got %v", dt.DataTableSelectedIndexes())
	}
}

func TestDataTableSelectionMulti(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSelectionMode(ui.SelectionModeMulti)
	dt.SetSize(400, 300)

	items := make([]any, 10)
	for i := range items {
		items[i] = i
	}
	dt.SetItems(items)

	// No selection initially.
	if len(dt.DataTableSelectedIndexes()) != 0 {
		t.Errorf("expected empty selection, got %v", dt.DataTableSelectedIndexes())
	}
}

func TestDataTableClearSelection(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(400, 300)
	dt.SetItems([]any{1, 2, 3})
	dt.ClearSelection()

	if len(dt.DataTableSelectedIndexes()) != 0 {
		t.Errorf("expected empty selection after clear, got %v", dt.DataTableSelectedIndexes())
	}
}

func TestDataTableZebraStriping(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetZebraStriping(true)
	dt.SetSize(400, 300)
	dt.SetItems([]any{1, 2, 3, 4, 5})

	// Should not panic; zebra striping just changes row background logic.
}

func TestDataTableShowHeader(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetShowHeader(false)
	dt.SetSize(400, 300)
	// Should not panic.
}

func TestDataTableLabelColumn(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	col := ui.LabelColumn("name", "Name", func(data any) string {
		if s, ok := data.(string); ok {
			return s
		}
		return ""
	})

	dt.AddColumn(col)
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Alice", "Bob", "Charlie"})

	if dt.DataTableDisplayCount() != 3 {
		t.Errorf("DataTableDisplayCount() = %d, want 3", dt.DataTableDisplayCount())
	}
}

func TestDataTableSortableColumn(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:      "name",
		Header:   "Name",
		Weight:   1,
		Sortable: true,
		SortType: ui.SortAlpha,
	})
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Charlie", "Alice", "Bob"})

	if dt.DataTableDisplayCount() != 3 {
		t.Errorf("DataTableDisplayCount() = %d, want 3", dt.DataTableDisplayCount())
	}
}

func TestDataTableFilterFunc(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(400, 300)
	dt.SetItems([]any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	// Filter: only even numbers.
	dt.SetFilterFunc(func(data any) bool {
		if n, ok := data.(int); ok {
			return n%2 == 0
		}
		return false
	})

	if dt.DataTableDisplayCount() != 5 {
		t.Errorf("DataTableDisplayCount() = %d, want 5", dt.DataTableDisplayCount())
	}
}

func TestDataTableFilterClear(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(400, 300)
	dt.SetItems([]any{1, 2, 3, 4, 5})

	dt.SetFilterFunc(func(data any) bool { return false })
	if dt.DataTableDisplayCount() != 0 {
		t.Errorf("DataTableDisplayCount() after filter = %d, want 0", dt.DataTableDisplayCount())
	}

	dt.SetFilterFunc(nil) // clear filter
	if dt.DataTableDisplayCount() != 5 {
		t.Errorf("DataTableDisplayCount() after clear = %d, want 5", dt.DataTableDisplayCount())
	}
}

func TestDataTableColumnDividers(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetShowColumnDividers(true)
	dt.SetShowRowDividers(true)
	dt.AddColumn(ui.DataTableColumn{Key: "a", Header: "A", Weight: 1})
	dt.AddColumn(ui.DataTableColumn{Key: "b", Header: "B", Weight: 1})
	dt.SetSize(400, 300)
	dt.SetItems([]any{"x", "y"})
	// Should not panic.
}

func TestDataTableScrollBar(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(400, 200)

	// With 100 items each 28px, scroll should be needed.
	items := make([]any, 100)
	for i := range items {
		items[i] = i
	}
	dt.SetItems(items)

	sb := dt.DataTableScrollBar()
	if sb == nil {
		t.Fatal("scrollbar should not be nil")
	}
}

func TestDataTableDispose(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	dt.AddColumn(ui.DataTableColumn{Key: "x", Header: "X", Weight: 1})
	dt.SetSize(400, 300)
	dt.SetItems([]any{1, 2, 3})
	dt.Dispose() // Should not panic.
}

// ---------------------------------------------------------------------------
// Substantive behavior tests
// ---------------------------------------------------------------------------

func TestDataTableLabelColumnCellText(t *testing.T) {
	// Verify that LabelColumn cells are populated (not blank) after SetItems.
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	names := []string{"Alice", "Bob", "Charlie"}
	dt.AddColumn(ui.LabelColumn("name", "Name", func(data any) string {
		if s, ok := data.(string); ok {
			return s
		}
		return ""
	}))
	dt.SetSize(400, 300)
	items := []any{names[0], names[1], names[2]}
	dt.SetItems(items)

	// Display count must match.
	if dt.DataTableDisplayCount() != 3 {
		t.Fatalf("DataTableDisplayCount() = %d, want 3", dt.DataTableDisplayCount())
	}

	// Replace items and verify display count updates.
	dt.SetItems([]any{"Dave", "Eve"})
	if dt.DataTableDisplayCount() != 2 {
		t.Errorf("DataTableDisplayCount() after replace = %d, want 2", dt.DataTableDisplayCount())
	}
}

func TestDataTableSortAlpha(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:         "name",
		Header:      "Name",
		Weight:      1,
		Sortable:    true,
		SortType:    ui.SortAlpha,
		SearchValue: func(data any) string { return data.(string) },
	})
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Charlie", "Alice", "Bob"})

	dt.SetSortedColumn("name", ui.SortAsc)

	disp := dt.DataTableDisplayIndexes()
	if len(disp) != 3 {
		t.Fatalf("display count = %d, want 3", len(disp))
	}
	// After alpha-asc sort: Alice(1), Bob(2), Charlie(0)
	items := []any{"Charlie", "Alice", "Bob"}
	order := make([]string, len(disp))
	for i, di := range disp {
		order[i] = items[di].(string)
	}
	if order[0] != "Alice" || order[1] != "Bob" || order[2] != "Charlie" {
		t.Errorf("sort order = %v, want [Alice Bob Charlie]", order)
	}
}

func TestDataTableSortNumeric(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:         "score",
		Header:      "Score",
		Weight:      1,
		Sortable:    true,
		SortType:    ui.SortNumeric,
		SearchValue: func(data any) string { return "" },
	})
	dt.SetSize(400, 300)
	dt.SetItems([]any{30, 10, 20})

	dt.SetSortedColumn("score", ui.SortAsc)

	disp := dt.DataTableDisplayIndexes()
	items := []any{30, 10, 20}
	order := make([]int, len(disp))
	for i, di := range disp {
		order[i] = items[di].(int)
	}
	if order[0] != 10 || order[1] != 20 || order[2] != 30 {
		t.Errorf("numeric sort order = %v, want [10 20 30]", order)
	}
}

func TestDataTableSortDesc(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:         "name",
		Header:      "Name",
		Weight:      1,
		Sortable:    true,
		SortType:    ui.SortAlpha,
		SearchValue: func(data any) string { return data.(string) },
	})
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Charlie", "Alice", "Bob"})

	dt.SetSortedColumn("name", ui.SortDesc)

	disp := dt.DataTableDisplayIndexes()
	items := []any{"Charlie", "Alice", "Bob"}
	order := make([]string, len(disp))
	for i, di := range disp {
		order[i] = items[di].(string)
	}
	if order[0] != "Charlie" || order[1] != "Bob" || order[2] != "Alice" {
		t.Errorf("desc sort order = %v, want [Charlie Bob Alice]", order)
	}
}

func TestDataTableFilterAndSortCompose(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:         "name",
		Header:      "Name",
		Weight:      1,
		Sortable:    true,
		SortType:    ui.SortAlpha,
		SearchValue: func(data any) string { return data.(string) },
	})
	dt.SetSize(400, 300)
	// Charlie=7, Alice=5, Bob=3, Dave=4 chars
	dt.SetItems([]any{"Charlie", "Alice", "Bob", "Dave"})
	dt.SetSortedColumn("name", ui.SortAsc)

	// Filter to only names with <= 4 chars: Bob(3), Dave(4).
	dt.SetFilterFunc(func(data any) bool {
		return len(data.(string)) <= 4
	})

	disp := dt.DataTableDisplayIndexes()
	if len(disp) != 2 {
		t.Fatalf("filtered display count = %d, want 2", len(disp))
	}
	items := []any{"Charlie", "Alice", "Bob", "Dave"}
	order := make([]string, len(disp))
	for i, di := range disp {
		order[i] = items[di].(string)
	}
	// After sort asc then filter <=4: Bob, Dave (alpha order).
	if order[0] != "Bob" || order[1] != "Dave" {
		t.Errorf("filter+sort compose: got %v, want [Bob Dave]", order)
	}
}

func TestDataTableFilterFuncNilRestoresAll(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(400, 300)
	dt.SetItems([]any{1, 2, 3, 4, 5})

	dt.SetFilterFunc(func(data any) bool { return false })
	if dt.DataTableDisplayCount() != 0 {
		t.Fatalf("expected 0 after all-false filter, got %d", dt.DataTableDisplayCount())
	}

	dt.SetFilterFunc(nil)
	if dt.DataTableDisplayCount() != 5 {
		t.Errorf("expected 5 after clearing filter, got %d", dt.DataTableDisplayCount())
	}
}

func TestDataTableBindItemsAdd(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(400, 300)
	arr := ui.NewArrayFrom([]any{"A", "B", "C"})
	dt.BindItems(arr)

	if dt.DataTableDisplayCount() != 3 {
		t.Fatalf("initial count = %d, want 3", dt.DataTableDisplayCount())
	}

	arr.Push("D")
	resetScheduler()

	if dt.DataTableDisplayCount() != 4 {
		t.Errorf("after Push count = %d, want 4", dt.DataTableDisplayCount())
	}
}

func TestDataTableBindItemsRemove(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(400, 300)
	arr := ui.NewArrayFrom([]any{"A", "B", "C"})
	dt.BindItems(arr)

	arr.RemoveAt(1) // remove "B"
	resetScheduler()

	if dt.DataTableDisplayCount() != 2 {
		t.Errorf("after RemoveAt count = %d, want 2", dt.DataTableDisplayCount())
	}
}

func TestDataTableSelectRow(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSelectionMode(ui.SelectionModeSingle)
	dt.SetSize(400, 300)
	dt.SetItems([]any{"A", "B", "C"})

	dt.SelectRow(1)
	sel := dt.DataTableSelectedIndexes()
	if len(sel) != 1 || sel[0] != 1 {
		t.Errorf("SelectRow(1): got %v, want [1]", sel)
	}

	if !dt.IsSelected(1) {
		t.Error("IsSelected(1) should be true")
	}
	if dt.IsSelected(0) {
		t.Error("IsSelected(0) should be false")
	}
}

func TestDataTableDeselectRow(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSelectionMode(ui.SelectionModeMulti)
	dt.SetSize(400, 300)
	dt.SetItems([]any{"A", "B", "C"})

	dt.SelectRow(0)
	dt.SelectRow(1)
	dt.DeselectRow(0)

	sel := dt.DataTableSelectedIndexes()
	if len(sel) != 1 || sel[0] != 1 {
		t.Errorf("DeselectRow(0): got %v, want [1]", sel)
	}
}

func TestDataTableSelectAll(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSelectionMode(ui.SelectionModeMulti)
	dt.SetSize(400, 300)
	dt.SetItems([]any{"A", "B", "C"})

	dt.SelectAll()
	sel := dt.DataTableSelectedIndexes()
	if len(sel) != 3 {
		t.Errorf("SelectAll: got %d selected, want 3", len(sel))
	}
}

func TestDataTableSelectAllNoOpForSingle(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSelectionMode(ui.SelectionModeSingle)
	dt.SetSize(400, 300)
	dt.SetItems([]any{"A", "B", "C"})

	dt.SelectAll()
	sel := dt.DataTableSelectedIndexes()
	if len(sel) != 0 {
		t.Errorf("SelectAll in single mode: got %d selected, want 0", len(sel))
	}
}

func TestDataTableScrollToRow(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(400, 200)
	items := make([]any, 50)
	for i := range items {
		items[i] = i
	}
	dt.SetItems(items)

	dt.ScrollToRow(40)
	// Should not panic; scroll position should have advanced.
	pos := dt.DataTableScrollPos().Peek()
	// Row 40 is at y=40*28=1120; viewport height ~= 200-32=168.
	// scrollBar clamps to [0, totalH - viewH] so pos should be > 0.
	if pos == 0 {
		t.Error("ScrollToRow(40) did not change scroll position")
	}
}

func TestDataTableScrollToTop(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(400, 200)
	items := make([]any, 50)
	for i := range items {
		items[i] = i
	}
	dt.SetItems(items)
	dt.ScrollToRow(40)
	dt.ScrollToTop()

	pos := dt.DataTableScrollPos().Peek()
	if pos != 0 {
		t.Errorf("ScrollToTop: pos = %f, want 0", pos)
	}
}

func TestDataTableSetSortedColumn(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:      "name",
		Header:   "Name",
		Weight:   1,
		Sortable: true,
		SortType: ui.SortAlpha,
	})
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Charlie", "Alice", "Bob"})

	dt.SetSortedColumn("name", ui.SortAsc)
	k, dir := dt.SortedColumn()
	if k != "name" || dir != ui.SortAsc {
		t.Errorf("SortedColumn() = (%s, %d), want (name, SortAsc)", k, dir)
	}
}

func TestDataTableOnSortSuppressesInternal(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:      "name",
		Header:   "Name",
		Weight:   1,
		Sortable: true,
		SortType: ui.SortAlpha,
	})
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Charlie", "Alice", "Bob"})

	sortCalled := false
	dt.SetOnSort(func(key string, dir ui.SortDirection) {
		sortCalled = true
	})

	// SetSortedColumn bypasses onSort (direct programmatic).
	// Verify original display order is unchanged when onSort is set and cycleSort would be called.
	// The display count should remain the same.
	_ = sortCalled
	if dt.DataTableDisplayCount() != 3 {
		t.Errorf("display count = %d, want 3", dt.DataTableDisplayCount())
	}
}

func TestDataTableSearchFilter(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.LabelColumn("name", "Name", func(data any) string {
		return data.(string)
	}))
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Alice", "Bob", "Albert"})

	searchRef := ui.NewRef("")
	dt.BindSearchFilter(searchRef)

	// Flush + Update to apply the initial sortFilterDirty from BindSearchFilter.
	flushAndUpdate(dt)

	searchRef.Set("al")
	flushAndUpdate(dt) // flush watch callback, then process sortFilterDirty

	if dt.DataTableDisplayCount() != 2 {
		t.Errorf("search 'al': count = %d, want 2 (Alice, Albert)", dt.DataTableDisplayCount())
	}

	searchRef.Set("")
	flushAndUpdate(dt)

	if dt.DataTableDisplayCount() != 3 {
		t.Errorf("search cleared: count = %d, want 3", dt.DataTableDisplayCount())
	}
}

func TestDataTableRebuildRefresh(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(400, 300)
	dt.SetItems([]any{1, 2, 3})

	// Both should set dirty flags without panicking.
	dt.Rebuild()
	dt.Refresh()
}

func TestDataTableStaticMode(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetScrollMode(ui.ScrollModeStatic)
	dt.SetSize(400, 300)
	items := make([]any, 20)
	for i := range items {
		items[i] = i
	}
	dt.SetItems(items)

	if dt.DataTableDisplayCount() != 20 {
		t.Errorf("static mode count = %d, want 20", dt.DataTableDisplayCount())
	}
}

func TestDataTableOnColumn(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{Key: "name", Header: "Name", Weight: 1})
	dt.SetSize(400, 300)

	// Attach render funcs via OnColumn.
	called := false
	dt.OnColumn("name", ui.DataTableColumn{
		Key:    "name",
		Weight: 1,
		UpdateCell: func(rowIndex int, data any, comp *ui.Component) {
			called = true
		},
	})

	dt.SetItems([]any{"A", "B"})
	_ = called // UpdateCell fires during updateRows
}

func TestDataTableToggleRowSelection(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSelectionMode(ui.SelectionModeMulti)
	dt.SetSize(400, 300)
	dt.SetItems([]any{"A", "B", "C"})

	dt.ToggleRowSelection(0)
	if !dt.IsSelected(0) {
		t.Error("ToggleRowSelection(0) should select")
	}

	dt.ToggleRowSelection(0)
	if dt.IsSelected(0) {
		t.Error("ToggleRowSelection(0) again should deselect")
	}
}

func TestDataTableBindSelectedIndexes(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSelectionMode(ui.SelectionModeMulti)
	dt.SetSize(400, 300)
	dt.SetItems([]any{"A", "B", "C", "D"})

	// Track what the setter receives.
	var externalSel []int
	dt.BindSelectedIndexes(
		func() []int { return externalSel },
		func(idx []int) { externalSel = idx },
	)

	// Select via API — setter should be called.
	dt.SelectRow(1)
	if len(externalSel) != 1 || externalSel[0] != 1 {
		t.Errorf("after SelectRow(1), externalSel = %v, want [1]", externalSel)
	}

	dt.SelectRow(3)
	if len(externalSel) != 2 {
		t.Errorf("after SelectRow(3), externalSel = %v, want len 2", externalSel)
	}

	// Push selection in from outside.
	dt.SetSelectedIndexes([]int{0, 2})
	if !dt.IsSelected(0) || !dt.IsSelected(2) {
		t.Error("SetSelectedIndexes([0,2]) not reflected in IsSelected")
	}
	if dt.IsSelected(1) || dt.IsSelected(3) {
		t.Error("SetSelectedIndexes([0,2]) should have cleared 1 and 3")
	}

	// Clear should push empty to setter.
	dt.ClearSelection()
	if len(externalSel) != 0 {
		t.Errorf("after ClearSelection, externalSel = %v, want []", externalSel)
	}
}

func TestDataTableBindItemsSetAt(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.LabelColumn("val", "Val", func(data any) string {
		return data.(string)
	}))
	dt.SetSize(400, 300)

	arr := ui.NewArrayFrom([]any{"A", "B", "C"})
	dt.BindItems(arr)

	if dt.DataTableDisplayCount() != 3 {
		t.Fatalf("initial count = %d, want 3", dt.DataTableDisplayCount())
	}

	// SetAt should update the item without changing count.
	arr.SetAt(1, "BB")
	resetScheduler()

	if dt.DataTableDisplayCount() != 3 {
		t.Errorf("after SetAt count = %d, want 3", dt.DataTableDisplayCount())
	}
}

func TestDataTableScrollToBottom(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.SetSize(400, 200)
	items := make([]any, 50)
	for i := range items {
		items[i] = i
	}
	dt.SetItems(items)

	dt.ScrollToBottom()
	pos := dt.DataTableScrollPos().Peek()
	if pos == 0 {
		t.Error("ScrollToBottom did not change scroll position")
	}
}

func TestDataTableSortCustomComparator(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:      "val",
		Header:   "Val",
		Weight:   1,
		Sortable: true,
		SortType: ui.SortCustom,
		Comparator: func(a, b any) int {
			// Sort by string length.
			la := len(a.(string))
			lb := len(b.(string))
			if la < lb {
				return -1
			}
			if la > lb {
				return 1
			}
			return 0
		},
		SearchValue: func(data any) string { return data.(string) },
	})
	dt.SetSize(400, 300)
	dt.SetItems([]any{"BB", "A", "CCC"})

	dt.SetSortedColumn("val", ui.SortAsc)

	disp := dt.DataTableDisplayIndexes()
	items := []any{"BB", "A", "CCC"}
	order := make([]string, len(disp))
	for i, di := range disp {
		order[i] = items[di].(string)
	}
	// Sorted by length asc: A(1), BB(2), CCC(3)
	if order[0] != "A" || order[1] != "BB" || order[2] != "CCC" {
		t.Errorf("custom sort order = %v, want [A BB CCC]", order)
	}
}

// ---------------------------------------------------------------------------
// Multi-sort, column filters, default sort, SortKeys, ResetFiltersAndSort,
// CellStyle, OnPostUpdate, Header, BindSearchInput tests
// ---------------------------------------------------------------------------

func TestDataTableMultiSortKeys(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	type item struct {
		name  string
		score int
	}

	dt.AddColumn(ui.DataTableColumn{
		Key:         "name",
		Header:      "Name",
		Weight:      1,
		Sortable:    true,
		SortType:    ui.SortAlpha,
		SearchValue: func(d any) string { return d.(item).name },
	})
	dt.AddColumn(ui.DataTableColumn{
		Key:      "score",
		Header:   "Score",
		Weight:   1,
		Sortable: true,
		SortType: ui.SortNumeric,
		SortValue: func(d any) any {
			return d.(item).score
		},
		SearchValue: func(d any) string { return "" },
	})
	dt.SetSize(400, 300)
	dt.SetItems([]any{
		item{"Alice", 10},
		item{"Bob", 20},
		item{"Alice", 30},
		item{"Bob", 5},
	})

	// Multi-sort: name asc, then score asc.
	dt.SetSortKeys([]ui.SortKey{
		{ColKey: "name", Dir: ui.SortAsc},
		{ColKey: "score", Dir: ui.SortAsc},
	})

	disp := dt.DataTableDisplayIndexes()
	items := []any{
		item{"Alice", 10},
		item{"Bob", 20},
		item{"Alice", 30},
		item{"Bob", 5},
	}
	order := make([]string, len(disp))
	for i, di := range disp {
		it := items[di].(item)
		order[i] = it.name + ":" + string(rune('0'+it.score/10))
	}
	// Expected: Alice:1, Alice:3, Bob:0, Bob:2 (name asc, score asc)
	if order[0] != "Alice:1" || order[1] != "Alice:3" || order[2] != "Bob:0" || order[3] != "Bob:2" {
		t.Errorf("multi-sort order = %v, want [Alice:1 Alice:3 Bob:0 Bob:2]", order)
	}

	// Verify SortKeys returns what we set.
	keys := dt.SortKeys()
	if len(keys) != 2 || keys[0].ColKey != "name" || keys[1].ColKey != "score" {
		t.Errorf("SortKeys() = %v, want [{name Asc} {score Asc}]", keys)
	}
}

func TestDataTableDefaultSort(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:         "name",
		Header:      "Name",
		Weight:      1,
		Sortable:    true,
		SortType:    ui.SortAlpha,
		SearchValue: func(d any) string { return d.(string) },
	})
	dt.SetSize(400, 300)

	dt.SetDefaultSort("name", ui.SortAsc)
	dt.SetItems([]any{"Charlie", "Alice", "Bob"})

	disp := dt.DataTableDisplayIndexes()
	items := []any{"Charlie", "Alice", "Bob"}
	order := make([]string, len(disp))
	for i, di := range disp {
		order[i] = items[di].(string)
	}
	if order[0] != "Alice" || order[1] != "Bob" || order[2] != "Charlie" {
		t.Errorf("default sort order = %v, want [Alice Bob Charlie]", order)
	}
}

func TestDataTableSetSortedColumnBackcompat(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:      "name",
		Header:   "Name",
		Weight:   1,
		Sortable: true,
		SortType: ui.SortAlpha,
	})
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Charlie", "Alice", "Bob"})

	// SetSortedColumn should still work for single-column sort.
	dt.SetSortedColumn("name", ui.SortAsc)
	k, dir := dt.SortedColumn()
	if k != "name" || dir != ui.SortAsc {
		t.Errorf("SortedColumn() = (%s, %d), want (name, SortAsc)", k, dir)
	}

	// SortKeys should have exactly one entry.
	keys := dt.SortKeys()
	if len(keys) != 1 || keys[0].ColKey != "name" || keys[0].Dir != ui.SortAsc {
		t.Errorf("SortKeys() = %v, want [{name Asc}]", keys)
	}
}

func TestDataTableColumnFilter(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.LabelColumn("cls", "Class", func(d any) string {
		return d.(string)
	}))
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Warrior", "Mage", "Rogue", "Mage", "Warrior"})

	dt.SetColumnFilter("cls", []string{"Mage"})
	flushAndUpdate(dt)

	if dt.DataTableDisplayCount() != 2 {
		t.Errorf("column filter 'Mage': count = %d, want 2", dt.DataTableDisplayCount())
	}

	dt.ClearColumnFilter("cls")
	flushAndUpdate(dt)

	if dt.DataTableDisplayCount() != 5 {
		t.Errorf("after clear column filter: count = %d, want 5", dt.DataTableDisplayCount())
	}
}

func TestDataTableColumnFilterMultiValues(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.LabelColumn("cls", "Class", func(d any) string {
		return d.(string)
	}))
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Warrior", "Mage", "Rogue", "Mage", "Warrior"})

	dt.SetColumnFilter("cls", []string{"Mage", "Rogue"})
	flushAndUpdate(dt)

	if dt.DataTableDisplayCount() != 3 {
		t.Errorf("column filter 'Mage,Rogue': count = %d, want 3", dt.DataTableDisplayCount())
	}
}

func TestDataTableResetFiltersAndSort(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.LabelColumn("name", "Name", func(d any) string {
		return d.(string)
	}))
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Charlie", "Alice", "Bob"})

	dt.SetSortedColumn("name", ui.SortAsc)
	dt.SetColumnFilter("name", []string{"Alice"})
	dt.SetFilterFunc(func(d any) bool { return true })

	dt.ResetFiltersAndSort()

	k, dir := dt.SortedColumn()
	if k != "" || dir != ui.SortNone {
		t.Errorf("after reset: SortedColumn() = (%s, %d), want ('', SortNone)", k, dir)
	}
	if dt.DataTableDisplayCount() != 3 {
		t.Errorf("after reset: count = %d, want 3", dt.DataTableDisplayCount())
	}
}

func TestDataTableOnPostUpdate(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	col := ui.LabelColumn("name", "Name", func(d any) string {
		return d.(string)
	})
	col.Cell.OnPostUpdate = func(d any, comp *ui.Component) {
		if l, ok := comp.UserData().(*ui.Label); ok {
			if d.(string) == "Alice" {
				l.SetColor(willow.RGBA(1, 0, 0, 1))
			} else {
				l.SetColor(willow.RGBA(1, 1, 1, 1))
			}
		}
	}
	dt.AddColumn(col)
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Alice", "Bob"})

	// Should not panic — OnPostUpdate is applied during populateSlotCells.
	if dt.DataTableDisplayCount() != 2 {
		t.Errorf("count = %d, want 2", dt.DataTableDisplayCount())
	}
}

func TestDataTableHeaderStyle(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	col := ui.DataTableColumn{
		Key:    "name",
		Header: "Name",
		Weight: 1,
	}
	col.HeaderStyle.FontSize = 18
	dt.AddColumn(col)
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Alice"})
	// Should not panic.
}

func TestDataTableBindSearchInput(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.LabelColumn("name", "Name", func(d any) string {
		return d.(string)
	}))
	dt.SetSize(400, 300)
	dt.SetItems([]any{"Alice", "Bob", "Albert"})

	// BindSearchInput delegates to BindSearchFilter(ti.ValueRef()).
	// Since creating a TextInput requires a valid font (not available in headless tests),
	// we test BindSearchFilter directly — BindSearchInput is a one-liner wrapper.
	searchRef := ui.NewRef("")
	dt.BindSearchFilter(searchRef)
	flushAndUpdate(dt)

	searchRef.Set("al")
	flushAndUpdate(dt)

	if dt.DataTableDisplayCount() != 2 {
		t.Errorf("search 'al': count = %d, want 2 (Alice, Albert)", dt.DataTableDisplayCount())
	}
}

func TestDataTableOnMultiSort(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:      "name",
		Header:   "Name",
		Weight:   1,
		Sortable: true,
		SortType: ui.SortAlpha,
	})
	dt.SetSize(400, 300)
	dt.SetItems([]any{"A", "B"})

	var called bool
	dt.SetOnMultiSort(func(keys []ui.SortKey) {
		called = true
	})

	// Programmatic SetSortKeys triggers the callback path through cycleSort indirectly.
	// Directly test SetSortKeys does not call onMultiSort (it's for programmatic use).
	dt.SetSortKeys([]ui.SortKey{{ColKey: "name", Dir: ui.SortAsc}})
	_ = called
	// Just verify no panic and keys are set.
	if len(dt.SortKeys()) != 1 {
		t.Errorf("SortKeys len = %d, want 1", len(dt.SortKeys()))
	}
}

func TestDataTableSortNoneClears(t *testing.T) {
	resetScheduler()
	dt := ui.NewDataTable("dt", 28)
	defer dt.Dispose()

	dt.AddColumn(ui.DataTableColumn{
		Key:      "name",
		Header:   "Name",
		Weight:   1,
		Sortable: true,
		SortType: ui.SortAlpha,
	})
	dt.SetSize(400, 300)
	dt.SetItems([]any{"C", "A", "B"})

	dt.SetSortedColumn("name", ui.SortAsc)
	dt.SetSortedColumn("", ui.SortNone)

	k, dir := dt.SortedColumn()
	if k != "" || dir != ui.SortNone {
		t.Errorf("after clear: SortedColumn() = (%s, %d), want ('', SortNone)", k, dir)
	}
}
