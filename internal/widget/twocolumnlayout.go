package widget

import "github.com/devthicket/willowui/internal/sg"

// rowEntry stores the left and right components of a two-column row.
// When span is true, left holds a child that fills the full available width
// and right is nil (clear:both / full-width row).
type rowEntry struct {
	left  *Component
	right *Component // nil when span is true
	span  bool
}

// TwoColumnLayout arranges children in labeled two-column rows. Each column
// is independently aligned. The primary use case is settings screens, stat
// sheets, and form dialogs where the left column contains labels and the
// right column contains input widgets.
//
// Children are added in pairs via AddRow. The left column defaults to
// right-aligned (AlignEnd) for labels; the right column defaults to
// left-aligned (AlignStart) for values and inputs.
type TwoColumnLayout struct {
	Panel
	rows []rowEntry

	// LeftAlign controls alignment within the left column.
	// Default: AlignEnd (right-aligned, suits labels).
	LeftAlign Alignment

	// RightAlign controls alignment within the right column.
	// Default: AlignStart (left-aligned, suits inputs).
	RightAlign Alignment

	// LeftWidth sets a fixed pixel width for the left column.
	// When zero, ColumnRatio or an even split determines the width.
	LeftWidth float64

	// ColumnRatio sets the fraction of available width (after Gap) given to the
	// left column. 0.3 means left=30%, right=70%. Only used when LeftWidth is 0.
	// When both are zero, the columns split evenly (ratio 0.5).
	ColumnRatio float64

	// Gap is the horizontal space between the two columns.
	Gap float64

	// RowSpacing is the vertical space between rows.
	RowSpacing float64
}

// NewTwoColumnLayout creates a TwoColumnLayout container with sensible defaults:
// left column right-aligned, right column left-aligned, 12px gap, 8px row spacing.
func NewTwoColumnLayout(name string) *TwoColumnLayout {
	tl := &TwoColumnLayout{
		LeftAlign:  AlignEnd,
		RightAlign: AlignStart,
		Gap:        12,
		RowSpacing: 8,
	}
	initComponent(&tl.Component, name)
	tl.initBackground(name)
	tl.initBorder(name)

	tl.onThemeChange = func() { tl.applyThemeColors() }
	tl.applyThemeColors()

	tl.onLayout = func() { tl.positionRows() }

	return tl
}

// AddRow adds a row. When both left and right are provided, they are placed in
// their respective columns. When right is nil, left spans the full available
// width across both columns (clear:both equivalent).
func (tl *TwoColumnLayout) AddRow(left, right UIElement) {
	if left == nil {
		return
	}
	tl.Component.AddChild(left)
	if right != nil {
		tl.Component.AddChild(right)
		tl.rows = append(tl.rows, rowEntry{left: left.base(), right: right.base()})
	} else {
		tl.rows = append(tl.rows, rowEntry{left: left.base(), span: true})
	}
}

// RemoveRow removes the row containing left (or the spanning child), detaching it.
func (tl *TwoColumnLayout) RemoveRow(left UIElement) {
	if left == nil {
		return
	}
	lc := left.base()
	for i, row := range tl.rows {
		if row.left == lc {
			tl.Component.RemoveChild(row.left)
			if row.right != nil {
				tl.Component.RemoveChild(row.right)
			}
			tl.rows = append(tl.rows[:i], tl.rows[i+1:]...)
			return
		}
	}
}

// positionRows sets each child's X/Y based on the column widths and alignments.
// Called from the onLayout hook inside UpdateLayout; layoutNone syncs the
// positions to willow nodes immediately after.
func (tl *TwoColumnLayout) positionRows() {
	padL := tl.Padding.Left
	padT := tl.Padding.Top
	availW := tl.Width - padL - tl.Padding.Right

	leftW := tl.resolveLeftWidth(availW)
	rightW := availW - leftW - tl.Gap
	rightX := padL + leftW + tl.Gap

	y := padT
	firstVisible := true
	for _, row := range tl.rows {
		lc := row.left
		if !lc.IsVisible() && (row.right == nil || !row.right.IsVisible()) {
			continue
		}
		if !firstVisible {
			y += tl.RowSpacing
		}
		firstVisible = false

		if row.span {
			// Full-width row: child stretches across both columns.
			lc.X = padL
			lc.Y = y
			lc.Width = availW
			y += lc.Height
			continue
		}

		rc := row.right
		rowH := 0.0
		if lc.IsVisible() && lc.Height > rowH {
			rowH = lc.Height
		}
		if rc != nil && rc.IsVisible() && rc.Height > rowH {
			rowH = rc.Height
		}

		if lc.IsVisible() {
			lc.Y = y + (rowH-lc.Height)/2
			switch tl.LeftAlign {
			case AlignCenter:
				lc.X = padL + (leftW-lc.Width)/2
			case AlignEnd:
				lc.X = padL + leftW - lc.Width
			default:
				lc.X = padL
			}
		}

		if rc != nil && rc.IsVisible() {
			rc.Y = y + (rowH-rc.Height)/2
			switch tl.RightAlign {
			case AlignCenter:
				rc.X = rightX + (rightW-rc.Width)/2
			case AlignEnd:
				rc.X = rightX + rightW - rc.Width
			default:
				rc.X = rightX
			}
		}

		y += rowH
	}
}

