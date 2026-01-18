// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_transformer_sarvam

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	sarvam_internal "github.com/rapidaai/api/assistant-api/internal/transformer/sarvam/internal"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type sarvamSpeechToText struct {
	*sarvamOption

	ctx       context.Context
	ctxCancel context.CancelFunc

	// Single mutex protects connection
	mu         sync.Mutex
	connection *websocket.Conn

	logger   commons.Logger
	onPacket func(pkt ...internal_type.Packet) error
}

// Name implements internal_transformer.SpeechToTextTransformer.
func (*sarvamSpeechToText) Name() string {
	return "sarvam-speech-to-text"
}

func NewSarvamSpeechToText(
	ctx context.Context,
	logger commons.Logger,
	credential *protos.VaultCredential,
	audioConfig *protos.AudioConfig,
	onPacket func(pkt ...internal_type.Packet) error,
	opts utils.Option,
) (internal_type.SpeechToTextTransformer, error) {

	sarvamOpts, err := NewSarvamOption(logger, credential, audioConfig, opts)
	if err != nil {
		logger.Errorf("sarvam-stt: initializing sarvam failed %+v", err)
		return nil, err
	}

	ct, ctxCancel := context.WithCancel(ctx)

	return &sarvamSpeechToText{
		ctx:          ct,
		ctxCancel:    ctxCancel,
		logger:       logger,
		sarvamOption: sarvamOpts,
		onPacket:     onPacket,
	}, nil
}

func (cst *sarvamSpeechToText) speechToTextCallback(conn *websocket.Conn, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			cst.logger.Infof("sarvam-stt: context cancelled, stopping response listener")
			return
		default:
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			cst.logger.Errorf("sarvam-stt: error reading from WebSocket: %v", err)
			return
		}

		var response sarvam_internal.SarvamSpeechToTextResponse
		if err := json.Unmarshal(msg, &response); err != nil {
			cst.logger.Errorf("sarvam-stt: failed to unmarshal response: %v", err)
			continue
		}

		switch response.Type {
		case "data":
			if transcriptionData, err := response.AsTranscription(); err == nil {
				if cst.onPacket != nil {
					cst.onPacket(
						internal_type.InterruptionPacket{Source: internal_type.InterruptionSourceWord},
						internal_type.SpeechToTextPacket{
							Script:     transcriptionData.Transcript,
							Confidence: 0.9,
							Language:   *transcriptionData.LanguageCode,
							Interim:    true,
						},
					)
				}
			}

		case "error":
			if errorData, err := response.AsError(); err == nil {
				cst.logger.Errorf(
					"sarvam-stt: error from server: %v",
					errorData,
				)
			}

		case "events":
			cst.logger.Infof(
				"sarvam-stt: event received: %s",
				string(response.Data),
			)

		default:
			cst.logger.Warnf(
				"sarvam-stt: unknown response type: %s",
				response.Type,
			)
		}
	}
}

func (cst *sarvamSpeechToText) Initialize() error {
	headers := make(map[string][]string)
	headers["Api-Subscription-Key"] = []string{cst.GetKey()}

	conn, _, err := websocket.DefaultDialer.Dial(cst.speechToTextUrl(), headers)
	if err != nil {
		return fmt.Errorf("sarvam-stt: failed to connect to Sarvam WebSocket: %w", err)
	}

	cst.mu.Lock()
	cst.connection = conn
	cst.mu.Unlock()

	go cst.speechToTextCallback(conn, cst.ctx)
	return nil
}

func (cst *sarvamSpeechToText) Transform(ctx context.Context, in internal_type.UserAudioPacket) error {

	vl, err := cst.speechToTextMessage(in.Audio)
	if err != nil {
		return fmt.Errorf("sarvam-stt: unable to encode byte to base64: %w", err)
	}

	cst.mu.Lock()
	connection := cst.connection
	cst.mu.Unlock()

	if connection == nil {
		return fmt.Errorf("sarvam-stt: websocket connection is not initialized")
	}

	if err := connection.WriteMessage(websocket.TextMessage, vl); err != nil {
		return fmt.Errorf("sarvam-stt: failed to send audio data: %w", err)
	}

	return nil
}

func (cst *sarvamSpeechToText) Close(ctx context.Context) error {
	cst.ctxCancel()
	cst.mu.Lock()
	defer cst.mu.Unlock()

	if cst.connection != nil {
		if err := cst.connection.Close(); err != nil {
			cst.logger.Errorf("sarvam-stt: error closing websocket connection: %w", err)
		}
		cst.connection = nil
		cst.logger.Info("sarvam-stt: websocket connection closed")
	}

	return nil
}
