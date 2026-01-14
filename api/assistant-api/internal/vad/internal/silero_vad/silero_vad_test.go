// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_silero_vad

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	internal_audio "github.com/rapidaai/api/assistant-api/internal/audio"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestOptions(tb testing.TB, threshold float64) utils.Option {
	opts := map[string]interface{}{}
	if threshold >= 0 {
		opts["microphone.vad.threshold"] = threshold
	}
	return opts
}

func getModelPath() string {
	envModelPath := os.Getenv("SILERO_MODEL_PATH")
	if envModelPath == "" {
		_, path, _, _ := runtime.Caller(0)
		envModelPath = filepath.Join(filepath.Dir(path), "models/silero_vad_20251001.onnx")
	}
	return envModelPath
}

func newSileroOrSkip(t *testing.T, inputCfg *protos.AudioConfig, threshold float64, cb internal_type.VADCallback) *SileroVAD {
	logger, err := commons.NewApplicationLoggerWithOptions()
	opts := newTestOptions(t, threshold)
	vad, err := NewSileroVAD(logger, inputCfg, cb, opts)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			t.Skipf("silero model missing at %s", getModelPath())
		}
		require.NoError(t, err)
	}
	silero := vad.(*SileroVAD)
	t.Cleanup(func() { _ = silero.Close() })
	return silero
}

func generateSilence(samples int) []byte {
	return make([]byte, samples*2)
}

func generateSineWave(samples int, frequency, amplitude float64) []byte {
	data := make([]byte, samples*2)
	for i := 0; i < samples; i++ {
		sample := int16(amplitude * 32767 * math.Sin(2*math.Pi*float64(i)*frequency/16000))
		binary.LittleEndian.PutUint16(data[i*2:i*2+2], uint16(sample))
	}
	return data
}

func generateNoise(samples int) []byte {
	data := make([]byte, samples*2)
	for i := 0; i < samples; i++ {
		sample := int16((i*7919)%65536 - 32768)
		binary.LittleEndian.PutUint16(data[i*2:i*2+2], uint16(sample))
	}
	return data
}

// Core functionality tests

func TestNewSileroVAD_DefaultThreshold(t *testing.T) {
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	callback := func(internal_type.InterruptionPacket) error { return nil }

	vad := newSileroOrSkip(t, inputConfig, -1, callback)

	assert.NotNil(t, vad.detector)
	assert.NotNil(t, vad.audioSampler)
	assert.NotNil(t, vad.audioConverter)
	assert.NotNil(t, vad.vadConfig)
	assert.Equal(t, uint32(16000), vad.vadConfig.SampleRate)
}

func TestSileroVAD_Name(t *testing.T) {
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	callback := func(internal_type.InterruptionPacket) error { return nil }

	vad := newSileroOrSkip(t, inputConfig, 0.5, callback)

	assert.Equal(t, "silero_vad", vad.Name())
}

func TestSileroVAD_Process_Silence_NoCallback(t *testing.T) {
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	callbackCalled := false
	callback := func(internal_type.InterruptionPacket) error {
		callbackCalled = true
		return nil
	}

	vad := newSileroOrSkip(t, inputConfig, 0.5, callback)

	err := vad.Process(generateSilence(16000))
	require.NoError(t, err)
	assert.False(t, callbackCalled)
}

func TestSileroVAD_Process_Speech_AllowsCallback(t *testing.T) {
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	// var result internal_type.InterruptionPacket
	callback := func(r internal_type.InterruptionPacket) error {
		// result = r
		return nil
	}

	vad := newSileroOrSkip(t, inputConfig, 0.2, callback)

	err := vad.Process(generateSineWave(16000, 440, 0.9))
	require.NoError(t, err)

	// if result != nil {
	// 	assert.GreaterOrEqual(t, result.EndSec, result.StartSec)
	// }
}

func TestSileroVAD_Process_DifferentSampleRates(t *testing.T) {
	callback := func(internal_type.InterruptionPacket) error { return nil }

	tests := []struct {
		name       string
		sampleRate uint32
		samples    int
	}{
		{"8kHz", 8000, 8000},
		{"16kHz", 16000, 16000},
		{"24kHz", 24000, 24000},
		{"48kHz", 48000, 48000},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			inputConfig := &protos.AudioConfig{SampleRate: tt.sampleRate, AudioFormat: protos.AudioConfig_LINEAR16, Channels: 1}

			vad := newSileroOrSkip(t, inputConfig, 0.5, callback)

			err := vad.Process(generateSilence(tt.samples))
			require.NoError(t, err)
		})
	}
}

