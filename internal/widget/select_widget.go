package widget

import (
	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// SelectOption is a single choice in a Select widget.
type SelectOption struct {
	Label string
	Value any // optional typed payload
}

// Select is a dropdown widget: shows the currently selected option and opens
// a MenuPopup when clicked.
type Select struct {
	Component
	options     []SelectOption
	selected    int
	popup       *MenuPopup
	label       *sg.Node // text node for selected label
	chevron     *sg.Node // down-pointing chevron glyph sprite
	onChange    func(index int, option SelectOption)
	font        *sg.FontFamily
	source      *sg.FontFamily
	displaySize float64

	// Reactive bindings.
	stopOptions func()      // stops options Array binding
	selectedRef *Ref[int]   // set when BindSelected is active
	selWatch    WatchHandle // watches selectedRef for external changes
	ignoreWatch bool        // prevents feedback loop during user interaction
}

// chevronGlyphSize is the desired display size for the dropdown chevron.
const chevronGlyphSize = 9

// NewSelect creates a Select with the given name, options list, font source, and display size.
func NewSelect(name string, options []SelectOption, source *sg.FontFamily, displaySize float64) *Select {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	s := &Select{
		font:        font,
		source:      source,
		displaySize: displaySize,
		selected:    0,
	}
	initComponent(&s.Component, name)
	s.initBackground(name)
	s.initBorder(name)

	s.options = options
	s.Focusable = true
	s.AllowTab = true

	// Label text node.
	s.label = sg.NewText(name+"-label", "", font)
	s.label.TextBlock.FontSize = displaySize
	s.node.AddChild(s.label)

	// Chevron: reuse the same pixel-painted 9×9 down-pointing "v" glyph as the
	// tree-list collapse toggle. SetColor tints it to the theme chevron color.
	s.chevron = sg.NewSprite(name+"-chevron", sg.TextureRegion{})
	s.chevron.SetCustomImage(treeCollapseGlyph())
	s.chevron.SetSize(chevronGlyphSize, chevronGlyphSize)
	s.chevron.Interactable = false
	s.node.AddChild(s.chevron)

	// Click opens the popup.
	s.node.OnClick(func(ctx sg.ClickContext) {
		if ctx.Button != sg.MouseButtonLeft || !s.enabled {
			return
		}
		s.openPopup()
	})

	s.wireVisualCallbacks(s.UpdateVisuals)

	// Keyboard: Space or Enter opens popup when focused.
	s.handleKeyFn = func(k engine.Key) bool {
		if k == engine.KeyEnter || k == engine.KeySpace {
			if !DefaultMenuPopupManager.IsOpen() {
				s.openPopup()
			}
			return true
		}
		return false
	}

	if len(options) > 0 {
		s.updateLabel()
	}

	return s
}

// SetOptions replaces the option list. The selection resets to index 0.
func (s *Select) SetOptions(options []SelectOption) {
	s.options = options
	s.selected = 0
	s.updateLabel()
}

// BindOptions binds the options list to a reactive Array[SelectOption].
// When the array changes the dropdown re-syncs and clamps the selection.
// Pass nil to detach.
func (s *Select) BindOptions(arr *Array[SelectOption]) {
	if s.stopOptions != nil {
		s.stopOptions()
		s.stopOptions = nil
	}
	if arr == nil {
		return
	}
	sync := func() {
		s.options = s.options[:0]
		arr.ForEach(func(_ int, opt SelectOption) {
			s.options = append(s.options, opt)
		})
		if s.selected >= len(s.options) {
			s.selected = max(0, len(s.options)-1)
		}
		s.updateLabel()
	}
	sync()
	h := arr.OnChange(func() { sync() })
	s.stopOptions = func() { h.Stop() }
}

// BindSelected binds the selection index to a reactive Ref[int].
// External changes to the ref update the widget; user interaction updates
// the ref. Replaces any previous binding.
func (s *Select) BindSelected(ref *Ref[int]) {
	s.selWatch.Stop()
	s.selectedRef = ref
	s.SetSelected(ref.Peek())
	s.selWatch = WatchValue(ref, func(_, newIdx int) {
		if s.ignoreWatch {
			return
		}
		s.SetSelected(newIdx)
	})
}

// SelectedRef returns the reactive Ref backing the selection, or nil if
// BindSelected has not been called.
func (s *Select) SelectedRef() *Ref[int] {
	return s.selectedRef
}

// SetSelected selects the option at the given index.
func (s *Select) SetSelected(idx int) {
	if idx < 0 || idx >= len(s.options) {
		return
	}
	s.selected = idx
	s.updateLabel()
}

// Selected returns the index of the currently selected option.
func (s *Select) Selected() int { return s.selected }

// SelectedOption returns the currently selected SelectOption.
func (s *Select) SelectedOption() SelectOption {
	if len(s.options) == 0 {
		return SelectOption{}
	}
	return s.options[s.selected]
}

// SetOnChange registers a callback fired when the user selects an option.
func (s *Select) SetOnChange(fn func(index int, option SelectOption)) {
	s.onChange = fn
}

// SetSize sets the explicit dimensions of the trigger button.
func (s *Select) SetSize(w, h float64) {
	s.Width = w
	s.Height = h
	s.resizeBackground(w, h)
	s.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	s.resizeBorder(w, h)
	s.UpdateVisuals()
	s.MarkLayoutDirty()
}

// UpdateVisuals applies theme colors and repositions children.
func (s *Select) UpdateVisuals() {
	g := s.EffectiveTheme().Select.Group(s.Variant())

	// Corner radius.
	cr := resolveCornerRadius(g.CornerRadius, s.Height)
	s.applyCornerRadius(cr)

	bg := g.Background.Resolve(s.state)
	s.applyBackground(bg)
	s.applyBorder(g.Border.Resolve(s.state), g.BorderWidth, bg)
	s.applyFocusRing(g.FocusColor.Resolve(core.StateFocus), g.FocusRingWidth)

	// Label text color and position.
	textColor := g.TextColor.Resolve(s.state)
	s.label.SetTextColor(textColor)

	pad := g.Padding
	if pad.IsAuto() {
		pad = render.Insets{Top: 6, Right: 8, Bottom: 6, Left: 8}
	}
	s.label.SetPosition(pad.Left, pad.Top)

	// Chevron: right-aligned, vertically centred, tinted to theme color.
	s.chevron.SetColor(g.ChevronColor.Resolve(s.state))
	cx := s.Width - pad.Right - chevronGlyphSize
	cy := (s.Height - chevronGlyphSize) / 2
	s.chevron.SetPosition(cx, cy)

	s.MarkDrawDirty()
}

// Dispose stops reactive watches and disposes the component.
func (s *Select) Dispose() {
	if s.stopOptions != nil {
		s.stopOptions()
	}
	s.selWatch.Stop()
	s.Component.Dispose()
}

func (s *Select) updateLabel() {
	if len(s.options) == 0 {
		s.label.SetContent("")
	} else {
		s.label.SetContent(s.options[s.selected].Label)
	}
	s.MarkDrawDirty()
}

func (s *Select) openPopup() {
	if len(s.options) == 0 {
		return
	}
	if s.popup == nil {
		s.popup = NewMenuPopup(s.source, s.displaySize)
	}

	// Build menu items from options.
	items := make([]MenuItem, len(s.options))
	for i, opt := range s.options {
		idx := i
		items[i] = MenuItem{
			Label: opt.Label,
			OnSelect: func() {
				prev := s.selected
				s.selected = idx
				s.updateLabel()
				// Update reactive ref if bound.
				if s.selectedRef != nil {
					s.ignoreWatch = true
					s.selectedRef.Set(idx)
					DefaultScheduler.Flush()
					s.ignoreWatch = false
				}
				if s.onChange != nil && prev != idx {
					s.onChange(idx, s.options[idx])
				}
			},
		}
	}
	s.popup.SetItems(items)
	s.popup.minWidth = s.Width
	s.popup.selectedIdx = s.selected
	DefaultMenuPopupManager.Show(s.popup, &s.Component)
	// Pre-highlight and scroll to the current selection.
	DefaultMenuPopupManager.setHighlight(s.selected)
	s.popup.scrollToHighlighted()
}
