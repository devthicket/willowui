// ImageCropper widget demo.
// Shows an image with an interactive crop rectangle.
package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 700
	screenH = 500
)

func main() {
	font := ui.MustLoadDefaultFont()

	// Load the test image.
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

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "ImageCropper Widget", font)
	title.TextBlock.FontSize = 22
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", float64(screenW))
	div.SetPosition(0, 46)
	screen.AddNode(div)

	// Status label.
	status := ui.NewLabel("status", "Crop: (drag handles to resize)", font, 13)
	status.SetPosition(430, 70)
	screen.Add(status)

	// Free-form cropper (left).
	freeLabel := willow.NewText("free-label", "Free crop", font)
	freeLabel.TextBlock.FontSize = 14
	freeLabel.TextBlock.Color = willow.RGBA(0.5, 0.6, 0.7, 1)
	freeLabel.SetPosition(30, 55)
	screen.AddNode(freeLabel)

	cropper := ui.NewImageCropper("free-crop")
	cropper.SetImage(src)
	cropper.SetCropRect(20, 15, 90, 80)
	cropper.SetShowGrid(true)
	cropper.SetOnCropChanged(func(rect image.Rectangle) {
		status.SetText(fmt.Sprintf("Crop: %d,%d %dx%d", rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy()))
	})
	cropper.SetSize(380, 380)
	cropper.SetPosition(30, 80)
	screen.Add(cropper)

	// Square-constrained cropper (right).
	sqLabel := willow.NewText("sq-label", "1:1 aspect ratio", font)
	sqLabel.TextBlock.FontSize = 14
	sqLabel.TextBlock.Color = willow.RGBA(0.5, 0.6, 0.7, 1)
	sqLabel.SetPosition(430, 105)
	screen.AddNode(sqLabel)

	sqCropper := ui.NewImageCropper("sq-crop")
	sqCropper.SetImage(src)
	sqCropper.SetAspectRatio(1, 1)
	sqCropper.SetCropRect(25, 10, 80, 80)
	sqCropper.SetShowGrid(true)
	sqCropper.SetMinSize(20, 20)
	sqCropper.SetSize(240, 240)
	sqCropper.SetPosition(430, 130)
	screen.Add(sqCropper)

	// Info labels.
	info1 := willow.NewText("info1", "Drag corners/edges to resize", font)
	info1.TextBlock.FontSize = 12
	info1.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	info1.SetPosition(430, 390)
	screen.AddNode(info1)

	info2 := willow.NewText("info2", "Drag inside crop to move", font)
	info2.TextBlock.FontSize = 12
	info2.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	info2.SetPosition(430, 410)
	screen.AddNode(info2)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "ImageCropper Widget",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
