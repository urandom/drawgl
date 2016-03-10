package transform

import (
	"encoding/json"
	"fmt"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/interpolator"
	"github.com/urandom/drawgl/operation/transform/matrix"
	"github.com/urandom/graph"
	"github.com/urandom/graph/base"
)

type Scale struct {
	base.Node

	opts ScaleOptions
}

type ScaleOptions struct {
	Width, Height int
	Interpolator  interpolator.Interpolator
	Channel       drawgl.Channel
	Mask          drawgl.Mask
	Linear        bool
}

type jsonScaleOptions struct {
	ScaleOptions
	InterpolatorType string `json:"Interpolator"`
}

func NewScaleLinker(opts ScaleOptions) (graph.Linker, error) {
	if opts.Width <= 0 && opts.Height <= 0 {
		return nil, fmt.Errorf("invalid width %f and height %f, at least one has to be positive", opts.Width, opts.Height)
	}

	if opts.Interpolator == nil {
		opts.Interpolator = interpolator.BiLinear
	}

	opts.Channel.Normalize(true)
	return base.NewLinkerNode(Scale{
		Node: base.NewNode(),
		opts: opts,
	}), nil
}

func (n Scale) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
	var err error
	var buf *drawgl.FloatImage
	res := drawgl.Result{Id: n.Id()}

	defer func() {
		res.Buffer = buf
		if err != nil {
			res.Error = fmt.Errorf("applying scale using %v: %v", n.opts, err)
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

	m := matrix.New3()
	m[0][0] = float64(b.Dx()) / float64(tW)
	m[1][1] = float64(b.Dy()) / float64(tH)

	buf = affine(transformOperation{matrix: m, interpolator: n.opts.Interpolator, dstB: dr}, src, n.opts.Mask, n.opts.Channel, n.opts.Linear)
}

func init() {
	graph.RegisterLinker("Scale", func(opts json.RawMessage) (graph.Linker, error) {
		var o jsonScaleOptions

		if err := json.Unmarshal([]byte(opts), &o); err != nil {
			return nil, fmt.Errorf("constructing Scale: %v", err)
		}

		o.ScaleOptions.Interpolator = interpolator.Inst(o.InterpolatorType)

		return NewScaleLinker(o.ScaleOptions)
	})
}
