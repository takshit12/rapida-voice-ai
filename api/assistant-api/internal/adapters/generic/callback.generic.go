// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_adapter_generic

import (
	"context"
	"errors"
	"io"
	"time"

	internal_adapter_request_customizers "github.com/rapidaai/api/assistant-api/internal/adapters/customizers"
	internal_adapter_telemetry "github.com/rapidaai/api/assistant-api/internal/telemetry"
	internal_telemetry "github.com/rapidaai/api/assistant-api/internal/telemetry"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
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

func (talking *GenericRequestor) callSentenceAssembler(ctx context.Context, vl internal_type.Packet) error {
	if talking.sentenceAssembler != nil {
		if err := talking.sentenceAssembler.Assemble(ctx, vl); err != nil {
			talking.logger.Debugf("unable to send packet to assembler %v", err)
		}
		return nil
	}
	return errors.New("sentenceAssembler not configured")
}

func (talking *GenericRequestor) callRecording(ctx context.Context, vl internal_type.Packet) error {
	if talking.recorder != nil {
		utils.Go(ctx, func() {
			if err := talking.recorder.Record(ctx, vl); err != nil {
				talking.logger.Errorf("recorder error: %v", err)
			}
		})
		return nil
	}
	return errors.New("recording not configured")
}

func (talking *GenericRequestor) callCreateMessage(ctx context.Context, vl internal_type.MessagePacket) error {
	utils.Go(ctx, func() {
		if err := talking.onCreateMessage(ctx, vl); err != nil {
			talking.logger.Errorf("Error in onCreateMessage: %v", err)
		}
	})
	return nil
}

func (talking *GenericRequestor) callVadProcess(ctx context.Context, vl internal_type.UserAudioPacket) error {
	if talking.vad != nil {
		utils.Go(ctx, func() {
			if err := talking.vad.Process(ctx, vl); err != nil {
				talking.logger.Warnf("error while processing with vad %s", err.Error())
			}
		})
	}
	return nil
}

func (talking *GenericRequestor) callSpeechToText(ctx context.Context, vl internal_type.UserAudioPacket) error {
	if talking.speechToTextTransformer != nil {
		utils.Go(ctx, func() {
			if err := talking.speechToTextTransformer.Transform(ctx, vl); err != nil {
				if !errors.Is(err, io.EOF) {
					talking.logger.Tracef(ctx, "error while transforming input %s and error %s", talking.speechToTextTransformer.Name(), err.Error())
				}
			}
		})
	}
	return nil
}

/**/
func (talking *GenericRequestor) OnPacket(ctx context.Context, pkts ...internal_type.Packet) error {
	for _, p := range pkts {
		switch vl := p.(type) {
		case internal_type.UserTextPacket:
			// calling end of speech analyzer
			interim := talking.messaging.Create(vl.Text)
			if err := talking.Notify(talking.Context(), &protos.AssistantConversationUserMessage{Id: interim.GetId(), Completed: false, Message: &protos.AssistantConversationUserMessage_Text{Text: &protos.AssistantConversationMessageTextContent{Content: interim.String()}}, Time: timestamppb.Now()}); err != nil {
				talking.logger.Tracef(talking.Context(), "error while notifying the text input from user: %w", err)
			}

			if err := talking.messaging.Transition(internal_adapter_request_customizers.UserSpeaking); err != nil {
				talking.logger.Errorf("messaging transition error: %v", err)
			}
			// send to end of speech analyzer
			vl.ContextID = interim.GetId()
			if err := talking.callEndOfSpeech(ctx, vl); err != nil {
				talking.OnPacket(ctx, internal_type.EndOfSpeechPacket{ContextID: vl.ContextID, Speech: vl.Text})
			}
			// end of speech not configured so directly send end of speech packet
			continue

		case internal_type.UserAudioPacket:
			if talking.denoiser != nil && !vl.NoiseReduced {
				vl.NoiseReduced = true
				dnOut, _, err := talking.denoiser.Denoise(ctx, vl.Audio)
				if err != nil {
					talking.logger.Warnf("error while denoising process | will process actual audio byte")
					talking.OnPacket(ctx, vl)
				} else {
					vl.Audio = dnOut
					talking.OnPacket(ctx, vl)
				}
				continue
			}

			if err := talking.callRecording(ctx, vl); err != nil {
				talking.logger.Errorf("recorder error: %v", err)
			}

			if err := talking.callVadProcess(ctx, vl); err != nil {
				talking.logger.Errorf("VAD process error: %v", err)
			}

			if err := talking.callSpeechToText(ctx, vl); err != nil {
				talking.logger.Errorf("speech to text transform error: %v", err)
			}
			continue
		case internal_type.StaticPacket:
			// when static packet is received it means that rapida system has something to speak
			// do not abrupt it just send it to the assembler
			if err := talking.callCreateMessage(ctx, vl); err != nil {
				talking.logger.Errorf("unable to create message from static packet %v", err)
			}

			// sending static packat to executor for any post processing
			if err := talking.messaging.Transition(internal_adapter_request_customizers.LLMGenerating); err != nil {
				talking.logger.Errorf("messaging transition error: %v", err)
			}

			//
			if err := talking.assistantExecutor.Execute(ctx, talking, vl); err != nil {
				talking.logger.Errorf("assistant executor error: %v", err)
			}

			if err := talking.messaging.Transition(internal_adapter_request_customizers.LLMGenerating); err != nil {
				talking.logger.Errorf("messaging transition error: %v", err)
			}

			if err := talking.callSentenceAssembler(ctx, internal_type.LLMStreamPacket{ContextID: vl.ContextId(), Text: vl.Text}); err != nil {
				talking.logger.Debugf("unable to send static packet to tokenizer %v", err)
			}

			if err := talking.messaging.Transition(internal_adapter_request_customizers.LLMGenerated); err != nil {
				talking.logger.Errorf("messaging transition error: %v", err)
			}
			if err := talking.callSentenceAssembler(ctx, internal_type.LLMMessagePacket{ContextID: vl.ContextId()}); err != nil {
				talking.logger.Debugf("unable to send static packet to tokenizer %v", err)
			}

			continue
		case internal_type.InterruptionPacket:
			ctx, span, _ := talking.Tracer().StartSpan(talking.Context(), utils.AssistantUtteranceStage)
			defer span.EndSpan(ctx, utils.AssistantUtteranceStage)

			// calling end of speech analyzer
			if err := talking.callEndOfSpeech(ctx, vl); err != nil {
				talking.logger.Errorf("end of speech error: %v", err)
			}
			//
			// recorder interrupted
			if err := talking.callRecording(ctx, vl); err != nil {
				talking.logger.Errorf("recorder error: %v", err)
			}

			switch vl.Source {
			case internal_type.InterruptionSourceWord:
				span.AddAttributes(ctx, internal_telemetry.KV{K: "activity_type", V: internal_telemetry.StringValue("word_interrupt")})
				talking.resetIdleTimeoutTimer(talking.Context())

				if err := talking.sentenceAssembler.Assemble(ctx, vl); err != nil {
					talking.logger.Debugf("unable to send interruption packet to assembler %v", err)
				}
				//
				if err := talking.messaging.Transition(internal_adapter_request_customizers.Interrupted); err != nil {
					talking.logger.Errorf("messaging transition error: %v", err)
					continue
				}
				talking.Notify(ctx, &protos.AssistantConversationInterruption{Type: protos.AssistantConversationInterruption_INTERRUPTION_TYPE_WORD, Time: timestamppb.Now()})
				continue
			default:
				// might be noise at first
				if vl.StartAt < 3 {
					talking.logger.Warn("interrupt: very early interruption")
					continue
				}
				span.AddAttributes(ctx, internal_telemetry.KV{K: "activity_type", V: internal_telemetry.StringValue("vad_interrupt")})
				if err := talking.messaging.Transition(internal_adapter_request_customizers.Interrupt); err != nil {
					talking.logger.Errorf("messaging transition error: %v", err)
					continue
				}
				talking.Notify(ctx, &protos.AssistantConversationInterruption{Type: protos.AssistantConversationInterruption_INTERRUPTION_TYPE_VAD, Time: timestamppb.Now()})
				continue
			}
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
			talking.messaging.Transition(internal_adapter_request_customizers.UserSpeaking)
			// send to end of speech analyzer
			if err := talking.callEndOfSpeech(ctx, vl); err != nil {
				if !vl.Interim {
					msi := talking.messaging.Create(vl.Script)
					talking.OnPacket(ctx, internal_type.EndOfSpeechPacket{ContextID: msi.Id, Speech: msi.String()})
				}
			}

			if !vl.Interim {
				msi := talking.messaging.Create(vl.Script)
				talking.Notify(ctx, &protos.AssistantConversationUserMessage{Id: msi.GetId(), Message: &protos.AssistantConversationUserMessage_Text{Text: &protos.AssistantConversationMessageTextContent{Content: msi.String()}}, Completed: false, Time: timestamppb.New(time.Now())})
			}
			continue

		case internal_type.EndOfSpeechPacket:
			ctx, span, _ := talking.Tracer().StartSpan(talking.Context(), utils.AssistantUtteranceStage)
			span.EndSpan(ctx,
				utils.AssistantUtteranceStage,
				internal_telemetry.KV{K: "activity_type", V: internal_telemetry.StringValue("SpeechEndActivity")},
				internal_telemetry.KV{K: "speech", V: internal_telemetry.StringValue(vl.Speech)},
			)

			msg, err := talking.messaging.GetMessage()
			if err != nil {
				talking.logger.Errorf("the message should have gotten created from speech to text packet or user text packet %v", err)
				continue
			}
			if err := talking.Notify(ctx,
				&protos.AssistantConversationUserMessage{Id: msg.GetId(), Message: &protos.AssistantConversationUserMessage_Text{Text: &protos.AssistantConversationMessageTextContent{Content: msg.String()}}, Completed: true, Time: timestamppb.New(time.Now())}); err != nil {
				talking.logger.Tracef(ctx, "might be returing processing the duplicate message so cut it out.")
				continue
			}
			utils.Go(ctx, func() {
				if err := talking.onCreateMessage(ctx, internal_type.UserTextPacket{ContextID: msg.GetId(), Text: msg.String()}); err != nil {
					talking.logger.Errorf("Error in onCreateMessage: %v", err)
				}
			})

			//
			talking.messaging.Transition(internal_adapter_request_customizers.UserCompleted)
			if err := talking.assistantExecutor.Execute(ctx, talking, internal_type.UserTextPacket{ContextID: msg.GetId(), Text: msg.String()}); err != nil {
				talking.logger.Errorf("assistant executor error: %v", err)
				talking.OnError(ctx)
				continue
			}
		case internal_type.LLMStreamPacket:
			// bot had spoken reset the timer
			talking.resetIdleTimeoutTimer(talking.Context())
			// packet from llm reciecved
			inputMessage, err := talking.messaging.GetMessage()
			if err != nil {
				continue
			}
			// might be stale packet
			if vl.ContextID != inputMessage.GetId() {
				continue
			}

			if err := talking.messaging.Transition(internal_adapter_request_customizers.LLMGenerating); err != nil {
				talking.logger.Errorf("messaging transition error: %v", err)
			}
			// sending to assembler for assembling sentences
			if err := talking.callSentenceAssembler(ctx, vl); err != nil {
				talking.logger.Errorf("sentence assembler error: %v, calling speak directly", err)
				if err := talking.callSpeaking(ctx, vl); err != nil {
					talking.logger.Errorf("speaking error: %v", err)
				}
			}
			// end of speech analyzer in case histoyrical data is to be used
			if err := talking.callEndOfSpeech(ctx, vl); err != nil {
				talking.logger.Errorf("end of speech error: %v", err)
			}
		case internal_type.LLMMessagePacket:
			talking.resetIdleTimeoutTimer(talking.Context())
			inputMessage, err := talking.messaging.GetMessage()
			if err != nil {
				continue
			}
			// might be stale packet
			if vl.ContextID != inputMessage.GetId() {
				continue
			}
			if err := talking.messaging.Transition(internal_adapter_request_customizers.LLMGenerated); err != nil {
				talking.logger.Errorf("messaging transition error: %v", err)
			}
			if err := talking.callCreateMessage(ctx, vl); err != nil {
				talking.logger.Errorf("error creating message: %v", err)
			}
			if err := talking.callEndOfSpeech(ctx, vl); err != nil {
				talking.logger.Errorf("end of speech error: %v", err)
			}

			if err := talking.callSentenceAssembler(ctx, vl); err != nil {
				talking.logger.Errorf("sentence assembler error: %v calling speak directly", err)
				if err := talking.callSpeaking(ctx, vl); err != nil {
					talking.logger.Errorf("speaking error: %v", err)
				}
			}
			continue
		case internal_type.LLMToolPacket:
			talking.Notify(ctx, &protos.AssistantMessagingResponse_Action{
				Action: &protos.AssistantConversationAction{
					Name:   vl.ContextID,
					Action: vl.Action,
				},
			})
			continue
		case internal_type.MetricPacket:
			// metrics update for the message
			// later this can be used at each stage to calculate various metrics
			if len(vl.Metrics) > 0 {
				if err := talking.OnMessageMetric(talking.Context(), vl.ContextID, vl.Metrics); err != nil {
					talking.logger.Errorf("Error in OnUpdateMessage: %v", err)
				}
			}
		case internal_type.TextToSpeechEndPacket:
			// notify the user about completion of tts
			inputMessage, err := talking.messaging.GetMessage()
			if err != nil {
				continue
			}
			// might be stale packet
			if vl.ContextID != inputMessage.GetId() {
				continue
			}
			if err := talking.Notify(talking.Context(), &protos.AssistantConversationAssistantMessage{Time: timestamppb.Now(), Id: vl.ContextID, Completed: true}); err != nil {
				talking.logger.Tracef(talking.ctx, "error while outputing chunk to the user: %w", err)
			}
			continue
		case internal_type.TextToSpeechAudioPacket:
			// get current input message
			inputMessage, err := talking.messaging.GetMessage()
			if err != nil {
				continue
			}
			// might be stale packet
			if vl.ContextID != inputMessage.GetId() {
				continue
			}

			// notify the user about audio chunk
			if err := talking.Notify(talking.Context(), &protos.AssistantConversationAssistantMessage{Time: timestamppb.Now(), Id: vl.ContextID, Message: &protos.AssistantConversationAssistantMessage_Audio{Audio: &protos.AssistantConversationMessageAudioContent{Content: vl.AudioChunk}}}); err != nil {
				talking.logger.Tracef(talking.ctx, "error while outputing chunk to the user: %w", err)
			}

			// for recording puposes
			if err := talking.callRecording(ctx, vl); err != nil {
				talking.logger.Errorf("recorder error: %v", err)
			}
			continue
		default:
			talking.logger.Warnf("unknown packet type received in OnGeneration %T", vl)
		}
	}

	return nil
}

func (spk *GenericRequestor) callSpeaking(ctx context.Context, result internal_type.LLMPacket) error {
	switch res := result.(type) {
	case internal_type.LLMMessagePacket:
		if spk.textToSpeechTransformer != nil {
			ctx, span, _ := spk.Tracer().StartSpan(spk.Context(), utils.AssistantSpeakingStage)
			defer span.EndSpan(ctx, utils.AssistantSpeakingStage)
			span.AddAttributes(ctx,
				internal_adapter_telemetry.MessageKV(res.ContextID),
				internal_adapter_telemetry.KV{K: "activity", V: internal_adapter_telemetry.StringValue("finish_speaking")},
			)
			if err := spk.textToSpeechTransformer.Transform(spk.Context(), res); err != nil {
				spk.logger.Errorf("speak: failed to send flush to text to speech transformer error: %v", err)
			}
		}
	case internal_type.LLMStreamPacket:
		if spk.textToSpeechTransformer != nil {
			ctx, span, _ := spk.Tracer().StartSpan(spk.Context(), utils.AssistantSpeakingStage)
			defer span.EndSpan(ctx, utils.AssistantSpeakingStage)
			span.AddAttributes(ctx,
				internal_adapter_telemetry.MessageKV(res.ContextID),
				internal_adapter_telemetry.KV{K: "activity", V: internal_adapter_telemetry.StringValue("speak")},
				internal_adapter_telemetry.KV{K: "script", V: internal_adapter_telemetry.StringValue(res.Text)},
			)
			if err := spk.textToSpeechTransformer.Transform(spk.Context(), res); err != nil {
				spk.logger.Errorf("speak: failed to send flush to text to speech transformer error: %v", err)
			}
		}

		inputMessage, err := spk.messaging.GetMessage()
		if err != nil {
			return nil
		}
		// might be stale packet
		if res.ContextId() != inputMessage.GetId() {
			return nil
		}

		if err := spk.Notify(ctx, &protos.AssistantConversationAssistantMessage{Time: timestamppb.Now(), Id: res.ContextId(), Completed: true, Message: &protos.AssistantConversationAssistantMessage_Text{Text: &protos.AssistantConversationMessageTextContent{Content: res.Text}}}); err != nil {
			spk.logger.Tracef(ctx, "error while outputting chunk to the user: %w", err)
		}
	default:
	}
	return nil
}
