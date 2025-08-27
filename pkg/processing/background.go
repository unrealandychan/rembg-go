package processing

import (
    "image"
)

// RemoveBackground applies a model session to an image and returns an RGBA image with alpha applied.
// This is a stub: real implementation must call ONNX runtime.
func RemoveBackground(img image.Image) (image.Image, error) {
    // For the scaffold, just convert to RGBA and return the same image.
    rgba := image.NewRGBA(img.Bounds())
    drawImage(rgba, img)
    return rgba, nil
}

// drawImage is a tiny helper to copy src into dst.
func drawImage(dst *image.RGBA, src image.Image) {
    b := src.Bounds()
    for y := b.Min.Y; y < b.Max.Y; y++ {
        for x := b.Min.X; x < b.Max.X; x++ {
            dst.Set(x, y, src.At(x, y))
        }
    }
}
