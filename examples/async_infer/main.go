package main

import (
    "context"
    "fmt"
    "io/ioutil"
    "time"

    "rembg-go/pkg/backends"
)

func main() {
    // Load an example image into memory
    payload, err := ioutil.ReadFile("examples/simple/example.png")
    if err != nil {
        fmt.Println("read example image failed:", err)
        return
    }

    // Create a backend (change to your endpoint/addr)
    b := backends.NewSageMakerBackend("my-endpoint")

    // Fire multiple concurrent inferences using InferAsync
    ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
    defer cancel()

    ch1 := backends.InferAsync(ctx, b, payload)
    ch2 := backends.InferAsync(ctx, b, payload)

    for i := 0; i < 2; i++ {
        select {
        case r := <-ch1:
            if r.Err != nil {
                fmt.Println("infer1 failed:", r.Err)
            } else {
                fmt.Println("infer1 got bytes:", len(r.Data))
            }
        case r := <-ch2:
            if r.Err != nil {
                fmt.Println("infer2 failed:", r.Err)
            } else {
                fmt.Println("infer2 got bytes:", len(r.Data))
            }
        case <-ctx.Done():
            fmt.Println("timed out waiting for inferences")
            return
        }
    }
}
