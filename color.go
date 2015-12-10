package drawgl

import "image/color"

type FloatColor struct {
	R, G, B, A float64
}

const (
	maxColor = 0xffff
)

var (
	FloatColorModel color.Model = color.ModelFunc(floatColorModel)
)

func (c FloatColor) RGBA() (uint32, uint32, uint32, uint32) {
	return uint32(ClampColor(c.R) * maxColor),
		uint32(ClampColor(c.G) * maxColor),
		uint32(ClampColor(c.B) * maxColor),
		uint32(ClampColor(c.A) * maxColor)
}

func floatColorModel(c color.Color) color.Color {
	if _, ok := c.(FloatColor); ok {
		return c
	}

	r, g, b, a := c.RGBA()
	return FloatColor{float64(r) / maxColor, float64(g) / maxColor,
		float64(b) / maxColor, float64(a) / maxColor}
}

func ClampColor(in float64) float64 {
	if in < 0 {
		return 0
	} else if in > maxColor {
		return maxColor
	}

	return in
}
