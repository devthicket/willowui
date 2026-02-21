// NavDrawer - slide-out navigation panel demo.
// Shows a left-anchored drawer that opens/closes with a hamburger button,
// and a right-anchored drawer toggled by a second button.
// The pin toggle sits in the drawer's title row.
package main

import (
	"image"
	"image/color"

	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 600
)

// makeHamburgerIcon creates a 14x12 white three-line hamburger icon.
func makeHamburgerIcon() *ebiten.Image {
	const w, h = 14, 12
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	for x := 0; x < w; x++ {
		// Three horizontal lines at rows 0-1, 5-6, 10-11.
		img.SetNRGBA(x, 0, white)
		img.SetNRGBA(x, 1, white)
		img.SetNRGBA(x, 5, white)
		img.SetNRGBA(x, 6, white)
		img.SetNRGBA(x, 10, white)
		img.SetNRGBA(x, 11, white)
	}
	return ebiten.NewImageFromImage(img)
}

func main() {
	font := ui.MustLoadDefaultFont()

	screen := ui.NewScreen()

	// --- Toolbar ---

	hamburgerIcon := makeHamburgerIcon()
	menuBtn := ui.NewIconButton("menu-btn")
	menuBtn.SetIconImage(hamburgerIcon)
	menuBtn.SetIconSize(14, 12)
	menuBtn.SetSize(40, 32)
	menuBtn.SetPosition(16, 16)
	screen.Add(menuBtn)

	detailBtn := ui.NewButton("detail-btn", "Details", font, 14)
	detailBtn.SetPosition(screenW-96, 16)
	detailBtn.SetSize(80, 32)
	screen.Add(detailBtn)

	statusLabel := ui.NewLabel("status", "Click Menu or Details to open drawers", font, 14)
	statusLabel.SetColor(willow.RGBA(0.6, 0.6, 0.6, 1))
	statusLabel.SetPosition(16, screenH-30)
	screen.Add(statusLabel)

	// --- Left NavDrawer ---

	leftDrawer := ui.NewNavDrawer("left-drawer")
	leftDrawer.SetSize(screenW, screenH)
	leftDrawer.SetWidth(220)
	leftDrawer.Node().SetZIndex(100)

	// Build nav content — add children first, then set size to trigger layout.
	navPanel := ui.NewPanel("nav-content")
	navPanel.SetLayout(ui.LayoutVBox)
	navPanel.SetSpacing(6)

	// Title row: "Navigation" on the left, pin toggle on the right.
	titleRow := ui.NewPanel("title-row")
	titleRow.SetLayout(ui.LayoutHBox)
	titleRow.SetJustify(ui.AlignSpaceBetween)
	titleRow.SetAlignment(ui.AlignCenter)

	navTitle := ui.NewLabel("nav-title", "Navigation", font, 18)
	navTitle.SetColor(willow.RGBA(1, 1, 1, 1))
	titleRow.AddChild(navTitle)

	pinToggle := ui.NewToggle("pin-toggle")
	titleRow.AddChild(pinToggle)

	titleRow.SetSize(188, 28)
	navPanel.AddChild(titleRow)

	for _, item := range []string{"Home", "Profile", "Settings", "About"} {
		btn := ui.NewButton("nav-"+item, item, font, 14)
		btn.SetSize(188, 32)
		navPanel.AddChild(btn)
	}

	navPanel.SetSize(220, screenH)
	leftDrawer.SetContent(navPanel)
	screen.Add(leftDrawer)

	// --- Right NavDrawer ---

	rightDrawer := ui.NewNavDrawer("right-drawer")
	rightDrawer.SetSize(screenW, screenH)
	rightDrawer.SetWidth(200)
	rightDrawer.SetAnchor(ui.NavDrawerRight)
	rightDrawer.Node().SetZIndex(100)

	detailPanel := ui.NewPanel("detail-content")
	detailPanel.SetLayout(ui.LayoutVBox)
	detailPanel.SetSpacing(6)

	detailTitle := ui.NewLabel("detail-title", "Details", font, 18)
	detailTitle.SetColor(willow.RGBA(1, 1, 1, 1))
	detailPanel.AddChild(detailTitle)

	detailInfo := ui.NewLabel("detail-info", "Right-side panel", font, 14)
	detailInfo.SetColor(willow.RGBA(0.8, 0.8, 0.8, 1))
	detailPanel.AddChild(detailInfo)

	detailPanel.SetSize(200, screenH)
	rightDrawer.SetContent(detailPanel)
	screen.Add(rightDrawer)

	// --- Wire callbacks ---

	menuBtn.SetOnClick(func() { leftDrawer.Toggle() })
	detailBtn.SetOnClick(func() { rightDrawer.Toggle() })
	pinToggle.SetOnChange(func(v bool) { leftDrawer.SetPinned(v) })

	leftDrawer.SetOnOpen(func() { statusLabel.SetText("Left drawer opened") })
	leftDrawer.SetOnClose(func() { statusLabel.SetText("Left drawer closed") })
	rightDrawer.SetOnOpen(func() { statusLabel.SetText("Right drawer opened") })
	rightDrawer.SetOnClose(func() { statusLabel.SetText("Right drawer closed") })

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "NavDrawer Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.12, 0.12, 0.14, 1),
	})
}
