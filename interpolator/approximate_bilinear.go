package interpolator

import (
	"image"

	"github.com/urandom/drawgl"
)

type approximateBilinear struct {
	bias image.Point
}

func newApproximageBilinear(bias image.Point) approximateBilinear {
	return approximateBilinear{bias: bias}
}

// Copy from golang.org/x/image/draw
// Copyright (c) 2009 The Go Authors. All rights reserved.
func (i approximateBilinear) Get(src *drawgl.FloatImage, fx, fy float64) drawgl.FloatColor {
	b := src.Bounds()
	fx -= 0.5
	fy -= 0.5

	ix0 := int(fx)
	iy0 := int(fy)

	xFrac0 := drawgl.ColorValue(fx - float64(ix0))
	xFrac1 := drawgl.ColorValue(1 - xFrac0)
	yFrac0 := drawgl.ColorValue(fy - float64(iy0))
	yFrac1 := drawgl.ColorValue(1 - yFrac0)

	ix0 += i.bias.X
	ix1 := ix0 + 1
	if ix0 < b.Min.X {
		ix0, ix1 = b.Min.X, b.Min.X
	} else if ix1 >= b.Max.X {
		ix0, ix1 = b.Max.X-1, b.Max.X-1
		xFrac0, xFrac1 = 1, 0
	}

	iy0 += i.bias.Y
	iy1 := iy0 + 1
	if iy0 < b.Min.Y {
		iy0, iy1 = b.Min.Y, b.Min.Y
	} else if iy1 >= b.Max.Y {
		iy0, iy1 = b.Max.Y-1, b.Max.Y-1
		yFrac0, yFrac1 = 1, 0
	}

	s00i := (iy0-b.Min.Y)*src.Stride + (ix0-b.Min.X)*4
	s00r := src.Pix[s00i+0]
	s00g := src.Pix[s00i+1]
	s00b := src.Pix[s00i+2]
	s00a := src.Pix[s00i+3]

	s10i := (iy0-b.Min.Y)*src.Stride + (ix1-b.Min.X)*4
	s10r := xFrac0*src.Pix[s10i+0] + xFrac1*s00r
	s10g := xFrac0*src.Pix[s10i+1] + xFrac1*s00g
	s10b := xFrac0*src.Pix[s10i+2] + xFrac1*s00b
	s10a := xFrac0*src.Pix[s10i+3] + xFrac1*s00a

	s01i := (iy1-b.Min.Y)*src.Stride + (ix0-b.Min.X)*4
	s01r := src.Pix[s01i+0]
	s01g := src.Pix[s01i+1]
	s01b := src.Pix[s01i+2]
	s01a := src.Pix[s01i+3]

	s11i := (iy1-b.Min.Y)*src.Stride + (ix1-b.Min.X)*4
	s11r := xFrac0*src.Pix[s11i+0] + xFrac1*s01r
	s11g := xFrac0*src.Pix[s11i+1] + xFrac1*s01g
	s11b := xFrac0*src.Pix[s11i+2] + xFrac1*s01b
	s11a := xFrac0*src.Pix[s11i+3] + xFrac1*s01a

	return drawgl.FloatColor{
		R: yFrac0*s11r + yFrac1*s10r,
		G: yFrac0*s11g + yFrac1*s10g,
		B: yFrac0*s11b + yFrac1*s10b,
		A: yFrac0*s11a + yFrac1*s10a,
	}
}
