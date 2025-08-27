package utils

import (
    "image"
    "image/color"
    "image/draw"
    "testing"
)

func TestNormalizeToFloat32CHW(t *testing.T) {
    // create a simple 2x2 red image
    img := image.NewRGBA(image.Rect(0, 0, 2, 2))
    draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 255, G: 0, B: 0, A: 255}}, image.Point{}, draw.Src)

    mean := [3]float32{0.0, 0.0, 0.0}
    std := [3]float32{1.0, 1.0, 1.0}
    out, err := NormalizeToFloat32CHW(img, 2, 2, mean, std)
    if err != nil {
        t.Fatal(err)
    }
    if len(out) != 3*2*2 {
        t.Fatalf("unexpected length: %d", len(out))
    }
    // Check that red channel is 1.0 (approx) and others 0.0
    if out[0] < 0.99 || out[0] > 1.01 {
        t.Fatalf("unexpected red value: %v", out[0])
    }
    if out[1] != 0.0 || out[2] != 0.0 {
        t.Fatalf("unexpected green/blue values: %v %v", out[1], out[2])
    }
}
