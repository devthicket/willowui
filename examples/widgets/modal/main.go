// Modal demonstrates WillowUI's modal window mode: a confirmation dialog with
// a darkened backdrop overlay that blocks all input to the content behind it.
// Toggles and RGBA sliders let you experiment with the overlay appearance live.
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
	title := willow.NewText("title", "WillowUI: Modal Dialog Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW-48)
	div.SetPosition(24, 50)
	screen.AddNode(div)

	// ── Background interactive content ───────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall,
		"Background content -- clicking is blocked while the modal is open", 24, 62)

	counts := [3]int{}
	countRefs := [3]*ui.Ref[string]{
		ui.NewRef("Item A: 0"),
		ui.NewRef("Item B: 0"),
		ui.NewRef("Item C: 0"),
	}
	for i := 0; i < 3; i++ {
		idx := i
		btn := ui.NewButton(fmt.Sprintf("bg-btn-%d", i), "", font, sizeMedium)
		btn.BindText(countRefs[idx])
		btn.SetSize(160, 40)
		btn.SetPosition(40+float64(i)*180, 82)
		btn.SetOnClick(func() {
			counts[idx]++
			countRefs[idx].Set(fmt.Sprintf("Item %c: %d", 'A'+idx, counts[idx]))
		})
		screen.Add(btn)
	}

	addSectionLabel(screen, font, sizeSmall,
		"Each button counts clicks -- they should not increment while modal is open", 24, 134)

	// ── Open dialog button ───────────────────────────────────────────────────
	openBtn := ui.NewButton("open-btn", "Open Confirmation Dialog", font, sizeMedium)
	openBtn.SetSize(260, 44)
	openBtn.SetVariant(ui.Accent)
	openBtn.SetPosition((screenW-260)/2, 156)
	screen.Add(openBtn)

	// ── Status label ─────────────────────────────────────────────────────────
	statusRef := ui.NewRef("Status: Ready")
	statusLabel := ui.NewLabel("status", "", font, sizeSmall)
	statusLabel.BindText(statusRef)
	statusLabel.SetColor(willow.RGBA(0.55, 0.8, 0.55, 1))
	statusLabel.SetPosition(24, screenH-24)
	screen.Add(statusLabel)

	// ── Modal dialog window ───────────────────────────────────────────────────
	const (
		modalW = 360.0
		modalH = 200.0
	)
	modal := ui.NewWindow("confirm-modal", "Confirm Action", font, sizeMedium)
	modal.SetModal(true)
	modal.SetMovable(false)
	modal.SetResizable(false)
	modal.SetCloseable(false)
	modal.SetSize(modalW, modalH)
	modal.SetPosition((screenW-modalW)/2, (screenH-modalH)/2)
	screen.Add(modal)
	ui.DefaultWindowManager.Add(modal)
	modal.SetVisible(false)

	// overlayPreview is assigned later after the stepper rows are built.
	var overlayPreview *willow.Node

	// showResult is called by the modal's result handler for every dismiss path.
	showResult := func(reason string) {
		if overlayPreview != nil {
			overlayPreview.SetVisible(true)
		}
		statusRef.Set("Status: " + reason)
	}

	// Body: message + Cancel / Confirm (manual positions via LayoutNone).
	body := modal.Body()
	body.SetLayout(ui.LayoutNone)

	msgLabel := ui.NewLabel("msg",
		"Are you sure you want to proceed?\nThis action cannot be undone.",
		font, sizeSmall)
	msgLabel.SetColor(willow.RGBA(0.9, 0.9, 0.9, 1))
	msgLabel.SetPosition(8, 8)
	body.AddChild(msgLabel)

	cancelBtn := ui.NewButton("cancel-btn", "Cancel", font, sizeSmall)
	cancelBtn.SetSize(100, 36)
	cancelBtn.SetVariant(ui.Neutral)
	cancelBtn.SetPosition(76, 80)
	cancelBtn.SetOnClick(func() { modal.FireResult("cancel", nil) })
	body.AddChild(cancelBtn)

	confirmBtn := ui.NewButton("confirm-btn", "Confirm", font, sizeSmall)
	confirmBtn.SetSize(100, 36)
	confirmBtn.SetVariant(ui.Danger)
	confirmBtn.SetPosition(184, 80)
	confirmBtn.SetOnClick(func() { modal.FireResult("confirm", nil) })
	body.AddChild(confirmBtn)

	body.UpdateLayout()

	// Toggle state — kept in sync by the toggle handlers below.
	// Initial values mirror the toggle defaults so openBtn applies them correctly.
	backdropEnabled := true
	escEnabled := true
	enterEnabled := true

	// Wire the result handler and apply current toggle settings before each open.
	openBtn.SetOnClick(func() {
		modal.SetOnResult(func(key string, _ any) {
			switch key {
			case "confirm":
				showResult("Action confirmed!")
			case "cancel":
				showResult("Action cancelled")
			case "backdrop":
				showResult("Dismissed (backdrop click)")
			case "esc":
				showResult("Closed (Esc)")
			case "enter":
				showResult("Confirmed (Enter)")
			}
		})
		if backdropEnabled {
			modal.SetOnModalOverlayClick(func() { modal.FireResult("backdrop", nil) })
		} else {
			modal.SetOnModalOverlayClick(nil)
		}
		if escEnabled {
			modal.SetEscResult("esc")
		} else {
			modal.SetEscResult("")
		}
		if enterEnabled {
			modal.SetEnterResult("enter")
		} else {
			modal.SetEnterResult("")
		}
		modal.SetVisible(true)
		if overlayPreview != nil {
			overlayPreview.SetVisible(false)
		}
	})

	// ── Dialog options ────────────────────────────────────────────────────────
	div2 := ui.NewDivider("divider-2", screenW-48)
	div2.SetPosition(24, 218)
	screen.AddNode(div2)
	addSectionLabel(screen, font, sizeSmall, "Dialog Options", 24, 228)

	const (
		// Toggles — 2 columns × 3 rows.
		tglLabelXL = 40.0
		tglXL      = 340.0
		tglLabelXR = 440.0
		tglXR      = 740.0
		rowY0      = 248.0
		rowStep    = 36.0
	)

	// Overlay color state — shared by the dim toggle and RGBA steppers.
	isDim := true
	ovR, ovG, ovB, ovA := 0.0, 0.0, 0.0, 0.5

	// applyOverlayColor pushes the current RGBA (or transparent) to the overlay
	// and refreshes the preview swatch.
	// When dim is off, alpha 0.001 is visually transparent but avoids the
	// all-zero white rendering fallback while still blocking input.
	var applyOverlayColor func()
	applyOverlayColor = func() {
		if isDim {
			modal.SetModalOverlayColor(willow.RGBA(ovR, ovG, ovB, ovA))
		} else {
			modal.SetModalOverlayColor(willow.RGBA(0, 0, 0, 0.001))
		}
		if overlayPreview != nil {
			overlayPreview.SetColor(willow.RGBA(ovR, ovG, ovB, ovA))
		}
	}

	// addToggleRow places a label + toggle at the given column (0=left, 1=right)
	// and row index within the 2-column grid.
	addToggleRow := func(name, label string, initial bool, col, row int, onChange func(bool)) *ui.Toggle {
		labelX := tglLabelXL
		tglX := tglXL
		if col == 1 {
			labelX, tglX = tglLabelXR, tglXR
		}
		y := rowY0 + float64(row)*rowStep

		n := willow.NewText(name+"-lbl", label, font)
		n.TextBlock.FontSize = sizeSmall
		n.TextBlock.Color = willow.RGBA(0.85, 0.85, 0.85, 1)
		n.SetPosition(labelX, y+4)
		screen.AddNode(n)

		tgl := ui.NewToggle(name)
		tgl.SetValue(initial)
		tgl.SetOnChange(onChange)
		tgl.SetPosition(tglX, y)
		screen.Add(tgl)
		return tgl
	}

	//  col 0 (left)                              col 1 (right)
	addToggleRow("tgl-dim", "Dim overlay", true, 0, 0, func(v bool) {
		isDim = v
		applyOverlayColor()
	})
	addToggleRow("tgl-backdrop", "Close on backdrop click", true, 1, 0, func(v bool) {
		backdropEnabled = v
		if v {
			modal.SetOnModalOverlayClick(func() { modal.FireResult("backdrop", nil) })
		} else {
			modal.SetOnModalOverlayClick(nil)
		}
	})
	addToggleRow("tgl-close", "Show close button", false, 0, 1, func(v bool) {
		modal.SetCloseable(v)
		if v {
			modal.SetOnClose(func() { modal.FireResult("cancel", nil) })
		} else {
			modal.SetOnClose(nil)
		}
	})
	addToggleRow("tgl-move", "Movable", false, 1, 1, func(v bool) {
		modal.SetMovable(v)
	})
	addToggleRow("tgl-esc", "Esc to close", true, 0, 2, func(v bool) {
		escEnabled = v
		if v {
			modal.SetEscResult("esc")
		} else {
			modal.SetEscResult("")
		}
	})
	addToggleRow("tgl-enter", "Enter to confirm", true, 1, 2, func(v bool) {
		enterEnabled = v
		if v {
			modal.SetEnterResult("enter")
		} else {
			modal.SetEnterResult("")
		}
	})

	// ── Overlay Color (RGBA steppers — 2 columns × 2 rows) ───────────────────
	const colorSectionY = rowY0 + 3*rowStep + 8 // 364
	div3 := ui.NewDivider("divider-3", screenW-48)
	div3.SetPosition(24, colorSectionY)
	screen.AddNode(div3)
	addSectionLabel(screen, font, sizeSmall, "Overlay Color (RGBA, 0–1)", 24, colorSectionY+10)

	const (
		stepperLabelX = 40.0
		stepperX      = 58.0
		stepperW      = 190.0
		stepperH      = 28.0
		colorRowY0    = colorSectionY + 30
		colorRowH     = 36.0
	)

	addStepperRow := func(name, label string, initial float64, rowIdx int, set func(float64)) {
		y := colorRowY0 + float64(rowIdx)*colorRowH

		lbl := willow.NewText(name+"-lbl", label, font)
		lbl.TextBlock.FontSize = sizeSmall
		lbl.TextBlock.Color = willow.RGBA(0.85, 0.85, 0.85, 1)
		lbl.SetPosition(stepperLabelX, y+8)
		screen.AddNode(lbl)

		ns := ui.NewNumberStepper(name, font, sizeSmall)
		ns.SetSize(stepperW, stepperH)
		ns.SetMin(0)
		ns.SetMax(1)
		ns.SetStep(0.05)
		ns.SetDecimals(2)
		ns.SetValue(initial)
		ns.SetOnChange(func(v float64) {
			set(v)
		})
		ns.SetPosition(stepperX, y)
		screen.Add(ns)
	}

	addStepperRow("ovr-r", "R", ovR, 0, func(v float64) { ovR = v; applyOverlayColor() })
	addStepperRow("ovr-g", "G", ovG, 1, func(v float64) { ovG = v; applyOverlayColor() })
	addStepperRow("ovr-b", "B", ovB, 2, func(v float64) { ovB = v; applyOverlayColor() })
	addStepperRow("ovr-a", "A", ovA, 3, func(v float64) { ovA = v; applyOverlayColor() })

	// 48×48 color swatch centred alongside the four stepper rows.
	const (
		swatchSize = 48.0
		swatchX    = stepperX + stepperW + 12
		swatchY    = colorRowY0 + (colorRowH*4-swatchSize)/2
	)
	overlayPreview = willow.NewSprite("overlay-preview", willow.TextureRegion{})
	overlayPreview.SetPosition(swatchX, swatchY)
	overlayPreview.SetScale(swatchSize, swatchSize)
	overlayPreview.SetColor(willow.RGBA(ovR, ovG, ovB, ovA))
	screen.AddNode(overlayPreview)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Modal Dialog Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addSectionLabel(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("section-"+text[:4], text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
