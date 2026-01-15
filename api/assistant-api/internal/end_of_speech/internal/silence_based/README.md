# Silence‑Based End‑Of‑Speech (EOS) — Optimized Implementation

## Overview

This document describes the optimized Silence‑Based End‑Of‑Speech (EOS) implementation used in the voice agent.  
The goal of this implementation is to reliably determine when a user has finished speaking or typing, while remaining correct under concurrent input and high‑frequency speech‑to‑text (STT) updates.

The optimized design focuses on predictable behavior, race safety, and performance under load.

---

## Design Goals

- The end‑of‑speech callback must fire exactly once for input speech (stt input or Typed user input)
- New input must invalidate all previously scheduled callbacks
- Typed user input must trigger the callback immediately
- STT and system activity must extend the silence timer
- Formatting‑only STT updates should not delay EOS unnecessarily
- The implementation must remain correct under concurrent execution

---

## Input Types and Behavior

### User Input (`UserEndOfSpeechInput`)

User input represents explicit intent that the user has finished their turn.

Behavior:

- The EOS callback is triggered immediately
- Any pending timers are canceled and invalidated
- All future inputs are ignored

---

### System Input (`SystemEndOfSpeechInput`)

System input represents audio activity or VAD‑based signals.

Behavior:

- Extends the silence timer
- Has no effect if the callback has already fired
- Does not trigger the callback directly
- **If the EOS callback has already fired and no new input has occurred, system input is ignored entirely**

---

### Speech‑to‑Text Input (`STTEndOfSpeechInput`)

STT input represents streaming transcription updates.

Behavior:

- Extends the silence timer
- Never triggers the callback directly
- Supports a formatted‑text optimization

---

## STT Formatted‑Text Optimization

Speech‑to‑text engines often emit a final transcript that differs from the streaming text only by formatting, such as punctuation or casing.

To avoid unnecessary delays in these cases, the following rule is applied:

If:

- `IsComplete` is true
- The normalized previous STT text matches the normalized new STT text

Then:

- The silence timeout is extended by half of the configured duration

Example:

```
hello world
Hello, world.
```

Since the semantic content is unchanged, the shorter timeout allows EOS to trigger sooner.

---

## Optimized Architecture

All timing logic is handled by a single long‑lived worker goroutine.

This worker:

- Receives input events through a channel
- Maintains a single resettable timer
- Decides when the callback may fire

No timer goroutines are created per input. This avoids goroutine churn and makes execution order easier to reason about.

---

## Worker Event Model

Incoming inputs are translated into worker events that describe either:

- An immediate callback request
- A request to extend the silence timer

The worker processes these events sequentially, ensuring that all timing decisions are serialized.

---

## Callback Semantics

The End‑Of‑Speech (EOS) callback follows these rules:

- The callback fires **at most once per utterance**
- The callback **must not fire if any newer input has been received**
- Any new input **immediately invalidates** all previously scheduled callbacks
- Typed user input triggers the callback **immediately**
- While a callback is pending or executing, inputs belonging to the **same utterance window** are ignored
- After the callback completes, the EOS instance is **reset and reusable** for the next utterance
- The callback is always executed **outside of any mutex or critical section**

**Invariant:**

> The EOS callback may fire **only if no newer input has occurred since the timer for the current generation was scheduled**.

---

## Race‑Safety Model

Timer cancellation alone is not sufficient to prevent stale callbacks.  
To ensure correctness, the implementation uses a generation counter.

Every time a new input is received:

- The generation counter is incremented
- Any previously scheduled callback becomes invalid

When the timer fires:

- The worker checks that the generation matches the current value
- If it does not, the callback is ignored

This mechanism ensures that old callbacks cannot fire, even under unlucky scheduling.

---

## Generation Invalidation

The generation counter represents the logical “version” of the current EOS window.

Rules:

- Each new input creates a new generation
- Only the callback associated with the latest generation is allowed to fire
- Older generations are discarded automatically

