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

// OutputAudioTransformer is an interface for transforming output audio data.
// It extends the Transformers interface, specifying that it transforms
// from string (processed audio representation) to []byte (raw audio data).
type TextToSpeechTransformer interface {
	Name() string

	//
	Transformers[Packet]
}

// OutputAudioTransformerOptions defines the interface for handling audio output transformation
type TextToSpeechInitializeOptions struct {

	// audio config
	AudioConfig *protos.AudioConfig

	//
	OnSpeech func(pkt ...Packet) error

	// options of model
	ModelOptions utils.Option
}
