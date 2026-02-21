package widget

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
	"github.com/devthicket/willowui/internal/theme"
)

// ---------------------------------------------------------------------------
// Enum types
// ---------------------------------------------------------------------------

// SortType controls how a column's values are compared during sort.
type SortType int

const (
	SortAlpha   SortType = iota // lexicographic (string) comparison
	SortNumeric                 // numeric comparison
	SortCustom                  // use the column's Comparator function
)

// CellClipMode controls how cell content is clipped when it overflows.
type CellClipMode int

const (
	ClipEllipsis CellClipMode = iota // truncate with "..."
	ClipMask                         // use a mask to clip content
)

// ScrollMode controls how the DataTable scrolls.
type ScrollMode int

const (
	ScrollModeVirtual ScrollMode = iota // virtualized: only visible rows are rendered
	ScrollModeStatic                    // static pool: all slots always present
)

// SelectionMode controls row selection behavior.
type SelectionMode int

const (
	SelectionModeNone   SelectionMode = iota // no selection
	SelectionModeSingle                      // single row selection
	SelectionModeMulti                       // multi-row selection
)

// SortDirection indicates the sort order for a column.
type SortDirection int

const (
	SortNone SortDirection = iota // no sort
	SortAsc                       // ascending
	SortDesc                      // descending
)

// OnSortScroll controls scroll behavior after a sort operation.
type OnSortScroll int

const (
	OnSortScrollNone        OnSortScroll = iota // do not scroll
	OnSortScrollToSelection                     // scroll to current selection
	OnSortScrollToTop                           // scroll to top
)

// CellCoord identifies a cell by row and column.
type CellCoord struct {
	Row, Col int
}

// scrollWheelSpeedDataTable is the pixel distance per wheel tick.
const scrollWheelSpeedDataTable = 40

// ---------------------------------------------------------------------------
// DataTableColumn
// ---------------------------------------------------------------------------

// CellStyle holds styling overrides for DataTable cells and headers.
// Zero values mean "use default" (theme color, table font/size, left-aligned).
type CellStyle struct {
	Align        sg.TextAlign                    // text alignment within the cell
	Color        sg.Color                        // text color override (zero = theme default)
	FontSize     float64                         // font size override (0 = table font size)
	Sharpness    float64                         // SDF sharpness override (0 = default)
	OnPostUpdate func(data any, comp *Component) // called after UpdateCell with the row data and cell component
}

// LabelStyle is a deprecated alias for CellStyle.
type LabelStyle = CellStyle

// SortKey identifies a column and its sort direction in a multi-sort stack.
type SortKey struct {
	ColKey string
	Dir    SortDirection
}

// DataTableColumn defines a column in a DataTable.
type DataTableColumn struct {
	Key         string
	Header      string
	Tooltip     string
	Weight      float64 // flex weight (0 means use FixedWidth)
	FixedWidth  float64 // fixed pixel width (0 means use Weight)
	MinWidth    float64 // minimum pixel width for flex columns
	MaxWidth    float64 // maximum pixel width for flex columns
	Sortable    bool
	Filterable  bool // enables per-column filter UI
	SortType    SortType
	SortValue   func(data any) any // override sort key; returns string or number
	Comparator  func(a, b any) int // used when SortType == SortCustom
	Searchable  bool
	SearchValue func(data any) string // used by BindSearchFilter; Searchable must also be true
	RenderCell  func(rowIndex int, data any) *Component
	UpdateCell  func(rowIndex int, data any, comp *Component)
	Cell        CellStyle // styling and post-update hook for data cells
	HeaderStyle CellStyle // styling overrides for the header text
	ClipMode    CellClipMode

	// Hidden returns true when this column should collapse to zero width and
	// not render any content. The function is called during layout resolution,
	// so the result can change dynamically (e.g. bound to a reactive Ref).
	// A nil Hidden func means the column is always visible.
	Hidden func() bool

	// internal holds private state for special column types (e.g. SelectionColumn).
	// Not intended for external use.
	internal any
}

// LabelColumn creates a simple text-label column with the given key, header and accessor.
func LabelColumn(key, header string, accessor func(data any) string) DataTableColumn {
	return DataTableColumn{
		Key:         key,
		Header:      header,
		Weight:      1,
		Searchable:  true,
		SearchValue: accessor,
		UpdateCell: func(rowIndex int, data any, comp *Component) {
			text := accessor(data)
			if l, ok := comp.UserData().(*Label); ok {
				l.SetText(text)
			} else {
				label := NewLabel(key+"-cell", text, nil, 0)
				comp.node.AddChild(label.Node())
				comp.SetUserData(label)
			}
		},
		RenderCell: nil, // UpdateCell handles both init and update
	}
}

// defaultMeterCellHeight is the height of the meter bar inside a table cell.
const defaultMeterCellHeight = 14

// MeterColumn creates a column that renders an inline MeterBar for each row.
// The accessor returns a float64 in [0, 1]. Use Cell.OnPostUpdate to
// customize the fill color dynamically.
func MeterColumn(key, header string, accessor func(data any) float64) DataTableColumn {
	return DataTableColumn{
		Key:        key,
		Header:     header,
		Weight:     1,
		Sortable:   true,
		SortType:   SortNumeric,
		SortValue:  func(data any) any { return accessor(data) },
		Searchable: false,
		UpdateCell: func(rowIndex int, data any, comp *Component) {
			v := accessor(data)
			if mb, ok := comp.UserData().(*MeterBar); ok {
				// Resize to match current cell width.
				if comp.Width > 0 && mb.Width != comp.Width {
					mb.SetSize(comp.Width, defaultMeterCellHeight)
				}
				mb.SetValue(v)
			} else {
				mb := NewMeterBar(key + "-cell-meter")
				w := comp.Width
				if w <= 0 {
					w = 100
				}
				mb.SetSize(w, defaultMeterCellHeight)
				mb.SetValue(v)
				comp.node.AddChild(mb.Node())
				comp.Height = defaultMeterCellHeight
				comp.SetUserData(mb)
			}
		},
	}
}

// ---------------------------------------------------------------------------
// SelectionColumn — checkbox (multi) or radio (single) selection per row
// ---------------------------------------------------------------------------

// selectionColumnState holds the shared mutable state for a SelectionColumn.
// A pointer to this struct is stored in DataTableColumn.internal so that the
// DataTable can wire itself up when SetColumns is called.
type selectionColumnState struct {
	// visible controls whether the column is shown. When the Ref value is
	// false the column collapses to zero width and cells are hidden.
	visible *Ref[bool]

	// multi controls the selection widget type:
	//   true  -> checkboxes (multi-select: toggle individual rows)
	//   false -> radio dots  (single-select: only one row at a time)
	multi *Ref[bool]

	// rowClickSelects mirrors DataTable.rowClickSelects but is scoped to
	// the selection column. When true, clicking anywhere on the row
	// triggers the same toggle/select as clicking the checkbox/radio.
	rowClickSelects bool

	// table is set by DataTable.SetColumns so that the column's UpdateCell
	// closure can call into the table's selection API.
	table *DataTable

	// prevMulti tracks the last-observed value of multi so that cells can
	// detect a mode switch and recreate their widget.
	prevMulti bool
}

// defaultSelectionColWidth is the fixed pixel width for the selection column.
// Sized to comfortably fit a checkbox or radio dot with padding.
const defaultSelectionColWidth = 36

// SelectionColumn creates a column that renders a selection indicator
// (checkbox or radio button) for each row. The column is designed for
// "batch mode" UX patterns — hide it in normal viewing mode and show it
// when the user activates batch operations.
//
// Parameters:
//
//   - key: unique column key (used internally and for identification).
//
//   - visible: reactive bool that controls column visibility. When false
//     the column collapses to zero width and its header/cells are hidden.
//     Bind this to a button or toggle to let users enter/exit batch mode.
//
//   - multi: reactive bool that controls the selection widget type.
//     true renders checkboxes (multi-select), false renders radio dots
//     (single-select). Switching at runtime swaps the widget in every
//     visible cell.
//
// The column is non-sortable, non-searchable, and uses a fixed width of
// 36px. Selection state is owned by the DataTable — this column is purely
// the visual companion to the table's SelectionMode / SelectedIndexes API.
//
// The DataTable's SelectionMode must match the multi flag for consistent
// behavior: SelectionModeMulti when multi is true, SelectionModeSingle
// when false. The column does NOT change SelectionMode automatically.
//
// Example:
//
//	selVisible := ui.NewRef(false)
//	selMulti   := ui.NewRef(true)
//
//	cols := []ui.DataTableColumn{
//	    ui.SelectionColumn("sel", selVisible, selMulti),
//	    ui.LabelColumn("name", "Name", func(d any) string { ... }),
//	}
//
//	table := ui.NewDataTable("tbl", 30)
//	table.SetColumns(cols)
//	table.SetSelectionMode(ui.SelectionModeMulti)
//
//	// Toggle batch mode from a button:
//	selVisible.Set(true)
func SelectionColumn(key string, visible, multi *Ref[bool]) DataTableColumn {
	state := &selectionColumnState{
		visible:   visible,
		multi:     multi,
		prevMulti: multi.Peek(),
	}

	return DataTableColumn{
		Key:        key,
		Header:     "",
		FixedWidth: defaultSelectionColWidth,
		Sortable:   false,
		Searchable: false,

		// The column collapses to zero width when visible is false.
		Hidden: func() bool {
			return !state.visible.Peek()
		},

		UpdateCell: func(rowIndex int, data any, comp *Component) {
			dt := state.table
			if dt == nil {
				return
			}

			isMulti := state.multi.Peek()
			modeChanged := isMulti != state.prevMulti

			// selCellState caches the per-cell widget so we can reuse it
			// across scroll recycling without recreating nodes every frame.
			type selCellState struct {
				checkbox *Checkbox // used in multi-select mode
				radio    *Radio    // single-option group for single-select mode
				isMulti  bool      // which widget type is currently active
			}

			var cellState *selCellState
			if ud, ok := comp.UserData().(*selCellState); ok && !modeChanged {
				cellState = ud
			} else {
				// First init or mode changed — clear old children and create
				// the appropriate widget.
				for comp.node.NumChildren() > 0 {
					comp.node.RemoveChildAt(0)
				}

				// Switching between check/radio resets the selection state so
				// stale selections from the previous mode don't carry over.
				if modeChanged {
					dt.ClearSelection()
				}

				cellState = &selCellState{isMulti: isMulti}

				// Build a selection-column theme that overrides Checkbox/Radio
				// colors to ensure visibility against the DataTable row background.
				selTheme := selectionColumnTheme(dt)

				if isMulti {
					cb := NewCheckbox(key+"-sel-cb", "", nil, 0)
					cb.SetTheme(selTheme)
					cellState.checkbox = cb
					// Set comp.Height for vertical centering in the row.
					comp.Height = DefaultCheckboxSize
					// Purely visual — pointer events pass through to the
					// row so hovering the checkbox does not break row hover.
					cb.Node().Interactable = false
					comp.node.AddChild(cb.Node())
				} else {
					rg := NewRadio(key + "-sel-rg")
					rg.SetTheme(selTheme)
					rg.AddOption("", nil, 0)
					cellState.radio = rg
					// Set comp.Height for vertical centering in the row.
					comp.Height = DefaultRadioSize
					// Purely visual — same as checkbox above.
					rg.Node().Interactable = false
					comp.node.AddChild(rg.Node())
				}

				comp.SetUserData(cellState)
				state.prevMulti = isMulti
			}

			// Update the widget to reflect the current selection state.
			selected := dt.isSelected(rowIndex)

			if cellState.isMulti && cellState.checkbox != nil {
				cellState.checkbox.SetChecked(selected)
				cellState.checkbox.SetOnChange(func(checked bool) {
					if checked {
						dt.SelectRow(rowIndex)
					} else {
						dt.DeselectRow(rowIndex)
					}
				})
			} else if cellState.radio != nil {
				if selected {
					cellState.radio.SetSelected(0)
				} else {
					cellState.radio.SetSelected(-1)
				}
				cellState.radio.SetOnChange(func(_ int) {
					// In single-select mode, selecting a new row must first
					// clear the previous selection so only one row is active.
					dt.ClearSelection()
					dt.SelectRow(rowIndex)
				})
			}
		},

		// Store the state pointer so SetColumns can wire up the table ref.
		internal: state,
	}
}

