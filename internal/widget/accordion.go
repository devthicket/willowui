package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

// Accordion is a vertically stacked list of collapsible sections, each with
// a header row and an arbitrary content panel.
type Accordion struct {
	Component
	sections  []*accordionSection
	exclusive bool // only one section open at a time (default true)
	animated  bool // animate expand/collapse (default true)
	onToggle  func(id string, expanded bool)

	font        *sg.FontFamily
	displaySize float64

	// Reactive binding for the expanded section ID.
	expandedRef   *Ref[string]
	expandedWatch WatchHandle
}

// AccordionSection defines a section to add to an Accordion.
type AccordionSection struct {
	ID      string
	Label   string
	Icon    sg.TextureRegion // zero value = no icon
	Content *Component
}

// accordionSection is the internal representation of an accordion section.
type accordionSection struct {
	id       string
	label    string
	icon     sg.TextureRegion
	content  *Component
	expanded bool

	// Node tree
	headerNode  *sg.Node // container for header row
	headerBg    *sg.Node // background sprite for header
	labelNode   *sg.Node // text node for label
	iconNode    *sg.Node // sprite for icon (nil if no icon)
	chevronNode *sg.Node // chevron indicator

	contentWrap *sg.Node // container that holds the content node
	maskRoot    *sg.Node // mask for clipping content during animation
	maskSprite  *sg.Node // mask rectangle

	// Divider
	divider *sg.Node

	// Animation
	tween         *gween.Tween
	currentHeight float64 // current animated height of content area
	targetHeight  float64 // target height (content natural height or 0)
}

// NewAccordion creates an Accordion with default settings.
func NewAccordion(name string) *Accordion {
	a := &Accordion{
		exclusive: true,
		animated:  true,
	}
	initComponent(&a.Component, name)
	a.initBackground(name)
	a.initBorder(name)

	a.onThemeChange = func() { a.applyThemeColors() }
	a.applyThemeColors()

	// OnUpdate drives animation.
	a.node.OnUpdate = func(dt float64) {
		a.Update(float32(dt))
	}

	return a
}

// SetFont sets the font source used for section header labels.
func (a *Accordion) SetFont(source *sg.FontFamily, size float64) {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	a.font = font
	a.displaySize = size
}

