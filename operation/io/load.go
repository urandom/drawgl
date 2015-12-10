package io

import (
	"errors"
	"fmt"
	"image"
	"io"
	"os"

	"github.com/urandom/drawgl"
	"github.com/urandom/graph"
	"github.com/urandom/graph/base"
	"golang.org/x/image/draw"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const (
	InputPath   = "input-path"
	InputFormat = "input-format"
)

type Load struct {
	base.Node
	opts LoadOptions
}

type LoadOptions struct {
	Reader io.Reader
	Path   string
}

func NewLoadLinker(opts LoadOptions) (graph.Linker, error) {
	if opts.Reader == nil && opts.Path == "" {
		return nil, errors.New("No input")
	}
	return base.NewLinkerNode(Load{Node: base.NewNode(), opts: opts}), nil
}

func (n Load) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
	var err error
	res := drawgl.Result{Id: n.Id()}

	defer func() {
		if err != nil {
			res.Error = fmt.Errorf("Error loading image using %v: %v", n.opts, err)
		}
		output <- res

		wd.Close()
	}()

	reader := n.opts.Reader
	if reader == nil {
		reader, err = os.Open(n.opts.Path)

		if err != nil {
			return
		}

	}

	res.Meta = drawgl.Meta{InputPath: n.opts.Path}

	var img image.Image
	img, res.Meta[InputFormat], err = image.Decode(reader)

	if err == nil {
		if d, ok := img.(*drawgl.FloatImage); ok {
			res.Buffer = d
		} else {
			b := img.Bounds()
			res.Buffer = drawgl.NewFloatImage(b)
			draw.Draw(res.Buffer, b, img, b.Min, draw.Src)
		}
	}
}