// SetRowClickSelects configures whether clicking anywhere on a row triggers
// the selection toggle for a SelectionColumn. Call this on the
// DataTableColumn returned by SelectionColumn.
//
// By default row-click selection is off — only clicking the checkbox/radio
// toggles selection. Enable it for an email-inbox-style UX where clicking
// the entire row selects it.
func SetRowClickSelects(col *DataTableColumn, enable bool) {
	if s, ok := col.internal.(*selectionColumnState); ok {
		s.rowClickSelects = enable
	}
}

// hasSelectionColumn returns true if any column is a SelectionColumn.
func (dt *DataTable) hasSelectionColumn() bool {
	for _, col := range dt.columns {
		if _, ok := col.internal.(*selectionColumnState); ok {
			return true
		}
	}
	return false
}

// selectionColumnWantsRowClick returns true if any SelectionColumn in the table
// has RowClickSelects enabled and is currently visible.
func (dt *DataTable) selectionColumnWantsRowClick() bool {
	for _, col := range dt.columns {
		if s, ok := col.internal.(*selectionColumnState); ok {
			if s.rowClickSelects && s.visible.Peek() {
				return true
			}
		}
	}
	return false
}

// clickInSelectionColumn returns true if the given local X coordinate falls
// within any visible selection column's cell bounds. Used so that clicks on
// non-interactable checkbox/radio widgets still toggle selection.
func (dt *DataTable) clickInSelectionColumn(localX float64) bool {
	for i, col := range dt.columns {
		if _, ok := col.internal.(*selectionColumnState); !ok {
			continue
		}
		s := col.internal.(*selectionColumnState)
		if !s.visible.Peek() {
			continue
		}
		if i >= len(dt.colOffsets) || i >= len(dt.colWidths) {
			continue
		}
		x0 := dt.colOffsets[i]
		x1 := x0 + dt.colWidths[i]
		if localX >= x0 && localX < x1 {
			return true
		}
	}
	return false
}

// selectionColumnTheme builds a theme override for SelectionColumn widgets.
// It copies the effective theme and replaces the Checkbox.BoxColor and
// Radio.CircleColor default state with the DataTable's CellText color,
// ensuring the indicators are visible against row backgrounds.
func selectionColumnTheme(dt *DataTable) *Theme {
	base := dt.EffectiveTheme()
	// Shallow copy — most fields are value types (arrays/structs) so this
	// produces an independent theme we can safely mutate.
	t := *base

	dtGroup := base.DataTable.Group(dt.Variant())
	cellText := dtGroup.CellText.Resolve(StateDefault)

	// Checkbox: use cell text for the unchecked border so it stands out.
	t.Checkbox.Primary.BoxColor = NewColorPropStates(map[ComponentState]sg.Color{
		StateDefault:     cellText,
		StateActive:      base.Checkbox.Primary.BoxColor.Resolve(StateActive),
		StateFocus:       cellText,
		StateFocusActive: base.Checkbox.Primary.BoxColor.Resolve(StateFocusActive),
		StateDisabled:    base.Checkbox.Primary.BoxColor.Resolve(StateDisabled),
	})

	// Radio: use cell text for the unselected circle border.
	t.Radio.Primary.CircleColor = NewColorPropStates(map[ComponentState]sg.Color{
		StateDefault:     cellText,
		StateActive:      base.Radio.Primary.CircleColor.Resolve(StateActive),
		StateFocus:       cellText,
		StateFocusActive: base.Radio.Primary.CircleColor.Resolve(StateFocusActive),
	})

	return &t
}

// compAsLabelComp retrieves a *Label from a *Component via UserData.
func compAsLabelComp(c *Component) *Label {
	if c == nil {
		return nil
	}
	if ud := c.UserData(); ud != nil {
		if l, ok := ud.(*Label); ok {
			return l
		}
	}
	return nil
}

// EllipsisLabel creates a Label for use in cell rendering. The label uses nil font
// which falls back to the default font. Returns both the label and its component.
func EllipsisLabel(name, text string) *Label {
	l := NewLabel(name, text, nil, 0)
	l.SetUserData(l)
	return l
}

// UpdateEllipsisLabel updates the text of an EllipsisLabel component.
func UpdateEllipsisLabel(comp *Component, text string) {
	if l := compAsLabelComp(comp); l != nil {
		l.SetText(text)
	}
}

// ---------------------------------------------------------------------------
// dtRowSlot: a recycled row slot
// ---------------------------------------------------------------------------

type dtRowSlot struct {
	node       *sg.Node
	bgNode     *sg.Node
	cells      []*Component
	dividers   []*sg.Node // vertical column dividers
	rowDivider *sg.Node   // horizontal row bottom border
	displayPos int        // which display index this slot shows (-1 = empty)
	hovered    bool
}

// ---------------------------------------------------------------------------
// DataTable
// ---------------------------------------------------------------------------

// DataTable is a virtualized, sortable, filterable data grid widget.
type DataTable struct {
	Component

	// columns
	columns    []DataTableColumn
	colWidths  []float64 // resolved pixel widths
	colOffsets []float64 // precomputed x offsets per column

	// data
	items     []any
	stopArray func()

	// display order (after sort/filter)
	displayIndexes []int

	// scroll
	scrollPos      *Ref[float64]
	scrollPosRef   *Ref[float64] // externally bound ref
	scrollPosWatch WatchHandle
	scrollBar      *ScrollBar
	scrollMode     ScrollMode
	showScrollBar  bool

	// selection
	selectionMode   SelectionMode
	selectedIndexes []int
	anchorRow       int
	rowClickSelects bool
	selIdxGet       func() []int
	selIdxSet       func([]int)

	// sort
	sortKeys        []SortKey
	defaultSortKeys []SortKey
	onSort          func(key string, dir SortDirection)
	onMultiSort     func([]SortKey)
	onSortScroll    OnSortScroll

	// filter
	filterFunc    func(data any) bool
	columnFilters map[string]map[string]bool // colKey -> allowed values
	searchRef     *Ref[string]
	searchWatch   WatchHandle

	// appearance
	rowHeight          float64
	headerHeight       float64
	showHeader         bool
	zebraStriping      bool
	showColumnDividers bool
	showRowDividers    bool
	font               *sg.FontFamily
	source             *sg.FontFamily
	fontSize           float64

	// nodes
	headerContainer *sg.Node
	rowViewport     *sg.Node
	rowContent      *sg.Node // positioned content container inside viewport
	maskRoot        *sg.Node // reused mask container
	maskSprite      *sg.Node // reused mask sprite
	emptyComp       *Component

	// header cell hover state
	headerHovered []bool

	// row pool (slots)
	rowSlots    []*dtRowSlot
	topRowIndex int // for static mode

	// callbacks
	onCellClick        func(coord CellCoord, data any)
	onCellDoubleClick  func(coord CellCoord, data any)
	onSelectionChanged func(indexes []int)

	// dirty flags
	layoutDirty     bool
	dataDirty       bool
	sortFilterDirty bool
	selectionDirty  bool
	scrollDirty     bool

	// selColWatchers tracks reactive watchers set up for SelectionColumn refs
	// so they can be stopped when columns are replaced.
	selColWatchers []WatchHandle
}

// NewDataTable creates a DataTable with the given name and row height.
func NewDataTable(name string, rowHeight float64) *DataTable {
	dt := &DataTable{
		rowHeight:       rowHeight,
		headerHeight:    32,
		showHeader:      true,
		showScrollBar:   true,
		rowClickSelects: true,
		selectionMode:   SelectionModeSingle,
		anchorRow:       -1,
		scrollPos:       NewRef(0.0),
	}
	initComponent(&dt.Component, name)
	dt.initBackground(name)
	dt.initBorder(name)

	// Apply background from theme.
	group := dt.EffectiveTheme().DataTable.Group(dt.Variant())
	bg := group.Background.Resolve(StateDefault)
	dt.applyBackground(bg)

	// Header container.
	dt.headerContainer = sg.NewContainer(name + "-header")
	dt.headerContainer.Interactable = true
	dt.node.AddChild(dt.headerContainer)

	// Viewport for rows.
	dt.rowViewport = sg.NewContainer(name + "-viewport")
	dt.rowViewport.Interactable = true
	dt.node.AddChild(dt.rowViewport)

	// Content inside viewport (scrolled).
	dt.rowContent = sg.NewContainer(name + "-content")
	dt.rowContent.Interactable = true
	dt.rowViewport.AddChild(dt.rowContent)

	// Reusable mask (resized in layoutInternals).
	dt.maskRoot = sg.NewContainer(name + "-mask")
	dt.maskSprite = sg.NewSprite(name+"-mask-rect", sg.TextureRegion{})
	dt.maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	dt.maskRoot.AddChild(dt.maskSprite)
	dt.rowViewport.SetMask(dt.maskRoot)

	// ScrollBar.
	dt.scrollBar = NewScrollBar(name + "-scrollbar")
	dt.scrollBar.SetOnChange(func(pos float64) {
		dt.scrollPos.Set(pos)
		DefaultScheduler.Flush()
		dt.updateRows()
	})
	dt.scrollBar.AddToNode(dt.node)
	dt.scrollBar.parent = &dt.Component

	// Interactable header so hover events fire.
	dt.headerContainer.OnPointerEnter(func(_ sg.PointerContext) {})
	dt.headerContainer.OnPointerLeave(func(_ sg.PointerContext) {})

	// Auto-update hook.
	dt.node.OnUpdate = func(_ float64) {
		dt.Update()
	}

	// Click on the table area to potentially focus.
	dt.node.OnPointerDown(func(_ sg.PointerContext) {
		DefaultFocusManager.SetFocus(&dt.Component)
	})

	dt.onThemeChange = func() { dt.applyThemeColors() }

	dt.SetSize(400, 300)
	return dt
}

// ---------------------------------------------------------------------------
// Configuration setters
// ---------------------------------------------------------------------------

