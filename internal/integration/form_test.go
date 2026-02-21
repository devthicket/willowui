package integration

import (
	"testing"

	"github.com/atotto/clipboard"
	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/widget"
)

// ---------------------------------------------------------------------------
// Toggle
// ---------------------------------------------------------------------------

func TestNewToggleDefaults(t *testing.T) {
	resetScheduler()
	tgl := ui.NewToggle("tgl")
	defer tgl.Dispose()

	if tgl.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if tgl.Value() != false {
		t.Error("initial value should be false")
	}
	if tgl.TrackNode() == nil {
		t.Fatal("track should not be nil")
	}
	if tgl.ThumbNode() == nil {
		t.Fatal("thumb should not be nil")
	}
}

func TestToggleSetValueUpdatesState(t *testing.T) {
	resetScheduler()
	tgl := ui.NewToggle("tgl")
	defer tgl.Dispose()

	tgl.SetValue(true)
	if !tgl.Value() {
		t.Error("Value() should be true after SetValue(true)")
	}

	tgl.SetValue(false)
	if tgl.Value() {
		t.Error("Value() should be false after SetValue(false)")
	}
}

func TestToggleClickTogglesValue(t *testing.T) {
	resetScheduler()
	tgl := ui.NewToggle("tgl")
	defer tgl.Dispose()

	// Simulate click via OnClick callback.
	tgl.Node().GetOnClick()(willow.ClickContext{Node: tgl.Node()})
	if !tgl.Value() {
		t.Error("click should toggle value to true")
	}

	tgl.Node().GetOnClick()(willow.ClickContext{Node: tgl.Node()})
	if tgl.Value() {
		t.Error("second click should toggle value back to false")
	}
}

func TestToggleDisabledBlocksClick(t *testing.T) {
	resetScheduler()
	tgl := ui.NewToggle("tgl")
	defer tgl.Dispose()

	tgl.SetEnabled(false)
	tgl.Node().GetOnClick()(willow.ClickContext{Node: tgl.Node()})
	if tgl.Value() {
		t.Error("click should not toggle when disabled")
	}
}

func TestToggleOnChange(t *testing.T) {
	resetScheduler()
	tgl := ui.NewToggle("tgl")
	defer tgl.Dispose()

	var got bool
	tgl.SetOnChange(func(v bool) { got = v })

	tgl.Node().GetOnClick()(willow.ClickContext{Node: tgl.Node()})
	if !got {
		t.Error("onChange should fire with true")
	}
}

func TestToggleBindValue(t *testing.T) {
	resetScheduler()
	tgl := ui.NewToggle("tgl")
	defer tgl.Dispose()

	ref := ui.NewRef(true)
	tgl.BindValue(ref)

	if !tgl.Value() {
		t.Error("binding should sync initial value")
	}

	ref.Set(false)
	ui.DefaultScheduler.Flush()
	if tgl.Value() {
		t.Error("reactive update should set value to false")
	}
}

func TestToggleUpdateVisuals(t *testing.T) {
	resetScheduler()
	tgl := ui.NewToggle("tgl")
	defer tgl.Dispose()

	group := ui.DefaultTheme.Toggle.Group(ui.Primary)

	// Default theme has CornerRadius > 0, so colors go to trackPoly/thumbPoly.
	tgl.UpdateVisuals()
	wantOff := group.TrackColor.Resolve(ui.StateDefault)
	if tgl.TrackPoly().Color() != wantOff {
		t.Errorf("off track poly color = %v, want %v", tgl.TrackPoly().Color(), wantOff)
	}

	// On state: track should use StateActive color.
	tgl.SetValue(true)
	tgl.UpdateVisuals()
	wantOn := group.TrackColor.Resolve(ui.StateActive)
	if tgl.TrackPoly().Color() != wantOn {
		t.Errorf("on track poly color = %v, want %v", tgl.TrackPoly().Color(), wantOn)
	}
}

func TestToggleCornerRadius(t *testing.T) {
	resetScheduler()
	tgl := ui.NewToggle("tgl")
	defer tgl.Dispose()

	group := ui.DefaultTheme.Toggle.Group(ui.Primary)

	// Default theme should have CornerRadius == -1 (auto: 50% of height).
	if group.CornerRadius != -1 {
		t.Fatalf("default CornerRadius = %v, want -1", group.CornerRadius)
	}

	tgl.UpdateVisuals()

	// trackPoly should be created and visible.
	if tgl.TrackPoly() == nil {
		t.Fatal("trackPoly should not be nil with CornerRadius > 0")
	}
	if !tgl.TrackPoly().Visible() {
		t.Error("trackPoly should be visible")
	}

	// Flat track sprite should be hidden.
	if tgl.TrackNode().Visible() {
		t.Error("flat track sprite should be hidden when rounded")
	}

	// thumbPoly should be created and visible.
	if tgl.ThumbPoly() == nil {
		t.Fatal("thumbPoly should not be nil with CornerRadius > 0")
	}
	if !tgl.ThumbPoly().Visible() {
		t.Error("thumbPoly should be visible")
	}

	// Flat thumb sprite should be hidden.
	if tgl.ThumbNode().Visible() {
		t.Error("flat thumb sprite should be hidden when rounded")
	}
}

