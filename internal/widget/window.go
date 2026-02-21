package widget

import (
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// Window is a draggable, closeable, optionally resizable container with a
// title bar and body panel.
type Window struct {
	Component
	titleBar            *sg.Node        // container for the title bar area
	titleBg             *sg.Node        // polygon: title bar background (rounded top corners)
	titleBgFlat         *sg.Node        // flat sprite fallback when cornerRadius == 0
	titleGradientMesh   *sg.Node        // gradient mesh for title bar background
	titleSliceContainer *sg.Node        // nine-grid container for title bar (lazy)
	titleSliceNodes     *nineSliceNodes // nine-grid sprites (lazy)
	titleSlice          *NineSlice      // current nine-slice config
	titleCenterFillMesh *sg.Node        // gradient mesh replacing nine-grid center cell (lazy)
	titleLabel          *Label
	closeBtn            *IconButton
	body                *Panel

	contentPaneUnderTitleBar bool // body background extends to Y=0, behind the title bar

	dragging bool
	dragOffX float64
	dragOffY float64

	movable bool // whether the window can be dragged by the title bar

	resizable    bool
	resizeHandle *sg.Node // polygon: bottom-right resize grip (rounded corner)
	resizeFlat   *sg.Node // flat sprite fallback when cornerRadius == 0
	resizeIcon   *sg.Node // sprite node for theme resize icon
	resizing     bool
	resizeStartX float64 // global X at drag start
	resizeStartY float64 // global Y at drag start
	resizeStartW float64 // window width at drag start
	resizeStartH float64 // window height at drag start

	modal               bool     // whether the window is modal (blocks input below)
	modalOverlay        *sg.Node // full-screen interactable overlay rendered behind the window
	modalOverlayColor   sg.Color // color of the modal overlay (default: semi-transparent black)
	modalOverlayOnClick func()   // called when the user clicks the overlay (nil = no-op)

	minWidth          float64
	minHeight         float64
	titleBarHeight    float64
	onClose           func()
	onResult          func(string, any) // called by FireResult; nil = no handler
	escResultKey      string            // non-empty: Esc fires FireResult(key, nil)
	enterResultKey    string            // non-empty: Enter fires FireResult(key, nil)
	closeIcon         engine.Image      // nil = use theme or procedural default
	appliedCloseIcon  engine.Image      // tracks last applied theme close icon
	appliedResizeIcon engine.Image      // tracks last applied theme resize icon
}

// modalOverlayHalfSize is the half-extent used to create the modal overlay.
// The overlay is centered at the window's scene origin and sized 2× this in
// each dimension, covering any practical viewport size.
const modalOverlayHalfSize = 10000.0

// Default window dimensions.
const (
	DefaultWindowWidth   = 400
	DefaultWindowHeight  = 300
	DefaultWindowMinW    = 120
	DefaultWindowMinH    = 80
	WindowTitleBarHeight = 32
	WindowResizeGripSize = 16
	DefaultWindowPadding = 8
)

// defaultCloseIcon returns the close X glyph from the default spritesheet.
func defaultCloseIcon() engine.Image { return IconCloseX() }

// NewWindow creates a window with the given title, font source, and display size.
// The window is draggable by the title bar and includes a close button.
func NewWindow(name string, title string, source *sg.FontFamily, displaySize float64) *Window {
	w := &Window{
		minWidth:          DefaultWindowMinW,
		minHeight:         DefaultWindowMinH,
		titleBarHeight:    WindowTitleBarHeight,
		movable:           true,
		modalOverlayColor: sg.RGBA(0, 0, 0, 0.5),
	}
	initComponent(&w.Component, name)

	w.initBackground(name)
	w.initBorder(name)

	// Title bar container.
	w.titleBar = sg.NewContainer(name + "-titlebar")
	w.titleBar.Interactable = true
	w.node.AddChild(w.titleBar)

	// Title bar background: flat sprite (fallback) and polygon (rounded).
	w.titleBgFlat = sg.NewSprite(name+"-titlebg-flat", sg.TextureRegion{})
	w.titleBar.AddChild(w.titleBgFlat)
	placeholder := render.RoundedRectPointsPerCorner(1, 1, 0, 0, 0, 0, defaultCornerSegments)
	w.titleBg = sg.NewPolygon(name+"-titlebg", placeholder)
	w.titleBg.SetVisible(false)
	w.titleBar.AddChild(w.titleBg)

	// Title label.
	w.titleLabel = NewLabel(name+"-title", title, source, displaySize)
	w.titleBar.AddChild(w.titleLabel.Node())

	// Close button (icon-based; icon region set via SetCloseIcon).
	w.closeBtn = NewIconButton(name + "-close")
	w.closeBtn.SetVariant(Custom1)
	w.closeBtn.SetSize(w.titleBarHeight-12, w.titleBarHeight-12)
	w.closeBtn.SetOnClick(func() {
		w.Close()
	})
	// Tint the icon: untinted (white) in default state, theme TextColor on hover/active.
	white := sg.RGBA(1, 1, 1, 1)
	origVisualChange := w.closeBtn.onVisualStateChange
	w.closeBtn.onVisualStateChange = func() {
		if origVisualChange != nil {
			origVisualChange()
		}
		// The close button position is managed by the title bar layout,
		// so undo any theme-driven offsets applied by UpdateVisuals.
		if w.closeBtn.OffsetX != 0 || w.closeBtn.OffsetY != 0 {
			w.closeBtn.OffsetX = 0
			w.closeBtn.OffsetY = 0
			w.updateTitleLayout()
		}
		st := computeState(w.closeBtn.enabled, w.closeBtn.focused, w.closeBtn.hovered, w.closeBtn.pressed)
		if st == StateDefault {
			w.closeBtn.icon.SetColor(white)
		} else {
			group := w.closeBtn.EffectiveTheme().Button.Group(w.closeBtn.Variant())
			w.closeBtn.icon.SetColor(group.TextColor.Resolve(st))
		}
	}
	w.closeBtn.onThemeChange = w.closeBtn.onVisualStateChange
	w.titleBar.AddChild(w.closeBtn.Node())

	// Apply a default close icon so windows show a visible X out of the box.
	// Callers can override with SetCloseIcon at any time.
	w.SetCloseIcon(defaultCloseIcon())

	// Body panel — added to both the Component tree (for activation
	// bubbling) and the willow node tree (for rendering).
	w.body = NewPanel(name + "-body")
	w.body.SetLayout(LayoutVBox)
	w.body.parent = &w.Component
	w.node.AddChild(w.body.Node())

	// Resize handle: flat sprite (fallback) and polygon (rounded).
	w.resizeFlat = sg.NewSprite(name+"-resize-flat", sg.TextureRegion{})
	w.resizeFlat.Interactable = true
	w.resizeFlat.SetVisible(false)
	w.node.AddChild(w.resizeFlat)
	w.resizeHandle = sg.NewPolygon(name+"-resize", placeholder)
	w.resizeHandle.Interactable = true
	w.resizeHandle.SetVisible(false)
	w.node.AddChild(w.resizeHandle)
	w.resizeIcon = sg.NewSprite(name+"-resize-icon", sg.TextureRegion{})
	w.resizeIcon.Interactable = true
	w.resizeIcon.SetVisible(false)
	w.node.AddChild(w.resizeIcon)

	// Title bar pointer down activates the window via bubbleActivation.
	w.titleBar.OnPointerDown(func(_ sg.PointerContext) {
		w.bubbleActivation()
	})

	// Wire drag on title bar (only when movable).
	w.titleBar.OnDragStart(func(ctx sg.DragContext) {
		if !w.movable {
			return
		}
		w.dragging = true
		w.dragOffX = ctx.LocalX
		w.dragOffY = ctx.LocalY
	})
	w.titleBar.OnDrag(func(ctx sg.DragContext) {
		if !w.dragging {
			return
		}
		// Move the window node's position, keeping c.X/c.Y in sync
		// so that a later UpdateLayout pass doesn't snap back.
		newX := ctx.GlobalX - w.dragOffX
		newY := ctx.GlobalY - w.dragOffY
		w.X = newX
		w.Y = newY
		w.node.SetPosition(newX+w.OffsetX, newY+w.OffsetY)
	})
	w.titleBar.OnDragEnd(func(_ sg.DragContext) {
		w.dragging = false
	})

	// Wire drag on both resize handle nodes (polygon and flat).
	resizeDragStart := func(ctx sg.DragContext) {
		if !w.resizable {
			return
		}
		w.bubbleActivation()
		w.resizing = true
		w.resizeStartX = ctx.GlobalX
		w.resizeStartY = ctx.GlobalY
		w.resizeStartW = w.Width
		w.resizeStartH = w.Height
	}
	resizeDrag := func(ctx sg.DragContext) {
		if !w.resizing {
			return
		}
		newW := w.resizeStartW + (ctx.GlobalX - w.resizeStartX)
		newH := w.resizeStartH + (ctx.GlobalY - w.resizeStartY)
		if newW < w.minWidth {
			newW = w.minWidth
		}
		if newH < w.minHeight {
			newH = w.minHeight
		}
		w.SetSize(newW, newH)
	}
	resizeDragEnd := func(_ sg.DragContext) {
		w.resizing = false
	}
	w.resizeHandle.OnDragStart(resizeDragStart)
	w.resizeHandle.OnDrag(resizeDrag)
	w.resizeHandle.OnDragEnd(resizeDragEnd)
	w.resizeFlat.OnDragStart(resizeDragStart)
	w.resizeFlat.OnDrag(resizeDrag)
	w.resizeFlat.OnDragEnd(resizeDragEnd)
	w.resizeIcon.OnDragStart(resizeDragStart)
	w.resizeIcon.OnDrag(resizeDrag)
	w.resizeIcon.OnDragEnd(resizeDragEnd)

	// Show diagonal resize cursor over the resize handle.
	resizeEnter := func(_ sg.PointerContext) {
		engine.SetCursorShape(engine.CursorShapeNWSEResize)
	}
	resizeLeave := func(_ sg.PointerContext) {
		engine.SetCursorShape(engine.CursorShapeDefault)
	}
	w.resizeHandle.OnPointerEnter(resizeEnter)
	w.resizeHandle.OnPointerLeave(resizeLeave)
	w.resizeFlat.OnPointerEnter(resizeEnter)
	w.resizeFlat.OnPointerLeave(resizeLeave)
	w.resizeIcon.OnPointerEnter(resizeEnter)
	w.resizeIcon.OnPointerLeave(resizeLeave)

	w.onThemeChange = func() { w.applyThemeColors() }
	w.applyThemeColors()

	// Default size.
	w.SetSize(DefaultWindowWidth, DefaultWindowHeight)

	// Per-frame key handling for Esc/Enter result shortcuts.
	w.node.OnUpdate = func(_ float64) {
		if !w.node.Visible() {
			return
		}
		im := DefaultInputManager
		if w.escResultKey != "" && im.IsKeyJustAvailable(engine.KeyEscape) {
			im.Consume(engine.KeyEscape)
			w.FireResult(w.escResultKey, nil)
			return
		}
		if w.enterResultKey != "" && im.IsKeyJustAvailable(engine.KeyEnter) {
			im.Consume(engine.KeyEnter)
			w.FireResult(w.enterResultKey, nil)
		}
	}

	return w
}

func (w *Window) applyThemeColors() {
	group := w.EffectiveTheme().Window.Group(w.Variant())
	w.applyCornerRadius(group.CornerRadius)
	bg := group.Background.Resolve(StateDefault)
	w.applyBackground(bg)
	w.applyBorder(group.Border.Resolve(StateDefault), group.BorderWidth, bg)
	w.titleLabel.SetColor(group.TitleTextColor.Resolve(StateDefault))

	// Update content-pane-under-title-bar mode: body renders behind title bar
	// but in front of the window background. ZIndex ordering:
	//   normal:  all 0 → insertion order (bg → titleBar → body → resize)
	//   enabled: bg=0, body=1, titleBar=2, resize=3
	if w.contentPaneUnderTitleBar != group.ContentPaneUnderTitleBar {
		w.contentPaneUnderTitleBar = group.ContentPaneUnderTitleBar
		if w.contentPaneUnderTitleBar {
			w.body.SetZIndex(1)
			w.titleBar.SetZIndex(2)
			w.resizeFlat.SetZIndex(3)
			w.resizeHandle.SetZIndex(3)
			w.resizeIcon.SetZIndex(3)
		} else {
			w.body.SetZIndex(0)
			w.titleBar.SetZIndex(0)
			w.resizeFlat.SetZIndex(0)
			w.resizeHandle.SetZIndex(0)
			w.resizeIcon.SetZIndex(0)
		}
	}

	// Apply theme close icon (per-instance override takes priority).
	if w.closeIcon == nil && group.CloseIcon.Set && w.appliedCloseIcon != group.CloseIcon.Image {
		w.appliedCloseIcon = group.CloseIcon.Image
		w.applyCloseImage(group.CloseIcon.Image)
	}

	// Propagate theme to child components so they pick up the
	// window's effective theme (e.g. after SetTheme is called).
	if w.body != nil {
		w.body.propagateThemeChange()
	}
	if w.closeBtn != nil {
		w.closeBtn.SetTheme(w.EffectiveTheme())
	}

	// Determine title bar background.
	var titleBg Background
	if DefaultWindowManager.Active() == w {
		titleBg = group.TitleBackground.Resolve(StateFocus)
	} else {
		titleBg = group.TitleBackground.Resolve(StateDefault)
	}

	cr := group.CornerRadius
	w.applyTitleBackground(titleBg, cr)

	if group.ResizeIcon.Set {
		w.resizeIcon.SetColor(group.ResizeHandleColor.Resolve(StateDefault))
	} else if cr > 0 {
		w.resizeHandle.SetColor(group.ResizeHandleColor.Resolve(StateDefault))
		w.updateResizePoly(cr)
	} else {
		w.resizeFlat.SetColor(group.ResizeHandleColor.Resolve(StateDefault))
	}

	// Body panel needs per-corner rounding: sharp top, rounded bottom.
	w.applyBodyCorners()

	w.MarkDrawDirty()
}

// applyTitleBackground sets the title bar background, handling solid colors,
// gradients, and nine-grids (with optional gradient center fill).
func (w *Window) applyTitleBackground(bg Background, cr float64) {
	tbH := w.titleBarHeight

	// Hide all title bg renderers first; the active case re-shows its nodes.
	w.titleBg.SetVisible(false)
	w.titleBgFlat.SetVisible(false)
	if w.titleGradientMesh != nil {
		w.titleGradientMesh.SetVisible(false)
	}
	if w.titleSliceContainer != nil {
		w.titleSliceContainer.SetVisible(false)
	}
	if w.titleCenterFillMesh != nil {
		w.titleCenterFillMesh.SetVisible(false)
	}

	switch bg.Type {
	case BgNineSlice:
		// Nine-grid title bar (with optional gradient center fill).
		w.titleSlice = bg.Slice
		if w.titleSliceContainer == nil {
			w.titleSliceContainer = sg.NewContainer(w.node.Name + "-title-9")
			w.titleBar.AddChildAt(w.titleSliceContainer, 0)
		}
		if w.titleSliceNodes == nil || w.titleSlice != bg.Slice {
			if w.titleSliceNodes != nil {
				w.titleSliceContainer.RemoveChild(w.titleSliceNodes.TL)
				w.titleSliceContainer.RemoveChild(w.titleSliceNodes.T)
				w.titleSliceContainer.RemoveChild(w.titleSliceNodes.TR)
				w.titleSliceContainer.RemoveChild(w.titleSliceNodes.L)
				w.titleSliceContainer.RemoveChild(w.titleSliceNodes.C)
				w.titleSliceContainer.RemoveChild(w.titleSliceNodes.R)
				w.titleSliceContainer.RemoveChild(w.titleSliceNodes.BL)
				w.titleSliceContainer.RemoveChild(w.titleSliceNodes.B)
				w.titleSliceContainer.RemoveChild(w.titleSliceNodes.BR)
			}
			w.titleSliceNodes = render.CreateNineSliceNodes(w.node.Name+"-title-9", w.titleSliceContainer, bg.Slice)
		}
		render.LayoutNineSlice(w.titleSliceNodes, bg.Slice, w.Width, tbH)
		w.titleSliceContainer.SetVisible(true)

		// Center fill gradient — rendered at full title bar size *under* the
		// nine-grid so the border tiles draw on top with no gap.
		if bg.Slice.CenterFill != nil {
			verts, inds := render.RoundedRectGradientMesh(w.Width, tbH, 0, defaultCornerSegments, bg.Slice.CenterFill)
			if w.titleCenterFillMesh == nil {
				w.titleCenterFillMesh = sg.NewMesh(w.node.Name+"-title-cf", sg.WhitePixel, verts, inds)
				w.titleCenterFillMesh.SetColor(sg.RGBA(1, 1, 1, 1))
				w.titleSliceContainer.AddChildAt(w.titleCenterFillMesh, 0)
			} else {
				w.titleCenterFillMesh.SetMeshVertices(verts)
				w.titleCenterFillMesh.SetMeshIndices(inds)
				w.titleCenterFillMesh.InvalidateMeshAABB()
			}
			w.titleCenterFillMesh.SetPosition(0, 0)
			w.titleCenterFillMesh.SetVisible(true)
		}

	case BgGradient:
		// Full gradient title bar (no nine-grid frame).
		verts, inds := render.RoundedRectGradientMeshPerCorner(
			w.Width, tbH, cr, cr, 0, 0, defaultCornerSegments, bg.Gradient)
		if w.titleGradientMesh == nil {
			w.titleGradientMesh = sg.NewMesh(w.node.Name+"-title-grad", sg.WhitePixel, verts, inds)
			w.titleGradientMesh.SetColor(sg.RGBA(1, 1, 1, 1))
			w.titleBar.AddChildAt(w.titleGradientMesh, 0)
		} else {
			w.titleGradientMesh.SetMeshVertices(verts)
			w.titleGradientMesh.SetMeshIndices(inds)
			w.titleGradientMesh.InvalidateMeshAABB()
		}
		w.titleGradientMesh.SetVisible(true)

	default:
		// Solid color title bar.
		if cr > 0 {
			w.titleBg.SetColor(bg.Color)
			w.titleBg.SetVisible(true)
			w.updateTitlePoly(cr)
		} else {
			w.titleBgFlat.SetColor(bg.Color)
			w.titleBgFlat.SetVisible(true)
		}
	}
}

// SetTitle updates the window title text.
func (w *Window) SetTitle(t string) {
	w.titleLabel.SetText(t)
	w.updateTitleLayout()
}

// SetSize sets the window dimensions and recalculates internal layout.
func (w *Window) SetSize(width, height float64) {
	if width < w.minWidth {
		width = w.minWidth
	}
	if height < w.minHeight {
		height = w.minHeight
	}
	w.Width = width
	w.Height = height

	w.resizeBackground(width, height)
	w.resizeBorder(width, height)
	w.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: width, Height: height}

	w.updateTitleLayout()
	w.updateBodyLayout()
	w.updateResizeHandle()
	w.MarkLayoutDirty()
}

