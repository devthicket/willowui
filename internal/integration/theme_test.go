package integration

import (
	"math"
	"testing"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/core"
	interntheme "github.com/devthicket/willowui/internal/theme"
)

// ---------- UIElement / base() ----------

func TestComponentImplementsUIElement(t *testing.T) {
	c := ui.NewComponent("ui-elem")
	var _ ui.UIElement = c // compile-time check
	if c.BaseComp() != c {
		t.Error("BaseComp() should return the same component")
	}
}

func TestButtonImplementsUIElement(t *testing.T) {
	b := ui.NewButton("btn", "OK", newTestFont(), 16)
	var _ ui.UIElement = b // compile-time check
	if b.BaseComp() != &b.Component {
		t.Error("BaseComp() should return embedded component")
	}
}

func TestLabelImplementsUIElement(t *testing.T) {
	l := ui.NewLabel("lbl", "hi", newTestFont(), 16)
	var _ ui.UIElement = l // compile-time check
	if l.BaseComp() != &l.Component {
		t.Error("BaseComp() should return embedded component")
	}
}

// ---------- EffectiveTheme ----------

func TestEffectiveThemeDefaultsToDefaultTheme(t *testing.T) {
	c := ui.NewComponent("c")
	if c.EffectiveTheme() != ui.DefaultTheme {
		t.Error("orphan component should use DefaultTheme")
	}
}

func TestSetThemeOverridesDefault(t *testing.T) {
	custom := &ui.Theme{}
	c := ui.NewComponent("c")
	c.SetTheme(custom)
	if c.EffectiveTheme() != custom {
		t.Error("SetTheme should make EffectiveTheme return the custom theme")
	}
}

func TestSetThemeNilRevertsToDefault(t *testing.T) {
	custom := &ui.Theme{}
	c := ui.NewComponent("c")
	c.SetTheme(custom)
	c.SetTheme(nil)
	if c.EffectiveTheme() != ui.DefaultTheme {
		t.Error("SetTheme(nil) should revert to DefaultTheme")
	}
}

// ---------- Theme inheritance through tree ----------

func TestChildInheritsParentTheme(t *testing.T) {
	custom := &ui.Theme{}
	parent := ui.NewComponent("parent")
	parent.SetTheme(custom)

	child := ui.NewComponent("child")
	parent.AddChild(child)

	if child.EffectiveTheme() != custom {
		t.Error("child should inherit parent theme")
	}
}

func TestGrandchildInheritsAncestorTheme(t *testing.T) {
	custom := &ui.Theme{}
	root := ui.NewComponent("root")
	root.SetTheme(custom)

	mid := ui.NewComponent("mid")
	root.AddChild(mid)

	leaf := ui.NewComponent("leaf")
	mid.AddChild(leaf)

	if leaf.EffectiveTheme() != custom {
		t.Error("grandchild should inherit ancestor theme")
	}
}

func TestChildExplicitThemeOverridesParent(t *testing.T) {
	parentTheme := &ui.Theme{}
	childTheme := &ui.Theme{}

	parent := ui.NewComponent("parent")
	parent.SetTheme(parentTheme)

	child := ui.NewComponent("child")
	child.SetTheme(childTheme)
	parent.AddChild(child)

	if child.EffectiveTheme() != childTheme {
		t.Error("child explicit theme should override parent")
	}
}

func TestPropagationStopsAtExplicitTheme(t *testing.T) {
	theme1 := &ui.Theme{}
	theme2 := &ui.Theme{}

	root := ui.NewComponent("root")
	root.SetTheme(theme1)

	child := ui.NewComponent("child")
	child.SetTheme(theme2)
	root.AddChild(child)

	grandchild := ui.NewComponent("grandchild")
	child.AddChild(grandchild)

	// grandchild should inherit from child (theme2), not root (theme1)
	if grandchild.EffectiveTheme() != theme2 {
		t.Error("propagation should stop at child's explicit theme")
	}

	// Changing root's theme should not affect grandchild
	theme3 := &ui.Theme{}
	root.SetTheme(theme3)
	if grandchild.EffectiveTheme() != theme2 {
		t.Error("grandchild should still inherit from child, not root")
	}
}

// ---------- Theme propagation triggers onThemeChange ----------

func TestSetThemePropagatesOnThemeChange(t *testing.T) {
	called := 0
	parent := ui.NewComponent("parent")
	child := ui.NewComponent("child")
	child.SetOnThemeChangeForTest(func() { called++ })
	parent.AddChild(child)

	custom := &ui.Theme{}
	called = 0
	parent.SetTheme(custom)

	if called != 1 {
		t.Errorf("onThemeChange called %d times, want 1", called)
	}
}

