// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_adapter_generic

import (
	"context"
	"errors"
	"strings"
	"time"

	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	type_enums "github.com/rapidaai/pkg/types/enums"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
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

func (communication *GenericRequestor) InitializeBehavior(ctx context.Context) error {

	behavior, err := communication.GetBehavior()
	if err != nil {
		communication.logger.Errorf("error while fetching deployment behavior: %v", err)
		return nil
	}

	if behavior.Greeting != nil {
		greetingCnt := communication.templateParser.Parse(*behavior.Greeting, communication.GetArgs())
		if strings.TrimSpace(greetingCnt) != "" {
			message := communication.messaging.Create(type_enums.UserActor, "")
			if err := communication.OnPacket(ctx, internal_type.StaticPacket{ContextID: message.GetId(), Text: greetingCnt}); err != nil {
				communication.logger.Errorf("error while sending on error message: %v", err)
			}
		}
	}

	if behavior.IdealTimeout != nil && *behavior.IdealTimeout > 0 {
		// start the ideal timeout timer
		communication.StartIdealTimeoutTimer(ctx)
	}

	if behavior.MaxSessionDuration != nil && *behavior.MaxSessionDuration > 0 {
		timeoutDuration := time.Duration(*behavior.MaxSessionDuration) * time.Minute
		time.AfterFunc(timeoutDuration, func() {
			communication.logger.Infof("conversation timeout reached for assistant: %s", communication.assistant.Id)
			communication.OnPacket(ctx, internal_type.LLMToolPacket{ContextID: communication.messaging.GetId(), Action: protos.AssistantConversationAction_END_CONVERSATION})
		})
	}

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
	if err := communication.OnPacket(ctx, internal_type.StaticPacket{ContextID: messageId, Text: mistakeContent}); err != nil {
		communication.logger.Errorf("error while sending on error message: %v", err)
		return nil
	}
	return nil
}

// OnIdealTimeout handles the behavior when the bot has spoken but the user has not responded for the ideal timeout duration.
// If configured, it will ask the user if they are still there.
func (communication *GenericRequestor) OnIdealTimeout(ctx context.Context) error {
	behavior, err := communication.GetBehavior()
	if err != nil {
		communication.logger.Debugf("no ideal timeout behavior setup for assistant.")
		return nil
	}

	// Check if ideal timeout is configured
	if behavior.IdealTimeout == nil || *behavior.IdealTimeout == 0 {
		return nil
	}

	if behavior.IdealTimeoutBackoff != nil && *behavior.IdealTimeoutBackoff > 0 && communication.idealTimeoutCount >= *behavior.IdealTimeoutBackoff {
		communication.OnPacket(ctx, internal_type.LLMToolPacket{ContextID: communication.messaging.GetId(), Action: protos.AssistantConversationAction_END_CONVERSATION})
		return nil
	}

	// Use default or configured timeout message
	communication.idealTimeoutCount++
	timeoutContent := "Are you still there?"
	if behavior.IdealTimeoutMessage != nil && strings.TrimSpace(*behavior.IdealTimeoutMessage) != "" {
		timeoutContent = communication.templateParser.Parse(*behavior.IdealTimeoutMessage, communication.GetArgs())
	}

	if strings.TrimSpace(timeoutContent) == "" {
		communication.logger.Warnf("empty ideal timeout message")
		return nil
	}
	if err := communication.OnPacket(ctx, internal_type.StaticPacket{ContextID: communication.messaging.GetId(), Text: timeoutContent}); err != nil {
		communication.logger.Errorf("error while sending ideal timeout message: %v", err)
		return nil
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
	timeoutDuration := time.Duration(*behavior.IdealTimeout) * time.Minute
	communication.idealTimeoutTimer = time.AfterFunc(timeoutDuration, func() {
		if err := communication.OnIdealTimeout(ctx); err != nil {
			communication.logger.Errorf("error while handling ideal timeout: %v", err)
		}
	})
}

// ResetIdealTimeoutTimer resets the ideal timeout timer when the user speaks (indicating they are still there).
func (communication *GenericRequestor) ResetIdealTimeoutTimer(ctx context.Context) {
	if communication.idealTimeoutTimer == nil {
		return
	}
	communication.StartIdealTimeoutTimer(ctx)
}
