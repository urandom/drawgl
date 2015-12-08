package drawgl

import (
	"fmt"
	"image"

	"github.com/urandom/graph"
)

type Graph struct {
}

type Channel int

const (
	All Channel = iota << 1
	Red
	Green
	Blue
	Alpha
)

type Result struct {
	Id     graph.Id
	Buffer *image.NRGBA64
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
							r.Buffer = CopyImage(r.Buffer)
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

func CopyImage(img *image.NRGBA64) *image.NRGBA64 {
	cp := new(image.NRGBA64)
	*cp = *img
	cp.Pix = make([]uint8, len(img.Pix))
	copy(cp.Pix, img.Pix)

	return cp
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

func ClampUint16(in float64) uint16 {
	if in > 0xffff {
		return 0xffff
	} else if in < 0 {
		return 0
	}

	return uint16(in)
}