// AddColumn appends a column definition.
func (dt *DataTable) AddColumn(col DataTableColumn) {
	dt.wireSelectionColumn(&col)
	dt.columns = append(dt.columns, col)
	dt.headerHovered = append(dt.headerHovered, false)
	dt.layoutDirty = true
	dt.dataDirty = true
}

// SetColumns replaces all columns.
func (dt *DataTable) SetColumns(cols []DataTableColumn) {
	// Stop any existing selection column watchers from the previous column set.
	for _, w := range dt.selColWatchers {
		w.Stop()
	}
	dt.selColWatchers = nil

	dt.columns = append([]DataTableColumn(nil), cols...)
	dt.headerHovered = make([]bool, len(cols))
	dt.layoutDirty = true
	dt.dataDirty = true

	for i := range dt.columns {
		dt.wireSelectionColumn(&dt.columns[i])
	}
}

// wireSelectionColumn detects a SelectionColumn, wires the table back-reference,
// and sets up reactive watchers on its visible/multi refs so that changes
// trigger a layout rebuild.
func (dt *DataTable) wireSelectionColumn(col *DataTableColumn) {
	s, ok := col.internal.(*selectionColumnState)
	if !ok {
		return
	}
	s.table = dt

	// Watch the visible ref — when it changes, the Hidden func return value
	// changes, so we need a full layout + slot rebuild.
	w1 := WatchValue(s.visible, func(_, _ bool) {
		dt.layoutDirty = true
		dt.dataDirty = true
	})
	dt.selColWatchers = append(dt.selColWatchers, w1)

	// Watch the multi ref — when it changes, cells need to swap between
	// checkbox and radio widgets, so mark data dirty to repopulate cells.
	w2 := WatchValue(s.multi, func(_, _ bool) {
		dt.dataDirty = true
	})
	dt.selColWatchers = append(dt.selColWatchers, w2)
}

// OnColumn merges render/update funcs into an existing column identified by key.
// Used to attach RenderCell/UpdateCell factories after loading columns from XML.
func (dt *DataTable) OnColumn(key string, col DataTableColumn) {
	for i, c := range dt.columns {
		if c.Key == key {
			col.Key = c.Key
			if col.Header == "" {
				col.Header = c.Header
			}
			dt.columns[i] = col
			return
		}
	}
}

// SetItems replaces the data slice and refreshes the display.
func (dt *DataTable) SetItems(items []any) {
	if dt.stopArray != nil {
		dt.stopArray()
		dt.stopArray = nil
	}
	dt.items = append([]any(nil), items...)
	dt.scrollPos.Set(0)
	DefaultScheduler.Flush()
	// Ensure layout is resolved before building slots.
	if dt.layoutDirty || len(dt.colWidths) != len(dt.columns) {
		dt.layoutDirty = false
		dt.layoutInternals()
	}
	if len(dt.sortKeys) == 0 && len(dt.defaultSortKeys) > 0 {
		dt.sortKeys = append([]SortKey(nil), dt.defaultSortKeys...)
	}
	dt.rebuildDisplayIndexes()
	dt.rebuildSlots()
	dt.updateScrollBar()
	dt.updateRows()
}

// BindItems binds the DataTable to a reactive Array[any]. Mutations to the array
// are reflected incrementally without full re-renders where possible.
func (dt *DataTable) BindItems(items *Array[any]) {
	if dt.stopArray != nil {
		dt.stopArray()
		dt.stopArray = nil
	}
	if items == nil {
		dt.items = nil
		dt.rebuildDisplayIndexes()
		dt.updateScrollBar()
		dt.updateRows()
		return
	}

	dt.items = dt.items[:0]
	items.ForEach(func(_ int, item any) { dt.items = append(dt.items, item) })
	if len(dt.sortKeys) == 0 && len(dt.defaultSortKeys) > 0 {
		dt.sortKeys = append([]SortKey(nil), dt.defaultSortKeys...)
	}
	dt.rebuildDisplayIndexes()
	dt.rebuildSlots()
	dt.updateScrollBar()
	dt.updateRows()

	h1 := items.OnAdded(func(idx int, item any) {
		dt.items = append(dt.items, nil)
		copy(dt.items[idx+1:], dt.items[idx:])
		dt.items[idx] = item
		dt.rebuildDisplayIndexes()
		dt.updateScrollBar()
		dt.updateRows()
	})
	h2 := items.OnRemoved(func(idx int, _ any) {
		dt.items = append(dt.items[:idx], dt.items[idx+1:]...)
		dt.rebuildDisplayIndexes()
		dt.updateScrollBar()
		dt.updateRows()
	})
	h3 := items.OnItemChanged(func(idx int, item any) {
		if idx >= 0 && idx < len(dt.items) {
			dt.items[idx] = item
		}
		dt.updateAffectedSlots(idx)
	})
	syncAll := func() {
		dt.items = dt.items[:0]
		items.ForEach(func(_ int, item any) { dt.items = append(dt.items, item) })
		dt.rebuildDisplayIndexes()
		dt.updateScrollBar()
		dt.updateRows()
	}
	h4 := items.OnReplaced(syncAll)
	h5 := items.OnMoved(func(_, _ int) { syncAll() })
	dt.stopArray = func() { h1.Stop(); h2.Stop(); h3.Stop(); h4.Stop(); h5.Stop() }
}

// updateAffectedSlots refreshes only the visible slots that display the given data index.
func (dt *DataTable) updateAffectedSlots(dataIdx int) {
	group := dt.dtTheme()
	for _, slot := range dt.rowSlots {
		if slot.displayPos < 0 || slot.displayPos >= len(dt.displayIndexes) {
			continue
		}
		if dt.displayIndexes[slot.displayPos] != dataIdx {
			continue
		}
		data := dt.items[dataIdx]
		for ci, col := range dt.columns {
			cell := slot.cells[ci]
			if cell == nil {
				continue
			}
			if col.UpdateCell != nil {
				col.UpdateCell(dataIdx, data, cell)
			}
		}
		dt.applyRowBg(slot, slot.displayPos, dataIdx, group)
	}
}

// SetRowHeight sets the height of each data row.
func (dt *DataTable) SetRowHeight(h float64) {
	dt.rowHeight = h
	dt.layoutDirty = true
}

// SetHeaderHeight sets the height of the header row.
func (dt *DataTable) SetHeaderHeight(h float64) {
	dt.headerHeight = h
	dt.layoutDirty = true
}

// SetShowHeader controls whether the header row is visible.
func (dt *DataTable) SetShowHeader(v bool) {
	dt.showHeader = v
	dt.layoutDirty = true
}

// SetShowScrollBar controls whether the scrollbar is shown.
func (dt *DataTable) SetShowScrollBar(v bool) {
	dt.showScrollBar = v
	dt.layoutDirty = true
}

// SetZebraStriping enables or disables alternating row backgrounds.
func (dt *DataTable) SetZebraStriping(v bool) {
	dt.zebraStriping = v
	dt.selectionDirty = true
}

// SetShowColumnDividers shows/hides vertical dividers between columns.
func (dt *DataTable) SetShowColumnDividers(v bool) {
	dt.showColumnDividers = v
	dt.dataDirty = true
}

// SetShowRowDividers shows/hides horizontal dividers between rows.
func (dt *DataTable) SetShowRowDividers(v bool) {
	dt.showRowDividers = v
	dt.dataDirty = true
}

// SetScrollMode switches between virtual and static scroll modes.
func (dt *DataTable) SetScrollMode(m ScrollMode) {
	dt.scrollMode = m
	dt.dataDirty = true
}

// SetFont sets the font source and display size used for header text and default cell labels.
// Must be called before SetItems/BindItems so cells pick up the font.
func (dt *DataTable) SetFont(source *sg.FontFamily, size float64) {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	dt.font = font
	dt.source = source
	dt.fontSize = size
	dt.layoutDirty = true
}

// SetSelectionMode sets the selection behavior.
func (dt *DataTable) SetSelectionMode(m SelectionMode) {
	dt.selectionMode = m
	dt.selectedIndexes = nil
	dt.selectionDirty = true
}

// SetRowClickSelects enables/disables selection on row click.
func (dt *DataTable) SetRowClickSelects(v bool) {
	dt.rowClickSelects = v
}

// SetOnSortScroll sets the scroll behavior after sorting.
func (dt *DataTable) SetOnSortScroll(v OnSortScroll) {
	dt.onSortScroll = v
}

// SetFilterFunc sets a function to filter visible rows.
// Pass nil to clear the filter; scroll position is reset to top.
func (dt *DataTable) SetFilterFunc(fn func(data any) bool) {
	dt.filterFunc = fn
	dt.scrollPos.Set(0)
	DefaultScheduler.Flush()
	dt.rebuildDisplayIndexes()
	dt.updateScrollBar()
	dt.updateRows()
}

// BindSearchFilter binds a reactive string ref to the search filter.
// Columns with Searchable=true and SearchValue set are filtered when the ref changes.
func (dt *DataTable) BindSearchFilter(ref *Ref[string]) {
	dt.searchWatch.Stop()
	dt.searchRef = ref
	dt.sortFilterDirty = true
	dt.searchWatch = WatchValue(ref, func(_, _ string) {
		dt.sortFilterDirty = true
	})
}

// BindSearchInput binds a TextInput's reactive value to the search filter.
func (dt *DataTable) BindSearchInput(ti *TextInput) {
	dt.BindSearchFilter(ti.ValueRef())
}

// SetColumnFilter sets the allowed values for a per-column filter.
// Only rows whose column SearchValue is in values will be shown.
func (dt *DataTable) SetColumnFilter(colKey string, values []string) {
	if dt.columnFilters == nil {
		dt.columnFilters = make(map[string]map[string]bool)
	}
	allowed := make(map[string]bool, len(values))
	for _, v := range values {
		allowed[v] = true
	}
	dt.columnFilters[colKey] = allowed
	dt.sortFilterDirty = true
}

// ClearColumnFilter removes the per-column filter for the given column.
func (dt *DataTable) ClearColumnFilter(colKey string) {
	delete(dt.columnFilters, colKey)
	dt.sortFilterDirty = true
}

// ResetFiltersAndSort clears all sort keys, filters, search, and column filters,
// scrolls to top, and rebuilds.
func (dt *DataTable) ResetFiltersAndSort() {
	dt.sortKeys = nil
	dt.filterFunc = nil
	dt.columnFilters = nil
	if dt.searchRef != nil {
		dt.searchRef.Set("")
		DefaultScheduler.Flush()
	}
	dt.scrollPos.Set(0)
	DefaultScheduler.Flush()
	dt.rebuildDisplayIndexes()
	dt.buildHeader()
	dt.updateScrollBar()
	dt.updateRows()
}

// SetOnCellClick sets the callback for cell click events.
func (dt *DataTable) SetOnCellClick(fn func(coord CellCoord, data any)) {
	dt.onCellClick = fn
}

// SetOnCellDoubleClick sets the callback for cell double-click events.
func (dt *DataTable) SetOnCellDoubleClick(fn func(coord CellCoord, data any)) {
	dt.onCellDoubleClick = fn
}

