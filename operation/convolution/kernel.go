package convolution

import (
	"errors"
	"math"

	"github.com/urandom/drawgl"
)

type Kernel interface {
	Weights() []drawgl.ColorValue
	Normalized() ([]drawgl.ColorValue, drawgl.ColorValue)
}

type HVKernel interface {
	HWeights() []drawgl.ColorValue
	VWeights() []drawgl.ColorValue
	HNormalized() ([]drawgl.ColorValue, drawgl.ColorValue)
	VNormalized() ([]drawgl.ColorValue, drawgl.ColorValue)
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

func (k kernel) Weights() []drawgl.ColorValue {
	w := make([]drawgl.ColorValue, len(k))
	for i := range k {
		w[i] = drawgl.ColorValue(k[i])
	}
	return w
}

func (k kernel) Normalized() ([]drawgl.ColorValue, drawgl.ColorValue) {
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

func (k hvkernel) HWeights() []drawgl.ColorValue {
	w := make([]drawgl.ColorValue, len(k.h))
	for i := range k.h {
		w[i] = drawgl.ColorValue(k.h[i])
	}
	return w
}

func (k hvkernel) VWeights() []drawgl.ColorValue {
	w := make([]drawgl.ColorValue, len(k.v))
	for i := range k.v {
		w[i] = drawgl.ColorValue(k.v[i])
	}
	return w
}

func (k hvkernel) HNormalized() ([]drawgl.ColorValue, drawgl.ColorValue) {
	return NormalizeData(k.h)
}

func (k hvkernel) VNormalized() ([]drawgl.ColorValue, drawgl.ColorValue) {
	return NormalizeData(k.v)
}

func NormalizeData(data []float32) (normalized []drawgl.ColorValue, offset drawgl.ColorValue) {
	var sum drawgl.ColorValue

	normalized = make([]drawgl.ColorValue, len(data))
	for i := range data {
		d := drawgl.ColorValue(data[i])
		sum += d
		normalized[i] = d
	}

	var div drawgl.ColorValue
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
		for i := range normalized {
			normalized[i] = normalized[i] / div
		}

	}

	return
}