// SetTitleBarHeight sets the title bar height and recalculates layout.
func (w *Window) SetTitleBarHeight(h float64) {
	w.titleBarHeight = h
	w.closeBtn.SetSize(h-12, h-12)
	w.applyThemeColors()
	w.SetSize(w.Width, w.Height)
}

// SetMinSize sets the minimum window dimensions enforced during resize.
func (w *Window) SetMinSize(minW, minH float64) {
	w.minWidth = minW
	w.minHeight = minH
}

// SizeToContent sizes the window body to fit its children, then resizes the
// window to match (adding the title bar height). The resulting size is also
// set as the minimum size.
func (w *Window) SizeToContent() {
	if w.body != nil {
		w.body.SizeToContent()
		width := w.body.Width
		height := w.body.Height + w.titleBarHeight
		w.SetSize(width, height)
		w.SetMinSize(width, height)
	}
}

// MinSize returns the current minimum window dimensions.
func (w *Window) MinSize() (float64, float64) {
	return w.minWidth, w.minHeight
}

// SetResizable enables or disables the resize handle at the bottom-right.
func (w *Window) SetResizable(v bool) {
	w.resizable = v
	w.updateResizeHandle()
}

// SetMovable controls whether the window can be repositioned by dragging the
// title bar. Defaults to true. Set to false for fixed-position dialogs.
func (w *Window) SetMovable(v bool) {
	w.movable = v
}

