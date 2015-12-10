package drawgl

import "image/color"

type ColorValue float32

type FloatColor struct {
	R, G, B, A ColorValue
}

const (
	maxColor = 0xffff
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
