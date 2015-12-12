package tests

import (
	"fmt"
	"os"
	"strings"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/io"
	"github.com/urandom/graph"
)

func Pix() []drawgl.ColorValue {
	// 4x4
	return []drawgl.ColorValue{
		1, 1, 1, 1, 1, 0, 0, 1, 1, 1, 0, 1, 0.39215687, 0.78431374, 0, 1,
		0, 0, 0, 1, 0, 1, 0, 1, 0, 1, 1, 1, 0.39215687, 0, 0.78431374, 1,
		1, 1, 1, 1, 0, 0, 1, 1, 1, 0, 1, 1, 0, 0.39215687, 0.78431374, 1,
		0, 0, 0, 1, 0, 0.78431374, 0.39215687, 1, 0.78431374, 0, 0.39215687, 1, 0.78431374, 0.39215687, 0, 1,
	}
}

func Colors() (c [4][4]drawgl.FloatColor) {
	pix := Pix()

	x, y := 0, 0
	for i := 0; i < len(pix); i += 4 {
		color := drawgl.FloatColor{pix[i], pix[i+1], pix[i+2], pix[i+3]}
		c[y][x] = color

		x++
		if x%4 == 0 {
			x = 0
			y++
		}
	}

	return
}

func PrepareLinker(l graph.Linker, connectors ...graph.Connector) (drawgl.Processor, graph.WalkData, chan drawgl.Result) {
	node := l.Node()

	done := make(chan struct{})
	wd := graph.NewWalkData(node, connectors, done)

	output := make(chan drawgl.Result)
	return node.(drawgl.Processor), wd, output
}

func TestDataDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ""
	}

	parent := "/drawgl"
	idx := strings.Index(cwd, parent)

	return cwd[:idx] + parent + "/operation/tests/.test_data"
}

func ReadTestData() (*drawgl.FloatImage, error) {
	jpg, err := io.NewLoadLinker(io.LoadOptions{Path: TestDataDir() + "/test.png"})
	if err != nil {
		return nil, err
	}

	pb := make(map[graph.ConnectorName]drawgl.Result)
	p, wd, output := PrepareLinker(jpg)

	go p.Process(wd, pb, output)

	r := <-output

	if r.Error != nil {
		return nil, r.Error
	}

	return r.Buffer, nil
}
