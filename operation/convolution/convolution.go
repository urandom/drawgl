package convolution

import (
	"errors"
	"fmt"
	"image"
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
	Region    image.Rectangle
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

	hasRegion := false
	if !n.opts.Region.Empty() {
		hasRegion = true
	}

	var weights []float64
	var offset float64
	if n.opts.Normalize {
		weights, offset = n.opts.Kernel.Normalized()
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

	it.Iterate(func(pt image.Point) {
		if hasRegion && !pt.In(n.opts.Region) {
			return
		}

		var rsum, gsum, bsum, asum float64
		for cy := pt.Y - half; cy <= pt.Y+half; cy++ {
			for cx := pt.X - half; cx <= pt.X+half; cx++ {
				coeff := weights[l-((cy-pt.Y+half)*size+cx-pt.X+half)-1]

				mx := cx
				my := cy
				if mx < b.Min.X {
					mx = 2*b.Min.X - mx
				} else if mx >= b.Max.X {
					mx = (b.Max.X-1)*2 - mx
				}

				if my < b.Min.Y {
					my = 2*b.Min.Y - my
				} else if my >= b.Max.Y {
					my = (b.Max.Y-1)*2 - my
				}

				c := src.NRGBA64At(mx, my)

				if n.opts.Channel&drawgl.Red > 0 {
					rsum += coeff * float64(c.R)
				}
				if n.opts.Channel&drawgl.Green > 0 {
					gsum += coeff * float64(c.G)
				}
				if n.opts.Channel&drawgl.Blue > 0 {
					bsum += coeff * float64(c.B)
				}
				if n.opts.Alpha && n.opts.Channel&drawgl.Alpha > 0 {
					asum += coeff * float64(c.A)
				}
			}
		}

		c := src.NRGBA64At(pt.X, pt.Y)

		if n.opts.Channel&drawgl.Red > 0 {
			c.R = drawgl.ClampUint16(rsum + offset)
		}
		if n.opts.Channel&drawgl.Green > 0 {
			c.G = drawgl.ClampUint16(gsum + offset)
		}
		if n.opts.Channel&drawgl.Blue > 0 {
			c.B = drawgl.ClampUint16(bsum + offset)
		}
		if n.opts.Alpha && n.opts.Channel&drawgl.Alpha > 0 {
			c.A = drawgl.ClampUint16(asum + offset)
		}

		buf.SetNRGBA64(pt.X, pt.Y, c)
	})

	res.Buffer = buf
}
