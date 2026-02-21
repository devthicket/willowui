package integration

import (
	"math"
	"testing"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/widget"
)

// testNineSlice creates a NineSlice with a 48x48 region and 8px insets.
func testNineSlice() *ui.NineSlice {
	return &ui.NineSlice{
		Region: willow.TextureRegion{Page: 0, X: 0, Y: 0, Width: 48, Height: 48, OriginalW: 48, OriginalH: 48},
		Insets: ui.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8},
	}
}

func TestNineSlice_SubRegions(t *testing.T) {
	base := willow.TextureRegion{Page: 1, X: 10, Y: 20, Width: 64, Height: 64, OriginalW: 64, OriginalH: 64}

	sr := widget.SubRegion(base, 5, 10, 20, 30)
	if sr.X != 15 || sr.Y != 30 || sr.Width != 20 || sr.Height != 30 {
		t.Errorf("SubRegion = {X:%d Y:%d W:%d H:%d}, want {X:15 Y:30 W:20 H:30}",
			sr.X, sr.Y, sr.Width, sr.Height)
	}
	if sr.Page != 1 {
		t.Error("SubRegion should share the same page")
	}
}

func TestNineSlice_NodeCount(t *testing.T) {
	ns := testNineSlice()
	container := willow.NewContainer("test-container")
	nodes := widget.CreateNineSliceNodes("test", container, ns)

	if nodes == nil {
		t.Fatal("CreateNineSliceNodes should return non-nil")
	}

	// Verify all 9 nodes exist and are distinct.
	all := []*willow.Node{nodes.TL, nodes.T, nodes.TR, nodes.L, nodes.C, nodes.R, nodes.BL, nodes.B, nodes.BR}
	for i, n := range all {
		if n == nil {
			t.Fatalf("node %d is nil", i)
		}
	}

	// Check uniqueness.
	seen := make(map[*willow.Node]bool)
	for _, n := range all {
		if seen[n] {
			t.Error("duplicate node found")
		}
		seen[n] = true
	}
}

func TestNineSlice_CornerPositions(t *testing.T) {
	ns := testNineSlice()
	container := willow.NewContainer("test-container")
	nodes := widget.CreateNineSliceNodes("test", container, ns)

	w, h := 200.0, 100.0
	widget.LayoutNineSlice(nodes, ns, w, h)

	// Top-left corner at (0, 0).
	if nodes.TL.X() != 0 || nodes.TL.Y() != 0 {
		t.Errorf("TL position = (%f, %f), want (0, 0)", nodes.TL.X(), nodes.TL.Y())
	}

	// Top-right corner at (w-inR, 0).
	if nodes.TR.X() != w-8 || nodes.TR.Y() != 0 {
		t.Errorf("TR position = (%f, %f), want (%f, 0)", nodes.TR.X(), nodes.TR.Y(), w-8)
	}

	// Bottom-left corner at (0, h-inB).
	if nodes.BL.X() != 0 || nodes.BL.Y() != h-8 {
		t.Errorf("BL position = (%f, %f), want (0, %f)", nodes.BL.X(), nodes.BL.Y(), h-8)
	}

	// Bottom-right corner at (w-inR, h-inB).
	if nodes.BR.X() != w-8 || nodes.BR.Y() != h-8 {
		t.Errorf("BR position = (%f, %f), want (%f, %f)", nodes.BR.X(), nodes.BR.Y(), w-8, h-8)
	}
}

func TestNineSlice_EdgeScaling(t *testing.T) {
	ns := testNineSlice()
	container := willow.NewContainer("test-container")
	nodes := widget.CreateNineSliceNodes("test", container, ns)

	w, h := 200.0, 100.0
	widget.LayoutNineSlice(nodes, ns, w, h)

	midW := w - 8 - 8       // 184
	midH := h - 8 - 8       // 84
	srcMidW := 48.0 - 8 - 8 // 32
	srcMidH := 48.0 - 8 - 8 // 32

	// Top edge: stretches X.
	wantScaleX := midW / srcMidW
	if math.Abs(nodes.T.ScaleX()-wantScaleX) > 0.001 {
		t.Errorf("top edge ScaleX = %f, want %f", nodes.T.ScaleX(), wantScaleX)
	}

	// Left edge: stretches Y.
	wantScaleY := midH / srcMidH
	if math.Abs(nodes.L.ScaleY()-wantScaleY) > 0.001 {
		t.Errorf("left edge ScaleY = %f, want %f", nodes.L.ScaleY(), wantScaleY)
	}
}

