package integration

import (
	"fmt"
	"testing"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

// ---------------------------------------------------------------------------
// TabBar
// ---------------------------------------------------------------------------

func TestNewTabBarDefaults(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	if tb.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if tb.TabCount() != 0 {
		t.Errorf("TabCount() = %d, want 0", tb.TabCount())
	}
	if tb.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", tb.Selected())
	}
}

func TestTabBarAddTab(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	panel1 := ui.NewComponent("panel1")
	panel2 := ui.NewComponent("panel2")

	idx0 := tb.AddTab("Tab A", panel1)
	idx1 := tb.AddTab("Tab B", panel2)

	if idx0 != 0 {
		t.Errorf("first tab index = %d, want 0", idx0)
	}
	if idx1 != 1 {
		t.Errorf("second tab index = %d, want 1", idx1)
	}
	if tb.TabCount() != 2 {
		t.Errorf("TabCount() = %d, want 2", tb.TabCount())
	}
}

func TestTabBarSwitchesPanels(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	panel1 := ui.NewComponent("panel1")
	panel2 := ui.NewComponent("panel2")
	panel3 := ui.NewComponent("panel3")

	tb.AddTab("A", panel1)
	tb.AddTab("B", panel2)
	tb.AddTab("C", panel3)

	// Initially tab 0 is selected.
	if !panel1.IsVisible() {
		t.Error("panel1 should be visible (tab 0 selected)")
	}
	if panel2.IsVisible() {
		t.Error("panel2 should be hidden")
	}
	if panel3.IsVisible() {
		t.Error("panel3 should be hidden")
	}

	// Switch to tab 1.
	tb.SetSelected(1)
	if panel1.IsVisible() {
		t.Error("panel1 should be hidden after selecting tab 1")
	}
	if !panel2.IsVisible() {
		t.Error("panel2 should be visible after selecting tab 1")
	}

	// Switch to tab 2.
	tb.SetSelected(2)
	if panel2.IsVisible() {
		t.Error("panel2 should be hidden after selecting tab 2")
	}
	if !panel3.IsVisible() {
		t.Error("panel3 should be visible after selecting tab 2")
	}
}

func TestTabBarOnChange(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	tb.AddTab("A", ui.NewComponent("p1"))
	tb.AddTab("B", ui.NewComponent("p2"))

	var got int
	tb.SetOnChange(func(idx int) { got = idx })

	tb.SetSelected(1)
	if got != 1 {
		t.Errorf("onChange got %d, want 1", got)
	}
}

func TestTabBarBindSelected(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	tb.AddTab("A", ui.NewComponent("p1"))
	tb.AddTab("B", ui.NewComponent("p2"))
	tb.AddTab("C", ui.NewComponent("p3"))

	ref := ui.NewRef(2)
	tb.BindSelected(ref)

	if tb.Selected() != 2 {
		t.Errorf("binding should sync initial value, got %d", tb.Selected())
	}

	ref.Set(0)
	ui.DefaultScheduler.Flush()
	if tb.Selected() != 0 {
		t.Errorf("reactive update should select 0, got %d", tb.Selected())
	}
}

func TestTabBarRemoveTab(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	tb.AddTab("A", ui.NewComponent("p1"))
	tb.AddTab("B", ui.NewComponent("p2"))
	tb.AddTab("C", ui.NewComponent("p3"))

	tb.RemoveTab(1)

	if tb.TabCount() != 2 {
		t.Errorf("TabCount() = %d, want 2 after remove", tb.TabCount())
	}
}

func TestTabBarRemoveSelectedTab(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	tb.AddTab("A", ui.NewComponent("p1"))
	tb.AddTab("B", ui.NewComponent("p2"))
	tb.AddTab("C", ui.NewComponent("p3"))

	tb.SetSelected(2)
	tb.RemoveTab(2)

	// Selection should adjust to last tab.
	if tb.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1 after removing selected tab", tb.Selected())
	}
}

func TestTabBarSetSelectedOutOfRange(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	tb.AddTab("A", ui.NewComponent("p1"))
	tb.SetSelected(0)

	// Out of range should be ignored.
	tb.SetSelected(5)
	if tb.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0 (out of range ignored)", tb.Selected())
	}
}

func TestTabBarNilPanel(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	// Adding a tab with nil content should not panic.
	tb.AddTab("Empty", nil)
	if tb.TabCount() != 1 {
		t.Errorf("TabCount() = %d, want 1", tb.TabCount())
	}
	tb.SetSelected(0) // should not panic
}

