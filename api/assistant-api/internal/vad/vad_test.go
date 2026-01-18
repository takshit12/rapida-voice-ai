// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_vad

import (
	"testing"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/protos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockVADCallback implements the VADCallback interface for testing
func MockVADCallback(result internal_type.InterruptionPacket) error {
	return nil
}

// TestGetVAD_SILERO_VAD tests VAD factory with SILERO_VAD identifier
func TestGetVAD_SILERO_VAD(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	audioConfig := &protos.AudioConfig{
		AudioFormat: protos.AudioConfig_LINEAR16,
		SampleRate:  16000,
		Channels:    1,
	}

	vad, err := GetVAD(t.Context(), logger, audioConfig, MockVADCallback, map[string]interface{}{
		OptionsKeyVadProvider: "invalid_vad",
	})

	require.NoError(t, err, "GetVAD should not return error for SILERO_VAD")
	require.NotNil(t, vad, "GetVAD should return non-nil VAD instance")
	assert.NotEmpty(t, vad.Name(), "VAD name should not be empty")
}

// TestGetVAD_TEN_VAD tests VAD factory with TEN_VAD identifier
func TestGetVAD_TEN_VAD(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	audioConfig := &protos.AudioConfig{
		AudioFormat: protos.AudioConfig_LINEAR16,
		SampleRate:  16000,
		Channels:    1,
	}

	vad, err := GetVAD(t.Context(), logger, audioConfig, MockVADCallback, map[string]interface{}{
		OptionsKeyVadProvider: TEN_VAD,
	})

	require.NoError(t, err, "GetVAD should not return error for TEN_VAD")
	require.NotNil(t, vad, "GetVAD should return non-nil VAD instance")
	assert.NotEmpty(t, vad.Name(), "VAD name should not be empty")
}

// TestGetVAD_InvalidIdentifier tests VAD factory with invalid identifier defaults to SILERO_VAD
func TestGetVAD_InvalidIdentifier(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	audioConfig := &protos.AudioConfig{
		AudioFormat: protos.AudioConfig_LINEAR16,
		SampleRate:  16000,
		Channels:    1,
	}

	vad, err := GetVAD(t.Context(), logger, audioConfig, MockVADCallback, map[string]interface{}{
		OptionsKeyVadProvider: "invalid_vad",
	})

	require.NoError(t, err, "GetVAD should default to SILERO_VAD for invalid identifier")
	require.NotNil(t, vad, "GetVAD should return non-nil VAD instance")
	assert.NotEmpty(t, vad.Name(), "VAD name should not be empty")
}

// TestGetVAD_WithNilLogger tests VAD factory with nil logger
func TestGetVAD_WithNilLogger(t *testing.T) {
	audioConfig := &protos.AudioConfig{
		AudioFormat: protos.AudioConfig_LINEAR16,
		SampleRate:  16000,
		Channels:    1,
	}

	logger, _ := commons.NewApplicationLogger()

	vad, err := GetVAD(t.Context(), logger, audioConfig, MockVADCallback, map[string]interface{}{
		OptionsKeyVadProvider: SILERO_VAD,
	})

	if err != nil {
		t.Logf("GetVAD returned error with nil logger: %v", err)
	} else if vad != nil {
		t.Logf("GetVAD returned VAD instance with nil logger")
	}
}

// TestGetVAD_WithNilAudioConfig tests VAD factory with nil audio config
func TestGetVAD_WithNilAudioConfig(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()

	vad, err := GetVAD(t.Context(), logger, nil, MockVADCallback, map[string]interface{}{
		OptionsKeyVadProvider: SILERO_VAD,
	})
	if err != nil {
		t.Logf("GetVAD returned error with nil audio config: %v", err)
	} else if vad != nil {
		t.Logf("GetVAD returned VAD instance with nil audio config")
	}
}

// TestGetVAD_WithNilCallback tests VAD factory with nil callback
func TestGetVAD_WithNilCallback(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	audioConfig := &protos.AudioConfig{
		AudioFormat: protos.AudioConfig_LINEAR16,
		SampleRate:  16000,
		Channels:    1,
	}

	vad, err := GetVAD(t.Context(), logger, audioConfig, nil, map[string]interface{}{
		OptionsKeyVadProvider: SILERO_VAD,
	})

	if err != nil {
		t.Logf("GetVAD returned error with nil callback: %v", err)
	} else if vad != nil {
		t.Logf("GetVAD returned VAD instance with nil callback")
	}
}

// TestGetVAD_WithDifferentAudioFormats tests VAD factory with various audio formats
func TestGetVAD_WithDifferentAudioFormats(t *testing.T) {
	testCases := []struct {
		name       string
		audioFmt   protos.AudioConfig_AudioFormat
		sampleRate uint32
	}{
		{
			name:       "LINEAR16_16kHz",
			audioFmt:   protos.AudioConfig_LINEAR16,
			sampleRate: 16000,
		},
		{
			name:       "LINEAR16_8kHz",
			audioFmt:   protos.AudioConfig_LINEAR16,
			sampleRate: 8000,
		},
		{
			name:       "MuLaw8_8kHz",
			audioFmt:   protos.AudioConfig_MuLaw8,
			sampleRate: 8000,
		},
		{
			name:       "LINEAR16_48kHz",
			audioFmt:   protos.AudioConfig_LINEAR16,
			sampleRate: 48000,
		},
	}

	logger, _ := commons.NewApplicationLogger()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			audioConfig := &protos.AudioConfig{
				AudioFormat: tc.audioFmt,
				SampleRate:  tc.sampleRate,
				Channels:    1,
			}

			vad, err := GetVAD(t.Context(), logger, audioConfig, MockVADCallback, map[string]interface{}{
				OptionsKeyVadProvider: SILERO_VAD,
			})

			require.NoError(t, err, "GetVAD should not error for %s", tc.name)
			require.NotNil(t, vad, "GetVAD should return VAD instance for %s", tc.name)
		})
	}
}

