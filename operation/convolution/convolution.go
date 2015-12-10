package convolution

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"math"

	"github.com/urandom/drawgl"
	"github.com/urandom/graph"
	"github.com/urandom/graph/base"
)

type Convolution struct {
	base.Node
	opts ConvolutionOptions
}

type ConvolutionOptions struct {
	Kernel    Kernel
	Channel   drawgl.Channel
	Normalize bool
	Mask      drawgl.Mask
	Alpha     bool
	Linear    bool
}

func NewConvolutionLinker(opts ConvolutionOptions) (graph.Linker, error) {
	if len(opts.Kernel.Weights()) == 0 {
		return nil, errors.New("Empty kernel")
	}

	opts.Channel.Normalize()

	return base.NewLinkerNode(Convolution{Node: base.NewNode(), opts: opts}), nil
}

func (n Convolution) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
	var err error
	res := drawgl.Result{Id: n.Id()}

	defer func() {
		if err != nil {
			res.Error = fmt.Errorf("Error applying convolution using %v: %v", n.opts, err)
		}
		output <- res

		wd.Close()
	}()

	r := buffers[graph.InputName]
	buf := r.Buffer
	res.Meta = r.Meta
	if buf == nil {
		err = fmt.Errorf("no input buffer")
		return
	}

	var weights []float32
	var offset drawgl.ColorValue
	if n.opts.Normalize {
		var o float32
		weights, o = n.opts.Kernel.Normalized()
		offset = drawgl.ColorValue(o)
	} else {
		weights = n.opts.Kernel.Weights()
	}

	src := drawgl.CopyImage(buf)
	b := buf.Bounds()
	l := len(weights)
	size := int(math.Sqrt(float64(l)))
	half := int(size / 2)

	var it drawgl.RectangleIterator
	if n.opts.Linear {
		it = drawgl.LinearRectangleIterator(b)
	} else {
		it = drawgl.ParallelRectangleIterator(b)
	}

	it.Iterate(n.opts.Mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		var rsum, gsum, bsum, asum drawgl.ColorValue
		var center drawgl.FloatColor
		for cy := pt.Y - half; cy <= pt.Y+half; cy++ {
			for cx := pt.X - half; cx <= pt.X+half; cx++ {
				coeff := drawgl.ColorValue(weights[l-((cy-pt.Y+half)*size+cx-pt.X+half)-1])

				mx := cx
				my := cy
				if mx < b.Min.X {
					mx = b.Min.X
				} else if mx >= b.Max.X {
					mx = b.Max.X - 1
				}

				if my < b.Min.Y {
					my = b.Min.Y
				} else if my >= b.Max.Y {
					my = b.Max.Y - 1
				}

				c := src.UnsafeFloatAt(mx, my)
				if mx == pt.X && my == pt.Y {
					center = c
				}

				if n.opts.Channel&drawgl.Red > 0 {
					rsum += coeff * c.R
				}
				if n.opts.Channel&drawgl.Green > 0 {
					gsum += coeff * c.G
				}
				if n.opts.Channel&drawgl.Blue > 0 {
					bsum += coeff * c.B
				}
				if n.opts.Alpha && n.opts.Channel&drawgl.Alpha > 0 {
					asum += coeff * c.A
				}
			}
		}

		offset = 0
		cs := drawgl.FloatColor{
			R: (rsum + offset),
			G: (gsum + offset),
			B: (bsum + offset),
			A: center.A,
		}

		if n.opts.Alpha {
			cs.A = asum + offset
		}

		buf.UnsafeSetColor(pt.X, pt.Y,
			drawgl.MaskColor(center, cs, n.opts.Channel, f, draw.Over))
	})

	res.Buffer = buf
}
