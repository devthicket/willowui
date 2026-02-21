// Tag demonstrates the WillowUI Tag and TagBar widgets: pill-shaped
// category labels with optional remove (×) and toggle modes, plus a
// tag-input bar where pressing Space creates chips.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 600
	screenH = 480
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		titleSize = 18.0
		labelSize = 12.0
		tagSize   = 12.0
	)

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	titleNode := willow.NewText("title", "WillowUI: Tag & TagBar", font)
	titleNode.TextBlock.FontSize = titleSize
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetPosition(24, 18)
	screen.AddNode(titleNode)

	// ── Static tags with variants ──────────────────────────────────────────
	y := 58.0
	addSection(screen, font, labelSize, "STATIC TAGS", 24, y)
	y += 22

	x := 24.0
	for _, tc := range []struct {
		name    string
		variant ui.Variant
	}{
		{"Go", ui.Primary},
		{"Rust", ui.Success},
		{"Python", ui.Warning},
		{"Ruby", ui.Danger},
		{"C++", ui.Neutral},
	} {
		tag := ui.NewTag("tag-"+tc.name, font, tagSize)
		tag.SetText(tc.name)
		tag.SetVariant(tc.variant)
		tag.SizeToContent()
		tag.SetPosition(x, y)
		screen.Add(tag)
		x += tag.Width + 8
	}

	// ── Removable tags ──────────────────────────────────────────────────────
	y += 40
	addSection(screen, font, labelSize, "REMOVABLE", 24, y)
	y += 22

	x = 24.0
	for _, name := range []string{"TypeScript", "React", "Tailwind"} {
		chip := ui.NewTag("chip-"+name, font, tagSize)
		chip.SetText(name)
		chip.SetRemovable(true)
		chip.SetVariant(ui.Primary)
		chip.SizeToContent()
		chip.SetPosition(x, y)
		localChip := chip
		chip.SetOnRemove(func() {
			screen.Remove(localChip)
		})
		screen.Add(chip)
		x += chip.Width + 8
	}

	// ── Selectable (toggle) tags ────────────────────────────────────────────
	y += 40
	addSection(screen, font, labelSize, "SELECTABLE", 24, y)
	y += 22

	x = 24.0
	for _, name := range []string{"Open Source", "MIT License", "Active"} {
		toggle := ui.NewTag("sel-"+name, font, tagSize)
		toggle.SetText(name)
		toggle.SetSelectable(true)
		toggle.SizeToContent()
		toggle.SetPosition(x, y)
		localName := name
		toggle.SetOnToggle(func(selected bool) {
			fmt.Printf("toggle %q → %v\n", localName, selected)
		})
		screen.Add(toggle)
		x += toggle.Width + 8
	}

	// ── TagBar ──────────────────────────────────────────────────────────────
	y += 50
	addSection(screen, font, labelSize, "TAG BAR (type + Space)", 24, y)
	y += 22

	bar := ui.NewTagBar("tagbar", font, tagSize)
	bar.SetPlaceholder("Add tags...")
	bar.SetSize(400, 36)
	bar.SetPosition(24, y)
	bar.SetOnChange(func(tags []string) {
		fmt.Printf("tags changed: %v\n", tags)
	})
	screen.Add(bar)

	// Pre-populated TagBar to show tags rendering.
	y += 50
	addSection(screen, font, labelSize, "TAG BAR (pre-populated)", 24, y)
	y += 22
	bar2 := ui.NewTagBar("tagbar2", font, tagSize)
	bar2.SetSize(400, 36)
	bar2.SetPosition(24, y)
	bar2.SetTags([]string{"Go", "Rust", "TypeScript"})
	screen.Add(bar2)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Tag & TagBar — WillowUI",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.12, 0.12, 0.15, 1),
	})
}

func addSection(screen *ui.Screen, font *willow.FontFamily, size float64, text string, x, y float64) {
	lbl := ui.NewLabel("section-"+text, text, font, size)
	lbl.SetColor(willow.RGBA(0.45, 0.55, 0.65, 1))
	lbl.SetPosition(x, y)
	screen.Add(lbl)
}
