// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_denoiser_rnnoise

import (
	"context"

	internal_audio_resampler "github.com/rapidaai/api/assistant-api/internal/audio/resampler"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type rnnoiseDenoiser struct {
	rnNoise        *RNNoise
	logger         commons.Logger
	denoiserConfig *protos.AudioConfig
	inputConfig    *protos.AudioConfig
	audioSampler   internal_type.AudioResampler
	audioConverter internal_type.AudioConverter
}

// NewDenoiser creates a new denoiser instance
func NewRnnoiseDenoiser(
	ctx context.Context,
	logger commons.Logger, inputConfig *protos.AudioConfig, options utils.Option,
) (internal_type.Denoiser, error) {
	rn, err := NewRNNoise()
	if err != nil {
		return nil, err
	}
	sampler, err := internal_audio_resampler.GetResampler(logger)
	if err != nil {
		return nil, err
	}
	converter, err := internal_audio_resampler.GetConverter(logger)
	if err != nil {
		return nil, err
	}

	return &rnnoiseDenoiser{
		audioSampler:   sampler,
		audioConverter: converter,
		rnNoise:        rn,
		denoiserConfig: &protos.AudioConfig{
			SampleRate:  48000,
			AudioFormat: protos.AudioConfig_LINEAR16,
		},
		inputConfig: inputConfig,
		logger:      logger}, nil
}

// ProcessStream processes a continuous audio stream
func (rnd *rnnoiseDenoiser) Denoise(ctx context.Context, input []byte) ([]byte, float64, error) {
	idi, err := rnd.audioSampler.Resample(input, rnd.inputConfig, rnd.denoiserConfig)
	if err != nil {
		return nil, 0, err
	}

	floatSample, err := rnd.audioConverter.ConvertToFloat32Samples(idi, rnd.denoiserConfig)
	if err != nil {
		return nil, 0, err
	}

	var combinedCleanedAudio []float32
	var combinedCnf float64

	for i := 0; i < len(floatSample); i += 480 {
		end := i + 480
		if end > len(floatSample) {
			end = len(floatSample)
		}

		// Extract chunk and pad if it is less than 480 samples
		chunk := floatSample[i:end]
		if len(chunk) < 480 {
			padding := make([]float32, 480-len(chunk))
			chunk = append(chunk, padding...)
		}

		cnf, cleanedAudio, err := rnd.rnNoise.SuppressNoise(chunk)
		if err != nil {
			return nil, 0, err
		}

		// Append results
		combinedCleanedAudio = append(combinedCleanedAudio, cleanedAudio...)
		combinedCnf += cnf
	}

	// Average the confidence scores
	if len(combinedCleanedAudio) > 0 {
		combinedCnf /= float64((len(floatSample)-1)/480 + 1)
	}

	ido, err := rnd.audioConverter.ConvertToByteSamples(combinedCleanedAudio, rnd.denoiserConfig)
	if err != nil {
		return nil, 0, err
	}

	idm, err := rnd.audioSampler.Resample(ido, rnd.denoiserConfig, rnd.inputConfig)
	if err != nil {
		return nil, 0, err
	}
	return idm, combinedCnf, err
}

// Close releases resources
func (d *rnnoiseDenoiser) Flush() {
	if d.rnNoise != nil {
		d.rnNoise.Close()
	}
}