func TestAddChildPropagatesTheme(t *testing.T) {
	custom := &ui.Theme{}
	parent := ui.NewComponent("parent")
	parent.SetTheme(custom)

	called := 0
	child := ui.NewComponent("child")
	child.SetOnThemeChangeForTest(func() { called++ })

	parent.AddChild(child)

	if called != 1 {
		t.Errorf("onThemeChange called %d times on AddChild, want 1", called)
	}
	if child.EffectiveTheme() != custom {
		t.Error("child should have parent's theme after AddChild")
	}
}

func TestReparentUpdatesTheme(t *testing.T) {
	theme1 := &ui.Theme{}
	theme2 := &ui.Theme{}

	parent1 := ui.NewComponent("p1")
	parent1.SetTheme(theme1)
	parent2 := ui.NewComponent("p2")
	parent2.SetTheme(theme2)

	child := ui.NewComponent("child")
	parent1.AddChild(child)
	if child.EffectiveTheme() != theme1 {
		t.Error("child should inherit theme1 from parent1")
	}

	parent2.AddChild(child)
	if child.EffectiveTheme() != theme2 {
		t.Error("child should inherit theme2 from parent2 after reparent")
	}
}

// ---------- Variant ----------

func TestVariantDefaultIsPrimary(t *testing.T) {
	c := ui.NewComponent("c")
	if c.Variant() != ui.Primary {
		t.Error("default variant should be Primary")
	}
}

func TestSetVariantTriggersOnThemeChange(t *testing.T) {
	called := 0
	c := ui.NewComponent("c")
	c.SetOnThemeChangeForTest(func() { called++ })
	c.SetVariant(ui.Danger)

	if c.Variant() != ui.Danger {
		t.Error("variant should be Danger")
	}
	if called != 1 {
		t.Errorf("onThemeChange called %d times, want 1", called)
	}
}

func TestSetVariantSameValueNoOp(t *testing.T) {
	called := 0
	c := ui.NewComponent("c")
	c.SetOnThemeChangeForTest(func() { called++ })
	c.SetVariant(ui.Primary) // same as default

	if called != 0 {
		t.Errorf("onThemeChange called %d times for same variant, want 0", called)
	}
}

// ---------- Concrete widgets respond to theme changes ----------

func TestButtonUpdatesVisualsOnThemeChange(t *testing.T) {
	wantBg := willow.RGBA(1, 0, 0, 1)
	custom := &ui.Theme{
		Button: ui.ButtonConfig{
			Primary: ui.ButtonGroup{
				Background: ui.NewSolidBgPropUniform(wantBg),
				TextColor:  ui.NewColorPropUniform(willow.RGBA(0, 1, 0, 1)),
				Padding:    ui.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
			},
		},
	}
	b := ui.NewButton("btn", "Click", newTestFont(), 16)
	b.SetTheme(custom)

	if b.BgNode().Color() != wantBg {
		t.Errorf("button background = %v, want %v", b.BgNode().Color(), wantBg)
	}
}

func TestToggleUpdatesVisualsOnThemeChange(t *testing.T) {
	resetScheduler()
	wantTrack := willow.RGBA(0.5, 0.5, 0.5, 1)
	custom := &ui.Theme{
		Toggle: ui.ToggleConfig{
			Primary: ui.ToggleGroup{
				TrackColor: ui.NewColorPropUniform(wantTrack),
				ThumbColor: ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
			},
		},
	}
	tgl := ui.NewToggle("tgl")
	tgl.SetTheme(custom)

	if tgl.TrackNode().Color() != wantTrack {
		t.Errorf("toggle track = %v, want %v", tgl.TrackNode().Color(), wantTrack)
	}
}

func TestLabelThemeChangesColor(t *testing.T) {
	wantColor := willow.RGBA(0, 0, 1, 1)
	custom := &ui.Theme{
		Label: ui.LabelConfig{
			Primary: ui.LabelGroup{
				TextColor: ui.NewColorPropUniform(wantColor),
			},
		},
	}
	l := ui.NewLabel("lbl", "test", newTestFont(), 16)
	l.SetTheme(custom)

	if l.TextNode().TextBlock.Color != wantColor {
		t.Errorf("label color = %v, want %v", l.TextNode().TextBlock.Color, wantColor)
	}
}

// ---------- SliceBackground constructor ----------

func TestSliceBackgroundConstructor(t *testing.T) {
	ns := &ui.NineSlice{
		Insets: ui.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4},
	}
	bg := ui.SliceBackground(ns)
	if bg.Type != ui.BgNineSlice {
		t.Errorf("Type = %d, want BgNineSlice (%d)", bg.Type, ui.BgNineSlice)
	}
	if bg.Slice != ns {
		t.Error("Slice should point to the provided NineSlice")
	}
}

