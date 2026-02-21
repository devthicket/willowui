// Filter Search — demo.
// Shows a SearchBox filtering a list in real time.
// Features: search icon, clear button, debounce, result count, lifecycle labels.
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 720
	screenH = 600
	colX    = 40.0
)

var allItems = []string{
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

	// Title.
	title := willow.NewText("title", "Filter Search", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	y := 70.0

	addLabel := func(text string, x, yy float64, size, alpha float64) *willow.Node {
		n := willow.NewText(text, text, font)
		n.TextBlock.FontSize = size
		n.TextBlock.Color = willow.RGBA(1, 1, 1, alpha)
		n.SetPosition(x, yy)
		screen.AddNode(n)
		return n
	}

	// Status label.
	statusNode := willow.NewText("status", fmt.Sprintf("Showing all %d items", len(allItems)), font)
	statusNode.TextBlock.FontSize = sizeSmall
	statusNode.TextBlock.Color = willow.RGBA(0.7, 0.7, 0.7, 1)
	statusNode.SetPosition(colX, y)
	screen.AddNode(statusNode)
	y += 26

	// SearchBox.
	results := ui.NewArray[string]()

	sb := ui.NewSearchBox("search", font, sizeMedium)
	sb.SetWidth(400)
	sb.SetPlaceholder("Search items...")
	sb.SetDebounce(120 * time.Millisecond)

	ui.SetSearchBoxFunc(sb, results, func(q string) []string {
		q = strings.ToLower(q)
		var out []string
		for _, item := range allItems {
			if strings.Contains(strings.ToLower(item), q) {
				out = append(out, item)
			}
		}
		return out
	})

	sb.SetOnSearchStart(func(q string) {
		statusNode.SetContent(fmt.Sprintf("Searching for %q...", q))
	})
	sb.SetOnSearchFinish(func(q string, count int) {
		statusNode.SetContent(fmt.Sprintf("%d result(s) for %q", count, q))
	})
	sb.SetOnSearchEmpty(func(q string) {
		statusNode.SetContent(fmt.Sprintf("No results for %q", q))
	})
	sb.SetOnClear(func() {
		statusNode.SetContent(fmt.Sprintf("Cleared - showing all %d items", len(allItems)))
		results.Batch(func() {
			results.Clear()
			for _, item := range allItems {
				results.Push(item)
			}
		})
	})

	sb.Component.OffsetX = colX
	sb.Component.OffsetY = y
	screen.Add(&sb.Component)
	y += sb.Height + 16

	// Result count readout.
	countNode := addLabel(fmt.Sprintf("Results: %d / %d", len(allItems), len(allItems)), colX, y, sizeSmall, 0.6)
	y += 22

	// Result list using a willow List.
	list := ui.NewList("results-list", 26)
	list.SetSize(400, 280)
	list.SetSelectable(true)
	list.Component.OffsetX = colX
	list.Component.OffsetY = y
	screen.Add(&list.Component)

	list.SetRenderItem(func(_ int, data any) *ui.Component {
		lbl := ui.NewLabel("item", data.(string), font, sizeMedium)
		return &lbl.Component
	})

	list.SetOnChange(func(idx int) {
		if item := list.SelectedItem(); item != nil {
			statusNode.SetContent(fmt.Sprintf("Selected: %s", item.(string)))
		}
	})

	// Wire results → list.
	buildListItems := func() []ui.ListItem {
		items := make([]ui.ListItem, results.Len())
		results.ForEach(func(i int, item string) {
			items[i] = ui.ListItem{Data: item}
		})
		return items
	}
	results.OnReplaced(func() {
		list.SetItems(buildListItems())
		countNode.SetContent(fmt.Sprintf("Results: %d / %d", results.Len(), len(allItems)))
	})

	// Populate with all items initially (batch push → single OnReplaced notification).
	results.Batch(func() {
		for _, item := range allItems {
			results.Push(item)
		}
	})

	// Right column: show/hide options.
	rx := colX + 440.0
	ry := 70.0

	addLabel("Options", rx, ry, sizeMedium, 1.0)
	ry += 30

	showIconToggle := ui.NewToggle("show-icon")
	showIconToggle.SetValue(true)
	showIconToggle.SetOnChange(func(v bool) {
		sb.SetShowSearchIcon(v)
	})
	showIconToggle.Component.OffsetX = rx
	showIconToggle.Component.OffsetY = ry
	screen.Add(&showIconToggle.Component)

	iconLabel := addLabel("Show search icon", rx+54, ry+3, sizeSmall, 0.8)
	_ = iconLabel
	ry += 36

	showClearToggle := ui.NewToggle("show-clear")
	showClearToggle.SetValue(true)
	showClearToggle.SetOnChange(func(v bool) {
		sb.SetShowClearButton(v)
	})
	showClearToggle.Component.OffsetX = rx
	showClearToggle.Component.OffsetY = ry
	screen.Add(&showClearToggle.Component)

	clearLabel := addLabel("Show clear button", rx+54, ry+3, sizeSmall, 0.8)
	_ = clearLabel
	ry += 36

	// Debounce note.
	addLabel("Debounce: 120 ms", rx, ry, sizeSmall, 0.5)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Filter Search Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.09, 0.11, 1),
	})
}
