package utils

import (
    "image"
    "github.com/nfnt/resize"
)

// ResizeImage resizes to the specified width and height using Lanczos resampling.
func ResizeImage(img image.Image, width, height uint) image.Image {
    return resize.Resize(width, height, img, resize.Lanczos3)
}
