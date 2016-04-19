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
	maxColor          = 0xffff
	defaultFloatDelta = 0.005
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

// ApproxEqual is used by tests to check whether a color is approximately equal
// to another
func (c FloatColor) ApproxEqual(o FloatColor) bool {
	if c != o {
		if c.R > o.R {
			if c.R-o.R > defaultFloatDelta {
				return false
			}
		} else if c.R < o.R {
			if o.R-c.R > defaultFloatDelta {
				return false
			}
		}

		if c.G > o.G {
			if c.G-o.G > defaultFloatDelta {
				return false
			}
		} else if c.G < o.G {
			if o.G-c.G > defaultFloatDelta {
				return false
			}
		}

		if c.B > o.B {
			if c.B-o.B > defaultFloatDelta {
				return false
			}
		} else if c.B < o.B {
			if o.B-c.B > defaultFloatDelta {
				return false
			}
		}

		if c.A > o.A {
			if c.A-o.A > defaultFloatDelta {
				return false
			}
		} else if c.A < o.A {
			if o.A-c.A > defaultFloatDelta {
				return false
			}
		}

	}

	return true
}

// Clamped returns a clamped color value, with a range 0-1
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

// Normalize returns a normalized version a channel value. If the value is the
// composite RGB value, it is transformed to Red | Green | Blue. If
// includeAlpha is given and true, the alpha is also added to the channel
func (c Channel) Normalize(includeAlpha ...bool) Channel {
	if c == RGB {
		c = Red | Green | Blue
		if len(includeAlpha) > 0 && includeAlpha[0] {
			c |= Alpha
		}
	}

	return c
}

// Is checks whether the channel contains a given value
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
