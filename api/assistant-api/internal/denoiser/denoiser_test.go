// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_denoiser

import (
	"testing"

	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
	"github.com/stretchr/testify/assert"
)

// TestDenoiserIdentifierString tests the String method for DenoiserIdentifier
func TestDenoiserIdentifierString(t *testing.T) {
	tests := []struct {
		name       string
		identifier DenoiserIdentifier
		expected   string
	}{
		{name: "RN_NOISE", identifier: RN_NOISE, expected: "rn_noise"},
		{name: "KRISP", identifier: KRISP, expected: "krisp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.identifier))
		})
	}
}

// TestGetDenoiserWithValidTypes tests factory with valid denoiser types
func TestGetDenoiserWithValidTypes(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	config := &protos.AudioConfig{SampleRate: 16000}

	tests := []struct {
		name       string
		identifier DenoiserIdentifier
	}{
		{name: "KRISP", identifier: KRISP},
		{name: "RN_NOISE", identifier: RN_NOISE},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			opts := utils.Option{DenoiserOptionsKeyProvider: tt.identifier}

			denoiser, err := GetDenoiser(t.Context(), mockLogger, config, opts)
			assert.NoError(t, err)
			assert.NotNil(t, denoiser)
		})
	}
}

// TestGetDenoiserWithInvalidIdentifiers tests factory with invalid identifiers
func TestGetDenoiserWithInvalidIdentifiers(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()
	config := &protos.AudioConfig{SampleRate: 16000}
	tests := []struct {
		name       string
		identifier DenoiserIdentifier
	}{
		{name: "Empty - defaults to RN_NOISE", identifier: DenoiserIdentifier("")},
		{name: "Unknown - defaults to RN_NOISE", identifier: DenoiserIdentifier("unknown")},
		{name: "Typo - defaults to RN_NOISE", identifier: DenoiserIdentifier("kris")},
		{name: "Case sensitive - defaults to RN_NOISE", identifier: DenoiserIdentifier("KRISP")},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			opts := utils.Option{DenoiserOptionsKeyProvider: tt.identifier}

			denoiser, err := GetDenoiser(t.Context(), mockLogger, config, opts)
			assert.NoError(t, err)
			assert.NotNil(t, denoiser)
		})
	}
}

// TestGetDenoiserWithNilLogger tests with nil logger
func TestGetDenoiserWithNilLogger(t *testing.T) {
	config := &protos.AudioConfig{SampleRate: 16000}

	tests := []struct {
		name       string
		identifier DenoiserIdentifier
	}{
		{name: "KRISP with nil logger", identifier: KRISP},
		{name: "RN_NOISE with nil logger", identifier: RN_NOISE},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			opts := utils.Option{DenoiserOptionsKeyProvider: tt.identifier}

			denoiser, _ := GetDenoiser(t.Context(), nil, config, opts)
			assert.NotNil(t, denoiser)
		})
	}
}

// TestGetDenoiserWithNilAudioConfig tests with nil config
func TestGetDenoiserWithNilAudioConfig(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	tests := []struct {
		name       string
		identifier DenoiserIdentifier
	}{
		{name: "KRISP with nil config", identifier: KRISP},
		{name: "RN_NOISE with nil config", identifier: RN_NOISE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := utils.Option{DenoiserOptionsKeyProvider: tt.identifier}

			denoiser, _ := GetDenoiser(t.Context(), mockLogger, nil, opts)
			assert.NotNil(t, denoiser)
		})
	}
}

// TestAllDenoiserTypesCallFactory tests all types work
func TestAllDenoiserTypesCallFactory(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	config := &protos.AudioConfig{SampleRate: 16000}

	tests := []struct {
		name       string
		identifier DenoiserIdentifier
	}{
		{name: "RN_NOISE", identifier: RN_NOISE},
		{name: "KRISP", identifier: KRISP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := utils.Option{DenoiserOptionsKeyProvider: tt.identifier}
			denoiser, err := GetDenoiser(t.Context(), mockLogger, config, opts)
			assert.NoError(t, err)
			assert.NotNil(t, denoiser)
		})
	}
}

// TestDenoiserIdentifierStringConsistency validates consistency
func TestDenoiserIdentifierStringConsistency(t *testing.T) {
	tests := []struct {
		name          string
		identifier    DenoiserIdentifier
		expectedValue string
	}{
		{name: "RN_NOISE", identifier: RN_NOISE, expectedValue: "rn_noise"},
		{name: "KRISP", identifier: KRISP, expectedValue: "krisp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedValue, string(tt.identifier))
		})
	}
}

