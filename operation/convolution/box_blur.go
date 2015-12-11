package convolution

import (
	"errors"
	"fmt"
	"image"
	"image/draw"

	"github.com/urandom/drawgl"
	"github.com/urandom/graph"
	"github.com/urandom/graph/base"
)

type BoxBlur struct {
	base.Node
	kernel HVKernel
	opts   BoxBlurOptions
}

type BoxBlurOptions struct {
	Radius  int
	Channel drawgl.Channel
	Mask    drawgl.Mask
	Linear  bool
}

func NewBoxBlurLinker(opts BoxBlurOptions) (graph.Linker, error) {
	if opts.Radius < 0 {
		return nil, errors.New("Radius cannot be less than 0")
	}

	k := make([]float32, 2*opts.Radius+1)
	for i := 0; i < 2*opts.Radius+1; i++ {
		k[i] = 1
	}
	kernel, err := NewHVKernel(k, k)
	if err != nil {
		return nil, err
	}

	opts.Channel.Normalize()
	return base.NewLinkerNode(BoxBlur{
		Node:   base.NewNode(),
		kernel: kernel,
		opts:   opts,
	}), nil
}

func (n BoxBlur) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
	var err error
	var buf *drawgl.FloatImage
	res := drawgl.Result{Id: n.Id()}

	defer func() {
		res.Buffer = buf
		if err != nil {
			res.Error = fmt.Errorf("Error applying box blur using %v: %v", n.opts, err)
		}
		output <- res

		wd.Close()
	}()

	r := buffers[graph.InputName]
	buf = r.Buffer
	res.Meta = r.Meta
	if buf == nil {
		err = fmt.Errorf("no input buffer")
		return
	}

	hk, _ := n.kernel.HNormalized()
	hcoeff := drawgl.ColorValue(hk[0])
	vk, _ := n.kernel.VNormalized()
	vcoeff := drawgl.ColorValue(vk[0])

	src := drawgl.CopyImage(buf)
	b := buf.Bounds()

	it := drawgl.DefaultRectangleIterator(b, n.opts.Linear)

	it.Iterate(n.opts.Mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		var rsum, gsum, bsum, asum drawgl.ColorValue
		var center drawgl.FloatColor
		if pt.X == b.Min.X {
			for cx := pt.X - n.opts.Radius; cx <= pt.X+n.opts.Radius; cx++ {
				mx := cx

				if mx < b.Min.X {
					mx = b.Min.X
				} else if mx >= b.Max.X {
					mx = b.Max.X - 1
				}

				c := src.UnsafeFloatAt(mx, pt.Y)
				if mx == pt.X {
					center = c
				}

				if n.opts.Channel.Is(drawgl.Red) {
					rsum += hcoeff * c.R
				}
				if n.opts.Channel.Is(drawgl.Green) {
					gsum += hcoeff * c.G
				}
				if n.opts.Channel.Is(drawgl.Blue) {
					bsum += hcoeff * c.B
				}
				if n.opts.Channel.Is(drawgl.Alpha) {
					asum += hcoeff * c.A
				}
			}
		} else {
			center = src.UnsafeFloatAt(pt.X, pt.Y)
			prev := src.UnsafeFloatAt(pt.X-1, pt.Y)

			var leftmost, rightmost drawgl.FloatColor
			if pt.X-n.opts.Radius < b.Min.X {
				leftmost = src.UnsafeFloatAt(b.Min.X, pt.Y)
			}
			if pt.X+n.opts.Radius > b.Max.X-1 {
				rightmost = src.UnsafeFloatAt(b.Max.X-1, pt.Y)
			}

			if n.opts.Channel.Is(drawgl.Red) {
				rsum += prev.R - hcoeff*leftmost.R + hcoeff*rightmost.R
			}
			if n.opts.Channel.Is(drawgl.Green) {
				gsum += prev.G - hcoeff*leftmost.G + hcoeff*rightmost.G
			}
			if n.opts.Channel.Is(drawgl.Blue) {
				bsum += prev.B - hcoeff*leftmost.B + hcoeff*rightmost.B
			}
			if n.opts.Channel.Is(drawgl.Alpha) {
				asum += prev.A - hcoeff*leftmost.A + hcoeff*rightmost.A
			}
		}

		cs := drawgl.FloatColor{
			R: rsum,
			G: gsum,
			B: bsum,
			A: asum,
		}

		buf.UnsafeSetColor(pt.X, pt.Y,
			drawgl.MaskColor(center, cs, n.opts.Channel, f, draw.Over))
	})

	src = drawgl.CopyImage(buf)
	it.VerticalIterate(n.opts.Mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		var rsum, gsum, bsum, asum drawgl.ColorValue
		var center drawgl.FloatColor
		if pt.Y == b.Min.Y {
			for cy := pt.Y - n.opts.Radius; cy <= pt.Y+n.opts.Radius; cy++ {
				my := cy

				if my < b.Min.Y {
					my = b.Min.Y
				} else if my >= b.Max.Y {
					my = b.Max.Y - 1
				}

				c := src.UnsafeFloatAt(pt.X, my)
				if my == pt.Y {
					center = c
				}

				if n.opts.Channel.Is(drawgl.Red) {
					rsum += vcoeff * c.R
				}
				if n.opts.Channel.Is(drawgl.Green) {
					gsum += vcoeff * c.G
				}
				if n.opts.Channel.Is(drawgl.Blue) {
					bsum += vcoeff * c.B
				}
				if n.opts.Channel.Is(drawgl.Alpha) {
					asum += vcoeff * c.A
				}
			}
		} else {
			center = src.UnsafeFloatAt(pt.X, pt.Y)
			prev := src.UnsafeFloatAt(pt.X, pt.Y-1)

			var leftmost, rightmost drawgl.FloatColor
			if pt.Y-n.opts.Radius < b.Min.Y {
				leftmost = src.UnsafeFloatAt(pt.X, b.Min.Y)
			}
			if pt.Y+n.opts.Radius > b.Max.Y-1 {
				rightmost = src.UnsafeFloatAt(pt.X, b.Max.Y-1)
			}

			if n.opts.Channel.Is(drawgl.Red) {
				rsum += prev.R - hcoeff*leftmost.R + hcoeff*rightmost.R
			}
			if n.opts.Channel.Is(drawgl.Green) {
				gsum += prev.G - hcoeff*leftmost.G + hcoeff*rightmost.G
			}
			if n.opts.Channel.Is(drawgl.Blue) {
				bsum += prev.B - hcoeff*leftmost.B + hcoeff*rightmost.B
			}
			if n.opts.Channel.Is(drawgl.Alpha) {
				asum += prev.A - hcoeff*leftmost.A + hcoeff*rightmost.A
			}
		}

		cs := drawgl.FloatColor{
			R: rsum,
			G: gsum,
			B: bsum,
			A: asum,
		}

		buf.UnsafeSetColor(pt.X, pt.Y,
			drawgl.MaskColor(center, cs, n.opts.Channel, f, draw.Over))
	})
}
