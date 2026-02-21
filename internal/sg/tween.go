package sg

import "github.com/devthicket/willow"

type TweenGroup = willow.TweenGroup
type TweenConfig = willow.TweenConfig

func TweenAlpha(n *Node, alpha float64, config TweenConfig) *TweenGroup {
	return willow.TweenAlpha(n, alpha, config)
}

func TweenPosition(n *Node, toX, toY float64, config TweenConfig) *TweenGroup {
	return willow.TweenPosition(n, toX, toY, config)
}
