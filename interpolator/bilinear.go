package interpolator

import "github.com/urandom/drawgl"

type bilinear struct {
}

func (i bilinear) Get(src *drawgl.FloatImage, fx, fy float64, edgeHandler drawgl.EdgeHandler) (dstC drawgl.FloatColor) {
	x, y := int(fx), int(fy)
	b := src.Bounds()

	tlX, tlY := drawgl.TranslateCoords(x, y, b, edgeHandler)
	trX, trY := drawgl.TranslateCoords(x+1, y, b, edgeHandler)
	blX, blY := drawgl.TranslateCoords(x, y+1, b, edgeHandler)
	brX, brY := drawgl.TranslateCoords(x+1, y+1, b, edgeHandler)

	tlI := (tlY-b.Min.Y)*src.Stride + (tlX-b.Min.X)*4
	trI := (trY-b.Min.Y)*src.Stride + (trX-b.Min.X)*4
	blI := (blY-b.Min.Y)*src.Stride + (blX-b.Min.X)*4
	brI := (brY-b.Min.Y)*src.Stride + (brX-b.Min.X)*4

	ix, iy := drawgl.ColorValue(fx-float64(x)), drawgl.ColorValue(fy-float64(y))

	botRightW := ix * iy
	botLeftW := iy - botRightW
	topRightW := ix - botRightW
	topLeftW := 1 - (ix - botLeftW)

	dstC.R = topLeftW*src.Pix[tlI] + topRightW*src.Pix[trI] + botLeftW*src.Pix[blI] + botRightW*src.Pix[brI]
	dstC.G = topLeftW*src.Pix[tlI+1] + topRightW*src.Pix[trI+1] + botLeftW*src.Pix[blI+1] + botRightW*src.Pix[brI+1]
	dstC.B = topLeftW*src.Pix[tlI+2] + topRightW*src.Pix[trI+2] + botLeftW*src.Pix[blI+2] + botRightW*src.Pix[brI+2]
	dstC.A = topLeftW*src.Pix[tlI+3] + topRightW*src.Pix[trI+3] + botLeftW*src.Pix[blI+3] + botRightW*src.Pix[brI+3]

	return
}