func TestToggleCornerRadiusZero(t *testing.T) {
	resetScheduler()
	tgl := ui.NewToggle("tgl")
	defer tgl.Dispose()

	// Use a custom theme with zero corner radius.
	theme := *ui.DefaultTheme
	theme.Toggle.Primary.CornerRadius = 0
	tgl.SetTheme(&theme)
	tgl.UpdateVisuals()

	// Flat sprites should be visible.
	if !tgl.TrackNode().Visible() {
		t.Error("flat track sprite should be visible with CornerRadius 0")
	}
	if !tgl.ThumbNode().Visible() {
		t.Error("flat thumb sprite should be visible with CornerRadius 0")
	}

	// Polys should be hidden (may exist from initial UpdateVisuals with default theme).
	if tgl.TrackPoly() != nil && tgl.TrackPoly().Visible() {
		t.Error("trackPoly should be hidden with CornerRadius 0")
	}
	if tgl.ThumbPoly() != nil && tgl.ThumbPoly().Visible() {
		t.Error("thumbPoly should be hidden with CornerRadius 0")
	}
}

// ---------------------------------------------------------------------------
// Checkbox
// ---------------------------------------------------------------------------

func TestNewCheckboxDefaults(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cb := ui.NewCheckbox("cb", "Check me", font, 0)
	defer cb.Dispose()

	if cb.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if cb.Checked() {
		t.Error("initial value should be false")
	}
	if cb.BoxNode() == nil {
		t.Fatal("box should not be nil")
	}
	if cb.CheckNode() == nil {
		t.Fatal("check should not be nil")
	}
	if cb.CheckboxLabel() == nil {
		t.Fatal("label should not be nil")
	}
}

func TestCheckboxSetChecked(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cb := ui.NewCheckbox("cb", "Test", font, 0)
	defer cb.Dispose()

	cb.SetChecked(true)
	if !cb.Checked() {
		t.Error("Checked() should be true")
	}

	cb.SetChecked(false)
	if cb.Checked() {
		t.Error("Checked() should be false")
	}
}

func TestCheckboxClickToggles(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cb := ui.NewCheckbox("cb", "Test", font, 0)
	defer cb.Dispose()

	cb.Node().GetOnClick()(willow.ClickContext{Node: cb.Node()})
	if !cb.Checked() {
		t.Error("click should check")
	}

	cb.Node().GetOnClick()(willow.ClickContext{Node: cb.Node()})
	if cb.Checked() {
		t.Error("second click should uncheck")
	}
}

func TestCheckboxDisabledBlocksClick(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cb := ui.NewCheckbox("cb", "Test", font, 0)
	defer cb.Dispose()

	cb.SetEnabled(false)
	cb.Node().GetOnClick()(willow.ClickContext{Node: cb.Node()})
	if cb.Checked() {
		t.Error("click should not check when disabled")
	}
}

func TestCheckboxOnChange(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cb := ui.NewCheckbox("cb", "Test", font, 0)
	defer cb.Dispose()

	var got bool
	cb.SetOnChange(func(v bool) { got = v })
	cb.Node().GetOnClick()(willow.ClickContext{Node: cb.Node()})
	if !got {
		t.Error("onChange should fire with true")
	}
}

func TestCheckboxBindValue(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cb := ui.NewCheckbox("cb", "Test", font, 0)
	defer cb.Dispose()

	ref := ui.NewRef(true)
	cb.BindValue(ref)

	if !cb.Checked() {
		t.Error("binding should sync initial value")
	}

	ref.Set(false)
	ui.DefaultScheduler.Flush()
	if cb.Checked() {
		t.Error("reactive update should uncheck")
	}
}

// ---------------------------------------------------------------------------
// Radio / RadioButton
// ---------------------------------------------------------------------------

func TestRadioAddOption(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	rg := ui.NewRadio("rg")
	defer rg.Dispose()

	rg.AddOption("Option A", font, 0)
	rg.AddOption("Option B", font, 0)
	rg.AddOption("Option C", font, 0)

	if len(rg.Buttons()) != 3 {
		t.Fatalf("expected 3 buttons, got %d", len(rg.Buttons()))
	}
	if rg.Selected() != -1 {
		t.Errorf("initial selection should be -1, got %d", rg.Selected())
	}
}

func TestRadioClickSelects(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	rg := ui.NewRadio("rg")
	defer rg.Dispose()

	rg.AddOption("A", font, 0)
	rg.AddOption("B", font, 0)

	// Click option B (index 1).
	rg.Buttons()[1].Node().GetOnClick()(willow.ClickContext{Node: rg.Buttons()[1].Node()})
	if rg.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1", rg.Selected())
	}

	// Click option A (index 0) -- deselects B.
	rg.Buttons()[0].Node().GetOnClick()(willow.ClickContext{Node: rg.Buttons()[0].Node()})
	if rg.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", rg.Selected())
	}
}

func TestRadioOnChange(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	rg := ui.NewRadio("rg")
	defer rg.Dispose()

	rg.AddOption("A", font, 0)
	rg.AddOption("B", font, 0)

	var got int
	rg.SetOnChange(func(idx int) { got = idx })

	rg.Buttons()[1].Node().GetOnClick()(willow.ClickContext{Node: rg.Buttons()[1].Node()})
	if got != 1 {
		t.Errorf("onChange got %d, want 1", got)
	}
}

func TestRadioBindSelected(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	rg := ui.NewRadio("rg")
	defer rg.Dispose()

	rg.AddOption("A", font, 0)
	rg.AddOption("B", font, 0)

	ref := ui.NewRef(1)
	rg.BindSelected(ref)

	if rg.Selected() != 1 {
		t.Errorf("binding should sync initial value, got %d", rg.Selected())
	}

	ref.Set(0)
	ui.DefaultScheduler.Flush()
	if rg.Selected() != 0 {
		t.Errorf("reactive update should select 0, got %d", rg.Selected())
	}
}

