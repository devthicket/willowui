// ToggleButtonBar demonstrates WillowUI's ToggleButtonBar component: a
// segmented control for single-selection among labeled buttons. A label
// below the bar shows the currently selected option.
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

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 24.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "WillowUI: ToggleButtonBar Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW-48)
	div.SetPosition(24, 48)
	screen.AddNode(div)

	// Section: ToggleButtonBar.
	addSectionLabel(screen, font, sizeSmall, "ToggleButtonBar: single-selection segmented control", 24, 58)

	bar := ui.NewToggleButtonBar("demo-bar", font, sizeMedium)
	bar.AddButton("Option A")
	bar.AddButton("Option B")
	bar.AddButton("Option C")
	bar.AddButton("Option D")
	bar.SetSize(500, 40)
	bar.SetPosition(24, 80)
	screen.Add(bar)

	// Selection label.
	selText := ui.NewRef("Selected: Option A (index 0)")
	selLabel := ui.NewLabel("sel-label", "", font, sizeMedium)
	selLabel.BindText(selText)
	selLabel.SetPosition(24, 140)
	screen.Add(selLabel)

	options := []string{"Option A", "Option B", "Option C", "Option D"}
	bar.SetOnChange(func(idx int) {
		name := options[idx]
		selText.Set(fmt.Sprintf("Selected: %s (index %d)", name, idx))
	})

	// Section: Second bar with reactive binding.
	addSectionLabel(screen, font, sizeSmall, "Reactive binding: selection bound to Ref[int]", 24, 180)

	bar2 := ui.NewToggleButtonBar("demo-bar2", font, sizeMedium)
	bar2.AddButton("Small")
	bar2.AddButton("Medium")
	bar2.AddButton("Large")
	bar2.SetSize(400, 36)
	bar2.SetPosition(24, 200)
	screen.Add(bar2)

	sizeRef := ui.NewRef(1) // start at "Medium"
	bar2.BindSelected(sizeRef)

	sizeLabel := ui.NewLabel("size-label", "Size: Medium", font, sizeMedium)
	sizeLabel.SetPosition(24, 250)
	screen.Add(sizeLabel)

	sizes := []string{"Small", "Medium", "Large"}
	bar2.SetOnChange(func(idx int) {
		sizeLabel.SetText(fmt.Sprintf("Size: %s", sizes[idx]))
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — ToggleButtonBar Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addSectionLabel(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("section", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
