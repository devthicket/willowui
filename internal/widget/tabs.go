package widget

import (
	"fmt"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// ---------------------------------------------------------------------------
// TabOverflowMode
// ---------------------------------------------------------------------------

// TabOverflowMode controls what happens when tabs overflow the bar width.
type TabOverflowMode int

const (
	TabOverflowClip   TabOverflowMode = iota // default — tabs are clipped at the edge
	TabOverflowScroll                        // scroll arrows appear on overflow
)

// ---------------------------------------------------------------------------
// tabEntry
// ---------------------------------------------------------------------------

// tabEntry pairs a tab button with its content panel.
type tabEntry struct {
	button *Button
	panel  *Component
}

// ---------------------------------------------------------------------------
// TabBar
// ---------------------------------------------------------------------------

// TabBar is a horizontal row of tab buttons. Selecting a tab shows its
// associated panel and hides all others.
type TabBar struct {
	Component
	tabs        []*tabEntry
	selected    *Ref[int]
	watch       WatchHandle
	onChange    func(int)
	source      *sg.FontFamily
	font        *sg.FontFamily
	displaySize float64

	bar   *Component // horizontal row of buttons
	body  *Component // container for panels
	barBg *sg.Node

	// Scroll overflow fields.
	overflowMode    TabOverflowMode
	barScrollOffset float64
	leftArrow       *Component // nil until scroll mode activated
	rightArrow      *Component // nil until scroll mode activated
	leftGlyph       *sg.Node   // sprite for left arrow glyph
	rightGlyph      *sg.Node   // sprite for right arrow glyph
	maskRoot        *sg.Node   // reused mask container
	maskSprite      *sg.Node   // reused mask sprite
}

// NewTabBar creates a new tab bar with the given font source and display size.
func NewTabBar(name string, source *sg.FontFamily, displaySize float64) *TabBar {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	tb := &TabBar{
		selected:    NewRef(0),
		source:      source,
		font:        font,
		displaySize: displaySize,
	}
	initComponent(&tb.Component, name)

	// Bar background.
	tb.barBg = sg.NewSprite(name+"-bar-bg", sg.TextureRegion{})
	tb.node.AddChild(tb.barBg)

	// Bar container for buttons.
	tb.bar = NewComponent(name + "-bar")
	tb.bar.Layout = LayoutHBox
	tb.bar.Spacing = 2
	tb.Component.AddChild(tb.bar)

	// Body container for panels.
	tb.body = NewComponent(name + "-body")
	tb.body.Layout = LayoutNone
	tb.Component.AddChild(tb.body)

	tb.onThemeChange = func() { tb.applyThemeColors() }
	tb.applyThemeColors()

	// Default size.
	tb.SetSize(400, 300)

	return tb
}

func (tb *TabBar) applyThemeColors() {
	group := tb.EffectiveTheme().Tabs.Group(tb.Variant())
	tb.barBg.SetColor(group.BarBackground.Resolve(StateDefault).Color)
	tb.updateButtonVisuals()
	tb.styleArrowButtons()
	tb.MarkDrawDirty()
}

// AddTab adds a new tab with the given label and content panel, returning
// the tab index.
func (tb *TabBar) AddTab(label string, content *Component) int {
	idx := len(tb.tabs)

	btn := NewButton(fmt.Sprintf("%s-tab-%d", tb.node.Name, idx), label, tb.source, tb.displaySize)
	btn.SetSize(btn.Width, 30)
	tabIdx := idx
	btn.SetOnClick(func() {
		tb.SetSelected(tabIdx)
	})

	// Override visual-state and theme hooks so tab styling is re-applied
	// after Button's own UpdateVisuals (which would otherwise override
	// our colors with the button theme).
	btn.onVisualStateChange = func() {
		btn.UpdateVisuals()
		tb.styleTabButton(tabIdx)
	}
	btn.onThemeChange = func() {
		btn.UpdateVisuals()
		tb.styleTabButton(tabIdx)
	}

	tb.bar.AddChild(btn)

	// Add panel to body.
	if content != nil {
		content.SetPosition(0, 0)
		tb.body.AddChild(content)
	}

	entry := &tabEntry{
		button: btn,
		panel:  content,
	}
	tb.tabs = append(tb.tabs, entry)

	// Show/hide based on current selection.
	tb.updatePanelVisibility()
	tb.updateButtonVisuals()
	tb.UpdateLayout()
	tb.recalcScrollBar()

	return idx
}

// AddTabPage creates a new page component with the given layout, spacing, and
// padding, adds it as a tab with the given label, and returns the page and its
// tab index. The page is sized to match the tab body.
func (tb *TabBar) AddTabPage(label string, layout LayoutMode, spacing float64, padding Insets) (*Component, int) {
	page := NewComponent(label + "-page")
	page.Layout = layout
	page.Spacing = spacing
	page.Padding = padding
	page.Width = tb.body.Width
	page.Height = tb.body.Height
	idx := tb.AddTab(label, page)
	return page, idx
}

// RemoveTab removes a tab at the given index.
func (tb *TabBar) RemoveTab(index int) {
	if index < 0 || index >= len(tb.tabs) {
		return
	}

	entry := tb.tabs[index]
	tb.bar.RemoveChild(entry.button)
	entry.button.Dispose()
	if entry.panel != nil {
		tb.body.RemoveChild(entry.panel)
		entry.panel.Dispose()
	}

	// Remove from slice.
	copy(tb.tabs[index:], tb.tabs[index+1:])
	tb.tabs[len(tb.tabs)-1] = nil
	tb.tabs = tb.tabs[:len(tb.tabs)-1]

	// Re-wire click handlers with updated indices.
	for i, e := range tb.tabs {
		tabIdx := i
		e.button.SetOnClick(func() {
			tb.SetSelected(tabIdx)
		})
	}

	// Adjust selection.
	sel := tb.selected.Peek()
	if sel >= len(tb.tabs) && len(tb.tabs) > 0 {
		tb.SetSelected(len(tb.tabs) - 1)
	} else if len(tb.tabs) == 0 {
		tb.selected.Set(-1)
		DefaultScheduler.Flush()
	} else {
		tb.updatePanelVisibility()
		tb.updateButtonVisuals()
	}

	tb.recalcScrollBar()
}

// Selected returns the currently selected tab index.
func (tb *TabBar) Selected() int {
	return tb.selected.Peek()
}

// SetSelected sets the selected tab index.
func (tb *TabBar) SetSelected(idx int) {
	if idx < 0 || idx >= len(tb.tabs) {
		return
	}
	old := tb.selected.Peek()
	tb.selected.Set(idx)
	DefaultScheduler.Flush()
	tb.updatePanelVisibility()
	tb.updateButtonVisuals()
	if tb.overflowMode == TabOverflowScroll {
		tb.ScrollToTab(idx)
	}
	if idx != old && tb.onChange != nil {
		tb.onChange(idx)
	}
}

// BindSelected binds the tab selection to a reactive Ref[int].
func (tb *TabBar) BindSelected(ref *Ref[int]) {
	tb.selected = ref
	bindRef(&tb.watch, ref, tb.SetSelected)
}

// SetOnChange sets the callback for tab selection changes.
func (tb *TabBar) SetOnChange(fn func(int)) {
	tb.onChange = fn
}

// SetSize sets the tab bar dimensions.
func (tb *TabBar) SetSize(w, h float64) {
	tb.Width = w
	tb.Height = h
	tb.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}

	barH := 32.0
	tb.barBg.SetScale(w, barH)
	tb.bar.Width = w
	tb.bar.Height = barH

	tb.body.X = 0
	tb.body.Y = barH
	tb.body.Width = w
	tb.body.Height = h - barH

	tb.MarkLayoutDirty()
	tb.recalcScrollBar()
}

