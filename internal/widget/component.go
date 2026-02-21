package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// UIElement is implemented by all UI component types in this package.
// The unexported method prevents external types from implementing it.
type UIElement interface {
	base() *Component
}

// Component is the base type for all WillowUI widgets. It wraps a sg.Node
// container and provides layout, state tracking, and dirty flag propagation.
type Component struct {
	node     *sg.Node // root container node
	parent   *Component
	children []*Component

	// Layout mode and configuration.
	Layout      LayoutMode
	GridColumns int
	Spacing     float64
	FlowRowGap  float64   // vertical gap between rows in LayoutFlow; falls back to Spacing when zero
	Align       Alignment // cross-axis alignment
	Justify     Alignment // main-axis alignment

	// Position and size.
	X, Y             float64
	OffsetX, OffsetY float64 // user offset applied on top of layout-computed position
	Width, Height    float64

	// Constraints.
	MinWidth, MinHeight float64
	MaxWidth, MaxHeight float64

	// Fill stretches the child to match the parent's content dimension.
	Fill FillMode

	// Grow is a flex-like weight for proportional sizing in HBox/VBox.
	// Children with Grow > 0 share remaining space after fixed children.
	Grow int

	// Insets.
	Padding Insets
	Margin  Insets

	// State.
	enabled bool
	focused bool
	hovered bool
	pressed bool
	state   ComponentState

	// Focus navigation flags. Disabled by default; interactive widgets
	// opt in by setting these in their constructors.
	Focusable          bool // can receive keyboard focus
	AllowTab           bool // participates in Tab/Shift+Tab cycling
	AllowSpatial       bool // participates in arrow-key spatial navigation
	InterceptArrows    bool // wants to handle arrow keys before spatial nav
	ConsumeHandledKeys bool // when HandleKey returns true, also consume from InputManager (default true)

	// handleKeyFn is called by HandleKey when this component has focus.
	// It is a pure query with no side effects — return true if the widget
	// will act on the key this frame, false to let FocusManager do spatial nav.
	handleKeyFn func(engine.Key) bool

	// Cursor shape shown when hovering over this component.
	// 0 (CursorShapeDefault) means no override.
	cursorShape engine.CursorShapeType

	// Dirty flags.
	dirtyLayout bool
	dirtyDraw   bool

	// onFocusChange is called after the focused state changes. Widgets
	// that need to update visuals on focus/blur set this in their
	// constructor.
	onFocusChange func(bool)

	// onVisualStateChange is called after hovered or pressed state
	// changes. Widgets set this to their UpdateVisuals method so
	// colors update automatically without per-frame polling.
	onVisualStateChange func()

	// onActivate is called when this component is activated (e.g. a
	// Window being brought to front). Set by WindowManager.
	onActivate func()

	// customEvents holds named event handlers registered by custom widgets
	// for use with XML template on:* bindings (e.g. on:change, on:select).
	customEvents map[string]func()

	// ensureChildVisible is set by scrollable containers (ScrollPanel) to
	// scroll a focused child into view. Called during SetFocused ancestor walk.
	ensureChildVisible func(child *Component)

	// onLayout is called during UpdateLayout before children are recursed.
	onLayout func()

	// anchoredChildren stores per-child anchor metadata for LayoutAnchor mode.
	// Lazily initialized. The zero value (AnchorTopLeft, 0, 0) is the default
	// applied when a child is added via AddChild without explicit anchor data.
	anchoredChildren map[*Component]anchorEntry

	// Background node fields (owned by Component, managed via initBackground).
	bgNode           *sg.Node        // WhitePixel sprite for solid color backgrounds
	bgContainer      *sg.Node        // container for nine-slice sprites (lazy)
	bgSliceNodes     *nineSliceNodes // the 9 sprites (lazy, nil until first nine-slice use)
	bgSlice          *NineSlice      // current nine-slice config (nil when solid/none)
	bgCenterFillMesh *sg.Node        // gradient mesh replacing nine-grid center cell (lazy)

	// Gradient mesh (lazy, nil until first gradient use).
	bgGradientMesh *sg.Node

	// Corner radius fields (owned by Component).
	cornerRadius    float64    // uniform corner radius (0 = sharp)
	perCornerRadius bool       // true when per-corner radii are active
	cornerRadii     [4]float64 // TL, TR, BR, BL — used when perCornerRadius is true
	bgPoly          *sg.Node   // polygon mesh for rounded bg (nil when sharp)
	borderPoly      *sg.Node   // mesh for rounded border (nil when sharp)

	// Border node fields (owned by Component, managed via initBorder).
	borderTop    *sg.Node
	borderRight  *sg.Node
	borderBot    *sg.Node
	borderLeft   *sg.Node
	borderWidth_ float64

	// Focus ring node (lazy-initialized on first focus).
	focusRing      *sg.Node
	focusRingShown bool
	focusRingTween *sg.TweenGroup // active fade tween; cancelled before starting a new one

	// theme is the explicit theme set on this component. nil means inherit
	// from the parent component.
	theme *Theme

	// variant selects the color group for this component.
	variant Variant

	// onThemeChange is called when the effective theme changes.
	// Components set this to their theme-reapplication method.
	onThemeChange func()

	// Reactive binding handles for enabled and visible state.
	enabledWatch WatchHandle
	visibleWatch WatchHandle

	// Tooltip attached to this component (nil if none).
	tooltip       *Tooltip
	onTooltipShow func() // fires just before tooltip becomes visible
	onTooltipHide func() // fires just after tooltip is hidden

	// ContextMenu attached to this component (nil if none).
	contextMenu *ContextMenu
}

// Insets represents spacing on four sides (top, right, bottom, left).
// It is an alias for render.Insets.
type Insets = render.Insets

// AutoPadding is the sentinel Insets value meaning "use the component's
// built-in default padding". All four fields are -1.
var AutoPadding = core.AutoPadding

// resolveAutoInsets returns fallback if i is auto, otherwise returns i unchanged.
func resolveAutoInsets(i, fallback Insets) Insets {
	return core.ResolveAutoInsets(i, fallback)
}

// Default auto-padding values per component type.
var (
	defaultButtonPadding    = core.DefaultButtonPadding
	defaultTextInputPadding = core.DefaultTextInputPadding
	defaultTextAreaPadding  = core.DefaultTextAreaPadding
	defaultBarPadding       = core.DefaultBarPadding
)

// NewComponent creates a new Component with sensible defaults.
// The underlying sg.Node is a container with pointer callbacks wired
// for hover and pressed state tracking.
func NewComponent(name string) *Component {
	c := &Component{}
	initComponent(c, name)
	return c
}

// initComponent wires up the sg.Node and pointer callbacks for c.
// It must be called exactly once. Use this instead of NewComponent when
// c is already allocated (e.g. as an embedded field in a larger struct)
// to avoid the dangling-closure problem that arises from value-copying
// a Component returned by NewComponent.
func initComponent(c *Component, name string) {
	c.enabled = true
	c.ConsumeHandledKeys = true
	c.dirtyLayout = true
	c.dirtyDraw = true

	node := sg.NewContainer(name)
	node.Interactable = true
	c.node = node

	// Store back-reference so pointer callbacks can find the Component.
	node.UserData = c

	// Track hovered state via enter/leave.
	node.OnPointerEnter(func(_ sg.PointerContext) {
		c.hovered = true
		if c.cursorShape != engine.CursorShapeDefault {
			engine.SetCursorShape(c.cursorShape)
		}
		c.MarkDrawDirty()
		if c.onVisualStateChange != nil {
			c.onVisualStateChange()
		}
		DefaultTooltipManager.onTriggerEnter(c)
	})
	node.OnPointerLeave(func(_ sg.PointerContext) {
		c.hovered = false
		c.pressed = false
		if c.cursorShape != engine.CursorShapeDefault {
			engine.SetCursorShape(engine.CursorShapeDefault)
		}
		c.MarkDrawDirty()
		if c.onVisualStateChange != nil {
			c.onVisualStateChange()
		}
		DefaultTooltipManager.onTriggerLeave(c)
	})

	// Track pressed state via down/up.
	node.OnPointerDown(func(ctx sg.PointerContext) {
		// Right-click: show context menu if attached.
		if c.contextMenu != nil && ctx.Button == sg.MouseButtonRight && c.enabled {
			c.contextMenu.ShowAt(ctx.GlobalX, ctx.GlobalY)
			return
		}
		if c.enabled {
			c.pressed = true
			c.MarkDrawDirty()
			if c.onVisualStateChange != nil {
				c.onVisualStateChange()
			}
		}
		c.bubbleActivation()
	})
	node.OnPointerUp(func(_ sg.PointerContext) {
		c.pressed = false
		c.MarkDrawDirty()
		if c.onVisualStateChange != nil {
			c.onVisualStateChange()
		}
	})
}

