// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_audio_soxr_resampler

import (
	"encoding/binary"
	"fmt"

	internal_audio "github.com/rapidaai/api/assistant-api/internal/audio"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/protos"
	resampling "github.com/tphakala/go-audio-resampler"
	"github.com/zaf/g711"
)

// libsoxrResampler provides high-quality audio resampling
// using pure-Go libsoxr-equivalent implementation
type libsoxrResampler struct {
	logger commons.Logger
}

// NewLibsoxrAudioResampler creates a new audio resampler
func NewLibsoxrAudioResampler(logger commons.Logger) internal_audio.AudioResampler {
	return &libsoxrResampler{
		logger: logger,
	}
}

// Resample converts audio data using high-quality resampling
func (r *libsoxrResampler) Resample(
	data []byte,
	source, target *protos.AudioConfig,
) ([]byte, error) {

	if source == nil || target == nil {
		return nil, fmt.Errorf("source and target configs are required")
	}

	if len(data) == 0 {
		return []byte{}, nil
	}

	// No-op if identical
	if source.SampleRate == target.SampleRate &&
		source.Channels == target.Channels &&
		source.AudioFormat == target.AudioFormat {
		return data, nil
	}

	// Convert input → LINEAR16
	pcm := data
	if source.AudioFormat != protos.AudioConfig_LINEAR16 {
		var err error
		pcm, err = r.convertToLinear16(data, source)
		if err != nil {
			return nil, err
		}
	}

	// Resample sample rate
	if source.SampleRate != target.SampleRate {
		var err error
		pcm, err = r.resamplePCM16(
			pcm,
			source.SampleRate,
			target.SampleRate,
			int(source.Channels),
		)
		if err != nil {
			return nil, err
		}
	}

	// Convert channels
	if source.Channels != target.Channels {
		pcm = r.convertChannels(pcm, source.Channels, target.Channels)
	}

	// Convert to target format
	if target.AudioFormat != protos.AudioConfig_LINEAR16 {
		var err error
		pcm, err = r.convertFromLinear16(pcm, target)
		if err != nil {
			return nil, err
		}
	}

	return pcm, nil
}

// =======================
// Resampling (Pure Go)
// =======================
func (r *libsoxrResampler) resamplePCM16(
	pcm []byte,
	srcRate, dstRate uint32,
	channels int,
) ([]byte, error) {

	// PCM16 → float64
	floatIn := pcm16ToFloat64(pcm)

	cfg := &resampling.Config{
		InputRate:      float64(srcRate),
		OutputRate:     float64(dstRate),
		Channels:       channels,
		EnableParallel: channels > 1,
		Quality: resampling.QualitySpec{
			Preset: resampling.QualityHigh,
		},
	}

	rs, err := resampling.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("resampler init failed: %w", err)
	}

	out, err := rs.Process(floatIn)
	if err != nil {
		return nil, fmt.Errorf("resample failed: %w", err)
	}

	tail, err := rs.Flush()
	if err != nil {
		return nil, fmt.Errorf("flush failed: %w", err)
	}

	out = append(out, tail...)

	return float64ToPCM16(out), nil
}

func pcm16ToFloat64(data []byte) []float64 {
	out := make([]float64, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		s := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		out[i/2] = float64(s) / 32768.0
	}
	return out
}

func float64ToPCM16(data []float64) []byte {
	out := make([]byte, len(data)*2)
	for i, v := range data {
		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}
		s := int16(v * 32767.0)
		binary.LittleEndian.PutUint16(out[i*2:i*2+2], uint16(s))
	}
	return out
}

// =======================
// Format Conversion
// =======================

func (r *libsoxrResampler) convertToLinear16(
	data []byte,
	cfg *protos.AudioConfig,
) ([]byte, error) {

	switch cfg.AudioFormat {
	case protos.AudioConfig_LINEAR16:
		return data, nil
	case protos.AudioConfig_MuLaw8:
		return r.muLawToLinear16(data), nil
	default:
		return nil, fmt.Errorf("unsupported input format: %v", cfg.AudioFormat)
	}
}

func (r *libsoxrResampler) convertFromLinear16(
	data []byte,
	cfg *protos.AudioConfig,
) ([]byte, error) {

	switch cfg.AudioFormat {
	case protos.AudioConfig_LINEAR16:
		return data, nil
	case protos.AudioConfig_MuLaw8:
		return r.linear16ToMuLaw(data), nil
	default:
		return nil, fmt.Errorf("unsupported output format: %v", cfg.AudioFormat)
	}
}

// =======================
// μ-law
// =======================

// mu-law (8-bit) → linear PCM16 (little-endian)
func (r *libsoxrResampler) muLawToLinear16(data []byte) []byte {
	return g711.DecodeUlaw(data) // returns int16
}

// linear PCM16 (little-endian) → mu-law (8-bit)
func (r *libsoxrResampler) linear16ToMuLaw(data []byte) []byte {
	return g711.EncodeUlaw(data)
}

// =======================
// Channel Conversion
// =======================

func (r *libsoxrResampler) convertChannels(
	data []byte,
	src, dst uint32,
) []byte {

	if src == dst {
		return data
	}

	// Mono → Stereo
	if src == 1 && dst == 2 {
		out := make([]byte, len(data)*2)
		for i := 0; i < len(data); i += 2 {
			copy(out[i*2:], data[i:i+2])
			copy(out[i*2+2:], data[i:i+2])
		}
		return out
	}

	// Stereo → Mono
	if src == 2 && dst == 1 {
		out := make([]byte, len(data)/2)
		for i := 0; i < len(data); i += 4 {
			l := int16(binary.LittleEndian.Uint16(data[i:]))
			r := int16(binary.LittleEndian.Uint16(data[i+2:]))
			m := int16((int32(l) + int32(r)) / 2)
			binary.LittleEndian.PutUint16(out[i/2:], uint16(m))
		}
		return out
	}

	return data
}