func TestNineSlice_CenterScaling(t *testing.T) {
	ns := testNineSlice()
	container := willow.NewContainer("test-container")
	nodes := widget.CreateNineSliceNodes("test", container, ns)

	w, h := 200.0, 100.0
	widget.LayoutNineSlice(nodes, ns, w, h)

	midW := w - 8 - 8
	midH := h - 8 - 8
	srcMidW := 48.0 - 8 - 8
	srcMidH := 48.0 - 8 - 8

	wantSX := midW / srcMidW
	wantSY := midH / srcMidH

	if math.Abs(nodes.C.ScaleX()-wantSX) > 0.001 || math.Abs(nodes.C.ScaleY()-wantSY) > 0.001 {
		t.Errorf("center scale = (%f, %f), want (%f, %f)",
			nodes.C.ScaleX(), nodes.C.ScaleY(), wantSX, wantSY)
	}

	// Center position should be at (inL, inT).
	if nodes.C.X() != 8 || nodes.C.Y() != 8 {
		t.Errorf("center position = (%f, %f), want (8, 8)", nodes.C.X(), nodes.C.Y())
	}
}

func TestNineSlice_Resize(t *testing.T) {
	ns := testNineSlice()
	container := willow.NewContainer("test-container")
	nodes := widget.CreateNineSliceNodes("test", container, ns)

	widget.LayoutNineSlice(nodes, ns, 200, 100)
	oldCenterSX := nodes.C.ScaleX()

	// Resize to different dimensions.
	widget.LayoutNineSlice(nodes, ns, 300, 200)
	if nodes.C.ScaleX() == oldCenterSX {
		t.Error("center ScaleX should change on resize")
	}

	// Verify corners are still in correct positions.
	if nodes.BR.X() != 300-8 || nodes.BR.Y() != 200-8 {
		t.Errorf("BR position after resize = (%f, %f), want (292, 192)",
			nodes.BR.X(), nodes.BR.Y())
	}
}

func TestNineSlice_LazyCreation(t *testing.T) {
	c := ui.NewComponent("test-lazy")
	defer c.Dispose()

	c.InitBackgroundForTest("test-lazy")

	// Before any nine-slice usage, bgContainer and bgSliceNodes should be nil.
	if c.BgContainer() != nil {
		t.Error("bgContainer should be nil before first nine-slice use")
	}
	if c.HasBgSliceNodes() {
		t.Error("bgSliceNodes should be nil before first nine-slice use")
	}

	// Apply a nine-slice background.
	ns := testNineSlice()
	c.Width = 100
	c.Height = 100
	c.ApplyBackgroundForTest(ui.SliceBackground(ns))

	if c.BgContainer() == nil {
		t.Error("bgContainer should be created after nine-slice apply")
	}
	if !c.HasBgSliceNodes() {
		t.Error("bgSliceNodes should be created after nine-slice apply")
	}
}

func TestNineSlice_SolidHidesContainer(t *testing.T) {
	c := ui.NewComponent("test-hide")
	defer c.Dispose()

	c.InitBackgroundForTest("test-hide")
	c.Width = 100
	c.Height = 100

	// Apply nine-slice first.
	ns := testNineSlice()
	c.ApplyBackgroundForTest(ui.SliceBackground(ns))
	if !c.BgContainer().Visible() {
		t.Error("bgContainer should be visible after nine-slice apply")
	}

	// Switch back to solid.
	c.ApplyBackgroundForTest(ui.SolidBackground(willow.RGBA(1, 0, 0, 1)))
	if c.BgContainer().Visible() {
		t.Error("bgContainer should be hidden after switching to solid")
	}
	if !c.BgNode().Visible() {
		t.Error("bgNode should be visible after switching to solid")
	}
}

func TestNineSlice_MinimumSize(t *testing.T) {
	ns := testNineSlice() // 8px insets
	container := willow.NewContainer("test-container")
	nodes := widget.CreateNineSliceNodes("test", container, ns)

	// Layout with size smaller than insets.
	widget.LayoutNineSlice(nodes, ns, 10, 6)

	// Should not panic and corners should clamp gracefully.
	// Width 10 < 16 (8+8), height 6 < 16 (8+8).
	// TL should be at (0,0), BR should be at correct clamped position.
	if nodes.TL.X() != 0 || nodes.TL.Y() != 0 {
		t.Errorf("TL should still be at (0,0), got (%f, %f)", nodes.TL.X(), nodes.TL.Y())
	}
}
