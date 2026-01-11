// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package deepgram_internal

type DeepgramTextToSpeechResponse struct {
	Type       string  `json:"type"`
	SequenceID float64 `json:"sequence_id,omitempty"`
	Code       string  `json:"code,omitempty"`
	Message    string  `json:"description,omitempty"`
}