// AddSection appends a section to the accordion and returns the accordion for chaining.
func (a *Accordion) AddSection(section AccordionSection) *Accordion {
	group := a.EffectiveTheme().Accordion.Group(a.Variant())

	s := &accordionSection{
		id:      section.ID,
		label:   section.Label,
		icon:    section.Icon,
		content: section.Content,
	}

	headerH := group.HeaderHeight
	if headerH <= 0 {
		headerH = 36
	}

	// Header container.
	s.headerNode = sg.NewContainer(a.node.Name + "-header-" + section.ID)
	s.headerNode.HitShape = sg.HitRect{X: 0, Y: 0, Width: a.Width, Height: headerH}
	s.headerNode.Interactable = true

	// Header background.
	s.headerBg = sg.NewSprite(a.node.Name+"-hdr-bg-"+section.ID, sg.TextureRegion{})
	s.headerBg.SetScale(a.Width, headerH)
	s.headerNode.AddChild(s.headerBg)

	// Chevron (positioned correctly after image is set via updateChevron).
	s.chevronNode = sg.NewSprite(a.node.Name+"-chevron-"+section.ID, sg.TextureRegion{})
	s.headerNode.AddChild(s.chevronNode)

	pad := group.HeaderPadding
	chevronSize := group.ChevronSize
	if chevronSize <= 0 {
		chevronSize = 12
	}

	// Icon (optional).
	iconOffset := pad.Left + chevronSize + group.HeaderIconGap
	if section.Icon != (sg.TextureRegion{}) {
		iconSize := group.HeaderIconSize
		if iconSize <= 0 {
			iconSize = 16
		}
		s.iconNode = sg.NewSprite(a.node.Name+"-icon-"+section.ID, section.Icon)
		s.iconNode.SetScale(iconSize, iconSize)
		s.iconNode.SetPosition(iconOffset, (headerH-iconSize)/2)
		s.headerNode.AddChild(s.iconNode)
		iconOffset += iconSize + group.HeaderIconGap
	}

	// Label text.
	if a.font != nil {
		s.labelNode = sg.NewText(a.node.Name+"-label-"+section.ID, section.Label, a.font)
		s.labelNode.TextBlock.FontSize = a.displaySize
		s.labelNode.SetPosition(iconOffset, (headerH-a.displaySize)/2)
		s.headerNode.AddChild(s.labelNode)
	}

	// Click handler on header.
	s.headerNode.OnClick(func(ctx sg.ClickContext) {
		a.Toggle(section.ID)
	})

	// Hover cursor.
	s.headerNode.OnPointerEnter(func(ctx sg.PointerContext) {
		engine.SetCursorShape(engine.CursorShapePointer)
	})
	s.headerNode.OnPointerLeave(func(ctx sg.PointerContext) {
		engine.SetCursorShape(engine.CursorShapeDefault)
	})

	a.node.AddChild(s.headerNode)

	// Content wrapper with mask for clipping during animation.
	s.contentWrap = sg.NewContainer(a.node.Name + "-content-" + section.ID)
	if section.Content != nil {
		section.Content.Node().SetPosition(group.ContentPadding.Left, group.ContentPadding.Top)
		s.contentWrap.AddChild(section.Content.Node())
	}
	s.contentWrap.SetVisible(false)
	s.currentHeight = 0

	// Mask for content clipping.
	s.maskRoot = sg.NewContainer(a.node.Name + "-cmask-" + section.ID)
	s.maskSprite = sg.NewSprite(a.node.Name+"-cmask-rect-"+section.ID, sg.TextureRegion{})
	s.maskSprite.SetColor(sg.RGBA(1, 1, 1, 1))
	s.maskSprite.SetScale(a.Width, 0)
	s.maskRoot.AddChild(s.maskSprite)
	s.contentWrap.SetMask(s.maskRoot)

	a.node.AddChild(s.contentWrap)

	// Divider between sections.
	divH := group.DividerHeight
	if divH <= 0 {
		divH = 1
	}
	s.divider = sg.NewSprite(a.node.Name+"-div-"+section.ID, sg.TextureRegion{})
	s.divider.SetScale(a.Width, divH)
	a.node.AddChild(s.divider)

	a.sections = append(a.sections, s)

	// Apply theme colors to this section.
	a.applyThemeColors()
	a.layoutSections()
	return a
}

// RemoveSection removes a section by ID.
func (a *Accordion) RemoveSection(id string) {
	for i, s := range a.sections {
		if s.id == id {
			a.node.RemoveChild(s.headerNode)
			a.node.RemoveChild(s.contentWrap)
			a.node.RemoveChild(s.divider)
			a.sections = append(a.sections[:i], a.sections[i+1:]...)
			a.layoutSections()
			return
		}
	}
}

// Section returns a pointer to the AccordionSection data for the given ID, or nil.
func (a *Accordion) Section(id string) *AccordionSection {
	for _, s := range a.sections {
		if s.id == id {
			return &AccordionSection{
				ID:      s.id,
				Label:   s.label,
				Icon:    s.icon,
				Content: s.content,
			}
		}
	}
	return nil
}

// Open expands a section by ID.
func (a *Accordion) Open(id string) {
	a.SetExpanded(id, true)
}

// Close collapses a section by ID.
func (a *Accordion) Close(id string) {
	a.SetExpanded(id, false)
}

// Toggle toggles a section's expanded state.
func (a *Accordion) Toggle(id string) {
	for _, s := range a.sections {
		if s.id == id {
			a.SetExpanded(id, !s.expanded)
			return
		}
	}
}

// SetExpanded sets the expanded state of a section.
func (a *Accordion) SetExpanded(id string, v bool) {
	for _, s := range a.sections {
		if s.id == id {
			if s.expanded == v {
				return
			}
			// In exclusive mode, close others first.
			if v && a.exclusive {
				for _, other := range a.sections {
					if other.id != id && other.expanded {
						a.collapseSection(other)
					}
				}
			}
			if v {
				a.expandSection(s)
			} else {
				a.collapseSection(s)
			}
			if a.expandedRef != nil {
				a.expandedRef.Set(a.ExpandedID())
			}
			if a.onToggle != nil {
				a.onToggle(id, v)
			}
			return
		}
	}
}