// bubbleActivation walks up the willow node tree and fires the first
// onActivate callback found on an ancestor Component. This lets clicks
// on any descendant (e.g. a TextInput inside a Window) activate the
// ancestor container, regardless of whether the Component parent chain
// or just the willow node tree was used to add children.
func (c *Component) bubbleActivation() {
	for n := c.node; n != nil; n = n.Parent {
		if comp, ok := n.UserData.(*Component); ok && comp.onActivate != nil {
			comp.onActivate()
			return
		}
	}
}

func (c *Component) base() *Component { return c }

// BaseComp returns the embedded Component. Used for testing.
func (c *Component) BaseComp() *Component { return c }

// SetOnThemeChangeForTest sets the onThemeChange callback. Used for testing.
func (c *Component) SetOnThemeChangeForTest(fn func()) { c.onThemeChange = fn }

// BgPoly returns the background polygon node (for rounded corners). Used for testing.
func (c *Component) BgPoly() *sg.Node { return c.bgPoly }

// BorderPoly returns the border polygon node (for rounded corners). Used for testing.
func (c *Component) BorderPoly() *sg.Node { return c.borderPoly }

// SetOnActivate sets the callback invoked when this component is activated
// (e.g. a Window being brought to front, or a generic click handler via templates).
func (c *Component) SetOnActivate(fn func()) {
	c.onActivate = fn
}

// SetOnEvent registers a named event handler for use with XML template
// on:* bindings. Custom widgets call FireEvent(name) to trigger the handler.
func (c *Component) SetOnEvent(name string, fn func()) {
	if c.customEvents == nil {
		c.customEvents = make(map[string]func())
	}
	c.customEvents[name] = fn
}

// FireEvent triggers a named custom event handler previously registered
// via SetOnEvent or XML template on:* bindings.
func (c *Component) FireEvent(name string) {
	if c.customEvents != nil {
		if fn := c.customEvents[name]; fn != nil {
			fn()
		}
	}
}

// BgNode returns the component's background willow node, or nil if none has been
// initialized. Used for testing visual state.
func (c *Component) BgNode() *sg.Node {
	return c.bgNode
}

// BgGradientMesh returns the component's gradient mesh node, or nil if none has
// been created. Used for testing gradient background state.
func (c *Component) BgGradientMesh() *sg.Node {
	return c.bgGradientMesh
}

// BgContainer returns the component's nine-slice container node, or nil if
// none has been created yet. Used for testing nine-slice background state.
func (c *Component) BgContainer() *sg.Node {
	return c.bgContainer
}

// HasBgSliceNodes reports whether the nine-slice sprite nodes have been created.
// Used for testing nine-slice lazy initialization.
func (c *Component) HasBgSliceNodes() bool {
	return c.bgSliceNodes != nil
}

// BorderTop returns the top border node, or nil if not initialized.
func (c *Component) BorderTop() *sg.Node { return c.borderTop }

// BorderRight returns the right border node, or nil if not initialized.
func (c *Component) BorderRight() *sg.Node { return c.borderRight }

// BorderBot returns the bottom border node, or nil if not initialized.
func (c *Component) BorderBot() *sg.Node { return c.borderBot }

// BorderLeft returns the left border node, or nil if not initialized.
func (c *Component) BorderLeft() *sg.Node { return c.borderLeft }

// BorderWidth returns the current border width.
func (c *Component) BorderWidth() float64 { return c.borderWidth_ }

// InitBorderForTest calls the internal initBorder method. Intended for unit
// tests that need to set up border nodes outside of a widget constructor.
func (c *Component) InitBorderForTest(name string) {
	c.initBorder(name)
}

// ApplyBorderForTest calls the internal applyBorder method. Intended for unit
// tests that need to drive border-switching logic directly.
func (c *Component) ApplyBorderForTest(color sg.Color, width float64, bg Background) {
	c.applyBorder(color, width, bg)
}

// ResizeBorderForTest calls the internal resizeBorder method. Intended for unit
// tests that need to drive border resize logic directly.
func (c *Component) ResizeBorderForTest(w, h float64) {
	c.resizeBorder(w, h)
}

// ResizeBackgroundForTest calls the internal resizeBackground method. Intended
// for unit tests that need to drive background resize logic directly.
func (c *Component) ResizeBackgroundForTest(w, h float64) {
	c.resizeBackground(w, h)
}

// SimulateHover sets the hovered state directly. Intended for unit tests that
// need to drive visual-state logic without a running input pipeline.
func (c *Component) SimulateHover(v bool) {
	c.hovered = v
}

// SimulatePress sets the pressed state directly. Intended for unit tests that
// need to drive visual-state logic without a running input pipeline.
func (c *Component) SimulatePress(v bool) {
	c.pressed = v
}

// SimulateDirtyDraw sets the dirtyDraw flag directly. Intended for unit tests
// that need to reset dirty state before testing MarkDrawDirty.
func (c *Component) SimulateDirtyDraw(v bool) {
	c.dirtyDraw = v
}

// InitBackgroundForTest calls the internal initBackground method. Intended for
// unit tests that need to set up a Component background outside of a widget
// constructor (e.g. to test gradient or nine-slice behaviour in isolation).
func (c *Component) InitBackgroundForTest(name string) {
	c.initBackground(name)
}

// ApplyBackgroundForTest calls the internal applyBackground method. Intended for
// unit tests that need to drive background-switching logic directly.
func (c *Component) ApplyBackgroundForTest(bg Background) {
	c.applyBackground(bg)
}

// State returns the current ComponentState.
func (c *Component) State() ComponentState {
	return c.state
}

// Node returns the underlying willow node. Use Component methods (SetPosition,
// SetVisible, SetZIndex, etc.) in preference to calling Node() directly.
// Node() is an escape hatch for low-level willow operations not covered by
// Component wrappers (e.g. HitShape, Interactable, AddChild, event callbacks).
func (c *Component) Node() *sg.Node {
	return c.node
}

// Name returns the name of the underlying willow node.
func (c *Component) Name() string { return c.node.Name }

// IsDisposed reports whether the underlying node has been disposed.
func (c *Component) IsDisposed() bool { return c.node.IsDisposed() }

// ZIndex returns the z-index of the underlying node.
func (c *Component) ZIndex() int { return c.node.ZIndex() }

// SetZIndex sets the z-index of the underlying node.
func (c *Component) SetZIndex(z int) { c.node.SetZIndex(z) }

// UserData returns the UserData field of the underlying node.
func (c *Component) UserData() any { return c.node.UserData }

// SetUserData sets the UserData field of the underlying node.
func (c *Component) SetUserData(v any) { c.node.UserData = v }

// SetInteractable sets whether the underlying node receives pointer events.
// Unlike SetEnabled, this does not affect visual state tracking.
func (c *Component) SetInteractable(v bool) { c.node.Interactable = v }

// IsInteractable reports whether the underlying node receives pointer events.
func (c *Component) IsInteractable() bool { return c.node.Interactable }

// OnClick registers a click callback on the underlying node.
// If no hit shape has been established yet and the component has dimensions,
// one is created automatically so the component becomes a pointer target.
func (c *Component) OnClick(fn func(sg.ClickContext)) {
	if c.node.HitShape == nil && c.Width > 0 && c.Height > 0 {
		c.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: c.Width, Height: c.Height}
	}
	c.node.OnClick(fn)
}

// SetHitShape sets the hit-test shape for the underlying node.
func (c *Component) SetHitShape(s sg.HitShape) { c.node.HitShape = s }

// HitShape returns the current hit-test shape of the underlying node.
func (c *Component) HitShape() sg.HitShape { return c.node.HitShape }

// AddRawChild adds a raw willow node as a child of this component's node.
// Use this when adding nodes that are not WillowUI components, such as
// nodes created with willow.NewSprite or willow.NewText.
func (c *Component) AddRawChild(node *sg.Node) {
	if node == nil {
		return
	}
	c.node.AddChild(node)
}