// IsMovable reports whether the window can be dragged.
func (w *Window) IsMovable() bool {
	return w.movable
}

// SetModal enables or disables modal mode. When modal, a semi-transparent
// overlay fills the viewport behind the window and intercepts all pointer
// input so nothing below it can be clicked. The overlay is created lazily
// the first time the window has a parent in the scene graph; call this after
// adding the window to the scene (or show it after SetModal) to guarantee the
// overlay is attached.
func (w *Window) SetModal(v bool) {
	if w.modal == v {
		return
	}
	w.modal = v
	if v {
		w.ensureModalOverlay()
		if w.modalOverlay != nil {
			w.modalOverlay.SetVisible(w.IsVisible())
			w.syncModalOverlayZIndex()
		}
	} else if w.modalOverlay != nil {
		w.modalOverlay.SetVisible(false)
	}
}

// IsModal reports whether the window is in modal mode.
func (w *Window) IsModal() bool {
	return w.modal
}

// SetModalOverlayColor sets the color of the modal backdrop overlay.
// The default is semi-transparent black (RGBA 0,0,0,0.5).
func (w *Window) SetModalOverlayColor(c sg.Color) {
	w.modalOverlayColor = c
	if w.modalOverlay != nil {
		w.modalOverlay.SetColor(c)
	}
}

