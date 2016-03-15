package interpolator

import "github.com/urandom/drawgl"

type bilinear struct {
}

func (i bilinear) Get(src *drawgl.FloatImage, fx, fy float64) (dstC drawgl.FloatColor) {
	/*
		x, y := int(fx), int(fy)
		b := src.Bounds()

		ix, iy := drawgl.ColorValue(fx-float64(x)), drawgl.ColorValue(fy-float64(y))
		if ix == 0 && iy == 0 {
			dstC = src.FloatAt(x, y)
			return
		}

		var tl, tr, bl, br [4]drawgl.ColorValue

		tlX, tlY, err := drawgl.TranslateCoords(x, y, b, edgeHandler)
		if err == nil {
			tlI := (tlY-b.Min.Y)*src.Stride + (tlX-b.Min.X)*4
			tl = [4]drawgl.ColorValue{src.Pix[tlI],
				src.Pix[tlI+1], src.Pix[tlI+2], src.Pix[tlI+3]}
		} else if err == drawgl.ErrOutOfBounds {
			tl = [4]drawgl.ColorValue{0, 0, 0, 0}
		} // No other kind of error exists in this context

		trX, trY, err := drawgl.TranslateCoords(x+1, y, b, edgeHandler)
		if err == nil {
			trI := (trY-b.Min.Y)*src.Stride + (trX-b.Min.X)*4
			tr = [4]drawgl.ColorValue{src.Pix[trI],
				src.Pix[trI+1], src.Pix[trI+2], src.Pix[trI+3]}
		} else if err == drawgl.ErrOutOfBounds {
			tr = [4]drawgl.ColorValue{0, 0, 0, 0}
		}

		blX, blY, err := drawgl.TranslateCoords(x, y+1, b, edgeHandler)
		if err == nil {
			blI := (blY-b.Min.Y)*src.Stride + (blX-b.Min.X)*4
			bl = [4]drawgl.ColorValue{src.Pix[blI],
				src.Pix[blI+1], src.Pix[blI+2], src.Pix[blI+3]}
		} else if err == drawgl.ErrOutOfBounds {
			bl = [4]drawgl.ColorValue{0, 0, 0, 0}
		}

		brX, brY, err := drawgl.TranslateCoords(x+1, y+1, b, edgeHandler)
		if err == nil {
			brI := (brY-b.Min.Y)*src.Stride + (brX-b.Min.X)*4
			br = [4]drawgl.ColorValue{src.Pix[brI],
				src.Pix[brI+1], src.Pix[brI+2], src.Pix[brI+3]}
		} else if err == drawgl.ErrOutOfBounds {
			br = [4]drawgl.ColorValue{0, 0, 0, 0}
		}

		botRightW := ix * iy
		botLeftW := iy - botRightW
		topRightW := ix - botRightW
		topLeftW := 1 - (ix - botLeftW)

		dstC.R = topLeftW*tl[0] + topRightW*tr[0] + botLeftW*bl[0] + botRightW*br[0]
		dstC.G = topLeftW*tl[1] + topRightW*tr[1] + botLeftW*bl[1] + botRightW*br[1]
		dstC.B = topLeftW*tl[2] + topRightW*tr[2] + botLeftW*bl[2] + botRightW*br[2]
		dstC.A = topLeftW*tl[3] + topRightW*tr[3] + botLeftW*bl[3] + botRightW*br[3]

	*/
	return
}
