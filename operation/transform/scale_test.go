package transform_test

import (
	"testing"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/tests"
	"github.com/urandom/drawgl/operation/transform"
)

func TestScale(t *testing.T) {
	_, err := transform.NewScaleLinker(transform.ScaleOptions{Interpolator: "something"})
	if err == nil {
		t.Fatalf("Expected an error\n")
	}

	_, err = transform.NewScaleLinker(transform.ScaleOptions{})
	if err == nil {
		t.Fatalf("Expected an error\n")
	}

	l, err := transform.NewScaleLinker(transform.ScaleOptions{Width: 2, Height: 2})
	if err != nil {
		t.Fatalf("Error creating a scale linker: %v\n", err)
	}

	buffers := tests.ImageBuffers(t)
	p, wd, output := tests.PrepareLinker(l)

	go p.Process(wd, buffers, output)

	r := <-output
	if r.Error != nil {
		t.Fatalf("Error processing: %v\n", r.Error)
	}

	b := r.Buffer.Bounds()
	if b.Dx() != 4 || b.Dy() != 4 {
		t.Fatalf("Bounds size doesn't match (10, 10): %v", b)
	}

	exp := expectedScaleResult1()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := r.Buffer.FloatAt(x, y)
			if !c.ApproxEqual(exp[y][x]) {
				t.Fatalf("At %d:%d, color %v doesn't match %v\n", x, y, c, exp[y][x])
			}
		}
	}
}

func expectedScaleResult1() (c [2][2]drawgl.FloatColor) {
	c = [2][2]drawgl.FloatColor{
		[2]drawgl.FloatColor{
			drawgl.FloatColor{0.49999237, 0.49999237, 0.24998856, 1},
			drawgl.FloatColor{0.44606698, 0.6960708, 0.44606698, 1},
		},
		[2]drawgl.FloatColor{
			drawgl.FloatColor{0.24998856, 0.44606698, 0.5980316, 1},
			drawgl.FloatColor{0.6421454, 0.19607843, 0.5441062, 1},
		},
	}

	return
}
