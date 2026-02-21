// Sprite-icons demonstrates WillowUI's theme-driven icon system. A small
// sprite sheet is generated at startup containing a checkmark, radio dot,
// close X, and tree expand/collapse arrows. The theme JSON references these
// sprites so that checkbox, radio, window, and tree list widgets all render
// with custom icons instead of the default procedural glyphs.
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
	screenW = 700
	screenH = 500
)

func main() {
	_, src, _, _ := runtime.Caller(0)
	dir := filepath.Dir(src)

	generateIconsPNG(filepath.Join(dir, "icons.png"))

	theme, err := ui.LoadThemeFromFile(filepath.Join(dir, "theme.json"))
	if err != nil {
		log.Fatalf("load theme: %v", err)
	}

	font := ui.MustLoadDefaultFont()

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "WillowUI: Sprite Icons Demo", font)
	title.TextBlock.FontSize = 22
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(20, 14)
	screen.AddNode(title)

	// --- Left column: Checkbox + Radio ---
	leftPanel := ui.NewPanel("left-panel")
	leftPanel.SetTheme(theme)
	leftPanel.SetLayout(ui.LayoutVBox)
	leftPanel.SetSpacing(16)
	leftPanel.Padding = ui.Insets{Top: 14, Left: 14, Right: 14, Bottom: 14}
	leftPanel.SetSize(240, 340)
	leftPanel.SetPosition(20, 55)
	screen.Add(leftPanel)

	// Checkbox section.
	cbHeading := ui.NewLabel("cb-heading", "Checkboxes (theme icon)", font, 14)
	leftPanel.AddChild(cbHeading)

	cb1 := ui.NewCheckbox("cb1", "Option Alpha", font, 14)
	cb1.SetChecked(true)
	leftPanel.AddChild(cb1)

	cb2 := ui.NewCheckbox("cb2", "Option Beta", font, 14)
	leftPanel.AddChild(cb2)

	cb3 := ui.NewCheckbox("cb3", "Disabled", font, 14)
	cb3.SetEnabled(false)
	leftPanel.AddChild(cb3)

	// Radio section.
	radioHeading := ui.NewLabel("radio-heading", "Radio (theme icon)", font, 14)
	leftPanel.AddChild(radioHeading)

	rg := ui.NewRadio("demo-radio")
	rg.AddOption("Small", font, 14)
	rg.AddOption("Medium", font, 14)
	rg.AddOption("Large", font, 14)
	rg.SetSelected(1)
	leftPanel.AddChild(rg)

	leftPanel.UpdateLayout()

	// --- Right column: TreeList ---
	treePanel := ui.NewPanel("tree-panel")
	treePanel.SetTheme(theme)
	treePanel.SetLayout(ui.LayoutVBox)
	treePanel.SetSpacing(8)
	treePanel.Padding = ui.Insets{Top: 14, Left: 14, Right: 14, Bottom: 14}
	treePanel.SetSize(300, 340)
	treePanel.SetPosition(280, 55)
	screen.Add(treePanel)

	treeHeading := ui.NewLabel("tree-heading", "TreeList (theme icons)", font, 14)
	treePanel.AddChild(treeHeading)

	tree := ui.NewTreeList("demo-tree", 24)
	tree.SetSize(272, 270)
	tree.SetDefaultTextRenderer(font, 13, 272, 24)
	tree.SetRoots([]*ui.TreeNode{
		{Data: "Documents", Children: []*ui.TreeNode{
			{Data: "Work"},
			{Data: "Personal", Children: []*ui.TreeNode{
				{Data: "Photos"},
				{Data: "Notes"},
			}},
		}},
		{Data: "Downloads", Children: []*ui.TreeNode{
			{Data: "Apps"},
			{Data: "Music"},
		}},
		{Data: "Desktop"},
	})
	tree.ExpandAll()
	treePanel.AddChild(tree)

	treePanel.UpdateLayout()

	// --- Window with themed close icon ---
	win := ui.NewWindow("demo-win", "Theme Close Icon", font, 14)
	win.SetTheme(theme)
	win.SetSize(260, 180)
	win.SetPosition(200, 160)
	win.SetResizable(true)
	ui.DefaultWindowManager.Add(win)

	winLabel := ui.NewLabel("win-label", "This window uses the\ntheme's close-x sprite.", font, 13)
	win.Body().AddChild(winLabel)
	win.Body().UpdateLayout()

	screen.Add(win)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI: Sprite Icons",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

