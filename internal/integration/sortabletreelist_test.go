package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

// ---------------------------------------------------------------------------
// SortableTreeList
// ---------------------------------------------------------------------------

func sampleTreeItems() []ui.SortableTreeItem {
	return []ui.SortableTreeItem{
		{ID: "bg", Label: "Background", ParentID: ""},
		{ID: "chars", Label: "Characters", ParentID: ""},
		{ID: "hero", Label: "Hero", ParentID: "chars"},
		{ID: "villain", Label: "Villain", ParentID: "chars"},
		{ID: "ui", Label: "UI Layer", ParentID: ""},
	}
}

func TestSortableTreeListDefaults(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	if st.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if st.Selected() != -1 {
		t.Errorf("Selected() = %d, want -1", st.Selected())
	}
	if len(st.Items()) != 0 {
		t.Errorf("Items() len = %d, want 0", len(st.Items()))
	}
}

func TestSortableTreeListSetItems(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	st.SetItems(sampleTreeItems())
	st.ExpandAll()

	items := st.Items()
	if len(items) != 5 {
		t.Fatalf("Items() len = %d, want 5", len(items))
	}
	// Verify depths are computed: root items have depth 0, children have depth 1.
	for _, item := range items {
		if item.ParentID == "" && item.Depth != 0 {
			t.Errorf("item %q depth = %d, want 0", item.ID, item.Depth)
		}
		if item.ParentID == "chars" && item.Depth != 1 {
			t.Errorf("item %q depth = %d, want 1", item.ID, item.Depth)
		}
	}
}

func TestSortableTreeListExpandCollapse(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	st.SetItems(sampleTreeItems())

	// Initially nothing is expanded.
	if st.IsExpanded("chars") {
		t.Error("chars should not be expanded initially")
	}

	// Expand chars.
	st.SetExpanded("chars", true)
	if !st.IsExpanded("chars") {
		t.Error("chars should be expanded after SetExpanded(true)")
	}

	// ExpandAll then CollapseAll.
	st.ExpandAll()
	if !st.IsExpanded("chars") {
		t.Error("chars should be expanded after ExpandAll")
	}
	st.CollapseAll()
	if st.IsExpanded("chars") {
		t.Error("chars should not be expanded after CollapseAll")
	}
}

func TestSortableTreeListSelection(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	st.SetItems(sampleTreeItems())
	st.ExpandAll()

	st.SetSelected(0)
	if st.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", st.Selected())
	}

	st.SetSelected(2)
	if st.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2", st.Selected())
	}

	// Out of range should be ignored.
	st.SetSelected(99)
	if st.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2 (out of range ignored)", st.Selected())
	}
}

func TestSortableTreeListOnChange(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	st.SetItems(sampleTreeItems())
	st.ExpandAll()

	var gotIdx int
	called := false
	st.SetOnChange(func(idx int) {
		called = true
		gotIdx = idx
	})

	st.SetSelected(1)
	if !called {
		t.Fatal("OnChange not called")
	}
	if gotIdx != 1 {
		t.Errorf("OnChange got %d, want 1", gotIdx)
	}
}

func TestSortableTreeListMoveSelectedUp(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	st.SetItems(sampleTreeItems())
	st.ExpandAll()

	// Flat list with all expanded: bg(0), chars(1), hero(2), villain(3), ui(4)
	// Select "villain" (flat index 3, sibling index 1 under "chars").
	st.SetSelected(3)

	var reorderID, reorderParent string
	var reorderIdx int
	st.SetOnReorder(func(itemID, newParentID string, newIndex int) {
		reorderID = itemID
		reorderParent = newParentID
		reorderIdx = newIndex
	})

	st.MoveSelectedUp()

	// villain should now be before hero under chars.
	if reorderID != "villain" {
		t.Errorf("OnReorder itemID = %q, want villain", reorderID)
	}
	if reorderParent != "chars" {
		t.Errorf("OnReorder newParentID = %q, want chars", reorderParent)
	}
	if reorderIdx != 0 {
		t.Errorf("OnReorder newIndex = %d, want 0", reorderIdx)
	}

	// Verify items order: villain should now be at sibling index 0 under chars.
	items := st.Items()
	var charsChildren []string
	for _, item := range items {
		if item.ParentID == "chars" {
			charsChildren = append(charsChildren, item.ID)
		}
	}
	if len(charsChildren) != 2 || charsChildren[0] != "villain" || charsChildren[1] != "hero" {
		t.Errorf("children of chars = %v, want [villain hero]", charsChildren)
	}
}

