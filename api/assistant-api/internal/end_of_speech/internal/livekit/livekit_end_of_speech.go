// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_livekit

import (
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
)

func NewLivekitEndOfSpeech(
	logger commons.Logger, onCallback internal_type.EndOfSpeechCallback, opts utils.Option,
) (internal_type.EndOfSpeech, error) {
	return nil, nil
}