// SetOnSelectionChanged sets the callback for selection changes.
func (dt *DataTable) SetOnSelectionChanged(fn func(indexes []int)) {
	dt.onSelectionChanged = fn
}

// SetOnSort sets the callback invoked when the sort column/direction changes.
// When registered, the DataTable does NOT apply sorting internally — the caller owns sort logic.
func (dt *DataTable) SetOnSort(fn func(key string, dir SortDirection)) {
	dt.onSort = fn
}

// SelectedIndexes returns the current selected data indexes (into items slice).
func (dt *DataTable) SelectedIndexes() []int {
	return append([]int(nil), dt.selectedIndexes...)
}

// BindSelectedIndexes binds selection state to external getter/setter functions.
// When the DataTable selection changes, setFn is called with the new indexes.
// When external code changes the selection, call SetSelectedIndexes to push it in.
func (dt *DataTable) BindSelectedIndexes(getFn func() []int, setFn func([]int)) {
	dt.selIdxGet = getFn
	dt.selIdxSet = setFn
	if getFn != nil {
		dt.selectedIndexes = append([]int(nil), getFn()...)
		dt.selectionDirty = true
	}
}

// SetSelectedIndexes replaces the selection from external code and refreshes visuals.
func (dt *DataTable) SetSelectedIndexes(indexes []int) {
	dt.selectedIndexes = append([]int(nil), indexes...)
	dt.selectionDirty = true
	dt.syncSelIdx()
}

// syncSelIdx pushes the current selectedIndexes to the bound setter, if any.
func (dt *DataTable) syncSelIdx() {
	if dt.selIdxSet != nil {
		dt.selIdxSet(append([]int(nil), dt.selectedIndexes...))
	}
}

// ClearSelection deselects all rows.
func (dt *DataTable) ClearSelection() {
	dt.selectedIndexes = nil
	dt.selectionDirty = true
	dt.syncSelIdx()
}

// SelectRow selects the row at the given data index.
func (dt *DataTable) SelectRow(index int) {
	if dt.selectionMode == SelectionModeNone {
		return
	}
	if dt.selectionMode == SelectionModeSingle {
		dt.selectedIndexes = []int{index}
	} else {
		if !dt.isSelected(index) {
			dt.selectedIndexes = append(dt.selectedIndexes, index)
		}
	}
	dt.anchorRow = index
	dt.selectionDirty = true
	dt.syncSelIdx()
	if dt.onSelectionChanged != nil {
		dt.onSelectionChanged(dt.selectedIndexes)
	}
}

// DeselectRow removes the given data index from the selection.
func (dt *DataTable) DeselectRow(index int) {
	newSel := dt.selectedIndexes[:0]
	for _, s := range dt.selectedIndexes {
		if s != index {
			newSel = append(newSel, s)
		}
	}
	dt.selectedIndexes = newSel
	dt.selectionDirty = true
	dt.syncSelIdx()
	if dt.onSelectionChanged != nil {
		dt.onSelectionChanged(dt.selectedIndexes)
	}
}

// ToggleRowSelection toggles selection of the given data index.
func (dt *DataTable) ToggleRowSelection(index int) {
	if dt.isSelected(index) {
		dt.DeselectRow(index)
	} else {
		dt.SelectRow(index)
	}
}

// SelectAll selects all rows (multi-selection mode only; no-op otherwise).
func (dt *DataTable) SelectAll() {
	if dt.selectionMode != SelectionModeMulti {
		return
	}
	dt.selectedIndexes = make([]int, len(dt.items))
	for i := range dt.items {
		dt.selectedIndexes[i] = i
	}
	dt.selectionDirty = true
	dt.syncSelIdx()
	if dt.onSelectionChanged != nil {
		dt.onSelectionChanged(dt.selectedIndexes)
	}
}

// IsSelected returns whether the given data index is selected.
func (dt *DataTable) IsSelected(index int) bool {
	return dt.isSelected(index)
}

// ---------------------------------------------------------------------------
// Scroll API
// ---------------------------------------------------------------------------

// ScrollToRow scrolls so the given original-array index is visible.
func (dt *DataTable) ScrollToRow(index int) {
	for i, di := range dt.displayIndexes {
		if di == index {
			targetY := float64(i) * dt.rowHeight
			dt.scrollBar.SetScrollPos(targetY)
			return
		}
	}
}

// ScrollToTop scrolls to the top of the table.
func (dt *DataTable) ScrollToTop() {
	dt.scrollBar.SetScrollPos(0)
}

// ScrollToBottom scrolls to the bottom of the table.
func (dt *DataTable) ScrollToBottom() {
	totalH := float64(len(dt.displayIndexes)) * dt.rowHeight
	dt.scrollBar.SetScrollPos(totalH)
}

// BindScrollPos binds an external Ref[float64] to the scroll position.
func (dt *DataTable) BindScrollPos(ref *Ref[float64]) {
	dt.scrollPosWatch.Stop()
	dt.scrollPosRef = ref
	if ref != nil {
		dt.scrollPos.Set(ref.Peek())
		dt.updateRows()
		dt.scrollPosWatch = WatchValue(ref, func(_, newPos float64) {
			dt.scrollBar.SetScrollPos(newPos)
		})
	}
}

// ---------------------------------------------------------------------------
// Sort setters
// ---------------------------------------------------------------------------

// SetSortedColumn programmatically sets a single sort column and direction.
// Clears any multi-column sort.
func (dt *DataTable) SetSortedColumn(key string, dir SortDirection) {
	if dir == SortNone || key == "" {
		dt.sortKeys = nil
	} else {
		dt.sortKeys = []SortKey{{ColKey: key, Dir: dir}}
	}
	dt.rebuildDisplayIndexes()
	dt.buildHeader()
	dt.updateRows()
}

// SortedColumn returns the first sort column key and direction.
// For multi-sort use SortKeys().
func (dt *DataTable) SortedColumn() (string, SortDirection) {
	if len(dt.sortKeys) == 0 {
		return "", SortNone
	}
	return dt.sortKeys[0].ColKey, dt.sortKeys[0].Dir
}

// SetSortKeys replaces the entire sort key stack.
func (dt *DataTable) SetSortKeys(keys []SortKey) {
	dt.sortKeys = append([]SortKey(nil), keys...)
	dt.rebuildDisplayIndexes()
	dt.buildHeader()
	dt.updateRows()
}

// SortKeys returns the current sort key stack.
func (dt *DataTable) SortKeys() []SortKey {
	return append([]SortKey(nil), dt.sortKeys...)
}

// SetDefaultSort sets the default sort applied when sortKeys is empty (e.g. on BindItems).
func (dt *DataTable) SetDefaultSort(key string, dir SortDirection) {
	if dir == SortNone || key == "" {
		dt.defaultSortKeys = nil
	} else {
		dt.defaultSortKeys = []SortKey{{ColKey: key, Dir: dir}}
	}
}

// SetOnMultiSort sets the callback invoked on multi-sort changes.
func (dt *DataTable) SetOnMultiSort(fn func([]SortKey)) {
	dt.onMultiSort = fn
}

// ---------------------------------------------------------------------------
// Dirty flag helpers
// ---------------------------------------------------------------------------

// Rebuild marks the layout as dirty, triggering a full re-layout on the next
// frame. This recreates all row slots from scratch, so it picks up structural
// changes like column additions, row height changes, or Label style updates.
// Not needed for normal operations (scrolling, sorting, filtering, selection,
// binding new data) — those reuse existing slots automatically.
func (dt *DataTable) Rebuild() {
	dt.layoutDirty = true
}

// Refresh marks data as dirty, triggering a slot rebuild on the next frame.
func (dt *DataTable) Refresh() {
	dt.dataDirty = true
}

// ---------------------------------------------------------------------------
// Empty state
// ---------------------------------------------------------------------------

// SetEmptyComponent sets a component to display when the table has no visible rows.
func (dt *DataTable) SetEmptyComponent(comp *Component) {
	if dt.emptyComp != nil {
		dt.node.RemoveChild(dt.emptyComp.Node())
	}
	dt.emptyComp = comp
	if comp != nil {
		dt.node.AddChild(comp.Node())
		comp.SetVisible(len(dt.displayIndexes) == 0)
	}
}

// ---------------------------------------------------------------------------
// SetSize
// ---------------------------------------------------------------------------

// SetSize sets the widget dimensions and rebuilds internal layout.
func (dt *DataTable) SetSize(w, h float64) {
	dt.Width = w
	dt.Height = h

	dt.resizeBackground(w, h)
	dt.resizeBorder(w, h)
	dt.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}

	dt.layoutInternals()
	dt.rebuildSlots()
	dt.updateScrollBar()
	dt.updateRows()
	dt.MarkLayoutDirty()
}

// ---------------------------------------------------------------------------
// Internal layout
// ---------------------------------------------------------------------------

func (dt *DataTable) sbWidth() float64 {
	if dt.showScrollBar {
		return float64(DefaultScrollBarWidth)
	}
	return 0
}

func (dt *DataTable) viewportHeight() float64 {
	hh := 0.0
	if dt.showHeader {
		hh = dt.headerHeight
	}
	return dt.Height - hh
}

func (dt *DataTable) contentWidth() float64 {
	return dt.Width - dt.sbWidth()
}

func (dt *DataTable) layoutInternals() {
	w := dt.Width
	h := dt.Height
	sbW := dt.sbWidth()
	hh := 0.0
	if dt.showHeader {
		hh = dt.headerHeight
	}
	viewH := h - hh
	contentW := w - sbW

	// Header.
	dt.headerContainer.SetPosition(0, 0)
	dt.headerContainer.SetVisible(dt.showHeader)
	dt.headerContainer.HitShape = sg.HitRect{X: 0, Y: 0, Width: contentW, Height: hh}

	// Viewport.
	dt.rowViewport.SetPosition(0, hh)
	dt.rowViewport.HitShape = sg.HitRect{X: 0, Y: 0, Width: contentW, Height: viewH}

	// Resize mask sprite (no reallocation).
	dt.maskSprite.SetScale(contentW, viewH)

	// ScrollBar.
	dt.scrollBar.SetSize(sbW, h)
	dt.scrollBar.SetPosition(w-sbW, 0)
	dt.scrollBar.SetVisible(dt.showScrollBar)

	// Resolve column widths.
	dt.resolveColumnWidths()

	// Rebuild header cells.
	dt.buildHeader()
}

