// Anchor layout demonstrates pinning children to edges and corners of a
// parent container using LayoutAnchor. This is the primary tool for HUD
// composition — health bars bottom-left, minimaps top-right, score counters
// top-center, etc.
package main

import (
	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 600
)

var (
	colorBg      = willow.RGBA(0.12, 0.12, 0.14, 1)
	colorPanel   = willow.RGBA(0.18, 0.20, 0.24, 1)
	colorAccent  = willow.RGBA(0.26, 0.52, 0.96, 1)
	colorSuccess = willow.RGBA(0.25, 0.78, 0.45, 1)
	colorWarning = willow.RGBA(0.95, 0.68, 0.20, 1)
	colorDanger  = willow.RGBA(0.90, 0.30, 0.30, 1)
	colorInfo    = willow.RGBA(0.30, 0.70, 0.90, 1)
	colorText    = willow.RGBA(0.93, 0.93, 0.93, 1)
	colorDim     = willow.RGBA(0.50, 0.55, 0.60, 1)
	colorCenter  = willow.RGBA(0.55, 0.35, 0.85, 1)
)

type controller struct {
	font *willow.FontFamily
}

func (c *controller) OnCreate(s *ui.Screen) {
	font := c.font
	const fontSize = 14.0

	// ── Title ──────────────────────────────────────────────────────────────
	title := willow.NewText("title", "LayoutAnchor: HUD Composition", font)
	title.TextBlock.FontSize = 22
	title.TextBlock.Color = colorText
	title.SetPosition(24, 12)
	s.AddNode(title)

	subtitle := willow.NewText("subtitle", "Children pinned to edges and corners with pixel offsets", font)
	subtitle.TextBlock.FontSize = 16
	subtitle.TextBlock.Color = colorDim
	subtitle.SetPosition(24, 40)
	s.AddNode(subtitle)

	// ── Anchor Layout (simulated game viewport) ────────────────────────────
	// Use a plain Panel with Layout = LayoutAnchor — no special widget needed.
	al := ui.NewPanel("hud")
	al.Layout = ui.LayoutAnchor
	al.SetSize(screenW-48, screenH-80)
	al.SetBackground(colorPanel)
	al.SetBorder(willow.RGBA(0.30, 0.30, 0.33, 1), 1)
	al.SetPosition(24, 60)
	al.Padding = ui.Insets{Top: 12, Right: 12, Bottom: 12, Left: 12}
	s.Add(al)

	// Top-left: health bar area
	healthBox := makeHUDPanel("Health", 180, 50, colorDanger, font, fontSize)
	al.AddAnchoredChild(healthBox, ui.AnchorTopLeft, 0, 0)

	// Top-center: wave/round indicator
	waveBox := makeHUDPanel("Wave 3 / 10", 140, 36, colorAccent, font, fontSize)
	al.AddAnchoredChild(waveBox, ui.AnchorTopCenter, 0, 0)

	// Top-right: minimap placeholder
	minimapBox := makeHUDPanel("Minimap", 120, 120, colorInfo, font, fontSize)
	al.AddAnchoredChild(minimapBox, ui.AnchorTopRight, 0, 0)

	// Middle-left: quest tracker
	questBox := makeHUDPanel("Quests", 160, 100, colorDim, font, fontSize)
	al.AddAnchoredChild(questBox, ui.AnchorMiddleLeft, 0, 0)

	// Center: pause/notification area
	centerBox := makeHUDPanel("PAUSED", 160, 60, colorCenter, font, fontSize)
	al.AddAnchoredChild(centerBox, ui.AnchorCenter, 0, 0)

	// Middle-right: buffs/debuffs
	buffsBox := makeHUDPanel("Buffs", 100, 80, colorSuccess, font, fontSize)
	al.AddAnchoredChild(buffsBox, ui.AnchorMiddleRight, 0, 0)

	// Bottom-left: chat log
	chatBox := makeHUDPanel("Chat", 220, 80, colorDim, font, fontSize)
	al.AddAnchoredChild(chatBox, ui.AnchorBottomLeft, 0, 0)

	// Bottom-center: action bar
	actionBar := makeHUDPanel("Action Bar", 300, 50, colorWarning, font, fontSize)
	al.AddAnchoredChild(actionBar, ui.AnchorBottomCenter, 0, 0)

	// Bottom-right: gold/inventory
	goldBox := makeHUDPanel("1,234 Gold", 130, 36, colorSuccess, font, fontSize)
	al.AddAnchoredChild(goldBox, ui.AnchorBottomRight, 0, 0)

	al.UpdateLayout()
}

func (c *controller) OnUpdate(dt float64) {}
func (c *controller) OnDestroy()          {}

func main() {
	font := ui.MustLoadDefaultFont()

	ui.Stage.Add(ui.NewScreen(ui.WithController(&controller{font: font})))
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Anchor Layout",
		Width:      screenW,
		Height:     screenH,
		ClearColor: colorBg,
	})
}

// makeHUDPanel creates a small labeled panel representing a HUD element.
func makeHUDPanel(label string, w, h float64, accentColor willow.Color, font *willow.FontFamily, fontSize float64) *ui.Panel {
	p := ui.NewPanel(label)
	p.SetSize(w, h)
	p.SetBackground(willow.RGBA(0.10, 0.10, 0.12, 0.85))
	p.SetBorder(accentColor, 2)

	// Accent strip along the top.
	strip := willow.NewSprite(label+"-strip", willow.TextureRegion{})
	strip.SetScale(w, 3)
	strip.SetColor(accentColor)
	p.AddRawChild(strip)

	// Centered label.
	textW, _ := font.MeasureString(label, 0, false, false)
	scale := fontSize / font.LineHeight(0, false, false)
	lbl := willow.NewText(label+"-lbl", label, font)
	lbl.TextBlock.FontSize = fontSize
	lbl.TextBlock.Color = colorText
	lbl.SetPosition((w-textW*scale)/2, h/2-fontSize/2)
	p.AddRawChild(lbl)

	return p
}
