// Badge demonstrates the WillowUI Badge widget: pill-shaped count/label
// overlays and dot-mode status indicators with per-variant color presets.
package main

import (
	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 520
	screenH = 460
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		titleSize = 18.0
		labelSize = 12.0
		badgeSize = 11.0
	)

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	titleNode := willow.NewText("title", "WillowUI: Badge", font)
	titleNode.TextBlock.FontSize = titleSize
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetPosition(24, 18)
	screen.AddNode(titleNode)

	// ── Count badges ─────────────────────────────────────────────────────────
	y := 58.0
	addSection(screen, font, labelSize, "COUNTS", 24, y)
	y += 22

	x := 24.0
	for _, tc := range []struct {
		name    string
		count   int
		variant ui.Variant
	}{
		{"primary", 3, ui.Primary},
		{"success", 12, ui.Success},
		{"warning", 42, ui.Warning},
		{"danger", 99, ui.Danger},
		{"neutral", 0, ui.Neutral},
	} {
		b := ui.NewBadge("count-"+tc.name, font, badgeSize)
		b.SetCount(tc.count)
		b.SetVariant(tc.variant)
		b.SizeToContent()
		b.SetPosition(x, y)
		screen.Add(b)

		lbl := ui.NewLabel("lbl-"+tc.name, tc.name, font, labelSize)
		lbl.SetColor(willow.RGBA(0.6, 0.65, 0.7, 1))
		lbl.SetPosition(x, y+b.Height+4)
		screen.Add(lbl)

		x += b.Width + 32
	}

	// ── Max count truncation ─────────────────────────────────────────────────
	y += 56
	addSection(screen, font, labelSize, "MAX COUNT (99)", 24, y)
	y += 22

	overBadge := ui.NewBadge("over", font, badgeSize)
	overBadge.SetMaxCount(99)
	overBadge.SetCount(150)
	overBadge.SetVariant(ui.Danger)
	overBadge.SizeToContent()
	overBadge.SetPosition(24, y)
	screen.Add(overBadge)

	overLbl := ui.NewLabel("over-lbl", "150 → 99+", font, labelSize)
	overLbl.SetColor(willow.RGBA(0.6, 0.65, 0.7, 1))
	overLbl.SetPosition(24+overBadge.Width+10, y+2)
	screen.Add(overLbl)

	// ── Text labels ──────────────────────────────────────────────────────────
	y += 44
	addSection(screen, font, labelSize, "TEXT LABELS", 24, y)
	y += 22

	x = 24
	for _, tc := range []struct {
		name    string
		text    string
		variant ui.Variant
	}{
		{"new", "NEW", ui.Primary},
		{"rare", "Rare", ui.Success},
		{"hot", "HOT", ui.Danger},
		{"beta", "BETA", ui.Warning},
	} {
		b := ui.NewBadge("text-"+tc.name, font, badgeSize)
		b.SetText(tc.text)
		b.SetVariant(tc.variant)
		b.SizeToContent()
		b.SetPosition(x, y)
		screen.Add(b)
		x += b.Width + 16
	}

	// ── Padding comparison ──────────────────────────────────────────────────
	y += 44
	addSection(screen, font, labelSize, "PADDING", 24, y)
	y += 22

	x = 24
	for _, tc := range []struct {
		name                     string
		top, right, bottom, left float64
	}{
		{"tight", 1, 3, 1, 3},
		{"default", 2, 6, 2, 6},
		{"medium", 4, 10, 4, 10},
		{"roomy", 6, 16, 6, 16},
		{"wide", 2, 20, 2, 20},
	} {
		b := ui.NewBadge("pad-"+tc.name, font, badgeSize)
		b.SetText("42")
		b.SetVariant(ui.Primary)
		b.SetPadding(tc.top, tc.right, tc.bottom, tc.left)
		b.SizeToContent()
		b.SetPosition(x, y)
		screen.Add(b)

		lbl := ui.NewLabel("pad-lbl-"+tc.name, tc.name, font, labelSize)
		lbl.SetColor(willow.RGBA(0.6, 0.65, 0.7, 1))
		lbl.SetPosition(x, y+b.Height+4)
		screen.Add(lbl)

		x += b.Width + 20
	}

	// ── Dot mode (status indicators) ─────────────────────────────────────────
	y += 44
	addSection(screen, font, labelSize, "DOT MODE (STATUS)", 24, y)
	y += 22

	x = 24
	for _, tc := range []struct {
		name    string
		variant ui.Variant
	}{
		{"online", ui.Success},
		{"away", ui.Warning},
		{"busy", ui.Danger},
		{"offline", ui.Neutral},
	} {
		dot := ui.NewBadge("dot-"+tc.name, font, badgeSize)
		dot.SetDotMode(true)
		dot.SetVariant(tc.variant)
		dot.SizeToContent()
		dot.SetPosition(x, y+2)
		screen.Add(dot)

		lbl := ui.NewLabel("dot-lbl-"+tc.name, tc.name, font, labelSize)
		lbl.SetColor(willow.RGBA(0.6, 0.65, 0.7, 1))
		lbl.SetPosition(x+dot.Width+6, y)
		screen.Add(lbl)

		x += dot.Width + lbl.Width + 28
	}

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Badge — WillowUI",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.10, 0.10, 0.12, 1),
	})
}

func addSection(screen *ui.Screen, font *willow.FontFamily, size float64, label string, x, y float64) {
	node := willow.NewText("sec-"+label, label, font)
	node.TextBlock.FontSize = size
	node.TextBlock.Color = willow.RGBA(0.45, 0.50, 0.70, 1)
	node.SetPosition(x, y)
	screen.AddNode(node)
}
