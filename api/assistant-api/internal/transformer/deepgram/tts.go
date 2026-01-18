// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_deepgram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	deepgram_internal "github.com/rapidaai/api/assistant-api/internal/transformer/deepgram/internal"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	utils "github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

/*
Deepgram Continuous Streaming TTS
Reference: https://developers.deepgram.com/reference/text-to-speech/speak-streaming
*/

type deepgramTTS struct {
	*deepgramOption
	// context management
	ctx       context.Context
	ctxCancel context.CancelFunc
	contextId string
	mu        sync.Mutex

	logger     commons.Logger
	connection *websocket.Conn
	onPacket   func(pkt ...internal_type.Packet) error
}

func NewDeepgramTextToSpeech(ctx context.Context, logger commons.Logger, credential *protos.VaultCredential, audioConfig *protos.AudioConfig,
	onPacket func(pkt ...internal_type.Packet) error,
	opts utils.Option) (internal_type.TextToSpeechTransformer, error) {

	dGoptions, err := NewDeepgramOption(logger, credential, audioConfig, opts)
	if err != nil {
		logger.Errorf("deepgram-tts: error while intializing deepgram text to speech")
		return nil, err
	}
	ctx2, cancel := context.WithCancel(ctx)

	return &deepgramTTS{
		deepgramOption: dGoptions,
		ctx:            ctx2,
		ctxCancel:      cancel,
		logger:         logger,
		onPacket:       onPacket,
	}, nil
}

// Initialize implements internal_transformer.OutputAudioTransformer.
func (t *deepgramTTS) Initialize() error {

	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("token %s", t.GetKey()))
	conn, resp, err := websocket.DefaultDialer.Dial(t.GetTextToSpeechConnectionString(), header)
	if err != nil {
		t.logger.Errorf("deepgram-tts: websocket dial failed err=%v resp=%v", err, resp)
		return err
	}

	t.mu.Lock()
	t.connection = conn
	t.mu.Unlock()

	go t.textToSpeechCallback(conn, t.ctx)
	return nil
}

// Name implements internal_transformer.OutputAudioTransformer.
func (*deepgramTTS) Name() string {
	return "deepgram-text-to-speech"
}

// readLoop handles server → client messages
func (t *deepgramTTS) textToSpeechCallback(conn *websocket.Conn, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			t.logger.Infof("deepgram-tts: context cancelled, stopping read loop")
			return
		default:
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				if errors.Is(err, io.EOF) || websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					t.logger.Infof("deepgram-tts: websocket closed gracefully")
					return
				}
				t.logger.Errorf("deepgram-tts: read error %v", err)
				return
			}

			if msgType == websocket.BinaryMessage {
				t.onPacket(internal_type.TextToSpeechAudioPacket{
					ContextID:  t.contextId,
					AudioChunk: data,
				})
				continue
			}

			var envelope *deepgram_internal.DeepgramTextToSpeechResponse
			if err := json.Unmarshal(data, &envelope); err != nil {
				continue
			}

			switch envelope.Type {
			case "Metadata":
				// ignoreing metadata for now
				continue

			case "Flushed":
				t.onPacket(internal_type.TextToSpeechEndPacket{
					ContextID: t.contextId,
				})
				continue

			case "Cleared":
				// ignoreing metadata for now
				continue

			case "Warning":
				t.logger.Warnf("deepgram-tts warning code=%s message=%s", envelope.Code, envelope.Message)
			}
		}
	}
}

// Transform streams text into Deepgram
func (t *deepgramTTS) Transform(ctx context.Context, in internal_type.LLMPacket) error {
	t.mu.Lock()
	conn := t.connection
	currentCtx := t.contextId
	if in.ContextId() != t.contextId {
		t.contextId = in.ContextId()
	}
	t.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("deepgram-tts: websocket not initialized")
	}

	if currentCtx != t.contextId && currentCtx != "" {
		_ = conn.WriteJSON(map[string]interface{}{
			"type": "Clear",
		})
	}

	switch input := in.(type) {
	case internal_type.LLMStreamPacket:
		// if the request is for complete then we just flush the stream
		if err := conn.WriteJSON(map[string]interface{}{
			"type": "Speak",
			"text": input.Text,
		}); err != nil {
			t.logger.Errorf("deepgram-tts: failed to send Speak message %v", err)
		}

		return nil
	case internal_type.LLMMessagePacket:
		t.logger.Debugf("flushing %s", input.ContextID)
		if err := conn.WriteJSON(map[string]string{"type": "Flush"}); err != nil {
			t.logger.Errorf("deepgram-tts: failed to send Flush %v", err)
			return err
		}
		return nil
	default:
		return fmt.Errorf("deepgram-tts: unsupported input type %T", in)
	}

}

// Close gracefully closes the Deepgram connection
func (t *deepgramTTS) Close(ctx context.Context) error {
	t.ctxCancel()
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connection != nil {
		_ = t.connection.WriteJSON(map[string]string{
			"type": "Close",
		})
		t.connection.Close()
		t.connection = nil
	}
	return nil
}