func TestRadioSetSelected(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	rg := ui.NewRadio("rg")
	defer rg.Dispose()

	rg.AddOption("A", font, 0)
	rg.AddOption("B", font, 0)

	rg.SetSelected(1)
	if rg.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1", rg.Selected())
	}

	rg.SetSelected(0)
	if rg.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", rg.Selected())
	}
}

// ---------------------------------------------------------------------------
// TextInput
// ---------------------------------------------------------------------------

func TestNewTextInputDefaults(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	if ti.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if ti.Value() != "" {
		t.Errorf("initial value should be empty, got %q", ti.Value())
	}
	if ti.BgNode() == nil {
		t.Fatal("bgNode should not be nil")
	}
	if ti.TextNode() == nil {
		t.Fatal("textNode should not be nil")
	}
	if ti.CursorNode() == nil {
		t.Fatal("cursor should not be nil")
	}
}

func TestTextInputSetValue(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello")
	if ti.Value() != "hello" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "hello")
	}
}

func TestTextInputMaxLength(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetMaxLength(5)
	ti.SetValue("hello world")
	if ti.Value() != "hello" {
		t.Errorf("Value() = %q, want %q (truncated to maxLength)", ti.Value(), "hello")
	}
}

func TestTextInputPlaceholder(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetPlaceholder("Enter text...")
	if ti.GetPlaceholder() != "Enter text..." {
		t.Errorf("placeholder = %q, want %q", ti.GetPlaceholder(), "Enter text...")
	}
}

func TestTextInputBindValue(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ref := ui.NewRef("initial")
	ti.BindValue(ref)

	if ti.Value() != "initial" {
		t.Errorf("binding should sync initial value, got %q", ti.Value())
	}

	ref.Set("updated")
	ui.DefaultScheduler.Flush()
	if ti.Value() != "updated" {
		t.Errorf("reactive update should set value, got %q", ti.Value())
	}
}

func TestTextInputInsertText(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.InsertText("abc")
	if ti.Value() != "abc" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "abc")
	}
	if ti.GetCursorPos() != 3 {
		t.Errorf("cursorPos = %d, want 3", ti.GetCursorPos())
	}

	// Insert in middle.
	ti.SetCursorPos(1)
	ti.InsertText("X")
	if ti.Value() != "aXbc" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "aXbc")
	}
}

func TestTextInputDeleteBack(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("abc")
	ti.SetCursorPos(3)

	ti.DeleteBack()
	if ti.Value() != "ab" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "ab")
	}
	if ti.GetCursorPos() != 2 {
		t.Errorf("cursorPos = %d, want 2", ti.GetCursorPos())
	}
}

func TestTextInputDeleteForward(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("abc")
	ti.SetCursorPos(1)

	ti.DeleteForward()
	if ti.Value() != "ac" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "ac")
	}
}

func TestTextInputOnChange(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	var got string
	ti.SetOnChange(func(v string) { got = v })

	ti.InsertText("hi")
	if got != "hi" {
		t.Errorf("onChange got %q, want %q", got, "hi")
	}
}

func TestTextInputOnSubmit(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	var got string
	ti.SetOnSubmit(func(v string) { got = v })

	ti.SetValue("test")
	ti.Submit()
	if got != "test" {
		t.Errorf("onSubmit got %q, want %q", got, "test")
	}
}

func TestTextInputSetSize(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetSize(200, 30)
	if ti.Width != 200 || ti.Height != 30 {
		t.Errorf("size = %fx%f, want 200x30", ti.Width, ti.Height)
	}
}

// ---------------------------------------------------------------------------
// TextArea
// ---------------------------------------------------------------------------

func TestNewTextAreaDefaults(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	if ta.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if ta.Value() != "" {
		t.Errorf("initial value should be empty, got %q", ta.Value())
	}
}

func TestTextAreaSetValue(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("line1\nline2\nline3")
	if ta.Value() != "line1\nline2\nline3" {
		t.Errorf("Value() = %q, want multiline", ta.Value())
	}
}

func TestTextAreaSetRows(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetRows(5)
	// Height should be 5 * lineHeight (16 for test font) + some padding.
	expectedMinH := 5 * 16.0
	if ta.Height < expectedMinH {
		t.Errorf("Height = %f, should be at least %f for 5 rows", ta.Height, expectedMinH)
	}
}

func TestTextAreaInsertNewline(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.InsertText("hello")
	ta.InsertText("\n")
	ta.InsertText("world")
	if ta.Value() != "hello\nworld" {
		t.Errorf("Value() = %q, want %q", ta.Value(), "hello\nworld")
	}
}

func TestTextAreaBindValue(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ref := ui.NewRef("initial")
	ta.BindValue(ref)

	if ta.Value() != "initial" {
		t.Errorf("binding should sync initial value, got %q", ta.Value())
	}

	ref.Set("updated")
	ui.DefaultScheduler.Flush()
	if ta.Value() != "updated" {
		t.Errorf("reactive update should set value, got %q", ta.Value())
	}
}

func TestTextAreaOnChange(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	var got string
	ta.SetOnChange(func(v string) { got = v })

	ta.InsertText("test")
	if got != "test" {
		t.Errorf("onChange got %q, want %q", got, "test")
	}
}

func TestTextAreaDeleteBack(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("ab\ncd")
	ta.SetCursorPos(5) // end

	ta.DeleteBack()
	if ta.Value() != "ab\nc" {
		t.Errorf("Value() = %q, want %q", ta.Value(), "ab\nc")
	}
}

