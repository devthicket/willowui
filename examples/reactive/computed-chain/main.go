// Computed Chain - reactive demo.
// Models an RPG character sheet where base stat inputs flow through three
// levels of Computed nodes: base inputs → effective strength → damage range
// + crit modifier → average DPS. Changing any input cascades automatically
// through all downstream computeds. Ref.Update drives the +/- strength buttons.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 800
	screenH  = 540
	colLeft  = 40.0
	colRight = 440.0
)

var (
	equipNames   = []string{"None", "Iron (+3)", "Steel (+6)", "Enchanted (+12)"}
	equipBonuses = []float64{0, 3, 6, 12}
	critNames    = []string{"Normal (1.0×)", "Sharp (1.5×)", "Critical (2.0×)"}
	critMults    = []float64{1.0, 1.5, 2.0}
)

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	screen := ui.NewScreen()

	title := willow.NewText("title", "Reactive: Computed Chain (RPG Stats)", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	// ── Reactive inputs ───────────────────────────────────────────────────────
	strengthRef := ui.NewRef(10.0) // 1–20
	equipRef := ui.NewRef(0)
	critRef := ui.NewRef(0)

	// ── Left: input controls ──────────────────────────────────────────────────
	y := 62.0

	addHeader(screen, font, sizeSmall, "BASE STATS", colLeft, y)
	y += 20

	// Strength slider
	strSlider := ui.NewSlider("str-slider")
	strSlider.SetRange(1, 20)
	strSlider.SetSize(200, 20)
	strSlider.SetPosition(colLeft, y+4)
	strSlider.BindValue(strengthRef)
	screen.Add(strSlider)

	// +/- buttons — demonstrate Ref.Update
	minusBtn := ui.NewButton("str-minus", "-", font, sizeMedium)
	minusBtn.SetSize(28, 28)
	minusBtn.SetPosition(colLeft+210, y)
	screen.Add(minusBtn)
	minusBtn.SetOnClick(func() {
		strengthRef.Update(func(v float64) float64 {
			if v > 1 {
				return v - 1
			}
			return v
		})
	})

	plusBtn := ui.NewButton("str-plus", "+", font, sizeMedium)
	plusBtn.SetSize(28, 28)
	plusBtn.SetPosition(colLeft+244, y)
	screen.Add(plusBtn)
	plusBtn.SetOnClick(func() {
		strengthRef.Update(func(v float64) float64 {
			if v < 20 {
				return v + 1
			}
			return v
		})
	})

	strValLbl := ui.NewLabel("str-val", "10", font, sizeSmall)
	strValLbl.SetColor(willow.RGBA(0.7, 0.82, 1, 1))
	strValLbl.SetPosition(colLeft+280, y+6)
	screen.Add(strValLbl)
	ui.WatchValue(strengthRef, func(_, v float64) {
		strValLbl.SetText(fmt.Sprintf("%.0f", v))
	})

	y += 36 + 16

	addHeader(screen, font, sizeSmall, "EQUIPMENT", colLeft, y)
	y += 20

	equipRG := ui.NewRadio("equip")
	for _, name := range equipNames {
		equipRG.AddOption(name, font, sizeSmall)
	}
	equipRG.BindSelected(equipRef)
	equipRG.SetPosition(colLeft, y)
	screen.Add(equipRG)

	// 4 options × 20px + 3 gaps × 8px = 104px total
	y += 104 + 18

	addHeader(screen, font, sizeSmall, "CRITICAL HIT", colLeft, y)
	y += 20

	critRG := ui.NewRadio("crit")
	for _, name := range critNames {
		critRG.AddOption(name, font, sizeSmall)
	}
	critRG.BindSelected(critRef)
	critRG.SetPosition(colLeft, y)
	screen.Add(critRG)

	// ── Computed chain ────────────────────────────────────────────────────────

	// Level 1: base + equipment bonus
	effectiveStr := ui.NewComputed(func() float64 {
		return strengthRef.Get() + equipBonuses[equipRef.Get()]
	})

	// Level 2a: damage range from effective strength
	dmgMin := ui.NewComputed(func() float64 {
		return effectiveStr.Get() * 0.8
	})
	dmgMax := ui.NewComputed(func() float64 {
		return effectiveStr.Get() * 1.5
	})

	// Level 2b: crit multiplier from crit tier
	critMultiplier := ui.NewComputed(func() float64 {
		return critMults[critRef.Get()]
	})

	// Level 3: avg DPS = (dmgMin + dmgMax × critMult) / 2
	avgDps := ui.NewComputed(func() float64 {
		return (dmgMin.Get() + dmgMax.Get()*critMultiplier.Get()) / 2.0
	})

	// ── Right: derived stats display ──────────────────────────────────────────
	rx := colRight
	ry := 62.0

	addHeader(screen, font, sizeSmall, "DERIVED STATS", rx, ry)
	ry += 20

	addRule(screen, rx, ry, 340)
	ry += 10

	effStrLbl := addStatRow(screen, font, rx, ry, "Effective Str", "Computed [lvl 1]")
	ry += 26

	addRule(screen, rx, ry, 340)
	ry += 8

	dmgRangeLbl := addStatRow(screen, font, rx, ry, "Damage Range", "Computed [lvl 2]")
	ry += 24
	critMultLbl := addStatRow(screen, font, rx, ry, "Crit Modifier", "Computed [lvl 2]")
	ry += 26

	addRule(screen, rx, ry, 340)
	ry += 8

	avgDpsLbl := addStatRow(screen, font, rx, ry, "Avg DPS", "Computed [lvl 3]")
	ry += 32

	// DPS progress bar. Max DPS ≈ (32×0.8 + 32×1.5×2.0)/2 = 61.6 ≈ 64 for clean display.
	const maxDps = 64.0
	dpsRef := ui.NewRef(0.0)
	dpsBar := ui.NewProgressBar("dps-bar")
	dpsBar.SetSize(340, 14)
	dpsBar.SetPosition(rx, ry)
	dpsBar.BindValue(dpsRef)
	screen.Add(dpsBar)
	ry += 20

	dpsNoteLbl := ui.NewLabel("dps-note", "0 / 64 max", font, sizeSmall)
	dpsNoteLbl.SetColor(willow.RGBA(0.46, 0.52, 0.58, 1))
	dpsNoteLbl.SetPosition(rx, ry)
	screen.Add(dpsNoteLbl)

	// ── WatchEffects propagate computed values to labels ──────────────────────
	ui.WatchEffect(func() {
		effStrLbl.SetText(fmt.Sprintf("%.0f + %.0f = %.0f",
			strengthRef.Get(), equipBonuses[equipRef.Get()], effectiveStr.Get()))
	})

	ui.WatchEffect(func() {
		dmgRangeLbl.SetText(fmt.Sprintf("%.1f – %.1f", dmgMin.Get(), dmgMax.Get()))
	})

	ui.WatchEffect(func() {
		critMultLbl.SetText(fmt.Sprintf("%.1f×", critMultiplier.Get()))
	})

	ui.WatchEffect(func() {
		dps := avgDps.Get()
		avgDpsLbl.SetText(fmt.Sprintf("%.1f", dps))
		dpsRef.Set(dps / maxDps)
		dpsNoteLbl.SetText(fmt.Sprintf("%.1f / %.0f max", dps, maxDps))
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive — Computed Chain",
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

func addRule(screen *ui.Screen, x, y, width float64) {
	d := willow.NewSprite("rule", willow.TextureRegion{})
	d.SetPosition(x, y)
	d.SetScale(width, 1)
	d.SetColor(willow.RGBA(0.20, 0.25, 0.30, 1))
	screen.AddNode(d)
}

func addStatRow(screen *ui.Screen, font *willow.FontFamily, x, y float64, label, badge string) *ui.Label {
	f := font
	nameNode := willow.NewText("sn", label+":", f)
	nameNode.TextBlock.FontSize = 16
	nameNode.TextBlock.Color = willow.RGBA(0.46, 0.55, 0.65, 1)
	nameNode.SetPosition(x, y)
	screen.AddNode(nameNode)

	badgeNode := willow.NewText("sb", badge, f)
	badgeNode.TextBlock.FontSize = 14
	badgeNode.TextBlock.Color = willow.RGBA(0.28, 0.32, 0.38, 1)
	badgeNode.SetPosition(x+228, y+3)
	screen.AddNode(badgeNode)

	valLbl := ui.NewLabel("sv", "--", font, 16)
	valLbl.SetColor(willow.RGBA(1, 0.92, 0.60, 1))
	valLbl.SetPosition(x+142, y)
	screen.Add(valLbl)
	return valLbl
}
