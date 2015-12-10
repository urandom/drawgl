package drawgl

import (
	"image"
	"image/color"
	"image/draw"
	"runtime"
	"sync"
)

type Channel int

type RectangleIterator interface {
	Iterate(mask Mask, fn func(pt image.Point, factor float64))
}

type ParallelRectangleIterator image.Rectangle
type LinearRectangleIterator image.Rectangle

type Mask struct {
	Image image.Image
	Rect  image.Rectangle

	hasImage bool
	hasRect  bool
}

const (
	All Channel = iota
	Red         = 1 << iota
	Green
	Blue
	Alpha

	m = 1<<16 - 1
)

func (c *Channel) Normalize() {
	if *c == All {
		*c = Red | Green | Blue | Alpha
	}
}

func (c Channel) Is(o Channel) bool {
	return c&o == o
}

func (rect ParallelRectangleIterator) Iterate(mask Mask, fn func(pt image.Point, factor float64)) {
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
						pt := image.Pt(x, y)
						f := MaskFactor(pt, mask)
						fn(pt, f)
					}
				}
			}
		}()
	}

	wg.Wait()
}

func (rect LinearRectangleIterator) Iterate(mask Mask, fn func(pt image.Point, factor float64)) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			pt := image.Pt(x, y)
			f := MaskFactor(pt, mask)
			fn(pt, f)
		}
	}
}

func NewMask(image image.Image, rect image.Rectangle) Mask {
	return Mask{Image: image, Rect: rect, hasImage: image != nil, hasRect: !rect.Empty()}

}

func CopyImage(img *image.NRGBA64) *image.NRGBA64 {
	cp := new(image.NRGBA64)
	*cp = *img
	cp.Pix = make([]uint8, len(img.Pix))
	copy(cp.Pix, img.Pix)

	return cp
}

func MaskFactor(pt image.Point, mask Mask) (factor float64) {
	if mask.hasRect && !pt.In(mask.Rect) {
		return 0
	}

	if mask.hasImage {
		_, _, _, ma := mask.Image.At(pt.X, pt.Y).RGBA()

		return float64(ma) / float64(m)
	}

	return 1
}

func MaskColor(dst color.NRGBA64, src color.NRGBA64, c Channel, f float64, op draw.Op) color.NRGBA64 {
	switch op {
	case draw.Over:
		switch f {
		case 0:
		case 1:
			if c.Is(Red) {
				dst.R = src.R
			}
			if c.Is(Green) {
				dst.G = src.G
			}
			if c.Is(Blue) {
				dst.B = src.B
			}
			if c.Is(Alpha) {
				dst.A = src.A
			}
		default:
			if c.Is(Red) {
				dst.R = uint16(float64(dst.R)/f + float64(src.R)*f)
			}
			if c.Is(Green) {
				dst.G = uint16(float64(dst.G)/f + float64(src.G)*f)
			}
			if c.Is(Blue) {
				dst.B = uint16(float64(dst.B)/f + float64(src.B)*f)
			}
			if c.Is(Alpha) {
				dst.A = uint16(float64(dst.A)/f + float64(src.A)*f)
			}
		}
	case draw.Src:
		switch f {
		case 0:
			if c.Is(Red) {
				dst.R = 0
			}
			if c.Is(Green) {
				dst.G = 0
			}
			if c.Is(Blue) {
				dst.B = 0
			}
			if c.Is(Alpha) {
				dst.A = 0
			}
		case 1:
			if c.Is(Red) {
				dst.R = src.R
			}
			if c.Is(Green) {
				dst.G = src.G
			}
			if c.Is(Blue) {
				dst.B = src.B
			}
			if c.Is(Alpha) {
				dst.A = src.A
			}
		default:
			if c.Is(Red) {
				dst.R = uint16(float64(src.R) * f)
			}
			if c.Is(Green) {
				dst.G = uint16(float64(src.G) * f)
			}
			if c.Is(Blue) {
				dst.B = uint16(float64(src.B) * f)
			}
			if c.Is(Alpha) {
				dst.A = uint16(float64(src.A) * f)
			}
		}
	}

	return dst
}
