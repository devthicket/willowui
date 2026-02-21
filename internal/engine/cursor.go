package engine

import "github.com/hajimehoshi/ebiten/v2"

type CursorShapeType = ebiten.CursorShapeType

const (
	CursorShapeDefault    = ebiten.CursorShapeDefault
	CursorShapePointer    = ebiten.CursorShapePointer
	CursorShapeText       = ebiten.CursorShapeText
	CursorShapeNSResize   = ebiten.CursorShapeNSResize
	CursorShapeEWResize   = ebiten.CursorShapeEWResize
	CursorShapeNWSEResize = ebiten.CursorShapeNWSEResize
)

func SetCursorShape(shape CursorShapeType) { ebiten.SetCursorShape(shape) }
