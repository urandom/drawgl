package convolution

import (
	"errors"
	"math"
)

type Kernel interface {
	Weights() []float64
	Normalized() ([]float64, float64)
}

type kernel []float64

func (k kernel) Weights() []float64 {
	return k
}

func NewKernel(data []float64) (k Kernel, err error) {
	size := int(math.Sqrt(float64(len(data))))
	if size%2 == 0 || size*size != len(data) {
		err = errors.New("Kernel has to be an odd square")
		return
	}

	k = kernel(data)

	return
}

func (k kernel) Normalized() ([]float64, float64) {
	return NormalizeData(k)
}

func NormalizeData(data []float64) ([]float64, float64) {
	var sum float64
	for _, d := range data {
		sum += d
	}

	var div, offset float64
	if sum > 0 {
		div = sum
		offset = 0
	} else if sum < 0 {
		div = -sum
		offset = 1
	} else {
		div = 1
		offset = 0.5
	}

	if div != 1 {
		for i := range data {
			data[i] = data[i] / div
		}

	}

	return data, offset
}
