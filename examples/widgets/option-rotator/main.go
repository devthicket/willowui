// OptionRotator demonstrates WillowUI's OptionRotator component: a compact
// left/right chevron control for cycling through a fixed list of options.
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
	title := willow.NewText("title", "WillowUI: OptionRotator Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW-48)
	div.SetPosition(24, 52)
	screen.AddNode(div)

	// -----------------------------------------------------------------------
	// Section 1: Basic usage with onChange callback.
	// -----------------------------------------------------------------------
	addSectionLabel(screen, font, sizeSmall, "Basic: onChange callback", 24, 64)

	diff := ui.NewOptionRotator("difficulty", []string{"Easy", "Normal", "Hard", "Nightmare"}, font, sizeMedium)
	diff.SetSelected(1) // start on Normal
	diff.SetSize(220, 32)
	diff.SetPosition(24, 82)
	screen.Add(diff)

	diffStatus := ui.NewRef("Difficulty: Normal (index 1)")
	diffLabel := ui.NewLabel("diff-label", "", font, sizeMedium)
	diffLabel.BindText(diffStatus)
	diffLabel.SetPosition(260, 90)
	screen.Add(diffLabel)

	diff.SetOnChange(func(idx int, val string) {
		diffStatus.Set(fmt.Sprintf("Difficulty: %s (index %d)", val, idx))
	})

	// -----------------------------------------------------------------------
	// Section 2: Reactive binding to Ref[int].
	// -----------------------------------------------------------------------
	addSectionLabel(screen, font, sizeSmall, "BindSelected: two-way Ref[int]", 24, 140)

	qualityRef := ui.NewRef(2) // starts at "High"
	quality := ui.NewOptionRotator("quality", []string{"Low", "Medium", "High", "Ultra"}, font, sizeMedium)
	quality.BindSelected(qualityRef)
	quality.SetSize(220, 32)
	quality.SetPosition(24, 158)
	screen.Add(quality)

	qualLabel := ui.NewLabel("qual-label", "", font, sizeMedium)
	qualLabel.SetPosition(260, 166)
	screen.Add(qualLabel)

	quality.SetOnChange(func(idx int, val string) {
		qualLabel.SetText(fmt.Sprintf("ref = %d  value = %q", qualityRef.Peek(), val))
	})
	qualLabel.SetText(fmt.Sprintf("ref = %d  value = %q", qualityRef.Peek(), quality.Value()))

	// -----------------------------------------------------------------------
	// Section 3: Reactive binding to Ref[string].
	// -----------------------------------------------------------------------
	addSectionLabel(screen, font, sizeSmall, "BindValue: two-way Ref[string]", 24, 218)

	resRef := ui.NewRef("1920×1080")
	res := ui.NewOptionRotator("resolution", []string{
		"1280×720", "1920×1080", "2560×1440", "3840×2160",
	}, font, sizeMedium)
	res.BindValue(resRef)
	res.SetSize(220, 32)
	res.SetPosition(24, 236)
	screen.Add(res)

	resLabel := ui.NewLabel("res-label", "", font, sizeMedium)
	resLabel.SetPosition(260, 244)
	screen.Add(resLabel)

	res.SetOnChange(func(_ int, val string) {
		resLabel.SetText(fmt.Sprintf("ref = %q", resRef.Peek()))
	})
	resLabel.SetText(fmt.Sprintf("ref = %q", resRef.Peek()))

	// -----------------------------------------------------------------------
	// Section 4: No-wrap — chevrons disable at the ends.
	// -----------------------------------------------------------------------
	addSectionLabel(screen, font, sizeSmall, "SetWrap(false): disabled chevrons at boundaries", 24, 296)

	page := ui.NewOptionRotator("page", makePageLabels(8), font, sizeMedium)
	page.SetWrap(false)
	page.SetSize(220, 32)
	page.SetPosition(24, 314)
	screen.Add(page)

	pageLabel := ui.NewLabel("page-label", "Page 1 of 8", font, sizeMedium)
	pageLabel.SetPosition(260, 322)
	screen.Add(pageLabel)

	page.SetOnChange(func(idx int, _ string) {
		pageLabel.SetText(fmt.Sprintf("Page %d of 8", idx+1))
	})

	// -----------------------------------------------------------------------
	// Section 5: Disabled state.
	// -----------------------------------------------------------------------
	addSectionLabel(screen, font, sizeSmall, "SetEnabled(false): entire widget disabled", 24, 374)

	frozen := ui.NewOptionRotator("frozen", []string{"Spring", "Summer", "Autumn", "Winter"}, font, sizeMedium)
	frozen.SetSelected(1)
	frozen.SetEnabled(false)
	frozen.SetSize(220, 32)
	frozen.SetPosition(24, 392)
	screen.Add(frozen)

	frozenLabel := ui.NewLabel("frozen-label", "This rotator is disabled", font, sizeMedium)
	frozenLabel.SetPosition(260, 400)
	frozenLabel.SetColor(willow.RGBA(0.5, 0.5, 0.5, 1))
	screen.Add(frozenLabel)

	// -----------------------------------------------------------------------
	// Footer hint.
	// -----------------------------------------------------------------------
	hint := willow.NewText("hint", "Tab to focus · Left/Right arrows to cycle", font)
	hint.TextBlock.FontSize = sizeSmall
	hint.TextBlock.Color = willow.RGBA(0.35, 0.45, 0.55, 1)
	hint.SetPosition(24, screenH-28)
	screen.AddNode(hint)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — OptionRotator Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func makePageLabels(n int) []string {
	labels := make([]string, n)
	for i := range labels {
		labels[i] = fmt.Sprintf("Page %d", i+1)
	}
	return labels
}

func addSectionLabel(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("section", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
