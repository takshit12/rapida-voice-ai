// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_agent_executor_llm

import (
	"context"
	"errors"

	internal_agent_executor "github.com/rapidaai/api/assistant-api/internal/agent/executor"
	internal_model "github.com/rapidaai/api/assistant-api/internal/agent/executor/llm/internal/model"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	type_enums "github.com/rapidaai/pkg/types/enums"
)

type assistantExecutor struct {
	logger   commons.Logger
	executor internal_agent_executor.AssistantExecutor
}

func NewAssistantExecutor(logger commons.Logger) internal_agent_executor.AssistantExecutor {
	return &assistantExecutor{
		logger: logger,
	}
}

// Init implements internal_executors.AssistantExecutor.
func (a *assistantExecutor) Initialize(ctx context.Context, communication internal_type.Communication) error {
	switch communication.Assistant().AssistantProvider {
	// case type_enums.AGENTKIT:
	// 	a.executor = internal_agentkit.NewAgentKitAssistantExecutor(a.logger)
	// case type_enums.WEBSOCKET:
	// 	a.executor = internal_websocket.NewWebsocketAssistantExecutor(a.logger)
	case type_enums.MODEL:
		a.executor = internal_model.NewModelAssistantExecutor(a.logger)
	default:
		return errors.New("illegal assistant executor")
	}
	return a.executor.Initialize(ctx, communication)
}

// Name implements internal_executors.AssistantExecutor.
func (a *assistantExecutor) Name() string {
	return a.executor.Name()
}

// Talk implements internal_executors.AssistantExecutor.
func (a *assistantExecutor) Execute(ctx context.Context, communication internal_type.Communication, pctk internal_type.Packet) error {
	if a.executor == nil {
		return errors.New("assistant executor not initialized")
	}
	return a.executor.Execute(ctx, communication, pctk)
}

func (a *assistantExecutor) Close(ctx context.Context, communication internal_type.Communication) error {
	if a.executor == nil {
		return errors.New("assistant executor not initialized")
	}
	return a.executor.Close(ctx, communication)
}
