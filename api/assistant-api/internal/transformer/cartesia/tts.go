// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_cartesia

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	protos "github.com/rapidaai/protos"
)

type cartesiaTTS struct {
	logger     commons.Logger
	connection *websocket.Conn
	ctx        context.Context
	mu         sync.Mutex
	//
	transformerOptions *internal_transformer.TextToSpeechInitializeOptions
	providerOptions    CartesiaOption
	lastUsed           time.Time
}

func NewCartesiaTextToSpeech(
	ctx context.Context,
	logger commons.Logger,
	credential *protos.VaultCredential,
	opts *internal_transformer.TextToSpeechInitializeOptions,
) (internal_transformer.TextToSpeechTransformer, error) {
	//create cartesia option
	options, err := NewCartesiaOption(logger, credential,
		opts.AudioConfig,
		opts.ModelOptions)
	if err != nil {
		logger.Errorf("intializing cartesia failed %+v", err)
		return nil, err
	}

	return &cartesiaTTS{
		logger:             logger,
		ctx:                ctx,
		lastUsed:           time.Now(),
		providerOptions:    options,
		transformerOptions: opts,
	}, nil
}

func (ct *cartesiaTTS) connectWebSocket() error {
	conn, _, err := websocket.DefaultDialer.Dial(ct.providerOptions.GetTextToSpeechConnectionString(), nil)
	if err != nil {
		return err
	}
	ct.connection = conn
	ct.logger.Info("Connected to Cartesia WebSocket")
	return nil
}

func (ct *cartesiaTTS) Initialize() error {
	if err := ct.connectWebSocket(); err != nil {
		return err
	}
	ct.lastUsed = time.Now()
	utils.Go(ct.ctx, func() {
		ct.handleAudioStream()
	})

	return nil
}

func (ct *cartesiaTTS) ensureValidConnection() error {
	if ct.connection == nil || time.Since(ct.lastUsed) > 1*time.Minute {
		if ct.connection != nil {
			_ = ct.connection.Close()
		}
		if err := ct.Initialize(); err != nil {
			return err
		}
	}
	ct.lastUsed = time.Now()
	return nil
}

func (ct *cartesiaTTS) handleAudioStream() error {
	for {
		select {
		case <-ct.ctx.Done():
			return nil
		default:
			_, msg, err := ct.connection.ReadMessage()
			if err != nil {
				return err
			}

			var payload TextToSpeechOuput
			if err := json.Unmarshal(msg, &payload); err != nil {
				ct.logger.Error("Invalid JSON from Cartesia", "error", err)
				continue
			}

			if payload.Done {
				_ = ct.transformerOptions.OnComplete(payload.ContextID)
				continue
			}

			if payload.Data == "" {
				continue
			}

			decoded, err := base64.StdEncoding.DecodeString(payload.Data)
			if err != nil {
				ct.logger.Error("Failed to decode audio payload", "error", err)
				continue
			}

			_ = ct.transformerOptions.OnSpeech(payload.ContextID, decoded)
		}
	}
}

// Name returns the name of this transformer.
func (*cartesiaTTS) Name() string {
	return "cartesia-text-to-speech"
}

func (ct *cartesiaTTS) Transform(ctx context.Context, in string, opts *internal_transformer.TextToSpeechOption) error {
	ct.logger.Infof("cartesia-tts: speak %s with context id = %s and completed = %t", in, opts.ContextId, opts.IsComplete)
	if err := ct.ensureValidConnection(); err != nil {
		return err
	}
	message := ct.providerOptions.GetTextToSpeechInput(
		in,
		map[string]interface{}{
			"continue":   !opts.IsComplete,
			"context_id": opts.ContextId,
		})

	ct.mu.Lock()         // Lock before writing to the WebSocket
	defer ct.mu.Unlock() // Ensure the lock is released when the function returns

	if err := ct.connection.WriteJSON(message); err != nil {
		return err
	}
	return nil
}

func (ct *cartesiaTTS) Close(ctx context.Context) error {
	ct.logger.Info("CartesiaTTS: Closing Cartesia connection")
	if ct.connection != nil {
		_ = ct.connection.Close()
	}
	return nil
}
