// Text Input - reactive demo.
// Shows TextInput.BindValue(Ref[string]) and TextArea.BindValue(Ref[string]).
// A Computed label derives word count and character count from the text area.
package main

import (
	"fmt"
	"strings"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 800
	screenH  = 460
	colLeft  = 40.0
	colRight = 450.0
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	title := willow.NewText("title", "Reactive - Text Input", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	y := 62.0

	// ── 1. TextInput ─────────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "TextInput.BindValue - Ref[string], echo label on the right", colLeft, y)
	y += 20

	nameRef := ui.NewRef("")

	ti := ui.NewTextInput("ti", font, sizeMedium)
	ti.SetWidth(340)
	ti.SetPlaceholder("Enter your name...")
	ti.BindValue(nameRef)
	ti.SetPosition(colLeft, y)
	screen.Add(ti)

	echoLbl := addStatus(screen, font, sizeMedium, colRight, y+8)
	echoLbl.SetColor(willow.RGBA(0.4, 0.9, 0.6, 1))

	greetComputed := ui.NewComputed(func() string {
		n := strings.TrimSpace(nameRef.Get())
		if n == "" {
			return "Hello, stranger!"
		}
		return fmt.Sprintf("Hello, %s!", n)
	})
	ui.WatchEffect(func() {
		echoLbl.SetText(greetComputed.Get())
	})

	// Programmatic set
	setBtn := ui.NewButton("set", "Set to \"WillowUI\"", font, sizeSmall)
	setBtn.SetSize(160, 28)
	setBtn.SetPosition(colLeft, y+44)
	screen.Add(setBtn)
	setBtn.SetOnClick(func() {
		nameRef.Set("WillowUI")
	})

	clearBtn := ui.NewButton("clear", "Clear", font, sizeSmall)
	clearBtn.SetSize(80, 28)
	clearBtn.SetPosition(colLeft+170, y+44)
	screen.Add(clearBtn)
	clearBtn.SetOnClick(func() {
		nameRef.Set("")
	})

	y += 80 + 20

	// ── 2. TextArea ───────────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "TextArea.BindValue - Computed word/char count on the right", colLeft, y)
	y += 20

	bodyRef := ui.NewRef("")

	ta := ui.NewTextArea("ta", font, sizeMedium)
	ta.SetSize(340, 100)
	ta.BindValue(bodyRef)
	ta.SetPosition(colLeft, y)
	screen.Add(ta)

	wordCount := ui.NewComputed(func() int {
		return len(strings.Fields(bodyRef.Get()))
	})
	charCount := ui.NewComputed(func() int {
		return len([]rune(bodyRef.Get()))
	})

	wcLbl := addStatus(screen, font, sizeSmall, colRight, y+10)
	ccLbl := addStatus(screen, font, sizeSmall, colRight, y+28)
	ccLbl.SetColor(willow.RGBA(1, 0.85, 0.4, 1))

	ui.WatchEffect(func() {
		wcLbl.SetText(fmt.Sprintf("words: %d", wordCount.Get()))
	})
	ui.WatchEffect(func() {
		ccLbl.SetText(fmt.Sprintf("chars: %d", charCount.Get()))
	})

	// Preset buttons
	setBodyBtn := ui.NewButton("set-body", "Load sample", font, sizeSmall)
	setBodyBtn.SetSize(120, 28)
	setBodyBtn.SetPosition(colLeft, y+108)
	screen.Add(setBodyBtn)
	setBodyBtn.SetOnClick(func() {
		bodyRef.Set("The quick brown fox\njumps over the lazy dog.")
	})

	clearBodyBtn := ui.NewButton("clear-body", "Clear", font, sizeSmall)
	clearBodyBtn.SetSize(80, 28)
	clearBodyBtn.SetPosition(colLeft+130, y+108)
	screen.Add(clearBodyBtn)
	clearBodyBtn.SetOnClick(func() {
		bodyRef.Set("")
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive - Text Input",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addHeader(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("hdr", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.45, 0.55, 0.65, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}

func addStatus(screen *ui.Screen, font *willow.FontFamily, fontSize, x, y float64) *ui.Label {
	lbl := ui.NewLabel("status", "...", font, fontSize)
	lbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	lbl.SetPosition(x, y)
	screen.Add(lbl)
	return lbl
}
