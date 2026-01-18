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
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

var errDeploymentNotEnabled = errors.New("deployment is not enabled for source")

// GetBehavior retrieves the deployment behavior configuration based on the source type.
func (r *GenericRequestor) GetBehavior() (*internal_assistant_entity.AssistantDeploymentBehavior, error) {
	if r.assistant == nil {
		return nil, errDeploymentNotEnabled
	}

	switch r.source {
	case utils.PhoneCall:
		if r.assistant.AssistantPhoneDeployment != nil {
			return &r.assistant.AssistantPhoneDeployment.AssistantDeploymentBehavior, nil
		}
	case utils.Whatsapp:
		if r.assistant.AssistantWhatsappDeployment != nil {
			return &r.assistant.AssistantWhatsappDeployment.AssistantDeploymentBehavior, nil
		}
	case utils.SDK:
		if r.assistant.AssistantApiDeployment != nil {
			return &r.assistant.AssistantApiDeployment.AssistantDeploymentBehavior, nil
		}
	case utils.WebPlugin:
		if r.assistant.AssistantWebPluginDeployment != nil {
			return &r.assistant.AssistantWebPluginDeployment.AssistantDeploymentBehavior, nil
		}
	case utils.Debugger:
		if r.assistant.AssistantDebuggerDeployment != nil {
			return &r.assistant.AssistantDebuggerDeployment.AssistantDeploymentBehavior, nil
		}
	}

	return nil, errDeploymentNotEnabled
}

// InitializeBehavior sets up the initial behavior configuration including greeting,
// idle timeout, and max session duration timers.
func (r *GenericRequestor) InitializeBehavior(ctx context.Context) error {
	behavior, err := r.GetBehavior()
	if err != nil {
		r.logger.Errorf("error while fetching deployment behavior: %v", err)
		return nil
	}

	r.initializeGreeting(ctx, behavior)
	r.initializeIdleTimeout(ctx, behavior)
	r.initializeMaxSessionDuration(ctx, behavior)

	return nil
}

// initializeGreeting sends the greeting message if configured.
func (r *GenericRequestor) initializeGreeting(ctx context.Context, behavior *internal_assistant_entity.AssistantDeploymentBehavior) {
	if behavior.Greeting == nil {
		return
	}

	greetingContent := r.templateParser.Parse(*behavior.Greeting, r.GetArgs())
	if strings.TrimSpace(greetingContent) == "" {
		return
	}

	message := r.messaging.Create("")
	if err := r.OnPacket(ctx, internal_type.StaticPacket{ContextID: message.GetId(), Text: greetingContent}); err != nil {
		r.logger.Errorf("error while sending greeting message: %v", err)
	}
}

// initializeIdleTimeout starts the idle timeout timer if configured.
func (r *GenericRequestor) initializeIdleTimeout(ctx context.Context, behavior *internal_assistant_entity.AssistantDeploymentBehavior) {
	if behavior.IdealTimeout == nil || *behavior.IdealTimeout <= 0 {
		return
	}
	r.startIdleTimeoutTimer(ctx)
}

// initializeMaxSessionDuration sets up the max session duration timer if configured.
func (r *GenericRequestor) initializeMaxSessionDuration(ctx context.Context, behavior *internal_assistant_entity.AssistantDeploymentBehavior) {
	if behavior.MaxSessionDuration == nil || *behavior.MaxSessionDuration <= 0 {
		return
	}

	timeoutDuration := time.Duration(*behavior.MaxSessionDuration) * time.Minute
	r.maxSessionTimer = time.AfterFunc(timeoutDuration, func() {
		inputMessage, err := r.messaging.GetMessage()
		if err != nil {
			inputMessage = r.messaging.Create("")
		}
		r.OnPacket(ctx, internal_type.LLMToolPacket{
			ContextID: inputMessage.GetId(),
			Action:    protos.AssistantConversationAction_END_CONVERSATION,
		})
	})
}

