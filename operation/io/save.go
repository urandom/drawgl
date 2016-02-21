package io

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"strings"

	"github.com/urandom/drawgl"
	"github.com/urandom/graph"
	"github.com/urandom/graph/base"

	"image/gif"
	"image/jpeg"
	"image/png"
)

const (
	OutputPath   = "output-path"
	OutputFormat = "output-format"
)

type Save struct {
	base.Node
	opts SaveOptions
}

type SaveOptions struct {
	Writer      io.Writer
	Path        string
	Type        string
	JpegOptions *jpeg.Options
	GifOptions  *gif.Options
}

func NewSaveLinker(opts SaveOptions) (graph.Linker, error) {
	if opts.Writer == nil && opts.Path == "" {
		return nil, errors.New("No output")
	}
	return base.NewLinkerNode(Save{Node: base.NewNode(), opts: opts}), nil
}

func (n Save) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
	var err error
	res := drawgl.Result{Id: n.Id()}

	defer func() {
		if err != nil {
			res.Error = fmt.Errorf("Error saving image using %v: %v", n.opts, err)
		}
		output <- res

		wd.Close()
	}()

	r := buffers[graph.InputName]
	res.Meta = r.Meta
	if r.Buffer == nil {
		err = fmt.Errorf("no input buffer")
		return
	}

	kind := "jpeg"
	if n.opts.Type != "" {
		kind = n.opts.Type
	} else if n.opts.Path != "" {
		idx := strings.LastIndexByte(n.opts.Path, '.')
		if idx != -1 {
			mime := mime.TypeByExtension(n.opts.Path[idx:])
			kind = mime[strings.LastIndexByte(mime, '/')+1:]
		}
	}

	w := n.opts.Writer
	if w == nil && n.opts.Path != "" {
		w, err = os.Create(n.opts.Path)
		defer w.(*os.File).Close()
		if err != nil {
			return
		}
	}

	if w == nil {
		err = fmt.Errorf("nowhere to save the image")
	} else {
		res.Meta[OutputFormat] = kind
		res.Meta[OutputPath] = n.opts.Path
		switch kind {
		case "jpeg":
			jpeg.Encode(w, r.Buffer, n.opts.JpegOptions)
		case "png":
			png.Encode(w, r.Buffer)
		case "gif":
			gif.Encode(w, r.Buffer, n.opts.GifOptions)
		default:
			err = fmt.Errorf("unknown format %s", kind)
		}
	}
}

func init() {
	graph.RegisterLinker("Save", func(opts json.RawMessage) (graph.Linker, error) {
		var o SaveOptions

		if err := json.Unmarshal([]byte(opts), &o); err != nil {
			return nil, fmt.Errorf("constructing Save: %v", err)
		}

		return NewSaveLinker(o), nil
	})
}
