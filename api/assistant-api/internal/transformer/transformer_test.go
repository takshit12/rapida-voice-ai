// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer

import (
	"context"
	"testing"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/protos"
	"github.com/stretchr/testify/assert"
)

// TestAudioTransformerString tests the String method
func TestAudioTransformerString(t *testing.T) {
	tests := []struct {
		name     string
		input    AudioTransformer
		expected string
	}{
		{
			name:     "Deepgram",
			input:    DEEPGRAM,
			expected: "deepgram",
		},
		{
			name:     "Google Speech Service",
			input:    GOOGLE_SPEECH_SERVICE,
			expected: "google-speech-service",
		},
		{
			name:     "Azure Speech Service",
			input:    AZURE_SPEECH_SERVICE,
			expected: "azure-speech-service",
		},
		{
			name:     "Cartesia",
			input:    CARTESIA,
			expected: "cartesia",
		},
		{
			name:     "RevAI",
			input:    REVAI,
			expected: "revai",
		},
		{
			name:     "Sarvam",
			input:    SARVAM,
			expected: "sarvamai",
		},
		{
			name:     "ElevenLabs",
			input:    ELEVENLABS,
			expected: "elevenlabs",
		},
		{
			name:     "AssemblyAI",
			input:    ASSEMBLYAI,
			expected: "assemblyai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetTextToSpeechTransformer tests text-to-speech transformer creation
func TestGetTextToSpeechTransformer(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()
	ctx := context.Background()
	credential := &protos.VaultCredential{}
	opts := &internal_type.TextToSpeechInitializeOptions{}

	tests := []struct {
		name            string
		transformerType AudioTransformer
		shouldError     bool
	}{
		{
			name:            "Deepgram TTS",
			transformerType: DEEPGRAM,
			shouldError:     true, // Will fail due to missing credentials, but factory works
		},
		{
			name:            "Invalid TTS",
			transformerType: AudioTransformer("invalid"),
			shouldError:     true, // Should fail with factory error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer, err := GetTextToSpeechTransformer(tt.transformerType, ctx, mockLogger, credential, opts)

			if tt.transformerType == AudioTransformer("invalid") {
				// Invalid transformer type should return factory error
				assert.Error(t, err)
				assert.Nil(t, transformer)
				assert.Equal(t, "illegal text to speech idenitfier", err.Error())
			} else if tt.shouldError {
				// Valid transformer type but credential issues
				assert.Error(t, err) // Expected to fail due to credentials, but not factory error
				assert.Nil(t, transformer)
			}
		})
	}
}

// TestGetSpeechToTextTransformer tests speech-to-text transformer creation
func TestGetSpeechToTextTransformer(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()
	ctx := context.Background()
	credential := &protos.VaultCredential{}
	opts := &internal_type.SpeechToTextInitializeOptions{}

	tests := []struct {
		name            string
		transformerType AudioTransformer
		shouldError     bool
	}{
		{
			name:            "Deepgram STT",
			transformerType: DEEPGRAM,
			shouldError:     true, // Will fail due to missing credentials, but factory works
		},
		{
			name:            "Invalid STT",
			transformerType: AudioTransformer("invalid"),
			shouldError:     true, // Should fail with factory error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer, err := GetSpeechToTextTransformer(tt.transformerType, ctx, mockLogger, credential, opts)

			if tt.transformerType == AudioTransformer("invalid") {
				// Invalid transformer type should return factory error
				assert.Error(t, err)
				assert.Nil(t, transformer)
				assert.Equal(t, "illegal speech to text idenitfier", err.Error())
			} else if tt.shouldError {
				// Valid transformer type but credential issues
				assert.Error(t, err) // Expected to fail due to credentials, but not factory error
				assert.Nil(t, transformer)
			}
		})
	}
}

// TestInvalidAudioTransformerTypesCombinations tests all types of invalid inputs
func TestInvalidAudioTransformerTypesCombinations(t *testing.T) {
	ctx := context.Background()
	mockLogger, _ := commons.NewApplicationLogger()
	credential := &protos.VaultCredential{}
	optsTTS := &internal_type.TextToSpeechInitializeOptions{}
	optsSTT := &internal_type.SpeechToTextInitializeOptions{}

	tests := []struct {
		name       string
		ttsType    AudioTransformer
		sttType    AudioTransformer
		wantTTSErr bool
		wantSTTErr bool
	}{
		{
			name:       "Empty string transformer",
			ttsType:    AudioTransformer(""),
			sttType:    AudioTransformer(""),
			wantTTSErr: true,
			wantSTTErr: true,
		},
		{
			name:       "Unknown transformer",
			ttsType:    AudioTransformer("unknown-provider"),
			sttType:    AudioTransformer("unknown-provider"),
			wantTTSErr: true,
			wantSTTErr: true,
		},
		{
			name:       "Case sensitive test",
			ttsType:    AudioTransformer("DEEPGRAM"),
			sttType:    AudioTransformer("DEEPGRAM"),
			wantTTSErr: true,
			wantSTTErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ttsErr := GetTextToSpeechTransformer(tt.ttsType, ctx, mockLogger, credential, optsTTS)
			if tt.wantTTSErr {
				assert.Error(t, ttsErr)
				assert.Equal(t, "illegal text to speech idenitfier", ttsErr.Error())
			} else {
				assert.NoError(t, ttsErr)
			}

			_, sttErr := GetSpeechToTextTransformer(tt.sttType, ctx, mockLogger, credential, optsSTT)
			if tt.wantSTTErr {
				assert.Error(t, sttErr)
				assert.Equal(t, "illegal speech to text idenitfier", sttErr.Error())
			} else {
				assert.NoError(t, sttErr)
			}
		})
	}
}

