// Forms demonstrates WillowUI's form controls: toggle, checkbox, radio group,
// text input, and text area with reactive bindings and visual states.
package main

import (
	"fmt"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 600
)

func main() {
	font := ui.MustLoadDefaultFont()
	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	const (
		sizeLarge  = 24.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI: Form Controls Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	// ── Toggle ───────────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Toggle", 24, 48)

	toggleStatus := ui.NewRef("Toggle: OFF")
	toggleLabel := ui.NewLabel("tgl-label", "", font, sizeMedium)
	toggleLabel.BindText(toggleStatus)
	toggleLabel.SetPosition(100, 68)
	screen.Add(toggleLabel)

	tgl := ui.NewToggle("tgl")
	tgl.SetOnChange(func(v bool) {
		if v {
			toggleStatus.Set("Toggle: ON")
		} else {
			toggleStatus.Set("Toggle: OFF")
		}
	})
	tgl.SetPosition(40, 64)
	screen.Add(tgl)

	// ── Checkbox ─────────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Checkbox", 24, 100)

	cbStatus := ui.NewRef("Unchecked")
	cbLabel := ui.NewLabel("cb-status", "", font, sizeMedium)
	cbLabel.BindText(cbStatus)
	cbLabel.SetPosition(200, 118)
	screen.Add(cbLabel)

	cb := ui.NewCheckbox("cb", "Accept terms", font, sizeMedium)
	cb.SetOnChange(func(v bool) {
		if v {
			cbStatus.Set("Checked")
		} else {
			cbStatus.Set("Unchecked")
		}
	})
	cb.SetPosition(40, 116)
	screen.Add(cb)

	// ── Radio Group ──────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Radio Group", 24, 150)

	radioStatus := ui.NewRef("Selected: none")
	radioLabel := ui.NewLabel("radio-status", "", font, sizeMedium)
	radioLabel.BindText(radioStatus)
	radioLabel.SetPosition(200, 170)
	screen.Add(radioLabel)

	rg := ui.NewRadio("rg")
	rg.AddOption("Option A", font, sizeMedium)
	rg.AddOption("Option B", font, sizeMedium)
	rg.AddOption("Option C", font, sizeMedium)
	rg.SetOnChange(func(idx int) {
		radioStatus.Set(fmt.Sprintf("Selected: %d", idx))
	})
	rg.UpdateLayout()
	rg.SetPosition(40, 168)
	screen.Add(rg)

	// ── Text Input ───────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Text Input", 400, 48)

	tiStatus := ui.NewRef("Input: (empty)")
	tiLabel := ui.NewLabel("ti-status", "", font, sizeMedium)
	tiLabel.BindText(tiStatus)
	tiLabel.SetPosition(420, 100)
	screen.Add(tiLabel)

	ti := ui.NewTextInput("ti", font, sizeMedium)
	ti.SetPlaceholder("Enter text...")
	ti.SetWidth(250)
	ti.SetOnChange(func(v string) {
		tiStatus.Set(fmt.Sprintf("Input: %s", v))
	})
	ti.SetPosition(420, 64)
	screen.Add(ti)

	// ── Text Area ────────────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Text Area", 400, 130)

	taStatus := ui.NewRef("Lines: 0")
	taLabel := ui.NewLabel("ta-status", "", font, sizeMedium)
	taLabel.BindText(taStatus)
	taLabel.SetPosition(420, 300)
	screen.Add(taLabel)

	ta := ui.NewTextArea("ta", font, sizeMedium)
	ta.SetSize(300, 120)
	ta.SetOnChange(func(v string) {
		lines := 1
		for _, r := range v {
			if r == '\n' {
				lines++
			}
		}
		taStatus.Set(fmt.Sprintf("Lines: %d", lines))
	})
	ta.SetPosition(420, 148)
	screen.Add(ta)

	// ── Disabled variants ────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Disabled Controls", 24, 300)

	disToggle := ui.NewToggle("dis-tgl")
	disToggle.SetEnabled(false)
	disToggle.SetPosition(40, 320)
	screen.Add(disToggle)

	disCb := ui.NewCheckbox("dis-cb", "Disabled", font, sizeMedium)
	disCb.SetEnabled(false)
	disCb.SetPosition(100, 320)
	screen.Add(disCb)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Form Controls Demo",
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
