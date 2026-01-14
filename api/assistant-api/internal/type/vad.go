// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_type

// Activity represents a detected Audio segment.

type VADCallback func(InterruptionPacket) error

type Vad interface {
	Name() string
	Process(frame []byte) error
	Close() error
}
