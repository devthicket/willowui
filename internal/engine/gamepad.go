package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type GamepadID = ebiten.GamepadID
type GamepadButton = ebiten.GamepadButton

const GamepadButtonMax = ebiten.GamepadButtonMax

func AppendGamepadIDs(buf []GamepadID) []GamepadID { return ebiten.AppendGamepadIDs(buf) }
func IsGamepadButtonJustPressed(id GamepadID, btn GamepadButton) bool {
	return inpututil.IsGamepadButtonJustPressed(id, btn)
}
