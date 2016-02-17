package io

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/urandom/drawgl"
	"github.com/urandom/graph"
	"github.com/urandom/graph/base"
)

type ExecType int

const (
	Exiftool ExecType = iota
)

type CopyExif struct {
	base.Node
	opts CopyExifOptions
}

type CopyExifOptions struct {
	Executable     string
	ExecutableType ExecType
	InputPath      string
	OutputPath     string
}

func NewCopyExifLinker(opts CopyExifOptions) graph.Linker {
	return base.NewLinkerNode(CopyExif{Node: base.NewNode(), opts: opts})
}

func (n CopyExif) Process(wd graph.WalkData, buffers map[graph.ConnectorName]drawgl.Result, output chan<- drawgl.Result) {
	var err error
	res := drawgl.Result{Id: n.Id()}

	defer func() {
		if err != nil {
			res.Error = fmt.Errorf("Error copying exif data using %v: %v", n.opts, err)
		}
		output <- res

		wd.Close()
	}()

	r := buffers[graph.InputName]
	res.Meta = r.Meta

	inputPath := n.opts.InputPath
	if inputPath == "" {
		if s, ok := res.Meta[InputPath].(string); ok {
			inputPath = s
		}
	}

	if inputPath == "" {
		err = fmt.Errorf("no input path")
		return
	}

	if s, ok := res.Meta[InputFormat].(string); !ok || s != "jpeg" {
		fmt.Fprintln(os.Stderr, "CopyExif: input format is not supported")
		// err = fmt.Errorf("input format is not supported")
		return
	}

	outputPath := n.opts.OutputPath
	if outputPath == "" {
		if s, ok := res.Meta[OutputPath].(string); ok {
			outputPath = s
		}
	}

	if outputPath == "" {
		err = fmt.Errorf("no output path")
	}

	if s, ok := res.Meta[OutputFormat].(string); !ok || s != "jpeg" {
		fmt.Fprintln(os.Stderr, "CopyExif: output format is not supported")
		// err = fmt.Errorf("output format is not supported")
		return
	}

	var prog string
	if n.opts.Executable == "" {
		prog, err = exec.LookPath("exiftool")
	} else {
		prog, err = exec.LookPath(n.opts.Executable)
	}
	if err != nil {
		return
	}

	var cmd *exec.Cmd
	switch n.opts.ExecutableType {
	case Exiftool:
		cmd = exec.Command(prog, "-overwrite_original", "-TagsFromFile", inputPath, outputPath)
	}

	if cmd == nil {
		err = fmt.Errorf("no executable")
	} else {
		if err = cmd.Start(); err != nil {
			return
		}

		if err = cmd.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					err = fmt.Errorf("error code: %d", status)
				}
			}
		}
	}
}

func init() {
	drawgl.RegisterOperation("CopyExif", func(opts json.RawMessage) (graph.Linker, error) {
		var o CopyExifOptions

		if err := json.Unmarshal([]byte(opts), &o); err != nil {
			return nil, fmt.Errorf("constructing CopyExif: %v", err)
		}

		return NewCopyExifLinker(o)
	})
}
