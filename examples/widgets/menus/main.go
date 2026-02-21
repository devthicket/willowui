// menus showcases WillowUI's menu system: Select dropdowns, right-click
// ContextMenus, and programmatically triggered menus.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 900
	screenH = 640
	colL    = 24.0  // left column x
	colR    = 500.0 // right column x (status log)
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 15.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	// -----------------------------------------------------------------------
	// Header
	// -----------------------------------------------------------------------
	title := willow.NewText("title", "WillowUI: Menus & Dropdowns", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(0.93, 0.93, 0.93, 1)
	title.SetPosition(colL, 16)
	screen.AddNode(title)

	screen.AddNode(hline("div-top", colL, 50, screenW-48))

	// -----------------------------------------------------------------------
	// Status log (right column) — shared by all sections
	// -----------------------------------------------------------------------
	logHeader := sectionLabel("log-header", "Last action", font, sizeSmall)
	logHeader.SetPosition(colR, 62)
	screen.Add(logHeader)

	logRef := ui.NewRef("-- nothing selected yet --")

	logBg := ui.NewPanel("log-bg")
	logBg.SetSize(screenW-colR-24, screenH-120)
	logBg.SetPosition(colR, 80)
	screen.Add(logBg)

	logText := ui.NewLabel("log-text", "", font, sizeMedium)
	logText.BindText(logRef)
	logText.SetColor(willow.RGBA(0.75, 0.85, 0.95, 1))
	logText.SetPosition(12, 12)
	logBg.AddChild(logText)

	record := func(msg string) { logRef.Set(msg) }

	// -----------------------------------------------------------------------
	// Section 1 — Select dropdowns
	// -----------------------------------------------------------------------
	screen.AddNode(sectionNode("s1-hdr", "Select  /  Dropdown", font, sizeSmall, colL, 62))

	// Resolution
	screen.AddNode(inlineLabel("res-lbl", "Resolution", font, sizeMedium, colL, 84))
	resSel := ui.NewSelect("resolution",
		[]ui.SelectOption{
			{Label: "1280 × 720  (HD)", Value: "720p"},
			{Label: "1920 × 1080  (Full HD)", Value: "1080p"},
			{Label: "2560 × 1440  (QHD)", Value: "1440p"},
			{Label: "3840 × 2160  (4K)", Value: "4k"},
		},
		font, sizeMedium,
	)
	resSel.SetSize(240, 32)
	resSel.SetPosition(colL+120, 80)
	resSel.SetSelected(1) // start on 1080p
	resSel.SetOnChange(func(_ int, opt ui.SelectOption) {
		record(fmt.Sprintf("Resolution → %s  (value=%q)", opt.Label, opt.Value))
	})
	screen.Add(resSel)

	// Quality
	screen.AddNode(inlineLabel("q-lbl", "Quality", font, sizeMedium, colL, 126))
	qualSel := ui.NewSelect("quality",
		[]ui.SelectOption{
			{Label: "Low", Value: 0},
			{Label: "Medium", Value: 1},
			{Label: "High", Value: 2},
			{Label: "Ultra (may reduce framerate)", Value: 3},
		},
		font, sizeMedium,
	)
	qualSel.SetSize(240, 32)
	qualSel.SetPosition(colL+120, 122)
	qualSel.SetSelected(2) // start on High
	qualSel.SetOnChange(func(idx int, opt ui.SelectOption) {
		record(fmt.Sprintf("Quality → %s  (index %d)", opt.Label, idx))
	})
	screen.Add(qualSel)

	// Language (disabled options example)
	screen.AddNode(inlineLabel("lang-lbl", "Language", font, sizeMedium, colL, 168))
	langSel := ui.NewSelect("language",
		[]ui.SelectOption{
			{Label: "English", Value: "en"},
			{Label: "Español", Value: "es"},
			{Label: "Français", Value: "fr"},
			{Label: "日本語", Value: "ja"},
		},
		font, sizeMedium,
	)
	langSel.SetSize(240, 32)
	langSel.SetPosition(colL+120, 164)
	langSel.SetOnChange(func(_ int, opt ui.SelectOption) {
		record(fmt.Sprintf("Language → %s", opt.Label))
	})
	screen.Add(langSel)

	// Timezone (long list — triggers scrollbar)
	screen.AddNode(inlineLabel("tz-lbl", "Timezone", font, sizeMedium, colL, 210))
	tzSel := ui.NewSelect("timezone",
		[]ui.SelectOption{
			{Label: "UTC−12  Baker Island", Value: "Etc/GMT+12"},
			{Label: "UTC−11  American Samoa", Value: "Pacific/Pago_Pago"},
			{Label: "UTC−10  Hawaii", Value: "Pacific/Honolulu"},
			{Label: "UTC−9   Alaska", Value: "America/Anchorage"},
			{Label: "UTC−8   Pacific Time", Value: "America/Los_Angeles"},
			{Label: "UTC−7   Mountain Time", Value: "America/Denver"},
			{Label: "UTC−6   Central Time", Value: "America/Chicago"},
			{Label: "UTC−5   Eastern Time", Value: "America/New_York"},
			{Label: "UTC−4   Atlantic Time", Value: "America/Halifax"},
			{Label: "UTC−3   Buenos Aires", Value: "America/Argentina/Buenos_Aires"},
			{Label: "UTC−2   South Georgia", Value: "Atlantic/South_Georgia"},
			{Label: "UTC−1   Azores", Value: "Atlantic/Azores"},
			{Label: "UTC+0   London", Value: "Europe/London"},
			{Label: "UTC+1   Paris / Berlin", Value: "Europe/Paris"},
			{Label: "UTC+2   Cairo / Helsinki", Value: "Africa/Cairo"},
			{Label: "UTC+3   Moscow / Nairobi", Value: "Europe/Moscow"},
			{Label: "UTC+4   Dubai", Value: "Asia/Dubai"},
			{Label: "UTC+5   Karachi", Value: "Asia/Karachi"},
			{Label: "UTC+5:30  Mumbai", Value: "Asia/Kolkata"},
			{Label: "UTC+6   Dhaka", Value: "Asia/Dhaka"},
			{Label: "UTC+7   Bangkok / Jakarta", Value: "Asia/Bangkok"},
			{Label: "UTC+8   Beijing / Singapore", Value: "Asia/Singapore"},
			{Label: "UTC+9   Tokyo / Seoul", Value: "Asia/Tokyo"},
			{Label: "UTC+10  Sydney", Value: "Australia/Sydney"},
			{Label: "UTC+11  Solomon Islands", Value: "Pacific/Guadalcanal"},
			{Label: "UTC+12  Auckland", Value: "Pacific/Auckland"},
		},
		font, sizeMedium,
	)
	tzSel.SetSize(260, 32)
	tzSel.SetPosition(colL+120, 206)
	tzSel.SetSelected(12) // UTC+0
	tzSel.SetOnChange(func(_ int, opt ui.SelectOption) {
		record(fmt.Sprintf("Timezone → %s", opt.Label))
	})
	screen.Add(tzSel)

	screen.AddNode(hline("div1", colL, 252, colR-colL-32))

	// -----------------------------------------------------------------------
	// Section 2 — Right-click ContextMenu
	// -----------------------------------------------------------------------
	screen.AddNode(sectionNode("s2-hdr", "Context Menu  /  Right-click", font, sizeSmall, colL, 264))

	// Three "file" cards that each have their own context menu.
	files := []string{"hero_sprite.png", "tilemap.json", "theme.json"}
	for i, name := range files {
		fi := i
		fn := name

		card := ui.NewPanel(fmt.Sprintf("card-%d", i))
		card.SetSize(310, 36)
		card.SetPosition(colL, 284+float64(i)*44)
		screen.Add(card)

		lbl := ui.NewLabel(fmt.Sprintf("card-lbl-%d", i), "▸  "+fn, font, sizeMedium)
		lbl.SetColor(willow.RGBA(0.75, 0.82, 0.95, 1))
		lbl.SetPosition(12, 10)
		card.AddChild(lbl)

		cm := ui.NewContextMenu(font, sizeMedium)
		cm.SetItems([]ui.MenuItem{
			{Label: "Open", OnSelect: func() { record(fmt.Sprintf("Open  →  %s", fn)) }},
			{Label: "Rename...", OnSelect: func() { record(fmt.Sprintf("Rename  →  %s", fn)) }},
			{Label: "Duplicate", OnSelect: func() { record(fmt.Sprintf("Duplicate  →  %s", fn)) }},
			{Separator: true},
			{Label: "Copy path", OnSelect: func() { record(fmt.Sprintf("Copied path of  %s", fn)) }},
			{Separator: true},
			{Label: fmt.Sprintf("Delete  file %d", fi+1), OnSelect: func() { record(fmt.Sprintf("Delete  →  %s", fn)) }},
		})
		card.SetContextMenu(cm)
	}

	hint2 := willow.NewText("hint2", "← right-click any row", font)
	hint2.TextBlock.FontSize = sizeSmall
	hint2.TextBlock.Color = willow.RGBA(0.35, 0.45, 0.55, 1)
	hint2.SetPosition(colL+320, 300)
	screen.AddNode(hint2)

	screen.AddNode(hline("div2", colL, 422, colR-colL-32))

	// -----------------------------------------------------------------------
	// Section 3 — Programmatic menu from a button
	// -----------------------------------------------------------------------
	screen.AddNode(sectionNode("s3-hdr", "Programmatic  /  ShowAt", font, sizeSmall, colL, 434))

	// "Edit" menu button — opens below the button.
	editBtn := ui.NewButton("edit-btn", "Edit", font, sizeMedium)
	editBtn.SetSize(80, 32)
	editBtn.SetPosition(colL, 454)
	screen.Add(editBtn)

	editMenu := ui.NewContextMenu(font, sizeMedium)
	editMenu.SetItems([]ui.MenuItem{
		{Label: "Undo", OnSelect: func() { record("Undo") }},
		{Label: "Redo", OnSelect: func() { record("Redo") }},
		{Separator: true},
		{Label: "Cut", OnSelect: func() { record("Cut") }},
		{Label: "Copy", OnSelect: func() { record("Copy") }},
		{Label: "Paste", OnSelect: func() { record("Paste") }},
		{Separator: true},
		{Label: "Select All", OnSelect: func() { record("Select All") }},
	})
	editBtn.SetOnClick(func() { editMenu.ShowAt(colL, 490) })

	// "View" menu button.
	viewBtn := ui.NewButton("view-btn", "View", font, sizeMedium)
	viewBtn.SetSize(80, 32)
	viewBtn.SetPosition(colL+88, 454)
	screen.Add(viewBtn)

	viewMenu := ui.NewContextMenu(font, sizeMedium)
	viewMenu.SetItems([]ui.MenuItem{
		{Label: "Zoom In", OnSelect: func() { record("Zoom In") }},
		{Label: "Zoom Out", OnSelect: func() { record("Zoom Out") }},
		{Label: "Zoom to Fit", OnSelect: func() { record("Zoom to Fit") }},
		{Separator: true},
		{Label: "Show Grid", OnSelect: func() { record("Show Grid") }},
		{Label: "Show Rulers", OnSelect: func() { record("Show Rulers") }},
		{Separator: true},
		{Label: "Fullscreen", OnSelect: func() { record("Fullscreen") }},
	})
	viewBtn.SetOnClick(func() { viewMenu.ShowAt(colL+88, 490) })

	// "Insert" menu button — with a disabled item.
	insertBtn := ui.NewButton("insert-btn", "Insert", font, sizeMedium)
	insertBtn.SetSize(88, 32)
	insertBtn.SetPosition(colL+176, 454)
	screen.Add(insertBtn)

	insertMenu := ui.NewContextMenu(font, sizeMedium)
	insertMenu.SetItems([]ui.MenuItem{
		{Label: "Text Block", OnSelect: func() { record("Insert → Text Block") }},
		{Label: "Image...", OnSelect: func() { record("Insert → Image") }},
		{Label: "Table...", OnSelect: func() { record("Insert → Table") }},
		{Separator: true},
		{Label: "Symbol...", OnSelect: func() { record("Insert → Symbol") }},
		{Label: "Emoji (unavailable)", Disabled: true},
	})
	insertBtn.SetOnClick(func() { insertMenu.ShowAt(colL+176, 490) })

	// -----------------------------------------------------------------------
	// Footer
	// -----------------------------------------------------------------------
	screen.AddNode(hline("div-bot", colL, screenH-44, float64(screenW-48)))

	footer := willow.NewText("footer", "Tab · Shift+Tab to move focus    Space · Enter to open dropdown    Esc to dismiss", font)
	footer.TextBlock.FontSize = sizeSmall
	footer.TextBlock.Color = willow.RGBA(0.30, 0.40, 0.50, 1)
	footer.SetPosition(colL, screenH-28)
	screen.AddNode(footer)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Menus Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.09, 0.11, 1),
	})
}

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

func hline(name string, x, y, w float64) *willow.Node {
	n := willow.NewSprite(name, willow.TextureRegion{})
	n.SetScale(w, 1)
	n.SetPosition(x, y)
	n.SetColor(willow.RGBA(0.22, 0.26, 0.30, 1))
	return n
}

func sectionNode(name, text string, font *willow.FontFamily, size, x, y float64) *willow.Node {
	n := willow.NewText(name, text, font)
	n.TextBlock.FontSize = size
	n.TextBlock.Color = willow.RGBA(0.40, 0.50, 0.62, 1)
	n.SetPosition(x, y)
	return n
}

func sectionLabel(name, text string, font *willow.FontFamily, size float64) *ui.Label {
	l := ui.NewLabel(name, text, font, size)
	l.SetColor(willow.RGBA(0.40, 0.50, 0.62, 1))
	return l
}

func inlineLabel(name, text string, font *willow.FontFamily, size, x, y float64) *willow.Node {
	n := willow.NewText(name, text, font)
	n.TextBlock.FontSize = size
	n.TextBlock.Color = willow.RGBA(0.60, 0.65, 0.72, 1)
	n.SetPosition(x, y+8)
	return n
}