// SetOnModalOverlayClick sets a callback invoked when the user clicks the
// modal backdrop overlay. Set to nil to disable. Use this to implement
// "click outside to dismiss" behavior.
func (w *Window) SetOnModalOverlayClick(fn func()) {
	w.modalOverlayOnClick = fn
}

// SetVisible overrides Component.SetVisible to keep the modal overlay in sync.
func (w *Window) SetVisible(v bool) {
	w.Component.SetVisible(v)
	if w.modal {
		if v {
			w.ensureModalOverlay()
		}
		if w.modalOverlay != nil {
			w.modalOverlay.SetVisible(v)
		}
	}
}

// ensureModalOverlay lazily creates and attaches the modal backdrop node to
// the window's parent. Does nothing if the overlay already exists or the
// window has no parent yet.
func (w *Window) ensureModalOverlay() {
	if w.modalOverlay != nil || !w.modal {
		return
	}
	parent := w.node.Parent
	if parent == nil {
		return
	}
	overlay := sg.NewSprite(w.node.Name+"-modal-overlay", sg.TextureRegion{})
	overlay.Interactable = true
	overlay.HitShape = sg.HitRect{
		X: 0, Y: 0,
		Width:  modalOverlayHalfSize * 2,
		Height: modalOverlayHalfSize * 2,
	}
	overlay.SetPosition(-modalOverlayHalfSize, -modalOverlayHalfSize)
	overlay.SetScale(modalOverlayHalfSize*2, modalOverlayHalfSize*2)
	overlay.SetColor(w.modalOverlayColor)
	overlay.SetVisible(false) // caller sets visible state
	// Wire click handler once; w.modalOverlayOnClick is read at call time so
	// it can be swapped after creation without re-wiring.
	overlay.OnPointerDown(func(_ sg.PointerContext) {
		if w.modalOverlayOnClick != nil {
			w.modalOverlayOnClick()
		}
	})
	parent.AddChild(overlay)
	w.modalOverlay = overlay
	w.syncModalOverlayZIndex()
}

