package video

import (
	"bytes"
	"context"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/unrealandychan/rembg-go/pkg/backends"
	"github.com/unrealandychan/rembg-go/pkg/processing"
)

type frameTask struct {
	img  image.Image
	path string
}

func saveFrameWorker(tasks <-chan frameTask, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		of, err := os.Create(task.path)
		if err != nil {
			fmt.Printf("create output file: %v\n", err)
			continue
		}
		if err := png.Encode(of, task.img); err != nil {
			fmt.Printf("encode png: %v\n", err)
		}
		of.Close()
	}
}

// ExtractFramesGocv opens videoPath with OpenCV and writes frames as PNG into outDir.
func ExtractFramesGocv(videoPath, outDir string) error {
	start := time.Now()
	if videoPath == "" {
		return fmt.Errorf("videoPath required")
	}
	if outDir == "" {
		outDir = "frames"
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	captureFile, err := gocv.VideoCaptureFile(videoPath)
	if err != nil {
		return fmt.Errorf("open video: %w", err)
	}
	defer func(cap *gocv.VideoCapture) {
		err := cap.Close()
		if err != nil {
			fmt.Printf("release video capture: %v\n", err)
		}
	}(captureFile)

	img := gocv.NewMat()
	defer func(img *gocv.Mat) {
		err := img.Close()
		if err != nil {
			fmt.Printf("release mat: %v\n", err)
		}
	}(&img)
	var workerCount = runtime.NumCPU()
	tasks := make(chan frameTask, workerCount*2)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go saveFrameWorker(tasks, &wg)
	}

	idx := 0
	for {
		if ok := captureFile.Read(&img); !ok {
			break
		}
		if img.Empty() {
			continue
		}
		fmt.Printf("Extracting frame %d\r", idx)
		imgGo, err := img.ToImage()
		if err != nil {
			return fmt.Errorf("convert mat to image: %w", err)
		}
		rmbgImg, err := processing.RemoveBackground(imgGo, processing.RemoveBackgroundOptions{
			PostProcessMask: true,
			OnlyMask:        false,
			AlphaMatting:    false,
			PutAlpha:        true,
			BackgroundColor: nil,
			ReturnType:      "image",
			ModelPath:       "./u2net.onnx",
		})
		if err != nil {
			return fmt.Errorf("remove background: %w\n", err)
		}
		outPath := filepath.Join(outDir, fmt.Sprintf("frame_%04d.png", idx))
		tasks <- frameTask{img: rmbgImg, path: outPath}
		idx++
	}
	close(tasks)
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Frame extraction took %s\n", elapsed)
	return nil
}

// RemoveBackgroundForVideo processes all PNG frames in inputDir, removes background, and saves to outputDir.
func RemoveBackgroundForVideo(ctx context.Context, backend backends.Backend, inputDir, outputDir string, opts processing.RemoveBackgroundOptions) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	frames, err := filepath.Glob(filepath.Join(inputDir, "frame_*.png"))
	if err != nil {
		return err
	}
	if len(frames) == 0 {
		return fmt.Errorf("no frames found in %s", inputDir)
	}
	workerCount := runtime.NumCPU()
	tasks := make(chan string, workerCount*2)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for framePath := range tasks {
				frameBytes, err := os.ReadFile(framePath)
				if err != nil {
					fmt.Printf("read frame: %v\n", err)
					continue
				}
				result, err := backends.RemoveBackgroundWithBackend(ctx, backend, frameBytes, opts)
				if err != nil {
					fmt.Printf("remove bg: %v\n", err)
					continue
				}
				outPath := filepath.Join(outputDir, filepath.Base(framePath))
				var outBytes []byte
				switch v := result.(type) {
				case []byte:
					outBytes = v
				case image.Image:
					buf := new(bytes.Buffer)
					if err := png.Encode(buf, v); err != nil {
						fmt.Printf("encode png: %v\n", err)
						continue
					}
					outBytes = buf.Bytes()
				default:
					fmt.Printf("unknown result type for %s\n", framePath)
					continue
				}
				if err := os.WriteFile(outPath, outBytes, 0o644); err != nil {
					fmt.Printf("write output: %v\n", err)
				}
			}
		}()
	}
	for _, frame := range frames {
		tasks <- frame
	}
	close(tasks)
	wg.Wait()
	return nil
}
