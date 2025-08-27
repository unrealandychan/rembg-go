package models

import "fmt"

// Session represents a loaded ONNX model session.
type Session struct {
    ModelPath string
}

// NewSession initializes a new model session from a local model path.
func NewSession(modelPath string) (*Session, error) {
    // TODO: integrate ONNX runtime; return stub for now.
    if modelPath == "" {
        return nil, fmt.Errorf("modelPath required")
    }
    return &Session{ModelPath: modelPath}, nil
}
