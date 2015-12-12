package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
)

func main() {
	f, err := os.Open("test.png")
	if err != nil {
		panic(err)
	}
	t, _, _ := image.Decode(f)

	f, err = os.Create("/tmp/test1.jpg")
	if err != nil {
		panic(err)
	}
	jpeg.Encode(f, t, nil)

	f, _ = os.Create("/tmp/test1.gif")
	gif.Encode(f, t, nil)

	fmt.Printf("%T ", t)
	fmt.Println(t)
	for y := t.Bounds().Min.Y; y < t.Bounds().Max.Y; y++ {
		for x := t.Bounds().Min.X; x < t.Bounds().Max.X; x++ {
			fmt.Print(t.At(x, y).RGBA())
			fmt.Print(", ")
		}
	}
	fmt.Println(" ")

	f, _ = os.Open("test.jpg")
	t, _, _ = image.Decode(f)

	f, _ = os.Create("/tmp/test2.png")
	png.Encode(f, t)

	f, _ = os.Create("/tmp/test2.gif")
	gif.Encode(f, t, nil)

	fmt.Printf("%T ", t)
	fmt.Println(t)
	for y := t.Bounds().Min.Y; y < t.Bounds().Max.Y; y++ {
		for x := t.Bounds().Min.X; x < t.Bounds().Max.X; x++ {
			fmt.Print(t.At(x, y).RGBA())
			fmt.Print(", ")
		}
	}
	fmt.Println(" ")

	f, _ = os.Open("test.gif")
	t, _, _ = image.Decode(f)

	f, _ = os.Create("/tmp/test3.png")
	png.Encode(f, t)

	f, _ = os.Create("/tmp/test3.jpg")
	jpeg.Encode(f, t, nil)

	fmt.Printf("%T ", t)
	fmt.Println(t)
	for y := t.Bounds().Min.Y; y < t.Bounds().Max.Y; y++ {
		for x := t.Bounds().Min.X; x < t.Bounds().Max.X; x++ {
			fmt.Print(t.At(x, y).RGBA())
			fmt.Print(", ")
		}
	}
	fmt.Println(" ")

}
