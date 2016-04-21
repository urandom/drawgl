package transform_test

import (
	"testing"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/tests"
	"github.com/urandom/drawgl/operation/transform"
)

func TestScale(t *testing.T) {
	_, err := transform.NewScaleLinker(transform.ScaleOptions{})
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

func expectedScaleResult1() (c [4][4]drawgl.FloatColor) {
	c = [4][4]drawgl.FloatColor{
		[4]drawgl.FloatColor{
			drawgl.FloatColor{1, 1, 1, 1},
			drawgl.FloatColor{1, 1, 0, 1},
			drawgl.FloatColor{0, 0, 0, 0},
			drawgl.FloatColor{0, 0, 0, 0},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{1, 1, 1, 1},
			drawgl.FloatColor{1, 0, 1, 1},
			drawgl.FloatColor{0, 0, 0, 0},
			drawgl.FloatColor{0, 0, 0, 0},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0, 0, 0, 0},
			drawgl.FloatColor{0, 0, 0, 0},
			drawgl.FloatColor{0, 0, 0, 0},
			drawgl.FloatColor{0, 0, 0, 0},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0, 0, 0, 0},
			drawgl.FloatColor{0, 0, 0, 0},
			drawgl.FloatColor{0, 0, 0, 0},
			drawgl.FloatColor{0, 0, 0, 0},
		},
	}

	return
}
