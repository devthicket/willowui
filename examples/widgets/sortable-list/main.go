// SortableList example — demonstrates drag-handle reordering and keyboard
// reorder commands (Alt+Up / Alt+Down) with a reactive array backing.
// Two lists are shown: a quest priority list and a loadout ordering list.
package main

import (
	"fmt"
	"strings"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 560
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 15.0
		sizeSmall  = 13.0
	)

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "SortableList", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	// ── Quest Priority List ─────────────────────────────────────────────────

	questHeader := willow.NewText("quest-header", "Quest Priority", font)
	questHeader.TextBlock.FontSize = sizeMedium
	questHeader.TextBlock.Color = willow.RGBA(0.75, 0.85, 1.0, 1)
	questHeader.SetPosition(28, 56)
	screen.AddNode(questHeader)

	quests := ui.NewArrayFrom([]string{
		"Defeat the Dragon",
		"Find the Lost Artifact",
		"Escort the Merchant",
		"Clear the Dungeon",
		"Gather Herbs",
		"Rescue the Villagers",
		"Forge the Legendary Sword",
		"Map the Underground",
	})

	sl := ui.NewSortableList("quests", 34)
	sl.SetSize(260, 340)
	sl.SetPosition(20, 78)
	sl.SetSelected(0)

	sl.SetRenderItem(func(index int, data any) *ui.Component {
		text := data.(string)
		lbl := ui.NewLabel("item", fmt.Sprintf("%d. %s", index+1, text), font, sizeSmall)
		return &lbl.Component
	})

	ui.BindSortableListItems(sl, quests)

	// Status label — shows feedback after reorder.
	questStatus := ui.NewLabel("quest-status", "Click or drag to reorder", font, sizeSmall)
	questStatus.SetPosition(20, 426)
	screen.Add(questStatus)

	sl.SetOnReorder(func(from, to int) {
		item := quests.At(to)
		questStatus.SetText(fmt.Sprintf("Moved \"%s\" to #%d", item, to+1))
	})

	screen.Add(sl)

	// ── Loadout List ────────────────────────────────────────────────────────

	loadoutHeader := willow.NewText("loadout-header", "Loadout Order", font)
	loadoutHeader.TextBlock.FontSize = sizeMedium
	loadoutHeader.TextBlock.Color = willow.RGBA(1.0, 0.85, 0.55, 1)
	loadoutHeader.SetPosition(308, 56)
	screen.AddNode(loadoutHeader)

	loadout := ui.NewArrayFrom([]string{
		"Iron Sword",
		"Wooden Shield",
		"Health Potion",
		"Fire Scroll",
		"Rope",
		"Torch",
	})

	sl2 := ui.NewSortableList("loadout", 34)
	sl2.SetSize(220, 250)
	sl2.SetPosition(300, 78)
	sl2.SetHandleSide(ui.SortHandleRight)
	sl2.SetSelected(0)

	sl2.SetRenderItem(func(index int, data any) *ui.Component {
		text := data.(string)
		lbl := ui.NewLabel("item", text, font, sizeSmall)
		return &lbl.Component
	})

	ui.BindSortableListItems(sl2, loadout)
	screen.Add(sl2)

	// ── Order Readout ───────────────────────────────────────────────────────

	orderLabel := ui.NewLabel("order", "", font, sizeSmall)
	orderLabel.SetPosition(548, 78)
	screen.Add(orderLabel)

	updateOrder := func() {
		var parts []string
		quests.ForEach(func(i int, v string) {
			parts = append(parts, fmt.Sprintf("%d. %s", i+1, v))
		})
		orderLabel.SetText("Quest Order:\n" + strings.Join(parts, "\n"))
	}
	updateOrder()

	sl.SetOnReorder(func(from, to int) {
		item := quests.At(to)
		questStatus.SetText(fmt.Sprintf("Moved \"%s\" to #%d", item, to+1))
		updateOrder()
	})

	// ── Instructions ────────────────────────────────────────────────────────

	instrLabel := ui.NewLabel("instructions",
		"Controls:\n"+
			"  Up/Down        Navigate\n"+
			"  Alt+Up/Down    Move item\n"+
			"  Drag handle    Reorder\n\n"+
			"Left list: handles on left\n"+
			"Right list: handles on right",
		font, sizeSmall)
	instrLabel.SetPosition(548, 280)
	screen.Add(instrLabel)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "SortableList Example",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.1, 0.1, 0.12, 1),
	})
}