// generateIconsPNG creates a sprite sheet with all icons laid out horizontally.
// Layout (each icon in its own 12-16px cell):
//
//	x=0:  11x11 checkmark
//	x=16:  8x8  radio dot (at y=1 to center in 11px row)
//	x=28: 11x11 close X
//	x=44:  9x9  tree expand (right-pointing triangle, at y=1)
//	x=56:  9x9  tree collapse (down-pointing triangle, at y=1)
//	x=68: 12x12 resize grip (diagonal dots)
func generateIconsPNG(path string) {
	const w, h = 84, 12
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	// Checkmark (11x11) at (0,0): a V-shape.
	drawCheckmark(img, 0, 0, white)

	// Radio dot (8x8) at (16,1): filled circle.
	drawCircle(img, 16, 1, 8, white)

	// Close X (11x11) at (28,0): two diagonal lines.
	drawCloseX(img, 28, 0, 11, white)

	// Tree expand (9x9) at (44,1): right-pointing triangle.
	drawTriangleRight(img, 44, 1, 9, white)

	// Tree collapse (9x9) at (56,1): down-pointing triangle.
	drawTriangleDown(img, 56, 1, 9, white)

	// Resize grip (12x12) at (68,0): diagonal dot pattern.
	drawResizeGrip(img, 68, 0, 12, white)

	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatalf("encode %s: %v", path, err)
	}
}

func drawCheckmark(img *image.NRGBA, ox, oy int, c color.NRGBA) {
	// V-shape: descend from (1,5) to (4,8), then ascend from (4,8) to (9,3).
	// Two pixel wide strokes.
	pts := [][2]int{
		{1, 5}, {2, 6}, {3, 7}, {4, 8},
		{5, 7}, {6, 6}, {7, 5}, {8, 4}, {9, 3},
	}
	for _, p := range pts {
		img.SetNRGBA(ox+p[0], oy+p[1], c)
		img.SetNRGBA(ox+p[0], oy+p[1]-1, c)
	}
}

func drawCircle(img *image.NRGBA, ox, oy, size int, c color.NRGBA) {
	r := float64(size) / 2
	cx, cy := r, r
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			if dx*dx+dy*dy <= r*r {
				img.SetNRGBA(ox+x, oy+y, c)
			}
		}
	}
}

func drawCloseX(img *image.NRGBA, ox, oy, size int, c color.NRGBA) {
	for i := 0; i < size; i++ {
		img.SetNRGBA(ox+i, oy+i, c)
		img.SetNRGBA(ox+size-1-i, oy+i, c)
		if i+1 < size {
			img.SetNRGBA(ox+i+1, oy+i, c)
			img.SetNRGBA(ox+size-2-i, oy+i, c)
		}
	}
}

func drawTriangleRight(img *image.NRGBA, ox, oy, size int, c color.NRGBA) {
	// Right-pointing filled triangle.
	mid := size / 2
	for row := 0; row < size; row++ {
		dist := mid - row
		if dist < 0 {
			dist = -dist
		}
		cols := mid - dist + 1
		for col := 0; col < cols; col++ {
			img.SetNRGBA(ox+col, oy+row, c)
		}
	}
}

func drawTriangleDown(img *image.NRGBA, ox, oy, size int, c color.NRGBA) {
	// Down-pointing filled triangle.
	mid := size / 2
	for col := 0; col < size; col++ {
		dist := mid - col
		if dist < 0 {
			dist = -dist
		}
		rows := mid - dist + 1
		for row := 0; row < rows; row++ {
			img.SetNRGBA(ox+col, oy+row, c)
		}
	}
}

func drawResizeGrip(img *image.NRGBA, ox, oy, size int, c color.NRGBA) {
	// Three diagonal rows of 2x2 dots in the bottom-right triangle,
	// resembling a standard resize grip.
	dots := [][2]int{
		// Bottom-right single dot.
		{9, 9},
		// Middle diagonal.
		{5, 9}, {9, 5},
		// Top diagonal.
		{1, 9}, {5, 5}, {9, 1},
	}
	for _, d := range dots {
		for dy := 0; dy < 2; dy++ {
			for dx := 0; dx < 2; dx++ {
				px, py := d[0]+dx, d[1]+dy
				if px < size && py < size {
					img.SetNRGBA(ox+px, oy+py, c)
				}
			}
		}
	}
}
