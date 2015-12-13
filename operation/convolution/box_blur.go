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
	opts BoxBlurOptions
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
	} else if opts.Radius == 0 {
		opts.Radius = 4
	}

	opts.Channel.Normalize()
	return base.NewLinkerNode(BoxBlur{
		Node: base.NewNode(),
		opts: opts,
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

	coeff := 1 / drawgl.ColorValue(2*n.opts.Radius+1)

	src := drawgl.CopyImage(buf)
	b := buf.Bounds()

	it := drawgl.DefaultRectangleIterator(b, n.opts.Linear)
	// it := drawgl.LinearRectangleIterator(b)

	it.Iterate(n.opts.Mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		var rsum, gsum, bsum, asum drawgl.ColorValue
		var center drawgl.FloatColor
		if pt.X == b.Min.X {
			for cx := pt.X - n.opts.Radius; cx <= pt.X+n.opts.Radius; cx++ {
				mx, _ := drawgl.TranslateCoords(cx, pt.Y, b, drawgl.Extend)

				c := src.UnsafeFloatAt(mx, pt.Y)
				if mx == pt.X {
					center = c
				}

				if n.opts.Channel.Is(drawgl.Red) {
					rsum += coeff * c.R
				}
				if n.opts.Channel.Is(drawgl.Green) {
					gsum += coeff * c.G
				}
				if n.opts.Channel.Is(drawgl.Blue) {
					bsum += coeff * c.B
				}
				if n.opts.Channel.Is(drawgl.Alpha) {
					asum += coeff * c.A
				}
			}
		} else {
			center = src.UnsafeFloatAt(pt.X, pt.Y)
			prev := buf.UnsafeFloatAt(pt.X-1, pt.Y)

			var leftmost, rightmost drawgl.FloatColor
			if pt.X-n.opts.Radius < b.Min.X {
				leftmost = src.UnsafeFloatAt(b.Min.X, pt.Y)
			} else {
				leftmost = src.UnsafeFloatAt(pt.X-n.opts.Radius, pt.Y)
			}
			if pt.X+n.opts.Radius > b.Max.X-1 {
				rightmost = src.UnsafeFloatAt(b.Max.X-1, pt.Y)
			} else {
				rightmost = src.UnsafeFloatAt(pt.X+n.opts.Radius, pt.Y)
			}

			if n.opts.Channel.Is(drawgl.Red) {
				rsum += prev.R - coeff*leftmost.R + coeff*rightmost.R
			}
			if n.opts.Channel.Is(drawgl.Green) {
				gsum += prev.G - coeff*leftmost.G + coeff*rightmost.G
			}
			if n.opts.Channel.Is(drawgl.Blue) {
				bsum += prev.B - coeff*leftmost.B + coeff*rightmost.B
			}
			if n.opts.Channel.Is(drawgl.Alpha) {
				asum += prev.A - coeff*leftmost.A + coeff*rightmost.A
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
				_, my := drawgl.TranslateCoords(pt.X, cy, b, drawgl.Extend)

				c := src.UnsafeFloatAt(pt.X, my)
				if my == pt.Y {
					center = c
				}

				if n.opts.Channel.Is(drawgl.Red) {
					rsum += coeff * c.R
				}
				if n.opts.Channel.Is(drawgl.Green) {
					gsum += coeff * c.G
				}
				if n.opts.Channel.Is(drawgl.Blue) {
					bsum += coeff * c.B
				}
				if n.opts.Channel.Is(drawgl.Alpha) {
					asum += coeff * c.A
				}
			}
		} else {
			center = src.UnsafeFloatAt(pt.X, pt.Y)
			prev := buf.UnsafeFloatAt(pt.X, pt.Y-1)

			var leftmost, rightmost drawgl.FloatColor
			if pt.Y-n.opts.Radius < b.Min.Y {
				leftmost = src.UnsafeFloatAt(pt.X, b.Min.Y)
			} else {
				leftmost = src.UnsafeFloatAt(pt.X, pt.Y-n.opts.Radius)
			}
			if pt.Y+n.opts.Radius > b.Max.Y-1 {
				rightmost = src.UnsafeFloatAt(pt.X, b.Max.Y-1)
			} else {
				rightmost = src.UnsafeFloatAt(pt.X, pt.Y+n.opts.Radius)
			}

			if n.opts.Channel.Is(drawgl.Red) {
				rsum += prev.R - coeff*leftmost.R + coeff*rightmost.R
			}
			if n.opts.Channel.Is(drawgl.Green) {
				gsum += prev.G - coeff*leftmost.G + coeff*rightmost.G
			}
			if n.opts.Channel.Is(drawgl.Blue) {
				bsum += prev.B - coeff*leftmost.B + coeff*rightmost.B
			}
			if n.opts.Channel.Is(drawgl.Alpha) {
				asum += prev.A - coeff*leftmost.A + coeff*rightmost.A
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
