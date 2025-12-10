// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_sarvam

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/dvonthenen/websocket"
	internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/protos"
)

type sarvamTextToSpeech struct {
	*sarvamOption
	ctx       context.Context
	mu        sync.Mutex
	contextId string

	logger     commons.Logger
	connection *websocket.Conn
	options    *internal_transformer.TextToSpeechInitializeOptions
}

func NewSarvamTextToSpeech(
	ctx context.Context,
	logger commons.Logger,
	credential *protos.VaultCredential,
	opts *internal_transformer.TextToSpeechInitializeOptions) (internal_transformer.TextToSpeechTransformer, error) {
	sarvamOpts, err := NewSarvamOption(logger,
		credential,
		opts.AudioConfig,
		opts.ModelOptions)
	if err != nil {
		logger.Errorf("sarvam-stt: intializing sarvam failed %+v", err)
		return nil, err
	}

	return &sarvamTextToSpeech{
		ctx:          ctx,
		logger:       logger,
		sarvamOption: sarvamOpts,
		options:      opts,
	}, nil
}

// Initialize implements internal_transformer.OutputAudioTransformer.
func (rt *sarvamTextToSpeech) Initialize() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	headers := map[string][]string{
		"Api-Subscription-Key": {rt.GetKey()},
	}
	conn, _, err := websocket.DefaultDialer.Dial(rt.textToSpeechUrl(), headers)
	if err != nil {
		rt.logger.Errorf("sarvam-tts: unable to connect to websocket err: %v", err)
		return err
	}
	rt.connection = conn
	if err := rt.connection.WriteJSON(rt.configureTextToSpeech()); err != nil {
		rt.logger.Errorf("sarvam-tts: error sending configuration: %v", err)
		return err
	}

	go rt.textToSpeechCallback(rt.ctx)
	return nil
}

// Name implements internal_transformer.OutputAudioTransformer.
func (*sarvamTextToSpeech) Name() string {
	return "sarvam-text-to-speech"
}

func (rt *sarvamTextToSpeech) textToSpeechCallback(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			rt.logger.Infof("sarvam-tts: context cancelled, stopping response listener")
			return
		default:
			_, audioChunk, err := rt.connection.ReadMessage()
			if err != nil {
				rt.logger.Errorf("sarvam-tts: error reading from WebSocket: %v", err)
				return
			}

			var response map[string]interface{}
			if err := json.Unmarshal(audioChunk, &response); err != nil {
				rt.logger.Errorf("sarvam-tts: error parsing response chunk: %v", err)
				continue
			}

			// Handle different message types based on AsyncAPI spec
			switch response["type"] {
			case "audio":
				audioData, ok := response["data"].(map[string]interface{})
				if !ok {
					rt.logger.Errorf("sarvam-tts: invalid audio data format")
					continue
				}
				payload, ok := audioData["audio"].(string)
				if !ok {
					continue
				}
				rawAudioData, err := base64.StdEncoding.DecodeString(payload)
				if err != nil {
					rt.logger.Errorf("sarvam-tts: error decoding audio data: %v", err)
					continue
				}
				rt.options.OnSpeech(rt.contextId, rawAudioData)
			case "event":
				eventData := response["data"]
				rt.logger.Infof("sarvam-tts: received event data: %v", eventData)
			case "error":
				rt.logger.Errorf("sarvam-tts: received error response: %v", response)
			}
		}
	}
}

func (rt *sarvamTextToSpeech) Transform(ctx context.Context, in string, opts *internal_transformer.TextToSpeechOption) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.connection == nil {
		return fmt.Errorf("sarvam-tts: websocket connection is not initialized")
	}
	rt.contextId = opts.ContextId
	if opts.IsComplete {
		flushMsg := map[string]interface{}{
			"type": "flush",
		}
		rt.logger.Debugf("sending request flush %v", flushMsg)
		if err := rt.connection.WriteJSON(flushMsg); err != nil {
			rt.logger.Errorf("sarvam-tts: error sending flush signal to websocket: %v", err)
			return err
		}
		return nil
	}
	textMsg := map[string]interface{}{
		"type": "text",
		"data": map[string]interface{}{
			"text": in,
		},
	}
	if err := rt.connection.WriteJSON(textMsg); err != nil {
		rt.logger.Errorf("sarvam-tts: error writing text message to websocket: %v", err)
		return err
	}
	return nil
}

func (rt *sarvamTextToSpeech) Close(ctx context.Context) error {
	if rt.connection != nil {
		rt.connection.Close()
	}
	return nil
}