// TabCount returns the number of tabs.
func (tb *TabBar) TabCount() int {
	return len(tb.tabs)
}

// Dispose cleans up the tab bar.
func (tb *TabBar) Dispose() {
	tb.watch.Stop()
	for _, entry := range tb.tabs {
		entry.button.Dispose()
		if entry.panel != nil {
			entry.panel.Dispose()
		}
	}
	tb.tabs = nil
	if tb.leftArrow != nil {
		tb.leftArrow.Dispose()
	}
	if tb.rightArrow != nil {
		tb.rightArrow.Dispose()
	}
	tb.Component.Dispose()
}

func (tb *TabBar) updatePanelVisibility() {
	sel := tb.selected.Peek()
	for i, entry := range tb.tabs {
		if entry.panel != nil {
			entry.panel.SetVisible(i == sel)
		}
	}
}

func (tb *TabBar) updateButtonVisuals() {
	for i := range tb.tabs {
		tb.styleTabButton(i)
	}
}

// styleTabButton overrides a tab button's visuals so that the button theme
// (background, border) is replaced with the tab bar's selected/unselected
// colors and backgrounds.
func (tb *TabBar) styleTabButton(idx int) {
	if idx < 0 || idx >= len(tb.tabs) {
		return
	}
	entry := tb.tabs[idx]
	sel := tb.selected.Peek()
	group := tb.EffectiveTheme().Tabs.Group(tb.Variant())

	// Clear button-theme border.
	entry.button.applyBorder(sg.Color{}, 0, Background{})

	// Apply tab-specific background and text color.
	state := entry.button.state
	if idx == sel {
		entry.button.applyBackground(group.SelectedTabBackground.Resolve(state))
		entry.button.label.SetColor(group.SelectedTabColor.Resolve(state))
	} else {
		entry.button.applyBackground(group.UnselectedTabBackground.Resolve(state))
		entry.button.label.SetColor(group.UnselectedTabColor.Resolve(state))
	}
}

