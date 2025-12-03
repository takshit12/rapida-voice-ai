// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_deepgram

import (
	"context"
	"errors"
	"sync"

	"github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	client "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/speak"
	internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
	"github.com/rapidaai/pkg/commons"
	protos "github.com/rapidaai/protos"
)

type DeepgramSpeaking interface {
	Speak(string) error
	Flush() error
	Reset() error
	Connect() bool
}

type deepgramTTS struct {
	logger    commons.Logger
	client    DeepgramSpeaking
	contextId string
	mu        sync.Mutex

	providerOption    DeepgramOption
	transformerOption *internal_transformer.TextToSpeechInitializeOptions
}

func NewDeepgramTextToSpeech(
	ctx context.Context,
	logger commons.Logger,
	credential *protos.VaultCredential,
	opts *internal_transformer.TextToSpeechInitializeOptions) (internal_transformer.TextToSpeechTransformer, error) {

	//create deepgram option
	dGoptions, err := NewDeepgramOption(
		logger,
		credential,
		opts.AudioConfig,
		opts.ModelOptions,
	)
	if err != nil {
		logger.Errorf("error while intializing deepgram text to speech")
		return nil, err
	}

	dg := &deepgramTTS{
		logger:            logger,
		providerOption:    dGoptions,
		transformerOption: opts,
	}
	dg.client, err = client.NewWSUsingCallback(ctx,
		dGoptions.GetKey(),
		&interfaces.ClientOptions{
			APIKey:          dGoptions.GetKey(),
			EnableKeepAlive: true,
		},
		dGoptions.TextToSpeechOptions(),
		NewDeepgramSpeakCallback(logger, dg.onspeech, dg.oncomplete),
	)
	if err != nil {
		logger.Errorf("unable create dg client with error %+v", err.Error())
		return nil, err
	}
	return dg, nil
}

// Deepgram service using the WebSocket client `dg.client`.
func (dg *deepgramTTS) Initialize() error {
	if !dg.client.Connect() {
		return errors.New("unable to connect")
	}
	return nil
}

func (dg *deepgramTTS) onspeech(b []byte) error {
	return dg.transformerOption.OnSpeech(dg.contextId, b)
}

func (dg *deepgramTTS) oncomplete() error {
	return dg.transformerOption.OnComplete(dg.contextId)
}

func (dg *deepgramTTS) Transform(
	ctx context.Context,
	sentence string,
	opts *internal_transformer.TextToSpeechOption) error {
	dg.logger.Infof("deepgram-tts: speak %s with context id = %s and completed = %t", sentence, opts.ContextId, opts.IsComplete)
	dg.mu.Lock()
	dg.contextId = opts.ContextId
	dg.mu.Unlock()

	dg.client.Speak(sentence)
	if opts.IsComplete {
		dg.client.Flush()
	}
	return nil

}

func (dg *deepgramTTS) Close(ctx context.Context) error {
	dg.client.Reset()
	return nil
}
