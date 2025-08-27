package processing

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"

	"github.com/owulveryck/onnx-go"
	"github.com/owulveryck/onnx-go/backend/x/gorgonnx"
	"gorgonia.org/tensor"
)

// RemoveBackgroundOptions holds options for background removal.
type RemoveBackgroundOptions struct {
	PostProcessMask                 bool
	OnlyMask                        bool
	AlphaMatting                    bool
	PutAlpha                        bool
	AlphaMattingForegroundThreshold float32
	AlphaMattingBackgroundThreshold float32
	AlphaMattingErodeSize           int
	BackgroundColor                 *color.Color
	ReturnType                      string // "image", "bytes"
}

// RemoveBackground applies U2Net ONNX model to an image and returns RGBA with alpha mask.
func RemoveBackground(img image.Image, opts RemoveBackgroundOptions) (interface{}, error) {
	// 1. Load ONNX model
	modelPath := "u2net.onnx" // Update with actual path
	f, err := os.Open(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open model: %w", err)
	}
	defer f.Close()

	backend := gorgonnx.NewGraph()
	model := onnx.NewModel(backend)
	modelBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read model file: %w", err)
	}
	err = model.UnmarshalBinary(modelBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to load model: %w", err)
	}

	// 2. Preprocess image: resize to 320x320, normalize
	inputTensor, err := preprocessImage(img)
	if err != nil {
		return nil, fmt.Errorf("preprocess error: %w", err)
	}

	// 3. Run inference
	err = model.SetInput(0, inputTensor)
	if err != nil {
		return nil, fmt.Errorf("set input error: %w", err)
	}
	err = backend.Run()
	if err != nil {
		return nil, fmt.Errorf("inference error: %w", err)
	}
	output, err := model.GetOutputTensors()
	if err != nil {
		return nil, fmt.Errorf("get output error: %w", err)
	}

	// 4. Postprocess mask: normalize, resize to original size
	maskImg, err := postprocessMask(output[0], img.Bounds().Dx(), img.Bounds().Dy())
	if err != nil {
		return nil, fmt.Errorf("mask postprocess error: %w", err)
	}

	// 5. Post-process mask if requested
	if opts.PostProcessMask {
		maskImg = postProcessMaskGo(maskImg) // You need to implement postProcessMaskGo
	}

	var cutout image.Image
	if opts.OnlyMask {
		cutout = maskImg
	} else if opts.AlphaMatting {
		am, err := alphaMattingCutoutGo(img, maskImg, opts.AlphaMattingForegroundThreshold, opts.AlphaMattingBackgroundThreshold, opts.AlphaMattingErodeSize)
		if err != nil {
			if opts.PutAlpha {
				cutout = putAlphaCutoutGo(img, maskImg)
			} else {
				cutout = naiveCutoutGo(img, maskImg)
			}
		} else {
			cutout = am
		}
	} else {
		if opts.PutAlpha {
			cutout = putAlphaCutoutGo(img, maskImg)
		} else {
			cutout = naiveCutoutGo(img, maskImg)
		}
	}

	// Apply background color if requested and not only mask
	if opts.BackgroundColor != nil && !opts.OnlyMask {
		cutout = applyBackgroundColorGo(cutout, opts.BackgroundColor)
	}

	if opts.ReturnType == "image" {
		return cutout, nil
	}

	if opts.ReturnType == "bytes" {
		buf := new(bytes.Buffer)
		err := png.Encode(buf, cutout)
		if err != nil {
			return nil, fmt.Errorf("png encode error: %w", err)
		}
		return buf.Bytes(), nil
	}

	return cutout, nil
}

// preprocessImage resizes and normalizes the image for U2Net input.
func preprocessImage(img image.Image) (*tensor.Dense, error) {
	// Resize to 320x320
	width, height := 320, 320
	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	origBounds := img.Bounds()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Nearest neighbor
			srcX := origBounds.Min.X + x*(origBounds.Dx())/width
			srcY := origBounds.Min.Y + y*(origBounds.Dy())/height
			resized.Set(x, y, img.At(srcX, srcY))
		}
	}

	// Normalize to [0,1] and convert to float32 tensor [1,3,320,320]
	data := make([]float32, 3*width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pix := color.NRGBAModel.Convert(resized.At(x, y)).(color.NRGBA)
			idx := y*width + x
			data[idx] = float32(pix.R) / 255.0                // Red
			data[width*height+idx] = float32(pix.G) / 255.0   // Green
			data[2*width*height+idx] = float32(pix.B) / 255.0 // Blue
		}
	}

	shape := []int{1, 3, width, height}
	t := tensor.New(tensor.WithShape(shape...), tensor.WithBacking(data))
	return t, nil
}

// postprocessMask normalizes and resizes the mask to original image size.
func postprocessMask(t tensor.Tensor, width, height int) (image.Image, error) {
	// Assume t is [1,1,320,320] float32
	data := t.Data().([]float32)
	if len(data) == 0 {
		return nil, fmt.Errorf("empty mask tensor")
	}
	// Normalize to [0,255]
	minVal, maxVal := data[0], data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	img320 := image.NewGray(image.Rect(0, 0, 320, 320))
	for y := 0; y < 320; y++ {
		for x := 0; x < 320; x++ {
			idx := y*320 + x
			val := (data[idx] - minVal) / (maxVal - minVal)
			img320.SetGray(x, y, color.Gray{Y: uint8(val * 255)})
		}
	}
	// Resize to original size
	imgResized := image.NewGray(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Nearest neighbor
			srcX := x * 320 / width
			srcY := y * 320 / height
			imgResized.SetGray(x, y, img320.GrayAt(srcX, srcY))
		}
	}
	return imgResized, nil
}

// postProcessMaskGo can be a no-op or simple smoothing (identity for now)
func postProcessMaskGo(mask image.Image) image.Image {
	// No-op, just return mask
	return mask
}

// applyMaskToImage sets the mask as alpha channel on the original image.
func applyMaskToImage(img image.Image, mask image.Image) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			orig := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			alpha := color.GrayModel.Convert(mask.At(x, y)).(color.Gray).Y
			rgba.Set(x, y, color.NRGBA{R: orig.R, G: orig.G, B: orig.B, A: alpha})
		}
	}
	return rgba
}

// Helper stubs for cutout and background color
func alphaMattingCutoutGo(img image.Image, mask image.Image, fgThresh, bgThresh float32, erodeSize int) (image.Image, error) {
	// TODO: Implement alpha matting cutout
	return nil, fmt.Errorf("alphaMattingCutoutGo not implemented")
}

func putAlphaCutoutGo(img image.Image, mask image.Image) image.Image {
	return applyMaskToImage(img, mask)
}

func naiveCutoutGo(img image.Image, mask image.Image) image.Image {
	return applyMaskToImage(img, mask)
}

func applyBackgroundColorGo(img image.Image, bgColor *color.Color) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	var bg color.NRGBA
	if bgColor != nil {
		// Convert *color.Color to color.NRGBA
		bg = color.NRGBAModel.Convert(*bgColor).(color.NRGBA)
	} else {
		bg = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // default white
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pix := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			alpha := float64(pix.A) / 255.0
			outR := uint8(float64(pix.R)*alpha + float64(bg.R)*(1-alpha))
			outG := uint8(float64(pix.G)*alpha + float64(bg.G)*(1-alpha))
			outB := uint8(float64(pix.B)*alpha + float64(bg.B)*(1-alpha))
			out.Set(x, y, color.NRGBA{R: outR, G: outG, B: outB, A: 255})
		}
	}
	return out
}
