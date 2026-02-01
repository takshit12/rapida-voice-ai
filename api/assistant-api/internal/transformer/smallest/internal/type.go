// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package smallest_internal

type SmallestTTSResponse struct {
	RequestID string              `json:"request_id"`
	Status    string              `json:"status"` // "chunk" | "complete" | "error"
	Data      *SmallestTTSAudioData `json:"data"`
	Message   string              `json:"message"`
	Done      bool                `json:"done"`
}

type SmallestTTSAudioData struct {
	Audio string `json:"audio"` // base64 encoded audio
}
