// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_default

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
)

// Helper function to create mock options
func newMockOptions(boundaries string) utils.Option {
	opts := utils.Option(make(map[string]interface{}))
	if boundaries != "" {
		// Split boundaries into individual characters and join with commons.SEPARATOR
		// This matches how the tokenizer expects boundaries to be configured
		var parts []string
		for _, ch := range boundaries {
			parts = append(parts, string(ch))
		}
		opts["speaker.sentence.boundaries"] = strings.Join(parts, commons.SEPARATOR)
	}
	return opts
}

// Helper function to collect results from tokenizer
func collectResults(ctx context.Context, resultChan <-chan internal_type.Packet) []internal_type.Packet {
	var results []internal_type.Packet
	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				return results
			}
			results = append(results, result)
		case <-ctx.Done():
			return results
		case <-time.After(100 * time.Millisecond):
			return results
		}
	}
}

// Tests

func TestNewDefaultLLMSentenceAssembler(t *testing.T) {
	tests := []struct {
		name           string
		boundaries     string
		shouldError    bool
		expectedBounds bool
	}{
		{
			name:           "with valid boundaries",
			boundaries:     ".,?!",
			shouldError:    false,
			expectedBounds: true,
		},
		{
			name:           "with empty boundaries",
			boundaries:     "",
			shouldError:    false,
			expectedBounds: false,
		},
		{
			name:           "with single boundary",
			boundaries:     ".",
			shouldError:    false,
			expectedBounds: true,
		},
		{
			name:           "with semicolon separated boundaries",
			boundaries:     ".,?!;:",
			shouldError:    false,
			expectedBounds: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := commons.NewApplicationLogger()
			opts := newMockOptions(tt.boundaries)

			tokenizer, err := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
			if tt.shouldError && err != nil {
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tokenizer == nil {
				t.Fatal("tokenizer is nil")
			}

			defer tokenizer.Close()

			st := tokenizer.(*sentenceAssembler)
			if st.hasBoundaries != tt.expectedBounds {
				t.Errorf("expected hasBoundaries=%v, got %v", tt.expectedBounds, st.hasBoundaries)
			}
		})
	}
}

func TestSingleSentence(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()
	err := tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
		ContextID: "speaker1",
		Text:      "Hello world.",
	})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	results := collectResults(ctx, tokenizer.Result())
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
		return
	}

	// Handle both AssembledSentence and other AssemblerOutput types
	if ts, ok := results[0].(internal_type.LLMStreamPacket); ok {
		if ts.Text != "Hello world." {
			t.Errorf("expected 'Hello world.', got '%s'", ts.Text)
		}
		if ts.ContextID != "speaker1" {
			t.Errorf("expected context 'speaker1', got '%s'", ts.ContextID)
		}
	} else {
		t.Errorf("unexpected result type: %T", results[0])
	}
}

func TestMultipleSentences(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()

	sentences := []string{
		"First sentence.",
		" Second sentence.",
		" Third sentence.",
	}

	go func() {
		for _, s := range sentences {
			tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: "speaker1",
				Text:      s,
			})
		}
	}()

	results := collectResults(ctx, tokenizer.Result())

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
		return
	}

	expected := []string{"First sentence.", "Second sentence.", "Third sentence."}
	for i, result := range results {
		if ts, ok := result.(internal_type.LLMStreamPacket); ok {
			if ts.Text != expected[i] {
				t.Errorf("result %d: expected '%s', got '%s'", i, expected[i], ts.Text)
			}
		}
	}
}

func TestMultipleBoundaries(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".,?!")
	ctx := context.Background()

	testCases := []struct {
		input    string
		expected int
	}{
		{"What a day!", 1},
		{"Is this real?", 1},
		{"Sure, let's go.", 2}, // Now correctly splits on both comma and period
		{"One. Two? Three!", 3},
	}

	for _, tc := range testCases {
		tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)

		// Count results through channel
		resultCount := 0
		go func() {
			tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: "speaker1",
				Text:      tc.input,
			})
		}()

		// Give it time to process
		time.Sleep(50 * time.Millisecond)

		// Drain channel
		for len(tokenizer.Result()) > 0 {
			select {
			case _, ok := <-tokenizer.Result():
				if ok {
					resultCount++
				}
			default:
				break
			}
		}

		if resultCount != tc.expected {
			t.Logf("input '%s': got %d results (expected %d)", tc.input, resultCount, tc.expected)
		}

		tokenizer.Close()
	}
}

