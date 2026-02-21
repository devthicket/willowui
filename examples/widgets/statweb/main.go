// StatWeb demonstrates the WillowUI StatWeb widget: a spider/radar chart
// for visualizing multi-axis attributes with optional editable handles.
package main

import (
	"fmt"
	"log"
	"math"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 640
	screenH = 480
)

func main() {
	font := ui.MustLoadDefaultFont()
	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	screen := ui.NewScreen()

	// ── Title ────────────────────────────────────────────────────────────────
	titleNode := willow.NewText("title", "WillowUI - StatWeb", font)
	titleNode.TextBlock.FontSize = 18
	titleNode.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	titleNode.SetPosition(24, 18)
	screen.AddNode(titleNode)

	// ── Read-only stat web (5 axes) ─────────────────────────────────────────
	sectionLabel(screen, font, "READ-ONLY", 24, 54)

	web1 := ui.NewStatWeb("readonly-stats", font, 11)
	web1.SetAxes([]ui.StatAxis{
		{Name: "STR", Min: 0, Max: 100, Value: 75},
		{Name: "AGI", Min: 0, Max: 100, Value: 60},
		{Name: "VIT", Min: 0, Max: 100, Value: 85},
		{Name: "WIS", Min: 0, Max: 100, Value: 40},
		{Name: "LCK", Min: 0, Max: 100, Value: 55},
	})
	web1.SetSize(240, 240)
	web1.SetPosition(20, 74)
	screen.Add(web1)

	// ── Point allocation web ────────────────────────────────────────────────
	// Fixed budget: dragging one axis up forces the others down proportionally.
	const totalBudget = 30.0

	sectionLabel(screen, font, "POINT ALLOCATION (budget: 30)", 350, 54)

	allocWeb := ui.NewStatWeb("alloc-stats", font, 11)
	allocAxes := []ui.StatAxis{
		{Name: "ATK", Min: 0, Max: 15, Value: 6},
		{Name: "DEF", Min: 0, Max: 15, Value: 6},
		{Name: "SPD", Min: 0, Max: 15, Value: 6},
		{Name: "INT", Min: 0, Max: 15, Value: 6},
		{Name: "LCK", Min: 0, Max: 15, Value: 6},
	}
	allocWeb.SetAxes(allocAxes)
	allocWeb.SetEditable(true)
	allocWeb.SetSize(240, 240)
	allocWeb.SetPosition(370, 74)

	budgetLbl := ui.NewLabel("budget", fmtBudget(allocAxes, totalBudget), font, 11)
	budgetLbl.SetColor(willow.RGBA(0.5, 0.7, 0.5, 1))
	budgetLbl.SetPosition(370, 326)
	screen.Add(budgetLbl)

	allocWeb.SetOnValueChanged(func(index int, value float64) {
		axes := allocWeb.Axes()
		used := sumValues(axes)
		over := used - totalBudget
		if over > 0.5 {
			// Steal proportionally from the other axes.
			redistributeBudget(allocWeb, axes, index, over)
			axes = allocWeb.Axes()
		}
		budgetLbl.SetText(fmtBudget(axes, totalBudget))
	})
	screen.Add(allocWeb)

	// ── Weighted axes (different max per axis) ──────────────────────────────
	sectionLabel(screen, font, "WEIGHTED AXES", 24, 338)

	web3 := ui.NewStatWeb("weighted-stats", font, 10)
	web3.SetAxes([]ui.StatAxis{
		{Name: "HP", Min: 0, Max: 999, Value: 450},
		{Name: "MP", Min: 0, Max: 500, Value: 320},
		{Name: "ATK", Min: 0, Max: 200, Value: 85},
		{Name: "DEF", Min: 0, Max: 200, Value: 120},
	})
	web3.SetEditable(true)
	web3.SetSize(200, 130)
	web3.SetPosition(50, 360)
	screen.Add(web3)

	// ── 3-axis minimal ──────────────────────────────────────────────────────
	sectionLabel(screen, font, "MINIMAL (3 axes)", 340, 338)

	web4 := ui.NewStatWeb("triangle-stats", font, 10)
	web4.SetAxes([]ui.StatAxis{
		{Name: "Power", Min: 0, Max: 10, Value: 7},
		{Name: "Speed", Min: 0, Max: 10, Value: 5},
		{Name: "Guard", Min: 0, Max: 10, Value: 8},
	})
	web4.SetSize(160, 130)
	web4.SetPosition(400, 360)
	screen.Add(web4)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "StatWeb - WillowUI",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.10, 0.10, 0.12, 1),
	})
}

// redistributeBudget steals `over` points proportionally from axes other than
// the one the user just dragged, keeping all values >= 0.
func redistributeBudget(web *ui.StatWeb, axes []ui.StatAxis, changed int, over float64) {
	// Sum of the other axes' values.
	otherSum := 0.0
	for i, a := range axes {
		if i != changed {
			otherSum += a.Value
		}
	}
	if otherSum < 0.01 {
		// Nothing to steal from — clamp the changed axis instead.
		web.SetValue(changed, axes[changed].Value-over)
		return
	}
	for i := range axes {
		if i == changed {
			continue
		}
		share := axes[i].Value / otherSum * over
		newVal := math.Max(0, axes[i].Value-share)
		web.SetValue(i, math.Round(newVal*10)/10)
	}
}

func sumValues(axes []ui.StatAxis) float64 {
	s := 0.0
	for _, a := range axes {
		s += a.Value
	}
	return s
}

func fmtBudget(axes []ui.StatAxis, budget float64) string {
	used := sumValues(axes)
	return fmt.Sprintf("Used: %.0f / %.0f", used, budget)
}

func sectionLabel(screen *ui.Screen, font *willow.FontFamily, text string, x, y float64) {
	node := willow.NewText("sec-"+text, text, font)
	node.TextBlock.FontSize = 12
	node.TextBlock.Color = willow.RGBA(0.45, 0.50, 0.70, 1)
	node.SetPosition(x, y)
	screen.AddNode(node)
}
