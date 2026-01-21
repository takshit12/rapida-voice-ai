// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_telephony_base

import (
	"bytes"
	"context"
	"encoding/base64"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rapidaai/protos"
)

// TestCreateVoiceRequest tests the CreateVoiceRequest method
func TestCreateVoiceRequest(t *testing.T) {
	streamer := &BaseTelephonyStreamer{}

	tests := []struct {
		name      string
		audioData []byte
	}{
		{
			name:      "Empty audio data",
			audioData: []byte{},
		},
		{
			name:      "Small audio chunk",
			audioData: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name:      "Nil audio data",
			audioData: nil,
		},
		{
			name:      "Large audio chunk (1KB)",
			audioData: make([]byte, 1024),
		},
		{
			name:      "Large audio chunk (10KB)",
			audioData: make([]byte, 10*1024),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := streamer.CreateVoiceRequest(tt.audioData)

			require.NotNil(t, request)
			require.NotNil(t, request.GetMessage())
			require.NotNil(t, request.GetMessage().GetAudio())

			// Verify audio content matches
			if tt.audioData != nil {
				assert.Equal(t, tt.audioData, request.GetMessage().GetAudio().Content)
			}
		})
	}
}

// TestContext tests the Context method
func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	streamer := &BaseTelephonyStreamer{
		ctx: ctx,
	}

	retrievedCtx := streamer.Context()
	assert.Equal(t, ctx, retrievedCtx)

	// Verify context is active
	select {
	case <-retrievedCtx.Done():
		t.Error("Context should not be done initially")
	default:
		// Expected
	}
}

// TestCancel tests the Cancel method
func TestCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	streamer := &BaseTelephonyStreamer{
		ctx:        ctx,
		cancelFunc: cancel,
		conn:       nil, // No real connection for this test
	}

	// Verify context is active before cancel
	select {
	case <-streamer.ctx.Done():
		t.Error("Context should not be cancelled before Cancel()")
	default:
		// Expected
	}

	// Cancel the streamer
	err := streamer.Cancel()
	assert.NoError(t, err)

	// Verify context is cancelled after Cancel()
	select {
	case <-streamer.ctx.Done():
		// Expected - context is cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after Cancel()")
	}

	// Verify context error
	assert.Equal(t, context.Canceled, streamer.ctx.Err())

	// Verify connection is nil
	assert.Nil(t, streamer.conn)
}

// TestEncoder tests the Encoder method
func TestEncoder(t *testing.T) {
	streamer := &BaseTelephonyStreamer{
		encoder: base64.StdEncoding,
	}

	encoder := streamer.Encoder()
	assert.NotNil(t, encoder)
	assert.Equal(t, base64.StdEncoding, encoder)

	// Verify encoder works correctly
	testData := []byte("test audio data 123")
	encoded := encoder.EncodeToString(testData)
	decoded, err := encoder.DecodeString(encoded)
	require.NoError(t, err)
	assert.Equal(t, testData, decoded)
}

// TestCredential tests the Credential and VaultCredential methods
func TestCredential(t *testing.T) {
	credential := &protos.VaultCredential{
		Id:   42,
		Name: "test-credential",
	}

	streamer := &BaseTelephonyStreamer{
		vaultCredential: credential,
	}

	// Test Credential()
	cred := streamer.Credential()
	require.NotNil(t, cred)
	assert.Equal(t, uint64(42), cred.Id)
	assert.Equal(t, "test-credential", cred.Name)

	// Test VaultCredential()
	vaultCred := streamer.VaultCredential()
	require.NotNil(t, vaultCred)
	assert.Equal(t, credential, vaultCred)
	assert.Same(t, credential, vaultCred) // Same pointer
}

// TestInputBuffer tests the InputBuffer getter
func TestInputBuffer(t *testing.T) {
	buffer := new(bytes.Buffer)
	streamer := &BaseTelephonyStreamer{
		inputAudioBuffer: buffer,
	}

	retrievedBuffer := streamer.InputBuffer()
	assert.Same(t, buffer, retrievedBuffer)

	// Test writing to buffer
	testData := []byte("audio input data")
	retrievedBuffer.Write(testData)
	assert.Equal(t, len(testData), retrievedBuffer.Len())
	assert.Equal(t, testData, retrievedBuffer.Bytes())
}

// TestOutputBuffer tests the OutputBuffer getter
func TestOutputBuffer(t *testing.T) {
	buffer := new(bytes.Buffer)
	streamer := &BaseTelephonyStreamer{
		outputAudioBuffer: buffer,
	}

	retrievedBuffer := streamer.OutputBuffer()
	assert.Same(t, buffer, retrievedBuffer)

	// Test writing to buffer
	testData := []byte("audio output data")
	retrievedBuffer.Write(testData)
	assert.Equal(t, len(testData), retrievedBuffer.Len())
	assert.Equal(t, testData, retrievedBuffer.Bytes())
}

