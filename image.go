package drawgl

import (
	"image"
	"runtime"
	"sync"
)

type Channel int

type RectangleIterator interface {
	Iterate(fn func(pt image.Point))
}

type ParallelRectangleIterator image.Rectangle
type LinearRectangleIterator image.Rectangle

const (
	All Channel = iota
	Red         = 1 << iota
	Green
	Blue
	Alpha
)

func (c *Channel) Normalize() {
	if *c == All {
		*c = Red | Green | Blue | Alpha
	}
}

func (rect ParallelRectangleIterator) Iterate(fn func(pt image.Point)) {
	var wg sync.WaitGroup

	rowchan := make(chan []int)

	go func() {
		defer close(rowchan)

		capacity := 200
		i := 0
		chunk := make([]int, i, capacity)
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			chunk = append(chunk, y)

			if i == cap(chunk) || y == rect.Max.Y-1 {
				rowchan <- chunk

				if y != rect.Max.Y-1 {
					i = 0
					chunk = make([]int, 0, capacity)
				}
			} else {
				i++
			}
		}
	}()

	count := runtime.GOMAXPROCS(0)
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			for chunk := range rowchan {
				for _, y := range chunk {
					for x := rect.Min.X; x < rect.Max.X; x++ {
						fn(image.Pt(x, y))
					}
				}
			}
		}()
	}

	wg.Wait()
}

func (rect LinearRectangleIterator) Iterate(fn func(pt image.Point)) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			fn(image.Pt(x, y))
		}
	}
}

func CopyImage(img *image.NRGBA64) *image.NRGBA64 {
	cp := new(image.NRGBA64)
	*cp = *img
	cp.Pix = make([]uint8, len(img.Pix))
	copy(cp.Pix, img.Pix)

	return cp
}
