package interpolator

import (
	"image"
	"math"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/transform/matrix"
)

var (
	BiLinear Interpolator = bilinear{}
)

type Interpolator interface {
	Get(src *drawgl.FloatImage, x, y float64, edgeHandler drawgl.EdgeHandler) drawgl.FloatColor
}

func New(
	kind string,
	src *drawgl.FloatImage,
	edgeHandler drawgl.EdgeHandler,
	m matrix.Matrix3,
	bias image.Point,
) Interpolator {

	var i = BiLinear
	switch kind {
	case "BiLinear":
		k := newKernel(src, edgeHandler, m, bias)
		k.Support = 1
		k.At = func(t float64) float64 {
			return 1 - t
		}

		return k
	case "CatmullRom":
		k := newKernel(src, edgeHandler, m, bias)
		k.Support = 2
		k.At = func(t float64) float64 {
			if t < 1 {
				return (1.5*t-2.5)*t*t + 1
			}
			return ((-0.5*t+2.5)*t-4)*t + 2
		}

		return k
	case "Lanczos":
		k := newKernel(src, edgeHandler, m, bias)
		k.Support = 3
		k.At = func(t float64) float64 {
			t = math.Abs(t)
			if t < 3 {
				return sinc(t) * sinc(t/3)
			}
			return 0
		}

		return k
	}

	return i
}

func sinc(x float64) float64 {
	if x == 0 {
		return 1.0
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}