// ---------- initBackground / applyBackground / resizeBackground ----------

func TestInitBackground(t *testing.T) {
	c := ui.NewComponent("test-init-bg")
	defer c.Dispose()

	c.InitBackgroundForTest("test-init-bg")
	if c.BgNode() == nil {
		t.Fatal("bgNode should not be nil after initBackground")
	}
}

func TestApplyBackground_Solid(t *testing.T) {
	c := ui.NewComponent("test-apply-solid")
	defer c.Dispose()

	c.InitBackgroundForTest("test-apply-solid")
	red := willow.RGBA(1, 0, 0, 1)
	c.ApplyBackgroundForTest(ui.SolidBackground(red))
	if c.BgNode().Color() != red {
		t.Errorf("bgNode.Color = %v, want %v", c.BgNode().Color(), red)
	}
}

func TestApplyBackground_None(t *testing.T) {
	c := ui.NewComponent("test-apply-none")
	defer c.Dispose()

	c.InitBackgroundForTest("test-apply-none")
	c.ApplyBackgroundForTest(ui.SolidBackground(willow.RGBA(1, 0, 0, 1)))
	c.ApplyBackgroundForTest(ui.Background{Type: ui.BgNone})
	zero := willow.Color{}
	if c.BgNode().Color() != zero {
		t.Errorf("bgNode.Color = %v, want zero", c.BgNode().Color())
	}
}

func TestResizeBackground(t *testing.T) {
	c := ui.NewComponent("test-resize-bg")
	defer c.Dispose()

	c.InitBackgroundForTest("test-resize-bg")
	c.ResizeBackgroundForTest(200, 100)
	if c.BgNode().ScaleX() != 200 || c.BgNode().ScaleY() != 100 {
		t.Errorf("bgNode scale = %fx%f, want 200x100", c.BgNode().ScaleX(), c.BgNode().ScaleY())
	}
}

// ---------- initBorder / applyBorder / resizeBorder ----------

func TestInitBorder(t *testing.T) {
	c := ui.NewComponent("test-init-border")
	defer c.Dispose()

	c.InitBorderForTest("test-init-border")
	if c.BorderTop() == nil || c.BorderRight() == nil || c.BorderBot() == nil || c.BorderLeft() == nil {
		t.Fatal("border nodes should not be nil after initBorder")
	}
	// All should be invisible by default.
	if c.BorderTop().Visible() || c.BorderRight().Visible() {
		t.Error("border nodes should be invisible by default")
	}
}

func TestApplyBorder_SolidBg(t *testing.T) {
	c := ui.NewComponent("test-apply-border")
	defer c.Dispose()

	c.InitBorderForTest("test-apply-border")
	c.Width = 100
	c.Height = 50
	green := willow.RGBA(0, 1, 0, 1)
	c.ApplyBorderForTest(green, 2, ui.Background{Type: ui.BgSolid})

	if !c.BorderTop().Visible() {
		t.Error("border should be visible with solid bg")
	}
	if c.BorderTop().Color() != green {
		t.Errorf("border color = %v, want green", c.BorderTop().Color())
	}
	if c.BorderWidth() != 2 {
		t.Errorf("borderWidth = %f, want 2", c.BorderWidth())
	}
}

func TestApplyBorder_NineSliceHidesBorder(t *testing.T) {
	c := ui.NewComponent("test-border-9s")
	defer c.Dispose()

	c.InitBorderForTest("test-border-9s")
	green := willow.RGBA(0, 1, 0, 1)
	c.ApplyBorderForTest(green, 2, ui.Background{Type: ui.BgSolid})
	if !c.BorderTop().Visible() {
		t.Fatal("border should be visible before nine-slice")
	}
	// Now apply border with nine-slice bg — should hide borders.
	c.ApplyBorderForTest(green, 2, ui.Background{Type: ui.BgNineSlice})
	if c.BorderTop().Visible() || c.BorderRight().Visible() || c.BorderBot().Visible() || c.BorderLeft().Visible() {
		t.Error("border should be hidden when background is BgNineSlice")
	}
}

func TestApplyBorder_TransparentColorHidesBorder(t *testing.T) {
	c := ui.NewComponent("test-border-transparent")
	defer c.Dispose()

	c.InitBorderForTest("test-border-transparent")
	transparent := willow.RGBA(0, 0, 0, 0)
	c.ApplyBorderForTest(transparent, 2, ui.Background{Type: ui.BgSolid})
	if c.BorderTop().Visible() {
		t.Error("border should be hidden with transparent color")
	}
}