// syncModalOverlayZIndex positions the modal overlay z-index just below the
// window so it sits above all other content but below this window.
func (w *Window) syncModalOverlayZIndex() {
	if w.modalOverlay != nil {
		w.modalOverlay.SetZIndex(w.node.ZIndex() - 1)
	}
}

// SetCloseIcon sets a custom image for the close button icon at its native
// size and centers it within the button.
func (w *Window) SetCloseIcon(img engine.Image) {
	w.closeIcon = img
	w.applyCloseImage(img)
}

// closeIconDisplaySize is the desired pixel size for the close button icon.
const closeIconDisplaySize = 11.0

// applyCloseImage applies an icon image to the close button.
// SetIconSize + SetIconImage ensures layoutChildren uses the divide-by-dims
// path, scaling the image to the desired display size regardless of native
// resolution (e.g. a 48px spritesheet glyph renders at 11px).
func (w *Window) applyCloseImage(img engine.Image) {
	w.closeBtn.SetIconImage(img)
	w.closeBtn.SetIconSize(closeIconDisplaySize, closeIconDisplaySize)
}

// SetCloseable shows or hides the close button.
func (w *Window) SetCloseable(v bool) {
	w.closeBtn.SetVisible(v)
	w.updateTitleLayout()
}

// Body returns the body panel where content should be added.
func (w *Window) Body() *Panel {
	return w.body
}

