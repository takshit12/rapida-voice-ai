// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_adapter_request_generic

import (
	"context"
	"errors"
	"strings"
	"time"

	internal_adapter_request_customizers "github.com/rapidaai/api/assistant-api/internal/adapters/customizers"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/types"
	type_enums "github.com/rapidaai/pkg/types/enums"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (gr *GenericRequestor) GetBehavior() (*internal_assistant_entity.AssistantDeploymentBehavior, error) {
	switch gr.source {
	case utils.PhoneCall:
		if a := gr.assistant; a != nil && a.AssistantPhoneDeployment != nil {
			return &a.AssistantPhoneDeployment.AssistantDeploymentBehavior, nil
		}
	case utils.Whatsapp:
		if a := gr.assistant; a != nil && a.AssistantWhatsappDeployment != nil {
			return &a.AssistantWhatsappDeployment.AssistantDeploymentBehavior, nil
		}
	case utils.SDK:
		if a := gr.assistant; a != nil && a.AssistantApiDeployment != nil {
			return &a.AssistantApiDeployment.AssistantDeploymentBehavior, nil
		}
	case utils.WebPlugin:
		if a := gr.assistant; a != nil && a.AssistantWebPluginDeployment != nil {
			return &a.AssistantWebPluginDeployment.AssistantDeploymentBehavior, nil
		}
	case utils.Debugger:
		if a := gr.assistant; a != nil && a.AssistantDebuggerDeployment != nil {
			return &a.AssistantDebuggerDeployment.AssistantDeploymentBehavior, nil
		}
	}
	return nil, errors.New("deployment is not enabled for source")
}

func (communication *GenericRequestor) OnGreet(ctx context.Context) error {

	behavior, err := communication.GetBehavior()
	if err != nil {
		communication.logger.Errorf("error while fetching deployment behavior: %v", err)
		return nil
	}

	if behavior.Greeting == nil {
		communication.logger.Errorf("error while fetching deployment behavior: %v", err)
		return nil
	}
	greetingCnt := communication.templateParser.Parse(*behavior.Greeting, communication.GetArgs())
	if strings.TrimSpace(greetingCnt) == "" {
		communication.logger.Warnf("empty greeting message, could be space in the table or argument contains space")
		return nil
	}

	message := communication.messaging.Create(type_enums.UserActor, "")
	utils.Go(ctx, func() {
		if err := communication.OnCreateMessage(ctx, message.GetId(), message); err != nil {
			communication.logger.Errorf("Error in OnCreateMessage: %v", err)
		}
	})

	greetings := &types.Message{
		Id:   message.GetId(),
		Role: "assistant",
		Contents: []*types.Content{{
			ContentType:   commons.TEXT_CONTENT.String(),
			ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
			Content:       []byte(greetingCnt),
		}}}

	if err := communication.Notify(ctx, &protos.AssistantConversationAssistantMessage{
		Time:      timestamppb.Now(),
		Id:        greetings.GetId(),
		Completed: true,
		Message: &protos.AssistantConversationAssistantMessage_Text{
			Text: &protos.AssistantConversationMessageTextContent{
				Content: greetingCnt,
			},
		},
	}); err != nil {
		communication.logger.Tracef(ctx, "error while outputting chunk to the user: %w", err)
	}

	communication.AssistantCallback(ctx, greetings.GetId(), greetings, nil)
	// audio processing
	if communication.messaging.GetInputMode().Audio() {
		if err := communication.Speak(internal_type.TextPacket{ContextID: greetings.GetId(), Text: greetingCnt}, internal_type.FlushPacket{ContextID: greetings.GetId()}); err == nil {
			communication.logger.Debugf("finished speaking greeting message")
		}
	}
	// Notify the response if there is no user message
	if err := communication.Notify(ctx, &protos.AssistantConversationMessage{
		MessageId:               greetings.GetId(),
		AssistantId:             communication.assistant.Id,
		AssistantConversationId: communication.assistantConversation.Id,
		Response:                greetings.ToProto(),
	}); err != nil {
		communication.logger.Tracef(ctx, "error while outputting chunk to the user: %w", err)
	}
	communication.messaging.Transition(internal_adapter_request_customizers.AgentCompleted)
	return nil
}

func (communication *GenericRequestor) OnError(ctx context.Context, messageId string) error {
	behavior, err := communication.GetBehavior()
	if err != nil {
		communication.logger.Warnf("no on error message setup for assistant.")
		return nil
	}

	mistakeContent := "Oops! It looks like something went wrong. Let me look into that for you right away. I really appreciate your patience—hang tight while I get this sorted!"
	if behavior.Mistake != nil {
		mistakeContent = communication.templateParser.Parse(*behavior.Mistake, communication.GetArgs())
	}

	msg := types.NewMessage(
		"assistant", &types.Content{
			ContentType:   commons.TEXT_CONTENT.String(),
			ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
			Content:       []byte(mistakeContent),
		})
	if err := communication.OnGeneration(ctx, messageId, msg); err != nil {
		communication.logger.Errorf("error while sending on error message: %v", err)
		return nil
	}
	if err := communication.OnGenerationComplete(ctx, messageId, msg, nil); err != nil {
		communication.logger.Errorf("error while completing on error message: %v", err)
	}
	return nil
}

// OnIdealTimeout handles the behavior when the bot has spoken but the user has not responded for the ideal timeout duration.
// If configured, it will ask the user if they are still there.
func (communication *GenericRequestor) OnIdealTimeout(ctx context.Context) error {

	communication.logger.Errorf("will speak something on ideal timeout")
	behavior, err := communication.GetBehavior()
	if err != nil {
		communication.logger.Debugf("no ideal timeout behavior setup for assistant.")
		return nil
	}

	// Check if ideal timeout is configured
	if behavior.IdealTimeout == nil || *behavior.IdealTimeout == 0 {
		return nil
	}

	// Use default or configured timeout message
	timeoutContent := "Are you still there?"
	if behavior.IdealTimeoutMessage != nil && strings.TrimSpace(*behavior.IdealTimeoutMessage) != "" {
		timeoutContent = communication.templateParser.Parse(*behavior.IdealTimeoutMessage, communication.GetArgs())
	}

	if strings.TrimSpace(timeoutContent) == "" {
		communication.logger.Warnf("empty ideal timeout message")
		return nil
	}

	message := communication.messaging.Create(type_enums.UserActor, "")

	timeoutMsg := &types.Message{
		Id:   message.GetId(),
		Role: "assistant",
		Contents: []*types.Content{{
			ContentType:   commons.TEXT_CONTENT.String(),
			ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
			Content:       []byte(timeoutContent),
		}}}

	if err := communication.Notify(ctx, &protos.AssistantConversationAssistantMessage{
		Time:      timestamppb.Now(),
		Id:        timeoutMsg.GetId(),
		Completed: true,
		Message: &protos.AssistantConversationAssistantMessage_Text{
			Text: &protos.AssistantConversationMessageTextContent{
				Content: timeoutContent,
			},
		},
	}); err != nil {
		communication.logger.Tracef(ctx, "error while outputting ideal timeout message to the user: %w", err)
	}

	communication.AssistantCallback(ctx, timeoutMsg.GetId(), timeoutMsg, nil)

	// audio processing
	if communication.messaging.GetInputMode().Audio() {
		communication.Speak(
			internal_type.TextPacket{ContextID: timeoutMsg.GetId(), Text: timeoutContent},
			internal_type.FlushPacket{ContextID: timeoutMsg.GetId()},
		)

	}

	// Notify the response
	if err := communication.Notify(ctx, &protos.AssistantConversationMessage{
		MessageId:               timeoutMsg.GetId(),
		AssistantId:             communication.assistant.Id,
		AssistantConversationId: communication.assistantConversation.Id,
		Response:                timeoutMsg.ToProto(),
	}); err != nil {
		communication.logger.Tracef(ctx, "error while outputting ideal timeout message: %w", err)
	}

	return nil
}

// StartIdealTimeoutTimer starts a timer that triggers OnIdealTimeout when the bot has spoken but user hasn't responded for the configured duration.
func (communication *GenericRequestor) StartIdealTimeoutTimer(ctx context.Context) {
	if communication.idealTimeoutTimer != nil {
		communication.idealTimeoutTimer.Stop()
	}
	behavior, err := communication.GetBehavior()
	if err != nil {
		return
	}
	if behavior.IdealTimeout == nil || *behavior.IdealTimeout == 0 {
		return
	}
	communication.lastAssistantMessageTime = time.Now()
	timeoutDuration := time.Duration(*behavior.IdealTimeout) * time.Minute
	communication.idealTimeoutTimer = time.AfterFunc(timeoutDuration, func() {
		if err := communication.OnIdealTimeout(ctx); err != nil {
			communication.logger.Errorf("error while handling ideal timeout: %v", err)
		}
		communication.StartIdealTimeoutTimer(ctx)
	})
}

// ResetIdealTimeoutTimer resets the ideal timeout timer when the user speaks (indicating they are still there).
func (communication *GenericRequestor) ResetIdealTimeoutTimer(ctx context.Context) {
	if communication.idealTimeoutTimer != nil {
		communication.idealTimeoutTimer.Stop()
	}
	behavior, err := communication.GetBehavior()
	if err != nil {
		return
	}
	if behavior.IdealTimeout == nil || *behavior.IdealTimeout == 0 {
		return
	}
	communication.StartIdealTimeoutTimer(ctx)
}