// RemoveRawChild removes a raw willow node from this component's node.
func (c *Component) RemoveRawChild(node *sg.Node) {
	if node == nil {
		return
	}
	c.node.RemoveChild(node)
}

// SetPosition sets the component position in local parent space.
// This forwards to the underlying node while keeping layout fields in sync.
func (c *Component) SetPosition(x, y float64) {
	c.X = x
	c.Y = y
	c.node.SetPosition(x+c.OffsetX, y+c.OffsetY)
	c.MarkLayoutDirty()
}

// AddToNode adds this component's node as a child of parent.
// It also applies any pending layout so components with VBox/HBox children
// position correctly without requiring a Screen controller.
func (c *Component) AddToNode(parent *sg.Node) {
	if parent == nil {
		return
	}
	parent.AddChild(c.node)
	c.UpdateLayout()
}

// AddToScene adds this component to the scene root node.
// It also applies any pending layout so components with VBox/HBox children
// position correctly without requiring a Screen controller.
func (c *Component) AddToScene(scene *sg.Scene) {
	if scene == nil || scene.Root == nil {
		return
	}
	scene.Root.AddChild(c.node)
	c.UpdateLayout()
}

// --- Enabled / Visible ---

// SetEnabled sets whether the component responds to input.
func (c *Component) SetEnabled(v bool) {
	if c.enabled == v {
		return
	}
	c.enabled = v
	c.node.Interactable = v
	if !v {
		c.pressed = false
		c.hovered = false
	}
	if c.onVisualStateChange != nil {
		c.onVisualStateChange()
	}
	c.MarkDrawDirty()
}

// IsEnabled reports whether the component is enabled.
func (c *Component) IsEnabled() bool {
	return c.enabled
}

// SetVisible sets the component's visibility.
func (c *Component) SetVisible(v bool) {
	c.node.SetVisible(v)
	c.MarkLayoutDirty()
}

// IsVisible reports whether the component is visible.
func (c *Component) IsVisible() bool {
	return c.node.Visible()
}

// BindEnabled binds the component's enabled state to a reactive Ref[bool].
// Any previous binding is stopped first.
func (c *Component) BindEnabled(ref *Ref[bool]) {
	c.enabledWatch.Stop()
	c.enabledWatch = WatchValue(ref, func(_, v bool) {
		c.SetEnabled(v)
	})
}

// BindVisible binds the component's visibility to a reactive Ref[bool].
// Any previous binding is stopped first.
func (c *Component) BindVisible(ref *Ref[bool]) {
	c.visibleWatch.Stop()
	c.visibleWatch = WatchValue(ref, func(_, v bool) {
		c.SetVisible(v)
	})
}

// SetCursorShape sets the cursor shape shown when hovering over this component.
// Use engine.CursorShapeDefault to clear any override.
func (c *Component) SetCursorShape(shape engine.CursorShapeType) {
	c.cursorShape = shape
}

// --- Tooltip ---

// SetTooltip attaches a tooltip to this component, replacing any existing one.
// Pass nil to clear.
func (c *Component) SetTooltip(tt *Tooltip) {
	if c.tooltip != nil && c.tooltip != tt {
		DefaultTooltipManager.onTriggerCleared(c)
	}
	c.tooltip = tt
}

// SetTooltipText is a convenience method that creates a single-label tooltip
// and attaches it to this component.
func (c *Component) SetTooltipText(text string, source *sg.FontFamily, size float64) {
	tt := NewTooltip(c.node.Name + "-tooltip")
	tt.SetText(text, source, size)
	c.SetTooltip(tt)
}

// ClearTooltip removes any attached tooltip from this component.
func (c *Component) ClearTooltip() {
	c.SetTooltip(nil)
}

// GetTooltip returns the tooltip attached to this component, or nil if none.
func (c *Component) GetTooltip() *Tooltip {
	return c.tooltip
}

// SetContextMenu attaches a ContextMenu to this component. Right-clicking
// the component will show the menu at the cursor position.
func (c *Component) SetContextMenu(cm *ContextMenu) {
	c.contextMenu = cm
}

// ClearContextMenu removes any attached context menu from this component.
func (c *Component) ClearContextMenu() {
	c.contextMenu = nil
}

// SetOnTooltipShow registers a callback that fires just before the tooltip
// becomes visible. Use this to update tooltip content dynamically.
func (c *Component) SetOnTooltipShow(fn func()) {
	c.onTooltipShow = fn
}

// SetOnTooltipHide registers a callback that fires just after the tooltip
// is hidden.
func (c *Component) SetOnTooltipHide(fn func()) {
	c.onTooltipHide = fn
}

// --- State queries ---

// IsHovered reports whether the pointer is over this component.
func (c *Component) IsHovered() bool {
	return c.hovered
}

// containsCursor reports whether the mouse cursor is within this component's
// bounding box. Unlike IsHovered, this works even when a child node captures
// the pointer enter/leave events (e.g. list items inside a scrollable list).
func (c *Component) containsCursor() bool {
	cx, cy := engine.CursorPosition()
	lx, ly := c.node.WorldToLocal(float64(cx), float64(cy))
	return lx >= 0 && lx < c.Width && ly >= 0 && ly < c.Height
}

// IsPressed reports whether the component is being pressed.
func (c *Component) IsPressed() bool {
	return c.pressed
}

// IsFocused reports whether the component currently has focus.
func (c *Component) IsFocused() bool {
	return c.focused
}

// SetFocused sets the focus state directly. Prefer using FocusManager.SetFocus
// for proper focus tracking across the UI.
func (c *Component) SetFocused(v bool) {
	if c.focused == v {
		return
	}
	c.focused = v
	c.MarkDrawDirty()
	if c.onFocusChange != nil {
		c.onFocusChange(v)
	}
	// When gaining focus, ask any ancestor ScrollPanel to scroll this
	// component into view.
	if v {
		c.ensureVisibleInScrollPanel()
	}
}

// ensureVisibleInScrollPanel walks up the parent chain looking for a
// ScrollPanel ancestor and asks it to scroll this component into view.
func (c *Component) ensureVisibleInScrollPanel() {
	for p := c.parent; p != nil; p = p.parent {
		if p.ensureChildVisible != nil {
			p.ensureChildVisible(c)
			return
		}
	}
}

// HandleKey delegates to the component's handleKeyFn. Returns false if no
// handler is set. This is a pure query — it must not perform the action
// itself; the widget's own Update() handles the actual behavior.
func (c *Component) HandleKey(key engine.Key) bool {
	if c.handleKeyFn == nil {
		return false
	}
	return c.handleKeyFn(key)
}

// SetHandleKey sets the function called by HandleKey. Widgets with
// InterceptArrows set this in their constructor.
func (c *Component) SetHandleKey(fn func(engine.Key) bool) {
	c.handleKeyFn = fn
}

// WorldBounds returns the component's bounding rectangle in world-space
// coordinates using the underlying node's transform.
func (c *Component) WorldBounds() (x, y, w, h float64) {
	wx, wy := c.node.LocalToWorld(0, 0)
	return wx, wy, c.Width, c.Height
}

// WorldCenter returns the component's center point in world-space coordinates.
func (c *Component) WorldCenter() (cx, cy float64) {
	wx, wy := c.node.LocalToWorld(0, 0)
	return wx + c.Width/2, wy + c.Height/2
}

// --- Children ---

// AddChild appends a child component. Both the Component tree and the
// underlying willow Node tree are updated. Accepts any UIElement (all
// widget types implement this interface).
func (c *Component) AddChild(child UIElement) {
	if child == nil {
		return
	}
	cc := child.base()
	// Remove from previous parent if any.
	if cc.parent != nil {
		cc.parent.RemoveChild(child)
	}
	cc.parent = c
	c.children = append(c.children, cc)
	c.node.AddChild(cc.node)
	c.MarkLayoutDirty()

	// If the child inherits its theme (no explicit override), it may now
	// have a different effective theme from its new parent. Propagate.
	if cc.theme == nil {
		cc.propagateThemeChange()
	}
}