// ---------------------------------------------------------------------------
// TextInput — Selection
// ---------------------------------------------------------------------------

func TestTextInputHasSelectionDefault(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	if ti.HasSelection() {
		t.Error("new TextInput should have no selection")
	}
}

func TestTextInputSelectAll(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello")
	ti.SelectAll()
	if !ti.HasSelection() {
		t.Fatal("SelectAll should create a selection")
	}
	if ti.SelectedText() != "hello" {
		t.Errorf("SelectedText() = %q, want %q", ti.SelectedText(), "hello")
	}
	if ti.GetSelStart() != 0 || ti.GetSelEnd() != 5 {
		t.Errorf("selStart=%d selEnd=%d, want 0,5", ti.GetSelStart(), ti.GetSelEnd())
	}
}

func TestTextInputInsertReplacesSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello")
	ti.SetSelStart(1)
	ti.SetSelEnd(4)
	ti.SetCursorPos(4)

	ti.InsertText("X")
	if ti.Value() != "hXo" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "hXo")
	}
	if ti.HasSelection() {
		t.Error("selection should be cleared after insert")
	}
}

func TestTextInputDeleteBackWithSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("abcdef")
	ti.SetSelStart(1)
	ti.SetSelEnd(4)
	ti.SetCursorPos(4)

	ti.DeleteBack()
	if ti.Value() != "aef" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "aef")
	}
	if ti.GetCursorPos() != 1 {
		t.Errorf("cursorPos = %d, want 1", ti.GetCursorPos())
	}
}

func TestTextInputDeleteForwardWithSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("abcdef")
	ti.SetSelStart(2)
	ti.SetSelEnd(5)
	ti.SetCursorPos(5)

	ti.DeleteForward()
	if ti.Value() != "abf" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "abf")
	}
	if ti.GetCursorPos() != 2 {
		t.Errorf("cursorPos = %d, want 2", ti.GetCursorPos())
	}
}

func TestTextInputMoveCursorLeftCollapsesSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello")
	ti.SetSelStart(1)
	ti.SetSelEnd(4)
	ti.SetCursorPos(4)

	ti.MoveCursorLeft() // should collapse to left edge (1)
	if ti.HasSelection() {
		t.Error("selection should be cleared")
	}
	if ti.GetCursorPos() != 1 {
		t.Errorf("cursorPos = %d, want 1", ti.GetCursorPos())
	}
}

func TestTextInputMoveCursorRightCollapsesSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello")
	ti.SetSelStart(1)
	ti.SetSelEnd(4)
	ti.SetCursorPos(4)

	ti.MoveCursorRight() // should collapse to right edge (4)
	if ti.HasSelection() {
		t.Error("selection should be cleared")
	}
	if ti.GetCursorPos() != 4 {
		t.Errorf("cursorPos = %d, want 4", ti.GetCursorPos())
	}
}

func TestTextInputShiftArrowExtendsSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello")
	ti.SetCursorPos(2)
	ti.ClearSelectionForTest()

	// Shift+Right to extend selection.
	ti.MoveCursorRightShiftForTest(true)
	if !ti.HasSelection() {
		t.Fatal("shift+right should create selection")
	}
	if ti.GetSelStart() != 2 || ti.GetSelEnd() != 3 {
		t.Errorf("sel = [%d,%d], want [2,3]", ti.GetSelStart(), ti.GetSelEnd())
	}

	// Shift+Right again.
	ti.MoveCursorRightShiftForTest(true)
	if ti.GetSelStart() != 2 || ti.GetSelEnd() != 4 {
		t.Errorf("sel = [%d,%d], want [2,4]", ti.GetSelStart(), ti.GetSelEnd())
	}

	// Shift+Left shrinks.
	ti.MoveCursorLeftShiftForTest(true)
	if ti.GetSelStart() != 2 || ti.GetSelEnd() != 3 {
		t.Errorf("sel = [%d,%d], want [2,3]", ti.GetSelStart(), ti.GetSelEnd())
	}
}

func TestTextInputSetValueClearsSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello")
	ti.SelectAll()
	ti.SetValue("world")
	if ti.HasSelection() {
		t.Error("SetValue should clear selection")
	}
}

func TestTextInputSelectionRect(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	if ti.SelRectNode() == nil {
		t.Fatal("selRect should not be nil")
	}

	// No selection — rect hidden.
	ti.SetValue("hello")
	ti.UpdateSelectionRectForTest()
	if ti.SelRectVisible() {
		t.Error("selRect should be hidden when no selection")
	}

	// Select chars 1..3 — rect shown.
	ti.SetSelStart(1)
	ti.SetSelEnd(3)
	ti.SetCursorPos(3)
	ti.UpdateSelectionRectForTest()
	if !ti.SelRectVisible() {
		t.Error("selRect should be visible when selection exists")
	}
}

