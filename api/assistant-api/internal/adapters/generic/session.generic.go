// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_adapter_generic

import (
	"context"
	"fmt"
	"sync"
	"time"

	internal_adapter_request_customizers "github.com/rapidaai/api/assistant-api/internal/adapters/customizers"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_telemetry "github.com/rapidaai/api/assistant-api/internal/telemetry"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/types"
	type_enums "github.com/rapidaai/pkg/types/enums"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Disconnect handles the entire disconnection lifecycle for a conversation,
// including closing listeners, speakers, persisting recordings, and exporting metrics.
func (talking *GenericRequestor) Disconnect() {
	ctx, span, _ := talking.Tracer().StartSpan(talking.Context(), utils.AssistantDisconnectStage)
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(1)
	utils.Go(talking.Context(), func() {
		defer wg.Done()
		metrics := []*types.Metric{
			{
				Name:        type_enums.TIME_TAKEN.String(),
				Value:       fmt.Sprintf("%d", int64(time.Since(talking.StartedAt))),
				Description: "Time taken to complete conversation",
			},
			{
				Name:        type_enums.STATUS.String(),
				Value:       type_enums.RECORD_COMPLETE.String(),
				Description: "Status of the given conversation",
			},
		}
		talking.AddMetrics(talking.Auth(), metrics...)
	})
	wg.Add(1)
	utils.Go(talking.Context(), func() {
		defer wg.Done()
		if err := talking.CloseListener(ctx); err != nil {
			talking.logger.Tracef(ctx, "unable to cleanup input transformer with error %+v", err)
			return
		}
	})
	wg.Add(1)
	utils.Go(talking.Context(), func() {
		defer wg.Done()
		if err := talking.CloseSpeaker(); err != nil {
			talking.logger.Tracef(ctx, "unable to cleanup input transformer with error %+v", err)
			return
		}
	})
	wg.Wait()
	talking.OnEndConversation()
	utils.Go(talking.Context(), func() {
		byt, err := talking.recorder.Persist()
		if err != nil {
			talking.logger.Tracef(ctx, "unable to persist the audio %+v", err)
			return
		}
		if err = talking.CreateConversationRecording(byt); err != nil {
			talking.logger.Tracef(ctx, "unable to create conversation recording %+v", err)
			return
		}
	})
	span.EndSpan(ctx, utils.AssistantDisconnectStage)
	if err := talking.tracer.Export(ctx, talking.auth, &internal_telemetry.VoiceAgentExportOption{
		AssistantId:              talking.assistant.Id,
		AssistantProviderModelId: talking.assistant.AssistantProviderId,
		AssistantConversationId:  talking.assistantConversation.Id,
	}); err != nil {
		talking.logger.Errorf("error while exporting telementry %v", err)
	}
	if err := talking.assistantExecutor.Close(ctx, talking); err != nil {
		talking.logger.Errorf("error while closing assistant executor %v", err)
	}

	if talking.idealTimeoutTimer != nil {
		talking.idealTimeoutTimer.Stop()
	}
	talking.logger.Benchmark("talking.OnEndSession", time.Since(start))
}

// Connect initializes a new assistant session or resumes an existing one, based on the provided conversation configuration.
func (talking *GenericRequestor) Connect(ctx context.Context, iAuth types.SimplePrinciple, identifier string, req *protos.AssistantConversationConfiguration) error {
	ctx, span, err := talking.Tracer().StartSpan(ctx, utils.AssistantConnectStage)
	defer span.EndSpan(ctx, utils.AssistantConnectStage)
	customization, err := internal_adapter_request_customizers.NewRequestBaseCustomizer(req)
	if err != nil {
		talking.logger.Errorf("unable to initialize customizer %+v", err)
		return err
	}
	talking.SetAuth(iAuth)
	assistant, err := talking.GetAssistant(iAuth, req.Assistant.AssistantId, req.Assistant.Version)

	if err != nil {
		talking.logger.Errorf("unable to initialize assistant %+v", err)
		return err
	}

	if req.GetAssistantConversationId() > 0 {
		span.AddAttributes(ctx, internal_telemetry.KV{
			K: "conversation_initiation",
			V: internal_telemetry.StringValue("resume"),
		}, internal_telemetry.KV{
			K: "conversation_id",
			V: internal_telemetry.IntValue(req.GetAssistantConversationId()),
		})
		return talking.OnResumeSession(ctx, req.GetInputConfig(), req.GetOutputConfig(), assistant, identifier, req.GetAssistantConversationId(), customization)
	}
	span.AddAttributes(ctx, internal_telemetry.KV{
		K: "conversation_initiation",
		V: internal_telemetry.StringValue("new"),
	})
	return talking.OnCreateSession(ctx, req.GetInputConfig(), req.GetOutputConfig(), assistant, identifier, customization)
}

// OnCreateSession initializes a new assistant session, sets up listeners and speakers,
// starts voice recording, and sends configuration notifications.
func (talking *GenericRequestor) OnCreateSession(ctx context.Context, inCfg, strmCfg *protos.StreamConfig, assistant *internal_assistant_entity.Assistant, identifier string, customization internal_type.Customization) error {
	ctx, span, err := talking.Tracer().StartSpan(ctx, utils.AssistantCreateConversationStage)
	defer span.EndSpan(ctx, utils.AssistantCreateConversationStage)
	wg, _ := errgroup.WithContext(ctx)

	conversation, err := talking.BeginConversation(talking.Auth(), assistant, type_enums.DIRECTION_INBOUND, identifier, customization.GetArgs(), customization.GetMetadata(), customization.GetOptions())
	if err != nil {
		talking.logger.Errorf("unable to begin convsersation %+v", err)
		return err
	}

	audioInConfig := inCfg.GetAudio()
	if audioInConfig != nil {
		talking.messaging.SwitchInputMode(type_enums.AudioMode)
	}

	audioOutConfig := strmCfg.GetAudio()
	if audioOutConfig != nil {
		talking.messaging.SwitchOutputMode(type_enums.AudioMode)
	}
	//

	wg.Go(func() error {
		if err := talking.assistantExecutor.Initialize(ctx, talking); err != nil {
			talking.logger.Tracef(ctx, "unable to init executor %+v", err)
			return err
		}
		return nil
	})

	wg.Go(func() error {
		if err := talking.Notify(ctx,
			&protos.AssistantConversationConfiguration{
				AssistantConversationId: conversation.Id,
				Assistant: &protos.AssistantDefinition{
					AssistantId: assistant.Id,
					Version:     fmt.Sprintf("vrsn_%d", assistant.AssistantProviderId),
				},
				Time: timestamppb.Now(),
			},
		); err != nil {
			talking.logger.Errorf("Error sending configuration: %v\n", err)
		}
		return nil
	})

	utils.Go(ctx, func() {
		if audioInConfig != nil && audioOutConfig != nil {
			if err := talking.recorder.Initialize(audioInConfig, audioOutConfig); err != nil {
				talking.logger.Tracef(ctx, "unable to init recorder %+v", err)
			}
		}
	})

	wg.Go(func() error {
		if audioOutConfig != nil {
			if err := talking.ConnectSpeaker(ctx, audioOutConfig); err != nil {
				talking.logger.Tracef(ctx, "unable to connect speaker %+v", err)
			}
		}
		if err := talking.InitializeBehavior(ctx); err != nil {
			talking.logger.Errorf("unable to greet user with error %+v", err)
		}
		return nil
	})

	// establish listener
	utils.Go(ctx, func() {
		if audioInConfig != nil {
			if err := talking.ConnectListener(ctx, audioInConfig); err != nil {
				talking.logger.Tracef(ctx, "unable to init analyzer %+v", err)
			}
		}
	})

	utils.Go(ctx, func() {
		talking.AddMetrics(talking.Auth(), &types.Metric{
			Name:        type_enums.STATUS.String(),
			Value:       type_enums.RECORD_IN_PROGRESS.String(),
			Description: "Status of the given conversation",
		})
	})

	utils.Go(ctx, func() {
		if client := types.GetClientInfoFromGrpcContext(ctx); client != nil {
			if clj, err := client.ToJson(); err == nil {
				talking.SetMetadata(talking.Auth(), map[string]interface{}{"talk.client_information": clj})
			}
		}
	})

	utils.Go(ctx, func() {
		if err := talking.OnBeginConversation(); err != nil {
			talking.logger.Errorf("error while begin conversation error %+v", err)
		}
	})
	return wg.Wait()
}

// OnResumeSession resumes an existing assistant session, re-initializes listeners and speakers,
// and sends configuration notifications while also restoring ongoing conversation details.
func (talking *GenericRequestor) OnResumeSession(ctx context.Context, inCfg, strmCfg *protos.StreamConfig, assistant *internal_assistant_entity.Assistant, identifier string, assistantConversationId uint64, customization internal_type.Customization) error {
	ctx, span, err := talking.Tracer().StartSpan(talking.Context(), utils.AssistantResumeConverstaionStage)
	defer span.EndSpan(ctx, utils.AssistantResumeConverstaionStage)
	wg, _ := errgroup.WithContext(ctx)

	// resume the conversation
	conversation, err := talking.ResumeConversation(talking.Auth(), assistant, assistantConversationId, identifier)
	if err != nil {
		talking.logger.Errorf("unable to resume convsersation %+v", err)
		return err
	}

	audioInConfig := inCfg.GetAudio()
	if audioInConfig != nil {
		talking.messaging.SwitchInputMode(type_enums.AudioMode)
	}

	audioOutConfig := strmCfg.GetAudio()
	if audioOutConfig != nil {
		talking.messaging.SwitchOutputMode(type_enums.AudioMode)
	}

	// important to handle the llm request
	// need to initilize before
	wg.Go(func() error {
		if err := talking.assistantExecutor.Initialize(ctx, talking); err != nil {
			talking.logger.Tracef(ctx, "unable to init executor %+v", err)
		}
		return nil
	})

	// notify the configuration
	wg.Go(func() error {
		if err := talking.Notify(ctx, &protos.AssistantConversationConfiguration{
			AssistantConversationId: conversation.Id,
			Assistant: &protos.AssistantDefinition{
				AssistantId: assistant.Id,
				Version:     fmt.Sprintf("vrsn_%d", assistant.AssistantProviderId),
			},
			Time: timestamppb.Now(),
		}); err != nil {
			talking.logger.Errorf("Error sending configuration: %v\n", err)
		}
		return nil
	})

	utils.Go(ctx, func() {
		if audioOutConfig != nil && audioInConfig != nil {
			if err := talking.recorder.Initialize(audioInConfig, audioOutConfig); err != nil {
				talking.logger.Tracef(ctx, "unable to init recorder %+v", err)
			}
		}
	})

	// if voice then we need to support reconnecting the speaker
	wg.Go(func() error {
		if audioOutConfig != nil {
			if err := talking.ConnectSpeaker(ctx, audioOutConfig); err != nil {
				talking.logger.Tracef(ctx, "unable to connect speaker %+v", err)
			}
			// changing to audio mode
			talking.messaging.SwitchOutputMode(type_enums.AudioMode)
		}
		if err := talking.InitializeBehavior(ctx); err != nil {
			talking.logger.Errorf("unable to greet user with error %+v", err)
		}
		return nil

	})

	// establish listener
	utils.Go(ctx, func() {
		if audioInConfig != nil {
			if err := talking.ConnectListener(ctx, audioInConfig); err != nil {
				talking.logger.Tracef(ctx, "unable to init analyzer %+v", err)
			}
		}
	})

	utils.Go(ctx, func() {
		client := types.GetClientInfoFromGrpcContext(ctx)
		if client != nil {
			if clj, err := client.ToJson(); err == nil {
				talking.SetMetadata(talking.Auth(), map[string]interface{}{
					"talk.client_information": clj,
				})
			}
		}
	})

	utils.Go(ctx, func() {
		talking.AddMetrics(
			talking.Auth(),
			&types.Metric{
				Name:        type_enums.STATUS.String(),
				Value:       type_enums.RECORD_IN_PROGRESS.String(),
				Description: "Status of the given conversation",
			})
	})

	if err := talking.OnResumeConversation(); err != nil {
		talking.logger.Errorf("Error while resume the conversation: %v", err)
	}
	return wg.Wait()
}
