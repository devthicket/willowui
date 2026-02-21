package widget

import (
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// Checkbox is a toggle control with a box and check mark, plus an optional label.
type Checkbox struct {
	Component
	box         *sg.Node // border square (WhitePixel)
	check       *sg.Node // inner fill (WhitePixel)
	label       *Label
	value       *Ref[bool]
	watch       WatchHandle
	onChange    func(bool)
	checkIcon   engine.Image // nil = use theme or procedural default
	appliedIcon engine.Image // tracks last applied icon to avoid redundant work
}

// DefaultCheckboxSize is the default box dimension.
const DefaultCheckboxSize = 20

// DefaultCheckboxInnerSize is the check mark inner size.
const DefaultCheckboxInnerSize = 12

// DefaultCheckboxGap is the spacing between the box and label.
const DefaultCheckboxGap = 8

// NewCheckbox creates a Checkbox with the given label text, font source, and display size.
func NewCheckbox(name string, text string, source *sg.FontFamily, displaySize float64) *Checkbox {
	c := &Checkbox{
		value: NewRef(false),
	}
	initComponent(&c.Component, name)

	// Box border.
	c.box = sg.NewSprite(name+"-box", sg.TextureRegion{})
	c.box.SetScale(DefaultCheckboxSize, DefaultCheckboxSize)
	c.node.AddChild(c.box)

	// Check icon (default: built-in checkmark glyph).
	c.check = sg.NewSprite(name+"-check", sg.TextureRegion{})
	c.check.SetVisible(false)
	c.node.AddChild(c.check)
	c.applyCheckImage(IconCheckmark())

	// Label.
	c.label = NewLabel(name+"-label", text, source, displaySize)
	c.label.SetPosition(DefaultCheckboxSize+DefaultCheckboxGap, (DefaultCheckboxSize-c.label.Height)/2)
	c.label.AddToNode(c.node)

	c.Width = DefaultCheckboxSize + DefaultCheckboxGap + c.label.Width
	c.Height = DefaultCheckboxSize
	if c.label.Height > c.Height {
		c.Height = c.label.Height
	}
	c.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: c.Width, Height: c.Height}

	// Click toggles.
	c.node.OnClick(func(ctx sg.ClickContext) {
		if !c.enabled {
			return
		}
		newVal := !c.value.Peek()
		c.value.Set(newVal)
		DefaultScheduler.Flush()
		c.check.SetVisible(newVal)
		c.UpdateVisuals()
		if c.onChange != nil {
			c.onChange(newVal)
		}
	})

	c.wireVisualCallbacks(c.UpdateVisuals)
	c.SetCursorShape(engine.CursorShapePointer)

	// Focus: checkboxes participate in tab and spatial nav.
	c.enableFocusNavigation()
	c.node.OnUpdate = func(_ float64) {
		if !c.focused || !c.enabled {
			return
		}
		if DefaultInputManager.IsKeyJustAvailable(engine.KeySpace) {
			newVal := !c.value.Peek()
			c.value.Set(newVal)
			DefaultScheduler.Flush()
			c.check.SetVisible(newVal)
			c.UpdateVisuals()
			if c.onChange != nil {
				c.onChange(newVal)
			}
			DefaultInputManager.Consume(engine.KeySpace)
		}
	}

	c.UpdateVisuals()
	return c
}

// Checked returns the current checked state.
func (c *Checkbox) Checked() bool {
	return c.value.Peek()
}

// SetChecked sets the checked state.
func (c *Checkbox) SetChecked(v bool) {
	c.value.Set(v)
	DefaultScheduler.Flush()
	c.check.SetVisible(v)
	c.UpdateVisuals()
}

// SetText sets the checkbox label text.
func (c *Checkbox) SetText(text string) {
	if c.label == nil {
		return
	}
	c.label.SetText(text)
	c.Width = DefaultCheckboxSize + DefaultCheckboxGap + c.label.Width
	c.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: c.Width, Height: c.Height}
	c.node.SetScale(c.Width, c.Height)
	c.node.Invalidate()
}

// SetOnChange sets the callback invoked when checked state changes.
func (c *Checkbox) SetOnChange(fn func(bool)) {
	c.onChange = fn
}

// BindValue binds the checkbox to a reactive Ref[bool].
func (c *Checkbox) BindValue(ref *Ref[bool]) {
	c.watch.Stop()
	c.value = ref
	c.SetChecked(ref.Peek())
	c.watch = WatchValue(ref, func(_, newVal bool) {
		c.check.SetVisible(newVal)
		c.UpdateVisuals()
	})
}

// SetCheckIcon replaces the default filled-square check mark with a custom
// image (e.g. a pixel-art checkmark glyph). The image is rendered at 1:1
// pixel scale, centered in the box.
func (c *Checkbox) SetCheckIcon(img engine.Image) {
	c.checkIcon = img
	c.applyCheckImage(img)
}

// applyCheckImage applies an icon image to the check node, scaling it to fit
// DefaultCheckboxInnerSize and centering it within the box.
func (c *Checkbox) applyCheckImage(img engine.Image) {
	c.appliedIcon = img
	c.check.SetCustomImage(img)
	s := GlyphScale(img, DefaultCheckboxSize)
	c.check.SetScale(s, s)
	offset := 0.0
	c.check.SetPosition(offset, offset)
	c.check.Invalidate()
}

// SetEnabled overrides Component.SetEnabled to also update visuals.
func (c *Checkbox) SetEnabled(v bool) {
	c.Component.SetEnabled(v)
	c.UpdateVisuals()
}

// UpdateVisuals applies theme colors based on current state.
func (c *Checkbox) UpdateVisuals() {
	c.state = computeState(c.enabled, c.focused, c.hovered, c.value.Peek())
	group := c.EffectiveTheme().Checkbox.Group(c.Variant())
	c.box.SetColor(group.BoxColor.Resolve(c.state))
	c.check.SetColor(group.CheckColor.Resolve(c.state))

	// Apply icon: per-instance override > theme icon > default fill.
	// Only re-apply when the resolved image changes.
	if c.checkIcon == nil && group.CheckIcon.Set && c.appliedIcon != group.CheckIcon.Image {
		c.applyCheckImage(group.CheckIcon.Image)
	}

	c.applyFocusRingSize(group.FocusColor.Resolve(c.state), group.FocusRingWidth, DefaultCheckboxSize, DefaultCheckboxSize)
	c.MarkDrawDirty()
}

// Dispose stops reactive watches and disposes children.
func (c *Checkbox) Dispose() {
	c.watch.Stop()
	if c.label != nil {
		c.label.Dispose()
	}
	c.Component.Dispose()
}

// BoxNode returns the border box node. Used for testing.
func (c *Checkbox) BoxNode() *sg.Node { return c.box }

// CheckNode returns the inner check fill node. Used for testing.
func (c *Checkbox) CheckNode() *sg.Node { return c.check }

// CheckboxLabel returns the optional label widget, or nil. Used for testing.
func (c *Checkbox) CheckboxLabel() *Label { return c.label }
