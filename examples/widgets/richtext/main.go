// RichText demonstrates WillowUI's RichText component: multi-span styled text
// with bold, italic, bold+italic variants from a single FontFamily, colors,
// outlines, word wrapping, and alignment options — all using the span API
// (AddSpan, AddBoldSpan, AddItalicSpan, AddBoldItalicSpan, AddStyledSpan).
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

type richTextController struct {
	allRTs    []*ui.RichText
	rtDynamic *ui.RichText
	colors    []willow.Color
	frame     int
}

func (c *richTextController) OnCreate(s *ui.Screen) {
	font := ui.DefaultFont

	const (
		sizeLarge   = 28.0
		sizeRegular = 18.0
		sizeSmall   = 16.0
	)

	// ── Title ────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI -- RichText Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	s.AddNode(title)

	div := ui.NewDivider("divider", screenW-48)
	div.SetPosition(24, 52)
	s.AddNode(div)

	// ── Section 1: Bold, italic, bold+italic via span API ───────────────────
	addSectionLabel(s, font, sizeSmall, "Bold, italic, and bold+italic from a single FontFamily", 24, 62)

	white := willow.RGBA(0.9, 0.9, 0.9, 1)
	gold := willow.RGBA(1, 0.9, 0.3, 1)
	cyan := willow.RGBA(0.3, 0.8, 1, 1)
	pink := willow.RGBA(1, 0.5, 0.8, 1)

	rt1 := ui.NewRichText("styles", font, sizeRegular)
	rt1.SetWrapWidth(500)
	rt1.SetColor(white)
	rt1.AddSpan("This is ")
	rt1.AddBoldSpan("bold text", gold)
	rt1.AddSpan(", this is ")
	rt1.AddItalicSpan("italic text", cyan)
	rt1.AddSpan(", and this is ")
	rt1.AddBoldItalicSpan("bold italic text", pink)
	rt1.AddSpan(". All three variants resolve from the same FontFamily.")
	rt1.SetPosition(24, 82)
	s.Add(rt1)

	// ── Section 2: Mixed styles with colors and outlines ────────────────────
	addSectionLabel(s, font, sizeSmall, "Outlines on selected spans", 24, 180)

	rt2 := ui.NewRichText("outlines", font, sizeRegular)
	rt2.SetWrapWidth(500)
	rt2.SetColor(willow.RGBA(1, 1, 1, 1))
	rt2.AddSpan("Normal text, then ")
	rt2.AddStyledSpan("outlined bold!", nil,
		willow.RGBA(1, 0.4, 0.4, 1),
		&ui.Outline{Color: willow.RGBA(0, 0, 0, 1), Thickness: 2},
	)
	rt2.AddSpan(" Back to normal, then ")
	rt2.AddStyledSpan("thick outline", nil,
		willow.RGBA(1, 1, 0, 1),
		&ui.Outline{Color: willow.RGBA(0.2, 0, 0, 1), Thickness: 4},
	)
	rt2.AddSpan(" to finish.")
	rt2.SetPosition(24, 200)
	s.Add(rt2)

	// ── Section 3: Word wrapping ────────────────────────────────────────────
	addSectionLabel(s, font, sizeSmall, "Word wrapping at 300px with style changes", 24, 290)

	wrapBG := willow.NewSprite("wrap-bg", willow.TextureRegion{})
	wrapBG.SetPosition(24, 310)
	wrapBG.SetScale(300, 120)
	wrapBG.SetColor(willow.RGBA(0.12, 0.12, 0.15, 1))
	s.AddNode(wrapBG)

	rt3 := ui.NewRichText("wrapping", font, sizeRegular)
	rt3.SetWrapWidth(300)
	rt3.SetColor(willow.RGBA(0.9, 0.9, 0.9, 1))
	rt3.AddSpan("This paragraph demonstrates word wrapping. ")
	rt3.AddBoldSpan("Bold spans ", willow.RGBA(0.6, 0.9, 1, 1))
	rt3.AddSpan("and ")
	rt3.AddItalicSpan("italic spans ", willow.RGBA(0.9, 1, 0.6, 1))
	rt3.AddSpan("wrap correctly across span boundaries.")
	rt3.SetPosition(24, 310)
	s.Add(rt3)

	// ── Section 4: Alignment ────────────────────────────────────────────────
	addSectionLabel(s, font, sizeSmall, "Text alignment: left, center, right", 400, 290)

	type alignEntry struct {
		label string
		align willow.TextAlign
		y     float64
		bold  bool
		ital  bool
	}
	alignments := []alignEntry{
		{"Left-aligned (bold)", willow.TextAlignLeft, 310, true, false},
		{"Center-aligned (italic)", willow.TextAlignCenter, 370, false, true},
		{"Right-aligned (bold+italic)", willow.TextAlignRight, 430, true, true},
	}

	var alignRTs []*ui.RichText
	for i, a := range alignments {
		bg := willow.NewSprite("align-bg", willow.TextureRegion{})
		bg.SetPosition(400, a.y)
		bg.SetScale(350, 50)
		bg.SetColor(willow.RGBA(0.12, 0.12, 0.15, 1))
		s.AddNode(bg)

		rt := ui.NewRichText(fmt.Sprintf("align-%d", i), font, sizeRegular)
		rt.SetWrapWidth(350)
		rt.SetAlign(a.align)
		rt.SetColor(willow.RGBA(0.85, 0.85, 0.85, 1))
		rt.AddTextSpan(ui.TextSpan{Text: a.label, Bold: a.bold, Italic: a.ital})
		rt.SetPosition(400, a.y)
		s.Add(rt)
		alignRTs = append(alignRTs, rt)
	}

	// ── Section 5: Dynamic content ──────────────────────────────────────────
	addSectionLabel(s, font, sizeSmall, "Dynamic: spans rebuilt each second", 24, 470)

	rtDynamic := ui.NewRichText("dynamic", font, sizeRegular)
	rtDynamic.SetWrapWidth(700)
	rtDynamic.SetColor(willow.RGBA(0.8, 0.8, 0.8, 1))
	rtDynamic.SetPosition(24, 490)
	s.Add(rtDynamic)
	c.rtDynamic = rtDynamic

	c.allRTs = append([]*ui.RichText{rt1, rt2, rt3, rtDynamic}, alignRTs...)
}

