// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_tool_local

import (
	"context"

	internal_tool "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/protos"
)

type endOfConversationCaller struct {
	toolCaller
}

func (afkTool *endOfConversationCaller) Call(ctx context.Context, pkt internal_type.LLMPacket, toolId string, args string, communication internal_type.Communication) internal_type.LLMToolPacket {
	return internal_type.LLMToolPacket{ContextID: pkt.ContextID, Action: protos.AssistantConversationAction_END_CONVERSATION, Result: afkTool.Result("Disconnected successfully.", true)}
}

func NewEndOfConversationCaller(logger commons.Logger, toolOptions *internal_assistant_entity.AssistantTool, communcation internal_type.Communication,
) (internal_tool.ToolCaller, error) {
	return &endOfConversationCaller{
		toolCaller: toolCaller{
			logger:      logger,
			toolOptions: toolOptions,
		},
	}, nil
}
