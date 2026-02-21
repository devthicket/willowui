package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

// ---------------------------------------------------------------------------
// Mask parsing / slot kinds
// ---------------------------------------------------------------------------

func TestMaskedInputSetMask(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99/9999")
	if mi.Mask() != "99/99/9999" {
		t.Errorf("Mask() = %q, want %q", mi.Mask(), "99/99/9999")
	}
	if mi.IsEmpty() == false {
		t.Error("fresh mask should be empty")
	}
}

func TestMaskedInputCapacity(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("AAAA-9999") // 4 + 4 = 8 slots
	mi.SetRawValue("ABCD1234")
	if !mi.IsComplete() {
		t.Error("IsComplete() should be true after filling all slots")
	}
	if mi.RawValue() != "ABCD1234" {
		t.Errorf("RawValue() = %q, want %q", mi.RawValue(), "ABCD1234")
	}
}

// ---------------------------------------------------------------------------
// Typing: valid characters
// ---------------------------------------------------------------------------

func TestMaskedInputTypeDigit(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99")
	mi.SetRawValue("12")
	if mi.RawValue() != "12" {
		t.Errorf("RawValue() = %q, want %q", mi.RawValue(), "12")
	}
}

func TestMaskedInputAutoUppercase(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("AAAA-9999")
	mi.SetRawValue("abcd1234")
	if mi.RawValue() != "ABCD1234" {
		t.Errorf("RawValue() = %q, want ABCD1234 (auto-uppercased)", mi.RawValue())
	}
}

func TestMaskedInputUpperAlnum(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("XXX")
	mi.SetRawValue("a2b")
	if mi.RawValue() != "A2B" {
		t.Errorf("RawValue() = %q, want A2B (auto-uppercased X slot)", mi.RawValue())
	}
}

func TestMaskedInputRejectInvalid(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	// '9' slot should reject letters.
	mi.SetMask("9")
	mi.SetRawValue("A") // 'A' is not a digit
	if mi.RawValue() != "" {
		t.Errorf("RawValue() = %q, want empty (invalid char for digit slot)", mi.RawValue())
	}
}

// ---------------------------------------------------------------------------
// Value vs RawValue
// ---------------------------------------------------------------------------

func TestMaskedInputValueFormatted(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("AAAA-9999")
	mi.SetRawValue("ABCD1234")

	if mi.Value() != "ABCD-1234" {
		t.Errorf("Value() = %q, want %q", mi.Value(), "ABCD-1234")
	}
	if mi.RawValue() != "ABCD1234" {
		t.Errorf("RawValue() = %q, want %q", mi.RawValue(), "ABCD1234")
	}
}

func TestMaskedInputValuePartial(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99/9999")
	mi.SetRawValue("01")

	// Formatted: "01/__/____" but without placeholder set, empties are spaces.
	raw := mi.RawValue()
	if raw != "01" {
		t.Errorf("RawValue() = %q, want %q", raw, "01")
	}
}

// ---------------------------------------------------------------------------
// SetValue (formatted input)
// ---------------------------------------------------------------------------

func TestMaskedInputSetValueFormatted(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99/9999")
	mi.SetValue("01/23/2026")

	if mi.RawValue() != "01232026" {
		t.Errorf("RawValue() = %q, want %q", mi.RawValue(), "01232026")
	}
	if !mi.IsComplete() {
		t.Error("IsComplete() should be true")
	}
}

// ---------------------------------------------------------------------------
// Paste behavior
// ---------------------------------------------------------------------------

func TestMaskedInputPasteSkipsLiterals(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99/9999")
	// Paste a formatted date — the '/' separators are invalid for digit slots and
	// should be skipped, while the digit chars fill the slots sequentially.
	mi.InsertText("01-23-2026")
	if mi.RawValue() != "01232026" {
		t.Errorf("RawValue() = %q, want %q", mi.RawValue(), "01232026")
	}
}

func TestMaskedInputPastePartial(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("AAAA-9999")
	// InsertText fills sequentially: A→slot0(upper), B→slot1(upper), '!' skipped (not letter),
	// C→slot2(upper), '1' skipped (not letter for slot3=upper), '2' skipped.
	// Result: "ABC" in slots 0-2; slot3 (upper) and digit slots remain empty.
	mi.InsertText("AB!C12")
	raw := mi.RawValue()
	if raw != "ABC" {
		t.Errorf("RawValue() = %q, want %q", raw, "ABC")
	}
}