// SizeToContent resizes the layout to tightly wrap all rows. Children must
// have their sizes set before calling this. UpdateLayout is called internally.
func (tl *TwoColumnLayout) SizeToContent() {
	// When ColumnRatio is set, derive widths from the ratio rather than
	// measuring children — the ratio implies a fixed relationship, so we
	// measure one side and compute the other.
	var leftW, rightW float64
	if tl.LeftWidth > 0 {
		leftW = tl.LeftWidth
		for _, row := range tl.rows {
			if row.right != nil && row.right.IsVisible() && row.right.Width > rightW {
				rightW = row.right.Width
			}
		}
	} else if tl.ColumnRatio > 0 {
		// Measure the wider of the two sides and back-calculate total.
		var maxLeft, maxRight float64
		for _, row := range tl.rows {
			if !row.span && row.left.IsVisible() && row.left.Width > maxLeft {
				maxLeft = row.left.Width
			}
			if row.right != nil && row.right.IsVisible() && row.right.Width > maxRight {
				maxRight = row.right.Width
			}
		}
		ratio := tl.ColumnRatio
		// Choose whichever side constrains total width more.
		totalFromLeft := (maxLeft / ratio) + tl.Gap
		totalFromRight := (maxRight / (1 - ratio)) + tl.Gap
		total := totalFromLeft
		if totalFromRight > total {
			total = totalFromRight
		}
		leftW = (total - tl.Gap) * ratio
		rightW = total - tl.Gap - leftW
	} else {
		for _, row := range tl.rows {
			if !row.span && row.left.IsVisible() && row.left.Width > leftW {
				leftW = row.left.Width
			}
			if row.right != nil && row.right.IsVisible() && row.right.Width > rightW {
				rightW = row.right.Width
			}
		}
	}

	// Span rows may also constrain the minimum total width.
	var minSpanW float64
	for _, row := range tl.rows {
		if row.span && row.left.IsVisible() && row.left.Width > minSpanW {
			minSpanW = row.left.Width
		}
	}

	totalH := tl.Padding.Vertical()
	firstVisible := true
	for _, row := range tl.rows {
		lc := row.left
		if !lc.IsVisible() && (row.right == nil || !row.right.IsVisible()) {
			continue
		}
		if !firstVisible {
			totalH += tl.RowSpacing
		}
		firstVisible = false

		rowH := lc.Height
		if !row.span && row.right != nil && row.right.IsVisible() && row.right.Height > rowH {
			rowH = row.right.Height
		}
		totalH += rowH
	}

	totalW := tl.Padding.Horizontal() + leftW + tl.Gap + rightW
	if minSpanW+tl.Padding.Horizontal() > totalW {
		totalW = minSpanW + tl.Padding.Horizontal()
	}
	tl.Width = totalW
	tl.Height = totalH
	tl.resizeBackground(totalW, totalH)
	tl.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: totalW, Height: totalH}
	tl.resizeBorder(totalW, totalH)
	tl.MarkLayoutDirty()
	tl.UpdateLayout()
}

// resolveLeftWidth returns the pixel width of the left column given the
// available width (padding already subtracted). Priority: LeftWidth > ColumnRatio > even split.
func (tl *TwoColumnLayout) resolveLeftWidth(availW float64) float64 {
	if tl.LeftWidth > 0 {
		return tl.LeftWidth
	}
	if tl.ColumnRatio > 0 {
		return (availW - tl.Gap) * tl.ColumnRatio
	}
	return (availW - tl.Gap) / 2
}
