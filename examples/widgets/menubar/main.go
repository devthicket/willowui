// menubar demonstrates a desktop-style menu bar with File, Edit, View, and
// Help menus. Each menu entry opens a dropdown with keyboard shortcut hints.
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

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 14.0
		sizeSmall  = 15.0
	)

	screen := ui.NewScreen()

	// Status label shows which action was last selected.
	statusRef := ui.NewRef("-- no action yet --")

	statusBg := ui.NewPanel("status-bg")
	statusBg.SetSize(screenW-48, screenH-80)
	statusBg.SetPosition(24, 56)
	screen.Add(statusBg)

	statusLabel := ui.NewLabel("status", "", font, sizeMedium)
	statusLabel.BindText(statusRef)
	statusLabel.SetColor(willow.RGBA(0.7, 0.8, 0.9, 1))
	statusLabel.SetPosition(16, 16)
	statusBg.AddChild(statusLabel)

	record := func(msg string) { statusRef.Set(msg) }

	// -----------------------------------------------------------------------
	// MenuBar
	// -----------------------------------------------------------------------
	menuBar := ui.NewMenuBar("menu-bar", font, sizeSmall)
	menuBar.SetSize(screenW, 32)
	menuBar.SetPosition(0, 0)
	menuBar.SetEntries([]ui.MenuBarEntry{
		{
			Label: "File",
			Items: []ui.MenuItem{
				{Label: "New", Shortcut: "Ctrl+N", OnSelect: func() { record("File > New") }},
				{Label: "Open", Shortcut: "Ctrl+O", OnSelect: func() { record("File > Open") }},
				{Label: "Save", Shortcut: "Ctrl+S", OnSelect: func() { record("File > Save") }},
				{Label: "Save As", Shortcut: "Ctrl+Shift+S", OnSelect: func() { record("File > Save As") }},
				{Separator: true},
				{Label: "Export PNG", OnSelect: func() { record("File > Export PNG") }},
				{Label: "Export SVG", Disabled: true},
				{Separator: true},
				{Label: "Exit", OnSelect: func() { record("File > Exit") }},
			},
		},
		{
			Label: "Edit",
			Items: []ui.MenuItem{
				{Label: "Undo", Shortcut: "Ctrl+Z", OnSelect: func() { record("Edit > Undo") }},
				{Label: "Redo", Shortcut: "Ctrl+Y", OnSelect: func() { record("Edit > Redo") }},
				{Separator: true},
				{Label: "Cut", Shortcut: "Ctrl+X", OnSelect: func() { record("Edit > Cut") }},
				{Label: "Copy", Shortcut: "Ctrl+C", OnSelect: func() { record("Edit > Copy") }},
				{Label: "Paste", Shortcut: "Ctrl+V", OnSelect: func() { record("Edit > Paste") }},
				{Separator: true},
				{Label: "Select All", Shortcut: "Ctrl+A", OnSelect: func() { record("Edit > Select All") }},
			},
		},
		{
			Label: "View",
			Items: []ui.MenuItem{
				{Label: "Zoom In", Shortcut: "Ctrl++", OnSelect: func() { record("View > Zoom In") }},
				{Label: "Zoom Out", Shortcut: "Ctrl+-", OnSelect: func() { record("View > Zoom Out") }},
				{Label: "Reset Zoom", Shortcut: "Ctrl+0", OnSelect: func() { record("View > Reset Zoom") }},
				{Separator: true},
				{Label: "Show Grid", OnSelect: func() { record("View > Show Grid") }},
				{Label: "Show Rulers", OnSelect: func() { record("View > Show Rulers") }},
				{Separator: true},
				{Label: "Fullscreen", Shortcut: "F11", OnSelect: func() { record("View > Fullscreen") }},
			},
		},
		{
			Label: "Help",
			Items: []ui.MenuItem{
				{Label: "Documentation", OnSelect: func() { record("Help > Documentation") }},
				{Label: "Report Issue", OnSelect: func() { record("Help > Report Issue") }},
				{Separator: true},
				{Label: fmt.Sprintf("About WillowUI v%s", "0.1"), OnSelect: func() { record("Help > About") }},
			},
		},
	})

	menuBar.SetOnMenuOpen(func(idx int) {
		fmt.Printf("menu opened: index %d\n", idx)
	})
	menuBar.SetOnMenuClose(func() {
		fmt.Println("menu closed")
	})

	screen.Add(menuBar)

	// Footer hint.
	footer := willow.NewText("footer", "Click a menu entry to open  |  Hover to switch  |  Arrow keys to navigate  |  Esc to close", font)
	footer.TextBlock.FontSize = sizeSmall
	footer.TextBlock.Color = willow.RGBA(0.30, 0.40, 0.50, 1)
	footer.SetPosition(24, screenH-24)
	screen.AddNode(footer)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI -- MenuBar Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.09, 0.11, 1),
	})
}
