package transform_test

import (
	"testing"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/tests"
	"github.com/urandom/drawgl/operation/transform"
)

func TestTransformInvalid(t *testing.T) {
	_, err := transform.NewTransformLinker(transform.TransformOptions{Operator: 99})
	if err == nil {
		t.Fatalf("Expected an error\n")
	}

	_, err = transform.NewTransformLinker(transform.TransformOptions{})
	if err == nil {
		t.Fatalf("Expected an error\n")
	}
}

func TestTransformFlipH(t *testing.T) {
	l, err := transform.NewTransformLinker(transform.TransformOptions{Operator: transform.FlipHOperator})
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
		t.Fatalf("Bounds size doesn't match (4, 4): %v", b)
	}

	exp := expectedTransformFlipHResult()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := r.Buffer.FloatAt(x, y)
			if !c.ApproxEqual(exp[y][x]) {
				t.Errorf("At %d:%d, color %v doesn't match %v\n", x, y, c, exp[y][x])
			}
		}
	}
}

func expectedTransformFlipHResult() (c [4][4]drawgl.FloatColor) {
	c = [4][4]drawgl.FloatColor{
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.39215687, 0.78431374, 0, 1},
			drawgl.FloatColor{1, 1, 0, 1},
			drawgl.FloatColor{1, 0, 0, 1},
			drawgl.FloatColor{1, 1, 1, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.39215687, 0, 0.78431374, 1},
			drawgl.FloatColor{0, 1, 1, 1},
			drawgl.FloatColor{0, 1, 0, 1},
			drawgl.FloatColor{0, 0, 0, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0, 0.39215687, 0.78431374, 1},
			drawgl.FloatColor{1, 0, 1, 1},
			drawgl.FloatColor{0, 0, 1, 1},
			drawgl.FloatColor{1, 1, 1, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.78431374, 0.39215687, 0, 1},
			drawgl.FloatColor{0.78431374, 0, 0.39215687, 1},
			drawgl.FloatColor{0, 0.78431374, 0.39215687, 1},
			drawgl.FloatColor{0, 0, 0, 1},
		},
	}

	return
}

func TestTransformFlipV(t *testing.T) {
	l, err := transform.NewTransformLinker(transform.TransformOptions{Operator: transform.FlipVOperator})
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
		t.Fatalf("Bounds size doesn't match (4, 4): %v", b)
	}

	exp := expectedTransformFlipVResult()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := r.Buffer.FloatAt(x, y)
			if c != exp[y][x] {
				t.Errorf("At %d:%d, color %v doesn't match %v\n", x, y, c, exp[y][x])
			}
		}
	}
}

func expectedTransformFlipVResult() (c [4][4]drawgl.FloatColor) {
	c = [4][4]drawgl.FloatColor{
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0, 0, 0, 1},
			drawgl.FloatColor{0, 0.78431374, 0.39215687, 1},
			drawgl.FloatColor{0.78431374, 0, 0.39215687, 1},
			drawgl.FloatColor{0.78431374, 0.39215687, 0, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{1, 1, 1, 1},
			drawgl.FloatColor{0, 0, 1, 1},
			drawgl.FloatColor{1, 0, 1, 1},
			drawgl.FloatColor{0, 0.39215687, 0.78431374, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0, 0, 0, 1},
			drawgl.FloatColor{0, 1, 0, 1},
			drawgl.FloatColor{0, 1, 1, 1},
			drawgl.FloatColor{0.39215687, 0, 0.78431374, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{1, 1, 1, 1},
			drawgl.FloatColor{1, 0, 0, 1},
			drawgl.FloatColor{1, 1, 0, 1},
			drawgl.FloatColor{0.39215687, 0.78431374, 0, 1},
		},
	}

	return
}