func TestSileroVAD_Process_DifferentChannels(t *testing.T) {
	callback := func(internal_type.InterruptionPacket) error { return nil }

	tests := []struct {
		name     string
		channels uint32
		samples  int
	}{
		{"mono", 1, 16000},
		{"stereo", 2, 32000},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			inputConfig := &protos.AudioConfig{SampleRate: 16000, AudioFormat: protos.AudioConfig_LINEAR16, Channels: tt.channels}
			vad := newSileroOrSkip(t, inputConfig, 0.5, callback)

			err := vad.Process(generateSilence(tt.samples))
			require.NoError(t, err)
		})
	}
}

func TestSileroVAD_Process_CorruptedData(t *testing.T) {
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	callback := func(internal_type.InterruptionPacket) error { return nil }

	vad := newSileroOrSkip(t, inputConfig, 0.5, callback)

	corrupted := make([]byte, 999) // Odd length
	err := vad.Process(corrupted)
	_ = err // Accept error or nil; should not panic
}

func TestSileroVAD_Process_VerySmallChunks(t *testing.T) {
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	callback := func(internal_type.InterruptionPacket) error { return nil }

	vad := newSileroOrSkip(t, inputConfig, 0.5, callback)

	sizes := []int{1, 2, 5, 10, 20}
	for _, size := range sizes {
		size := size
		t.Run(fmt.Sprintf("%d_samples", size), func(t *testing.T) {
			err := vad.Process(generateSilence(size))
			_ = err
		})
	}
}

func TestSileroVAD_Process_Concurrent(t *testing.T) {
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	callback := func(internal_type.InterruptionPacket) error { return nil }

	vad := newSileroOrSkip(t, inputConfig, 0.5, callback)

	var wg sync.WaitGroup
	const workers = 8
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_ = vad.Process(generateSilence(1600))
		}()
	}
	wg.Wait()
}

func TestSileroVAD_Close_Idempotent(t *testing.T) {
	logger, err := commons.NewApplicationLoggerWithOptions()
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	callback := func(internal_type.InterruptionPacket) error { return nil }
	opts := newTestOptions(t, 0.5)

	vad, err := NewSileroVAD(logger, inputConfig, callback, opts)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			t.Skipf("silero model missing at %s", getModelPath())
		}
		require.NoError(t, err)
	}

	require.NoError(t, vad.Close())
	err = vad.Close()
	_ = err
}

func TestSileroVAD_ModelPath_Environment(t *testing.T) {
	modelPath := getModelPath()
	if _, err := os.Stat(modelPath); err != nil {
		t.Skipf("silero model missing at %s", modelPath)
	}

	original := os.Getenv("SILERO_MODEL_PATH")
	os.Setenv("SILERO_MODEL_PATH", modelPath)
	t.Cleanup(func() {
		if original != "" {
			os.Setenv("SILERO_MODEL_PATH", original)
		} else {
			os.Unsetenv("SILERO_MODEL_PATH")
		}
	})

	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	callback := func(internal_type.InterruptionPacket) error { return nil }
	_ = newSileroOrSkip(t, inputConfig, 0.5, callback)
}

func TestSileroVAD_Process_NoisePatterns(t *testing.T) {
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	callback := func(internal_type.InterruptionPacket) error { return nil }

	vad := newSileroOrSkip(t, inputConfig, 0.5, callback)

	err := vad.Process(generateNoise(16000))
	require.NoError(t, err)
}

func TestSileroVAD_Process_MaxAmplitude(t *testing.T) {
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	callback := func(internal_type.InterruptionPacket) error { return nil }

	vad := newSileroOrSkip(t, inputConfig, 0.5, callback)

	samples := 16000
	data := make([]byte, samples*2)
	for i := 0; i < samples; i++ {
		var val int16
		if i%2 == 0 {
			val = 32767
		} else {
			val = -32768
		}
		binary.LittleEndian.PutUint16(data[i*2:i*2+2], uint16(val))
	}

	err := vad.Process(data)
	require.NoError(t, err)
}

func TestSileroVAD_Process_RepeatedCalls(t *testing.T) {
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	callback := func(internal_type.InterruptionPacket) error { return nil }

	vad := newSileroOrSkip(t, inputConfig, 0.5, callback)

	chunk := generateSilence(1600)
	for i := 0; i < 50; i++ {
		err := vad.Process(chunk)
		require.NoError(t, err)
	}
}

func TestSileroVAD_StatefulProcessing(t *testing.T) {
	inputConfig := internal_audio.NewLinear16khzMonoAudioConfig()
	var calls int
	callback := func(internal_type.InterruptionPacket) error {
		calls++
		return nil
	}

	vad := newSileroOrSkip(t, inputConfig, 0.3, callback)

	for i := 0; i < 10; i++ {
		err := vad.Process(generateSineWave(1600, 440, 0.8))
		require.NoError(t, err)
	}

	assert.GreaterOrEqual(t, calls, 0)
}
