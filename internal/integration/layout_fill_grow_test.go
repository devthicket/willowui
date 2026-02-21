package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func makeBox(name string, w, h float64) *ui.Component {
	c := ui.NewComponent(name)
	c.Width = w
	c.Height = h
	return c
}

// --- Fill tests ---

func TestFillWidthInVBox(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutVBox
	p.Width = 200
	p.Height = 100

	child := makeBox("child", 50, 30)
	child.Fill = ui.FillWidth
	p.AddChild(child)
	p.UpdateLayout()

	if child.Width != 200 {
		t.Errorf("child.Width = %v, want 200 (parent width)", child.Width)
	}
	if child.Height != 30 {
		t.Errorf("child.Height = %v, want 30 (unchanged)", child.Height)
	}
}

func TestFillHeightInHBox(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutHBox
	p.Width = 200
	p.Height = 100

	child := makeBox("child", 50, 30)
	child.Fill = ui.FillHeight
	p.AddChild(child)
	p.UpdateLayout()

	if child.Width != 50 {
		t.Errorf("child.Width = %v, want 50 (unchanged)", child.Width)
	}
	if child.Height != 100 {
		t.Errorf("child.Height = %v, want 100 (parent height)", child.Height)
	}
}

func TestFillBothInNone(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutNone
	p.Width = 300
	p.Height = 200

	child := makeBox("child", 10, 10)
	child.Fill = ui.FillBoth
	p.AddChild(child)
	p.UpdateLayout()

	if child.Width != 300 {
		t.Errorf("child.Width = %v, want 300", child.Width)
	}
	if child.Height != 200 {
		t.Errorf("child.Height = %v, want 200", child.Height)
	}
}

func TestFillRespectsMargin(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutVBox
	p.Width = 200
	p.Height = 100

	child := makeBox("child", 50, 30)
	child.Fill = ui.FillWidth
	child.Margin = ui.Insets{Left: 10, Right: 10}
	p.AddChild(child)
	p.UpdateLayout()

	if child.Width != 180 {
		t.Errorf("child.Width = %v, want 180 (200 - 10 - 10)", child.Width)
	}
}

func TestFillRespectsPadding(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutVBox
	p.Width = 200
	p.Height = 100
	p.Padding = ui.Insets{Left: 20, Right: 20}

	child := makeBox("child", 50, 30)
	child.Fill = ui.FillWidth
	p.AddChild(child)
	p.UpdateLayout()

	if child.Width != 160 {
		t.Errorf("child.Width = %v, want 160 (200 - 20 - 20)", child.Width)
	}
}

// --- Grow tests ---

func TestGrowEqualWeightInHBox(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutHBox
	p.Width = 300
	p.Height = 50

	a := makeBox("a", 0, 30)
	a.Grow = 1
	b := makeBox("b", 0, 30)
	b.Grow = 1
	c := makeBox("c", 0, 30)
	c.Grow = 1

	p.AddChild(a)
	p.AddChild(b)
	p.AddChild(c)
	p.UpdateLayout()

	if a.Width != 100 {
		t.Errorf("a.Width = %v, want 100", a.Width)
	}
	if b.Width != 100 {
		t.Errorf("b.Width = %v, want 100", b.Width)
	}
	if c.Width != 100 {
		t.Errorf("c.Width = %v, want 100", c.Width)
	}
}

func TestGrowWeightedInHBox(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutHBox
	p.Width = 300
	p.Height = 50

	a := makeBox("a", 0, 30)
	a.Grow = 1
	b := makeBox("b", 0, 30)
	b.Grow = 2

	p.AddChild(a)
	p.AddChild(b)
	p.UpdateLayout()

	if a.Width != 100 {
		t.Errorf("a.Width = %v, want 100 (1/3 of 300)", a.Width)
	}
	if b.Width != 200 {
		t.Errorf("b.Width = %v, want 200 (2/3 of 300)", b.Width)
	}
}

func TestGrowWithFixedChildInHBox(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutHBox
	p.Width = 300
	p.Height = 50

	fixed := makeBox("fixed", 100, 30) // fixed width
	flex := makeBox("flex", 0, 30)
	flex.Grow = 1

	p.AddChild(fixed)
	p.AddChild(flex)
	p.UpdateLayout()

	if fixed.Width != 100 {
		t.Errorf("fixed.Width = %v, want 100", fixed.Width)
	}
	if flex.Width != 200 {
		t.Errorf("flex.Width = %v, want 200 (300 - 100)", flex.Width)
	}
}

func TestGrowInVBox(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutVBox
	p.Width = 100
	p.Height = 300

	fixed := makeBox("fixed", 80, 100)
	flex := makeBox("flex", 80, 0)
	flex.Grow = 1

	p.AddChild(fixed)
	p.AddChild(flex)
	p.UpdateLayout()

	if fixed.Height != 100 {
		t.Errorf("fixed.Height = %v, want 100", fixed.Height)
	}
	if flex.Height != 200 {
		t.Errorf("flex.Height = %v, want 200 (300 - 100)", flex.Height)
	}
}

func TestGrowWithSpacingInHBox(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutHBox
	p.Width = 310
	p.Height = 50
	p.Spacing = 10

	fixed := makeBox("fixed", 100, 30)
	flex := makeBox("flex", 0, 30)
	flex.Grow = 1

	p.AddChild(fixed)
	p.AddChild(flex)
	p.UpdateLayout()

	// remaining = 310 - 100 (fixed) - 10 (spacing) = 200
	if flex.Width != 200 {
		t.Errorf("flex.Width = %v, want 200 (310 - 100 - 10 spacing)", flex.Width)
	}
}

func TestGrowNoGrowChildrenUnchanged(t *testing.T) {
	resetScheduler()
	p := ui.NewComponent("parent")
	defer p.Dispose()

	p.Layout = ui.LayoutHBox
	p.Width = 300
	p.Height = 50

	a := makeBox("a", 80, 30)
	b := makeBox("b", 120, 30)

	p.AddChild(a)
	p.AddChild(b)
	p.UpdateLayout()

	if a.Width != 80 {
		t.Errorf("a.Width = %v, want 80 (no grow, unchanged)", a.Width)
	}
	if b.Width != 120 {
		t.Errorf("b.Width = %v, want 120 (no grow, unchanged)", b.Width)
	}
}
