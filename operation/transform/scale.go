package transform

import (
	"encoding/json"
	"fmt"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/transform/matrix"
	"github.com/urandom/graph"
	"github.com/urandom/graph/base"
)

type Scale struct {
	base.Node

	opts ScaleOptions
}

type ScaleOptions struct {
	Width, Height               int
	WidthPercent, HeightPercent float64
	Interpolator                string
	Channel                     drawgl.Channel
	Mask                        drawgl.Mask
	Linear                      bool
}

type jsonScaleOptions struct {
	ScaleOptions
	Width, Height string
}

func NewScaleLinker(opts ScaleOptions) (graph.Linker, error) {
	if opts.Width <= 0 && opts.Height <= 0 && opts.WidthPercent <= 0 && opts.HeightPercent <= 0 {
		return nil, fmt.Errorf("invalid width %f and height %f, at least one has to be positive", opts.Width, opts.Height)
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

	width, height := b.Dx(), b.Dy()

	var tW, tH int
	if n.opts.Width != 0 {
		tW = n.opts.Width
	} else if n.opts.WidthPercent != 0 {
		tW = int(n.opts.WidthPercent * float64(width))
	}

	if n.opts.Height != 0 {
		tH = n.opts.Height
	} else if n.opts.HeightPercent != 0 {
		tH = int(n.opts.HeightPercent * float64(height))
	}

	if tW <= 0 || tH <= 0 {
		if tW <= 0 {
			tW = tH * width / height
		} else {
			tH = tW * height / width
		}
	}

	m := matrix.New3()
	m[0][0] = float64(tW) / float64(b.Dx())
	m[1][1] = float64(tH) / float64(b.Dy())

	buf = affine(transformOperation{matrix: m, interpolator: n.opts.Interpolator}, src, n.opts.Mask, n.opts.Channel, n.opts.Linear)
}

func init() {
	graph.RegisterLinker("Scale", func(opts json.RawMessage) (graph.Linker, error) {
		var o jsonScaleOptions
		var err error

		if err = json.Unmarshal([]byte(opts), &o); err != nil {
			return nil, fmt.Errorf("constructing Scale: %v", err)
		}

		if o.ScaleOptions.Width, o.ScaleOptions.WidthPercent, err =
			drawgl.ParseLength(o.Width); err != nil {
			return nil, fmt.Errorf("constructing Scale: parsing Width: %v", err)
		}

		if o.ScaleOptions.Height, o.ScaleOptions.HeightPercent, err =
			drawgl.ParseLength(o.Height); err != nil {
			return nil, fmt.Errorf("constructing Scale: parsing Height: %v", err)
		}

		return NewScaleLinker(o.ScaleOptions)
	})
}
