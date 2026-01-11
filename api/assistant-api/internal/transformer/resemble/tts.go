// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_resemble

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
	"github.com/rapidaai/pkg/commons"
	protos "github.com/rapidaai/protos"
)

type resembleTTS struct {
	*resembleOption

	// context management
	ctx       context.Context
	ctxCancel context.CancelFunc

	// mutex for thread-safe access
	mu         sync.Mutex
	contextId  string
	connection *websocket.Conn

	logger  commons.Logger
	options *internal_transformer.TextToSpeechInitializeOptions
}

func NewResembleTextToSpeech(
	ctx context.Context,
	logger commons.Logger,
	credential *protos.VaultCredential,
	options *internal_transformer.TextToSpeechInitializeOptions,
) (internal_transformer.TextToSpeechTransformer, error) {
	rsmblOpts, err := NewResembleOption(logger, credential, options.AudioConfig, options.ModelOptions)
	if err != nil {
		logger.Errorf("resemble-tts: initializing resembleai failed %+v", err)
		return nil, err
	}

	ct, ctxCancel := context.WithCancel(ctx)
	return &resembleTTS{
		resembleOption: rsmblOpts,
		ctx:            ct,
		ctxCancel:      ctxCancel,
		logger:         logger,
		options:        options,
	}, nil
}

// Initialize implements internal_transformer.TextToSpeechTransformer.
func (rt *resembleTTS) Initialize() error {
	headers := map[string][]string{
		"Authorization": {"Bearer " + rt.GetKey()},
	}
	conn, _, err := websocket.DefaultDialer.Dial("wss://websocket.cluster.resemble.ai/stream", headers)
	if err != nil {
		rt.logger.Errorf("resemble-tts: unable to connect to websocket err: %v", err)
		return err
	}

	rt.mu.Lock()
	rt.connection = conn
	rt.mu.Unlock()

	rt.logger.Debugf("resemble-tts: connection established")
	go rt.textToSpeechCallback(conn, rt.ctx)
	return nil
}

// Name implements internal_transformer.TextToSpeechTransformer.
func (*resembleTTS) Name() string {
	return "resemble-text-to-speech"
}

func (rt *resembleTTS) textToSpeechCallback(conn *websocket.Conn, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			rt.logger.Infof("resemble-tts: context cancelled, stopping response listener")
			return
		default:
		}

		_, audioChunk, err := conn.ReadMessage()
		if err != nil {
			rt.logger.Errorf("resemble-tts: error reading from Resemble WebSocket: %v", err)
			return
		}

		var audioData map[string]interface{}
		if err := json.Unmarshal(audioChunk, &audioData); err != nil {
			rt.logger.Errorf("resemble-tts: error parsing audio chunk: %v", err)
			continue
		}

		// Handle different message types
		messageType, ok := audioData["type"].(string)
		if !ok {
			rt.logger.Errorf("resemble-tts: invalid message type format")
			continue
		}

		switch messageType {
		case "audio_end":
			rt.logger.Infof("resemble-tts: received audio_end event")
			return

		case "audio":
			payload, ok := audioData["audio_content"].(string)
			if !ok {
				rt.logger.Errorf("resemble-tts: invalid audio_content format")
				continue
			}

			rawAudioData, err := base64.StdEncoding.DecodeString(payload)
			if err != nil {
				rt.logger.Errorf("resemble-tts: error decoding base64 string: %v", err)
				continue
			}

			// Get contextId safely under lock
			rt.mu.Lock()
			contextId := rt.contextId
			rt.mu.Unlock()

			rt.options.OnSpeech(contextId, rawAudioData)

		default:
			rt.logger.Debugf("resemble-tts: received unknown message type: %s", messageType)
		}
	}
}

func (rt *resembleTTS) Transform(ctx context.Context, in string, opts *internal_transformer.TextToSpeechOption) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.connection == nil {
		return fmt.Errorf("resemble-tts: connection is not initialized")
	}

	rt.contextId = opts.ContextId

	if err := rt.connection.WriteJSON(rt.GetTextToSpeechRequest(opts.ContextId, in)); err != nil {
		rt.logger.Errorf("resemble-tts: error while writing request to websocket: %v", err)
		return err
	}

	return nil
}

func (rt *resembleTTS) Close(ctx context.Context) error {
	rt.ctxCancel()

	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.connection != nil {
		rt.connection.Close()
		rt.connection = nil
	}

	return nil
}
