package video

// StreamProcessor would accept RGB24 frames and forward them to the background removal pipeline.
type StreamProcessor struct{
    // TODO: add fields for session, workers, channels
}

func NewStreamProcessor() *StreamProcessor {
    return &StreamProcessor{}
}
