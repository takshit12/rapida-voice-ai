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
	cartesia_internal "github.com/rapidaai/api/assistant-api/internal/transformer/cartesia/internal"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type cartesiaTTS struct {
	*cartesiaOption
	mu sync.Mutex
	// context management
	ctx       context.Context
	ctxCancel context.CancelFunc

	contextId string

	logger     commons.Logger
	connection *websocket.Conn
	onPacket   func(pkt ...internal_type.Packet) error
}

func NewCartesiaTextToSpeech(ctx context.Context, logger commons.Logger, credential *protos.VaultCredential,
	audioConfig *protos.AudioConfig,
	onPacket func(pkt ...internal_type.Packet) error,
	opts utils.Option) (internal_type.TextToSpeechTransformer, error) {
	cartesiaOpts, err := NewCartesiaOption(logger, credential, audioConfig, opts)
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
		onPacket:       onPacket,
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
				_ = cst.onPacket(internal_type.TextToSpeechEndPacket{
					ContextID: payload.ContextID,
				})
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
			_ = cst.onPacket(internal_type.TextToSpeechAudioPacket{
				ContextID:  payload.ContextID,
				AudioChunk: decoded,
			})
		}
	}
}

// Name returns the name of this transformer.
func (*cartesiaTTS) Name() string {
	return "cartesia-text-to-speech"
}

func (ct *cartesiaTTS) Transform(ctx context.Context, in internal_type.LLMPacket) error {
	ct.mu.Lock()
	conn := ct.connection
	currentCtx := ct.contextId
	if in.ContextId() != ct.contextId {
		ct.contextId = in.ContextId()
	}
	ct.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("cartesia-tts: websocket connection is not initialized")
	}

	if currentCtx != in.ContextId() && currentCtx != "" {
		_ = conn.WriteJSON(map[string]interface{}{
			"context_id": currentCtx,
			"cancel":     true,
		})
	}

	switch input := in.(type) {
	case internal_type.LLMStreamPacket:
		message := ct.GetTextToSpeechInput(input.Text, map[string]interface{}{"continue": true, "context_id": ct.contextId, "max_buffer_delay_ms": "0ms"})
		if err := conn.WriteJSON(message); err != nil {
			return err
		}
	case internal_type.LLMMessagePacket:
		message := ct.GetTextToSpeechInput("", map[string]interface{}{"continue": false, "flush": true, "context_id": ct.contextId})
		if err := conn.WriteJSON(message); err != nil {
			return err
		}
	default:
		return fmt.Errorf("azure-tts: unsupported input type %T", in)
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