func TestTextInputDragToSelect(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	// Use short text "abc" so it fits within the default TextInput width.
	ti.SetValue("abc")
	pad := ui.DefaultTheme.TextInput.Group(ui.Primary).Padding.Left

	// Simulate OnPointerDown at the start (pos 0).
	if !ti.Node().HasOnPointerDown() {
		t.Fatal("OnPointerDown should be wired")
	}
	ti.Node().GetOnPointerDown()(willow.PointerContext{
		Node:   ti.Node(),
		LocalX: pad + 1, // near start
	})

	if ti.HasSelection() {
		t.Error("pointer down should not create selection")
	}
	anchor := ti.GetCursorPos()

	// Simulate OnDrag extending rightward.
	if !ti.Node().HasOnDrag() {
		t.Fatal("OnDrag should be wired")
	}
	// Drag to just past the full text width to select everything.
	fullW, _ := font.MeasureString("abc", 0, false, false)
	ti.Node().GetOnDrag()(willow.DragContext{
		Node:   ti.Node(),
		LocalX: pad + fullW + 1,
	})

	if !ti.HasSelection() {
		t.Fatal("drag should create selection")
	}
	if ti.GetSelStart() != anchor {
		t.Errorf("selStart = %d, want %d (anchor)", ti.GetSelStart(), anchor)
	}
	if ti.GetSelEnd() <= anchor {
		t.Errorf("selEnd = %d should be > anchor %d", ti.GetSelEnd(), anchor)
	}
	if ti.SelectedText() == "" {
		t.Error("SelectedText should not be empty during drag")
	}
}

// ---------------------------------------------------------------------------
// TextArea — Selection
// ---------------------------------------------------------------------------

func TestTextAreaHasSelectionDefault(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	if ta.HasSelection() {
		t.Error("new TextArea should have no selection")
	}
}

func TestTextAreaSelectAll(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("hello\nworld")
	ta.SelectAll()
	if !ta.HasSelection() {
		t.Fatal("SelectAll should create a selection")
	}
	if ta.SelectedText() != "hello\nworld" {
		t.Errorf("SelectedText() = %q, want %q", ta.SelectedText(), "hello\nworld")
	}
}

func TestTextAreaInsertReplacesSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("abcdef")
	ta.SetSelStart(1)
	ta.SetSelEnd(4)
	ta.SetCursorPos(4)

	ta.InsertText("X")
	if ta.Value() != "aXef" {
		t.Errorf("Value() = %q, want %q", ta.Value(), "aXef")
	}
}

func TestTextAreaDeleteBackWithSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("ab\ncd")
	ta.SetSelStart(1)
	ta.SetSelEnd(4)
	ta.SetCursorPos(4)

	ta.DeleteBack()
	if ta.Value() != "ad" {
		t.Errorf("Value() = %q, want %q", ta.Value(), "ad")
	}
	if ta.GetCursorPos() != 1 {
		t.Errorf("cursorPos = %d, want 1", ta.GetCursorPos())
	}
}

func TestTextAreaShiftArrowExtendsSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("hello\nworld")
	ta.SetCursorPos(2)
	ta.ClearSelectionForTest()

	ta.MoveCursorRightShiftForTest(true)
	if !ta.HasSelection() {
		t.Fatal("shift+right should create selection")
	}
	if ta.GetSelStart() != 2 || ta.GetSelEnd() != 3 {
		t.Errorf("sel = [%d,%d], want [2,3]", ta.GetSelStart(), ta.GetSelEnd())
	}
}

func TestTextAreaShiftUpDownExtendsSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("hello\nworld")
	ta.SetCursorPos(6) // "w" on line 2, col 0
	ta.ClearSelectionForTest()

	ta.MoveCursorUpShiftForTest(true)
	if !ta.HasSelection() {
		t.Fatal("shift+up should create selection")
	}
	// Should move to line 1, col 0 (pos 0).
	if ta.GetCursorPos() != 0 {
		t.Errorf("cursorPos = %d, want 0", ta.GetCursorPos())
	}
	if ta.GetSelEnd() != 0 {
		t.Errorf("selEnd = %d, want 0", ta.GetSelEnd())
	}
}

func TestTextAreaMoveCursorLeftCollapsesSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("hello")
	ta.SetSelStart(1)
	ta.SetSelEnd(4)
	ta.SetCursorPos(4)

	ta.MoveCursorLeft()
	if ta.HasSelection() {
		t.Error("selection should be cleared")
	}
	if ta.GetCursorPos() != 1 {
		t.Errorf("cursorPos = %d, want 1", ta.GetCursorPos())
	}
}

func TestTextAreaSetValueClearsSelection(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("hello")
	ta.SelectAll()
	ta.SetValue("world")
	if ta.HasSelection() {
		t.Error("SetValue should clear selection")
	}
}

// ---------------------------------------------------------------------------
// TextInput — Clipboard
// ---------------------------------------------------------------------------

func TestTextInputCopyToClipboard(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello world")
	ti.SelectAll()
	sel := ti.SelectedText()
	clipboard.WriteAll(sel)

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("clipboard.ReadAll: %v", err)
	}
	if got != "hello world" {
		t.Errorf("clipboard = %q, want %q", got, "hello world")
	}
	// Value unchanged after copy.
	if ti.Value() != "hello world" {
		t.Errorf("Value = %q, want %q", ti.Value(), "hello world")
	}
}

func TestTextInputCutToClipboard(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello world")
	ti.SelectAll()
	sel := ti.SelectedText()
	clipboard.WriteAll(sel)
	ti.DeleteSelectionForTest()

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("clipboard.ReadAll: %v", err)
	}
	if got != "hello world" {
		t.Errorf("clipboard = %q, want %q", got, "hello world")
	}
	if ti.Value() != "" {
		t.Errorf("Value = %q, want empty after cut", ti.Value())
	}
}

func TestTextInputPasteFromClipboard(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello ")
	clipboard.WriteAll("world")
	text, _ := clipboard.ReadAll()
	ti.InsertText(text)

	if ti.Value() != "hello world" {
		t.Errorf("Value = %q, want %q", ti.Value(), "hello world")
	}
}

