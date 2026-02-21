package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/vector"
)

func StrokeLine(dst Image, x0, y0, x1, y1, strokeWidth float32, clr color.Color, antialias bool) {
	vector.StrokeLine(dst, x0, y0, x1, y1, strokeWidth, clr, antialias)
}