// ---------------------------------------------------------------------------
// Backspace and Delete
// ---------------------------------------------------------------------------

func TestMaskedInputDeleteBack(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99")
	mi.SetRawValue("12")
	// cursorPos = 2 after SetRawValue
	mi.DeleteBack()
	if mi.RawValue() != "1" {
		t.Errorf("after DeleteBack: RawValue() = %q, want %q", mi.RawValue(), "1")
	}
	if mi.GetCursorPos() != 1 {
		t.Errorf("after DeleteBack: cursorPos = %d, want 1", mi.GetCursorPos())
	}
}

func TestMaskedInputDeleteForward(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99")
	mi.SetRawValue("12")
	mi.SetCursorPos(0)
	mi.DeleteForward()
	if mi.RawValue() != "2" {
		t.Errorf("after DeleteForward: RawValue() = %q, want %q", mi.RawValue(), "2")
	}
	if mi.GetCursorPos() != 0 {
		t.Errorf("after DeleteForward: cursorPos = %d, want 0", mi.GetCursorPos())
	}
}

func TestMaskedInputDeleteBackAtStart(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("9")
	mi.SetRawValue("5")
	mi.SetCursorPos(0)
	mi.DeleteBack() // no-op: already at start
	if mi.RawValue() != "5" {
		t.Errorf("DeleteBack at start should be no-op, got %q", mi.RawValue())
	}
}

// ---------------------------------------------------------------------------
// Cursor movement
// ---------------------------------------------------------------------------

func TestMaskedInputCursorMovement(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("AAAA-9999")
	mi.SetRawValue("ABCD1234")
	mi.SetCursorPos(4)
	mi.MoveCursorLeft()
	if mi.GetCursorPos() != 3 {
		t.Errorf("after left: cursorPos = %d, want 3", mi.GetCursorPos())
	}
	mi.MoveCursorRight()
	if mi.GetCursorPos() != 4 {
		t.Errorf("after right: cursorPos = %d, want 4", mi.GetCursorPos())
	}
}

func TestMaskedInputCursorAtBoundaries(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99")
	mi.SetCursorPos(0)
	mi.MoveCursorLeft() // no-op at start
	if mi.GetCursorPos() != 0 {
		t.Errorf("left at 0 should stay at 0, got %d", mi.GetCursorPos())
	}
	mi.SetCursorPos(2)
	mi.MoveCursorRight() // no-op at end
	if mi.GetCursorPos() != 2 {
		t.Errorf("right at capacity should stay at 2, got %d", mi.GetCursorPos())
	}
}

// ---------------------------------------------------------------------------
// Selection
// ---------------------------------------------------------------------------

func TestMaskedInputSelectAll(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99")
	mi.SetRawValue("1234")
	mi.SelectAll()
	if !mi.HasSelection() {
		t.Error("HasSelection() should be true after SelectAll()")
	}
	if mi.GetSelStart() != 0 {
		t.Errorf("selStart = %d, want 0", mi.GetSelStart())
	}
	if mi.GetSelEnd() != 4 {
		t.Errorf("selEnd = %d, want 4", mi.GetSelEnd())
	}
}

func TestMaskedInputSelectedText(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("AAAA-9999")
	mi.SetRawValue("ABCD1234")
	mi.SelectAll()
	// Formatted value includes the '-' separator.
	sel := mi.SelectedText()
	if sel != "ABCD-1234" {
		t.Errorf("SelectedText() = %q, want %q", sel, "ABCD-1234")
	}
}

func TestMaskedInputDeleteSelection(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99")
	mi.SetRawValue("1234")
	mi.SelectAll()
	mi.DeleteBack()
	if mi.RawValue() != "" {
		t.Errorf("after deleting selection: RawValue() = %q, want empty", mi.RawValue())
	}
	if !mi.IsEmpty() {
		t.Error("IsEmpty() should be true after deleting all")
	}
}

// ---------------------------------------------------------------------------
// IsComplete / IsEmpty
// ---------------------------------------------------------------------------

