// Copyright (c) Rapida
// Author: Prashant <prashant@rapida.ai>
//
// Licensed under the Rapida internal use license.
// This file is part of Rapida's proprietary software and is not open source.
// Unauthorized copying, modification, or redistribution is strictly prohibited.

package internal_transformer_google

import (
	"context"
	"io"
	"sync"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/protos"
)

type googleTextToSpeech struct {
	mu                 sync.Mutex
	ctx                context.Context
	contextId          string
	logger             commons.Logger
	client             texttospeechpb.TextToSpeech_StreamingSynthesizeClient
	providerOptions    GoogleOption
	transformerOptions *internal_transformer.TextToSpeechInitializeOptions
}

func NewGoogleTextToSpeech(
	ctx context.Context,
	logger commons.Logger,
	credential *protos.VaultCredential,
	opts *internal_transformer.TextToSpeechInitializeOptions) (internal_transformer.TextToSpeechTransformer, error) {
	cOptions, err := NewGoogleOption(logger, credential, opts.AudioConfig, opts.ModelOptions)
	if err != nil {
		logger.Errorf("intializing google failed %+v", err)
		return nil, err
	}

	client, err := texttospeech.NewClient(
		ctx,
		cOptions.GetClientOptions()...)
	if err != nil {
		logger.Errorf("error while creating client for google tts %+v", err)
		return nil, err
	}
	stream, err := client.StreamingSynthesize(ctx)
	if err != nil {
		logger.Errorf("error while creating by directional for google tts %+v", err)
		return nil, err
	}
	return &googleTextToSpeech{
		ctx:                ctx,
		logger:             logger,
		client:             stream,
		transformerOptions: opts,
		providerOptions:    cOptions,
	}, nil
}

func (g *googleTextToSpeech) Close(ctx context.Context) error {
	g.client.CloseSend()
	return nil
}

func (google *googleTextToSpeech) Initialize() error {
	google.mu.Lock()
	defer google.mu.Unlock()
	req := texttospeechpb.StreamingSynthesizeRequest{
		StreamingRequest: &texttospeechpb.
			StreamingSynthesizeRequest_StreamingConfig{
			StreamingConfig: google.providerOptions.TextToSpeechOptions(),
		},
	}
	err := google.client.Send(&req)
	if err != nil {
		google.logger.Errorf("error while intiializing google text to speech")
		return err
	}
	go google.TextToSpeechCallback(google.ctx)
	return nil

}

func (*googleTextToSpeech) Name() string {
	return "google-text-to-speech"
}

func (google *googleTextToSpeech) Transform(ctx context.Context, in string, opts *internal_transformer.TextToSpeechOption) error {
	google.logger.Infof("google-tts: speak %s with context id = %s and completed = %t", in, opts.ContextId, opts.IsComplete)
	google.mu.Lock()
	defer google.mu.Unlock()
	google.contextId = opts.ContextId
	req := texttospeechpb.StreamingSynthesizeRequest{
		StreamingRequest: &texttospeechpb.StreamingSynthesizeRequest_Input{
			Input: &texttospeechpb.StreamingSynthesisInput{
				InputSource: &texttospeechpb.StreamingSynthesisInput_Text{Text: in},
			},
		},
	}
	err := google.client.Send(&req)
	if err != nil {
		google.logger.Errorf("unable to Synthesize text %v", err)
	}
	return nil

}

// SpeechToTextCallback processes streaming responses with context awareness.
func (g *googleTextToSpeech) TextToSpeechCallback(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			g.logger.Infof("Google STT: context cancelled, stopping response listener")
			return
		default:
			resp, err := g.client.Recv()
			if err == io.EOF {
				g.logger.Infof("Google STT: stream ended (EOF)")
				continue
			}
			if err != nil {
				// reconnect
				g.logger.Errorf("Google STT: recv error: %v match it", err.Error())
				return
			}
			if resp != nil {
				g.transformerOptions.OnSpeech(
					g.contextId,
					resp.GetAudioContent())
			}

		}
	}
}
