// Checkbox & Radio - reactive demo.
// Shows Checkbox.BindValue(Ref[bool]) and Radio.BindSelected(Ref[int]).
// A computed summary label reads both refs to produce a derived string.
package main

import (
	"fmt"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 800
	screenH  = 620
	colLeft  = 40.0
	colRight = 450.0
)

func main() {
	font := ui.MustLoadDefaultFont()

	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	title := willow.NewText("title", "Reactive - Checkbox & Radio", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	y := 62.0

	// ── 1. Checkbox ──────────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Checkbox.BindValue - two-way Ref[bool]", colLeft, y)
	y += 20

	enabledRef := ui.NewRef(false)

	cb := ui.NewCheckbox("cb", "Enable notifications", font, sizeMedium)
	cb.BindValue(enabledRef)
	cb.SetPosition(colLeft, y)
	screen.Add(cb)

	st1 := addStatus(screen, font, sizeSmall, colRight, y)
	ui.WatchValue(enabledRef, func(_, v bool) {
		if v {
			st1.SetText("enabled: true")
		} else {
			st1.SetText("enabled: false")
		}
	})

	// Programmatic toggle button - drives the same Ref
	progBtn := ui.NewButton("prog", "Toggle via Ref", font, sizeSmall)
	progBtn.SetSize(130, 28)
	progBtn.SetPosition(colLeft, y+30)
	screen.Add(progBtn)
	progBtn.SetOnClick(func() {
		enabledRef.Set(!enabledRef.Peek())
	})

	y += 70 + 20

	// ── 2. Radio ─────────────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Radio.BindSelected - Ref[int]", colLeft, y)
	y += 20

	themeOptions := []string{"System", "Light", "Dark"}
	themeRef := ui.NewRef(0)

	rg := ui.NewRadio("rg")
	for _, opt := range themeOptions {
		rg.AddOption(opt, font, sizeMedium)
	}
	rg.BindSelected(themeRef)
	rg.SetPosition(colLeft, y)
	screen.Add(rg)

	st2 := addStatus(screen, font, sizeSmall, colRight, y+24)
	ui.WatchValue(themeRef, func(_, v int) {
		if v < 0 {
			st2.SetText("selected: (none)")
		} else {
			st2.SetText(fmt.Sprintf("selected: %d - %s", v, themeOptions[v]))
		}
	})

	rgH := float64(len(themeOptions))*26 + 10
	y += rgH + 20

	// ── 3. Multi-column Radio ─────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Radio.SetColumns(3) - horizontal-first (default)", colLeft, y)
	y += 20

	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	dayRef := ui.NewRef(0)

	rgMulti := ui.NewRadio("rg-multi")
	for _, d := range days {
		rgMulti.AddOption(d, font, sizeMedium)
	}
	rgMulti.SetColumns(3)
	rgMulti.BindSelected(dayRef)
	rgMulti.SetPosition(colLeft, y)
	screen.Add(rgMulti)

	st3 := addStatus(screen, font, sizeSmall, colRight, y+24)
	ui.WatchValue(dayRef, func(_, v int) {
		if v >= 0 && v < len(days) {
			st3.SetText(fmt.Sprintf("day: %s", days[v]))
		}
	})

	y += rgMulti.Height + 30

	// ── 4. Computed summary ───────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Computed - reads both Refs, updates when either changes", colLeft, y)
	y += 20

	summary := ui.NewComputed(func() string {
		t := themeRef.Get()
		e := enabledRef.Get()
		themeName := "none"
		if t >= 0 && t < len(themeOptions) {
			themeName = themeOptions[t]
		}
		notif := "off"
		if e {
			notif = "on"
		}
		return fmt.Sprintf("theme=%s  notifications=%s", themeName, notif)
	})

	summaryLbl := ui.NewLabel("summary", "", font, sizeMedium)
	summaryLbl.SetColor(willow.RGBA(1, 0.85, 0.4, 1))
	summaryLbl.SetPosition(colLeft, y)
	screen.Add(summaryLbl)
	ui.WatchEffect(func() {
		summaryLbl.SetText(summary.Get())
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive - Checkbox & Radio",
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
