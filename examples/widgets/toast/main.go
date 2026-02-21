// Toast demonstrates WillowUI's Toast notification system: fire-and-forget
// messages that appear at a screen corner, auto-dismiss, stack, and animate.
package main

import (
	"log"
	"time"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 740
	screenH = 560
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		titleSize   = 18.0
		sectionSize = 12.0
		buttonSize  = 14.0
		toastSize   = 14.0
	)

	// Configure the default toast manager once at startup.
	ui.DefaultToastManager.SetFont(font, toastSize)
	ui.DefaultToastManager.SetAnchor(ui.ToastBottomRight)
	ui.DefaultToastManager.SetMaxStack(4)
	ui.DefaultToastManager.SetMargin(16, 16)

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	titleNode := willow.NewText("title", "WillowUI: Toast", font)
	titleNode.TextBlock.FontSize = titleSize
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetPosition(24, 18)
	screen.AddNode(titleNode)

	// Two-column layout: left column x=24, right column x=260
	const (
		colX1  = 24.0
		colX2  = 260.0
		bw     = 210.0
		bh     = 36.0
		bgap   = 8.0
		secGap = 18.0
	)

	// ── Left column: Variants ────────────────────────────────────────────────
	y1 := 58.0

	addSection(screen, font, sectionSize, "VARIANTS", colX1, y1)
	y1 += secGap

	addBtn(screen, font, buttonSize, "btn-info", "Info", colX1, y1, bw, bh, func() {
		ui.ShowToast("This is an info message.", ui.Info)
	})
	y1 += bh + bgap

	addBtn(screen, font, buttonSize, "btn-success", "Success", colX1, y1, bw, bh, func() {
		ui.ShowToast("Operation completed!", ui.Success)
	})
	y1 += bh + bgap

	addBtn(screen, font, buttonSize, "btn-warning", "Warning", colX1, y1, bw, bh, func() {
		ui.ShowToast("Low memory warning.", ui.Warning,
			ui.WithDuration(5*time.Second))
	})
	y1 += bh + bgap

	addBtn(screen, font, buttonSize, "btn-error", "Error", colX1, y1, bw, bh, func() {
		ui.ShowToast("Connection failed.", ui.Danger)
	})
	y1 += bh + bgap*2 + 4

	// ── Left column: Options ──────────────────────────────────────────────────
	addSection(screen, font, sectionSize, "OPTIONS", colX1, y1)
	y1 += secGap

	addBtn(screen, font, buttonSize, "btn-progress", "With progress bar", colX1, y1, bw, bh, func() {
		ui.ShowToast("Saving...", ui.Info,
			ui.WithProgress(true),
			ui.WithDuration(4*time.Second))
	})
	y1 += bh + bgap

	addBtn(screen, font, buttonSize, "btn-nodismiss", "No click-dismiss", colX1, y1, bw, bh, func() {
		ui.ShowToast("Click won't close this.", ui.Primary,
			ui.WithDismissOnClick(false),
			ui.WithDuration(3*time.Second))
	})
	y1 += bh + bgap

	addBtn(screen, font, buttonSize, "btn-callback", "On-dismiss callback", colX1, y1, bw, bh, func() {
		ui.ShowToast("Watch the console.", ui.Primary,
			ui.WithOnDismiss(func() {
				log.Println("toast dismissed")
			}))
	})
	y1 += bh + bgap

	addBtn(screen, font, buttonSize, "btn-short", "Short (1 s)", colX1, y1, bw, bh, func() {
		ui.ShowToast("Gone in a second!", ui.Primary, ui.WithDuration(time.Second))
	})

	// ── Right column: Stack / Control ─────────────────────────────────────────
	y2 := 58.0

	addSection(screen, font, sectionSize, "STACK", colX2, y2)
	y2 += secGap

	addBtn(screen, font, buttonSize, "btn-stack", "Spam 4 toasts", colX2, y2, bw, bh, func() {
		ui.ShowToast("First", ui.Info)
		ui.ShowToast("Second", ui.Success)
		ui.ShowToast("Third", ui.Warning)
		ui.ShowToast("Fourth", ui.Danger)
	})
	y2 += bh + bgap

	addBtn(screen, font, buttonSize, "btn-dismiss-all", "Dismiss all", colX2, y2, bw, bh, func() {
		ui.DefaultToastManager.DismissAll()
	})
	y2 += bh + bgap*2 + 4

	// ── Right column: Anchor ──────────────────────────────────────────────────
	addSection(screen, font, sectionSize, "ANCHOR CORNER", colX2, y2)
	y2 += secGap

	type anchorRow struct {
		id     string
		label  string
		anchor ui.ToastAnchor
	}
	for _, a := range []anchorRow{
		{"br", "Bottom-right (default)", ui.ToastBottomRight},
		{"bl", "Bottom-left", ui.ToastBottomLeft},
		{"tr", "Top-right", ui.ToastTopRight},
		{"tl", "Top-left", ui.ToastTopLeft},
	} {
		a := a
		addBtn(screen, font, buttonSize, "btn-anchor-"+a.id, a.label, colX2, y2, bw, bh, func() {
			ui.DefaultToastManager.SetAnchor(a.anchor)
			ui.ShowToast("Moved to: "+a.label, ui.Primary)
		})
		y2 += bh + bgap
	}

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Toast — WillowUI",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.10, 0.10, 0.12, 1),
	})
}

func addBtn(screen *ui.Screen, font *willow.FontFamily, size float64, id, label string, x, y, w, h float64, fn func()) {
	btn := ui.NewButton(id, label, font, size)
	btn.SetSize(w, h)
	btn.SetPosition(x, y)
	btn.SetOnClick(fn)
	screen.Add(btn)
}

func addSection(screen *ui.Screen, font *willow.FontFamily, size float64, label string, x, y float64) {
	node := willow.NewText("sec-"+label, label, font)
	node.TextBlock.FontSize = size
	node.TextBlock.Color = willow.RGBA(0.45, 0.50, 0.70, 1)
	node.SetPosition(x, y)
	screen.AddNode(node)
}
