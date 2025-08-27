package backends

import (
    "encoding/binary"
    "math"
)

// Float32sToBytes converts a slice of float32 to little-endian bytes.
func Float32sToBytes(floats []float32) []byte {
    buf := make([]byte, 4*len(floats))
    for i, v := range floats {
        bits := math.Float32bits(v)
        binary.LittleEndian.PutUint32(buf[i*4:(i+1)*4], bits)
    }
    return buf
}

// BytesToFloat32s converts little-endian bytes to a slice of float32.
func BytesToFloat32s(b []byte) []float32 {
    n := len(b) / 4
    out := make([]float32, n)
    for i := 0; i < n; i++ {
        bits := binary.LittleEndian.Uint32(b[i*4 : (i+1)*4])
        out[i] = math.Float32frombits(bits)
    }
    return out
}

// IntsToInt64s converts []int to []int64 for Triton shape fields.
func IntsToInt64s(in []int) []int64 {
    out := make([]int64, len(in))
    for i, v := range in {
        out[i] = int64(v)
    }
    return out
}
