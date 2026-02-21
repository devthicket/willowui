// Labels & Buttons - reactive demo.
// Shows Label.BindText, NewComputed driving a label, Button.OnClick
// updating a shared Ref[int] that two labels observe independently,
// and Button.BindText / BindEnabled / BindVisible.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 800
	screenH  = 520
	colLeft  = 40.0
	colRight = 450.0
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	title := willow.NewText("title", "Reactive - Labels & Buttons", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	y := 62.0

	// ── 1. Label + BindText ──────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Label.BindText - bound to a Ref[string]", colLeft, y)
	y += 20

	textRef := ui.NewRef("Hello, WillowUI!")

	lbl := ui.NewLabel("lbl", "", font, sizeMedium)
	lbl.SetColor(willow.RGBA(0.4, 0.9, 0.6, 1))
	lbl.BindText(textRef)
	lbl.SetPosition(colLeft, y)
	screen.Add(lbl)

	phrases := []string{"Hello, WillowUI!", "Reactive text!", "Bound to a Ref", "Click to cycle"}
	idx := 0
	cycleBtn := ui.NewButton("cycle", "Cycle", font, sizeSmall)
	cycleBtn.SetSize(80, 28)
	cycleBtn.SetPosition(colLeft+240, y)
	screen.Add(cycleBtn)
	cycleBtn.SetOnClick(func() {
		idx = (idx + 1) % len(phrases)
		textRef.Set(phrases[idx])
	})

	st1 := addStatus(screen, font, sizeSmall, colRight, y)
	ui.WatchValue(textRef, func(_, v string) {
		st1.SetText(fmt.Sprintf("value: %q", v))
	})

	y += 28 + 30

	// ── 2. Computed label ────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "NewComputed - derived label, re-evaluates when deps change", colLeft, y)
	y += 20

	countRef := ui.NewRef(0)

	computed := ui.NewComputed(func() string {
		n := countRef.Get()
		switch {
		case n == 0:
			return "idle"
		case n < 5:
			return "warming up..."
		case n < 10:
			return "running"
		default:
			return "done!"
		}
	})

	computedLbl := ui.NewLabel("computed-lbl", "", font, sizeMedium)
	computedLbl.SetColor(willow.RGBA(0.7, 0.85, 1, 1))
	computedLbl.SetPosition(colLeft, y)
	screen.Add(computedLbl)

	ui.WatchEffect(func() {
		computedLbl.SetText(fmt.Sprintf("status: %s  (count=%d)", computed.Get(), countRef.Get()))
	})

	y += 28 + 30

	// ── 3. Buttons mutating Ref[int] ─────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Button.OnClick → Ref[int] → two independent watchers", colLeft, y)
	y += 20

	incBtn := ui.NewButton("inc", "Increment", font, sizeMedium)
	incBtn.SetSize(130, 36)
	incBtn.SetPosition(colLeft, y)
	screen.Add(incBtn)
	incBtn.SetOnClick(func() {
		countRef.Set(countRef.Peek() + 1)
	})

	resetBtn := ui.NewButton("reset", "Reset", font, sizeMedium)
	resetBtn.SetSize(90, 36)
	resetBtn.SetPosition(colLeft+140, y)
	screen.Add(resetBtn)
	resetBtn.SetOnClick(func() {
		countRef.Set(0)
	})

	disBtn := ui.NewButton("dis", "Disabled", font, sizeMedium)
	disBtn.SetSize(110, 36)
	disBtn.SetEnabled(false)
	disBtn.SetPosition(colLeft+240, y)
	screen.Add(disBtn)

	// Watcher A: raw count
	stA := addStatus(screen, font, sizeSmall, colRight, y)
	ui.WatchValue(countRef, func(_, v int) {
		stA.SetText(fmt.Sprintf("count: %d", v))
	})

	// Watcher B: even/odd
	stB := addStatus(screen, font, sizeSmall, colRight, y+20)
	stB.SetColor(willow.RGBA(1, 0.85, 0.4, 1))
	ui.WatchValue(countRef, func(_, v int) {
		if v%2 == 0 {
			stB.SetText("parity: even")
		} else {
			stB.SetText("parity: odd")
		}
	})

	y += 36 + 24

	// ── 4. BindText / BindEnabled / BindVisible ───────────────────────────────
	addHeader(screen, font, sizeSmall, "BindText (live label)  BindEnabled (>=5)  BindVisible (>=10)", colLeft, y)
	y += 20

	// BindText + BindEnabled: button label shows live count, enabled at count >= 5.
	countLblRef := ui.NewRef("")
	ui.WatchValue(countRef, func(_, v int) {
		countLblRef.Set(fmt.Sprintf("Count: %d", v))
	})

	halfwayRef := ui.NewRef(false)
	ui.WatchValue(countRef, func(_, v int) { halfwayRef.Set(v >= 5) })

	countDisplayBtn := ui.NewButton("count-display", "", font, sizeMedium)
	countDisplayBtn.SetSize(130, 36)
	countDisplayBtn.BindText(countLblRef)
	countDisplayBtn.BindEnabled(halfwayRef)
	countDisplayBtn.SetPosition(colLeft, y)
	screen.Add(countDisplayBtn)

	// BindEnabled: reset only enabled when count > 0.
	canResetRef := ui.NewRef(false)
	ui.WatchValue(countRef, func(_, v int) { canResetRef.Set(v > 0) })

	bindResetBtn := ui.NewButton("bind-reset", "Reset", font, sizeMedium)
	bindResetBtn.SetSize(90, 36)
	bindResetBtn.SetPosition(colLeft+140, y)
	bindResetBtn.BindEnabled(canResetRef)
	screen.Add(bindResetBtn)
	bindResetBtn.SetOnClick(func() {
		countRef.Set(0)
	})

	// BindVisible: "Done!" label only visible when count >= 10.
	doneRef := ui.NewRef(false)
	ui.WatchValue(countRef, func(_, v int) { doneRef.Set(v >= 10) })

	doneLbl := ui.NewLabel("done", "Done!", font, sizeMedium)
	doneLbl.SetColor(willow.RGBA(0.3, 1, 0.5, 1))
	doneLbl.SetPosition(colLeft+240, y+8)
	doneLbl.BindVisible(doneRef)
	screen.Add(doneLbl)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive - Labels & Buttons",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addHeader(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("hdr", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.45, 0.55, 0.65, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}

func addStatus(screen *ui.Screen, font *willow.FontFamily, fontSize, x, y float64) *ui.Label {
	lbl := ui.NewLabel("status", "...", font, fontSize)
	lbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	lbl.SetPosition(x, y)
	screen.Add(lbl)
	return lbl
}
