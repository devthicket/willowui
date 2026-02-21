package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func TestNewOptionRotator(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	opts := []string{"Low", "Medium", "High"}
	or := ui.NewOptionRotator("or", opts, font, 16)
	defer or.Dispose()

	if or.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if or.Name() != "or" {
		t.Errorf("Name() = %q, want %q", or.Name(), "or")
	}
	if or.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", or.Selected())
	}
	if or.Value() != "Low" {
		t.Errorf("Value() = %q, want %q", or.Value(), "Low")
	}
}

func TestNewOptionRotator_EmptyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty options slice")
		}
	}()
	font := newTestFont()
	ui.NewOptionRotator("or", []string{}, font, 16)
}

func TestOptionRotator_SetSelected(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	or := ui.NewOptionRotator("or", []string{"A", "B", "C"}, font, 16)
	defer or.Dispose()

	or.SetSelected(2)
	if or.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2", or.Selected())
	}
	if or.Value() != "C" {
		t.Errorf("Value() = %q, want %q", or.Value(), "C")
	}

	// Clamp above range.
	or.SetSelected(99)
	if or.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2 after clamping above range", or.Selected())
	}

	// Clamp below range.
	or.SetSelected(-5)
	if or.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0 after clamping below range", or.Selected())
	}

	// Same index: no change.
	or.SetSelected(0)
	or.SetSelected(0) // should be a no-op
	if or.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", or.Selected())
	}
}

func TestOptionRotator_NextPrev(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	or := ui.NewOptionRotator("or", []string{"A", "B", "C"}, font, 16)
	defer or.Dispose()

	or.Next()
	if or.Selected() != 1 {
		t.Errorf("after Next: Selected() = %d, want 1", or.Selected())
	}

	or.Prev()
	if or.Selected() != 0 {
		t.Errorf("after Prev: Selected() = %d, want 0", or.Selected())
	}

	// Wrap: Prev at index 0 should wrap to last.
	or.Prev()
	if or.Selected() != 2 {
		t.Errorf("wrap Prev at 0: Selected() = %d, want 2", or.Selected())
	}

	// Wrap: Next at last should wrap to 0.
	or.Next()
	if or.Selected() != 0 {
		t.Errorf("wrap Next at last: Selected() = %d, want 0", or.Selected())
	}
}

func TestOptionRotator_NoWrap(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	or := ui.NewOptionRotator("or", []string{"A", "B", "C"}, font, 16)
	defer or.Dispose()

	or.SetWrap(false)

	// Prev at index 0: no-op.
	or.Prev()
	if or.Selected() != 0 {
		t.Errorf("no-wrap Prev at 0: Selected() = %d, want 0", or.Selected())
	}

	or.SetSelected(2)

	// Next at last: no-op.
	or.Next()
	if or.Selected() != 2 {
		t.Errorf("no-wrap Next at last: Selected() = %d, want 2", or.Selected())
	}
}

func TestOptionRotator_SetOptions(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	or := ui.NewOptionRotator("or", []string{"A", "B", "C"}, font, 16)
	defer or.Dispose()

	or.SetSelected(2)
	or.SetOptions([]string{"X", "Y"})

	// Index clamped to new length - 1.
	if or.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1 after SetOptions with shorter list", or.Selected())
	}
	if or.Value() != "Y" {
		t.Errorf("Value() = %q, want %q", or.Value(), "Y")
	}

	// Options() returns a copy.
	got := or.Options()
	if len(got) != 2 || got[0] != "X" || got[1] != "Y" {
		t.Errorf("Options() = %v, want [X Y]", got)
	}
}

func TestOptionRotator_SetOptions_FiresOnChange(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	or := ui.NewOptionRotator("or", []string{"A", "B", "C"}, font, 16)
	defer or.Dispose()

	var called int
	or.SetOnChange(func(idx int, val string) { called++ })

	or.SetSelected(2)
	called = 0 // reset after initial SetSelected

	// SetOptions with shorter list clamps index → fires onChange.
	or.SetOptions([]string{"X", "Y"})
	if called != 1 {
		t.Errorf("onChange called %d times after SetOptions clamping, want 1", called)
	}

	called = 0
	// SetOptions without index change → no onChange.
	or.SetOptions([]string{"P", "Q"})
	if called != 0 {
		t.Errorf("onChange called %d times when index unchanged, want 0", called)
	}
}

