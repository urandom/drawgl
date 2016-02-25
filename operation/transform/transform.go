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

type TransformOp int

type Transform struct {
	base.Node

	opts TransformOptions
}

type TransformOptions struct {
	Operator TransformOp
	Channel  drawgl.Channel
	Mask     drawgl.Mask
	Linear   bool
}

const (
	_                          = iota
	transformFlipH TransformOp = iota
	transformFlipV
	// FlipH + Rotate270
	transformTranspose
	// FlipV + Rotate270
	transformTransverse
	transformRotate90
	transformRotate180
	transformRotate270
)

func NewTransformLinker(opts TransformOptions) (graph.Linker, error) {
	if opts.Operator == 0 || opts.Operator > transformRotate270 {
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

func (o TransformOp) MarshalJSON() (b []byte, err error) {
	switch o {
	case transformFlipH:
		b = []byte(`"flip-horizontal"`)
	case transformFlipV:
		b = []byte(`"flip-vertical"`)
	case transformTranspose:
		b = []byte(`"transpose"`)
	case transformTransverse:
		b = []byte(`"transverse"`)
	case transformRotate90:
		b = []byte(`"rotate-90"`)
	case transformRotate180:
		b = []byte(`"rotate-180"`)
	case transformRotate270:
		b = []byte(`"rotate-270"`)
	}
	return
}

func (o *TransformOp) UnmarshalJSON(b []byte) (err error) {
	var val string
	if err = json.Unmarshal(b, &val); err == nil {
		switch val {
		case "flip-horizontal":
			*o = transformFlipH
		case "flip-vertical":
			*o = transformFlipV
		case "transpose":
			*o = transformTranspose
		case "transverse":
			*o = transformTransverse
		case "rotate-90":
			*o = transformRotate90
		case "rotate-180":
			*o = transformRotate180
		case "rotate-270":
			*o = transformRotate270
		default:
			err = errors.New("unknown transform operator " + val)
		}
	}
	return
}

func transform(op TransformOp, src *drawgl.FloatImage, mask drawgl.Mask, channel drawgl.Channel, forceLinear bool) (dst *drawgl.FloatImage) {
	srcB := src.Bounds()
	dstB := srcB

	switch op {
	case transformTranspose, transformTransverse, transformRotate90, transformRotate270:
		dstB = image.Rect(srcB.Min.Y, srcB.Min.X, srcB.Max.Y, srcB.Max.X)
	}

	var offsetX, offsetY int

	switch op {
	case transformFlipH:
		offsetX = srcB.Min.X + srcB.Max.X - 1
	case transformFlipV:
		offsetY = srcB.Min.Y + srcB.Max.Y - 1
	case transformTranspose:
		offsetX = dstB.Min.X - srcB.Min.Y
		offsetY = dstB.Min.Y - srcB.Min.X
	case transformTransverse:
		offsetX = dstB.Min.Y + srcB.Max.Y - 1
		offsetY = dstB.Min.X + srcB.Max.X - 1
	case transformRotate90:
		offsetX = dstB.Min.X - srcB.Min.Y
		offsetY = dstB.Min.Y + srcB.Max.X - 1
	case transformRotate180:
		offsetX = dstB.Min.X + srcB.Max.X - 1
		offsetY = dstB.Min.Y + srcB.Max.Y - 1
	case transformRotate270:
		offsetX = dstB.Min.X + srcB.Max.Y - 1
		offsetY = dstB.Min.Y - srcB.Min.X
	}

	dst = drawgl.NewFloatImage(srcB)

	it := drawgl.DefaultRectangleIterator(srcB, forceLinear)

	it.Iterate(mask, func(pt image.Point, f float32) {
		if f == 0 {
			return
		}

		var px, py int

		switch op {
		case transformFlipH:
			px, py = offsetX-pt.X, pt.Y
		case transformFlipV:
			px, py = pt.X, offsetY-pt.Y
		case transformTranspose:
			px, py = offsetX+pt.Y, offsetY+pt.X
		case transformTransverse:
			px, py = offsetX-pt.Y, offsetY-pt.X
		case transformRotate90:
			px, py = offsetX+pt.Y, offsetY-pt.X
		case transformRotate180:
			px, py = offsetX-pt.X, offsetY-pt.Y
		case transformRotate270:
			px, py = offsetX-pt.Y, offsetY+pt.X
		}

		srcColor := src.UnsafeFloatAt(pt.X, pt.Y)
		dstColor := src.UnsafeFloatAt(px, py)

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
