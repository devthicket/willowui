package widget

import (
	"image"
	"image/color"
	"sort"
	"strings"
	"sync"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

// TableColumn defines a column in a TreeTable.
type TableColumn struct {
	Key      string
	Label    string
	Width    float64
	Sortable bool
}

// SortDir indicates the sort order for a TreeTable column.
type SortDir int

const (
	SortDirAsc  SortDir = iota
	SortDirDesc
)

// TreeTableRow represents a row in the TreeTable hierarchy.
type TreeTableRow struct {
	ID       string
	ParentID string // empty = root row
	Cells    map[string]string
	Children []TreeTableRow
}

// ---------------------------------------------------------------------------
// Layout constants
// ---------------------------------------------------------------------------

const (
	ttHeaderHeight = 30.0
	ttRowHeight    = 28.0
	ttIndentWidth  = 20.0
	ttChevronSize  = 12.0
	ttPadLeft      = 6.0
	ttHandleWidth  = 6.0
	ttScrollBarW   = 16.0
)

// ---------------------------------------------------------------------------
// Internal flat entry
// ---------------------------------------------------------------------------

type treeTableFlatEntry struct {
	row      TreeTableRow
	depth    int
	hasKids  bool
	expanded bool
}

// ---------------------------------------------------------------------------
// TreeTable widget
// ---------------------------------------------------------------------------

// TreeTable is a hybrid tree + column grid where rows can be expanded/collapsed
// while each row spans multiple data columns.
type TreeTable struct {
	Component

	headerNode *sg.Node
	viewport   *sg.Node
	content    *sg.Node
	scrollBar  *ScrollBar

	columns  []TableColumn
	rows     []TreeTableRow // original hierarchical data
	flatRows []TreeTableRow // flattened from Children
	flatList []treeTableFlatEntry
	expanded map[string]bool

	sortKey string
	sortDir SortDir

	scrollPos float64
	selected  int // flat list index, -1 = none

	// Visual nodes.
	selHighlight *sg.Node
	rowNodes     []*treeTableRowNode
	headerCells  []*sg.Node

	// Column resize state.
	resizing       bool
	resizeColIndex int
	resizeStartX   float64
	resizeOrigW    float64

	// Callbacks.
	onRowClick  func(id string)
	onRowExpand func(id string, expanded bool)

	// Font.
	font        *sg.FontFamily
	displaySize float64

	width, height float64
}

type treeTableRowNode struct {
	container *sg.Node
	flatIdx   int
}

// NewTreeTable creates a new TreeTable with the given name, font, and display size.
func NewTreeTable(name string, source *sg.FontFamily, displaySize float64) *TreeTable {
	tt := &TreeTable{
		expanded:    make(map[string]bool),
		selected:    -1,
		font:        source,
		displaySize: displaySize,
		sortDir:     SortDirAsc,
	}

	initComponent(&tt.Component, name)
	tt.initBackground(name)
	tt.initBorder(name)

	group := tt.EffectiveTheme().TreeTable.Group(tt.Variant())
	tt.applyBackground(group.Background.Resolve(StateDefault))
	tt.applyBorder(group.Border.Resolve(StateDefault), group.BorderWidth, group.Background.Resolve(StateDefault))

	// Header container.
	tt.headerNode = sg.NewContainer(name + "-header")
	tt.headerNode.Interactable = true
	tt.node.AddChild(tt.headerNode)

	// Viewport (clips content).
	tt.viewport = sg.NewContainer(name + "-viewport")
	tt.viewport.Interactable = true
	tt.node.AddChild(tt.viewport)

	// Content container inside viewport.
	tt.content = sg.NewContainer(name + "-content")
	tt.content.Interactable = true
	tt.viewport.AddChild(tt.content)

	// ScrollBar.
	tt.scrollBar = NewScrollBar(name + "-scrollbar")
	tt.scrollBar.SetOnChange(func(pos float64) {
		old := tt.scrollPos
		tt.scrollPos = pos
		if pos != old {
			tt.rebuild()
		}
	})
	tt.scrollBar.AddToNode(tt.node)

	// Selection highlight.
	tt.selHighlight = sg.NewSprite(name+"-sel-hl", sg.TextureRegion{})
	tt.selHighlight.SetVisible(false)
	tt.selHighlight.SetZIndex(-1)
	tt.content.AddChild(tt.selHighlight)

	// Auto-update: mouse wheel scrolling.
	tt.node.OnUpdate = func(_ float64) {
		tt.update()
	}

	tt.scrollBar.parent = &tt.Component
	tt.onThemeChange = func() {
		tt.applyThemeColors()
		tt.scrollBar.applyThemeColors()
		tt.rebuildHeader()
		tt.rebuild()
	}

	tt.SetSize(500, 400)
	return tt
}

// applyThemeColors applies the current theme to the TreeTable background and border.
func (tt *TreeTable) applyThemeColors() {
	group := tt.EffectiveTheme().TreeTable.Group(tt.Variant())
	tt.applyBackground(group.Background.Resolve(StateDefault))
	tt.applyBorder(group.Border.Resolve(StateDefault), group.BorderWidth, group.Background.Resolve(StateDefault))
	tt.MarkDrawDirty()
}

// ---------------------------------------------------------------------------
// Accessors
// ---------------------------------------------------------------------------

// Node returns the root node for adding to a scene.
func (tt *TreeTable) Node() *sg.Node {
	return tt.node
}

// Dispose cleans up resources.
func (tt *TreeTable) Dispose() {
	tt.clearRows()
}

// ---------------------------------------------------------------------------
// Public API: structure
// ---------------------------------------------------------------------------

// SetColumns sets the column definitions.
func (tt *TreeTable) SetColumns(cols []TableColumn) {
	tt.columns = make([]TableColumn, len(cols))
	copy(tt.columns, cols)
	tt.rebuildHeader()
	tt.rebuild()
}

// Columns returns a copy of the current column definitions.
func (tt *TreeTable) Columns() []TableColumn {
	out := make([]TableColumn, len(tt.columns))
	copy(out, tt.columns)
	return out
}

// SetRows sets the row data. Rows may be nested via Children or flat via ParentID.
func (tt *TreeTable) SetRows(rows []TreeTableRow) {
	tt.rows = rows
	tt.flatRows = flattenTreeTableRows(rows, "")
	tt.rebuildFlatList()
	tt.updateScrollBar()
	tt.rebuild()
}

// Rows returns a copy of the original row data.
func (tt *TreeTable) Rows() []TreeTableRow {
	out := make([]TreeTableRow, len(tt.rows))
	copy(out, tt.rows)
	return out
}

// ---------------------------------------------------------------------------
// Public API: sort
// ---------------------------------------------------------------------------

// SetSortColumn sets the sort column key and direction.
func (tt *TreeTable) SetSortColumn(key string, dir SortDir) {
	prevID := tt.selectedRowID()
	tt.sortKey = key
	tt.sortDir = dir
	tt.rebuildFlatList()
	tt.remapSelection(prevID)
	tt.updateScrollBar()
	tt.rebuild()
	tt.rebuildHeader()
}

// ---------------------------------------------------------------------------
// Public API: expand state
// ---------------------------------------------------------------------------

// ExpandAll expands all rows that have children.
func (tt *TreeTable) ExpandAll() {
	prevID := tt.selectedRowID()
	for _, row := range tt.flatRows {
		if tt.hasChildren(row.ID) {
			tt.expanded[row.ID] = true
		}
	}
	tt.rebuildFlatList()
	tt.remapSelection(prevID)
	tt.updateScrollBar()
	tt.rebuild()
}

// CollapseAll collapses all rows.
func (tt *TreeTable) CollapseAll() {
	prevID := tt.selectedRowID()
	tt.expanded = make(map[string]bool)
	tt.rebuildFlatList()
	tt.remapSelection(prevID)
	tt.updateScrollBar()
	tt.rebuild()
}

// SetExpanded sets the expansion state for a given row ID.
func (tt *TreeTable) SetExpanded(id string, v bool) {
	prevID := tt.selectedRowID()
	if v {
		tt.expanded[id] = true
	} else {
		delete(tt.expanded, id)
	}
	tt.rebuildFlatList()
	tt.remapSelection(prevID)
	tt.updateScrollBar()
	tt.rebuild()
	if tt.onRowExpand != nil {
		tt.onRowExpand(id, v)
	}
}

// IsExpanded reports whether the given row ID is expanded.
func (tt *TreeTable) IsExpanded(id string) bool {
	return tt.expanded[id]
}

// ---------------------------------------------------------------------------
// Public API: callbacks
// ---------------------------------------------------------------------------

// SetOnRowClick sets the callback for interactive row clicks.
func (tt *TreeTable) SetOnRowClick(fn func(id string)) {
	tt.onRowClick = fn
}

// SetOnRowExpand sets the callback for expand/collapse events.
func (tt *TreeTable) SetOnRowExpand(fn func(id string, expanded bool)) {
	tt.onRowExpand = fn
}

// ---------------------------------------------------------------------------
// Public API: selection
// ---------------------------------------------------------------------------

// Selected returns the currently selected flat list index, or -1.
func (tt *TreeTable) Selected() int {
	return tt.selected
}

// SetSelected sets the selected flat list index programmatically.
// This does NOT fire OnRowClick.
func (tt *TreeTable) SetSelected(idx int) {
	n := len(tt.flatList)
	if idx < -1 || idx >= n {
		return
	}
	tt.selected = idx
	tt.updateHighlight()
	if idx >= 0 {
		tt.scrollToIndex(idx)
	}
}

// VisibleRowCount returns the number of rows currently in the flat list.
func (tt *TreeTable) VisibleRowCount() int {
	return len(tt.flatList)
}

// RowIDAt returns the row ID at the given flat list index, or "" if out of range.
func (tt *TreeTable) RowIDAt(idx int) string {
	if idx < 0 || idx >= len(tt.flatList) {
		return ""
	}
	return tt.flatList[idx].row.ID
}

// ---------------------------------------------------------------------------
// SetSize
// ---------------------------------------------------------------------------

// SetSize sets the widget dimensions.
func (tt *TreeTable) SetSize(w, h float64) {
	tt.width = w
	tt.height = h
	tt.Width = w
	tt.Height = h

	tt.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}

	// Header.
	tt.headerNode.SetPosition(0, 0)
	tt.headerNode.HitShape = sg.HitRect{X: 0, Y: 0, Width: w - ttScrollBarW, Height: ttHeaderHeight}

	// Viewport below header.
	bodyH := h - ttHeaderHeight
	tt.viewport.SetPosition(0, ttHeaderHeight)

	// Clipping mask for viewport.
	maskRoot := sg.NewContainer(tt.node.Name + "-mask")
	maskSprite := sg.NewSprite(tt.node.Name+"-mask-rect", sg.TextureRegion{})
	maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	maskSprite.SetScale(w-ttScrollBarW, bodyH)
	maskRoot.AddChild(maskSprite)
	tt.viewport.SetMask(maskRoot)

	// ScrollBar.
	tt.scrollBar.SetSize(ttScrollBarW, bodyH)
	tt.scrollBar.SetPosition(w-ttScrollBarW, ttHeaderHeight)

	tt.rebuildHeader()
	tt.updateScrollBar()
	tt.rebuild()
	tt.MarkLayoutDirty()
}