// OnError handles error scenarios by sending a configured or default error message.
func (r *GenericRequestor) OnError(ctx context.Context) error {
	behavior, err := r.GetBehavior()
	if err != nil {
		r.logger.Warnf("no error message configured for assistant")
		return nil
	}

	const defaultMistakeMessage = "Oops! It looks like something went wrong. Let me look into that for you right away. I really appreciate your patience—hang tight while I get this sorted!"

	mistakeContent := defaultMistakeMessage
	if behavior.Mistake != nil {
		mistakeContent = r.templateParser.Parse(*behavior.Mistake, r.GetArgs())
	}

	inputMessage, err := r.messaging.GetMessage()
	if err != nil {
		inputMessage = r.messaging.Create("")
	}

	if err := r.OnPacket(ctx, internal_type.StaticPacket{ContextID: inputMessage.GetId(), Text: mistakeContent}); err != nil {
		r.logger.Errorf("error while sending error message: %v", err)
	}

	return nil
}

// OnIdleTimeout handles the behavior when the bot has spoken but the user
// has not responded within the idle timeout duration.
// If configured, it will prompt the user or end the conversation after max retries.
func (r *GenericRequestor) onIdleTimeout(ctx context.Context) error {
	behavior, err := r.GetBehavior()
	if err != nil {
		r.logger.Debugf("no idle timeout behavior configured for assistant")
		return nil
	}

	if behavior.IdealTimeout == nil || *behavior.IdealTimeout == 0 {
		return nil
	}

	inputMessage, err := r.messaging.GetMessage()
	if err != nil {
		inputMessage = r.messaging.Create("")
	}

	// Check if max backoff retries reached
	if behavior.IdealTimeoutBackoff != nil && *behavior.IdealTimeoutBackoff > 0 {
		if r.idleTimeoutCount >= *behavior.IdealTimeoutBackoff {
			r.OnPacket(ctx, internal_type.LLMToolPacket{
				ContextID: inputMessage.GetId(),
				Action:    protos.AssistantConversationAction_END_CONVERSATION,
			})
			return nil
		}
	}

	r.idleTimeoutCount++
	timeoutContent := r.getIdleTimeoutMessage(behavior)
	if timeoutContent == "" {
		r.logger.Warnf("empty idle timeout message")
		return nil
	}

	if err := r.OnPacket(ctx, internal_type.StaticPacket{ContextID: inputMessage.GetId(), Text: timeoutContent}); err != nil {
		r.logger.Errorf("error while sending idle timeout message: %v", err)
	}

	return nil
}

// getIdleTimeoutMessage returns the configured or default idle timeout message.
func (r *GenericRequestor) getIdleTimeoutMessage(behavior *internal_assistant_entity.AssistantDeploymentBehavior) string {
	const defaultTimeoutMessage = "Are you still there?"

	if behavior.IdealTimeoutMessage != nil && strings.TrimSpace(*behavior.IdealTimeoutMessage) != "" {
		return r.templateParser.Parse(*behavior.IdealTimeoutMessage, r.GetArgs())
	}

	return defaultTimeoutMessage
}

// StartIdleTimeoutTimer starts a timer that triggers OnIdleTimeout when the bot
// has spoken but the user hasn't responded within the configured duration.
func (r *GenericRequestor) startIdleTimeoutTimer(ctx context.Context) {
	if r.idleTimeoutTimer != nil {
		r.idleTimeoutTimer.Stop()
	}

	behavior, err := r.GetBehavior()
	if err != nil {
		return
	}

	if behavior.IdealTimeout == nil || *behavior.IdealTimeout == 0 {
		return
	}

	timeoutDuration := time.Duration(*behavior.IdealTimeout) * time.Minute
	r.idleTimeoutTimer = time.AfterFunc(timeoutDuration, func() {
		if err := r.onIdleTimeout(ctx); err != nil {
			r.logger.Errorf("error while handling idle timeout: %v", err)
		}
	})
}

// ResetIdleTimeoutTimer resets the idle timeout timer when the user responds,
// indicating they are still engaged in the conversation.
func (r *GenericRequestor) resetIdleTimeoutTimer(ctx context.Context) {
	if r.idleTimeoutTimer == nil {
		return
	}
	r.startIdleTimeoutTimer(ctx)
}
