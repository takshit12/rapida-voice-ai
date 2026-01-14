// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_type

import (
	"context"
)

type EndOfSpeechCallback func(context.Context, EndOfSpeechPacket) error

type EndOfSpeech interface {
	Name() string
	Analyze(ctx context.Context, s Packet) error
	Close() error
}
