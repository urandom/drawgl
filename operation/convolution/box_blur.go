package convolution

import (
	"encoding/json"
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

	edgeHandler := drawgl.Extend

	it.Iterate(n.opts.Mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		var center, acc drawgl.FloatColor
		if pt.X == b.Min.X {
			for cx := pt.X - n.opts.Radius; cx <= pt.X+n.opts.Radius; cx++ {
				mx, _ := drawgl.TranslateCoords(cx, pt.Y, b, edgeHandler)

				c := src.UnsafeFloatAt(mx, pt.Y)
				if mx == pt.X {
					center = c
				}

				acc = ColorAccumulator(acc, c, drawgl.FloatColor{}, coeff, n.opts.Channel)
			}
		} else {
			center = src.UnsafeFloatAt(pt.X, pt.Y)
			prev := buf.UnsafeFloatAt(pt.X-1, pt.Y)

			mx, _ := drawgl.TranslateCoords(pt.X-n.opts.Radius-1, pt.Y, b, edgeHandler)
			leftmost := src.UnsafeFloatAt(mx, pt.Y)

			mx, _ = drawgl.TranslateCoords(pt.X+n.opts.Radius, pt.Y, b, edgeHandler)
			rightmost := src.UnsafeFloatAt(mx, pt.Y)

			acc = ColorAccumulator(prev, rightmost, leftmost, coeff, n.opts.Channel)
		}

		buf.UnsafeSetColor(pt.X, pt.Y,
			drawgl.MaskColor(center, acc, n.opts.Channel, f, draw.Over))
	})

	src = drawgl.CopyImage(buf)
	it.VerticalIterate(n.opts.Mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		var center, acc drawgl.FloatColor
		if pt.Y == b.Min.Y {
			for cy := pt.Y - n.opts.Radius; cy <= pt.Y+n.opts.Radius; cy++ {
				_, my := drawgl.TranslateCoords(pt.X, cy, b, edgeHandler)

				c := src.UnsafeFloatAt(pt.X, my)
				if my == pt.Y {
					center = c
				}

				acc = ColorAccumulator(acc, c, drawgl.FloatColor{}, coeff, n.opts.Channel)
			}
		} else {
			center = src.UnsafeFloatAt(pt.X, pt.Y)
			prev := buf.UnsafeFloatAt(pt.X, pt.Y-1)

			_, my := drawgl.TranslateCoords(pt.X, pt.Y-n.opts.Radius-1, b, edgeHandler)
			leftmost := src.UnsafeFloatAt(pt.X, my)

			_, my = drawgl.TranslateCoords(pt.X, pt.Y+n.opts.Radius, b, edgeHandler)
			rightmost := src.UnsafeFloatAt(pt.X, my)

			acc = ColorAccumulator(prev, rightmost, leftmost, coeff, n.opts.Channel)
		}

		buf.UnsafeSetColor(pt.X, pt.Y,
			drawgl.MaskColor(center, acc, n.opts.Channel, f, draw.Over))
	})
}

func init() {
	drawgl.RegisterOperation("BoxBlur", func(opts json.RawMessage) (graph.Linker, error) {
		var o BoxBlurOptions

		if err := json.Unmarshal([]byte(opts), &o); err != nil {
			return nil, fmt.Errorf("constructing BoxBlur: %v", err)
		}

		return NewBoxBlurLinker(o)
	})
}