func TestResizeBorder(t *testing.T) {
	c := ui.NewComponent("test-resize-border")
	defer c.Dispose()

	c.InitBorderForTest("test-resize-border")
	green := willow.RGBA(0, 1, 0, 1)
	c.ApplyBorderForTest(green, 2, ui.Background{Type: ui.BgSolid})
	c.ResizeBorderForTest(100, 80)

	// Top border: full width, borderWidth tall.
	if c.BorderTop().ScaleX() != 100 || c.BorderTop().ScaleY() != 2 {
		t.Errorf("top scale = %fx%f, want 100x2", c.BorderTop().ScaleX(), c.BorderTop().ScaleY())
	}
	// Right border at x=98.
	if c.BorderRight().X() != 98 {
		t.Errorf("right X = %f, want 98", c.BorderRight().X())
	}
	// Bottom border at y=78.
	if c.BorderBot().Y() != 78 {
		t.Errorf("bottom Y = %f, want 78", c.BorderBot().Y())
	}
}

func TestResizeBorder_NilSafe(t *testing.T) {
	c := ui.NewComponent("test-resize-nil")
	defer c.Dispose()

	// Should not panic when border nodes are nil.
	c.ResizeBorderForTest(100, 100)
}

// ---------- Panel theme auto-apply ----------

func TestPanelAppliesThemeOnConstruction(t *testing.T) {
	wantBg := willow.RGBA(0.5, 0.5, 0.5, 1)
	wantBorder := willow.RGBA(1, 0, 0, 1)
	custom := &ui.Theme{
		Panel: ui.PanelConfig{
			Primary: ui.PanelGroup{
				Background:  ui.NewSolidBgPropUniform(wantBg),
				Border:      ui.NewColorPropUniform(wantBorder),
				BorderWidth: 3,
			},
		},
	}

	// Set DefaultTheme temporarily to test construction.
	old := ui.DefaultTheme
	ui.DefaultTheme = custom
	defer func() { ui.DefaultTheme = old }()

	p := ui.NewPanel("themed-panel")
	defer p.Dispose()
	p.SetSize(200, 100)

	if p.BgNode().Color() != wantBg {
		t.Errorf("panel bg = %v, want %v", p.BgNode().Color(), wantBg)
	}
	if p.BorderTop().Color() != wantBorder {
		t.Errorf("panel border color = %v, want %v", p.BorderTop().Color(), wantBorder)
	}
	if !p.BorderTop().Visible() {
		t.Error("panel border should be visible")
	}
}

func TestPanelSetBackgroundOverridesTheme(t *testing.T) {
	themeBg := willow.RGBA(0.2, 0.2, 0.2, 1)
	custom := &ui.Theme{
		Panel: ui.PanelConfig{
			Primary: ui.PanelGroup{
				Background: ui.NewSolidBgPropUniform(themeBg),
			},
		},
	}

	p := ui.NewPanel("panel-override-bg")
	defer p.Dispose()
	p.SetTheme(custom)

	// Verify theme applied.
	if p.BgNode().Color() != themeBg {
		t.Fatalf("expected theme bg %v, got %v", themeBg, p.BgNode().Color())
	}

	// Manual override.
	manualBg := willow.RGBA(1, 0, 0, 1)
	p.SetBackground(manualBg)
	if p.BgNode().Color() != manualBg {
		t.Errorf("manual bg = %v, want %v", p.BgNode().Color(), manualBg)
	}

	// Theme change should NOT overwrite the manual override.
	p.SetTheme(&ui.Theme{
		Panel: ui.PanelConfig{
			Primary: ui.PanelGroup{
				Background: ui.NewSolidBgPropUniform(willow.RGBA(0, 1, 0, 1)),
			},
		},
	})
	if p.BgNode().Color() != manualBg {
		t.Errorf("after theme change, bg = %v, want manual override %v", p.BgNode().Color(), manualBg)
	}
}

func TestPanelSetBorderOverridesTheme(t *testing.T) {
	themeBorder := willow.RGBA(0.5, 0.5, 0.5, 1)
	custom := &ui.Theme{
		Panel: ui.PanelConfig{
			Primary: ui.PanelGroup{
				Background:  ui.NewSolidBgPropUniform(willow.RGBA(0.1, 0.1, 0.1, 1)),
				Border:      ui.NewColorPropUniform(themeBorder),
				BorderWidth: 2,
			},
		},
	}

	p := ui.NewPanel("panel-override-border")
	defer p.Dispose()
	p.SetSize(100, 100)
	p.SetTheme(custom)

	// Verify theme border applied.
	if p.BorderTop().Color() != themeBorder {
		t.Fatalf("expected theme border %v, got %v", themeBorder, p.BorderTop().Color())
	}

	// Manual override.
	manualBorder := willow.RGBA(1, 1, 0, 1)
	p.SetBorder(manualBorder, 3)
	if p.BorderTop().Color() != manualBorder {
		t.Errorf("manual border = %v, want %v", p.BorderTop().Color(), manualBorder)
	}

	// Theme change should NOT overwrite the manual border.
	p.SetTheme(&ui.Theme{
		Panel: ui.PanelConfig{
			Primary: ui.PanelGroup{
				Background:  ui.NewSolidBgPropUniform(willow.RGBA(0.1, 0.1, 0.1, 1)),
				Border:      ui.NewColorPropUniform(willow.RGBA(0, 0, 1, 1)),
				BorderWidth: 1,
			},
		},
	})
	if p.BorderTop().Color() != manualBorder {
		t.Errorf("after theme change, border = %v, want manual override %v", p.BorderTop().Color(), manualBorder)
	}
}

