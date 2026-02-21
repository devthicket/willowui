package integration

import (
	"math"
	"testing"

	ui "github.com/devthicket/willowui"
)

func newTestStatWeb(t *testing.T) *ui.StatWeb {
	t.Helper()
	resetScheduler()
	font := newTestFont()
	return ui.NewStatWeb("sw", font, 11)
}

func TestStatWebSetAxesRendersNSpokes(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	sw.SetAxes([]ui.StatAxis{
		{Name: "A", Min: 0, Max: 100, Value: 50},
		{Name: "B", Min: 0, Max: 100, Value: 50},
		{Name: "C", Min: 0, Max: 100, Value: 50},
		{Name: "D", Min: 0, Max: 100, Value: 50},
	})

	axes := sw.Axes()
	if len(axes) != 4 {
		t.Fatalf("Axes() len = %d, want 4", len(axes))
	}
}

func TestStatWebMaxEightAxes(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	axes := make([]ui.StatAxis, 10)
	for i := range axes {
		axes[i] = ui.StatAxis{Name: "X", Min: 0, Max: 100, Value: 50}
	}
	sw.SetAxes(axes)

	if len(sw.Axes()) != 8 {
		t.Errorf("Axes() len = %d, want 8 (max)", len(sw.Axes()))
	}
}

func TestStatWebSetValuesUpdates(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	sw.SetAxes([]ui.StatAxis{
		{Name: "A", Min: 0, Max: 100, Value: 50},
		{Name: "B", Min: 0, Max: 100, Value: 50},
		{Name: "C", Min: 0, Max: 100, Value: 50},
	})

	sw.SetValues([]float64{10, 20, 30})
	vals := sw.Values()
	if vals[0] != 10 || vals[1] != 20 || vals[2] != 30 {
		t.Errorf("Values() = %v, want [10 20 30]", vals)
	}
}

func TestStatWebSetValueClamps(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	sw.SetAxes([]ui.StatAxis{
		{Name: "A", Min: 10, Max: 90, Value: 50},
		{Name: "B", Min: 0, Max: 100, Value: 50},
		{Name: "C", Min: 0, Max: 100, Value: 50},
	})

	sw.SetValue(0, 200)
	if sw.Value(0) != 90 {
		t.Errorf("Value(0) = %f, want 90 (clamped to max)", sw.Value(0))
	}

	sw.SetValue(0, -5)
	if sw.Value(0) != 10 {
		t.Errorf("Value(0) = %f, want 10 (clamped to min)", sw.Value(0))
	}
}

func TestStatWebReadOnlyHasNoVisibleHandles(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	sw.SetAxes([]ui.StatAxis{
		{Name: "A", Min: 0, Max: 100, Value: 50},
		{Name: "B", Min: 0, Max: 100, Value: 50},
		{Name: "C", Min: 0, Max: 100, Value: 50},
	})

	if sw.IsEditable() {
		t.Error("new StatWeb should not be editable")
	}
}

func TestStatWebEditableShowsHandles(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	sw.SetAxes([]ui.StatAxis{
		{Name: "A", Min: 0, Max: 100, Value: 50},
		{Name: "B", Min: 0, Max: 100, Value: 50},
		{Name: "C", Min: 0, Max: 100, Value: 50},
	})

	sw.SetEditable(true)
	if !sw.IsEditable() {
		t.Error("expected IsEditable() = true after SetEditable(true)")
	}
}

func TestStatWebOnValueChangedFires(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	sw.SetAxes([]ui.StatAxis{
		{Name: "A", Min: 0, Max: 100, Value: 50},
		{Name: "B", Min: 0, Max: 100, Value: 50},
		{Name: "C", Min: 0, Max: 100, Value: 50},
	})

	var cbIndex int
	var cbValue float64
	sw.SetOnValueChanged(func(index int, value float64) {
		cbIndex = index
		cbValue = value
	})

	sw.SetValue(1, 75)
	if cbIndex != 1 {
		t.Errorf("callback index = %d, want 1", cbIndex)
	}
	if cbValue != 75 {
		t.Errorf("callback value = %f, want 75", cbValue)
	}
}

func TestStatWebOnValueChangedDoesNotFireWhenUnchanged(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	sw.SetAxes([]ui.StatAxis{
		{Name: "A", Min: 0, Max: 100, Value: 50},
		{Name: "B", Min: 0, Max: 100, Value: 50},
		{Name: "C", Min: 0, Max: 100, Value: 50},
	})

	fired := false
	sw.SetOnValueChanged(func(int, float64) { fired = true })

	sw.SetValue(0, 50) // same value
	if fired {
		t.Error("callback should not fire when value unchanged")
	}
}

func TestStatWebAxesReturnsCopy(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	sw.SetAxes([]ui.StatAxis{
		{Name: "A", Min: 0, Max: 100, Value: 50},
		{Name: "B", Min: 0, Max: 100, Value: 50},
		{Name: "C", Min: 0, Max: 100, Value: 50},
	})

	axes := sw.Axes()
	axes[0].Value = 999
	if sw.Value(0) != 50 {
		t.Error("Axes() should return a copy, not a reference")
	}
}

func TestStatWebSetSizeUpdatesWidthHeight(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	sw.SetSize(400, 400)
	if sw.Width != 400 || sw.Height != 400 {
		t.Errorf("size = (%f, %f), want (400, 400)", sw.Width, sw.Height)
	}
}

func TestStatWebValueOutOfBoundsIndex(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	sw.SetAxes([]ui.StatAxis{
		{Name: "A", Min: 0, Max: 100, Value: 50},
		{Name: "B", Min: 0, Max: 100, Value: 50},
		{Name: "C", Min: 0, Max: 100, Value: 50},
	})

	// Should not panic.
	sw.SetValue(-1, 100)
	sw.SetValue(10, 100)
	if v := sw.Value(-1); v != 0 {
		t.Errorf("Value(-1) = %f, want 0", v)
	}
	if v := sw.Value(10); v != 0 {
		t.Errorf("Value(10) = %f, want 0", v)
	}
}

func TestStatWebWeightedAxesNormalize(t *testing.T) {
	sw := newTestStatWeb(t)
	defer sw.Dispose()

	sw.SetAxes([]ui.StatAxis{
		{Name: "HP", Min: 0, Max: 999, Value: 500},
		{Name: "MP", Min: 0, Max: 500, Value: 250},
		{Name: "ATK", Min: 0, Max: 200, Value: 100},
	})

	// All should be at 50% — verify values stored correctly.
	vals := sw.Values()
	if math.Abs(vals[0]-500) > 0.01 {
		t.Errorf("HP value = %f, want 500", vals[0])
	}
	if math.Abs(vals[1]-250) > 0.01 {
		t.Errorf("MP value = %f, want 250", vals[1])
	}
	if math.Abs(vals[2]-100) > 0.01 {
		t.Errorf("ATK value = %f, want 100", vals[2])
	}
}