// ---------------------------------------------------------------------------
// Internal: update
// ---------------------------------------------------------------------------

func (tt *TreeTable) update() {
	if tt.containsCursor() {
		_, wy := engine.Wheel()
		if wy != 0 {
			newPos := tt.scrollPos - wy*40
			tt.scrollBar.SetScrollPos(newPos)
		}
	}
}

// ---------------------------------------------------------------------------
// Internal: selection helpers
// ---------------------------------------------------------------------------

func (tt *TreeTable) selectedRowID() string {
	if tt.selected >= 0 && tt.selected < len(tt.flatList) {
		return tt.flatList[tt.selected].row.ID
	}
	return ""
}

func (tt *TreeTable) remapSelection(id string) {
	if id == "" {
		return
	}
	newIdx := tt.flatIndexOf(id)
	if newIdx < 0 {
		tt.selected = -1
	} else {
		tt.selected = newIdx
	}
}

// ---------------------------------------------------------------------------
// Internal: data management
// ---------------------------------------------------------------------------

func flattenTreeTableRows(rows []TreeTableRow, parentID string) []TreeTableRow {
	var out []TreeTableRow
	for _, r := range rows {
		flat := TreeTableRow{
			ID:       r.ID,
			ParentID: parentID,
			Cells:    r.Cells,
		}
		out = append(out, flat)
		if len(r.Children) > 0 {
			out = append(out, flattenTreeTableRows(r.Children, r.ID)...)
		}
	}
	return out
}

