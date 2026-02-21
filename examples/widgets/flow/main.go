// Flow layout demonstrates LayoutFlow: children placed left-to-right with
// automatic row wrapping. Contrasts LayoutGrid (uniform cells) with
// LayoutFlow (natural-width wrapping).
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 860
	screenH = 680
)

var (
	colorBG      = willow.RGBA(0.08, 0.08, 0.10, 1)
	colorPanel   = willow.RGBA(0.13, 0.13, 0.16, 1)
	colorSection = willow.RGBA(0.35, 0.45, 0.55, 1)
	colorText    = willow.RGBA(0.93, 0.93, 0.93, 1)
)

// Status-effect icon colors for the buff bar demo.
var buffColors = []willow.Color{
	willow.RGBA(0.9, 0.3, 0.3, 1), // bleed
	willow.RGBA(0.3, 0.7, 0.3, 1), // regen
	willow.RGBA(0.3, 0.5, 0.9, 1), // shield
	willow.RGBA(0.9, 0.7, 0.2, 1), // haste
	willow.RGBA(0.7, 0.3, 0.9, 1), // stun
	willow.RGBA(0.2, 0.8, 0.8, 1), // chill
	willow.RGBA(0.9, 0.5, 0.2, 1), // burn
	willow.RGBA(0.5, 0.9, 0.5, 1), // poison
	willow.RGBA(0.8, 0.8, 0.3, 1), // bless
	willow.RGBA(0.5, 0.5, 0.9, 1), // slow
}

// Tag widths for the chip/tag demo — varied to show natural-width wrapping.
var tagLabels = []string{
	"gameplay", "art", "audio", "bug", "ui", "backend",
	"performance", "critical", "wontfix", "enhancement",
	"good first issue", "help wanted", "in progress",
}

type flowController struct {
	font *willow.FontFamily
}

func (c *flowController) OnCreate(s *ui.Screen) {
	font := c.font

	// ── Title ──────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI: LayoutFlow Demo", font)
	title.TextBlock.FontSize = 22
	title.TextBlock.Color = colorText
	title.SetPosition(24, 14)
	s.AddNode(title)

	div := ui.NewDivider("div", float64(screenW)-48)
	div.SetPosition(24, 44)
	s.AddNode(div)

	y := 58.0

	// ── Section 1: Buff bar (icon wrapping) ────────────────────────────────
	addLabel(s, font, "LayoutFlow: status-effect icons wrapping in a fixed-width panel", 24, y)
	y += 18

	buffBar := makeContainer("buff-bar", 300, 0)
	buffBar.Layout = ui.LayoutFlow
	buffBar.Spacing = 6
	buffBar.FlowRowGap = 6
	buffBar.Padding = ui.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8}
	buffBar.Align = ui.AlignCenter
	buffBar.SetPosition(24, y)
	s.Add(buffBar)

	for i, col := range buffColors {
		icon := makeSquare(fmt.Sprintf("buff-%d", i), 28, col)
		buffBar.AddChild(icon)
	}
	buffBar.SizeToContent()
	resizeContainer(buffBar)

	// ── Section 2: Chip/tag buttons ────────────────────────────────────────
	y += buffBar.Height + 20
	addLabel(s, font, "LayoutFlow: tag chips with natural widths", 24, y)
	y += 18

	tagPanel := makeContainer("tags", 500, 0)
	tagPanel.Layout = ui.LayoutFlow
	tagPanel.Spacing = 6
	tagPanel.FlowRowGap = 6
	tagPanel.Padding = ui.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8}
	tagPanel.SetPosition(24, y)
	s.Add(tagPanel)

	for _, lbl := range tagLabels {
		chip := makeChip(lbl, font)
		tagPanel.AddChild(chip)
	}
	tagPanel.SizeToContent()
	resizeContainer(tagPanel)

	// ── Section 3: Justify variants ────────────────────────────────────────
	y += tagPanel.Height + 20
	addLabel(s, font, "LayoutFlow justify=start / center / end  (5 items, 3 fit per row)", 24, y)
	y += 18

	justifyModes := []struct {
		name    string
		justify ui.Alignment
	}{
		{"start", ui.AlignStart},
		{"center", ui.AlignCenter},
		{"end", ui.AlignEnd},
	}

	jx := 24.0
	for _, jm := range justifyModes {
		lbl := willow.NewText("jlbl-"+jm.name, jm.name, font)
		lbl.TextBlock.FontSize = 13
		lbl.TextBlock.Color = colorSection
		lbl.SetPosition(jx, y)
		s.AddNode(lbl)

		fp := makeContainer("flow-"+jm.name, 240, 100)
		fp.Layout = ui.LayoutFlow
		fp.Spacing = 8
		fp.FlowRowGap = 6
		fp.Padding = ui.Insets{Top: 6, Right: 8, Bottom: 6, Left: 8}
		fp.Justify = jm.justify
		fp.Align = ui.AlignCenter
		fp.SetPosition(jx, y+16)
		s.Add(fp)

		for i := range 5 {
			v := 0.35 + float64(i)*0.12
			box := makeSquare(fmt.Sprintf("%s-%d", jm.name, i), 50, willow.RGBA(v*0.6, v, v*0.8, 1))
			fp.AddChild(box)
		}
		fp.UpdateLayout()
		jx += 280
	}

	// ── Section 4: Grid vs Flow contrast ────────────────────────────────────
	y += 140
	addLabel(s, font, "LayoutGrid (uniform cells)  vs  LayoutFlow (natural widths)", 24, y)
	y += 18

	gridDemo := makeContainer("grid-demo", 280, 110)
	gridDemo.Layout = ui.LayoutGrid
	gridDemo.GridColumns = 4
	gridDemo.Spacing = 6
	gridDemo.Padding = ui.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8}
	gridDemo.SetPosition(24, y)
	s.Add(gridDemo)

	gridLabel := willow.NewText("grid-lbl", "LayoutGrid", font)
	gridLabel.TextBlock.FontSize = 12
	gridLabel.TextBlock.Color = colorSection
	gridLabel.SetPosition(24, y+114)
	s.AddNode(gridLabel)

	for i := range 8 {
		v := 0.4 + float64(i)*0.07
		box := makeSquare(fmt.Sprintf("g-%d", i), 40, willow.RGBA(v, v*0.5, v*0.3, 1))
		gridDemo.AddChild(box)
	}
	gridDemo.UpdateLayout()

	flowDemo := makeContainer("flow-demo", 280, 110)
	flowDemo.Layout = ui.LayoutFlow
	flowDemo.Spacing = 6
	flowDemo.FlowRowGap = 6
	flowDemo.Padding = ui.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8}
	flowDemo.SetPosition(330, y)
	s.Add(flowDemo)

	flowLabel := willow.NewText("flow-lbl", "LayoutFlow", font)
	flowLabel.TextBlock.FontSize = 12
	flowLabel.TextBlock.Color = colorSection
	flowLabel.SetPosition(330, y+114)
	s.AddNode(flowLabel)

	mixedWidths := []float64{40, 60, 25, 50, 35, 55, 45, 30}
	for i, w := range mixedWidths {
		v := 0.4 + float64(i)*0.07
		box := makeRect(fmt.Sprintf("f-%d", i), w, 40, willow.RGBA(v*0.3, v*0.5, v, 1))
		flowDemo.AddChild(box)
	}
	flowDemo.UpdateLayout()
}