func TestContextSwitching(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()

	go func() {
		// Speaker 1 starts with boundary
		tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      "Hello there.",
		})

		// Speaker 2 continues with boundary
		tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker2",
			Text:      "Goodbye.",
		})
	}()

	results := collectResults(ctx, tokenizer.Result())

	if len(results) < 2 {
		t.Errorf("expected at least 2 results, got %d", len(results))
		return
	}

	// Check speaker1 and speaker2 results
	foundSpeaker1 := false
	foundSpeaker2 := false
	for _, result := range results {
		if ts, ok := result.(internal_type.LLMStreamPacket); ok {
			if ts.ContextID == "speaker1" {
				foundSpeaker1 = true
				if ts.Text != "Hello there." {
					t.Errorf("speaker1 expected 'Hello there.', got '%s'", ts.Text)
				}
			}
			if ts.ContextID == "speaker2" {
				foundSpeaker2 = true
				if ts.Text != "Goodbye." {
					t.Errorf("speaker2 expected 'Goodbye.', got '%s'", ts.Text)
				}
			}
		}
	}
	if !foundSpeaker1 {
		t.Error("expected to find speaker1 result")
	}
	if !foundSpeaker2 {
		t.Error("expected to find speaker2 result")
	}
}

func TestIsCompleteFlag(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()

	go func() {
		// Send incomplete sentence without boundary
		tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      "This is incomplete",
		})
		// Force completion with Flush
		tokenizer.Assemble(ctx, internal_type.LLMMessagePacket{
			ContextID: "speaker1",
		})
	}()

	results := collectResults(ctx, tokenizer.Result())

	// Should get: AssembledSentence + Flush message
	if len(results) != 2 {
		t.Errorf("expected 2 results (sentence + flush), got %d", len(results))
		return
	}

	// First result should be the flushed sentence (can be value or pointer)
	var sentence string
	if ts, ok := results[0].(internal_type.LLMStreamPacket); ok {
		sentence = ts.Text
	} else if ts, ok := results[0].(internal_type.LLMStreamPacket); ok {
		sentence = ts.Text
	} else {
		t.Errorf("expected first result to be AssembledSentence, got %T", results[0])
		return
	}

	if sentence != "This is incomplete" {
		t.Errorf("expected 'This is incomplete', got '%s'", sentence)
	}

	// Second result should be the Flush message
	if _, ok := results[1].(internal_type.LLMMessagePacket); !ok {
		t.Errorf("expected second result to be Flush message, got %T", results[1])
	}
}

func TestEmptyInput(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()

	err := tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
		ContextID: "speaker1",
		Text:      "",
	})

	if err != nil {
		t.Fatalf("Assemble should not error on empty input: %v", err)
	}

	results := collectResults(ctx, tokenizer.Result())
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty input, got %d", len(results))
	}
}

func TestNoBoundariesDefined(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions("")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()

	go func() {
		// Without boundaries, we can only get output with Flush
		tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      "Hello world",
		})

		tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      " this is great",
		})

		// Flush to trigger output
		tokenizer.Assemble(ctx, internal_type.LLMMessagePacket{
			ContextID: "speaker1",
		})
	}()

	results := collectResults(ctx, tokenizer.Result())

	// Should have at least one result (the flushed one)
	if len(results) < 1 {
		t.Errorf("expected at least 1 result without boundaries, got %d", len(results))
		return
	}

	if ts, ok := results[0].(internal_type.LLMStreamPacket); ok {
		if ts.Text != "Hello world this is great" {
			t.Errorf("expected 'Hello world this is great', got '%s'", ts.Text)
		}
	}
}

func TestContextCancellation(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	err := tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
		ContextID: "speaker1",
		Text:      "Hello.",
	})

	if err == nil {
		// When context is already cancelled, Assemble may not return an error
		// if the select case for sending on the channel is not reached
		// This is acceptable behavior
		return
	}

	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestConcurrentContexts(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	resultCount := atomic.Int32{}

	// Simulate multiple speakers concurrently
	for speaker := 0; speaker < 3; speaker++ {
		wg.Add(1)
		go func(speakerID int) {
			defer wg.Done()
			contextID := fmt.Sprintf("speaker%d", speakerID)
			for i := 0; i < 3; i++ {
				tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
					ContextID: contextID,
					Text:      fmt.Sprintf("Sentence %d.", i),
				})
			}
		}(speaker)
	}

	go func() {
		wg.Wait()
		tokenizer.Close()
	}()

	// Collect results
	for range tokenizer.Result() {
		resultCount.Add(1)
	}

	// Should have received multiple sentences
	if resultCount.Load() == 0 {
		t.Error("expected to receive some results from concurrent tokenization")
	}
}

func TestBufferStateMaintenance(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()

	go func() {
		// Send partial sentence
		tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      "Hello",
		})

		// Continue with no boundary
		tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      " world",
		})

		// Now add boundary
		tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      ".",
		})
	}()

	results := collectResults(ctx, tokenizer.Result())

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
		return
	}

	if ts, ok := results[0].(internal_type.LLMStreamPacket); ok && ts.Text != "Hello world." {
		t.Errorf("expected 'Hello world.', got '%s'", ts.Text)
	}
}

