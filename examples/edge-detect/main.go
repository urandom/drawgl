package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/convolution"
	"github.com/urandom/drawgl/operation/io"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "No input file given")
		return
	}

	in := args[0]
	out := "/tmp/out.png"
	if len(args) > 1 {
		out = args[1]
	}

	kernel, err := convolution.NewKernel([]float32{-1, -1, -1, -1, 8, -1, -1, -1, -1})
	exitWithError(err)
	edge, err := convolution.NewConvolutionLinker(convolution.ConvolutionOptions{
		Kernel:    kernel,
		Normalize: true,
	})
	exitWithError(err)

	load, err := io.NewLoadLinker(io.LoadOptions{Path: in})
	exitWithError(err)
	save, err := io.NewSaveLinker(io.SaveOptions{Path: out})
	exitWithError(err)

	load.Link(edge)
	edge.Link(save)

	graph := drawgl.Graph{}
	err = graph.Process(load)

	if err == nil {
		fmt.Printf("Converted image saved to '%s'\n", out)
	} else {
		fmt.Fprintf(os.Stderr, "Error convertig image %s: %v\n", in, err)
	}
}

func exitWithError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}
