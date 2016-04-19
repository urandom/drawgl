package convolution_test

import (
	"testing"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/convolution"
	"github.com/urandom/drawgl/operation/tests"
)

func TestBoxBlur(t *testing.T) {
	_, err := convolution.NewBoxBlurLinker(convolution.BoxBlurOptions{Radius: -5})
	if err == nil {
		t.Fatalf("Expected an error\n")
	}

	l, err := convolution.NewBoxBlurLinker(convolution.BoxBlurOptions{})
	if err != nil {
		t.Fatalf("Error creating a convolution linker: %v\n", err)
	}

	buffers := tests.ImageBuffers(t)
	p, wd, output := tests.PrepareLinker(l)

	go p.Process(wd, buffers, output)

	r := <-output
	if r.Error != nil {
		t.Fatalf("Error processing: %v\n", r.Error)
	}

	exp := expectedBoxBlurResult1()
	b := r.Buffer.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := r.Buffer.FloatAt(x, y)
			if !c.ApproxEqual(exp[y][x]) {
				t.Fatalf("At %d:%d, color %v doesn't match %v\n", x, y, c, exp[y][x])
			}
		}
	}
}

func expectedBoxBlurResult1() (c [4][4]drawgl.FloatColor) {
	c = [4][4]drawgl.FloatColor{
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.6223675, 0.6020335, 0.46550477, 1},
			drawgl.FloatColor{0.59670776, 0.59089816, 0.41079646, 1},
			drawgl.FloatColor{0.5710481, 0.5797629, 0.35608816, 1},
			drawgl.FloatColor{0.54538846, 0.5686276, 0.30137986, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.5553135, 0.5279594, 0.41345924, 1},
			drawgl.FloatColor{0.5468409, 0.5243283, 0.37109664, 1},
			drawgl.FloatColor{0.53836834, 0.52069724, 0.328734, 1},
			drawgl.FloatColor{0.52989584, 0.51706624, 0.28637138, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.48825955, 0.45388532, 0.36141372, 1},
			drawgl.FloatColor{0.49697405, 0.45775843, 0.33139682, 1},
			drawgl.FloatColor{0.5056886, 0.46163163, 0.30137986, 1},
			drawgl.FloatColor{0.5144032, 0.46550485, 0.2713629, 1},
		},
		[4]drawgl.FloatColor{
			drawgl.FloatColor{0.42120558, 0.37981123, 0.3093682, 1},
			drawgl.FloatColor{0.4471072, 0.39118856, 0.291697, 1},
			drawgl.FloatColor{0.47300887, 0.40256602, 0.2740257, 1},
			drawgl.FloatColor{0.4989106, 0.41394347, 0.25635442, 1},
		},
	}

	return
}