// ---------------------------------------------------------------------------
// Overflow mode
// ---------------------------------------------------------------------------

// OverflowMode returns the current overflow mode.
func (tb *TabBar) OverflowMode() TabOverflowMode {
	return tb.overflowMode
}

// ScrollOffset returns the current scroll offset in pixels.
// Always 0 when OverflowMode is TabOverflowClip.
func (tb *TabBar) ScrollOffset() float64 {
	return tb.barScrollOffset
}

// LeftArrowVisible reports whether the left scroll arrow is visible.
func (tb *TabBar) LeftArrowVisible() bool {
	return tb.leftArrow != nil && tb.leftArrow.IsVisible()
}

// RightArrowVisible reports whether the right scroll arrow is visible.
func (tb *TabBar) RightArrowVisible() bool {
	return tb.rightArrow != nil && tb.rightArrow.IsVisible()
}

// SetOverflowMode sets the overflow mode for the tab bar.
func (tb *TabBar) SetOverflowMode(mode TabOverflowMode) {
	if tb.overflowMode == mode {
		return
	}
	tb.overflowMode = mode

	if mode == TabOverflowScroll {
		tb.ensureArrows()
	} else {
		// Reset to clip mode.
		tb.barScrollOffset = 0
		tb.hideArrows()
		tb.bar.node.SetMask(nil)
		tb.bar.Width = tb.Width
		tb.bar.SetPosition(0, 0)
		tb.MarkLayoutDirty()
	}
	tb.recalcScrollBar()
}

// ScrollToTab scrolls the bar so that the tab at idx is fully visible.
// No-op when OverflowMode is TabOverflowClip or when the tab is already visible.
func (tb *TabBar) ScrollToTab(idx int) {
	if tb.overflowMode != TabOverflowScroll || idx < 0 || idx >= len(tb.tabs) {
		return
	}

	aw := tb.arrowWidth()

	// Compute the left edge of this tab in the natural (unscrolled) strip.
	left := 0.0
	for i := 0; i < idx; i++ {
		left += tb.tabs[i].button.Width + tb.bar.Spacing
	}
	right := left + tb.tabs[idx].button.Width

	// Use the worst-case visible width (both arrows present) so that the
	// target tab is fully visible regardless of which arrows end up shown.
	visibleW := tb.Width - 2*aw
	if visibleW < 0 {
		visibleW = 0
	}

	if left < tb.barScrollOffset {
		tb.barScrollOffset = left
	} else if right > tb.barScrollOffset+visibleW {
		tb.barScrollOffset = right - visibleW
	}

	tb.clampScrollOffset(aw)
	tb.recalcScrollBar()
}

// ---------------------------------------------------------------------------
// Internal scroll helpers
// ---------------------------------------------------------------------------

