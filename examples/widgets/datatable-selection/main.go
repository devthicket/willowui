// DataTable SelectionColumn example — email inbox demo.
//
// Demonstrates the SelectionColumn convenience constructor for batch
// operations on table rows. The UI has two modes:
//
//	Normal mode:  click a row to "view" it (status label updates).
//	Batch mode:   toggle the "Batch" button to reveal checkboxes in every
//	              row. Select multiple messages, then "Archive" or "Delete"
//	              them. Switch to radio mode with the "Radio" button to
//	              restrict selection to a single row at a time.
//
// Archive toggles a flag on selected rows (demonstrating in-place data
// binding updates via SetAt). Delete removes rows entirely. The message
// count label updates reactively via Array.LenRef().
//
// Key APIs exercised:
//   - ui.SelectionColumn(key, visible, multi)
//   - ui.SetRowClickSelects(&col, true)
//   - table.SetSelectionMode / SelectedIndexes / ClearSelection
//   - Reactive Ref[bool] binding for column visibility and mode
//   - Array.SetAt for in-place item updates
//   - Array.LenRef + WatchEffect for reactive count display
package main

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 820
	screenH = 560
)

// Email represents a single inbox message.
type Email struct {
	From     string
	Subject  string
	Date     string
	Read     bool
	Archived bool
}

// inbox is the initial dataset.
var inbox = []Email{
	{"Alice", "Project update for Q2", "Mar 10", false, false},
	{"Bob", "Lunch tomorrow?", "Mar 10", true, false},
	{"Carol", "Invoice #4821 attached", "Mar 09", false, false},
	{"Dave", "Re: deployment checklist", "Mar 09", true, false},
	{"Eve", "New hire onboarding docs", "Mar 08", false, false},
	{"Frank", "Weekly standup notes", "Mar 08", true, false},
	{"Grace", "Bug report: login timeout", "Mar 07", false, false},
	{"Hank", "Vacation request approved", "Mar 07", true, false},
	{"Iris", "Design review feedback", "Mar 06", false, false},
	{"Jack", "Server migration plan", "Mar 06", false, false},
	{"Kate", "Team outing poll", "Mar 05", true, false},
	{"Leo", "Security audit results", "Mar 05", false, false},
	{"Maya", "Sprint retrospective", "Mar 04", true, false},
	{"Nathan", "API deprecation notice", "Mar 04", false, false},
	{"Olivia", "Customer feedback summary", "Mar 03", false, false},
}

