// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package elevenlabs_internal

type ElevenlabTextToSpeechResponse struct {
	Audio     string  `json:"audio"`
	ContextId *string `json:"contextId"`
	IsFinal   *bool   `json:"isFinal"`
}
