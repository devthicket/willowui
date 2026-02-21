package sg

import (
	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
)

type Node = willow.Node
type Vertex = ebiten.Vertex

func NewContainer(name string) *Node                    { return willow.NewContainer(name) }
func NewSprite(name string, region TextureRegion) *Node { return willow.NewSprite(name, region) }
func NewText(name, content string, font *FontFamily) *Node {
	return willow.NewText(name, content, font)
}
func NewPolygon(name string, pts []Vec2) *Node { return willow.NewPolygon(name, pts) }
func SetPolygonPoints(n *Node, pts []Vec2)     { willow.SetPolygonPoints(n, pts) }

func NewMesh(name string, img *ebiten.Image, verts []Vertex, inds []uint16) *Node {
	return willow.NewMesh(name, img, verts, inds)
}
