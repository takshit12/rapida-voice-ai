// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_openai

import (
	"context"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
	"github.com/rapidaai/pkg/commons"
)

type openaiSpeechToText struct {
	logger commons.Logger
	client openai.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func (o *openaiSpeechToText) Initialize() error {
	o.ctx, o.cancel = context.WithCancel(context.Background())
	o.client = openai.NewClient(option.WithAPIKey("YOUR_API_KEY"))
	return nil
}

func (o *openaiSpeechToText) Close(ctx context.Context) error {
	if o.cancel != nil {
		o.cancel()
	}
	o.logger.Infof("OpenAI SpeechToText connection closed.")
	return nil
}

func (o *openaiSpeechToText) Name() string {
	return "openai-speech-to-text"
}

// Transform receives a stream of bytes (audioStream) and prints transcribed text in realtime.
func (o *openaiSpeechToText) Transform(ctx context.Context,
	byt []byte,
	opt *internal_transformer.SpeechToTextOption) error {
	return nil
}

func NewOpenaiSpeechToText(
	ctx context.Context,
	logger commons.Logger,
	opts internal_transformer.SpeechToTextTransformer,
) (internal_transformer.SpeechToTextTransformer, error) {
	stt := &openaiSpeechToText{
		logger: logger,
	}
	return stt, nil
}