// IsExpanded returns whether a section is expanded.
func (a *Accordion) IsExpanded(id string) bool {
	for _, s := range a.sections {
		if s.id == id {
			return s.expanded
		}
	}
	return false
}

// SetExclusive sets whether only one section can be open at a time.
func (a *Accordion) SetExclusive(v bool) {
	a.exclusive = v
}

// SetAnimated enables or disables expand/collapse animation.
func (a *Accordion) SetAnimated(v bool) {
	a.animated = v
}

// SetOnToggle sets the callback invoked when a section is toggled.
func (a *Accordion) SetOnToggle(fn func(id string, expanded bool)) {
	a.onToggle = fn
}

// BindExpanded binds the expanded section ID to a reactive Ref. In exclusive
// mode this is the single open section's ID (or "" when all are collapsed).
// In non-exclusive mode this tracks the most recently toggled section.
func (a *Accordion) BindExpanded(ref *Ref[string]) {
	a.expandedWatch.Stop()
	a.expandedRef = ref
	// Sync current ref value into widget.
	if id := ref.Peek(); id != "" {
		a.SetExpanded(id, true)
	}
	a.expandedWatch = WatchValue(ref, func(_, newID string) {
		if newID == "" {
			// Close all sections.
			for _, s := range a.sections {
				if s.expanded {
					a.collapseSection(s)
				}
			}
		} else {
			a.SetExpanded(newID, true)
		}
	})
}

// ExpandedID returns the ID of the currently expanded section (first found),
// or "" if none are expanded.
func (a *Accordion) ExpandedID() string {
	for _, s := range a.sections {
		if s.expanded {
			return s.id
		}
	}
	return ""
}

// SetSize sets the accordion dimensions.
func (a *Accordion) SetSize(w, h float64) {
	a.Width = w
	a.Height = h
	a.resizeBackground(w, h)
	a.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	a.resizeBorder(w, h)
	a.layoutSections()
	a.MarkLayoutDirty()
}

// Update advances any active expand/collapse animations.
func (a *Accordion) Update(dt float32) {
	needsLayout := false
	for _, s := range a.sections {
		if s.tween != nil {
			h, done := s.tween.Update(dt)
			s.currentHeight = float64(h)
			if done {
				s.tween = nil
				if !s.expanded {
					s.contentWrap.SetVisible(false)
				}
			}
			// Update mask to clip content.
			s.maskSprite.SetScale(a.Width, s.currentHeight)
			needsLayout = true
		}
	}
	if needsLayout {
		a.layoutSections()
	}
}

// Dispose cleans up the accordion.
func (a *Accordion) Dispose() {
	a.expandedWatch.Stop()
	for _, s := range a.sections {
		s.tween = nil
	}
	a.Component.Dispose()
}

// expandSection starts expanding a section.
func (a *Accordion) expandSection(s *accordionSection) {
	s.expanded = true
	s.contentWrap.SetVisible(true)

	group := a.EffectiveTheme().Accordion.Group(a.Variant())
	s.targetHeight = a.contentHeight(s)

	if a.animated {
		dur := float32(group.AnimationDuration)
		if dur <= 0 {
			dur = 0.2
		}
		s.tween = gween.New(float32(s.currentHeight), float32(s.targetHeight), dur, ease.OutCubic)
	} else {
		s.currentHeight = s.targetHeight
		s.tween = nil
		s.maskSprite.SetScale(a.Width, s.currentHeight)
	}

	a.updateChevron(s)
	a.layoutSections()
}

// collapseSection starts collapsing a section.
func (a *Accordion) collapseSection(s *accordionSection) {
	s.expanded = false
	s.targetHeight = 0

	group := a.EffectiveTheme().Accordion.Group(a.Variant())

	if a.animated {
		dur := float32(group.AnimationDuration)
		if dur <= 0 {
			dur = 0.2
		}
		s.tween = gween.New(float32(s.currentHeight), 0, dur, ease.OutCubic)
	} else {
		s.currentHeight = 0
		s.tween = nil
		s.contentWrap.SetVisible(false)
		s.maskSprite.SetScale(a.Width, 0)
	}

	a.updateChevron(s)
	a.layoutSections()
}

// contentHeight computes the full height of a section's content including padding.
func (a *Accordion) contentHeight(s *accordionSection) float64 {
	if s.content == nil {
		return 0
	}
	group := a.EffectiveTheme().Accordion.Group(a.Variant())
	return s.content.Height + group.ContentPadding.Top + group.ContentPadding.Bottom
}