func (c *richTextController) OnUpdate(dt float64) {
	c.frame++

	// Rebuild dynamic content every 60 frames.
	if c.frame%60 == 0 {
		idx := (c.frame / 60) % len(c.colors)
		c.rtDynamic.ClearSpans()
		c.rtDynamic.AddSpan("Frame ")
		c.rtDynamic.AddBoldSpan(fmt.Sprintf("%d", c.frame), c.colors[idx])
		c.rtDynamic.AddSpan(" -- ")
		c.rtDynamic.AddBoldSpan("Bold", willow.RGBA(1, 1, 1, 1))
		c.rtDynamic.AddSpan(", ")
		c.rtDynamic.AddItalicSpan("italic", willow.RGBA(1, 1, 1, 1))
		c.rtDynamic.AddSpan(", and ")
		c.rtDynamic.AddBoldItalicSpan("bold italic", willow.RGBA(1, 1, 1, 1))
		c.rtDynamic.AddSpan(" all from one FontFamily.")
	}

	for _, rt := range c.allRTs {
		rt.Render()
	}
}

func (c *richTextController) OnDestroy() {}

func main() {
	ui.MustLoadDefaultFont()

	colors := []willow.Color{
		willow.RGBA(1, 0.3, 0.3, 1),
		willow.RGBA(0.3, 1, 0.3, 1),
		willow.RGBA(0.3, 0.3, 1, 1),
		willow.RGBA(1, 1, 0.3, 1),
		willow.RGBA(1, 0.3, 1, 1),
		willow.RGBA(0.3, 1, 1, 1),
	}

	ctrl := &richTextController{colors: colors}

	ui.Stage.Add(ui.NewScreen(ui.WithController(ctrl)))
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI -- RichText Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addSectionLabel(s *ui.Screen, font *willow.FontFamily, fontSize float64, label string, x, y float64) {
	n := willow.NewText("section", label, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	s.AddNode(n)
}
