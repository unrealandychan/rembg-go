package backends

import (
	"context"
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
