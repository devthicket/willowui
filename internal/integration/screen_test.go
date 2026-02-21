package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func TestClearTemplateTree(t *testing.T) {
	resetScheduler()

	screen := ui.NewScreen()

	// Add some children.
	screen.Add(ui.NewComponent("child1"))
	screen.Add(ui.NewComponent("child2"))

	// Track a disposable.
	stopped := false
	screen.TrackRef(&mockDisposable{onStop: func() { stopped = true }})

	if screen.NumChildren() != 2 {
		t.Fatalf("before clear: children = %d, want 2", screen.NumChildren())
	}

	screen.ClearTemplateTree()

	if screen.NumChildren() != 0 {
		t.Errorf("after clear: children = %d, want 0", screen.NumChildren())
	}
	if !stopped {
		t.Error("expected tracked ref to be stopped")
	}
}

type mockDisposable struct {
	onStop func()
}

func (m *mockDisposable) Stop() {
	if m.onStop != nil {
		m.onStop()
	}
}
