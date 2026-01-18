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

// BenchmarkNewDefaultLLMSentenceAssembler measures the creation time of a assembler
func BenchmarkNewDefaultLLMSentenceAssembler(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": ".,?!"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		assembler.Close()
	}
}

// BenchmarkNewDefaultLLMSentenceAssemblerNoBoundaries measures creation without boundaries
func BenchmarkNewDefaultLLMSentenceAssemblerNoBoundaries(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		assembler.Close()
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
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		assembler.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      "Hello world.",
		})
		assembler.Close()
	}
}

// BenchmarkMultipleSentences measures processing multiple sentences
func BenchmarkMultipleSentences(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}
	ctx := context.Background()

	sentences := []*internal_type.LLMStreamPacket{
		{ContextID: "speaker1", Text: "First sentence."},
		{ContextID: "speaker1", Text: " Second sentence."},
		{ContextID: "speaker1", Text: " Third sentence."},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		for _, s := range sentences {
			assembler.Assemble(ctx, s)
		}
		assembler.Close()
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
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		assembler.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      largeSentence,
		})
		assembler.Close()
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
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		for _, s := range testSentences {
			assembler.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: "speaker1",
				Text:      s,
			})
		}
		assembler.Close()
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
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		for speaker := 0; speaker < 5; speaker++ {
			for j := 0; j < 3; j++ {
				assembler.Assemble(ctx, internal_type.LLMStreamPacket{
					ContextID: fmt.Sprintf("speaker%d", speaker),
					Text:      fmt.Sprintf("Sentence %d.", j),
				})
			}
		}
		assembler.Close()
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
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)

		// Send sentences
		for j := 0; j < 10; j++ {
			assembler.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: "speaker1",
				Text:      fmt.Sprintf("Sentence %d.", j),
			})
		}

		assembler.Close()
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
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		assembler.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      "This is a test",
		}, internal_type.LLMMessagePacket{
			ContextID: "speaker1",
		})

		assembler.Close()
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
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		for j := 0; j < 5; j++ {
			assembler.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: "speaker1",
				Text:      "Text segment",
			})
		}
		assembler.Close()
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
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		for _, chunk := range chunks {
			assembler.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: "speaker1",
				Text:      chunk,
			})
		}
		assembler.Close()
	}
}

// BenchmarkClosing measures the cost of closing a assembler
func BenchmarkClosing(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		assembler.Close()
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
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		// Send empty with complete flag
		assembler.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      "",
		}, internal_type.LLMMessagePacket{
			ContextID: "speaker1",
		})
		assembler.Close()
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
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)

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
			assembler.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: turn.speaker,
				Text:      turn.text,
			})
		}

		// Flush remaining
		assembler.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "alice",
			Text:      "",
		}, internal_type.LLMMessagePacket{
			ContextID: "alice",
		})

		assembler.Close()
	}
}

// BenchmarkParallelProcessing measures parallel token processing
func BenchmarkParallelProcessing(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	opts := utils.Option{"speaker.sentence.boundaries": "."}
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
			assembler.Assemble(ctx, internal_type.LLMStreamPacket{
				ContextID: "speaker1",
				Text:      "Hello world.",
			})
			assembler.Close()
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
		assembler, _ := NewDefaultLLMSentenceAssembler(b.Context(), logger, opts)
		assembler.Assemble(ctx, internal_type.LLMStreamPacket{
			ContextID: "speaker1",
			Text:      textWithWhitespace,
		})
		assembler.Close()
	}
}
