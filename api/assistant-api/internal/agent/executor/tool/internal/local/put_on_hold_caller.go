// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_tool_local

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	internal_tool "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type putOnHoldToolCaller struct {
	toolCaller
	maxHoldTime uint64
}

func (tc *putOnHoldToolCaller) argument(args string) uint64 {
	var input map[string]interface{}
	err := json.Unmarshal([]byte(args), &input)
	if err != nil {
		tc.logger.Debugf("illegal input from llm check and pushing the llm response as incomplete %v", args)
		return tc.maxHoldTime
	}
	var duration uint64
	switch v := input["duration"].(type) {
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return tc.maxHoldTime
		}
		return utils.MaxUint64(duration, parsed)
	case float64:
		return utils.MaxUint64(duration, uint64(v))
	case int:
		return utils.MaxUint64(duration, uint64(v))
	case uint64:
		return utils.MaxUint64(duration, v)
	}
	return tc.maxHoldTime

}

func (afkTool *putOnHoldToolCaller) Call(ctx context.Context, pkt internal_type.LLMPacket, toolId string, args string, communication internal_type.Communication) internal_type.LLMToolPacket {
	return internal_type.LLMToolPacket{ContextID: pkt.ContextID, Action: protos.AssistantConversationAction_PUT_ON_HOLD, Result: afkTool.Result("Putting on hold.", true)}
}

func NewPutOnHoldToolCaller(
	logger commons.Logger,
	toolOptions *internal_assistant_entity.AssistantTool,
	communication internal_type.Communication,
) (internal_tool.ToolCaller, error) {

	opts := toolOptions.GetOptions()
	var maxHoldTime uint64
	switch v := opts["tool.max_hold_time"].(type) {
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("tool.endpoint_id is not a valid number: %v", err)
		}
		maxHoldTime = parsed
	case float64:
		maxHoldTime = uint64(v)
	case int:
		maxHoldTime = uint64(v)
	case uint64:
		maxHoldTime = v
	default:
		return nil, fmt.Errorf("tool.max_hold_time is not a recognized type, got %T", v)
	}

	return &putOnHoldToolCaller{
		toolCaller: toolCaller{
			logger:      logger,
			toolOptions: toolOptions,
		},
		maxHoldTime: maxHoldTime,
	}, nil
}
