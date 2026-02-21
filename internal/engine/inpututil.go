package engine

import "github.com/hajimehoshi/ebiten/v2/inpututil"

func IsKeyJustPressed(k Key) bool                 { return inpututil.IsKeyJustPressed(k) }
func IsKeyJustReleased(k Key) bool                { return inpututil.IsKeyJustReleased(k) }
func AppendJustPressedKeys(buf []Key) []Key       { return inpututil.AppendJustPressedKeys(buf) }
func IsMouseButtonJustPressed(b MouseButton) bool { return inpututil.IsMouseButtonJustPressed(b) }
