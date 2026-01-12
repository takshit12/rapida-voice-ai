// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_tokenizer

import (
	"context"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
)

// SentenceTokenizer defines the contract for components that transform
// streamed or batched text inputs into tokenized sentence outputs.
//
// Implementations are expected to:
//   - Accept inputs via Tokenize
//   - Emit results asynchronously on the Result channel
//   - Release resources and stop processing on Close
type SentenceTokenizer interface {
	// Tokenize consumes a tokenizer input (such as an LLMStreamChunk
	// or Finalize signal). Implementations should respect context
	// cancellation and deadlines.
	Tokenize(ctx context.Context, in ...internal_type.Packet) error

	// Result returns a read-only channel on which tokenized outputs
	// are delivered.
	Result() <-chan internal_type.Packet

	// Close terminates the tokenizer, releases resources,
	// and closes the Result channel.
	Close() error
}
