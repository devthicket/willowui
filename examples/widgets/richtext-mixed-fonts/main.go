// RichText-Mixed-Fonts demonstrates rendering RichText with two different
// FontFamily instances (Lato sans-serif and Noto Serif) in the same view,
// mixing them within the same RichText via AddStyledSpan and AddTextSpan.
//
// Prerequisites:
//
//	go generate ./examples/widgets/richtext-mixed-fonts/
package main

//go:generate go run gen.go

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 640
)

func loadBundle(name string) *willow.FontFamily {
	_, src, _, _ := runtime.Caller(0)
	p := filepath.Join(filepath.Dir(src), "..", "..", "_assets", "fonts", name)
	data, err := os.ReadFile(p)
	if err != nil {
		log.Fatalf("load %s: %v", name, err)
	}
	ff, err := willow.NewFontFamilyFromFontBundle(data)
	if err != nil {
		log.Fatalf("parse %s: %v", name, err)
	}
	return ff
}

type controller struct {
	lato *willow.FontFamily
	noto *willow.FontFamily
	rts  []*ui.RichText
}

func (c *controller) OnCreate(s *ui.Screen) {
	lato := c.lato
	noto := c.noto

	const (
		sizeLarge   = 28.0
		sizeRegular = 18.0
		sizeSmall   = 14.0
	)

	white := willow.RGBA(1, 1, 1, 1)
	gold := willow.RGBA(1, 0.85, 0.3, 1)
	cyan := willow.RGBA(0.4, 0.85, 1, 1)
	pink := willow.RGBA(1, 0.5, 0.75, 1)
	green := willow.RGBA(0.5, 1, 0.6, 1)
	muted := willow.RGBA(0.4, 0.5, 0.6, 1)

	// ── Title ────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "Mixed Fonts -- Lato + Noto Serif", lato)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Bold = true
	title.TextBlock.Color = white
	title.SetPosition(24, 16)
	s.AddNode(title)

	div := ui.NewDivider("div", screenW-48)
	div.SetPosition(24, 56)
	s.AddNode(div)

	// ── Section 1: Same text, two fonts ─────────────────────────────────────
	secLabel(s, lato, "Same sentence in Lato (sans) vs Noto Serif", 24, 68, muted)

	rt1 := ui.NewRichText("lato-sentence", lato, sizeRegular)
	rt1.SetWrapWidth(screenW - 48)
	rt1.SetColor(white)
	rt1.AddSpan("The quick brown fox jumps over the lazy dog. ")
	rt1.AddBoldSpan("Bold. ", gold)
	rt1.AddItalicSpan("Italic. ", cyan)
	rt1.AddBoldItalicSpan("Bold+Italic.", pink)
	rt1.SetPosition(24, 90)
	s.Add(rt1)

	rt2 := ui.NewRichText("noto-sentence", noto, sizeRegular)
	rt2.SetWrapWidth(screenW - 48)
	rt2.SetColor(white)
	rt2.AddSpan("The quick brown fox jumps over the lazy dog. ")
	rt2.AddBoldSpan("Bold. ", gold)
	rt2.AddItalicSpan("Italic. ", cyan)
	rt2.AddBoldItalicSpan("Bold+Italic.", pink)
	rt2.SetPosition(24, 130)
	s.Add(rt2)

	// ── Section 2: Mixed fonts in one RichText ──────────────────────────────
	secLabel(s, lato, "Mixing Lato and Noto Serif in a single RichText", 24, 180, muted)

	rt3 := ui.NewRichText("mixed", lato, sizeRegular)
	rt3.SetWrapWidth(screenW - 48)
	rt3.SetColor(white)
	rt3.AddSpan("This sentence is in Lato, but ")
	rt3.AddStyledSpan("this part switches to Noto Serif", noto, cyan, nil)
	rt3.AddSpan(", then back to Lato. You can ")
	rt3.AddTextSpan(ui.TextSpan{Text: "mix bold serif ", Source: noto, Bold: true, Color: gold, ColorSet: true})
	rt3.AddSpan("with ")
	rt3.AddItalicSpan("italic sans ", pink)
	rt3.AddSpan("freely.")
	rt3.SetPosition(24, 202)
	s.Add(rt3)

	// ── Section 3: Style grid ───────────────────────────────────────────────
	secLabel(s, lato, "Style grid: Regular / Bold / Italic / Bold+Italic", 24, 272, muted)

	type styleRow struct {
		label  string
		font   *willow.FontFamily
		bold   bool
		italic bool
		color  willow.Color
	}
	rows := []styleRow{
		{"Lato Regular", lato, false, false, white},
		{"Lato Bold", lato, true, false, gold},
		{"Lato Italic", lato, false, true, cyan},
		{"Lato Bold+Italic", lato, true, true, pink},
		{"Noto Regular", noto, false, false, white},
		{"Noto Bold", noto, true, false, gold},
		{"Noto Italic", noto, false, true, cyan},
		{"Noto Bold+Italic", noto, true, true, pink},
	}

	y := 294.0
	for i, r := range rows {
		x := 24.0
		if i >= 4 {
			x = 410.0
		}
		if i == 4 {
			y = 294.0
		}
		rt := ui.NewRichText("grid", r.font, sizeRegular)
		rt.SetColor(white)
		rt.AddTextSpan(ui.TextSpan{Text: r.label, Bold: r.bold, Italic: r.italic, Color: r.color, ColorSet: true})
		rt.SetPosition(x, y)
		s.Add(rt)
		c.rts = append(c.rts, rt)
		y += 30
	}

	// ── Section 4: Paragraph comparison ─────────────────────────────────────
	secLabel(s, lato, "Side-by-side paragraphs with word wrap", 24, 440, muted)

	bg1 := willow.NewSprite("bg1", willow.TextureRegion{})
	bg1.SetPosition(24, 462)
	bg1.SetScale(360, 150)
	bg1.SetColor(willow.RGBA(0.12, 0.12, 0.15, 1))
	s.AddNode(bg1)

	rt4 := ui.NewRichText("para-lato", lato, sizeRegular)
	rt4.SetWrapWidth(344)
	rt4.SetColor(willow.RGBA(0.9, 0.9, 0.9, 1))
	rt4.AddBoldSpan("Lato (sans-serif): ", green)
	rt4.AddSpan("A clean, modern typeface designed for readability on screen. ")
	rt4.AddItalicSpan("Italics add emphasis ", cyan)
	rt4.AddSpan("while keeping a light feel.")
	rt4.SetPosition(32, 470)
	s.Add(rt4)

	bg2 := willow.NewSprite("bg2", willow.TextureRegion{})
	bg2.SetPosition(410, 462)
	bg2.SetScale(366, 150)
	bg2.SetColor(willow.RGBA(0.12, 0.12, 0.15, 1))
	s.AddNode(bg2)

	rt5 := ui.NewRichText("para-noto", noto, sizeRegular)
	rt5.SetWrapWidth(350)
	rt5.SetColor(willow.RGBA(0.9, 0.9, 0.9, 1))
	rt5.AddBoldSpan("Noto Serif: ", green)
	rt5.AddSpan("A classic serif typeface with broad Unicode coverage. ")
	rt5.AddItalicSpan("Italics evoke tradition ", cyan)
	rt5.AddSpan("and add a formal touch.")
	rt5.SetPosition(418, 470)
	s.Add(rt5)

	c.rts = append(c.rts, rt1, rt2, rt3, rt4, rt5)
}

func (c *controller) OnUpdate(dt float64) {
	for _, rt := range c.rts {
		rt.Render()
	}
}

func (c *controller) OnDestroy() {}

func main() {
	ui.MustLoadDefaultFont()
	lato := loadBundle("lato.fontbundle")
	noto := loadBundle("notoserif.fontbundle")

	ctrl := &controller{lato: lato, noto: noto}
	ui.Stage.Add(ui.NewScreen(ui.WithController(ctrl)))
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI -- Mixed Fonts Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func secLabel(s *ui.Screen, font *willow.FontFamily, text string, x, y float64, color willow.Color) {
	n := willow.NewText("sec", text, font)
	n.TextBlock.FontSize = 14
	n.TextBlock.Color = color
	n.SetPosition(x, y)
	s.AddNode(n)
}
