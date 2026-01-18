// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

// Package internal_default_sentence_tokenizer provides sentence-level tokenization
// for streaming text with support for multiple concurrent contexts.
//
// It uses boundary detection (configurable delimiters like ".", "?", "!") to
// identify sentence boundaries and emit complete sentences through a channel.
//
// Example usage:
//
//	tokenizer, err := NewSentenceTokenizer(logger, options)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer tokenizer.Close()
//
//	go func() {
//		tokenizer.Tokenize(ctx, Sentence{
//			ContextId: "speaker1",
//			Sentence:  "Hello. ",
//		})
//	}()
//
//	for result := range tokenizer.Result() {
//		fmt.Println(result.Sentence)
//	}
package internal_default

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
)

// SentenceTokenizer implements sentence-level tokenization for streaming text.
type sentenceAssembler struct {
	logger commons.Logger
	ctx    context.Context

	result chan internal_type.Packet

	mu sync.RWMutex

	// buffering state
	buffer         strings.Builder
	currentContext string
	boundaryRegex  *regexp.Regexp
	hasBoundaries  bool
}

// NewSentenceTokenizer creates a new sentence tokenizer with the given logger and options.
//
// The options parameter should contain "speaker.sentence.boundaries" which is a
// comma-separated list of sentence delimiters (e.g., ".,?!").
//
// Returns an error if the boundary regex compilation fails.
//
// Example:
//
//	tokenizer, err := NewSentenceTokenizer(logger, options)
//	if err != nil {
//		return err
//	}
func NewDefaultLLMSentenceAssembler(context context.Context, logger commons.Logger, options utils.Option) (internal_type.LLMSentenceAssembler, error) {
	st := &sentenceAssembler{
		ctx:    context,
		logger: logger,
		result: make(chan internal_type.Packet, 16),
	}
	if err := st.initializeBoundaries(options, logger); err != nil {
		return nil, err
	}
	return st, nil
}

// initializeBoundaries sets up the boundary regex from configuration options.
// It safely handles missing or invalid boundary configurations.
func (st *sentenceAssembler) initializeBoundaries(options utils.Option, logger commons.Logger) error {
	boundariesRaw, err := options.GetString("speaker.sentence.boundaries")
	if err != nil || boundariesRaw == "" {
		return nil
	}

	boundaries := strings.Split(boundariesRaw, commons.SEPARATOR)
	validBoundaries := filterBoundaries(boundaries)

	if len(validBoundaries) == 0 {
		return nil
	}
	st.hasBoundaries = true
	parts := make([]string, 0, len(validBoundaries))
	for _, b := range validBoundaries {
		parts = append(parts, regexp.QuoteMeta(b))
	}

	pattern := fmt.Sprintf(`(%s)\s*`, strings.Join(parts, "|"))
	regex, err := regexp.Compile(pattern)
	if err != nil {
		logger.Errorf("Invalid boundary regex: %v", err)
		st.hasBoundaries = false
		return nil
	}

	st.boundaryRegex = regex
	logger.Debugf("Custom sentence boundaries: %v", validBoundaries)
	return nil
}

// filterBoundaries removes empty and whitespace-only boundaries.
func filterBoundaries(boundaries []string) []string {
	var valid []string
	for _, b := range boundaries {
		if b = strings.TrimSpace(b); b != "" {
			valid = append(valid, b)
		}
	}
	return valid
}