// TestBufferLocking tests thread-safe buffer access with mutexes
func TestBufferLocking(t *testing.T) {
	t.Run("Input buffer concurrent writes", func(t *testing.T) {
		streamer := &BaseTelephonyStreamer{
			inputAudioBuffer: new(bytes.Buffer),
		}

		var wg sync.WaitGroup
		iterations := 100

		// Concurrent writes to input buffer with locking
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(val byte) {
				defer wg.Done()
				streamer.LockInputAudioBuffer()
				streamer.InputBuffer().Write([]byte{val})
				streamer.UnlockInputAudioBuffer()
			}(byte(i))
		}

		wg.Wait()
		assert.Equal(t, iterations, streamer.InputBuffer().Len())
	})

	t.Run("Output buffer concurrent writes", func(t *testing.T) {
		streamer := &BaseTelephonyStreamer{
			outputAudioBuffer: new(bytes.Buffer),
		}

		var wg sync.WaitGroup
		iterations := 100

		// Concurrent writes to output buffer with locking
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func(val byte) {
				defer wg.Done()
				streamer.LockOutputAudioBuffer()
				streamer.OutputBuffer().Write([]byte{val})
				streamer.UnlockOutputAudioBuffer()
			}(byte(i))
		}

		wg.Wait()
		assert.Equal(t, iterations, streamer.OutputBuffer().Len())
	})

	t.Run("Mixed buffer access", func(t *testing.T) {
		streamer := &BaseTelephonyStreamer{
			inputAudioBuffer:  new(bytes.Buffer),
			outputAudioBuffer: new(bytes.Buffer),
		}

		var wg sync.WaitGroup
		iterations := 50

		// Concurrent access to both buffers
		for i := 0; i < iterations; i++ {
			wg.Add(2)

			go func(val byte) {
				defer wg.Done()
				streamer.LockInputAudioBuffer()
				streamer.InputBuffer().Write([]byte{val})
				streamer.UnlockInputAudioBuffer()
			}(byte(i))

			go func(val byte) {
				defer wg.Done()
				streamer.LockOutputAudioBuffer()
				streamer.OutputBuffer().Write([]byte{val})
				streamer.UnlockOutputAudioBuffer()
			}(byte(i + 100))
			}

		wg.Wait()
		assert.Equal(t, iterations, streamer.InputBuffer().Len())
		assert.Equal(t, iterations, streamer.OutputBuffer().Len())
	})
}

// TestConnection tests the Connection getter
func TestConnection(t *testing.T) {
	// Note: We can't easily test with a real WebSocket connection
	// This test just verifies the getter returns what was set
	streamer := &BaseTelephonyStreamer{
		conn: nil, // Would be a real *websocket.Conn in production
	}

	conn := streamer.Connection()
	assert.Nil(t, conn) // Nil in this test case
}

// TestContextCancellation tests that context cancellation works properly
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	streamer := &BaseTelephonyStreamer{
		ctx:        ctx,
		cancelFunc: cancel,
		conn:       nil,
	}

	// Start a goroutine that waits for context cancellation
	done := make(chan bool)
	go func() {
		<-streamer.Context().Done()
		done <- true
	}()

	// Cancel the streamer
	err := streamer.Cancel()
	assert.NoError(t, err)

	// Wait for goroutine to receive cancellation
	select {
	case <-done:
		// Expected - goroutine received cancellation
	case <-time.After(1 * time.Second):
		t.Error("Goroutine did not receive context cancellation")
	}
}

// Benchmark tests
func BenchmarkCreateVoiceRequest(b *testing.B) {
	streamer := &BaseTelephonyStreamer{}
	audioData := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = streamer.CreateVoiceRequest(audioData)
	}
}

func BenchmarkCreateVoiceRequest_LargeAudio(b *testing.B) {
	streamer := &BaseTelephonyStreamer{}
	audioData := make([]byte, 10*1024) // 10KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = streamer.CreateVoiceRequest(audioData)
	}
}

func BenchmarkBufferLocking(b *testing.B) {
	streamer := &BaseTelephonyStreamer{
		inputAudioBuffer: new(bytes.Buffer),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		streamer.LockInputAudioBuffer()
		streamer.UnlockInputAudioBuffer()
	}
}

func BenchmarkBufferWrite(b *testing.B) {
	streamer := &BaseTelephonyStreamer{
		inputAudioBuffer: new(bytes.Buffer),
	}
	data := []byte{0x01}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		streamer.LockInputAudioBuffer()
		streamer.InputBuffer().Write(data)
		streamer.UnlockInputAudioBuffer()
	}
}

func BenchmarkEncoder(b *testing.B) {
	streamer := &BaseTelephonyStreamer{
		encoder: base64.StdEncoding,
	}
	data := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder := streamer.Encoder()
		_ = encoder.EncodeToString(data)
	}
}
