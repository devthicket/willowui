package integration

import (
	"fmt"
	"testing"

	ui "github.com/devthicket/willowui"
)

// ---------------------------------------------------------------------------
// SortableList
// ---------------------------------------------------------------------------

func TestNewSortableListDefaults(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	if sl.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if sl.SortableListScrollBar() == nil {
		t.Fatal("scrollBar should not be nil")
	}
	if sl.ItemCount() != 0 {
		t.Errorf("ItemCount() = %d, want 0", sl.ItemCount())
	}
	if sl.Selected() != -1 {
		t.Errorf("Selected() = %d, want -1", sl.Selected())
	}
}

func TestSortableListBindItems(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B", "C", "D"})
	ui.BindSortableListItems(sl, items)

	if sl.ItemCount() != 4 {
		t.Errorf("ItemCount() = %d, want 4", sl.ItemCount())
	}
}

func TestSortableListMoveItem(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B", "C", "D"})
	ui.BindSortableListItems(sl, items)

	sl.MoveItem(0, 2)

	// After Move(0,2): [B, C, A, D]
	if items.At(0) != "B" {
		t.Errorf("items[0] = %q, want B", items.At(0))
	}
	if items.At(2) != "A" {
		t.Errorf("items[2] = %q, want A", items.At(2))
	}
}

func TestSortableListMoveItemSelectionFollows(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B", "C", "D"})
	ui.BindSortableListItems(sl, items)

	// Select item A at index 0, then move it to index 2.
	sl.SetSelected(0)
	sl.MoveItem(0, 2)

	// Selection should follow the moved item to index 2.
	if sl.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2 (selection should follow moved item)", sl.Selected())
	}
}

func TestSortableListMoveSelectedUp(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B", "C"})
	ui.BindSortableListItems(sl, items)

	sl.SetSelected(2)
	sl.MoveSelectedUp()

	// C should now be at index 1.
	if items.At(1) != "C" {
		t.Errorf("items[1] = %q, want C", items.At(1))
	}
	if sl.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1", sl.Selected())
	}
}

func TestSortableListMoveSelectedDown(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B", "C"})
	ui.BindSortableListItems(sl, items)

	sl.SetSelected(0)
	sl.MoveSelectedDown()

	// A should now be at index 1.
	if items.At(1) != "A" {
		t.Errorf("items[1] = %q, want A", items.At(1))
	}
	if sl.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1", sl.Selected())
	}
}

func TestSortableListDragClamp(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B", "C"})
	ui.BindSortableListItems(sl, items)

	// Move beyond bounds should be no-op.
	sl.SetSelected(0)
	sl.MoveSelectedUp() // already at top
	if sl.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0 (should stay at top)", sl.Selected())
	}

	sl.SetSelected(2)
	sl.MoveSelectedDown() // already at bottom
	if sl.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2 (should stay at bottom)", sl.Selected())
	}
}

func TestSortableListOnReorder(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B", "C"})
	ui.BindSortableListItems(sl, items)

	var gotFrom, gotTo int
	sl.SetOnReorder(func(from, to int) {
		gotFrom = from
		gotTo = to
	})

	sl.MoveItem(0, 2)

	if gotFrom != 0 || gotTo != 2 {
		t.Errorf("OnReorder got (%d, %d), want (0, 2)", gotFrom, gotTo)
	}
}

func TestSortableListSetSelected(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B", "C"})
	ui.BindSortableListItems(sl, items)

	sl.SetSelected(1)
	if sl.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1", sl.Selected())
	}

	// Out of range should be no-op.
	sl.SetSelected(10)
	if sl.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1 (out of range ignored)", sl.Selected())
	}
}

func TestSortableListBindSelected(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B", "C"})
	ui.BindSortableListItems(sl, items)

	ref := ui.NewRef(2)
	sl.BindSelected(ref)

	if sl.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2", sl.Selected())
	}
}

func TestSortableListHandleSide(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetHandleSide(ui.SortHandleRight)

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B"})
	ui.BindSortableListItems(sl, items)

	// Should not panic; handles should be on the right.
	if sl.ItemCount() != 2 {
		t.Errorf("ItemCount() = %d, want 2", sl.ItemCount())
	}
}

func TestSortableListDragDisabled(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetDragEnabled(false)

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B"})
	ui.BindSortableListItems(sl, items)

	// MoveItem should still work (it's a programmatic API).
	sl.MoveItem(0, 1)
	if items.At(0) != "B" {
		t.Errorf("items[0] = %q, want B", items.At(0))
	}
}

func TestSortableListSelectedItem(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B", "C"})
	ui.BindSortableListItems(sl, items)

	sl.SetSelected(1)
	got := sl.SelectedItem()
	if got != "B" {
		t.Errorf("SelectedItem() = %v, want B", got)
	}

	sl.SetSelected(-1)
	if sl.SelectedItem() != nil {
		t.Error("SelectedItem() should be nil when nothing selected")
	}
}

func TestSortableListSelectionAdjustOnNonSelectedMove(t *testing.T) {
	resetScheduler()
	sl := ui.NewSortableList("sl", 30)
	defer sl.Dispose()

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		return ui.NewComponent(fmt.Sprintf("item-%d", index))
	})

	items := ui.NewArrayFrom([]string{"A", "B", "C", "D"})
	ui.BindSortableListItems(sl, items)

	// Select C at index 2, move A (index 0) to index 3.
	sl.SetSelected(2)
	sl.MoveItem(0, 3)
	// [B, C, D, A] — C was at 2, now at 1 because A was removed before it.
	if sl.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1 (should adjust when item moves past)", sl.Selected())
	}
}