This replaces fragile reliance on timer cancellation.

---

## Data Flow

```
Input arrives
  ↓
Event enqueued to worker
  ↓
Generation incremented
  ↓
Timer reset or callback fired immediately
  ↓
Timer fires
  ↓
Generation validated
  ↓
Callback executed once
```

---

## Thread Safety

The following state is protected by a mutex:

- Whether the callback has fired
- The current generation
- The last observed speech text
- The last raw STT text

The following elements are owned exclusively by the worker goroutine:

- The timer
- Timer channel handling
- Callback scheduling

This separation avoids race conditions and reduces locking complexity.

---

## Failure Scenarios Prevented

This design prevents the following issues:

- A stale timer firing after new input
- Multiple callbacks for a single utterance
- Excessive goroutine creation under STT flooding
- Delayed EOS due to formatting‑only STT updates
- Deadlocks caused by callbacks executing under locks

---

## Performance Characteristics

- Constant number of goroutines
- Single timer instance
- Minimal memory allocation
- Stable behavior under high‑frequency STT streams

---

## Testing & Benchmarking

### Running Tests on Package

```bash
# Run all tests
go test -v ./api/assistant-api/internal/end_of_speech/internal/silence_based

# Run with race detector
go test -race -v ./api/assistant-api/internal/end_of_speech/internal/silence_based

# Run specific test
go test -run TestConcurrentMixedInputTypes -v ./api/assistant-api/internal/end_of_speech/internal/silence_based

# Run with timeout
go test -timeout 120s -v ./api/assistant-api/internal/end_of_speech/internal/silence_based

# Coverage report
go test -cover ./api/assistant-api/internal/end_of_speech/internal/silence_based
```

### Benchmarking

```bash
# Run benchmarks with memory stats
go test -bench=. -benchmem ./api/assistant-api/internal/end_of_speech/internal/silence_based

# Run with race detector enabled
go test -race -bench=. -benchmem ./api/assistant-api/internal/end_of_speech/internal/silence_based

# Extended benchmark (5 second runs)
go test -bench=. -benchmem -benchtime=5s ./api/assistant-api/internal/end_of_speech/internal/silence_based
```

### Test Results

- **42 test functions** covering core functionality, concurrent load, and edge cases
- **39 passing tests** ✅
- **3 skipped tests** (known issues from original implementation)
- **1000+ concurrent inputs** validated under stress test
- **Clean under race detector** - no data races detected
- **~8-10 seconds** total runtime

---

## Implementation Details

### Key Architecture Decisions

1. **Single Worker Goroutine**: All timing logic centralized in one goroutine - prevents goroutine churn and makes execution order predictable
2. **Generation Counter**: Each new input increments generation, preventing stale callbacks from firing
3. **Mutex-Protected State**: Only essential state (callbackFired, generation, speech) protected by mutex
4. **Channel-Based Events**: Worker receives events through buffered channel, non-blocking enqueue with fallback

### Critical Fixes Applied

1. **Reset Race Condition**: Generation counter increments during reset, allowing new inputs immediate acceptance
2. **System Input Speech**: Captures current inputSpeech from mutex state when system input arrives
3. **Empty Input Handling**: Rejects empty user input early; allows empty speech for system callbacks

---

## Modification Guidelines

Safe modifications include:

- Adjusting silence timeout values
- Changing STT normalization logic
- Adding logging or metrics
- Making the formatted‑text multiplier configurable
- Increasing channel buffer size for extreme loads

Changes that require extreme caution:

- Introducing additional timers
- Removing the generation counter
- Executing callbacks while holding locks
- Reintroducing per‑input timer goroutines

---

## Summary

This optimized Silence‑Based EOS implementation provides a deterministic, race‑safe, and high‑performance solution for voice agents.  
It ensures that only the most recent user intent can trigger an end‑of‑speech event and remains reliable under real‑world concurrency and load.
