package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

// makeBox returns a Component with given width and height.
func makeFlowBox(name string, w, h float64) *ui.Component {
	c := ui.NewComponent(name)
	c.Width = w
	c.Height = h
	return c
}

// TestLayoutFlowWrap verifies wrapping occurs when the next child would exceed
// available width and no wrap occurs when the child exactly fits.
func TestLayoutFlowWrap(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 100
	p.Height = 200
	p.Spacing = 0

	a := makeFlowBox("a", 60, 20)
	b := makeFlowBox("b", 40, 20) // a + b = 100, exactly fits
	c := makeFlowBox("c", 10, 20) // would exceed → wrap

	p.AddChild(a)
	p.AddChild(b)
	p.AddChild(c)
	p.UpdateLayout()

	// a and b should be on row 0 (y=0)
	if a.Y != 0 {
		t.Errorf("a.Y = %v, want 0", a.Y)
	}
	if b.Y != 0 {
		t.Errorf("b.Y = %v, want 0", b.Y)
	}
	// b.X should be 60 (after a)
	if b.X != 60 {
		t.Errorf("b.X = %v, want 60", b.X)
	}
	// c should wrap to the next row
	if c.Y != 20 {
		t.Errorf("c.Y = %v, want 20 (wrapped row)", c.Y)
	}
	if c.X != 0 {
		t.Errorf("c.X = %v, want 0 (start of new row)", c.X)
	}
}

// TestLayoutFlowOversizedChild verifies a single child wider than available
// width occupies its own row.
func TestLayoutFlowOversizedChild(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 80
	p.Height = 200
	p.Spacing = 0

	a := makeFlowBox("a", 100, 20) // wider than parent
	b := makeFlowBox("b", 30, 20)

	p.AddChild(a)
	p.AddChild(b)
	p.UpdateLayout()

	if a.Y != 0 {
		t.Errorf("a.Y = %v, want 0", a.Y)
	}
	// b must be on the next row since a is oversized
	if b.Y != 20 {
		t.Errorf("b.Y = %v, want 20", b.Y)
	}
}

// TestLayoutFlowJustifyCenter verifies each row is centered.
func TestLayoutFlowJustifyCenter(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 100
	p.Height = 200
	p.Spacing = 0
	p.Justify = ui.AlignCenter

	a := makeFlowBox("a", 40, 20) // single item, width=40 → row.width=40 → startX=(100-40)/2=30

	p.AddChild(a)
	p.UpdateLayout()

	if a.X != 30 {
		t.Errorf("a.X = %v, want 30 (centered)", a.X)
	}
}

// TestLayoutFlowJustifyEnd verifies each row is right-aligned.
func TestLayoutFlowJustifyEnd(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 100
	p.Height = 200
	p.Spacing = 0
	p.Justify = ui.AlignEnd

	a := makeFlowBox("a", 40, 20) // row.width=40 → startX=100-40=60

	p.AddChild(a)
	p.UpdateLayout()

	if a.X != 60 {
		t.Errorf("a.X = %v, want 60 (end-aligned)", a.X)
	}
}

// TestLayoutFlowAlignCenter verifies children are vertically centered in the row.
func TestLayoutFlowAlignCenter(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 200
	p.Height = 200
	p.Spacing = 0
	p.Align = ui.AlignCenter

	a := makeFlowBox("a", 40, 20) // rowHeight=40 (tallest is b), center y = (40-20)/2=10
	b := makeFlowBox("b", 40, 40) // rowHeight=40, center y = (40-40)/2=0

	p.AddChild(a)
	p.AddChild(b)
	p.UpdateLayout()

	if a.Y != 10 {
		t.Errorf("a.Y = %v, want 10 (vertically centered in row)", a.Y)
	}
	if b.Y != 0 {
		t.Errorf("b.Y = %v, want 0 (tallest item, fills row)", b.Y)
	}
}

// TestLayoutFlowAlignEnd verifies children are bottom-aligned in the row.
func TestLayoutFlowAlignEnd(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 200
	p.Height = 200
	p.Spacing = 0
	p.Align = ui.AlignEnd

	a := makeFlowBox("a", 40, 20) // rowHeight=40, end y = 40-20=20
	b := makeFlowBox("b", 40, 40) // rowHeight=40, end y = 40-40=0

	p.AddChild(a)
	p.AddChild(b)
	p.UpdateLayout()

	if a.Y != 20 {
		t.Errorf("a.Y = %v, want 20 (bottom-aligned)", a.Y)
	}
	if b.Y != 0 {
		t.Errorf("b.Y = %v, want 0 (tallest item fills row)", b.Y)
	}
}

// TestLayoutFlowMargins verifies margins are included in wrap and placement.
func TestLayoutFlowMargins(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 100
	p.Height = 200
	p.Spacing = 0

	a := makeFlowBox("a", 60, 20)
	a.Margin = ui.Insets{Left: 5, Right: 5} // outer width = 70
	b := makeFlowBox("b", 30, 20)           // outer width = 30 → 70+30=100 exactly fits

	p.AddChild(a)
	p.AddChild(b)
	p.UpdateLayout()

	// a.X = margin.Left = 5
	if a.X != 5 {
		t.Errorf("a.X = %v, want 5", a.X)
	}
	// b.X = 5(a.margin.left) + 60(a.width) + 5(a.margin.right) = 70
	if b.X != 70 {
		t.Errorf("b.X = %v, want 70", b.X)
	}
	// both on row 0
	if a.Y != 0 || b.Y != 0 {
		t.Errorf("expected both on row 0, got a.Y=%v b.Y=%v", a.Y, b.Y)
	}
}

