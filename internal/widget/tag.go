package widget

import (
	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
	"github.com/devthicket/willowui/internal/theme"
)

// Tag is a compact pill widget used as a category marker, filter chip, or
// item classifier. It supports optional remove (×) and selectable (toggle)
// modes that may be combined.
type Tag struct {
	Component
	textNode    *sg.Node
	removeNode  *sg.Node // × affordance (visible when removable)
	font        *sg.FontFamily
	displaySize float64
	text        string
	removable   bool
	selectable  bool
	selected    bool
	padOverride bool

	onRemove func()
	onToggle func(selected bool)
}

// NewTag creates a Tag with the given name, font source, and display size.
func NewTag(name string, source *sg.FontFamily, displaySize float64) *Tag {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	t := &Tag{
		font:        font,
		displaySize: displaySize,
	}
	initComponent(&t.Component, name)
	t.initBackground(name)

	t.textNode = sg.NewText(name+"-text", "", font)
	t.textNode.TextBlock.FontSize = displaySize
	t.node.AddChild(t.textNode)

	// × remove button — hidden by default.
	t.removeNode = sg.NewText(name+"-remove", "x", font)
	t.removeNode.TextBlock.FontSize = displaySize
	t.removeNode.SetTextColor(sg.RGBA(0.7, 0.7, 0.75, 1))
	t.removeNode.SetVisible(false)
	t.removeNode.Interactable = true
	t.removeNode.OnClick(func(ctx sg.ClickContext) {
		if t.onRemove != nil {
			t.onRemove()
		}
	})
	t.node.AddChild(t.removeNode)

	// Click on the tag body toggles selection (when selectable).
	t.node.OnClick(func(ctx sg.ClickContext) {
		if !t.enabled {
			return
		}
		if t.selectable {
			t.selected = !t.selected
			t.UpdateVisuals()
			if t.onToggle != nil {
				t.onToggle(t.selected)
			}
		}
	})

	t.onThemeChange = func() { t.UpdateVisuals() }

	return t
}

// Text returns the current text content.
func (t *Tag) Text() string { return t.text }

// SetText updates the displayed text.
func (t *Tag) SetText(text string) {
	t.text = text
	t.textNode.SetContent(text)
	t.MarkDrawDirty()
}

// SetRemovable shows or hides the × affordance.
func (t *Tag) SetRemovable(v bool) {
	t.removable = v
	t.removeNode.SetVisible(v)
}

// SetSelectable enables or disables toggle behaviour on click.
func (t *Tag) SetSelectable(v bool) {
	t.selectable = v
	t.UpdateVisuals()
}

// SetSelected sets the toggle state (only meaningful when selectable).
func (t *Tag) SetSelected(v bool) {
	t.selected = v
	t.UpdateVisuals()
}

// Selected returns the current toggle state.
func (t *Tag) Selected() bool { return t.selected }

// SetOnRemove sets the callback invoked when the × button is clicked.
func (t *Tag) SetOnRemove(fn func()) { t.onRemove = fn }

// SetOnToggle sets the callback invoked when the tag is toggled.
func (t *Tag) SetOnToggle(fn func(selected bool)) { t.onToggle = fn }

// SetPadding overrides the theme padding with per-instance values.
func (t *Tag) SetPadding(top, right, bottom, left float64) {
	t.Padding = render.Insets{Top: top, Right: right, Bottom: bottom, Left: left}
	t.padOverride = true
}

// SetSize overrides the auto-size dimensions.
func (t *Tag) SetSize(w, h float64) {
	t.Width = w
	t.Height = h
	t.resizeBackground(w, h)
	t.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	t.positionContent()
	t.UpdateVisuals()
	t.MarkLayoutDirty()
}

// SizeToContent auto-sizes the tag to fit its text plus padding.
func (t *Tag) SizeToContent() {
	group := t.effectiveGroup()
	pad := t.resolvedPadding(group)
	tw, th := measureDisplay(t.font, t.text, t.displaySize)

	w := tw + pad.Horizontal()
	h := th + pad.Vertical()

	if t.removable {
		rw, _ := measureDisplay(t.font, "x", t.displaySize)
		w += group.Gap + rw
	}

	// Ensure minimum pill width.
	if w < h {
		w = h
	}

	t.Width = w
	t.Height = h
	t.resizeBackground(w, h)
	t.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	t.positionContent()
	t.UpdateVisuals()
	t.MarkLayoutDirty()
}

// UpdateVisuals applies theme colors and corner radius.
func (t *Tag) UpdateVisuals() {
	group := t.effectiveGroup()

	// Corner radius: -1 means pill.
	cr := resolveCornerRadius(group.CornerRadius, t.Height)
	t.applyCornerRadius(cr)

	// Background: selected vs normal.
	if t.selected {
		t.applyBackground(group.SelectedBackground[core.StateDefault])
		t.textNode.SetTextColor(group.SelectedTextColor.Resolve(core.StateDefault))
	} else {
		t.applyBackground(group.Background[core.StateDefault])
		t.textNode.SetTextColor(group.TextColor.Resolve(core.StateDefault))
	}

	// Remove button color.
	if t.removable {
		t.removeNode.SetTextColor(group.RemoveButtonColor.Resolve(core.StateDefault))
	}

	t.MarkDrawDirty()
}

func (t *Tag) effectiveGroup() *theme.TagGroup {
	return t.EffectiveTheme().Tag.Group(t.Variant())
}

func (t *Tag) resolvedPadding(group *theme.TagGroup) render.Insets {
	if t.padOverride {
		return t.Padding
	}
	return group.Padding
}

func (t *Tag) positionContent() {
	if t.font == nil {
		return
	}
	group := t.effectiveGroup()
	pad := t.resolvedPadding(group)

	tw, th := measureDisplay(t.font, t.text, t.displaySize)

	if t.removable {
		rw, rh := measureDisplay(t.font, "x", t.displaySize)
		totalW := tw + group.Gap + rw
		startX := (t.Width - totalW) / 2
		if startX < pad.Left {
			startX = pad.Left
		}
		t.textNode.SetPosition(startX, (t.Height-th)/2)
		t.removeNode.SetPosition(startX+tw+group.Gap, (t.Height-rh)/2)

		// Set hit shape on remove node for click area.
		rmSize := group.RemoveButtonSize
		if rmSize <= 0 {
			rmSize = rw + 4
		}
		t.removeNode.HitShape = sg.HitRect{
			X: -(rmSize - rw) / 2, Y: -(rmSize - rh) / 2,
			Width: rmSize, Height: rmSize,
		}
	} else {
		t.textNode.SetPosition(
			(t.Width-tw)/2,
			(t.Height-th)/2,
		)
	}
}
