// Component demonstrates WillowUI's base Component system: layout modes
// (VBox, HBox, Grid), theming, focus management, dirty flags, and the
// component lifecycle. Each section is labeled and color-coded.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 600
)

// Demo color palette — mirrors the default theme values for display purposes.
var (
	colorPrimary    = willow.RGBA(0.26, 0.52, 0.96, 1)
	colorSecondary  = willow.RGBA(0.40, 0.40, 0.45, 1)
	colorBackground = willow.RGBA(0.15, 0.15, 0.17, 1)
	colorText       = willow.RGBA(0.93, 0.93, 0.93, 1)
	colorBorder     = willow.RGBA(0.30, 0.30, 0.33, 1)
	colorDisabled   = willow.RGBA(0.45, 0.45, 0.48, 0.6)
	colorHover      = willow.RGBA(0.30, 0.58, 1.00, 1)
	colorPressed    = willow.RGBA(0.20, 0.42, 0.80, 1)
	colorFocused    = willow.RGBA(0.35, 0.65, 1.00, 1)
)

type layoutController struct {
	font        *willow.FontFamily
	fm          *ui.FocusManager
	focusBoxes  []*ui.Component
	focusLabels []*willow.Node
	focusStatus *willow.Node
	frame       int
	tabFrame    int
}

