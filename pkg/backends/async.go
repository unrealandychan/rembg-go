package backends

import (
    "context"
)

// InferResult holds an asynchronous inference result.
type InferResult struct {
    Data []byte
    Err  error
}

// InferAsync calls Backend.Infer in a new goroutine and returns a channel
// which will receive the result. The channel is buffered with size 1.
// The caller should select on the returned channel and ctx.Done() as needed.
func InferAsync(ctx context.Context, b Backend, payload []byte) <-chan InferResult {
    ch := make(chan InferResult, 1)
    go func() {
        data, err := b.Infer(ctx, payload)
        select {
        case ch <- InferResult{Data: data, Err: err}:
        case <-ctx.Done():
            // If context was cancelled before we could send, prefer returning ctx.Err
            ch <- InferResult{Data: nil, Err: ctx.Err()}
        }
    }()
    return ch
}

// pooledJob is an internal job submitted to the pool.
type pooledJob struct {
    ctx     context.Context
    payload []byte
    resp    chan InferResult
}

// PooledBackend is a wrapper that runs underlying Backend.Infer calls
// on a fixed-size pool of worker goroutines. It implements Backend.
type PooledBackend struct {
    jobs chan pooledJob
}

// NewPooledBackend wraps an existing Backend with a worker pool of size workers.
// Use workers = runtime.GOMAXPROCS(0) or a tuned number for your workload.
func NewPooledBackend(b Backend, workers int) *PooledBackend {
    if workers <= 0 {
        workers = 4
    }
    p := &PooledBackend{jobs: make(chan pooledJob)}

    // start workers
    for i := 0; i < workers; i++ {
        go func() {
            for job := range p.jobs {
                data, err := b.Infer(job.ctx, job.payload)
                // try to deliver result, but respect job ctx
                select {
                case job.resp <- InferResult{Data: data, Err: err}:
                case <-job.ctx.Done():
                    job.resp <- InferResult{Data: nil, Err: job.ctx.Err()}
                }
            }
        }()
    }
    return p
}

// Infer submits the payload to the worker pool and blocks until a result is available
// or the provided context is cancelled.
func (p *PooledBackend) Infer(ctx context.Context, payload []byte) ([]byte, error) {
    resp := make(chan InferResult, 1)
    job := pooledJob{ctx: ctx, payload: payload, resp: resp}

    select {
    case p.jobs <- job:
        // submitted
    case <-ctx.Done():
        return nil, ctx.Err()
    }

    select {
    case r := <-resp:
        return r.Data, r.Err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

// Close shuts down the worker pool by closing the jobs channel. After Close,
// further Infer calls will panic; create a new pooled backend instead.
func (p *PooledBackend) Close() {
    close(p.jobs)
}
