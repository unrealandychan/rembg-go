package backends

import (
    "reflect"
    "testing"
)

func TestFloat32sBytesRoundtrip(t *testing.T) {
    src := []float32{0.1, -1.5, 3.1415}
    b := Float32sToBytes(src)
    got := BytesToFloat32s(b)
    if !reflect.DeepEqual(src, got) {
        t.Fatalf("roundtrip mismatch: want=%v got=%v", src, got)
    }
}

func TestIntsToInt64s(t *testing.T) {
    in := []int{1, 3, 320, 320}
    out := IntsToInt64s(in)
    want := []int64{1, 3, 320, 320}
    if !reflect.DeepEqual(out, want) {
        t.Fatalf("converted shapes mismatch: want=%v got=%v", want, out)
    }
}
