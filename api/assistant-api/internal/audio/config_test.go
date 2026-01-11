// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_audio

import (
	"sync"
	"testing"

	"github.com/rapidaai/protos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMulaw8khzMonoAudioConfig validates Mulaw 8kHz mono audio configuration
func TestNewMulaw8khzMonoAudioConfig(t *testing.T) {
	tests := []struct {
		name      string
		assertion func(t *testing.T, config *protos.AudioConfig)
	}{
		{
			name: "returns non-nil config",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.NotNil(t, config)
			},
		},
		{
			name: "sets correct sample rate to 8000",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(8000), config.SampleRate)
			},
		},
		{
			name: "sets correct audio format to MuLaw8",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, protos.AudioConfig_MuLaw8, config.AudioFormat)
			},
		},
		{
			name: "sets correct channels to mono (1)",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(1), config.Channels)
			},
		},
		{
			name: "all fields set correctly",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(8000), config.SampleRate)
				assert.Equal(t, protos.AudioConfig_MuLaw8, config.AudioFormat)
				assert.Equal(t, uint32(1), config.Channels)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewMulaw8khzMonoAudioConfig()
			tt.assertion(t, config)
		})
	}
}

// TestNewLinear24khzMonoAudioConfig validates Linear 24kHz mono audio configuration
func TestNewLinear24khzMonoAudioConfig(t *testing.T) {
	tests := []struct {
		name      string
		assertion func(t *testing.T, config *protos.AudioConfig)
	}{
		{
			name: "returns non-nil config",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.NotNil(t, config)
			},
		},
		{
			name: "sets correct sample rate to 24000",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(24000), config.SampleRate)
			},
		},
		{
			name: "sets correct audio format to LINEAR16",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, protos.AudioConfig_LINEAR16, config.AudioFormat)
			},
		},
		{
			name: "sets correct channels to mono (1)",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(1), config.Channels)
			},
		},
		{
			name: "all fields set correctly",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(24000), config.SampleRate)
				assert.Equal(t, protos.AudioConfig_LINEAR16, config.AudioFormat)
				assert.Equal(t, uint32(1), config.Channels)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewLinear24khzMonoAudioConfig()
			tt.assertion(t, config)
		})
	}
}

// TestNewLinear16khzMonoAudioConfig validates Linear 16kHz mono audio configuration
func TestNewLinear16khzMonoAudioConfig(t *testing.T) {
	tests := []struct {
		name      string
		assertion func(t *testing.T, config *protos.AudioConfig)
	}{
		{
			name: "returns non-nil config",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.NotNil(t, config)
			},
		},
		{
			name: "sets correct sample rate to 16000",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(16000), config.SampleRate)
			},
		},
		{
			name: "sets correct audio format to LINEAR16",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, protos.AudioConfig_LINEAR16, config.AudioFormat)
			},
		},
		{
			name: "sets correct channels to mono (1)",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(1), config.Channels)
			},
		},
		{
			name: "all fields set correctly",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(16000), config.SampleRate)
				assert.Equal(t, protos.AudioConfig_LINEAR16, config.AudioFormat)
				assert.Equal(t, uint32(1), config.Channels)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewLinear16khzMonoAudioConfig()
			tt.assertion(t, config)
		})
	}
}

// TestNewLinear8khzMonoAudioConfig validates Linear 8kHz mono audio configuration
func TestNewLinear8khzMonoAudioConfig(t *testing.T) {
	tests := []struct {
		name      string
		assertion func(t *testing.T, config *protos.AudioConfig)
	}{
		{
			name: "returns non-nil config",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.NotNil(t, config)
			},
		},
		{
			name: "sets correct sample rate to 8000",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(8000), config.SampleRate)
			},
		},
		{
			name: "sets correct audio format to LINEAR16",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, protos.AudioConfig_LINEAR16, config.AudioFormat)
			},
		},
		{
			name: "sets correct channels to mono (1)",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(1), config.Channels)
			},
		},
		{
			name: "all fields set correctly",
			assertion: func(t *testing.T, config *protos.AudioConfig) {
				assert.Equal(t, uint32(8000), config.SampleRate)
				assert.Equal(t, protos.AudioConfig_LINEAR16, config.AudioFormat)
				assert.Equal(t, uint32(1), config.Channels)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewLinear8khzMonoAudioConfig()
			tt.assertion(t, config)
		})
	}
}

// TestAllAudioConfigsIsolated tests that configs are independent and not shared
func TestAllAudioConfigsIsolated(t *testing.T) {
	config1 := NewMulaw8khzMonoAudioConfig()
	config2 := NewMulaw8khzMonoAudioConfig()

	// Verify they are different instances
	assert.False(t, config1 == config2, "configs should be different instances")

	// Modify one and ensure other is not affected
	originalRate := config2.SampleRate
	config1.SampleRate = 16000
	assert.Equal(t, originalRate, config2.SampleRate, "modifying config1 should not affect config2")
}

