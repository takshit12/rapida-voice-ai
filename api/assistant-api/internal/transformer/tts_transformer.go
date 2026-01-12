// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer

import (
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

// OutputAudioTransformer is an interface for transforming output audio data.
// It extends the Transformers interface, specifying that it transforms
// from string (processed audio representation) to []byte (raw audio data).
type TextToSpeechTransformer interface {
	Name() string

	//
	Transformers[internal_type.Packet]
}

// OutputAudioTransformerOptions defines the interface for handling audio output transformation
type TextToSpeechInitializeOptions struct {

	// audio config
	AudioConfig *protos.AudioConfig
	// OnSpeech is called when speech is detected in the audio stream
	// It receives a byte slice containing the speech audio data
	// Returns an error if there's an issue processing the speech
	OnSpeech func(string, []byte) error

	// OnComplete is called when the audio transformation is complete
	// Returns an error if there's an issue finalizing the transformation
	OnComplete func(string) error

	// options of model
	ModelOptions utils.Option
}
