package interpolator

import "github.com/urandom/drawgl"

var (
	BiLinear Interpolator = bilinear{}
)

type Interpolator interface {
	Get(src *drawgl.FloatImage, x, y float64, edgeHandler drawgl.EdgeHandler) drawgl.FloatColor
}

func Inst(typ string) Interpolator {
	switch typ {
	case "BiLinear":
		return BiLinear
	default:
		return BiLinear
	}
}