// TestAudioConfigsDifferent tests that different config types have correct values
func TestAudioConfigsDifferent(t *testing.T) {
	mulaw8 := NewMulaw8khzMonoAudioConfig()
	linear8 := NewLinear8khzMonoAudioConfig()
	linear16 := NewLinear16khzMonoAudioConfig()
	linear24 := NewLinear24khzMonoAudioConfig()

	// Test sample rates
	assert.Equal(t, uint32(8000), mulaw8.SampleRate)
	assert.Equal(t, uint32(8000), linear8.SampleRate)
	assert.Equal(t, uint32(16000), linear16.SampleRate)
	assert.Equal(t, uint32(24000), linear24.SampleRate)

	// Test audio formats
	assert.Equal(t, protos.AudioConfig_MuLaw8, mulaw8.AudioFormat)
	assert.Equal(t, protos.AudioConfig_LINEAR16, linear8.AudioFormat)
	assert.Equal(t, protos.AudioConfig_LINEAR16, linear16.AudioFormat)
	assert.Equal(t, protos.AudioConfig_LINEAR16, linear24.AudioFormat)

	// All should be mono
	assert.Equal(t, uint32(1), mulaw8.Channels)
	assert.Equal(t, uint32(1), linear8.Channels)
	assert.Equal(t, uint32(1), linear16.Channels)
	assert.Equal(t, uint32(1), linear24.Channels)
}

// TestAudioConfigsConcurrentAccess tests thread-safe access to configs
func TestAudioConfigsConcurrentAccess(t *testing.T) {
	const numGoroutines = 100
	const iterationsPerGoroutine = 100

	var wg sync.WaitGroup
	var errorCount int
	var errorMutex sync.Mutex

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < iterationsPerGoroutine; j++ {
				var config *protos.AudioConfig
				switch goroutineID % 4 {
				case 0:
					config = NewMulaw8khzMonoAudioConfig()
				case 1:
					config = NewLinear8khzMonoAudioConfig()
				case 2:
					config = NewLinear16khzMonoAudioConfig()
				case 3:
					config = NewLinear24khzMonoAudioConfig()
				}

				if config == nil {
					errorMutex.Lock()
					errorCount++
					errorMutex.Unlock()
					continue
				}

				// Verify all fields are set
				if config.SampleRate == 0 || config.Channels == 0 {
					errorMutex.Lock()
					errorCount++
					errorMutex.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()
	assert.Zero(t, errorCount, "no errors should occur during concurrent access")
}

// BenchmarkNewMulaw8khzMonoAudioConfig benchmarks Mulaw config creation
func BenchmarkNewMulaw8khzMonoAudioConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewMulaw8khzMonoAudioConfig()
	}
}

// BenchmarkNewLinear8khzMonoAudioConfig benchmarks Linear 8kHz config creation
func BenchmarkNewLinear8khzMonoAudioConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLinear8khzMonoAudioConfig()
	}
}

// BenchmarkNewLinear16khzMonoAudioConfig benchmarks Linear 16kHz config creation
func BenchmarkNewLinear16khzMonoAudioConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLinear16khzMonoAudioConfig()
	}
}

// BenchmarkNewLinear24khzMonoAudioConfig benchmarks Linear 24kHz config creation
func BenchmarkNewLinear24khzMonoAudioConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLinear24khzMonoAudioConfig()
	}
}

// BenchmarkAllConfigsCreation benchmarks creating all config types
func BenchmarkAllConfigsCreation(b *testing.B) {
	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewMulaw8khzMonoAudioConfig()
			_ = NewLinear8khzMonoAudioConfig()
			_ = NewLinear16khzMonoAudioConfig()
			_ = NewLinear24khzMonoAudioConfig()
		}
	})
}

// TestAudioConfigFieldValidation validates all fields are properly initialized
func TestAudioConfigFieldValidation(t *testing.T) {
	configs := map[string]*protos.AudioConfig{
		"Mulaw8":   NewMulaw8khzMonoAudioConfig(),
		"Linear8":  NewLinear8khzMonoAudioConfig(),
		"Linear16": NewLinear16khzMonoAudioConfig(),
		"Linear24": NewLinear24khzMonoAudioConfig(),
	}

	for name, config := range configs {
		t.Run(name, func(t *testing.T) {
			// All configs should be non-nil
			require.NotNil(t, config)

			// All configs should have positive sample rates
			assert.Greater(t, config.SampleRate, uint32(0), "sample rate should be positive")

			// All configs should have valid AudioFormat (either LINEAR16 or MuLaw8)
			validFormat := config.AudioFormat == protos.AudioConfig_LINEAR16 || config.AudioFormat == protos.AudioConfig_MuLaw8
			assert.True(t, validFormat, "audio format should be LINEAR16 or MuLaw8")

			// All configs should be mono (channels = 1)
			assert.Equal(t, uint32(1), config.Channels, "all configs should be mono")
		})
	}
}

