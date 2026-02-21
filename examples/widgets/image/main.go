// Image widget demo.
// Shows all five scale modes side by side using whelp.png.
package main

import (
	"image/png"
	"log"
	"os"

	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 960
	screenH = 310
)

func main() {
	font := ui.MustLoadDefaultFont()

	// Load whelp.png (128×128).
	f, err := os.Open("examples/_assets/whelp.png")
	if err != nil {
		log.Fatalf("open whelp.png: %v", err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		log.Fatalf("decode whelp.png: %v", err)
	}
	src := ebiten.NewImageFromImage(decoded)

	const (
		sizeLarge = 22.0
		sizeSmall = 13.0
	)

	screen := ui.NewScreen()

	title := willow.NewText("title", "Image Widget  /  Scale Modes", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", float64(screenW))
	div.SetPosition(0, 46)
	screen.AddNode(div)

	const (
		startX  = 30.0
		startY  = 80.0
		colW    = 160.0
		gapX    = 20.0
		boxH    = 100.0 // non-square (160×100) so Stretch/Fit/Fill all behave differently
		bgColor = 0.14  // subtle dark-slate background behind each image widget
		captY   = startY + boxH + 8
	)

	col := func(n int) float64 { return startX + float64(n)*(colW+gapX) }

	// Dark background rect + header label for each column.
	addCol := func(n int, label string) {
		x := col(n)
		bg := willow.NewRect("bg"+label, colW, boxH, willow.RGBA(bgColor, bgColor, bgColor+0.04, 1))
		bg.SetPosition(x, startY)
		screen.AddNode(bg)

		hdr := willow.NewText("lbl-"+label, label, font)
		hdr.TextBlock.FontSize = sizeSmall
		hdr.TextBlock.Color = willow.RGBA(0.45, 0.55, 0.65, 1)
		hdr.SetPosition(x, startY-18)
		hdr.ZIndex_ = 10
		hdr.Invalidate()
		screen.AddNode(hdr)
	}

	caption := func(name, text string, x, y float64) {
		n := willow.NewText(name, text, font)
		n.TextBlock.FontSize = sizeSmall - 1
		n.TextBlock.Color = willow.RGBA(0.55, 0.65, 0.75, 1)
		n.SetPosition(x, y)
		screen.AddNode(n)
	}

	// ── 0. Stretch ──────────────────────────────────────────────────────────
	addCol(0, "Stretch")
	stretch := ui.NewImage("img-stretch")
	stretch.SetImage(src)
	stretch.SetScaleMode(ui.ImageScaleStretch)
	stretch.SetSize(colW, boxH)
	stretch.SetPosition(col(0), startY)
	screen.Add(stretch)
	caption("cap-stretch", "fills bounds,\nignores aspect ratio", col(0), captY)

	// ── 1. Fit (letterbox) ──────────────────────────────────────────────────
	addCol(1, "Fit (letterbox)")
	fit := ui.NewImage("img-fit")
	fit.SetImage(src)
	fit.SetScaleMode(ui.ImageScaleFit)
	fit.SetSize(colW, boxH)
	fit.SetPosition(col(1), startY)
	screen.Add(fit)
	caption("cap-fit", "uniform scale to fit,\nbars show bg", col(1), captY)

	// ── 2. Fill (crop) ──────────────────────────────────────────────────────
	addCol(2, "Fill (crop)")
	fill := ui.NewImage("img-fill")
	fill.SetImage(src)
	fill.SetScaleMode(ui.ImageScaleFill)
	fill.SetSize(colW, boxH)
	fill.SetPosition(col(2), startY)
	screen.Add(fill)
	caption("cap-fill", "uniform scale to fill,\noverflow outside bounds", col(2), captY)

	// ── 3. Center ───────────────────────────────────────────────────────────
	addCol(3, "Center")
	center := ui.NewImage("img-center")
	center.SetImage(src)
	center.SetScaleMode(ui.ImageScaleCenter)
	center.SetSize(colW, boxH)
	center.SetPosition(col(3), startY)
	screen.Add(center)
	caption("cap-center", "native pixel size,\ncentered in bounds", col(3), captY)

	// ── 4. Tile ─────────────────────────────────────────────────────────────
	addCol(4, "Tile")
	tile := ui.NewImage("img-tile")
	tile.SetImage(src)
	tile.SetScaleMode(ui.ImageScaleTile)
	tile.SetSize(colW, boxH)
	tile.SetPosition(col(4), startY)
	screen.Add(tile)
	caption("cap-tile", "repeats at native size\nto fill bounds", col(4), captY)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Image Widget",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