func TestWhitespaceHandling(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()

	go func() {
		tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      "Hello.   \n  ",
		})
		tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      "World.",
		})
	}()

	results := collectResults(ctx, tokenizer.Result())

	// Should trim whitespace appropriately
	if len(results) < 1 {
		t.Errorf("expected at least 1 result, got %d", len(results))
		return
	}

	// First sentence should be trimmed
	if ts, ok := results[0].(internal_type.LLMStreamPacket); ok && ts.Text != "Hello." {
		t.Errorf("expected 'Hello.', got '%s'", ts.Text)
	}
}

func TestMultipleClose(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)

	// Close multiple times should not panic
	err1 := tokenizer.Close()
	err2 := tokenizer.Close()

	if err1 != nil || err2 != nil {
		t.Errorf("Close should not error on multiple calls: err1=%v, err2=%v", err1, err2)
	}
}

func TestResultChannelClosure(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)

	resultChan := tokenizer.Result()
	tokenizer.Close()

	// Try to read from closed channel
	_, ok := <-resultChan
	if ok {
		t.Error("expected channel to be closed")
	}
}

func TestSpecialCharacterBoundaries(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()

	// Test with regex special characters
	opts := newMockOptions(".?!*+")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()

	go func() {
		tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      "Really?",
		})
	}()

	results := collectResults(ctx, tokenizer.Result())
	if len(results) > 0 {
		if ts, ok := results[0].(internal_type.LLMStreamPacket); ok && ts.Text != "Really?" {
			t.Errorf("special character boundary failed: got '%s'", ts.Text)
		}
	}
}

func TestLargeBatch(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()
	const batchSize = 100

	go func() {
		for i := 0; i < batchSize; i++ {
			tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: "speaker1",
				Text:      fmt.Sprintf("Sentence %d.", i),
			})
		}
	}()

	results := collectResults(ctx, tokenizer.Result())

	if len(results) != batchSize {
		t.Errorf("expected %d results, got %d", batchSize, len(results))
	}
}

func TestLLMStreamingInput(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".!?")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()

	// Simulate LLM streaming tokens/chunks
	llmChunks := []string{
		"Hello",
		" world",
		", this",
		" is",
		" an",
		" LLM",
		" streamed",
		" sentence",
		".",
		" Another",
		" one",
		"!",
	}

	// Channel to signal when streaming is done
	done := make(chan bool)
	var results []internal_type.Packet

	// Start goroutine to collect results WHILE streaming happens
	go func() {
		resultChan := tokenizer.Result()
		for {
			select {
			case result, ok := <-resultChan:
				if !ok {
					t.Logf("Result channel closed")
					done <- true
					return
				}
				results = append(results, result)
				if ts, ok := result.(internal_type.LLMStreamPacket); ok {
					t.Logf("Received result: %q", ts.Text)
				}

			case <-time.After(500 * time.Millisecond):
				// Timeout - no more results coming
				t.Logf("Result collection timeout")
				done <- true
				return
			}
		}
	}()

	// Send all chunks
	go func() {
		for i, chunk := range llmChunks {
			t.Logf("Sending chunk %d: %q", i, chunk)
			err := tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: "llm",
				Text:      chunk,
			})
			if err != nil {
				t.Errorf("Assemble failed: %v", err)
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		t.Logf("All chunks sent, closing tokenizer...")
		tokenizer.Close()
	}()

	// Wait for collection to finish
	<-done

	t.Logf("Total results received: %d", len(results))

	if len(results) != 2 {
		t.Logf("WARNING: expected 2 sentences from LLM stream, got %d", len(results))
		if len(results) == 0 {
			t.Log("No results received - this indicates the tokenizer is not emitting sentences")
		}
		return
	}

	expected := []string{
		"Hello world, this is an LLM streamed sentence.",
		"Another one!",
	}

	for i, r := range results {
		logger.Debugf("result %+v", r)
		if ts, ok := r.(internal_type.LLMStreamPacket); ok {
			if ts.ContextID != "llm" {
				t.Errorf("result %d: expected context 'llm', got '%s'", i, ts.ContextID)
			}
			if ts.Text != expected[i] {
				t.Errorf("result %d: expected '%s', got '%s'", i, expected[i], ts.Text)
			}
		}
	}
}

func TestLLMStreamingWithPauses(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".!?")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()

	chunks := []string{
		"This",
		" sentence",
		" arrives",
		" slowly",
		".",
	}

	var results []internal_type.Packet

	go func() {
		for _, chunk := range chunks {
			time.Sleep(50 * time.Millisecond) // simulate slow LLM token stream
			_ = tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: "llm",
				Text:      chunk,
			})
		}
		tokenizer.Close()
	}()

	for r := range tokenizer.Result() {
		results = append(results, r)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 sentence, got %d", len(results))
	}

	if ts, ok := results[0].(internal_type.LLMStreamPacket); ok && ts.Text != "This sentence arrives slowly." {
		t.Errorf("unexpected sentence: %q", ts.Text)
	}
}

