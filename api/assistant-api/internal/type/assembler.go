// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_type

import (
	"context"
)

// LLMSentenceAssembler defines the contract for components that transform
// streamed or batched text inputs into tokenized sentence outputs.
//
// Implementations are expected to:
//   - Accept inputs via Tokenize
//   - Emit results asynchronously on the Result channel
//   - Release resources and stop processing on Close
type LLMSentenceAssembler interface {
	// Tokenize consumes a tokenizer input (such as an LLMStreamChunk
	// or Finalize signal). Implementations should respect context
	// cancellation and deadlines.
	Assemble(ctx context.Context, in ...LLMPacket) error

	// Result returns a read-only channel on which tokenized outputs
	// are delivered.
	Result() <-chan Packet

	// Close terminates the tokenizer, releases resources,
	// and closes the Result channel.
	Close() error
}
