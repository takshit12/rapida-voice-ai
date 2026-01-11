# Silero VAD — Benchmark Report (macOS, Apple M1 Pro)

This report summarizes performance of `silero_vad` on Apple M1 Pro.

## Environment

- OS: macOS (Darwin, arm64)
- CPU: Apple M1 Pro
- Go: go1.25 (toolchain)
- Model: Silero VAD (onnx)
- Sample format: 16 kHz mono LINEAR16 (internal), resampled from inputs

## Summary

- Initialization: ~39.2 ms/op
- Real-time throughput: ~4.65M samples/sec (~290× real-time)
- Typical processing cost (500 ms chunk): ~1.68 ms, ~101 KB, ~199 allocs
- Typical processing cost (1 s chunk): ~3.45 ms, ~202 KB, ~407 allocs
- Small chunk (20 ms): ~1.33 µs, ~4.0 KB, ~6 allocs
- Parallel scaling: moderate overhead (8 streams: ~3.28 ms/op, ~809 KB, ~1609 allocs)

## Detailed Results (selected)

- Silence 500 ms: 1,681,311 ns/op; 101,041 B/op; 199 allocs/op
- Speech 500 ms: 1,679,572 ns/op; 101,041 B/op; 199 allocs/op
- Silence 1 s: 3,493,318 ns/op; 202,162 B/op; 407 allocs/op
- Speech 1 s: 3,447,289 ns/op; 202,161 B/op; 407 allocs/op
- Chunk 20 ms: 1,330 ns/op; 4,090 B/op; 6 allocs/op
- Chunk 50 ms: 119,992 ns/op; 10,000 B/op; 17 allocs/op
- Chunk 200 ms: 667,062 ns/op; 41,984 B/op; 82 allocs/op
- Chunk 2 s: 6,873,835 ns/op; 404,230 B/op; 810 allocs/op
- Threshold 0.1 / 0.5 / 0.9: ~1.68–1.71 ms/op; ~101 KB; ~199 allocs/op
- Parallel 2 streams: 1,866,699 ns/op; 202,257 B/op; 403 allocs/op
- Parallel 4 streams: 2,006,656 ns/op; 404,563 B/op; 805 allocs/op
- Parallel 8 streams: 3,283,569 ns/op; 809,039 B/op; 1609 allocs/op
- Sequential 10 chunks: 3,366,144 ns/op; 207,200 B/op; 430 allocs/op
- Sequential 50 chunks: 16,803,266 ns/op; 1,036,007 B/op; 2150 allocs/op
- Sequential 100 chunks: 33,728,309 ns/op; 2,072,032 B/op; 4300 allocs/op
- Resample 8 kHz: 1,701,114 ns/op; 215,730 B/op; 202 allocs/op
- Resample 24 kHz: 1,712,377 ns/op; 281,267 B/op; 202 allocs/op
- Resample 48 kHz: 1,732,141 ns/op; 379,569 B/op; 202 allocs/op
- Mixed speech+silence: 3,336,965 ns/op; 202,083 B/op; 398 allocs/op
- Mixed alternating: 1,359,788 ns/op; 82,881 B/op; 172 allocs/op
- Initialization: 39,215,726 ns/op; 2,360 B/op; 10 allocs/op
- Memory pressure (small chunks): 69,045 ns/op; 204,522 B/op; 300 allocs/op
- Memory pressure (large chunks): 3,452,378 ns/op; 202,163 B/op; 407 allocs/op
- Throughput (real-time): 3,440,264 ns/op; 4,650,820 samples/sec; ~290.7× real-time; 202,162 B/op; 407 allocs/op

## Observations

- Efficiency: Processing is orders of magnitude faster than wall clock (≥290× real-time).
- Allocations: Scale linearly with chunk duration; ~200 allocs per 0.5–1 s chunk.
- Parallelism: Overhead grows with concurrent streams; use bounded workers.
- Resampling: Higher input rates increase transient memory, but CPU impact is minimal.

## Guidance

- For real-time voice: 20–50 ms chunks keep latency and allocations low.
- Concurrency: Prefer limited goroutine pools; avoid unbounded parallelism.
- Memory: Reuse buffers for streaming to reduce per-chunk allocations.
- Threshold: Performance insensitive to threshold; choose based on detection quality.

Generated from `go test -bench=. -benchmem ./api/assistant-api/internal/vad/silero_vad` on Apple M1 Pro.
