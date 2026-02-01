// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_smallest

import (
	"fmt"

	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type smallestOption struct {
	key         string
	logger      commons.Logger
	mdlOpts     utils.Option
	audioConfig *protos.AudioConfig
}

func NewSmallestOption(logger commons.Logger, vaultCredential *protos.VaultCredential,
	audioConfig *protos.AudioConfig,
	opts utils.Option) (*smallestOption, error) {
	cx, ok := vaultCredential.GetValue().AsMap()["key"]
	if !ok {
		return nil, fmt.Errorf("smallest: illegal vault config")
	}
	return &smallestOption{
		key:         cx.(string),
		audioConfig: audioConfig,
		mdlOpts:     opts,
		logger:      logger,
	}, nil
}

func (co *smallestOption) GetKey() string {
	return co.key
}

func (co *smallestOption) GetOutputFormat() string {
	switch co.audioConfig.GetAudioFormat() {
	case protos.AudioConfig_MuLaw8:
		return "mulaw"
	case protos.AudioConfig_LINEAR16:
		return "pcm"
	default:
		return "pcm"
	}
}

func (co *smallestOption) GetSampleRate() uint32 {
	return co.audioConfig.GetSampleRate()
}

func (co *smallestOption) GetTextToSpeechConnectionString() string {
	model := "lightning-v2"
	if modelValue, err := co.mdlOpts.GetString("speak.model"); err == nil {
		model = modelValue
	}

	return fmt.Sprintf("wss://waves-api.smallest.ai/api/v1/%s/get_speech/stream?timeout=30", model)
}