// ---------- Button border ----------

func TestButtonBorderFromTheme(t *testing.T) {
	wantBorder := willow.RGBA(0.3, 0.3, 0.35, 1)
	custom := &ui.Theme{
		Button: ui.ButtonConfig{
			Primary: ui.ButtonGroup{
				Background:  ui.NewSolidBgPropUniform(willow.RGBA(0.2, 0.4, 0.8, 1)),
				TextColor:   ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				Border:      ui.NewColorPropUniform(wantBorder),
				BorderWidth: 1,
				Padding:     ui.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
			},
		},
	}

	b := ui.NewButton("btn-border", "Test", newTestFont(), 16)
	defer b.Dispose()
	b.SetTheme(custom)

	if b.BorderTop().Color() != wantBorder {
		t.Errorf("button border = %v, want %v", b.BorderTop().Color(), wantBorder)
	}
	if !b.BorderTop().Visible() {
		t.Error("button border should be visible")
	}
}

func TestButtonBorderHiddenWhenTransparent(t *testing.T) {
	custom := &ui.Theme{
		Button: ui.ButtonConfig{
			Primary: ui.ButtonGroup{
				Background:  ui.NewSolidBgPropUniform(willow.RGBA(0.2, 0.4, 0.8, 1)),
				TextColor:   ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				BorderWidth: 1,
				// Border is zero-value (transparent) — should remain hidden.
				Padding: ui.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
			},
		},
	}

	b := ui.NewButton("btn-no-border", "Test", newTestFont(), 16)
	defer b.Dispose()
	b.SetTheme(custom)

	if b.BorderTop().Visible() {
		t.Error("button border should be hidden when border color is transparent")
	}
}

// ---------- Corner Radius ----------

func TestCornerRadiusBgPoly(t *testing.T) {
	custom := &ui.Theme{
		Button: ui.ButtonConfig{
			Primary: ui.ButtonGroup{
				Background:   ui.NewSolidBgPropUniform(willow.RGBA(0.2, 0.4, 0.8, 1)),
				TextColor:    ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				BorderWidth:  1,
				Border:       ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				CornerRadius: 8,
				Padding:      ui.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
			},
		},
	}

	b := ui.NewButton("btn-rounded", "Test", newTestFont(), 16)
	defer b.Dispose()
	b.SetTheme(custom)

	// bgPoly should be created and visible.
	if b.BgPoly() == nil {
		t.Fatal("bgPoly should be created when cornerRadius > 0")
	}
	if !b.BgPoly().Visible() {
		t.Error("bgPoly should be visible when cornerRadius > 0")
	}
	// bgNode (flat sprite) should be hidden.
	if b.BgNode().Visible() {
		t.Error("bgNode should be hidden when cornerRadius > 0")
	}
}

func TestCornerRadiusZeroUsesSprite(t *testing.T) {
	custom := &ui.Theme{
		Button: ui.ButtonConfig{
			Primary: ui.ButtonGroup{
				Background:   ui.NewSolidBgPropUniform(willow.RGBA(0.2, 0.4, 0.8, 1)),
				TextColor:    ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				BorderWidth:  1,
				Border:       ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				CornerRadius: 0,
				Padding:      ui.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
			},
		},
	}

	b := ui.NewButton("btn-sharp", "Test", newTestFont(), 16)
	defer b.Dispose()
	b.SetTheme(custom)

	// bgNode should be visible, bgPoly nil or hidden.
	if !b.BgNode().Visible() {
		t.Error("bgNode should be visible when cornerRadius == 0")
	}
	if b.BgPoly() != nil && b.BgPoly().Visible() {
		t.Error("bgPoly should be hidden when cornerRadius == 0")
	}
}

