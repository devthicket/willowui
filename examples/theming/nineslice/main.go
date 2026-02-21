// Nine-slice demonstrates WillowUI nine-slice backgrounds. A 48x48 rounded
// rectangle PNG is generated at startup and used as a nine-slice source for
// button backgrounds. The example shows both nine-slice and solid-color buttons
// side by side for comparison.
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 600
	screenH = 400
)

func main() {
	// Resolve paths relative to this source file so `go run` works from any
	// working directory.
	_, src, _, _ := runtime.Caller(0)
	dir := filepath.Dir(src)

	// Generate a 48x48 rounded-rect nine-slice source image.
	generateButtonPNG(filepath.Join(dir, "button_bg.png"))

	// Load theme with nine-slice backgrounds.
	jsonPath := filepath.Join(dir, "theme.json")
	theme, err := ui.LoadThemeFromFile(jsonPath)
	if err != nil {
		log.Fatalf("load theme: %v", err)
	}

	font := ui.MustLoadDefaultFont()

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "WillowUI: Nine-Slice Demo", font)
	title.TextBlock.FontSize = 22
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(20, 14)
	screen.AddNode(title)

	// Panel with nine-slice theme.
	panel := ui.NewPanel("demo-panel")
	panel.SetTheme(theme)
	panel.SetLayout(ui.LayoutVBox)
	panel.SetSpacing(12)
	panel.Padding = ui.Insets{Top: 14, Left: 14, Right: 14, Bottom: 14}
	panel.SetSize(360, 300)
	panel.SetBackground(willow.RGBA(0.15, 0.15, 0.16, 1))
	panel.SetBorder(willow.RGBA(0.30, 0.30, 0.33, 1), 1)
	panel.SetPosition(20, 55)
	screen.Add(panel)

	// Heading.
	heading := ui.NewLabel("heading", "Nine-Slice Buttons", font, 16)
	panel.AddChild(heading)

	// Nine-slice button (primary variant uses the nine-slice background).
	nsBtn := ui.NewButton("ns-btn", "Nine-Slice Button", font, 14)
	nsBtn.SetSize(200, 40)
	panel.AddChild(nsBtn)

	// Another nine-slice button, wider to show stretching.
	nsBtn2 := ui.NewButton("ns-btn-wide", "Wide Nine-Slice", font, 14)
	nsBtn2.SetSize(300, 40)
	panel.AddChild(nsBtn2)

	// Solid-color button (accent variant falls back to solid).
	solidBtn := ui.NewButton("solid-btn", "Solid Accent Button", font, 14)
	solidBtn.SetSize(200, 40)
	solidBtn.SetVariant(ui.Accent)
	panel.AddChild(solidBtn)

	// Disabled button.
	disBtn := ui.NewButton("dis-btn", "Disabled", font, 14)
	disBtn.SetSize(200, 40)
	disBtn.SetEnabled(false)
	panel.AddChild(disBtn)

	panel.UpdateLayout()

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI: Nine-Slice",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

// generateButtonPNG creates a 48x48 rounded-rectangle PNG suitable for
// nine-slice with 8px insets. The image has a colored border with rounded
// corners and a lighter interior fill.
func generateButtonPNG(path string) {
	const size = 48
	const radius = 8
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	border := color.RGBA{R: 58, G: 122, B: 254, A: 255} // blue border
	fill := color.RGBA{R: 40, G: 80, B: 180, A: 255}    // darker blue fill

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if insideRoundedRect(x, y, size, size, radius) {
				if x < 2 || x >= size-2 || y < 2 || y >= size-2 ||
					!insideRoundedRect(x, y, size, size, radius-2) {
					img.Set(x, y, border)
				} else {
					img.Set(x, y, fill)
				}
			}
		}
	}

	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatalf("encode %s: %v", path, err)
	}
}

// insideRoundedRect returns true if (px, py) is inside a rounded rectangle
// of the given dimensions with the specified corner radius.
func insideRoundedRect(px, py, w, h, r int) bool {
	// Check corners.
	corners := [][2]int{
		{r, r},         // top-left
		{w - r, r},     // top-right (note: last pixel is w-1)
		{r, h - r},     // bottom-left
		{w - r, h - r}, // bottom-right
	}
	for _, c := range corners {
		cx, cy := c[0], c[1]
		dx := 0
		dy := 0
		if px < r {
			dx = cx - px
		} else if px >= w-r {
			dx = px - (cx - 1)
		}
		if py < r {
			dy = cy - py
		} else if py >= h-r {
			dy = py - (cy - 1)
		}
		if dx > 0 && dy > 0 {
			if dx*dx+dy*dy > r*r {
				return false
			}
		}
	}
	return px >= 0 && px < w && py >= 0 && py < h
}
