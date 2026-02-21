package engine

import "github.com/hajimehoshi/ebiten/v2/text/v2"

type GoTextFaceSource = text.GoTextFaceSource
type GoTextFace = text.GoTextFace
type DrawOptions = text.DrawOptions

func NewGoTextFaceSource(r interface{ Read([]byte) (int, error) }) (*GoTextFaceSource, error) {
	return text.NewGoTextFaceSource(r)
}

func TextAdvance(s string, face *GoTextFace) float64 {
	return text.Advance(s, face)
}

func TextDraw(dst Image, s string, face *GoTextFace, op *DrawOptions) {
	text.Draw(dst, s, face, op)
}
