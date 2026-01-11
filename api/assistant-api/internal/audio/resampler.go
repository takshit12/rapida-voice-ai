// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_audio

import (
	"fmt"

	"github.com/rapidaai/protos"
)

// AudioInfo holds information about audio data
type AudioInfo struct {
	SampleRate        uint32
	Format            protos.AudioConfig_AudioFormat
	Channels          uint32
	SamplesPerChannel int
	BytesPerSample    int
	TotalBytes        int
	DurationSeconds   float64
}

type AudioConverter interface {
	ConvertToFloat32Samples(data []byte, config *protos.AudioConfig) ([]float32, error)
	ConvertToByteSamples(samples []float32, config *protos.AudioConfig) ([]byte, error)
}

type AudioResampler interface {
	Resample(data []byte, source, target *protos.AudioConfig) ([]byte, error)
}

// String returns a formatted string representation of AudioInfo
func (info AudioInfo) String() string {
	formatName := "Unknown"
	switch info.Format {
	case protos.AudioConfig_LINEAR16:
		formatName = "Linear16"
	case protos.AudioConfig_MuLaw8:
		formatName = "μ-law 8-bit"
	}

	channelName := "Mono"
	if info.Channels == 2 {
		channelName = "Stereo"
	} else if info.Channels > 2 {
		channelName = fmt.Sprintf("%d channels", info.Channels)
	}

	return fmt.Sprintf("Audio: %s, %d Hz, %s, %.2f seconds (%d samples, %d bytes)",
		formatName, info.SampleRate, channelName, info.DurationSeconds,
		info.SamplesPerChannel, info.TotalBytes)
}