// TitleLabel returns the window's title Label widget.
// Used for testing window internals.
func (w *Window) TitleLabel() *Label { return w.titleLabel }

// CloseBtn returns the window's close IconButton widget.
// Used for testing window internals.
func (w *Window) CloseBtn() *IconButton { return w.closeBtn }

// ResizeHandle returns the polygon resize grip node.
// Used for testing resize handle visibility.
func (w *Window) ResizeHandle() *sg.Node { return w.resizeHandle }

// ResizeFlat returns the flat sprite resize handle node.
// Used for testing resize handle visibility.
func (w *Window) ResizeFlat() *sg.Node { return w.resizeFlat }

// SetMinWidth sets the minimum width constraint for the window.
// Used for testing resize clamping.
func (w *Window) SetMinWidth(v float64) { w.minWidth = v }

// SetMinHeight sets the minimum height constraint for the window.
// Used for testing resize clamping.
func (w *Window) SetMinHeight(v float64) { w.minHeight = v }

// SetOnClose sets the callback invoked when the window is closed.
func (w *Window) SetOnClose(fn func()) {
	w.onClose = fn
}

// SetOnResult registers the handler called when FireResult is invoked.
// key identifies the outcome (e.g. "confirm", "cancel"); data is optional payload.
// Set to nil to remove. Intended to be set by the caller each time the modal
// is opened so each invocation can handle its own outcome independently.
func (w *Window) SetOnResult(fn func(key string, data any)) {
	w.onResult = fn
}

// FireResult hides the window and delivers a result to the registered handler.
// Typical use: button clicks or keyboard shortcuts inside the modal call
// FireResult("confirm", nil), FireResult("cancel", nil), etc.
func (w *Window) FireResult(key string, data any) {
	w.SetVisible(false)
	if w.onResult != nil {
		w.onResult(key, data)
	}
}

// SetEscResult configures Escape to call FireResult(key, nil) while the window
// is visible. Pass "" to disable.
func (w *Window) SetEscResult(key string) {
	w.escResultKey = key
}

// SetEnterResult configures Enter to call FireResult(key, nil) while the window
// is visible. Pass "" to disable. Note: if a TextInput inside the dialog has
// focus, Enter is consumed by that input first.
func (w *Window) SetEnterResult(key string) {
	w.enterResultKey = key
}

// BringToFront moves this window in front of siblings by setting a high
// Z-index via the DefaultWindowManager.
func (w *Window) BringToFront() {
	DefaultWindowManager.BringToFront(w)
}

// setActive updates the title bar color to reflect active or inactive state.
func (w *Window) setActive(active bool) {
	group := w.EffectiveTheme().Window.Group(w.Variant())
	var titleBg Background
	if active {
		titleBg = group.TitleBackground.Resolve(StateFocus)
	} else {
		titleBg = group.TitleBackground.Resolve(StateDefault)
	}
	w.applyTitleBackground(titleBg, group.CornerRadius)
}

// ---------------------------------------------------------------------------
// WindowManager
// ---------------------------------------------------------------------------

// WindowManager tracks a set of windows and handles z-ordering so that
// clicking any registered window automatically brings it to the front.
// It also tracks the active window and updates title bar colors.
type WindowManager struct {
	windows  []*Window
	active   *Window
	zCounter int
}

// DefaultWindowManager is the package-level window manager. Register windows
// here so they automatically sort to the top on click.
var DefaultWindowManager = NewWindowManager()

// NewWindowManager creates an empty WindowManager.
func NewWindowManager() *WindowManager {
	return &WindowManager{}
}

