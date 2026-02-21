package sg

import (
	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
)

type Atlas = willow.Atlas

func LoadAtlas(jsonData []byte, pages []*ebiten.Image) (*Atlas, error) {
	return willow.LoadAtlas(jsonData, pages)
}

func NewBatchAtlas() *Atlas { return willow.NewBatchAtlas() }
