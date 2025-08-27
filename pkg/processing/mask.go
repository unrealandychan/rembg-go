package processing

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// ApplyMask applies an alpha mask (grayscale) to src and returns an RGBA image.
// maskImg should be the same bounds as src and be a grayscale or alpha image.
func ApplyMask(src image.Image, maskImg image.Image) *image.RGBA {
	b := src.Bounds()
	out := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, _ := src.At(x, y).RGBA()
			// mask pixel: use red channel as alpha if grayscale
			mr, _, _, _ := maskImg.At(x, y).RGBA()
			alpha := uint8(mr >> 8)
			out.Set(x, y, color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(bb >> 8), A: alpha})
		}
	}
	return out
}

// DecodePNG decodes PNG bytes into an image.Image
func DecodePNG(b []byte) (image.Image, error) {
	return png.Decode(bytes.NewReader(b))
}
