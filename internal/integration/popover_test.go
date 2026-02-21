package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

// ---------------------------------------------------------------------------
// Construction / API
// ---------------------------------------------------------------------------

func TestNewPopoverDoesNotPanic(t *testing.T) {
	resetScheduler()
	p := ui.NewPopover("test-pop")
	if p == nil {
		t.Fatal("NewPopover returned nil")
	}
}

func TestPopoverIsOpenDefaultFalse(t *testing.T) {
	resetScheduler()
	p := ui.NewPopover("pop")
	if p.IsOpen() {
		t.Error("new popover should not be open")
	}
}

func TestPopoverSetPreferredSide(t *testing.T) {
	resetScheduler()
	p := ui.NewPopover("pop")
	p.SetPreferredSide(ui.PopoverAbove)
	p.SetPreferredSide(ui.PopoverBelow)
	p.SetPreferredSide(ui.PopoverLeft)
	p.SetPreferredSide(ui.PopoverRight)
}

func TestPopoverSetContentSize(t *testing.T) {
	resetScheduler()
	p := ui.NewPopover("pop")
	p.SetContentSize(200, 150)
}

func TestPopoverSetTitle(t *testing.T) {
	resetScheduler()
	p := ui.NewPopover("pop")
	p.SetTitle("Filters", newTestFont(), 13)
}

func TestPopoverSetShowCloseButton(t *testing.T) {
	resetScheduler()
	p := ui.NewPopover("pop")
	p.SetShowCloseButton(true)
	p.SetShowCloseButton(false)
}

func TestPopoverSetContent(t *testing.T) {
	resetScheduler()
	content := ui.NewPanel("content")
	p := ui.NewPopover("pop")
	p.SetContent(content)
}

func TestPopoverOnOpenCallback(t *testing.T) {
	resetScheduler()
	called := false
	p := ui.NewPopover("pop")
	p.SetOnOpen(func() { called = true })
	_ = called
}

func TestPopoverOnCloseCallback(t *testing.T) {
	resetScheduler()
	called := false
	p := ui.NewPopover("pop")
	p.SetOnClose(func() { called = true })
	_ = called
}

// ---------------------------------------------------------------------------
// Manager: open/close without scene
// ---------------------------------------------------------------------------

func TestPopoverCloseNoOpWhenNotOpen(t *testing.T) {
	resetScheduler()
	p := ui.NewPopover("pop")
	// Should not panic.
	p.Close()
}

func TestPopoverOpenWithoutSceneDoesNotPanic(t *testing.T) {
	resetScheduler()
	mgr := &ui.PopoverManager{}
	p := ui.NewPopover("pop")
	p.SetContentSize(200, 100)
	// Trigger is nil — should center on screen without panic.
	mgr.Open(p, nil)
}

func TestPopoverOpenSetsIsOpen(t *testing.T) {
	resetScheduler()
	// Use a fresh manager (no scene) to avoid touching global state.
	mgr := &ui.PopoverManager{}
	p := ui.NewPopover("pop")
	p.SetContentSize(200, 100)
	mgr.Open(p, nil)
	if !p.IsOpen() {
		t.Error("popover should be open after Open()")
	}
}

func TestPopoverCloseResetsIsOpen(t *testing.T) {
	resetScheduler()
	mgr := &ui.PopoverManager{}
	p := ui.NewPopover("pop")
	p.SetContentSize(200, 100)
	mgr.Open(p, nil)
	mgr.Close(p)
	if p.IsOpen() {
		t.Error("popover should be closed after Close()")
	}
}

func TestPopoverOnOpenFires(t *testing.T) {
	resetScheduler()
	mgr := &ui.PopoverManager{}
	opened := false
	p := ui.NewPopover("pop")
	p.SetContentSize(200, 100)
	p.SetOnOpen(func() { opened = true })
	mgr.Open(p, nil)
	if !opened {
		t.Error("OnOpen callback was not fired")
	}
}

func TestPopoverOnCloseFires(t *testing.T) {
	resetScheduler()
	mgr := &ui.PopoverManager{}
	closed := false
	p := ui.NewPopover("pop")
	p.SetContentSize(200, 100)
	p.SetOnClose(func() { closed = true })
	mgr.Open(p, nil)
	mgr.Close(p)
	if !closed {
		t.Error("OnClose callback was not fired")
	}
}

func TestPopoverOpenSecondClosesFirst(t *testing.T) {
	resetScheduler()
	mgr := &ui.PopoverManager{}
	closed := false
	p1 := ui.NewPopover("pop1")
	p1.SetContentSize(200, 100)
	p1.SetOnClose(func() { closed = true })

	p2 := ui.NewPopover("pop2")
	p2.SetContentSize(200, 100)

	mgr.Open(p1, nil)
	mgr.Open(p2, nil) // should close p1 first
	if !closed {
		t.Error("opening a second popover should close the first")
	}
	if p1.IsOpen() {
		t.Error("first popover should be closed")
	}
	if !p2.IsOpen() {
		t.Error("second popover should be open")
	}
}

// ---------------------------------------------------------------------------
// PopoverSide constants
// ---------------------------------------------------------------------------

func TestPopoverSideConstantsDistinct(t *testing.T) {
	sides := []ui.PopoverSide{
		ui.PopoverBelow,
		ui.PopoverAbove,
		ui.PopoverRight,
		ui.PopoverLeft,
	}
	seen := map[ui.PopoverSide]bool{}
	for _, s := range sides {
		if seen[s] {
			t.Errorf("duplicate PopoverSide value %d", s)
		}
		seen[s] = true
	}
}

// ---------------------------------------------------------------------------
// DefaultPopoverManager
// ---------------------------------------------------------------------------

func TestDefaultPopoverManagerIsNotNil(t *testing.T) {
	if ui.DefaultPopoverManager == nil {
		t.Fatal("DefaultPopoverManager must not be nil")
	}
}