func (dt *DataTable) resolveColumnWidths() {
	if len(dt.columns) == 0 {
		dt.colWidths = nil
		dt.colOffsets = nil
		return
	}
	cw := dt.contentWidth()
	dt.colWidths = make([]float64, len(dt.columns))

	totalFixed := 0.0
	totalWeight := 0.0
	for i, col := range dt.columns {
		// Hidden columns collapse to zero width and are excluded from layout.
		if col.Hidden != nil && col.Hidden() {
			dt.colWidths[i] = 0
			continue
		}
		if col.FixedWidth > 0 {
			totalFixed += col.FixedWidth
			dt.colWidths[i] = col.FixedWidth
		} else {
			w := col.Weight
			if w <= 0 {
				w = 1
			}
			totalWeight += w
		}
	}

	remaining := cw - totalFixed
	if remaining < 0 {
		remaining = 0
	}

	for i, col := range dt.columns {
		if (col.Hidden != nil && col.Hidden()) || col.FixedWidth > 0 {
			continue
		}
		w := col.Weight
		if w <= 0 {
			w = 1
		}
		colW := 0.0
		if totalWeight > 0 {
			colW = (w / totalWeight) * remaining
		}
		if col.MinWidth > 0 && colW < col.MinWidth {
			colW = col.MinWidth
		}
		if col.MaxWidth > 0 && colW > col.MaxWidth {
			colW = col.MaxWidth
		}
		dt.colWidths[i] = colW
	}

	// Precompute column x offsets.
	dt.colOffsets = make([]float64, len(dt.columns))
	x := 0.0
	for i, w := range dt.colWidths {
		dt.colOffsets[i] = x
		x += w
	}
}

// buildHeader recreates header cell nodes.
func (dt *DataTable) buildHeader() {
	// Remove all existing header children.
	for dt.headerContainer.NumChildren() > 0 {
		dt.headerContainer.RemoveChildAt(0)
	}

	if !dt.showHeader || len(dt.columns) == 0 {
		return
	}

	group := dt.EffectiveTheme().DataTable.Group(dt.Variant())
	hh := dt.headerHeight
	cw := dt.contentWidth()

	// Header background.
	headerBg := sg.NewSprite(dt.node.Name+"-header-bg", sg.TextureRegion{})
	headerBg.SetColor(group.HeaderBackground.Resolve(StateDefault).Color)
	headerBg.SetScale(cw, hh)
	dt.headerContainer.AddChild(headerBg)

	multiSort := len(dt.sortKeys) > 1

	x := 0.0
	for i, col := range dt.columns {
		colW := dt.colWidths[i]
		// Skip hidden columns entirely — no header cell created.
		if col.Hidden != nil && col.Hidden() {
			continue
		}
		ci := i // capture

		cellNode := sg.NewContainer(dt.node.Name + "-hcell-" + col.Key)
		cellNode.Interactable = true
		cellNode.HitShape = sg.HitRect{X: 0, Y: 0, Width: colW, Height: hh}
		cellNode.SetPosition(x, 0)

		// Header bg sprite for hover.
		hbg := sg.NewSprite(dt.node.Name+"-hcell-bg-"+col.Key, sg.TextureRegion{})
		hbg.SetColor(group.HeaderBackground.Resolve(StateDefault).Color)
		hbg.SetScale(colW, hh)
		cellNode.AddChild(hbg)

		// Text label.
		labelText := col.Header
		textNode := sg.NewText(dt.node.Name+"-hcell-text-"+col.Key, labelText, dt.font)
		fontSize := dt.fontSize
		if col.HeaderStyle.FontSize > 0 {
			fontSize = col.HeaderStyle.FontSize
		}
		if fontSize > 0 {
			textNode.TextBlock.FontSize = fontSize
		}
		if col.HeaderStyle.Color != (sg.Color{}) {
			textNode.TextBlock.Color = col.HeaderStyle.Color
		} else {
			textNode.TextBlock.Color = group.HeaderText.Resolve(StateDefault)
		}
		if col.HeaderStyle.Sharpness > 0 {
			textNode.TextBlock.Sharpness = col.HeaderStyle.Sharpness
		}
		px := group.CellPadding
		textNode.SetPosition(px, (hh-textNode.Height())/2)
		cellNode.AddChild(textNode)

		// Sort glyph (sprite-based, scaled from spritesheet).
		var glyphNode *sg.Node
		if col.Sortable {
			glyphNode = sg.NewSprite(dt.node.Name+"-hcell-glyph-"+col.Key, sg.TextureRegion{})
			dt.applySortGlyph(glyphNode, col.Key, group)
			glyphNode.SetPosition(colW-sortGlyphDisplaySize-px, (hh-sortGlyphDisplaySize)/2)
			cellNode.AddChild(glyphNode)

			// Sort badge number for multi-sort.
			if multiSort {
				skIdx := dt.sortKeyIndex(col.Key)
				if skIdx >= 0 && dt.font != nil {
					badgeText := strconv.Itoa(skIdx + 1)
					badge := sg.NewText(dt.node.Name+"-hcell-badge-"+col.Key, badgeText, dt.font)
					badgeFontSize := 9.0
					if dt.fontSize > 0 {
						badgeFontSize = dt.fontSize * 0.7
					}
					badge.TextBlock.FontSize = badgeFontSize
					badge.TextBlock.Color = group.SortGlyphColor.Resolve(StateDefault)
					badge.SetPosition(colW-sortGlyphDisplaySize-px-badgeFontSize, (hh-sortGlyphDisplaySize)/2)
					cellNode.AddChild(badge)
				}
			}
		}

		// Filter glyph for filterable columns.
		if col.Filterable {
			filterGlyph := sg.NewSprite(dt.node.Name+"-hcell-filter-"+col.Key, sg.TextureRegion{})
			filterGlyph.SetCustomImage(IconFilter())
			filterGlyph.SetSize(sortGlyphDisplaySize, sortGlyphDisplaySize)
			// Color: active if column has an active filter, inactive otherwise.
			if _, hasFilter := dt.columnFilters[col.Key]; hasFilter {
				filterGlyph.SetColor(group.SortGlyphColor.Resolve(StateDefault))
			} else {
				filterGlyph.SetColor(group.SortGlyphInactive.Resolve(StateDefault))
			}
			filterX := colW - sortGlyphDisplaySize - px
			if col.Sortable {
				filterX -= 14 // leave room for sort glyph
			}
			filterGlyph.SetPosition(filterX, (hh-sortGlyphDisplaySize)/2)
			filterGlyph.Interactable = true
			filterGlyph.HitShape = sg.HitRect{X: -2, Y: -2, Width: sortGlyphDisplaySize + 4, Height: sortGlyphDisplaySize + 4}
			capturedCol := col
			filterGlyph.OnClick(func(_ sg.ClickContext) {
				dt.openColumnFilterPopover(capturedCol, ci)
			})
			cellNode.AddChild(filterGlyph)
		}

		// Hover/click wiring.
		if col.Sortable {
			cellNode.OnPointerEnter(func(_ sg.PointerContext) {
				dt.headerHovered[ci] = true
				dt.updateHeaderCellBg(ci, hbg, glyphNode)
			})
			cellNode.OnPointerLeave(func(_ sg.PointerContext) {
				dt.headerHovered[ci] = false
				dt.updateHeaderCellBg(ci, hbg, glyphNode)
			})
			cellNode.OnClick(func(_ sg.ClickContext) {
				dt.cycleSort(ci, glyphNode)
			})
		}

		// Column divider (right edge).
		if dt.showColumnDividers && i < len(dt.columns)-1 {
			div := sg.NewSprite(dt.node.Name+"-hcol-div-"+col.Key, sg.TextureRegion{})
			div.SetColor(group.DividerColor.Resolve(StateDefault))
			div.SetScale(group.DividerWidth, hh)
			div.SetPosition(colW-group.DividerWidth, 0)
			cellNode.AddChild(div)
		}

		// Header bottom border.
		bord := sg.NewSprite(dt.node.Name+"-hcell-bot-"+col.Key, sg.TextureRegion{})
		bord.SetColor(group.HeaderBorderColor.Resolve(StateDefault))
		bord.SetScale(colW, group.HeaderBorderWidth)
		bord.SetPosition(0, hh-group.HeaderBorderWidth)
		cellNode.AddChild(bord)

		dt.headerContainer.AddChild(cellNode)
		x += colW
	}
}

func (dt *DataTable) updateHeaderCellBg(ci int, hbg *sg.Node, glyphNode *sg.Node) {
	group := dt.EffectiveTheme().DataTable.Group(dt.Variant())
	if dt.headerHovered[ci] {
		hbg.SetColor(group.HeaderHoverColor.Resolve(StateDefault))
	} else {
		hbg.SetColor(group.HeaderBackground.Resolve(StateDefault).Color)
	}
	if glyphNode != nil {
		if dt.sortKeyIndex(dt.columns[ci].Key) >= 0 {
			glyphNode.SetColor(group.SortGlyphColor.Resolve(StateDefault))
		} else {
			glyphNode.SetColor(group.SortGlyphInactive.Resolve(StateDefault))
		}
	}
}

// sortKeyIndex returns the index of the given column key in sortKeys, or -1.
func (dt *DataTable) sortKeyIndex(colKey string) int {
	for i, sk := range dt.sortKeys {
		if sk.ColKey == colKey {
			return i
		}
	}
	return -1
}

// sortKeyDir returns the SortDirection for the given column key, or SortNone.
func (dt *DataTable) sortKeyDir(colKey string) SortDirection {
	for _, sk := range dt.sortKeys {
		if sk.ColKey == colKey {
			return sk.Dir
		}
	}
	return SortNone
}

// applySortGlyph sets the correct image, color, and visibility on a sort glyph sprite node.
func (dt *DataTable) applySortGlyph(glyphNode *sg.Node, key string, group *theme.DataTableGroup) {
	dir := dt.sortKeyDir(key)
	if dir == SortNone {
		glyphNode.SetVisible(false)
		return
	}
	glyphNode.SetVisible(true)
	switch dir {
	case SortAsc:
		glyphNode.SetCustomImage(IconArrowUp())
	case SortDesc:
		glyphNode.SetCustomImage(IconArrowDown())
	}
	glyphNode.SetSize(sortGlyphDisplaySize, sortGlyphDisplaySize)
	glyphNode.SetColor(group.SortGlyphColor.Resolve(StateDefault))
}

func (dt *DataTable) cycleSort(colIndex int, _ *sg.Node) {
	col := dt.columns[colIndex]
	shift := engine.IsKeyPressed(engine.KeyShift)

	if shift && len(dt.sortKeys) > 0 {
		// Multi-sort: append/cycle/remove this column in the sort stack.
		idx := dt.sortKeyIndex(col.Key)
		if idx >= 0 {
			// Column already in stack — cycle or remove.
			switch dt.sortKeys[idx].Dir {
			case SortAsc:
				dt.sortKeys[idx].Dir = SortDesc
			case SortDesc:
				dt.sortKeys = append(dt.sortKeys[:idx], dt.sortKeys[idx+1:]...)
			}
		} else {
			dt.sortKeys = append(dt.sortKeys, SortKey{ColKey: col.Key, Dir: SortAsc})
		}
	} else {
		// Single-sort: same as before.
		idx := dt.sortKeyIndex(col.Key)
		if idx == 0 && len(dt.sortKeys) == 1 {
			switch dt.sortKeys[0].Dir {
			case SortAsc:
				dt.sortKeys[0].Dir = SortDesc
			case SortDesc:
				dt.sortKeys = nil
			}
		} else {
			dt.sortKeys = []SortKey{{ColKey: col.Key, Dir: SortAsc}}
		}
	}

	// Rebuild header so all glyph states update (active column + reset others).
	dt.buildHeader()

	// When onSort is registered the caller owns sort logic; do not mutate displayIndexes.
	if dt.onSort != nil {
		key, dir := "", SortNone
		if len(dt.sortKeys) > 0 {
			key = dt.sortKeys[0].ColKey
			dir = dt.sortKeys[0].Dir
		}
		dt.onSort(key, dir)
		return
	}
	if dt.onMultiSort != nil {
		dt.onMultiSort(dt.SortKeys())
	}

	dt.rebuildDisplayIndexes()
	dt.updateRows()

	switch dt.onSortScroll {
	case OnSortScrollToTop:
		dt.scrollBar.SetScrollPos(0)
	case OnSortScrollToSelection:
		dt.scrollToFirstSelected()
	}
}

