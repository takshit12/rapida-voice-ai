// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_vad

import (
	"context"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	internal_vad_silero "github.com/rapidaai/api/assistant-api/internal/vad/internal/silero_vad"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type VADIdentifier string

const (
	SILERO_VAD            VADIdentifier = "silero_vad"
	TEN_VAD               VADIdentifier = "ten_vad"
	OptionsKeyVadProvider               = "microphone.vad.provider"
)

// logger, audioConfig, opts
func GetVAD(ctx context.Context, logger commons.Logger, intputAudio *protos.AudioConfig, callback internal_type.VADCallback, options utils.Option) (internal_type.Vad, error) {
	typ, _ := options.GetString(OptionsKeyVadProvider)
	switch VADIdentifier(typ) {
	case SILERO_VAD:
		return internal_vad_silero.NewSileroVAD(ctx, logger, intputAudio, callback, options)
	default:
		return internal_vad_silero.NewSileroVAD(ctx, logger, intputAudio, callback, options)
	}
}
