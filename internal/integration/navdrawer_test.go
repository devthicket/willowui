package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func TestNewNavDrawerCreates(t *testing.T) {
	resetScheduler()
	nd := ui.NewNavDrawer("drawer")
	defer nd.Dispose()

	if nd.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if nd.Name() != "drawer" {
		t.Errorf("Name() = %q, want %q", nd.Name(), "drawer")
	}
}

func TestNavDrawerOpenClose(t *testing.T) {
	resetScheduler()
	nd := ui.NewNavDrawer("drawer")
	defer nd.Dispose()

	if nd.IsOpen() {
		t.Error("drawer should start closed")
	}

	nd.Open()
	if !nd.IsOpen() {
		t.Error("drawer should be open after Open()")
	}

	nd.Close()
	if nd.IsOpen() {
		t.Error("drawer should be closed after Close()")
	}
}

func TestNavDrawerToggle(t *testing.T) {
	resetScheduler()
	nd := ui.NewNavDrawer("drawer")
	defer nd.Dispose()

	nd.Toggle()
	if !nd.IsOpen() {
		t.Error("Toggle() should open a closed drawer")
	}

	nd.Toggle()
	if nd.IsOpen() {
		t.Error("Toggle() should close an open drawer")
	}
}

func TestNavDrawerCallbacks(t *testing.T) {
	resetScheduler()
	nd := ui.NewNavDrawer("drawer")
	defer nd.Dispose()

	openCalled := false
	closeCalled := false
	nd.SetOnOpen(func() { openCalled = true })
	nd.SetOnClose(func() { closeCalled = true })

	nd.Open()
	if !openCalled {
		t.Error("OnOpen callback should have fired")
	}

	nd.Close()
	if !closeCalled {
		t.Error("OnClose callback should have fired")
	}
}

func TestNavDrawerPinned(t *testing.T) {
	resetScheduler()
	nd := ui.NewNavDrawer("drawer")
	defer nd.Dispose()

	nd.SetPinned(true)
	if !nd.IsPinned() {
		t.Error("should be pinned after SetPinned(true)")
	}
	if !nd.IsOpen() {
		t.Error("pinning should force drawer open")
	}

	// Backdrop should be hidden when pinned.
	if nd.Backdrop().Visible() {
		t.Error("backdrop should not be visible when pinned")
	}
}

func TestNavDrawerBackdropHiddenWhenPinned(t *testing.T) {
	resetScheduler()
	nd := ui.NewNavDrawer("drawer")
	defer nd.Dispose()
	nd.SetSize(800, 600)

	// Open in overlay mode — backdrop visible.
	nd.Open()
	if !nd.Backdrop().Visible() {
		t.Error("backdrop should be visible when open (overlay mode)")
	}

	// Pin — backdrop should hide.
	nd.SetPinned(true)
	if nd.Backdrop().Visible() {
		t.Error("backdrop should be hidden when pinned")
	}
}

func TestNavDrawerSetSize(t *testing.T) {
	resetScheduler()
	nd := ui.NewNavDrawer("drawer")
	defer nd.Dispose()

	nd.SetSize(800, 600)
	if nd.Width != 800 {
		t.Errorf("Width = %f, want 800", nd.Width)
	}
	if nd.Height != 600 {
		t.Errorf("Height = %f, want 600", nd.Height)
	}
}

func TestNavDrawerAnchorRight(t *testing.T) {
	resetScheduler()
	nd := ui.NewNavDrawer("drawer")
	defer nd.Dispose()

	nd.SetSize(800, 600)
	nd.SetAnchor(ui.NavDrawerRight)

	// When closed, drawer should be off-screen to the right.
	panel := nd.DrawerPanel()
	if panel.X != 800 {
		t.Errorf("right-anchored closed X = %f, want 800", panel.X)
	}
}

func TestNavDrawerSetContent(t *testing.T) {
	resetScheduler()
	nd := ui.NewNavDrawer("drawer")
	defer nd.Dispose()

	p := ui.NewPanel("content")
	nd.SetContent(p)

	// Drawer panel should have the content as child.
	dp := nd.DrawerPanel()
	children := dp.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0].Name() != "content" {
		t.Errorf("child name = %q, want %q", children[0].Name(), "content")
	}

	// Replacing content should remove old child.
	p2 := ui.NewPanel("content2")
	nd.SetContent(p2)
	children = dp.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child after replace, got %d", len(children))
	}
	if children[0].Name() != "content2" {
		t.Errorf("child name = %q, want %q", children[0].Name(), "content2")
	}
}

func TestNavDrawerAnimationProducesIntermediatePosition(t *testing.T) {
	resetScheduler()
	nd := ui.NewNavDrawer("drawer")
	defer nd.Dispose()
	nd.SetSize(800, 600)
	nd.SetWidth(200)

	nd.Open()

	// Advance part of the animation.
	nd.Update(0.05) // 50ms into 250ms animation

	// The slide position should be between 0 and 1 (not fully open).
	panel := nd.DrawerPanel()
	// For left anchor: closed = -200, open = 0. Intermediate should be between.
	if panel.X <= -200 || panel.X >= 0 {
		t.Errorf("expected intermediate position, got X=%f", panel.X)
	}
}
