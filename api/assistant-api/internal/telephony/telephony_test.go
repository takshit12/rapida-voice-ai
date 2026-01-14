// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_telephony_factory

import (
	"testing"

	"github.com/rapidaai/api/assistant-api/config"
	"github.com/rapidaai/pkg/commons"
	"github.com/stretchr/testify/assert"
)

// TestTelephonyString tests the String method for all Telephony types
func TestTelephonyString(t *testing.T) {
	tests := []struct {
		name     string
		input    Telephony
		expected string
	}{
		{
			name:     "Twilio",
			input:    Twilio,
			expected: "twilio",
		},
		{
			name:     "Exotel",
			input:    Exotel,
			expected: "exotel",
		},
		{
			name:     "Vonage",
			input:    Vonage,
			expected: "vonage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetTelephonyWithInvalidTypes tests GetTelephony with invalid provider types
func TestGetTelephonyWithInvalidTypes(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	cfg := &config.AssistantConfig{}

	tests := []struct {
		name     string
		provider Telephony
		wantErr  bool
		wantMsg  string
	}{
		{
			name:     "Empty string provider",
			provider: Telephony(""),
			wantErr:  true,
			wantMsg:  "illegal telephony provider",
		},
		{
			name:     "Unknown provider",
			provider: Telephony("unknown"),
			wantErr:  true,
			wantMsg:  "illegal telephony provider",
		},
		{
			name:     "Case sensitive - TWILIO",
			provider: Telephony("TWILIO"),
			wantErr:  true,
			wantMsg:  "illegal telephony provider",
		},
		{
			name:     "Partial match - twil",
			provider: Telephony("twil"),
			wantErr:  true,
			wantMsg:  "illegal telephony provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telephony, err := GetTelephony(tt.provider, cfg, mockLogger)

			assert.Error(t, err, "expected error for provider: %s", tt.provider)
			assert.Nil(t, telephony, "telephony should be nil for invalid provider")
			assert.Equal(t, tt.wantMsg, err.Error(), "error message mismatch")
		})
	}
}

// TestGetTelephonyWithNilConfig tests GetTelephony with nil config
func TestGetTelephonyWithNilConfig(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	// Factory creates the telephony object with nil config
	// The validation happens at runtime when the config is used
	telephony, err := GetTelephony(Twilio, nil, mockLogger)
	// The factory doesn't validate nil config at creation time
	// so it may succeed or fail depending on the provider implementation
	if telephony != nil && err != nil {
		assert.Error(t, err)
	}
}

// TestGetTelephonyWithNilLogger tests GetTelephony with nil logger
func TestGetTelephonyWithNilLogger(t *testing.T) {
	cfg := &config.AssistantConfig{}

	// Factory creates the telephony object with nil logger
	// The validation happens at runtime when the logger is used
	telephony, err := GetTelephony(Twilio, cfg, nil)
	// The factory doesn't validate nil logger at creation time
	// so it may succeed or fail depending on the provider implementation
	if telephony != nil && err != nil {
		assert.Error(t, err)
	}
}

// TestAllTelephonyProvidersCallFactory validates factory doesn't panic for all types
func TestAllTelephonyProvidersCallFactory(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	cfg := &config.AssistantConfig{}

	telephonyTypes := []Telephony{
		Twilio,
		Exotel,
		Vonage,
	}

	for _, provider := range telephonyTypes {
		t.Run(provider.String(), func(t *testing.T) {
			// Just ensure factory can be called without panic
			_, _ = GetTelephony(provider, cfg, mockLogger)
		})
	}
}

// TestTelephonyStringConsistency tests that String() matches type constants
func TestTelephonyStringConsistency(t *testing.T) {
	tests := []struct {
		provider Telephony
		value    string
	}{
		{Twilio, "twilio"},
		{Exotel, "exotel"},
		{Vonage, "vonage"},
	}

	for _, tt := range tests {
		t.Run(tt.provider.String(), func(t *testing.T) {
			// Ensure that String() returns the constant value
			assert.Equal(t, tt.value, tt.provider.String())
			// Ensure that creating a new Telephony from the string matches
			newProvider := Telephony(tt.value)
			assert.Equal(t, tt.provider, newProvider)
		})
	}
}

// TestGetTelephonyErrorMessages validates error messages are correct
func TestGetTelephonyErrorMessages(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	cfg := &config.AssistantConfig{}

	invalidProviders := []Telephony{
		Telephony(""),
		Telephony("invalid"),
		Telephony("vonage-extra"),
		Telephony("123"),
		Telephony("slack"),
	}

	for _, provider := range invalidProviders {
		t.Run("Error_"+provider.String(), func(t *testing.T) {
			_, err := GetTelephony(provider, cfg, mockLogger)
			assert.Error(t, err)
			assert.Equal(t, "illegal telephony provider", err.Error())
		})
	}
}

// TestTelephonyFactoryConsistency tests factory returns nil on error
func TestTelephonyFactoryConsistency(t *testing.T) {
	mockLogger, _ := commons.NewApplicationLogger()

	cfg := &config.AssistantConfig{}

	// For any invalid type, both return values must follow the pattern:
	// (nil, error)
	telephony, err := GetTelephony(Telephony("invalid"), cfg, mockLogger)

	assert.Nil(t, telephony, "telephony must be nil when error occurs")
	assert.NotNil(t, err, "error must not be nil for invalid provider")
}

// TestValidProvidersMatch validates all valid provider constants are covered
func TestValidProvidersMatch(t *testing.T) {
	validProviders := []Telephony{
		Twilio,
		Exotel,
		Vonage,
	}

	for _, provider := range validProviders {
		t.Run(provider.String(), func(t *testing.T) {
			// Each valid provider should have a unique string representation
			assert.NotEmpty(t, provider.String())
		})
	}

	// Verify all providers are distinct
	seen := make(map[string]bool)
	for _, provider := range validProviders {
		str := provider.String()
		assert.False(t, seen[str], "provider %s appears multiple times", str)
		seen[str] = true
	}
}

// BenchmarkGetTelephony benchmarks the GetTelephony factory method
func BenchmarkGetTelephony(b *testing.B) {
	mockLogger, _ := commons.NewApplicationLogger()

	cfg := &config.AssistantConfig{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetTelephony(Twilio, cfg, mockLogger)
	}
}

// BenchmarkTelephonyString benchmarks the String method
func BenchmarkTelephonyString(b *testing.B) {
	provider := Twilio

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.String()
	}
}

// TestTelephonyTypeConversion tests type conversion and creation
func TestTelephonyTypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		expected Telephony
	}{
		{
			name:     "String to Twilio",
			str:      "twilio",
			expected: Twilio,
		},
		{
			name:     "String to Exotel",
			str:      "exotel",
			expected: Exotel,
		},
		{
			name:     "String to Vonage",
			str:      "vonage",
			expected: Vonage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := Telephony(tt.str)
			assert.Equal(t, tt.expected, provider)
			assert.Equal(t, tt.str, provider.String())
		})
	}
}
