// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_elevenlabs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	elevenlabs_internal "github.com/rapidaai/api/assistant-api/internal/transformer/elevenlabs/internal"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type elevenlabsTTS struct {
	*elevenLabsOption
	// context management
	ctx       context.Context
	ctxCancel context.CancelFunc

	// mutex
	mu sync.Mutex

	logger     commons.Logger
	connection *websocket.Conn
	onPacket   func(pkt ...internal_type.Packet) error
}

func NewElevenlabsTextToSpeech(ctx context.Context, logger commons.Logger, credential *protos.VaultCredential, audioConfig *protos.AudioConfig,
	onPacket func(pkt ...internal_type.Packet) error,
	opts utils.Option) (internal_type.TextToSpeechTransformer, error) {
	eleOpts, err := NewElevenLabsOption(logger, credential, audioConfig, opts)
	if err != nil {
		logger.Errorf("elevenlabs-tts: intializing elevenlabs failed %+v", err)
		return nil, err
	}
	ctx2, contextCancel := context.WithCancel(ctx)
	return &elevenlabsTTS{
		ctx:              ctx2,
		ctxCancel:        contextCancel,
		onPacket:         onPacket,
		logger:           logger,
		elevenLabsOption: eleOpts,
	}, nil
}

// Initialize implements internal_transformer.OutputAudioTransformer.
func (ct *elevenlabsTTS) Initialize() error {
	header := http.Header{}
	header.Set("xi-api-key", ct.GetKey())
	conn, resp, err := websocket.DefaultDialer.Dial(ct.GetTextToSpeechConnectionString(), header)
	if err != nil {
		ct.logger.Errorf("elevenlab-tts: error while elevenlabs %s with response %v", err, resp)
		return err
	}

	ct.mu.Lock()
	ct.connection = conn
	defer ct.mu.Unlock()

	go ct.textToSpeechCallback(conn, ct.ctx)
	return nil
}

// Name implements internal_transformer.OutputAudioTransformer.
func (*elevenlabsTTS) Name() string {
	return "elevenlabs-text-to-speech"
}

func (elt *elevenlabsTTS) textToSpeechCallback(conn *websocket.Conn, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			elt.logger.Infof("elevenlabs-tts: context cancelled, stopping response listener")
			return
		default:
			_, audioChunk, err := conn.ReadMessage()
			if err != nil {
				if errors.Is(err, io.EOF) || websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					elt.logger.Infof("elevenlabs-tts: websocket closed gracefully")
					return
				}

				elt.logger.Errorf("elevenlabs-tts: websocket read error: %v", err)
				return
			}
			var audioData elevenlabs_internal.ElevenlabTextToSpeechResponse
			if err := json.Unmarshal(audioChunk, &audioData); err != nil {
				elt.logger.Errorf("elevenlab-tts: Error parsing audio chunk: %v", err)
				continue
			}

			if rawAudioData, err := base64.StdEncoding.DecodeString(audioData.Audio); err == nil {
				if audioData.ContextId != nil {
					elt.onPacket(internal_type.TextToSpeechAudioPacket{
						ContextID:  *audioData.ContextId,
						AudioChunk: rawAudioData,
					})
				}
			}

			if audioData.IsFinal != nil && *audioData.IsFinal {
				if audioData.ContextId != nil {
					elt.onPacket(internal_type.TextToSpeechEndPacket{
						ContextID: *audioData.ContextId,
					})
				}
			}
		}
	}

}

func (t *elevenlabsTTS) Transform(ctx context.Context, in internal_type.LLMPacket) error {
	t.mu.Lock()
	cnn := t.connection
	currentCtx := in.ContextId()
	t.mu.Unlock()

	if cnn == nil {
		return fmt.Errorf("elevenlabs-tts: websocket connection is not initialized")
	}

	switch input := in.(type) {
	case internal_type.LLMStreamPacket:
		if err := cnn.WriteJSON(map[string]interface{}{
			"text":       input.Text,
			"context_id": currentCtx,
			"flush":      false,
		}); err != nil {
			t.logger.Errorf("elevenlab-tts: unable to write json for text to speech: %v", err)
		}
		if err := cnn.WriteJSON(map[string]interface{}{
			"context_id": currentCtx,
			"flush":      true,
		}); err != nil {
			t.logger.Errorf("elevenlab-tts: unable to write json for text to speech: %v", err)
			return err
		}
	case internal_type.LLMMessagePacket:
		return nil
	default:
		return fmt.Errorf("elevenlab-tts: unsupported input type %T", in)
	}
	return nil
}

func (t *elevenlabsTTS) Close(ctx context.Context) error {
	t.ctxCancel()
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connection != nil {
		t.connection.Close()
		t.connection = nil
	}
	return nil
}
