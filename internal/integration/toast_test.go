package integration

import (
	"testing"
	"time"

	ui "github.com/devthicket/willowui"
)

// ---------------------------------------------------------------------------
// Construction / API
// ---------------------------------------------------------------------------

func TestShowToastDoesNotPanicWithoutScene(t *testing.T) {
	resetScheduler()
	// Without a scene the manager has no overlayNode; Show should return silently.
	mgr := &ui.ToastManager{}
	mgr.Show("hello", ui.Primary)
}

func TestShowToastVariantInfo(t *testing.T) {
	resetScheduler()
	mgr := &ui.ToastManager{}
	mgr.SetMaxStack(4)
	mgr.SetMargin(16, 16)
	// No-op without scene; just verify no panic.
	mgr.Show("info", ui.Info)
}

func TestShowToastAllVariants(t *testing.T) {
	resetScheduler()
	mgr := &ui.ToastManager{}
	// Verify all variant values are accepted.
	mgr.Show("primary", ui.Primary)
	mgr.Show("info", ui.Info)
	mgr.Show("success", ui.Success)
	mgr.Show("warning", ui.Warning)
	mgr.Show("danger", ui.Danger)
}

func TestToastManagerSetAnchor(t *testing.T) {
	resetScheduler()
	mgr := &ui.ToastManager{}
	mgr.SetAnchor(ui.ToastTopLeft)
	mgr.SetAnchor(ui.ToastTopRight)
	mgr.SetAnchor(ui.ToastBottomLeft)
	mgr.SetAnchor(ui.ToastBottomRight)
}

func TestToastManagerSetMaxStack(t *testing.T) {
	resetScheduler()
	mgr := &ui.ToastManager{}
	mgr.SetMaxStack(3)
	// Clamp to minimum 1.
	mgr.SetMaxStack(0)
}

func TestToastManagerSetMargin(t *testing.T) {
	resetScheduler()
	mgr := &ui.ToastManager{}
	mgr.SetMargin(20, 20)
}

func TestToastManagerDismissAllNoOp(t *testing.T) {
	resetScheduler()
	mgr := &ui.ToastManager{}
	// Should not panic with no active toasts.
	mgr.DismissAll()
}

// ---------------------------------------------------------------------------
// Options
// ---------------------------------------------------------------------------

func TestWithDurationOption(t *testing.T) {
	resetScheduler()
	called := false
	fn := ui.WithDuration(5 * time.Second)
	if fn == nil {
		t.Fatal("WithDuration should return a non-nil option")
	}
	_ = called
}

func TestWithDismissOnClickOption(t *testing.T) {
	resetScheduler()
	fn := ui.WithDismissOnClick(false)
	if fn == nil {
		t.Fatal("WithDismissOnClick should return a non-nil option")
	}
}

func TestWithProgressOption(t *testing.T) {
	resetScheduler()
	fn := ui.WithProgress(true)
	if fn == nil {
		t.Fatal("WithProgress should return a non-nil option")
	}
}

func TestWithOnDismissOption(t *testing.T) {
	resetScheduler()
	dismissed := false
	fn := ui.WithOnDismiss(func() { dismissed = true })
	if fn == nil {
		t.Fatal("WithOnDismiss should return a non-nil option")
	}
	_ = dismissed
}

// ---------------------------------------------------------------------------
// Package-level helpers
// ---------------------------------------------------------------------------

func TestShowToastGlobalHelperDoesNotPanic(t *testing.T) {
	resetScheduler()
	// Reset DefaultToastManager to a clean state without scene.
	prev := ui.DefaultToastManager
	ui.DefaultToastManager = &ui.ToastManager{}
	defer func() { ui.DefaultToastManager = prev }()

	ui.ShowToast("hello", ui.Primary)
	ui.ShowToast("done", ui.Success)
}

// ---------------------------------------------------------------------------
// Anchor constants
// ---------------------------------------------------------------------------

func TestToastAnchorConstantValues(t *testing.T) {
	anchors := []ui.ToastAnchor{
		ui.ToastBottomRight,
		ui.ToastBottomLeft,
		ui.ToastTopRight,
		ui.ToastTopLeft,
	}
	seen := map[ui.ToastAnchor]bool{}
	for _, a := range anchors {
		if seen[a] {
			t.Errorf("duplicate ToastAnchor value %d", a)
		}
		seen[a] = true
	}
}
