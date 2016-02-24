package transform

import (
	"encoding/json"
	"fmt"

	"github.com/urandom/drawgl"
	"github.com/urandom/graph"
	"github.com/urandom/graph/base"
	"golang.org/x/image/draw"
)

type Scale struct {
	base.Node

	opts         ScaleOptions
	interpolator draw.Interpolator
}

type ScaleOptions struct {
	Width, Height int
	Interpolator  InterpolatorOp
}

func NewScaleLinker(opts ScaleOptions) (graph.Linker, error) {
	interpolator, err := opts.Interpolator.Inst()

	if err != nil {
		return nil, err
	}

	if opts.Width <= 0 && opts.Height <= 0 {
		return nil, fmt.Errorf("Invalid width %f and height %f, at least one has to be positive", opts.Width, opts.Height)
	}

	return base.NewLinkerNode(Scale{
		Node:         base.NewNode(),
		opts:         opts,
		interpolator: interpolator,
	}), nil
}

func (n Scale) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
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
	src := r.Buffer
	res.Meta = r.Meta
	if src == nil {
		err = fmt.Errorf("no input buffer")
		return
	}

	b := src.Bounds()
	dr := src.Bounds()

	tW, tH := n.opts.Width, n.opts.Height
	if tW <= 0 || tH <= 0 {
		width, height := b.Max.X-b.Min.X, b.Max.Y-b.Min.Y

		if tW <= 0 {
			tW = tH * width / height
		} else {
			tH = tW * height / width
		}
	}

	dr.Max.X = dr.Min.X + tW
	dr.Max.Y = dr.Min.Y + tH

	buf = drawgl.NewFloatImage(dr)

	n.interpolator.Scale(buf, dr, src, b, draw.Src, nil)
}

func init() {
	graph.RegisterLinker("Scale", func(opts json.RawMessage) (graph.Linker, error) {
		var o ScaleOptions

		if err := json.Unmarshal([]byte(opts), &o); err != nil {
			return nil, fmt.Errorf("constructing Scale: %v", err)
		}

		return NewScaleLinker(o)
	})
}