// RemoveChild detaches a child component from this component.
func (c *Component) RemoveChild(child UIElement) {
	if child == nil {
		return
	}
	cc := child.base()
	if cc.parent != c {
		return
	}
	for i, ch := range c.children {
		if ch == cc {
			copy(c.children[i:], c.children[i+1:])
			c.children[len(c.children)-1] = nil
			c.children = c.children[:len(c.children)-1]
			break
		}
	}
	if c.anchoredChildren != nil {
		delete(c.anchoredChildren, cc)
	}
	cc.parent = nil
	c.node.RemoveChild(cc.node)
	c.MarkLayoutDirty()
}

// Parent returns the parent component, or nil if this is a root.
func (c *Component) Parent() *Component {
	return c.parent
}

// Children returns a read-only view of the child components.
func (c *Component) Children() []*Component {
	return c.children
}

// NumChildren returns the number of child components.
func (c *Component) NumChildren() int {
	return len(c.children)
}

// --- Theme & Variant ---

// SetTheme sets an explicit theme on this component. Pass nil to revert
// to parent inheritance. Propagates to descendants.
func (c *Component) SetTheme(t *Theme) {
	c.theme = t
	c.propagateThemeChange()
}

// EffectiveTheme returns the theme in effect for this component:
// its own if set, otherwise the nearest ancestor's, falling back to
// DefaultTheme if no ancestor has an explicit theme.
func (c *Component) EffectiveTheme() *Theme {
	if c.theme != nil {
		return c.theme
	}
	if c.parent != nil {
		return c.parent.EffectiveTheme()
	}
	return getDefaultTheme()
}

// SetVariant sets the visual variant for this component.
func (c *Component) SetVariant(v Variant) {
	if c.variant != v {
		c.variant = v
		if c.onThemeChange != nil {
			c.onThemeChange()
		} else {
			c.MarkDrawDirty()
		}
	}
}

// Variant returns the current visual variant.
func (c *Component) Variant() Variant {
	return c.variant
}

// propagateThemeChange notifies this component and all descendants that
// the effective theme may have changed.
func (c *Component) propagateThemeChange() {
	if c.onThemeChange != nil {
		c.onThemeChange()
	} else {
		c.MarkDrawDirty()
	}
	for _, child := range c.children {
		// Only propagate to children that don't have their own theme.
		if child.theme == nil {
			child.propagateThemeChange()
		}
	}
}

// --- Dispose ---

// Dispose removes this component from its parent, disposes the underlying
// willow node, and recursively disposes all children.
func (c *Component) Dispose() {
	c.enabledWatch.Stop()
	c.visibleWatch.Stop()
	// Clear callbacks before unregistering so that SetFocused(false) during
	// Unregister doesn't trigger UpdateVisuals on a partially-disposed widget.
	c.onFocusChange = nil
	c.onVisualStateChange = nil
	// Unregister from focus manager so tab order and keybinds are cleaned up.
	if c.Focusable {
		DefaultFocusManager.Unregister(c)
	}
	// Notify the tooltip manager so any active tooltip is hidden.
	DefaultTooltipManager.onTriggerDisposed(c)
	if c.parent != nil {
		c.parent.RemoveChild(c)
	}
	// Dispose children (iterate over a copy since Dispose modifies the slice).
	for len(c.children) > 0 {
		c.children[0].Dispose()
	}
	c.node.Dispose()
	c.parent = nil
	c.children = nil
}

// --- Background ---

// initBackground creates the solid-color background sprite and adds it to
// the root node. Call this in widget constructors instead of creating a
// background sprite manually.
func (c *Component) initBackground(name string) {
	c.bgNode = sg.NewSprite(name+"-bg", sg.TextureRegion{})
	c.node.AddChild(c.bgNode)
}

// hideBackground hides the background node. Used by composite widgets to
// suppress the inner component's own background.
func (c *Component) hideBackground() {
	if c.bgNode != nil {
		c.bgNode.SetVisible(false)
	}
	if c.bgPoly != nil {
		c.bgPoly.SetVisible(false)
	}
	if c.bgGradientMesh != nil {
		c.bgGradientMesh.SetVisible(false)
	}
}

// hideBorder hides all border nodes.
func (c *Component) hideBorder() {
	for _, b := range []*sg.Node{c.borderTop, c.borderRight, c.borderBot, c.borderLeft} {
		if b != nil {
			b.SetVisible(false)
		}
	}
	if c.borderPoly != nil {
		c.borderPoly.SetVisible(false)
	}
}

// hideFocusRing hides the focus ring.
func (c *Component) hideFocusRing() {
	if c.focusRing != nil {
		c.focusRing.SetVisible(false)
	}
	c.focusRingShown = false
}

// resolveCornerRadius returns the resolved corner radius value. A negative
// value means "full pill/circle" and is resolved to half the given dimension.
func resolveCornerRadius(cr, fallbackDim float64) float64 {
	if cr < 0 {
		return fallbackDim / 2
	}
	return cr
}

// wireVisualCallbacks sets onVisualStateChange, onThemeChange, and
// onFocusChange to all invoke the same function (typically UpdateVisuals).
func (c *Component) wireVisualCallbacks(fn func()) {
	c.onVisualStateChange = func() { fn() }
	c.onThemeChange = func() { fn() }
	c.onFocusChange = func(bool) { fn() }
}

// enableFocusNavigation sets the standard focus flags (Focusable, AllowTab,
// AllowSpatial) and registers with the DefaultFocusManager.
func (c *Component) enableFocusNavigation() {
	c.Focusable = true
	c.AllowTab = true
	c.AllowSpatial = true
	DefaultFocusManager.Register(c)
}

// applyCornerRadius sets the corner radius. Call before applyBackground/applyBorder.
func (c *Component) applyCornerRadius(r float64) {
	c.cornerRadius = r
	c.perCornerRadius = false
}

// applyCornerRadiiPerCorner sets independent radii for each corner (TL, TR, BR, BL).
// Call before applyBackground/applyBorder.
func (c *Component) applyCornerRadiiPerCorner(tl, tr, br, bl float64) {
	c.cornerRadii = [4]float64{tl, tr, br, bl}
	c.cornerRadius = tl // largest used as fallback for border mesh
	for _, r := range c.cornerRadii {
		if r > c.cornerRadius {
			c.cornerRadius = r
		}
	}
	c.perCornerRadius = true
}

// applyBackground updates the background visuals based on the given Background.
func (c *Component) applyBackground(bg Background) {
	// Fully transparent solid is equivalent to no background.
	if bg.Type == BgSolid && bg.Color.A() == 0 {
		bg.Type = BgNone
	}
	switch bg.Type {
	case BgSolid:
		if c.cornerRadius > 0 {
			// Rounded: use polygon mesh, hide flat sprite.
			c.bgNode.SetVisible(false)
			c.ensureBgPoly()
			var pts []sg.Vec2
			if c.perCornerRadius {
				r := c.cornerRadii
				pts = render.RoundedRectPointsPerCorner(c.Width, c.Height, r[0], r[1], r[2], r[3], defaultCornerSegments)
			} else {
				pts = render.RoundedRectPoints(c.Width, c.Height, c.cornerRadius, defaultCornerSegments)
			}
			sg.SetPolygonPoints(c.bgPoly, pts)
			c.bgPoly.SetColor(bg.Color)
			c.bgPoly.SetVisible(true)
		} else {
			// Sharp: use flat sprite, hide polygon.
			c.bgNode.SetColor(bg.Color)
			c.bgNode.SetVisible(true)
			if c.bgPoly != nil {
				c.bgPoly.SetVisible(false)
			}
		}
		if c.bgContainer != nil {
			c.bgContainer.SetVisible(false)
		}
		if c.bgGradientMesh != nil {
			c.bgGradientMesh.SetVisible(false)
		}
		if c.bgCenterFillMesh != nil {
			c.bgCenterFillMesh.SetVisible(false)
		}
		c.bgSlice = nil
	case BgNineSlice:
		c.bgNode.SetVisible(false)
		if c.bgPoly != nil {
			c.bgPoly.SetVisible(false)
		}
		if c.bgGradientMesh != nil {
			c.bgGradientMesh.SetVisible(false)
		}
		c.bgSlice = bg.Slice
		c.ensureNineSlice(bg.Slice)
		c.bgContainer.SetVisible(true)
		render.LayoutNineSlice(c.bgSliceNodes, bg.Slice, c.Width, c.Height)
		// If the nine-grid specifies a center fill gradient, replace the
		// center cell with a gradient mesh.
		if bg.Slice.CenterFill != nil {
			c.ensureCenterFillMesh(bg.Slice)
		} else if c.bgCenterFillMesh != nil {
			c.bgCenterFillMesh.SetVisible(false)
		}
	case BgGradient:
		// Hide all other bg types.
		c.bgNode.SetVisible(false)
		if c.bgPoly != nil {
			c.bgPoly.SetVisible(false)
		}
		if c.bgContainer != nil {
			c.bgContainer.SetVisible(false)
		}
		if c.bgCenterFillMesh != nil {
			c.bgCenterFillMesh.SetVisible(false)
		}
		c.bgSlice = nil
		// Create or update gradient mesh.
		c.ensureGradientMesh(bg)
		c.bgGradientMesh.SetVisible(true)
	case BgNone:
		c.bgNode.SetColor(sg.Color{})
		c.bgNode.SetVisible(false)
		if c.bgPoly != nil {
			c.bgPoly.SetVisible(false)
		}
		if c.bgContainer != nil {
			c.bgContainer.SetVisible(false)
		}
		if c.bgGradientMesh != nil {
			c.bgGradientMesh.SetVisible(false)
		}
		if c.bgCenterFillMesh != nil {
			c.bgCenterFillMesh.SetVisible(false)
		}
		c.bgSlice = nil
	}
}

