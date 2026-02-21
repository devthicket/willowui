package main

import (
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 320
	screenH = 240
)

func main() {
	font := ui.MustLoadDefaultFont()
	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	screen := ui.NewScreen()

	acc := ui.NewAccordion("acc")
	acc.SetFont(font, 14)
	acc.SetPosition((screenW-260)/2, 20)

	sections := []struct{ id, label, text string }{
		{"a", "Section A", "Content for section A"},
		{"b", "Section B", "Content for section B"},
		{"c", "Section C", "Content for section C"},
	}
	for _, s := range sections {
		p := ui.NewPanel(s.id + "-panel")
		p.SetLayout(ui.LayoutVBox)
		lbl := ui.NewLabel(s.id+"-text", s.text, font, 12)
		lbl.SetColor(willow.RGBA(0.8, 0.8, 0.8, 1))
		p.AddChild(lbl)
		p.SetSize(240, 40)
		acc.AddSection(ui.AccordionSection{ID: s.id, Label: s.label, Content: &p.Component})
	}
	acc.SetSize(260, 0)

	screen.Add(acc)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Accordion",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
