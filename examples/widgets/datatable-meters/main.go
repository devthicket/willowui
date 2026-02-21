// DataTable meters example — demonstrates MeterColumn with dynamic fill colors.
//
// Shows a party of characters with HP and XP as inline meter bars.
// HP uses OnPostUpdate to color-code: red < 30%, yellow < 70%, green >= 70%.
// XP uses the default theme fill color.
package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 700
	screenH = 500
)

type Character struct {
	Name  string
	Class string
	Level int
	HP    float64 // 0-1
	XP    float64 // 0-1
}

var classes = []string{
	"Warrior", "Mage", "Rogue", "Cleric", "Paladin",
	"Archer", "Druid", "Bard", "Monk", "Necromancer",
	"Ranger", "Berserker",
}

var names = []string{
	"Alice", "Bob", "Charlie", "Diana", "Eve",
	"Frank", "Grace", "Hank", "Iris", "Jack",
	"Kate", "Leo", "Maya", "Nathan", "Olivia",
	"Peter", "Quinn", "Rosa", "Sam", "Tara",
}

func makeRoster(n int) []Character {
	out := make([]Character, n)
	for i := range out {
		// Vary HP/XP using sine waves so values spread across 0-1.
		hp := 0.5 + 0.5*math.Sin(float64(i)*0.7)
		xp := 0.5 + 0.5*math.Cos(float64(i)*0.5)
		out[i] = Character{
			Name:  fmt.Sprintf("%s%d", names[i%len(names)], i+1),
			Class: classes[i%len(classes)],
			Level: 1 + (i*7)%60,
			HP:    hp,
			XP:    xp,
		}
	}
	return out
}

func main() {
	font := ui.MustLoadDefaultFont()

	screen := ui.NewScreen()

	// Title
	title := willow.NewText("title", "DataTable - Meter Columns", font)
	title.TextBlock.FontSize = 20
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 12)
	screen.AddNode(title)

	// DataTable
	dt := ui.NewDataTable("party", 32)
	dt.SetFont(font, 13)
	dt.SetSize(652, 420)
	dt.SetPosition(24, 50)
	dt.SetHeaderHeight(36)
	dt.SetZebraStriping(true)
	dt.SetShowColumnDividers(true)

	// Name column
	nameCol := ui.LabelColumn("name", "Name", func(d any) string {
		return d.(Character).Name
	})
	nameCol.Weight = 2
	nameCol.Sortable = true
	nameCol.SortType = ui.SortAlpha
	dt.AddColumn(nameCol)

	// Class column
	classCol := ui.LabelColumn("class", "Class", func(d any) string {
		return d.(Character).Class
	})
	classCol.Weight = 1.5
	classCol.Sortable = true
	dt.AddColumn(classCol)

	// Level column
	levelCol := ui.LabelColumn("level", "Lv", func(d any) string {
		return strconv.Itoa(d.(Character).Level)
	})
	levelCol.FixedWidth = 50
	levelCol.Sortable = true
	levelCol.SortType = ui.SortNumeric
	levelCol.Cell.Align = willow.TextAlignRight
	dt.AddColumn(levelCol)

	// HP meter — color-coded via OnPostUpdate
	hpCol := ui.MeterColumn("hp", "HP", func(d any) float64 {
		return d.(Character).HP
	})
	hpCol.Weight = 2
	hpCol.Cell.OnPostUpdate = func(d any, comp *ui.Component) {
		hp := d.(Character).HP
		if mb, ok := comp.UserData().(*ui.MeterBar); ok {
			if hp < 0.3 {
				mb.SetFillColor(willow.RGBA(0.9, 0.2, 0.2, 1)) // red
			} else if hp < 0.7 {
				mb.SetFillColor(willow.RGBA(0.9, 0.8, 0.2, 1)) // yellow
			} else {
				mb.SetFillColor(willow.RGBA(0.2, 0.9, 0.3, 1)) // green
			}
		}
	}
	dt.AddColumn(hpCol)

	// XP meter — default theme color
	xpCol := ui.MeterColumn("xp", "XP", func(d any) float64 {
		return d.(Character).XP
	})
	xpCol.Weight = 2
	dt.AddColumn(xpCol)

	dt.BindItems(ui.NewArrayFromAny(makeRoster(60)))
	screen.Add(dt)

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "DataTable Meters",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.07, 0.08, 0.10, 1),
	})
}
