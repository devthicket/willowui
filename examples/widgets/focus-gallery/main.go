// Focus Gallery demonstrates WillowUI's focus control system: Tab cycling,
// spatial arrow-key navigation, keyboard activation, and hotkey dispatch
// across a mixed VBox / HBox / absolute layout.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
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
	statusRef := ui.NewRef("Press Tab to cycle focus, arrow keys to navigate")

	// ── Title ────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI: Focus Gallery", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW-48)
	div.SetPosition(24, 48)
	screen.AddNode(div)

	// ── Status label (shows last action) ─────────────────────────────────────
	statusLabel := ui.NewLabel("status", "", font, sizeSmall)
	statusLabel.BindText(statusRef)
	statusLabel.SetColor(willow.RGBA(0.4, 0.8, 1.0, 1))
	statusLabel.SetPosition(24, screenH-30)
	screen.Add(statusLabel)

	// ── Section 1: Buttons in a column (Tab cycling) ─────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Buttons: Tab cycles, Space/Enter activates", 24, 58)

	for i := range 4 {
		btn := ui.NewButton(fmt.Sprintf("vbtn-%d", i), fmt.Sprintf("Button %d", i+1), font, sizeMedium)
		btn.SetSize(180, 36)
		btn.SetPosition(40, 80+float64(i)*44)
		btn.SetOnClick(func() {
			statusRef.Set(fmt.Sprintf("Clicked Button %d", i+1))
		})
		screen.Add(btn)
	}

	// ── Section 2: Controls in a row (spatial navigation) ────────────────────
	addSectionLabel(screen, font, sizeSmall, "Controls: Arrow keys navigate spatially", 280, 58)

	chk := ui.NewCheckbox("chk", "Check me", font, sizeMedium)
	chk.SetPosition(296, 80)
	chk.SetOnChange(func(v bool) {
		statusRef.Set(fmt.Sprintf("Checkbox toggled: %v", v))
	})
	screen.Add(chk)

	tgl := ui.NewToggle("tgl")
	tgl.SetPosition(440, 80)
	tgl.SetOnChange(func(v bool) {
		statusRef.Set(fmt.Sprintf("Toggle switched: %v", v))
	})
	screen.Add(tgl)

	accentBtn := ui.NewButton("accent-btn", "Accent", font, sizeMedium)
	accentBtn.SetSize(100, 36)
	accentBtn.SetVariant(ui.Accent)
	accentBtn.SetPosition(540, 80)
	accentBtn.SetOnClick(func() {
		statusRef.Set("Clicked Accent Button")
	})
	screen.Add(accentBtn)

	// ── Section 3: Text inputs (boundary escape) ─────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Text inputs: arrows escape at cursor boundaries", 24, 260)

	ti1 := ui.NewTextInput("input-1", font, sizeMedium)
	ti1.SetSize(220, 36)
	ti1.SetPlaceholder("First name...")
	ti1.SetPosition(40, 284)
	screen.Add(ti1)

	ti2 := ui.NewTextInput("input-2", font, sizeMedium)
	ti2.SetSize(220, 36)
	ti2.SetPlaceholder("Last name...")
	ti2.SetPosition(280, 284)
	screen.Add(ti2)

	ta := ui.NewTextArea("textarea", font, sizeMedium)
	ta.SetSize(460, 80)
	ta.SetPosition(40, 330)
	screen.Add(ta)

	// ── Section 4: Slider (arrow key interception) ───────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Slider: Left/Right adjust value, Up/Down navigate", 24, 430)

	slider := ui.NewSlider("slider")
	slider.SetSize(300, 24)
	slider.SetPosition(40, 454)
	screen.Add(slider)

	sliderLabel := ui.NewLabel("slider-val", "50%", font, sizeSmall)
	sliderLabel.SetPosition(360, 456)
	screen.Add(sliderLabel)

	slider.SetOnChange(func(v float64) {
		sliderLabel.SetText(fmt.Sprintf("%.0f%%", v*100))
		statusRef.Set(fmt.Sprintf("Slider: %.0f%%", v*100))
	})

	// ── Section 5: ToggleButtonBar ───────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "ToggleButtonBar: Left/Right cycle tabs, Up/Down navigate", 24, 500)

	tbb := ui.NewToggleButtonBar("tbb", font, sizeMedium)
	tbb.SetSize(400, 40)
	tbb.AddButton("Alpha")
	tbb.AddButton("Beta")
	tbb.AddButton("Gamma")
	tbb.SetPosition(40, 524)
	tbb.SetOnChange(func(idx int) {
		labels := []string{"Alpha", "Beta", "Gamma"}
		statusRef.Set(fmt.Sprintf("ToggleButtonBar: %s", labels[idx]))
	})
	screen.Add(tbb)

	// ── Hotkey: Ctrl+R resets status ─────────────────────────────────────────
	ui.FM.Bind(ui.Key(ebiten.KeyR, ui.ModCtrl), func() bool {
		statusRef.Set("Status reset via Ctrl+R hotkey")
		return true
	})

	// ── Hotkey: Escape clears focus ──────────────────────────────────────────
	ui.FM.Bind(ui.Key(ebiten.KeyEscape, ui.ModNone), func() bool {
		ui.FM.ClearFocus()
		statusRef.Set("Focus cleared via Escape")
		return true
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Focus Gallery",
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