func TestSortableTreeListMoveSelectedDown(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	st.SetItems(sampleTreeItems())
	st.ExpandAll()

	// Select "hero" (flat index 2, sibling index 0 under "chars").
	st.SetSelected(2)

	var reorderID string
	var reorderIdx int
	st.SetOnReorder(func(itemID, newParentID string, newIndex int) {
		reorderID = itemID
		reorderIdx = newIndex
	})

	st.MoveSelectedDown()

	if reorderID != "hero" {
		t.Errorf("OnReorder itemID = %q, want hero", reorderID)
	}
	if reorderIdx != 1 {
		t.Errorf("OnReorder newIndex = %d, want 1", reorderIdx)
	}
}

func TestSortableTreeListMoveSelectedUpAtTop(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	st.SetItems(sampleTreeItems())
	st.ExpandAll()

	// Select first item "bg" (flat index 0) — move up should be a no-op.
	st.SetSelected(0)

	called := false
	st.SetOnReorder(func(itemID, newParentID string, newIndex int) {
		called = true
	})

	st.MoveSelectedUp()

	if called {
		t.Error("OnReorder should not fire when move up is a no-op")
	}
	if st.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", st.Selected())
	}
}

func TestSortableTreeListIndentSelected(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	st.SetAllowCrossLevel(true)
	st.SetItems(sampleTreeItems())
	st.ExpandAll()

	// Flat list: bg(0), chars(1), hero(2), villain(3), ui(4)
	// Select "ui" (flat index 4, root level). Preceding sibling at root is "chars".
	// Indenting should make "ui" a child of "chars".
	st.SetSelected(4)

	var reorderID, reorderParent string
	st.SetOnReorder(func(itemID, newParentID string, newIndex int) {
		reorderID = itemID
		reorderParent = newParentID
	})

	st.IndentSelected()

	if reorderID != "ui" {
		t.Errorf("OnReorder itemID = %q, want ui", reorderID)
	}
	if reorderParent != "chars" {
		t.Errorf("OnReorder newParentID = %q, want chars", reorderParent)
	}

	// Verify the item's parent changed.
	for _, item := range st.Items() {
		if item.ID == "ui" && item.ParentID != "chars" {
			t.Errorf("ui ParentID = %q, want chars", item.ParentID)
		}
	}
}

func TestSortableTreeListOutdentSelected(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	st.SetAllowCrossLevel(true)
	st.SetItems(sampleTreeItems())
	st.ExpandAll()

	// Select "hero" (child of "chars"). Outdenting should move hero to root level.
	st.SetSelected(2)

	var reorderID, reorderParent string
	st.SetOnReorder(func(itemID, newParentID string, newIndex int) {
		reorderID = itemID
		reorderParent = newParentID
	})

	st.OutdentSelected()

	if reorderID != "hero" {
		t.Errorf("OnReorder itemID = %q, want hero", reorderID)
	}
	if reorderParent != "" {
		t.Errorf("OnReorder newParentID = %q, want empty (root)", reorderParent)
	}

	// Verify the item is now at root.
	for _, item := range st.Items() {
		if item.ID == "hero" && item.ParentID != "" {
			t.Errorf("hero ParentID = %q, want empty", item.ParentID)
		}
	}
}

func TestSortableTreeListOutdentAtRoot(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	st.SetAllowCrossLevel(true)
	st.SetItems(sampleTreeItems())
	st.ExpandAll()

	// Select "bg" (already root). Outdent should be a no-op.
	st.SetSelected(0)

	called := false
	st.SetOnReorder(func(itemID, newParentID string, newIndex int) {
		called = true
	})

	st.OutdentSelected()

	if called {
		t.Error("OnReorder should not fire when outdent is a no-op (already root)")
	}
}

func TestSortableTreeListSetSize(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	defer st.Dispose()

	st.SetItems(sampleTreeItems())
	st.SetSize(300, 500)

	if st.Width != 300 {
		t.Errorf("Width = %f, want 300", st.Width)
	}
	if st.Height != 500 {
		t.Errorf("Height = %f, want 500", st.Height)
	}
}

func TestSortableTreeListDispose(t *testing.T) {
	resetScheduler()
	st := ui.NewSortableTreeList("st", newTestFont(), 13)
	st.SetItems(sampleTreeItems())
	st.ExpandAll()

	// Should not panic.
	st.Dispose()
}