func TestTextInputPasteReplacesSelection(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello world")
	ti.SelectAll()
	clipboard.WriteAll("goodbye")
	text, _ := clipboard.ReadAll()
	ti.InsertText(text)

	if ti.Value() != "goodbye" {
		t.Errorf("Value = %q, want %q", ti.Value(), "goodbye")
	}
}

func TestTextInputCopyNoSelectionIsNoop(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	clipboard.WriteAll("original")
	ti.SetValue("hello")
	// No selection — copy should not change clipboard.
	sel := ti.SelectedText()
	if sel != "" {
		t.Fatal("expected no selection")
	}

	got, _ := clipboard.ReadAll()
	if got != "original" {
		t.Errorf("clipboard = %q, want %q (should be unchanged)", got, "original")
	}
}

// ---------------------------------------------------------------------------
// TextArea — Clipboard
// ---------------------------------------------------------------------------

func TestTextAreaCopyToClipboard(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("line one\nline two")
	ta.SelectAll()
	sel := ta.SelectedText()
	clipboard.WriteAll(sel)

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("clipboard.ReadAll: %v", err)
	}
	if got != "line one\nline two" {
		t.Errorf("clipboard = %q, want %q", got, "line one\nline two")
	}
}

func TestTextAreaCutToClipboard(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("line one\nline two")
	ta.SelectAll()
	sel := ta.SelectedText()
	clipboard.WriteAll(sel)
	ta.DeleteSelectionForTest()

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("clipboard.ReadAll: %v", err)
	}
	if got != "line one\nline two" {
		t.Errorf("clipboard = %q, want %q", got, "line one\nline two")
	}
	if ta.Value() != "" {
		t.Errorf("Value = %q, want empty after cut", ta.Value())
	}
}

func TestTextAreaPasteFromClipboard(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("hello ")
	clipboard.WriteAll("world")
	text, _ := clipboard.ReadAll()
	ta.InsertText(text)

	if ta.Value() != "hello world" {
		t.Errorf("Value = %q, want %q", ta.Value(), "hello world")
	}
}

func TestTextAreaPasteReplacesSelection(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetValue("old text")
	ta.SelectAll()
	clipboard.WriteAll("new text")
	text, _ := clipboard.ReadAll()
	ta.InsertText(text)

	if ta.Value() != "new text" {
		t.Errorf("Value = %q, want %q", ta.Value(), "new text")
	}
}

// ---------------------------------------------------------------------------
// TextInput — Horizontal Scrolling
// ---------------------------------------------------------------------------

func TestTextInputScrollXOnLongText(t *testing.T) {
	resetScheduler()
	font := newTestFont() // 8px per char
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	// Default width 200, padding each side → inner width.
	tiPad := ui.DefaultTheme.TextInput.Group(ui.Primary).Padding
	innerW := ti.Width - tiPad.Left - tiPad.Right

	// Type a string longer than the visible area.
	longText := "abcdefghijklmnopqrstuvwxyz0123" // 30 chars = 240px
	ti.SetValue(longText)

	// After SetValue, cursor is at end. scrollX should have adjusted
	// so the cursor is visible.
	textW, _ := font.MeasureString(longText, 0, false, false)
	if ti.GetScrollX() <= 0 {
		t.Errorf("scrollX = %f, want > 0 for overflowing text", ti.GetScrollX())
	}
	cursorVisible := textW - ti.GetScrollX()
	if cursorVisible < 0 || cursorVisible > innerW {
		t.Errorf("cursor at %f, scrollX=%f, visible offset=%f not in [0, %f]",
			textW, ti.GetScrollX(), cursorVisible, innerW)
	}
}

func TestTextInputScrollXResetOnShortText(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	// Short text that fits — scrollX should be 0.
	ti.SetValue("hi")
	if ti.GetScrollX() != 0 {
		t.Errorf("scrollX = %f, want 0 for short text", ti.GetScrollX())
	}
}

func TestTextInputScrollXOnCursorLeft(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	// Fill with long text so we have a nonzero scrollX.
	longText := "abcdefghijklmnopqrstuvwxyz0123"
	ti.SetValue(longText)
	initialScroll := ti.GetScrollX()

	// Move cursor to start — scrollX should reach 0.
	for ti.GetCursorPos() > 0 {
		ti.MoveCursorLeftShiftForTest(false)
	}
	if ti.GetScrollX() != 0 {
		t.Errorf("scrollX = %f, want 0 after moving cursor to start (was %f)",
			ti.GetScrollX(), initialScroll)
	}
}

// ---------------------------------------------------------------------------
// TextArea — Word Wrapping (visual lines)
// ---------------------------------------------------------------------------

func TestTextAreaGetVisualLinesNoWrap(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	// Set a wide width so no wrapping occurs.
	ta.SetSize(500, 100)
	ta.SetValue("hello\nworld")

	vlines := ta.GetVisualLinesForTest()
	if len(vlines) != 2 {
		t.Fatalf("got %d visual lines, want 2", len(vlines))
	}
	runes := []rune(ta.Value())
	if string(runes[vlines[0].RuneStart:vlines[0].RuneEnd]) != "hello" {
		t.Errorf("line 0 = %q, want %q", string(runes[vlines[0].RuneStart:vlines[0].RuneEnd]), "hello")
	}
	if string(runes[vlines[1].RuneStart:vlines[1].RuneEnd]) != "world" {
		t.Errorf("line 1 = %q, want %q", string(runes[vlines[1].RuneStart:vlines[1].RuneEnd]), "world")
	}
}

