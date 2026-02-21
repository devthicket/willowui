package integration

import (
	"testing"

	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

// --- NewButton ---

func TestNewButtonCreatesWithText(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Click Me", font, 0)
	defer btn.Dispose()

	if btn.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if btn.Name() != "btn" {
		t.Errorf("Name() = %q, want %q", btn.Name(), "btn")
	}
	if btn.LabelLabel() == nil {
		t.Fatal("label should not be nil")
	}
	if btn.LabelText() != "Click Me" {
		t.Errorf("LabelText() = %q, want %q", btn.LabelText(), "Click Me")
	}
	if btn.BgNode() == nil {
		t.Fatal("background should not be nil")
	}
}

// --- SetOnClick ---

func TestButtonSetOnClickFiresOnClick(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Click", font, 0)
	defer btn.Dispose()

	clicked := false
	btn.SetOnClick(func() {
		clicked = true
	})

	if !btn.Node().HasOnClick() {
		t.Fatal("OnClick should be wired on the button node")
	}
	btn.Node().GetOnClick()(willow.ClickContext{Node: btn.Node()})

	if !clicked {
		t.Error("onClick callback should have fired")
	}
}

// --- Disabled blocks click ---

func TestButtonDisabledBlocksClick(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Click", font, 0)
	defer btn.Dispose()

	clicked := false
	btn.SetOnClick(func() {
		clicked = true
	})

	btn.SetEnabled(false)
	btn.Node().GetOnClick()(willow.ClickContext{Node: btn.Node()})

	if clicked {
		t.Error("onClick should NOT fire when button is disabled")
	}
}

// --- Visual state changes ---

func TestButtonVisualStateChanges(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "State", font, 0)
	defer btn.Dispose()

	btn.SetSize(120, 40)

	group := ui.DefaultTheme.Button.Group(ui.Primary)

	// Normal state: primary color.
	btn.UpdateVisuals()
	wantDefault := group.Background.Resolve(ui.StateDefault).Color
	if btn.BgNode().Color() != wantDefault {
		t.Errorf("normal background = %v, want %v", btn.BgNode().Color(), wantDefault)
	}

	// Simulate hover.
	btn.SimulateHover(true)
	btn.UpdateVisuals()
	wantHover := group.Background.Resolve(ui.StateHover).Color
	if btn.BgNode().Color() != wantHover {
		t.Errorf("hovered background = %v, want %v", btn.BgNode().Color(), wantHover)
	}

	// Simulate pressed (while hovered).
	btn.SimulatePress(true)
	btn.UpdateVisuals()
	wantActive := group.Background.Resolve(ui.StateActive).Color
	if btn.BgNode().Color() != wantActive {
		t.Errorf("pressed background = %v, want %v", btn.BgNode().Color(), wantActive)
	}

	// Disabled overrides everything.
	btn.SimulateHover(false)
	btn.SimulatePress(false)
	btn.SetEnabled(false)
	btn.UpdateVisuals()
	wantDisabled := group.Background.Resolve(ui.StateDisabled).Color
	if btn.BgNode().Color() != wantDisabled {
		t.Errorf("disabled background = %v, want %v", btn.BgNode().Color(), wantDisabled)
	}
}

// --- SetSize ---

func TestButtonSetSizeUpdatesDimensions(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Size", font, 0)
	defer btn.Dispose()

	btn.SetSize(200, 50)

	if btn.Width != 200 {
		t.Errorf("Width = %f, want 200", btn.Width)
	}
	if btn.Height != 50 {
		t.Errorf("Height = %f, want 50", btn.Height)
	}
	if btn.BgNode().ScaleX() != 200 {
		t.Errorf("background.ScaleX = %f, want 200", btn.BgNode().ScaleX())
	}
	if btn.BgNode().ScaleY() != 50 {
		t.Errorf("background.ScaleY = %f, want 50", btn.BgNode().ScaleY())
	}
}

func TestButtonSetSizeSetsHitShape(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Hit", font, 0)
	defer btn.Dispose()

	btn.SetSize(150, 45)

	hs, ok := btn.HitShape().(willow.HitRect)
	if !ok {
		t.Fatal("HitShape should be HitRect after SetSize")
	}
	if hs.Width != 150 || hs.Height != 45 {
		t.Errorf("HitShape = %v, want {0 0 150 45}", hs)
	}
}

// --- Pointer callbacks capture correct Component ---

func TestButtonPointerCallbacksCaptureCorrectComponent(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Hover", font, 0)
	defer btn.Dispose()

	btn.SetSize(100, 40)

	// Simulate pointer enter via the node callback.
	if !btn.Node().HasOnPointerEnter() {
		t.Fatal("OnPointerEnter should be set")
	}
	btn.Node().GetOnPointerEnter()(willow.PointerContext{})

	// The hovered state should be on the Button's embedded Component,
	// not on a dangling copy.
	if !btn.IsHovered() {
		t.Error("pointer enter should set hovered on the Button's Component")
	}

	btn.Node().GetOnPointerDown()(willow.PointerContext{})
	if !btn.IsPressed() {
		t.Error("pointer down should set pressed on the Button's Component")
	}
}

// --- SetText ---

func TestButtonSetTextUpdatesLabel(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Old", font, 0)
	defer btn.Dispose()

	btn.SetText("New")
	if btn.LabelText() != "New" {
		t.Errorf("LabelText() = %q, want %q", btn.LabelText(), "New")
	}
}

// --- AutoSize ---