func TestMaskedInputCompleteIncomplete(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99")
	if !mi.IsEmpty() {
		t.Error("fresh field should be empty")
	}
	if mi.IsComplete() {
		t.Error("fresh field should not be complete")
	}
	mi.SetRawValue("1")
	if mi.IsEmpty() {
		t.Error("partial field should not be empty")
	}
	if mi.IsComplete() {
		t.Error("partial field should not be complete")
	}
	mi.SetRawValue("12")
	if !mi.IsComplete() {
		t.Error("fully filled field should be complete")
	}
}

// ---------------------------------------------------------------------------
// Callbacks
// ---------------------------------------------------------------------------

func TestMaskedInputOnComplete(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99")
	var completeCalled bool
	mi.SetOnComplete(func(raw, formatted string) {
		completeCalled = true
		if raw != "12" {
			t.Errorf("OnComplete raw = %q, want %q", raw, "12")
		}
		if formatted != "12" {
			t.Errorf("OnComplete formatted = %q, want %q", formatted, "12")
		}
	})
	mi.SetRawValue("12")
	if !completeCalled {
		t.Error("OnComplete should have been called")
	}
}

func TestMaskedInputOnIncomplete(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99")
	var incompleteCalled bool
	mi.SetOnIncomplete(func(raw, formatted string) {
		incompleteCalled = true
	})
	mi.SetRawValue("12") // complete
	mi.Clear()           // → incomplete
	if !incompleteCalled {
		t.Error("OnIncomplete should have been called when reverting from complete to empty")
	}
}

func TestMaskedInputOnChange(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99")
	var got string
	mi.SetOnChange(func(v string) { got = v })
	mi.SetRawValue("5")
	if got == "" {
		t.Error("OnChange should have been called")
	}
}

func TestMaskedInputOnRawChange(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("AAAA-9999")
	var got string
	mi.SetOnRawChange(func(v string) { got = v })
	mi.SetRawValue("ABCD1234")
	if got != "ABCD1234" {
		t.Errorf("OnRawChange got %q, want %q", got, "ABCD1234")
	}
}

// ---------------------------------------------------------------------------
// Reactive bindings
// ---------------------------------------------------------------------------

func TestMaskedInputBindValue(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("AAAA-9999")
	ref := ui.NewRef("ABCD-1234")
	mi.BindValue(ref)

	if mi.RawValue() != "ABCD1234" {
		t.Errorf("BindValue: RawValue() = %q, want %q", mi.RawValue(), "ABCD1234")
	}

	// Change from external ref: field should update.
	ref.Set("EFGH-5678")
	// The watch fires via the scheduler; flush to process pending updates.
	ui.DefaultScheduler.Flush()
	if mi.RawValue() != "EFGH5678" {
		t.Errorf("after ref.Set: RawValue() = %q, want %q", mi.RawValue(), "EFGH5678")
	}
}

func TestMaskedInputBindRawValue(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99")
	ref := ui.NewRef("1234")
	mi.BindRawValue(ref)

	if mi.Value() != "12/34" {
		t.Errorf("BindRawValue: Value() = %q, want %q", mi.Value(), "12/34")
	}
}

// ---------------------------------------------------------------------------
// Clear
// ---------------------------------------------------------------------------

func TestMaskedInputClear(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99")
	mi.SetRawValue("1234")
	mi.Clear()
	if !mi.IsEmpty() {
		t.Error("after Clear(): IsEmpty() should be true")
	}
	if mi.GetCursorPos() != 0 {
		t.Errorf("after Clear(): cursorPos = %d, want 0", mi.GetCursorPos())
	}
}

// ---------------------------------------------------------------------------
// MaskPlaceholder
// ---------------------------------------------------------------------------

func TestMaskedInputMaskPlaceholder(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99")
	mi.SetMaskPlaceholder('_')
	mi.SetRawValue("12")

	// With placeholder '_', display should be "12/__"
	if mi.Value() != "12/__" {
		t.Errorf("Value() = %q, want %q", mi.Value(), "12/__")
	}
}

// ---------------------------------------------------------------------------
// RawToDisplayIndex mapping
// ---------------------------------------------------------------------------

