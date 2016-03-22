package interpolator

import (
	"math"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/transform/matrix"
)

type kernel struct {
	// Support is the kernel support and must be >= 0. At(t) is assumed to be
	// zero when t >= Support.
	Support float64
	// At is the kernel function. It will only be called with t in the
	// range [0, Support).
	At func(t float64) float64

	halfWidth, kernelArgScale [2]float64
	weights                   [2][]float64
}

func newKernel(
	src *drawgl.FloatImage,
	m matrix.Matrix3,
) kernel {
	i := kernel{}

	xscale := abs(m[0][0])
	if s := abs(m[0][1]); xscale < s {
		xscale = s
	}

	yscale := abs(m[1][0])
	if s := abs(m[1][1]); yscale < s {
		yscale = s
	}

	i.halfWidth[0], i.halfWidth[1] = i.Support, i.Support
	i.kernelArgScale[0], i.kernelArgScale[1] = 1.0, 1.0

	if xscale > 1 {
		i.halfWidth[0] *= xscale
		i.kernelArgScale[0] = 1 / xscale
	}
	if yscale > 1 {
		i.halfWidth[1] *= yscale
		i.kernelArgScale[1] = 1 / yscale
	}

	i.weights[0] = make([]float64, 1+2*int(math.Ceil(i.halfWidth[0])))
	i.weights[1] = make([]float64, 1+2*int(math.Ceil(i.halfWidth[1])))

	return i
}

// Copy from golang.org/x/image/draw
// Copyright (c) 2009 The Go Authors. All rights reserved.
func (i kernel) Get(src *drawgl.FloatImage, fx, fy float64) drawgl.FloatColor {
	b := src.Bounds()

	totalWeights := [2]float64{}

	fx -= 0.5
	ix := int(math.Floor(fx - i.halfWidth[0]))

	if ix < b.Min.X {
		ix = b.Min.X
	}

	jx := int(math.Ceil(fx + i.halfWidth[1]))
	if jx > b.Max.X {
		jx = b.Max.X
	}

	for kx := ix; kx < jx; kx++ {
		w := 0.0
		if t := abs((fx - float64(kx)) * i.kernelArgScale[0]); t < i.Support {
			w = i.At(t)
		}
		i.weights[0][kx-ix] = w
		totalWeights[0] += w
	}

	for x := range i.weights[0][:jx-ix] {
		i.weights[0][x] /= totalWeights[0]
	}

	fy -= 0.5
	iy := int(math.Floor(fy - i.halfWidth[1]))
	if iy < b.Min.Y {
		iy = b.Min.Y
	}

	jy := int(math.Ceil(fy + i.halfWidth[1]))
	if jy > b.Max.Y {
		jy = b.Max.Y
	}

	for ky := iy; ky < jy; ky++ {
		w := 0.0
		if t := abs((fy - float64(ky)) * i.kernelArgScale[1]); t < i.Support {
			w = i.At(t)
		}

		i.weights[1][ky-iy] = w
		totalWeights[1] += w
	}

	for y := range i.weights[1][:jy-iy] {
		i.weights[1][y] /= totalWeights[1]
	}

	var pr, pg, pb, pa drawgl.ColorValue
	for ky := iy; ky < jy; ky++ {
		if yw := i.weights[1][ky-iy]; yw != 0 {
			for kx := ix; kx < jx; kx++ {
				if xw := drawgl.ColorValue(i.weights[0][kx-ix] * yw); xw != 0 {
					pi := (ky-b.Min.Y)*src.Stride + (kx-b.Min.X)*4
					pr += src.Pix[pi+0] * xw
					pg += src.Pix[pi+1] * xw
					pb += src.Pix[pi+2] * xw
					pa += src.Pix[pi+3] * xw
				}
			}
		}
	}

	if pr > pa {
		pr = pa
	}

	if pg > pa {
		pg = pa
	}

	if pb > pa {
		pb = pa
	}

	return drawgl.FloatColor{R: pr, G: pg, B: pb, A: pa}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