// ---------------------------------------------------------------------------
// Display indexes
// ---------------------------------------------------------------------------

func (dt *DataTable) rebuildDisplayIndexes() {
	n := len(dt.items)
	dt.displayIndexes = make([]int, 0, n)
	for i := 0; i < n; i++ {
		dt.displayIndexes = append(dt.displayIndexes, i)
	}

	// Sort first (spec order: sort then filter).
	if len(dt.sortKeys) > 0 {
		dt.sortDisplayIndexes()
	}

	// Column filters.
	if len(dt.columnFilters) > 0 {
		filtered := dt.displayIndexes[:0]
		for _, idx := range dt.displayIndexes {
			if dt.passesColumnFilters(dt.items[idx]) {
				filtered = append(filtered, idx)
			}
		}
		dt.displayIndexes = filtered
	}

	// Filter.
	if dt.filterFunc != nil {
		filtered := dt.displayIndexes[:0]
		for _, idx := range dt.displayIndexes {
			if dt.filterFunc(dt.items[idx]) {
				filtered = append(filtered, idx)
			}
		}
		dt.displayIndexes = filtered
	}

	// Search filter (text search across searchable columns).
	if dt.searchRef != nil {
		query := strings.ToLower(dt.searchRef.Peek())
		if query != "" {
			filtered := dt.displayIndexes[:0]
			for _, idx := range dt.displayIndexes {
				if dt.itemMatchesSearch(dt.items[idx], query) {
					filtered = append(filtered, idx)
				}
			}
			dt.displayIndexes = filtered
		}
	}
}

// passesColumnFilters returns true if the item passes all active column filters.
func (dt *DataTable) passesColumnFilters(item any) bool {
	for colKey, allowed := range dt.columnFilters {
		if len(allowed) == 0 {
			continue
		}
		for _, col := range dt.columns {
			if col.Key != colKey {
				continue
			}
			val := ""
			if col.SearchValue != nil {
				val = col.SearchValue(item)
			} else {
				val = itemToString(item)
			}
			if !allowed[val] {
				return false
			}
			break
		}
	}
	return true
}

func (dt *DataTable) itemMatchesSearch(item any, query string) bool {
	for _, col := range dt.columns {
		if !col.Searchable || col.SearchValue == nil {
			continue
		}
		if strings.Contains(strings.ToLower(col.SearchValue(item)), query) {
			return true
		}
	}
	return false
}

func (dt *DataTable) sortDisplayIndexes() {
	// Build a list of resolved sort columns (column definition + direction).
	type resolvedSort struct {
		col DataTableColumn
		dir SortDirection
	}
	var resolved []resolvedSort
	for _, sk := range dt.sortKeys {
		for _, col := range dt.columns {
			if col.Key == sk.ColKey {
				resolved = append(resolved, resolvedSort{col, sk.Dir})
				break
			}
		}
	}
	if len(resolved) == 0 {
		return
	}

	sort.SliceStable(dt.displayIndexes, func(a, b int) bool {
		ia := dt.displayIndexes[a]
		ib := dt.displayIndexes[b]
		da := dt.items[ia]
		db := dt.items[ib]
		for _, rs := range resolved {
			cmp := dt.compareItems(rs.col, da, db)
			if rs.dir == SortDesc {
				cmp = -cmp
			}
			if cmp != 0 {
				return cmp < 0
			}
		}
		return false
	})
}

func (dt *DataTable) compareItems(col DataTableColumn, da, db any) int {
	if col.SortType == SortCustom && col.Comparator != nil {
		return col.Comparator(da, db)
	}
	if col.SortType == SortNumeric {
		na := toFloat64(col, da)
		nb := toFloat64(col, db)
		if na < nb {
			return -1
		}
		if na > nb {
			return 1
		}
		return 0
	}
	// SortAlpha: case-insensitive string comparison using column accessor.
	sa := strings.ToLower(colSortString(col, da))
	sb := strings.ToLower(colSortString(col, db))
	if sa < sb {
		return -1
	}
	if sa > sb {
		return 1
	}
	return 0
}

// colSortString returns the column-specific string for alpha sorting.
// Priority: SortValue > SearchValue > itemToString.
func colSortString(col DataTableColumn, v any) string {
	if col.SortValue != nil {
		return itemToString(col.SortValue(v))
	}
	if col.SearchValue != nil {
		return col.SearchValue(v)
	}
	return itemToString(v)
}

func toFloat64(col DataTableColumn, v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case int32:
		return float64(t)
	}
	// Prefer SortValue (may return a number directly), fall back to SearchValue.
	if col.SortValue != nil {
		sv := col.SortValue(v)
		// Try direct numeric types first.
		switch n := sv.(type) {
		case float64:
			return n
		case float32:
			return float64(n)
		case int:
			return float64(n)
		case int64:
			return float64(n)
		case int32:
			return float64(n)
		}
		if s, ok := sv.(string); ok {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return f
			}
		}
	}
	if col.SearchValue != nil {
		if f, err := strconv.ParseFloat(col.SearchValue(v), 64); err == nil {
			return f
		}
	}
	return 0
}

func itemToString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	case fmt_stringer:
		return t.String()
	}
	return ""
}

type fmt_stringer interface {
	String() string
}

// ---------------------------------------------------------------------------
// Row slot management
// ---------------------------------------------------------------------------

func (dt *DataTable) rebuildSlots() {
	// Clear existing slots.
	for dt.rowContent.NumChildren() > 0 {
		dt.rowContent.RemoveChildAt(0)
	}
	dt.rowSlots = nil

	if dt.rowHeight <= 0 || len(dt.columns) == 0 {
		return
	}

	viewH := dt.viewportHeight()
	cw := dt.contentWidth()

	var poolSize int
	if dt.scrollMode == ScrollModeVirtual {
		poolSize = int(math.Ceil(viewH/dt.rowHeight)) + 2
	} else {
		poolSize = int(math.Ceil(viewH/dt.rowHeight)) + 1
	}
	if poolSize < 1 {
		poolSize = 1
	}

	group := dt.EffectiveTheme().DataTable.Group(dt.Variant())

	for s := 0; s < poolSize; s++ {
		slot := &dtRowSlot{displayPos: -1}

		// Row container.
		slotNode := sg.NewContainer(dt.node.Name + "-row-slot")
		slotNode.Interactable = true

		// Background sprite.
		bgNode := sg.NewSprite(dt.node.Name+"-row-bg", sg.TextureRegion{})
		bgNode.SetColor(group.RowBackground.Resolve(StateDefault).Color)
		bgNode.SetScale(cw, dt.rowHeight)
		slotNode.AddChild(bgNode)

		slot.node = slotNode
		slot.bgNode = bgNode

		// Cells: one per column.
		slot.cells = make([]*Component, len(dt.columns))
		x := 0.0
		for ci := range dt.columns {
			if ci >= len(dt.colWidths) {
				break
			}
			col := dt.columns[ci]
			colW := dt.colWidths[ci]
			pad := group.CellPadding

			// Create placeholder cell: real content set by updateRows.
			cellComp := NewComponent(dt.node.Name + "-cell-" + col.Key)
			cellComp.Width = colW - 2*pad
			cellComp.node.SetPosition(x+pad, 0)
			slotNode.AddChild(cellComp.node)
			slot.cells[ci] = cellComp

			// Column divider.
			if dt.showColumnDividers && ci < len(dt.columns)-1 {
				div := sg.NewSprite(dt.node.Name+"-col-div", sg.TextureRegion{})
				div.SetColor(group.DividerColor.Resolve(StateDefault))
				div.SetScale(group.DividerWidth, dt.rowHeight)
				div.SetPosition(x+colW-group.DividerWidth, 0)
				slotNode.AddChild(div)
				slot.dividers = append(slot.dividers, div)
			}

			x += colW
		}

		// Row bottom divider.
		if dt.showRowDividers {
			rowDiv := sg.NewSprite(dt.node.Name+"-row-div", sg.TextureRegion{})
			rowDiv.SetColor(group.DividerColor.Resolve(StateDefault))
			rowDiv.SetScale(cw, group.DividerWidth)
			rowDiv.SetPosition(0, dt.rowHeight-group.DividerWidth)
			slotNode.AddChild(rowDiv)
			slot.rowDivider = rowDiv
		}

		slotNode.HitShape = sg.HitRect{X: 0, Y: 0, Width: cw, Height: dt.rowHeight}
		dt.rowContent.AddChild(slotNode)
		dt.rowSlots = append(dt.rowSlots, slot)

		// Wire click handlers once per slot — read displayPos dynamically at event time.
		sRef := slot
		sRef.node.OnPointerEnter(func(_ sg.PointerContext) {
			if sRef.displayPos < 0 {
				return
			}
			sRef.hovered = true
			if !dt.isSelected(dt.displayIndexes[sRef.displayPos]) {
				sRef.bgNode.SetColor(dt.dtTheme().RowHoverColor.Resolve(StateDefault))
			}
		})
		sRef.node.OnPointerLeave(func(_ sg.PointerContext) {
			sRef.hovered = false
			if sRef.displayPos >= 0 && sRef.displayPos < len(dt.displayIndexes) {
				g := dt.dtTheme()
				dt.applyRowBg(sRef, sRef.displayPos, dt.displayIndexes[sRef.displayPos], g)
			}
		})
		sRef.node.OnClick(func(ctx sg.ClickContext) {
			if sRef.displayPos < 0 || sRef.displayPos >= len(dt.displayIndexes) {
				return
			}
			dataIdx := dt.displayIndexes[sRef.displayPos]
			// Selection column direct click: checkbox/radio are non-interactable
			// so the click falls through to the row. Use individual toggle so
			// multi-select works like a checkbox (add/remove one row) rather
			// than the replace-selection behavior of a plain row click.
			if dt.clickInSelectionColumn(ctx.LocalX) && dt.selectionMode != SelectionModeNone {
				if dt.isSelected(dataIdx) {
					dt.DeselectRow(dataIdx)
				} else {
					dt.SelectRow(dataIdx)
				}
			} else {
				// Row-click selection: triggered by the table's rowClickSelects
				// flag or any SelectionColumn with RowClickSelects enabled.
				wantsRowClick := dt.rowClickSelects
				if !wantsRowClick {
					wantsRowClick = dt.selectionColumnWantsRowClick()
				}
				if wantsRowClick && dt.selectionMode != SelectionModeNone {
					dt.handleRowSelectionWithModifiers(dataIdx)
				}
			}
			if dt.onCellClick != nil {
				dt.onCellClick(CellCoord{Row: dataIdx, Col: 0}, dt.items[dataIdx])
			}
		})
	}
}

