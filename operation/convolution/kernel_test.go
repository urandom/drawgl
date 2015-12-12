package convolution_test

import (
	"testing"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/convolution"
)

func TestKernel(t *testing.T) {
	_, err := convolution.NewKernel([]float32{})
	if err == nil {
		t.Fatalf("Expected an error\n")
	}

	_, err = convolution.NewKernel([]float32{1, 2, 3, 4, 5})
	if err == nil {
		t.Fatalf("Expected an error\n")
	}

	data := []float32{1, 2, 1, 2, 4, 2, 1, 2, 1}
	k, err := convolution.NewKernel(data)
	if err != nil {
		t.Fatalf("Error creating kernel: %v\n", err)
	}

	w := k.Weights()
	if len(w) != len(data) {
		t.Fatalf("Expected %d, got %d\n", len(data), len(w))
	}
	for i := range w {
		if w[i] != drawgl.ColorValue(data[i]) {
			t.Fatalf("Expected %v at %d, got %v\n", w[i], i, data[i])
		}
	}

	n, o := k.Normalized()
	if o != 0 {
		t.Fatalf("Expected 0, got %d\n", o)
	}

	if len(n) != len(data) {
		t.Fatalf("Expected %d, got %d\n", len(data), len(n))
	}
	for i := range n {
		if n[i] != drawgl.ColorValue(data[i])/16 {
			t.Fatalf("Expected %v at %d, got %v\n", n[i], i, data[i]/16)
		}
	}

}
