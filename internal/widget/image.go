package widget

import (
	"math"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// ImageScaleMode controls how the image is laid out within the widget bounds.
type ImageScaleMode int

const (
	ImageScaleStretch ImageScaleMode = iota // fill bounds exactly, ignoring aspect ratio
	ImageScaleFit                           // scale uniformly to fit inside bounds (letterbox)
	ImageScaleFill                          // scale uniformly to fill bounds (crop overflow)
	ImageScaleCenter                        // native pixel size, centered, no scaling
	ImageScaleTile                          // tile at native size to fill bounds
)

// Image is a display-only component that renders a sprite, texture region,
// or engine.Image with configurable fit/fill modes, tinting, and optional
// corner radius.
//
// For Fill and Center modes the image is positioned so overflow extends
// outside the widget bounds. Clipping at the widget edge is achieved
// automatically when the Image is placed inside a container that uses cache
// or mask — bare usage renders without pixel-level clip.
type Image struct {
	Component

	scaleMode ImageScaleMode
	tint      sg.Color

	// Source — at most one is set at a time.
	region    sg.TextureRegion
	nativeImg engine.Image
	hasImage  bool // true when either region or nativeImg is set

	// imgNode is a willow sprite child of the component node.
	imgNode *sg.Node

	// tileCanvas holds a pre-composited image for ImageScaleTile.
	// It is rebuilt whenever the widget size or source image changes.
	tileCanvas engine.Image

	// imgW/imgH track the pixel-space size applied to imgNode (for testing).
	imgW, imgH float64
	// imgX/imgY track the pixel-space position applied to imgNode (for testing).
	imgX, imgY float64
}

// NewImage creates an Image widget with no source set.
func NewImage(name string) *Image {
	im := &Image{
		tint: sg.RGBA(1, 1, 1, 1),
	}
	initComponent(&im.Component, name)
	im.initBackground(name)

	im.imgNode = sg.NewSprite(name+"-img", sg.TextureRegion{})
	im.imgNode.SetVisible(false)
	im.node.AddChild(im.imgNode)

	im.onThemeChange = func() { im.UpdateVisuals() }
	im.SetSize(64, 64)
	return im
}

// --- Source ---

// SetRegion sets the image source to a TextureRegion (atlas sprite).
// Clears any previously set engine.Image.
func (im *Image) SetRegion(region sg.TextureRegion) {
	im.region = region
	im.nativeImg = nil
	im.hasImage = region != (sg.TextureRegion{})
	im.imgNode.SetTextureRegion(region)
	im.imgNode.SetCustomImage(nil)
	im.imgNode.SetVisible(im.hasImage)
	im.rebuildLayout()
}

// SetImage sets the image source to a raw engine.Image.
// Clears any previously set TextureRegion.
func (im *Image) SetImage(img engine.Image) {
	im.nativeImg = img
	im.region = sg.TextureRegion{}
	im.hasImage = img != nil
	if img != nil {
		im.imgNode.SetCustomImage(img)
	} else {
		im.imgNode.SetCustomImage(nil)
	}
	im.imgNode.SetTextureRegion(sg.TextureRegion{})
	im.imgNode.SetVisible(im.hasImage)
	im.rebuildLayout()
}

// ClearImage removes the current image source, leaving the widget empty.
func (im *Image) ClearImage() {
	im.region = sg.TextureRegion{}
	im.nativeImg = nil
	im.hasImage = false
	im.imgNode.SetTextureRegion(sg.TextureRegion{})
	im.imgNode.SetCustomImage(nil)
	im.imgNode.SetVisible(false)
	im.tileCanvas = nil
	im.MarkDrawDirty()
}

// --- Layout ---

// SetScaleMode sets how the image is scaled within the widget bounds.
func (im *Image) SetScaleMode(mode ImageScaleMode) {
	im.scaleMode = mode
	im.rebuildLayout()
}

// ScaleMode returns the current scale mode.
func (im *Image) ScaleMode() ImageScaleMode { return im.scaleMode }

// --- Appearance ---

// SetTint sets the color multiplied over the image.
func (im *Image) SetTint(c sg.Color) {
	im.tint = c
	im.imgNode.SetColor(c)
	im.MarkDrawDirty()
}

// Tint returns the current tint color.
func (im *Image) Tint() sg.Color { return im.tint }

// SetAlpha sets the alpha of the tint (convenience wrapper).
func (im *Image) SetAlpha(a float32) {
	im.tint = sg.RGBA(im.tint.R(), im.tint.G(), im.tint.B(), float64(a))
	im.imgNode.SetColor(im.tint)
	im.MarkDrawDirty()
}

// SetCornerRadius sets the corner rounding. -1 means full pill (half the
// shorter dimension). 0 means sharp corners.
func (im *Image) SetCornerRadius(r float64) {
	cr := r
	if cr < 0 {
		cr = math.Min(im.Width, im.Height) / 2
	}
	im.applyCornerRadius(cr)
	im.MarkDrawDirty()
}

// --- Sizing ---

// SetSize sets the widget dimensions.
func (im *Image) SetSize(w, h float64) {
	im.Width = w
	im.Height = h
	im.resizeBackground(w, h)
	im.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	im.rebuildLayout()
	im.MarkLayoutDirty()
}

// SizeToContent resizes the widget to the image's native pixel dimensions.
// Has no effect if no image is set.
func (im *Image) SizeToContent() {
	nw, nh := im.nativeDimensions()
	if nw <= 0 || nh <= 0 {
		return
	}
	im.SetSize(float64(nw), float64(nh))
}

// UpdateVisuals applies theme colors based on current state.
func (im *Image) UpdateVisuals() {
	im.state = computeState(im.enabled, im.focused, im.hovered, im.pressed)
	group := im.EffectiveTheme().Image.Group(im.Variant())

	bg := group.Background.Resolve(im.state)
	im.applyBackground(bg)

	cr := resolveCornerRadius(group.CornerRadius, math.Min(im.Width, im.Height))
	im.applyCornerRadius(cr)

	im.MarkDrawDirty()
}

// --- Internal ---

// nativeDimensions returns the native pixel size of the current image source.
func (im *Image) nativeDimensions() (w, h int) {
	if im.nativeImg != nil {
		b := im.nativeImg.Bounds()
		return b.Dx(), b.Dy()
	}
	if im.region != (sg.TextureRegion{}) {
		return int(im.region.Width), int(im.region.Height)
	}
	return 0, 0
}

// rebuildLayout repositions and scales imgNode according to the current
// scale mode and widget size.
func (im *Image) rebuildLayout() {
	if !im.hasImage {
		im.MarkDrawDirty()
		return
	}

	imgW, imgH := im.nativeDimensions()
	if imgW <= 0 || imgH <= 0 {
		im.MarkDrawDirty()
		return
	}

	widW := im.Width
	widH := im.Height
	fImgW := float64(imgW)
	fImgH := float64(imgH)

	var applyW, applyH, applyX, applyY float64

	switch im.scaleMode {
	case ImageScaleStretch:
		applyW, applyH = widW, widH
		applyX, applyY = 0, 0

	case ImageScaleFit:
		s := math.Min(widW/fImgW, widH/fImgH)
		applyW = fImgW * s
		applyH = fImgH * s
		applyX = (widW - applyW) / 2
		applyY = (widH - applyH) / 2

	case ImageScaleFill:
		s := math.Max(widW/fImgW, widH/fImgH)
		applyW = fImgW * s
		applyH = fImgH * s
		applyX = (widW - applyW) / 2
		applyY = (widH - applyH) / 2
		im.rebuildClipCanvas(applyX, applyY, applyW, applyH)
		im.MarkDrawDirty()
		return

	case ImageScaleCenter:
		applyW, applyH = fImgW, fImgH
		applyX = (widW - fImgW) / 2
		applyY = (widH - fImgH) / 2
		im.rebuildClipCanvas(applyX, applyY, applyW, applyH)
		im.MarkDrawDirty()
		return

	case ImageScaleTile:
		im.rebuildTileCanvas(imgW, imgH)
		im.MarkDrawDirty()
		return
	}

	im.imgW, im.imgH = applyW, applyH
	im.imgX, im.imgY = applyX, applyY
	im.imgNode.SetScale(applyW/fImgW, applyH/fImgH)
	im.imgNode.SetPosition(applyX, applyY)
	im.MarkDrawDirty()
}

// rebuildClipCanvas pre-renders the image at (drawX, drawY) with size
// (drawW, drawH) onto a widget-sized canvas, providing pixel-exact clipping
// for Fill and Center modes. Only works for engine.Image sources.
func (im *Image) rebuildClipCanvas(drawX, drawY, drawW, drawH float64) {
	canvasW := int(math.Ceil(im.Width))
	canvasH := int(math.Ceil(im.Height))
	if canvasW <= 0 || canvasH <= 0 {
		return
	}
	if im.nativeImg == nil {
		return
	}

	if im.tileCanvas == nil ||
		im.tileCanvas.Bounds().Dx() != canvasW ||
		im.tileCanvas.Bounds().Dy() != canvasH {
		im.tileCanvas = engine.NewImage(canvasW, canvasH)
	} else {
		im.tileCanvas.Clear()
	}

	srcW := float64(im.nativeImg.Bounds().Dx())
	srcH := float64(im.nativeImg.Bounds().Dy())
	var op engine.DrawImageOptions
	op.GeoM.Scale(drawW/srcW, drawH/srcH)
	op.GeoM.Translate(drawX, drawY)
	im.tileCanvas.DrawImage(im.nativeImg, &op)

	im.imgW = float64(canvasW)
	im.imgH = float64(canvasH)
	im.imgX = 0
	im.imgY = 0
	im.imgNode.SetCustomImage(im.tileCanvas)
	im.imgNode.SetTextureRegion(sg.TextureRegion{})
	im.imgNode.SetScale(1, 1)
	im.imgNode.SetPosition(0, 0)
}

// rebuildTileCanvas creates a pre-composited engine.Image of widget size
// with the source image tiled across it, then assigns it to imgNode.
// Only works for engine.Image sources; TextureRegion sources are a no-op.
func (im *Image) rebuildTileCanvas(imgW, imgH int) {
	canvasW := int(math.Ceil(im.Width))
	canvasH := int(math.Ceil(im.Height))
	if canvasW <= 0 || canvasH <= 0 {
		return
	}

	if im.nativeImg == nil {
		return
	}

	// Reuse canvas if dimensions match.
	if im.tileCanvas == nil ||
		im.tileCanvas.Bounds().Dx() != canvasW ||
		im.tileCanvas.Bounds().Dy() != canvasH {
		im.tileCanvas = engine.NewImage(canvasW, canvasH)
	} else {
		im.tileCanvas.Clear()
	}

	var op engine.DrawImageOptions
	for y := 0; y < canvasH; y += imgH {
		for x := 0; x < canvasW; x += imgW {
			op.GeoM.Reset()
			op.GeoM.Translate(float64(x), float64(y))
			im.tileCanvas.DrawImage(im.nativeImg, &op)
		}
	}

	im.imgW = float64(canvasW)
	im.imgH = float64(canvasH)
	im.imgX = 0
	im.imgY = 0
	im.imgNode.SetCustomImage(im.tileCanvas)
	im.imgNode.SetTextureRegion(sg.TextureRegion{})
	im.imgNode.SetScale(1, 1) // tile canvas is already canvas-sized; render at native 1:1
	im.imgNode.SetPosition(0, 0)
}

// ImgNode returns the internal image sprite node. Used for testing.
func (im *Image) ImgNode() *sg.Node { return im.imgNode }

// ImgSize returns the pixel-space width and height applied to the image sprite.
// Used for testing layout calculations.
func (im *Image) ImgSize() (w, h float64) { return im.imgW, im.imgH }

// ImgPosition returns the pixel-space x/y position applied to the image sprite.
// Used for testing layout calculations.
func (im *Image) ImgPosition() (x, y float64) { return im.imgX, im.imgY }
