package models

// U2Net represents configuration for a U2Net ONNX model.
type U2Net struct {
    Path string
}

func NewU2Net(path string) *U2Net {
    return &U2Net{Path: path}
}