// insertInfraNode inserts a lazily-created infrastructure node (bgPoly,
// bgContainer, borderPoly) into c.node before any content children. It
// scans from the front and inserts at the first position that isn't a
// known infrastructure node (bgNode, border sprites, etc.).
func (c *Component) insertInfraNode(child *sg.Node) {
	infra := []*sg.Node{
		c.bgNode, c.bgPoly, c.bgGradientMesh, c.bgContainer,
		c.borderTop, c.borderRight, c.borderBot, c.borderLeft, c.borderPoly,
		c.focusRing,
	}
	children := c.node.Children()
	idx := 0
	for idx < len(children) {
		match := false
		for _, n := range infra {
			if n != nil && children[idx] == n {
				match = true
				break
			}
		}
		if !match {
			break
		}
		idx++
	}
	c.node.AddChildAt(child, idx)
}

// ensureBgPoly lazily creates the polygon node for rounded backgrounds.
func (c *Component) ensureBgPoly() {
	if c.bgPoly != nil {
		return
	}
	// Create with a placeholder rectangle; real points set by caller.
	pts := render.RoundedRectPoints(1, 1, 0, defaultCornerSegments)
	c.bgPoly = sg.NewPolygon(c.node.Name+"-bg-poly", pts)
	c.bgPoly.SetVisible(false)
	c.insertInfraNode(c.bgPoly)
}

// ensureNineSlice lazily creates the nine-slice container and its 9 child
// sprites. If the NineSlice configuration changes (different image), the
// old sprites are replaced.
func (c *Component) ensureNineSlice(ns *NineSlice) {
	if c.bgContainer == nil {
		c.bgContainer = sg.NewContainer(c.node.Name + "-bg9")
		// Insert before content children so it stays behind them.
		c.insertInfraNode(c.bgContainer)
	}
	// Recreate sprites if the NineSlice changed or first use.
	if c.bgSliceNodes == nil || c.bgSlice != ns {
		// Remove old children if any.
		if c.bgSliceNodes != nil {
			c.bgContainer.RemoveChild(c.bgSliceNodes.TL)
			c.bgContainer.RemoveChild(c.bgSliceNodes.T)
			c.bgContainer.RemoveChild(c.bgSliceNodes.TR)
			c.bgContainer.RemoveChild(c.bgSliceNodes.L)
			c.bgContainer.RemoveChild(c.bgSliceNodes.C)
			c.bgContainer.RemoveChild(c.bgSliceNodes.R)
			c.bgContainer.RemoveChild(c.bgSliceNodes.BL)
			c.bgContainer.RemoveChild(c.bgSliceNodes.B)
			c.bgContainer.RemoveChild(c.bgSliceNodes.BR)
		}
		c.bgSliceNodes = render.CreateNineSliceNodes(c.node.Name+"-bg9", c.bgContainer, ns)
	}
}

// ensureCenterFillMesh lazily creates or updates a gradient mesh that
// replaces the center cell of a nine-grid background.
func (c *Component) ensureCenterFillMesh(ns *NineSlice) {
	inL := ns.Insets.Left
	inR := ns.Insets.Right
	inT := ns.Insets.Top
	inB := ns.Insets.Bottom
	midW := c.Width - inL - inR
	midH := c.Height - inT - inB
	if midW <= 0 || midH <= 0 {
		if c.bgCenterFillMesh != nil {
			c.bgCenterFillMesh.SetVisible(false)
		}
		return
	}

	verts, inds := render.RoundedRectGradientMesh(midW, midH, 0, defaultCornerSegments, ns.CenterFill)
	if c.bgCenterFillMesh == nil {
		c.bgCenterFillMesh = sg.NewMesh(c.node.Name+"-bg-cf", sg.WhitePixel, verts, inds)
		c.bgCenterFillMesh.SetColor(sg.RGBA(1, 1, 1, 1))
		c.bgCenterFillMesh.SetVisible(false)
		// Add inside the nine-grid container so it renders with the nine-grid.
		if c.bgContainer != nil {
			c.bgContainer.AddChild(c.bgCenterFillMesh)
		}
	} else {
		c.bgCenterFillMesh.SetMeshVertices(verts)
		c.bgCenterFillMesh.SetMeshIndices(inds)
		c.bgCenterFillMesh.InvalidateMeshAABB()
	}
	c.bgCenterFillMesh.SetPosition(inL, inT)
	c.bgCenterFillMesh.SetVisible(true)
}

// ensureGradientMesh lazily creates or updates the gradient mesh node.
func (c *Component) ensureGradientMesh(bg Background) {
	verts, inds := render.RoundedRectGradientMesh(c.Width, c.Height, c.cornerRadius, defaultCornerSegments, bg.Gradient)
	if c.bgGradientMesh == nil {
		c.bgGradientMesh = sg.NewMesh(c.node.Name+"-bg-grad", sg.WhitePixel, verts, inds)
		c.bgGradientMesh.SetColor(sg.RGBA(1, 1, 1, 1))
		c.bgGradientMesh.SetVisible(false)
		c.insertInfraNode(c.bgGradientMesh)
	} else {
		c.bgGradientMesh.SetMeshVertices(verts)
		c.bgGradientMesh.SetMeshIndices(inds)
		c.bgGradientMesh.InvalidateMeshAABB()
		c.bgGradientMesh.Invalidate()
	}
}

// resizeBackground scales the background node to the given dimensions.
// When a nine-slice is active, it also re-layouts the 9 sprites.
func (c *Component) resizeBackground(w, h float64) {
	if c.cornerRadius > 0 && c.bgPoly != nil && c.bgPoly.Visible() {
		var pts []sg.Vec2
		if c.perCornerRadius {
			r := c.cornerRadii
			pts = render.RoundedRectPointsPerCorner(w, h, r[0], r[1], r[2], r[3], defaultCornerSegments)
		} else {
			pts = render.RoundedRectPoints(w, h, c.cornerRadius, defaultCornerSegments)
		}
		sg.SetPolygonPoints(c.bgPoly, pts)
	}
	if c.bgNode != nil {
		c.bgNode.SetScale(w, h)
	}
	if c.bgSlice != nil && c.bgSliceNodes != nil {
		render.LayoutNineSlice(c.bgSliceNodes, c.bgSlice, w, h)
		// Resize center fill gradient if active.
		if c.bgSlice.CenterFill != nil {
			c.ensureCenterFillMesh(c.bgSlice)
		}
	}
	// Gradient mesh is regenerated via applyBackground when the theme/state
	// changes. No separate resize path needed because applyBackground
	// calls ensureGradientMesh which rebuilds with current dimensions.
}

// --- Border ---

// initBorder creates 4 WhitePixel sprites for the component border and adds
// them to the root node. All start invisible. Call in widget constructors.
func (c *Component) initBorder(name string) {
	c.borderTop = sg.NewSprite(name+"-border-t", sg.TextureRegion{})
	c.borderRight = sg.NewSprite(name+"-border-r", sg.TextureRegion{})
	c.borderBot = sg.NewSprite(name+"-border-b", sg.TextureRegion{})
	c.borderLeft = sg.NewSprite(name+"-border-l", sg.TextureRegion{})
	for _, b := range []*sg.Node{c.borderTop, c.borderRight, c.borderBot, c.borderLeft} {
		b.SetColor(sg.Color{})
		b.SetVisible(false)
		c.node.AddChild(b)
	}
}

