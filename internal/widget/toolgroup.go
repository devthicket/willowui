package widget

// ToolGroup manages mutually-exclusive selection among a set of IconButtons.
// When one button is activated, all others in the group are deactivated —
// similar to a radio group but for toolbar icon buttons.
type ToolGroup struct {
	buttons   []*IconButton
	selected  *Ref[int]
	watch     WatchHandle
	allowNone bool // if true, clicking the active button deselects it
	onChange  func(int)
}

// NewToolGroup creates a new ToolGroup. By default one button must always be
// selected; call SetAllowNone(true) to permit deselecting all.
func NewToolGroup() *ToolGroup {
	return &ToolGroup{
		selected: NewRef(-1),
	}
}

// SetAllowNone controls whether clicking the active button deselects it
// (selected becomes -1). Default is false.
func (tg *ToolGroup) SetAllowNone(v bool) {
	tg.allowNone = v
}

// Add appends an IconButton to the group and wires its click handler to
// select it (deselecting any previously active button in the group).
func (tg *ToolGroup) Add(btn *IconButton) {
	idx := len(tg.buttons)
	tg.buttons = append(tg.buttons, btn)

	btn.SetOnClick(func() {
		cur := tg.selected.Peek()
		if cur == idx && tg.allowNone {
			tg.SetSelected(-1)
		} else {
			tg.SetSelected(idx)
		}
	})
}

// Selected returns the index of the currently active button, or -1 if none.
func (tg *ToolGroup) Selected() int {
	return tg.selected.Peek()
}

// SetSelected activates the button at idx (pass -1 to deselect all).
func (tg *ToolGroup) SetSelected(idx int) {
	old := tg.selected.Peek()
	if idx == old {
		return
	}
	tg.selected.Set(idx)
	DefaultScheduler.Flush()

	for i, btn := range tg.buttons {
		btn.SetActive(i == idx)
	}

	if tg.onChange != nil {
		tg.onChange(idx)
	}
}

// BindSelected binds the group's selection to an external Ref[int].
func (tg *ToolGroup) BindSelected(ref *Ref[int]) {
	tg.selected = ref
	bindRef(&tg.watch, ref, tg.SetSelected)
}

// SetOnChange sets the callback invoked when the selection changes.
// The callback receives the new selected index (-1 if deselected).
func (tg *ToolGroup) SetOnChange(fn func(int)) {
	tg.onChange = fn
}

// ButtonCount returns the number of buttons in the group.
func (tg *ToolGroup) ButtonCount() int {
	return len(tg.buttons)
}

// Dispose stops any reactive watches.
func (tg *ToolGroup) Dispose() {
	tg.watch.Stop()
	tg.buttons = nil
}
