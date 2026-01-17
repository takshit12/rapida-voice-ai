// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_adapter_generic

import (
	"context"
	"time"

	internal_denoiser "github.com/rapidaai/api/assistant-api/internal/denoiser"
	internal_end_of_speech "github.com/rapidaai/api/assistant-api/internal/end_of_speech"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"

	internal_telemetry "github.com/rapidaai/api/assistant-api/internal/telemetry"
	internal_transformer "github.com/rapidaai/api/assistant-api/internal/transformer"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	internal_vad "github.com/rapidaai/api/assistant-api/internal/vad"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
	"golang.org/x/sync/errgroup"
)

func (listening *GenericRequestor) initializeSpeechToText(ctx context.Context, transformerConfig *internal_assistant_entity.AssistantDeploymentAudio, audioConfig *protos.AudioConfig, options utils.Option) error {
	credentialId, err := options.GetUint64("rapida.credential_id")
	if err != nil {
		listening.logger.Errorf("unable to find credential from options %+v", err)
		return err
	}
	credential, err := listening.VaultCaller().GetCredential(ctx, listening.Auth(), credentialId)
	if err != nil {
		listening.logger.Errorf("Api call to find credential failed %+v", err)
		return err
	}

	atransformer, err := internal_transformer.GetSpeechToTextTransformer(internal_transformer.AudioTransformer(transformerConfig.AudioProvider), listening.Context(), listening.logger, credential, &internal_type.SpeechToTextInitializeOptions{AudioConfig: audioConfig, OnPacket: func(pkt ...internal_type.Packet) error { return listening.OnPacket(ctx, pkt...) }, ModelOptions: options})
	if err != nil {
		listening.logger.Errorf("unable to create input audio transformer with error %v", err)
		return err
	}
	err = atransformer.Initialize()
	if err != nil {
		listening.logger.Errorf("unable to initilize transformer %v", err)
		return err
	}
	listening.speechToTextTransformer = atransformer
	return nil
}

// Init initializes the audio talking system for a given assistant persona.
// It sets up both audio input and output transformer.
// This function is typically called at the beginning of a communication session.
func (listening *GenericRequestor) ConnectListener(ctx context.Context, audioConfig *protos.AudioConfig) error {
	ctx, span, _ := listening.Tracer().StartSpan(ctx, utils.AssistantListenConnectStage)
	defer span.EndSpan(ctx, utils.AssistantListenConnectStage)

	eGroup, ctx := errgroup.WithContext(ctx)
	options := map[string]interface{}{
		"microphone.eos.timeout": 500,
		"microphone.eos.enabled": true,
	}

	if audioConfig != nil {
		transformerConfig, err := listening.GetSpeechToTextTransformer()
		if err != nil {
			listening.logger.Warnf("error during getting transformer for assistant.")
		} else {
			options = utils.MergeMaps(options, transformerConfig.GetOptions())
			span.AddAttributes(ctx,
				internal_telemetry.KV{K: "options", V: internal_telemetry.JSONValue(options)},
				internal_telemetry.KV{K: "provider", V: internal_telemetry.StringValue(transformerConfig.AudioProvider)},
			)
			//
			eGroup.Go(func() error {
				err := listening.initializeSpeechToText(ctx, transformerConfig, audioConfig, options)
				if err != nil {
					listening.logger.Errorf("unable to initialize transformer %+v", err)
				}
				return nil
			})

			eGroup.Go(func() error {
				err := listening.initializeVAD(ctx, audioConfig, options)
				if err != nil {
					listening.logger.Errorf("illegal input audio transformer, check the config and re-init")
				}
				return nil
			})

			eGroup.Go(func() error {
				err := listening.initializeDenoiser(ctx, audioConfig, options)
				if err != nil {
					listening.logger.Errorf("illegal input audio transformer, check the config and re-init")
				}
				return nil
			})

		}
	}

	eGroup.Go(func() error {
		err := listening.initializeEndOfSpeech(ctx, options)
		if err != nil {
			listening.logger.Errorf("illegal input audio transformer, check the config and re-init")
		}
		return nil
	})

	if err := eGroup.Wait(); err != nil {
		return err
	}
	return nil
}

func (listening *GenericRequestor) CloseListener(ctx context.Context) error {
	if listening.speechToTextTransformer != nil {
		if err := listening.speechToTextTransformer.Close(ctx); err != nil {
			listening.logger.Warnf("cancel all output transformer with error %v", err)
		}
	}

	if listening.endOfSpeech != nil {
		if err := listening.endOfSpeech.Close(); err != nil {
			listening.logger.Warnf("cancel end of speech with error %v", err)
		}
	}

	if listening.vad != nil {
		if err := listening.vad.Close(); err != nil {
			listening.logger.Warnf("cancel vad with error %v", err)
		}
	}
	return nil
}

func (listening *GenericRequestor) initializeEndOfSpeech(
	ctx context.Context,
	options utils.Option,
) error {
	start := time.Now()
	provider, err := options.GetString("microphone.eos.provider")
	if err != nil {
		listening.logger.Errorf("denoising.provider is not set, please check the configuration")
		return err
	}
	endOfSpeech, err := internal_end_of_speech.GetEndOfSpeech(
		internal_end_of_speech.EndOfSpeechIdentifier(provider),
		listening.logger,
		func(_ctx context.Context, act internal_type.EndOfSpeechPacket) error {
			return listening.OnPacket(_ctx, act)
		},
		options)
	if err != nil {
		listening.logger.Warnf("unable to initialize text analyzer %+v", err)
		return err
	}
	listening.endOfSpeech = endOfSpeech
	listening.logger.Benchmark("listen.endOfSpeech", time.Since(start))
	return nil
}

func (listening *GenericRequestor) initializeDenoiser(ctx context.Context, audioConfig *protos.AudioConfig, options utils.Option) error {
	provider, err := options.GetString("microphone.denoising.provider")
	if err != nil {
		listening.logger.Errorf("denoising.provider is not set, please check the configuration")
		return err
	}
	denoise, err := internal_denoiser.GetDenoiser(internal_denoiser.DenoiserIdentifier(provider), listening.logger, audioConfig, options)
	if err != nil {
		listening.logger.Errorf("error wile intializing denoiser %+v", err)
	}
	listening.denoiser = denoise
	return nil
}

func (listening *GenericRequestor) initializeVAD(ctx context.Context, audioConfig *protos.AudioConfig, options utils.Option,
) error {
	start := time.Now()
	provider, err := options.GetString("microphone.vad.provider")
	if err != nil {
		listening.logger.Errorf("vad.provider is not set, please check the configuration")
		return err
	}

	vad, err := internal_vad.GetVAD(internal_vad.VADIdentifier(provider), listening.logger, audioConfig, func(vr internal_type.InterruptionPacket) error { return listening.OnPacket(listening.Context(), vr) }, options)
	if err != nil {
		listening.logger.Errorf("error wile intializing vad %+v", err)
		return err
	}
	listening.vad = vad
	listening.logger.Benchmark("listen.initializeVAD", time.Since(start))
	return nil
}
