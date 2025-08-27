package backends

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

// TritonInferRequest is a minimal struct for /v2/models/{model}/infer HTTP body.
type TritonInferRequest struct {
    Inputs []map[string]interface{} `json:"inputs"`
}

// InferTritonHTTP sends an HTTP inference request to Triton V2 HTTP API. It currently wraps a single raw input.
func InferTritonHTTP(ctx context.Context, addr, modelName, inputName string, inputData []byte, shape []int, dtype string) ([]byte, error) {
    url := fmt.Sprintf("http://%s/v2/models/%s/infer", addr, modelName)

    inp := map[string]interface{}{
        "name": inputName,
        "shape": shape,
        "datatype": dtype,
        // Triton accepts "data" for JSON-encoded array, or "raw_input" alternative; here we use "data" with base64 not implemented.
        "data": []interface{}{},
    }

    reqObj := TritonInferRequest{Inputs: []map[string]interface{}{inp}}
    body, err := json.Marshal(reqObj)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("triton infer failed: %s: %s", resp.Status, string(b))
    }
    return io.ReadAll(resp.Body)
}
