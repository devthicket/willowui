package sg

import "github.com/devthicket/willow"

type Scene = willow.Scene

func NewScene() *Scene                         { return willow.NewScene() }
func Run(scene *Scene, config RunConfig) error { return willow.Run(scene, config) }
