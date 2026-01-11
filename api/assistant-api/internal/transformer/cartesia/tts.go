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
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
	cartesia_internal "github.com/rapidaai/api/assistant-api/internal/transformer/cartesia/internal"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/protos"
)

type cartesiaTTS struct {
	*cartesiaOption
	mu sync.Mutex
	// context management
	ctx       context.Context
	ctxCancel context.CancelFunc

	logger     commons.Logger
	connection *websocket.Conn
	options    *internal_transformer.TextToSpeechInitializeOptions
}

func NewCartesiaTextToSpeech(ctx context.Context, logger commons.Logger, credential *protos.VaultCredential, opts *internal_transformer.TextToSpeechInitializeOptions) (internal_transformer.TextToSpeechTransformer, error) {
	cartesiaOpts, err := NewCartesiaOption(logger, credential, opts.AudioConfig, opts.ModelOptions)
	if err != nil {
		logger.Errorf("intializing cartesia failed %+v", err)
		return nil, err
	}

	ct, ctxCancel := context.WithCancel(ctx)
	return &cartesiaTTS{
		cartesiaOption: cartesiaOpts,
		logger:         logger,
		ctx:            ct,
		ctxCancel:      ctxCancel,
		options:        opts,
	}, nil
}

func (ct *cartesiaTTS) Initialize() error {
	conn, _, err := websocket.DefaultDialer.Dial(ct.GetTextToSpeechConnectionString(), nil)
	if err != nil {
		ct.logger.Errorf("cartesia-stt: unable to dial %v", err)
		return err
	}
	ct.mu.Lock()
	ct.connection = conn
	ct.mu.Unlock()

	go ct.textToSpeechCallback(ct.connection, ct.ctx)
	ct.logger.Debugf("cartesia-stt: connection established")
	return nil
}

func (cst *cartesiaTTS) textToSpeechCallback(conn *websocket.Conn, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			cst.logger.Infof("cartesia-tts: context cancelled, stopping response listener")
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var payload cartesia_internal.TextToSpeechOuput
			if err := json.Unmarshal(msg, &payload); err != nil {
				cst.logger.Errorf("cartesia-tts: invalid json from cartesia error : %v", err)
				continue
			}
			if payload.Done {
				_ = cst.options.OnComplete(payload.ContextID)
				continue
			}
			if payload.Data == "" {
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(payload.Data)
			if err != nil {
				cst.logger.Error("cartesia-tts: failed to decode audio payload error: %v", err)
				continue
			}
			_ = cst.options.OnSpeech(payload.ContextID, decoded)
		}
	}
}

// Name returns the name of this transformer.
func (*cartesiaTTS) Name() string {
	return "cartesia-text-to-speech"
}

func (ct *cartesiaTTS) Transform(ctx context.Context, in string, opts *internal_transformer.TextToSpeechOption) error {
	ct.mu.Lock()
	conn := ct.connection
	ct.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("cartesia-tts: websocket connection is not initialized")
	}
	message := ct.GetTextToSpeechInput(in, map[string]interface{}{"continue": !opts.IsComplete, "context_id": opts.ContextId})
	if err := conn.WriteJSON(message); err != nil {
		return err
	}
	return nil
}

func (ct *cartesiaTTS) Close(ctx context.Context) error {
	ct.ctxCancel()

	ct.mu.Lock()
	defer ct.mu.Unlock()

	if ct.connection != nil {
		_ = ct.connection.Close()
	}
	return nil
}
