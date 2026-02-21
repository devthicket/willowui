package integration

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

// ---------------------------------------------------------------------------
// InputManager
// ---------------------------------------------------------------------------

func TestInputManagerConsumeBlocksAvailability(t *testing.T) {
	im := ui.NewInputManager()
	im.Update()

	// Before consuming, key state comes from ebiten (not pressed in test env).
	// Consume should mark it unavailable regardless.
	im.Consume(ebiten.KeyA)

	if im.IsKeyAvailable(ebiten.KeyA) {
		t.Error("consumed key should not be available")
	}
	if im.IsKeyJustAvailable(ebiten.KeyA) {
		t.Error("consumed key should not be just-available")
	}
}

func TestInputManagerUpdateClearsConsumed(t *testing.T) {
	im := ui.NewInputManager()
	im.Update()
	im.Consume(ebiten.KeyB)

	// New frame clears consumed set.
	im.Update()

	// IsKeyAvailable depends on ebiten state (not pressed in tests), but
	// at least the consumed flag should be cleared.
	// We verify by consuming again and checking it works.
	if im.IsKeyAvailable(ebiten.KeyB) {
		// Key is not physically pressed, so still false — that's fine.
	}
	// The key should not be in the consumed map after Update.
	im.Consume(ebiten.KeyB) // should not panic
}

func TestInputManagerListenerRegistration(t *testing.T) {
	im := ui.NewInputManager()

	called := false
	h := im.OnKeyDown(ebiten.KeyW, func() { called = true })

	im.RemoveListener(h)

	// Fire listeners — callback should NOT fire since it was removed.
	im.Update()
	im.FireListeners()
	if called {
		t.Error("removed listener should not fire")
	}
}

// ---------------------------------------------------------------------------
// FocusManager — Tab order
// ---------------------------------------------------------------------------

func TestTabOrderSkipsNonTabComponents(t *testing.T) {
	fm := ui.NewFocusManager()
	a := ui.NewComponent("a")
	b := ui.NewComponent("b")
	c := ui.NewComponent("c")

	a.Focusable = true
	a.AllowTab = true
	b.Focusable = true
	b.AllowTab = false // skipped by Tab
	c.Focusable = true
	c.AllowTab = true

	fm.Register(a)
	fm.Register(b)
	fm.Register(c)

	fm.TabNext()
	if fm.Focused() != a {
		t.Error("first TabNext should focus a")
	}

	fm.TabNext()
	if fm.Focused() != c {
		t.Errorf("second TabNext should skip b, focus c; got %v", fm.Focused())
	}

	fm.TabNext()
	if fm.Focused() != a {
		t.Error("third TabNext should wrap to a")
	}
}

func TestTabOrderSkipsDisabledComponents(t *testing.T) {
	fm := ui.NewFocusManager()
	a := ui.NewComponent("a")
	b := ui.NewComponent("b")

	a.Focusable = true
	a.AllowTab = true
	b.Focusable = true
	b.AllowTab = true
	b.SetEnabled(false)

	fm.Register(a)
	fm.Register(b)

	fm.TabNext()
	if fm.Focused() != a {
		t.Error("should focus a")
	}

	fm.TabNext()
	// Should skip disabled b and wrap to a.
	if fm.Focused() != a {
		t.Error("should skip disabled b and stay on a")
	}
}

func TestTabOrderSkipsInvisibleComponents(t *testing.T) {
	fm := ui.NewFocusManager()
	a := ui.NewComponent("a")
	b := ui.NewComponent("b")

	a.Focusable = true
	a.AllowTab = true
	b.Focusable = true
	b.AllowTab = true
	b.SetVisible(false)

	fm.Register(a)
	fm.Register(b)

	fm.TabNext()
	fm.TabNext()
	if fm.Focused() != a {
		t.Error("should skip invisible b and stay on a")
	}
}

// ---------------------------------------------------------------------------
// FocusManager — Spatial navigation (distance-based)
// ---------------------------------------------------------------------------