func TestTransformTranspose(t *testing.T) {
	l, err := transform.NewTransformLinker(transform.TransformOptions{Operator: transform.TransposeOperator})
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
		t.Fatalf("Bounds size doesn't match (4, 4): %v", b)
	}

	exp := expectedTransformTransposeResult()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := r.Buffer.FloatAt(x, y)
			if c != exp[y][x] {
				t.Errorf("At %d:%d, color %v doesn't match %v\n", x, y, c, exp[y][x])
			}
		}
	}
}

func expectedTransformTransposeResult() (c [4][4]drawgl.FloatColor) {
	c = [4][4]drawgl.FloatColor{
		[4]drawgl.FloatColor{
			drawgl.FloatColor{1, 1, 1, 1},
			drawgl.FloatColor{0, 0, 0, 1},
			drawgl.FloatColor{1, 1, 1, 1},
			drawgl.FloatColor{0, 0, 0, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{1, 0, 0, 1},
			drawgl.FloatColor{0, 1, 0, 1},
			drawgl.FloatColor{0, 0, 1, 1},
			drawgl.FloatColor{0, 0.78431374, 0.39215687, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{1, 1, 0, 1},
			drawgl.FloatColor{0, 1, 1, 1},
			drawgl.FloatColor{1, 0, 1, 1},
			drawgl.FloatColor{0.78431374, 0, 0.39215687, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.39215687, 0.78431374, 0, 1},
			drawgl.FloatColor{0.39215687, 0, 0.78431374, 1},
			drawgl.FloatColor{0, 0.39215687, 0.78431374, 1},
			drawgl.FloatColor{0.78431374, 0.39215687, 0, 1},
		},
	}

	return
}

func TestTransformTransverse(t *testing.T) {
	l, err := transform.NewTransformLinker(transform.TransformOptions{Operator: transform.TransverseOperator})
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
		t.Fatalf("Bounds size doesn't match (4, 4): %v", b)
	}

	exp := expectedTransformTransverseResult()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := r.Buffer.FloatAt(x, y)
			if c != exp[y][x] {
				t.Errorf("At %d:%d, color %v doesn't match %v\n", x, y, c, exp[y][x])
			}
		}
	}
}

func expectedTransformTransverseResult() (c [4][4]drawgl.FloatColor) {
	c = [4][4]drawgl.FloatColor{
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.78431374, 0.39215687, 0, 1},
			drawgl.FloatColor{0, 0.39215687, 0.78431374, 1},
			drawgl.FloatColor{0.39215687, 0, 0.78431374, 1},
			drawgl.FloatColor{0.39215687, 0.78431374, 0, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.78431374, 0, 0.39215687, 1},
			drawgl.FloatColor{1, 0, 1, 1},
			drawgl.FloatColor{0, 1, 1, 1},
			drawgl.FloatColor{1, 1, 0, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0, 0.78431374, 0.39215687, 1},
			drawgl.FloatColor{0, 0, 1, 1},
			drawgl.FloatColor{0, 1, 0, 1},
			drawgl.FloatColor{1, 0, 0, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0, 0, 0, 1},
			drawgl.FloatColor{1, 1, 1, 1},
			drawgl.FloatColor{0, 0, 0, 1},
			drawgl.FloatColor{1, 1, 1, 1},
		},
	}

	return
}

func TestTransformRotate90(t *testing.T) {
	l, err := transform.NewTransformLinker(transform.TransformOptions{Operator: transform.Rotate90Operator})
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
		t.Fatalf("Bounds size doesn't match (4, 4): %v", b)
	}

	exp := expectedTransformRotate90Result()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := r.Buffer.FloatAt(x, y)
			if c != exp[y][x] {
				t.Errorf("At %d:%d, color %v doesn't match %v\n", x, y, c, exp[y][x])
			}
		}
	}
}

