package video

import (
    "fmt"
    "image/png"
    "os"
    "path/filepath"

    "gocv.io/x/gocv"
)

// ExtractFramesGocv opens videoPath with OpenCV and writes frames as PNG into outDir.
func ExtractFramesGocv(videoPath, outDir string) error {
    if videoPath == "" {
        return fmt.Errorf("videoPath required")
    }
    if outDir == "" {
        outDir = "frames"
    }
    if err := os.MkdirAll(outDir, 0o755); err != nil {
        return err
    }

    cap, err := gocv.VideoCaptureFile(videoPath)
    if err != nil {
        return fmt.Errorf("open video: %w", err)
    }
    defer cap.Close()

    img := gocv.NewMat()
    defer img.Close()

    idx := 0
    for {
        if ok := cap.Read(&img); !ok {
            break // end of stream
        }
        if img.Empty() {
            continue
        }

        outPath := filepath.Join(outDir, fmt.Sprintf("frame_%04d.png", idx))
        // Convert Mat to image.Image and encode PNG
        imgGo, err := img.ToImage()
        if err != nil {
            return fmt.Errorf("convert mat to image: %w", err)
        }
        of, err := os.Create(outPath)
        if err != nil {
            return fmt.Errorf("create output file: %w", err)
        }
        if err := png.Encode(of, imgGo); err != nil {
            of.Close()
            return fmt.Errorf("encode png: %w", err)
        }
        of.Close()
        idx++
    }
    return nil
}
