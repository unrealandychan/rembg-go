//go:build triton
// +build triton

package backends

import (
    "context"
    "fmt"

    tritonpb "rembg-go/pkg/backends/tritonpb"
    "google.golang.org/grpc"
)

// This file provides a real implementation that uses generated Triton protos.
// It is build-tagged with 'triton' to avoid breaking builds before codegen.

func InferTritonGRPCStub(ctx context.Context, conn *grpc.ClientConn, modelName, inputName string, inputData []byte, shape []int, dtype string) ([]byte, error) {
    client := tritonpb.NewGRPCInferenceServiceClient(conn)

    // Build ModelInferRequest here according to Triton protos. This is a template — adapt as needed.
    // Example:
    req := &tritonpb.InferRequest{
        ModelName: modelName,
        // ... fill inputs and parameters
    }

    resp, err := client.Infer(ctx, req)
    if err != nil {
        return nil, err
    }

    if len(resp.Outputs) == 0 {
        return nil, fmt.Errorf("no outputs from Triton")
    }
    // return the raw contents of the first output
    return resp.Outputs[0].RawContents, nil
}
