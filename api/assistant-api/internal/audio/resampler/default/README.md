# Audio Resampler

High-performance, thread-safe audio resampler for the Rapida Voice AI platform. Supports multiple audio formats, sample rates, and channel configurations with efficient parallel processing.

## Features

- **Multiple Format Support**: Linear16 (PCM), MuLaw8 (G.711)
- **Sample Rate Conversion**: 8kHz, 16kHz, 24kHz, 48kHz
- **Channel Conversion**: Mono ↔ Stereo
- **Thread-Safe**: Concurrent processing with no race conditions
- **High Performance**: Parallel processing with 3.5x speedup on 8 cores
- **Low Memory Footprint**: Minimal allocations per operation
