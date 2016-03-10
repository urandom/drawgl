package transform

import (
	"image"
	"image/draw"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/interpolator"
	"github.com/urandom/drawgl/operation/transform/matrix"
)

type transformOperation struct {
	matrix       matrix.Matrix3
	interpolator interpolator.Interpolator
	dstB         image.Rectangle
}

func affine(op transformOperation, src *drawgl.FloatImage, mask drawgl.Mask, channel drawgl.Channel, forceLinear bool) (dst *drawgl.FloatImage) {
	srcB := src.Bounds()
	dstB := op.dstB

	if dstB.Empty() {
		dstB = srcB
	}

	dst = drawgl.NewFloatImage(dstB)

	edgeHandler := drawgl.Transparent

	it := drawgl.DefaultRectangleIterator(dstB, forceLinear)
	it.Iterate(mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		sx := float64(pt.X)*op.matrix[0][0] + float64(pt.Y)*op.matrix[0][1] + op.matrix[0][2]
		sy := float64(pt.X)*op.matrix[1][0] + float64(pt.Y)*op.matrix[1][1] + op.matrix[1][2]

		orig := src.FloatAt(pt.X, pt.Y)
		srcC := op.interpolator.Get(src, sx, sy, edgeHandler)

		dst.UnsafeSetColor(pt.X, pt.Y, drawgl.MaskColor(orig, srcC, channel, f, draw.Over))
	})

	return
}