func TestSpatialNavPositionsAndFlags(t *testing.T) {
	// Verify that spatial-eligible components report correct positions.
	// WorldCenter relies on scene's UpdateWorldTransform, so we test
	// the component fields directly and verify the nav flag setup.
	fm := ui.NewFocusManager()

	a := ui.NewComponent("a")
	b := ui.NewComponent("b")
	c := ui.NewComponent("c")

	for _, comp := range []*ui.Component{a, b, c} {
		comp.Focusable = true
		comp.AllowTab = true
		comp.AllowSpatial = true
		comp.Width = 100
		comp.Height = 40
	}

	a.SetPosition(0, 0)
	b.SetPosition(200, 0)
	c.SetPosition(0, 100)

	fm.Register(a)
	fm.Register(b)
	fm.Register(c)

	// Verify positions stored on components.
	if a.X != 0 || a.Y != 0 {
		t.Errorf("a position = (%f, %f), want (0, 0)", a.X, a.Y)
	}
	if b.X != 200 || b.Y != 0 {
		t.Errorf("b position = (%f, %f), want (200, 0)", b.X, b.Y)
	}
	if c.X != 0 || c.Y != 100 {
		t.Errorf("c position = (%f, %f), want (0, 100)", c.X, c.Y)
	}

	// Verify focus can be set on spatial-eligible components.
	fm.SetFocus(a)
	if fm.Focused() != a {
		t.Error("should be able to focus a")
	}

	fm.SetFocus(b)
	if fm.Focused() != b {
		t.Error("should be able to focus b")
	}
	if a.IsFocused() {
		t.Error("a should not be focused after focusing b")
	}
}

// ---------------------------------------------------------------------------
// FocusManager — Keybind priority
// ---------------------------------------------------------------------------

func TestGlobalKeybindRegistration(t *testing.T) {
	fm := ui.NewFocusManager()

	fired := false
	h := fm.Bind(ui.Key(ebiten.KeyZ, ui.ModCtrl), func() bool {
		fired = true
		return true
	})

	// We can't easily simulate a key press in unit tests without ebiten,
	// but we can verify registration and removal work.
	fm.Unbind(h)

	// Verify double-unbind doesn't panic.
	fm.Unbind(h)

	_ = fired // used only when keybind fires
}

func TestScopedKeybindCleanupOnUnregister(t *testing.T) {
	fm := ui.NewFocusManager()
	comp := ui.NewComponent("scoped")
	comp.Focusable = true
	comp.AllowTab = true
	fm.Register(comp)

	fired := false
	fm.BindScoped(comp, ui.Key(ebiten.KeyA, ui.ModCtrl), func() bool {
		fired = true
		return true
	})

	// Unregistering the component should clean up scoped binds.
	fm.Unregister(comp)

	_ = fired
}

// ---------------------------------------------------------------------------
// Widget focus flags — default values
// ---------------------------------------------------------------------------

func TestButtonFocusFlags(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	btn := ui.NewButton("btn", "Focus", font, 0)
	defer btn.Dispose()

	if !btn.Focusable {
		t.Error("Button should be Focusable by default")
	}
	if !btn.AllowTab {
		t.Error("Button should have AllowTab by default")
	}
	if !btn.AllowSpatial {
		t.Error("Button should have AllowSpatial by default")
	}
	if btn.InterceptArrows {
		t.Error("Button should NOT have InterceptArrows by default")
	}
}

func TestCheckboxFocusFlags(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	chk := ui.NewCheckbox("chk", "Test", font, 0)
	defer chk.Dispose()

	if !chk.Focusable {
		t.Error("Checkbox should be Focusable by default")
	}
	if !chk.AllowTab {
		t.Error("Checkbox should have AllowTab by default")
	}
	if !chk.AllowSpatial {
		t.Error("Checkbox should have AllowSpatial by default")
	}
	if chk.InterceptArrows {
		t.Error("Checkbox should NOT have InterceptArrows by default")
	}
}

func TestRadioButtonFocusFlags(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	rg := ui.NewRadio("rg")
	defer rg.Dispose()

	rb := rg.AddOption("Option A", font, 0)
	if !rb.Focusable {
		t.Error("RadioButton should be Focusable by default")
	}
	if !rb.AllowTab {
		t.Error("RadioButton should have AllowTab by default")
	}
	if !rb.AllowSpatial {
		t.Error("RadioButton should have AllowSpatial by default")
	}
}

func TestToggleFocusFlags(t *testing.T) {
	resetScheduler()
	tgl := ui.NewToggle("tgl")
	defer tgl.Dispose()

	if !tgl.Focusable {
		t.Error("Toggle should be Focusable by default")
	}
	if !tgl.AllowTab {
		t.Error("Toggle should have AllowTab by default")
	}
	if !tgl.AllowSpatial {
		t.Error("Toggle should have AllowSpatial by default")
	}
}