// applyBorder sets the border color, width, and visibility. If the background
// is a nine-slice, borders are hidden (the nine-slice provides its own border).
func (c *Component) applyBorder(color sg.Color, width float64, bg Background) {
	c.borderWidth_ = width
	if bg.Type == BgNineSlice {
		// Nine-slice backgrounds provide their own border.
		for _, b := range []*sg.Node{c.borderTop, c.borderRight, c.borderBot, c.borderLeft} {
			b.SetVisible(false)
		}
		if c.borderPoly != nil {
			c.borderPoly.SetVisible(false)
		}
		return
	}
	vis := color.A() > 0 && width > 0
	if c.cornerRadius > 0 && vis {
		// Rounded: use border mesh, hide edge sprites.
		for _, b := range []*sg.Node{c.borderTop, c.borderRight, c.borderBot, c.borderLeft} {
			b.SetVisible(false)
		}
		c.ensureBorderPoly()
		verts, inds := render.RoundedRectBorderMesh(c.Width, c.Height, c.cornerRadius, width, defaultCornerSegments)
		c.borderPoly.SetMeshVertices(verts)
		c.borderPoly.SetMeshIndices(inds)
		c.borderPoly.InvalidateMeshAABB()
		c.borderPoly.SetColor(color)
		c.borderPoly.SetVisible(true)
	} else {
		// Sharp or invisible: use edge sprites, hide polygon.
		if c.borderPoly != nil {
			c.borderPoly.SetVisible(false)
		}
		for _, b := range []*sg.Node{c.borderTop, c.borderRight, c.borderBot, c.borderLeft} {
			b.SetColor(color)
			b.SetVisible(vis)
		}
		c.resizeBorder(c.Width, c.Height)
	}
}

// ensureBorderPoly lazily creates the mesh node for rounded borders.
func (c *Component) ensureBorderPoly() {
	if c.borderPoly != nil {
		return
	}
	// Create with a minimal mesh; real geometry set by caller.
	verts, inds := render.RoundedRectBorderMesh(1, 1, 0, 1, defaultCornerSegments)
	c.borderPoly = sg.NewMesh(c.node.Name+"-border-poly", sg.WhitePixel, verts, inds)
	c.borderPoly.SetVisible(false)
	c.insertInfraNode(c.borderPoly)
}

// resizeBorder positions the 4 border sprites along the edges of the
// given dimensions. Uses the same layout as Panel's old updateBorderLayout.
func (c *Component) resizeBorder(w, h float64) {
	// Rounded border: rebuild mesh.
	if c.cornerRadius > 0 && c.borderPoly != nil && c.borderPoly.Visible() {
		verts, inds := render.RoundedRectBorderMesh(w, h, c.cornerRadius, c.borderWidth_, defaultCornerSegments)
		c.borderPoly.SetMeshVertices(verts)
		c.borderPoly.SetMeshIndices(inds)
		c.borderPoly.InvalidateMeshAABB()
		return
	}

	if c.borderTop == nil {
		return
	}
	bw := c.borderWidth_

	// Top: full width, borderWidth tall, at top edge.
	c.borderTop.SetPosition(0, 0)
	c.borderTop.SetScale(w, bw)

	// Bottom: full width, borderWidth tall, at bottom edge.
	c.borderBot.SetPosition(0, h-bw)
	c.borderBot.SetScale(w, bw)

	// Left: borderWidth wide, full height minus top/bottom borders.
	c.borderLeft.SetPosition(0, bw)
	c.borderLeft.SetScale(bw, h-bw*2)

	// Right: borderWidth wide, full height minus top/bottom borders.
	c.borderRight.SetPosition(w-bw, bw)
	c.borderRight.SetScale(bw, h-bw*2)
}

// ---------------------------------------------------------------------------
// Focus ring
// ---------------------------------------------------------------------------

// defaultFocusRingOffset is how far outside the component bounds the focus ring
// is drawn (in pixels).
const defaultFocusRingOffset = 2.0

// applyFocusRing shows or hides the focus ring outline around this component.
// The ring is lazy-initialized on first focus. Color and width come from the
// widget's theme group. Call this from UpdateVisuals after applyBorder.
func (c *Component) applyFocusRing(color sg.Color, width float64) {
	c.applyFocusRingEx(color, width, c.Width, c.Height, c.cornerRadius)
}

// applyFocusRingSize is like applyFocusRing but uses the given w/h instead of
// the component's full dimensions. Use this when the interactive area is
// smaller than the full component (e.g. a checkbox box without its label).
func (c *Component) applyFocusRingSize(color sg.Color, width, w, h float64) {
	c.applyFocusRingEx(color, width, w, h, c.cornerRadius)
}

// applyFocusRingEx is the underlying implementation for applyFocusRing and
// applyFocusRingSize, accepting an explicit corner radius.
func (c *Component) applyFocusRingEx(color sg.Color, width, w, h, cornerRadius float64) {
	show := c.focused && color.A() > 0 && width > 0

	cr := cornerRadius
	if cr > 0 {
		cr += defaultFocusRingOffset
	}
	ringW := w + defaultFocusRingOffset*2
	ringH := h + defaultFocusRingOffset*2

	if show && c.focusRing == nil {
		// Lazy-init: create the mesh node on first focus.
		verts, inds := render.RoundedRectBorderMesh(ringW, ringH, cr, width, defaultCornerSegments)
		c.focusRing = sg.NewMesh(c.node.Name+"-focus-ring", sg.WhitePixel, verts, inds)
		c.focusRing.SetPosition(-defaultFocusRingOffset, -defaultFocusRingOffset)
		c.focusRing.SetColor(color)
		c.focusRing.Alpha_ = 0
		c.focusRing.SetVisible(false)
		c.insertInfraNode(c.focusRing)
	}

	if c.focusRing == nil {
		return
	}

	if show {
		// Update geometry and color in case size or theme changed.
		verts, inds := render.RoundedRectBorderMesh(ringW, ringH, cr, width, defaultCornerSegments)
		c.focusRing.SetMeshVertices(verts)
		c.focusRing.SetMeshIndices(inds)
		c.focusRing.InvalidateMeshAABB()
		c.focusRing.SetPosition(-defaultFocusRingOffset, -defaultFocusRingOffset)
		c.focusRing.SetColor(color)
	}

	if show && !c.focusRingShown {
		c.focusRingShown = true
		c.focusRing.SetVisible(true)
		if c.focusRingTween != nil {
			c.focusRingTween.Cancel()
		}
		c.focusRingTween = sg.TweenAlpha(c.focusRing, 1.0, sg.TweenConfig{Duration: 0.1})
	} else if !show && c.focusRingShown {
		c.focusRingShown = false
		// Tween to alpha 0; the node stays visible but fully transparent.
		if c.focusRingTween != nil {
			c.focusRingTween.Cancel()
		}
		c.focusRingTween = sg.TweenAlpha(c.focusRing, 0, sg.TweenConfig{Duration: 0.1})
	}
}

// --- Dirty flags ---

// MarkLayoutDirty marks this component as needing layout recalculation.
// Propagates up to the parent.
func (c *Component) MarkLayoutDirty() {
	c.dirtyLayout = true
	c.dirtyDraw = true
	if c.parent != nil {
		c.parent.MarkLayoutDirty()
	}
}

// MarkDrawDirty marks this component as needing a visual update.
func (c *Component) MarkDrawDirty() {
	c.dirtyDraw = true
}

// IsLayoutDirty reports whether the component needs layout recalculation.
func (c *Component) IsLayoutDirty() bool {
	return c.dirtyLayout
}

// IsDrawDirty reports whether the component needs a visual update.
func (c *Component) IsDrawDirty() bool {
	return c.dirtyDraw
}

// --- Layout ---