func TestCornerRadiusBorderPoly(t *testing.T) {
	custom := &ui.Theme{
		Button: ui.ButtonConfig{
			Primary: ui.ButtonGroup{
				Background:   ui.NewSolidBgPropUniform(willow.RGBA(0.2, 0.4, 0.8, 1)),
				TextColor:    ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				BorderWidth:  2,
				Border:       ui.NewColorPropUniform(willow.RGBA(1, 0, 0, 1)),
				CornerRadius: 8,
				Padding:      ui.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
			},
		},
	}

	b := ui.NewButton("btn-border-rounded", "Test", newTestFont(), 16)
	defer b.Dispose()
	b.SetTheme(custom)

	// borderPoly should be created and visible.
	if b.BorderPoly() == nil {
		t.Fatal("borderPoly should be created when cornerRadius > 0 and border visible")
	}
	if !b.BorderPoly().Visible() {
		t.Error("borderPoly should be visible when cornerRadius > 0")
	}
	// Edge sprites should be hidden.
	if b.BorderTop().Visible() || b.BorderRight().Visible() || b.BorderBot().Visible() || b.BorderLeft().Visible() {
		t.Error("edge border sprites should be hidden when cornerRadius > 0")
	}
}

func TestCornerRadiusToggleBackToZero(t *testing.T) {
	roundedTheme := &ui.Theme{
		Button: ui.ButtonConfig{
			Primary: ui.ButtonGroup{
				Background:   ui.NewSolidBgPropUniform(willow.RGBA(0.2, 0.4, 0.8, 1)),
				TextColor:    ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				BorderWidth:  2,
				Border:       ui.NewColorPropUniform(willow.RGBA(1, 0, 0, 1)),
				CornerRadius: 8,
				Padding:      ui.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
			},
		},
	}
	sharpTheme := &ui.Theme{
		Button: ui.ButtonConfig{
			Primary: ui.ButtonGroup{
				Background:   ui.NewSolidBgPropUniform(willow.RGBA(0.2, 0.4, 0.8, 1)),
				TextColor:    ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				BorderWidth:  2,
				Border:       ui.NewColorPropUniform(willow.RGBA(1, 0, 0, 1)),
				CornerRadius: 0,
				Padding:      ui.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
			},
		},
	}

	b := ui.NewButton("btn-toggle", "Test", newTestFont(), 16)
	defer b.Dispose()

	// Start rounded.
	b.SetTheme(roundedTheme)
	if b.BgPoly() == nil || !b.BgPoly().Visible() {
		t.Fatal("bgPoly should be visible with rounded theme")
	}
	if b.BorderPoly() == nil || !b.BorderPoly().Visible() {
		t.Fatal("borderPoly should be visible with rounded theme")
	}

	// Switch to sharp.
	b.SetTheme(sharpTheme)
	if b.BgPoly().Visible() {
		t.Error("bgPoly should be hidden after switching to sharp theme")
	}
	if b.BorderPoly().Visible() {
		t.Error("borderPoly should be hidden after switching to sharp theme")
	}
	if !b.BgNode().Visible() {
		t.Error("bgNode should be visible after switching to sharp theme")
	}
	if !b.BorderTop().Visible() {
		t.Error("edge border sprites should be visible after switching to sharp theme")
	}
}

// ---------- TextInput border ----------

func TestTextInputBorderFromTheme(t *testing.T) {
	resetScheduler()
	wantBorder := willow.RGBA(1, 0, 0, 1)
	custom := &ui.Theme{
		TextInput: ui.TextInputConfig{
			Primary: ui.TextInputGroup{
				Background:  ui.NewSolidBgPropUniform(willow.RGBA(0.1, 0.1, 0.1, 1)),
				TextColor:   ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				CursorColor: ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				Border:      ui.NewColorPropUniform(wantBorder),
				BorderWidth: 2,
				Padding:     ui.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4},
			},
		},
	}

	ti := ui.NewTextInput("ti-border", newTestFont(), 12)
	defer ti.Dispose()
	ti.SetTheme(custom)

	if ti.BorderTop() == nil {
		t.Fatal("borderTop should not be nil")
	}
	if ti.BorderTop().Color() != wantBorder {
		t.Errorf("textinput border = %v, want %v", ti.BorderTop().Color(), wantBorder)
	}
	if !ti.BorderTop().Visible() {
		t.Error("textinput border should be visible")
	}
}

// ---------- TextArea border ----------