// Add registers a window with the manager. Clicking anywhere on the window
// (title bar, body, resize handle) automatically brings it to the front.
func (wm *WindowManager) Add(w *Window) {
	for _, existing := range wm.windows {
		if existing == w {
			return
		}
	}
	wm.windows = append(wm.windows, w)

	// Set the activation hook so all internal interaction points
	// (title bar, body, resize handle) route through BringToFront.
	w.onActivate = func() {
		wm.BringToFront(w)
	}

	// Assign an initial z-index and mark as active (last added is on top).
	wm.zCounter++
	w.node.SetZIndex(wm.zCounter)
	w.syncModalOverlayZIndex()
	if wm.active != nil {
		wm.active.setActive(false)
	}
	wm.active = w
	w.setActive(true)
}

// Remove unregisters a window from the manager.
func (wm *WindowManager) Remove(w *Window) {
	for i, existing := range wm.windows {
		if existing == w {
			wm.windows = append(wm.windows[:i], wm.windows[i+1:]...)
			w.onActivate = nil
			if wm.active == w {
				wm.active = nil
			}
			return
		}
	}
}

// BringToFront moves the given window in front of all other managed windows
// and marks it as the active window, updating title bar colors.
func (wm *WindowManager) BringToFront(w *Window) {
	wm.zCounter++
	w.node.SetZIndex(wm.zCounter)
	w.syncModalOverlayZIndex()
	if wm.active != w {
		if wm.active != nil {
			wm.active.setActive(false)
		}
		wm.active = w
		w.setActive(true)
	}
}

// Active returns the currently active (foremost) window, or nil.
func (wm *WindowManager) Active() *Window {
	return wm.active
}

// Windows returns the currently registered windows.
func (wm *WindowManager) Windows() []*Window {
	return wm.windows
}

// Close hides the window and fires the onClose callback.
func (w *Window) Close() {
	w.SetVisible(false)
	if w.onClose != nil {
		w.onClose()
	}
}

// Dispose cleans up the window and its children.
func (w *Window) Dispose() {
	if w.modalOverlay != nil {
		if w.modalOverlay.Parent != nil {
			w.modalOverlay.Parent.RemoveChild(w.modalOverlay)
		}
		w.modalOverlay.Dispose()
		w.modalOverlay = nil
	}
	if w.titleLabel != nil {
		w.titleLabel.Dispose()
	}
	if w.closeBtn != nil {
		w.closeBtn.Dispose()
	}
	if w.body != nil {
		w.body.Dispose()
	}
	w.Component.Dispose()
}

// updateTitleLayout positions the title bar, label, and close button.
func (w *Window) updateTitleLayout() {
	tbH := w.titleBarHeight

	// Title bar spans the full width.
	w.titleBar.HitShape = sg.HitRect{X: 0, Y: 0, Width: w.Width, Height: tbH}

	// Title bar background: re-apply to handle resize.
	w.titleBgFlat.SetScale(w.Width, tbH)
	group := w.EffectiveTheme().Window.Group(w.Variant())
	cr := group.CornerRadius
	if cr > 0 {
		w.updateTitlePoly(cr)
	}
	// Rebuild title background (gradient, nine-grid, etc.) at new dimensions.
	var titleBg Background
	if DefaultWindowManager.Active() == w {
		titleBg = group.TitleBackground.Resolve(StateFocus)
	} else {
		titleBg = group.TitleBackground.Resolve(StateDefault)
	}
	if titleBg.Type != BgNone {
		w.applyTitleBackground(titleBg, cr)
	}

	// When the body overlaps the title bar's bottom border, center
	// elements within the visible portion of the title bar.
	visibleH := tbH
	if w.titleSlice != nil {
		visibleH = tbH - w.titleSlice.Insets.Bottom
	}

	// Title label: vertically centered in visible area, left-padded.
	pad := float64(DefaultWindowPadding)
	ly := (visibleH - w.titleLabel.Height) / 2
	w.titleLabel.SetPosition(pad, ly)

	// Close button: right-aligned, vertically centered in visible area,
	// matching title left padding.
	btnW := w.closeBtn.Width
	btnH := w.closeBtn.Height
	bx := w.Width - btnW - pad
	by := (visibleH-btnH)/2 + 1
	w.closeBtn.SetPosition(bx, by)
}

// updateTitlePoly rebuilds the title bar polygon with rounded top corners.
func (w *Window) updateTitlePoly(cr float64) {
	tbH := w.titleBarHeight
	pts := render.RoundedRectPointsPerCorner(w.Width, tbH, cr, cr, 0, 0, defaultCornerSegments)
	sg.SetPolygonPoints(w.titleBg, pts)
}

