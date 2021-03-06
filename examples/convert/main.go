package main

import (
	"fmt"
	"image/jpeg"
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

	load, err := io.NewLoadLinker(io.LoadOptions{Path: os.Args[1]})
	exitWithError(err)
	save, err := io.NewSaveLinker(io.SaveOptions{Path: out, JpegOptions: &jpeg.Options{Quality: 100}})
	exitWithError(err)
	exif := io.NewCopyExifLinker(io.CopyExifOptions{})

	load.Link(save)
	save.Link(exif)

	graph := drawgl.Graph{}
	err = graph.Process(load)

	if err == nil {
		fmt.Printf("Converted image saved to '%s'\n", out)
	} else {
		fmt.Fprintf(os.Stderr, "Error convertig image %s: %v\n", os.Args[1], err)
	}
}

func exitWithError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}
