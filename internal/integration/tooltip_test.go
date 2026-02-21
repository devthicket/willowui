package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func TestNewTooltipDefaults(t *testing.T) {
	resetScheduler()
	tt := ui.NewTooltip("tt")
	defer tt.Dispose()

	if tt.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if tt.Name() != "tt" {
		t.Errorf("Name() = %q, want %q", tt.Name(), "tt")
	}
	if tt.IsShowing() {
		t.Error("IsShowing() should be false by default")
	}
}

func TestTooltipIsNotShowingByDefault(t *testing.T) {
	resetScheduler()
	tt := ui.NewTooltip("tt")
	defer tt.Dispose()

	// Without a scene, Show/Hide require node parenting — just verify default state.
	if tt.IsShowing() {
		t.Error("IsShowing() should be false by default")
	}
}

func TestTooltipSetText(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tt := ui.NewTooltip("tt")
	defer tt.Dispose()

	// SetText must not panic.
	tt.SetText("Hover info", font, 14)
}

func TestTooltipSetSize(t *testing.T) {
	resetScheduler()
	tt := ui.NewTooltip("tt")
	defer tt.Dispose()

	tt.SetSize(120, 30)
	if tt.Width != 120 {
		t.Errorf("Width = %f, want 120", tt.Width)
	}
	if tt.Height != 30 {
		t.Errorf("Height = %f, want 30", tt.Height)
	}
}

func TestTooltipShowDelayDefault(t *testing.T) {
	resetScheduler()
	tt := ui.NewTooltip("tt")
	defer tt.Dispose()

	if tt.ShowDelay <= 0 {
		t.Errorf("ShowDelay = %d, want > 0 (default delay)", tt.ShowDelay)
	}
}

func TestTooltipSetShowDelay(t *testing.T) {
	resetScheduler()
	tt := ui.NewTooltip("tt")
	defer tt.Dispose()

	tt.SetShowDelay(60)
	if tt.ShowDelay != 60 {
		t.Errorf("ShowDelay = %d, want 60", tt.ShowDelay)
	}
}

func TestTooltipSetHideDelay(t *testing.T) {
	resetScheduler()
	tt := ui.NewTooltip("tt")
	defer tt.Dispose()

	tt.SetHideDelay(10)
	if tt.HideDelay != 10 {
		t.Errorf("HideDelay = %d, want 10", tt.HideDelay)
	}
}

func TestTooltipSetAnchor(t *testing.T) {
	resetScheduler()
	tt := ui.NewTooltip("tt")
	defer tt.Dispose()

	tt.SetAnchor(ui.TooltipAbove)
	if tt.Anchor != ui.TooltipAbove {
		t.Errorf("Anchor = %v, want TooltipAbove", tt.Anchor)
	}

	tt.SetAnchor(ui.TooltipBelow)
	if tt.Anchor != ui.TooltipBelow {
		t.Errorf("Anchor = %v, want TooltipBelow", tt.Anchor)
	}
}

func TestTooltipSetOffset(t *testing.T) {
	resetScheduler()
	tt := ui.NewTooltip("tt")
	defer tt.Dispose()

	tt.SetOffset(5, 10)
	if tt.OffsetX != 5 || tt.OffsetY != 10 {
		t.Errorf("Offset = (%f, %f), want (5, 10)", tt.OffsetX, tt.OffsetY)
	}
}

func TestTooltipSetFadeIn(t *testing.T) {
	resetScheduler()
	tt := ui.NewTooltip("tt")
	defer tt.Dispose()

	tt.SetFadeIn(0.2)
	if tt.FadeInDuration != 0.2 {
		t.Errorf("FadeInDuration = %f, want 0.2", tt.FadeInDuration)
	}
}

func TestTooltipSetClampToScreen(t *testing.T) {
	resetScheduler()
	tt := ui.NewTooltip("tt")
	defer tt.Dispose()

	tt.SetClampToScreen(true)
	if !tt.ClampToScreen {
		t.Error("ClampToScreen should be true after SetClampToScreen(true)")
	}

	tt.SetClampToScreen(false)
	if tt.ClampToScreen {
		t.Error("ClampToScreen should be false after SetClampToScreen(false)")
	}
}

func TestTooltipDispose(t *testing.T) {
	resetScheduler()
	tt := ui.NewTooltip("tt")
	tt.Dispose()
	if !tt.IsDisposed() {
		t.Error("IsDisposed() should be true after Dispose()")
	}
}
