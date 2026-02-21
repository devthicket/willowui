package widget

import (
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// toggleButtonEntry pairs a sub-component (button) with its label.
type toggleButtonEntry struct {
	comp  Component
	label *Label
}

// ToggleButtonBar is a segmented control for single-selection among a set
// of labeled buttons. Only one button is selected at a time.
type ToggleButtonBar struct {
	Component
	entries     []*toggleButtonEntry
	selected    *Ref[int]
	watch       WatchHandle
	stopButtons func() // stops reactive buttons Array binding
	onChange    func(int)
	source      *sg.FontFamily
	font        *sg.FontFamily
	displaySize float64
}

// NewToggleButtonBar creates a new toggle button bar with the given font
// and display size. If displaySize is 0, the native atlas size is used.
func NewToggleButtonBar(name string, source *sg.FontFamily, displaySize float64) *ToggleButtonBar {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	t := &ToggleButtonBar{
		selected:    NewRef(0),
		source:      source,
		font:        font,
		displaySize: displaySize,
	}
	initComponent(&t.Component, name)

	t.initBackground(name)
	t.initBorder(name)

	t.enableFocusNavigation()

	// Left/Right cycle through tabs; Up/Down navigate like Shift+Tab / Tab.
	DefaultFocusManager.BindScoped(&t.Component, Key(engine.KeyLeft, ModNone), func() bool {
		n := len(t.entries)
		if n > 0 {
			t.SetSelected((t.selected.Peek() - 1 + n) % n)
		}
		return true
	})
	DefaultFocusManager.BindScoped(&t.Component, Key(engine.KeyRight, ModNone), func() bool {
		n := len(t.entries)
		if n > 0 {
			t.SetSelected((t.selected.Peek() + 1) % n)
		}
		return true
	})
	DefaultFocusManager.BindScoped(&t.Component, Key(engine.KeyUp, ModNone), func() bool {
		DefaultFocusManager.TabPrev()
		return true
	})
	DefaultFocusManager.BindScoped(&t.Component, Key(engine.KeyDown, ModNone), func() bool {
		DefaultFocusManager.TabNext()
		return true
	})

	t.onThemeChange = func() { t.updateVisuals() }
	t.onFocusChange = func(bool) { t.updateVisuals() }
	t.updateVisuals()

	// Default size.
	t.SetSize(400, 40)

	return t
}

// AddButton adds a labeled button to the bar and updates layout and visuals.
func (t *ToggleButtonBar) AddButton(label string) {
	idx := len(t.entries)
	name := t.node.Name + "-entry"

	entry := &toggleButtonEntry{}
	initComponent(&entry.comp, name)
	entry.comp.initBackground(name)
	entry.comp.initBorder(name)
	entry.comp.node.Interactable = true
	entry.comp.SetCursorShape(engine.CursorShapePointer)

	entry.label = NewLabel(name+"-label", label, t.source, t.displaySize)
	entry.comp.node.AddChild(entry.label.Node())

	// Wire click handler.
	tabIdx := idx
	entry.comp.node.OnClick(func(ctx sg.ClickContext) {
		t.SetSelected(tabIdx)
	})

	t.node.AddChild(entry.comp.node)
	t.entries = append(t.entries, entry)

	t.updateLayout()
	t.updateVisuals()
}

// RemoveButton removes the button at the given index.
func (t *ToggleButtonBar) RemoveButton(idx int) {
	if idx < 0 || idx >= len(t.entries) {
		return
	}

	entry := t.entries[idx]
	t.node.RemoveChild(entry.comp.node)
	entry.label.Dispose()
	entry.comp.Dispose()

	// Remove from slice.
	copy(t.entries[idx:], t.entries[idx+1:])
	t.entries[len(t.entries)-1] = nil
	t.entries = t.entries[:len(t.entries)-1]

	// Re-wire click handlers with updated indices.
	for i, e := range t.entries {
		tabIdx := i
		e.comp.node.OnClick(func(ctx sg.ClickContext) {
			t.SetSelected(tabIdx)
		})
	}

	// Adjust selection.
	sel := t.selected.Peek()
	if sel >= len(t.entries) && len(t.entries) > 0 {
		t.SetSelected(len(t.entries) - 1)
	} else if len(t.entries) == 0 {
		t.selected.Set(-1)
		DefaultScheduler.Flush()
	} else {
		t.updateLayout()
		t.updateVisuals()
	}
}

// Selected returns the currently selected button index.
func (t *ToggleButtonBar) Selected() int {
	return t.selected.Peek()
}

// SetSelected sets the selected button index.
func (t *ToggleButtonBar) SetSelected(idx int) {
	if idx < 0 || idx >= len(t.entries) {
		return
	}
	old := t.selected.Peek()
	t.selected.Set(idx)
	DefaultScheduler.Flush()
	t.updateVisuals()
	if idx != old && t.onChange != nil {
		t.onChange(idx)
	}
}

// BindButtons binds the button labels to a reactive Array[string]. When the
// array changes the bar is rebuilt from the new labels, preserving the
// selection index where possible.
// Pass nil to detach.
func (t *ToggleButtonBar) BindButtons(arr *Array[string]) {
	if t.stopButtons != nil {
		t.stopButtons()
		t.stopButtons = nil
	}
	if arr == nil {
		return
	}
	rebuild := func() {
		sel := t.selected.Peek()
		// Remove all existing buttons.
		for len(t.entries) > 0 {
			t.RemoveButton(0)
		}
		// Add new buttons.
		arr.ForEach(func(_ int, label string) {
			t.AddButton(label)
		})
		// Restore selection, clamped to new length.
		n := len(t.entries)
		if n > 0 {
			if sel >= n {
				sel = n - 1
			}
			t.SetSelected(sel)
		}
	}
	rebuild()
	h := arr.OnChange(func() { rebuild() })
	t.stopButtons = func() { h.Stop() }
}

// BindSelected binds the selection to a reactive Ref[int].
func (t *ToggleButtonBar) BindSelected(ref *Ref[int]) {
	t.selected = ref
	bindRef(&t.watch, ref, t.SetSelected)
}

// SetOnChange sets the callback for selection changes.
func (t *ToggleButtonBar) SetOnChange(fn func(int)) {
	t.onChange = fn
}

// SetSize sets the toggle button bar dimensions and updates layout.
func (t *ToggleButtonBar) SetSize(w, h float64) {
	t.Width = w
	t.Height = h
	t.resizeBackground(w, h)
	t.resizeBorder(w, h)
	t.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	t.updateLayout()
	t.MarkLayoutDirty()
}

// ButtonCount returns the number of buttons.
func (t *ToggleButtonBar) ButtonCount() int {
	return len(t.entries)
}

// Dispose cleans up the toggle button bar and all its entries.
func (t *ToggleButtonBar) Dispose() {
	if t.stopButtons != nil {
		t.stopButtons()
	}
	t.watch.Stop()
	for _, entry := range t.entries {
		entry.label.Dispose()
		entry.comp.Dispose()
	}
	t.entries = nil
	t.Component.Dispose()
}

// EntriesIsNil reports whether the entries slice is nil (after dispose). Used for testing.
func (t *ToggleButtonBar) EntriesIsNil() bool { return t.entries == nil }

// updateLayout positions buttons evenly within the bar, accounting for
// padding and spacing from the theme.
func (t *ToggleButtonBar) updateLayout() {
	n := len(t.entries)
	if n == 0 {
		return
	}

	group := t.EffectiveTheme().ToggleButtonBar.Group(t.Variant())
	pad := resolveAutoInsets(group.Padding, defaultBarPadding)
	spacing := group.Spacing

	availW := t.Width - pad.Left - pad.Right - spacing*float64(n-1)
	btnW := availW / float64(n)
	btnH := t.Height - pad.Top - pad.Bottom

	x := pad.Left
	for _, entry := range t.entries {
		entry.comp.Width = btnW
		entry.comp.Height = btnH
		entry.comp.resizeBackground(btnW, btnH)
		entry.comp.resizeBorder(btnW, btnH)
		entry.comp.node.SetPosition(x, pad.Top)
		entry.comp.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: btnW, Height: btnH}

		// Center label within button.
		if entry.label != nil {
			lx := (btnW - entry.label.Width) / 2
			ly := (btnH - entry.label.Height) / 2
			entry.label.SetPosition(lx, ly)
		}

		x += btnW + spacing
	}
}

