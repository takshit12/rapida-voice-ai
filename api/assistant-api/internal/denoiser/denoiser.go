// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_denoiser

import (
	internal_denoiser_krisp "github.com/rapidaai/api/assistant-api/internal/denoiser/internal/krisp"
	internal_denoiser_rnnoise "github.com/rapidaai/api/assistant-api/internal/denoiser/internal/rn_noise"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type DenoiserIdentifier string

const (
	RN_NOISE DenoiserIdentifier = "rn_noise"
	KRISP    DenoiserIdentifier = "krisp"
)

// logger, audioConfig, opts
func GetDenoiser(aa DenoiserIdentifier, logger commons.Logger, inCfg *protos.AudioConfig, options utils.Option) (internal_type.Denoiser, error) {
	switch aa {
	case KRISP:
		return internal_denoiser_krisp.NewKrispDenoiser(logger, inCfg, options)
	default:
		return internal_denoiser_rnnoise.NewRnnoiseDenoiser(logger, inCfg, options)
	}
}
