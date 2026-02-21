package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

// ---------------------------------------------------------------------------
// TreeTable
// ---------------------------------------------------------------------------

func sampleTreeTableColumns() []ui.TableColumn {
	return []ui.TableColumn{
		{Key: "name", Label: "Name", Width: 200, Sortable: true},
		{Key: "type", Label: "Type", Width: 100, Sortable: false},
		{Key: "value", Label: "Value", Width: 150, Sortable: false},
	}
}

func sampleTreeTableRows() []ui.TreeTableRow {
	return []ui.TreeTableRow{
		{
			ID:    "entity",
			Cells: map[string]string{"name": "Entity", "type": "Group"},
			Children: []ui.TreeTableRow{
				{
					ID:    "transform",
					Cells: map[string]string{"name": "Transform", "type": "Object"},
					Children: []ui.TreeTableRow{
						{ID: "pos-x", Cells: map[string]string{"name": "Position X", "type": "float", "value": "10.5"}},
						{ID: "pos-y", Cells: map[string]string{"name": "Position Y", "type": "float", "value": "20.3"}},
					},
				},
				{
					ID:    "renderer",
					Cells: map[string]string{"name": "Renderer", "type": "Object"},
					Children: []ui.TreeTableRow{
						{ID: "material", Cells: map[string]string{"name": "Material", "type": "string", "value": "metal"}},
					},
				},
			},
		},
		{
			ID:    "player",
			Cells: map[string]string{"name": "Player", "type": "Group"},
			Children: []ui.TreeTableRow{
				{ID: "health", Cells: map[string]string{"name": "Health", "type": "int", "value": "100"}},
			},
		},
	}
}

func TestTreeTableDefaults(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	if tt.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if tt.Selected() != -1 {
		t.Errorf("Selected() = %d, want -1", tt.Selected())
	}
	if len(tt.Columns()) != 0 {
		t.Errorf("Columns() len = %d, want 0", len(tt.Columns()))
	}
	if len(tt.Rows()) != 0 {
		t.Errorf("Rows() len = %d, want 0", len(tt.Rows()))
	}
}

func TestTreeTableSetColumnsAndRows(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	tt.SetColumns(sampleTreeTableColumns())
	tt.SetRows(sampleTreeTableRows())

	cols := tt.Columns()
	if len(cols) != 3 {
		t.Fatalf("Columns() len = %d, want 3", len(cols))
	}
	if cols[0].Key != "name" {
		t.Errorf("cols[0].Key = %q, want %q", cols[0].Key, "name")
	}

	rows := tt.Rows()
	if len(rows) != 2 {
		t.Fatalf("Rows() len = %d, want 2 (root rows)", len(rows))
	}
}

func TestTreeTableRootRowsVisible(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	tt.SetColumns(sampleTreeTableColumns())
	tt.SetRows(sampleTreeTableRows())

	if tt.VisibleRowCount() != 2 {
		t.Errorf("VisibleRowCount() = %d, want 2", tt.VisibleRowCount())
	}
}

func TestTreeTableExpandCollapse(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	tt.SetColumns(sampleTreeTableColumns())
	tt.SetRows(sampleTreeTableRows())

	if tt.IsExpanded("entity") {
		t.Error("entity should not be expanded initially")
	}

	tt.SetExpanded("entity", true)
	if !tt.IsExpanded("entity") {
		t.Error("entity should be expanded after SetExpanded(true)")
	}
	if tt.VisibleRowCount() != 4 {
		t.Errorf("VisibleRowCount() = %d, want 4", tt.VisibleRowCount())
	}

	tt.SetExpanded("entity", false)
	if tt.IsExpanded("entity") {
		t.Error("entity should be collapsed")
	}
	if tt.VisibleRowCount() != 2 {
		t.Errorf("VisibleRowCount() = %d, want 2", tt.VisibleRowCount())
	}
}

func TestTreeTableExpandAll(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	tt.SetColumns(sampleTreeTableColumns())
	tt.SetRows(sampleTreeTableRows())
	tt.ExpandAll()

	if tt.VisibleRowCount() != 8 {
		t.Errorf("VisibleRowCount() = %d, want 8", tt.VisibleRowCount())
	}
}

