// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_azure

import (
	"context"
	"sync"

	"github.com/Microsoft/cognitive-services-speech-sdk-go/audio"
	"github.com/Microsoft/cognitive-services-speech-sdk-go/speech"
	internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
	"github.com/rapidaai/pkg/commons"
	protos "github.com/rapidaai/protos"
)

type azureTextToSpeech struct {
	contextId   string
	mu          sync.Mutex
	logger      commons.Logger
	audioConfig *audio.AudioConfig
	client      *speech.SpeechSynthesizer

	providerOption    AzureOption
	transformerOption *internal_transformer.TextToSpeechInitializeOptions
}

func NewAzureTextToSpeech(
	ctx context.Context,
	logger commons.Logger,
	credential *protos.VaultCredential,
	iOption *internal_transformer.TextToSpeechInitializeOptions) (internal_transformer.TextToSpeechTransformer, error) {
	providerOption, err := NewAzureOption(logger, credential,
		iOption.AudioConfig,
		iOption.ModelOptions)
	if err != nil {
		logger.Errorf("Unable to initilize azure option", err)
		return nil, err
	}

	return &azureTextToSpeech{
		logger:            logger,
		providerOption:    providerOption,
		transformerOption: iOption,
	}, nil
}

func (azure *azureTextToSpeech) Name() string {
	return "azure-text-to-speech"
}

func (azure *azureTextToSpeech) Close(ctx context.Context) error {
	azure.client.Close()
	azure.audioConfig.Close()
	return nil
}

func (azure *azureTextToSpeech) Initialize() (err error) {
	stream, err := audio.CreatePullAudioOutputStream()
	if err != nil {
		return err
	}
	azure.audioConfig, err = audio.NewAudioConfigFromStreamOutput(stream)
	if err != nil {
		return err
	}

	speechConfig, err := azure.providerOption.TextToSpeechOption()
	if err != nil {
		return err
	}

	azure.client, err = speech.NewSpeechSynthesizerFromConfig(speechConfig, azure.audioConfig)
	if err != nil {
		return err
	}
	azure.client.SynthesisStarted(azure.OnStart)
	azure.client.Synthesizing(azure.OnSpeech)
	azure.client.SynthesisCompleted(azure.OnComplete)
	azure.client.SynthesisCanceled(azure.OnCancel)
	return nil
}

func (azure *azureTextToSpeech) Transform(ctx context.Context, text string, opts *internal_transformer.TextToSpeechOption) error {
	azure.mu.Lock()
	azure.contextId = opts.ContextId
	azure.mu.Unlock()
	azure.logger.Infof("azure-tts: speak %s with context id = %s and completed = %t", text, opts.ContextId, opts.IsComplete)
	res := <-azure.client.SpeakTextAsync(text)
	if res.Error != nil {
		return res.Error
	}
	return nil
}
func (azCallback *azureTextToSpeech) OnStart(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
}

func (azCallback *azureTextToSpeech) OnSpeech(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
	azCallback.transformerOption.OnSpeech(azCallback.contextId, event.Result.AudioData)
}

func (azCallback *azureTextToSpeech) OnComplete(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
	azCallback.transformerOption.OnComplete(azCallback.contextId)
}

func (azCallback *azureTextToSpeech) OnCancel(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
}
