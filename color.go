package drawgl

import (
	"errors"
	"image/color"
)

type ColorValue float32

type FloatColor struct {
	R, G, B, A ColorValue
}

type Channel int

const (
	maxColor = 0xffff
)

const (
	RGB Channel = iota
	Red         = 1 << iota
	Green
	Blue
	Alpha

	m = 1<<16 - 1
)

var (
	FloatColorModel color.Model = color.ModelFunc(floatColorModel)
)

func (c FloatColor) RGBA() (uint32, uint32, uint32, uint32) {
	return uint32(c.R.Clamped()*maxColor + 0.5),
		uint32(c.G.Clamped()*maxColor + 0.5),
		uint32(c.B.Clamped()*maxColor + 0.5),
		uint32(c.A.Clamped()*maxColor + 0.5)
}

func (v ColorValue) Clamped() ColorValue {
	if v < 0 {
		v = 0
	} else if v > 1 {
		v = 1
	}

	return v
}

func floatColorModel(c color.Color) color.Color {
	if _, ok := c.(FloatColor); ok {
		return c
	}

	r, g, b, a := c.RGBA()
	return FloatColor{ColorValue(r) / maxColor, ColorValue(g) / maxColor,
		ColorValue(b) / maxColor, ColorValue(a) / maxColor}
}

func (c *Channel) Normalize(includeAlpha ...bool) {
	if *c == RGB {
		*c = Red | Green | Blue
		if len(includeAlpha) > 0 && includeAlpha[0] {
			*c |= Alpha
		}
	}
}

func (c Channel) Is(o Channel) bool {
	return c&o == o
}

func (c Channel) MarshalJSON() (b []byte, err error) {
	if c == RGB {
		b = []byte(`"RGB"`)
	} else {
		b = []byte{'"'}
		if c.Is(Red) {
			b = append(b, 'R')
		}
		if c.Is(Green) {
			b = append(b, 'G')
		}
		if c.Is(Blue) {
			b = append(b, 'B')
		}
		if c.Is(Alpha) {
			b = append(b, 'A')
		}
		b = append(b, '"')
	}

	return
}

func (c *Channel) UnmarshalJSON(b []byte) (err error) {
	if b[0] == 34 && b[len(b)-1] == 34 {
		for i := 1; i < len(b)-1; i++ {
			switch b[i] {
			case 'R':
				*c |= Red
			case 'G':
				*c |= Green
			case 'B':
				*c |= Blue
			case 'A':
				*c |= Alpha
			default:
				err = errors.New("unknown channel value " + string(b))
				break
			}
		}
	} else {
		err = errors.New("unknown channel value " + string(b))
	}
	return
}
