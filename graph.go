package drawgl

import (
	"fmt"
	"image"

	"github.com/urandom/graph"
	"golang.org/x/image/draw"
)

type Graph struct {
}

type Result struct {
	Id     graph.Id
	Buffer draw.Image
	Meta   Meta
	Error  error
}

type Processor interface {
	Process(wd graph.WalkData, buffers map[graph.ConnectorName]Result, output chan<- Result)
}

type Meta map[string]interface{}

func (g Graph) Process(start graph.Linker) error {
	walker := graph.NewWalker(start)
	data := walker.Walk()

	output := make(chan Result)
	resultSet := make(map[graph.Id]Result)

	for {
		select {
		case wd, open := <-data:
			if open {
				if p, ok := wd.Node.(Processor); ok {
					pb := make(map[graph.ConnectorName]Result)

					for _, p := range wd.Parents {
						r := resultSet[p.Node.Id()]
						if p.From != graph.OutputName && r.Buffer != nil {
							// If the image buffer comes from a secondary output, clone it
							r.Buffer = copyImage(r.Buffer)
						}
						r.Meta = copyMeta(r.Meta)
						pb[p.To] = r
					}

					go p.Process(wd, pb, output)
				} else {
					wd.Close()
				}
			} else {
				return nil
			}
		case r := <-output:
			if r.Error != nil {
				return fmt.Errorf("Error processing node %v: %v\n", r.Id, r.Error)
			}
			resultSet[r.Id] = r
		}
	}
}

func copyImage(img draw.Image) draw.Image {
	switch i := img.(type) {
	case *image.Alpha:
		var cp *image.Alpha
		*cp = *i
		copy(cp.Pix, i.Pix)

		return cp
	case *image.Alpha16:
		var cp *image.Alpha16
		*cp = *i
		copy(cp.Pix, i.Pix)

		return cp
	case *image.CMYK:
		var cp *image.CMYK
		*cp = *i
		copy(cp.Pix, i.Pix)

		return cp
	case *image.Gray:
		var cp *image.Gray
		*cp = *i
		copy(cp.Pix, i.Pix)

		return cp
	case *image.Gray16:
		var cp *image.Gray16
		*cp = *i
		copy(cp.Pix, i.Pix)

		return cp
	case *image.NRGBA:
		var cp *image.NRGBA
		*cp = *i
		copy(cp.Pix, i.Pix)

		return cp
	case *image.NRGBA64:
		var cp *image.NRGBA64
		*cp = *i
		copy(cp.Pix, i.Pix)

		return cp
	case *image.Paletted:
		var cp *image.Paletted
		*cp = *i
		copy(cp.Pix, i.Pix)

		return cp
	case *image.RGBA:
		var cp *image.RGBA
		*cp = *i
		copy(cp.Pix, i.Pix)

		return cp
	case *image.RGBA64:
		var cp *image.RGBA64
		*cp = *i
		copy(cp.Pix, i.Pix)

		return cp
	}

	return img
}

func copyMeta(meta Meta) (cp Meta) {
	cp = make(Meta)

	if meta != nil {
		for k, v := range meta {
			cp[k] = v
		}
	}

	return
}
