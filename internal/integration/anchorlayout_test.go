package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func TestNewAnchorLayoutDefaults(t *testing.T) {
	resetScheduler()
	al := ui.NewAnchorLayout("al")
	defer al.Dispose()

	if al.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if al.BgNode() == nil {
		t.Fatal("bgNode should not be nil")
	}
}

func TestAnchorLayoutPositioning(t *testing.T) {
	resetScheduler()
	al := ui.NewAnchorLayout("al")
	defer al.Dispose()

	al.Padding = ui.Insets{} // zero padding for predictable positions
	al.SetSize(400, 300)

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
		c := ui.NewComponent(tt.name)
		c.Width = 50
		c.Height = 50
		al.AddAnchoredChild(c, tt.anchor, 0, 0)
		children[i] = c
	}

	al.UpdateLayout()

	for i, tt := range tests {
		c := children[i]
		if c.X != tt.wantX || c.Y != tt.wantY {
			t.Errorf("%s: got (%g, %g), want (%g, %g)", tt.name, c.X, c.Y, tt.wantX, tt.wantY)
		}
	}
}

func TestAnchorLayoutWithPadding(t *testing.T) {
	resetScheduler()
	al := ui.NewAnchorLayout("al")
	defer al.Dispose()

	al.SetSize(400, 300)
	al.Padding = ui.Insets{Top: 10, Right: 20, Bottom: 30, Left: 40}

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50

	// Bottom-right with padding: available area = (400-40-20) x (300-10-30) = 340 x 260
	// Position = (340-50, 260-50) + padding offset (40, 10) = (330, 220)
	al.AddAnchoredChild(child, ui.AnchorBottomRight, 0, 0)
	al.UpdateLayout()

	wantX := 40 + (340 - 50) // 330
	wantY := 10 + (260 - 50) // 220
	if child.X != float64(wantX) || child.Y != float64(wantY) {
		t.Errorf("got (%g, %g), want (%d, %d)", child.X, child.Y, wantX, wantY)
	}
}

func TestAnchorLayoutOffsets(t *testing.T) {
	resetScheduler()
	al := ui.NewAnchorLayout("al")
	defer al.Dispose()

	al.Padding = ui.Insets{} // zero padding for predictable positions
	al.SetSize(400, 300)

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50

	al.AddAnchoredChild(child, ui.AnchorTopLeft, 15, 25)
	al.UpdateLayout()

	if child.X != 15 || child.Y != 25 {
		t.Errorf("got (%g, %g), want (15, 25)", child.X, child.Y)
	}
}

func TestAnchorLayoutSetAnchor(t *testing.T) {
	resetScheduler()
	al := ui.NewAnchorLayout("al")
	defer al.Dispose()

	al.Padding = ui.Insets{} // zero padding for predictable positions
	al.SetSize(200, 200)

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50

	al.AddAnchoredChild(child, ui.AnchorTopLeft, 0, 0)
	al.UpdateLayout()

	if child.X != 0 || child.Y != 0 {
		t.Errorf("before SetAnchor: got (%g, %g), want (0, 0)", child.X, child.Y)
	}

	al.SetAnchor(child, ui.AnchorBottomRight, 0, 0)
	al.UpdateLayout()

	if child.X != 150 || child.Y != 150 {
		t.Errorf("after SetAnchor: got (%g, %g), want (150, 150)", child.X, child.Y)
	}
}

func TestAnchorLayoutRemoveChild(t *testing.T) {
	resetScheduler()
	al := ui.NewAnchorLayout("al")
	defer al.Dispose()

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50

	al.AddAnchoredChild(child, ui.AnchorCenter, 0, 0)
	if al.NumChildren() != 1 {
		t.Fatalf("expected 1 child, got %d", al.NumChildren())
	}

	al.RemoveChild(child)
	if al.NumChildren() != 0 {
		t.Fatalf("expected 0 children after remove, got %d", al.NumChildren())
	}
}

func TestAnchorLayoutResize(t *testing.T) {
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
		t.Errorf("before resize: got (%g, %g), want (150, 150)", child.X, child.Y)
	}

	al.SetSize(400, 400)
	al.UpdateLayout()

	if child.X != 350 || child.Y != 350 {
		t.Errorf("after resize: got (%g, %g), want (350, 350)", child.X, child.Y)
	}
}

func TestAnchorLayoutChildOffset(t *testing.T) {
	resetScheduler()
	al := ui.NewAnchorLayout("al")
	defer al.Dispose()

	al.Padding = ui.Insets{} // zero padding for predictable positions
	al.SetSize(200, 200)

	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50
	child.OffsetX = 5
	child.OffsetY = 10

	al.AddAnchoredChild(child, ui.AnchorTopLeft, 0, 0)
	al.UpdateLayout()

	// X/Y set by anchor, but node position includes OffsetX/OffsetY.
	if child.X != 0 || child.Y != 0 {
		t.Errorf("child X/Y: got (%g, %g), want (0, 0)", child.X, child.Y)
	}
	// Node position = X + OffsetX, Y + OffsetY
	nx, ny := child.Node().X(), child.Node().Y()
	if nx != 5 || ny != 10 {
		t.Errorf("node position: got (%g, %g), want (5, 10)", nx, ny)
	}
}

func TestAnchorLayoutDefaultAddChild(t *testing.T) {
	resetScheduler()
	al := ui.NewAnchorLayout("al")
	defer al.Dispose()

	al.Padding = ui.Insets{} // zero padding for predictable positions
	al.SetSize(200, 200)

	// Adding via inherited AddChild should default to AnchorTopLeft.
	child := ui.NewComponent("child")
	child.Width = 50
	child.Height = 50
	al.AddChild(child)
	al.UpdateLayout()

	if child.X != 0 || child.Y != 0 {
		t.Errorf("default AddChild: got (%g, %g), want (0, 0)", child.X, child.Y)
	}
}