func TestTreeTableCollapseAll(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	tt.SetColumns(sampleTreeTableColumns())
	tt.SetRows(sampleTreeTableRows())
	tt.ExpandAll()
	tt.CollapseAll()

	if tt.VisibleRowCount() != 2 {
		t.Errorf("VisibleRowCount() = %d, want 2", tt.VisibleRowCount())
	}
}

func TestTreeTableOnRowClickNotFiredBySetSelected(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	tt.SetColumns(sampleTreeTableColumns())
	tt.SetRows(sampleTreeTableRows())

	callCount := 0
	tt.SetOnRowClick(func(id string) { callCount++ })

	tt.SetSelected(0)
	if callCount != 0 {
		t.Errorf("OnRowClick fired %d times on SetSelected, want 0", callCount)
	}
}

func TestTreeTableOnRowExpand(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	tt.SetColumns(sampleTreeTableColumns())
	tt.SetRows(sampleTreeTableRows())

	var expandedID string
	var expandedState bool
	tt.SetOnRowExpand(func(id string, expanded bool) {
		expandedID = id
		expandedState = expanded
	})

	tt.SetExpanded("entity", true)
	if expandedID != "entity" || !expandedState {
		t.Errorf("OnRowExpand got (%q, %v), want (%q, true)", expandedID, expandedState, "entity")
	}

	tt.SetExpanded("entity", false)
	if expandedID != "entity" || expandedState {
		t.Errorf("OnRowExpand got (%q, %v), want (%q, false)", expandedID, expandedState, "entity")
	}
}

func TestTreeTableSortColumn(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	tt.SetColumns(sampleTreeTableColumns())
	tt.SetRows(sampleTreeTableRows())

	tt.SetSortColumn("name", ui.SortDirAsc)
	if id := tt.RowIDAt(0); id != "entity" {
		t.Errorf("asc row 0 = %q, want %q", id, "entity")
	}
	if id := tt.RowIDAt(1); id != "player" {
		t.Errorf("asc row 1 = %q, want %q", id, "player")
	}

	tt.SetSortColumn("name", ui.SortDirDesc)
	if id := tt.RowIDAt(0); id != "player" {
		t.Errorf("desc row 0 = %q, want %q", id, "player")
	}

	tt.ExpandAll()
	tt.SetSortColumn("name", ui.SortDirAsc)
	if tt.VisibleRowCount() != 8 {
		t.Fatalf("VisibleRowCount() = %d, want 8", tt.VisibleRowCount())
	}
	// Under entity, sorted children: Renderer < Transform.
	if id := tt.RowIDAt(1); id != "renderer" {
		t.Errorf("entity child 0 = %q, want %q", id, "renderer")
	}
}

func TestTreeTableSelection(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	tt.SetColumns(sampleTreeTableColumns())
	tt.SetRows(sampleTreeTableRows())

	if tt.Selected() != -1 {
		t.Errorf("Selected() = %d, want -1", tt.Selected())
	}
	tt.SetSelected(0)
	if tt.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", tt.Selected())
	}
	tt.SetSelected(99) // out of range
	if tt.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0 (out of range ignored)", tt.Selected())
	}
}

func TestTreeTableSetSize(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	tt.SetSize(600, 400)
	if tt.Width != 600 {
		t.Errorf("Width = %f, want 600", tt.Width)
	}
	if tt.Height != 400 {
		t.Errorf("Height = %f, want 400", tt.Height)
	}
}

func TestTreeTableDispose(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	tt.SetColumns(sampleTreeTableColumns())
	tt.SetRows(sampleTreeTableRows())
	tt.ExpandAll()
	tt.Dispose() // should not panic
}

func TestTreeTableVirtualRowCount(t *testing.T) {
	resetScheduler()
	tt := ui.NewTreeTable("tt", newTestFont(), 13)
	defer tt.Dispose()

	tt.SetColumns(sampleTreeTableColumns())
	tt.SetRows(sampleTreeTableRows())

	if tt.VisibleRowCount() != 2 {
		t.Errorf("VisibleRowCount() = %d, want 2", tt.VisibleRowCount())
	}
	tt.SetExpanded("entity", true)
	if tt.VisibleRowCount() != 4 {
		t.Errorf("VisibleRowCount() = %d, want 4", tt.VisibleRowCount())
	}
	tt.SetExpanded("transform", true)
	if tt.VisibleRowCount() != 6 {
		t.Errorf("VisibleRowCount() = %d, want 6", tt.VisibleRowCount())
	}
}