// UpdateLayout recalculates position and size of this component and its
// children, then syncs the results to the underlying willow nodes.
func (c *Component) UpdateLayout() {
	if !c.dirtyLayout {
		return
	}

	// Clamp dimensions to min/max constraints.
	c.Width = clampDim(c.Width, c.MinWidth, c.MaxWidth)
	c.Height = clampDim(c.Height, c.MinHeight, c.MaxHeight)

	// Sync this component's position to the node, including any user offset.
	c.node.SetPosition(c.X+c.OffsetX, c.Y+c.OffsetY)

	// Sync hit shape to component dimensions only if one was already established.
	// Components without an explicit hit shape are transparent to pointer events,
	// which prevents decorative panels and labels from silently absorbing clicks.
	if c.node.HitShape != nil && c.Width > 0 && c.Height > 0 {
		c.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: c.Width, Height: c.Height}
	}

	// Custom layout hook (e.g. AnchorLayout).
	if c.onLayout != nil {
		c.onLayout()
	}

	// Apply layout to children based on mode.
	switch c.Layout {
	case LayoutNone:
		c.layoutNone()
	case LayoutVBox:
		c.layoutVBox()
	case LayoutHBox:
		c.layoutHBox()
	case LayoutGrid:
		c.layoutGrid()
	case LayoutFlow:
		c.layoutFlow()
	case LayoutAnchor:
		c.layoutAnchor()
	}

	// Recursively update children.
	for _, child := range c.children {
		child.UpdateLayout()
	}

	c.dirtyLayout = false
}

// SizeToContent resizes the component to tightly fit its children based on
// the current layout mode (VBox or HBox). Children should already have their
// sizes set. After resizing, UpdateLayout is called to reposition children.
func (c *Component) SizeToContent() {
	switch c.Layout {
	case LayoutVBox:
		h := c.Padding.Vertical()
		w := 0.0
		for i, ch := range c.Children() {
			if i > 0 {
				h += c.Spacing
			}
			h += ch.Margin.Vertical() + ch.Height
			cw := ch.Margin.Horizontal() + ch.Width
			if cw > w {
				w = cw
			}
		}
		w += c.Padding.Horizontal()
		c.Width, c.Height = w, h
	case LayoutHBox:
		w := c.Padding.Horizontal()
		h := 0.0
		for i, ch := range c.Children() {
			if i > 0 {
				w += c.Spacing
			}
			w += ch.Margin.Horizontal() + ch.Width
			ch2 := ch.Margin.Vertical() + ch.Height
			if ch2 > h {
				h = ch2
			}
		}
		h += c.Padding.Vertical()
		c.Width, c.Height = w, h
	case LayoutFlow:
		rowGap := c.flowRowGap()
		availW := c.Width - c.Padding.Horizontal()
		if availW <= 0 {
			// Width is zero: treat all items as a single row to derive natural dimensions.
			rows := c.buildFlowRows(math.MaxFloat64)
			if len(rows) > 0 {
				c.Width = rows[0].width + c.Padding.Horizontal()
				c.Height = rows[0].height + c.Padding.Vertical()
			} else {
				c.Width = c.Padding.Horizontal()
				c.Height = c.Padding.Vertical()
			}
		} else {
			// Fixed width: compute required height for wrapping layout.
			rows := c.buildFlowRows(availW)
			totalH := c.Padding.Vertical()
			for i, row := range rows {
				if i > 0 {
					totalH += rowGap
				}
				totalH += row.height
			}
			c.Height = totalH
		}
	}
	c.UpdateLayout()
}

// clampDim clamps v between min and max. A zero max means no upper bound.
func clampDim(v, min, max float64) float64 {
	return core.ClampDim(v, min, max)
}

// layoutNone syncs each child's node position from the child's X/Y.
func (c *Component) layoutNone() {
	c.applyFill()
	for _, child := range c.children {
		if !child.IsVisible() {
			continue
		}
		child.node.SetPosition(child.X+child.OffsetX, child.Y+child.OffsetY)
	}
}

// applyFill sets the dimensions of children that have a Fill flag, based on
// the parent's content area.
func (c *Component) applyFill() {
	availX := c.Width - c.Padding.Left - c.Padding.Right
	availY := c.Height - c.Padding.Top - c.Padding.Bottom
	for _, child := range c.children {
		if !child.IsVisible() || child.Fill == FillNone {
			continue
		}
		if child.Fill&FillWidth != 0 {
			child.Width = availX - child.Margin.Horizontal()
		}
		if child.Fill&FillHeight != 0 {
			child.Height = availY - child.Margin.Vertical()
		}
	}
}

// layoutVBox stacks visible children vertically, starting at Padding.Top.
func (c *Component) layoutVBox() {
	availX := c.Width - c.Padding.Left - c.Padding.Right
	availY := c.Height - c.Padding.Top - c.Padding.Bottom

	c.applyFill()

	// Distribute remaining space to grow children.
	var totalGrow int
	var fixedH float64
	var visCount int
	first := true
	for _, child := range c.children {
		if !child.IsVisible() {
			continue
		}
		visCount++
		if child.Grow > 0 {
			totalGrow += child.Grow
		} else {
			fixedH += child.Margin.Top + child.Height + child.Margin.Bottom
		}
		if !first {
			fixedH += c.Spacing
		}
		first = false
	}
	if totalGrow > 0 {
		// Remaining space after fixed children and all spacing gaps.
		remaining := availY - fixedH
		if remaining < 0 {
			remaining = 0
		}
		for _, child := range c.children {
			if !child.IsVisible() || child.Grow <= 0 {
				continue
			}
			child.Height = remaining * float64(child.Grow) / float64(totalGrow)
		}
	}

	// Compute total content height for main-axis justify.
	var totalH float64
	first = true
	for _, child := range c.children {
		if !child.IsVisible() {
			continue
		}
		totalH += child.Margin.Top + child.Height + child.Margin.Bottom
		if !first {
			totalH += c.Spacing
		}
		first = false
	}

	y := c.Padding.Top
	spaceBetweenGap := c.Spacing
	switch c.Justify {
	case AlignCenter:
		y += (availY - totalH) / 2
	case AlignEnd:
		y += availY - totalH
	case AlignSpaceBetween:
		if visCount > 1 {
			spaceBetweenGap = (availY - (totalH - c.Spacing*float64(visCount-1))) / float64(visCount-1)
		}
	}

	for _, child := range c.children {
		if !child.IsVisible() {
			continue
		}
		switch c.Align {
		case AlignCenter:
			child.X = c.Padding.Left + (availX-child.Width)/2
		case AlignEnd:
			child.X = c.Padding.Left + availX - child.Width - child.Margin.Right
		default:
			child.X = c.Padding.Left + child.Margin.Left
		}
		child.Y = y + child.Margin.Top
		child.node.SetPosition(child.X+child.OffsetX, child.Y+child.OffsetY)
		y = child.Y + child.Height + child.Margin.Bottom + spaceBetweenGap
	}
}

// layoutHBox stacks visible children horizontally, starting at Padding.Left.
func (c *Component) layoutHBox() {
	availX := c.Width - c.Padding.Left - c.Padding.Right
	availY := c.Height - c.Padding.Top - c.Padding.Bottom

	c.applyFill()

	// Distribute remaining space to grow children.
	var totalGrow int
	var fixedW float64
	var visCount int
	first := true
	for _, child := range c.children {
		if !child.IsVisible() {
			continue
		}
		visCount++
		if child.Grow > 0 {
			totalGrow += child.Grow
		} else {
			fixedW += child.Margin.Left + child.Width + child.Margin.Right
		}
		if !first {
			fixedW += c.Spacing
		}
		first = false
	}
	if totalGrow > 0 {
		remaining := availX - fixedW
		if remaining < 0 {
			remaining = 0
		}
		for _, child := range c.children {
			if !child.IsVisible() || child.Grow <= 0 {
				continue
			}
			child.Width = remaining * float64(child.Grow) / float64(totalGrow)
		}
	}

	// Compute total content width for main-axis justify.
	var totalW float64
	first = true
	for _, child := range c.children {
		if !child.IsVisible() {
			continue
		}
		totalW += child.Margin.Left + child.Width + child.Margin.Right
		if !first {
			totalW += c.Spacing
		}
		first = false
	}

	x := c.Padding.Left
	spaceBetweenGap := c.Spacing
	switch c.Justify {
	case AlignCenter:
		x += (availX - totalW) / 2
	case AlignEnd:
		x += availX - totalW
	case AlignSpaceBetween:
		if visCount > 1 {
			spaceBetweenGap = (availX - (totalW - c.Spacing*float64(visCount-1))) / float64(visCount-1)
		}
	}

	for _, child := range c.children {
		if !child.IsVisible() {
			continue
		}
		child.X = x + child.Margin.Left
		switch c.Align {
		case AlignCenter:
			child.Y = c.Padding.Top + (availY-child.Height)/2
		case AlignEnd:
			child.Y = c.Padding.Top + availY - child.Height - child.Margin.Bottom
		default:
			child.Y = c.Padding.Top + child.Margin.Top
		}
		child.node.SetPosition(child.X+child.OffsetX, child.Y+child.OffsetY)
		x = child.X + child.Width + child.Margin.Right + spaceBetweenGap
	}
}

