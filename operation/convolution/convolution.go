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
	Linear    bool
}

func NewConvolutionLinker(opts ConvolutionOptions) (graph.Linker, error) {
	if opts.Kernel == nil || len(opts.Kernel.Weights()) == 0 {
		return nil, errors.New("Empty kernel")
	}

	opts.Channel.Normalize()

	return base.NewLinkerNode(Convolution{Node: base.NewNode(), opts: opts}), nil
}

func (n Convolution) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
	var err error
	var buf *drawgl.FloatImage
	res := drawgl.Result{Id: n.Id()}

	defer func() {
		res.Buffer = buf
		if err != nil {
			res.Error = fmt.Errorf("Error applying convolution using %v: %v", n.opts, err)
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

	var weights []drawgl.ColorValue
	var offset drawgl.ColorValue
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

	it := drawgl.DefaultRectangleIterator(b, n.opts.Linear)

	it.Iterate(n.opts.Mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		var rsum, gsum, bsum, asum drawgl.ColorValue
		var center drawgl.FloatColor
		for cy := pt.Y - half; cy <= pt.Y+half; cy++ {
			for cx := pt.X - half; cx <= pt.X+half; cx++ {
				coeff := weights[l-((cy-pt.Y+half)*size+cx-pt.X+half)-1]

				mx, my := drawgl.TranslateCoords(cx, cy, b, drawgl.Extend)

				c := src.UnsafeFloatAt(mx, my)
				if mx == pt.X && my == pt.Y {
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
		}

		cs := drawgl.FloatColor{
			R: rsum + offset,
			G: gsum + offset,
			B: bsum + offset,
			A: asum + offset,
		}

		buf.UnsafeSetColor(pt.X, pt.Y,
			drawgl.MaskColor(center, cs, n.opts.Channel, f, draw.Over))
	})
}
