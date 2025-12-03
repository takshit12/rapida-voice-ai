// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_cartesia

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	protos "github.com/rapidaai/protos"
)

type cartesiaSpeechToText struct {
	logger             commons.Logger
	connection         *websocket.Conn
	providerOptions    CartesiaOption
	transformerOptions *internal_transformer.SpeechToTextInitializeOptions
}

// Name implements internal_transformer.SpeechToTextTransformer.
func (*cartesiaSpeechToText) Name() string {
	return "cartesia-speech-to-text"
}

func NewCartesiaSpeechToText(ctx context.Context,
	logger commons.Logger,
	credential *protos.VaultCredential,
	transformerOptions *internal_transformer.SpeechToTextInitializeOptions,
) (internal_transformer.SpeechToTextTransformer, error) {
	cOptions, err := NewCartesiaOption(logger,
		credential,
		transformerOptions.AudioConfig,
		transformerOptions.ModelOptions)
	if err != nil {
		logger.Errorf("intializing cartesia failed %+v", err)
		return nil, err
	}

	return &cartesiaSpeechToText{
		logger:             logger,
		providerOptions:    cOptions,
		transformerOptions: transformerOptions,
	}, nil
}

func (cst *cartesiaSpeechToText) Initialize() error {
	cst.logger.Debugf("intializing cartesia %s", cst.providerOptions.GetSpeechToTextConnectionString())
	conn, _, err := websocket.DefaultDialer.Dial(cst.providerOptions.GetSpeechToTextConnectionString(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Cartesia WebSocket: %w", err)
	}

	cst.connection = conn
	cst.logger.Info("connected to Cartesia STT WebSocket")

	// Start the read/callback goroutine after connection is established
	utils.Go(context.Background(), func() {
		for {
			select {
			default:
				_, msg, err := cst.connection.ReadMessage()
				if err != nil {
					cst.logger.Error("error reading from Cartesia WebSocket: ", err)
					return
				}
				var resp SpeechToTextOutput
				if err := json.Unmarshal(msg, &resp); err == nil && resp.Text != "" {
					cst.logger.Debug("Received transcription: %+v", resp)
					if cst.transformerOptions.OnTranscript != nil {
						cst.transformerOptions.OnTranscript(
							resp.Text,
							0.9,
							resp.Language,
							resp.IsFinal,
						)
					}
				}
			}
		}
	})

	return nil
}

func (cst *cartesiaSpeechToText) Transform(ctx context.Context, in []byte, opts *internal_transformer.SpeechToTextOption) error {
	if cst.connection == nil {
		return fmt.Errorf("WebSocket connection is not initialized")
	}
	if err := cst.connection.WriteMessage(
		websocket.BinaryMessage, in); err != nil {
		return fmt.Errorf("failed to send audio data: %w", err)
	}

	return nil
}

func (cst *cartesiaSpeechToText) Close(ctx context.Context) error {
	if cst.connection != nil {
		err := cst.connection.Close()
		if err != nil {
			return fmt.Errorf("error closing WebSocket connection: %w", err)
		}
		cst.logger.Info("Cartesia WebSocket connection closed")
	}
	return nil
}
