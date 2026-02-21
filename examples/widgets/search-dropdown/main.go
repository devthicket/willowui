// Search Dropdown — autocomplete demo.
// SearchBox drives a MenuPopup with live suggestions.
// Down/Up arrow: navigate. Enter or click: select. Esc: dismiss.
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 620
	screenH = 460
	colX    = 40.0
	sbWidth = 380.0
)

var catalog = []string{
	"Potion", "Hi-Potion", "X-Potion", "Mega-Potion",
	"Phoenix Down", "Mega Phoenix", "Elixir", "Megalixir",
	"Ether", "Turbo Ether", "Tent", "Cottage",
	"Antidote", "Eye Drops", "Echo Screen", "Maiden's Kiss",
	"Remedy", "Soft", "Holy Water", "Vaccine",
}

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 13.0
	)

	screen := ui.NewScreen()

	// ── Title & divider ──────────────────────────────────────────────────────
	title := willow.NewText("title", "Search Dropdown", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	// ── Hint ─────────────────────────────────────────────────────────────────
	hint := willow.NewText("hint", "Type to filter   ↓/↑ navigate   ↩ or click to select   Esc dismiss", font)
	hint.TextBlock.FontSize = sizeSmall
	hint.TextBlock.Color = willow.RGBA(0.5, 0.5, 0.5, 1)
	hint.SetPosition(colX, 70)
	screen.AddNode(hint)

	// ── Picked label ─────────────────────────────────────────────────────────
	pickedNode := willow.NewText("picked", "Nothing selected", font)
	pickedNode.TextBlock.FontSize = sizeMedium
	pickedNode.TextBlock.Color = willow.RGBA(0.75, 0.75, 0.75, 1)
	pickedNode.SetPosition(colX, 200)
	screen.AddNode(pickedNode)

	countNode := willow.NewText("count", "", font)
	countNode.TextBlock.FontSize = sizeSmall
	countNode.TextBlock.Color = willow.RGBA(0.45, 0.45, 0.45, 1)
	countNode.SetPosition(colX, 228)
	screen.AddNode(countNode)

	// ── SearchBox ────────────────────────────────────────────────────────────
	results := ui.NewArray[string]()
	popup := ui.NewMenuPopup(font, sizeMedium)

	sb := ui.NewSearchBox("search", font, sizeMedium)

	showPopup := func(items []string) {
		menuItems := make([]ui.MenuItem, len(items))
		for i, item := range items {
			label := item
			menuItems[i] = ui.MenuItem{
				Label: label,
				OnSelect: func() {
					pickedNode.SetContent(fmt.Sprintf("Selected: %s", label))
					sb.SetValue(label)
					sb.CancelPendingSearch()
				},
			}
		}
		popup.SetItems(menuItems)
	}
	sb.SetWidth(sbWidth)
	sb.SetPlaceholder("Type to search...")
	sb.SetDebounce(80 * time.Millisecond)

	ui.SetSearchBoxFunc(sb, results, func(q string) []string {
		q = strings.ToLower(q)
		var out []string
		for _, item := range catalog {
			if strings.Contains(strings.ToLower(item), q) {
				out = append(out, item)
			}
		}
		return out
	})

	sb.SetOnSearchStart(func(_ string) {
		countNode.SetContent("")
	})
	sb.SetOnSearchFinish(func(q string, count int) {
		countNode.SetContent(fmt.Sprintf("%d match(es)", count))

		suggestions := make([]string, results.Len())
		results.ForEach(func(i int, item string) { suggestions[i] = item })
		showPopup(suggestions)
		ui.DefaultMenuPopupManager.Show(popup, &sb.Component)
	})
	sb.SetOnSearchEmpty(func(_ string) {
		countNode.SetContent("No matches")
		ui.DefaultMenuPopupManager.Hide()
	})
	sb.SetOnClear(func() {
		countNode.SetContent("")
		ui.DefaultMenuPopupManager.Hide()
	})

	sb.Component.OffsetX = colX
	sb.Component.OffsetY = 96
	screen.Add(&sb.Component)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Search Dropdown",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.09, 0.11, 1),
	})
}