// Tokenize processes a text segment and emits complete sentences.
//
// It handles context switching by flushing any pending text in the previous context.
// When sentence.IsComplete is true, any remaining buffered text is emitted as a complete sentence.
//
// The method respects context cancellation and returns immediately if ctx is cancelled.
//
// Example:
//
//	err := tokenizer.Tokenize(ctx, Sentence{
//		ContextId:  "speaker1",
//		Sentence:   "Hello world.",
//		IsComplete: true,
//	})
func (st *sentenceAssembler) Assemble(ctx context.Context, sentences ...internal_type.LLMPacket) error {
	for _, sentence := range sentences {
		toEmit := st.extractAndQueue(sentence)
		// Send queued results while respecting context cancellation
		for _, s := range toEmit {
			select {
			case st.result <- s:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}

// extractAndQueue extracts complete sentences from the input and returns them
// in emission order. It safely manages buffer state and context switching.

func (st *sentenceAssembler) extractAndQueue(sentence internal_type.LLMPacket) []internal_type.Packet {
	st.mu.Lock()
	defer st.mu.Unlock()
	var toEmit []internal_type.Packet

	switch input := sentence.(type) {
	case internal_type.LLMStreamPacket:
		// Handle context switch - just clean buffer, do NOT emit
		if input.ContextID != st.currentContext && st.currentContext != "" {
			st.buffer.Reset()
		}

		st.currentContext = input.ContextID
		st.buffer.WriteString(input.Text)

		// Extract sentences at boundaries
		if st.hasBoundaries {
			toEmit = append(toEmit, st.extractSentencesByBoundary(input.ContextID)...)
		}
	case internal_type.LLMMessagePacket:
		if remaining := st.getBufferContent(); remaining != "" {
			toEmit = append(toEmit, internal_type.LLMStreamPacket{
				ContextID: st.currentContext,
				Text:      remaining,
			})
			st.buffer.Reset()
		}
		toEmit = append(toEmit, input)
	default:
		st.logger.Warnf("Unsupported tokenizer input type: %T", sentence)
		return nil
	}

	return toEmit
}

// extractSentencesByBoundary extracts all sentences that end at boundaries.
// Called with lock held.
func (st *sentenceAssembler) extractSentencesByBoundary(contextId string) []internal_type.Packet {
	var sentences []internal_type.Packet

	for {
		sentence, remaining := st.extractSentence(st.buffer.String())
		if sentence == "" {
			break
		}
		sentences = append(sentences, internal_type.LLMStreamPacket{ContextID: contextId, Text: sentence})
		st.buffer.Reset()
		st.buffer.WriteString(remaining)
	}

	return sentences
}

// getBufferContent returns the trimmed buffer content.
// Called with lock held.
func (st *sentenceAssembler) getBufferContent() string {
	return strings.TrimSpace(st.buffer.String())
}

// extractSentence extracts a single sentence from text using the boundary regex.
// Returns the sentence and remaining text.
// Called with lock held.
func (st *sentenceAssembler) extractSentence(text string) (string, string) {
	if st.boundaryRegex == nil || text == "" {
		return "", text
	}

	loc := st.boundaryRegex.FindStringIndex(text)
	if loc != nil {
		return strings.TrimSpace(text[:loc[1]]), text[loc[1]:]
	}

	return "", text
}

// Result returns a read-only channel that receives complete sentences.
//
// The channel is closed when Close() is called. Callers should use a range loop
// to safely consume sentences until the channel closes.
//
// Example:
//
//	for sentence := range tokenizer.Result() {
//		fmt.Println(sentence.Sentence)
//	}
func (st *sentenceAssembler) Result() <-chan internal_type.Packet {
	return st.result
}

// Close gracefully shuts down the tokenizer and closes the result channel.
//
// After Close() is called, the tokenizer cannot be reused. Any subsequent calls
// to Tokenize() will panic when trying to send on the closed channel.
//
// It is safe to call Close() multiple times; subsequent calls are no-ops.
func (st *sentenceAssembler) Close() error {
	st.mu.Lock()
	defer st.mu.Unlock()

	// Clear state
	st.buffer.Reset()
	st.currentContext = ""

	// Close result channel
	if st.result != nil {
		close(st.result)
		st.result = nil
	}

	return nil
}

// String returns a string representation of the tokenizer for debugging.
func (st *sentenceAssembler) String() string {
	st.mu.RLock()
	defer st.mu.RUnlock()

	return fmt.Sprintf("SentenceAssembler{context=%q, bufferLen=%d, hasBoundaries=%v}", st.currentContext, st.buffer.Len(), st.hasBoundaries)
}
