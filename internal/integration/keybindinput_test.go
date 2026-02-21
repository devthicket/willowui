package integration

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

func TestKeybindInputNewIsUnset(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	kb := ui.NewKeybindInput("kb", font, 14)
	defer kb.Dispose()

	if kb.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if !kb.Binding().IsUnset {
		t.Error("new keybind input should have IsUnset = true")
	}
	if kb.IsListening() {
		t.Error("new keybind input should not be listening")
	}
}

func TestKeybindInputSetBindingUpdatesBinding(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	kb := ui.NewKeybindInput("kb", font, 14)
	defer kb.Dispose()

	binding := ui.KeyBinding{Key: ebiten.KeySpace}
	kb.SetBinding(binding)

	got := kb.Binding()
	if got.Key != ebiten.KeySpace {
		t.Errorf("Binding().Key = %v, want KeySpace", got.Key)
	}
	if got.IsUnset {
		t.Error("Binding().IsUnset should be false after SetBinding")
	}
	if got.IsGamepad {
		t.Error("Binding().IsGamepad should be false")
	}
}

func TestKeybindInputSetListeningEntersListeningMode(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	kb := ui.NewKeybindInput("kb", font, 14)
	defer kb.Dispose()

	kb.SetListening(true)
	if !kb.IsListening() {
		t.Error("should be listening after SetListening(true)")
	}

	kb.SetListening(false)
	if kb.IsListening() {
		t.Error("should not be listening after SetListening(false)")
	}
}

func TestKeybindInputClearBindingSetsUnset(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	kb := ui.NewKeybindInput("kb", font, 14)
	defer kb.Dispose()

	kb.SetBinding(ui.KeyBinding{Key: ebiten.KeyA})
	kb.ClearBinding()

	if !kb.Binding().IsUnset {
		t.Error("Binding().IsUnset should be true after ClearBinding")
	}
}

func TestKeybindInputClearBindingFiresCallback(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	kb := ui.NewKeybindInput("kb", font, 14)
	defer kb.Dispose()

	var called bool
	kb.SetOnBindingChanged(func(b ui.KeyBinding) {
		called = true
		if !b.IsUnset {
			t.Error("callback binding should be unset")
		}
	})

	kb.SetBinding(ui.KeyBinding{Key: ebiten.KeyA})
	kb.ClearBinding()

	if !called {
		t.Error("OnBindingChanged should have been called")
	}
}

func TestKeybindInputSetListeningFalseRetainsBinding(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	kb := ui.NewKeybindInput("kb", font, 14)
	defer kb.Dispose()

	kb.SetBinding(ui.KeyBinding{Key: ebiten.KeyW})
	kb.SetListening(true)
	kb.SetListening(false) // cancel without capturing

	got := kb.Binding()
	if got.Key != ebiten.KeyW {
		t.Errorf("binding should be retained, got Key = %v", got.Key)
	}
}

func TestKeybindInputSetSizeUpdatesHitShape(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	kb := ui.NewKeybindInput("kb", font, 14)
	defer kb.Dispose()

	kb.SetSize(300, 40)

	if kb.Width != 300 {
		t.Errorf("Width = %f, want 300", kb.Width)
	}
	if kb.Height != 40 {
		t.Errorf("Height = %f, want 40", kb.Height)
	}
}

func TestKeybindInputDisabledStopsListening(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	kb := ui.NewKeybindInput("kb", font, 14)
	defer kb.Dispose()

	kb.SetListening(true)
	kb.SetEnabled(false)

	if kb.IsListening() {
		t.Error("should stop listening when disabled")
	}
}

func TestKeybindInputCombosEnabledByDefault(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	kb := ui.NewKeybindInput("kb", font, 14)
	defer kb.Dispose()

	if !kb.CombosEnabled() {
		t.Error("combos should be enabled by default")
	}
}

func TestKeybindInputSetCombosEnabled(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	kb := ui.NewKeybindInput("kb", font, 14)
	defer kb.Dispose()

	kb.SetCombosEnabled(false)
	if kb.CombosEnabled() {
		t.Error("combos should be disabled after SetCombosEnabled(false)")
	}
}

func TestKeyBindingDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		binding ui.KeyBinding
		want    string
	}{
		{"unset", ui.KeyBinding{IsUnset: true}, ""},
		{"simple key", ui.KeyBinding{Key: ebiten.KeyA}, "A"},
		{"space", ui.KeyBinding{Key: ebiten.KeySpace}, "SPACE"},
		{"ctrl+A", ui.KeyBinding{Key: ebiten.KeyA, Ctrl: true}, "Ctrl+A"},
		{"shift+ctrl+A", ui.KeyBinding{Key: ebiten.KeyA, Ctrl: true, Shift: true}, "Ctrl+Shift+A"},
		{"gamepad", ui.KeyBinding{GamepadButton: 0, IsGamepad: true}, "GP:0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.binding.DisplayName()
			if got != tt.want {
				t.Errorf("DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKeybindInputSetBindingDoesNotFireCallback(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	kb := ui.NewKeybindInput("kb", font, 14)
	defer kb.Dispose()

	var called bool
	kb.SetOnBindingChanged(func(b ui.KeyBinding) {
		called = true
	})

	kb.SetBinding(ui.KeyBinding{Key: ebiten.KeyA})

	if called {
		t.Error("SetBinding should not fire OnBindingChanged")
	}
}