func (tt *TreeTable) hasChildren(parentID string) bool {
	for _, row := range tt.flatRows {
		if row.ParentID == parentID {
			return true
		}
	}
	return false
}

func (tt *TreeTable) childrenOf(parentID string) []TreeTableRow {
	var result []TreeTableRow
	for _, row := range tt.flatRows {
		if row.ParentID == parentID {
			result = append(result, row)
		}
	}
	return result
}

func (tt *TreeTable) rebuildFlatList() {
	tt.flatList = tt.flatList[:0]
	if tt.sortKey != "" {
		tt.flattenEntriesSorted("", 0)
	} else {
		tt.flattenEntries("", 0)
	}
}

func (tt *TreeTable) flattenEntries(parentID string, depth int) {
	children := tt.childrenOf(parentID)
	for _, child := range children {
		hasKids := tt.hasChildren(child.ID)
		expanded := tt.expanded[child.ID]
		tt.flatList = append(tt.flatList, treeTableFlatEntry{
			row: child, depth: depth, hasKids: hasKids, expanded: expanded,
		})
		if hasKids && expanded {
			tt.flattenEntries(child.ID, depth+1)
		}
	}
}

func (tt *TreeTable) flattenEntriesSorted(parentID string, depth int) {
	children := tt.childrenOf(parentID)
	key := tt.sortKey
	dir := tt.sortDir
	sort.SliceStable(children, func(i, j int) bool {
		a := children[i].Cells[key]
		b := children[j].Cells[key]
		cmp := strings.Compare(strings.ToLower(a), strings.ToLower(b))
		if dir == SortDirDesc {
			cmp = -cmp
		}
		return cmp < 0
	})
	for _, child := range children {
		hasKids := tt.hasChildren(child.ID)
		expanded := tt.expanded[child.ID]
		tt.flatList = append(tt.flatList, treeTableFlatEntry{
			row: child, depth: depth, hasKids: hasKids, expanded: expanded,
		})
		if hasKids && expanded {
			tt.flattenEntriesSorted(child.ID, depth+1)
		}
	}
}

