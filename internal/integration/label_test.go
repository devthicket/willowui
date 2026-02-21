package integration

import (
	"testing"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

// --- NewLabel ---

func TestNewLabelCreatesWithCorrectText(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	lbl := ui.NewLabel("greeting", "Hello", font, 0)
	defer lbl.Dispose()

	if lbl.Text() != "Hello" {
		t.Errorf("Text() = %q, want %q", lbl.Text(), "Hello")
	}
	if lbl.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if lbl.Name() != "greeting" {
		t.Errorf("Name() = %q, want %q", lbl.Name(), "greeting")
	}
	if lbl.TextNode() == nil {
		t.Fatal("textNode should not be nil")
	}
	if lbl.TextNode().TextBlock == nil {
		t.Fatal("textNode.TextBlock should not be nil")
	}
	if lbl.TextNode().TextBlock.Content != "Hello" {
		t.Errorf("TextBlock.Content = %q, want %q", lbl.TextNode().TextBlock.Content, "Hello")
	}
}

// --- SetText ---

func TestLabelSetTextUpdatesContent(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	lbl := ui.NewLabel("lbl", "before", font, 0)
	defer lbl.Dispose()

	lbl.SetText("after")

	if lbl.Text() != "after" {
		t.Errorf("Text() = %q, want %q", lbl.Text(), "after")
	}
	if lbl.TextNode().TextBlock.Content != "after" {
		t.Errorf("TextBlock.Content = %q, want %q", lbl.TextNode().TextBlock.Content, "after")
	}
}

// --- Text ---

func TestLabelTextReturnsCurrentText(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	lbl := ui.NewLabel("lbl", "initial", font, 0)
	defer lbl.Dispose()

	if lbl.Text() != "initial" {
		t.Errorf("Text() = %q, want %q", lbl.Text(), "initial")
	}

	lbl.SetText("changed")
	if lbl.Text() != "changed" {
		t.Errorf("Text() = %q, want %q", lbl.Text(), "changed")
	}
}

// --- SetColor ---

func TestLabelSetColorChangesTextColor(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	lbl := ui.NewLabel("lbl", "test", font, 0)
	defer lbl.Dispose()

	red := willow.RGBA(1, 0, 0, 1)
	lbl.SetColor(red)

	got := lbl.TextNode().TextBlock.Color
	if got != red {
		t.Errorf("TextBlock.Color = %v, want %v", got, red)
	}
}

// --- SetAlign ---

func TestLabelSetAlignChangesAlignment(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	lbl := ui.NewLabel("lbl", "test", font, 0)
	defer lbl.Dispose()

	lbl.SetAlign(willow.TextAlignCenter)
	if lbl.TextNode().TextBlock.Align != willow.TextAlignCenter {
		t.Errorf("Align = %v, want TextAlignCenter", lbl.TextNode().TextBlock.Align)
	}

	lbl.SetAlign(willow.TextAlignRight)
	if lbl.TextNode().TextBlock.Align != willow.TextAlignRight {
		t.Errorf("Align = %v, want TextAlignRight", lbl.TextNode().TextBlock.Align)
	}
}

// --- SetWrapWidth ---

func TestLabelSetWrapWidthChangesWrapWidth(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	lbl := ui.NewLabel("lbl", "test", font, 0)
	defer lbl.Dispose()

	lbl.SetWrapWidth(200)
	if lbl.TextNode().TextBlock.WrapWidth != 200 {
		t.Errorf("WrapWidth = %f, want 200", lbl.TextNode().TextBlock.WrapWidth)
	}
}

// --- BindText ---

func TestLabelBindTextAutoUpdatesOnRefChange(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	lbl := ui.NewLabel("lbl", "initial", font, 0)
	defer lbl.Dispose()

	ref := ui.NewRef("reactive")
	lbl.BindText(ref)

	if lbl.Text() != "reactive" {
		t.Errorf("Text() = %q, want %q after BindText", lbl.Text(), "reactive")
	}

	ref.Set("updated")
	ui.DefaultScheduler.Flush()

	if lbl.Text() != "updated" {
		t.Errorf("Text() = %q, want %q after ref change + Flush", lbl.Text(), "updated")
	}
}

func TestLabelBindTextCanBeReplacedWithNewBinding(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	lbl := ui.NewLabel("lbl", "initial", font, 0)
	defer lbl.Dispose()

	ref1 := ui.NewRef("ref1")
	lbl.BindText(ref1)

	if lbl.Text() != "ref1" {
		t.Fatalf("Text() = %q, want %q", lbl.Text(), "ref1")
	}

	ref2 := ui.NewRef("ref2")
	lbl.BindText(ref2)

	if lbl.Text() != "ref2" {
		t.Errorf("Text() = %q, want %q after rebind", lbl.Text(), "ref2")
	}

	// Changing the old ref should have no effect.
	ref1.Set("ref1-changed")
	ui.DefaultScheduler.Flush()

	if lbl.Text() != "ref2" {
		t.Errorf("Text() = %q, want %q (old ref change should be ignored)", lbl.Text(), "ref2")
	}

	// Changing the new ref should work.
	ref2.Set("ref2-changed")
	ui.DefaultScheduler.Flush()

	if lbl.Text() != "ref2-changed" {
		t.Errorf("Text() = %q, want %q", lbl.Text(), "ref2-changed")
	}
}

// --- Dispose ---

func TestLabelDisposeStopsReactiveWatch(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	lbl := ui.NewLabel("lbl", "initial", font, 0)

	ref := ui.NewRef("bound")
	lbl.BindText(ref)

	if lbl.Text() != "bound" {
		t.Fatalf("Text() = %q, want %q", lbl.Text(), "bound")
	}

	lbl.Dispose()

	// After dispose, changing the ref should not panic.
	ref.Set("after-dispose")
	ui.DefaultScheduler.Flush()
}

// --- SetFont ---

func TestLabelSetFontChangesFont(t *testing.T) {
	resetScheduler()
	font1 := newTestFont()
	lbl := ui.NewLabel("lbl", "test", font1, 0)
	defer lbl.Dispose()

	font2 := newTestFont()
	lbl.SetFont(font2)

	if lbl.TextNode().TextBlock.Font != font2 {
		t.Error("Font should have been updated")
	}
}

// --- Width/Height measurement ---

func TestLabelUpdatesMeasuredSize(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	lbl := ui.NewLabel("lbl", "Hello", font, 0)
	defer lbl.Dispose()

	// Measure expected sizes from the actual font.
	wantW, wantH := font.MeasureString("Hello", 0, false, false)
	if lbl.Width != wantW {
		t.Errorf("Width = %f, want %f", lbl.Width, wantW)
	}
	if lbl.Height != wantH {
		t.Errorf("Height = %f, want %f", lbl.Height, wantH)
	}

	lbl.SetText("Hi")
	wantW2, _ := font.MeasureString("Hi", 0, false, false)
	if lbl.Width != wantW2 {
		t.Errorf("Width = %f, want %f after SetText", lbl.Width, wantW2)
	}
}
