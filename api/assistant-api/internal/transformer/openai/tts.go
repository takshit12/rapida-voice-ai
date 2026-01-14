// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_openai

import (
	"context"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
)

func NewOpenaiTextToSpeech(
	ctx context.Context,
	logger commons.Logger,
	onSpeech func([]byte) error) (internal_type.TextToSpeechTransformer, error) {
	return nil, nil
}