// updateRows repositions and populates the row slots based on current scroll position.
func (dt *DataTable) updateRows() {
	if len(dt.rowSlots) == 0 || len(dt.columns) == 0 {
		if dt.emptyComp != nil {
			dt.emptyComp.SetVisible(len(dt.displayIndexes) == 0)
		}
		return
	}

	pos := dt.scrollPos.Peek()
	displayCount := len(dt.displayIndexes)

	group := dt.EffectiveTheme().DataTable.Group(dt.Variant())

	if dt.scrollMode == ScrollModeVirtual {
		// Virtual mode: shift slots around as content scrolls.
		dt.rowContent.SetPosition(0, -pos)

		startRow := int(math.Floor(pos / dt.rowHeight))
		if startRow < 0 {
			startRow = 0
		}

		for s, slot := range dt.rowSlots {
			dispIdx := startRow + s
			if dispIdx >= displayCount {
				slot.node.SetVisible(false)
				slot.displayPos = -1
				continue
			}
			slot.node.SetVisible(true)
			slot.displayPos = dispIdx
			slot.node.SetPosition(0, float64(dispIdx)*dt.rowHeight)

			dataIdx := dt.displayIndexes[dispIdx]
			data := dt.items[dataIdx]

			dt.populateSlotCells(slot, dataIdx, data, group)
			dt.applyRowBg(slot, dispIdx, dataIdx, group)
		}
	} else {
		// Static mode.
		dt.rowContent.SetPosition(0, 0)

		startRow := int(math.Floor(pos / dt.rowHeight))
		if startRow < 0 {
			startRow = 0
		}
		dt.topRowIndex = startRow

		for s, slot := range dt.rowSlots {
			dispIdx := startRow + s
			slot.node.SetPosition(0, float64(s)*dt.rowHeight)

			if dispIdx >= displayCount {
				slot.node.SetVisible(false)
				slot.displayPos = -1
				continue
			}
			slot.node.SetVisible(true)
			slot.displayPos = dispIdx

			dataIdx := dt.displayIndexes[dispIdx]
			data := dt.items[dataIdx]

			dt.populateSlotCells(slot, dataIdx, data, group)
			dt.applyRowBg(slot, dispIdx, dataIdx, group)
		}
	}

	// Show/hide empty state component.
	if dt.emptyComp != nil {
		dt.emptyComp.SetVisible(len(dt.displayIndexes) == 0)
	}
}

// populateSlotCells updates cell content for a slot given a data item.
func (dt *DataTable) populateSlotCells(slot *dtRowSlot, dataIdx int, data any, group *theme.DataTableGroup) {
	for ci, col := range dt.columns {
		cell := slot.cells[ci]
		if cell == nil {
			continue
		}
		// Skip hidden columns — hide the cell node so it takes no space.
		if col.Hidden != nil && col.Hidden() {
			cell.node.SetVisible(false)
			continue
		}
		cell.node.SetVisible(true)
		colW := dt.colWidths[ci]
		pad := group.CellPadding
		if col.UpdateCell != nil {
			col.UpdateCell(dataIdx, data, cell)
		} else if col.RenderCell != nil {
			// Re-render: clear old children and add new component's node.
			for cell.node.NumChildren() > 0 {
				cell.node.RemoveChildAt(0)
			}
			newComp := col.RenderCell(dataIdx, data)
			if newComp != nil {
				cell.node.AddChild(newComp.node)
				slot.cells[ci] = newComp
			}
		}
		// Inject font and apply label style on first init.
		if l, ok := cell.UserData().(*Label); ok && l != nil && l.Font() == nil {
			if dt.source != nil {
				l.SetFont(dt.source)
			}
			fontSize := dt.fontSize
			if col.Cell.FontSize > 0 {
				fontSize = col.Cell.FontSize
			}
			if fontSize > 0 {
				l.SetFontSize(fontSize)
			}
			if col.Cell.Color != (sg.Color{}) {
				l.SetColor(col.Cell.Color)
			}
			if col.Cell.Sharpness > 0 {
				l.SetSharpness(col.Cell.Sharpness)
			}
			l.SetAlign(col.Cell.Align)
			if col.Cell.Align != 0 {
				l.SetWrapWidth(colW - 2*pad)
			}
		}

		// Apply OnPostUpdate hook every populate cycle.
		if col.Cell.OnPostUpdate != nil {
			col.Cell.OnPostUpdate(data, cell)
		}

		// Truncation with tooltip for ClipEllipsis mode.
		if col.ClipMode == ClipEllipsis {
			if l, ok := cell.UserData().(*Label); ok && l != nil {
				fullText := l.Text()
				maxW := colW - 2*pad
				font := l.Font()
				fSize := l.displaySize
				if font != nil && fSize > 0 {
					tw, _ := measureDisplay(font, fullText, fSize)
					if tw > maxW {
						truncated := dt.truncateText(font, fullText, fSize, maxW)
						l.SetText(truncated)
						cell.SetTooltipText(fullText, dt.source, fSize)
					} else {
						cell.ClearTooltip()
					}
				}
			}
		}

		// Compute content height for vertical centering.
		contentH := cell.Height
		if l, ok := cell.UserData().(*Label); ok && l != nil {
			contentH = l.Height
		}
		// Reposition using precomputed offset.
		x := dt.colOffsets[ci]
		cell.node.SetPosition(x+pad, (dt.rowHeight-contentH)/2)
		cell.Width = colW - 2*pad
	}
}

// truncateText shortens text to fit within maxW, appending "..." at the end.
func (dt *DataTable) truncateText(font *sg.FontFamily, text string, fontSize, maxW float64) string {
	runes := []rune(text)
	ellipsis := "..."
	ellipsisW, _ := measureDisplay(font, ellipsis, fontSize)
	target := maxW - ellipsisW
	if target <= 0 {
		return ellipsis
	}
	for i := len(runes); i > 0; i-- {
		w, _ := measureDisplay(font, string(runes[:i]), fontSize)
		if w <= target {
			return string(runes[:i]) + ellipsis
		}
	}
	return ellipsis
}

func (dt *DataTable) applyRowBg(slot *dtRowSlot, dispIdx, dataIdx int, group *theme.DataTableGroup) {
	selected := dt.isSelected(dataIdx)

	var color sg.Color
	if selected {
		color = group.SelectionColor.Resolve(StateDefault)
	} else if slot.hovered {
		color = group.RowHoverColor.Resolve(StateDefault)
	} else if dt.zebraStriping && dispIdx%2 == 1 {
		color = group.RowBackgroundAlt.Resolve(StateDefault).Color
	} else {
		color = group.RowBackground.Resolve(StateDefault).Color
	}
	slot.bgNode.SetColor(color)
}

func (dt *DataTable) isSelected(dataIdx int) bool {
	for _, s := range dt.selectedIndexes {
		if s == dataIdx {
			return true
		}
	}
	return false
}

func (dt *DataTable) handleRowSelectionWithModifiers(dataIdx int) {
	shift := engine.IsKeyPressed(engine.KeyShift)
	ctrl := engine.IsKeyPressed(engine.KeyControl) || engine.IsKeyPressed(engine.KeyMeta)

	switch dt.selectionMode {
	case SelectionModeSingle:
		if len(dt.selectedIndexes) == 1 && dt.selectedIndexes[0] == dataIdx {
			dt.selectedIndexes = nil
		} else {
			dt.selectedIndexes = []int{dataIdx}
		}
		dt.anchorRow = dataIdx

	case SelectionModeMulti:
		if shift && dt.anchorRow >= 0 {
			// Range select: find display positions of anchor and clicked row.
			anchorDisp := -1
			clickedDisp := -1
			for i, di := range dt.displayIndexes {
				if di == dt.anchorRow {
					anchorDisp = i
				}
				if di == dataIdx {
					clickedDisp = i
				}
			}
			if anchorDisp >= 0 && clickedDisp >= 0 {
				lo, hi := anchorDisp, clickedDisp
				if lo > hi {
					lo, hi = hi, lo
				}
				dt.selectedIndexes = nil
				for i := lo; i <= hi; i++ {
					dt.selectedIndexes = append(dt.selectedIndexes, dt.displayIndexes[i])
				}
			}
		} else if ctrl {
			// Toggle individual row without moving anchor.
			if dt.isSelected(dataIdx) {
				newSel := dt.selectedIndexes[:0]
				for _, s := range dt.selectedIndexes {
					if s != dataIdx {
						newSel = append(newSel, s)
					}
				}
				dt.selectedIndexes = newSel
			} else {
				dt.selectedIndexes = append(dt.selectedIndexes, dataIdx)
			}
		} else {
			// Plain click: replace selection.
			if dt.isSelected(dataIdx) && len(dt.selectedIndexes) == 1 {
				dt.selectedIndexes = nil
			} else {
				dt.selectedIndexes = []int{dataIdx}
			}
			dt.anchorRow = dataIdx
		}
	}

	dt.selectionDirty = true
	dt.syncSelIdx()
	if dt.onSelectionChanged != nil {
		dt.onSelectionChanged(dt.selectedIndexes)
	}
}

// handleRowSelection is kept for programmatic selection (no modifier keys).
func (dt *DataTable) handleRowSelection(dataIdx int) {
	switch dt.selectionMode {
	case SelectionModeSingle:
		if len(dt.selectedIndexes) == 1 && dt.selectedIndexes[0] == dataIdx {
			dt.selectedIndexes = nil
		} else {
			dt.selectedIndexes = []int{dataIdx}
		}
		dt.anchorRow = dataIdx
	case SelectionModeMulti:
		if dt.isSelected(dataIdx) {
			newSel := dt.selectedIndexes[:0]
			for _, s := range dt.selectedIndexes {
				if s != dataIdx {
					newSel = append(newSel, s)
				}
			}
			dt.selectedIndexes = newSel
		} else {
			dt.selectedIndexes = append(dt.selectedIndexes, dataIdx)
		}
		dt.anchorRow = dataIdx
	}

	dt.selectionDirty = true
	dt.syncSelIdx()
	if dt.onSelectionChanged != nil {
		dt.onSelectionChanged(dt.selectedIndexes)
	}
}

// ---------------------------------------------------------------------------
// Scrollbar
// ---------------------------------------------------------------------------

func (dt *DataTable) updateScrollBar() {
	totalH := float64(len(dt.displayIndexes)) * dt.rowHeight
	viewH := dt.viewportHeight()
	dt.scrollBar.SetContentSize(totalH, viewH)
	dt.scrollBar.SetVisible(dt.showScrollBar && totalH > viewH)
}

func (dt *DataTable) scrollToFirstSelected() {
	if len(dt.selectedIndexes) == 0 {
		return
	}
	selData := dt.selectedIndexes[0]
	for i, di := range dt.displayIndexes {
		if di == selData {
			targetY := float64(i) * dt.rowHeight
			dt.scrollBar.SetScrollPos(targetY)
			return
		}
	}
}