func TestTextAreaGetVisualLinesWordWrap(t *testing.T) {
	resetScheduler()
	font := newTestFont() // 8px per char
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	// Inner width for wrapping: width - left pad - right pad.
	// Set width so inner width = 80px → 10 chars per line.
	taPad := ui.DefaultTheme.TextArea.Group(ui.Primary).Padding
	ta.SetSize(80+taPad.Left+taPad.Right, 100)

	// "hello world" is 11 chars (88px) > 80px wrap width.
	// Should wrap at word boundary: "hello" + "world".
	ta.SetValue("hello world")

	vlines := ta.GetVisualLinesForTest()
	if len(vlines) != 2 {
		t.Fatalf("got %d visual lines, want 2", len(vlines))
	}
	runes := []rune(ta.Value())
	line0 := string(runes[vlines[0].RuneStart:vlines[0].RuneEnd])
	line1 := string(runes[vlines[1].RuneStart:vlines[1].RuneEnd])
	if line0 != "hello" {
		t.Errorf("line 0 = %q, want %q", line0, "hello")
	}
	if line1 != "world" {
		t.Errorf("line 1 = %q, want %q", line1, "world")
	}
}

func TestTextAreaCursorVisualLineCol(t *testing.T) {
	resetScheduler()
	font := newTestFont() // 8px per char
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	taPad := ui.DefaultTheme.TextArea.Group(ui.Primary).Padding
	ta.SetSize(80+taPad.Left+taPad.Right, 100) // 10 chars per line

	ta.SetValue("hello world")

	// Cursor at position 0 → line 0, col 0.
	line, col := ta.CursorVisualLineColForTest(0)
	if line != 0 || col != 0 {
		t.Errorf("pos 0: line=%d col=%d, want 0,0", line, col)
	}

	// Cursor at position 5 → "hello" ends at runeEnd=5 for line 0,
	// and line 1 starts at runeStart=6 (after space). Position 5 is
	// at line 0 end.
	line, col = ta.CursorVisualLineColForTest(5)
	if line != 0 || col != 5 {
		t.Errorf("pos 5: line=%d col=%d, want 0,5", line, col)
	}

	// Cursor at position 6 → start of "world" on line 1.
	line, col = ta.CursorVisualLineColForTest(6)
	if line != 1 || col != 0 {
		t.Errorf("pos 6: line=%d col=%d, want 1,0", line, col)
	}
}

func TestTextAreaVerticalAutoScroll(t *testing.T) {
	resetScheduler()
	font := newTestFont() // lineHeight = 16
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	// 2 visible rows.
	ta.SetRows(2)
	// Insert enough lines to overflow.
	ta.SetValue("line1\nline2\nline3\nline4")

	// After SetValue, cursor is at end (line4). scrollY should have
	// adjusted to keep the cursor visible.
	if ta.GetScrollY() <= 0 {
		t.Errorf("scrollY = %f, want > 0 when cursor is past visible rows", ta.GetScrollY())
	}
}

func TestTextAreaScrollbarVisibility(t *testing.T) {
	resetScheduler()
	font := newTestFont() // lineHeight = 16
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetRows(2) // 32px visible

	// Short text — scrollbar hidden.
	ta.SetValue("hello")
	if ta.ScrollBar().IsVisible() {
		t.Error("scrollbar should be hidden when content fits")
	}

	// Enough lines to overflow — scrollbar visible.
	ta.SetValue("line1\nline2\nline3\nline4")
	if !ta.ScrollBar().IsVisible() {
		t.Error("scrollbar should be visible when content overflows")
	}

	// Back to short text — scrollbar hidden again.
	ta.SetValue("short")
	if ta.ScrollBar().IsVisible() {
		t.Error("scrollbar should hide when content fits again")
	}
}

// ---------------------------------------------------------------------------
// TextInput — Home / End navigation
// ---------------------------------------------------------------------------

func TestTextInputHome(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello world")
	// Cursor is at end after SetValue.
	ti.MoveCursorHomeShiftForTest(false)
	if ti.GetCursorPos() != 0 {
		t.Errorf("cursorPos = %d, want 0 after Home", ti.GetCursorPos())
	}
	if ti.HasSelection() {
		t.Error("Home without shift should clear selection")
	}
}

func TestTextInputEnd(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello world")
	ti.SetCursorPos(0)
	ti.ClearSelectionForTest()
	ti.MoveCursorEndShiftForTest(false)
	if ti.GetCursorPos() != 11 {
		t.Errorf("cursorPos = %d, want 11 after End", ti.GetCursorPos())
	}
}

func TestTextInputHomeShiftSelects(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello")
	// Cursor at end (5). Shift+Home should select all.
	ti.MoveCursorHomeShiftForTest(true)
	if ti.GetCursorPos() != 0 {
		t.Errorf("cursorPos = %d, want 0", ti.GetCursorPos())
	}
	if !ti.HasSelection() {
		t.Error("Shift+Home should create selection")
	}
	if ti.SelectedText() != "hello" {
		t.Errorf("selected = %q, want %q", ti.SelectedText(), "hello")
	}
}

func TestTextInputEndShiftSelects(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello")
	ti.SetCursorPos(0)
	ti.SetSelStart(0)
	ti.SetSelEnd(0)
	ti.MoveCursorEndShiftForTest(true)
	if ti.GetCursorPos() != 5 {
		t.Errorf("cursorPos = %d, want 5", ti.GetCursorPos())
	}
	if ti.SelectedText() != "hello" {
		t.Errorf("selected = %q, want %q", ti.SelectedText(), "hello")
	}
}

