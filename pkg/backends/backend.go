package backends

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"

	"github.com/unrealandychan/rembg-go/pkg/processing"
)

// Backend is a generic inference backend used by the processing pipeline.
type Backend interface {
	// Infer sends payload (image bytes) and returns raw response bytes (e.g., PNG mask bytes).
	Infer(ctx context.Context, payload []byte) ([]byte, error)
}

// SageMakerBackend wraps a SageMaker endpoint.
type SageMakerBackend struct {
	Endpoint string
}

func NewSageMakerBackend(endpoint string) *SageMakerBackend {
	return &SageMakerBackend{Endpoint: endpoint}
}

func (s *SageMakerBackend) Infer(ctx context.Context, payload []byte) ([]byte, error) {
	return InvokeSageMaker(ctx, s.Endpoint, payload)
}

// TritonHTTPBackend wraps the Triton HTTP v2 infer endpoint.
type TritonHTTPBackend struct {
	Addr      string
	Model     string
	InputName string
	Shape     []int
	DType     string
}

func NewTritonHTTPBackend(addr, model, inputName string, shape []int, dtype string) *TritonHTTPBackend {
	return &TritonHTTPBackend{Addr: addr, Model: model, InputName: inputName, Shape: shape, DType: dtype}
}

func (t *TritonHTTPBackend) Infer(ctx context.Context, payload []byte) ([]byte, error) {
	return InferTritonHTTP(ctx, t.Addr, t.Model, t.InputName, payload, t.Shape, t.DType)
}

// TritonGRPCBackend wraps a Triton gRPC connection. It uses the triton_grpc helper stub.
type TritonGRPCBackend struct {
	Addr      string
	Model     string
	InputName string
	Shape     []int
	DType     string
}

func NewTritonGRPCBackend(addr, model, inputName string, shape []int, dtype string) *TritonGRPCBackend {
	return &TritonGRPCBackend{Addr: addr, Model: model, InputName: inputName, Shape: shape, DType: dtype}
}

func (t *TritonGRPCBackend) Infer(ctx context.Context, payload []byte) ([]byte, error) {
	conn, err := NewTritonGRPCConn(t.Addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return InferTritonGRPCStub(ctx, conn, t.Model, t.InputName, payload, t.Shape, t.DType)
}

// RemoveBackgroundWithBackend runs inference using the given backend and applies mask post-processing.
func RemoveBackgroundWithBackend(ctx context.Context, backend Backend, imgBytes []byte, opts processing.RemoveBackgroundOptions) (interface{}, error) {
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	maskBytes, err := backend.Infer(ctx, imgBytes)
	if err != nil {
		return nil, fmt.Errorf("backend infer: %w", err)
	}
	maskImg, err := png.Decode(bytes.NewReader(maskBytes))
	if err != nil {
		return nil, fmt.Errorf("decode mask: %w", err)
	}
	var cutout image.Image
	if opts.OnlyMask {
		cutout = maskImg
	} else if opts.AlphaMatting {
		am, err := processing.AlphaMattingCutoutGo(img, maskImg, opts.AlphaMattingForegroundThreshold, opts.AlphaMattingBackgroundThreshold, opts.AlphaMattingErodeSize)
		if err != nil {
			if opts.PutAlpha {
				cutout = processing.PutAlphaCutoutGo(img, maskImg)
			} else {
				cutout = processing.NaiveCutoutGo(img, maskImg)
			}
		} else {
			cutout = am
		}
	} else {
		if opts.PutAlpha {
			cutout = processing.PutAlphaCutoutGo(img, maskImg)
		} else {
			cutout = processing.NaiveCutoutGo(img, maskImg)
		}
	}
	if opts.BackgroundColor != nil && !opts.OnlyMask {
		cutout = processing.ApplyBackgroundColorGo(cutout, opts.BackgroundColor)
	}
	switch opts.ReturnType {
	case "image":
		return cutout, nil
	case "bytes":
		buf := new(bytes.Buffer)
		err := png.Encode(buf, cutout)
		if err != nil {
			return nil, fmt.Errorf("png encode error: %w", err)
		}
		return buf.Bytes(), nil
	default:
		return cutout, nil
	}
}
