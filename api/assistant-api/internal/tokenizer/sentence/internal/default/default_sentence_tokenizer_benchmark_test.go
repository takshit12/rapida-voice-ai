// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_default

import (
	"context"
	"fmt"
	"testing"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
)

// BenchmarkNewSentenceTokenizer measures the creation time of a tokenizer
func BenchmarkNewSentenceTokenizer(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": ".,?!"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		tokenizer.Close()
	}
}

// BenchmarkNewSentenceTokenizerNoBoundaries measures creation without boundaries
func BenchmarkNewSentenceTokenizerNoBoundaries(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		tokenizer.Close()
	}
}

// BenchmarkSingleSentenceTokenization measures processing a single sentence
func BenchmarkSingleSentenceTokenization(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		tokenizer.Tokenize(ctx, internal_type.TextPacket{
			ContextID: "speaker1",
			Text:      "Hello world.",
		})
		tokenizer.Close()
	}
}

// BenchmarkMultipleSentences measures processing multiple sentences
func BenchmarkMultipleSentences(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}
	ctx := context.Background()

	sentences := []*internal_type.TextPacket{
		{ContextID: "speaker1", Text: "First sentence."},
		{ContextID: "speaker1", Text: " Second sentence."},
		{ContextID: "speaker1", Text: " Third sentence."},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		for _, s := range sentences {
			tokenizer.Tokenize(ctx, s)
		}
		tokenizer.Close()
	}
}

// BenchmarkLargeSentences measures processing large sentences
func BenchmarkLargeSentences(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}
	ctx := context.Background()

	// Create a large sentence
	largeSentence := ""
	for i := 0; i < 1000; i++ {
		largeSentence += "word "
	}
	largeSentence += "."

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		tokenizer.Tokenize(ctx, internal_type.TextPacket{
			ContextID: "speaker1",
			Text:      largeSentence,
		})
		tokenizer.Close()
	}
}

// BenchmarkMultipleBoundaries measures processing with multiple boundaries
func BenchmarkMultipleBoundaries(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": ".,?!;:"}
	ctx := context.Background()

	testSentences := []string{
		"What is this?",
		"I don't know!",
		"Let's try.",
		"Really; absolutely.",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		for _, s := range testSentences {
			tokenizer.Tokenize(ctx, internal_type.TextPacket{
				ContextID: "speaker1",
				Text:      s,
			})
		}
		tokenizer.Close()
	}
}

// BenchmarkContextSwitching measures context switching overhead
func BenchmarkContextSwitching(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		for speaker := 0; speaker < 5; speaker++ {
			for j := 0; j < 3; j++ {
				tokenizer.Tokenize(ctx, internal_type.TextPacket{
					ContextID: fmt.Sprintf("speaker%d", speaker),
					Text:      fmt.Sprintf("Sentence %d.", j),
				})
			}
		}
		tokenizer.Close()
	}
}

// BenchmarkResultChannelConsumption measures the overhead of consuming results
func BenchmarkResultChannelConsumption(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)

		// Send sentences
		for j := 0; j < 10; j++ {
			tokenizer.Tokenize(ctx, internal_type.TextPacket{
				ContextID: "speaker1",
				Text:      fmt.Sprintf("Sentence %d.", j),
			})
		}

		tokenizer.Close()
	}
}

// BenchmarkCompleteFlag measures processing with IsComplete flag
func BenchmarkCompleteFlag(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		tokenizer.Tokenize(ctx, internal_type.TextPacket{
			ContextID: "speaker1",
			Text:      "This is a test",
		}, internal_type.FlushPacket{
			ContextID: "speaker1",
		})

		tokenizer.Close()
	}
}

// BenchmarkBufferingWithoutBoundaries measures buffering with no boundaries
func BenchmarkBufferingWithoutBoundaries(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		for j := 0; j < 5; j++ {
			tokenizer.Tokenize(ctx, internal_type.TextPacket{
				ContextID: "speaker1",
				Text:      "Text segment",
			})
		}
		tokenizer.Close()
	}
}

// BenchmarkStreamingLargeText measures processing streaming text
func BenchmarkStreamingLargeText(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}
	ctx := context.Background()

	// Simulate streaming text chunks
	chunks := []string{
		"The quick brown fox ",
		"jumps over the ",
		"lazy dog.",
		" This is a test.",
		" Another sentence follows.",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		for _, chunk := range chunks {
			tokenizer.Tokenize(ctx, internal_type.TextPacket{
				ContextID: "speaker1",
				Text:      chunk,
			})
		}
		tokenizer.Close()
	}
}

// BenchmarkClosing measures the cost of closing a tokenizer
func BenchmarkClosing(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		tokenizer.Close()
	}
}

// BenchmarkEmptyAndCompleteFlush measures flushing empty buffers
func BenchmarkEmptyAndCompleteFlush(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		// Send empty with complete flag
		tokenizer.Tokenize(ctx, internal_type.TextPacket{
			ContextID: "speaker1",
			Text:      "",
		}, internal_type.FlushPacket{
			ContextID: "speaker1",
		})
		tokenizer.Close()
	}
}

// BenchmarkComplexScenario measures a realistic complex scenario
func BenchmarkComplexScenario(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": ".,?!;:"}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)

		// Simulate a realistic conversation
		conversationTurns := []struct {
			speaker string
			text    string
		}{
			{"alice", "Hello, "},
			{"alice", "how are you today?"},
			{"bob", " I'm doing great!"},
			{"bob", " How about you."},
			{"alice", " Not bad; "},
			{"alice", "just working on code."},
		}

		for _, turn := range conversationTurns {
			tokenizer.Tokenize(ctx, internal_type.TextPacket{
				ContextID: turn.speaker,
				Text:      turn.text,
			})
		}

		// Flush remaining
		tokenizer.Tokenize(ctx, internal_type.TextPacket{
			ContextID: "alice",
			Text:      "",
		}, internal_type.FlushPacket{
			ContextID: "alice",
		})

		tokenizer.Close()
	}
}

// BenchmarkParallelProcessing measures parallel token processing
func BenchmarkParallelProcessing(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tokenizer, _ := NewSentenceTokenizer(logger, opts)
			tokenizer.Tokenize(ctx, internal_type.TextPacket{
				ContextID: "speaker1",
				Text:      "Hello world.",
			})
			tokenizer.Close()
		}
	})
}

// BenchmarkWhitespaceProcessing measures text with various whitespace
func BenchmarkWhitespaceProcessing(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}
	ctx := context.Background()

	textWithWhitespace := "  \n\tHello  \n  world.  \t\n  "

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tokenizer, _ := NewSentenceTokenizer(logger, opts)
		tokenizer.Tokenize(ctx, internal_type.TextPacket{
			ContextID: "speaker1",
			Text:      textWithWhitespace,
		})
		tokenizer.Close()
	}
}
