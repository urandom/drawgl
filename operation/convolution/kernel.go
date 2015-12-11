package convolution

import (
	"errors"
	"math"
)

type Kernel interface {
	Weights() []float32
	Normalized() ([]float32, float32)
}

type HVKernel interface {
	HWeights() []float32
	VWeights() []float32
	HNormalized() ([]float32, float32)
	VNormalized() ([]float32, float32)
}

type kernel []float32
type hvkernel struct {
	h []float32
	v []float32
}

const (
	halfOffset = 0.5 / 0xffff
	fullOffset = 1 / 0xffff
)

func NewKernel(data []float32) (k Kernel, err error) {
	size := int(math.Sqrt(float64(len(data))))
	if size%2 == 0 || size*size != len(data) {
		err = errors.New("Kernel has to be an odd square")
		return
	}

	k = kernel(data)

	return
}

func (k kernel) Weights() []float32 {
	return k
}

func (k kernel) Normalized() ([]float32, float32) {
	return NormalizeData(k)
}

func NewHVKernel(h, v []float32) (k HVKernel, err error) {
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

func (k hvkernel) HWeights() []float32 {
	return k.h
}

func (k hvkernel) VWeights() []float32 {
	return k.v
}

func (k hvkernel) HNormalized() ([]float32, float32) {
	return NormalizeData(k.h)
}

func (k hvkernel) VNormalized() ([]float32, float32) {
	return NormalizeData(k.v)
}

func NormalizeData(data []float32) ([]float32, float32) {
	var sum float32
	for _, d := range data {
		sum += d
	}

	var div, offset float32
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
