package engine

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Image = *ebiten.Image
type DrawImageOptions = ebiten.DrawImageOptions
type Blend = ebiten.Blend

var BlendDestinationOut = ebiten.BlendDestinationOut

func NewImage(w, h int) Image                          { return ebiten.NewImage(w, h) }
func NewImageFromImage(img image.Image) Image          { return ebiten.NewImageFromImage(img) }
func SubImage(src Image, r image.Rectangle) Image      { return src.SubImage(r).(Image) }
func ImageBounds(img Image) image.Rectangle            { return img.Bounds() }
func ImageFill(img Image, c color.Color)               { img.Fill(c) }
func ImageClear(img Image)                             { img.Clear() }
func ImageDraw(dst, src Image, opts *DrawImageOptions) { dst.DrawImage(src, opts) }
func ImageDeallocate(img Image)                        { img.Deallocate() }
func ImageSetPixel(img Image, x, y int, c color.Color) { img.Set(x, y, c) }
