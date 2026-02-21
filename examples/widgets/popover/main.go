// Popover demonstrates WillowUI's Popover widget: a floating interactive panel
// anchored to a trigger button, dismissed by clicking outside.
package main

import (
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 640
	screenH = 480
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		fontSize  = 14.0
		titleSize = 13.0
	)

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	titleNode := willow.NewText("title", "WillowUI: Popover", font)
	titleNode.TextBlock.FontSize = 18.0
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetPosition(24, 18)
	screen.AddNode(titleNode)

	// ── Filter popover (opens below trigger) ─────────────────────────────────

	// Content: radio group for filter options.
	filterGroup := ui.NewRadio("filter-group")
	filterGroup.AddOption("All", font, fontSize)
	filterGroup.AddOption("Active", font, fontSize)
	filterGroup.AddOption("Archived", font, fontSize)
	filterGroup.SetSelected(0)

	filterPopover := ui.NewPopover("filter-pop")
	filterPopover.SetTitle("Filters", font, titleSize)
	filterPopover.SetShowCloseButton(true)
	filterPopover.SetContentSize(200, 120)
	filterPopover.SetContent(filterGroup)
	filterPopover.SetPreferredSide(ui.PopoverBelow)
	filterPopover.SetOnOpen(func() {
		log.Println("filter popover opened")
	})
	filterPopover.SetOnClose(func() {
		log.Println("filter popover closed")
	})

	filterBtn := ui.NewButton("filter-btn", "Filter ▾", font, fontSize)
	filterBtn.SetSize(120, 36)
	filterBtn.SetPosition(80, 100)
	filterBtn.SetOnClick(func() {
		if filterPopover.IsOpen() {
			filterPopover.Close()
		} else {
			filterPopover.Open(&filterBtn.Component)
		}
	})
	screen.Add(filterBtn)

	// ── Stats popover (opens above trigger) ──────────────────────────────────

	statsPanel := ui.NewPanel("stats-panel")
	statsPanel.SetSize(220, 120)
	statsPanel.SetLayout(ui.LayoutVBox)
	statsPanel.Spacing = 4

	for _, line := range []string{"HP:   82 / 100", "ATK:  45", "DEF:  30", "SPD:  67"} {
		lbl := ui.NewLabel("stat-"+line, line, font, fontSize)
		statsPanel.AddChild(lbl)
	}

	statsPopover := ui.NewPopover("stats-pop")
	statsPopover.SetTitle("Stats", font, titleSize)
	statsPopover.SetShowCloseButton(false)
	statsPopover.SetContentSize(220, 120)
	statsPopover.SetContent(statsPanel)
	statsPopover.SetPreferredSide(ui.PopoverAbove)

	statsBtn := ui.NewButton("stats-btn", "View Stats", font, fontSize)
	statsBtn.SetSize(120, 36)
	statsBtn.SetPosition(80, 310)
	statsBtn.SetOnClick(func() {
		if statsPopover.IsOpen() {
			statsPopover.Close()
		} else {
			statsPopover.Open(&statsBtn.Component)
		}
	})
	screen.Add(statsBtn)

	// ── Info labels ──────────────────────────────────────────────────────────
	info1 := willow.NewText("info1", "Click 'Filter ▾' to open a popover below it", font)
	info1.TextBlock.FontSize = 12
	info1.TextBlock.Color = willow.RGBA(0.7, 0.7, 0.7, 1)
	info1.SetPosition(24, 160)
	screen.AddNode(info1)

	info2 := willow.NewText("info2", "Click 'View Stats' to open a popover above it", font)
	info2.TextBlock.FontSize = 12
	info2.TextBlock.Color = willow.RGBA(0.7, 0.7, 0.7, 1)
	info2.SetPosition(24, 360)
	screen.AddNode(info2)

	info3 := willow.NewText("info3", "Click outside any popover to dismiss it", font)
	info3.TextBlock.FontSize = 12
	info3.TextBlock.Color = willow.RGBA(0.7, 0.7, 0.7, 1)
	info3.SetPosition(24, 440)
	screen.AddNode(info3)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Popover",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.12, 0.12, 0.14, 1),
	})
}
