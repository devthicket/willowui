// Ranges demonstrates WillowUI's range controls: slider, scrollbar, progress
// bar, and meter bar with reactive bindings, drag interaction, and visual states.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 600
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 24.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI: Range Controls Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	// ── Slider ───────────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Slider (0–100, step 1)", 24, 48)

	sliderVal := ui.NewRef(0.0)
	sliderLabel := ui.NewLabel("slider-label", "Value: 0", font, sizeMedium)
	sliderLabel.SetPosition(300, 68)
	screen.Add(sliderLabel)

	slider := ui.NewSlider("slider")
	slider.SetRange(0, 100)
	slider.SetStep(1)
	slider.SetSize(250, 20)
	slider.SetOnChange(func(v float64) {
		sliderLabel.SetText(fmt.Sprintf("Value: %.0f", v))
		sliderVal.Set(v)
	})
	slider.SetPosition(40, 66)
	screen.Add(slider)

	// ── Vertical Slider ──────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Vertical Slider", 600, 48)

	vSlider := ui.NewSlider("v-slider")
	vSlider.SetRange(0, 100)
	vSlider.SetStep(5)
	vSlider.SetOrientation(ui.Vertical)
	vSlider.SetSize(20, 150)
	vSlider.SetPosition(620, 66)
	screen.Add(vSlider)

	vSliderLabel := ui.NewLabel("v-slider-label", "0", font, sizeSmall)
	vSliderLabel.SetPosition(650, 130)
	screen.Add(vSliderLabel)

	vSlider.SetOnChange(func(v float64) {
		vSliderLabel.SetText(fmt.Sprintf("%.0f", v))
	})

	// ── Vertical ScrollBar ──────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Vertical ScrollBar", 700, 48)

	vSb := ui.NewScrollBar("v-scrollbar")
	vSb.SetSize(16, 150)
	vSb.SetContentSize(600, 150)
	vSb.SetPosition(720, 66)
	screen.Add(vSb)

	vSbLabel := ui.NewLabel("v-sb-label", "0", font, sizeSmall)
	vSbLabel.SetPosition(745, 130)
	screen.Add(vSbLabel)

	vSb.SetOnChange(func(v float64) {
		vSbLabel.SetText(fmt.Sprintf("%.0f", v))
	})

	// ── ScrollBar ────────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "ScrollBar", 24, 110)

	sbLabel := ui.NewLabel("sb-label", "Scroll: 0", font, sizeMedium)
	sbLabel.SetPosition(300, 130)
	screen.Add(sbLabel)

	sb := ui.NewScrollBar("scrollbar")
	sb.SetOrientation(ui.Horizontal)
	sb.SetSize(250, 16)
	sb.SetContentSize(500, 250)
	sb.SetOnChange(func(v float64) {
		sbLabel.SetText(fmt.Sprintf("Scroll: %.0f", v))
	})
	sb.SetPosition(40, 130)
	screen.Add(sb)

	// ── Progress Bar ─────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Progress Bar", 24, 170)

	progress := ui.NewProgressBar("progress")
	progress.SetSize(300, 24)
	progress.SetShowLabel(true, font, sizeSmall)
	progress.SetPosition(40, 190)
	screen.Add(progress)

	// Bind progress to slider value (0-100 -> 0-1).
	ui.WatchEffect(func() {
		progress.SetValue(sliderVal.Get() / 100.0)
	})

	// ── Meter Bar ────────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Meter Bar (0–200)", 24, 230)

	meter := ui.NewMeterBar("meter")
	meter.SetRange(0, 200)
	meter.SetValue(75)
	meter.SetSize(300, 24)
	meter.SetShowLabel(true, font, sizeSmall)
	meter.SetPosition(40, 250)
	screen.Add(meter)

	// ── Disabled Slider ──────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Disabled Slider", 24, 300)

	disSlider := ui.NewSlider("dis-slider")
	disSlider.SetRange(0, 100)
	disSlider.SetValue(40)
	disSlider.SetSize(200, 20)
	disSlider.SetEnabled(false)
	disSlider.SetPosition(40, 320)
	screen.Add(disSlider)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Range Controls Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addSectionLabel(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("section", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