// ensureDisplayRowVisible scrolls the minimum amount needed so that the row
// at the given display index is fully visible in the viewport.
func (dt *DataTable) ensureDisplayRowVisible(displayIdx int) {
	rowTop := float64(displayIdx) * dt.rowHeight
	rowBot := rowTop + dt.rowHeight
	viewH := dt.viewportHeight()
	pos := dt.scrollPos.Peek()
	if rowTop < pos {
		dt.scrollBar.SetScrollPos(rowTop)
	} else if rowBot > pos+viewH {
		dt.scrollBar.SetScrollPos(rowBot - viewH)
	}
}

// ---------------------------------------------------------------------------
// Theme
// ---------------------------------------------------------------------------

func (dt *DataTable) applyThemeColors() {
	group := dt.EffectiveTheme().DataTable.Group(dt.Variant())
	bg := group.Background.Resolve(StateDefault)
	dt.applyBackground(bg)
	dt.MarkDrawDirty()
}

// DataTable theme accessor helper.
func (dt *DataTable) dtTheme() *theme.DataTableGroup {
	return dt.EffectiveTheme().DataTable.Group(dt.Variant())
}

// ---------------------------------------------------------------------------
// Update (auto-registered via OnUpdate)
// ---------------------------------------------------------------------------

// Update is called automatically once per frame via node.OnUpdate.
func (dt *DataTable) Update() {
	// Mouse wheel scrolling.
	if dt.containsCursor() {
		_, wy := engine.Wheel()
		if wy != 0 {
			newPos := dt.scrollPos.Peek() - wy*scrollWheelSpeedDataTable
			dt.scrollBar.SetScrollPos(newPos)
		}
	}

	// Arrow-key navigation: up/down moves selection when the cursor is over
	// the table and a selection mode is active.
	if dt.selectionMode != SelectionModeNone && dt.containsCursor() && len(dt.displayIndexes) > 0 {
		delta := 0
		if core.IsKeyJustPressed(engine.KeyArrowDown) {
			delta = 1
		} else if core.IsKeyJustPressed(engine.KeyArrowUp) {
			delta = -1
		}
		if delta != 0 {
			// Find the current display position of the first selected row.
			curDisplay := -1
			if len(dt.selectedIndexes) > 0 {
				sel := dt.selectedIndexes[0]
				for i, di := range dt.displayIndexes {
					if di == sel {
						curDisplay = i
						break
					}
				}
			}
			var nextDisplay int
			if curDisplay < 0 {
				// Nothing selected: down selects first row, up selects last.
				if delta > 0 {
					nextDisplay = 0
				} else {
					nextDisplay = len(dt.displayIndexes) - 1
				}
			} else {
				nextDisplay = curDisplay + delta
				if nextDisplay < 0 {
					nextDisplay = 0
				} else if nextDisplay >= len(dt.displayIndexes) {
					nextDisplay = len(dt.displayIndexes) - 1
				}
			}
			dataIdx := dt.displayIndexes[nextDisplay]
			dt.SelectRow(dataIdx)
			dt.ensureDisplayRowVisible(nextDisplay)
		}
	}

	// Process dirty flags in a single batched pass so that rebuildSlots and
	// updateRows are each called at most once per frame regardless of how many
	// flags are set simultaneously (e.g. BindItems with a sort configured sets
	// all three at once).
	needLayout := dt.layoutDirty
	needData := dt.dataDirty
	needSortFilter := dt.sortFilterDirty
	dt.layoutDirty = false
	dt.dataDirty = false
	dt.sortFilterDirty = false

	if needLayout {
		dt.layoutInternals()
	}
	if needLayout || needData {
		dt.rebuildSlots()
	}
	// rebuildDisplayIndexes must precede updateScrollBar because the scroll bar
	// reads len(displayIndexes) to compute total content height.
	if needSortFilter {
		dt.rebuildDisplayIndexes()
	}
	if needLayout || needSortFilter {
		dt.updateScrollBar()
	}
	if needLayout || needData || needSortFilter {
		dt.updateRows()
	}

	if dt.selectionDirty {
		dt.selectionDirty = false
		group := dt.dtTheme()
		hasSelCol := dt.hasSelectionColumn()
		for _, slot := range dt.rowSlots {
			if slot.displayPos < 0 || slot.displayPos >= len(dt.displayIndexes) {
				continue
			}
			dataIdx := dt.displayIndexes[slot.displayPos]
			dt.applyRowBg(slot, slot.displayPos, dataIdx, group)
			// When a SelectionColumn is present, repopulate cells so
			// checkbox/radio widgets update their checked/selected state.
			if hasSelCol {
				dt.populateSlotCells(slot, dataIdx, dt.items[dataIdx], group)
			}
		}
	}
}

// sortGlyphDisplaySize is the visual display size (in pixels) for sort and
// filter glyphs in data table headers.
const sortGlyphDisplaySize = 9.0

// sortAscGlyph returns the upward arrow glyph from the default spritesheet.
func sortAscGlyph() engine.Image { return IconArrowUp() }

// sortDescGlyph returns the downward arrow glyph from the default spritesheet.
func sortDescGlyph() engine.Image { return IconArrowDown() }

// filterGlyphImg returns the filter funnel glyph from the default spritesheet.
func filterGlyphImg() engine.Image { return IconFilter() }

// ---------------------------------------------------------------------------
// Per-column filter popover
// ---------------------------------------------------------------------------

// openColumnFilterPopover opens a popover with checkboxes for unique values in the column.
func (dt *DataTable) openColumnFilterPopover(col DataTableColumn, colIndex int) {
	// Gather unique values from all items using the column's SearchValue or itemToString.
	uniqueMap := make(map[string]bool)
	for _, item := range dt.items {
		val := ""
		if col.SearchValue != nil {
			val = col.SearchValue(item)
		} else {
			val = itemToString(item)
		}
		uniqueMap[val] = true
	}

	// Sort unique values.
	uniqueVals := make([]string, 0, len(uniqueMap))
	for v := range uniqueMap {
		uniqueVals = append(uniqueVals, v)
	}
	sort.Strings(uniqueVals)

	// Determine which values are currently allowed.
	currentFilter := dt.columnFilters[col.Key]
	allSelected := currentFilter == nil || len(currentFilter) == 0

	// Build content panel with checkboxes.
	panel := NewPanel(dt.node.Name + "-filter-panel-" + col.Key)

	// "All" and "None" buttons.
	allBtn := NewButton(dt.node.Name+"-filter-all-"+col.Key, "All", dt.source, 11)
	allBtn.SetSize(50, 22)
	allBtn.SetPosition(4, 4)

	noneBtn := NewButton(dt.node.Name+"-filter-none-"+col.Key, "None", dt.source, 11)
	noneBtn.SetSize(50, 22)
	noneBtn.SetPosition(58, 4)

	// Create checkboxes for each value.
	checkboxes := make([]*Checkbox, len(uniqueVals))
	yOff := 30.0
	maxW := 110.0
	for i, val := range uniqueVals {
		cb := NewCheckbox(dt.node.Name+"-filter-cb-"+col.Key+"-"+strconv.Itoa(i), val, dt.source, 11)
		if allSelected || (currentFilter != nil && currentFilter[val]) {
			cb.SetChecked(true)
		}
		cb.SetPosition(4, yOff)
		panel.AddChild(cb)
		checkboxes[i] = cb
		yOff += 24
		if cb.Width+8 > maxW {
			maxW = cb.Width + 8
		}
	}

	panel.AddChild(allBtn)
	panel.AddChild(noneBtn)

	contentH := yOff + 30 // room for apply button
	if contentH > 300 {
		contentH = 300
	}
	contentW := maxW
	if contentW < 120 {
		contentW = 120
	}

	// Apply button at the bottom.
	applyBtn := NewButton(dt.node.Name+"-filter-apply-"+col.Key, "Apply", dt.source, 11)
	applyBtn.SetSize(contentW-8, 24)
	applyBtn.SetPosition(4, yOff)
	panel.AddChild(applyBtn)

	panel.SetSize(contentW, contentH)

	// Wire "All" and "None" buttons.
	allBtn.SetOnClick(func() {
		for _, cb := range checkboxes {
			cb.SetChecked(true)
		}
	})
	noneBtn.SetOnClick(func() {
		for _, cb := range checkboxes {
			cb.SetChecked(false)
		}
	})

	colKey := col.Key
	applyBtn.SetOnClick(func() {
		var selected []string
		allChecked := true
		for i, cb := range checkboxes {
			if cb.Checked() {
				selected = append(selected, uniqueVals[i])
			} else {
				allChecked = false
			}
		}
		if allChecked || len(selected) == 0 {
			dt.ClearColumnFilter(colKey)
		} else {
			dt.SetColumnFilter(colKey, selected)
		}
		DefaultPopoverManager.dismiss()
		dt.buildHeader()
	})

	pop := NewPopover(dt.node.Name + "-filter-pop-" + col.Key)
	pop.SetTitle("Filter: "+col.Header, dt.source, 12)
	pop.SetShowCloseButton(true)
	pop.SetContentSize(contentW, contentH)
	pop.SetContent(panel)
	pop.SetPreferredSide(PopoverBelow)

	// Use a temporary component anchored at the header cell position.
	triggerComp := NewComponent(dt.node.Name + "-filter-trigger-" + col.Key)
	triggerComp.Width = dt.colWidths[colIndex]
	triggerComp.Height = dt.headerHeight
	wx, wy := dt.headerContainer.LocalToWorld(dt.colOffsets[colIndex], 0)
	triggerComp.node.SetPosition(wx, wy)
	// Add to scene root temporarily for positioning.
	if sc := currentScene(); sc != nil && sc.Root != nil {
		sc.Root.AddChild(triggerComp.node)
		pop.SetOnClose(func() {
			sc.Root.RemoveChild(triggerComp.node)
		})
	}

	pop.Open(triggerComp)
}

// ---------------------------------------------------------------------------
// Public API helpers
// ---------------------------------------------------------------------------

// Dispose cleans up watches and child components.
func (dt *DataTable) Dispose() {
	if dt.stopArray != nil {
		dt.stopArray()
	}
	dt.searchWatch.Stop()
	dt.scrollPosWatch.Stop()
	dt.scrollBar.Dispose()
	dt.Component.Dispose()
}

// DataTableScrollBar returns the internal scrollbar for testing.
func (dt *DataTable) DataTableScrollBar() *ScrollBar { return dt.scrollBar }

// DataTableScrollPos returns the reactive scroll position ref for testing.
func (dt *DataTable) DataTableScrollPos() *Ref[float64] { return dt.scrollPos }

// DataTableDisplayCount returns the number of rows currently displayed (after filter/sort).
func (dt *DataTable) DataTableDisplayCount() int { return len(dt.displayIndexes) }

// DataTableDisplayIndexes returns a copy of the current display index order.
func (dt *DataTable) DataTableDisplayIndexes() []int {
	return append([]int(nil), dt.displayIndexes...)
}

// DataTableSelectedIndexes returns the current selection (data indexes).
func (dt *DataTable) DataTableSelectedIndexes() []int { return dt.selectedIndexes }
