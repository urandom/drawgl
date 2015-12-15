package transform

import (
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

	if opts.Width <= 0 {
		return nil, fmt.Errorf("Invalid width %f", opts.Width)
	}

	if opts.Height <= 0 {
		return nil, fmt.Errorf("Invalid height %f", opts.Height)
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

	dr.Max.X = dr.Min.X + n.opts.Width
	dr.Max.Y = dr.Min.Y + n.opts.Height

	buf = drawgl.NewFloatImage(dr)

	n.interpolator.Scale(buf, dr, src, b, draw.Src, nil)
}
