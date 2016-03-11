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

type Translate struct {
	base.Node

	opts TranslateOptions
}

type TranslateOptions struct {
	Offset        [2]int
	OffsetPercent [2]float64
	Interpolator  interpolator.Interpolator
	Channel       drawgl.Channel
	Mask          drawgl.Mask
	Linear        bool
}

type jsonTranslateOptions struct {
	TranslateOptions
	InterpolatorType string `json:"Interpolator"`
	Offset           [2]string
}

func NewTranslateLinker(opts TranslateOptions) (graph.Linker, error) {
	if opts.Interpolator == nil {
		opts.Interpolator = interpolator.BiLinear
	}

	opts.Channel.Normalize(true)
	return base.NewLinkerNode(Translate{
		Node: base.NewNode(),
		opts: opts,
	}), nil
}

func (n Translate) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
	var err error
	var buf *drawgl.FloatImage
	res := drawgl.Result{Id: n.Id()}

	defer func() {
		res.Buffer = buf
		if err != nil {
			res.Error = fmt.Errorf("applying translate using %v: %v", n.opts, err)
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

	if n.opts.Offset[0] == 0 && n.opts.Offset[1] == 0 &&
		n.opts.OffsetPercent[0] == 0 && n.opts.OffsetPercent[1] == 0 {
		buf = src
		return
	}

	m := matrix.New3()
	b := src.Bounds()
	if n.opts.Offset[0] != 0 {
		m[0][2] = -float64(n.opts.Offset[0])
	} else if n.opts.OffsetPercent[0] != 0 {
		m[0][2] = -n.opts.OffsetPercent[0] * float64(b.Dx())
	}

	if n.opts.Offset[1] != 0 {
		m[1][2] = -float64(n.opts.Offset[1])
	} else if n.opts.OffsetPercent[1] != 0 {
		m[1][2] = -n.opts.OffsetPercent[1] * float64(b.Dy())
	}

	buf = affine(transformOperation{matrix: m, interpolator: n.opts.Interpolator}, src, n.opts.Mask, n.opts.Channel, n.opts.Linear)
}

func init() {
	graph.RegisterLinker("Translate", func(opts json.RawMessage) (graph.Linker, error) {
		var o jsonTranslateOptions
		var err error

		if err = json.Unmarshal([]byte(opts), &o); err != nil {
			return nil, fmt.Errorf("constructing Translate: %v", err)
		}

		o.TranslateOptions.Interpolator = interpolator.Inst(o.InterpolatorType)

		if len(o.Offset) == 2 {
			for i := 0; i < 2; i++ {
				if o.TranslateOptions.Offset[i], o.TranslateOptions.OffsetPercent[i], err =
					drawgl.ParseLength(o.Offset[i]); err != nil {
					return nil, fmt.Errorf("constructing Translate: parsing Offset[%d]: %v", i, err)
				}
			}
		}

		return NewTranslateLinker(o.TranslateOptions)
	})
}
