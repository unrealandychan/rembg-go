package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"rembg-go/pkg/processing"
)

func main() {
	// This example reads `example.png` or `example.jpg`, runs the (stub) local
	// RemoveBackground function and writes `out.png`.
	f, err := os.Open("example.png")
	if err != nil {
		// fallback to jpeg
		f, err = os.Open("example.jpg")
		if err != nil {
			fmt.Println("open example image failed:", err)
			return
		}
	}
	defer f.Close()

	// decode generically
	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Println("decode failed:", err)
		return
	}

	out, err := processing.RemoveBackground(img)
	if err != nil {
		fmt.Println("remove failed:", err)
		return
	}
	of, _ := os.Create("out.png")
	defer of.Close()
	png.Encode(of, out)
	fmt.Println("wrote out.png")

	// For CLI usage, run: bin/rembg image examples/simple/example.png out.png --backend sagemaker --addr my-endpoint
}
