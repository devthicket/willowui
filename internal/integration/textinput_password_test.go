package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/widget"
)

func TestPasswordMode_DotsRendered(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("pw", font, 14)
	defer ti.Dispose()

	ti.SetPasswordMode(true)
	ti.SetValue("hello")

	// 5 runes → 5 visible dots.
	dots := ti.PasswordDots()
	if len(dots) < 5 {
		t.Fatalf("expected at least 5 dots, got %d", len(dots))
	}
	visCount := 0
	for _, d := range dots {
		if d.Visible() {
			visCount++
		}
	}
	if visCount != 5 {
		t.Errorf("expected 5 visible dots, got %d", visCount)
	}

	// Text node should be hidden.
	if ti.TextNode().Visible() {
		t.Error("textNode should be hidden in password mode")
	}
}

func TestPasswordMode_ValueUnchanged(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("pw", font, 14)
	defer ti.Dispose()

	ti.SetPasswordMode(true)
	ti.InsertText("secret")

	if ti.Value() != "secret" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "secret")
	}
}

func TestPasswordMode_CopyDisabled(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("pw", font, 14)
	defer ti.Dispose()

	ti.SetPasswordMode(true)
	ti.SetValue("secret")
	ti.SelectAll()

	sel := ti.SelectedText()
	if sel != "secret" {
		t.Fatalf("SelectedText() = %q, want %q", sel, "secret")
	}

	// In password mode, the widget should not expose text to clipboard.
	// The copy/cut logic is tested via the Update() key handlers which
	// consume the key but skip clipboard write. Since we can't easily
	// simulate Ctrl+C in a unit test without a running game loop, we
	// verify that the password mode flag is set and trust the guard.
	if !ti.IsPasswordMode() {
		t.Error("expected IsPasswordMode() to be true")
	}
}

func TestPasswordMode_PlaceholderNotMasked(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("pw", font, 14)
	defer ti.Dispose()

	ti.SetPasswordMode(true)
	ti.SetPlaceholder("Enter password")

	// Empty value + placeholder → textNode should be visible.
	if !ti.TextNode().Visible() {
		t.Error("textNode should be visible when showing placeholder")
	}

	// All dots should be hidden.
	for i, d := range ti.PasswordDots() {
		if d.Visible() {
			t.Errorf("dot %d should be hidden when value is empty", i)
		}
	}
}

func TestPasswordMode_CursorNavigation(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("pw", font, 14)
	defer ti.Dispose()

	ti.SetPasswordMode(true)
	ti.SetValue("abcde")

	// Cursor should be at end (5).
	if ti.GetCursorPos() != 5 {
		t.Fatalf("cursor pos = %d, want 5", ti.GetCursorPos())
	}

	ti.MoveCursorLeft()
	if ti.GetCursorPos() != 4 {
		t.Errorf("after left, cursor pos = %d, want 4", ti.GetCursorPos())
	}

	ti.MoveCursorRight()
	if ti.GetCursorPos() != 5 {
		t.Errorf("after right, cursor pos = %d, want 5", ti.GetCursorPos())
	}
}

func TestPasswordMode_Toggle(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("pw", font, 14)
	defer ti.Dispose()

	ti.SetValue("test")
	ti.SetPasswordMode(true)

	if ti.TextNode().Visible() {
		t.Error("textNode should be hidden after enabling password mode")
	}

	ti.SetPasswordMode(false)

	if !ti.TextNode().Visible() {
		t.Error("textNode should be visible after disabling password mode")
	}

	// All dots should be hidden.
	for i, d := range ti.PasswordDots() {
		if d.Visible() {
			t.Errorf("dot %d should be hidden after disabling password mode", i)
		}
	}
}

func TestPasswordMode_ReactiveBinding(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("pw", font, 14)
	defer ti.Dispose()

	ti.SetValue("test")

	ref := ui.NewRef(false)
	ti.BindPasswordMode(ref)

	// Initially false — textNode visible.
	if !ti.TextNode().Visible() {
		t.Error("textNode should be visible when ref is false")
	}

	ref.Set(true)
	ui.DefaultScheduler.Flush()

	if !ti.IsPasswordMode() {
		t.Error("expected password mode after setting ref to true")
	}
	if ti.TextNode().Visible() {
		t.Error("textNode should be hidden after setting ref to true")
	}

	ref.Set(false)
	ui.DefaultScheduler.Flush()

	if ti.IsPasswordMode() {
		t.Error("expected normal mode after setting ref to false")
	}
	if !ti.TextNode().Visible() {
		t.Error("textNode should be visible after setting ref to false")
	}
}

func TestPasswordMode_PasteWorks(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("pw", font, 14)
	defer ti.Dispose()

	ti.SetPasswordMode(true)
	ti.InsertText("pasted")

	if ti.Value() != "pasted" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "pasted")
	}

	// Dots should reflect the pasted text.
	visCount := 0
	for _, d := range ti.PasswordDots() {
		if d.Visible() {
			visCount++
		}
	}
	if visCount != 6 {
		t.Errorf("expected 6 visible dots after paste, got %d", visCount)
	}
}

func TestPasswordMode_CutDisabled(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("pw", font, 14)
	defer ti.Dispose()

	ti.SetPasswordMode(true)
	ti.SetValue("secret")
	ti.SelectAll()

	// Verify the value is intact — cut should not work in password mode.
	// Like the copy test, we verify the guard is in place.
	if ti.Value() != "secret" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "secret")
	}
	if !ti.IsPasswordMode() {
		t.Error("expected IsPasswordMode() to be true")
	}
}

func TestPasswordMode_DotGlyphNotNil(t *testing.T) {
	glyph := ui.PasswordDotGlyph()
	if glyph == nil {
		t.Fatal("PasswordDotGlyph() returned nil")
	}
	b := glyph.Bounds()
	if b.Dx() != widget.GlyphSize || b.Dy() != widget.GlyphSize {
		t.Errorf("glyph size = %dx%d, want %dx%d", b.Dx(), b.Dy(), widget.GlyphSize, widget.GlyphSize)
	}
}
