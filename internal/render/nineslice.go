package render

import "github.com/devthicket/willowui/internal/sg"

// NineSliceNodes holds the 9 sprite nodes that make up a nine-slice background.
// The naming convention follows CSS-style corners and edges:
//
//	TL  T  TR
//	L   C   R
//	BL  B  BR
type NineSliceNodes struct {
	TL, T, TR *sg.Node
	L, C, R   *sg.Node
	BL, B, BR *sg.Node
}

// SubRegion computes a sub-region within a base TextureRegion.
// x, y, w, h are in pixel coordinates relative to the base region's origin.
func SubRegion(base sg.TextureRegion, x, y, w, h uint16) sg.TextureRegion {
	return sg.TextureRegion{
		Page:      base.Page,
		X:         base.X + x,
		Y:         base.Y + y,
		Width:     w,
		Height:    h,
		OriginalW: w,
		OriginalH: h,
	}
}

// CreateNineSliceNodes creates 9 sprite nodes as children of the given
// container, each using the correct sub-region of the nine-slice image.
func CreateNineSliceNodes(name string, container *sg.Node, ns *NineSlice) *NineSliceNodes {
	base := ns.Region
	inL := uint16(ns.Insets.Left)
	inR := uint16(ns.Insets.Right)
	inT := uint16(ns.Insets.Top)
	inB := uint16(ns.Insets.Bottom)
	midW := base.Width - inL - inR
	midH := base.Height - inT - inB

	sprite := func(suffix string, r sg.TextureRegion) *sg.Node {
		s := sg.NewSprite(name+suffix, r)
		container.AddChild(s)
		return s
	}

	nodes := &NineSliceNodes{
		TL: sprite("-tl", SubRegion(base, 0, 0, inL, inT)),
		T:  sprite("-t", SubRegion(base, inL, 0, midW, inT)),
		TR: sprite("-tr", SubRegion(base, inL+midW, 0, inR, inT)),

		L: sprite("-l", SubRegion(base, 0, inT, inL, midH)),
		C: sprite("-c", SubRegion(base, inL, inT, midW, midH)),
		R: sprite("-r", SubRegion(base, inL+midW, inT, inR, midH)),

		BL: sprite("-bl", SubRegion(base, 0, inT+midH, inL, inB)),
		B:  sprite("-b", SubRegion(base, inL, inT+midH, midW, inB)),
		BR: sprite("-br", SubRegion(base, inL+midW, inT+midH, inR, inB)),
	}

	return nodes
}

// LayoutNineSlice positions and scales the 9 sprites to fill the given
// dimensions (w, h). Corners keep their natural pixel size; edges stretch
// along one axis; center stretches both.
func LayoutNineSlice(nodes *NineSliceNodes, ns *NineSlice, w, h float64) {
	inL := ns.Insets.Left
	inR := ns.Insets.Right
	inT := ns.Insets.Top
	inB := ns.Insets.Bottom

	// Clamp: if the component is smaller than the insets, shrink uniformly.
	if w < inL+inR {
		scale := w / (inL + inR)
		inL *= scale
		inR *= scale
	}
	if h < inT+inB {
		scale := h / (inT + inB)
		inT *= scale
		inB *= scale
	}

	midW := w - inL - inR
	midH := h - inT - inB

	base := ns.Region
	srcInL := float64(uint16(ns.Insets.Left))
	srcInR := float64(uint16(ns.Insets.Right))
	srcInT := float64(uint16(ns.Insets.Top))
	srcInB := float64(uint16(ns.Insets.Bottom))
	srcMidW := float64(base.Width) - srcInL - srcInR
	srcMidH := float64(base.Height) - srcInT - srcInB

	// Corners: position at corners, scale to destination inset sizes.
	nodes.TL.SetPosition(0, 0)
	nodes.TL.SetScale(inL/srcInL, inT/srcInT)

	nodes.TR.SetPosition(w-inR, 0)
	nodes.TR.SetScale(inR/srcInR, inT/srcInT)

	nodes.BL.SetPosition(0, h-inB)
	nodes.BL.SetScale(inL/srcInL, inB/srcInB)

	nodes.BR.SetPosition(w-inR, h-inB)
	nodes.BR.SetScale(inR/srcInR, inB/srcInB)

	// Horizontal edges: stretch X, natural Y.
	scaleX := 1.0
	if srcMidW > 0 {
		scaleX = midW / srcMidW
	}

	nodes.T.SetPosition(inL, 0)
	nodes.T.SetScale(scaleX, inT/srcInT)

	nodes.B.SetPosition(inL, h-inB)
	nodes.B.SetScale(scaleX, inB/srcInB)

	// Vertical edges: natural X, stretch Y.
	scaleY := 1.0
	if srcMidH > 0 {
		scaleY = midH / srcMidH
	}

	nodes.L.SetPosition(0, inT)
	nodes.L.SetScale(inL/srcInL, scaleY)

	nodes.R.SetPosition(w-inR, inT)
	nodes.R.SetScale(inR/srcInR, scaleY)

	// Center: stretch both (hidden when CenterFill is used).
	nodes.C.SetPosition(inL, inT)
	nodes.C.SetScale(scaleX, scaleY)
	if ns.CenterFill != nil {
		nodes.C.SetVisible(false)
	}
}