// layoutSections positions all section nodes vertically.
func (a *Accordion) layoutSections() {
	group := a.EffectiveTheme().Accordion.Group(a.Variant())
	headerH := group.HeaderHeight
	if headerH <= 0 {
		headerH = 36
	}
	divH := group.DividerHeight
	if divH <= 0 {
		divH = 1
	}

	y := 0.0
	for i, s := range a.sections {
		// Header.
		s.headerNode.SetPosition(0, y)
		s.headerBg.SetScale(a.Width, headerH)
		s.headerNode.HitShape = sg.HitRect{X: 0, Y: 0, Width: a.Width, Height: headerH}
		y += headerH

		// Content (may be animating).
		s.contentWrap.SetPosition(0, y)
		if s.expanded || s.tween != nil {
			y += s.currentHeight
		}

		// Divider (between sections, not after last).
		if i < len(a.sections)-1 {
			s.divider.SetPosition(0, y)
			s.divider.SetScale(a.Width, divH)
			s.divider.SetVisible(true)
			y += divH
		} else {
			s.divider.SetVisible(false)
		}
	}

	// Auto-size height if set to 0.
	if a.Height == 0 || a.Height != y {
		a.Height = y
		a.resizeBackground(a.Width, y)
		a.resizeBorder(a.Width, y)
		a.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: a.Width, Height: y}
	}

	a.MarkDrawDirty()
}

// updateChevron sets the correct chevron glyph for a section and positions it.
func (a *Accordion) updateChevron(s *accordionSection) {
	group := a.EffectiveTheme().Accordion.Group(a.Variant())
	headerH := group.HeaderHeight
	if headerH <= 0 {
		headerH = 36
	}
	chevronSize := group.ChevronSize
	if chevronSize <= 0 {
		chevronSize = 12
	}
	pad := group.HeaderPadding

	var glyph engine.Image
	if s.expanded {
		if group.CollapseIcon.Set {
			glyph = group.CollapseIcon.Image
		}
		if glyph == nil {
			glyph = treeCollapseGlyph()
		}
	} else {
		if group.ExpandIcon.Set {
			glyph = group.ExpandIcon.Image
		}
		if glyph == nil {
			glyph = treeExpandGlyph()
		}
	}

	s.chevronNode.SetCustomImage(glyph)
	// Display size: use the smaller of native image size and chevron allocation
	// so spritesheet glyphs (48px) scale down while custom icons stay native.
	b := glyph.Bounds()
	displayW := math.Min(float64(b.Dx()), chevronSize)
	displayH := math.Min(float64(b.Dy()), chevronSize)
	s.chevronNode.SetSize(displayW, displayH)
	// Center the glyph within the chevronSize allocation at pad.Left.
	offX := pad.Left + (chevronSize-displayW)/2
	offY := (headerH - displayH) / 2
	s.chevronNode.SetPosition(offX, offY)
}

// applyThemeColors applies theme colors to all sections.
func (a *Accordion) applyThemeColors() {
	group := a.EffectiveTheme().Accordion.Group(a.Variant())

	cr := resolveCornerRadius(group.CornerRadius, a.Height)
	a.applyCornerRadius(cr)
	a.applyBackground(group.Background.Resolve(a.state))
	a.applyBorder(group.BorderColor.Resolve(a.state), group.BorderWidth, group.Background.Resolve(a.state))

	for _, s := range a.sections {
		// Header background.
		hdrBg := group.HeaderBackground.Resolve(a.state)
		if hdrBg.Type != BgNone {
			s.headerBg.SetColor(hdrBg.Color)
			s.headerBg.SetVisible(true)
		} else {
			s.headerBg.SetVisible(false)
		}

		// Header text color.
		if s.labelNode != nil {
			s.labelNode.SetColor(group.HeaderTextColor.Resolve(a.state))
		}

		// Chevron color.
		s.chevronNode.SetColor(group.ChevronColor.Resolve(a.state))

		// Divider color.
		s.divider.SetColor(group.DividerColor.Resolve(a.state))

		// Set chevron image.
		a.updateChevron(s)
	}

	a.MarkDrawDirty()
}
