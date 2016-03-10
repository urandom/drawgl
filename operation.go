package drawgl

import (
	"errors"
	"image"
	"image/draw"
)

type Mask struct {
	Image image.Image
	Rect  image.Rectangle

	hasImage bool
	hasRect  bool
}

type EdgeHandler int

const (
	Extend EdgeHandler = iota
	Wrap
	Transparent
)

var (
	ErrOutOfBounds = errors.New("out of bounds")
)

func NewMask(image image.Image, rect image.Rectangle) Mask {
	return Mask{Image: image, Rect: rect, hasImage: image != nil, hasRect: !rect.Empty()}

}

func MaskFactor(pt image.Point, mask Mask) (factor float32) {
	if mask.hasRect && !pt.In(mask.Rect) {
		return 0
	}

	if mask.hasImage {
		_, _, _, ma := mask.Image.At(pt.X, pt.Y).RGBA()

		return float32(ma) / float32(m)
	}

	return 1
}

func MaskColor(dst FloatColor, src FloatColor, c Channel, f float32, op draw.Op) FloatColor {
	fv := ColorValue(f)
	switch op {
	case draw.Over:
		switch fv {
		case 0:
		case 1:
			if c.Is(Red) {
				dst.R = src.R
			}
			if c.Is(Green) {
				dst.G = src.G
			}
			if c.Is(Blue) {
				dst.B = src.B
			}
			if c.Is(Alpha) {
				dst.A = src.A
			}
		default:
			if c.Is(Red) {
				dst.R = dst.R/fv + src.R*fv
			}
			if c.Is(Green) {
				dst.G = dst.G/fv + src.G*fv
			}
			if c.Is(Blue) {
				dst.B = dst.B/fv + src.B*fv
			}
			if c.Is(Alpha) {
				dst.A = dst.A/fv + src.A*fv
			}
		}
	case draw.Src:
		switch fv {
		case 0:
			if c.Is(Red) {
				dst.R = 0
			}
			if c.Is(Green) {
				dst.G = 0
			}
			if c.Is(Blue) {
				dst.B = 0
			}
			if c.Is(Alpha) {
				dst.A = 0
			}
		case 1:
			if c.Is(Red) {
				dst.R = src.R
			}
			if c.Is(Green) {
				dst.G = src.G
			}
			if c.Is(Blue) {
				dst.B = src.B
			}
			if c.Is(Alpha) {
				dst.A = src.A
			}
		default:
			if c.Is(Red) {
				dst.R = src.R * fv
			}
			if c.Is(Green) {
				dst.G = src.G * fv
			}
			if c.Is(Blue) {
				dst.B = src.B * fv
			}
			if c.Is(Alpha) {
				dst.A = src.A * fv
			}
		}
	}

	return dst
}

func TranslateCoords(x, y int, b image.Rectangle, h EdgeHandler) (mx, my int, err error) {
	mx, my = x, y

	switch h {
	case Wrap:
		if mx < b.Min.X {
			mx = b.Max.X - b.Min.X + mx
		} else if mx >= b.Max.X {
			mx = b.Min.X - b.Max.X + mx
		}

		if my < b.Min.Y {
			my = b.Max.Y - b.Min.Y + my
		} else if my >= b.Max.Y {
			my = b.Min.Y - b.Max.Y + my
		}
	case Extend:
		if mx < b.Min.X {
			mx = b.Min.X
		} else if mx >= b.Max.X {
			mx = b.Max.X - 1
		}

		if my < b.Min.Y {
			my = b.Min.Y
		} else if my >= b.Max.Y {
			my = b.Max.Y - 1
		}
	case Transparent:
		if mx < b.Min.X || mx >= b.Max.X || my < b.Min.Y || my >= b.Max.Y {
			err = ErrOutOfBounds
		}
	}

	return
}