// TestMultipleCallsConsistency ensures multiple calls return consistent configurations
func TestMultipleCallsConsistency(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("Call_"+string(rune(i)), func(t *testing.T) {
			config1 := NewMulaw8khzMonoAudioConfig()
			config2 := NewMulaw8khzMonoAudioConfig()

			assert.Equal(t, config1.SampleRate, config2.SampleRate)
			assert.Equal(t, config1.AudioFormat, config2.AudioFormat)
			assert.Equal(t, config1.Channels, config2.Channels)
		})
	}
}

// TestAudioConfigsSampleRates validates all expected sample rates
func TestAudioConfigsSampleRates(t *testing.T) {
	tests := []struct {
		name         string
		factoryFunc  func() *protos.AudioConfig
		expectedRate uint32
	}{
		{
			name:         "Mulaw8kHz",
			factoryFunc:  NewMulaw8khzMonoAudioConfig,
			expectedRate: 8000,
		},
		{
			name:         "Linear8kHz",
			factoryFunc:  NewLinear8khzMonoAudioConfig,
			expectedRate: 8000,
		},
		{
			name:         "Linear16kHz",
			factoryFunc:  NewLinear16khzMonoAudioConfig,
			expectedRate: 16000,
		},
		{
			name:         "Linear24kHz",
			factoryFunc:  NewLinear24khzMonoAudioConfig,
			expectedRate: 24000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.factoryFunc()
			assert.Equal(t, tt.expectedRate, config.SampleRate)
		})
	}
}

// TestAudioFormatsDistinction validates different audio formats
func TestAudioFormatsDistinction(t *testing.T) {
	mulaw := NewMulaw8khzMonoAudioConfig()
	linear := NewLinear8khzMonoAudioConfig()

	// Same sample rate but different formats
	assert.Equal(t, mulaw.SampleRate, linear.SampleRate)
	assert.NotEqual(t, mulaw.AudioFormat, linear.AudioFormat)
	assert.Equal(t, protos.AudioConfig_MuLaw8, mulaw.AudioFormat)
	assert.Equal(t, protos.AudioConfig_LINEAR16, linear.AudioFormat)
}

// TestConfigPointerIndependence ensures returned pointers point to different objects
func TestConfigPointerIndependence(t *testing.T) {
	const callCount = 5

	configs := make([]*protos.AudioConfig, callCount)
	for i := 0; i < callCount; i++ {
		configs[i] = NewMulaw8khzMonoAudioConfig()
	}

	// Verify all pointers are unique
	for i := 0; i < callCount; i++ {
		for j := i + 1; j < callCount; j++ {
			assert.NotSame(t, configs[i], configs[j], "pointers should be different")
		}
	}
}

// TestConcurrentModificationResistance tests that concurrent mods to one instance don't affect others
func TestConcurrentModificationResistance(t *testing.T) {
	config1 := NewLinear16khzMonoAudioConfig()
	originalSampleRate := config1.SampleRate

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		config1.SampleRate = 48000
	}()

	go func() {
		defer wg.Done()
		// Create new config while another modifies config1
		config2 := NewLinear16khzMonoAudioConfig()
		assert.Equal(t, uint32(16000), config2.SampleRate)
	}()

	wg.Wait()
	// config1 was modified, but new configs should still have original values
	assert.Equal(t, originalSampleRate, NewLinear16khzMonoAudioConfig().SampleRate)
}

// TestAudioConfigsNotNil validates that all factory functions return non-nil configs
func TestAudioConfigsNotNil(t *testing.T) {
	factoryFunctions := map[string]func() *protos.AudioConfig{
		"NewMulaw8khzMonoAudioConfig":   NewMulaw8khzMonoAudioConfig,
		"NewLinear8khzMonoAudioConfig":  NewLinear8khzMonoAudioConfig,
		"NewLinear16khzMonoAudioConfig": NewLinear16khzMonoAudioConfig,
		"NewLinear24khzMonoAudioConfig": NewLinear24khzMonoAudioConfig,
	}

	for name, factory := range factoryFunctions {
		t.Run(name, func(t *testing.T) {
			config := factory()
			assert.NotNil(t, config)
		})
	}
}

// TestAudioChannelsAlwaysMono validates that all configs have channels = 1
func TestAudioChannelsAlwaysMono(t *testing.T) {
	configs := []*protos.AudioConfig{
		NewMulaw8khzMonoAudioConfig(),
		NewLinear8khzMonoAudioConfig(),
		NewLinear16khzMonoAudioConfig(),
		NewLinear24khzMonoAudioConfig(),
	}

	for _, config := range configs {
		assert.Equal(t, uint32(1), config.Channels, "all configs should be mono (channels=1)")
	}
}

// TestRaceConditionOnConcurrentReads tests multiple goroutines reading same config
func TestRaceConditionOnConcurrentReads(t *testing.T) {
	config := NewLinear16khzMonoAudioConfig()
	const numGoroutines = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			// Read the config multiple times
			_ = config.SampleRate
			_ = config.Channels
			_ = config.AudioFormat
		}()
	}

	wg.Wait()
}