func TestTextAreaBorderFromTheme(t *testing.T) {
	resetScheduler()
	wantBorder := willow.RGBA(0, 1, 0, 1)
	custom := &ui.Theme{
		TextArea: ui.TextAreaConfig{
			Primary: ui.TextAreaGroup{
				Background:  ui.NewSolidBgPropUniform(willow.RGBA(0.1, 0.1, 0.1, 1)),
				TextColor:   ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				CursorColor: ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				Border:      ui.NewColorPropUniform(wantBorder),
				BorderWidth: 2,
				Padding:     ui.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4},
			},
		},
		ScrollBar: ui.ScrollBarConfig{
			Primary: ui.ScrollBarGroup{
				Background:      ui.NewSolidBgPropUniform(willow.RGBA(0.1, 0.1, 0.1, 1)),
				ThumbBackground: ui.NewSolidBgPropUniform(willow.RGBA(0.5, 0.5, 0.5, 1)),
			},
		},
	}

	ta := ui.NewTextArea("ta-border", newTestFont(), 12)
	defer ta.Dispose()
	ta.SetTheme(custom)

	if ta.BorderTop() == nil {
		t.Fatal("borderTop should not be nil")
	}
	if ta.BorderTop().Color() != wantBorder {
		t.Errorf("textarea border = %v, want %v", ta.BorderTop().Color(), wantBorder)
	}
	if !ta.BorderTop().Visible() {
		t.Error("textarea border should be visible")
	}
}

// ---------- MeterBar border ----------

func TestMeterBarBorderFromTheme(t *testing.T) {
	resetScheduler()
	wantBorder := willow.RGBA(0, 0, 1, 1)
	custom := &ui.Theme{
		MeterBar: ui.MeterBarConfig{
			Primary: ui.MeterBarGroup{
				Background:     ui.NewSolidBgPropUniform(willow.RGBA(0.2, 0.2, 0.2, 1)),
				FillBackground: ui.NewSolidBgPropUniform(willow.RGBA(0, 1, 0, 1)),
				TextColor:      ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				Border:         ui.NewColorPropUniform(wantBorder),
				BorderWidth:    1,
			},
		},
	}

	old := ui.DefaultTheme
	ui.DefaultTheme = custom
	defer func() { ui.DefaultTheme = old }()

	pb := ui.NewProgressBar("pb-border")
	defer pb.Dispose()
	pb.SetSize(200, 20)

	if pb.BorderTop() == nil {
		t.Fatal("borderTop should not be nil")
	}
	if pb.BorderTop().Color() != wantBorder {
		t.Errorf("meterbar border = %v, want %v", pb.BorderTop().Color(), wantBorder)
	}
	if !pb.BorderTop().Visible() {
		t.Error("meterbar border should be visible")
	}
}

// ---------- Window border ----------

func TestWindowBorderFromTheme(t *testing.T) {
	resetScheduler()
	wantBorder := willow.RGBA(1, 1, 0, 1)
	custom := &ui.Theme{
		Window: ui.WindowConfig{
			Primary: ui.WindowGroup{
				Background:        ui.NewSolidBgPropUniform(willow.RGBA(0.1, 0.1, 0.1, 1)),
				TitleBackground:   ui.NewSolidBgPropUniform(willow.RGBA(0.3, 0.3, 0.3, 1)),
				TitleTextColor:    ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				ResizeHandleColor: ui.NewColorPropUniform(willow.RGBA(0.5, 0.5, 0.5, 1)),
				Border:            ui.NewColorPropUniform(wantBorder),
				BorderWidth:       2,
			},
		},
		Button: ui.ButtonConfig{
			Primary: ui.ButtonGroup{
				Background: ui.NewSolidBgPropUniform(willow.RGBA(0.2, 0.2, 0.2, 1)),
				TextColor:  ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				Padding:    ui.Insets{Top: 4, Right: 8, Bottom: 4, Left: 8},
			},
		},
		Label: ui.LabelConfig{
			Primary: ui.LabelGroup{
				TextColor: ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
			},
		},
		Panel: ui.PanelConfig{
			Primary: ui.PanelGroup{
				Background: ui.NewSolidBgPropUniform(willow.Color{}),
			},
		},
	}

	old := ui.DefaultTheme
	ui.DefaultTheme = custom
	defer func() { ui.DefaultTheme = old }()

	w := ui.NewWindow("win-border", "Test", newTestFont(), 12)
	defer w.Dispose()

	if w.BorderTop() == nil {
		t.Fatal("borderTop should not be nil")
	}
	if w.BorderTop().Color() != wantBorder {
		t.Errorf("window border = %v, want %v", w.BorderTop().Color(), wantBorder)
	}
	if !w.BorderTop().Visible() {
		t.Error("window border should be visible")
	}
}

// ---------- FloatProperty ----------

func TestFloatPropertyResolveUniform(t *testing.T) {
	fp := ui.NewFloatPropUniform(3.5)
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if fp.Resolve(s) != 3.5 {
			t.Errorf("state %d: got %f, want 3.5", s, fp.Resolve(s))
		}
	}
}