func (tb *TabBar) ensureArrows() {
	if tb.leftArrow != nil {
		return
	}
	name := tb.node.Name

	// Left arrow.
	tb.leftArrow = &Component{}
	initComponent(tb.leftArrow, name+"-arrow-left")
	tb.leftArrow.initBackground(name + "-arrow-left")
	tb.leftArrow.node.Interactable = true
	tb.leftArrow.SetCursorShape(engine.CursorShapePointer)
	tb.leftArrow.node.OnClick(func(_ sg.ClickContext) {
		tb.scrollByOneTab(-1)
	})
	tb.leftArrow.onVisualStateChange = func() { tb.styleArrowButtons() }

	tb.leftGlyph = sg.NewSprite(name+"-arrow-left-glyph", sg.TextureRegion{})
	tb.leftGlyph.SetCustomImage(optionRotatorLeftGlyph())
	tb.leftArrow.node.AddChild(tb.leftGlyph)
	tb.node.AddChild(tb.leftArrow.node)

	// Right arrow.
	tb.rightArrow = &Component{}
	initComponent(tb.rightArrow, name+"-arrow-right")
	tb.rightArrow.initBackground(name + "-arrow-right")
	tb.rightArrow.node.Interactable = true
	tb.rightArrow.SetCursorShape(engine.CursorShapePointer)
	tb.rightArrow.node.OnClick(func(_ sg.ClickContext) {
		tb.scrollByOneTab(+1)
	})
	tb.rightArrow.onVisualStateChange = func() { tb.styleArrowButtons() }

	tb.rightGlyph = sg.NewSprite(name+"-arrow-right-glyph", sg.TextureRegion{})
	tb.rightGlyph.SetCustomImage(treeExpandGlyph())
	tb.rightArrow.node.AddChild(tb.rightGlyph)
	tb.node.AddChild(tb.rightArrow.node)

	// Pre-allocate reusable mask nodes.
	tb.maskRoot = sg.NewContainer(name + "-bar-mask")
	tb.maskSprite = sg.NewSprite(name+"-bar-mask-rect", sg.TextureRegion{})
	tb.maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	tb.maskRoot.AddChild(tb.maskSprite)

	tb.styleArrowButtons()
}

func (tb *TabBar) hideArrows() {
	if tb.leftArrow != nil {
		tb.leftArrow.SetVisible(false)
	}
	if tb.rightArrow != nil {
		tb.rightArrow.SetVisible(false)
	}
}

func (tb *TabBar) totalTabsWidth() float64 {
	n := len(tb.tabs)
	if n == 0 {
		return 0
	}
	total := tb.bar.Spacing * float64(n-1)
	for _, e := range tb.tabs {
		total += e.button.Width
	}
	return total
}

func (tb *TabBar) arrowWidth() float64 {
	group := tb.EffectiveTheme().Tabs.Group(tb.Variant())
	aw := group.ScrollArrowWidth
	if aw <= 0 {
		aw = 24
	}
	return aw
}

// computeVisibleBarWidth returns the visible strip width given the current
// scroll offset and arrow width. Avoids redundant arrowWidth() lookups.
func (tb *TabBar) computeVisibleBarWidth(aw float64) float64 {
	w := tb.Width
	if tb.barScrollOffset > 0 {
		w -= aw
	}
	if tb.barScrollOffset < tb.computeMaxScrollOffset(aw) {
		w -= aw
	}
	return w
}

// computeMaxScrollOffset returns the maximum scroll offset. Uses both-arrows
// visible width since that is the worst case.
func (tb *TabBar) computeMaxScrollOffset(aw float64) float64 {
	visible := tb.Width - 2*aw
	if visible < 0 {
		visible = 0
	}
	max := tb.totalTabsWidth() - visible
	if max < 0 {
		max = 0
	}
	return max
}

func (tb *TabBar) clampScrollOffset(aw float64) {
	max := tb.computeMaxScrollOffset(aw)
	if tb.barScrollOffset < 0 {
		tb.barScrollOffset = 0
	}
	if tb.barScrollOffset > max {
		tb.barScrollOffset = max
	}
}

