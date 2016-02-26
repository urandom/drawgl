package transform

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

type Operator int

type Transform struct {
	base.Node

	opts TransformOptions
}

type TransformOptions struct {
	Operator Operator
	Channel  drawgl.Channel
	Mask     drawgl.Mask
	Linear   bool
}

const (
	_                      = iota
	FlipHOperator Operator = iota
	FlipVOperator
	// FlipH + Rotate270
	TransposeOperator
	// FlipV + Rotate270
	TransverseOperator
	Rotate90Operator
	Rotate180Operator
	Rotate270Operator
)

func NewTransformLinker(opts TransformOptions) (graph.Linker, error) {
	if opts.Operator == 0 || opts.Operator > Rotate270Operator {
		return nil, fmt.Errorf("unknown operator %d", opts.Operator)
	}

	opts.Channel.Normalize()

	return base.NewLinkerNode(Transform{
		Node: base.NewNode(),
		opts: opts,
	}), nil
}

func (n Transform) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
	var err error
	var buf *drawgl.FloatImage
	res := drawgl.Result{Id: n.Id()}

	defer func() {
		res.Buffer = buf
		if err != nil {
			res.Error = fmt.Errorf("applying transform using %v: %v", n.opts, err)
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

	buf = transform(n.opts.Operator, src, n.opts.Mask, n.opts.Channel, n.opts.Linear)
}

func (o Operator) MarshalJSON() (b []byte, err error) {
	switch o {
	case FlipHOperator:
		b = []byte(`"flip-horizontal"`)
	case FlipVOperator:
		b = []byte(`"flip-vertical"`)
	case TransposeOperator:
		b = []byte(`"transpose"`)
	case TransverseOperator:
		b = []byte(`"transverse"`)
	case Rotate90Operator:
		b = []byte(`"rotate-90"`)
	case Rotate180Operator:
		b = []byte(`"rotate-180"`)
	case Rotate270Operator:
		b = []byte(`"rotate-270"`)
	}
	return
}

func (o *Operator) UnmarshalJSON(b []byte) (err error) {
	var val string
	if err = json.Unmarshal(b, &val); err == nil {
		switch val {
		case "flip-horizontal":
			*o = FlipHOperator
		case "flip-vertical":
			*o = FlipVOperator
		case "transpose":
			*o = TransposeOperator
		case "transverse":
			*o = TransverseOperator
		case "rotate-90":
			*o = Rotate90Operator
		case "rotate-180":
			*o = Rotate180Operator
		case "rotate-270":
			*o = Rotate270Operator
		default:
			err = errors.New("unknown transform operator " + val)
		}
	}
	return
}

func transform(op Operator, src *drawgl.FloatImage, mask drawgl.Mask, channel drawgl.Channel, forceLinear bool) (dst *drawgl.FloatImage) {
	srcB := src.Bounds()
	dstB := srcB

	switch op {
	case TransposeOperator, TransverseOperator, Rotate90Operator, Rotate270Operator:
		dstB = image.Rect(srcB.Min.Y, srcB.Min.X, srcB.Max.Y, srcB.Max.X)
	}

	var offsetX, offsetY int

	switch op {
	case FlipHOperator:
		offsetX = srcB.Min.X + srcB.Max.X - 1
	case FlipVOperator:
		offsetY = srcB.Min.Y + srcB.Max.Y - 1
	case TransposeOperator:
		offsetX = dstB.Min.X - srcB.Min.Y
		offsetY = dstB.Min.Y - srcB.Min.X
	case TransverseOperator:
		offsetX = dstB.Min.Y + srcB.Max.Y - 1
		offsetY = dstB.Min.X + srcB.Max.X - 1
	case Rotate90Operator:
		offsetX = dstB.Min.X + srcB.Max.Y - 1
		offsetY = dstB.Min.Y - srcB.Min.X
	case Rotate180Operator:
		offsetX = dstB.Min.X + srcB.Max.X - 1
		offsetY = dstB.Min.Y + srcB.Max.Y - 1
	case Rotate270Operator:
		offsetX = dstB.Min.X - srcB.Min.Y
		offsetY = dstB.Min.Y + srcB.Max.X - 1
	}

	dst = drawgl.NewFloatImage(dstB)

	it := drawgl.DefaultRectangleIterator(srcB, forceLinear)

	it.Iterate(mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		var px, py int

		switch op {
		case FlipHOperator:
			px, py = offsetX-pt.X, pt.Y
		case FlipVOperator:
			px, py = pt.X, offsetY-pt.Y
		case TransposeOperator:
			px, py = offsetX+pt.Y, offsetY+pt.X
		case TransverseOperator:
			px, py = offsetX-pt.Y, offsetY-pt.X
		case Rotate90Operator:
			px, py = offsetX-pt.Y, offsetY+pt.X
		case Rotate180Operator:
			px, py = offsetX-pt.X, offsetY-pt.Y
		case Rotate270Operator:
			px, py = offsetX+pt.Y, offsetY-pt.X
		}

		srcColor := src.UnsafeFloatAt(pt.X, pt.Y)

		var dstColor drawgl.FloatColor
		if srcB == dstB || (image.Point{px, py}.In(src.Rect)) {
			dstColor = src.UnsafeFloatAt(px, py)
		} else {
			dstColor = drawgl.FloatColor{A: 1}
		}

		dst.UnsafeSetColor(px, py, drawgl.MaskColor(dstColor, srcColor, channel, f, draw.Over))
	})

	return
}

func init() {
	graph.RegisterLinker("Transform", func(opts json.RawMessage) (graph.Linker, error) {
		var o TransformOptions

		if err := json.Unmarshal([]byte(opts), &o); err != nil {
			return nil, fmt.Errorf("constructing Transform: %v", err)
		}

		return NewTransformLinker(o)
	})
}