// TestInvalidAudioTransformerTypesTTS tests various invalid transformer types for TTS
func TestInvalidAudioTransformerTypesTTS(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	ctx := context.Background()
	credential := &protos.VaultCredential{}
	opts := &internal_type.TextToSpeechInitializeOptions{}

	invalidTypes := []string{
		"",
		"invalid",
		"DEEPGRAM",
		"deepgram-extra",
		"unknown-service",
	}

	for _, invalidType := range invalidTypes {
		t.Run("Invalid_"+invalidType, func(t *testing.T) {
			transformer, err := GetTextToSpeechTransformer(AudioTransformer(invalidType), ctx, mockLogger, credential, opts)
			assert.Error(t, err)
			assert.Nil(t, transformer)
			assert.Equal(t, "illegal text to speech idenitfier", err.Error())
		})
	}
}

// TestInvalidAudioTransformerTypesSTT tests various invalid transformer types for STT
func TestInvalidAudioTransformerTypesSTT(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	ctx := context.Background()
	credential := &protos.VaultCredential{}
	opts := &internal_type.SpeechToTextInitializeOptions{}

	invalidTypes := []string{
		"",
		"invalid",
		"DEEPGRAM",
		"deepgram-extra",
		"unknown-service",
	}

	for _, invalidType := range invalidTypes {
		t.Run("Invalid_"+invalidType, func(t *testing.T) {
			transformer, err := GetSpeechToTextTransformer(AudioTransformer(invalidType), ctx, mockLogger, credential, opts)
			assert.Error(t, err)
			assert.Nil(t, transformer)
			assert.Equal(t, "illegal speech to text idenitfier", err.Error())
		})
	}
}

// TestAllTextToSpeechTransformersAreDifferent validates factory doesn't panic for all types
func TestAllTextToSpeechTransformersCallFactory(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	ctx := context.Background()
	credential := &protos.VaultCredential{}
	opts := &internal_type.TextToSpeechInitializeOptions{}

	transformerTypes := []AudioTransformer{
		DEEPGRAM,
		AZURE_SPEECH_SERVICE,
		CARTESIA,
		GOOGLE_SPEECH_SERVICE,
		REVAI,
		SARVAM,
		ELEVENLABS,
	}

	for _, tt := range transformerTypes {
		t.Run(tt.String(), func(t *testing.T) {
			// Just ensure factory can be called without panic
			_, _ = GetTextToSpeechTransformer(tt, ctx, mockLogger, credential, opts)
		})
	}
}

// TestAllSpeechToTextTransformersCallFactory validates factory doesn't panic for all types
func TestAllSpeechToTextTransformersCallFactory(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	ctx := context.Background()
	credential := &protos.VaultCredential{}
	opts := &internal_type.SpeechToTextInitializeOptions{}

	transformerTypes := []AudioTransformer{
		DEEPGRAM,
		AZURE_SPEECH_SERVICE,
		GOOGLE_SPEECH_SERVICE,
		ASSEMBLYAI,
		REVAI,
		SARVAM,
		CARTESIA,
	}

	for _, tt := range transformerTypes {
		t.Run(tt.String(), func(t *testing.T) {
			// Just ensure factory can be called without panic
			_, _ = GetSpeechToTextTransformer(tt, ctx, mockLogger, credential, opts)
		})
	}
}

// BenchmarkGetTextToSpeechTransformer benchmarks TTS factory performance
func BenchmarkGetTextToSpeechTransformer(b *testing.B) {
	mockLogger, _ := commons.NewApplicationLogger()

	ctx := context.Background()
	credential := &protos.VaultCredential{}
	opts := &internal_type.TextToSpeechInitializeOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetTextToSpeechTransformer(DEEPGRAM, ctx, mockLogger, credential, opts)
	}
}

// BenchmarkGetSpeechToTextTransformer benchmarks STT factory performance
func BenchmarkGetSpeechToTextTransformer(b *testing.B) {
	mockLogger, _ := commons.NewApplicationLogger()

	ctx := context.Background()
	credential := &protos.VaultCredential{}
	opts := &internal_type.SpeechToTextInitializeOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetSpeechToTextTransformer(DEEPGRAM, ctx, mockLogger, credential, opts)
	}
}
