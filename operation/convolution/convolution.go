package convolution

import (
	"errors"
	"fmt"
	"image"
	"image/color"
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
}

func NewConvolutionLinker(opts ConvolutionOptions) (graph.Linker, error) {
	if len(opts.Kernel.Weights()) == 0 {
		return nil, errors.New("Empty kernel")
	}

	if opts.Channel == drawgl.All {
		opts.Channel = drawgl.Red | drawgl.Green | drawgl.Blue | drawgl.Alpha
	}

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

	b := buf.Bounds()
	l := len(weights)
	size := int(math.Sqrt(float64(l)))
	half := int(size / 2)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if hasRegion && !image.Pt(x, y).In(n.opts.Region) {
				continue
			}

			var rsum, gsum, bsum, asum float64
			for cy := y - half; cy <= y+half; cy++ {
				for cx := x - half; cx <= x+half; cx++ {
					coeff := weights[l-((cy-y+half)*size+cx-x+half)-1]

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

					r, g, b, a := color.RGBA64Model.Convert(
						buf.At(mx, my)).RGBA()

					if n.opts.Channel&drawgl.Red > 0 {
						rsum += coeff * float64(r)
					}
					if n.opts.Channel&drawgl.Green > 0 {
						gsum += coeff * float64(g)
					}
					if n.opts.Channel&drawgl.Blue > 0 {
						bsum += coeff * float64(b)
					}
					if n.opts.Channel&drawgl.Alpha > 0 {
						asum += coeff * float64(a)
					}
				}
			}

			r, g, b, a := color.RGBA64Model.Convert(buf.At(x, y)).RGBA()

			if n.opts.Channel&drawgl.Red > 0 {
				r = drawgl.ClampUint32(rsum + offset)
			}
			if n.opts.Channel&drawgl.Green > 0 {
				g = drawgl.ClampUint32(gsum + offset)
			}
			if n.opts.Channel&drawgl.Blue > 0 {
				b = drawgl.ClampUint32(bsum + offset)
			}
			if n.opts.Channel&drawgl.Alpha > 0 {
				a = drawgl.ClampUint32(asum + offset)
			}

			c := color.RGBA64{
				drawgl.ClampUint16(r),
				drawgl.ClampUint16(g),
				drawgl.ClampUint16(b),
				drawgl.ClampUint16(a),
			}

			buf.Set(x, y, c)
		}
	}

	res.Buffer = buf
}
