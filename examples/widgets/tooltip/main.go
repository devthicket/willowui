// Tooltip demonstrates WillowUI's Tooltip component: hover delays, anchor
// positions (above, below, right, follow-mouse), rich multi-child content,
// shared tooltips with dynamic content, programmatic show/hide, and
// screen-edge clamping that keeps tooltips fully visible.
package main

import (
	"fmt"
	"log"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 760
)

func main() {
	font := ui.MustLoadDefaultFont()
	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	const (
		sizeLarge  = 20.0
		sizeMedium = 15.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "WillowUI: Tooltip Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div0 := ui.NewDivider("div0", screenW-48)
	div0.SetPosition(24, 42)
	screen.AddNode(div0)

	// ── Section 1: Anchor positions ──────────────────────────────────────────
	addSection(screen, font, sizeSmall, "Anchor positions: hover each button", 24, 52)

	// Below (default)
	btnBelow := ui.NewButton("btn-below", "Hover → Below", font, sizeMedium)
	btnBelow.SetSize(150, 38)
	btnBelow.SetPosition(24, 74)
	screen.Add(btnBelow)
	ttBelow := ui.NewTooltip("tip-below")
	ttBelow.Anchor = ui.TooltipBelow
	ttBelow.ShowDelay = 15
	ttBelow.SetText("This tooltip appears below.", font, sizeSmall)
	btnBelow.SetTooltip(ttBelow)

	// Above
	btnAbove := ui.NewButton("btn-above", "Hover → Above", font, sizeMedium)
	btnAbove.SetSize(150, 38)
	btnAbove.SetPosition(194, 74)
	screen.Add(btnAbove)
	ttAbove := ui.NewTooltip("tip-above")
	ttAbove.Anchor = ui.TooltipAbove
	ttAbove.ShowDelay = 15
	ttAbove.SetText("This tooltip appears above.", font, sizeSmall)
	btnAbove.SetTooltip(ttAbove)

	// Right
	btnRight := ui.NewButton("btn-right", "Hover → Right", font, sizeMedium)
	btnRight.SetSize(150, 38)
	btnRight.SetPosition(364, 74)
	screen.Add(btnRight)
	ttRight := ui.NewTooltip("tip-right")
	ttRight.Anchor = ui.TooltipRight
	ttRight.ShowDelay = 15
	ttRight.SetText("This tooltip appears to the right.", font, sizeSmall)
	btnRight.SetTooltip(ttRight)

	// SetTooltipText convenience
	btnConvenience := ui.NewButton("btn-conv", "SetTooltipText", font, sizeMedium)
	btnConvenience.SetSize(150, 38)
	btnConvenience.SetPosition(534, 74)
	screen.Add(btnConvenience)
	btnConvenience.SetTooltipText("Convenience shorthand -- one call.", font, sizeSmall)

	// ── Section 2: Rich content tooltip ──────────────────────────────────────
	addSection(screen, font, sizeSmall, "Rich content: multi-child tooltip with heading and body", 24, 134)

	itemPanel := ui.NewPanel("item-panel")
	itemPanel.SetSize(200, 60)
	itemPanel.SetBackground(willow.RGBA(0.18, 0.18, 0.22, 1))
	itemPanel.SetBorder(willow.RGBA(0.35, 0.35, 0.40, 1), 1)
	itemPanel.SetPosition(24, 156)
	screen.Add(itemPanel)

	itemLabel := ui.NewLabel("item-label", "⚔  Iron Sword", font, sizeMedium)
	itemLabel.SetColor(willow.RGBA(0.9, 0.85, 0.7, 1))
	itemLabel.SetPosition(12, 14)
	itemPanel.AddChild(itemLabel)

	itemSub := ui.NewLabel("item-sub", "hover for details", font, sizeSmall)
	itemSub.SetColor(willow.RGBA(0.5, 0.5, 0.55, 1))
	itemSub.SetPosition(12, 38)
	itemPanel.AddChild(itemSub)

	richTip := buildRichTooltip(font, sizeMedium, sizeSmall)
	itemPanel.SetTooltip(richTip)

	// ── Section 3: Follow-mouse tooltip ──────────────────────────────────────
	addSection(screen, font, sizeSmall, "Follow-mouse: tooltip tracks the cursor", 24, 240)

	hoverZone := ui.NewPanel("hover-zone")
	hoverZone.SetSize(280, 54)
	hoverZone.SetBackground(willow.RGBA(0.14, 0.20, 0.28, 1))
	hoverZone.SetBorder(willow.RGBA(0.26, 0.52, 0.96, 0.5), 1)
	hoverZone.SetPosition(24, 262)
	screen.Add(hoverZone)

	hoverHint := ui.NewLabel("hover-hint", "Move mouse over this zone", font, sizeSmall)
	hoverHint.SetColor(willow.RGBA(0.55, 0.75, 1.0, 1))
	hoverHint.SetPosition(16, 20)
	hoverZone.AddChild(hoverHint)

	followTip := ui.NewTooltip("tip-follow")
	followTip.Anchor = ui.TooltipFollowMouse
	followTip.OffsetX = 16
	followTip.OffsetY = 20
	followTip.ShowDelay = 0
	followTip.SetText("Follows the cursor!", font, sizeSmall)
	hoverZone.SetTooltip(followTip)

	// ── Section 4: Shared tooltip with dynamic content ───────────────────────
	addSection(screen, font, sizeSmall, "Shared tooltip: one tooltip, dynamic content per trigger", 24, 340)

	colorItems := []struct {
		name  string
		color willow.Color
	}{
		{"Crimson", willow.RGBA(0.86, 0.08, 0.24, 1)},
		{"Amber", willow.RGBA(1.0, 0.75, 0.0, 1)},
		{"Jade", willow.RGBA(0.0, 0.66, 0.42, 1)},
		{"Cobalt", willow.RGBA(0.0, 0.45, 0.85, 1)},
		{"Violet", willow.RGBA(0.54, 0.17, 0.89, 1)},
	}

	// Single shared tooltip with one label updated via onTooltipShow.
	sharedTip := ui.NewTooltip("tip-shared")
	sharedTip.Anchor = ui.TooltipBelow
	sharedTip.ShowDelay = 10
	sharedTip.Padding = ui.Insets{Top: 6, Right: 12, Bottom: 6, Left: 12}
	sharedLabel := ui.NewLabel("shared-lbl", "", font, sizeSmall)
	sharedTip.AddChild(sharedLabel)

	x := 24.0
	for _, item := range colorItems {
		item := item // capture
		swatch := ui.NewPanel("swatch-" + item.name)
		swatch.SetSize(50, 50)
		swatch.SetBackground(item.color)
		swatch.SetPosition(x, 362)
		screen.Add(swatch)
		swatch.SetTooltip(sharedTip)
		swatch.SetOnTooltipShow(func() {
			sharedLabel.SetText(item.name)
		})
		x += 60
	}

	// ── Section 5: Fade + programmatic show/hide ─────────────────────────────
	addSection(screen, font, sizeSmall, "Fade & programmatic show/hide", 24, 438)

	btnFade := ui.NewButton("btn-fade", "Hover -- Fade 20fr", font, sizeMedium)
	btnFade.SetSize(180, 38)
	btnFade.SetPosition(24, 460)
	screen.Add(btnFade)
	fadeTip := ui.NewTooltip("tip-fade")
	fadeTip.ShowDelay = 10
	fadeTip.FadeInDuration = 0.33
	fadeTip.FadeOutDuration = 0.33
	fadeTip.SetText("Fades in and out over 0.33 seconds.", font, sizeSmall)
	btnFade.SetTooltip(fadeTip)

	progTip := ui.NewTooltip("tip-prog")
	progTip.Anchor = ui.TooltipCornerBottomRight
	progTip.OffsetX = -16
	progTip.OffsetY = -16
	progTip.SetText("Shown programmatically at bottom-right.", font, sizeSmall)

	showing := false
	btnToggle := ui.NewButton("btn-toggle", "Toggle Tip", font, sizeMedium)
	btnToggle.SetSize(140, 38)
	btnToggle.SetPosition(224, 460)
	screen.Add(btnToggle)
	btnToggle.SetOnClick(func() {
		if showing {
			progTip.Hide()
			showing = false
			btnToggle.SetText("Toggle Tip")
		} else {
			progTip.Show(0, 0) // position is overridden by corner anchor
			showing = true
			btnToggle.SetText("Hide Tip")
		}
	})

	// Status label shows active tooltip name
	statusLabel := ui.NewLabel("status", "", font, sizeSmall)
	statusLabel.SetColor(willow.RGBA(0.5, 0.5, 0.55, 1))
	statusLabel.SetPosition(24, 520)
	screen.Add(statusLabel)

	// Wire onTooltipShow/Hide on the fade button to update status
	activeCount := 0
	btnFade.SetOnTooltipShow(func() {
		activeCount++
		statusLabel.SetText(fmt.Sprintf("onTooltipShow fired (%d total)", activeCount))
	})
	btnFade.SetOnTooltipHide(func() {
		statusLabel.SetText(fmt.Sprintf("onTooltipHide fired (shown %d times)", activeCount))
	})

	// ── Section 6: Screen-edge clamping ──────────────────────────────────────
	div6 := ui.NewDivider("div6", screenW-48)
	div6.SetPosition(24, 558)
	screen.AddNode(div6)
	addSection(screen, font, sizeSmall, "Screen-edge clamping: hover buttons near edges to see auto-adjustment", 24, 566)

	// Button near the right edge: TooltipRight would overflow → clamped left.
	btnEdgeRight := ui.NewButton("btn-edge-right", "Right-anchored →", font, sizeMedium)
	btnEdgeRight.SetSize(158, 38)
	btnEdgeRight.SetPosition(float64(screenW)-166, 588)
	screen.Add(btnEdgeRight)
	ttEdgeRight := ui.NewTooltip("tip-edge-right")
	ttEdgeRight.Anchor = ui.TooltipRight
	ttEdgeRight.ShowDelay = 10
	ttEdgeRight.SetText("Clamped: would overflow\nthe right screen edge.", font, sizeSmall)
	btnEdgeRight.SetTooltip(ttEdgeRight)

	// Wide zone: follow-mouse tooltip clamps as the cursor approaches any edge.
	followClampZone := ui.NewPanel("follow-clamp-zone")
	followClampZone.SetSize(float64(screenW)-48, 58)
	followClampZone.SetBackground(willow.RGBA(0.12, 0.18, 0.12, 1))
	followClampZone.SetBorder(willow.RGBA(0.2, 0.55, 0.2, 0.5), 1)
	followClampZone.SetPosition(24, 636)
	screen.Add(followClampZone)

	followClampHint := ui.NewLabel("follow-clamp-hint", "Move mouse to screen edges -- tooltip stays on-screen", font, sizeSmall)
	followClampHint.SetColor(willow.RGBA(0.4, 0.85, 0.4, 1))
	followClampHint.SetPosition(16, 22)
	followClampZone.AddChild(followClampHint)

	ttFollowClamp := ui.NewTooltip("tip-follow-clamp")
	ttFollowClamp.Anchor = ui.TooltipFollowMouse
	ttFollowClamp.OffsetX = 14
	ttFollowClamp.OffsetY = 18
	ttFollowClamp.ShowDelay = 0
	ttFollowClamp.SetText("I follow the mouse\nbut won't go off-screen!", font, sizeSmall)
	followClampZone.SetTooltip(ttFollowClamp)

	// Button near bottom-left: TooltipBelow would overflow → clamped upward.
	btnEdgeBottom := ui.NewButton("btn-edge-bottom", "Below-anchored ↓", font, sizeMedium)
	btnEdgeBottom.SetSize(158, 38)
	btnEdgeBottom.SetPosition(24, float64(screenH)-46)
	screen.Add(btnEdgeBottom)
	ttEdgeBottom := ui.NewTooltip("tip-edge-bottom")
	ttEdgeBottom.Anchor = ui.TooltipBelow
	ttEdgeBottom.ShowDelay = 10
	ttEdgeBottom.SetText("Clamped: would overflow\nthe bottom screen edge.", font, sizeSmall)
	btnEdgeBottom.SetTooltip(ttEdgeBottom)

	// Button near bottom-right corner: overflows both axes → clamped on both.
	btnEdgeCorner := ui.NewButton("btn-edge-corner", "Corner ↘", font, sizeMedium)
	btnEdgeCorner.SetSize(158, 38)
	btnEdgeCorner.SetPosition(float64(screenW)-166, float64(screenH)-46)
	screen.Add(btnEdgeCorner)
	ttEdgeCorner := ui.NewTooltip("tip-edge-corner")
	ttEdgeCorner.Anchor = ui.TooltipBelow
	ttEdgeCorner.ShowDelay = 10
	ttEdgeCorner.SetText("Clamped on both axes:\nbottom and right edges.", font, sizeSmall)
	btnEdgeCorner.SetTooltip(ttEdgeCorner)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Tooltip Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

// buildRichTooltip constructs a multi-child tooltip for the item card.
func buildRichTooltip(font *willow.FontFamily, sizeMedium, sizeSmall float64) *ui.Tooltip {
	tt := ui.NewTooltip("tip-rich")
	tt.Layout = ui.LayoutVBox
	tt.Spacing = 5
	tt.Padding = ui.Insets{Top: 10, Right: 14, Bottom: 10, Left: 14}
	tt.Anchor = ui.TooltipRight
	tt.ShowDelay = 10

	heading := ui.NewLabel("h", "Iron Sword", font, sizeMedium)
	heading.SetColor(willow.RGBA(1.0, 0.85, 0.35, 1))
	tt.AddChild(heading)

	divNode := ui.NewPanel("div")
	divNode.SetSize(180, 1)
	divNode.SetBackground(willow.RGBA(0.35, 0.35, 0.40, 1))
	tt.AddChild(divNode)

	stat1 := ui.NewLabel("s1", "Attack  +12", font, sizeSmall)
	stat1.SetColor(willow.RGBA(0.8, 0.9, 0.8, 1))
	tt.AddChild(stat1)

	stat2 := ui.NewLabel("s2", "Weight   2", font, sizeSmall)
	stat2.SetColor(willow.RGBA(0.8, 0.9, 0.8, 1))
	tt.AddChild(stat2)

	desc := ui.NewLabel("desc", "A sturdy iron blade,\nreliable and affordable.", font, sizeSmall)
	desc.SetColor(willow.RGBA(0.65, 0.65, 0.70, 1))
	tt.AddChild(desc)

	return tt
}

func addSection(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("sec", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