func expectedTransformRotate90Result() (c [4][4]drawgl.FloatColor) {
	c = [4][4]drawgl.FloatColor{
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0, 0, 0, 1},
			drawgl.FloatColor{1, 1, 1, 1},
			drawgl.FloatColor{0, 0, 0, 1},
			drawgl.FloatColor{1, 1, 1, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0, 0.78431374, 0.39215687, 1},
			drawgl.FloatColor{0, 0, 1, 1},
			drawgl.FloatColor{0, 1, 0, 1},
			drawgl.FloatColor{1, 0, 0, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.78431374, 0, 0.39215687, 1},
			drawgl.FloatColor{1, 0, 1, 1},
			drawgl.FloatColor{0, 1, 1, 1},
			drawgl.FloatColor{1, 1, 0, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.78431374, 0.39215687, 0, 1},
			drawgl.FloatColor{0, 0.39215687, 0.78431374, 1},
			drawgl.FloatColor{0.39215687, 0, 0.78431374, 1},
			drawgl.FloatColor{0.39215687, 0.78431374, 0, 1},
		},
	}

	return
}

func TestTransformRotate180(t *testing.T) {
	l, err := transform.NewTransformLinker(transform.TransformOptions{Operator: transform.Rotate180Operator})
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
		t.Fatalf("Bounds size doesn't match (4, 4): %v", b)
	}

	exp := expectedTransformRotate180Result()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := r.Buffer.FloatAt(x, y)
			if c != exp[y][x] {
				t.Errorf("At %d:%d, color %v doesn't match %v\n", x, y, c, exp[y][x])
			}
		}
	}
}

func expectedTransformRotate180Result() (c [4][4]drawgl.FloatColor) {
	c = [4][4]drawgl.FloatColor{
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.78431374, 0.39215687, 0, 1},
			drawgl.FloatColor{0.78431374, 0, 0.39215687, 1},
			drawgl.FloatColor{0, 0.78431374, 0.39215687, 1},
			drawgl.FloatColor{0, 0, 0, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0, 0.39215687, 0.78431374, 1},
			drawgl.FloatColor{1, 0, 1, 1},
			drawgl.FloatColor{0, 0, 1, 1},
			drawgl.FloatColor{1, 1, 1, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.39215687, 0, 0.78431374, 1},
			drawgl.FloatColor{0, 1, 1, 1},
			drawgl.FloatColor{0, 1, 0, 1},
			drawgl.FloatColor{0, 0, 0, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.39215687, 0.78431374, 0, 1},
			drawgl.FloatColor{1, 1, 0, 1},
			drawgl.FloatColor{1, 0, 0, 1},
			drawgl.FloatColor{1, 1, 1, 1},
		},
	}

	return
}

func TestTransformRotate270(t *testing.T) {
	l, err := transform.NewTransformLinker(transform.TransformOptions{Operator: transform.Rotate270Operator})
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
		t.Fatalf("Bounds size doesn't match (4, 4): %v", b)
	}

	exp := expectedTransformRotate270Result()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := r.Buffer.FloatAt(x, y)
			if c != exp[y][x] {
				t.Errorf("At %d:%d, color %v doesn't match %v\n", x, y, c, exp[y][x])
			}
		}
	}
}

func expectedTransformRotate270Result() (c [4][4]drawgl.FloatColor) {
	c = [4][4]drawgl.FloatColor{
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.39215687, 0.78431374, 0, 1},
			drawgl.FloatColor{0.39215687, 0, 0.78431374, 1},
			drawgl.FloatColor{0, 0.39215687, 0.78431374, 1},
			drawgl.FloatColor{0.78431374, 0.39215687, 0, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{1, 1, 0, 1},
			drawgl.FloatColor{0, 1, 1, 1},
			drawgl.FloatColor{1, 0, 1, 1},
			drawgl.FloatColor{0.78431374, 0, 0.39215687, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{1, 0, 0, 1},
			drawgl.FloatColor{0, 1, 0, 1},
			drawgl.FloatColor{0, 0, 1, 1},
			drawgl.FloatColor{0, 0.78431374, 0.39215687, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{1, 1, 1, 1},
			drawgl.FloatColor{0, 0, 0, 1},
			drawgl.FloatColor{1, 1, 1, 1},
			drawgl.FloatColor{0, 0, 0, 1},
		},
	}

	return
}
