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
	// Iterate iterates over the image buffer, calling the fn function for each
	// point. The cycle order is row -> column. Implementations must ensure
	// that all columns of a given row are received in a single goroutine
	Iterate(mask Mask, fn func(pt image.Point, factor float32))
	// VerticalIterate iterates over the image buffer, calling the fn function
	// for each point. The cycle order is column -> row. Implementations must
	// ensure that all rows of a given column are received in a single
	// goroutine
	VerticalIterate(mask Mask, fn func(pt image.Point, factor float32))
}

type ParallelRectangleIterator image.Rectangle
type LinearRectangleIterator image.Rectangle

type FloatImage struct {
	// Pix holds the image's pixels, in R, G, B, A order and big-endian format. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*4].
	Pix []ColorValue
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

type Mask struct {
	Image image.Image
	Rect  image.Rectangle

	hasImage bool
	hasRect  bool
}

const (
	RGB Channel = iota
	Red         = 1 << iota
	Green
	Blue
	Alpha

	m = 1<<16 - 1
)

func (c *Channel) Normalize() {
	if *c == RGB {
		*c = Red | Green | Blue
	}
}

func (c Channel) Is(o Channel) bool {
	return c&o == o
}

func DefaultRectangleIterator(rect image.Rectangle, forceLinear ...bool) RectangleIterator {
	if len(forceLinear) > 0 && forceLinear[0] || runtime.GOMAXPROCS(0) == 1 {
		return LinearRectangleIterator(rect)
	}

	return ParallelRectangleIterator(rect)
}

func (rect ParallelRectangleIterator) Iterate(mask Mask, fn func(pt image.Point, factor float32)) {
	count := runtime.GOMAXPROCS(0)
	if count == 1 {
		LinearRectangleIterator(rect).Iterate(mask, fn)
		return
	}

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

func (rect ParallelRectangleIterator) VerticalIterate(mask Mask, fn func(pt image.Point, factor float32)) {
	count := runtime.GOMAXPROCS(0)
	if count == 1 {
		LinearRectangleIterator(rect).VerticalIterate(mask, fn)
		return
	}

	var wg sync.WaitGroup

	rowchan := make(chan []int)

	go func() {
		defer close(rowchan)

		capacity := 200
		i := 0
		chunk := make([]int, i, capacity)
		for x := rect.Min.X; x < rect.Max.X; x++ {
			chunk = append(chunk, x)

			if i == cap(chunk) || x == rect.Max.X-1 {
				rowchan <- chunk

				if x != rect.Max.X-1 {
					i = 0
					chunk = make([]int, 0, capacity)
				}
			} else {
				i++
			}
		}
	}()

	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			for chunk := range rowchan {
				for _, x := range chunk {
					for y := rect.Min.Y; y < rect.Max.Y; y++ {
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

func (rect LinearRectangleIterator) Iterate(mask Mask, fn func(pt image.Point, factor float32)) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			pt := image.Pt(x, y)
			f := MaskFactor(pt, mask)
			fn(pt, f)
		}
	}
}

func (rect LinearRectangleIterator) VerticalIterate(mask Mask, fn func(pt image.Point, factor float32)) {
	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			pt := image.Pt(x, y)
			f := MaskFactor(pt, mask)
			fn(pt, f)
		}
	}
}

func (p *FloatImage) ColorModel() color.Model { return FloatColorModel }

func (p *FloatImage) Bounds() image.Rectangle { return p.Rect }

func (p *FloatImage) At(x, y int) color.Color {
	return p.FloatAt(x, y)
}

func (p *FloatImage) FloatAt(x, y int) FloatColor {
	if !(image.Point{x, y}.In(p.Rect)) {
		return FloatColor{}
	}
	return p.UnsafeFloatAt(x, y)
}

func (p *FloatImage) UnsafeFloatAt(x, y int) FloatColor {
	i := p.PixOffset(x, y)
	return FloatColor{
		p.Pix[i], p.Pix[i+1], p.Pix[i+2], p.Pix[i+3],
	}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *FloatImage) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*4
}