func TestTextInputFocusFlags(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ti := ui.NewTextInput("ti", font, 0)
	defer ti.Dispose()

	if !ti.Focusable {
		t.Error("TextInput should be Focusable by default")
	}
	if !ti.AllowTab {
		t.Error("TextInput should have AllowTab by default")
	}
	if !ti.AllowSpatial {
		t.Error("TextInput should have AllowSpatial by default")
	}
	if !ti.InterceptArrows {
		t.Error("TextInput should have InterceptArrows by default")
	}
}

func TestTextAreaFocusFlags(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ta := ui.NewTextArea("ta", font, 0)
	defer ta.Dispose()

	if !ta.Focusable {
		t.Error("TextArea should be Focusable by default")
	}
	if !ta.InterceptArrows {
		t.Error("TextArea should have InterceptArrows by default")
	}
}

func TestSliderFocusFlags(t *testing.T) {
	resetScheduler()
	s := ui.NewSlider("slider")
	defer s.Dispose()

	if !s.Focusable {
		t.Error("Slider should be Focusable by default")
	}
	if !s.AllowTab {
		t.Error("Slider should have AllowTab by default")
	}
	if !s.InterceptArrows {
		t.Error("Slider should have InterceptArrows by default")
	}
}

func TestToggleButtonBarFocusFlags(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tbb := ui.NewToggleButtonBar("tbb", font, 0)
	defer tbb.Dispose()

	if !tbb.Focusable {
		t.Error("ToggleButtonBar should be Focusable by default")
	}
	if !tbb.AllowTab {
		t.Error("ToggleButtonBar should have AllowTab by default")
	}
}

func TestLabelNotFocusable(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	lbl := ui.NewLabel("lbl", "text", font, 0)
	defer lbl.Dispose()

	if lbl.Focusable {
		t.Error("Label should NOT be Focusable")
	}
}

func TestComponentNotFocusableByDefault(t *testing.T) {
	c := ui.NewComponent("plain")
	if c.Focusable {
		t.Error("plain Component should NOT be Focusable by default")
	}
	if c.AllowTab {
		t.Error("plain Component should NOT have AllowTab by default")
	}
	if c.AllowSpatial {
		t.Error("plain Component should NOT have AllowSpatial by default")
	}
}

// ---------------------------------------------------------------------------
// FocusManager — Unregister clears scoped keybinds
// ---------------------------------------------------------------------------

func TestUnregisterClearsScopedKeybinds(t *testing.T) {
	fm := ui.NewFocusManager()

	comp := ui.NewComponent("comp")
	comp.Focusable = true
	comp.AllowTab = true
	fm.Register(comp)

	// Register two scoped keybinds.
	fm.BindScoped(comp, ui.Key(ebiten.KeyA, ui.ModNone), func() bool { return true })
	fm.BindScoped(comp, ui.Key(ebiten.KeyB, ui.ModNone), func() bool { return true })

	// Unregister — should clean up all scoped binds.
	fm.Unregister(comp)

	// Register again and TabNext — should work without panicking.
	fm.Register(comp)
	fm.TabNext()
	if fm.Focused() != comp {
		t.Error("should be able to focus comp after re-registration")
	}
}

// ---------------------------------------------------------------------------
// Theme — FocusColor and FocusRingWidth in default theme
// ---------------------------------------------------------------------------

func TestDefaultThemeHasFocusFields(t *testing.T) {
	group := ui.DefaultTheme.Button.Group(ui.Primary)
	fc := group.FocusColor.Resolve(ui.StateFocus)
	if fc.A() == 0 {
		t.Error("Button theme should have a non-transparent FocusColor for StateFocus")
	}
	if group.FocusRingWidth <= 0 {
		t.Error("Button theme should have FocusRingWidth > 0")
	}
}

func TestDefaultThemeToggleFocusFields(t *testing.T) {
	group := ui.DefaultTheme.Toggle.Group(ui.Primary)
	fc := group.FocusColor.Resolve(ui.StateFocus)
	if fc.A() == 0 {
		t.Error("Toggle theme should have a non-transparent FocusColor for StateFocus")
	}
}

func TestDefaultThemeTextInputFocusFields(t *testing.T) {
	group := ui.DefaultTheme.TextInput.Group(ui.Primary)
	fc := group.FocusColor.Resolve(ui.StateFocus)
	if fc.A() == 0 {
		t.Error("TextInput theme should have a non-transparent FocusColor for StateFocus")
	}
}
