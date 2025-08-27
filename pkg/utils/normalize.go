package utils

import (
    "image"
    "image/draw"

    "github.com/nfnt/resize"
)

// NormalizeImage is a compatibility wrapper that returns an NRGBA-converted/resized image.
// For model input tensors prefer using NormalizeToFloat32CHW which returns normalized float32 CHW data.
func NormalizeImage(img image.Image, width, height int) image.Image {
    // resize first
    resized := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
    // convert to NRGBA
    out := image.NewNRGBA(resized.Bounds())
    draw.Draw(out, out.Bounds(), resized, image.Point{}, draw.Src)
    return out
}

// NormalizeToFloat32CHW converts img to a float32 tensor in CHW order with given target size and per-channel mean/std.
// - width/height: target dimensions
// - mean/std: 3-element arrays for R,G,B
// The returned slice has length 3*width*height and layout [C][H][W] where channel order is R, G, B.
func NormalizeToFloat32CHW(img image.Image, width, height int, mean [3]float32, std [3]float32) ([]float32, error) {
    // resize
    resized := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

    // ensure NRGBA
    nrgba := image.NewNRGBA(resized.Bounds())
    draw.Draw(nrgba, nrgba.Bounds(), resized, image.Point{}, draw.Src)

    hw := width * height
    out := make([]float32, 3*hw)

    // iterate pixels and fill out in CHW order
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            idx := y*width + x
            r := nrgba.Pix[(y*nrgba.Stride)+(x*4)+0]
            g := nrgba.Pix[(y*nrgba.Stride)+(x*4)+1]
            b := nrgba.Pix[(y*nrgba.Stride)+(x*4)+2]

            // normalize to [0,1], then apply mean/std
            rf := (float32(r) / 255.0 - mean[0]) / std[0]
            gf := (float32(g) / 255.0 - mean[1]) / std[1]
            bf := (float32(b) / 255.0 - mean[2]) / std[2]

            out[0*hw+idx] = rf
            out[1*hw+idx] = gf
            out[2*hw+idx] = bf
        }
    }
    return out, nil
}