// updateBodyLayout positions the body panel below the title bar.
// When the title bar uses a nine-grid background, the body overlaps the
// title bar's bottom border so the two areas appear seamlessly connected.
// When ContentPaneUnderTitleBar is enabled, the body extends to Y=0 and
// uses Padding.Top to keep its content below the title bar.
func (w *Window) updateBodyLayout() {
	tbH := w.titleBarHeight

	var bodyY, bodyH float64
	if w.contentPaneUnderTitleBar {
		// Body fills the full window height, rendered behind the title bar.
		bodyY = 0
		bodyH = w.Height
		w.body.Padding.Top = tbH
	} else {
		// Overlap the body into the title bar's bottom border region so the
		// title and body appear seamlessly connected.
		overlap := 0.0
		if w.titleSlice != nil {
			overlap = w.titleSlice.Insets.Bottom
		}
		bodyY = tbH - overlap
		bodyH = w.Height - bodyY
		w.body.Padding.Top = 0
	}

	if bodyH < 0 {
		bodyH = 0
	}
	w.body.SetSize(w.Width, bodyH)
	w.body.SetPosition(0, bodyY)

	// Override body background polygon with per-corner radii after SetSize
	// (which rebuilds with uniform radius).
	w.applyBodyCorners()
}

// applyBodyCorners overrides the body panel's background polygon so the
// corners match the window's corner radius. When ContentPaneUnderTitleBar is
// enabled the body covers the full window so all four corners are rounded;
// otherwise only the bottom corners are rounded (top corners sit flush under
// the title bar).
func (w *Window) applyBodyCorners() {
	cr := w.EffectiveTheme().Window.Group(w.Variant()).CornerRadius
	if cr <= 0 || w.body == nil {
		return
	}
	// Force the body panel to use a polygon background so corners
	// can be rounded independently.
	w.body.cornerRadius = cr
	w.body.ensureBgPoly()
	var pts []sg.Vec2
	if w.contentPaneUnderTitleBar {
		pts = render.RoundedRectPointsPerCorner(w.body.Width, w.body.Height, cr, cr, cr, cr, defaultCornerSegments)
	} else {
		pts = render.RoundedRectPointsPerCorner(w.body.Width, w.body.Height, 0, 0, cr, cr, defaultCornerSegments)
	}
	sg.SetPolygonPoints(w.body.bgPoly, pts)
	// Apply the body panel's actual background color.
	bodyGroup := w.EffectiveTheme().Panel.Group(w.body.Variant())
	bodyBg := bodyGroup.Background.Resolve(w.body.state)
	w.body.bgPoly.SetColor(bodyBg.Color)
	w.body.bgPoly.SetVisible(true)
	if w.body.bgNode != nil {
		w.body.bgNode.SetVisible(false)
	}
}

// updateResizeHandle positions the resize grip at the bottom-right corner.
func (w *Window) updateResizeHandle() {
	gs := float64(WindowResizeGripSize)
	group := w.EffectiveTheme().Window.Group(w.Variant())
	cr := group.CornerRadius

	if group.ResizeIcon.Set {
		// Theme sprite icon for the resize handle.
		w.resizeHandle.SetVisible(false)
		w.resizeFlat.SetVisible(false)
		if w.appliedResizeIcon != group.ResizeIcon.Image {
			w.appliedResizeIcon = group.ResizeIcon.Image
			w.resizeIcon.SetCustomImage(group.ResizeIcon.Image)
			w.resizeIcon.SetScale(1, 1)
		}
		b := group.ResizeIcon.Image.Bounds()
		iw := float64(b.Dx())
		ih := float64(b.Dy())
		w.resizeIcon.SetPosition(w.Width-iw, w.Height-ih)
		w.resizeIcon.HitShape = sg.HitRect{X: 0, Y: 0, Width: iw, Height: ih}
		w.resizeIcon.SetColor(group.ResizeHandleColor.Resolve(StateDefault))
		w.resizeIcon.SetVisible(w.resizable)
	} else if cr > 0 {
		// Polygon with rounded bottom-right corner.
		w.resizeIcon.SetVisible(false)
		w.updateResizePoly(cr)
		w.resizeHandle.SetPosition(w.Width-gs, w.Height-gs)
		w.resizeHandle.HitShape = sg.HitRect{X: 0, Y: 0, Width: gs, Height: gs}
		w.resizeHandle.SetVisible(w.resizable)
		w.resizeFlat.SetVisible(false)
	} else {
		// Flat sprite.
		w.resizeIcon.SetVisible(false)
		w.resizeFlat.SetScale(gs, gs)
		w.resizeFlat.SetPosition(w.Width-gs, w.Height-gs)
		w.resizeFlat.HitShape = sg.HitRect{X: 0, Y: 0, Width: gs, Height: gs}
		w.resizeFlat.SetVisible(w.resizable)
		w.resizeHandle.SetVisible(false)
	}
}

// updateResizePoly rebuilds the resize handle polygon with a rounded bottom-right corner.
func (w *Window) updateResizePoly(cr float64) {
	gs := float64(WindowResizeGripSize)
	pts := render.RoundedRectPointsPerCorner(gs, gs, 0, 0, cr, 0, defaultCornerSegments)
	sg.SetPolygonPoints(w.resizeHandle, pts)
}
