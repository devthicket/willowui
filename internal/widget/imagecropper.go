package widget

import (
	"image"
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// ImageCropper displays an image with a draggable crop rectangle.
// Users can drag corner and edge handles to resize the crop region,
// or drag inside the crop rect to move it. The widget returns crop
// coordinates in image-pixel space via CropRect().
type ImageCropper struct {
	Component

	// Source image.
	srcImg engine.Image

	// Image display node (scaled to fit widget bounds).
	imgNode *sg.Node

	// Composite canvas used for rendering the dimmed overlay + crop border + grid.
	overlayCanvas engine.Image
	overlayNode   *sg.Node

	// Crop rectangle in image-pixel coordinates.
	cropX, cropY, cropW, cropH float64

	// Constraints.
	aspectW, aspectH float64 // 0,0 = free
	minW, minH       float64
	maxW, maxH       float64

	// Appearance.
	showGrid bool

	// Handles: 0-3 = corners (TL, TR, BR, BL), 4-7 = edges (T, R, B, L).
	handles    [8]Component
	handleSize float64

	// Drag state.
	dragging     bool
	dragHandle   int     // -1 = moving whole rect, 0-7 = handle index
	dragStartCX  float64 // crop rect at drag start
	dragStartCY  float64
	dragStartCW  float64
	dragStartCH  float64
	dragOriginGX float64 // global pointer pos at drag start
	dragOriginGY float64

	// Callback.
	onCropChanged func(rect image.Rectangle)

	// Cached display scale: how image pixels map to widget pixels.
	displayScale float64
	displayOffX  float64
	displayOffY  float64
	displayW     float64
	displayH     float64
}

const (
	defaultCropperSize       = 300
	defaultCropperHandleSize = 10
)

// NewImageCropper creates an ImageCropper widget.
func NewImageCropper(name string) *ImageCropper {
	c := &ImageCropper{
		handleSize: defaultCropperHandleSize,
		dragHandle: -1,
	}
	initComponent(&c.Component, name)
	c.initBackground(name)
	c.initBorder(name)

	// Image sprite node.
	c.imgNode = sg.NewSprite(name+"-img", sg.TextureRegion{})
	c.imgNode.SetVisible(false)
	c.node.AddChild(c.imgNode)

	// Overlay node for dim + border + grid.
	c.overlayNode = sg.NewSprite(name+"-overlay", sg.TextureRegion{})
	c.overlayNode.SetVisible(false)
	c.node.AddChild(c.overlayNode)

	// Create 8 handle sub-components.
	for i := 0; i < 8; i++ {
		h := &c.handles[i]
		initComponent(h, name+"-handle")
		h.initBackground(name + "-handle")
		c.node.AddChild(h.node)
		h.node.Interactable = true
		h.node.SetVisible(false)

		idx := i
		h.node.OnDragStart(func(ctx sg.DragContext) {
			if !c.enabled {
				return
			}
			c.startDrag(idx, ctx.GlobalX, ctx.GlobalY)
		})
		h.node.OnDrag(func(ctx sg.DragContext) {
			if !c.enabled || !c.dragging {
				return
			}
			c.handleDrag(ctx.GlobalX, ctx.GlobalY)
		})
		h.node.OnDragEnd(func(_ sg.DragContext) {
			c.dragging = false
		})
		h.node.OnPointerEnter(func(_ sg.PointerContext) {
			c.hovered = true
			c.MarkDrawDirty()
		})
		h.node.OnPointerLeave(func(_ sg.PointerContext) {
			if !c.dragging {
				c.hovered = false
				c.MarkDrawDirty()
			}
		})
	}

	// Drag on the main node (move crop rect).
	c.node.OnDragStart(func(ctx sg.DragContext) {
		if !c.enabled || c.srcImg == nil {
			return
		}
		// Check if pointer is inside crop rect (in widget coords).
		wx, wy := c.imageToWidget(c.cropX, c.cropY)
		ww, wh := c.cropW*c.displayScale, c.cropH*c.displayScale
		lx, ly := ctx.LocalX, ctx.LocalY
		if lx >= wx && lx <= wx+ww && ly >= wy && ly <= wy+wh {
			c.startDrag(-1, ctx.GlobalX, ctx.GlobalY)
		}
	})
	c.node.OnDrag(func(ctx sg.DragContext) {
		if !c.enabled || !c.dragging {
			return
		}
		c.handleDrag(ctx.GlobalX, ctx.GlobalY)
	})
	c.node.OnDragEnd(func(_ sg.DragContext) {
		c.dragging = false
	})

	c.onThemeChange = func() { c.UpdateVisuals() }
	c.SetSize(defaultCropperSize, defaultCropperSize)
	c.UpdateVisuals()
	return c
}

// SetImage sets the source image.
func (c *ImageCropper) SetImage(img engine.Image) {
	c.srcImg = img
	if img != nil {
		c.imgNode.SetCustomImage(img)
		c.imgNode.SetTextureRegion(sg.TextureRegion{})
		c.imgNode.SetVisible(true)
		// Default crop to full image if not set.
		b := img.Bounds()
		if c.cropW == 0 && c.cropH == 0 {
			c.cropX, c.cropY = 0, 0
			c.cropW, c.cropH = float64(b.Dx()), float64(b.Dy())
		}
	} else {
		c.imgNode.SetCustomImage(nil)
		c.imgNode.SetVisible(false)
	}
	c.rebuildLayout()
}

// SetCropRect sets the crop rectangle in image-pixel coordinates.
func (c *ImageCropper) SetCropRect(x, y, w, h float64) {
	c.cropX, c.cropY = x, y
	c.cropW, c.cropH = w, h
	c.clampCrop()
	c.rebuildOverlay()
	c.positionHandles()
	c.MarkDrawDirty()
}

// CropRect returns the current crop rectangle in image-pixel coordinates.
func (c *ImageCropper) CropRect() image.Rectangle {
	x0 := int(math.Round(c.cropX))
	y0 := int(math.Round(c.cropY))
	x1 := int(math.Round(c.cropX + c.cropW))
	y1 := int(math.Round(c.cropY + c.cropH))
	return image.Rect(x0, y0, x1, y1)
}

// SetAspectRatio constrains the crop to a fixed aspect ratio. Pass 0, 0 for free.
func (c *ImageCropper) SetAspectRatio(w, h float64) {
	c.aspectW, c.aspectH = w, h
	if w > 0 && h > 0 {
		c.enforceAspect()
		c.clampCrop()
		c.rebuildOverlay()
		c.positionHandles()
		c.MarkDrawDirty()
	}
}

// SetMinSize sets the minimum crop size in image pixels.
func (c *ImageCropper) SetMinSize(w, h float64) {
	c.minW, c.minH = w, h
}

// SetMaxSize sets the maximum crop size in image pixels.
func (c *ImageCropper) SetMaxSize(w, h float64) {
	c.maxW, c.maxH = w, h
}

// SetShowGrid enables/disables the rule-of-thirds grid overlay.
func (c *ImageCropper) SetShowGrid(v bool) {
	c.showGrid = v
	c.rebuildOverlay()
	c.MarkDrawDirty()
}

// SetOnCropChanged sets the callback fired during handle drags.
func (c *ImageCropper) SetOnCropChanged(fn func(rect image.Rectangle)) {
	c.onCropChanged = fn
}

// SetSize sets the widget dimensions.
func (c *ImageCropper) SetSize(w, h float64) {
	c.Width = w
	c.Height = h
	c.resizeBackground(w, h)
	c.resizeBorder(w, h)
	c.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	c.rebuildLayout()
	c.MarkLayoutDirty()
}

// UpdateVisuals applies theme colors.
func (c *ImageCropper) UpdateVisuals() {
	c.state = computeState(c.enabled, c.focused, c.hovered, c.pressed)
	group := c.EffectiveTheme().ImageCropper.Group(c.Variant())

	bg := group.Background.Resolve(c.state)
	c.applyBackground(bg)

	cr := resolveCornerRadius(group.CornerRadius, math.Min(c.Width, c.Height))
	c.applyCornerRadius(cr)

	// Update handle appearance.
	hs := group.HandleSize
	if hs <= 0 {
		hs = defaultCropperHandleSize
	}
	c.handleSize = hs

	for i := range c.handles {
		hState := c.state
		handleBg := group.HandleBackground.Resolve(hState)
		c.handles[i].applyBackground(handleBg)
		hcr := resolveCornerRadius(group.HandleCornerRadius, hs)
		c.handles[i].applyCornerRadius(hcr)
	}

	c.rebuildOverlay()
	c.positionHandles()
	c.MarkDrawDirty()
}

// Dispose releases resources.
func (c *ImageCropper) Dispose() {
	for i := range c.handles {
		c.handles[i].Dispose()
	}
	c.Component.Dispose()
}

// --- Internal ---

// rebuildLayout recomputes image scaling and repositions everything.
func (c *ImageCropper) rebuildLayout() {
	if c.srcImg == nil {
		c.overlayNode.SetVisible(false)
		for i := range c.handles {
			c.handles[i].node.SetVisible(false)
		}
		return
	}

	b := c.srcImg.Bounds()
	imgW, imgH := float64(b.Dx()), float64(b.Dy())
	if imgW <= 0 || imgH <= 0 {
		return
	}

	// Scale image to fit widget (letterbox).
	scale := math.Min(c.Width/imgW, c.Height/imgH)
	c.displayScale = scale
	c.displayW = imgW * scale
	c.displayH = imgH * scale
	c.displayOffX = (c.Width - c.displayW) / 2
	c.displayOffY = (c.Height - c.displayH) / 2

	// Position and scale the image node.
	c.imgNode.SetScale(scale, scale)
	c.imgNode.SetPosition(c.displayOffX, c.displayOffY)

	c.rebuildOverlay()
	c.positionHandles()
	c.MarkDrawDirty()
}

// imageToWidget converts image-pixel coordinates to widget-local coordinates.
func (c *ImageCropper) imageToWidget(ix, iy float64) (float64, float64) {
	return c.displayOffX + ix*c.displayScale, c.displayOffY + iy*c.displayScale
}

// widgetToImage converts widget-local coordinates to image-pixel coordinates.
func (c *ImageCropper) widgetToImage(wx, wy float64) (float64, float64) {
	if c.displayScale == 0 {
		return 0, 0
	}
	return (wx - c.displayOffX) / c.displayScale, (wy - c.displayOffY) / c.displayScale
}

// rebuildOverlay composites the dim overlay, crop border, and optional grid
// onto a canvas the same size as the widget.
func (c *ImageCropper) rebuildOverlay() {
	if c.srcImg == nil {
		c.overlayNode.SetVisible(false)
		return
	}

	canvasW := int(math.Ceil(c.Width))
	canvasH := int(math.Ceil(c.Height))
	if canvasW <= 0 || canvasH <= 0 {
		return
	}

	if c.overlayCanvas == nil ||
		c.overlayCanvas.Bounds().Dx() != canvasW ||
		c.overlayCanvas.Bounds().Dy() != canvasH {
		c.overlayCanvas = engine.NewImage(canvasW, canvasH)
	} else {
		c.overlayCanvas.Clear()
	}

	group := c.EffectiveTheme().ImageCropper.Group(c.Variant())

	// Crop rect in widget coordinates.
	cx, cy := c.imageToWidget(c.cropX, c.cropY)
	cw := c.cropW * c.displayScale
	ch := c.cropH * c.displayScale

	// Draw dim overlay (fill everything, then clear the crop area).
	dimColor := group.DimColor.Resolve(c.state)
	dimR := uint8(dimColor.R() * 255)
	dimG := uint8(dimColor.G() * 255)
	dimB := uint8(dimColor.B() * 255)
	dimA := uint8(dimColor.A() * 255)

	// Fill entire canvas with dim color.
	dimImg := engine.NewImage(canvasW, canvasH)
	dimImg.Fill(toStdColor(dimR, dimG, dimB, dimA))
	c.overlayCanvas.DrawImage(dimImg, nil)

	// Clear the crop region (punch a hole).
	cropInt := image.Rect(
		int(math.Round(cx)), int(math.Round(cy)),
		int(math.Round(cx+cw)), int(math.Round(cy+ch)),
	)
	// Clamp to canvas bounds.
	canvasBounds := image.Rect(0, 0, canvasW, canvasH)
	cropInt = cropInt.Intersect(canvasBounds)
	if !cropInt.Empty() {
		clearImg := engine.NewImage(cropInt.Dx(), cropInt.Dy())
		clearImg.Fill(toStdColor(255, 255, 255, 255))
		// Draw with BlendDestinationOut to punch through the dim.
		var op engine.DrawImageOptions
		op.GeoM.Translate(float64(cropInt.Min.X), float64(cropInt.Min.Y))
		op.Blend = engine.BlendDestinationOut
		c.overlayCanvas.DrawImage(clearImg, &op)
	}

	// Draw crop border.
	borderColor := group.CropBorderColor.Resolve(c.state)
	bw := group.CropBorderWidth
	if bw <= 0 {
		bw = 1
	}
	drawRectBorder(c.overlayCanvas, cx, cy, cw, ch, bw, borderColor)

	// Draw rule-of-thirds grid if enabled.
	if c.showGrid {
		gridColor := group.GridColor.Resolve(c.state)
		glw := group.GridLineWidth
		if glw <= 0 {
			glw = 1
		}
		// Vertical lines at 1/3 and 2/3.
		for i := 1; i <= 2; i++ {
			gx := cx + cw*float64(i)/3
			drawLine(c.overlayCanvas, gx, cy, gx, cy+ch, glw, gridColor)
		}
		// Horizontal lines at 1/3 and 2/3.
		for i := 1; i <= 2; i++ {
			gy := cy + ch*float64(i)/3
			drawLine(c.overlayCanvas, cx, gy, cx+cw, gy, glw, gridColor)
		}
	}

	c.overlayNode.SetCustomImage(c.overlayCanvas)
	c.overlayNode.SetTextureRegion(sg.TextureRegion{})
	c.overlayNode.SetScale(1, 1)
	c.overlayNode.SetPosition(0, 0)
	c.overlayNode.SetVisible(true)
}

// positionHandles places the 8 drag handles at the crop rect boundary.
func (c *ImageCropper) positionHandles() {
	if c.srcImg == nil {
		return
	}

	cx, cy := c.imageToWidget(c.cropX, c.cropY)
	cw := c.cropW * c.displayScale
	ch := c.cropH * c.displayScale
	hs := c.handleSize
	half := hs / 2

	// Handle positions: TL, TR, BR, BL, T-mid, R-mid, B-mid, L-mid.
	positions := [8][2]float64{
		{cx - half, cy - half},             // TL
		{cx + cw - half, cy - half},        // TR
		{cx + cw - half, cy + ch - half},   // BR
		{cx - half, cy + ch - half},        // BL
		{cx + cw/2 - half, cy - half},      // T-mid
		{cx + cw - half, cy + ch/2 - half}, // R-mid
		{cx + cw/2 - half, cy + ch - half}, // B-mid
		{cx - half, cy + ch/2 - half},      // L-mid
	}

	for i := range c.handles {
		h := &c.handles[i]
		h.Width = hs
		h.Height = hs
		h.resizeBackground(hs, hs)
		h.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: hs, Height: hs}
		h.node.SetPosition(positions[i][0], positions[i][1])
		h.node.SetVisible(true)
	}
}

