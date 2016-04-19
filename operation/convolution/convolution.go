package convolution

import (
	"encoding/json"
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
		return nil, errors.New("empty kernel")
	}

	opts.Channel = opts.Channel.Normalize()

	return base.NewLinkerNode(Convolution{Node: base.NewNode(), opts: opts}), nil
}

func (n Convolution) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
	var err error
	var buf *drawgl.FloatImage
	res := drawgl.Result{Id: n.Id()}

	defer func() {
		res.Buffer = buf
		if err != nil {
			res.Error = fmt.Errorf("applying convolution using %v: %v", n.opts, err)
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

		var center, acc drawgl.FloatColor
		for cy := pt.Y - half; cy <= pt.Y+half; cy++ {
			for cx := pt.X - half; cx <= pt.X+half; cx++ {
				coeff := weights[l-((cy-pt.Y+half)*size+cx-pt.X+half)-1]

				mx, my := drawgl.TranslateCoords(cx, cy, b, drawgl.Extend)

				c := src.UnsafeFloatAt(mx, my)
				if mx == pt.X && my == pt.Y {
					center = c
				}

				acc = ColorAccumulator(acc, c, drawgl.FloatColor{}, coeff, n.opts.Channel)
			}
		}

		cs := drawgl.FloatColor{
			R: acc.R + offset,
			G: acc.G + offset,
			B: acc.B + offset,
			A: acc.A + offset,
		}

		buf.UnsafeSetColor(pt.X, pt.Y,
			drawgl.MaskColor(center, cs, n.opts.Channel, f, draw.Over))
	})
}

func ColorAccumulator(acc, add, sub drawgl.FloatColor, coeff drawgl.ColorValue, channel drawgl.Channel) drawgl.FloatColor {
	if channel.Is(drawgl.Red) {
		acc.R += coeff*add.R - coeff*sub.R
	}
	if channel.Is(drawgl.Green) {
		acc.G += coeff*add.G - coeff*sub.G
	}
	if channel.Is(drawgl.Blue) {
		acc.B += coeff*add.B - coeff*sub.B
	}
	if channel.Is(drawgl.Alpha) {
		acc.A += coeff*add.A - coeff*sub.A
	}
	return acc
}

func init() {
	type jsonOptions struct {
		Kernel    kernel
		Channel   drawgl.Channel
		Normalize bool
		Mask      drawgl.Mask
		Linear    bool
	}

	graph.RegisterLinker("Convolution", func(opts json.RawMessage) (graph.Linker, error) {
		var o ConvolutionOptions
		var jsono jsonOptions

		if err := json.Unmarshal([]byte(opts), &jsono); err != nil {
			return nil, fmt.Errorf("constructing Convolution: %v", err)
		}

		o.Kernel = jsono.Kernel
		o.Channel = jsono.Channel
		o.Normalize = jsono.Normalize
		o.Mask = jsono.Mask
		o.Linear = jsono.Linear

		return NewConvolutionLinker(o)
	})
}
