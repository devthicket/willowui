package integration

import (
	"testing"
	"time"

	ui "github.com/devthicket/willowui"
)

func TestCalendarSelectorCreation(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cal := ui.NewCalendarSelector("cal", font, 13)
	defer cal.Dispose()

	if cal.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if cal.Name() != "cal" {
		t.Errorf("Name() = %q, want %q", cal.Name(), "cal")
	}
}

func TestCalendarSelectorSetDate(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cal := ui.NewCalendarSelector("cal", font, 13)
	defer cal.Dispose()

	target := time.Date(2026, time.March, 15, 0, 0, 0, 0, time.Local)
	cal.SetDate(target)

	got := cal.Date()
	if got.Year() != 2026 || got.Month() != time.March || got.Day() != 15 {
		t.Errorf("Date() = %v, want 2026-03-15", got)
	}
}

func TestCalendarSelectorSetDateNavigatesMonth(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cal := ui.NewCalendarSelector("cal", font, 13)
	defer cal.Dispose()

	now := time.Now()
	cal.SetDate(now)

	// Set to a different month -- the grid should navigate.
	target := time.Date(2025, time.January, 10, 0, 0, 0, 0, time.Local)
	cal.SetDate(target)

	got := cal.Date()
	if got.Year() != 2025 || got.Month() != time.January || got.Day() != 10 {
		t.Errorf("Date() = %v, want 2025-01-10", got)
	}
}

func TestCalendarSelectorSetMonth(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cal := ui.NewCalendarSelector("cal", font, 13)
	defer cal.Dispose()

	cal.SetDate(time.Date(2026, time.March, 1, 0, 0, 0, 0, time.Local))
	cal.SetMonth(2026, 4) // navigate to April

	// Setting a date in April should work.
	cal.SetDate(time.Date(2026, time.April, 5, 0, 0, 0, 0, time.Local))
	if cal.Date().Month() != time.April || cal.Date().Day() != 5 {
		t.Errorf("expected April 5, got %v", cal.Date())
	}
}

func TestCalendarSelectorMinMaxDisablesDays(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cal := ui.NewCalendarSelector("cal", font, 13)
	defer cal.Dispose()

	cal.SetDate(time.Date(2026, time.March, 15, 0, 0, 0, 0, time.Local))
	cal.SetMinDate(time.Date(2026, time.March, 10, 0, 0, 0, 0, time.Local))
	cal.SetMaxDate(time.Date(2026, time.March, 20, 0, 0, 0, 0, time.Local))

	// The selected date (15) is within range, so it should remain.
	if cal.Date().Day() != 15 {
		t.Errorf("expected day 15 (within range), got %d", cal.Date().Day())
	}
}

func TestCalendarSelectorPopupMode(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cal := ui.NewCalendarSelector("cal", font, 13)
	defer cal.Dispose()

	cal.SetPopupMode(true)
	if !cal.IsPopupMode() {
		t.Error("expected popup mode to be true")
	}

	// Trigger button should be visible.
	if !cal.TriggerButton().Node().Visible() {
		t.Error("trigger button should be visible in popup mode")
	}
}

func TestCalendarSelectorClearMinMax(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cal := ui.NewCalendarSelector("cal", font, 13)
	defer cal.Dispose()

	cal.SetMinDate(time.Date(2026, time.March, 10, 0, 0, 0, 0, time.Local))
	cal.SetMaxDate(time.Date(2026, time.March, 20, 0, 0, 0, 0, time.Local))
	cal.ClearMinDate()
	cal.ClearMaxDate()

	// After clearing, all dates should be enabled.
	cal.SetDate(time.Date(2026, time.March, 5, 0, 0, 0, 0, time.Local))
	if cal.Date().Day() != 5 {
		t.Errorf("expected day 5, got %d", cal.Date().Day())
	}
}

func TestCalendarSelectorOnDateSelected(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cal := ui.NewCalendarSelector("cal", font, 13)
	defer cal.Dispose()

	var selected time.Time
	cal.SetOnDateSelected(func(d time.Time) {
		selected = d
	})

	// SetDate should NOT fire OnDateSelected (only user clicks should).
	cal.SetDate(time.Date(2026, time.June, 20, 0, 0, 0, 0, time.Local))
	if !selected.IsZero() {
		t.Error("SetDate should not fire OnDateSelected")
	}
}

func TestCalendarSelectorPrevNextButton(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cal := ui.NewCalendarSelector("cal", font, 13)
	defer cal.Dispose()

	if cal.PrevButton() == nil {
		t.Fatal("PrevButton should not be nil")
	}
	if cal.NextButton() == nil {
		t.Fatal("NextButton should not be nil")
	}
}

func TestCalendarSelectorEnabled(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	cal := ui.NewCalendarSelector("cal", font, 13)
	defer cal.Dispose()

	cal.SetEnabled(false)
	if cal.IsEnabled() {
		t.Error("expected disabled")
	}

	cal.SetEnabled(true)
	if !cal.IsEnabled() {
		t.Error("expected enabled")
	}
}
