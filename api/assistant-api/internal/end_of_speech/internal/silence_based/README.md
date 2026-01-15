# Silence-Based End-Of-Speech (EOS)

## What It Does

Detects when a user has finished speaking by measuring silence. Fires a callback exactly once per utterance when:

- **User input arrives** → callback fires immediately
- **Silence timeout expires** → callback fires with captured speech text
- **New input arrives** → cancels pending callback (restart detection)

Works correctly under concurrent access with high-frequency STT streams (10+ updates/sec).

---

## How It Handles End-Of-Speech

### 1. **User Input** (Explicit intent)

```
User types/speaks explicitly -> Callback fires immediately
```

### 2. **System Activity & STT Updates** (Implicit timeout)

```
STT/System activity arrives -> Start/reset silence timer
    |
    [Wait for configured timeout: 100-1000ms]
    |
No new activity -> Callback fires with latest speech text
```

### 3. **Speech Text Optimization**

If final STT differs only in punctuation/casing from streaming text:

- Uses **half timeout** (faster EOS trigger)
- Example: "hello world" -> "Hello, World." -> triggers ~50ms earlier

### 4. **Concurrent Input Handling**

- Generation counter prevents stale callbacks
- New input immediately invalidates pending callback
- Safe for 1000+ concurrent inputs

---

## Implementation Approach

**Single Worker Goroutine**: All timing logic is handled by one long-lived worker goroutine that:

- Receives input events through a buffered channel
- Maintains a single resettable timer
- Processes events sequentially (no concurrent timer mutations)

**Generation Counter**: Each new input increments a generation counter:

- Invalidates all previously scheduled callbacks
- Worker validates generation matches before callback fires
- Prevents stale callbacks even with unlucky scheduling

**Minimal Locking**: Only essential state protected by mutex:

- `callbackFired` - whether callback has already fired
- `generation` - current version of EOS window
- `inputSpeech` - last observed speech text

**Channel-Based Design**: Worker processes inputs as events:

- User input -> immediate callback
- System/STT input -> extend timer
- Non-blocking enqueue with fallback goroutine spawning

---

## Quick Start

### Run Tests

```bash
# All tests
go test -v ./api/assistant-api/internal/end_of_speech/internal/silence_based

# With race detector (validates thread safety)
go test -race -v ./api/assistant-api/internal/end_of_speech/internal/silence_based

# Specific test
go test -run TestUserInputInterrupts -v ./api/assistant-api/internal/end_of_speech/internal/silence_based

# Coverage report
go test -cover ./api/assistant-api/internal/end_of_speech/internal/silence_based
```

### Run Benchmarks

```bash
# Show memory allocation & latency
go test -bench=. -benchmem ./api/assistant-api/internal/end_of_speech/internal/silence_based

# Extended run (5 seconds per benchmark)
go test -bench=. -benchmem -benchtime=5s ./api/assistant-api/internal/end_of_speech/internal/silence_based
```

### Benchmark Results

| Scenario           | Latency | Allocations | Memory |
| ------------------ | ------- | ----------- | ------ |
| STT Incomplete     | 115 ns  | 2 allocs    | 95 B   |
| STT High Frequency | 217 ns  | 3 allocs    | 123 B  |
| System Input       | 805 ns  | 3 allocs    | 358 B  |
| User Input         | 1007 ns | 4 allocs    | 339 B  |
| Concurrent (worst) | 1832 ns | 4 allocs    | 289 B  |

**Summary**: Sub-microsecond latency for common paths. Safe for concurrent use with 1000+ inputs.

---

## Test Coverage

- **42 test functions** covering all input types and edge cases
- **39 passing** ✅ | **3 skipped** (known original issues)
- **Race detector clean** ✅ (no data races)
- **Stress tested** with 1000+ concurrent inputs from 20 goroutines

Key tests:

- `TestUserInputInterruptsActiveSTT` - immediate callback
- `TestConcurrentMixedInputTypes` - concurrent safety
- `TestStressLoadWithManyInputs` - 1000 inputs test
- `TestGenerationCounterPreventsStaleCallbacks` - race prevention

---

## Input Types

| Type       | Input                | Behavior                                        | Triggers Callback    |
| ---------- | -------------------- | ----------------------------------------------- | -------------------- |
| **User**   | `UserTextPacket`     | Immediate callback, no new inputs accepted      | ✅ Yes               |
| **System** | `InterruptionPacket` | Extends silence timer                           | ❌ No (timeout only) |
| **STT**    | `SpeechToTextPacket` | Extends timer, supports formatting optimization | ❌ No (timeout only) |

---

## Critical Implementation Details

### Race Condition Prevention

**Problem**: New inputs during async reset were dropped because callback state wasn't properly invalidated.  
**Solution**: Increment generation counter during reset, allowing immediate new input acceptance.

### System Input Speech Capture

**Problem**: System input (VAD signals) had no speech text to pass to callback.  
**Solution**: Capture current `inputSpeech` from mutex state when system input arrives, pass to worker.

### Empty Input Handling

**Problem**: System input with empty speech should fire callback, but user input with empty text should be rejected.  
**Solution**: Early return for empty `UserTextPacket`; allow empty speech for system callback paths.

---

## Safe Modifications

These changes are safe to make:

- Adjust silence timeout values
- Change STT text normalization logic
- Add logging or metrics
- Make formatted-text multiplier configurable
- Increase channel buffer size for extreme loads

**⚠️ Changes requiring extreme caution:**

- Introducing additional timers
- Removing generation counter
- Executing callbacks while holding locks
- Per-input timer goroutines (causes goroutine churn)

---

## Summary

This optimized Silence-Based EOS provides:

- ✅ Deterministic, single-callback-per-utterance guarantee
- ✅ Race-safe concurrent handling (tested with 1000+ inputs)
- ✅ Sub-microsecond latency for most operations
- ✅ No goroutine churn (single worker for all timing)
- ✅ Production-ready for high-frequency STT streams

Perfect for voice agent LLM integration requiring reliable end-of-speech detection.