// updateVisuals applies theme colors to the bar and each button based on
// the current selection state.
func (t *ToggleButtonBar) updateVisuals() {
	t.state = computeState(t.enabled, t.focused, t.hovered, false)
	group := t.EffectiveTheme().ToggleButtonBar.Group(t.Variant())

	// Bar background and border.
	t.applyCornerRadius(group.CornerRadius)
	bg := group.Background.Resolve(StateDefault)
	t.applyBackground(bg)
	t.applyBorder(group.Border.Resolve(t.state), group.BorderWidth, bg)

	sel := t.selected.Peek()
	for i, entry := range t.entries {
		if i == sel {
			entry.comp.applyCornerRadius(group.SelectedCornerRadius)
			selBg := group.SelectedBackground.Resolve(StateDefault)
			entry.comp.applyBackground(selBg)
			entry.comp.applyBorder(group.SelectedBorder.Resolve(StateDefault), group.SelectedBorderWidth, selBg)
			if entry.label != nil {
				entry.label.SetColor(group.SelectedTextColor.Resolve(StateDefault))
			}
		} else {
			entry.comp.applyCornerRadius(group.UnselectedCornerRadius)
			unselBg := group.UnselectedBackground.Resolve(StateDefault)
			entry.comp.applyBackground(unselBg)
			entry.comp.applyBorder(group.UnselectedBorder.Resolve(StateDefault), group.UnselectedBorderWidth, unselBg)
			if entry.label != nil {
				entry.label.SetColor(group.UnselectedTextColor.Resolve(StateDefault))
			}
		}
	}

	t.applyFocusRing(group.FocusColor.Resolve(t.state), group.FocusRingWidth)
	t.MarkDrawDirty()
}
