// Font-Family demonstrates loading a custom FontFamily from a .fontbundle and
// exercising all style variants: regular, bold, italic, bold+italic, multiple
// display sizes, and RichText with inline markup.
//
// Prerequisites:
//
//	go generate ./examples/theming/font-family/
package main

//go:generate go run gen.go

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

func loadLatoBundle() []byte {
	_, src, _, _ := runtime.Caller(0)
	p := filepath.Join(filepath.Dir(src), "..", "..", "_assets", "fonts", "lato.fontbundle")
	data, err := os.ReadFile(p)
	if err != nil {
		log.Fatalf("load lato.fontbundle: %v", err)
	}
	return data
}

const (
	screenW = 800
	screenH = 640
)

func main() {
	lato, err := willow.NewFontFamilyFromFontBundle(loadLatoBundle())
	if err != nil {
		log.Fatalf("load lato bundle: %v", err)
	}

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "Lato Font Family Demo", lato)
	title.TextBlock.FontSize = 32
	title.TextBlock.Bold = true
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW-48)
	div.SetPosition(24, 56)
	screen.AddNode(div)

	// ── Section 1: Display sizes ─────────────────────────────────────────────
	sectionLabel(screen, lato, "Multiple display sizes from one family", 24, 68)

	sizes := []float64{12, 16, 24, 36}
	y := 90.0
	for _, sz := range sizes {
		lbl := ui.NewLabel("size", "The quick brown fox jumps over the lazy dog", lato, sz)
		lbl.SetPosition(40, y)
		screen.Add(lbl)
		y += sz + 12
	}

	// ── Section 2: Style variants ────────────────────────────────────────────
	y += 8
	sectionLabel(screen, lato, "Style variants: regular, bold, italic, bold+italic", 24, y)
	y += 24

	type variant struct {
		text   string
		bold   bool
		italic bool
	}
	variants := []variant{
		{"Regular style", false, false},
		{"Bold style", true, false},
		{"Italic style", false, true},
		{"Bold + Italic style", true, true},
	}
	for _, v := range variants {
		lbl := ui.NewLabel("variant", v.text, lato, 20)
		lbl.SetBold(v.bold)
		lbl.SetItalic(v.italic)
		lbl.SetPosition(40, y)
		screen.Add(lbl)
		y += 32
	}

	// ── Section 3: RichText with markup ──────────────────────────────────────
	y += 8
	sectionLabel(screen, lato, "RichText with <b>/<i> markup resolved automatically", 24, y)
	y += 24

	rt := ui.NewRichText("rich", lato, 18)
	rt.SetWrapWidth(screenW - 80)
	rt.SetPosition(40, y)
	if err := rt.SetMarkup(
		"This is <b>bold</b>, this is <i>italic</i>, and this is <b><i>bold italic</i></b>. " +
			"The FontFamily resolves each style variant at render time.",
	); err != nil {
		log.Printf("markup error: %v", err)
	}
	screen.Add(rt)
	y += 90

	// ── Section 4: Side-by-side size comparison ──────────────────────────────
	sectionLabel(screen, lato, "Side-by-side: 14 px vs 32 px (different atlas selection)", 24, y)
	y += 24

	small := ui.NewLabel("cmp-small", "Lato 14px", lato, 14)
	small.SetPosition(40, y)
	screen.Add(small)

	large := ui.NewLabel("cmp-large", "Lato 32px", lato, 32)
	large.SetPosition(200, y-8)
	screen.Add(large)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI: Font Family Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func sectionLabel(screen *ui.Screen, font *willow.FontFamily, text string, x, y float64) {
	n := willow.NewText("section", text, font)
	n.TextBlock.FontSize = 14
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
