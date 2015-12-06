package common

import (
	"fmt"
	"io"
	"mime"
	"os"
	"strings"

	"github.com/urandom/drawgl"
	"github.com/urandom/graph"
	"github.com/urandom/graph/base"
	"golang.org/x/image/draw"

	"image/gif"
	"image/jpeg"
	"image/png"
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

func NewSaveLinker(opts SaveOptions) graph.Linker {
	return base.NewLinkerNode(Save{Node: base.NewNode(), opts: opts})
}

func (n Save) Process(wd graph.WalkData, buffers map[graph.ConnectorName]draw.Image, output chan<- drawgl.Result) {
	var err error
	res := drawgl.Result{Id: n.Id()}

	defer func() {
		if err != nil {
			res.Error = fmt.Errorf("Error saving image using %v: %v", n.opts, err)
		}
		output <- res

		wd.Close()
	}()

	buf := buffers[graph.InputName]
	if buf == nil {
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
		if err != nil {
			return
		}
	}

	if w == nil {
		err = fmt.Errorf("nowhere to save the image")
	} else {
		switch kind {
		case "jpeg":
			jpeg.Encode(w, buf, n.opts.JpegOptions)
		case "png":
			png.Encode(w, buf)
		case "gif":
			gif.Encode(w, buf, n.opts.GifOptions)
		default:
			err = fmt.Errorf("unknown format %s", kind)
		}
	}
}
