package transform

import (
	"image"
	"image/draw"
	"math"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/interpolator"
	"github.com/urandom/drawgl/operation/transform/matrix"
)

type transformOperation struct {
	matrix       matrix.Matrix3
	interpolator string
	dstB         image.Rectangle
}

func affine(op transformOperation, src *drawgl.FloatImage, mask drawgl.Mask, channel drawgl.Channel, forceLinear bool) (dst *drawgl.FloatImage) {
	if op.matrix.IsIdentity() {
		dst = drawgl.CopyImage(src)
		return
	}

	srcB := src.Bounds()
	dstB := op.dstB
	if dstB.Empty() {
		dstB = srcB
	}

	adr := srcB.Intersect(affineTransformRect(op.matrix, srcB))
	dst = drawgl.NewFloatImage(dstB)

	if adr.Empty() || srcB.Empty() {
		return
	}

	inverse := op.matrix
	inverse.Invert()

	bias := affineTransformRect(inverse, adr).Min
	bias.X--
	bias.Y--
	inverse[0][2] -= float64(bias.X)
	inverse[1][2] -= float64(bias.Y)

	interpolator := interpolator.New(op.interpolator, src, inverse, bias)

	it := drawgl.DefaultRectangleIterator(adr, forceLinear)
	it.Iterate(mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		dx := float64(dstB.Min.X+pt.X) + 0.5
		dy := float64(dstB.Min.Y+pt.Y) + 0.5

		sx := inverse[0][0]*dx + inverse[0][1]*dy + inverse[0][2]
		sy := inverse[1][0]*dx + inverse[1][1]*dy + inverse[1][2]

		if !(image.Point{int(sx) + bias.X, int(sy) + bias.Y}).In(srcB) {
			return
		}

		sx += float64(bias.X)
		sy += float64(bias.Y)

		orig := src.FloatAt(pt.X, pt.Y)
		srcC := interpolator.Get(src, sx, sy)

		dst.UnsafeSetColor(pt.X, pt.Y, drawgl.MaskColor(orig, srcC, channel, f, draw.Over))
	})

	return
}

func affineTransformRect(m matrix.Matrix3, sr image.Rectangle) (dr image.Rectangle) {
	ps := [...]image.Point{
		{sr.Min.X, sr.Min.Y},
		{sr.Max.X, sr.Min.Y},
		{sr.Min.X, sr.Max.Y},
		{sr.Max.X, sr.Max.Y},
	}
	for i, p := range ps {
		sxf := float64(p.X)
		syf := float64(p.Y)
		dx := int(math.Floor(m[0][0]*sxf + m[0][1]*syf + m[0][2]))
		dy := int(math.Floor(m[1][0]*sxf + m[1][1]*syf + m[1][2]))

		// The +1 adjustments below are because an image.Rectangle is inclusive
		// on the low end but exclusive on the high end.

		if i == 0 {
			dr = image.Rectangle{
				Min: image.Point{dx + 0, dy + 0},
				Max: image.Point{dx + 1, dy + 1},
			}
			continue
		}

		if dr.Min.X > dx {
			dr.Min.X = dx
		}
		dx++
		if dr.Max.X < dx {
			dr.Max.X = dx
		}

		if dr.Min.Y > dy {
			dr.Min.Y = dy
		}
		dy++
		if dr.Max.Y < dy {
			dr.Max.Y = dy
		}
	}
	return
}