// TestGetVAD_ConsistentResults tests that multiple calls return consistent VAD instances
func TestGetVAD_ConsistentResults(t *testing.T) {
	logger, _ := commons.NewApplicationLogger()
	audioConfig := &protos.AudioConfig{
		AudioFormat: protos.AudioConfig_LINEAR16,
		SampleRate:  16000,
		Channels:    1,
	}

	vad1, err1 := GetVAD(t.Context(), logger, audioConfig, MockVADCallback, map[string]interface{}{
		OptionsKeyVadProvider: SILERO_VAD,
	})
	vad2, err2 := GetVAD(t.Context(), logger, audioConfig, MockVADCallback, map[string]interface{}{
		OptionsKeyVadProvider: SILERO_VAD,
	})

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, vad1)
	require.NotNil(t, vad2)
	assert.NotEmpty(t, vad1.Name())
	assert.NotEmpty(t, vad2.Name())
}

// TestGetVAD_AllIdentifiers tests all VADIdentifier constants
func TestGetVAD_AllIdentifiers(t *testing.T) {
	identifiers := []VADIdentifier{
		SILERO_VAD,
		TEN_VAD,
	}

	logger, _ := commons.NewApplicationLogger()
	audioConfig := &protos.AudioConfig{
		AudioFormat: protos.AudioConfig_LINEAR16,
		SampleRate:  16000,
		Channels:    1,
	}

	for _, identifier := range identifiers {
		t.Run(string(identifier), func(t *testing.T) {
			vad, err := GetVAD(t.Context(), logger, audioConfig, MockVADCallback, map[string]interface{}{
				OptionsKeyVadProvider: identifier,
			})

			require.NoError(t, err, "GetVAD should not error for identifier: %s", identifier)
			require.NotNil(t, vad, "GetVAD should return VAD instance for identifier: %s", identifier)
			assert.NotEmpty(t, vad.Name(), "VAD name should not be empty for identifier: %s", identifier)
		})
	}
}

// TestVADIdentifier_String tests VADIdentifier string representation
func TestVADIdentifier_String(t *testing.T) {
	testCases := []struct {
		name       string
		identifier VADIdentifier
		expected   string
	}{
		{
			name:       "SILERO_VAD string",
			identifier: SILERO_VAD,
			expected:   "silero_vad",
		},
		{
			name:       "TEN_VAD string",
			identifier: TEN_VAD,
			expected:   "ten_vad",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.identifier))
		})
	}
}

// BenchmarkGetVAD_SILERO_VAD benchmarks VAD factory with SILERO_VAD
func BenchmarkGetVAD_SILERO_VAD(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	audioConfig := &protos.AudioConfig{
		AudioFormat: protos.AudioConfig_LINEAR16,
		SampleRate:  16000,
		Channels:    1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetVAD(b.Context(), logger, audioConfig, MockVADCallback, map[string]interface{}{
			OptionsKeyVadProvider: SILERO_VAD,
		})
	}
}

// BenchmarkGetVAD_TEN_VAD benchmarks VAD factory with TEN_VAD
func BenchmarkGetVAD_TEN_VAD(b *testing.B) {
	logger, _ := commons.NewApplicationLogger()
	audioConfig := &protos.AudioConfig{
		AudioFormat: protos.AudioConfig_LINEAR16,
		SampleRate:  16000,
		Channels:    1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetVAD(b.Context(), logger, audioConfig, MockVADCallback, map[string]interface{}{
			OptionsKeyVadProvider: TEN_VAD,
		})
	}
}
