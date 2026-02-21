// Windows demonstrates WillowUI's container widgets: Panel, ScrollPanel, and
// Window with dragging, closing, resizing, and bring-to-front behaviour.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 900
	screenH = 650
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 24.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI \u2014 Containers Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	// ── Window 1: Form ──────────────────────────────────────────────────────
	win1 := ui.NewWindow("win-form", "Form Window", font, sizeMedium)
	win1.SetSize(320, 240)
	win1.SetResizable(true)
	win1.SetMinSize(320, 240)
	win1.SetPosition(40, 60)
	screen.Add(win1)
	ui.DefaultWindowManager.Add(win1)

	body1 := win1.Body()
	body1.SetLayout(ui.LayoutVBox)
	body1.SetSpacing(8)
	body1.Padding = ui.Insets{Top: 8, Left: 8, Right: 8, Bottom: 8}

	nameLabel := ui.NewLabel("name-lbl", "Name:", font, sizeSmall)
	body1.AddChild(nameLabel)

	nameInput := ui.NewTextInput("name-input", font, sizeSmall)
	nameInput.SetPlaceholder("Enter your name")
	nameInput.SetWidth(280)
	body1.AddChild(nameInput)

	emailLabel := ui.NewLabel("email-lbl", "Email:", font, sizeSmall)
	body1.AddChild(emailLabel)

	emailInput := ui.NewTextInput("email-input", font, sizeSmall)
	emailInput.SetPlaceholder("user@example.com")
	emailInput.SetWidth(280)
	body1.AddChild(emailInput)

	submitBtn := ui.NewButton("submit", "Submit", font, sizeSmall)
	submitBtn.SetSize(100, 28)
	body1.AddChild(submitBtn)

	body1.UpdateLayout()

	// ── Window 2: Scrollable List ───────────────────────────────────────────
	win2 := ui.NewWindow("win-list", "Scrollable List", font, sizeMedium)
	win2.SetSize(280, 250)
	win2.SetPosition(400, 80)
	screen.Add(win2)
	ui.DefaultWindowManager.Add(win2)

	// Use a scroll panel inside the body.
	scrollPanel := ui.NewScrollPanel("list-scroll")
	scrollPanel.SetSize(280, 218)
	scrollPanel.SetBackground(willow.RGBA(0.12, 0.12, 0.14, 1))
	scrollPanel.SetContentSize(260, 600)
	win2.Body().AddChild(scrollPanel)

	// Add items to the scroll panel's content node.
	for i := 0; i < 20; i++ {
		lbl := ui.NewLabel(
			fmt.Sprintf("item-%d", i),
			fmt.Sprintf("  Item %d -- click to select", i+1),
			font, sizeSmall,
		)
		lbl.SetPosition(4, float64(i)*28+4)
		scrollPanel.AddContent(lbl)
	}

	// ── Window 3: Text Content ──────────────────────────────────────────────
	win3 := ui.NewWindow("win-text", "About WillowUI", font, sizeMedium)
	win3.SetSize(340, 200)
	win3.SetResizable(true)
	win3.SetMinSize(340, 200)
	win3.SetPosition(200, 360)
	screen.Add(win3)
	ui.DefaultWindowManager.Add(win3)

	win3.Body().SetBackground(willow.RGBA(0.12, 0.12, 0.14, 1))

	aboutText := ui.NewLabel("about-text",
		"WillowUI is a UI library built on top\n"+
			"of the Willow game engine. It provides\n"+
			"reactive data binding, layout management,\n"+
			"form controls, and container widgets.\n"+
			"\n"+
			"Drag windows by their title bar.\n"+
			"Close them with the X button.\n"+
			"Resize with the bottom-right handle.",
		font, sizeSmall,
	)
	aboutText.SetColor(willow.RGBA(1, 1, 1, 1))
	aboutText.SetPosition(8, 8)
	win3.Body().AddChild(aboutText)

	// ── Status label ────────────────────────────────────────────────────────
	statusRef := ui.NewRef("Status: Ready")
	statusLabel := ui.NewLabel("status", "", font, sizeSmall)
	statusLabel.BindText(statusRef)
	statusLabel.SetPosition(24, screenH-24)
	screen.Add(statusLabel)

	win1.SetOnClose(func() {
		statusRef.Set("Status: Form window closed")
	})
	win2.SetOnClose(func() {
		statusRef.Set("Status: List window closed")
	})
	win3.SetOnClose(func() {
		statusRef.Set("Status: About window closed")
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI \u2014 Containers Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
