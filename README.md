# rembg-go — Minimal Golang rembg

This repository is a minimal scaffold of a Go port of rembg focused on video-to-image extraction and background removal. It provides a starting point with a CLI, package layout, and small examples.

See `copilot-instructions.md` for more implementation guidance.

## Quickstart

Prerequisites (macOS):

```bash
brew install go pkg-config opencv protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Clone and prepare:

```bash
cd /Users/228448
git clone <your-repo-url> rembg-golang
cd rembg-golang
go mod tidy
```

Build the CLI:

```bash
# build normal binary
go build -o bin/rembg ./cmd/rembg

# build with Triton gRPC support (after generating protos)
go build -tags triton -o bin/rembg-triton ./cmd/rembg
```

## CLI usage

Commands provided by the scaffold:

- `rembg image <input.png> <output.png> [--backend <backend>] [--addr <addr>]`
- `rembg video <input.mp4>` — extracts frames with OpenCV (gocv) into `./frames`

Examples:

```bash
# Process a PNG with a SageMaker endpoint named `my-endpoint`
bin/rembg image examples/simple/example.png out.png --backend sagemaker --addr my-endpoint

# Extract frames from a video (requires OpenCV and gocv)
bin/rembg video input.mp4

# Use Triton gRPC backend (requires proto generation and triton build)
bin/rembg-triton image examples/simple/example.png out.png --backend triton_grpc --addr triton-host:8001
```

Notes:
- The scaffold currently decodes PNG images for the `image` command. Add JPEG support in `cmd/rembg` if needed.
- `--addr` is the SageMaker endpoint name for `sagemaker` or `host:port` for Triton backends.

## Backends: SageMaker & Triton

SageMaker

- Uses `pkg/backends/sagemaker.go` to call `InvokeEndpoint` via the AWS SDK v2. Ensure AWS creds/region are available.

Triton (gRPC)

- The project includes helpers for Triton gRPC in `pkg/backends/triton_grpc.go` and a build-tagged template implementation in `pkg/backends/triton_grpc_impl.go`.

- To enable Triton gRPC support:


	- Generate Go protos (script provided):
		```bash
		./scripts/gen_triton_protos.sh
		```


	- Build with the `triton` tag:

		```bash
		go build -tags triton -o bin/rembg-triton ./cmd/rembg
		```

	- Implement request construction in `pkg/backends/triton_grpc_impl.go` to match your model input (shape, datatype, input name). See `copilot-instructions.md` for guidance.

Triton details:

- Triton expects tensor data in RawContents as little-endian binary (e.g., FP32). Use the helpers in `pkg/backends/triton_helpers.go` to convert float32 slices to bytes and back.
- Typical ONNX image model input: shape `[1,3,H,W]` (NCHW) and datatype `FP32`.

## Example: how Triton gRPC flow works (high level)

1. Preprocess image -> float32 CHW tensor (`pkg/utils/normalize.go`).
2. Convert []float32 -> []byte with `Float32sToBytes` (`pkg/backends/triton_helpers.go`).
3. Build `ModelInferRequest` with inputs (Name, Datatype, Shape) and set RawContents.
4. Call `client.ModelInfer(ctx, req)` on the generated Triton gRPC client.
5. Parse `RawContents` from outputs -> []float32 -> reshape to mask -> encode PNG mask -> composite.

## Tests

Run tests (requires Go):

```bash
go test ./...
```

Included tests:

- `pkg/backends/triton_helpers_test.go` — float32/byte roundtrip and shape helper tests.
- `pkg/utils/normalize_test.go` — image normalization tensor test.

## Troubleshooting & tips

- If you see issues building, run `go mod tidy` to refresh dependencies.
- If Triton calls fail, verify model names and input tensor metadata with:

```bash
curl http://triton-host:8000/v2/models/<model>/config
```

brew install go pkg-config opencv protoc

## Contributing / Next steps

- Implement ONNX runtime local inference (`pkg/models/session.go`) to run models locally.
- Fill `pkg/backends/triton_grpc_impl.go` after generating protos for a complete Triton gRPC example.
- Add JPEG support in the CLI and enhance video pipeline for streaming inference.

## Preprocessing & normalization

This project includes helpers to preprocess images for ONNX/Triton models.

- `utils.NormalizeImage(img image.Image, width, height int) image.Image` — resizes the image and returns an NRGBA image (convenience for image-based workflows).
- `utils.NormalizeToFloat32CHW(img image.Image, width, height int, mean [3]float32, std [3]float32) ([]float32, error)` — converts the image to a normalized float32 tensor in CHW order (R,G,B channels).

Common example (prepare tensor and bytes for Triton FP32 input):

```go
// load img (image.Image)
tensor, err := utils.NormalizeToFloat32CHW(img, 320, 320, [3]float32{0.485,0.456,0.406}, [3]float32{0.229,0.224,0.225})
if err != nil { /* handle */ }
payload := backends.Float32sToBytes(tensor)
// payload is ready to put into Triton request RawContents
```

Notes on normalization:

- The example uses ImageNet mean/std; many segmentation models (like U2Net) expect raw uint8 input or simple scaling by 1/255.0 — verify the model's preprocessing.
- `rembg video <input.mp4>` — extracts frames with OpenCV (gocv) into `./frames`


