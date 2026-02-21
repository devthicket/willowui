// Theme-Gallery demonstrates switching between WillowUI themes at runtime.
// A toggle button bar at the top lets you pick from the available themes
// loaded from examples/_themes/. The selected theme is applied to a window
// containing representative UI components. A full-screen background panel
// updates its color on theme switch.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 640
	screenH = 520
)

type themeEntry struct {
	name  string
	theme *Theme // nil = default (no theme)
	clear willow.Color
}

type Theme = ui.Theme

func main() {
	_, src, _, _ := runtime.Caller(0)
	examplesDir := filepath.Dir(filepath.Dir(filepath.Dir(src)))

	// Load available themes from sibling example directories.
	themes := []themeEntry{
		{name: "Default", theme: nil, clear: willow.RGBA(0.15, 0.15, 0.17, 1)},
	}

	// Scan for .json theme files in the shared _themes directory.
	themesDir := filepath.Join(examplesDir, "_themes")
	entries, _ := os.ReadDir(themesDir)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := e.Name()[:len(e.Name())-5] // strip .json
		t, err := ui.LoadThemeFromFile(filepath.Join(themesDir, e.Name()))
		if err != nil {
			log.Printf("skip %s: %v", e.Name(), err)
			continue
		}
		themes = append(themes, themeEntry{name: name, theme: t, clear: willow.RGBA(0.1, 0.1, 0.12, 1)})
	}

	font := ui.MustLoadDefaultFont()

	screen := ui.NewScreen()

	// --- Full-screen background panel (updated on theme switch) ---
	bgPanel := ui.NewPanel("bg-panel")
	bgPanel.SetSize(screenW, screenH)
	bgPanel.SetBackground(themes[0].clear)
	bgPanel.SetInteractable(false)
	bgPanel.SetZIndex(-1)
	screen.Add(bgPanel)

	// --- Theme selector bar (unthemed, top-right corner) ---
	var tbbNames []string
	for _, t := range themes {
		tbbNames = append(tbbNames, t.name)
	}
	selector := ui.NewToggleButtonBar("theme-selector", font, 13)
	for _, name := range tbbNames {
		selector.AddButton(name)
	}
	selector.SetSelected(0)
	tbbW := float64(len(themes) * 90)
	selector.SetSize(tbbW, 28)
	selector.SetPosition(screenW-tbbW-8, 6)
	screen.Add(selector)

	// --- Demo window ---
	win := ui.NewWindow("gallery-win", "Theme Gallery", font, 16)
	win.SetSize(540, 440)
	win.SetResizable(true)
	win.SetPosition(50, 48)
	screen.Add(win)
	ui.DefaultWindowManager.Add(win)

	body := win.Body()
	body.SetLayout(ui.LayoutVBox)
	body.SetSpacing(10)
	body.SetAlignment(ui.AlignCenter)
	body.Padding = ui.Insets{Top: 30, Left: 16, Right: 16, Bottom: 16}

	// --- Tabs ---
	tabs := ui.NewTabBar("demo-tabs", font, 13)
	tabs.SetSize(500, 310)

	tabPad := ui.Insets{Top: 10, Left: 12, Right: 12, Bottom: 10}

	// Tab 1: Controls — buttons, toggle, checkbox, radio.
	controlsPage, _ := tabs.AddTabPage("Controls", ui.LayoutVBox, 10, tabPad)
	controlsPage.Align = ui.AlignStart

	btnRow := ui.NewHBox("btn-row")
	btnRow.Spacing = 12
	btnRow.Align = ui.AlignCenter
	btnRow.Width = 460
	btnRow.Height = 36

	primaryBtn := ui.NewButton("primary-btn", "Primary", font, 13)
	primaryBtn.SetSize(120, 32)
	btnRow.AddChild(primaryBtn)

	accentBtn := ui.NewButton("accent-btn", "Accent", font, 13)
	accentBtn.SetSize(120, 32)
	accentBtn.SetVariant(ui.Accent)
	btnRow.AddChild(accentBtn)

	dangerBtn := ui.NewButton("danger-btn", "Danger", font, 13)
	dangerBtn.SetSize(120, 32)
	dangerBtn.SetVariant(ui.Danger)
	btnRow.AddChild(dangerBtn)

	controlsPage.AddChild(btnRow)

	disBtn := ui.NewButton("dis-btn", "Disabled", font, 13)
	disBtn.SetSize(120, 32)
	disBtn.SetEnabled(false)
	controlsPage.AddChild(disBtn)

	chk := ui.NewCheckbox("demo-chk", "Enable Feature", font, 13)
	controlsPage.AddChild(chk)

	radioGroup := ui.NewRadio("demo-radio")
	radioGroup.AddOption("Option A", font, 13)
	radioGroup.AddOption("Option B", font, 13)
	radioGroup.AddOption("Option C", font, 13)
	radioGroup.SetSelected(0)
	controlsPage.AddChild(radioGroup)

	tglRow := ui.NewHBox("tgl-row")
	tglRow.Align = ui.AlignCenter
	tglRow.Spacing = 8
	tglRow.Width = 200
	tglRow.Height = 28
	tglLbl := ui.NewLabel("tgl-label", "Toggle", font, 13)
	tgl := ui.NewToggle("demo-toggle")
	tglRow.AddChild(tglLbl)
	tglRow.AddChild(tgl)
	controlsPage.AddChild(tglRow)

	controlsPage.UpdateLayout()

	// Tab 2: Range & Text — slider, progress bar, text input, text area.
	rangePage, _ := tabs.AddTabPage("Range & Text", ui.LayoutVBox, 10, tabPad)
	rangePage.Align = ui.AlignStart

	sliderLbl := ui.NewLabel("slider-label", "Slider", font, 13)
	rangePage.AddChild(sliderLbl)

	slider := ui.NewSlider("demo-slider")
	slider.SetRange(0, 100)
	slider.SetStep(1)
	slider.SetValue(42)
	slider.SetSize(260, 20)
	rangePage.AddChild(slider)

	sliderVal := ui.NewLabel("slider-val", "value: 42", font, 11)
	slider.SetOnChange(func(v float64) { sliderVal.SetText(fmt.Sprintf("value: %.0f", v)) })
	rangePage.AddChild(sliderVal)

	hpLbl := ui.NewLabel("hp-label", "Progress", font, 13)
	rangePage.AddChild(hpLbl)

	hp := ui.NewProgressBar("demo-progress")
	hp.SetValue(0.65)
	hp.SetSize(260, 18)
	hp.SetShowLabel(true, font, 11)
	rangePage.AddChild(hp)

	inputLbl := ui.NewLabel("input-label", "Text Input", font, 13)
	rangePage.AddChild(inputLbl)

	input := ui.NewTextInput("demo-input", font, 13)
	input.SetPlaceholder("Type here...")
	rangePage.AddChild(input)

	taLbl := ui.NewLabel("ta-label", "Text Area", font, 13)
	rangePage.AddChild(taLbl)

	ta := ui.NewTextArea("demo-textarea", font, 13)
	ta.SetSize(rangePage.Width-24, 60)
	ta.SetValue("Multi-line text goes here...")
	rangePage.AddChild(ta)

	rangePage.UpdateLayout()

	// Tab 3: Lists — list, toggle button bar.
	listsPage, _ := tabs.AddTabPage("Lists", ui.LayoutVBox, 10, tabPad)
	listsPage.Align = ui.AlignStart

	list := ui.NewList("demo-list", 26)
	list.SetSize(300, 150)
	list.SetItems([]ui.ListItem{
		{Data: "Apple"}, {Data: "Banana"}, {Data: "Cherry"},
		{Data: "Date"}, {Data: "Elderberry"}, {Data: "Fig"},
		{Data: "Grape"}, {Data: "Honeydew"}, {Data: "Kiwi"},
	})
	list.SetSelectable(true)
	list.SetRenderItem(func(idx int, data any) *ui.Component {
		lbl := ui.NewLabel("item", data.(string), font, 12)
		return &lbl.Component
	})
	listsPage.AddChild(list)

	tbbLbl := ui.NewLabel("tbb-label", "Toggle Button Bar", font, 13)
	listsPage.AddChild(tbbLbl)

	tbb := ui.NewToggleButtonBar("demo-tbb", font, 12)
	tbb.AddButton("Alpha")
	tbb.AddButton("Beta")
	tbb.AddButton("Gamma")
	tbb.SetSelected(0)
	tbb.SetSize(260, 30)
	listsPage.AddChild(tbb)

	listsPage.UpdateLayout()

	body.AddChild(tabs)

	// --- Bottom buttons ---
	bottomRow := ui.NewHBox("bottom-row")
	bottomRow.Spacing = 12
	bottomRow.Align = ui.AlignCenter
	bottomRow.Justify = ui.AlignCenter
	bottomRow.Width = 500
	bottomRow.Height = 36

	okBtn := ui.NewButton("ok-btn", "OK", font, 13)
	okBtn.SetSize(100, 32)
	bottomRow.AddChild(okBtn)

	cancelBtn := ui.NewButton("cancel-btn", "Cancel", font, 13)
	cancelBtn.SetSize(100, 32)
	cancelBtn.SetVariant(ui.Danger)
	bottomRow.AddChild(cancelBtn)

	body.AddChild(bottomRow)
	body.UpdateLayout()

	// --- Theme switching logic ---
	applyTheme := func(idx int) {
		entry := themes[idx]
		win.SetTheme(entry.theme)
		win.SetTitle(fmt.Sprintf("Theme Gallery: %s", entry.name))
		bgPanel.SetBackground(entry.clear)
	}
	selector.SetOnChange(func(idx int) { applyTheme(idx) })
	applyTheme(0)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI: Theme Gallery",
		Width:      screenW,
		Height:     screenH,
		ClearColor: themes[0].clear,
	})
}
