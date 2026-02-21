// DataTable example — character roster demo.
// Demonstrates: LabelColumn, multi-sort, per-column filters, CellStyle,
// OnPostUpdate, BindSearchInput, filtering, single selection, arrow-key
// navigation, BindItems (reactive array), scroll virtualization, and the
// empty state.
//
// New features shown:
// - Shift+click header for multi-column sort (numbered badges appear)
// - Score column has OnPostUpdate (green > 10000, red < 5000)
// - Class column is Filterable (click funnel icon to filter by class)
// - Search box at top filters across all searchable columns
// - Reset button clears all sort/filter state
// - Default sort: Name ascending on load
package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 900
	screenH = 600
)

type Character struct {
	Name       string
	Level      int
	Class      string
	ClassOrder int // custom sort order for Class column
	Score      int
}

// classOrder defines a custom sort order for classes (tank > healer > DPS).
var classOrder = map[string]int{
	"Warrior":     1,
	"Paladin":     2,
	"Berserker":   3,
	"Cleric":      4,
	"Druid":       5,
	"Monk":        6,
	"Ranger":      7,
	"Archer":      8,
	"Rogue":       9,
	"Bard":        10,
	"Mage":        11,
	"Necromancer": 12,
}

// roster is the primary 25-entry dataset as a reactive array.
var roster = ui.NewArrayFromAny([]Character{
	{"Alice", 9, "Warrior", 1, 9800},
	{"Bob", 100, "Mage", 11, 7600},
	{"Charlie", 5, "Rogue", 9, 11200},
	{"Diana", 29, "Cleric", 4, 6400},
	{"Eve", 1, "Paladin", 2, 13500},
	{"Frank", 1000, "Archer", 8, 4200},
	{"Grace", 44, "Druid", 5, 9100},
	{"Hank", 3, "Bard", 10, 7000},
	{"Iris", 200, "Monk", 6, 10800},
	{"Jack", 10, "Necromancer", 12, 5900},
	{"Kate", 48, "Ranger", 7, 10200},
	{"Leo", 2, "Berserker", 3, 7800},
	{"Maya", 41, "Mage", 11, 9200},
	{"Nathan", 58, "Warrior", 1, 12100},
	{"Olivia", 7, "Rogue", 9, 5100},
	{"Peter", 46, "Cleric", 4, 9600},
	{"Quinn", 34, "Bard", 10, 7200},
	{"Rosa", 500, "Paladin", 2, 14800},
	{"Sam", 19, "Archer", 8, 3800},
	{"Tara", 53, "Druid", 5, 11000},
	{"Ugo", 11, "Necromancer", 12, 6100},
	{"Vera", 45, "Monk", 6, 9400},
	{"Will", 22, "Warrior", 1, 4600},
	{"Xena", 99, "Ranger", 7, 12600},
	{"Yuki", 8, "Mage", 11, 8300},
})

// largeClasses cycles through class names for the generated large dataset.
var largeClasses = []string{
	"Warrior", "Mage", "Rogue", "Cleric", "Paladin",
	"Archer", "Druid", "Bard", "Monk", "Necromancer",
}

// makeLargeRoster generates 100 entries for scroll virtualization testing.
func makeLargeRoster() *ui.Array[any] {
	names := []string{
		"Aiden", "Bella", "Connor", "Dana", "Ethan",
		"Faye", "George", "Hana", "Ivan", "Jade",
	}
	items := make([]Character, 100)
	for i := 0; i < 100; i++ {
		name := names[i%len(names)] + strconv.Itoa(i+1)
		class := largeClasses[i%len(largeClasses)]
		level := 10 + (i*3)%55
		score := 2000 + i*120
		items[i] = Character{name, level, class, classOrder[class], score}
	}
	return ui.NewArrayFromAny(items)
}

