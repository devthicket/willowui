package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func TestNewToggleButtonBar(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	bar := ui.NewToggleButtonBar("tbb", font, 16)
	defer bar.Dispose()

	if bar.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if bar.Name() != "tbb" {
		t.Errorf("Name() = %q, want %q", bar.Name(), "tbb")
	}
	if bar.ButtonCount() != 0 {
		t.Errorf("ButtonCount() = %d, want 0", bar.ButtonCount())
	}
	if bar.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", bar.Selected())
	}
}

func TestToggleButtonBar_AddButton(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	bar := ui.NewToggleButtonBar("tbb", font, 16)
	defer bar.Dispose()

	bar.AddButton("Option A")
	if bar.ButtonCount() != 1 {
		t.Errorf("ButtonCount() = %d, want 1", bar.ButtonCount())
	}

	bar.AddButton("Option B")
	if bar.ButtonCount() != 2 {
		t.Errorf("ButtonCount() = %d, want 2", bar.ButtonCount())
	}

	bar.AddButton("Option C")
	if bar.ButtonCount() != 3 {
		t.Errorf("ButtonCount() = %d, want 3", bar.ButtonCount())
	}
}

func TestToggleButtonBar_RemoveButton(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	bar := ui.NewToggleButtonBar("tbb", font, 16)
	defer bar.Dispose()

	bar.AddButton("A")
	bar.AddButton("B")
	bar.AddButton("C")

	bar.RemoveButton(1) // remove "B"
	if bar.ButtonCount() != 2 {
		t.Errorf("ButtonCount() = %d, want 2", bar.ButtonCount())
	}

	// Out-of-range removal should be a no-op.
	bar.RemoveButton(10)
	if bar.ButtonCount() != 2 {
		t.Errorf("ButtonCount() after invalid remove = %d, want 2", bar.ButtonCount())
	}

	bar.RemoveButton(-1)
	if bar.ButtonCount() != 2 {
		t.Errorf("ButtonCount() after negative remove = %d, want 2", bar.ButtonCount())
	}
}

func TestToggleButtonBar_SetSelected(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	bar := ui.NewToggleButtonBar("tbb", font, 16)
	defer bar.Dispose()

	bar.AddButton("A")
	bar.AddButton("B")
	bar.AddButton("C")

	bar.SetSelected(2)
	if bar.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2", bar.Selected())
	}

	bar.SetSelected(0)
	if bar.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", bar.Selected())
	}

	// Out-of-range: should be no-op.
	bar.SetSelected(10)
	if bar.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0 after invalid set", bar.Selected())
	}

	bar.SetSelected(-1)
	if bar.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0 after negative set", bar.Selected())
	}
}

func TestToggleButtonBar_Selected(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	bar := ui.NewToggleButtonBar("tbb", font, 16)
	defer bar.Dispose()

	bar.AddButton("A")
	bar.AddButton("B")

	// Default selection is 0.
	if bar.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", bar.Selected())
	}

	bar.SetSelected(1)
	if bar.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1", bar.Selected())
	}
}

func TestToggleButtonBar_BindSelected(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	bar := ui.NewToggleButtonBar("tbb", font, 16)
	defer bar.Dispose()

	bar.AddButton("A")
	bar.AddButton("B")
	bar.AddButton("C")

	ref := ui.NewRef(2)
	bar.BindSelected(ref)

	if bar.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2 after BindSelected", bar.Selected())
	}

	// Setting the ref should update the bar.
	ref.Set(1)
	ui.DefaultScheduler.Flush()
	if bar.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1 after ref.Set(1)", bar.Selected())
	}
}

func TestToggleButtonBar_SetOnChange(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	bar := ui.NewToggleButtonBar("tbb", font, 16)
	defer bar.Dispose()

	bar.AddButton("A")
	bar.AddButton("B")
	bar.AddButton("C")

	var called int
	var lastIdx int
	bar.SetOnChange(func(idx int) {
		called++
		lastIdx = idx
	})

	bar.SetSelected(1)
	if called != 1 {
		t.Errorf("onChange called %d times, want 1", called)
	}
	if lastIdx != 1 {
		t.Errorf("onChange lastIdx = %d, want 1", lastIdx)
	}

	// Setting to the same index should not fire onChange.
	bar.SetSelected(1)
	if called != 1 {
		t.Errorf("onChange called %d times after same-index set, want 1", called)
	}

	bar.SetSelected(2)
	if called != 2 {
		t.Errorf("onChange called %d times, want 2", called)
	}
	if lastIdx != 2 {
		t.Errorf("onChange lastIdx = %d, want 2", lastIdx)
	}
}

func TestToggleButtonBar_Dispose(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	bar := ui.NewToggleButtonBar("tbb", font, 16)

	bar.AddButton("A")
	bar.AddButton("B")

	// Dispose should not panic and should clean up entries.
	bar.Dispose()
	if !bar.EntriesIsNil() {
		t.Error("entries should be nil after Dispose")
	}
}

func TestToggleButtonBar_RemoveAdjustsSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	bar := ui.NewToggleButtonBar("tbb", font, 16)
	defer bar.Dispose()

	bar.AddButton("A")
	bar.AddButton("B")
	bar.AddButton("C")

	bar.SetSelected(2)  // select last
	bar.RemoveButton(2) // remove the selected button

	// Selection should adjust to the new last index.
	if bar.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1 after removing selected last button", bar.Selected())
	}
}