// startDrag begins a drag operation for the given handle (or -1 for move).
func (c *ImageCropper) startDrag(handleIdx int, gx, gy float64) {
	c.dragging = true
	c.dragHandle = handleIdx
	c.dragStartCX = c.cropX
	c.dragStartCY = c.cropY
	c.dragStartCW = c.cropW
	c.dragStartCH = c.cropH
	c.dragOriginGX = gx
	c.dragOriginGY = gy
}

// handleDrag processes a drag event, updating the crop rect.
func (c *ImageCropper) handleDrag(gx, gy float64) {
	if c.displayScale == 0 {
		return
	}

	// Delta in image pixels.
	dx := (gx - c.dragOriginGX) / c.displayScale
	dy := (gy - c.dragOriginGY) / c.displayScale

	switch c.dragHandle {
	case -1: // Move whole rect.
		c.cropX = c.dragStartCX + dx
		c.cropY = c.dragStartCY + dy
	case 0: // TL corner.
		c.cropX = c.dragStartCX + dx
		c.cropY = c.dragStartCY + dy
		c.cropW = c.dragStartCW - dx
		c.cropH = c.dragStartCH - dy
	case 1: // TR corner.
		c.cropY = c.dragStartCY + dy
		c.cropW = c.dragStartCW + dx
		c.cropH = c.dragStartCH - dy
	case 2: // BR corner.
		c.cropW = c.dragStartCW + dx
		c.cropH = c.dragStartCH + dy
	case 3: // BL corner.
		c.cropX = c.dragStartCX + dx
		c.cropW = c.dragStartCW - dx
		c.cropH = c.dragStartCH + dy
	case 4: // T edge.
		c.cropY = c.dragStartCY + dy
		c.cropH = c.dragStartCH - dy
	case 5: // R edge.
		c.cropW = c.dragStartCW + dx
	case 6: // B edge.
		c.cropH = c.dragStartCH + dy
	case 7: // L edge.
		c.cropX = c.dragStartCX + dx
		c.cropW = c.dragStartCW - dx
	}

	// Enforce minimum size.
	if c.minW > 0 && c.cropW < c.minW {
		if c.dragHandle == 0 || c.dragHandle == 3 || c.dragHandle == 7 {
			c.cropX = c.cropX + c.cropW - c.minW
		}
		c.cropW = c.minW
	}
	if c.minH > 0 && c.cropH < c.minH {
		if c.dragHandle == 0 || c.dragHandle == 1 || c.dragHandle == 4 {
			c.cropY = c.cropY + c.cropH - c.minH
		}
		c.cropH = c.minH
	}

	// Enforce maximum size.
	if c.maxW > 0 && c.cropW > c.maxW {
		c.cropW = c.maxW
	}
	if c.maxH > 0 && c.cropH > c.maxH {
		c.cropH = c.maxH
	}

	// Enforce aspect ratio for corner drags.
	if c.aspectW > 0 && c.aspectH > 0 && c.dragHandle >= 0 && c.dragHandle <= 3 {
		c.enforceAspectFromDrag()
	}

	c.clampCrop()
	c.rebuildOverlay()
	c.positionHandles()

	if c.onCropChanged != nil {
		c.onCropChanged(c.CropRect())
	}
}

