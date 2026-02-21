package main

import (
	"fmt"
	"time"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 600
)

func main() {
	font := ui.MustLoadDefaultFont()

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "WillowUI -- CalendarSelector Demo", font)
	title.TextBlock.FontSize = 22.0
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	// Status label updated on date selection.
	statusLabel := willow.NewText("status", "Click a date to select it", font)
	statusLabel.TextBlock.FontSize = 13.0
	statusLabel.TextBlock.Color = willow.RGBA(0.6, 0.65, 0.7, 1)
	statusLabel.SetPosition(40, 560)
	screen.AddNode(statusLabel)

	updateStatus := func(source string, t time.Time) {
		statusLabel.SetContent(fmt.Sprintf("%s: selected %s", source, t.Format("2006-01-02")))
	}

	// --- Inline calendar ---
	inlineLabel := willow.NewText("inline-lbl", "Inline Calendar", font)
	inlineLabel.TextBlock.FontSize = 14.0
	inlineLabel.TextBlock.Color = willow.RGBA(0.7, 0.75, 0.8, 1)
	inlineLabel.SetPosition(40, 52)
	screen.AddNode(inlineLabel)

	cal := ui.NewCalendarSelector("inline-cal", font, 13)
	cal.SetDate(time.Now())
	cal.SetSize(280, 300)
	cal.SetPosition(40, 74)
	cal.SetOnDateSelected(func(t time.Time) {
		updateStatus("Inline", t)
	})
	screen.Add(cal)

	// --- Calendar with min/max constraints ---
	constrainedLabel := willow.NewText("constrained-lbl", "Constrained (10th - 20th)", font)
	constrainedLabel.TextBlock.FontSize = 14.0
	constrainedLabel.TextBlock.Color = willow.RGBA(0.7, 0.75, 0.8, 1)
	constrainedLabel.SetPosition(360, 52)
	screen.AddNode(constrainedLabel)

	now := time.Now()
	minDate := time.Date(now.Year(), now.Month(), 10, 0, 0, 0, 0, time.Local)
	maxDate := time.Date(now.Year(), now.Month(), 20, 0, 0, 0, 0, time.Local)

	cal2 := ui.NewCalendarSelector("constrained-cal", font, 13)
	cal2.SetDate(time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, time.Local))
	cal2.SetMinDate(minDate)
	cal2.SetMaxDate(maxDate)
	cal2.SetSize(280, 300)
	cal2.SetPosition(360, 74)
	cal2.SetOnDateSelected(func(t time.Time) {
		updateStatus("Constrained", t)
	})
	screen.Add(cal2)

	// --- Popup calendar ---
	popupLabel := willow.NewText("popup-lbl", "Popup Calendar (click to open)", font)
	popupLabel.TextBlock.FontSize = 14.0
	popupLabel.TextBlock.Color = willow.RGBA(0.7, 0.75, 0.8, 1)
	popupLabel.SetPosition(40, 406)
	screen.AddNode(popupLabel)

	popupCal := ui.NewCalendarSelector("popup-cal", font, 13)
	popupCal.SetPopupMode(true)
	popupCal.SetSize(180, 34)
	popupCal.SetPosition(40, 430)
	popupCal.SetOnDateSelected(func(t time.Time) {
		updateStatus("Popup", t)
	})
	screen.Add(popupCal)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI -- CalendarSelector Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
