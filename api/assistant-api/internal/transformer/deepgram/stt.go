// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_deepgram

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"sync"

	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	client "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/listen"
	internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
	internal_transformer_deepgram_internal "github.com/rapidaai/api/assistant-api/internal/transformer/deepgram/internal"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/protos"
)

type deepgramSTT struct {
	*deepgramOption
	mu      sync.Mutex
	ctx     context.Context
	logger  commons.Logger
	client  *client.WSCallback
	options *internal_transformer.SpeechToTextInitializeOptions
}

func (*deepgramSTT) Name() string {
	return "deepgram-speech-to-text"
}

func NewDeepgramSpeechToText(ctx context.Context,
	logger commons.Logger,
	vaultCredential *protos.VaultCredential,
	opts *internal_transformer.SpeechToTextInitializeOptions,
) (internal_transformer.SpeechToTextTransformer, error) {
	deepgramOpts, err := NewDeepgramOption(
		logger,
		vaultCredential,
		opts.AudioConfig,
		opts.ModelOptions,
	)
	if err != nil {
		logger.Errorf("deepgram-stt: Key from credential failed %+v", err)
		return nil, err
	}

	//
	return &deepgramSTT{
		ctx:            ctx,
		options:        opts,
		logger:         logger,
		deepgramOption: deepgramOpts,
	}, nil
}

// The `Initialize` method in the `deepgram` struct is responsible for establishing a connection to the
// Deepgram service using the WebSocket client `dg.client`.
func (dg *deepgramSTT) Initialize() error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	dgClient, err := client.NewWSUsingCallback(
		dg.ctx,
		dg.GetKey(),
		&interfaces.ClientOptions{
			APIKey:          dg.GetKey(),
			EnableKeepAlive: true,
		},
		dg.SpeechToTextOptions(),
		internal_transformer_deepgram_internal.
			NewDeepgramSttCallback(dg.logger, dg.options.OnTranscript))

	if err != nil {
		dg.logger.Errorf("deepgram-stt: unable create dg client with error %+v", err.Error())
		return err
	}
	if !dgClient.Connect() {
		dg.logger.Errorf("deepgram-stt: unable to connect to deepgram service")
		return fmt.Errorf("deepgram-stt: connection failed")
	}
	dg.client = dgClient
	dg.logger.Debugf("deepgram-stt: connection established")
	return nil
}

// Transform implements internal_transformer.SpeechToTextTransformer.
// The `Transform` method in the `deepgram` struct is taking an input audio byte array `in`, creating a
// new `bufio.Reader` from it, and then passing that reader to the `Stream` method of the `dg.client`
// WebSocket client. This method is responsible for streaming the audio data to the Deepgram service
// for transcription. If there are any errors during the streaming process, they will be returned by
// the method.
func (dg *deepgramSTT) Transform(ctx context.Context, in []byte, opts *internal_transformer.SpeechToTextOption) error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if dg.client == nil {
		return fmt.Errorf("deepgram-stt: connection is not initialized")
	}
	err := dg.client.Stream(bufio.NewReader(bytes.NewReader(in)))
	if err != nil {
		if err.Error() == "EOF" {
			return nil
		}
		dg.logger.Errorf("deepgram-stt: error while calling deepgram: %v", err)
		return fmt.Errorf("deepgram stream error: %w", err)
	}
	return err
}

func (dg *deepgramSTT) Close(ctx context.Context) error {
	if dg.client != nil {
		dg.client.Stop()
	}
	return nil
}