// enforceAspect adjusts crop dimensions to match the aspect ratio.
func (c *ImageCropper) enforceAspect() {
	if c.aspectW <= 0 || c.aspectH <= 0 {
		return
	}
	ratio := c.aspectW / c.aspectH
	currentRatio := c.cropW / c.cropH
	if currentRatio > ratio {
		c.cropW = c.cropH * ratio
	} else {
		c.cropH = c.cropW / ratio
	}
}

// enforceAspectFromDrag adjusts crop dimensions based on which corner is dragged.
func (c *ImageCropper) enforceAspectFromDrag() {
	ratio := c.aspectW / c.aspectH
	// Use the larger delta to determine the dominant axis.
	newH := c.cropW / ratio
	switch c.dragHandle {
	case 0: // TL — anchor BR.
		brX := c.dragStartCX + c.dragStartCW
		brY := c.dragStartCY + c.dragStartCH
		c.cropH = newH
		c.cropX = brX - c.cropW
		c.cropY = brY - c.cropH
	case 1: // TR — anchor BL.
		blY := c.dragStartCY + c.dragStartCH
		c.cropH = newH
		c.cropY = blY - c.cropH
	case 2: // BR — anchor TL.
		c.cropH = newH
	case 3: // BL — anchor TR.
		trX := c.dragStartCX + c.dragStartCW
		c.cropH = newH
		c.cropX = trX - c.cropW
	}
}

