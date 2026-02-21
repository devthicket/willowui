package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func TestNewNumberStepperDefaults(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	defer ns.Dispose()

	if ns.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if ns.Name() != "ns" {
		t.Errorf("Name() = %q, want %q", ns.Name(), "ns")
	}
	if ns.Value() != 0 {
		t.Errorf("Value() = %f, want 0", ns.Value())
	}
	if ns.DecrementButton() == nil {
		t.Fatal("DecrementButton() should not be nil")
	}
	if ns.IncrementButton() == nil {
		t.Fatal("IncrementButton() should not be nil")
	}
	if ns.InputField() == nil {
		t.Fatal("InputField() should not be nil")
	}
}

func TestNumberStepperSetValue(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	defer ns.Dispose()

	ns.SetValue(42)
	if ns.Value() != 42 {
		t.Errorf("Value() = %f, want 42", ns.Value())
	}

	ns.SetValue(-10)
	if ns.Value() != -10 {
		t.Errorf("Value() = %f, want -10", ns.Value())
	}
}

func TestNumberStepperMinMax(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	defer ns.Dispose()

	ns.SetMin(0)
	ns.SetMax(100)

	ns.SetValue(-5)
	if ns.Value() != 0 {
		t.Errorf("Value() = %f, want 0 (clamped to min)", ns.Value())
	}

	ns.SetValue(150)
	if ns.Value() != 100 {
		t.Errorf("Value() = %f, want 100 (clamped to max)", ns.Value())
	}
}

func TestNumberStepperStepDownViaSetValue(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	defer ns.Dispose()

	ns.SetMin(0)
	ns.SetMax(100)
	ns.SetStep(5)
	ns.SetValue(20)

	// The decrement button calls SetValue(value - step) — verify that directly.
	ns.SetValue(ns.Value() - 5)
	if ns.Value() != 15 {
		t.Errorf("Value() = %f, want 15 after step-down", ns.Value())
	}
}

func TestNumberStepperStepUpViaSetValue(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	defer ns.Dispose()

	ns.SetMin(0)
	ns.SetMax(100)
	ns.SetStep(5)
	ns.SetValue(20)

	// The increment button calls SetValue(value + step) — verify that directly.
	ns.SetValue(ns.Value() + 5)
	if ns.Value() != 25 {
		t.Errorf("Value() = %f, want 25 after step-up", ns.Value())
	}
}

func TestNumberStepperStepDownClampsToMin(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	defer ns.Dispose()

	ns.SetMin(0)
	ns.SetMax(100)
	ns.SetStep(10)
	ns.SetValue(5)

	// Going below min should clamp.
	ns.SetValue(ns.Value() - 10)
	if ns.Value() != 0 {
		t.Errorf("Value() = %f, want 0 (clamped to min)", ns.Value())
	}
}

func TestNumberStepperSetOnChangeFires(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	defer ns.Dispose()

	var called int
	var lastVal float64
	ns.SetOnChange(func(v float64) {
		called++
		lastVal = v
	})

	ns.SetValue(7)
	if called != 1 {
		t.Errorf("onChange called %d times, want 1", called)
	}
	if lastVal != 7 {
		t.Errorf("onChange lastVal = %f, want 7", lastVal)
	}
}

func TestNumberStepperBindValue(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	defer ns.Dispose()

	ref := ui.NewRef(42.0)
	ns.BindValue(ref)

	if ns.Value() != 42 {
		t.Errorf("Value() = %f, want 42 after BindValue", ns.Value())
	}

	ref.Set(99.0)
	ui.DefaultScheduler.Flush()
	if ns.Value() != 99 {
		t.Errorf("Value() = %f, want 99 after ref.Set(99)", ns.Value())
	}
}

func TestNumberStepperSetDecimals(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	defer ns.Dispose()

	// Just ensure no panic and value unchanged.
	ns.SetDecimals(2)
	ns.SetValue(3.14159)
	if ns.Value() != 3.14159 {
		t.Errorf("Value() = %f, want 3.14159", ns.Value())
	}
}

func TestNumberStepperSetEnabled(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	defer ns.Dispose()

	ns.SetEnabled(false)
	if ns.IsEnabled() {
		t.Error("IsEnabled() should be false after SetEnabled(false)")
	}
}

func TestNumberStepperSetSize(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	defer ns.Dispose()

	ns.SetSize(180, 32)
	if ns.Width != 180 {
		t.Errorf("Width = %f, want 180", ns.Width)
	}
	if ns.Height != 32 {
		t.Errorf("Height = %f, want 32", ns.Height)
	}
}

func TestNumberStepperDispose(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	ns := ui.NewNumberStepper("ns", font, 16)
	ns.Dispose()
	if !ns.IsDisposed() {
		t.Error("IsDisposed() should be true after Dispose()")
	}
}
