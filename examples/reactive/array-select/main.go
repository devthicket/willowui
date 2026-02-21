// array-select demonstrates reactive Array bindings on selection widgets.
//
// Patterns shown:
//   - OptionRotator.BindOptions: cycle through a live-updated options array
//   - Select.BindOptions + BindSelected: dropdown with reactive options and
//     two-way selection binding via Ref[int]
//   - ToggleButtonBar.BindButtons: segmented control rebuilt from a live array
//   - Batch mutations: add/remove several options atomically
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 860
	screenH  = 560
	fontSize = 14.0
	colL     = 40.0
	colR     = 470.0
)

func main() {
	font := ui.MustLoadDefaultFont()

	screen := ui.NewScreen()

	// Title.
	titleNode := willow.NewText("title", "Reactive: Array Bindings: Selection Widgets", font)
	titleNode.TextBlock.FontSize = 18
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetPosition(24, 12)
	screen.AddNode(titleNode)
	screen.AddNode(divider("top-div", screenW-48, 40))

	y := 56.0

	// ── Shared options array ──────────────────────────────────────────────────
	// All three widgets in the left section share the same difficulty options.
	diffOptions := ui.NewArrayFrom([]string{"Easy", "Normal", "Hard", "Nightmare"})

	// ── Section 1: OptionRotator.BindOptions ─────────────────────────────────
	sectionLabel(screen, font, "OptionRotator.BindOptions", colL, y)
	y += 20

	diffIdx := ui.NewRef(1)

	rotator := ui.NewOptionRotator("diff-rotator", []string{"placeholder"}, font, fontSize)
	rotator.SetSize(280, 32)
	rotator.SetPosition(colL, y)
	rotator.BindOptions(diffOptions) // replaces the placeholder list
	rotator.BindSelected(diffIdx)
	screen.Add(rotator)

	rotatorStatus := ui.NewLabel("rot-status", "", font, fontSize)
	rotatorStatus.SetColor(willow.RGBA(0.7, 0.85, 1, 1))
	rotatorStatus.SetPosition(colL, y+38)
	screen.Add(rotatorStatus)
	ui.WatchEffect(func() {
		rotatorStatus.SetText(fmt.Sprintf("selected index: %d  (%d options)", diffIdx.Get(), diffOptions.Len()))
	})

	y += 72
	screen.AddNode(divider("d1", screenW-80, y))
	y += 14

	// ── Section 2: Select.BindOptions + BindSelected ─────────────────────────
	sectionLabel(screen, font, "Select.BindOptions + BindSelected  (same array + shared Ref)", colL, y)
	y += 20

	// diffIdx is shared — changing the rotator also moves this dropdown, and vice versa.
	dropdown := ui.NewSelect("diff-select", nil, font, fontSize)
	dropdown.SetSize(280, 32)
	dropdown.SetPosition(colL, y)
	dropdown.BindOptions(ui.NewArrayFrom(toSelectOpts(diffOptions)))
	dropdown.BindSelected(diffIdx)
	// Keep dropdown options in sync with diffOptions array changes.
	selectOpts := ui.NewArray[ui.SelectOption]()
	for i := 0; i < diffOptions.Len(); i++ {
		selectOpts.Push(ui.SelectOption{Label: diffOptions.At(i)})
	}
	diffOptions.OnChange(func() {
		selectOpts.Set(toSelectOpts(diffOptions))
	})
	dropdown.BindOptions(selectOpts)
	screen.Add(dropdown)

	selectStatus := ui.NewLabel("sel-status", "", font, fontSize)
	selectStatus.SetColor(willow.RGBA(0.7, 0.9, 0.7, 1))
	selectStatus.SetPosition(colL, y+38)
	screen.Add(selectStatus)
	ui.WatchEffect(func() {
		selectStatus.SetText(fmt.Sprintf(
			"diffIdx Ref = %d  (rotator and dropdown share this ref)", diffIdx.Get(),
		))
	})

	y += 72
	screen.AddNode(divider("d2", screenW-80, y))
	y += 14

	// ── Section 3: ToggleButtonBar.BindButtons ───────────────────────────────
	sectionLabel(screen, font, "ToggleButtonBar.BindButtons", colL, y)
	y += 20

	// Separate array for the tab bar to show independent mutation.
	tabLabels := ui.NewArrayFrom([]string{"Overview", "Settings", "Logs"})
	tabIdx := ui.NewRef(0)

	bar := ui.NewToggleButtonBar("view-bar", font, fontSize)
	bar.SetSize(380, 36)
	bar.SetPosition(colL, y)
	bar.BindButtons(tabLabels)
	bar.BindSelected(tabIdx)
	screen.Add(bar)

	barStatus := ui.NewLabel("bar-status", "", font, fontSize)
	barStatus.SetColor(willow.RGBA(1, 0.8, 0.6, 1))
	barStatus.SetPosition(colL, y+44)
	screen.Add(barStatus)
	ui.WatchEffect(func() {
		barStatus.SetText(fmt.Sprintf("tabIdx = %d  (%d buttons)", tabIdx.Get(), tabLabels.Len()))
	})

	y += 80

	// ── Right column: mutation controls ──────────────────────────────────────
	// Controls for diffOptions (affects rotator + dropdown).
	ry := 56.0
	sectionLabel(screen, font, "Mutate diffOptions -- rotator + dropdown update live", colR, ry)
	ry += 20

	addDiffBtn := ui.NewButton("add-diff", "+ Add Option", font, fontSize)
	addDiffBtn.SetSize(130, 28)
	addDiffBtn.SetPosition(colR, ry)
	addDiffBtn.SetOnClick(func() {
		n := diffOptions.Len() + 1
		newOpt := fmt.Sprintf("Level %d", n)
		diffOptions.Push(newOpt)
	})
	screen.Add(addDiffBtn)

	removeFirstBtn := ui.NewButton("remove-first-diff", "Remove First", font, fontSize)
	removeFirstBtn.SetSize(130, 28)
	removeFirstBtn.SetPosition(colR+138, ry)
	removeFirstBtn.SetOnClick(func() {
		if diffOptions.Len() > 1 {
			diffOptions.Shift()
		}
	})
	screen.Add(removeFirstBtn)

	ry += 36

	shuffleDiffBtn := ui.NewButton("shuffle-diff", "Shuffle", font, fontSize)
	shuffleDiffBtn.SetSize(130, 28)
	shuffleDiffBtn.SetPosition(colR, ry)
	shuffleDiffBtn.SetOnClick(func() { diffOptions.Shuffle() })
	screen.Add(shuffleDiffBtn)

	// Batch: replace all options at once.
	batchDiffBtn := ui.NewButton("batch-diff", "Reset to defaults", font, fontSize)
	batchDiffBtn.SetSize(138, 28)
	batchDiffBtn.SetPosition(colR+138, ry)
	batchDiffBtn.SetOnClick(func() {
		diffOptions.Set([]string{"Easy", "Normal", "Hard", "Nightmare"})
		diffIdx.Set(0)
	})
	screen.Add(batchDiffBtn)

	ry += 48
	screen.AddNode(divider("d3", screenW-80-colR+40, ry))
	ry += 14

	// Controls for tabLabels (affects ToggleButtonBar).
	sectionLabel(screen, font, "Mutate tabLabels -- ToggleButtonBar rebuilds live", colR, ry)
	ry += 20

	addTabBtn := ui.NewButton("add-tab", "+ Add Tab", font, fontSize)
	addTabBtn.SetSize(130, 28)
	addTabBtn.SetPosition(colR, ry)
	addTabBtn.SetOnClick(func() {
		tabLabels.Push(fmt.Sprintf("Tab %d", tabLabels.Len()+1))
	})
	screen.Add(addTabBtn)

	removeLastTabBtn := ui.NewButton("remove-last-tab", "Remove Last", font, fontSize)
	removeLastTabBtn.SetSize(130, 28)
	removeLastTabBtn.SetPosition(colR+138, ry)
	removeLastTabBtn.SetOnClick(func() {
		if tabLabels.Len() > 1 {
			tabLabels.Pop()
		}
	})
	screen.Add(removeLastTabBtn)

	ry += 36

	// Batch: add three tabs at once without triggering three separate rebuilds.
	batchTabsBtn := ui.NewButton("batch-tabs", "Batch: add 3 tabs", font, fontSize)
	batchTabsBtn.SetSize(200, 28)
	batchTabsBtn.SetPosition(colR, ry)
	batchTabsBtn.SetOnClick(func() {
		base := tabLabels.Len() + 1
		tabLabels.Batch(func() {
			tabLabels.Push(fmt.Sprintf("Extra %d", base))
			tabLabels.Push(fmt.Sprintf("Extra %d", base+1))
			tabLabels.Push(fmt.Sprintf("Extra %d", base+2))
		})
	})
	screen.Add(batchTabsBtn)

	resetTabsBtn := ui.NewButton("reset-tabs", "Reset", font, fontSize)
	resetTabsBtn.SetSize(90, 28)
	resetTabsBtn.SetPosition(colR+208, ry)
	resetTabsBtn.SetOnClick(func() {
		tabLabels.Set([]string{"Overview", "Settings", "Logs"})
		tabIdx.Set(0)
	})
	screen.Add(resetTabsBtn)

	_ = y // suppress unused warning — y used for left column layout above

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Reactive Array: Selection Widgets",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.07, 0.08, 0.10, 1),
	})
}

func toSelectOpts(arr *ui.Array[string]) []ui.SelectOption {
	opts := make([]ui.SelectOption, arr.Len())
	arr.ForEach(func(i int, s string) {
		opts[i] = ui.SelectOption{Label: s}
	})
	return opts
}

func sectionLabel(screen *ui.Screen, font *willow.FontFamily, text string, x, y float64) {
	n := willow.NewText("sec-"+text[:8], text, font)
	n.TextBlock.FontSize = 12
	n.TextBlock.Color = willow.RGBA(0.45, 0.55, 0.68, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}

func divider(name string, w, y float64) *willow.Node {
	d := willow.NewSprite(name, willow.TextureRegion{})
	d.SetPosition(40, y)
	d.SetScale(w, 1)
	d.SetColor(willow.RGBA(0.18, 0.23, 0.28, 1))
	return d
}