func main() {
	font := ui.MustLoadDefaultFont()

	theme, err := ui.LoadThemeRelative("../../_themes/dark.json")
	if err != nil {
		log.Fatalf("theme: %v", err)
	}
	ui.DefaultTheme = theme

	screen := ui.NewScreen()

	// -----------------------------------------------------------------------
	// Title
	// -----------------------------------------------------------------------
	title := willow.NewText("title", "DataTable - Character Roster", font)
	title.TextBlock.FontSize = 20
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 8)
	screen.AddNode(title)

	// -----------------------------------------------------------------------
	// Search box (top right, beside title)
	// -----------------------------------------------------------------------
	searchInput := ui.NewTextInput("search", font, 13)
	searchInput.SetWidth(180)
	searchInput.SetPosition(696, 6)
	searchInput.SetPlaceholder("Search...")
	screen.Add(searchInput)

	// -----------------------------------------------------------------------
	// Row-count label (updated by callbacks)
	// -----------------------------------------------------------------------
	countLabel := willow.NewText("count", "25 rows", font)
	countLabel.TextBlock.FontSize = 13
	countLabel.TextBlock.Color = willow.RGBA(0.6, 0.8, 0.6, 1)
	countLabel.SetPosition(810, 44)
	screen.AddNode(countLabel)

	updateCount := func(n int) {
		countLabel.SetContent(fmt.Sprintf("%d rows", n))
	}

	// -----------------------------------------------------------------------
	// DataTable
	// -----------------------------------------------------------------------
	dt := ui.NewDataTable("roster", 28)
	dt.SetFont(font, 13)
	dt.SetSize(852, 520)
	dt.SetPosition(24, 72)
	dt.SetHeaderHeight(36)
	dt.SetZebraStriping(true)
	dt.SetShowColumnDividers(true)
	dt.SetSelectionMode(ui.SelectionModeSingle)

	// Default sort: Name ascending.
	dt.SetDefaultSort("name", ui.SortAsc)

	// Name column — LabelColumn auto-populates SearchValue from the accessor.
	nameCol := ui.LabelColumn("name", "Name", func(data any) string {
		return data.(Character).Name
	})
	nameCol.Weight = 2
	nameCol.Sortable = true
	nameCol.SortType = ui.SortAlpha
	dt.AddColumn(nameCol)

	// Level column — numeric sort uses SearchValue string which toFloat64 parses.
	levelCol := ui.LabelColumn("level", "Level", func(data any) string {
		return strconv.Itoa(data.(Character).Level)
	})
	levelCol.Weight = 1
	levelCol.Sortable = true
	levelCol.SortType = ui.SortNumeric
	levelCol.Cell.Align = willow.TextAlignRight
	dt.AddColumn(levelCol)

	// Class column — sorts by ClassOrder (tank > healer > DPS) via SortValue.
	// Filterable: click the funnel icon in the header to filter by class.
	classCol := ui.LabelColumn("class", "Class", func(data any) string {
		return data.(Character).Class
	})
	classCol.Weight = 1.5
	classCol.Sortable = true
	classCol.Filterable = true
	classCol.SortType = ui.SortNumeric
	classCol.SortValue = func(data any) any {
		return data.(Character).ClassOrder
	}
	dt.AddColumn(classCol)

	// Score column — numeric sort with ColorFunc.
	// Green for high scores (>10000), red for low (<5000), white otherwise.
	scoreCol := ui.LabelColumn("score", "Score", func(data any) string {
		return strconv.Itoa(data.(Character).Score)
	})
	scoreCol.Weight = 1
	scoreCol.Sortable = true
	scoreCol.SortType = ui.SortNumeric
	scoreCol.Cell.Align = willow.TextAlignRight
	scoreCol.Cell.OnPostUpdate = func(data any, comp *ui.Component) {
		score := data.(Character).Score
		var c willow.Color
		if score > 10000 {
			c = willow.RGBA(0.3, 1.0, 0.3, 1) // green
		} else if score < 5000 {
			c = willow.RGBA(1.0, 0.3, 0.3, 1) // red
		} else {
			c = willow.RGBA(0.9, 0.9, 0.9, 1) // white-ish
		}
		if l, ok := comp.UserData().(*ui.Label); ok {
			l.SetColor(c)
		}
	}
	scoreCol.HeaderStyle = ui.CellStyle{
		Color: willow.RGBA(1.0, 0.85, 0.4, 1), // gold header
	}
	dt.AddColumn(scoreCol)

	// Bind search input to the table.
	dt.BindSearchInput(searchInput)

	// Empty state shown when no rows match the filter.
	emptyLbl := ui.NewLabel("empty", "No matching rows", font, 15)
	emptyLbl.SetPosition(24+836/2-70, 72+36+20)
	dt.SetEmptyComponent(&emptyLbl.Component)

	// Load initial dataset and wire count callback.
	dt.BindItems(roster)
	dt.SetOnSelectionChanged(func(_ []int) {})

	screen.Add(dt)

	// -----------------------------------------------------------------------
	// Toolbar button helper
	// -----------------------------------------------------------------------
	addBtn := func(name, label string, x, y, w float64, fn func()) {
		btn := ui.NewButton(name, label, font, 13)
		btn.SetSize(w, 28)
		btn.SetPosition(x, y)
		btn.SetOnClick(fn)
		screen.Add(btn)
	}

	// -----------------------------------------------------------------------
	// Row 1 (y=40): Dataset + Filter buttons
	// -----------------------------------------------------------------------
	addBtn("btn-roster", "Roster", 24, 40, 58, func() {
		dt.ResetFiltersAndSort()
		dt.BindItems(roster)
		updateCount(dt.DataTableDisplayCount())
	})
	addBtn("btn-large", "Large", 88, 40, 54, func() {
		dt.ResetFiltersAndSort()
		dt.BindItems(makeLargeRoster())
		updateCount(dt.DataTableDisplayCount())
	})
	addBtn("btn-empty", "Empty", 148, 40, 52, func() {
		dt.ResetFiltersAndSort()
		dt.BindItems(nil)
		updateCount(dt.DataTableDisplayCount())
	})

	// Separator gap, then filter buttons
	addBtn("btn-all", "All", 240, 40, 40, func() {
		dt.SetFilterFunc(nil)
		updateCount(dt.DataTableDisplayCount())
	})
	addBtn("btn-mages", "Mages", 286, 40, 56, func() {
		dt.SetFilterFunc(func(data any) bool {
			return data.(Character).Class == "Mage"
		})
		updateCount(dt.DataTableDisplayCount())
	})
	addBtn("btn-warriors", "Warriors", 348, 40, 72, func() {
		dt.SetFilterFunc(func(data any) bool {
			return data.(Character).Class == "Warrior"
		})
		updateCount(dt.DataTableDisplayCount())
	})
	addBtn("btn-lv40", "Lv>40", 426, 40, 56, func() {
		dt.SetFilterFunc(func(data any) bool {
			return data.(Character).Level > 40
		})
		updateCount(dt.DataTableDisplayCount())
	})
	addBtn("btn-reset", "Reset", 520, 40, 52, func() {
		dt.ResetFiltersAndSort()
		updateCount(dt.DataTableDisplayCount())
	})

	// Initialise count display.
	updateCount(dt.DataTableDisplayCount())

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "DataTable Example",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.07, 0.08, 0.10, 1),
	})
}