func (c *flowController) OnUpdate(_ float64) {}
func (c *flowController) OnDestroy()         {}

func main() {
	font := ui.MustLoadDefaultFont()

	ui.Stage.Add(ui.NewScreen(ui.WithController(&flowController{font: font})))
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — LayoutFlow Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: colorBG,
	})
}

// makeContainer creates a Component with a dark background sprite.
func makeContainer(name string, w, h float64) *ui.Component {
	c := ui.NewComponent(name)
	c.Width = w
	c.Height = h

	bg := willow.NewSprite(name+"-bg", willow.TextureRegion{})
	bg.SetScale(w, h)
	bg.SetColor(colorPanel)
	c.AddRawChild(bg)

	return c
}

// resizeContainer resizes the background sprite to match the component's
// current Width/Height (e.g. after SizeToContent).
func resizeContainer(c *ui.Component) {
	if len(c.Node().Children()) > 0 {
		bg := c.Node().Children()[0]
		bg.SetScale(c.Width, c.Height)
	}
}

// makeSquare creates a w×w colored square component.
func makeSquare(name string, w float64, col willow.Color) *ui.Component {
	return makeRect(name, w, w, col)
}

// makeRect creates a w×h colored rectangle component.
func makeRect(name string, w, h float64, col willow.Color) *ui.Component {
	c := ui.NewComponent(name)
	c.Width = w
	c.Height = h

	sp := willow.NewSprite(name+"-sp", willow.TextureRegion{})
	sp.SetScale(w, h)
	sp.SetColor(col)
	c.AddRawChild(sp)

	return c
}

// makeChip creates a chip/tag component sized to its text label.
func makeChip(text string, font *willow.FontFamily) *ui.Component {
	const fontSize = 13.0
	tw, _ := font.MeasureString(text, 0, false, false)
	scale := fontSize / font.LineHeight(0, false, false)
	w := tw*scale + 16
	h := 24.0

	c := ui.NewComponent("chip-" + text)
	c.Width = w
	c.Height = h

	bg := willow.NewSprite("chip-bg-"+text, willow.TextureRegion{})
	bg.SetScale(w, h)
	bg.SetColor(willow.RGBA(0.25, 0.35, 0.50, 1))
	c.AddRawChild(bg)

	lbl := willow.NewText("chip-lbl-"+text, text, font)
	lbl.TextBlock.FontSize = fontSize
	lbl.TextBlock.Color = colorText
	lbl.SetPosition(8, (h-fontSize*0.75)/2)
	c.AddRawChild(lbl)

	return c
}

// addLabel adds a small section heading to the screen.
func addLabel(s *ui.Screen, font *willow.FontFamily, text string, x, y float64) {
	n := willow.NewText("lbl", text, font)
	n.TextBlock.FontSize = 13
	n.TextBlock.Color = colorSection
	n.SetPosition(x, y)
	s.AddNode(n)
}
