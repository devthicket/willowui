// Masked Input — demo.
// Shows MaskedInput with various mask patterns: date, invite code, serial,
// plus formatted/raw value readout and complete/incomplete status labels.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 720
	screenH = 560
	colX    = 40.0
	labelX  = 300.0
	inputW  = 220.0
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 14.0
	)

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "Masked Input", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	y := 70.0
	gap := 70.0

	// ── Helper to add a section header ────────────────────────────────────
	addLabel := func(text string, x, y, size float64, alpha float64) *willow.Node {
		n := willow.NewText(text, text, font)
		n.TextBlock.FontSize = size
		n.TextBlock.Color = willow.RGBA(1, 1, 1, alpha)
		n.SetPosition(x, y)
		screen.AddNode(n)
		return n
	}

	addReadout := func(name string, x, y float64) func(v string) {
		lbl := willow.NewText(name+"-out", "", font)
		lbl.TextBlock.FontSize = sizeSmall
		lbl.TextBlock.Color = willow.RGBA(0.7, 0.9, 1.0, 1)
		lbl.SetPosition(x, y)
		screen.AddNode(lbl)
		return func(v string) {
			lbl.SetContent(v)
		}
	}

	addStatus := func(name string, x, y float64) func(complete bool) {
		lbl := willow.NewText(name+"-status", "○ incomplete", font)
		lbl.TextBlock.FontSize = sizeSmall
		lbl.TextBlock.Color = willow.RGBA(0.7, 0.7, 0.7, 1)
		lbl.SetPosition(x, y)
		screen.AddNode(lbl)
		return func(complete bool) {
			if complete {
				lbl.SetContent("● complete")
				lbl.SetTextColor(willow.RGBA(0.4, 1.0, 0.5, 1))
			} else {
				lbl.SetContent("○ incomplete")
				lbl.SetTextColor(willow.RGBA(0.7, 0.7, 0.7, 1))
			}
		}
	}

	// ── 1. Date field — 99/99/9999 ────────────────────────────────────────
	addLabel("Date   99/99/9999", colX, y-18, sizeSmall, 0.6)

	dateFormatted := addReadout("date-fmt", labelX, y)
	dateStatus := addStatus("date", labelX, y+18)
	addLabel("formatted:", labelX, y-18, sizeSmall, 0.5)

	date := ui.NewMaskedInput("date", font, sizeMedium)
	date.SetMask("99/99/9999")
	date.SetMaskPlaceholder('_')
	date.SetPlaceholder("MM/DD/YYYY")
	date.SetWidth(inputW)
	date.SetPosition(colX, y)
	date.SetOnChange(func(v string) {
		dateFormatted(fmt.Sprintf("value: %q  raw: %q", v, date.RawValue()))
	})
	date.SetOnComplete(func(raw, formatted string) { dateStatus(true) })
	date.SetOnIncomplete(func(raw, formatted string) { dateStatus(false) })
	screen.Add(date)
	y += gap

	// ── 2. Invite code — AAAA-9999 ────────────────────────────────────────
	addLabel("Invite code   AAAA-9999", colX, y-18, sizeSmall, 0.6)

	codeFormatted := addReadout("code-fmt", labelX, y)
	codeStatus := addStatus("code", labelX, y+18)
	addLabel("formatted:", labelX, y-18, sizeSmall, 0.5)

	code := ui.NewMaskedInput("code", font, sizeMedium)
	code.SetMask("AAAA-9999")
	code.SetMaskPlaceholder('_')
	code.SetPlaceholder("CODE-1234")
	code.SetWidth(inputW)
	code.SetPosition(colX, y)
	code.SetOnChange(func(v string) {
		codeFormatted(fmt.Sprintf("value: %q  raw: %q", v, code.RawValue()))
	})
	code.SetOnComplete(func(raw, formatted string) { codeStatus(true) })
	code.SetOnIncomplete(func(raw, formatted string) { codeStatus(false) })
	screen.Add(code)
	y += gap

	// ── 3. Serial — XXX-XXX-XXX ───────────────────────────────────────────
	addLabel("Serial   XXX-XXX-XXX", colX, y-18, sizeSmall, 0.6)

	serialFormatted := addReadout("serial-fmt", labelX, y)
	serialStatus := addStatus("serial", labelX, y+18)
	addLabel("formatted:", labelX, y-18, sizeSmall, 0.5)

	serial := ui.NewMaskedInput("serial", font, sizeMedium)
	serial.SetMask("XXX-XXX-XXX")
	serial.SetMaskPlaceholder('_')
	serial.SetWidth(inputW)
	serial.SetPosition(colX, y)
	serial.SetOnChange(func(v string) {
		serialFormatted(fmt.Sprintf("value: %q  raw: %q", v, serial.RawValue()))
	})
	serial.SetOnComplete(func(raw, formatted string) { serialStatus(true) })
	serial.SetOnIncomplete(func(raw, formatted string) { serialStatus(false) })
	screen.Add(serial)
	y += gap

	// ── 4. Reactive binding — 99/99 (MM/YY) ──────────────────────────────
	addLabel("Reactive   99/99  (bind:rawValue)", colX, y-18, sizeSmall, 0.6)

	expiryRef := ui.NewRef("")
	reactiveOut := addReadout("reactive-out", labelX, y)
	addLabel("ref:", labelX, y-18, sizeSmall, 0.5)

	expiry := ui.NewMaskedInput("expiry", font, sizeMedium)
	expiry.SetMask("99/99")
	expiry.SetMaskPlaceholder('_')
	expiry.SetPlaceholder("MM/YY")
	expiry.SetWidth(inputW)
	expiry.SetPosition(colX, y)
	expiry.BindRawValue(expiryRef)
	ui.WatchValue(expiryRef, func(_, v string) {
		reactiveOut(fmt.Sprintf("ref = %q", v))
	})
	screen.Add(expiry)
	y += gap + 10

	// ── Legend ─────────────────────────────────────────────────────────────
	addLabel("Slot keys: 9=digit  a=letter  A=upper-letter  X=upper-alnum  *=any", colX, y, 13, 0.4)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Masked Input",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.09, 0.11, 1),
	})
}