func (c *layoutController) OnCreate(s *ui.Screen) {
	font := c.font

	const (
		sizeLarge  = 24.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	// ── Title ────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI: Component Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	s.AddNode(title)

	div := ui.NewDivider("divider", screenW-48)
	div.SetPosition(24, 48)
	s.AddNode(div)

	// ── Section 1: VBox Layout ───────────────────────────────────────────────
	addSectionLabel(s, font, sizeSmall, "LayoutVBox: vertical stack with spacing", 24, 58)

	vbox := ui.NewComponent("vbox-container")
	vbox.X = 24
	vbox.Y = 78
	vbox.Width = 220
	vbox.Height = 230
	vbox.Layout = ui.LayoutVBox
	vbox.Spacing = 8
	vbox.Padding = ui.Insets{Top: 10, Right: 10, Bottom: 10, Left: 10}
	s.Add(vbox)
	addComponentBG(vbox, colorBackground)

	vboxColors := []struct {
		label string
		color willow.Color
	}{
		{"Primary", colorPrimary},
		{"Secondary", colorSecondary},
		{"Hover", colorHover},
		{"Pressed", colorPressed},
	}
	for _, col := range vboxColors {
		child := makeColorBox(col.label, 200, 40, col.color, font, sizeMedium)
		vbox.AddChild(child)
	}
	vbox.UpdateLayout()

	// ── Section 2: HBox Layout ───────────────────────────────────────────────
	addSectionLabel(s, font, sizeSmall, "LayoutHBox: horizontal stack", 270, 58)

	hbox := ui.NewComponent("hbox-container")
	hbox.X = 270
	hbox.Y = 78
	hbox.Width = 500
	hbox.Height = 80
	hbox.Layout = ui.LayoutHBox
	hbox.Spacing = 10
	hbox.Padding = ui.Insets{Top: 10, Right: 10, Bottom: 10, Left: 10}
	s.Add(hbox)
	addComponentBG(hbox, colorBackground)

	for i := range 5 {
		v := 0.3 + float64(i)*0.15
		box := makeColorBox(
			fmt.Sprintf("%d", i+1),
			60, 60,
			willow.RGBA(v*0.6, v*0.8, v, 1),
			font, sizeMedium,
		)
		hbox.AddChild(box)
		if i == 1 {
			hbox.AddChild(ui.NewSpacer("hbox-gap", 30, 0)) // 30px gap after item 2
		}
	}
	hbox.UpdateLayout()

	// ── Section 3: Grid Layout ───────────────────────────────────────────────
	addSectionLabel(s, font, sizeSmall, "LayoutGrid: 3 columns", 270, 170)

	grid := ui.NewComponent("grid-container")
	grid.X = 270
	grid.Y = 190
	grid.Width = 300
	grid.Height = 120
	grid.Layout = ui.LayoutGrid
	grid.GridColumns = 3
	grid.Spacing = 6
	grid.Padding = ui.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8}
	s.Add(grid)
	addComponentBG(grid, colorBackground)

	gridColors := []willow.Color{
		willow.RGBA(0.9, 0.3, 0.3, 1),
		willow.RGBA(0.3, 0.9, 0.3, 1),
		willow.RGBA(0.3, 0.3, 0.9, 1),
		willow.RGBA(0.9, 0.9, 0.3, 1),
		willow.RGBA(0.9, 0.3, 0.9, 1),
		willow.RGBA(0.3, 0.9, 0.9, 1),
	}
	for i, col := range gridColors {
		box := makeColorBox(fmt.Sprintf("G%d", i+1), 50, 40, col, font, sizeSmall)
		grid.AddChild(box)
	}
	grid.UpdateLayout()

	// ── Section 3b: SpaceBetween ────────────────────────────────────────────
	addSectionLabel(s, font, sizeSmall, "Justify: SpaceBetween", 590, 170)

	sbHBox := ui.NewComponent("sb-hbox")
	sbHBox.X = 590
	sbHBox.Y = 190
	sbHBox.Width = 180
	sbHBox.Height = 30
	sbHBox.Layout = ui.LayoutHBox
	sbHBox.Justify = ui.AlignSpaceBetween
	sbHBox.Padding = ui.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4}
	s.Add(sbHBox)
	addComponentBG(sbHBox, colorBackground)
	for i := range 3 {
		v := 0.4 + float64(i)*0.2
		box := makeColorBox("", 30, 22, willow.RGBA(v*0.6, v*0.8, v, 1), font, sizeSmall)
		sbHBox.AddChild(box)
	}
	sbHBox.UpdateLayout()

	sbVBox := ui.NewComponent("sb-vbox")
	sbVBox.X = 590
	sbVBox.Y = 230
	sbVBox.Width = 60
	sbVBox.Height = 80
	sbVBox.Layout = ui.LayoutVBox
	sbVBox.Justify = ui.AlignSpaceBetween
	sbVBox.Padding = ui.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4}
	s.Add(sbVBox)
	addComponentBG(sbVBox, colorBackground)
	for i := range 3 {
		v := 0.4 + float64(i)*0.2
		box := makeColorBox("", 52, 14, willow.RGBA(v, v*0.5, v*0.3, 1), font, sizeSmall)
		sbVBox.AddChild(box)
	}
	sbVBox.UpdateLayout()

	// ── Section 4: Theme Colors ──────────────────────────────────────────────
	addSectionLabel(s, font, sizeSmall, "DefaultTheme color palette", 24, 320)

	themeColors := []struct {
		label string
		color willow.Color
	}{
		{"Primary", colorPrimary},
		{"Secondary", colorSecondary},
		{"Background", colorBackground},
		{"Text", colorText},
		{"Border", colorBorder},
		{"Disabled", colorDisabled},
		{"Hover", colorHover},
		{"Pressed", colorPressed},
		{"Focused", colorFocused},
	}

	swatchX := 40.0
	for _, tc := range themeColors {
		swatch := willow.NewSprite("swatch", willow.TextureRegion{})
		swatch.SetPosition(swatchX, 342)
		swatch.SetScale(60, 30)
		swatch.SetColor(tc.color)
		s.AddNode(swatch)

		lbl := willow.NewText("swatch-lbl", tc.label, font)
		lbl.TextBlock.FontSize = sizeSmall
		lbl.TextBlock.Color = willow.RGBA(0.7, 0.7, 0.7, 1)
		lbl.SetPosition(swatchX, 376)
		s.AddNode(lbl)

		swatchX += 80
	}

	// ── Section 5: Focus Manager ─────────────────────────────────────────────
	addSectionLabel(s, font, sizeSmall, "FocusManager: click boxes or watch tab cycling", 24, 410)

	fm := ui.NewFocusManager()
	c.fm = fm
	c.focusBoxes = make([]*ui.Component, 4)
	c.focusLabels = make([]*willow.Node, 4)
	boxNames := []string{"A", "B", "C", "D"}

	for i := range 4 {
		box := ui.NewComponent(fmt.Sprintf("focus-%s", boxNames[i]))
		box.Width = 100
		box.Height = 50

		bg := willow.NewSprite("focus-bg", willow.TextureRegion{})
		bg.SetScale(box.Width, box.Height)
		bg.SetColor(colorSecondary)
		box.AddRawChild(bg)

		lbl := willow.NewText("focus-lbl", boxNames[i], font)
		lbl.TextBlock.FontSize = sizeMedium
		lbl.TextBlock.Color = colorText
		lbl.SetPosition(40, 16)
		box.AddRawChild(lbl)

		box.SetPosition(40+float64(i)*120, 432)

		box.OnClick(func(ctx willow.ClickContext) {
			comp := ctx.Node.UserData.(*ui.Component)
			fm.SetFocus(comp)
		})

		s.Add(box)
		fm.Register(box)
		c.focusBoxes[i] = box
		c.focusLabels[i] = lbl
	}

	focusStatus := willow.NewText("focus-status", "focused: none", font)
	focusStatus.TextBlock.FontSize = sizeMedium
	focusStatus.TextBlock.Color = willow.RGBA(0.5, 1, 0.5, 1)
	focusStatus.SetPosition(40, 496)
	s.AddNode(focusStatus)
	c.focusStatus = focusStatus

	// ── Section 6: Enable/Disable ────────────────────────────────────────────
	addSectionLabel(s, font, sizeSmall, "Enabled vs Disabled component", 24, 530)

	enabledBox := makeColorBox("Enabled", 120, 40, colorPrimary, font, sizeMedium)
	enabledBox.SetPosition(40, 550)
	s.Add(enabledBox)

	disabledBox := makeColorBox("Disabled", 120, 40, colorDisabled, font, sizeMedium)
	disabledBox.SetEnabled(false)
	disabledBox.SetPosition(180, 550)
	s.Add(disabledBox)
}

