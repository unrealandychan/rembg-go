package backends

import (
    "context"
    "fmt"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// NewTritonGRPCConn creates a gRPC connection to a Triton server.
// Use this connection with the generated Triton gRPC client (after you generate Go code from Triton's protos).
func NewTritonGRPCConn(addr string) (*grpc.ClientConn, error) {
    if addr == "" {
        return nil, fmt.Errorf("address required")
    }
    conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, err
    }
    return conn, nil
}

// InferTritonGRPCStub is a documented stub showing how to perform inference using the generated Triton gRPC client.
// It intentionally does not reference generated proto types so this file compiles without the generated code.
// Steps to complete:
// 1. Generate Go code from Triton protos (see copilot-instructions.md):
//    protoc --proto_path=server --go_out=. --go-grpc_out=. server/src/core/grpc_service.proto server/src/core/infer.proto
// 2. Import the generated package (for example: tritonpb "path/to/generated/package").
// 3. Replace the pseudo-code below with real calls using the generated client and request/response types.
func InferTritonGRPCStub(ctx context.Context, conn *grpc.ClientConn, modelName, inputName string, inputData []byte, shape []int, dtype string) ([]byte, error) {
    // PSEUDO-CODE (replace with generated types):
    // client := tritonpb.NewGRPCInferenceServiceClient(conn)
    // req := &tritonpb.InferRequest{ ModelName: modelName, /* build inputs using ModelInferRequest_InferInputTensor etc. */ }
    // resp, err := client.Infer(ctx, req)
    // if err != nil { return nil, err }
    // // extract raw bytes from resp outputs and return

    return nil, fmt.Errorf("InferTritonGRPCStub: implement using generated Triton proto types as documented in copilot-instructions.md")
}
