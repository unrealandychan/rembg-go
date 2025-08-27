package processing

import (
    "image"
    "image/draw"
)

// ToNRGBA converts any image to NRGBA for processing.
func ToNRGBA(img image.Image) *image.NRGBA {
    b := img.Bounds()
    out := image.NewNRGBA(b)
    draw.Draw(out, b, img, b.Min, draw.Src)
    return out
}
