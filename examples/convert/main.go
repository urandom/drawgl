package main

import (
	"fmt"
	"os"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/common"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "No input file given")
		return
	}

	load := common.NewLoadLinker(common.LoadOptions{Path: os.Args[1]})
	save := common.NewSaveLinker(common.SaveOptions{Path: "/tmp/out.png"})

	load.Link(save)

	graph := drawgl.Graph{}
	err := graph.Process(load)

	if err == nil {
		fmt.Println("Converted image saved to '/tmp/out.png'")
	} else {
		fmt.Fprintf(os.Stderr, "Error convertig image %s: %v\n", os.Args[1], err)
	}
}
