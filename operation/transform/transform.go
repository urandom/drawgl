package transform

import (
	"image"
	"image/draw"

	"github.com/urandom/drawgl"
)

type transformOp int

const (
	_                          = iota
	transformFlipH transformOp = iota
	transformFlipV
	transformTranspose
	transformTransverse
	transformRotate90
	transformRotate180
	transformRotate270
)

func transform(op transformOp, src *drawgl.FloatImage, it drawgl.RectangleIterator, mask drawgl.Mask, channel drawgl.Channel) (dst *drawgl.FloatImage) {
	b := src.Bounds()

	var offset int

	switch op {
	case transformFlipH:
		offset = b.Min.X + b.Max.X
	case transformFlipV:
		offset = b.Min.Y + b.Max.Y
	}

	dst = drawgl.NewFloatImage(b)

	it.Iterate(mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		var srcColor drawgl.FloatColor
		switch op {
		case transformFlipH:
			srcColor = src.UnsafeFloatAt(offset-pt.X-1, pt.Y)
		case transformFlipV:
			srcColor = src.UnsafeFloatAt(pt.X, offset-pt.Y-1)
		}

		dstColor := src.UnsafeFloatAt(pt.X, pt.Y)

		dst.UnsafeSetColor(pt.X, pt.Y, drawgl.MaskColor(dstColor, srcColor, channel, f, draw.Over))
	})

	return
}
