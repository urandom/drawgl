package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"

	"github.com/urandom/drawgl"
	_ "github.com/urandom/drawgl/operation"
	"github.com/urandom/graph"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	jsonfile   = flag.String("json", "-", "read graph definition from json file [defaults to standard input]")
)

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

	var jsonReader io.Reader
	if *jsonfile == "" {
		exitWithError(errors.New("No input json file given"))
	} else if *jsonfile == "-" {
		jsonReader = os.Stdin
	} else {
		if f, err := os.Open(*jsonfile); err == nil {
			defer f.Close()

			jsonReader = f
		} else {
			exitWithError(err)
		}
	}

	args := flag.Args()

	var err error
	var roots []graph.Linker

	if len(args) > 0 {
		roots, err = graph.ProcessJSON(jsonReader, &graph.JSONTemplateData{Args: args})
	} else {
		roots, err = graph.ProcessJSON(jsonReader, nil)
	}

	if err != nil {
		exitWithError(err)
	}

	if len(roots) == 0 {
		exitWithError(errors.New("Input file contains no node roots"))
	}

	graph := drawgl.Graph{}
	err = graph.Process(roots[0])

	if err == nil {
		fmt.Println("JSON processing done")
	} else {
		fmt.Fprintf(os.Stderr, "Error processing json: %v\n", err)
	}
}

func exitWithError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}
