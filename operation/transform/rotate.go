package transform

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/transform/matrix"
	"github.com/urandom/graph"
	"github.com/urandom/graph/base"
)

type Rotate struct {
	base.Node

	opts RotateOptions
}

type RotateOptions struct {
	Degrees       float64
	Center        [2]int
	CenterPercent [2]float64
	Interpolator  string
	Channel       drawgl.Channel
	Mask          drawgl.Mask
	Linear        bool
}

type jsonRotateOptions struct {
	RotateOptions
	Center [2]string
}

func NewRotateLinker(opts RotateOptions) (graph.Linker, error) {
	opts.Channel.Normalize(true)
	return base.NewLinkerNode(Rotate{
		Node: base.NewNode(),
		opts: opts,
	}), nil
}

func (n Rotate) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
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

	if n.opts.Degrees == 0 {
		buf = src
		return
	}

	rads := n.opts.Degrees * (2 * math.Pi / 360)

	cos := math.Cos(rads)
	sin := math.Sin(rads)
	m := matrix.New3()

	m[0][0] = cos
	m[0][1] = -sin
	m[1][0] = sin
	m[1][1] = cos

	b := src.Bounds()
	var h, k float64
	if n.opts.Center[0] != 0 {
		h = float64(n.opts.Center[0])
	} else if n.opts.CenterPercent[0] != 0 {
		h = n.opts.CenterPercent[0] * float64(b.Dx())
	}

	if n.opts.Center[1] != 0 {
		k = float64(n.opts.Center[1])
	} else if n.opts.CenterPercent[1] != 0 {
		k = n.opts.CenterPercent[1] * float64(b.Dy())
	}

	buf = src

	if h != 0 || k != 0 {
		m[0][2] = h - m[0][0]*h - m[0][1]*k
		m[1][2] = k - m[1][0]*h - m[1][1]*k
	}

	buf = affine(transformOperation{matrix: m, interpolator: n.opts.Interpolator}, buf, n.opts.Mask, n.opts.Channel, n.opts.Linear)
}

func init() {
	graph.RegisterLinker("Rotate", func(opts json.RawMessage) (graph.Linker, error) {
		var o jsonRotateOptions
		var err error

		if err = json.Unmarshal([]byte(opts), &o); err != nil {
			return nil, fmt.Errorf("constructing Rotate: %v", err)
		}

		if len(o.Center) == 2 {
			for i := 0; i < 2; i++ {
				if o.RotateOptions.Center[i], o.RotateOptions.CenterPercent[i], err =
					drawgl.ParseLength(o.Center[i]); err != nil {
					return nil, fmt.Errorf("constructing Rotate: parsing Center[%d]: %v", i, err)
				}
			}
		}

		return NewRotateLinker(o.RotateOptions)
	})
}
