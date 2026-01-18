// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_assemblyai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	assemblyai_internal "github.com/rapidaai/api/assistant-api/internal/transformer/assembly-ai/internal"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type assemblyaiSTT struct {
	*assemblyaiOption

	// context management
	ctx       context.Context
	ctxCancel context.CancelFunc

	// mutex for thread-safe access
	mu         sync.Mutex
	connection *websocket.Conn
	logger     commons.Logger
	onPacket   func(pkt ...internal_type.Packet) error
}

func NewAssemblyaiSpeechToText(ctx context.Context, logger commons.Logger, credential *protos.VaultCredential,
	audioConfig *protos.AudioConfig,
	onPacket func(pkt ...internal_type.Packet) error,
	opts utils.Option,
) (internal_type.SpeechToTextTransformer, error) {
	ayOptions, err := NewAssemblyaiOption(
		logger,
		credential,
		audioConfig,
		opts,
	)
	if err != nil {
		logger.Errorf("assembly-ai-stt: key from credential failed %v", err)
		return nil, err
	}
	ct, ctxCancel := context.WithCancel(ctx)
	return &assemblyaiSTT{
		ctx:              ct,
		ctxCancel:        ctxCancel,
		logger:           logger,
		assemblyaiOption: ayOptions,
		onPacket:         onPacket,
	}, nil
}

func (aai *assemblyaiSTT) Name() string {
	return "assemblyai-speech-to-text"
}

func (aai *assemblyaiSTT) Initialize() error {
	headers := http.Header{}
	headers.Set("Authorization", aai.GetKey())
	dialer := websocket.Dialer{
		Proxy:            nil,              // Skip proxy for direct connection
		HandshakeTimeout: 10 * time.Second, // Reduced handshake timeout for quick failover
	}

	connection, _, err := dialer.Dial(aai.GetSpeechToTextConnectionString(), headers)
	if err != nil {
		aai.logger.Errorf("assembly-ai-stt: failed to connect to websocket: %v", err)
		return fmt.Errorf("failed to connect to assemblyai websocket: %w", err)
	}

	aai.mu.Lock()
	aai.connection = connection
	aai.mu.Unlock()

	aai.logger.Debugf("assembly-ai-stt: connection established")
	go aai.speechToTextCallback(connection, aai.ctx)
	return nil
}

func (aai *assemblyaiSTT) speechToTextCallback(conn *websocket.Conn, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			aai.logger.Infof("assembly-ai-stt: read goroutine exiting due to context cancellation")
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				aai.logger.Errorf("assembly-ai-stt: read error: %v", err)
				return
			}

			var transcript assemblyai_internal.TranscriptMessage
			if err := json.Unmarshal(msg, &transcript); err != nil {
				aai.logger.Errorf("assembly-ai-stt: error unmarshalling transcript: %v", err)
				continue
			}

			switch transcript.Type {
			case "Turn":
				if len(transcript.Words) == 0 {
					aai.logger.Warnf("assembly-ai-stt: received Turn message with no words")
					continue
				}

				confidence := 0.0
				for _, v := range transcript.Words {
					confidence += v.Confidence
				}
				averageConfidence := confidence / float64(len(transcript.Words))
				aai.onPacket(
					internal_type.InterruptionPacket{Source: internal_type.InterruptionSourceWord},
					internal_type.SpeechToTextPacket{
						Script:     transcript.Transcript,
						Language:   "en",
						Confidence: averageConfidence,
						Interim:    !transcript.EndOfTurn,
					})

			case "Begin":
				aai.logger.Debugf("assembly-ai-stt: received Begin message")

			default:
				aai.logger.Debugf("assembly-ai-stt: received unknown message type: %s", transcript.Type)
			}
		}

	}
}

func (aai *assemblyaiSTT) Transform(ctx context.Context, in internal_type.UserAudioPacket) error {
	aai.mu.Lock()
	defer aai.mu.Unlock()

	if aai.connection == nil {
		return fmt.Errorf("assembly-ai-stt: websocket connection is not initialized")
	}

	if err := aai.connection.WriteMessage(websocket.BinaryMessage, in.Content()); err != nil {
		aai.logger.Errorf("assembly-ai-stt: error sending audio: %v", err)
		return fmt.Errorf("error sending audio: %w", err)
	}

	return nil
}

func (aai *assemblyaiSTT) Close(ctx context.Context) error {
	aai.ctxCancel()

	aai.mu.Lock()
	defer aai.mu.Unlock()

	if aai.connection != nil {
		aai.logger.Debugf("assembly-ai-stt: closing websocket connection")
		err := aai.connection.Close()
		aai.connection = nil
		return err
	}

	return nil
}
