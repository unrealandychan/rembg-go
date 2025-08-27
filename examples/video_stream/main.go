package main

import (
    "fmt"
    "image/png"
    "os"
    "path/filepath"
    "time"

    "gocv.io/x/gocv"
    "rembg-go/pkg/backends"
    "rembg-go/pkg/processing"
)

func main() {
    // open video capture (change input.mp4 to your file)
    cap, err := gocv.VideoCaptureFile("input.mp4")
    if err != nil {
        fmt.Println("open video failed:", err)
        return
    }
    defer cap.Close()

    // prepare output dir for masks
    outDir := "masks"
    os.MkdirAll(outDir, 0o755)

    // create backend and pooled wrapper
    b := backends.NewTritonHTTPBackend("triton-host:8000", "u2net", "INPUT__0", []int{1, 3, 320, 320}, "UINT8")
    pool := backends.NewPooledBackend(b, 8)
    defer pool.Close()

    mat := gocv.NewMat()
    defer mat.Close()

    idx := 0
    for {
        if ok := cap.Read(&mat); !ok {
            break
        }
        if mat.Empty() {
            continue
        }

        // convert Mat to image.Image
        img, err := mat.ToImage()
        if err != nil {
            fmt.Println("mat to image failed:", err)
            continue
        }

        // prepare payload — in this example we encode as PNG in memory
        fpath := filepath.Join(outDir, fmt.Sprintf("frame_%04d.png", idx))
        of, _ := os.Create(fpath)
        png.Encode(of, img)
        of.Close()

        // read bytes back (in real use you'd pipe bytes directly)
        payload, _ := os.ReadFile(fpath)

        // submit to pooled backend asynchronously by calling Infer in a goroutine
        go func(i int, data []byte) {
            ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
            defer cancel()
            resp, err := pool.Infer(ctx, data)
            if err != nil {
                fmt.Println("infer failed:", err)
                return
            }
            // decode mask and write mask image
            maskImg, err := processing.DecodePNG(resp)
            if err != nil {
                fmt.Println("decode mask failed:", err)
                return
            }
            mp := filepath.Join(outDir, fmt.Sprintf("mask_%04d.png", i))
            mf, _ := os.Create(mp)
            png.Encode(mf, maskImg)
            mf.Close()
            fmt.Println("wrote mask", mp)
        }(idx, payload)

        idx++
    }

    fmt.Println("streaming submitted, waiting a bit for pending jobs")
    time.Sleep(5 * time.Second)
}
