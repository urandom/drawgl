package transform

import (
	"math"

	"golang.org/x/image/draw"
)

var (
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

func sinc(x float64) float64 {
	if x == 0 {
		return 1.0
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}
