package main

import (
	"context"
	"fmt"
	"image/png"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"rembg-go/pkg/backends"
	"rembg-go/pkg/processing"
	"rembg-go/pkg/video"
)

var rootCmd = &cobra.Command{
	Use:   "rembg",
	Short: "rembg-go: minimal background removal CLI",
}

var imageCmd = &cobra.Command{
	Use:   "image [input] [output]",
	Short: "Process a single image (stub)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		inPath := args[0]
		outPath := args[1]
		backendType, _ := cmd.Flags().GetString("backend")
		backendAddr, _ := cmd.Flags().GetString("addr")

		f, err := os.Open(inPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "open input failed:", err)
			os.Exit(1)
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read input failed:", err)
			os.Exit(1)
		}

		var b backends.Backend
		switch backendType {
		case "sagemaker":
			b = backends.NewSageMakerBackend(backendAddr)
		case "triton_http":
			// simple defaults; expects model "u2net" and input name "INPUT__0"
			b = backends.NewTritonHTTPBackend(backendAddr, "u2net", "INPUT__0", []int{1, 3, 320, 320}, "UINT8")
		case "triton_grpc":
			b = backends.NewTritonGRPCBackend(backendAddr, "u2net", "INPUT__0", []int{1, 3, 320, 320}, "UINT8")
		default:
			fmt.Fprintln(os.Stderr, "unknown backend; choose sagemaker, triton_http or triton_grpc")
			os.Exit(1)
		}

		// Run inference asynchronously with a timeout so the command stays responsive.
		reqCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		resCh := backends.InferAsync(reqCtx, b, data)
		var resp []byte
		select {
		case r := <-resCh:
			if r.Err != nil {
				fmt.Fprintln(os.Stderr, "inference failed:", r.Err)
				os.Exit(1)
			}
			resp = r.Data
		case <-reqCtx.Done():
			fmt.Fprintln(os.Stderr, "inference timeout or canceled:", reqCtx.Err())
			os.Exit(1)
		}

		// assume response is PNG mask bytes
		maskImg, err := processing.DecodePNG(resp)
		if err != nil {
			fmt.Fprintln(os.Stderr, "decode mask failed:", err)
			os.Exit(1)
		}

		// decode original image again using image/png for simplicity (supports PNG only here)
		f2, _ := os.Open(inPath)
		defer f2.Close()
		srcImg, err := png.Decode(f2)
		if err != nil {
			fmt.Fprintln(os.Stderr, "decode source failed (PNG only):", err)
			os.Exit(1)
		}

		out := processing.ApplyMask(srcImg, maskImg)
		of, err := os.Create(outPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "create output failed:", err)
			os.Exit(1)
		}
		defer of.Close()
		png.Encode(of, out)
		fmt.Println("wrote", outPath)
	},
}

var videoCmd = &cobra.Command{
	Use:   "video [input]",
	Short: "Extract frames from video using OpenCV (gocv)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		in := args[0]
		if err := video.ExtractFramesGocv(in, "frames"); err != nil {
			fmt.Fprintln(os.Stderr, "extract frames failed:", err)
			os.Exit(1)
		}
		fmt.Println("frames extracted to ./frames")
	},
}

func init() {
	rootCmd.AddCommand(imageCmd)
	rootCmd.AddCommand(videoCmd)
	imageCmd.Flags().String("backend", "sagemaker", "backend to use: sagemaker|triton_http|triton_grpc")
	imageCmd.Flags().String("addr", "", "backend address or endpoint name (SageMaker endpoint name or triton host:port)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