// TestLayoutFlowInvisibleChildrenSkipped verifies invisible children are ignored.
func TestLayoutFlowInvisibleChildrenSkipped(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 100
	p.Height = 200
	p.Spacing = 0

	a := makeFlowBox("a", 40, 20)
	hidden := makeFlowBox("hidden", 40, 20)
	hidden.SetVisible(false)
	b := makeFlowBox("b", 40, 20)

	p.AddChild(a)
	p.AddChild(hidden)
	p.AddChild(b)
	p.UpdateLayout()

	// a and b should both be on row 0, adjacent
	if a.X != 0 {
		t.Errorf("a.X = %v, want 0", a.X)
	}
	if b.X != 40 {
		t.Errorf("b.X = %v, want 40 (hidden skipped)", b.X)
	}
	if a.Y != 0 || b.Y != 0 {
		t.Errorf("expected both on row 0, got a.Y=%v b.Y=%v", a.Y, b.Y)
	}
}

// TestLayoutFlowRowGap verifies FlowRowGap controls vertical row spacing.
func TestLayoutFlowRowGap(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 50
	p.Height = 200
	p.Spacing = 4
	p.FlowRowGap = 10

	a := makeFlowBox("a", 50, 20) // row 0 height=20
	b := makeFlowBox("b", 50, 20) // row 1, y = 20 + rowGap(10) = 30

	p.AddChild(a)
	p.AddChild(b)
	p.UpdateLayout()

	if a.Y != 0 {
		t.Errorf("a.Y = %v, want 0", a.Y)
	}
	if b.Y != 30 {
		t.Errorf("b.Y = %v, want 30 (rowGap=10)", b.Y)
	}
}

// TestLayoutFlowRowGapFallbackToSpacing verifies FlowRowGap=0 uses Spacing.
func TestLayoutFlowRowGapFallbackToSpacing(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 50
	p.Height = 200
	p.Spacing = 8
	p.FlowRowGap = 0 // should fall back to Spacing

	a := makeFlowBox("a", 50, 20)
	b := makeFlowBox("b", 50, 20) // y = 20 + spacing(8) = 28

	p.AddChild(a)
	p.AddChild(b)
	p.UpdateLayout()

	if b.Y != 28 {
		t.Errorf("b.Y = %v, want 28 (FlowRowGap falls back to Spacing)", b.Y)
	}
}

// TestLayoutFlowSizeToContentFixedWidth verifies SizeToContent computes the
// required height for a wrapping container with a fixed width.
func TestLayoutFlowSizeToContentFixedWidth(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 100
	p.Spacing = 0

	a := makeFlowBox("a", 60, 20)
	b := makeFlowBox("b", 60, 20) // wraps → row 1

	p.AddChild(a)
	p.AddChild(b)
	p.SizeToContent()

	// two rows of 20 each → total height = 40
	if p.Height != 40 {
		t.Errorf("p.Height = %v, want 40", p.Height)
	}
}

// TestLayoutFlowSizeToContentZeroWidth verifies SizeToContent with zero width
// treats items as a single row and computes natural dimensions.
func TestLayoutFlowSizeToContentZeroWidth(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 0
	p.Spacing = 4

	a := makeFlowBox("a", 30, 20)
	b := makeFlowBox("b", 50, 40)

	p.AddChild(a)
	p.AddChild(b)
	p.SizeToContent()

	// width = 30 + 4 + 50 = 84, height = max(20,40) = 40
	if p.Width != 84 {
		t.Errorf("p.Width = %v, want 84", p.Width)
	}
	if p.Height != 40 {
		t.Errorf("p.Height = %v, want 40", p.Height)
	}
}

// TestLayoutFlowPadding verifies padding offsets child positions.
func TestLayoutFlowPadding(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutFlow
	p.Width = 200
	p.Height = 200
	p.Spacing = 0
	p.Padding = ui.Insets{Top: 10, Left: 8, Right: 8, Bottom: 10}

	a := makeFlowBox("a", 40, 20)

	p.AddChild(a)
	p.UpdateLayout()

	if a.X != 8 {
		t.Errorf("a.X = %v, want 8 (padding.left)", a.X)
	}
	if a.Y != 10 {
		t.Errorf("a.Y = %v, want 10 (padding.top)", a.Y)
	}
}

// TestLayoutFlowXMLParsing verifies the template parser accepts layout="flow"
// and flow-row-gap.
func TestLayoutFlowXMLParsing(t *testing.T) {
	resetScheduler()

	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	src := `<Panel layout="flow" spacing="6" flow-row-gap="10" width="200" height="200"/>`
	if err := reg.RegisterXML("test", []byte(src)); err != nil {
		t.Fatalf("RegisterXML: %v", err)
	}

	ctrl := &xmlTestController{provider: &testDataProvider{refs: map[string]any{}}}
	comp, err := reg.Instantiate("test", ctrl, nil)
	if err != nil {
		t.Fatalf("Instantiate: %v", err)
	}
	defer comp.Dispose()

	if comp.Layout != ui.LayoutFlow {
		t.Errorf("Layout = %v, want LayoutFlow", comp.Layout)
	}
	if comp.Spacing != 6 {
		t.Errorf("Spacing = %v, want 6", comp.Spacing)
	}
	if comp.FlowRowGap != 10 {
		t.Errorf("FlowRowGap = %v, want 10", comp.FlowRowGap)
	}
}
