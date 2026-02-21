package widget

import (
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// nineSliceNodes is a package-private alias for render.NineSliceNodes.
type nineSliceNodes = render.NineSliceNodes

// NineSliceNodes is the public type for nine-slice node collections.
type NineSliceNodes = render.NineSliceNodes

// SubRegion is an exported wrapper for render.SubRegion. Used for testing.
func SubRegion(base sg.TextureRegion, x, y, w, h uint16) sg.TextureRegion {
	return render.SubRegion(base, x, y, w, h)
}

// CreateNineSliceNodes is an exported wrapper for render.CreateNineSliceNodes. Used for testing.
func CreateNineSliceNodes(name string, container *sg.Node, ns *NineSlice) *NineSliceNodes {
	return render.CreateNineSliceNodes(name, container, ns)
}

// LayoutNineSlice is an exported wrapper for render.LayoutNineSlice. Used for testing.
func LayoutNineSlice(nodes *NineSliceNodes, ns *NineSlice, w, h float64) {
	render.LayoutNineSlice(nodes, ns, w, h)
}