func TestFloatPropertyResolveStates(t *testing.T) {
	fp := ui.NewFloatPropStates(map[ui.ComponentState]float64{
		ui.StateDefault: 0,
		ui.StateHover:   -1,
		ui.StateActive:  2,
	})
	if fp.Resolve(ui.StateDefault) != 0 {
		t.Errorf("default: got %f, want 0", fp.Resolve(ui.StateDefault))
	}
	if fp.Resolve(ui.StateHover) != -1 {
		t.Errorf("hover: got %f, want -1", fp.Resolve(ui.StateHover))
	}
	if fp.Resolve(ui.StateActive) != 2 {
		t.Errorf("active: got %f, want 2", fp.Resolve(ui.StateActive))
	}
}

func TestFloatPropertyFallbackChain(t *testing.T) {
	// Only set default and hover; active should fallback to hover.
	fp := ui.NewFloatPropStates(map[ui.ComponentState]float64{
		ui.StateDefault: 0,
		ui.StateHover:   -2,
	})
	// StateActive fallback: hover → default. Should get hover's value.
	if fp.Resolve(ui.StateActive) != -2 {
		t.Errorf("active (fallback from hover): got %f, want -2", fp.Resolve(ui.StateActive))
	}
	// StateDisabled fallback: default. Should get 0.
	if fp.Resolve(ui.StateDisabled) != 0 {
		t.Errorf("disabled (fallback from default): got %f, want 0", fp.Resolve(ui.StateDisabled))
	}
}

func TestFloatPropertyZeroIsValid(t *testing.T) {
	// Zero is a valid offset, not a sentinel.
	fp := ui.NewFloatPropStates(map[ui.ComponentState]float64{
		ui.StateDefault: 0,
		ui.StateHover:   5,
	})
	if fp.Resolve(ui.StateDefault) != 0 {
		t.Errorf("default: got %f, want 0", fp.Resolve(ui.StateDefault))
	}
	if fp.Resolve(ui.StateHover) != 5 {
		t.Errorf("hover: got %f, want 5", fp.Resolve(ui.StateHover))
	}
}

func TestResolveFloatFallbacksNaNBecomesZero(t *testing.T) {
	// If only StateDefault is NaN (nothing set), it should become 0.
	var fp ui.FloatProperty
	for i := range fp {
		fp[i] = math.NaN()
	}
	interntheme.ResolveFloatFallbacks(&fp)
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if fp[s] != 0 {
			t.Errorf("state %d: got %f after fallback, want 0", s, fp[s])
		}
	}
}

// ---------- Button per-state offsets ----------

func TestButtonOffsetFromTheme(t *testing.T) {
	custom := &ui.Theme{
		Button: ui.ButtonConfig{
			Primary: ui.ButtonGroup{
				Background: ui.NewSolidBgPropUniform(willow.RGBA(0.2, 0.4, 0.8, 1)),
				TextColor:  ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				Padding:    ui.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
				OffsetY: ui.NewFloatPropStates(map[ui.ComponentState]float64{
					ui.StateDefault: 0,
					ui.StateHover:   -1,
					ui.StateActive:  2,
				}),
			},
		},
	}

	b := ui.NewButton("btn-offset", "Test", newTestFont(), 16)
	defer b.Dispose()
	b.SetTheme(custom)

	// Default state: offset should be 0.
	if b.OffsetY != 0 {
		t.Errorf("default OffsetY = %f, want 0", b.OffsetY)
	}
}

func TestButtonTextOffsetFromTheme(t *testing.T) {
	custom := &ui.Theme{
		Button: ui.ButtonConfig{
			Primary: ui.ButtonGroup{
				Background: ui.NewSolidBgPropUniform(willow.RGBA(0.2, 0.4, 0.8, 1)),
				TextColor:  ui.NewColorPropUniform(willow.RGBA(1, 1, 1, 1)),
				Padding:    ui.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
				TextOffsetY: ui.NewFloatPropStates(map[ui.ComponentState]float64{
					ui.StateDefault: 0,
					ui.StateHover:   -2,
					ui.StateActive:  3,
				}),
			},
		},
	}

	b := ui.NewButton("btn-toffset", "Test", newTestFont(), 16)
	defer b.Dispose()
	b.SetTheme(custom)

	// Default state: text offset should be 0, label centered normally.
	if b.TextOY() != 0 {
		t.Errorf("default textOY = %f, want 0", b.TextOY())
	}

	// Label should be at exact center (no offset).
	wantLY := (b.Height - b.LabelLabel().Height) / 2
	if b.LabelLabel().Node().Y() != wantLY {
		t.Errorf("label Y = %f, want %f (centered)", b.LabelLabel().Node().Y(), wantLY)
	}
}