func TestButtonAutoSizeDefault(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Hello", font, 0)
	defer btn.Dispose()

	if !btn.AutoSize() {
		t.Fatal("new button should have auto-size enabled")
	}
	if btn.Width <= 0 || btn.Height <= 0 {
		t.Fatalf("auto-sized button should have positive dimensions, got %fx%f", btn.Width, btn.Height)
	}
}

func TestButtonAutoSizeFitsText(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Hi", font, 0)
	defer btn.Dispose()

	shortW := btn.Width
	btn.SetText("This is a much longer label")
	if btn.Width <= shortW {
		t.Errorf("auto-size should grow: shortW=%f, newW=%f", shortW, btn.Width)
	}

	longW := btn.Width
	btn.SetText("Hi")
	if btn.Width >= longW {
		t.Errorf("auto-size should shrink: longW=%f, newW=%f", longW, btn.Width)
	}
}

func TestButtonSetSizeDisablesAutoSize(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Auto", font, 0)
	defer btn.Dispose()

	btn.SetSize(300, 60)
	if btn.AutoSize() {
		t.Fatal("SetSize should disable auto-size")
	}
	btn.SetText("Tiny")
	if btn.Width != 300 || btn.Height != 60 {
		t.Errorf("after SetSize, SetText should not resize: got %fx%f", btn.Width, btn.Height)
	}
}

func TestButtonSetAutoSizeReEnables(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Test", font, 0)
	defer btn.Dispose()

	btn.SetSize(300, 60)
	btn.SetAutoSize(true)
	if !btn.AutoSize() {
		t.Fatal("SetAutoSize(true) should re-enable auto-size")
	}
	// Should have re-fit immediately.
	if btn.Width == 300 {
		t.Error("SetAutoSize(true) should re-fit to text, but width unchanged")
	}
}

// --- IconButton ---

func TestIconButtonCreatesWithSprite(t *testing.T) {
	resetScheduler()
	ib := ui.NewIconButton("icon-btn")
	defer ib.Dispose()

	if ib.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if ib.IconNode() == nil {
		t.Fatal("icon should not be nil")
	}
}

func TestIconButtonSetIconImage(t *testing.T) {
	resetScheduler()
	ib := ui.NewIconButton("icon-btn")
	defer ib.Dispose()

	img := ebiten.NewImage(16, 16)
	ib.SetIconImage(img)
	if ib.IconNode() == nil {
		t.Fatal("icon should not be nil after SetIconImage")
	}
}

func TestIconButtonSetSize(t *testing.T) {
	resetScheduler()
	ib := ui.NewIconButton("icon-btn")
	defer ib.Dispose()

	ib.SetSize(64, 48)
	if ib.Width != 64 || ib.Height != 48 {
		t.Errorf("expected 64x48, got %.0fx%.0f", ib.Width, ib.Height)
	}
}

func TestIconButtonSetOnClick(t *testing.T) {
	resetScheduler()
	ib := ui.NewIconButton("icon-btn")
	defer ib.Dispose()

	clicked := false
	ib.SetOnClick(func() { clicked = true })
	ib.SimulateOnClick()
	if !clicked {
		t.Error("onClick should have been called")
	}
}

func TestIconButtonSetActive(t *testing.T) {
	resetScheduler()
	ib := ui.NewIconButton("icon-btn")
	defer ib.Dispose()

	ib.SetActive(true)
	if !ib.IsActive() {
		t.Error("active should be true after SetActive(true)")
	}
	ib.SetActive(false)
	if ib.IsActive() {
		t.Error("active should be false after SetActive(false)")
	}
}

func TestIconButtonBindActive(t *testing.T) {
	resetScheduler()
	ib := ui.NewIconButton("icon-btn")
	defer ib.Dispose()

	ref := ui.NewRef(false)
	ib.BindActive(ref)
	if ib.IsActive() {
		t.Error("should start inactive")
	}

	ref.Set(true)
	ui.DefaultScheduler.Flush()
	if !ib.IsActive() {
		t.Error("should be active after ref set to true")
	}
}

func TestIconButtonSetEnabled(t *testing.T) {
	resetScheduler()
	ib := ui.NewIconButton("icon-btn")
	defer ib.Dispose()

	ib.SetEnabled(false)
	if ib.IsEnabled() {
		t.Error("should be disabled")
	}
	ib.SetEnabled(true)
	if !ib.IsEnabled() {
		t.Error("should be enabled")
	}
}

func TestIconButtonLabelBelow(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ib := ui.NewIconButton("icon-btn")
	defer ib.Dispose()

	ib.SetSize(64, 64)
	ib.SetLabel("Save", font, 12)
	ib.SetLabelPosition(ui.IconLabelBelow)
	// After label is set, icon should be above center and label below.
	// Just ensure no panic and the widget has valid dimensions.
	if ib.Width != 64 || ib.Height != 64 {
		t.Errorf("size changed unexpectedly: %.0fx%.0f", ib.Width, ib.Height)
	}
}

func TestIconButtonLabelRight(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ib := ui.NewIconButton("icon-btn")
	defer ib.Dispose()

	ib.SetSize(120, 40)
	ib.SetLabel("Open", font, 12)
	ib.SetLabelPosition(ui.IconLabelRight)
	if ib.Width != 120 || ib.Height != 40 {
		t.Errorf("size changed unexpectedly: %.0fx%.0f", ib.Width, ib.Height)
	}
}

// --- Button Dispose ---

func TestButtonDispose(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Dispose", font, 0)

	btn.Dispose()

	if !btn.IsDisposed() {
		t.Error("node should be disposed after Button.Dispose()")
	}
}
