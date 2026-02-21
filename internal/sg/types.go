package sg

import "github.com/devthicket/willow"

type Color = willow.Color
type Vec2 = willow.Vec2
type TextureRegion = willow.TextureRegion
type HitRect = willow.HitRect
type HitShape = willow.HitShape
type MouseButton = willow.MouseButton
type TextAlign = willow.TextAlign
type FXAAConfig = willow.FXAAConfig
type RunConfig = willow.RunConfig
type EaseFunc = willow.EaseFunc

const (
	MouseButtonLeft  = willow.MouseButtonLeft
	MouseButtonRight = willow.MouseButtonRight
	ModShift         = willow.ModShift
	TextAlignLeft    = willow.TextAlignLeft
	TextAlignCenter  = willow.TextAlignCenter
	TextAlignRight   = willow.TextAlignRight
)

var (
	WhitePixel        = willow.WhitePixel
	DefaultFXAAConfig = willow.DefaultFXAAConfig
)

func RGBA(r, g, b, a float64) Color      { return willow.RGBA(r, g, b, a) }
func ColorFromHSV(h, s, v float64) Color { return willow.ColorFromHSV(h, s, v) }
