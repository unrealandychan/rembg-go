package main

import (
    "fmt"
    "rembg-go/pkg/video"
)

func main() {
    if err := video.ExtractFrames("input.mp4", "frames"); err != nil {
        fmt.Println("extract frames failed:", err)
        return
    }
    fmt.Println("frames extracted to ./frames")
}