func (c *layoutController) OnUpdate(dt float64) {
	c.frame++

	// Auto-cycle focus every 60 frames to show the focus manager in action.
	if c.frame%60 == 0 && c.tabFrame < 8 {
		c.fm.TabNext()
		c.tabFrame++
	}

	// Update focus visuals.
	for i, box := range c.focusBoxes {
		bg := box.Node().Children()[0]
		if box.IsFocused() {
			bg.SetColor(colorFocused)
			c.focusLabels[i].TextBlock.Color = willow.RGBA(0, 0, 0, 1)
		} else {
			bg.SetColor(colorSecondary)
			c.focusLabels[i].TextBlock.Color = colorText
		}
	}

	name := "none"
	if f := c.fm.Focused(); f != nil {
		name = f.Name()
	}
	c.focusStatus.SetContent(fmt.Sprintf("focused: %s", name))
}

func (c *layoutController) OnDestroy() {}

func main() {
	font := ui.MustLoadDefaultFont()

	ui.Stage.Add(ui.NewScreen(ui.WithController(&layoutController{font: font})))
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Component Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

// makeColorBox creates a Component with a colored background and centered text label.
func makeColorBox(label string, w, h float64, color willow.Color, font *willow.FontFamily, fontSize float64) *ui.Component {
	c := ui.NewComponent(label)
	c.Width = w
	c.Height = h

	bg := willow.NewSprite(label+"-bg", willow.TextureRegion{})
	bg.SetScale(w, h)
	bg.SetColor(color)
	c.AddRawChild(bg)

	textW, _ := font.MeasureString(label, 0, false, false)
	scale := fontSize / font.LineHeight(0, false, false)
	lbl := willow.NewText(label+"-lbl", label, font)
	lbl.TextBlock.FontSize = fontSize
	lbl.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	lbl.SetPosition((w-textW*scale)/2, h/2-8)
	c.AddRawChild(lbl)

	return c
}

func addSectionLabel(s *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("section", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	s.AddNode(n)
}

// addComponentBG adds a dark background sprite to a component's node.
func addComponentBG(c *ui.Component, color willow.Color) {
	bg := willow.NewSprite("container-bg", willow.TextureRegion{})
	bg.SetScale(c.Width, c.Height)
	bg.SetColor(color)
	c.AddRawChild(bg)
}
