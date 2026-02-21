package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

// testController records lifecycle calls for verification.
type testController struct {
	createCount  int
	updateCount  int
	destroyCount int
	screen       *ui.Screen
	lastDT       float64
}

func (c *testController) OnCreate(s *ui.Screen) {
	c.createCount++
	c.screen = s
}

func (c *testController) OnUpdate(dt float64) {
	c.updateCount++
	c.lastDT = dt
}

func (c *testController) OnDestroy() {
	c.destroyCount++
}

// testDisposable tracks Stop calls.
type testDisposable struct {
	stopped bool
}

func (d *testDisposable) Stop() {
	d.stopped = true
}

// ---------------------------------------------------------------------------
// Screen lifecycle
// ---------------------------------------------------------------------------

func TestScreen_ShowCallsOnCreate(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	ctrl := &testController{}
	s := ui.NewScreen(ui.WithScene(scene), ui.WithController(ctrl))

	s.Show()

	if ctrl.createCount != 1 {
		t.Fatalf("OnCreate called %d times, want 1", ctrl.createCount)
	}
	if ctrl.screen != s {
		t.Fatal("OnCreate did not receive the correct screen")
	}
}

func TestScreen_ShowOnlyCreatesOnce(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	ctrl := &testController{}
	s := ui.NewScreen(ui.WithScene(scene), ui.WithController(ctrl))

	s.Show()
	s.Hide()
	s.Show()

	if ctrl.createCount != 1 {
		t.Fatalf("OnCreate called %d times after show/hide/show, want 1", ctrl.createCount)
	}
}

func TestScreen_DestroyCallsOnDestroy(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	ctrl := &testController{}
	s := ui.NewScreen(ui.WithScene(scene), ui.WithController(ctrl))

	s.Show()
	s.Destroy()

	if ctrl.destroyCount != 1 {
		t.Fatalf("OnDestroy called %d times, want 1", ctrl.destroyCount)
	}
}

func TestScreen_DestroyWithoutShowDoesNotCallOnDestroy(t *testing.T) {
	resetScheduler()
	ctrl := &testController{}
	s := ui.NewScreen(ui.WithController(ctrl))

	s.Destroy()

	if ctrl.destroyCount != 0 {
		t.Fatalf("OnDestroy called %d times on never-shown screen, want 0", ctrl.destroyCount)
	}
}

// ---------------------------------------------------------------------------
// Hide/Show preserves state
// ---------------------------------------------------------------------------

func TestScreen_HideShowPreservesState(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	ctrl := &testController{}
	s := ui.NewScreen(ui.WithScene(scene), ui.WithController(ctrl))

	s.Show()

	// Add a child to prove state is preserved.
	child := ui.NewComponent("test-child")
	s.Add(child)

	s.Hide()
	if s.Visible() {
		t.Fatal("screen should not be visible after Hide")
	}

	s.Show()
	if !s.Visible() {
		t.Fatal("screen should be visible after Show")
	}

	if s.NumChildren() != 1 {
		t.Fatalf("expected 1 child after hide/show, got %d", s.NumChildren())
	}
}

// ---------------------------------------------------------------------------
// TrackRef auto-dispose
// ---------------------------------------------------------------------------

func TestScreen_TrackRefAutoDispose(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	ctrl := &testController{}
	s := ui.NewScreen(ui.WithScene(scene), ui.WithController(ctrl))

	d1 := &testDisposable{}
	d2 := &testDisposable{}

	s.Show()
	s.TrackRef(d1)
	s.TrackRef(d2)
	s.Destroy()

	if !d1.stopped {
		t.Fatal("disposable 1 was not stopped on Destroy")
	}
	if !d2.stopped {
		t.Fatal("disposable 2 was not stopped on Destroy")
	}
}

// ---------------------------------------------------------------------------
// Scheduler flushes on Update
// ---------------------------------------------------------------------------

func TestScreen_UpdateFlushesScheduler(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	ctrl := &testController{}
	s := ui.NewScreen(ui.WithScene(scene), ui.WithController(ctrl))
	s.Show()

	ref := ui.NewRef("hello")
	var observed string
	ui.WatchEffect(func() {
		observed = ref.Get()
	})

	ref.Set("world")

	// Before Update, the watch hasn't flushed.
	if observed != "hello" {
		t.Fatalf("watch fired before Update: got %q", observed)
	}

	s.Update(0.016)

	if observed != "world" {
		t.Fatalf("after Update, observed = %q, want %q", observed, "world")
	}
	if ctrl.updateCount != 1 {
		t.Fatalf("OnUpdate called %d times, want 1", ctrl.updateCount)
	}
}

// ---------------------------------------------------------------------------
// FindByName
// ---------------------------------------------------------------------------

func TestScreen_FindByName(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	ctrl := &testController{}
	s := ui.NewScreen(ui.WithScene(scene), ui.WithController(ctrl))
	s.Show()

	parent := ui.NewComponent("parent")
	child := ui.NewComponent("target")
	parent.AddChild(child)
	s.Add(parent)

	found := s.FindByName("target")
	if found != child {
		t.Fatal("FindByName did not find the target component")
	}

	notFound := s.FindByName("nonexistent")
	if notFound != nil {
		t.Fatal("FindByName should return nil for nonexistent name")
	}
}