func TestOptionRotator_OnChange(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	or := ui.NewOptionRotator("or", []string{"Easy", "Normal", "Hard"}, font, 16)
	defer or.Dispose()

	var callCount int
	var lastIdx int
	var lastVal string
	or.SetOnChange(func(idx int, val string) {
		callCount++
		lastIdx = idx
		lastVal = val
	})

	or.SetSelected(1)
	if callCount != 1 {
		t.Errorf("onChange called %d times, want 1", callCount)
	}
	if lastIdx != 1 || lastVal != "Normal" {
		t.Errorf("onChange got (%d, %q), want (1, Normal)", lastIdx, lastVal)
	}

	// Same index: no callback.
	or.SetSelected(1)
	if callCount != 1 {
		t.Errorf("onChange called %d times after same-index set, want 1", callCount)
	}

	or.Next()
	if callCount != 2 || lastIdx != 2 || lastVal != "Hard" {
		t.Errorf("after Next: onChange = (%d, %q), want (2, Hard)", lastIdx, lastVal)
	}
}

func TestOptionRotator_BindSelected(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	or := ui.NewOptionRotator("or", []string{"A", "B", "C"}, font, 16)
	defer or.Dispose()

	ref := ui.NewRef(2)
	or.BindSelected(ref)

	if or.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2 after BindSelected", or.Selected())
	}

	// External ref change updates widget.
	ref.Set(0)
	ui.DefaultScheduler.Flush()
	if or.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0 after ref.Set(0)", or.Selected())
	}

	// Widget change updates bound ref.
	or.SetSelected(1)
	if ref.Peek() != 1 {
		t.Errorf("ref.Peek() = %d, want 1 after SetSelected(1)", ref.Peek())
	}
}

func TestOptionRotator_BindValue(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	or := ui.NewOptionRotator("or", []string{"Low", "Medium", "High"}, font, 16)
	defer or.Dispose()

	ref := ui.NewRef("Medium")
	or.BindValue(ref)

	if or.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1 after BindValue(Medium)", or.Selected())
	}

	// External ref change updates widget.
	ref.Set("High")
	ui.DefaultScheduler.Flush()
	if or.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2 after ref.Set(High)", or.Selected())
	}

	// Widget change updates bound ref.
	or.SetSelected(0)
	if ref.Peek() != "Low" {
		t.Errorf("ref.Peek() = %q, want %q after SetSelected(0)", ref.Peek(), "Low")
	}

	// Value not in options: silently ignored (stays at current).
	ref.Set("Ultra")
	ui.DefaultScheduler.Flush()
	if or.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0 for unknown value", or.Selected())
	}
}

func TestOptionRotator_BindReplacesPrevious(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	or := ui.NewOptionRotator("or", []string{"A", "B", "C"}, font, 16)
	defer or.Dispose()

	idxRef := ui.NewRef(0)
	or.BindSelected(idxRef)

	valRef := ui.NewRef("B")
	or.BindValue(valRef) // replaces BindSelected

	if or.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1 after BindValue replaces BindSelected", or.Selected())
	}

	// Old idxRef should no longer drive the widget.
	idxRef.Set(2)
	ui.DefaultScheduler.Flush()
	if or.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1 — old idxRef should be inactive", or.Selected())
	}
}

func TestOptionRotator_SelectedRef(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	or := ui.NewOptionRotator("or", []string{"A", "B", "C"}, font, 16)
	defer or.Dispose()

	ref := or.SelectedRef()
	if ref == nil {
		t.Fatal("SelectedRef() should not be nil")
	}
	if ref.Peek() != 0 {
		t.Errorf("SelectedRef().Peek() = %d, want 0", ref.Peek())
	}

	or.SetSelected(2)
	if ref.Peek() != 2 {
		t.Errorf("SelectedRef().Peek() = %d, want 2 after SetSelected(2)", ref.Peek())
	}
}

func TestOptionRotator_Dispose(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	or := ui.NewOptionRotator("or", []string{"A", "B"}, font, 16)
	or.Dispose() // must not panic
}
