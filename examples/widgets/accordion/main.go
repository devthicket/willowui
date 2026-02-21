// Accordion - collapsible section panels.
// Demonstrates exclusive mode (only one open) and multi-open mode,
// programmatic open/close, and OnToggle callbacks.
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

	const fontSize = 14.0

	screen := ui.NewScreen()

	title := willow.NewText("title", "Accordion Widget", font)
	title.TextBlock.FontSize = 22
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	// Status label for toggle events.
	statusLbl := ui.NewLabel("status", "Click a section header to expand it", font, fontSize)
	statusLbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	statusLbl.SetPosition(24, 560)
	screen.Add(statusLbl)

	// ── Exclusive Accordion (left) ──────────────────────────────────────────
	accLabel := ui.NewLabel("acc-label", "Exclusive Mode", font, fontSize)
	accLabel.SetColor(willow.RGBA(0.5, 0.6, 0.7, 1))
	accLabel.SetPosition(24, 50)
	screen.Add(accLabel)

	acc := ui.NewAccordion("settings")
	acc.SetFont(font, fontSize)
	acc.SetExclusive(true)
	acc.SetPosition(24, 72)

	acc.AddSection(ui.AccordionSection{
		ID:      "general",
		Label:   "General Settings",
		Content: makeContentPanel("general", font, fontSize, "Language, region, and display preferences."),
	})
	acc.AddSection(ui.AccordionSection{
		ID:      "audio",
		Label:   "Audio Settings",
		Content: makeContentPanel("audio", font, fontSize, "Volume, mute, and output device."),
	})
	acc.AddSection(ui.AccordionSection{
		ID:      "video",
		Label:   "Video Settings",
		Content: makeContentPanel("video", font, fontSize, "Resolution, quality, and refresh rate."),
	})
	acc.SetSize(340, 0)
	acc.Open("general")

	acc.SetOnToggle(func(id string, expanded bool) {
		action := "collapsed"
		if expanded {
			action = "expanded"
		}
		statusLbl.SetText(fmt.Sprintf("Section '%s' %s", id, action))
	})
	screen.Add(acc)

	// ── Multi-open Accordion (right) ────────────────────────────────────────
	multiLabel := ui.NewLabel("multi-label", "Multi-open Mode", font, fontSize)
	multiLabel.SetColor(willow.RGBA(0.5, 0.6, 0.7, 1))
	multiLabel.SetPosition(410, 50)
	screen.Add(multiLabel)

	multi := ui.NewAccordion("faq")
	multi.SetFont(font, fontSize)
	multi.SetExclusive(false)
	multi.SetPosition(410, 72)

	multi.AddSection(ui.AccordionSection{
		ID:      "q1",
		Label:   "What is WillowUI?",
		Content: makeContentPanel("a1", font, fontSize, "A UI library for Ebitengine games."),
	})
	multi.AddSection(ui.AccordionSection{
		ID:      "q2",
		Label:   "Is it free?",
		Content: makeContentPanel("a2", font, fontSize, "Core widgets are free and open source."),
	})
	multi.AddSection(ui.AccordionSection{
		ID:      "q3",
		Label:   "How do I get started?",
		Content: makeContentPanel("a3", font, fontSize, "Check the docs and example gallery."),
	})
	multi.SetSize(340, 0)
	screen.Add(multi)

	// ── Control buttons ─────────────────────────────────────────────────────
	btnY := 420.0

	openBtn := ui.NewButton("open-all", "Open All (multi)", font, fontSize)
	openBtn.SetSize(160, 32)
	openBtn.SetPosition(410, btnY)
	openBtn.SetOnClick(func() {
		multi.Open("q1")
		multi.Open("q2")
		multi.Open("q3")
	})
	screen.Add(openBtn)

	closeBtn := ui.NewButton("close-all", "Close All (multi)", font, fontSize)
	closeBtn.SetSize(160, 32)
	closeBtn.SetPosition(580, btnY)
	closeBtn.SetOnClick(func() {
		multi.Close("q1")
		multi.Close("q2")
		multi.Close("q3")
	})
	screen.Add(closeBtn)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Accordion",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func makeContentPanel(name string, font *willow.FontFamily, fontSize float64, text string) *ui.Component {
	p := ui.NewPanel(name + "-panel")
	p.SetBackground(willow.RGBA(0.15, 0.15, 0.18, 1))
	p.SetLayout(ui.LayoutVBox)
	lbl := ui.NewLabel(name+"-text", text, font, fontSize)
	lbl.SetColor(willow.RGBA(0.8, 0.8, 0.8, 1))
	p.AddChild(lbl)
	p.SetSize(320, 50)
	return &p.Component
}