func (tb *TabBar) scrollByOneTab(dir int) {
	if len(tb.tabs) == 0 {
		return
	}

	aw := tb.arrowWidth()

	if dir > 0 {
		// Scroll right: find first tab whose right edge is beyond the visible area.
		visibleRight := tb.barScrollOffset + tb.computeVisibleBarWidth(aw)
		left := 0.0
		for _, e := range tb.tabs {
			right := left + e.button.Width
			if right > visibleRight+0.5 {
				tb.barScrollOffset += e.button.Width + tb.bar.Spacing
				break
			}
			left = right + tb.bar.Spacing
		}
	} else {
		// Scroll left: find the last tab whose left edge is before the visible
		// area, tracking its position in a single pass.
		left := 0.0
		targetOffset := 0.0
		found := false
		for _, e := range tb.tabs {
			if left < tb.barScrollOffset-0.5 {
				targetOffset = left
				found = true
			}
			left += e.button.Width + tb.bar.Spacing
		}
		if found {
			tb.barScrollOffset = targetOffset
		}
	}

	tb.clampScrollOffset(aw)
	tb.recalcScrollBar()
}

func (tb *TabBar) recalcScrollBar() {
	if tb.overflowMode != TabOverflowScroll || tb.leftArrow == nil {
		return
	}

	aw := tb.arrowWidth()
	tb.clampScrollOffset(aw)

	barH := 32.0
	total := tb.totalTabsWidth()
	maxOff := tb.computeMaxScrollOffset(aw)

	showLeft := tb.barScrollOffset > 0.5
	showRight := tb.barScrollOffset < maxOff-0.5

	// If total tabs fit without arrows, hide both and use full width.
	if total <= tb.Width {
		showLeft = false
		showRight = false
		tb.barScrollOffset = 0
	}

	tb.leftArrow.SetVisible(showLeft)
	tb.rightArrow.SetVisible(showRight)

	// Position and size arrows.
	tb.leftArrow.Width = aw
	tb.leftArrow.Height = barH
	tb.leftArrow.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: aw, Height: barH}
	tb.leftArrow.SetPosition(0, 0)

	tb.rightArrow.Width = aw
	tb.rightArrow.Height = barH
	tb.rightArrow.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: aw, Height: barH}
	tb.rightArrow.SetPosition(tb.Width-aw, 0)

	// Center glyphs within their arrow areas.
	centerGlyphIn(tb.leftGlyph, optionRotatorLeftGlyph(), aw, barH)
	centerGlyphIn(tb.rightGlyph, treeExpandGlyph(), aw, barH)

	// Compute visible bar region.
	barX := 0.0
	visibleW := tb.Width
	if showLeft {
		barX = aw
		visibleW -= aw
	}
	if showRight {
		visibleW -= aw
	}

	// Position bar with scroll offset applied.
	tb.bar.SetPosition(barX-tb.barScrollOffset, 0)
	tb.bar.Width = total

	// Update reusable clipping mask.
	tb.maskSprite.SetPosition(barX, 0)
	tb.maskSprite.SetScale(visibleW, barH)
	tb.bar.node.SetMask(tb.maskRoot)

	tb.styleArrowButtons()
	tb.MarkLayoutDirty()
}

// centerGlyphIn positions a glyph sprite centered within an area, scaling
// the glyph to a 9px display size to match the default chevron visual.
func centerGlyphIn(glyph *sg.Node, img engine.Image, areaW, areaH float64) {
	if glyph == nil || img == nil {
		return
	}
	const displaySize = 9.0
	glyph.SetSize(displaySize, displaySize)
	glyph.SetPosition((areaW-displaySize)/2, (areaH-displaySize)/2)
}

func (tb *TabBar) styleArrowButtons() {
	if tb.leftArrow == nil || tb.rightArrow == nil {
		return
	}
	group := tb.EffectiveTheme().Tabs.Group(tb.Variant())

	leftSt := tb.leftArrow.state
	rightSt := tb.rightArrow.state

	tb.leftArrow.applyBackground(group.ScrollArrowBackground.Resolve(leftSt))
	tb.rightArrow.applyBackground(group.ScrollArrowBackground.Resolve(rightSt))

	if tb.leftGlyph != nil {
		tb.leftGlyph.SetColor(group.ScrollArrowColor.Resolve(leftSt))
	}
	if tb.rightGlyph != nil {
		tb.rightGlyph.SetColor(group.ScrollArrowColor.Resolve(rightSt))
	}
}
