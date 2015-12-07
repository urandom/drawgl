package main

import (
	"fmt"
	"os"

	"github.com/urandom/drawgl"
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

	load := io.NewLoadLinker(io.LoadOptions{Path: os.Args[1]})
	save := io.NewSaveLinker(io.SaveOptions{Path: out})
	exif := io.NewCopyExifLinker(io.CopyExifOptions{})

	load.Link(save)
	save.Link(exif)

	graph := drawgl.Graph{}
	err := graph.Process(load)

	if err == nil {
		fmt.Printf("Converted image saved to '%s'\n", out)
	} else {
		fmt.Fprintf(os.Stderr, "Error convertig image %s: %v\n", os.Args[1], err)
	}
}
