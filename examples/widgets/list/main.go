// List - reactive demo.
// Shows List.BindSelected(Ref[int]).
// A second list is bound to the same Ref so both selections stay in sync.
// A detail label is driven by a Computed that reads the selection.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 480
)

var fruits = []string{
	"Apple", "Banana", "Cherry", "Date", "Elderberry",
	"Fig", "Grape", "Honeydew", "Kiwi", "Lemon",
}

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	title := willow.NewText("title", "Reactive - List", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	// Shared selection Ref.
	selRef := ui.NewRef(-1)

	items := make([]ui.ListItem, len(fruits))
	for i, f := range fruits {
		items[i] = ui.ListItem{Data: f}
	}

	renderItem := func(index int, data any) *ui.Component {
		text := data.(string)
		lbl := ui.NewLabel("item", text, font, sizeMedium)
		return &lbl.Component
	}

	// ── List A ────────────────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "List A", 40, 58)

	listA := ui.NewList("list-a", 28)
	listA.SetSize(200, 200)
	listA.SetSelectable(true)
	listA.SetRenderItem(renderItem)
	listA.SetItems(items)
	listA.BindSelected(selRef)
	listA.SetPosition(40, 78)
	screen.Add(listA)

	// ── List B - same Ref ─────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "List B (same Ref)", 260, 58)

	listB := ui.NewList("list-b", 28)
	listB.SetSize(200, 200)
	listB.SetSelectable(true)
	listB.SetRenderItem(renderItem)
	listB.SetItems(items)
	listB.BindSelected(selRef)
	listB.SetPosition(260, 78)
	screen.Add(listB)

	// ── Detail panel ──────────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Detail - Computed from Ref", 480, 58)

	detail := ui.NewComputed(func() string {
		idx := selRef.Get()
		if idx < 0 {
			return "(nothing selected)"
		}
		return fmt.Sprintf("#%d - %s\nindex: %d\nlen(name): %d",
			idx+1, fruits[idx], idx, len(fruits[idx]))
	})

	detailLbl := ui.NewLabel("detail", "", font, sizeMedium)
	detailLbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	detailLbl.SetWrapWidth(260)
	detailLbl.SetPosition(480, 78)
	screen.Add(detailLbl)
	ui.WatchEffect(func() {
		detailLbl.SetText(detail.Get())
	})

	// ── Navigation buttons ────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Navigate via Ref", 40, 300)

	prevBtn := ui.NewButton("prev", "← Prev", font, sizeSmall)
	prevBtn.SetSize(90, 30)
	prevBtn.SetPosition(40, 320)
	screen.Add(prevBtn)
	prevBtn.SetOnClick(func() {
		cur := selRef.Peek()
		if cur > 0 {
			selRef.Set(cur - 1)
		} else {
			selRef.Set(len(fruits) - 1)
		}
	})

	nextBtn := ui.NewButton("next", "Next →", font, sizeSmall)
	nextBtn.SetSize(90, 30)
	nextBtn.SetPosition(140, 320)
	screen.Add(nextBtn)
	nextBtn.SetOnClick(func() {
		cur := selRef.Peek()
		selRef.Set((cur + 1) % len(fruits))
	})

	clearBtn := ui.NewButton("clear", "Clear", font, sizeSmall)
	clearBtn.SetSize(80, 30)
	clearBtn.SetPosition(240, 320)
	screen.Add(clearBtn)
	clearBtn.SetOnClick(func() {
		selRef.Set(-1)
	})

	noteLbl := ui.NewLabel("note", "Both lists track the same Ref - selecting or navigating updates both.", font, sizeSmall)
	noteLbl.SetColor(willow.RGBA(0.5, 0.55, 0.6, 1))
	noteLbl.SetWrapWidth(460)
	noteLbl.SetPosition(40, 360)
	screen.Add(noteLbl)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive - List",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addHeader(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("hdr", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.45, 0.55, 0.65, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
