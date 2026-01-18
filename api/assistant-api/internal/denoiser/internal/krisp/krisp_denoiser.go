// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_denoiser_krisp

import (
	"context"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type krispDenoiser struct {
	logger commons.Logger
}

func NewKrispDenoiser(ctx context.Context, logger commons.Logger, inCfg *protos.AudioConfig, options utils.Option) (internal_type.Denoiser, error) {
	return &krispDenoiser{logger: logger}, nil
}

func (krisp *krispDenoiser) Denoise(ctx context.Context, input []byte) ([]byte, float64, error) {
	panic("not yet implimented")
}

func (krisp *krispDenoiser) Flush() {
	panic("not yet implimented")
}
