package transform

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/interpolator"
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
	Interpolator  interpolator.Interpolator
	Channel       drawgl.Channel
	Mask          drawgl.Mask
	Linear        bool
}

type jsonRotateOptions struct {
	RotateOptions
	InterpolatorType string `json:"Interpolator"`
	Center           [2]string
}

func NewRotateLinker(opts RotateOptions) (graph.Linker, error) {
	if opts.Interpolator == nil {
		opts.Interpolator = interpolator.BiLinear
	}

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
	m[1][1] = cos
	m[0][1] = sin
	m[1][0] = -sin

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
	dstB := b
	if h != 0 || k != 0 {
		move := matrix.New3()
		move[0][2] = h
		move[1][2] = k
		dstB.Min.X -= int(h)
		dstB.Min.Y -= int(k)
		buf = affine(transformOperation{matrix: move, interpolator: n.opts.Interpolator, dstB: dstB}, buf, n.opts.Mask, n.opts.Channel, n.opts.Linear)
	}

	buf = affine(transformOperation{matrix: m, interpolator: n.opts.Interpolator, dstB: dstB}, buf, n.opts.Mask, n.opts.Channel, n.opts.Linear)

	if h != 0 || k != 0 {
		move := matrix.New3()
		move[0][2] = -h
		move[1][2] = -k
		dstB.Min.X += int(h)
		dstB.Min.Y += int(k)
		buf = affine(transformOperation{matrix: move, interpolator: n.opts.Interpolator, dstB: dstB}, buf, n.opts.Mask, n.opts.Channel, n.opts.Linear)
	}
}

func init() {
	graph.RegisterLinker("Rotate", func(opts json.RawMessage) (graph.Linker, error) {
		var o jsonRotateOptions

		if err := json.Unmarshal([]byte(opts), &o); err != nil {
			return nil, fmt.Errorf("constructing Rotate: %v", err)
		}

		o.RotateOptions.Interpolator = interpolator.Inst(o.InterpolatorType)

		if len(o.Center) == 2 {
			for i := 0; i < 2; i++ {
				if o.Center[i] != "" {
					if strings.HasSuffix(o.Center[i], "%") {
						o.Center[i] = strings.TrimSpace(strings.TrimSuffix(o.Center[i], "%"))
						if pc, err := strconv.ParseFloat(o.Center[i], 64); err == nil {
							o.RotateOptions.CenterPercent[i] = pc / 100
						}
					} else {
						o.Center[i] = strings.TrimSpace(strings.TrimSuffix(o.Center[i], "px"))
						if px, err := strconv.Atoi(o.Center[i]); err == nil {
							o.RotateOptions.Center[i] = px
						}
					}
				}
			}
		}

		return NewRotateLinker(o.RotateOptions)
	})
}
