package convolution_test

import (
	"testing"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/convolution"
	"github.com/urandom/drawgl/operation/tests"
	"github.com/urandom/graph"
)

func TestConvolution(t *testing.T) {
	_, err := convolution.NewConvolutionLinker(convolution.ConvolutionOptions{})
	if err == nil {
		t.Fatalf("Expected an error\n")
	}

	k := kernel1()
	l, err := convolution.NewConvolutionLinker(convolution.ConvolutionOptions{Kernel: k})
	if err != nil {
		t.Fatalf("Error creating a convolution linker: %v\n", err)
	}

	buffers := getImageBuffers(t)
	p, wd, output := tests.PrepareLinker(l)

	go p.Process(wd, buffers, output)

	r := <-output
	if r.Error != nil {
		t.Fatalf("Error processing: %v\n", r.Error)
	}

	exp := expectedResult1()
	b := r.Buffer.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := r.Buffer.FloatAt(x, y)
			if c != exp[y][x] {
				t.Fatalf("At %d:%d, color %v doesn't match %v\n", x, y, c, exp[y][x])
			}
		}
	}
}

func kernel1() convolution.Kernel {
	k, _ := convolution.NewKernel([]float32{
		0, 1, 0,
		1, 2, 1,
		0, 1, 0,
	})

	return k
}

func expectedResult1() (c [4][4]drawgl.FloatColor) {
	c = [4][4]drawgl.FloatColor{
		[4]drawgl.FloatColor{
			drawgl.FloatColor{5, 4, 4, 1},
			drawgl.FloatColor{5, 3, 1, 1},
			drawgl.FloatColor{4.392157, 4.7843137, 1, 1},
			drawgl.FloatColor{2.9607842, 4.1372547, 0.78431374, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{2, 3, 2, 1},
			drawgl.FloatColor{1, 3, 2, 1},
			drawgl.FloatColor{2.3921568, 4, 3.7843137, 1},
			drawgl.FloatColor{1.5686275, 2.1764705, 4.1372547, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{3, 3, 4, 1},
			drawgl.FloatColor{2, 2.7843137, 4.392157, 1},
			drawgl.FloatColor{2.7843137, 1.3921568, 5.1764708, 1},
			drawgl.FloatColor{2.1764705, 1.5686275, 4.1372547, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{1, 1.7843137, 1.3921568, 1},
			drawgl.FloatColor{0.78431374, 2.3529413, 2.5686274, 1},
			drawgl.FloatColor{4.1372547, 1.1764706, 2.5686274, 1},
			drawgl.FloatColor{3.9215686, 1.9607843, 1.1764706, 1},
		},
	}

	return
}

func getImageBuffers(t *testing.T) map[graph.ConnectorName]drawgl.Result {
	img, err := tests.ReadTestData()
	if err != nil {
		t.Fatalf("Error reading test image: %v\n", err)
	}

	return map[graph.ConnectorName]drawgl.Result{
		graph.InputName: drawgl.Result{Buffer: img},
	}
}
