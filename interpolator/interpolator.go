package interpolator

import (
	"image"
	"math"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/transform/matrix"
)

type Interpolator interface {
	Get(src *drawgl.FloatImage, x, y float64) drawgl.FloatColor
}

func New(
	kind string,
	src *drawgl.FloatImage,
	m matrix.Matrix3,
	bias image.Point,
) Interpolator {

	switch kind {
	case "NearestNeighbor":
		return nearestNeighbor{}
	case "ApproximageBilinear":
		return newApproximageBilinear(bias)
	case "CatmullRom":
		k := newKernel(src, m)
		k.Support = 2
		k.At = func(t float64) float64 {
			if t < 1 {
				return (1.5*t-2.5)*t*t + 1
			}
			return ((-0.5*t+2.5)*t-4)*t + 2
		}

		return k
	case "Lanczos":
		k := newKernel(src, m)
		k.Support = 3
		k.At = func(t float64) float64 {
			t = math.Abs(t)
			if t < 3 {
				return sinc(t) * sinc(t/3)
			}
			return 0
		}

		return k
	case "Bilinear":
		fallthrough
	default:
		k := newKernel(src, m)
		k.Support = 1
		k.At = func(t float64) float64 {
			return 1 - t
		}

		return k
	}
}

func sinc(x float64) float64 {
	if x == 0 {
		return 1.0
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}
