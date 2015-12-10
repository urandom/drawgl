package convolution

import (
	"errors"
	"math"
)

type Kernel interface {
	Weights() []float64
	Normalized() ([]float64, float64)
}

type HVKernel interface {
	HWeights() []float64
	VWeights() []float64
}

type kernel []float64
type hvkernel struct {
	h []float64
	v []float64
}

const (
	halfOffset = 0.5 / 0xffff
	fullOffset = 1 / 0xffff
)

func NewKernel(data []float64) (k Kernel, err error) {
	size := int(math.Sqrt(float64(len(data))))
	if size%2 == 0 || size*size != len(data) {
		err = errors.New("Kernel has to be an odd square")
		return
	}

	k = kernel(data)

	return
}

func (k kernel) Weights() []float64 {
	return k
}

func (k kernel) Normalized() ([]float64, float64) {
	return NormalizeData(k)
}

func NewHVKernel(h, v []float64) (k HVKernel, err error) {
	if len(h)%2 == 0 {
		err = errors.New("Horizontal kernel has to be odd")
		return
	}
	if len(v)%2 == 0 {
		err = errors.New("Vertical kernel has to be odd")
		return
	}
	k = hvkernel{h, v}

	return
}

func (k hvkernel) HWeights() []float64 {
	return k.h
}

func (k hvkernel) VWeights() []float64 {
	return k.v
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
		offset = fullOffset
	} else {
		div = 1
		offset = halfOffset
	}

	if div != 1 {
		for i := range data {
			data[i] = data[i] / div
		}

	}

	return data, offset
}