// ---------------------------------------------------------------------------
// StageManager
// ---------------------------------------------------------------------------

func TestStage_Add(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	sm := ui.NewStageManager()
	sm.SetScene(scene)

	ctrl := &testController{}
	s := ui.NewScreen(ui.WithController(ctrl))

	sm.Add(s)
	if sm.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", sm.Size())
	}
	if sm.Top() != s {
		t.Fatal("top should be s")
	}
	if !s.Visible() {
		t.Fatal("s should be visible after Add")
	}
	if ctrl.createCount != 1 {
		t.Fatalf("OnCreate called %d times, want 1", ctrl.createCount)
	}
}

func TestStage_AddMultiple(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	sm := ui.NewStageManager()
	sm.SetScene(scene)

	ctrl1 := &testController{}
	ctrl2 := &testController{}
	s1 := ui.NewScreen(ui.WithController(ctrl1))
	s2 := ui.NewScreen(ui.WithController(ctrl2))

	sm.Add(s1)
	sm.Add(s2)

	// Both screens are visible in the new Add semantics.
	if sm.Size() != 2 {
		t.Fatalf("stack size = %d, want 2", sm.Size())
	}
	if !s1.Visible() {
		t.Fatal("s1 should remain visible after s2 Add")
	}
	if !s2.Visible() {
		t.Fatal("s2 should be visible after Add")
	}
}

func TestStage_Remove(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	sm := ui.NewStageManager()
	sm.SetScene(scene)

	ctrl1 := &testController{}
	ctrl2 := &testController{}
	s1 := ui.NewScreen(ui.WithController(ctrl1))
	s2 := ui.NewScreen(ui.WithController(ctrl2))

	sm.Add(s1)
	sm.Add(s2)

	sm.Remove(s2)

	if sm.Size() != 1 {
		t.Fatalf("stack size = %d after Remove, want 1", sm.Size())
	}
	if s2.Visible() {
		t.Fatal("s2 should not be visible after Remove")
	}
	if ctrl2.destroyCount != 1 {
		t.Fatal("s2 OnDestroy should have been called on Remove")
	}
	if !s1.Visible() {
		t.Fatal("s1 should still be visible after s2 Remove")
	}
}

func TestStage_Replace(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	sm := ui.NewStageManager()
	sm.SetScene(scene)

	ctrl1 := &testController{}
	ctrl2 := &testController{}
	s1 := ui.NewScreen(ui.WithController(ctrl1))
	s2 := ui.NewScreen(ui.WithController(ctrl2))

	sm.Add(s1)
	sm.Replace(s2)

	if sm.Size() != 1 {
		t.Fatalf("stack size = %d after replace, want 1", sm.Size())
	}
	if sm.Top() != s2 {
		t.Fatal("top should be s2 after replace")
	}
	if ctrl1.destroyCount != 1 {
		t.Fatal("s1 should be destroyed after replace")
	}
	if !s2.Visible() {
		t.Fatal("s2 should be visible after replace")
	}
}

func TestStage_CloseAll(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	sm := ui.NewStageManager()
	sm.SetScene(scene)

	ctrl1 := &testController{}
	ctrl2 := &testController{}
	sm.Add(ui.NewScreen(ui.WithController(ctrl1)))
	sm.Add(ui.NewScreen(ui.WithController(ctrl2)))

	sm.CloseAll()

	if sm.Size() != 0 {
		t.Fatalf("stack size = %d after CloseAll, want 0", sm.Size())
	}
	if ctrl1.destroyCount != 1 {
		t.Fatal("s1 OnDestroy should have been called")
	}
	if ctrl2.destroyCount != 1 {
		t.Fatal("s2 OnDestroy should have been called")
	}
}

func TestStage_Update(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	sm := ui.NewStageManager()
	sm.SetScene(scene)

	ctrl := &testController{}
	sm.Add(ui.NewScreen(ui.WithController(ctrl)))

	sm.Update(0.016)

	if ctrl.updateCount != 1 {
		t.Fatalf("OnUpdate called %d times, want 1", ctrl.updateCount)
	}
}

func TestStage_UpdateEmpty(t *testing.T) {
	resetScheduler()
	sm := ui.NewStageManager()

	// Should not panic.
	sm.Update(0.016)
}

// ---------------------------------------------------------------------------
// Reactive data through controller
// ---------------------------------------------------------------------------

type reactiveController struct {
	counter *ui.Ref[int]
}

func (c *reactiveController) OnCreate(s *ui.Screen) {
	c.counter = ui.NewRef(0)
	s.TrackRef(ui.WatchEffect(func() {
		_ = c.counter.Get()
	}))
}

func (c *reactiveController) OnUpdate(dt float64) {}
func (c *reactiveController) OnDestroy()          {}

func TestScreen_ReactiveDataThroughController(t *testing.T) {
	resetScheduler()
	scene := newTestScene()
	ctrl := &reactiveController{}
	s := ui.NewScreen(ui.WithScene(scene), ui.WithController(ctrl))

	s.Show()

	ctrl.counter.Set(42)
	s.Update(0.016)

	if ctrl.counter.Peek() != 42 {
		t.Fatalf("counter = %d, want 42", ctrl.counter.Peek())
	}
}