func TestMaskedInputRawToDisplayIndex(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	// mask AAAA-9999: tokens = A A A A - 9 9 9 9
	//                display =  0 1 2 3 4 5 6 7 8
	// rawToDisplayIndex(0) = 0, (1) = 1, ..., (4) = 4, (5) = 6, ..., (8) = 9
	mi.SetMask("AAAA-9999")

	tests := []struct {
		rawIdx  int
		wantIdx int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 6},
		{6, 7},
		{7, 8},
		{8, 9},
	}
	for _, tc := range tests {
		got := mi.RawToDisplayIndex(tc.rawIdx)
		if got != tc.wantIdx {
			t.Errorf("RawToDisplayIndex(%d) = %d, want %d", tc.rawIdx, got, tc.wantIdx)
		}
	}
}

// ---------------------------------------------------------------------------
// UX: snap to first empty on focus
// ---------------------------------------------------------------------------

func TestMaskedInputSnapToFirstEmpty(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99/99/9999")
	// Pre-fill month.
	mi.SetRawValue("12")
	// Simulate focus gain (tab/keyboard nav).
	mi.SnapCursorToFirstEmptyForTest()
	// Cursor should be at slot 2 (first empty slot after "12").
	if mi.GetCursorPos() != 2 {
		t.Errorf("after snap: cursorPos = %d, want 2", mi.GetCursorPos())
	}
}

func TestMaskedInputSnapWhenFull(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	mi.SetMask("99")
	mi.SetRawValue("12")
	mi.SnapCursorToFirstEmptyForTest()
	// All slots filled — cursor should be at end.
	if mi.GetCursorPos() != 2 {
		t.Errorf("after snap when full: cursorPos = %d, want 2 (end)", mi.GetCursorPos())
	}
}

// ---------------------------------------------------------------------------
// UX: visual display index advances past literals
// ---------------------------------------------------------------------------

func TestMaskedInputVisualDisplayIdx(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	// mask 99/99: tokens = [9,9,'/',9,9]
	// rawToDisplayIndex(2) = 2 (before the '/')
	// visualDisplayIdx(2) should = 3 (after the '/')
	mi.SetMask("99/99")
	got := mi.VisualDisplayIdxForTest(2)
	if got != 3 {
		t.Errorf("visualDisplayIdx(2) = %d, want 3 (past literal '/')", got)
	}
	// Position 0 has no preceding slot, no literal to skip at start.
	got = mi.VisualDisplayIdxForTest(0)
	if got != 0 {
		t.Errorf("visualDisplayIdx(0) = %d, want 0", got)
	}
	// Position 4 (end): rawToDisplayIndex(4)=5, no literals after → visual=5.
	got = mi.VisualDisplayIdxForTest(4)
	if got != 5 {
		t.Errorf("visualDisplayIdx(4) = %d, want 5 (end)", got)
	}
}

// ---------------------------------------------------------------------------
// Theme
// ---------------------------------------------------------------------------

func TestMaskedInputDefaultTheme(t *testing.T) {
	group := ui.DefaultTheme.MaskedInput.Group(ui.Primary)
	if group.FocusColor.Resolve(ui.StateFocus).A() == 0 {
		t.Error("MaskedInput theme should have a non-transparent FocusColor")
	}
	if group.LiteralColor.Resolve(ui.StateDefault).A() == 0 {
		t.Error("MaskedInput theme should have a non-transparent LiteralColor")
	}
	if group.MaskPlaceholderColor.Resolve(ui.StateDefault).A() == 0 {
		t.Error("MaskedInput theme should have a non-transparent MaskPlaceholderColor")
	}
}

// ---------------------------------------------------------------------------
// Focus flags
// ---------------------------------------------------------------------------

func TestMaskedInputFocusFlags(t *testing.T) {
	resetScheduler()
	mi := ui.NewMaskedInput("mi", newTestFont(), 14)
	defer mi.Dispose()

	if !mi.Focusable {
		t.Error("MaskedInput should be Focusable")
	}
	if !mi.AllowTab {
		t.Error("MaskedInput should have AllowTab")
	}
	if !mi.AllowSpatial {
		t.Error("MaskedInput should have AllowSpatial")
	}
	if !mi.InterceptArrows {
		t.Error("MaskedInput should have InterceptArrows")
	}
}