// clampCrop ensures the crop rect stays within the image bounds and has positive dimensions.
func (c *ImageCropper) clampCrop() {
	if c.srcImg == nil {
		return
	}
	b := c.srcImg.Bounds()
	imgW, imgH := float64(b.Dx()), float64(b.Dy())

	// Ensure positive dimensions.
	if c.cropW < 1 {
		c.cropW = 1
	}
	if c.cropH < 1 {
		c.cropH = 1
	}

	// Clamp to image bounds.
	if c.cropX < 0 {
		c.cropX = 0
	}
	if c.cropY < 0 {
		c.cropY = 0
	}
	if c.cropX+c.cropW > imgW {
		c.cropX = imgW - c.cropW
		if c.cropX < 0 {
			c.cropX = 0
			c.cropW = imgW
		}
	}
	if c.cropY+c.cropH > imgH {
		c.cropY = imgH - c.cropH
		if c.cropY < 0 {
			c.cropY = 0
			c.cropH = imgH
		}
	}
}

// --- Drawing helpers ---

// drawRectBorder draws a rectangular border on the canvas.
func drawRectBorder(canvas engine.Image, x, y, w, h, lineWidth float64, color sg.Color) {
	c := toStdColor(
		uint8(color.R()*255),
		uint8(color.G()*255),
		uint8(color.B()*255),
		uint8(color.A()*255),
	)
	lw := int(math.Max(1, math.Round(lineWidth)))

	// Top edge.
	top := engine.NewImage(int(math.Ceil(w)), lw)
	top.Fill(c)
	var op engine.DrawImageOptions
	op.GeoM.Translate(x, y)
	canvas.DrawImage(top, &op)

	// Bottom edge.
	op.GeoM.Reset()
	op.GeoM.Translate(x, y+h-float64(lw))
	canvas.DrawImage(top, &op)

	// Left edge.
	left := engine.NewImage(lw, int(math.Ceil(h)))
	left.Fill(c)
	op.GeoM.Reset()
	op.GeoM.Translate(x, y)
	canvas.DrawImage(left, &op)

	// Right edge.
	op.GeoM.Reset()
	op.GeoM.Translate(x+w-float64(lw), y)
	canvas.DrawImage(left, &op)
}