// ---------------------------------------------------------------------------
// TabBar — Scroll Overflow
// ---------------------------------------------------------------------------

// newOverflowTabBar creates a narrow (200px) TabBar with 10 tabs in scroll mode.
func newOverflowTabBar(t *testing.T) *ui.TabBar {
	t.Helper()
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	tb.SetSize(200, 100)
	tb.SetOverflowMode(ui.TabOverflowScroll)
	for i := 0; i < 10; i++ {
		tb.AddTab(fmt.Sprintf("T%d", i), nil)
	}
	return tb
}

func TestTabBarClipModeDefault(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	if tb.OverflowMode() != ui.TabOverflowClip {
		t.Errorf("default OverflowMode = %d, want TabOverflowClip", tb.OverflowMode())
	}

	// No arrows in clip mode.
	for i := 0; i < 10; i++ {
		tb.AddTab(fmt.Sprintf("T%d", i), nil)
	}
	if tb.LeftArrowVisible() {
		t.Error("left arrow should not be visible in clip mode")
	}
	if tb.RightArrowVisible() {
		t.Error("right arrow should not be visible in clip mode")
	}
}

func TestTabBarScrollModeArrowsAppear(t *testing.T) {
	tb := newOverflowTabBar(t)
	defer tb.Dispose()

	if tb.OverflowMode() != ui.TabOverflowScroll {
		t.Errorf("OverflowMode = %d, want TabOverflowScroll", tb.OverflowMode())
	}

	// With many overflowing tabs, right arrow should be visible.
	if !tb.RightArrowVisible() {
		t.Error("right arrow should be visible when tabs overflow")
	}
}

func TestTabBarLeftArrowHiddenAtStart(t *testing.T) {
	tb := newOverflowTabBar(t)
	defer tb.Dispose()

	// At scroll offset 0, left arrow should be hidden.
	if tb.ScrollOffset() != 0 {
		t.Errorf("initial ScrollOffset = %f, want 0", tb.ScrollOffset())
	}
	if tb.LeftArrowVisible() {
		t.Error("left arrow should be hidden at scroll start")
	}
}

func TestTabBarRightArrowHiddenAtEnd(t *testing.T) {
	tb := newOverflowTabBar(t)
	defer tb.Dispose()

	// Scroll to the very last tab.
	tb.SetSelected(9)

	// Right arrow should be hidden at the end.
	if tb.RightArrowVisible() {
		t.Error("right arrow should be hidden when scrolled to end")
	}
	// Left arrow should be visible.
	if !tb.LeftArrowVisible() {
		t.Error("left arrow should be visible when scrolled to end")
	}
}

func TestTabBarBothArrowsVisibleInMiddle(t *testing.T) {
	tb := newOverflowTabBar(t)
	defer tb.Dispose()

	// Select a middle tab to force scrolling away from both edges.
	tb.SetSelected(5)

	if !tb.LeftArrowVisible() {
		t.Error("left arrow should be visible in the middle")
	}
	if !tb.RightArrowVisible() {
		t.Error("right arrow should be visible in the middle")
	}
}

func TestTabBarScrollToTabRight(t *testing.T) {
	tb := newOverflowTabBar(t)
	defer tb.Dispose()

	tb.SetSelected(9)
	if tb.Selected() != 9 {
		t.Errorf("Selected = %d, want 9", tb.Selected())
	}
	if tb.ScrollOffset() <= 0 {
		t.Errorf("ScrollOffset = %f, want > 0 after scrolling right", tb.ScrollOffset())
	}
}

func TestTabBarScrollToTabLeft(t *testing.T) {
	tb := newOverflowTabBar(t)
	defer tb.Dispose()

	// Scroll to end, then back to start.
	tb.SetSelected(9)
	tb.SetSelected(0)
	if tb.Selected() != 0 {
		t.Errorf("Selected = %d, want 0", tb.Selected())
	}
	if tb.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset = %f, want 0 after scrolling back to start", tb.ScrollOffset())
	}
}

func TestTabBarScrollToTabNoop(t *testing.T) {
	tb := newOverflowTabBar(t)
	defer tb.Dispose()

	before := tb.ScrollOffset()
	tb.ScrollToTab(0)
	if tb.ScrollOffset() != before {
		t.Errorf("ScrollOffset changed from %f to %f, should be no-op", before, tb.ScrollOffset())
	}
}

