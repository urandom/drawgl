package io_test

import (
	"testing"

	"github.com/urandom/drawgl"
	"github.com/urandom/drawgl/operation/io"
	"github.com/urandom/drawgl/operation/tests"
	"github.com/urandom/graph"
)

func TestLoadJpg(t *testing.T) {
	jpg, err := io.NewLoadLinker(io.LoadOptions{Path: tests.TestDataDir() + "/test.jpg"})
	if err != nil {
		t.Fatalf("Error opening jpeg: %v\n", err)
	}

	pb := make(map[graph.ConnectorName]drawgl.Result)
	p, wd, output := tests.PrepareLinker(jpg)

	go p.Process(wd, pb, output)

	r := <-output

	if r.Error != nil {
		t.Fatalf("Error processing: %v\n", r.Error)
	}
	buf := r.Buffer
	b := buf.Bounds()

	if b.Dx() != 4 || b.Dy() != 4 {
		t.Fatalf("Wrong bounds: %v\n", b)
	}

	colors := tests.Colors()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := buf.FloatAt(x, y)

			// A bit more liberal with the jpeg
			if !testColorJpeg(c, colors[y][x]) {
				t.Fatalf("At %d:%d, color %v doesn't match %v\n", x, y, c, colors[y][x])
			}
		}
	}
}

func TestLoadPng(t *testing.T) {
	buf, err := tests.ReadTestData()
	if err != nil {
		t.Fatalf("Error opening png: %v\n", err)
	}

	b := buf.Bounds()

	if b.Dx() != 4 || b.Dy() != 4 {
		t.Fatalf("Wrong bounds: %v\n", b)
	}

	colors := tests.Colors()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := buf.FloatAt(x, y)

			if c != colors[y][x] {
				t.Fatalf("At %d:%d, color %v doesn't match %v\n", x, y, c, colors[y][x])
			}
		}
	}
}

func TestLoadGif(t *testing.T) {
	jpg, err := io.NewLoadLinker(io.LoadOptions{Path: tests.TestDataDir() + "/test.gif"})
	if err != nil {
		t.Fatalf("Error opening gif: %v\n", err)
	}

	pb := make(map[graph.ConnectorName]drawgl.Result)
	p, wd, output := tests.PrepareLinker(jpg)

	go p.Process(wd, pb, output)

	r := <-output

	if r.Error != nil {
		t.Fatalf("Error processing: %v\n", r.Error)
	}
	buf := r.Buffer
	b := buf.Bounds()

	if b.Dx() != 4 || b.Dy() != 4 {
		t.Fatalf("Wrong bounds: %v\n", b)
	}

	colors := tests.Colors()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := buf.FloatAt(x, y)

			if c != colors[y][x] {
				t.Fatalf("At %d:%d, color %v doesn't match %v\n", x, y, c, colors[y][x])
			}
		}
	}
}

func testColorJpeg(c, exp drawgl.FloatColor) bool {
	if c.R < exp.R-0.995 || c.R > exp.R+0.995 {
		return false
	}

	if c.G < exp.G-0.995 || c.G > exp.G+0.995 {
		return false
	}

	if c.B < exp.B-0.995 || c.B > exp.B+0.995 {
		return false
	}

	if c.A < exp.A-0.995 || c.A > exp.A+0.995 {
		return false
	}

	return true
}