func main() {
	font := ui.MustLoadDefaultFont()

	screen := ui.NewScreen()

	// -------------------------------------------------------------------
	// Title
	// -------------------------------------------------------------------
	title := willow.NewText("title", "Inbox - SelectionColumn Demo", font)
	title.TextBlock.FontSize = 18
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(20, 10)
	screen.AddNode(title)

	// -------------------------------------------------------------------
	// Status label — shows what was clicked or how many are selected
	// -------------------------------------------------------------------
	statusLabel := willow.NewText("status", "Click a row to view", font)
	statusLabel.TextBlock.FontSize = 12
	statusLabel.TextBlock.Color = willow.RGBA(0.6, 0.7, 0.9, 1)
	statusLabel.SetPosition(20, 520)
	screen.AddNode(statusLabel)

	setStatus := func(msg string) {
		statusLabel.SetContent(msg)
	}

	// -------------------------------------------------------------------
	// Selection column reactive state
	// -------------------------------------------------------------------
	// selVisible controls whether the checkbox/radio column is shown.
	// Starts hidden (normal viewing mode).
	selVisible := ui.NewRef(false)

	// selMulti controls the widget type: true = checkboxes, false = radio.
	selMulti := ui.NewRef(true)

	// -------------------------------------------------------------------
	// DataTable
	// -------------------------------------------------------------------
	dt := ui.NewDataTable("inbox", 30)
	dt.SetFont(font, 13)
	dt.SetSize(780, 460)
	dt.SetPosition(20, 50)
	dt.SetHeaderHeight(34)
	dt.SetZebraStriping(true)
	dt.SetShowColumnDividers(true)

	// Start in single-selection (view) mode. When batch mode is activated
	// the selection mode is switched to multi.
	dt.SetSelectionMode(ui.SelectionModeSingle)

	// --- Columns ---

	// Selection column: first column, checkbox or radio per row.
	selCol := ui.SelectionColumn("sel", selVisible, selMulti)
	// Enable row-click selection so clicking anywhere on the row toggles
	// the checkbox/radio, not just the widget itself.
	ui.SetRowClickSelects(&selCol, true)
	dt.AddColumn(selCol)

	// From
	fromCol := ui.LabelColumn("from", "From", func(data any) string {
		return data.(Email).From
	})
	fromCol.Weight = 1
	fromCol.Sortable = true
	fromCol.SortType = ui.SortAlpha
	dt.AddColumn(fromCol)

	// Subject
	subjCol := ui.LabelColumn("subject", "Subject", func(data any) string {
		return data.(Email).Subject
	})
	subjCol.Weight = 3
	subjCol.Sortable = true
	subjCol.SortType = ui.SortAlpha
	dt.AddColumn(subjCol)

	// Date
	dateCol := ui.LabelColumn("date", "Date", func(data any) string {
		return data.(Email).Date
	})
	dateCol.FixedWidth = 80
	dt.AddColumn(dateCol)

	// Archived flag — shows "Yes" or "" to demonstrate toggling data in-place.
	archCol := ui.LabelColumn("archived", "Archived", func(data any) string {
		if data.(Email).Archived {
			return "Yes"
		}
		return ""
	})
	archCol.FixedWidth = 70
	dt.AddColumn(archCol)

	// --- Data ---
	items := ui.NewArrayFromAny(inbox)
	dt.BindItems(items)

	// Row click in normal mode — show the selected email info.
	dt.SetOnCellClick(func(coord ui.CellCoord, data any) {
		email := data.(Email)
		setStatus(fmt.Sprintf("Viewing: %s - %q", email.From, email.Subject))
	})

	// Selection changed — update status with count.
	dt.SetOnSelectionChanged(func(indexes []int) {
		if len(indexes) == 0 {
			setStatus("No rows selected")
		} else {
			setStatus(fmt.Sprintf("%d row(s) selected", len(indexes)))
		}
	})

	screen.Add(dt)

	// -------------------------------------------------------------------
	// Toolbar buttons
	// -------------------------------------------------------------------
	btnX := 20.0
	addBtn := func(name, label string, w float64, fn func()) *ui.Button {
		btn := ui.NewButton(name, label, font, 13)
		btn.SetSize(w, 28)
		btn.SetPosition(btnX, 516)
		btn.SetOnClick(fn)
		screen.Add(btn)
		btnX += w + 6
		return btn
	}

	// Batch toggle — shows/hides the selection column and switches
	// the table between single and multi selection mode.
	batchActive := false
	var batchBtn *ui.Button
	batchBtn = addBtn("btn-batch", "Batch", 60, func() {
		batchActive = !batchActive
		selVisible.Set(batchActive)
		if batchActive {
			dt.SetSelectionMode(ui.SelectionModeMulti)
			batchBtn.SetText("Exit")
			setStatus("Batch mode: select rows then Archive or Delete")
		} else {
			dt.ClearSelection()
			dt.SetSelectionMode(ui.SelectionModeSingle)
			batchBtn.SetText("Batch")
			setStatus("Click a row to view")
		}
	})

	// Radio/Multi toggle — switches between radio (single-pick) and
	// checkbox (multi-pick) while in batch mode.
	isMulti := true
	addBtn("btn-mode", "Radio", 60, func() {
		isMulti = !isMulti
		selMulti.Set(isMulti)
		if isMulti {
			dt.SetSelectionMode(ui.SelectionModeMulti)
			setStatus("Switched to multi-select (checkboxes)")
		} else {
			dt.ClearSelection()
			dt.SetSelectionMode(ui.SelectionModeSingle)
			setStatus("Switched to single-select (radio)")
		}
	})

	// Select All
	addBtn("btn-selall", "Sel All", 60, func() {
		dt.SelectAll()
	})

	// Clear selection
	addBtn("btn-clear", "Clear", 54, func() {
		dt.ClearSelection()
		setStatus("Selection cleared")
	})

	// Archive — toggles the Archived flag on selected rows in-place.
	addBtn("btn-archive", "Archive", 64, func() {
		sel := dt.SelectedIndexes()
		if len(sel) == 0 {
			setStatus("Nothing selected to archive")
			return
		}
		toggled := 0
		for _, idx := range sel {
			email := items.At(idx).(Email)
			email.Archived = !email.Archived
			items.SetAt(idx, email)
			toggled++
		}
		dt.ClearSelection()
		setStatus(fmt.Sprintf("Toggled archive on %d message(s)", toggled))
	})

	// Delete — removes selected rows from the dataset.
	addBtn("btn-delete", "Delete", 58, func() {
		sel := dt.SelectedIndexes()
		if len(sel) == 0 {
			setStatus("Nothing selected to delete")
			return
		}
		dt.ClearSelection()
		// Remove in descending order so indexes stay valid.
		sort.Sort(sort.Reverse(sort.IntSlice(sel)))
		for _, idx := range sel {
			items.Remove(idx)
		}
		setStatus(fmt.Sprintf("Deleted %d message(s) (%d remaining)", len(sel), items.Len()))
	})

	// Row count display — reactively tracks the array length via LenRef.
	countLabel := willow.NewText("count", "", font)
	countLabel.TextBlock.FontSize = 12
	countLabel.TextBlock.Color = willow.RGBA(0.5, 0.7, 0.5, 1)
	countLabel.SetPosition(700, 520)
	screen.AddNode(countLabel)

	lenRef := items.LenRef()
	ui.WatchEffect(func() {
		countLabel.SetContent(strconv.Itoa(lenRef.Get()) + " messages")
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "DataTable SelectionColumn",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.07, 0.08, 0.10, 1),
	})
}