func TestTabBarSetSelectedScrollsInScrollMode(t *testing.T) {
	tb := newOverflowTabBar(t)
	defer tb.Dispose()

	tb.SetSelected(9)
	if tb.Selected() != 9 {
		t.Errorf("SetSelected(9) resulted in Selected() = %d", tb.Selected())
	}
	// Offset should have changed.
	if tb.ScrollOffset() <= 0 {
		t.Error("SetSelected should implicitly scroll in scroll mode")
	}
}

func TestTabBarAddTabBeyondVisible(t *testing.T) {
	tb := newOverflowTabBar(t)
	defer tb.Dispose()

	offsetBefore := tb.ScrollOffset()
	tb.AddTab("T10", nil)
	tb.AddTab("T11", nil)
	if tb.TabCount() != 12 {
		t.Errorf("TabCount = %d, want 12", tb.TabCount())
	}
	// Scroll state should not be disrupted.
	if tb.ScrollOffset() != offsetBefore {
		t.Errorf("scroll offset changed from %f to %f after adding tabs", offsetBefore, tb.ScrollOffset())
	}
	// Should still be able to select any tab.
	tb.SetSelected(11)
	if tb.Selected() != 11 {
		t.Errorf("Selected = %d, want 11", tb.Selected())
	}
}

func TestTabBarRemoveLastOverflowingTabHidesArrow(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	tb.SetSize(200, 100)
	tb.SetOverflowMode(ui.TabOverflowScroll)

	// Add tabs until they overflow.
	for i := 0; i < 10; i++ {
		tb.AddTab(fmt.Sprintf("T%d", i), nil)
	}
	if !tb.RightArrowVisible() {
		t.Fatal("right arrow should be visible with 10 tabs in 200px bar")
	}

	// Remove tabs until they all fit.
	for tb.TabCount() > 1 {
		tb.RemoveTab(tb.TabCount() - 1)
	}

	// Both arrows should be hidden when tabs fit.
	if tb.LeftArrowVisible() {
		t.Error("left arrow should be hidden when tabs fit")
	}
	if tb.RightArrowVisible() {
		t.Error("right arrow should be hidden when tabs fit")
	}
}

func TestTabBarSwitchBackToClipMode(t *testing.T) {
	tb := newOverflowTabBar(t)
	defer tb.Dispose()

	// Scroll somewhere.
	tb.SetSelected(5)
	if tb.ScrollOffset() <= 0 {
		t.Fatal("precondition: should have non-zero offset")
	}

	// Switch back to clip mode.
	tb.SetOverflowMode(ui.TabOverflowClip)

	if tb.OverflowMode() != ui.TabOverflowClip {
		t.Errorf("OverflowMode = %d, want TabOverflowClip", tb.OverflowMode())
	}
	if tb.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset = %f, want 0 after switching to clip mode", tb.ScrollOffset())
	}
	if tb.LeftArrowVisible() {
		t.Error("left arrow should be hidden in clip mode")
	}
	if tb.RightArrowVisible() {
		t.Error("right arrow should be hidden in clip mode")
	}
}

func TestTabBarThemeApplied(t *testing.T) {
	resetScheduler()

	group := ui.DefaultTheme.Tabs.Group(ui.Primary)
	if group.ScrollArrowWidth <= 0 {
		t.Errorf("ScrollArrowWidth = %f, want > 0", group.ScrollArrowWidth)
	}
	if group.ScrollArrowBackground.Resolve(ui.StateDefault).Type == 0 &&
		group.ScrollArrowColor.Resolve(ui.StateDefault) == (willow.Color{}) {
		t.Error("scroll arrow theme fields should have non-zero defaults")
	}
}

func TestTabBarXMLOverflowMode(t *testing.T) {
	resetScheduler()

	xml := `<TabBar name="tabs" width="200" height="100" overflowMode="scroll" />`
	ir, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatalf("CompileXML: %v", err)
	}
	if ir.ComponentType != "TabBar" {
		t.Errorf("type = %q, want TabBar", ir.ComponentType)
	}
	found := false
	for _, attr := range ir.Attributes {
		if attr.Name == "overflowMode" && attr.Static == "scroll" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected overflowMode='scroll' attribute in compiled IR")
	}
}

func TestTabBarScrollToTabClipModeNoop(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	tb := ui.NewTabBar("tabs", font, 0)
	defer tb.Dispose()

	tb.AddTab("A", nil)
	tb.AddTab("B", nil)

	tb.ScrollToTab(1)
	if tb.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset = %f, want 0 in clip mode", tb.ScrollOffset())
	}
}
