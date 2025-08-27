## Instructions: Minimal Golang rembg Project (Video-to-Image & Background Removal)

### 1. Install Required Go Packages

Add these dependencies to your `go.mod`:

- ONNX Runtime: `github.com/yalue/onnxruntime_go`
- CLI: `github.com/spf13/cobra`
- Image Resizing: `github.com/nfnt/resize`
- Standard library: `image`, `image/png`, `image/jpeg`, `net/http`, `os`, `os/exec`, `path/filepath`

---

### 2. Project Structure

Organize your project as follows:

```
rembg-go/
├── cmd/
│   └── rembg/
│       └── main.go
├── pkg/
│   ├── models/
│   │   ├── session.go
│   │   ├── u2net.go
│   │   └── downloader.go
│   ├── processing/
│   │   ├── background.go
│   │   ├── image.go
│   │   └── mask.go
│   ├── video/
│   │   └── stream.go
│   └── utils/
│       ├── normalize.go
│       └── resize.go
├── models/
├── examples/
│   ├── simple/
│   │   └── main.go
│   └── video/
│       └── main.go
├── go.mod
├── go.sum
└── README.md
```

---

### 3. Implementation Steps

#### a. Background Removal

- Implement `remove()` in `pkg/processing/background.go` to process images and masks.

#### b. Session Management

- Create `new_session()` logic in `pkg/models/session.go` for ONNX model loading and reuse.

#### c. Video Processing

- Use OpenCV via the `gocv` Go bindings (`gocv.io/x/gocv`) in `pkg/video/gocv_video.go` to extract frames from video files.

#### d. Model Support

- Start with a single ONNX model (e.g., `u2netp` or `silueta`) in `pkg/models/u2net.go`.

---

### 4. Usage Example

- Provide a simple CLI in `cmd/rembg/main.go` using Cobra.
- Add example scripts in `examples/simple/main.go` and `examples/video/main.go`.

---

### 5. Notes

- Focus first on the core `remove()` and session management logic.
- Video processing should extract RGB24 frames for background removal.
- Follow Go modular conventions for maintainability.

---

For more details, see the [Python API documentation](https://github.com/danielgatis/rembg/wiki/Python-API).

---

### Quick Start (macOS)

1. Install Go (if needed):

```bash
brew install go
```

2. Ensure OpenCV and `gocv` are installed for video examples:

```bash
brew install opencv pkg-config
```
# install gocv Go bindings (requires CGO and the OpenCV dev libs)
go install gocv.io/x/gocv@latest

3. From the repository root:

```bash
go mod tidy
go build ./...
```

4. Run examples:

```bash
# image example (place examples/simple/example.png first)
go run ./examples/simple

# video example (place input.mp4 in repo root)
go run ./examples/video
```

---

### Next implementation milestones

- Integrate ONNX runtime and implement `pkg/models/session.go` using `github.com/yalue/onnxruntime_go`.
- Implement model inference and mask postprocessing in `pkg/processing/background.go`.
- Add streaming OpenCV integration (use gocv.VideoCapture and a worker pool) to process frames concurrently.
- Add tests and example model download scripts.

---

### Alternative inference backends: SageMaker & Triton

If you don't want to run ONNX locally (or want a hosted/serving option), you can call inference backends remotely. Below are practical Go examples and tips for working with AWS SageMaker Endpoints and NVIDIA Triton Inference Server (gRPC).

#### A. AWS SageMaker Endpoint (InvokeEndpoint)

When you have a SageMaker endpoint deployed (for example serving a U2Net model that accepts an image and returns a mask), use the AWS SDK for Go v2 `sagemakerruntime` client to invoke the endpoint.

Key points:
- Send the image bytes (PNG/JPEG) as the request Body.
- Use ContentType that the model expects (e.g. `application/octet-stream` or `image/png`).
- The endpoint response Body contains the raw model output (mask bytes or serialized image) which you then decode and composite.

Example (Go):

```go
// InvokeSageMaker calls a SageMaker real-time endpoint with raw image bytes and returns the raw response bytes.
func InvokeSageMaker(ctx context.Context, endpoint string, payload []byte) ([]byte, error) {
	// Requires: github.com/aws/aws-sdk-go-v2/config and github.com/aws/aws-sdk-go-v2/service/sagemakerruntime
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := sagemakerruntime.NewFromConfig(cfg)

	input := &sagemakerruntime.InvokeEndpointInput{
		Body:        bytes.NewReader(payload),
		EndpointName: aws.String(endpoint),
		ContentType: aws.String("application/octet-stream"),
	}
	resp, err := client.InvokeEndpoint(ctx, input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
```

Suggested workflow:
- Preprocess and encode the image (PNG/JPEG) in memory.
- Call `InvokeSageMaker` with endpoint name.
- Parse returned bytes — if the model returns a PNG mask, decode with `image/png`; if it returns raw float mask, parse accordingly and convert to an alpha mask.

Security & performance notes:
- Make sure IAM credentials are available to the process (environment, profile, or EC2/ECS role).
- SageMaker charges per instance + inference; batch multiple images in one request if the model supports batching.

#### B. NVIDIA Triton Inference Server (gRPC) — outline

Triton supports both HTTP/REST and gRPC. gRPC is lower-latency and is recommended for high-throughput inference. The recommended flow:

1. Fetch/provide Triton proto files from the Triton repo (they define the gRPC API).
2. Generate Go bindings with `protoc` + `protoc-gen-go` / `protoc-gen-go-grpc`.
3. Use `google.golang.org/grpc` to dial the server and call the inference methods.

Important: Triton's exact request/response types are in the `inference` / `grpc` proto files shipped with the Triton repository — generate Go code from those protos before using the sample below.

High-level steps to generate Go stubs (on your dev machine):

```bash
# clone triton server (or download proto files)
git clone https://github.com/triton-inference-server/server.git

# generate Go code (example paths — adjust to where protos live)
protoc --proto_path=server --go_out=. --go-grpc_out=. \
  server/src/core/grpc_service.proto server/src/core/infer.proto
```

gRPC client outline (Go):

```go
// NOTE: this is an outline. Replace `grpcservice` with the generated package name.
func InferTritonGRPC(ctx context.Context, addr, modelName string, input []byte) ([]byte, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := grpcservice.NewGRPCInferenceServiceClient(conn)

	// Build an InferRequest according to the generated protos.
	req := &grpcservice.InferRequest{
		ModelName: modelName,
		// Inputs: create ModelInferRequest_InferInputTensor entries with RawContents=payload
	}

	resp, err := client.Infer(ctx, req)
	if err != nil {
		return nil, err
	}

	// Extract output raw bytes from resp.Outputs (depending on model, may be raw binary or tensor contents)
	return resp.Outputs[0].RawContents, nil // pseudo-code: adapt to real proto
}
```

Notes & tips for Triton gRPC:
- Triton expects structured requests: model name, input tensors with names, datatypes, and shapes. Use RawContents to send binary bytes.
- Many image models accept NCHW or NHWC tensors; convert and normalize images before calling.
- Use multiple concurrent gRPC streams for throughput; reuse gRPC connections.
- If you prefer not to generate protos, you can use Triton's HTTP/V2 API (simpler to call from Go with net/http) as a fallback.

References
- Triton server repo (protos & examples): https://github.com/triton-inference-server/server
- AWS SDK for Go v2: https://aws.github.io/aws-sdk-go-v2/


 