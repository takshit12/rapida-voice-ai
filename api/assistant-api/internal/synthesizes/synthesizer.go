// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_synthesizers

import (
	"context"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/utils"
)

// Synthesizer is a generic interface that defines the contract for any type of synthesizer.
// It uses a generic type parameter IN to allow for different input types.
// The interface defines two methods: Synthesize and Flush.
type Synthesizer interface {
	// Synthesize takes a context, a text string, and optional configuration.
	Synthesize(ctx context.Context, in internal_type.TextPacket) internal_type.TextPacket
}

type SynthesizerOptions struct {
	SpeakerOptions utils.Option
}
