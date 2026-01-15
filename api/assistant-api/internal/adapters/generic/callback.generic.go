// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_adapter_generic

import (
	"context"
	"errors"
	"time"

	internal_adapter_request_customizers "github.com/rapidaai/api/assistant-api/internal/adapters/customizers"
	internal_telemetry "github.com/rapidaai/api/assistant-api/internal/telemetry"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	type_enums "github.com/rapidaai/pkg/types/enums"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (talking *GenericRequestor) callEndOfSpeech(ctx context.Context, vl internal_type.Packet) error {
	if talking.endOfSpeech != nil {
		utils.Go(ctx, func() {
			if err := talking.endOfSpeech.Analyze(ctx, vl); err != nil {
				talking.logger.Errorf("end of speech analyze error: %v", err)
			}

		})
		return nil
	}
	return errors.New("end of speech analyzer not configured")
}

/**/
func (talking *GenericRequestor) OnPacket(ctx context.Context, pkt ...internal_type.Packet) error {
	for _, p := range pkt {
		switch vl := p.(type) {
		case internal_type.UserTextPacket:
			if talking.endOfSpeech != nil {
				talking.callEndOfSpeech(ctx, vl)
				continue
			}
			// end of speech not configured so directly send end of speech packet
			talking.OnPacket(ctx, internal_type.EndOfSpeechPacket{ContextID: vl.ContextID, Speech: vl.Text})
			continue
		case internal_type.StaticPacket:
			utils.Go(ctx, func() {
				// create a message for the static packet
				if err := talking.OnCreateMessage(ctx, vl); err != nil {
					talking.logger.Errorf("Error in OnCreateMessage: %v", err)
				}
			})
			// notify the user about static packet
			if err := talking.tokenizer.Tokenize(
				ctx,
				internal_type.TextPacket{ContextID: vl.ContextId(), Text: vl.Text},
				internal_type.FlushPacket{ContextID: vl.ContextId()}); err != nil {
				talking.logger.Debugf("unable to send static packet to tokenizer %v", err)
			}

			// sending static packat to executor for any post processing
			talking.assistantExecutor.Execute(ctx, talking, vl)
			//transition to completed
			talking.messaging.Transition(internal_adapter_request_customizers.AgentCompleted)
			continue
		case internal_type.InterruptionPacket:

			ctx, span, _ := talking.Tracer().StartSpan(talking.Context(), utils.AssistantUtteranceStage)
			defer span.EndSpan(ctx, utils.AssistantUtteranceStage)

			// notify end of speech first
			talking.callEndOfSpeech(ctx, vl)
			switch vl.Source {
			case "word":
				// user had spoken reset the timer
				span.AddAttributes(ctx, internal_telemetry.KV{K: "activity_type", V: internal_telemetry.StringValue("word_interrupt")})

				//
				talking.ResetIdealTimeoutTimer(talking.Context())
				//
				if err := talking.messaging.Transition(internal_adapter_request_customizers.Interrupted); err != nil {
					continue
				}

				talking.Notify(ctx, &protos.AssistantConversationInterruption{Type: protos.AssistantConversationInterruption_INTERRUPTION_TYPE_WORD, Time: timestamppb.Now()})
			default:
				// might be noise at first
				if vl.StartAt < 3 {
					talking.logger.Warn("interrupt: very early interruption")
					continue
				}
				span.AddAttributes(ctx, internal_telemetry.KV{K: "activity_type", V: internal_telemetry.StringValue("vad_interrupt")})
				if err := talking.messaging.Transition(internal_adapter_request_customizers.Interrupt); err != nil {
					continue
				}
				talking.Notify(ctx, &protos.AssistantConversationInterruption{Type: protos.AssistantConversationInterruption_INTERRUPTION_TYPE_VAD, Time: timestamppb.Now()})
			}
			// recorder interrupted
			if talking.messaging.GetInputMode().Audio() {
				talking.recorder.Interrupt()
			}
			continue
		case internal_type.SpeechToTextPacket:
			ctx, span, _ := talking.Tracer().StartSpan(talking.Context(), utils.AssistantListeningStage,
				internal_telemetry.KV{
					K: "transcript",
					V: internal_telemetry.StringValue(vl.Script),
				}, internal_telemetry.KV{
					K: "confidence",
					V: internal_telemetry.FloatValue(vl.Confidence),
				}, internal_telemetry.KV{
					K: "isCompleted",
					V: internal_telemetry.BoolValue(!vl.Interim),
				})
			defer span.EndSpan(ctx, utils.AssistantListeningStage)
			//

			msi := talking.messaging.Create(type_enums.UserActor, "")
			if !vl.Interim {
				msi = talking.messaging.Create(type_enums.UserActor, vl.Script)
				talking.Notify(ctx, &protos.AssistantConversationUserMessage{Id: msi.GetId(), Message: &protos.AssistantConversationUserMessage_Text{Text: &protos.AssistantConversationMessageTextContent{Content: msi.String()}}, Completed: false, Time: timestamppb.New(time.Now())})
			}

			// send to end of speech analyzer
			if err := talking.callEndOfSpeech(ctx, vl); err != nil {
				if !vl.Interim {
					talking.OnPacket(ctx, internal_type.EndOfSpeechPacket{ContextID: vl.ContextID, Speech: msi.String()})
				}
			}
			continue

		case internal_type.EndOfSpeechPacket:
			ctx, span, _ := talking.Tracer().StartSpan(talking.Context(), utils.AssistantUtteranceStage)
			span.EndSpan(ctx,
				utils.AssistantUtteranceStage,
				internal_telemetry.KV{
					K: "activity_type",
					V: internal_telemetry.StringValue("SpeechEndActivity"),
				},
				internal_telemetry.KV{
					K: "speech",
					V: internal_telemetry.StringValue(vl.Speech),
				},
			)

			msg, err := talking.messaging.GetMessage(type_enums.UserActor)
			if err != nil {
				talking.logger.Tracef(ctx, "illegal message state with error %v", err)
				continue
			}
			//
			if err := talking.Notify(ctx,
				&protos.AssistantConversationUserMessage{Id: msg.GetId(), Message: &protos.AssistantConversationUserMessage_Text{Text: &protos.AssistantConversationMessageTextContent{Content: msg.String()}}, Completed: true, Time: timestamppb.New(time.Now())}); err != nil {
				talking.logger.Tracef(ctx, "might be returing processing the duplicate message so cut it out.")
				continue
			}

			//
			talking.messaging.Transition(internal_adapter_request_customizers.UserCompleted)
			utils.Go(ctx, func() {
				if err := talking.OnCreateMessage(ctx, internal_type.UserTextPacket{ContextID: msg.GetId(), Text: msg.String()}); err != nil {
					talking.logger.Errorf("Error in OnCreateMessage: %v", err)
				}
			})

			//
			talking.messaging.Transition(internal_adapter_request_customizers.LLMGenerating)
			if err := talking.assistantExecutor.Execute(ctx, talking, internal_type.UserTextPacket{ContextID: msg.GetId(), Text: msg.String()}); err != nil {
				talking.logger.Errorf("assistant executor error: %v", err)
				talking.OnError(ctx, msg.GetId())
				continue
			}
		case internal_type.LLMStreamPacket:
			// bot had spoken reset the timer
			talking.ResetIdealTimeoutTimer(talking.Context())
			talking.tokenizer.Tokenize(ctx, internal_type.TextPacket{ContextID: vl.ContextID, Text: vl.Text})

		case internal_type.LLMPacket:
			talking.ResetIdealTimeoutTimer(talking.Context())
			utils.Go(ctx, func() {
				if err := talking.OnCreateMessage(ctx, vl); err != nil {
					talking.logger.Errorf("Error in OnCreateMessage: %v", err)
				}
			})

			if err := talking.tokenizer.Tokenize(ctx, internal_type.FlushPacket{ContextID: vl.ContextID}); err != nil {
				talking.logger.Warnf("unable to send finish speaking sentence to tokenizer %v", err)
			}

			talking.messaging.Transition(internal_adapter_request_customizers.AgentCompleted)
			continue
		case internal_type.LLMToolPacket:
			talking.
				Notify(
					ctx,
					&protos.AssistantMessagingResponse_Action{
						Action: &protos.AssistantConversationAction{
							Name:   vl.ContextID,
							Action: vl.Action,
						},
					},
				)
			continue
		case internal_type.MetricPacket:
			// metrics update for the message
			// later this can be used at each stage to calculate various metrics
			if len(vl.Metrics) > 0 {
				if err := talking.OnMessageMetric(talking.Context(), vl.ContextID, vl.Metrics); err != nil {
					talking.logger.Errorf("Error in OnUpdateMessage: %v", err)
				}
			}
		case internal_type.TextToSpeechFlushPacket:
			// notify the user about completion of tts
			if err := talking.Notify(talking.Context(), &protos.AssistantConversationAssistantMessage{Time: timestamppb.Now(), Id: vl.ContextID, Completed: true}); err != nil {
				talking.logger.Tracef(talking.ctx, "error while outputing chunk to the user: %w", err)
			}
			continue
		case internal_type.TextToSpeechPacket:
			inputMessage, err := talking.messaging.GetMessage(type_enums.UserActor)
			if err != nil {
				continue
			}
			// //
			if vl.ContextID != inputMessage.GetId() {
				continue
			}

			if err := talking.messaging.Transition(internal_adapter_request_customizers.AgentSpeaking); err != nil {
				continue
			}

			if err := talking.Notify(talking.Context(), &protos.AssistantConversationAssistantMessage{Time: timestamppb.Now(), Id: vl.ContextID, Message: &protos.AssistantConversationAssistantMessage_Audio{Audio: &protos.AssistantConversationMessageAudioContent{Content: vl.AudioChunk}}}); err != nil {
				talking.logger.Tracef(talking.ctx, "error while outputing chunk to the user: %w", err)
			}
			//
			utils.Go(context.Background(), func() {
				talking.recorder.System(vl.AudioChunk)
			})
			continue
		default:
			talking.logger.Warnf("unknown packet type received in OnGeneration %T", vl)
		}
	}
	return nil
}