// flowRow holds the children, total content width, max height, and per-item
// outer heights for one row in a LayoutFlow pass.
type flowRow struct {
	items  []*Component
	outerH []float64 // outer height (margin+height) for each item, parallel to items
	width  float64
	height float64
}

// buildFlowRows groups the visible children of c into rows that fit within
// availW, using c.Spacing as the horizontal gap between items.
func (c *Component) buildFlowRows(availW float64) []flowRow {
	var rows []flowRow
	var cur flowRow

	for _, child := range c.children {
		if !child.IsVisible() {
			continue
		}
		ow := child.Margin.Horizontal() + child.Width
		oh := child.Margin.Vertical() + child.Height

		// Wrap if this is not the first item in the row and it would exceed width.
		if len(cur.items) > 0 && cur.width+c.Spacing+ow > availW {
			rows = append(rows, cur)
			cur = flowRow{}
		}

		if len(cur.items) > 0 {
			cur.width += c.Spacing
		}
		cur.width += ow
		if oh > cur.height {
			cur.height = oh
		}
		cur.items = append(cur.items, child)
		cur.outerH = append(cur.outerH, oh)
	}
	if len(cur.items) > 0 {
		rows = append(rows, cur)
	}
	return rows
}

// flowRowGap returns the effective vertical gap between rows: FlowRowGap when
// non-zero, otherwise Spacing.
func (c *Component) flowRowGap() float64 {
	if c.FlowRowGap != 0 {
		return c.FlowRowGap
	}
	return c.Spacing
}

// layoutFlow arranges visible children left-to-right, wrapping to new rows
// when the available width is exceeded.
func (c *Component) layoutFlow() {
	c.applyFill()
	availW := c.Width - c.Padding.Horizontal()
	rowGap := c.flowRowGap()

	rows := c.buildFlowRows(availW)

	y := c.Padding.Top
	for i, row := range rows {
		if i > 0 {
			y += rowGap
		}

		// Compute the row's starting X based on Justify.
		rowX := c.Padding.Left
		switch c.Justify {
		case AlignCenter:
			rowX += (availW - row.width) / 2
		case AlignEnd:
			rowX += availW - row.width
		}

		x := rowX
		for j, child := range row.items {
			oh := row.outerH[j]
			child.X = x + child.Margin.Left
			switch c.Align {
			case AlignCenter:
				child.Y = y + (row.height-oh)/2 + child.Margin.Top
			case AlignEnd:
				child.Y = y + row.height - oh + child.Margin.Top
			default:
				child.Y = y + child.Margin.Top
			}
			child.node.SetPosition(child.X+child.OffsetX, child.Y+child.OffsetY)
			x += child.Margin.Horizontal() + child.Width + c.Spacing
		}

		y += row.height
	}
}

// --- Anchor layout ---

// anchorEntry stores the anchor position and pixel offsets for a child in
// LayoutAnchor mode.
type anchorEntry struct {
	Anchor  Anchor
	OffsetX float64
	OffsetY float64
}

// anchorPosition computes the x, y position of a child within the available
// area based on the anchor point and pixel offsets.
func anchorPosition(parentW, parentH, childW, childH float64, a Anchor, ox, oy float64) (float64, float64) {
	var x, y float64
	switch a {
	case AnchorTopLeft:
		x, y = 0, 0
	case AnchorTopCenter:
		x, y = (parentW-childW)/2, 0
	case AnchorTopRight:
		x, y = parentW-childW, 0
	case AnchorMiddleLeft:
		x, y = 0, (parentH-childH)/2
	case AnchorCenter:
		x, y = (parentW-childW)/2, (parentH-childH)/2
	case AnchorMiddleRight:
		x, y = parentW-childW, (parentH-childH)/2
	case AnchorBottomLeft:
		x, y = 0, parentH-childH
	case AnchorBottomCenter:
		x, y = (parentW-childW)/2, parentH-childH
	case AnchorBottomRight:
		x, y = parentW-childW, parentH-childH
	}
	return x + ox, y + oy
}

// layoutAnchor positions each visible child according to its anchor entry.
func (c *Component) layoutAnchor() {
	padL := c.Padding.Left
	padT := c.Padding.Top
	availW := c.Width - padL - c.Padding.Right
	availH := c.Height - padT - c.Padding.Bottom

	for _, child := range c.children {
		if !child.IsVisible() {
			continue
		}
		var entry anchorEntry
		if c.anchoredChildren != nil {
			entry = c.anchoredChildren[child]
		}
		x, y := anchorPosition(availW, availH, child.Width, child.Height, entry.Anchor, entry.OffsetX, entry.OffsetY)
		child.X = padL + x
		child.Y = padT + y
	}
}

// AddAnchoredChild adds a child pinned to the given anchor with pixel offsets.
// The parent's Layout must be LayoutAnchor for anchor positioning to take effect.
func (c *Component) AddAnchoredChild(child UIElement, anchor Anchor, offsetX, offsetY float64) {
	if child == nil {
		return
	}
	c.AddChild(child)
	if c.anchoredChildren == nil {
		c.anchoredChildren = make(map[*Component]anchorEntry)
	}
	c.anchoredChildren[child.base()] = anchorEntry{
		Anchor:  anchor,
		OffsetX: offsetX,
		OffsetY: offsetY,
	}
}

// SetAnchor changes the anchor and offset for an existing child.
// Has no effect if child is not a direct child of this component.
func (c *Component) SetAnchor(child UIElement, anchor Anchor, offsetX, offsetY float64) {
	if child == nil {
		return
	}
	cc := child.base()
	if cc.parent != c {
		return
	}
	if c.anchoredChildren == nil {
		c.anchoredChildren = make(map[*Component]anchorEntry)
	}
	c.anchoredChildren[cc] = anchorEntry{
		Anchor:  anchor,
		OffsetX: offsetX,
		OffsetY: offsetY,
	}
	c.MarkLayoutDirty()
}

// AnchorOf returns the anchor metadata for a child. Returns false if the
// child is not a direct child or has no explicit anchor data (defaults apply).
func (c *Component) AnchorOf(child UIElement) (anchor Anchor, offsetX, offsetY float64, ok bool) {
	if child == nil || c.anchoredChildren == nil {
		return AnchorTopLeft, 0, 0, false
	}
	cc := child.base()
	if cc.parent != c {
		return AnchorTopLeft, 0, 0, false
	}
	entry, exists := c.anchoredChildren[cc]
	if !exists {
		return AnchorTopLeft, 0, 0, false
	}
	return entry.Anchor, entry.OffsetX, entry.OffsetY, true
}

// layoutGrid arranges visible children in a grid with GridColumns columns.
func (c *Component) layoutGrid() {
	c.applyFill()
	cols := c.GridColumns
	if cols < 1 {
		cols = 1
	}

	// Determine uniform cell size from the largest child.
	var cellW, cellH float64
	for _, child := range c.children {
		if !child.IsVisible() {
			continue
		}
		w := child.Width + child.Margin.Horizontal()
		h := child.Height + child.Margin.Vertical()
		cellW = math.Max(cellW, w)
		cellH = math.Max(cellH, h)
	}

	col, row := 0, 0
	for _, child := range c.children {
		if !child.IsVisible() {
			continue
		}
		child.X = c.Padding.Left + float64(col)*(cellW+c.Spacing) + child.Margin.Left
		child.Y = c.Padding.Top + float64(row)*(cellH+c.Spacing) + child.Margin.Top
		child.node.SetPosition(child.X+child.OffsetX, child.Y+child.OffsetY)
		col++
		if col >= cols {
			col = 0
			row++
		}
	}
}
