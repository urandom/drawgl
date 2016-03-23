package interpolator

import "github.com/urandom/drawgl"

type nearestNeighbor struct{}

// Copy from golang.org/x/image/draw
// Copyright (c) 2009 The Go Authors. All rights reserved.
func (nearestNeighbor) Get(src *drawgl.FloatImage, fx, fy float64) drawgl.FloatColor {
	b := src.Bounds()
	pi := (int(fy)-b.Min.Y)*src.Stride + (int(fx)-b.Min.X)*4

	return drawgl.FloatColor{
		R: src.Pix[pi+0],
		G: src.Pix[pi+1],
		B: src.Pix[pi+2],
		A: src.Pix[pi+3],
	}
}
