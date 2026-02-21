package sg

import "github.com/devthicket/willow"

type FontFamily = willow.FontFamily
type FontFamilyConfig = willow.FontFamilyConfig
type TextBlock = willow.TextBlock

func NewFontFamilyFromFontBundle(data []byte) (*FontFamily, error) {
	return willow.NewFontFamilyFromFontBundle(data)
}

func NewFontFamilyFromTTF(config FontFamilyConfig) (*FontFamily, error) {
	return willow.NewFontFamilyFromTTF(config)
}
