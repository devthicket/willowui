package integration

import (
	"testing"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

// ---------------------------------------------------------------------------
// Panel
// ---------------------------------------------------------------------------

func TestNewPanelDefaults(t *testing.T) {
	resetScheduler()
	p := ui.NewPanel("panel")
	defer p.Dispose()

	if p.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if p.BgNode() == nil {
		t.Fatal("bgNode should not be nil")
	}
	if p.BorderTop() == nil || p.BorderRight() == nil || p.BorderBot() == nil || p.BorderLeft() == nil {
		t.Fatal("border sprites should not be nil")
	}
}

func TestPanelSetSize(t *testing.T) {
	resetScheduler()
	p := ui.NewPanel("panel")
	defer p.Dispose()

	p.SetSize(200, 150)
	if p.Width != 200 || p.Height != 150 {
		t.Errorf("size = %fx%f, want 200x150", p.Width, p.Height)
	}
	if p.BgNode().ScaleX() != 200 || p.BgNode().ScaleY() != 150 {
		t.Errorf("background scale = %fx%f, want 200x150", p.BgNode().ScaleX(), p.BgNode().ScaleY())
	}
}

func TestPanelSetBackground(t *testing.T) {
	resetScheduler()
	p := ui.NewPanel("panel")
	defer p.Dispose()

	red := willow.RGBA(1, 0, 0, 1)
	p.SetBackground(red)
	if p.BgNode().Color() != red {
		t.Errorf("background color not set correctly")
	}
}

func TestPanelSetBorder(t *testing.T) {
	resetScheduler()
	p := ui.NewPanel("panel")
	defer p.Dispose()

	p.SetSize(100, 80)
	green := willow.RGBA(0, 1, 0, 1)
	p.SetBorder(green, 2)

	if p.BorderWidth() != 2 {
		t.Errorf("borderWidth = %f, want 2", p.BorderWidth())
	}
	if p.BorderTop().Color() != green {
		t.Error("border top color not set correctly")
	}
	// Top border should span full width and be 2px tall.
	if p.BorderTop().ScaleX() != 100 || p.BorderTop().ScaleY() != 2 {
		t.Errorf("top border scale = %fx%f, want 100x2", p.BorderTop().ScaleX(), p.BorderTop().ScaleY())
	}
	// Right border should be at x=98 (100-2).
	if p.BorderRight().X() != 98 {
		t.Errorf("right border X = %f, want 98", p.BorderRight().X())
	}
}

func TestPanelVBoxLayout(t *testing.T) {
	resetScheduler()
	p := ui.NewPanel("panel")
	defer p.Dispose()

	p.Padding = ui.Insets{} // zero padding for predictable positions
	p.SetSize(200, 300)
	p.SetLayout(ui.LayoutVBox)
	p.SetSpacing(10)

	c1 := ui.NewComponent("c1")
	c1.Width = 100
	c1.Height = 30
	c2 := ui.NewComponent("c2")
	c2.Width = 100
	c2.Height = 40

	p.AddChild(c1)
	p.AddChild(c2)
	p.UpdateLayout()

	// VBox: c1 at Y=0 (no padding), c2 at Y=30+10=40.
	if c1.Y != 0 {
		t.Errorf("c1.Y = %f, want 0", c1.Y)
	}
	if c2.Y != 40 {
		t.Errorf("c2.Y = %f, want 40 (30 height + 10 spacing)", c2.Y)
	}
}

func TestPanelHBoxLayout(t *testing.T) {
	resetScheduler()
	p := ui.NewPanel("panel")
	defer p.Dispose()

	p.Padding = ui.Insets{} // zero padding for predictable positions
	p.SetSize(400, 100)
	p.SetLayout(ui.LayoutHBox)
	p.SetSpacing(5)

	c1 := ui.NewComponent("c1")
	c1.Width = 60
	c1.Height = 30
	c2 := ui.NewComponent("c2")
	c2.Width = 80
	c2.Height = 30

	p.AddChild(c1)
	p.AddChild(c2)
	p.UpdateLayout()

	if c1.X != 0 {
		t.Errorf("c1.X = %f, want 0", c1.X)
	}
	if c2.X != 65 {
		t.Errorf("c2.X = %f, want 65 (60 width + 5 spacing)", c2.X)
	}
}

func TestPanelVBoxLayoutWithPadding(t *testing.T) {
	resetScheduler()
	p := ui.NewPanel("panel")
	defer p.Dispose()

	p.SetSize(200, 300)
	p.SetLayout(ui.LayoutVBox)
	p.Padding = ui.Insets{Top: 10, Left: 15, Right: 15, Bottom: 10}
	p.SetSpacing(5)

	c1 := ui.NewComponent("c1")
	c1.Width = 100
	c1.Height = 30

	p.AddChild(c1)
	p.UpdateLayout()

	if c1.X != 15 {
		t.Errorf("c1.X = %f, want 15 (padding left)", c1.X)
	}
	if c1.Y != 10 {
		t.Errorf("c1.Y = %f, want 10 (padding top)", c1.Y)
	}
}

// ---------------------------------------------------------------------------
// ScrollPanel
// ---------------------------------------------------------------------------

func TestNewScrollPanelDefaults(t *testing.T) {
	resetScheduler()
	sp := ui.NewScrollPanel("scroll")
	defer sp.Dispose()

	if sp.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if sp.Viewport() == nil {
		t.Fatal("viewport should not be nil")
	}
	if sp.ContentNode() == nil {
		t.Fatal("content should not be nil")
	}
	if sp.VScrollBar() == nil {
		t.Fatal("vScrollBar should not be nil")
	}
	if sp.HScrollBar() == nil {
		t.Fatal("hScrollBar should not be nil")
	}
}

func TestScrollPanelSetContentSize(t *testing.T) {
	resetScheduler()
	sp := ui.NewScrollPanel("scroll")
	defer sp.Dispose()

	sp.SetSize(200, 150)
	sp.SetContentSize(200, 500)

	if sp.ContentH() != 500 {
		t.Errorf("contentH = %f, want 500", sp.ContentH())
	}
}

func TestScrollPanelScrollSync(t *testing.T) {
	resetScheduler()
	sp := ui.NewScrollPanel("scroll")
	defer sp.Dispose()

	sp.SetSize(200, 100)
	sp.SetContentSize(200, 500)

	sp.SetScrollY(50)
	if sp.ScrollY() != 50 {
		t.Errorf("ScrollY() = %f, want 50", sp.ScrollY())
	}
	// Content should be offset.
	if sp.ContentNode().Y() != -50 {
		t.Errorf("content.Y = %f, want -50", sp.ContentNode().Y())
	}
}

func TestScrollPanelScrollClamp(t *testing.T) {
	resetScheduler()
	sp := ui.NewScrollPanel("scroll")
	defer sp.Dispose()

	sp.SetSize(200, 100)
	sp.SetContentSize(200, 300)

	sp.SetScrollY(9999)
	if sp.ScrollY() > 300 {
		t.Errorf("ScrollY() = %f, should be clamped <= 300", sp.ScrollY())
	}

	sp.SetScrollY(-50)
	if sp.ScrollY() != 0 {
		t.Errorf("ScrollY() = %f, should be clamped to 0", sp.ScrollY())
	}
}

func TestScrollPanelShowHideScrollbars(t *testing.T) {
	resetScheduler()
	sp := ui.NewScrollPanel("scroll")
	defer sp.Dispose()

	sp.ShowVScroll(false)
	if sp.VScrollBar().IsVisible() {
		t.Error("vScrollBar should be hidden")
	}
	sp.ShowVScroll(true)
	if !sp.VScrollBar().IsVisible() {
		t.Error("vScrollBar should be visible")
	}
	sp.ShowHScroll(true)
	if !sp.HScrollBar().IsVisible() {
		t.Error("hScrollBar should be visible")
	}
}

func TestScrollPanelAddRemoveChild(t *testing.T) {
	resetScheduler()
	sp := ui.NewScrollPanel("scroll")
	defer sp.Dispose()

	child := ui.NewComponent("child")
	sp.AddChild(child)

	if sp.NumChildren() != 1 {
		t.Errorf("NumChildren() = %d, want 1", sp.NumChildren())
	}

	sp.RemoveChild(child)
	if sp.NumChildren() != 0 {
		t.Errorf("NumChildren() = %d, want 0 after remove", sp.NumChildren())
	}
}

// ---------------------------------------------------------------------------
// Window
// ---------------------------------------------------------------------------

func TestNewWindowDefaults(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	w := ui.NewWindow("win", "Test Window", font, 0)
	defer w.Dispose()

	if w.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if w.TitleLabel() == nil {
		t.Fatal("titleLabel should not be nil")
	}
	if w.CloseBtn() == nil {
		t.Fatal("closeBtn should not be nil")
	}
	if w.Body() == nil {
		t.Fatal("body should not be nil")
	}
	if w.BgNode() == nil {
		t.Fatal("bgNode should not be nil")
	}
	if w.Width != float64(ui.DefaultWindowWidth) {
		t.Errorf("Width = %f, want %d", w.Width, ui.DefaultWindowWidth)
	}
	if w.Height != float64(ui.DefaultWindowHeight) {
		t.Errorf("Height = %f, want %d", w.Height, ui.DefaultWindowHeight)
	}
}

func TestWindowSetTitle(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	w := ui.NewWindow("win", "Original", font, 0)
	defer w.Dispose()

	w.SetTitle("New Title")
	if w.TitleLabel().Text() != "New Title" {
		t.Errorf("title = %q, want %q", w.TitleLabel().Text(), "New Title")
	}
}

func TestWindowSetSize(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	w := ui.NewWindow("win", "Test", font, 0)
	defer w.Dispose()

	w.SetSize(500, 400)
	if w.Width != 500 || w.Height != 400 {
		t.Errorf("size = %fx%f, want 500x400", w.Width, w.Height)
	}
}

func TestWindowSetSizeRespectsMin(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	w := ui.NewWindow("win", "Test", font, 0)
	defer w.Dispose()

	w.SetMinWidth(200)
	w.SetMinHeight(150)

	w.SetSize(50, 50)
	if w.Width != 200 {
		t.Errorf("Width = %f, want 200 (min)", w.Width)
	}
	if w.Height != 150 {
		t.Errorf("Height = %f, want 150 (min)", w.Height)
	}
}

func TestWindowClose(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	w := ui.NewWindow("win", "Test", font, 0)
	defer w.Dispose()

	closed := false
	w.SetOnClose(func() { closed = true })

	w.Close()

	if w.IsVisible() {
		t.Error("window should be hidden after Close()")
	}
	if !closed {
		t.Error("onClose callback should have fired")
	}
}

func TestWindowBringToFront(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	// Use a fresh manager for deterministic tests.
	wm := ui.NewWindowManager()

	w1 := ui.NewWindow("w1", "First", font, 0)
	defer w1.Dispose()
	w2 := ui.NewWindow("w2", "Second", font, 0)
	defer w2.Dispose()

	wm.Add(w1)
	wm.Add(w2)

	wm.BringToFront(w1)
	z1 := w1.ZIndex()

	wm.BringToFront(w2)
	z2 := w2.ZIndex()

	if z2 <= z1 {
		t.Errorf("w2 ZIndex (%d) should be > w1 ZIndex (%d)", z2, z1)
	}
}

func TestWindowBody(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	w := ui.NewWindow("win", "Test", font, 0)
	defer w.Dispose()

	body := w.Body()
	if body == nil {
		t.Fatal("Body() should not be nil")
	}

	// Body should be positioned below the title bar.
	bodyY := w.Body().Node().Y()
	if bodyY != float64(ui.WindowTitleBarHeight) {
		t.Errorf("body Y = %f, want %f", bodyY, float64(ui.WindowTitleBarHeight))
	}
}

func TestWindowSetResizable(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	w := ui.NewWindow("win", "Test", font, 0)
	defer w.Dispose()

	// Either the polygon or flat sprite may be the active handle depending
	// on the theme's CornerRadius value.
	resizeVisible := func() bool {
		return w.ResizeHandle().Visible() || w.ResizeFlat().Visible()
	}

	if resizeVisible() {
		t.Error("resize handle should be hidden by default")
	}

	w.SetResizable(true)
	if !resizeVisible() {
		t.Error("resize handle should be visible after SetResizable(true)")
	}

	w.SetResizable(false)
	if resizeVisible() {
		t.Error("resize handle should be hidden after SetResizable(false)")
	}
}

func TestWindowSetCloseable(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	w := ui.NewWindow("win", "Test", font, 0)
	defer w.Dispose()

	w.SetCloseable(false)
	if w.CloseBtn().IsVisible() {
		t.Error("close button should be hidden after SetCloseable(false)")
	}

	w.SetCloseable(true)
	if !w.CloseBtn().IsVisible() {
		t.Error("close button should be visible after SetCloseable(true)")
	}
}