func (p *FloatImage) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	c1 := FloatColorModel.Convert(c).(FloatColor)
	p.Pix[i] = c1.R
	p.Pix[i+1] = c1.G
	p.Pix[i+2] = c1.B
	p.Pix[i+3] = c1.A
}

func (p *FloatImage) SetColor(x, y int, c FloatColor) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	p.UnsafeSetColor(x, y, c)
}

func (p *FloatImage) UnsafeSetColor(x, y int, c FloatColor) {
	i := p.PixOffset(x, y)
	p.Pix[i] = c.R
	p.Pix[i+1] = c.G
	p.Pix[i+2] = c.B
	p.Pix[i+3] = c.A
}

// SubImage returns an image representing the portion of the image p visible
// through r. The returned value shares pixels with the original image.
func (p *FloatImage) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(p.Rect)
	// If r1 and r2 are Rectangles, r1.Intersect(r2) is not guaranteed to be inside
	// either r1 or r2 if the intersection is empty. Without explicitly checking for
	// this, the Pix[i:] expression below can panic.
	if r.Empty() {
		return &FloatImage{}
	}
	i := p.PixOffset(r.Min.X, r.Min.Y)
	return &FloatImage{
		Pix:    p.Pix[i:],
		Stride: p.Stride,
		Rect:   r,
	}
}

// Opaque scans the entire image and reports whether it is fully opaque.
func (p *FloatImage) Opaque() bool {
	if p.Rect.Empty() {
		return true
	}
	i0, i1 := 3, p.Rect.Dx()*4
	for y := p.Rect.Min.Y; y < p.Rect.Max.Y; y++ {
		for i := i0; i < i1; i += 4 {
			if p.Pix[i] != 1 {
				return false
			}
		}
		i0 += p.Stride
		i1 += p.Stride
	}
	return true
}

// NewFloatImage returns a new FloatImage with the given bounds.
func NewFloatImage(r image.Rectangle) *FloatImage {
	w, h := r.Dx(), r.Dy()
	pix := make([]ColorValue, 4*w*h)
	return &FloatImage{pix, 4 * w, r}
}

func NewMask(image image.Image, rect image.Rectangle) Mask {
	return Mask{Image: image, Rect: rect, hasImage: image != nil, hasRect: !rect.Empty()}

}

func CopyImage(img *FloatImage) *FloatImage {
	cp := new(FloatImage)
	*cp = *img
	cp.Pix = make([]ColorValue, len(img.Pix))
	copy(cp.Pix, img.Pix)

	return cp
}

func MaskFactor(pt image.Point, mask Mask) (factor float32) {
	if mask.hasRect && !pt.In(mask.Rect) {
		return 0
	}

	if mask.hasImage {
		_, _, _, ma := mask.Image.At(pt.X, pt.Y).RGBA()

		return float32(ma) / float32(m)
	}

	return 1
}

func MaskColor(dst FloatColor, src FloatColor, c Channel, f float32, op draw.Op) FloatColor {
	fv := ColorValue(f)
	switch op {
	case draw.Over:
		switch fv {
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
				dst.R = dst.R/fv + src.R*fv
			}
			if c.Is(Green) {
				dst.G = dst.G/fv + src.G*fv
			}
			if c.Is(Blue) {
				dst.B = dst.B/fv + src.B*fv
			}
			if c.Is(Alpha) {
				dst.A = dst.A/fv + src.A*fv
			}
		}
	case draw.Src:
		switch fv {
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
				dst.R = src.R * fv
			}
			if c.Is(Green) {
				dst.G = src.G * fv
			}
			if c.Is(Blue) {
				dst.B = src.B * fv
			}
			if c.Is(Alpha) {
				dst.A = src.A * fv
			}
		}
	}

	return dst
}
