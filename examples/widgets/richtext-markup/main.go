// RichTextMarkup demonstrates XML-like markup parsing into WillowUI RichText spans.
// Tags like <b>, <i>, <u>, <strike>, <color>, <size>, <outline>, <link>, <h1>-<h3>,
// <ul>/<ol>/<li>, and <br/> are parsed and rendered as styled text.
//
// Includes an interactive textarea where you can type markup and see it rendered
// in real time.
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

type richTextMarkupController struct {
	family *willow.FontFamily
	allRTs []*ui.RichText
}

func (c *richTextMarkupController) OnCreate(s *ui.Screen) {
	family := c.family

	const (
		sizeLarge   = 28.0
		sizeRegular = 18.0
		sizeSmall   = 16.0
	)
	// Resolve the regular font for direct willow.NewText usage.
	regularFont := family
	lh := regularFont.LineHeight(0, false, false)
	scaleLarge := sizeLarge / lh
	scaleSmall := sizeSmall / lh

	// ── Title ───────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI \u2014 RichText Markup Demo", regularFont)
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetScale(scaleLarge, scaleLarge)
	title.SetPosition(24, 16)
	s.AddNode(title)

	div := ui.NewDivider("divider", screenW-48)
	div.SetPosition(24, 52)
	s.AddNode(div)

	// ── Section 1: Static examples ──────────────────────────────────────────
	addSectionLabel(s, family, scaleSmall, "Bold, italic, underline, color, outline, and links via markup tags", 24, 62)

	rt1 := ui.NewRichText("example1", family, sizeRegular)
	rt1.SetWrapWidth(screenW - 48)
	rt1.SetColor(willow.RGBA(0.9, 0.9, 0.9, 1))
	rt1.SetOnLinkClick(func(url string) {
		fmt.Println("Link clicked:", url)
	})
	rt1.SetMarkup(
		`This is <b>bold</b>, <i>italic</i>, and <b><i>bold italic</i></b>. ` +
			`<u>Underlined</u> and <strike>struck through</strike>. ` +
			`Colors: <color value="#ff3333">red</color>, <color value="#33ff33">green</color>. ` +
			`<outline thickness="2" color="black"><color value="white"><b>OUTLINED</b></color></outline>. ` +
			`<link url="https://example.com">Click this link</link>.`,
	)
	rt1.SetPosition(24, 82)
	s.Add(rt1)

	rt2 := ui.NewRichText("example2", family, sizeRegular)
	rt2.SetWrapWidth(screenW - 48)
	rt2.SetColor(willow.RGBA(0.85, 0.85, 0.85, 1))
	rt2.SetMarkup(
		`<h1>Heading 1</h1>` +
			`<h2>Heading 2</h2>` +
			`<h3>Heading 3</h3>` +
			`<ul><li>Bullet one</li><li>Bullet two</li></ul>` +
			`<ol><li>Numbered one</li><li>Numbered two</li></ol>`,
	)
	rt2.SetPosition(24, 130)
	s.Add(rt2)

	div2 := ui.NewDivider("divider-2", screenW-48)
	div2.SetPosition(24, 310)
	s.AddNode(div2)

	// ── Section 2: Interactive editor ───────────────────────────────────────
	addSectionLabel(s, family, scaleSmall, "Type markup below \u2014 live preview on the right", 24, 320)
	addSectionLabel(s, family, scaleSmall,
		"Tags: <b> <i> <u> <strike> <color> <size> <outline> <link> <br/> <h1-3> <ul> <ol> <li>",
		24, 336)

	const (
		taX = 24
		taY = 358
		taW = 360
		taH = 220
	)

	ta := ui.NewTextArea("markup-input", family, sizeRegular)
	ta.SetSize(taW, taH)
	ta.SetPosition(taX, taY)
	s.Add(ta)

	defaultMarkup := `<b><color value="#ffcc00">Hello!</color></b> Type your <i>markup</i> here.

Try: <color value="#33ccff">colored text</color>
Or: <b><i><color value="#ff9933">bold italic orange</color></i></b>
Or: <u>underlined</u> and <strike>struck</strike>
Or: <size value="24">big text</size>`
	ta.SetValue(defaultMarkup)

	// Right side: preview panel
	const (
		previewX = 400
		previewY = 358
		previewW = 376
		previewH = 220
	)

	addSectionLabel(s, family, scaleSmall, "Preview:", previewX, 320)

	previewBG := willow.NewSprite("preview-bg", willow.TextureRegion{})
	previewBG.SetPosition(previewX, previewY)
	previewBG.SetScale(previewW, previewH)
	previewBG.SetColor(willow.RGBA(0.12, 0.12, 0.15, 1))
	s.AddNode(previewBG)

	rtPreview := ui.NewRichText("preview", family, sizeRegular)
	rtPreview.SetWrapWidth(previewW - 16)
	rtPreview.SetColor(willow.RGBA(0.9, 0.9, 0.9, 1))
	rtPreview.SetPosition(previewX+8, previewY+8)
	s.Add(rtPreview)

	// Initial parse.
	rtPreview.SetMarkup(defaultMarkup)

	// Re-parse on every text change.
	ta.SetOnChange(func(v string) {
		err := rtPreview.SetMarkup(v)
		if err != nil {
			rtPreview.ClearSpans()
			rtPreview.AddSpan("Parse error: " + err.Error())
		}
	})

	c.allRTs = []*ui.RichText{rt1, rt2, rtPreview}
}

func (c *richTextMarkupController) OnUpdate(dt float64) {
	for _, rt := range c.allRTs {
		rt.Render()
	}
}

func (c *richTextMarkupController) OnDestroy() {}

func main() {
	family := ui.MustLoadDefaultFont()

	ctrl := &richTextMarkupController{family: family}

	ui.Stage.Add(ui.NewScreen(ui.WithController(ctrl)))
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI \u2014 RichText Markup Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addSectionLabel(s *ui.Screen, font *willow.FontFamily, scale float64, label string, x, y float64) {
	n := willow.NewText("section", label, font)
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetScale(scale, scale)
	n.SetPosition(x, y)
	s.AddNode(n)
}
