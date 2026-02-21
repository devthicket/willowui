package integration

import (
	"testing"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

// --- Lifecycle ---

func TestNewComponent(t *testing.T) {
	c := ui.NewComponent("test")
	if c.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if c.Name() != "test" {
		t.Errorf("Name() = %q, want %q", c.Name(), "test")
	}
	if !c.IsEnabled() {
		t.Error("new component should be enabled")
	}
	if !c.IsVisible() {
		t.Error("new component should be visible")
	}
	if !c.IsLayoutDirty() {
		t.Error("new component should have dirty layout")
	}
	if !c.IsDrawDirty() {
		t.Error("new component should have dirty draw")
	}
}

func TestDispose(t *testing.T) {
	parent := ui.NewComponent("parent")
	child := ui.NewComponent("child")
	grandchild := ui.NewComponent("grandchild")

	parent.AddChild(child)
	child.AddChild(grandchild)

	if parent.NumChildren() != 1 {
		t.Fatalf("parent should have 1 child, got %d", parent.NumChildren())
	}

	child.Dispose()

	if parent.NumChildren() != 0 {
		t.Errorf("parent should have 0 children after dispose, got %d", parent.NumChildren())
	}
	if child.Parent() != nil {
		t.Error("disposed child should have nil parent")
	}
	if grandchild.IsDisposed() == false {
		t.Error("grandchild node should be disposed")
	}
}

func TestDisposeRoot(t *testing.T) {
	root := ui.NewComponent("root")
	child := ui.NewComponent("child")
	root.AddChild(child)

	root.Dispose()

	if root.NumChildren() != 0 {
		t.Errorf("disposed root should have 0 children, got %d", root.NumChildren())
	}
}

// --- Dirty flag propagation ---

func TestDirtyFlagPropagation(t *testing.T) {
	parent := ui.NewComponent("parent")
	child := ui.NewComponent("child")
	parent.AddChild(child)

	// Clear dirty flags by running layout.
	parent.UpdateLayout()

	if parent.IsLayoutDirty() {
		t.Error("parent should be clean after UpdateLayout")
	}

	// Marking child dirty should propagate to parent.
	child.MarkLayoutDirty()

	if !child.IsLayoutDirty() {
		t.Error("child should be dirty")
	}
	if !parent.IsLayoutDirty() {
		t.Error("parent should be dirty after child.MarkLayoutDirty")
	}
}

func TestMarkDrawDirty(t *testing.T) {
	c := ui.NewComponent("c")
	c.UpdateLayout()
	c.SimulateDirtyDraw(false)

	c.MarkDrawDirty()
	if !c.IsDrawDirty() {
		t.Error("component should be draw-dirty after MarkDrawDirty")
	}
}

// --- Add/Remove children ---

func TestAddRemoveChildren(t *testing.T) {
	parent := ui.NewComponent("parent")
	a := ui.NewComponent("a")
	b := ui.NewComponent("b")

	parent.AddChild(a)
	parent.AddChild(b)

	if parent.NumChildren() != 2 {
		t.Fatalf("expected 2 children, got %d", parent.NumChildren())
	}
	if parent.Node().NumChildren() != 2 {
		t.Fatalf("expected 2 node children, got %d", parent.Node().NumChildren())
	}

	// Verify parent pointers.
	if a.Parent() != parent {
		t.Error("a.Parent() should be parent")
	}

	parent.RemoveChild(a)
	if parent.NumChildren() != 1 {
		t.Errorf("expected 1 child after remove, got %d", parent.NumChildren())
	}
	if parent.Node().NumChildren() != 1 {
		t.Errorf("expected 1 node child after remove, got %d", parent.Node().NumChildren())
	}
	if a.Parent() != nil {
		t.Error("removed child should have nil parent")
	}
}

func TestAddChildReparents(t *testing.T) {
	p1 := ui.NewComponent("p1")
	p2 := ui.NewComponent("p2")
	child := ui.NewComponent("child")

	p1.AddChild(child)
	if child.Parent() != p1 {
		t.Fatal("child should be parented to p1")
	}

	// Adding to p2 should remove from p1 first.
	p2.AddChild(child)
	if child.Parent() != p2 {
		t.Error("child should be reparented to p2")
	}
	if p1.NumChildren() != 0 {
		t.Errorf("p1 should have 0 children, got %d", p1.NumChildren())
	}
}

// --- Enable/Disable ---

func TestEnableDisable(t *testing.T) {
	c := ui.NewComponent("c")
	if !c.IsEnabled() {
		t.Fatal("should start enabled")
	}

	c.SetEnabled(false)
	if c.IsEnabled() {
		t.Error("should be disabled")
	}
	if c.IsInteractable() {
		t.Error("node should not be interactable when disabled")
	}

	c.SetEnabled(true)
	if !c.IsEnabled() {
		t.Error("should be re-enabled")
	}
	if !c.IsInteractable() {
		t.Error("node should be interactable when enabled")
	}
}

// --- Visibility ---

func TestVisibility(t *testing.T) {
	c := ui.NewComponent("c")
	c.SetVisible(false)
	if c.IsVisible() {
		t.Error("should be invisible")
	}
	if c.IsVisible() {
		t.Error("node should be invisible")
	}
	c.SetVisible(true)
	if !c.IsVisible() {
		t.Error("should be visible again")
	}
}

// --- Layout modes ---

func TestLayoutVBox(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutVBox
	parent.Spacing = 10
	parent.Padding = ui.Insets{Top: 5, Left: 5}
	parent.Width = 200
	parent.Height = 200

	a := ui.NewComponent("a")
	a.Width = 50
	a.Height = 20
	b := ui.NewComponent("b")
	b.Width = 50
	b.Height = 30

	parent.AddChild(a)
	parent.AddChild(b)
	parent.UpdateLayout()

	// a should be at padding offset.
	if a.X != 5 {
		t.Errorf("a.X = %f, want 5", a.X)
	}
	if a.Y != 5 {
		t.Errorf("a.Y = %f, want 5", a.Y)
	}

	// b should be below a + spacing.
	expectedBY := 5 + 20 + 10.0 // padding.Top + a.Height + spacing
	if b.Y != expectedBY {
		t.Errorf("b.Y = %f, want %f", b.Y, expectedBY)
	}
	if b.X != 5 {
		t.Errorf("b.X = %f, want 5", b.X)
	}
}

func TestLayoutHBox(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutHBox
	parent.Spacing = 8
	parent.Padding = ui.Insets{Top: 3, Left: 3}
	parent.Width = 300
	parent.Height = 100

	a := ui.NewComponent("a")
	a.Width = 40
	a.Height = 20
	b := ui.NewComponent("b")
	b.Width = 60
	b.Height = 20

	parent.AddChild(a)
	parent.AddChild(b)
	parent.UpdateLayout()

	if a.X != 3 {
		t.Errorf("a.X = %f, want 3", a.X)
	}
	if a.Y != 3 {
		t.Errorf("a.Y = %f, want 3", a.Y)
	}

	expectedBX := 3 + 40 + 8.0 // padding.Left + a.Width + spacing
	if b.X != expectedBX {
		t.Errorf("b.X = %f, want %f", b.X, expectedBX)
	}
	if b.Y != 3 {
		t.Errorf("b.Y = %f, want 3", b.Y)
	}
}

func TestLayoutGrid(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutGrid
	parent.GridColumns = 2
	parent.Spacing = 4
	parent.Padding = ui.Insets{Top: 2, Left: 2}
	parent.Width = 200
	parent.Height = 200

	items := make([]*ui.Component, 4)
	for i := range items {
		items[i] = ui.NewComponent("item")
		items[i].Width = 30
		items[i].Height = 20
		parent.AddChild(items[i])
	}

	parent.UpdateLayout()

	// Row 0, Col 0
	if items[0].X != 2 {
		t.Errorf("items[0].X = %f, want 2", items[0].X)
	}
	if items[0].Y != 2 {
		t.Errorf("items[0].Y = %f, want 2", items[0].Y)
	}
	// Row 0, Col 1
	expectedX1 := 2 + (30 + 4.0) // padding.Left + (cellW + spacing)
	if items[1].X != expectedX1 {
		t.Errorf("items[1].X = %f, want %f", items[1].X, expectedX1)
	}
	// Row 1, Col 0
	expectedY2 := 2 + (20 + 4.0)
	if items[2].Y != expectedY2 {
		t.Errorf("items[2].Y = %f, want %f", items[2].Y, expectedY2)
	}
	if items[2].X != 2 {
		t.Errorf("items[2].X = %f, want 2", items[2].X)
	}
}

func TestLayoutNone(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutNone
	parent.Width = 200
	parent.Height = 200

	child := ui.NewComponent("child")
	child.X = 42
	child.Y = 73
	child.Width = 10
	child.Height = 10
	parent.AddChild(child)
	parent.UpdateLayout()

	if child.Node().X() != 42 {
		t.Errorf("node.X = %f, want 42", child.Node().X())
	}
	if child.Node().Y() != 73 {
		t.Errorf("node.Y = %f, want 73", child.Node().Y())
	}
}

func TestLayoutSkipsInvisibleChildren(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutVBox
	parent.Spacing = 10
	parent.Width = 100
	parent.Height = 100

	a := ui.NewComponent("a")
	a.Width = 50
	a.Height = 20

	hidden := ui.NewComponent("hidden")
	hidden.Width = 50
	hidden.Height = 20
	hidden.SetVisible(false)

	b := ui.NewComponent("b")
	b.Width = 50
	b.Height = 20

	parent.AddChild(a)
	parent.AddChild(hidden)
	parent.AddChild(b)
	parent.UpdateLayout()

	// b should be right after a, since hidden is invisible.
	expectedBY := 0 + 20 + 10.0 // a.Y + a.Height + spacing
	if b.Y != expectedBY {
		t.Errorf("b.Y = %f, want %f (hidden child should be skipped)", b.Y, expectedBY)
	}
}

// --- Cross-axis alignment ---

func TestVBoxAlignCenter(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutVBox
	parent.Align = ui.AlignCenter
	parent.Width = 200
	parent.Height = 200
	parent.Padding = ui.Insets{Left: 10, Right: 10}

	a := ui.NewComponent("a")
	a.Width = 80
	a.Height = 20

	parent.AddChild(a)
	parent.UpdateLayout()

	// avail = 200 - 10 - 10 = 180; x = 10 + (180-80)/2 = 60
	if a.X != 60 {
		t.Errorf("VBox AlignCenter: a.X = %f, want 60", a.X)
	}
}

func TestVBoxAlignEnd(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutVBox
	parent.Align = ui.AlignEnd
	parent.Width = 200
	parent.Height = 200
	parent.Padding = ui.Insets{Left: 10, Right: 10}

	a := ui.NewComponent("a")
	a.Width = 80
	a.Height = 20
	a.Margin = ui.Insets{Right: 5}

	parent.AddChild(a)
	parent.UpdateLayout()

	// avail = 180; x = 10 + 180 - 80 - 5 = 105
	if a.X != 105 {
		t.Errorf("VBox AlignEnd: a.X = %f, want 105", a.X)
	}
}

func TestHBoxAlignCenter(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutHBox
	parent.Align = ui.AlignCenter
	parent.Width = 300
	parent.Height = 100
	parent.Padding = ui.Insets{Top: 10, Bottom: 10}

	a := ui.NewComponent("a")
	a.Width = 40
	a.Height = 30

	parent.AddChild(a)
	parent.UpdateLayout()

	// avail = 100 - 10 - 10 = 80; y = 10 + (80-30)/2 = 35
	if a.Y != 35 {
		t.Errorf("HBox AlignCenter: a.Y = %f, want 35", a.Y)
	}
}

func TestHBoxAlignEnd(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutHBox
	parent.Align = ui.AlignEnd
	parent.Width = 300
	parent.Height = 100
	parent.Padding = ui.Insets{Top: 10, Bottom: 10}

	a := ui.NewComponent("a")
	a.Width = 40
	a.Height = 30
	a.Margin = ui.Insets{Bottom: 5}

	parent.AddChild(a)
	parent.UpdateLayout()

	// avail = 80; y = 10 + 80 - 30 - 5 = 55
	if a.Y != 55 {
		t.Errorf("HBox AlignEnd: a.Y = %f, want 55", a.Y)
	}
}

// --- Main-axis justify ---

func TestHBoxJustifyCenter(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutHBox
	parent.Justify = ui.AlignCenter
	parent.Width = 300
	parent.Height = 100
	parent.Padding = ui.Insets{Left: 10, Right: 10}

	a := ui.NewComponent("a")
	a.Width = 40
	a.Height = 20
	b := ui.NewComponent("b")
	b.Width = 60
	b.Height = 20

	parent.AddChild(a)
	parent.AddChild(b)
	parent.Spacing = 10
	parent.UpdateLayout()

	// availX = 300 - 10 - 10 = 280
	// totalW = 40 + 10 + 60 = 110
	// offset = 10 + (280 - 110) / 2 = 10 + 85 = 95
	if a.X != 95 {
		t.Errorf("HBox JustifyCenter: a.X = %f, want 95", a.X)
	}
	expectedBX := 95 + 40 + 10.0 // a.X + a.Width + spacing
	if b.X != expectedBX {
		t.Errorf("HBox JustifyCenter: b.X = %f, want %f", b.X, expectedBX)
	}
}

func TestHBoxJustifyEnd(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutHBox
	parent.Justify = ui.AlignEnd
	parent.Width = 300
	parent.Height = 100
	parent.Padding = ui.Insets{Left: 10, Right: 10}

	a := ui.NewComponent("a")
	a.Width = 40
	a.Height = 20
	b := ui.NewComponent("b")
	b.Width = 60
	b.Height = 20

	parent.AddChild(a)
	parent.AddChild(b)
	parent.Spacing = 10
	parent.UpdateLayout()

	// availX = 280, totalW = 110
	// offset = 10 + 280 - 110 = 180
	if a.X != 180 {
		t.Errorf("HBox JustifyEnd: a.X = %f, want 180", a.X)
	}
	expectedBX := 180 + 40 + 10.0
	if b.X != expectedBX {
		t.Errorf("HBox JustifyEnd: b.X = %f, want %f", b.X, expectedBX)
	}
}

func TestVBoxJustifyCenter(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutVBox
	parent.Justify = ui.AlignCenter
	parent.Width = 200
	parent.Height = 300
	parent.Padding = ui.Insets{Top: 10, Bottom: 10}

	a := ui.NewComponent("a")
	a.Width = 50
	a.Height = 20
	b := ui.NewComponent("b")
	b.Width = 50
	b.Height = 30

	parent.AddChild(a)
	parent.AddChild(b)
	parent.Spacing = 10
	parent.UpdateLayout()

	// availY = 300 - 10 - 10 = 280
	// totalH = 20 + 10 + 30 = 60
	// offset = 10 + (280 - 60) / 2 = 10 + 110 = 120
	if a.Y != 120 {
		t.Errorf("VBox JustifyCenter: a.Y = %f, want 120", a.Y)
	}
	expectedBY := 120 + 20 + 10.0
	if b.Y != expectedBY {
		t.Errorf("VBox JustifyCenter: b.Y = %f, want %f", b.Y, expectedBY)
	}
}

func TestVBoxJustifyEnd(t *testing.T) {
	parent := ui.NewComponent("parent")
	parent.Layout = ui.LayoutVBox
	parent.Justify = ui.AlignEnd
	parent.Width = 200
	parent.Height = 300
	parent.Padding = ui.Insets{Top: 10, Bottom: 10}

	a := ui.NewComponent("a")
	a.Width = 50
	a.Height = 20
	b := ui.NewComponent("b")
	b.Width = 50
	b.Height = 30

	parent.AddChild(a)
	parent.AddChild(b)
	parent.Spacing = 10
	parent.UpdateLayout()

	// availY = 280, totalH = 60
	// offset = 10 + 280 - 60 = 230
	if a.Y != 230 {
		t.Errorf("VBox JustifyEnd: a.Y = %f, want 230", a.Y)
	}
	expectedBY := 230 + 20 + 10.0
	if b.Y != expectedBY {
		t.Errorf("VBox JustifyEnd: b.Y = %f, want %f", b.Y, expectedBY)
	}
}

// --- Constraints ---

func TestMinMaxConstraints(t *testing.T) {
	c := ui.NewComponent("c")
	c.MinWidth = 50
	c.MaxWidth = 200
	c.MinHeight = 30
	c.MaxHeight = 150

	// Below min.
	c.Width = 10
	c.Height = 5
	c.UpdateLayout()
	if c.Width != 50 {
		t.Errorf("Width = %f, want 50 (MinWidth)", c.Width)
	}
	if c.Height != 30 {
		t.Errorf("Height = %f, want 30 (MinHeight)", c.Height)
	}

	// Above max.
	c.Width = 999
	c.Height = 999
	c.MarkLayoutDirty()
	c.UpdateLayout()
	if c.Width != 200 {
		t.Errorf("Width = %f, want 200 (MaxWidth)", c.Width)
	}
	if c.Height != 150 {
		t.Errorf("Height = %f, want 150 (MaxHeight)", c.Height)
	}
}

// --- Focus Manager ---

func TestFocusManagerSetClear(t *testing.T) {
	fm := ui.NewFocusManager()
	a := ui.NewComponent("a")
	b := ui.NewComponent("b")

	fm.SetFocus(a)
	if fm.Focused() != a {
		t.Error("focused should be a")
	}
	if !a.IsFocused() {
		t.Error("a should report focused")
	}

	fm.SetFocus(b)
	if fm.Focused() != b {
		t.Error("focused should be b")
	}
	if a.IsFocused() {
		t.Error("a should no longer be focused")
	}
	if !b.IsFocused() {
		t.Error("b should report focused")
	}

	fm.ClearFocus()
	if fm.Focused() != nil {
		t.Error("focused should be nil after clear")
	}
	if b.IsFocused() {
		t.Error("b should no longer be focused after clear")
	}
}

func TestFocusManagerTabCycling(t *testing.T) {
	fm := ui.NewFocusManager()
	a := ui.NewComponent("a")
	b := ui.NewComponent("b")
	c := ui.NewComponent("c")
	for _, comp := range []*ui.Component{a, b, c} {
		comp.Focusable = true
		comp.AllowTab = true
	}

	fm.Register(a)
	fm.Register(b)
	fm.Register(c)

	// TabNext from nil should focus first.
	fm.TabNext()
	if fm.Focused() != a {
		t.Errorf("focused = %v, want a", fm.Focused())
	}

	fm.TabNext()
	if fm.Focused() != b {
		t.Errorf("focused = %v, want b", fm.Focused())
	}

	fm.TabNext()
	if fm.Focused() != c {
		t.Errorf("focused = %v, want c", fm.Focused())
	}

	// Wrap around.
	fm.TabNext()
	if fm.Focused() != a {
		t.Errorf("focused = %v, want a (wrap)", fm.Focused())
	}

	// TabPrev.
	fm.TabPrev()
	if fm.Focused() != c {
		t.Errorf("focused = %v, want c (prev wrap)", fm.Focused())
	}

	fm.TabPrev()
	if fm.Focused() != b {
		t.Errorf("focused = %v, want b", fm.Focused())
	}
}

func TestFocusManagerUnregister(t *testing.T) {
	fm := ui.NewFocusManager()
	a := ui.NewComponent("a")
	b := ui.NewComponent("b")

	fm.Register(a)
	fm.Register(b)
	fm.SetFocus(a)

	fm.Unregister(a)
	if fm.Focused() != nil {
		t.Error("focused should be nil after unregistering focused component")
	}
	if a.IsFocused() {
		t.Error("a should not be focused after unregister")
	}
}

func TestFocusManagerRegisterDuplicate(t *testing.T) {
	fm := ui.NewFocusManager()
	a := ui.NewComponent("a")
	a.Focusable = true
	a.AllowTab = true

	fm.Register(a)
	fm.Register(a) // duplicate

	// Should only appear once in tab order.
	fm.TabNext()
	fm.TabNext() // wraps to same
	if fm.Focused() != a {
		t.Error("should still be on a after cycling through single entry")
	}
}

func TestFocusManagerTabPrevFromNil(t *testing.T) {
	fm := ui.NewFocusManager()
	a := ui.NewComponent("a")
	b := ui.NewComponent("b")
	a.Focusable = true
	a.AllowTab = true
	b.Focusable = true
	b.AllowTab = true
	fm.Register(a)
	fm.Register(b)

	// TabPrev from nil should focus last.
	fm.TabPrev()
	if fm.Focused() != b {
		t.Errorf("focused = %v, want b (last)", fm.Focused())
	}
}

// --- Insets ---

func TestInsets(t *testing.T) {
	i := ui.Insets{Top: 10, Right: 20, Bottom: 30, Left: 40}
	if i.Horizontal() != 60 {
		t.Errorf("Horizontal() = %f, want 60", i.Horizontal())
	}
	if i.Vertical() != 40 {
		t.Errorf("Vertical() = %f, want 40", i.Vertical())
	}
}

// --- Theme ---

func TestDefaultThemeValues(t *testing.T) {
	// Button primary background should be opaque.
	bg := ui.DefaultTheme.Button.Group(ui.Primary).Background.Resolve(ui.StateDefault)
	if bg.Color.A() != 1 {
		t.Error("Button default background alpha should be 1")
	}
}

// --- Node tree sync ---

func TestNodeTreeSync(t *testing.T) {
	root := ui.NewComponent("root")
	a := ui.NewComponent("a")
	b := ui.NewComponent("b")

	root.AddChild(a)
	a.AddChild(b)

	// Verify willow node tree mirrors the component tree.
	rootNode := root.Node()
	if rootNode.NumChildren() != 1 {
		t.Fatalf("root node should have 1 child, got %d", rootNode.NumChildren())
	}
	aNode := rootNode.Children()[0]
	if aNode != a.Node() {
		t.Error("root's first node child should be a's node")
	}
	if aNode.NumChildren() != 1 {
		t.Fatalf("a node should have 1 child, got %d", aNode.NumChildren())
	}
	if aNode.Children()[0] != b.Node() {
		t.Error("a's first node child should be b's node")
	}
}

// --- Gradient background ---

func TestComponent_GradientBackground(t *testing.T) {
	c := ui.NewComponent("grad-test")
	c.Width = 100
	c.Height = 60
	c.InitBackgroundForTest("grad-test")

	g := &ui.GradientColors{
		TopLeft:     willow.RGBA(1, 0, 0, 1),
		TopRight:    willow.RGBA(0, 1, 0, 1),
		BottomRight: willow.RGBA(0, 0, 1, 1),
		BottomLeft:  willow.RGBA(1, 1, 0, 1),
	}
	bg := ui.GradientBackground(g)
	c.ApplyBackgroundForTest(bg)

	// Verify gradient mesh was created and is visible.
	if c.BgGradientMesh() == nil {
		t.Fatal("bgGradientMesh should be created")
	}
	if !c.BgGradientMesh().Visible() {
		t.Error("bgGradientMesh should be visible")
	}

	// Verify other bg types are hidden.
	if c.BgNode().Visible() {
		t.Error("bgNode should be hidden when gradient is active")
	}

	// Now switch to solid background.
	c.ApplyBackgroundForTest(ui.SolidBackground(willow.RGBA(1, 1, 1, 1)))
	if c.BgGradientMesh().Visible() {
		t.Error("bgGradientMesh should be hidden when solid bg is active")
	}
}
