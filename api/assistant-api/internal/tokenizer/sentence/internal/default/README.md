# SentenceTokenizer Test Suite

This directory contains comprehensive test cases and benchmarks for the `SentenceTokenizer` implementation.

## Files

- **default_sentence_tokenizer.go** - Main tokenizer implementation
- **default_sentence_tokenizer_test.go** - Comprehensive unit tests (18 KB, 17 test cases)
- **default_sentence_tokenizer_benchmark_test.go** - Performance benchmarks (11 KB, 16 benchmark cases)

## Running Tests

### Run all tests

```bash
go test -v ./api/assistant-api/internal/tokenizer/sentence/default/
```

### Run specific test

```bash
go test -v -run TestSingleSentence ./api/assistant-api/internal/tokenizer/sentence/default/
```

### Run benchmarks

```bash
go test -bench=. -benchmem ./api/assistant-api/internal/tokenizer/sentence/default/
```

### Run benchmarks with specific pattern

```bash
go test -bench=BenchmarkSingle -benchmem ./api/assistant-api/internal/tokenizer/sentence/default/
```

## Test Cases

### Initialization Tests

- `TestNewSentenceTokenizer` - Tests tokenizer creation with various boundary configurations

### Basic Functionality Tests

- `TestSingleSentence` - Tests tokenization of a single sentence
- `TestMultipleSentences` - Tests tokenization of multiple consecutive sentences
- `TestEmptyInput` - Tests handling of empty input
- `TestLargeBatch` - Tests processing 100 sentences in a batch

### Boundary & Delimiter Tests

- `TestMultipleBoundaries` - Tests various sentence delimiters (.,?!;:)
- `TestSpecialCharacterBoundaries` - Tests regex special characters as boundaries
- `TestNoBoundariesDefined` - Tests behavior when no boundaries are configured

### Context & State Management Tests

- `TestContextSwitching` - Tests switching between multiple speakers/contexts
- `TestBufferStateMaintenance` - Tests proper buffer state across multiple calls
- `TestWhitespaceHandling` - Tests trimming and handling of whitespace
- `TestIsCompleteFlag` - Tests the IsComplete flag for forcing sentence completion
- `TestContextCancellation` - Tests context cancellation handling

### Concurrency Tests

- `TestConcurrentContexts` - Tests multiple concurrent speaker contexts
- `TestStringRepresentation` - Tests the String() method for debugging

### Lifecycle Tests

- `TestMultipleClose` - Tests calling Close() multiple times safely
- `TestResultChannelClosure` - Tests that result channel is properly closed

## Benchmarks

Performance benchmarks measure various scenarios:

### Creation & Setup

- `BenchmarkNewSentenceTokenizer` - Tokenizer creation with boundaries (~1.7 μs)
- `BenchmarkNewSentenceTokenizerNoBoundaries` - Tokenizer creation without boundaries (~0.5 μs)
- `BenchmarkClosing` - Cost of closing tokenizer (~1.4 μs)

### Basic Operations

- `BenchmarkSingleSentenceTokenization` - Processing one sentence (~1.7 μs)
- `BenchmarkMultipleSentences` - Processing three sentences (~2.1 μs)
- `BenchmarkCompleteFlag` - Processing with IsComplete flag (~0.6 μs)

### Advanced Scenarios

- `BenchmarkLargeSentences` - Processing 1000+ words (~2.7 μs)
- `BenchmarkMultipleBoundaries` - Multiple delimiters (~2.2 μs)
- `BenchmarkContextSwitching` - Switching between 5 speakers (~6.6 μs)
- `BenchmarkStreamingLargeText` - Processing streaming chunks (~2.3 μs)
- `BenchmarkComplexScenario` - Realistic conversation (~2.6 μs)

### Specialized Scenarios

- `BenchmarkBufferingWithoutBoundaries` - Buffering without delimiters (~0.7 μs)
- `BenchmarkEmptyAndCompleteFlush` - Flushing empty buffers (~1.4 μs)
- `BenchmarkWhitespaceProcessing` - Text with various whitespace (~1.8 μs)
- `BenchmarkResultChannelConsumption` - Consuming results from channel (~4.3 μs)
- `BenchmarkParallelProcessing` - Parallel execution (~1.4 μs)

## Test Coverage

The test suite covers:

✅ **Initialization** - Creation with various configurations
✅ **Basic tokenization** - Single and multiple sentences
✅ **Boundary detection** - Multiple delimiters and special characters
✅ **Context management** - Context switching and cancellation
✅ **Buffer state** - Proper state maintenance across calls
✅ **Whitespace handling** - Trimming and formatting
✅ **Completion flags** - Force completion of incomplete sentences
✅ **Concurrent access** - Multiple simultaneous contexts
✅ **Lifecycle** - Safe cleanup and multiple Close() calls
✅ **Performance** - Benchmarks for various scenarios

## Mock Implementations

The test suite includes mock implementations:

- `mockLogger` - Full Logger interface implementation for unit tests
- `benchMockLogger` - Minimal Logger implementation for benchmarks
- `newMockOptions` - Helper to create utils.Option maps

## Key Metrics

From the benchmark results on Apple M1 Pro:

| Operation                           | Time   | Allocations |
| ----------------------------------- | ------ | ----------- |
| Create tokenizer with boundaries    | 1.7 μs | 33 allocs   |
| Create tokenizer without boundaries | 0.5 μs | 7 allocs    |
| Single sentence                     | 1.7 μs | 34 allocs   |
| Multiple sentences (3)              | 2.1 μs | 42 allocs   |
| Large sentences (1000 words)        | 2.7 μs | 34 allocs   |
| Context switching (5 speakers)      | 6.6 μs | 120 allocs  |
| Parallel processing                 | 1.4 μs | 34 allocs   |

## Notes

- All 17 unit tests pass
- All 16 benchmarks complete successfully
- No race conditions detected
- Memory allocations are consistent and predictable
- Performance scales well with input size
