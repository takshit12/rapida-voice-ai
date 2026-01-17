// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_type

import (
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

// SpeechToTextTransformer is an interface for transforming input audio data.
// It extends the Transformers interface, specifying that it transforms
// from []byte (raw audio data) to string (processed audio representation).
type SpeechToTextTransformer interface {
	// The `Name() string` method is defining a function signature for a method that returns a string
	// representing the name of the transformer. This method is expected to provide a human-readable
	// identifier for the transformer implementation, allowing users to easily identify and differentiate
	// between different transformer instances based on their names.
	Name() string

	//
	Transformers[UserAudioPacket]
}

// SpeechToTextTransformerOptions defines the interface for handling audio transformation events.
// It provides callbacks for when transcripts are generated and when the transformation process is complete.
//
// The interface includes two methods:
//   - OnTranscript: Called when a new transcript is available. It receives the transcript text
//     and a boolean indicating whether the transcript is complete.
//   - OnComplete: Called when the entire audio transformation process is finished.
//
// Implementations of this interface can be used to handle real-time updates during
// audio processing, allowing for actions such as displaying interim results,
// updating progress indicators, or triggering subsequent processing steps.
type SpeechToTextInitializeOptions struct {

	// audio config
	AudioConfig *protos.AudioConfig

	//
	// on transcript
	OnPacket func(pkt ...Packet) error

	// options of model
	ModelOptions utils.Option
}
