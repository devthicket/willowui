// DragHandle - resizable panel demo.
// Shows three panels with DragHandle edges: bottom (Y), right (X), and
// bottom-right corner (diagonal), plus live size readout and min/max annotations.
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

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeSmall  = 14.0
		handleSize = 8.0
	)

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "DragHandle - Resizable Panels", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	// ── 1. Panel resizable from bottom edge (Y axis) ───────────────────────
	sectionLabel(screen, font, sizeSmall, "Bottom edge resize (DragAxisY) -- min:60 max:300", 30, 58)

	panelY := ui.NewPanel("panel-y")
	panelY.SetSize(220, 120)
	panelY.SetPosition(30, 80)
	panelY.SetVariant(ui.Accent)
	screen.Add(panelY)

	handleY := ui.NewDragHandle("handle-y")
	handleY.SetSize(220, handleSize)
	handleY.SetAxis(ui.DragAxisY)
	handleY.SetTarget(&panelY.Component)
	handleY.SetMin(60)
	handleY.SetMax(300)
	handleY.SetPosition(30, 80+120)
	screen.Add(handleY)

	sizeYLabel := ui.NewLabel("size-y", "220 x 120", font, sizeSmall)
	sizeYLabel.SetColor(willow.RGBA(0.7, 0.9, 0.5, 1))
	sizeYLabel.SetPosition(30, 80+120+handleSize+4)
	screen.Add(sizeYLabel)

	handleY.SetOnDrag(func(_ float64) {
		sizeYLabel.SetText(fmt.Sprintf("%.0f x %.0f", panelY.Width, panelY.Height))
		// Move handle to stay attached to panel bottom.
		handleY.SetPosition(30, 80+panelY.Height)
		sizeYLabel.SetPosition(30, 80+panelY.Height+handleSize+4)
	})

	// ── 2. Panel resizable from right edge (X axis) ────────────────────────
	sectionLabel(screen, font, sizeSmall, "Right edge resize (DragAxisX) -- min:80 max:350", 300, 58)

	panelX := ui.NewPanel("panel-x")
	panelX.SetSize(200, 160)
	panelX.SetPosition(300, 80)
	panelX.SetVariant(ui.Success)
	screen.Add(panelX)

	handleX := ui.NewDragHandle("handle-x")
	handleX.SetSize(handleSize, 160)
	handleX.SetAxis(ui.DragAxisX)
	handleX.SetGripStyle(ui.DragGripLines)
	handleX.SetTarget(&panelX.Component)
	handleX.SetMin(80)
	handleX.SetMax(350)
	handleX.SetPosition(300+200, 80)
	screen.Add(handleX)

	sizeXLabel := ui.NewLabel("size-x", "200 x 160", font, sizeSmall)
	sizeXLabel.SetColor(willow.RGBA(0.7, 0.9, 0.5, 1))
	sizeXLabel.SetPosition(300, 248)
	screen.Add(sizeXLabel)

	handleX.SetOnDrag(func(_ float64) {
		sizeXLabel.SetText(fmt.Sprintf("%.0f x %.0f", panelX.Width, panelX.Height))
		handleX.SetPosition(300+panelX.Width, 80)
	})

	// ── 3. Panel resizable from corner (diagonal) ──────────────────────────
	sectionLabel(screen, font, sizeSmall, "Corner resize (DragAxisDiagonal) -- min:80 max:400", 30, 400)

	panelD := ui.NewPanel("panel-d")
	panelD.SetSize(200, 120)
	panelD.SetPosition(30, 422)
	panelD.SetVariant(ui.Warning)
	screen.Add(panelD)

	handleD := ui.NewDragHandle("handle-d")
	handleD.SetSize(12, 12)
	handleD.SetAxis(ui.DragAxisDiagonal)
	handleD.SetTarget(&panelD.Component)
	handleD.SetMin(80)
	handleD.SetMax(400)
	handleD.SetPosition(30+200-12, 422+120-12)
	screen.Add(handleD)

	sizeDLabel := ui.NewLabel("size-d", "200 x 120", font, sizeSmall)
	sizeDLabel.SetColor(willow.RGBA(0.7, 0.9, 0.5, 1))
	sizeDLabel.SetPosition(30, 422+120+4)
	screen.Add(sizeDLabel)

	handleD.SetOnDrag(func(_ float64) {
		sizeDLabel.SetText(fmt.Sprintf("%.0f x %.0f", panelD.Width, panelD.Height))
		handleD.SetPosition(30+panelD.Width-12, 422+panelD.Height-12)
		sizeDLabel.SetPosition(30, 422+panelD.Height+4)
	})

	// ── 4. Delegate mode (no target) ───────────────────────────────────────
	sectionLabel(screen, font, sizeSmall, "Delegate mode (no target) -- reports delta", 300, 400)

	delegateHandle := ui.NewDragHandle("delegate")
	delegateHandle.SetSize(200, 24)
	delegateHandle.SetAxis(ui.DragAxisY)
	delegateHandle.SetGripStyle(ui.DragGripDots)
	delegateHandle.SetPosition(300, 422)
	screen.Add(delegateHandle)

	deltaLabel := ui.NewLabel("delta", "delta: 0", font, sizeSmall)
	deltaLabel.SetColor(willow.RGBA(0.7, 0.8, 1, 1))
	deltaLabel.SetPosition(300, 452)
	screen.Add(deltaLabel)

	delegateHandle.SetOnDrag(func(delta float64) {
		deltaLabel.SetText(fmt.Sprintf("delta: %.1f", delta))
	})
	delegateHandle.SetOnDragEnd(func(v float64) {
		deltaLabel.SetText(fmt.Sprintf("final: %.1f", v))
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "DragHandle - Resizable Panels",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func sectionLabel(screen *ui.Screen, font *willow.FontFamily, size float64, text string, x, y float64) {
	n := willow.NewText("hdr", text, font)
	n.TextBlock.FontSize = size
	n.TextBlock.Color = willow.RGBA(0.45, 0.55, 0.65, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