func (tt *TreeTable) flatIndexOf(id string) int {
	for i, entry := range tt.flatList {
		if entry.row.ID == id {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Internal: header rendering
// ---------------------------------------------------------------------------

func (tt *TreeTable) rebuildHeader() {
	for _, c := range tt.headerCells {
		tt.headerNode.RemoveChild(c)
	}
	tt.headerCells = nil

	if len(tt.columns) == 0 {
		return
	}

	group := tt.EffectiveTheme().TreeTable.Group(tt.Variant())
	contentW := tt.width - ttScrollBarW

	headerBgColor := group.HeaderBg.Resolve(StateDefault).Color
	headerTextColor := group.HeaderText.Resolve(StateDefault)
	dividerColor := group.DividerColor.Resolve(StateDefault)
	sortColor := group.SortIndicator.Resolve(StateDefault)

	// Header background.
	headerBg := sg.NewSprite(tt.node.Name+"-header-bg", sg.TextureRegion{})
	headerBg.SetScale(contentW, ttHeaderHeight)
	headerBg.SetColor(headerBgColor)
	headerBg.SetZIndex(-1)
	tt.headerNode.AddChild(headerBg)
	tt.headerCells = append(tt.headerCells, headerBg)

	x := 0.0
	for colIdx, col := range tt.columns {
		colW := col.Width
		if colW <= 0 {
			colW = 100
		}

		// Header text.
		label := sg.NewText(tt.node.Name+"-hdr-lbl", col.Label, tt.font)
		label.TextBlock.FontSize = tt.displaySize
		label.TextBlock.Color = headerTextColor
		label.SetPosition(x+ttPadLeft, (ttHeaderHeight-tt.displaySize)/2)
		tt.headerNode.AddChild(label)
		tt.headerCells = append(tt.headerCells, label)

		// Sort indicator.
		if col.Sortable && col.Key == tt.sortKey {
			var glyph engine.Image
			if tt.sortDir == SortDirAsc {
				glyph = ttSortAscGlyph()
			} else {
				glyph = ttSortDescGlyph()
			}
			indicator := sg.NewSprite(tt.node.Name+"-sort-ind", sg.TextureRegion{})
			indicator.SetCustomImage(glyph)
			indicator.SetSize(9, 9)
			indicator.SetPosition(x+colW-16, (ttHeaderHeight-9)/2)
			indicator.SetColor(sortColor)
			tt.headerNode.AddChild(indicator)
			tt.headerCells = append(tt.headerCells, indicator)
		}

		// Column divider + resize handle.
		if colIdx < len(tt.columns)-1 {
			divider := sg.NewSprite(tt.node.Name+"-hdr-div", sg.TextureRegion{})
			divider.SetScale(1, ttHeaderHeight)
			divider.SetPosition(x+colW, 0)
			divider.SetColor(dividerColor)
			tt.headerNode.AddChild(divider)
			tt.headerCells = append(tt.headerCells, divider)

			handle := sg.NewSprite(tt.node.Name+"-resize-handle", sg.TextureRegion{})
			handle.SetScale(ttHandleWidth, ttHeaderHeight)
			handle.SetPosition(x+colW-ttHandleWidth/2, 0)
			handle.SetColor(headerBgColor)
			handle.Interactable = true
			handle.HitShape = sg.HitRect{X: 0, Y: 0, Width: ttHandleWidth, Height: ttHeaderHeight}

			ci := colIdx
			handle.OnDragStart(func(ctx sg.DragContext) {
				tt.resizing = true
				tt.resizeColIndex = ci
				tt.resizeStartX = ctx.GlobalX
				tt.resizeOrigW = tt.columns[ci].Width
			})
			handle.OnDrag(func(ctx sg.DragContext) {
				if !tt.resizing {
					return
				}
				delta := ctx.GlobalX - tt.resizeStartX
				newW := tt.resizeOrigW + delta
				if newW < 30 {
					newW = 30
				}
				tt.columns[tt.resizeColIndex].Width = newW
				tt.rebuild()
			})
			handle.OnDragEnd(func(_ sg.DragContext) {
				tt.resizing = false
				tt.rebuildHeader()
			})

			tt.headerNode.AddChild(handle)
			tt.headerCells = append(tt.headerCells, handle)
		}

		// Click to sort.
		if col.Sortable {
			sortArea := sg.NewSprite(tt.node.Name+"-sort-click", sg.TextureRegion{})
			sortArea.SetScale(colW, ttHeaderHeight)
			sortArea.SetPosition(x, 0)
			sortArea.SetColor(headerBgColor)
			sortArea.Interactable = true
			sortArea.HitShape = sg.HitRect{X: 0, Y: 0, Width: colW, Height: ttHeaderHeight}
			sortArea.SetZIndex(-2)

			colKey := col.Key
			sortArea.OnClick(func(_ sg.ClickContext) {
				if tt.sortKey == colKey {
					if tt.sortDir == SortDirAsc {
						tt.SetSortColumn(colKey, SortDirDesc)
					} else {
						tt.SetSortColumn(colKey, SortDirAsc)
					}
				} else {
					tt.SetSortColumn(colKey, SortDirAsc)
				}
			})

			tt.headerNode.AddChild(sortArea)
			tt.headerCells = append(tt.headerCells, sortArea)
		}

		x += colW
	}

	// Bottom border.
	border := sg.NewSprite(tt.node.Name+"-hdr-border", sg.TextureRegion{})
	border.SetScale(contentW, 1)
	border.SetPosition(0, ttHeaderHeight-1)
	border.SetColor(dividerColor)
	tt.headerNode.AddChild(border)
	tt.headerCells = append(tt.headerCells, border)
}

// ---------------------------------------------------------------------------
// Internal: row rendering (virtualized)
// ---------------------------------------------------------------------------

func (tt *TreeTable) visibleRange() (int, int) {
	viewH := tt.height - ttHeaderHeight
	first := int(tt.scrollPos/ttRowHeight) - 2
	if first < 0 {
		first = 0
	}
	last := int((tt.scrollPos+viewH)/ttRowHeight) + 3
	n := len(tt.flatList)
	if last > n {
		last = n
	}
	return first, last
}

func (tt *TreeTable) rebuild() {
	tt.clearRows()
	if len(tt.flatList) == 0 || len(tt.columns) == 0 {
		return
	}

	contentW := tt.width - ttScrollBarW
	tt.content.SetPosition(0, -tt.scrollPos)

	first, last := tt.visibleRange()
	for i := first; i < last; i++ {
		row := tt.createRow(i, tt.flatList[i], contentW)
		tt.rowNodes = append(tt.rowNodes, row)
	}

	tt.updateHighlight()
}

func (tt *TreeTable) createRow(flatIdx int, entry treeTableFlatEntry, contentW float64) *treeTableRowNode {
	group := tt.EffectiveTheme().TreeTable.Group(tt.Variant())
	row := &treeTableRowNode{flatIdx: flatIdx}

	row.container = sg.NewContainer(tt.node.Name + "-row")
	row.container.Interactable = true
	row.container.HitShape = sg.HitRect{X: 0, Y: 0, Width: contentW, Height: ttRowHeight}
	row.container.SetPosition(0, float64(flatIdx)*ttRowHeight)

	// Alternating background.
	var bgColor sg.Color
	if flatIdx%2 == 0 {
		bgColor = group.RowBg.Resolve(StateDefault).Color
	} else {
		bgColor = group.RowAltBg.Resolve(StateDefault).Color
	}
	bgSprite := sg.NewSprite(tt.node.Name+"-row-bg", sg.TextureRegion{})
	bgSprite.SetScale(contentW, ttRowHeight)
	bgSprite.SetColor(bgColor)
	bgSprite.SetZIndex(-2)
	row.container.AddChild(bgSprite)

	// Click for selection + callback.
	idx := flatIdx
	rowID := entry.row.ID
	row.container.OnClick(func(_ sg.ClickContext) {
		tt.SetSelected(idx)
		if tt.onRowClick != nil {
			tt.onRowClick(rowID)
		}
	})

	cellTextColor := group.CellText.Resolve(StateDefault)
	chevronColor := group.ChevronColor.Resolve(StateDefault)
	dividerColor := group.DividerColor.Resolve(StateDefault)

	// Render cells per column.
	x := 0.0
	for colIdx, col := range tt.columns {
		colW := col.Width
		if colW <= 0 {
			colW = 100
		}

		cellX := x + ttPadLeft

		// First column: indent + chevron.
		if colIdx == 0 {
			indent := float64(entry.depth) * ttIndentWidth
			cellX += indent

			if entry.hasKids {
				var glyph engine.Image
				if entry.expanded {
					glyph = ttCollapseGlyph()
				} else {
					glyph = ttExpandGlyph()
				}
				chevSprite := sg.NewSprite(tt.node.Name+"-chev", sg.TextureRegion{})
				chevSprite.SetCustomImage(glyph)
				b := glyph.Bounds()
				scaleX := ttChevronSize / float64(b.Dx())
				scaleY := ttChevronSize / float64(b.Dy())
				chevSprite.SetScale(scaleX, scaleY)
				chevSprite.SetPosition(cellX, (ttRowHeight-ttChevronSize)/2)
				chevSprite.SetColor(chevronColor)
				chevSprite.Interactable = true
				chevSprite.HitShape = sg.HitRect{X: 0, Y: 0, Width: ttChevronSize, Height: ttChevronSize}

				chevID := entry.row.ID
				chevSprite.OnClick(func(_ sg.ClickContext) {
					tt.SetExpanded(chevID, !tt.expanded[chevID])
				})
				row.container.AddChild(chevSprite)
			}

			cellX += ttChevronSize + 4
		}

		// Cell text.
		text := entry.row.Cells[col.Key]
		if text != "" {
			textNode := sg.NewText(tt.node.Name+"-cell", text, tt.font)
			textNode.TextBlock.FontSize = tt.displaySize
			textNode.TextBlock.Color = cellTextColor
			textNode.SetPosition(cellX, (ttRowHeight-tt.displaySize)/2)
			row.container.AddChild(textNode)
		}

		// Column divider.
		if colIdx < len(tt.columns)-1 {
			divider := sg.NewSprite(tt.node.Name+"-row-div", sg.TextureRegion{})
			divider.SetScale(1, ttRowHeight)
			divider.SetPosition(x+colW, 0)
			divider.SetColor(dividerColor)
			row.container.AddChild(divider)
		}

		x += colW
	}

	tt.content.AddChild(row.container)
	return row
}

func (tt *TreeTable) clearRows() {
	for _, row := range tt.rowNodes {
		tt.content.RemoveChild(row.container)
	}
	tt.rowNodes = nil
}

func (tt *TreeTable) updateHighlight() {
	group := tt.EffectiveTheme().TreeTable.Group(tt.Variant())
	idx := tt.selected
	n := len(tt.flatList)
	if idx < 0 || idx >= n {
		tt.selHighlight.SetVisible(false)
		return
	}
	itemW := tt.width - ttScrollBarW
	y := float64(idx) * ttRowHeight
	tt.selHighlight.SetPosition(0, y)
	tt.selHighlight.SetScale(itemW, ttRowHeight)
	tt.selHighlight.SetColor(group.RowSelectedBg.Resolve(StateDefault))
	tt.selHighlight.SetVisible(true)
}

func (tt *TreeTable) updateScrollBar() {
	n := len(tt.flatList)
	totalH := float64(n) * ttRowHeight
	bodyH := tt.height - ttHeaderHeight
	tt.scrollBar.SetContentSize(totalH, bodyH)
	tt.scrollBar.SetVisible(totalH > bodyH)
}

func (tt *TreeTable) scrollToIndex(idx int) {
	n := len(tt.flatList)
	if idx < 0 || idx >= n {
		return
	}
	itemTop := float64(idx) * ttRowHeight
	itemBottom := itemTop + ttRowHeight
	viewH := tt.height - ttHeaderHeight

	pos := tt.scrollPos
	if itemTop < pos {
		pos = itemTop
	} else if itemBottom > pos+viewH {
		pos = itemBottom - viewH
	}
	tt.scrollBar.SetScrollPos(pos)
}

// ---------------------------------------------------------------------------
// Procedural glyph images
// ---------------------------------------------------------------------------

var (
	ttExpandOnce   sync.Once
	ttExpandImg    engine.Image
	ttCollapseOnce sync.Once
	ttCollapseImg  engine.Image
	ttSortAscOnce  sync.Once
	ttSortAscImg   engine.Image
	ttSortDescOnce sync.Once
	ttSortDescImg  engine.Image
)

// ttExpandGlyph returns a 9x9 white right-pointing chevron ">".
func ttExpandGlyph() engine.Image {
	ttExpandOnce.Do(func() {
		const s = 9
		img := image.NewNRGBA(image.Rect(0, 0, s, s))
		px := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		for y := 0; y < s; y++ {
			x := y
			if y > s/2 {
				x = s - 1 - y
			}
			for dx := 0; dx < 3 && x+dx < s; dx++ {
				img.SetNRGBA(x+dx, y, px)
			}
		}
		ttExpandImg = engine.NewImageFromImage(img)
	})
	return ttExpandImg
}

// ttCollapseGlyph returns a 9x9 white downward-pointing chevron "v".
func ttCollapseGlyph() engine.Image {
	ttCollapseOnce.Do(func() {
		const s = 9
		img := image.NewNRGBA(image.Rect(0, 0, s, s))
		px := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		for x := 0; x < s; x++ {
			y := x
			if x > s/2 {
				y = s - 1 - x
			}
			for dy := 0; dy < 3 && y+dy < s; dy++ {
				img.SetNRGBA(x, y+dy, px)
			}
		}
		ttCollapseImg = engine.NewImageFromImage(img)
	})
	return ttCollapseImg
}

// ttSortAscGlyph returns a 9x9 white upward-pointing chevron "^".
func ttSortAscGlyph() engine.Image {
	ttSortAscOnce.Do(func() {
		const s = 9
		img := image.NewNRGBA(image.Rect(0, 0, s, s))
		px := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		for x := 0; x < s; x++ {
			y := s - 1 - x
			if x > s/2 {
				y = x
			}
			for dy := 0; dy < 3 && y-dy >= 0; dy++ {
				img.SetNRGBA(x, y-dy, px)
			}
		}
		ttSortAscImg = engine.NewImageFromImage(img)
	})
	return ttSortAscImg
}

// ttSortDescGlyph returns a 9x9 white downward-pointing chevron "v".
func ttSortDescGlyph() engine.Image {
	ttSortDescOnce.Do(func() {
		ttSortDescImg = ttCollapseGlyph()
	})
	return ttSortDescImg
}
