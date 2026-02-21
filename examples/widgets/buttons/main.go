// Buttons demonstrates WillowUI's Label and Button components: reactive text
// binding, click callbacks, visual states (normal, hovered, disabled), and
// an IconButton. A counter label increments each time the "Increment" button
// is clicked.
package main

import (
	"fmt"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 600
)

func main() {
	font := ui.MustLoadDefaultFont()
	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	const (
		sizeLarge  = 24.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI: Labels & Buttons Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW-48)
	div.SetPosition(24, 48)
	screen.AddNode(div)

	// ── Section 1: Labels ────────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Labels: static text, colors, alignment", 24, 58)

	staticLabel := ui.NewLabel("static", "Hello, WillowUI!", font, sizeMedium)
	staticLabel.SetPosition(40, 80)
	screen.Add(staticLabel)

	coloredLabel := ui.NewLabel("colored", "Colored Label", font, sizeMedium)
	coloredLabel.SetColor(willow.RGBA(0.3, 0.9, 0.5, 1))
	coloredLabel.SetPosition(40, 106)
	screen.Add(coloredLabel)

	wrappedLabel := ui.NewLabel("wrapped", "This label has word wrapping enabled so long text fits within a set width.", font, sizeMedium)
	wrappedLabel.SetWrapWidth(250)
	wrappedLabel.SetPosition(40, 132)
	screen.Add(wrappedLabel)

	// ── Section 2: Reactive Label ────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Reactive label: bound to a Ref, updates on click", 320, 58)

	counter := ui.NewRef(0)
	counterText := ui.NewRef("Count: 0")

	counterLabel := ui.NewLabel("counter-label", "", font, sizeLarge)
	counterLabel.SetColor(willow.RGBA(1, 0.85, 0.2, 1))
	counterLabel.BindText(counterText)
	counterLabel.SetPosition(340, 80)
	screen.Add(counterLabel)

	// ── Section 3: Buttons ───────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Buttons: normal, hover, disabled, click callback", 24, 210)

	incBtn := ui.NewButton("inc-btn", "Increment", font, sizeMedium)
	incBtn.SetSize(140, 40)
	incBtn.SetOnClick(func() {
		v := counter.Get() + 1
		counter.Set(v)
		counterText.Set(fmt.Sprintf("Count: %d", v))
	})
	incBtn.SetPosition(40, 234)
	screen.Add(incBtn)

	resetBtn := ui.NewButton("reset-btn", "Reset", font, sizeMedium)
	resetBtn.SetSize(100, 40)
	resetBtn.SetOnClick(func() {
		counter.Set(0)
		counterText.Set("Count: 0")
	})
	resetBtn.SetPosition(200, 234)
	screen.Add(resetBtn)

	disabledBtn := ui.NewButton("disabled-btn", "Disabled", font, sizeMedium)
	disabledBtn.SetSize(120, 40)
	disabledBtn.SetEnabled(false)
	disabledBtn.SetPosition(320, 234)
	screen.Add(disabledBtn)

	// ── Section 4: Multiple buttons row ──────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "Button row: styled with different sizes", 24, 300)

	buttons := []struct {
		text string
		w, h float64
	}{
		{"Small", 80, 32},
		{"Medium", 120, 40},
		{"Large", 160, 48},
		{"Extra Wide", 200, 40},
	}

	bx := 40.0
	for _, spec := range buttons {
		btn := ui.NewButton("row-"+spec.text, spec.text, font, sizeMedium)
		btn.SetSize(spec.w, spec.h)
		btn.SetPosition(bx, 324)
		screen.Add(btn)
		bx += spec.w + 16
	}

	// ── Section 5: IconButton ────────────────────────────────────────────────
	addSectionLabel(screen, font, sizeSmall, "IconButton: icon-only | icon+label below | icon+label right", 24, 400)

	// Icon-only.
	iconBtn := ui.NewIconButton("icon-btn")
	iconBtn.SetSize(48, 48)
	iconBtn.SetIconSize(28, 28)
	iconBtn.SetOnClick(func() {
		v := counter.Get() + 10
		counter.Set(v)
		counterText.Set(fmt.Sprintf("Count: %d", v))
	})
	iconBtn.SetPosition(40, 420)
	screen.Add(iconBtn)

	// Icon + label below.
	iconLabelBtn := ui.NewIconButton("icon-label-btn")
	iconLabelBtn.SetSize(64, 68)
	iconLabelBtn.SetIconSize(28, 28)
	iconLabelBtn.SetLabel("Save", font, sizeSmall)
	iconLabelBtn.SetLabelPosition(ui.IconLabelBelow)
	iconLabelBtn.SetOnClick(func() {
		counter.Set(0)
		counterText.Set("Count: 0")
	})
	iconLabelBtn.SetPosition(106, 416)
	screen.Add(iconLabelBtn)

	// Icon + label right.
	iconRightBtn := ui.NewIconButton("icon-right-btn")
	iconRightBtn.SetSize(88, 40)
	iconRightBtn.SetIconSize(18, 18)
	iconRightBtn.SetLabel("Open", font, sizeSmall)
	iconRightBtn.SetLabelPosition(ui.IconLabelRight)
	iconRightBtn.SetPosition(188, 430)
	screen.Add(iconRightBtn)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Labels & Buttons Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addSectionLabel(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("section", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
