package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

// --- LayoutAnchor mode on plain Component ---

func TestLayoutAnchorPositioning(t *testing.T) {
	resetScheduler()
	c := ui.NewComponent("container")
	defer c.Dispose()

	c.Layout = ui.LayoutAnchor
	c.Width = 400
	c.Height = 300
	c.MarkLayoutDirty()

	tests := []struct {
		name   string
		anchor ui.Anchor
		wantX  float64
		wantY  float64
	}{
		{"top-left", ui.AnchorTopLeft, 0, 0},
		{"top-center", ui.AnchorTopCenter, 175, 0},
		{"top-right", ui.AnchorTopRight, 350, 0},
		{"middle-left", ui.AnchorMiddleLeft, 0, 125},
		{"center", ui.AnchorCenter, 175, 125},
		{"middle-right", ui.AnchorMiddleRight, 350, 125},
		{"bottom-left", ui.AnchorBottomLeft, 0, 250},
		{"bottom-center", ui.AnchorBottomCenter, 175, 250},
		{"bottom-right", ui.AnchorBottomRight, 350, 250},
	}

	children := make([]*ui.Component, len(tests))
	for i, tt := range tests {
		child := ui.NewComponent(tt.name)
		child.Width = 50
		child.Height = 50
		c.AddAnchoredChild(child, tt.anchor, 0, 0)
		children[i] = child
	}

	c.UpdateLayout()

	for i, tt := range tests {
		child := children[i]
		if child.X != tt.wantX || child.Y != tt.wantY {
			t.Errorf("%s: got (%g, %g), want (%g, %g)", tt.name, child.X, child.Y, tt.wantX, tt.wantY)
		}
	}
}

func TestLayoutAnchorWithPadding(t *testing.T) {
	resetScheduler()
	c := ui.NewComponent("container")
	defer c.Dispose()

	c.Layout = ui.LayoutAnchor
	c.Width = 400
	c.Height = 300
	c.Padding = ui.Insets{Top: 10, Right: 20, Bottom: 30, Left: 40}
	c.MarkLayoutDirty()

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50

	// Bottom-right: available = (400-40-20)x(300-10-30) = 340x260
	// Position = padding + (340-50, 260-50) = (40+290, 10+210) = (330, 220)
	c.AddAnchoredChild(child, ui.AnchorBottomRight, 0, 0)
	c.UpdateLayout()

	wantX := 40 + (340 - 50)
	wantY := 10 + (260 - 50)
	if child.X != float64(wantX) || child.Y != float64(wantY) {
		t.Errorf("got (%g, %g), want (%d, %d)", child.X, child.Y, wantX, wantY)
	}
}

func TestLayoutAnchorOffsets(t *testing.T) {
	resetScheduler()
	c := ui.NewComponent("container")
	defer c.Dispose()

	c.Layout = ui.LayoutAnchor
	c.Width = 400
	c.Height = 300
	c.MarkLayoutDirty()

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50

	c.AddAnchoredChild(child, ui.AnchorTopLeft, 15, 25)
	c.UpdateLayout()

	if child.X != 15 || child.Y != 25 {
		t.Errorf("got (%g, %g), want (15, 25)", child.X, child.Y)
	}
}

func TestLayoutAnchorSetAnchor(t *testing.T) {
	resetScheduler()
	c := ui.NewComponent("container")
	defer c.Dispose()

	c.Layout = ui.LayoutAnchor
	c.Width = 200
	c.Height = 200
	c.MarkLayoutDirty()

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50

	c.AddAnchoredChild(child, ui.AnchorTopLeft, 0, 0)
	c.UpdateLayout()

	if child.X != 0 || child.Y != 0 {
		t.Errorf("before SetAnchor: got (%g, %g), want (0, 0)", child.X, child.Y)
	}

	c.SetAnchor(child, ui.AnchorBottomRight, 0, 0)
	c.UpdateLayout()

	if child.X != 150 || child.Y != 150 {
		t.Errorf("after SetAnchor: got (%g, %g), want (150, 150)", child.X, child.Y)
	}
}

func TestLayoutAnchorAnchorOf(t *testing.T) {
	resetScheduler()
	c := ui.NewComponent("container")
	defer c.Dispose()

	c.Layout = ui.LayoutAnchor

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50

	c.AddAnchoredChild(child, ui.AnchorCenter, 5, 10)

	anchor, ox, oy, ok := c.AnchorOf(child)
	if !ok {
		t.Fatal("AnchorOf should return ok=true for anchored child")
	}
	if anchor != ui.AnchorCenter {
		t.Errorf("anchor: got %d, want AnchorCenter", anchor)
	}
	if ox != 5 || oy != 10 {
		t.Errorf("offsets: got (%g, %g), want (5, 10)", ox, oy)
	}
}