// TestDenoiserTypeConversion tests type conversion
func TestDenoiserTypeConversion(t *testing.T) {
	tests := []struct {
		name        string
		stringValue string
		expected    DenoiserIdentifier
		shouldMatch bool
	}{
		{name: "RN_NOISE", stringValue: "rn_noise", expected: RN_NOISE, shouldMatch: true},
		{name: "KRISP", stringValue: "krisp", expected: KRISP, shouldMatch: true},
		{name: "Case mismatch", stringValue: "Krisp", expected: KRISP, shouldMatch: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converted := DenoiserIdentifier(tt.stringValue)
			if tt.shouldMatch {
				assert.Equal(t, tt.expected, converted)
			} else {
				assert.NotEqual(t, tt.expected, converted)
			}
		})
	}
}

// TestDenoiserFactoryDefaults tests default behavior
func TestDenoiserFactoryDefaults(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	config := &protos.AudioConfig{SampleRate: 16000}

	tests := []struct {
		name       string
		identifier DenoiserIdentifier
	}{
		{name: "Empty defaults to RN_NOISE", identifier: DenoiserIdentifier("")},
		{name: "Typo defaults to RN_NOISE", identifier: DenoiserIdentifier("krisp_typo")},
		{name: "Random defaults to RN_NOISE", identifier: DenoiserIdentifier("random_x")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := utils.Option{DenoiserOptionsKeyProvider: tt.identifier}
			denoiser, err := GetDenoiser(t.Context(), mockLogger, config, opts)
			assert.NoError(t, err)
			assert.NotNil(t, denoiser)
		})
	}
}

// TestValidDenoiserIdentifierMatching validates constants are distinct
func TestValidDenoiserIdentifierMatching(t *testing.T) {
	constants := []DenoiserIdentifier{RN_NOISE, KRISP}
	for i := 0; i < len(constants); i++ {
		for j := i + 1; j < len(constants); j++ {
			assert.NotEqual(t, constants[i], constants[j])
		}
	}
}

// TestDenoiserFactoryWithDifferentConfigs tests various configs
func TestDenoiserFactoryWithDifferentConfigs(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	tests := []struct {
		name       string
		identifier DenoiserIdentifier
		sampleRate uint32
	}{
		{name: "KRISP 8kHz", identifier: KRISP, sampleRate: 8000},
		{name: "KRISP 16kHz", identifier: KRISP, sampleRate: 16000},
		{name: "RN_NOISE 8kHz", identifier: RN_NOISE, sampleRate: 8000},
		{name: "RN_NOISE 16kHz", identifier: RN_NOISE, sampleRate: 16000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &protos.AudioConfig{SampleRate: tt.sampleRate}
			opts := utils.Option{DenoiserOptionsKeyProvider: tt.identifier}

			denoiser, err := GetDenoiser(t.Context(), mockLogger, config, opts)
			assert.NoError(t, err)
			assert.NotNil(t, denoiser)
		})
	}
}

// BenchmarkGetDenoiserKRISP benchmarks KRISP factory
func BenchmarkGetDenoiserKRISP(b *testing.B) {

	config := &protos.AudioConfig{SampleRate: 16000}
	opts := utils.Option{DenoiserOptionsKeyProvider: KRISP}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetDenoiser(b.Context(), nil, config, opts)
	}
}

// BenchmarkGetDenoiserRNNoise benchmarks RN_NOISE factory
func BenchmarkGetDenoiserRNNoise(b *testing.B) {

	config := &protos.AudioConfig{SampleRate: 16000}
	opts := utils.Option{DenoiserOptionsKeyProvider: RN_NOISE}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetDenoiser(b.Context(), nil, config, opts)
	}
}

// BenchmarkGetDenoiserDefault benchmarks default behavior
func BenchmarkGetDenoiserDefault(b *testing.B) {

	config := &protos.AudioConfig{SampleRate: 16000}
	opts := utils.Option{DenoiserOptionsKeyProvider: DenoiserIdentifier("")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetDenoiser(b.Context(), nil, config, opts)
	}
}

// BenchmarkDenoiserIdentifierString benchmarks string conversion
func BenchmarkDenoiserIdentifierString(b *testing.B) {
	identifier := KRISP
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = string(identifier)
	}
}
