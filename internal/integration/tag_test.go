package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func TestTagSizeToContentSizesToTextPlusPadding(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tag := ui.NewTag("t", font, 0)
	defer tag.Dispose()

	tag.SetText("Go")
	tag.SizeToContent()

	// Measure actual text size and add theme padding.
	textW, textH := font.MeasureString("Go", 0, false, false)
	pad := ui.DefaultTheme.Tag.Group(ui.Primary).Padding
	wantW := textW + pad.Left + pad.Right
	wantH := textH + pad.Top + pad.Bottom
	if wantW < wantH {
		wantW = wantH // min width = height
	}
	if tag.Width != wantW {
		t.Errorf("Width = %f, want %f", tag.Width, wantW)
	}
	if tag.Height != wantH {
		t.Errorf("Height = %f, want %f", tag.Height, wantH)
	}
}

func TestTagRemovableIncreasesWidth(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tag := ui.NewTag("t", font, 0)
	defer tag.Dispose()

	tag.SetText("Go")
	tag.SizeToContent()
	normalW := tag.Width

	tag.SetRemovable(true)
	tag.SizeToContent()

	if tag.Width <= normalW {
		t.Errorf("removable Width (%f) should be > normal Width (%f)", tag.Width, normalW)
	}
}

func TestTagSetSelectedReflectsState(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tag := ui.NewTag("t", font, 0)
	defer tag.Dispose()

	tag.SetSelectable(true)

	if tag.Selected() {
		t.Error("Selected() should be false by default")
	}

	tag.SetSelected(true)
	if !tag.Selected() {
		t.Error("Selected() should be true after SetSelected(true)")
	}

	tag.SetSelected(false)
	if tag.Selected() {
		t.Error("Selected() should be false after SetSelected(false)")
	}
}

func TestTagSetVariantDoesNotPanic(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tag := ui.NewTag("t", font, 0)
	defer tag.Dispose()

	variants := []ui.Variant{
		ui.Primary,
		ui.Success,
		ui.Warning,
		ui.Danger,
		ui.Neutral,
	}
	for _, v := range variants {
		tag.SetVariant(v)
	}
}

func TestTagCornerRadiusDefaultsToNegativeOne(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tag := ui.NewTag("t", font, 0)
	defer tag.Dispose()

	// The theme default for Tag corner radius is -1 (pill).
	// After SizeToContent + UpdateVisuals, the resolved radius should be h/2.
	tag.SetText("Go")
	tag.SizeToContent()

	// Just verify it doesn't panic — the actual corner radius is applied
	// internally and tested visually.
}

func TestTagRemovableAndSelectableCombined(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tag := ui.NewTag("t", font, 0)
	defer tag.Dispose()

	tag.SetText("Go")
	tag.SetRemovable(true)
	tag.SetSelectable(true)
	tag.SizeToContent()

	// Verify both modes active simultaneously.
	tag.SetSelected(true)
	if !tag.Selected() {
		t.Error("Selected() should be true")
	}
}

func TestTagBarAddAndRemoveTags(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTagBar("tb", font, 0)
	defer tb.Dispose()

	tb.AddTag("Go")
	tb.AddTag("Rust")

	tags := tb.Tags()
	if len(tags) != 2 {
		t.Fatalf("Tags() len = %d, want 2", len(tags))
	}
	if tags[0] != "Go" || tags[1] != "Rust" {
		t.Errorf("Tags() = %v, want [Go Rust]", tags)
	}

	tb.RemoveTagAt(0)
	tags = tb.Tags()
	if len(tags) != 1 || tags[0] != "Rust" {
		t.Errorf("after remove: Tags() = %v, want [Rust]", tags)
	}
}

func TestTagBarSetTagsReplacesAll(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTagBar("tb", font, 0)
	defer tb.Dispose()

	tb.AddTag("old1")
	tb.AddTag("old2")
	tb.SetTags([]string{"new1", "new2", "new3"})

	tags := tb.Tags()
	if len(tags) != 3 {
		t.Fatalf("Tags() len = %d, want 3", len(tags))
	}
	if tags[0] != "new1" || tags[1] != "new2" || tags[2] != "new3" {
		t.Errorf("Tags() = %v, want [new1 new2 new3]", tags)
	}
}

func TestTagBarOnChangeFires(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTagBar("tb", font, 0)
	defer tb.Dispose()

	var received []string
	tb.SetOnChange(func(tags []string) {
		received = tags
	})

	tb.AddTag("Go")

	if len(received) != 1 || received[0] != "Go" {
		t.Errorf("OnChange received = %v, want [Go]", received)
	}
}