func TestLayoutAnchorRemoveChild(t *testing.T) {
	resetScheduler()
	c := ui.NewComponent("container")
	defer c.Dispose()

	c.Layout = ui.LayoutAnchor

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50

	c.AddAnchoredChild(child, ui.AnchorCenter, 0, 0)
	if c.NumChildren() != 1 {
		t.Fatalf("expected 1 child, got %d", c.NumChildren())
	}

	c.RemoveChild(child)
	if c.NumChildren() != 0 {
		t.Fatalf("expected 0 children after remove, got %d", c.NumChildren())
	}

	// AnchorOf should no longer find the child.
	_, _, _, ok := c.AnchorOf(child)
	if ok {
		t.Error("AnchorOf should return ok=false after child removed")
	}
}

func TestLayoutAnchorInvisibleChildSkipped(t *testing.T) {
	resetScheduler()
	c := ui.NewComponent("container")
	defer c.Dispose()

	c.Layout = ui.LayoutAnchor
	c.Width = 200
	c.Height = 200
	c.MarkLayoutDirty()

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50
	child.SetVisible(false)

	c.AddAnchoredChild(child, ui.AnchorBottomRight, 0, 0)
	c.UpdateLayout()

	// X/Y should be unset (zero) since the child was skipped.
	if child.X != 0 || child.Y != 0 {
		t.Errorf("invisible child: got (%g, %g), want (0, 0)", child.X, child.Y)
	}
}

func TestLayoutAnchorDefaultAddChild(t *testing.T) {
	resetScheduler()
	c := ui.NewComponent("container")
	defer c.Dispose()

	c.Layout = ui.LayoutAnchor
	c.Width = 200
	c.Height = 200
	c.MarkLayoutDirty()

	// Children added via AddChild (no explicit anchor) default to AnchorTopLeft.
	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50
	c.AddChild(child)
	c.UpdateLayout()

	if child.X != 0 || child.Y != 0 {
		t.Errorf("default AddChild: got (%g, %g), want (0, 0)", child.X, child.Y)
	}
}

func TestLayoutAnchorResize(t *testing.T) {
	resetScheduler()
	c := ui.NewComponent("container")
	defer c.Dispose()

	c.Layout = ui.LayoutAnchor
	c.Width = 200
	c.Height = 200
	c.MarkLayoutDirty()

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50
	c.AddAnchoredChild(child, ui.AnchorBottomRight, 0, 0)
	c.UpdateLayout()

	if child.X != 150 || child.Y != 150 {
		t.Errorf("before resize: got (%g, %g), want (150, 150)", child.X, child.Y)
	}

	c.Width = 400
	c.Height = 400
	c.MarkLayoutDirty()
	c.UpdateLayout()

	if child.X != 350 || child.Y != 350 {
		t.Errorf("after resize: got (%g, %g), want (350, 350)", child.X, child.Y)
	}
}

// --- Migration tests: AnchorLayout wrapper still works ---

func TestAnchorLayoutWrapperIsLayoutAnchor(t *testing.T) {
	resetScheduler()
	al := ui.NewAnchorLayout("al")
	defer al.Dispose()

	if al.Layout != ui.LayoutAnchor {
		t.Errorf("AnchorLayout.Layout = %d, want LayoutAnchor (%d)", al.Layout, ui.LayoutAnchor)
	}
}

func TestAnchorLayoutWrapperPositioning(t *testing.T) {
	resetScheduler()
	al := ui.NewAnchorLayout("al")
	defer al.Dispose()

	al.Padding = ui.Insets{} // zero padding for predictable positions
	al.SetSize(200, 200)

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50
	al.AddAnchoredChild(child, ui.AnchorBottomRight, 0, 0)
	al.UpdateLayout()

	if child.X != 150 || child.Y != 150 {
		t.Errorf("got (%g, %g), want (150, 150)", child.X, child.Y)
	}
}

func TestAnchorLayoutWrapperSetAnchor(t *testing.T) {
	resetScheduler()
	al := ui.NewAnchorLayout("al")
	defer al.Dispose()

	al.SetSize(200, 200)

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50
	al.AddAnchoredChild(child, ui.AnchorTopLeft, 0, 0)
	al.UpdateLayout()

	al.SetAnchor(child, ui.AnchorCenter, 0, 0)
	al.UpdateLayout()

	if child.X != 75 || child.Y != 75 {
		t.Errorf("after SetAnchor: got (%g, %g), want (75, 75)", child.X, child.Y)
	}
}
