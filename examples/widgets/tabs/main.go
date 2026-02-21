// Tabs - reactive demo with scrollable overflow.
// Shows TabBar.BindSelected(Ref[int]), ToggleButtonBar.BindSelected(Ref[int]),
// and TabBar scrollable overflow mode with dynamic add/remove.
package main

import (
	"fmt"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 800
	screenH  = 600
	colLeft  = 40.0
	colRight = 450.0
)

var tabNames = []string{"Overview", "Details", "History"}

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

	title := willow.NewText("title", "Reactive - Tabs & ToggleButtonBar", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	y := 62.0

	// Shared selection Ref - both widgets observe and mutate the same value.
	selRef := ui.NewRef(0)

	// ── 1. TabBar ─────────────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "TabBar.BindSelected - Ref[int] (shared with ToggleButtonBar below)", colLeft, y)
	y += 20

	tabs := ui.NewTabBar("tabs", font, sizeMedium)
	tabs.SetSize(400, 110)
	for _, name := range tabNames {
		p := ui.NewComponent("tab-content-" + name)
		p.Width = 400
		p.Height = 74
		contentLbl := ui.NewLabel("tc-lbl-"+name, name+" panel content", font, sizeSmall)
		contentLbl.SetColor(willow.RGBA(0.6, 0.65, 0.7, 1))
		contentLbl.SetPosition(12, 24)
		p.AddChild(contentLbl)
		tabs.AddTab(name, p)
	}
	tabs.BindSelected(selRef)
	tabs.SetPosition(colLeft, y)
	screen.Add(tabs)

	st1 := addStatus(screen, font, sizeSmall, colRight, y+20)
	ui.WatchValue(selRef, func(_, v int) {
		st1.SetText(fmt.Sprintf("selected: %d - %s", v, tabNames[v]))
	})

	y += 110 + 28

	// ── 2. ToggleButtonBar - same Ref ─────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "ToggleButtonBar.BindSelected - same Ref[int] as TabBar above", colLeft, y)
	y += 20

	tbb := ui.NewToggleButtonBar("tbb", font, sizeMedium)
	tbb.SetSize(400, 36)
	for _, name := range tabNames {
		tbb.AddButton(name)
	}
	tbb.BindSelected(selRef)
	tbb.SetPosition(colLeft, y)
	screen.Add(tbb)

	noteLbl := ui.NewLabel("note", "Selecting a button moves the tab, and vice versa.", font, sizeSmall)
	noteLbl.SetColor(willow.RGBA(0.5, 0.55, 0.6, 1))
	noteLbl.SetPosition(colLeft, y+44)
	screen.Add(noteLbl)

	y += 36 + 48 + 20

	// ── 3. Programmatic jump ──────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Buttons mutate Ref directly - both widgets follow", colLeft, y)
	y += 20

	for i, name := range tabNames {
		idx := i
		btn := ui.NewButton(fmt.Sprintf("jump-%d", i), name, font, sizeSmall)
		btn.SetSize(110, 30)
		btn.SetPosition(colLeft+float64(i)*120, y)
		screen.Add(btn)
		btn.SetOnClick(func() {
			selRef.Set(idx)
		})
	}

	y += 50

	// ── 4. Scrollable TabBar ──────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Scrollable Overflow - many tabs with arrows", colLeft, y)
	y += 20

	scrollTabs := ui.NewTabBar("scroll-tabs", font, sizeMedium)
	scrollTabs.SetSize(400, 50)
	scrollTabs.SetOverflowMode(ui.TabOverflowScroll)
	scrollTabs.SetPosition(colLeft, y)
	screen.Add(scrollTabs)

	tabCounter := 0
	addScrollTab := func() {
		scrollTabs.AddTab(fmt.Sprintf("Tab %d", tabCounter), nil)
		tabCounter++
	}

	// Start with 8 tabs.
	for i := 0; i < 8; i++ {
		addScrollTab()
	}

	y += 55

	// Add tab button.
	addBtn := ui.NewButton("add-tab", "Add Tab", font, sizeSmall)
	addBtn.SetSize(100, 30)
	addBtn.SetPosition(colLeft, y)
	addBtn.SetOnClick(func() {
		addScrollTab()
	})
	screen.Add(addBtn)

	// Remove tab button.
	rmBtn := ui.NewButton("rm-tab", "Remove Tab", font, sizeSmall)
	rmBtn.SetSize(120, 30)
	rmBtn.SetPosition(colLeft+110, y)
	rmBtn.SetOnClick(func() {
		if scrollTabs.TabCount() > 0 {
			scrollTabs.RemoveTab(scrollTabs.TabCount() - 1)
		}
	})
	screen.Add(rmBtn)

	// Select tab 0 button.
	firstBtn := ui.NewButton("select-first", "Select First", font, sizeSmall)
	firstBtn.SetSize(120, 30)
	firstBtn.SetPosition(colLeft+240, y)
	firstBtn.SetOnClick(func() {
		scrollTabs.SetSelected(0)
	})
	screen.Add(firstBtn)

	// Select last tab button.
	lastBtn := ui.NewButton("select-last", "Select Last", font, sizeSmall)
	lastBtn.SetSize(120, 30)
	lastBtn.SetPosition(colLeft+370, y)
	lastBtn.SetOnClick(func() {
		if scrollTabs.TabCount() > 0 {
			scrollTabs.SetSelected(scrollTabs.TabCount() - 1)
		}
	})
	screen.Add(lastBtn)

	y += 38

	// Toggle overflow mode.
	modeToggle := ui.NewToggle("mode-toggle")
	modeToggle.SetValue(true) // starts in scroll mode
	modeToggle.SetPosition(colLeft, y+2)
	screen.Add(modeToggle)

	modeLbl := ui.NewLabel("mode-lbl", "Scroll mode", font, sizeSmall)
	modeLbl.SetColor(willow.RGBA(0.7, 0.75, 0.8, 1))
	modeLbl.SetPosition(colLeft+55, y+4)
	screen.Add(modeLbl)

	modeToggle.SetOnChange(func(on bool) {
		if on {
			scrollTabs.SetOverflowMode(ui.TabOverflowScroll)
			modeLbl.SetText("Scroll mode")
		} else {
			scrollTabs.SetOverflowMode(ui.TabOverflowClip)
			modeLbl.SetText("Clip mode")
		}
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive - Tabs & ToggleButtonBar",
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
