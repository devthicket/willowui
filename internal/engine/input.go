package engine

import "github.com/hajimehoshi/ebiten/v2"

func IsKeyPressed(k Key) bool            { return ebiten.IsKeyPressed(k) }
func CursorPosition() (int, int)         { return ebiten.CursorPosition() }
func Wheel() (float64, float64)          { return ebiten.Wheel() }
func AppendInputChars(buf []rune) []rune { return ebiten.AppendInputChars(buf) }
func WindowSize() (int, int)             { return ebiten.WindowSize() }
