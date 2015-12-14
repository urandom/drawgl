package drawgl

import (
	"fmt"

	"github.com/urandom/graph"
)

type Graph struct {
}

type Result struct {
	Id           graph.Id
	Buffer       *FloatImage
	NamedBuffers map[graph.ConnectorName]*FloatImage
	Meta         Meta
	Error        error
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
						if p.From != graph.OutputName {
							// If the image buffer comes from a secondary output, clone it
							if nb, ok := r.NamedBuffers[p.From]; ok && nb != nil {
								r.Buffer = CopyImage(nb)
							} else if r.Buffer != nil {
								r.Buffer = CopyImage(r.Buffer)
							}
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

func copyMeta(meta Meta) (cp Meta) {
	cp = make(Meta)

	if meta != nil {
		for k, v := range meta {
			cp[k] = v
		}
	}

	return
}