// ---------------------------------------------------------------------------
// TextArea — Home / End / PageUp / PageDown navigation
// ---------------------------------------------------------------------------

func TestTextAreaHome(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetSize(500, 100) // wide enough for no wrapping
	ta.SetValue("hello\nworld")
	// Cursor is at end of "world" (pos 11).
	ta.MoveCursorHomeShiftForTest(false)
	// Should move to start of "world" line (pos 6).
	if ta.GetCursorPos() != 6 {
		t.Errorf("cursorPos = %d, want 6 after Home on line 2", ta.GetCursorPos())
	}
}

func TestTextAreaEnd(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetSize(500, 100)
	ta.SetValue("hello\nworld")
	// Move to start of line 1.
	ta.SetCursorPos(6)
	ta.ClearSelectionForTest()
	ta.MoveCursorEndShiftForTest(false)
	if ta.GetCursorPos() != 11 {
		t.Errorf("cursorPos = %d, want 11 after End on line 2", ta.GetCursorPos())
	}
}

func TestTextAreaPageDown(t *testing.T) {
	resetScheduler()
	font := newTestFont() // lineHeight = 16
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetRows(2)
	ta.SetSize(500, 100)
	ta.SetValue("line0\nline1\nline2\nline3\nline4")

	// Start at line 0, pos 0.
	ta.SetCursorPos(0)
	ta.ClearSelectionForTest()
	ta.MoveCursorPageDownShiftForTest(false)

	// rows=2, so PageDown should jump 2 lines → line 2, pos = start of "line2".
	line, _ := ta.CursorVisualLineColForTest(ta.GetCursorPos())
	if line != 2 {
		t.Errorf("after PageDown: visual line = %d, want 2", line)
	}
}

func TestTextAreaPageUp(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetRows(2)
	ta.SetSize(500, 100)
	ta.SetValue("line0\nline1\nline2\nline3\nline4")

	// Move to line 4 (end of text).
	runes := []rune(ta.Value())
	ta.SetCursorPos(len(runes))
	ta.ClearSelectionForTest()
	ta.MoveCursorPageUpShiftForTest(false)

	// From line 4, PageUp with rows=2 should jump to line 2.
	line, _ := ta.CursorVisualLineColForTest(ta.GetCursorPos())
	if line != 2 {
		t.Errorf("after PageUp: visual line = %d, want 2", line)
	}
}

func TestTextAreaPageDownShiftSelects(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetRows(2)
	ta.SetSize(500, 100)
	ta.SetValue("line0\nline1\nline2\nline3")

	ta.SetCursorPos(0)
	ta.SetSelStart(0)
	ta.SetSelEnd(0)
	ta.MoveCursorPageDownShiftForTest(true)

	if !ta.HasSelection() {
		t.Error("Shift+PageDown should create selection")
	}
	if ta.SelectedText() == "" {
		t.Error("expected non-empty selection after Shift+PageDown")
	}
}

// ---------------------------------------------------------------------------
// Double-click to select word
// ---------------------------------------------------------------------------

func TestWordBoundaries(t *testing.T) {
	tests := []struct {
		text   string
		pos    int
		wantLo int
		wantHi int
		desc   string
	}{
		{"hello world", 0, 0, 5, "start of first word"},
		{"hello world", 3, 0, 5, "middle of first word"},
		{"hello world", 5, 5, 6, "on space between words"},
		{"hello world", 6, 6, 11, "start of second word"},
		{"hello world", 11, 6, 11, "end of text"},
		{"hello", 2, 0, 5, "single word middle"},
		{"  spaces  ", 1, 0, 2, "leading spaces"},
		{"abc123", 3, 0, 6, "mixed letters and digits"},
		{"hello-world", 5, 5, 6, "on hyphen"},
		{"", 0, 0, 0, "empty string"},
	}
	for _, tt := range tests {
		runes := []rune(tt.text)
		lo, hi := widget.WordBoundaries(runes, tt.pos)
		if lo != tt.wantLo || hi != tt.wantHi {
			t.Errorf("%s: WordBoundaries(%q, %d) = (%d, %d), want (%d, %d)",
				tt.desc, tt.text, tt.pos, lo, hi, tt.wantLo, tt.wantHi)
		}
	}
}

func TestTextInputSelectWordAtCursor(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello world")
	ti.SetCursorPos(2) // middle of "hello"
	ti.ClearSelectionForTest()
	ti.SelectWordAtCursorForTest()

	if ti.SelectedText() != "hello" {
		t.Errorf("selected = %q, want %q", ti.SelectedText(), "hello")
	}
	if ti.GetCursorPos() != 5 {
		t.Errorf("cursorPos = %d, want 5 (end of word)", ti.GetCursorPos())
	}
}

func TestTextInputSelectWordOnSpace(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	ti.SetValue("hello world")
	ti.SetCursorPos(5) // on the space
	ti.ClearSelectionForTest()
	ti.SelectWordAtCursorForTest()

	// Space is a non-word char, should select just the space.
	if ti.SelectedText() != " " {
		t.Errorf("selected = %q, want %q", ti.SelectedText(), " ")
	}
}

func TestTextAreaSelectWordAtCursor(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	ta.SetSize(500, 100)
	ta.SetValue("hello world\nfoo bar")
	ta.SetCursorPos(14) // middle of "foo" (positions: 12=f, 13=o, 14=o)
	ta.ClearSelectionForTest()
	ta.SelectWordAtCursorForTest()

	if ta.SelectedText() != "foo" {
		t.Errorf("selected = %q, want %q", ta.SelectedText(), "foo")
	}
}
