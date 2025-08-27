package processing

import (
    "image/color"
    "image/draw"
    "testing"
)

func TestToNRGBA(t *testing.T) {
    rect := image.Rect(0, 0, 10, 10)
    img := image.NewRGBA(rect)
    draw.Draw(img, rect, &image.Uniform{C: color.RGBA{R: 10, G: 20, B: 30, A: 255}}, image.Point{}, draw.Src)

    out := ToNRGBA(img)
    if out.Bounds() != rect {
        t.Fatalf("expected bounds %v got %v", rect, out.Bounds())
    }
    c := out.NRGBAAt(0, 0)
    if c.R != 10 || c.G != 20 || c.B != 30 {
        t.Fatalf("unexpected pixel: %#v", c)
    }
}
