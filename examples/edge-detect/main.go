package main

import (
	"fmt"
	"os"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/convolution"
	"github.com/urandom/drawgl/operation/io"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "No input file given")
		return
	}

	out := "/tmp/out.png"
	if len(os.Args) > 2 {
		out = os.Args[2]
	}

	kernel, err := convolution.NewKernel([]float64{-1, -1, -1, -1, 8, -1, -1, -1, -1})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	edge, err := convolution.NewConvolutionLinker(convolution.ConvolutionOptions{
		Kernel:    kernel,
		Normalize: true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	load := io.NewLoadLinker(io.LoadOptions{Path: os.Args[1]})
	save := io.NewSaveLinker(io.SaveOptions{Path: out})

	load.Link(edge)
	edge.Link(save)

	graph := drawgl.Graph{}
	err = graph.Process(load)

	if err == nil {
		fmt.Printf("Converted image saved to '%s'\n", out)
	} else {
		fmt.Fprintf(os.Stderr, "Error convertig image %s: %v\n", os.Args[1], err)
	}
}