// drawLine draws a line between two points (axis-aligned only for simplicity).
func drawLine(canvas engine.Image, x1, y1, x2, y2, lineWidth float64, color sg.Color) {
	c := toStdColor(
		uint8(color.R()*255),
		uint8(color.G()*255),
		uint8(color.B()*255),
		uint8(color.A()*255),
	)
	lw := int(math.Max(1, math.Round(lineWidth)))

	if math.Abs(x2-x1) < 1 {
		// Vertical line.
		h := int(math.Ceil(math.Abs(y2 - y1)))
		if h <= 0 {
			return
		}
		line := engine.NewImage(lw, h)
		line.Fill(c)
		var op engine.DrawImageOptions
		op.GeoM.Translate(x1-float64(lw)/2, math.Min(y1, y2))
		canvas.DrawImage(line, &op)
	} else {
		// Horizontal line.
		w := int(math.Ceil(math.Abs(x2 - x1)))
		if w <= 0 {
			return
		}
		line := engine.NewImage(w, lw)
		line.Fill(c)
		var op engine.DrawImageOptions
		op.GeoM.Translate(math.Min(x1, x2), y1-float64(lw)/2)
		canvas.DrawImage(line, &op)
	}
}

// toStdColor converts RGBA bytes to a standard color.
func toStdColor(r, g, b, a uint8) stdColor {
	return stdColor{r, g, b, a}
}

type stdColor struct{ R, G, B, A uint8 }

func (c stdColor) RGBA() (r, g, b, a uint32) {
	return uint32(c.R) * 257, uint32(c.G) * 257, uint32(c.B) * 257, uint32(c.A) * 257
}
