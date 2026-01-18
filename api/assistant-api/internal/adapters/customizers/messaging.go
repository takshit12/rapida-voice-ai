// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_adapter_request_customizers

import (
	"fmt"
	"sync"

	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/types"
	type_enums "github.com/rapidaai/pkg/types/enums"
)

type Messaging interface {
	Create(msg string) *types.Message
	GetMessage() (*types.Message, error)
	Transition(state InteractionState) error
	GetInputMode() type_enums.MessageMode
	SwitchInputMode(mm type_enums.MessageMode)
	GetOutputMode() type_enums.MessageMode
	SwitchOutputMode(mm type_enums.MessageMode)
}

type InteractionState int

const (
	Unknown       InteractionState = 1
	UserSpeaking  InteractionState = 2
	UserCompleted InteractionState = 3

	//
	Interrupt   InteractionState = 6
	Interrupted InteractionState = 7

	//
	LLMGenerating InteractionState = 8
	LLMGenerated  InteractionState = 5
)

func (s InteractionState) String() string {
	switch s {
	case Unknown:
		return "Unknown"
	case UserSpeaking:
		return "UserSpeaking"
	case UserCompleted:
		return "UserCompleted"
	case LLMGenerated:
		return "LLMGenerated"
	case Interrupt:
		return "Interrupt"
	case Interrupted:
		return "Interrupted"
	case LLMGenerating:
		return "LLMGenerating"
	default:
		return "InvalidState"
	}
}

type messaging struct {
	logger commons.Logger
	in     *types.Message
	actor  type_enums.MessageActor
	state  InteractionState

	// rw mutex
	mutex sync.RWMutex

	inputMode  type_enums.MessageMode
	outputMode type_enums.MessageMode
}

func NewMessaging(logger commons.Logger) Messaging {
	return &messaging{
		logger:     logger,
		actor:      type_enums.UserActor,
		inputMode:  type_enums.TextMode,
		outputMode: type_enums.TextMode,
		state:      Unknown,
	}
}

// ============================================================================
// Input and output mode handling
// ============================================================================

func (ms *messaging) GetOutputMode() type_enums.MessageMode {
	return ms.outputMode
}

func (ms *messaging) SwitchOutputMode(mm type_enums.MessageMode) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.outputMode = mm
}

func (ms *messaging) GetInputMode() type_enums.MessageMode {
	return ms.inputMode
}

func (ms *messaging) SwitchInputMode(mm type_enums.MessageMode) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.inputMode = mm
}

func (ms *messaging) GetMessage() (*types.Message, error) {
	if ms.in != nil {
		return ms.in, nil
	}
	return nil, fmt.Errorf("invalid message for acting user")
}

func (ms *messaging) Create(msg string) *types.Message {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	if ms.in != nil {
		ms.in.MergeContent(&types.Content{
			ContentType:   commons.TEXT_CONTENT.String(),
			ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
			Content:       []byte(msg),
		})
	} else {
		ms.in = types.NewMessage("user", &types.Content{
			ContentType:   commons.TEXT_CONTENT.String(),
			ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
			Content:       []byte(msg),
		})
		ms.in.AddMetadata("mode", ms.inputMode.String())

	}
	return ms.in
}

func (ms *messaging) Transition(newState InteractionState) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	switch newState {
	case Unknown:
		return fmt.Errorf("Transition: invalid transition: cannot transition to Unknown state")
	case UserSpeaking:
	case UserCompleted:
	case LLMGenerating:
	case LLMGenerated:
	case Interrupt:
		if ms.state == Interrupted || ms.state == Interrupt {
			return fmt.Errorf("Transition: invalid transition: agent can't interrupt multiple times")
		}
		if ms.state == LLMGenerated || ms.state == LLMGenerating {
			ms.in = nil
		}
	case Interrupted:
		if ms.state == Interrupted {
			return fmt.Errorf("Transition: invalid transition: agent can't interrupted multiple times")
		}
		if ms.state == LLMGenerated || ms.state == LLMGenerating {
			ms.in = nil
		}

	default:
		return fmt.Errorf("Transition: invalid transition: unknown state %v", newState)
	}

	ms.state = newState
	return nil
}
