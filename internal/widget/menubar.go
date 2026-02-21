package widget

import (
	"fmt"

	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// menuBarOverlayZIndex is the same z as the dismiss node (menuPopupOverlayZIndex - 1).
// Because the bar overlay is added to scene.Root after the dismiss node, it
// wins hit-testing in the bar region. The popup overlay at menuPopupOverlayZIndex
// renders on top of both, so the dropdown is never hidden behind the bar.
const menuBarOverlayZIndex = menuPopupOverlayZIndex - 1

// MenuBarEntry defines one top-level menu in the bar.
type MenuBarEntry struct {
	Label string     // displayed in the bar (e.g. "File")
	Items []MenuItem // items shown when this entry is opened
}

// MenuBar is a horizontal bar of labeled menu buttons that open dropdown
// MenuPopup panels. Provides the standard desktop-style application menu.
type MenuBar struct {
	Component
	entries     []MenuBarEntry
	source      *sg.FontFamily
	font        *sg.FontFamily
	displaySize float64

	// Visual nodes.
	barBg      *sg.Node       // background sprite
	borderNode *sg.Node       // bottom border sprite
	entryNodes []*menuBarNode // per-entry visual data

	// High-z overlay: attached to the scene root when a menu is open so that
	// bar entries remain clickable above the MenuPopupManager's dismiss node.
	overlay *sg.Node // container at menuBarOverlayZIndex; children mirror entries

	// State.
	menuOpen      bool // true while a dropdown is visible
	activeIndex   int  // currently open/focused entry (-1 = none)
	hoveredIndex  int  // entry under pointer (-1 = none)
	openGuardTick int  // frame tick when menu was last opened; ignore events for 2 ticks

	// Callbacks.
	onMenuOpen  func(int)
	onMenuClose func()

	// Popup reused for showing items.
	popup     *MenuPopup
	tick      int  // incremented each frame via OnUpdate
	switching bool // true while openMenu is swapping popups (suppresses onDismiss)
}

type menuBarNode struct {
	bg    *sg.Node // entry background sprite (active — swapped to overlay when open)
	label *sg.Node // entry text node (active — swapped to overlay when open)
	hit   *sg.Node // hit container (in normal bar, not swapped)
	x     float64  // entry x position in bar
	w     float64  // entry width

	// Saved normal-bar references, set before the overlay swaps them.
	normalBg    *sg.Node
	normalLabel *sg.Node
}

// NewMenuBar creates a new MenuBar with the given font source and display size.
func NewMenuBar(name string, source *sg.FontFamily, displaySize float64) *MenuBar {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	mb := &MenuBar{
		source:        source,
		font:          font,
		displaySize:   displaySize,
		activeIndex:   -1,
		hoveredIndex:  -1,
		openGuardTick: -10,
	}
	initComponent(&mb.Component, name)

	// Background.
	mb.barBg = sg.NewSprite(name+"-bg", sg.TextureRegion{})
	mb.node.AddChild(mb.barBg)

	// Bottom border.
	mb.borderNode = sg.NewSprite(name+"-border", sg.TextureRegion{})
	mb.borderNode.SetZIndex(1)
	mb.node.AddChild(mb.borderNode)

	mb.onThemeChange = func() { mb.applyTheme() }

	// Keyboard navigation and tick counter via OnUpdate.
	mb.node.OnUpdate = func(_ float64) {
		mb.tick++
		mb.handleKeys()
	}

	mb.applyTheme()
	mb.SetSize(400, 28)

	return mb
}

// SetEntries sets the menu entries and rebuilds the bar.
func (mb *MenuBar) SetEntries(entries []MenuBarEntry) {
	mb.entries = entries
	mb.rebuild()
}

// SetEntry replaces a single entry at the given index.
func (mb *MenuBar) SetEntry(index int, entry MenuBarEntry) {
	if index < 0 || index >= len(mb.entries) {
		return
	}
	mb.entries[index] = entry
	mb.rebuild()
}

// SetSize sets the menu bar dimensions.
func (mb *MenuBar) SetSize(w, h float64) {
	mb.Width = w
	mb.Height = h
	mb.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	mb.applyTheme()
	mb.rebuild()
}

// SetOnMenuOpen sets a callback fired when a menu entry is opened.
func (mb *MenuBar) SetOnMenuOpen(fn func(int)) {
	mb.onMenuOpen = fn
}

// SetOnMenuClose sets a callback fired when the active menu is closed.
func (mb *MenuBar) SetOnMenuClose(fn func()) {
	mb.onMenuClose = fn
}

func (mb *MenuBar) group() *MenuBarGroup {
	return mb.EffectiveTheme().MenuBar.Group(mb.Variant())
}

func (mb *MenuBar) applyTheme() {
	g := mb.group()

	// Bar background.
	bgColor := g.Background.Resolve(core.StateDefault).Color
	mb.barBg.SetScale(mb.Width, mb.Height)
	mb.barBg.SetColor(bgColor)

	// Bottom border.
	bw := g.BorderWidth
	if bw > 0 {
		bc := g.BorderColor.Resolve(core.StateDefault)
		mb.borderNode.SetScale(mb.Width, bw)
		mb.borderNode.SetPosition(0, mb.Height-bw)
		mb.borderNode.SetColor(bc)
		mb.borderNode.SetVisible(true)
	} else {
		mb.borderNode.SetVisible(false)
	}

	// Update entry colors.
	mb.updateEntryVisuals()
}

func (mb *MenuBar) rebuild() {
	// Remove old entry nodes.
	for _, en := range mb.entryNodes {
		mb.node.RemoveChild(en.bg)
		mb.node.RemoveChild(en.label)
		mb.node.RemoveChild(en.hit)
	}
	mb.entryNodes = nil

	if len(mb.entries) == 0 {
		return
	}

	g := mb.group()
	pad := g.EntryPadding
	spacing := g.Spacing
	h := mb.Height

	// Vertically center text: measure actual text height and compute offset.
	_, textH := measureDisplay(mb.font, "Mg", mb.displaySize) // representative height
	textY := (h - textH) / 2
	if textY < 0 {
		textY = 0
	}

	x := 0.0
	for i, entry := range mb.entries {
		idx := i

		// Measure label width.
		lw, _ := measureDisplay(mb.font, entry.Label, mb.displaySize)
		entryW := lw + pad.Left + pad.Right

		// Background sprite.
		bg := sg.NewSprite(fmt.Sprintf("%s-entry-bg-%d", mb.node.Name, i), sg.TextureRegion{})
		bg.SetScale(entryW, h)
		bg.SetPosition(x, 0)
		bg.SetZIndex(2)
		bg.SetVisible(false)
		mb.node.AddChild(bg)

		// Label.
		lbl := sg.NewText(fmt.Sprintf("%s-entry-lbl-%d", mb.node.Name, i), entry.Label, mb.font)
		lbl.TextBlock.FontSize = mb.displaySize
		lbl.TextBlock.Color = g.EntryTextColor.Resolve(core.StateDefault)
		lbl.SetPosition(x+pad.Left, textY)
		lbl.SetZIndex(3)
		mb.node.AddChild(lbl)

		// Hit container for normal (non-overlay) interaction.
		hit := sg.NewContainer(fmt.Sprintf("%s-entry-hit-%d", mb.node.Name, i))
		hit.Interactable = true
		hit.HitShape = sg.HitRect{X: x, Y: 0, Width: entryW, Height: h}
		hit.SetZIndex(4)
		mb.wireEntryEvents(hit, idx)
		mb.node.AddChild(hit)

		mb.entryNodes = append(mb.entryNodes, &menuBarNode{
			bg:    bg,
			label: lbl,
			hit:   hit,
			x:     x,
			w:     entryW,
		})

		x += entryW + spacing
	}

	mb.updateEntryVisuals()
}

// wireEntryEvents attaches pointer callbacks for entry idx to the given node.
func (mb *MenuBar) wireEntryEvents(node *sg.Node, idx int) {
	node.OnPointerEnter(func(_ sg.PointerContext) {
		mb.hoveredIndex = idx
		if mb.menuOpen && idx != mb.activeIndex {
			mb.openMenu(idx)
		}
		mb.updateEntryVisuals()
	})
	node.OnPointerLeave(func(_ sg.PointerContext) {
		if mb.hoveredIndex == idx {
			mb.hoveredIndex = -1
		}
		mb.updateEntryVisuals()
	})
	node.OnPointerDown(func(_ sg.PointerContext) {
		// Guard: ignore events for 2 ticks after a menu open to avoid the
		// overlay's hit container immediately closing the menu it just opened.
		if mb.tick-mb.openGuardTick < 2 {
			return
		}
		if mb.menuOpen && mb.activeIndex == idx {
			mb.closeMenu()
		} else {
			mb.openMenu(idx)
		}
	})
}

func (mb *MenuBar) updateEntryVisuals() {
	g := mb.group()
	for i, en := range mb.entryNodes {
		var state core.ComponentState
		if i == mb.activeIndex {
			state = core.StateActive
		} else if i == mb.hoveredIndex {
			state = core.StateHover
		} else {
			state = core.StateDefault
		}

		// Entry background.
		bgColor := g.EntryBackground.Resolve(state).Color
		if bgColor.A() > 0 {
			en.bg.SetColor(bgColor)
			en.bg.SetVisible(true)
		} else {
			en.bg.SetVisible(false)
		}

		// Entry text color.
		textColor := g.EntryTextColor.Resolve(state)
		en.label.TextBlock.Color = textColor
		en.label.Invalidate()
	}
}

// showOverlay creates a transparent hit overlay at the scene root above the
// MenuPopupManager's dismiss node. This lets users hover/click bar entries
// while a dropdown is open.
func (mb *MenuBar) showOverlay() {
	sc := currentScene()
	if sc == nil || sc.Root == nil {
		return
	}

	if mb.overlay != nil {
		mb.hideOverlay()
	}

	barX, barY := mb.node.LocalToWorld(0, 0)

	mb.overlay = sg.NewContainer("menubar-overlay")
	mb.overlay.Interactable = true
	mb.overlay.SetPosition(barX, barY)
	mb.overlay.SetZIndex(menuBarOverlayZIndex)

	g := mb.group()
	bgColor := g.Background.Resolve(core.StateDefault).Color

	// Background so the bar visually layers above the dismiss node.
	bg := sg.NewSprite("menubar-overlay-bg", sg.TextureRegion{})
	bg.SetScale(mb.Width, mb.Height)
	bg.SetColor(bgColor)
	mb.overlay.AddChild(bg)

	// Hit shape covering the full bar so events don't leak to dismiss node.
	mb.overlay.HitShape = sg.HitRect{X: 0, Y: 0, Width: mb.Width, Height: mb.Height}

	// Vertically center text.
	_, textH := measureDisplay(mb.font, "Mg", mb.displaySize)
	textY := (mb.Height - textH) / 2
	if textY < 0 {
		textY = 0
	}

	pad := g.EntryPadding
	spacing := g.Spacing

	x := 0.0
	for i, entry := range mb.entries {
		idx := i
		en := mb.entryNodes[i]

		lw, _ := measureDisplay(mb.font, entry.Label, mb.displaySize)
		entryW := lw + pad.Left + pad.Right

		// Entry bg in overlay.
		oBg := sg.NewSprite(fmt.Sprintf("menubar-ov-bg-%d", i), sg.TextureRegion{})
		oBg.SetScale(entryW, mb.Height)
		oBg.SetPosition(x, 0)
		oBg.SetZIndex(2)

		var state core.ComponentState
		if i == mb.activeIndex {
			state = core.StateActive
		} else {
			state = core.StateDefault
		}
		entryBg := g.EntryBackground.Resolve(state).Color
		if entryBg.A() > 0 {
			oBg.SetColor(entryBg)
		} else {
			oBg.SetVisible(false)
		}
		mb.overlay.AddChild(oBg)

		// Save normal refs and swap to overlay.
		en.normalBg = en.bg
		en.normalLabel = en.label
		en.bg = oBg

		// Label in overlay.
		lbl := sg.NewText(fmt.Sprintf("menubar-ov-lbl-%d", i), entry.Label, mb.font)
		lbl.TextBlock.FontSize = mb.displaySize
		lbl.TextBlock.Color = g.EntryTextColor.Resolve(state)
		lbl.SetPosition(x+pad.Left, textY)
		lbl.SetZIndex(3)
		mb.overlay.AddChild(lbl)
		en.label = lbl

		// Hit container in overlay.
		hit := sg.NewContainer(fmt.Sprintf("menubar-ov-hit-%d", i))
		hit.Interactable = true
		hit.HitShape = sg.HitRect{X: x, Y: 0, Width: entryW, Height: mb.Height}
		hit.SetZIndex(4)
		mb.wireEntryEvents(hit, idx)
		mb.overlay.AddChild(hit)

		x += entryW + spacing
	}

	sc.Root.AddChild(mb.overlay)
}

// hideOverlay removes the high-z overlay from the scene and restores the
// normal entry node references.
func (mb *MenuBar) hideOverlay() {
	if mb.overlay == nil {
		return
	}
	if mb.overlay.Parent != nil {
		mb.overlay.Parent.RemoveChild(mb.overlay)
	}
	mb.overlay = nil

	// Restore entry node references to the saved normal-bar nodes.
	for _, en := range mb.entryNodes {
		if en.normalBg != nil {
			en.bg = en.normalBg
			en.normalBg = nil
		}
		if en.normalLabel != nil {
			en.label = en.normalLabel
			en.normalLabel = nil
		}
	}

	// Re-apply current visual state to the normal bar nodes.
	mb.updateEntryVisuals()
}

func (mb *MenuBar) openMenu(index int) {
	if index < 0 || index >= len(mb.entries) {
		return
	}

	entry := mb.entries[index]
	en := mb.entryNodes[index]

	if mb.popup == nil {
		mb.popup = NewMenuPopup(mb.source, mb.displaySize)
	}
	mb.popup.SetItems(entry.Items)
	mb.popup.minWidth = en.w
	mb.popup.SetOnDismiss(func() {
		if mb.switching {
			return // suppress during menu-to-menu switch
		}
		mb.menuOpen = false
		mb.activeIndex = -1
		mb.hideOverlay()
		mb.updateEntryVisuals()
		if mb.onMenuClose != nil {
			mb.onMenuClose()
		}
	})

	// Position below the entry with a small gap to clear the bottom border.
	// ShowAt internally calls hideActive which fires onDismiss — the
	// switching flag suppresses that.
	wx, wy := mb.node.LocalToWorld(en.x, mb.Height+1)
	mb.switching = true
	DefaultMenuPopupManager.ShowAt(mb.popup, wx, wy)
	mb.switching = false

	mb.menuOpen = true
	mb.activeIndex = index
	mb.openGuardTick = mb.tick

	// Show the high-z overlay so bar entries remain interactive above the
	// dismiss node.
	mb.showOverlay()
	mb.updateEntryVisuals()

	if mb.onMenuOpen != nil {
		mb.onMenuOpen(index)
	}
}

func (mb *MenuBar) closeMenu() {
	if !mb.menuOpen {
		return
	}
	DefaultMenuPopupManager.Hide()
	// onDismiss callback handles state cleanup + hideOverlay.
}

func (mb *MenuBar) handleKeys() {
	if len(mb.entries) == 0 {
		return
	}

	// Only handle keys when bar is focused or a menu is open.
	if !mb.focused && !mb.menuOpen {
		return
	}

	im := DefaultInputManager

	if im.IsKeyJustAvailable(engine.KeyArrowLeft) {
		im.Consume(engine.KeyArrowLeft)
		mb.cycleEntry(-1)
	} else if im.IsKeyJustAvailable(engine.KeyArrowRight) {
		im.Consume(engine.KeyArrowRight)
		mb.cycleEntry(1)
	} else if im.IsKeyJustAvailable(engine.KeyArrowDown) || im.IsKeyJustAvailable(engine.KeyEnter) {
		im.Consume(engine.KeyArrowDown)
		im.Consume(engine.KeyEnter)
		if !mb.menuOpen {
			idx := mb.activeIndex
			if idx < 0 {
				idx = 0
			}
			mb.openMenu(idx)
		}
	} else if im.IsKeyJustAvailable(engine.KeyEscape) {
		im.Consume(engine.KeyEscape)
		if mb.menuOpen {
			mb.closeMenu()
		}
	}
}

func (mb *MenuBar) cycleEntry(dir int) {
	n := len(mb.entries)
	if n == 0 {
		return
	}
	cur := mb.activeIndex
	if cur < 0 {
		cur = 0
	} else {
		cur = (cur + dir + n) % n
	}

	if mb.menuOpen {
		mb.openMenu(cur)
	} else {
		mb.activeIndex = cur
		mb.updateEntryVisuals()
	}
}
