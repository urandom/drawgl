package transform

import (
	"errors"
	"image"
	"math"

	"github.com/urandom/drawgl"

	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

type InterpolatorOp string

type nnInterpolator struct{}
type ablInterpolator struct{}

// TODO
// Parallel NestedNeighbor and ApproxBiLinear
var (
	NearestNeighbor = nnInterpolator{}
	ApproxBiLinear  = ablInterpolator{}

	Lanczos = &draw.Kernel{3, func(t float64) float64 {
		if t < 0 {
			t = -t
		}

		if t < 3 {
			return sinc(t) / sinc(t/3)
		}

		return 0
	}}
)

func (i InterpolatorOp) Inst() (draw.Interpolator, error) {
	var interpolator draw.Interpolator

	if i == "" {
		interpolator = draw.ApproxBiLinear
	} else {
		switch i {
		case "NearestNeighbor":
			interpolator = NearestNeighbor
		case "ApproxBiLinear":
			interpolator = ApproxBiLinear
		case "BiLinear":
			interpolator = draw.BiLinear
		case "CatmullRom":
			interpolator = draw.CatmullRom
		case "Lanczos":
			interpolator = Lanczos
		default:
			return nil, errors.New("Unknown interpolator: " + string(i))
		}
	}

	return interpolator, nil
}

func (z nnInterpolator) Scale(dst draw.Image, dr image.Rectangle, src image.Image, sr image.Rectangle, op draw.Op, opts *draw.Options) {
	fsrc, sok := src.(*drawgl.FloatImage)
	fdst, dok := dst.(*drawgl.FloatImage)

	var o draw.Options
	if opts != nil {
		o = *opts
	}

	// Try to simplify a Scale to a Copy.
	if !sok || !dok || dr.Size() == sr.Size() ||
		o.DstMask != nil || o.SrcMask != nil || !sr.In(src.Bounds()) {
		draw.NearestNeighbor.Scale(dst, dr, src, sr, op, opts)
		return
	}

	// adr is the affected destination pixels.
	adr := dst.Bounds().Intersect(dr)
	if adr.Empty() || sr.Empty() {
		return
	}
	// Make adr relative to dr.Min.
	adr = adr.Sub(dr.Min)
	if op == draw.Over && o.SrcMask == nil && opaque(src) {
		op = draw.Src
	}

	switch op {
	case draw.Over:
		z.scaleOver(fdst, dr, adr, fsrc, sr, &o)
	case draw.Src:
		z.scaleSrc(fdst, dr, adr, fsrc, sr, &o)
	}
}

func (z nnInterpolator) Transform(dst draw.Image, s2d f64.Aff3, src image.Image, sr image.Rectangle, op draw.Op, opts *draw.Options) {
	fsrc, sok := src.(*drawgl.FloatImage)
	fdst, dok := dst.(*drawgl.FloatImage)

	var o draw.Options
	if opts != nil {
		o = *opts
	}

	if !sok || !dok || o.DstMask != nil || o.SrcMask != nil || !sr.In(src.Bounds()) {
		draw.NearestNeighbor.Transform(dst, s2d, src, sr, op, opts)
		return
	}

	// Try to simplify a Transform to a Copy.
	if s2d[0] == 1 && s2d[1] == 0 && s2d[3] == 0 && s2d[4] == 1 {
		dx := int(s2d[2])
		dy := int(s2d[5])
		if float64(dx) == s2d[2] && float64(dy) == s2d[5] {
			draw.Copy(dst, image.Point{X: sr.Min.X + dx, Y: sr.Min.X + dy}, src, sr, op, opts)
			return
		}
	}

	dr := transformRect(&s2d, &sr)
	// adr is the affected destination pixels.
	adr := dst.Bounds().Intersect(dr)
	if adr.Empty() || sr.Empty() {
		return
	}
	if op == draw.Over && o.SrcMask == nil && opaque(src) {
		op = draw.Src
	}

	d2s := invert(&s2d)
	// bias is a translation of the mapping from dst coordinates to src
	// coordinates such that the latter temporarily have non-negative X
	// and Y coordinates. This allows us to write int(f) instead of
	// int(math.Floor(f)), since "round to zero" and "round down" are
	// equivalent when f >= 0, but the former is much cheaper. The X--
	// and Y-- are because the TransformLeaf methods have a "sx -= 0.5"
	// adjustment.
	bias := transformRect(&d2s, &adr).Min
	bias.X--
	bias.Y--
	d2s[2] -= float64(bias.X)
	d2s[5] -= float64(bias.Y)
	// Make adr relative to dr.Min.
	adr = adr.Sub(dr.Min)

	switch op {
	case draw.Over:
		z.transformOver(fdst, dr, adr, &d2s, fsrc, sr, bias, &o)
	case draw.Src:
		z.transformSrc(fdst, dr, adr, &d2s, fsrc, sr, bias, &o)
	}
}

func (z ablInterpolator) Scale(dst draw.Image, dr image.Rectangle, src image.Image, sr image.Rectangle, op draw.Op, opts *draw.Options) {
	fsrc, sok := src.(*drawgl.FloatImage)
	fdst, dok := dst.(*drawgl.FloatImage)

	var o draw.Options
	if opts != nil {
		o = *opts
	}

	// Try to simplify a Scale to a Copy.
	if !sok || !dok || dr.Size() == sr.Size() ||
		o.DstMask != nil || o.SrcMask != nil || !sr.In(src.Bounds()) {
		draw.NearestNeighbor.Scale(dst, dr, src, sr, op, opts)
		return
	}

	// adr is the affected destination pixels.
	adr := dst.Bounds().Intersect(dr)
	if adr.Empty() || sr.Empty() {
		return
	}
	// Make adr relative to dr.Min.
	adr = adr.Sub(dr.Min)
	if op == draw.Over && o.SrcMask == nil && opaque(src) {
		op = draw.Src
	}

	switch op {
	case draw.Over:
		z.scaleOver(fdst, dr, adr, fsrc, sr, &o)
	case draw.Src:
		z.scaleSrc(fdst, dr, adr, fsrc, sr, &o)
	}
}

func (z ablInterpolator) Transform(dst draw.Image, s2d f64.Aff3, src image.Image, sr image.Rectangle, op draw.Op, opts *draw.Options) {
	fsrc, sok := src.(*drawgl.FloatImage)
	fdst, dok := dst.(*drawgl.FloatImage)

	var o draw.Options
	if opts != nil {
		o = *opts
	}

	if !sok || !dok || o.DstMask != nil || o.SrcMask != nil || !sr.In(src.Bounds()) {
		draw.NearestNeighbor.Transform(dst, s2d, src, sr, op, opts)
		return
	}

	// Try to simplify a Transform to a Copy.
	if s2d[0] == 1 && s2d[1] == 0 && s2d[3] == 0 && s2d[4] == 1 {
		dx := int(s2d[2])
		dy := int(s2d[5])
		if float64(dx) == s2d[2] && float64(dy) == s2d[5] {
			draw.Copy(dst, image.Point{X: sr.Min.X + dx, Y: sr.Min.X + dy}, src, sr, op, opts)
			return
		}
	}

	dr := transformRect(&s2d, &sr)
	// adr is the affected destination pixels.
	adr := dst.Bounds().Intersect(dr)
	if adr.Empty() || sr.Empty() {
		return
	}
	if op == draw.Over && o.SrcMask == nil && opaque(src) {
		op = draw.Src
	}

	d2s := invert(&s2d)
	// bias is a translation of the mapping from dst coordinates to src
	// coordinates such that the latter temporarily have non-negative X
	// and Y coordinates. This allows us to write int(f) instead of
	// int(math.Floor(f)), since "round to zero" and "round down" are
	// equivalent when f >= 0, but the former is much cheaper. The X--
	// and Y-- are because the TransformLeaf methods have a "sx -= 0.5"
	// adjustment.
	bias := transformRect(&d2s, &adr).Min
	bias.X--
	bias.Y--
	d2s[2] -= float64(bias.X)
	d2s[5] -= float64(bias.Y)
	// Make adr relative to dr.Min.
	adr = adr.Sub(dr.Min)

	switch op {
	case draw.Over:
		z.transformOver(fdst, dr, adr, &d2s, fsrc, sr, bias, &o)
	case draw.Src:
		z.transformSrc(fdst, dr, adr, &d2s, fsrc, sr, bias, &o)
	}
}

func (nnInterpolator) scaleOver(dst *drawgl.FloatImage, dr, adr image.Rectangle, src *drawgl.FloatImage, sr image.Rectangle, opts *draw.Options) {
	dw2 := uint64(dr.Dx()) * 2
	dh2 := uint64(dr.Dy()) * 2
	sw := uint64(sr.Dx())
	sh := uint64(sr.Dy())
	for dy := int32(adr.Min.Y); dy < int32(adr.Max.Y); dy++ {
		sy := (2*uint64(dy) + 1) * sh / dh2
		d := (dr.Min.Y+int(dy)-dst.Rect.Min.Y)*dst.Stride + (dr.Min.X+adr.Min.X-dst.Rect.Min.X)*4
		for dx := int32(adr.Min.X); dx < int32(adr.Max.X); dx, d = dx+1, d+4 {
			sx := (2*uint64(dx) + 1) * sw / dw2
			pi := (sr.Min.Y+int(sy)-src.Rect.Min.Y)*src.Stride + (sr.Min.X+int(sx)-src.Rect.Min.X)*4
			pr := src.Pix[pi+0]
			pg := src.Pix[pi+1]
			pb := src.Pix[pi+2]
			pa := src.Pix[pi+3]
			pa1 := 1 - pa
			dst.Pix[d+0] = dst.Pix[d+0]*pa1 + pr
			dst.Pix[d+1] = dst.Pix[d+1]*pa1 + pg
			dst.Pix[d+2] = dst.Pix[d+2]*pa1 + pb
			dst.Pix[d+3] = dst.Pix[d+3]*pa1 + pa
		}
	}
}

func (nnInterpolator) scaleSrc(dst *drawgl.FloatImage, dr, adr image.Rectangle, src *drawgl.FloatImage, sr image.Rectangle, opts *draw.Options) {
	dw2 := uint64(dr.Dx()) * 2
	dh2 := uint64(dr.Dy()) * 2
	sw := uint64(sr.Dx())
	sh := uint64(sr.Dy())
	for dy := int32(adr.Min.Y); dy < int32(adr.Max.Y); dy++ {
		sy := (2*uint64(dy) + 1) * sh / dh2
		d := (dr.Min.Y+int(dy)-dst.Rect.Min.Y)*dst.Stride + (dr.Min.X+adr.Min.X-dst.Rect.Min.X)*4
		for dx := int32(adr.Min.X); dx < int32(adr.Max.X); dx, d = dx+1, d+4 {
			sx := (2*uint64(dx) + 1) * sw / dw2
			pi := (sr.Min.Y+int(sy)-src.Rect.Min.Y)*src.Stride + (sr.Min.X+int(sx)-src.Rect.Min.X)*4
			dst.Pix[d+0] = src.Pix[pi+0]
			dst.Pix[d+1] = src.Pix[pi+1]
			dst.Pix[d+2] = src.Pix[pi+2]
			dst.Pix[d+3] = src.Pix[pi+3]
		}
	}
}

func (nnInterpolator) transformOver(dst *drawgl.FloatImage, dr, adr image.Rectangle, d2s *f64.Aff3, src *drawgl.FloatImage, sr image.Rectangle, bias image.Point, opts *draw.Options) {
	for dy := int32(adr.Min.Y); dy < int32(adr.Max.Y); dy++ {
		dyf := float64(dr.Min.Y+int(dy)) + 0.5
		d := (dr.Min.Y+int(dy)-dst.Rect.Min.Y)*dst.Stride + (dr.Min.X+adr.Min.X-dst.Rect.Min.X)*4
		for dx := int32(adr.Min.X); dx < int32(adr.Max.X); dx, d = dx+1, d+4 {
			dxf := float64(dr.Min.X+int(dx)) + 0.5
			sx0 := int(d2s[0]*dxf+d2s[1]*dyf+d2s[2]) + bias.X
			sy0 := int(d2s[3]*dxf+d2s[4]*dyf+d2s[5]) + bias.Y
			if !(image.Point{sx0, sy0}).In(sr) {
				continue
			}
			pi := (sy0-src.Rect.Min.Y)*src.Stride + (sx0-src.Rect.Min.X)*4
			pr := src.Pix[pi+0]
			pg := src.Pix[pi+1]
			pb := src.Pix[pi+2]
			pa := src.Pix[pi+3]
			pa1 := 1 - pa
			dst.Pix[d+0] = dst.Pix[d+0]*pa1 + pr
			dst.Pix[d+1] = dst.Pix[d+1]*pa1 + pg
			dst.Pix[d+2] = dst.Pix[d+2]*pa1 + pb
			dst.Pix[d+3] = dst.Pix[d+3]*pa1 + pa
		}
	}
}

func (nnInterpolator) transformSrc(dst *drawgl.FloatImage, dr, adr image.Rectangle, d2s *f64.Aff3, src *drawgl.FloatImage, sr image.Rectangle, bias image.Point, opts *draw.Options) {
	for dy := int32(adr.Min.Y); dy < int32(adr.Max.Y); dy++ {
		dyf := float64(dr.Min.Y+int(dy)) + 0.5
		d := (dr.Min.Y+int(dy)-dst.Rect.Min.Y)*dst.Stride + (dr.Min.X+adr.Min.X-dst.Rect.Min.X)*4
		for dx := int32(adr.Min.X); dx < int32(adr.Max.X); dx, d = dx+1, d+4 {
			dxf := float64(dr.Min.X+int(dx)) + 0.5
			sx0 := int(d2s[0]*dxf+d2s[1]*dyf+d2s[2]) + bias.X
			sy0 := int(d2s[3]*dxf+d2s[4]*dyf+d2s[5]) + bias.Y
			if !(image.Point{sx0, sy0}).In(sr) {
				continue
			}
			pi := (sy0-src.Rect.Min.Y)*src.Stride + (sx0-src.Rect.Min.X)*4
			dst.Pix[d+0] = src.Pix[pi+0]
			dst.Pix[d+1] = src.Pix[pi+1]
			dst.Pix[d+2] = src.Pix[pi+2]
			dst.Pix[d+3] = src.Pix[pi+3]
		}
	}
}

func (ablInterpolator) scaleOver(dst *drawgl.FloatImage, dr, adr image.Rectangle, src *drawgl.FloatImage, sr image.Rectangle, opts *draw.Options) {
	sw := int32(sr.Dx())
	sh := int32(sr.Dy())
	yscale := float64(sh) / float64(dr.Dy())
	xscale := float64(sw) / float64(dr.Dx())
	swMinus1, shMinus1 := sw-1, sh-1

	for dy := int32(adr.Min.Y); dy < int32(adr.Max.Y); dy++ {
		sy := (float64(dy)+0.5)*yscale - 0.5
		// If sy < 0, we will clamp sy0 to 0 anyway, so it doesn't matter if
		// we say int32(sy) instead of int32(math.Floor(sy)). Similarly for
		// sx, below.
		sy0 := int32(sy)
		yFrac0 := sy - float64(sy0)
		yFrac1 := 1 - yFrac0
		sy1 := sy0 + 1
		if sy < 0 {
			sy0, sy1 = 0, 0
			yFrac0, yFrac1 = 0, 1
		} else if sy1 > shMinus1 {
			sy0, sy1 = shMinus1, shMinus1
			yFrac0, yFrac1 = 1, 0
		}
		d := (dr.Min.Y+int(dy)-dst.Rect.Min.Y)*dst.Stride + (dr.Min.X+adr.Min.X-dst.Rect.Min.X)*4

		for dx := int32(adr.Min.X); dx < int32(adr.Max.X); dx, d = dx+1, d+4 {
			sx := (float64(dx)+0.5)*xscale - 0.5
			sx0 := int32(sx)
			xFrac0 := sx - float64(sx0)
			xFrac1 := 1 - xFrac0
			sx1 := sx0 + 1
			if sx < 0 {
				sx0, sx1 = 0, 0
				xFrac0, xFrac1 = 0, 1
			} else if sx1 > swMinus1 {
				sx0, sx1 = swMinus1, swMinus1
				xFrac0, xFrac1 = 1, 0
			}

			s00i := (sr.Min.Y+int(sy0)-src.Rect.Min.Y)*src.Stride + (sr.Min.X+int(sx0)-src.Rect.Min.X)*4
			s00ru := src.Pix[s00i+0]
			s00gu := src.Pix[s00i+1]
			s00bu := src.Pix[s00i+2]
			s00au := src.Pix[s00i+3]
			s00r := float64(s00ru)
			s00g := float64(s00gu)
			s00b := float64(s00bu)
			s00a := float64(s00au)
			s10i := (sr.Min.Y+int(sy0)-src.Rect.Min.Y)*src.Stride + (sr.Min.X+int(sx1)-src.Rect.Min.X)*4
			s10ru := src.Pix[s10i+0]
			s10gu := src.Pix[s10i+1]
			s10bu := src.Pix[s10i+2]
			s10au := src.Pix[s10i+3]
			s10r := float64(s10ru)
			s10g := float64(s10gu)
			s10b := float64(s10bu)
			s10a := float64(s10au)
			s10r = xFrac1*s00r + xFrac0*s10r
			s10g = xFrac1*s00g + xFrac0*s10g
			s10b = xFrac1*s00b + xFrac0*s10b
			s10a = xFrac1*s00a + xFrac0*s10a
			s01i := (sr.Min.Y+int(sy1)-src.Rect.Min.Y)*src.Stride + (sr.Min.X+int(sx0)-src.Rect.Min.X)*4
			s01ru := src.Pix[s01i+0]
			s01gu := src.Pix[s01i+1]
			s01bu := src.Pix[s01i+2]
			s01au := src.Pix[s01i+3]
			s01r := float64(s01ru)
			s01g := float64(s01gu)
			s01b := float64(s01bu)
			s01a := float64(s01au)
			s11i := (sr.Min.Y+int(sy1)-src.Rect.Min.Y)*src.Stride + (sr.Min.X+int(sx1)-src.Rect.Min.X)*4
			s11ru := src.Pix[s11i+0]
			s11gu := src.Pix[s11i+1]
			s11bu := src.Pix[s11i+2]
			s11au := src.Pix[s11i+3]
			s11r := float64(s11ru)
			s11g := float64(s11gu)
			s11b := float64(s11bu)
			s11a := float64(s11au)
			s11r = xFrac1*s01r + xFrac0*s11r
			s11g = xFrac1*s01g + xFrac0*s11g
			s11b = xFrac1*s01b + xFrac0*s11b
			s11a = xFrac1*s01a + xFrac0*s11a
			s11r = yFrac1*s10r + yFrac0*s11r
			s11g = yFrac1*s10g + yFrac0*s11g
			s11b = yFrac1*s10b + yFrac0*s11b
			s11a = yFrac1*s10a + yFrac0*s11a
			pr := drawgl.ColorValue(s11r)
			pg := drawgl.ColorValue(s11g)
			pb := drawgl.ColorValue(s11b)
			pa := drawgl.ColorValue(s11a)
			pa1 := (1 - pa)
			dst.Pix[d+0] = dst.Pix[d+0]*pa1 + pr
			dst.Pix[d+1] = dst.Pix[d+1]*pa1 + pg
			dst.Pix[d+2] = dst.Pix[d+2]*pa1 + pb
			dst.Pix[d+3] = dst.Pix[d+3]*pa1 + pa
		}
	}
}

func (ablInterpolator) scaleSrc(dst *drawgl.FloatImage, dr, adr image.Rectangle, src *drawgl.FloatImage, sr image.Rectangle, opts *draw.Options) {
	sw := int32(sr.Dx())
	sh := int32(sr.Dy())
	yscale := float64(sh) / float64(dr.Dy())
	xscale := float64(sw) / float64(dr.Dx())
	swMinus1, shMinus1 := sw-1, sh-1

	for dy := int32(adr.Min.Y); dy < int32(adr.Max.Y); dy++ {
		sy := (float64(dy)+0.5)*yscale - 0.5
		// If sy < 0, we will clamp sy0 to 0 anyway, so it doesn't matter if
		// we say int32(sy) instead of int32(math.Floor(sy)). Similarly for
		// sx, below.
		sy0 := int32(sy)
		yFrac0 := sy - float64(sy0)
		yFrac1 := 1 - yFrac0
		sy1 := sy0 + 1
		if sy < 0 {
			sy0, sy1 = 0, 0
			yFrac0, yFrac1 = 0, 1
		} else if sy1 > shMinus1 {
			sy0, sy1 = shMinus1, shMinus1
			yFrac0, yFrac1 = 1, 0
		}
		d := (dr.Min.Y+int(dy)-dst.Rect.Min.Y)*dst.Stride + (dr.Min.X+adr.Min.X-dst.Rect.Min.X)*4

		for dx := int32(adr.Min.X); dx < int32(adr.Max.X); dx, d = dx+1, d+4 {
			sx := (float64(dx)+0.5)*xscale - 0.5
			sx0 := int32(sx)
			xFrac0 := sx - float64(sx0)
			xFrac1 := 1 - xFrac0
			sx1 := sx0 + 1
			if sx < 0 {
				sx0, sx1 = 0, 0
				xFrac0, xFrac1 = 0, 1
			} else if sx1 > swMinus1 {
				sx0, sx1 = swMinus1, swMinus1
				xFrac0, xFrac1 = 1, 0
			}

			s00i := (sr.Min.Y+int(sy0)-src.Rect.Min.Y)*src.Stride + (sr.Min.X+int(sx0)-src.Rect.Min.X)*4
			s00ru := src.Pix[s00i+0]
			s00gu := src.Pix[s00i+1]
			s00bu := src.Pix[s00i+2]
			s00au := src.Pix[s00i+3]
			s00r := float64(s00ru)
			s00g := float64(s00gu)
			s00b := float64(s00bu)
			s00a := float64(s00au)
			s10i := (sr.Min.Y+int(sy0)-src.Rect.Min.Y)*src.Stride + (sr.Min.X+int(sx1)-src.Rect.Min.X)*4
			s10ru := src.Pix[s10i+0]
			s10gu := src.Pix[s10i+1]
			s10bu := src.Pix[s10i+2]
			s10au := src.Pix[s10i+3]
			s10r := float64(s10ru)
			s10g := float64(s10gu)
			s10b := float64(s10bu)
			s10a := float64(s10au)
			s10r = xFrac1*s00r + xFrac0*s10r
			s10g = xFrac1*s00g + xFrac0*s10g
			s10b = xFrac1*s00b + xFrac0*s10b
			s10a = xFrac1*s00a + xFrac0*s10a
			s01i := (sr.Min.Y+int(sy1)-src.Rect.Min.Y)*src.Stride + (sr.Min.X+int(sx0)-src.Rect.Min.X)*4
			s01ru := src.Pix[s01i+0]
			s01gu := src.Pix[s01i+1]
			s01bu := src.Pix[s01i+2]
			s01au := src.Pix[s01i+3]
			s01r := float64(s01ru)
			s01g := float64(s01gu)
			s01b := float64(s01bu)
			s01a := float64(s01au)
			s11i := (sr.Min.Y+int(sy1)-src.Rect.Min.Y)*src.Stride + (sr.Min.X+int(sx1)-src.Rect.Min.X)*4
			s11ru := src.Pix[s11i+0]
			s11gu := src.Pix[s11i+1]
			s11bu := src.Pix[s11i+2]
			s11au := src.Pix[s11i+3]
			s11r := float64(s11ru)
			s11g := float64(s11gu)
			s11b := float64(s11bu)
			s11a := float64(s11au)
			s11r = xFrac1*s01r + xFrac0*s11r
			s11g = xFrac1*s01g + xFrac0*s11g
			s11b = xFrac1*s01b + xFrac0*s11b
			s11a = xFrac1*s01a + xFrac0*s11a
			s11r = yFrac1*s10r + yFrac0*s11r
			s11g = yFrac1*s10g + yFrac0*s11g
			s11b = yFrac1*s10b + yFrac0*s11b
			s11a = yFrac1*s10a + yFrac0*s11a
			pr := drawgl.ColorValue(s11r)
			pg := drawgl.ColorValue(s11g)
			pb := drawgl.ColorValue(s11b)
			pa := drawgl.ColorValue(s11a)
			dst.Pix[d+0] = pr
			dst.Pix[d+1] = pg
			dst.Pix[d+2] = pb
			dst.Pix[d+3] = pa
		}
	}
}

func (ablInterpolator) transformOver(dst *drawgl.FloatImage, dr, adr image.Rectangle, d2s *f64.Aff3, src *drawgl.FloatImage, sr image.Rectangle, bias image.Point, opts *draw.Options) {
	for dy := int32(adr.Min.Y); dy < int32(adr.Max.Y); dy++ {
		dyf := float64(dr.Min.Y+int(dy)) + 0.5
		d := (dr.Min.Y+int(dy)-dst.Rect.Min.Y)*dst.Stride + (dr.Min.X+adr.Min.X-dst.Rect.Min.X)*4
		for dx := int32(adr.Min.X); dx < int32(adr.Max.X); dx, d = dx+1, d+4 {
			dxf := float64(dr.Min.X+int(dx)) + 0.5
			sx := d2s[0]*dxf + d2s[1]*dyf + d2s[2]
			sy := d2s[3]*dxf + d2s[4]*dyf + d2s[5]
			if !(image.Point{int(sx) + bias.X, int(sy) + bias.Y}).In(sr) {
				continue
			}

			sx -= 0.5
			sx0 := int(sx)
			xFrac0 := sx - float64(sx0)
			xFrac1 := 1 - xFrac0
			sx0 += bias.X
			sx1 := sx0 + 1
			if sx0 < sr.Min.X {
				sx0, sx1 = sr.Min.X, sr.Min.X
				xFrac0, xFrac1 = 0, 1
			} else if sx1 >= sr.Max.X {
				sx0, sx1 = sr.Max.X-1, sr.Max.X-1
				xFrac0, xFrac1 = 1, 0
			}

			sy -= 0.5
			sy0 := int(sy)
			yFrac0 := sy - float64(sy0)
			yFrac1 := 1 - yFrac0
			sy0 += bias.Y
			sy1 := sy0 + 1
			if sy0 < sr.Min.Y {
				sy0, sy1 = sr.Min.Y, sr.Min.Y
				yFrac0, yFrac1 = 0, 1
			} else if sy1 >= sr.Max.Y {
				sy0, sy1 = sr.Max.Y-1, sr.Max.Y-1
				yFrac0, yFrac1 = 1, 0
			}

			s00i := (sy0-src.Rect.Min.Y)*src.Stride + (sx0-src.Rect.Min.X)*4
			s00ru := src.Pix[s00i+0]
			s00gu := src.Pix[s00i+1]
			s00bu := src.Pix[s00i+2]
			s00au := src.Pix[s00i+3]
			s00r := float64(s00ru)
			s00g := float64(s00gu)
			s00b := float64(s00bu)
			s00a := float64(s00au)
			s10i := (sy0-src.Rect.Min.Y)*src.Stride + (sx1-src.Rect.Min.X)*4
			s10ru := src.Pix[s10i+0]
			s10gu := src.Pix[s10i+1]
			s10bu := src.Pix[s10i+2]
			s10au := src.Pix[s10i+3]
			s10r := float64(s10ru)
			s10g := float64(s10gu)
			s10b := float64(s10bu)
			s10a := float64(s10au)
			s10r = xFrac1*s00r + xFrac0*s10r
			s10g = xFrac1*s00g + xFrac0*s10g
			s10b = xFrac1*s00b + xFrac0*s10b
			s10a = xFrac1*s00a + xFrac0*s10a
			s01i := (sy1-src.Rect.Min.Y)*src.Stride + (sx0-src.Rect.Min.X)*4
			s01ru := src.Pix[s01i+0]
			s01gu := src.Pix[s01i+1]
			s01bu := src.Pix[s01i+2]
			s01au := src.Pix[s01i+3]
			s01r := float64(s01ru)
			s01g := float64(s01gu)
			s01b := float64(s01bu)
			s01a := float64(s01au)
			s11i := (sy1-src.Rect.Min.Y)*src.Stride + (sx1-src.Rect.Min.X)*4
			s11ru := src.Pix[s11i+0]
			s11gu := src.Pix[s11i+1]
			s11bu := src.Pix[s11i+2]
			s11au := src.Pix[s11i+3]
			s11r := float64(s11ru)
			s11g := float64(s11gu)
			s11b := float64(s11bu)
			s11a := float64(s11au)
			s11r = xFrac1*s01r + xFrac0*s11r
			s11g = xFrac1*s01g + xFrac0*s11g
			s11b = xFrac1*s01b + xFrac0*s11b
			s11a = xFrac1*s01a + xFrac0*s11a
			s11r = yFrac1*s10r + yFrac0*s11r
			s11g = yFrac1*s10g + yFrac0*s11g
			s11b = yFrac1*s10b + yFrac0*s11b
			s11a = yFrac1*s10a + yFrac0*s11a
			pr := drawgl.ColorValue(s11r)
			pg := drawgl.ColorValue(s11g)
			pb := drawgl.ColorValue(s11b)
			pa := drawgl.ColorValue(s11a)
			pa1 := (1 - pa)
			dst.Pix[d+0] = dst.Pix[d+0]*pa1 + pr
			dst.Pix[d+1] = dst.Pix[d+1]*pa1 + pg
			dst.Pix[d+2] = dst.Pix[d+2]*pa1 + pb
			dst.Pix[d+3] = dst.Pix[d+3]*pa1 + pa
		}
	}
}

func (ablInterpolator) transformSrc(dst *drawgl.FloatImage, dr, adr image.Rectangle, d2s *f64.Aff3, src *drawgl.FloatImage, sr image.Rectangle, bias image.Point, opts *draw.Options) {
	for dy := int32(adr.Min.Y); dy < int32(adr.Max.Y); dy++ {
		dyf := float64(dr.Min.Y+int(dy)) + 0.5
		d := (dr.Min.Y+int(dy)-dst.Rect.Min.Y)*dst.Stride + (dr.Min.X+adr.Min.X-dst.Rect.Min.X)*4
		for dx := int32(adr.Min.X); dx < int32(adr.Max.X); dx, d = dx+1, d+4 {
			dxf := float64(dr.Min.X+int(dx)) + 0.5
			sx := d2s[0]*dxf + d2s[1]*dyf + d2s[2]
			sy := d2s[3]*dxf + d2s[4]*dyf + d2s[5]
			if !(image.Point{int(sx) + bias.X, int(sy) + bias.Y}).In(sr) {
				continue
			}

			sx -= 0.5
			sx0 := int(sx)
			xFrac0 := sx - float64(sx0)
			xFrac1 := 1 - xFrac0
			sx0 += bias.X
			sx1 := sx0 + 1
			if sx0 < sr.Min.X {
				sx0, sx1 = sr.Min.X, sr.Min.X
				xFrac0, xFrac1 = 0, 1
			} else if sx1 >= sr.Max.X {
				sx0, sx1 = sr.Max.X-1, sr.Max.X-1
				xFrac0, xFrac1 = 1, 0
			}

			sy -= 0.5
			sy0 := int(sy)
			yFrac0 := sy - float64(sy0)
			yFrac1 := 1 - yFrac0
			sy0 += bias.Y
			sy1 := sy0 + 1
			if sy0 < sr.Min.Y {
				sy0, sy1 = sr.Min.Y, sr.Min.Y
				yFrac0, yFrac1 = 0, 1
			} else if sy1 >= sr.Max.Y {
				sy0, sy1 = sr.Max.Y-1, sr.Max.Y-1
				yFrac0, yFrac1 = 1, 0
			}

			s00i := (sy0-src.Rect.Min.Y)*src.Stride + (sx0-src.Rect.Min.X)*4
			s00ru := src.Pix[s00i+0]
			s00gu := src.Pix[s00i+1]
			s00bu := src.Pix[s00i+2]
			s00au := src.Pix[s00i+3]
			s00r := float64(s00ru)
			s00g := float64(s00gu)
			s00b := float64(s00bu)
			s00a := float64(s00au)
			s10i := (sy0-src.Rect.Min.Y)*src.Stride + (sx1-src.Rect.Min.X)*4
			s10ru := src.Pix[s10i+0]
			s10gu := src.Pix[s10i+1]
			s10bu := src.Pix[s10i+2]
			s10au := src.Pix[s10i+3]
			s10r := float64(s10ru)
			s10g := float64(s10gu)
			s10b := float64(s10bu)
			s10a := float64(s10au)
			s10r = xFrac1*s00r + xFrac0*s10r
			s10g = xFrac1*s00g + xFrac0*s10g
			s10b = xFrac1*s00b + xFrac0*s10b
			s10a = xFrac1*s00a + xFrac0*s10a
			s01i := (sy1-src.Rect.Min.Y)*src.Stride + (sx0-src.Rect.Min.X)*4
			s01ru := src.Pix[s01i+0]
			s01gu := src.Pix[s01i+1]
			s01bu := src.Pix[s01i+2]
			s01au := src.Pix[s01i+3]
			s01r := float64(s01ru)
			s01g := float64(s01gu)
			s01b := float64(s01bu)
			s01a := float64(s01au)
			s11i := (sy1-src.Rect.Min.Y)*src.Stride + (sx1-src.Rect.Min.X)*4
			s11ru := src.Pix[s11i+0]
			s11gu := src.Pix[s11i+1]
			s11bu := src.Pix[s11i+2]
			s11au := src.Pix[s11i+3]
			s11r := float64(s11ru)
			s11g := float64(s11gu)
			s11b := float64(s11bu)
			s11a := float64(s11au)
			s11r = xFrac1*s01r + xFrac0*s11r
			s11g = xFrac1*s01g + xFrac0*s11g
			s11b = xFrac1*s01b + xFrac0*s11b
			s11a = xFrac1*s01a + xFrac0*s11a
			s11r = yFrac1*s10r + yFrac0*s11r
			s11g = yFrac1*s10g + yFrac0*s11g
			s11b = yFrac1*s10b + yFrac0*s11b
			s11a = yFrac1*s10a + yFrac0*s11a
			pr := drawgl.ColorValue(s11r)
			pg := drawgl.ColorValue(s11g)
			pb := drawgl.ColorValue(s11b)
			pa := drawgl.ColorValue(s11a)
			dst.Pix[d+0] = pr
			dst.Pix[d+1] = pg
			dst.Pix[d+2] = pb
			dst.Pix[d+3] = pa
		}
	}
}
func sinc(x float64) float64 {
	if x == 0 {
		return 1.0
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}

func opaque(m image.Image) bool {
	o, ok := m.(interface {
		Opaque() bool
	})
	return ok && o.Opaque()
}

func transformRect(s2d *f64.Aff3, sr *image.Rectangle) (dr image.Rectangle) {
	ps := [...]image.Point{
		{sr.Min.X, sr.Min.Y},
		{sr.Max.X, sr.Min.Y},
		{sr.Min.X, sr.Max.Y},
		{sr.Max.X, sr.Max.Y},
	}
	for i, p := range ps {
		sxf := float64(p.X)
		syf := float64(p.Y)
		dx := int(math.Floor(s2d[0]*sxf + s2d[1]*syf + s2d[2]))
		dy := int(math.Floor(s2d[3]*sxf + s2d[4]*syf + s2d[5]))

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
	return dr
}

func invert(m *f64.Aff3) f64.Aff3 {
	m00 := +m[3*1+1]
	m01 := -m[3*0+1]
	m02 := +m[3*1+2]*m[3*0+1] - m[3*1+1]*m[3*0+2]
	m10 := -m[3*1+0]
	m11 := +m[3*0+0]
	m12 := +m[3*1+0]*m[3*0+2] - m[3*1+2]*m[3*0+0]

	det := m00*m11 - m10*m01

	return f64.Aff3{
		m00 / det,
		m01 / det,
		m02 / det,
		m10 / det,
		m11 / det,
		m12 / det,
	}
}
