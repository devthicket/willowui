// KeybindInput demonstrates WillowUI's KeybindInput component for capturing
// keyboard and gamepad bindings in game settings screens. Shows multiple
// binding rows with labels, pre-set bindings, an unset binding, and a
// status label that reports captured keys.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 600
	screenH = 400
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 24.0
		sizeMedium = 16.0
		sizeSmall  = 14.0
	)

	screen := ui.NewScreen()

	// ── Title ──────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI -- KeybindInput Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW-48)
	div.SetPosition(24, 48)
	screen.AddNode(div)

	// ── Status label (shows last captured binding) ─────────────────────────
	statusRef := ui.NewRef("Last captured: (none)")
	statusLabel := ui.NewLabel("status", "", font, sizeMedium)
	statusLabel.BindText(statusRef)
	statusLabel.SetColor(willow.RGBA(0.3, 0.9, 0.5, 1))
	statusLabel.SetPosition(24, 340)
	screen.Add(statusLabel)

	// ── Binding rows ──────────────────────────────────────────────────────
	type bindRow struct {
		label   string
		name    string
		binding ui.KeyBinding
	}

	rows := []bindRow{
		{"Jump", "jump", ui.KeyBinding{Key: ebiten.KeySpace}},
		{"Move Left", "move-left", ui.KeyBinding{Key: ebiten.KeyA}},
		{"Move Right", "move-right", ui.KeyBinding{Key: ebiten.KeyD}},
		{"Attack", "attack", ui.KeyBinding{IsUnset: true}},
		{"Interact", "interact", ui.KeyBinding{Key: ebiten.KeyE}},
	}

	labelX := 40.0
	inputX := 180.0
	startY := 70.0
	rowH := 48.0

	for i, row := range rows {
		y := startY + float64(i)*rowH

		// Row label.
		lbl := ui.NewLabel(row.name+"-lbl", row.label, font, sizeMedium)
		lbl.SetColor(willow.RGBA(0.8, 0.8, 0.8, 1))
		lbl.SetPosition(labelX, y+8)
		screen.Add(lbl)

		// KeybindInput.
		kb := ui.NewKeybindInput(row.name, font, sizeSmall)
		kb.SetSize(200, 32)
		kb.SetBinding(row.binding)
		kb.SetPosition(inputX, y)
		screen.Add(kb)

		// Capture binding changes.
		actionName := row.label
		kb.SetOnBindingChanged(func(b ui.KeyBinding) {
			if b.IsUnset {
				statusRef.Set(fmt.Sprintf("Last captured: %s -> (cleared)", actionName))
			} else {
				statusRef.Set(fmt.Sprintf("Last captured: %s -> %s", actionName, b.DisplayName()))
			}
		})
	}

	// ── Section 2: Disabled binding ────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Disabled binding", 400, 70)

	disabledKb := ui.NewKeybindInput("disabled", font, sizeSmall)
	disabledKb.SetSize(160, 32)
	disabledKb.SetBinding(ui.KeyBinding{Key: ebiten.KeyF1})
	disabledKb.SetEnabled(false)
	disabledKb.SetPosition(400, 94)
	screen.Add(disabledKb)

	// ── Section 3: Combos disabled ─────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Single-key only (no combos)", 400, 150)

	singleKb := ui.NewKeybindInput("single-key", font, sizeSmall)
	singleKb.SetSize(160, 32)
	singleKb.SetCombosEnabled(false)
	singleKb.SetBinding(ui.KeyBinding{Key: ebiten.KeyTab})
	singleKb.SetPosition(400, 174)
	singleKb.SetOnBindingChanged(func(b ui.KeyBinding) {
		statusRef.Set(fmt.Sprintf("Last captured: Single-key -> %s", b.DisplayName()))
	})
	screen.Add(singleKb)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI -- KeybindInput Demo",
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