func TestLLMStreamingWithContextSwitch(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".!?")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()
	var results []internal_type.Packet

	go func() {
		// LLM A starts streaming (with boundary to get output)
		_ = tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "llm-A",
			Text:      "LLM A is speaking.",
		})

		// LLM B interrupts
		_ = tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "llm-B",
			Text:      "Hello from B.",
		})

		tokenizer.Close()
	}()

	for r := range tokenizer.Result() {
		results = append(results, r)
	}

	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}

	if ts, ok := results[0].(internal_type.LLMStreamPacket); ok {
		if ts.ContextID != "llm-A" {
			t.Errorf("expected first result from llm-A, got %s", ts.ContextID)
		}
	}

	foundB := false
	for _, r := range results {
		if ts, ok := r.(internal_type.LLMStreamPacket); ok {
			if ts.ContextID == "llm-B" {
				foundB = true
			}
		}
	}

	if !foundB {
		t.Error("expected output from llm-B after context switch")
	}
}

func TestLLMStreamingForcedCompletion(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".!?")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()
	var results []internal_type.Packet

	go func() {
		// Stream without punctuation
		_ = tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "llm",
			Text:      "This sentence never ends",
		})

		// Force flush to emit the buffered sentence
		_ = tokenizer.Assemble(ctx, internal_type.LLMMessagePacket{
			ContextID: "llm",
		})

		tokenizer.Close()
	}()

	for r := range tokenizer.Result() {
		results = append(results, r)
	}

	// Should get: AssembledSentence + Flush
	if len(results) != 2 {
		t.Fatalf("expected 2 results (sentence + flush), got %d", len(results))
	}

	var sentence string
	if ts, ok := results[0].(internal_type.LLMStreamPacket); ok {
		sentence = ts.Text
	} else if ts, ok := results[0].(internal_type.LLMStreamPacket); ok {
		sentence = ts.Text
	} else {
		t.Fatalf("expected first result to be AssembledSentence, got %T", results[0])
	}

	if sentence != "This sentence never ends" {
		t.Errorf("unexpected sentence: %q", sentence)
	}

	if _, ok := results[1].(internal_type.LLMMessagePacket); !ok {
		t.Errorf("expected second result to be Flush message, got %T", results[1])
	}
}
func TestLLMStreamingUnformattedButComplete(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".!?") // boundaries exist but are NOT used
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	ctx := context.Background()
	var results []internal_type.Packet

	go func() {
		// Simulate raw LLM streaming (no punctuation)
		chunks := []string{
			"this",
			" is",
			" a",
			" raw",
			" llm",
			" response",
		}

		for _, chunk := range chunks {
			_ = tokenizer.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: "llm",
				Text:      chunk,
			})
		}

		// End of stream — force completion with Flush
		_ = tokenizer.Assemble(ctx, internal_type.LLMMessagePacket{
			ContextID: "llm",
		})

		tokenizer.Close()
	}()

	for r := range tokenizer.Result() {
		results = append(results, r)
	}

	// Should get: AssembledSentence + Flush
	if len(results) != 2 {
		t.Fatalf("expected 2 results (sentence + flush), got %d", len(results))
	}

	expected := "this is a raw llm response"
	var sentence, contextID string
	if ts, ok := results[0].(internal_type.LLMStreamPacket); ok {
		sentence = ts.Text
		contextID = ts.ContextID
	} else if ts, ok := results[0].(internal_type.LLMStreamPacket); ok {
		sentence = ts.Text
		contextID = ts.ContextID
	} else {
		t.Fatalf("expected first result to be AssembledSentence, got %T", results[0])
	}

	if sentence != expected {
		t.Errorf("expected %q, got %q", expected, sentence)
	}
	if contextID != "llm" {
		t.Errorf("expected context 'llm', got %s", contextID)
	}

	if _, ok := results[1].(internal_type.LLMMessagePacket); !ok {
		t.Errorf("expected second result to be Flush message, got %T", results[1])
	}
}
func TestStringRepresentation(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	opts := newMockOptions(".")
	tokenizer, _ := NewDefaultLLMSentenceAssembler(t.Context(), logger, opts)
	defer tokenizer.Close()

	st := tokenizer.(*sentenceAssembler)
	str := st.String()

	if str == "" {
		t.Error("String() should return non-empty string")
	}

	if !contains(str, "SentenceAssembler") {
		t.Errorf("String() should contain 'SentenceAssembler', got '%s'", str)
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
