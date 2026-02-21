package integration

import (
	"testing"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/widget"
)

func resetTimePicker() {
	resetScheduler()
	widget.DefaultFocusManager = widget.NewFocusManager()
}

func newTestTimePicker(name string) *ui.TimePicker {
	return ui.NewTimePicker(name, newTestFont(), 16)
}

// clickButton simulates a click on a Button by invoking its OnClick handler.
func clickButton(btn *ui.Button) {
	btn.Node().GetOnClick()(willow.ClickContext{Node: btn.Node()})
}

func TestTimePickerSetTime24h(t *testing.T) {
	resetTimePicker()
	tp := newTestTimePicker("tp")
	defer tp.Dispose()

	tp.SetTime(14, 30, 0)
	if tp.Hour() != 14 {
		t.Errorf("Hour() = %d, want 14", tp.Hour())
	}
	if tp.Minute() != 30 {
		t.Errorf("Minute() = %d, want 30", tp.Minute())
	}
	if tp.Second() != 0 {
		t.Errorf("Second() = %d, want 0", tp.Second())
	}
}

func TestTimePickerOnTimeChangedFires(t *testing.T) {
	resetTimePicker()
	tp := newTestTimePicker("tp")
	defer tp.Dispose()

	tp.SetTime(14, 30, 0)

	var gotH, gotM, gotS int
	fired := false
	tp.SetOnTimeChanged(func(h, m, s int) {
		gotH, gotM, gotS = h, m, s
		fired = true
	})

	clickButton(tp.HourUpButton())
	if !fired {
		t.Fatal("OnTimeChanged should have fired")
	}
	if gotH != 15 {
		t.Errorf("hour = %d, want 15", gotH)
	}
	if gotM != 30 {
		t.Errorf("minute = %d, want 30", gotM)
	}
	if gotS != 0 {
		t.Errorf("second = %d, want 0", gotS)
	}
}

func TestTimePickerMinuteDecrementFires(t *testing.T) {
	resetTimePicker()
	tp := newTestTimePicker("tp")
	defer tp.Dispose()

	tp.SetTime(10, 15, 0)

	var gotM int
	tp.SetOnTimeChanged(func(h, m, s int) {
		gotM = m
	})

	clickButton(tp.MinuteDownButton())
	if gotM != 14 {
		t.Errorf("minute = %d, want 14", gotM)
	}
}

func TestTimePickerHourWraps24h(t *testing.T) {
	resetTimePicker()
	tp := newTestTimePicker("tp")
	defer tp.Dispose()

	tp.SetTime(23, 0, 0)

	var gotH int
	tp.SetOnTimeChanged(func(h, m, s int) {
		gotH = h
	})

	clickButton(tp.HourUpButton())
	if gotH != 0 {
		t.Errorf("hour should wrap from 23 to 0, got %d", gotH)
	}
}

func TestTimePickerMinuteWraps(t *testing.T) {
	resetTimePicker()
	tp := newTestTimePicker("tp")
	defer tp.Dispose()

	tp.SetTime(10, 59, 0)

	var gotM int
	tp.SetOnTimeChanged(func(h, m, s int) {
		gotM = m
	})

	clickButton(tp.MinuteUpButton())
	if gotM != 0 {
		t.Errorf("minute should wrap from 59 to 0, got %d", gotM)
	}
}

func TestTimePickerShowSeconds(t *testing.T) {
	resetTimePicker()
	tp := newTestTimePicker("tp")
	defer tp.Dispose()

	tp.SetShowSeconds(true)
	tp.SetTime(10, 30, 45)

	if tp.Second() != 45 {
		t.Errorf("Second() = %d, want 45", tp.Second())
	}

	var gotS int
	tp.SetOnTimeChanged(func(h, m, s int) {
		gotS = s
	})

	clickButton(tp.SecondUpButton())
	if gotS != 46 {
		t.Errorf("second = %d, want 46", gotS)
	}
}

func TestTimePickerFormat12h(t *testing.T) {
	resetTimePicker()
	tp := newTestTimePicker("tp")
	defer tp.Dispose()

	tp.SetFormat(ui.TimeFormat12h)
	tp.SetTime(15, 30, 0) // 3:30 PM

	if tp.Hour() != 15 {
		t.Errorf("Hour() should return 24h value, got %d", tp.Hour())
	}
}

func TestTimePickerHourWraps12h(t *testing.T) {
	resetTimePicker()
	tp := newTestTimePicker("tp")
	defer tp.Dispose()

	tp.SetFormat(ui.TimeFormat12h)
	tp.SetTime(11, 0, 0) // 11 AM, display=11

	var gotH int
	tp.SetOnTimeChanged(func(h, m, s int) {
		gotH = h
	})

	// 11 AM + 1 => 12 PM (crossing 11→12 toggles AM/PM). 24h=12.
	clickButton(tp.HourUpButton())
	if gotH != 12 {
		t.Errorf("hour should go from 11 AM to 12 PM (12 in 24h), got %d", gotH)
	}
}

func TestTimePickerAmPmToggle(t *testing.T) {
	resetTimePicker()
	tp := newTestTimePicker("tp")
	defer tp.Dispose()

	tp.SetFormat(ui.TimeFormat12h)
	tp.SetTime(8, 30, 0) // 8:30 AM

	var gotH int
	tp.SetOnTimeChanged(func(h, m, s int) {
		gotH = h
	})

	// Toggle AM -> PM.
	clickButton(tp.AmPmButton())
	if gotH != 20 {
		t.Errorf("toggling AM->PM for 8:30 should give hour=20, got %d", gotH)
	}
}

func TestTimePickerHourDown12hTogglesPeriod(t *testing.T) {
	resetTimePicker()
	tp := newTestTimePicker("tp")
	defer tp.Dispose()

	tp.SetFormat(ui.TimeFormat12h)
	tp.SetTime(12, 0, 0) // 12 PM (noon)

	var gotH int
	tp.SetOnTimeChanged(func(h, m, s int) {
		gotH = h
	})

	// 12 PM down => 11 AM (crossing 12→11 toggles PM→AM). 24h=11.
	clickButton(tp.HourDownButton())
	if gotH != 11 {
		t.Errorf("12 PM down should give 11 AM (11 in 24h), got %d", gotH)
	}
}

func TestTimePickerHour11PMUpGivesNoon(t *testing.T) {
	resetTimePicker()
	tp := newTestTimePicker("tp")
	defer tp.Dispose()

	tp.SetFormat(ui.TimeFormat12h)
	tp.SetTime(23, 0, 0) // 11 PM

	var gotH int
	tp.SetOnTimeChanged(func(h, m, s int) {
		gotH = h
	})

	// 11 PM + 1 => 12 AM (midnight). 24h=0.
	clickButton(tp.HourUpButton())
	if gotH != 0 {
		t.Errorf("11 PM up should give 12 AM (0 in 24h), got %d", gotH)
	}
}
