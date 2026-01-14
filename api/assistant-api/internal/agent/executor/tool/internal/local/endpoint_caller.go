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
	"strings"

	internal_tool "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	endpoint_client_builders "github.com/rapidaai/pkg/clients/endpoint/builders"
	"github.com/rapidaai/pkg/commons"
	protos "github.com/rapidaai/protos"
)

type endpointToolCaller struct {
	toolCaller
	endpointId         uint64
	endpointParameters map[string]string
	inputBuilder       endpoint_client_builders.InputInvokeBuilder
}

func NewEndpointToolCaller(
	logger commons.Logger,
	toolOptions *internal_assistant_entity.AssistantTool,
	communcation internal_type.Communication,
) (internal_tool.ToolCaller, error) {
	opts := toolOptions.GetOptions()
	endpointID, err := opts.GetUint64("tool.endpoint_id")
	if err != nil {
		return nil, fmt.Errorf("tool.endpoint_id is not a valid number: %v", err)
	}
	parameters, err := opts.GetStringMap("tool.parameters")
	if err != nil {
		return nil, fmt.Errorf("failed to parse tool.parameters: %v", err)
	}

	return &endpointToolCaller{
		toolCaller: toolCaller{
			logger:      logger,
			toolOptions: toolOptions,
		},
		endpointId:         endpointID,
		endpointParameters: parameters,
		inputBuilder:       endpoint_client_builders.NewInputInvokeBuilder(logger),
	}, nil
}

func (afkTool *endpointToolCaller) Call(ctx context.Context, pkt internal_type.LLMPacket, toolId string, args string, communication internal_type.Communication) internal_type.LLMToolPacket {
	body := afkTool.Parse(afkTool.endpointParameters, args, communication)
	ivk, err := communication.DeploymentCaller().Invoke(
		ctx,
		communication.Auth(),
		afkTool.inputBuilder.Invoke(
			&protos.EndpointDefinition{
				EndpointId: afkTool.endpointId,
				Version:    "latest",
			},
			afkTool.inputBuilder.Arguments(body, nil),
			nil,
			nil,
		),
	)
	if err != nil {
		afkTool.logger.Errorf("error while calling endpoint %+v", err)
		return internal_type.LLMToolPacket{ContextID: pkt.ContextID, Action: protos.AssistantConversationAction_ENDPOINT_CALL, Result: afkTool.Result("Failed to resolve", false)}
	}
	if ivk.GetSuccess() {
		if data := ivk.GetData(); len(data) > 0 {
			var contentData map[string]interface{}
			if err := json.Unmarshal(data[0].Content, &contentData); err != nil {
				return internal_type.LLMToolPacket{ContextID: pkt.ContextID, Action: protos.AssistantConversationAction_ENDPOINT_CALL, Result: map[string]interface{}{"result": string(data[0].Content)}}
			}
			return internal_type.LLMToolPacket{ContextID: pkt.ContextID, Action: protos.AssistantConversationAction_ENDPOINT_CALL, Result: contentData}
		}

	}
	return internal_type.LLMToolPacket{ContextID: pkt.ContextID, Action: protos.AssistantConversationAction_ENDPOINT_CALL, Result: afkTool.Result("Failed to resolve", false)}
}

func (md *endpointToolCaller) Parse(
	mapping map[string]string,
	args string,
	communication internal_type.Communication,
) map[string]interface{} {
	arguments := make(map[string]interface{})
	for key, value := range mapping {
		if k, ok := strings.CutPrefix(key, "tool."); ok {
			switch k {
			case "name":
				arguments[value] = md.Name()
			case "argument":
				var argMap map[string]interface{}
				err := json.Unmarshal([]byte(args), &argMap)
				if err != nil {
					md.logger.Debugf("the arugment might be string")
					arguments[value] = args
				} else {
					arguments[value] = argMap
				}
			}
		}
		if k, ok := strings.CutPrefix(key, "assistant."); ok {
			switch k {
			case "id":
				arguments[value] = fmt.Sprintf("%d", communication.Assistant().Id)
			case "version":
				arguments[value] = fmt.Sprintf("vrsn_%d", communication.Assistant().AssistantProviderModel.Id)
			}
		}
		if k, ok := strings.CutPrefix(key, "conversation."); ok {
			switch k {
			case "id":
				arguments[value] = fmt.Sprintf("%d", communication.Conversation().Id)
			case "messages":
				arguments[value] = md.SimplifyHistoy(communication.GetHistories())
			}
		}
		if k, ok := strings.CutPrefix(key, "argument."); ok {
			if aArg, ok := communication.GetArgs()[k]; ok {
				arguments[value] = aArg
			}
		}
		if k, ok := strings.CutPrefix(key, "metadata."); ok {
			if mtd, ok := communication.GetMetadata()[k]; ok {
				arguments[value] = mtd
			}
		}
		if k, ok := strings.CutPrefix(key, "option."); ok {
			if ot, ok := communication.GetOptions()[k]; ok {
				arguments[value] = ot
			}
		}
	}
	return arguments
}

func (md *endpointToolCaller) SimplifyHistoy(msgs []internal_type.MessagePacket) []map[string]string {
	out := make([]map[string]string, 0)
	for _, msg := range msgs {
		out = append(out, map[string]string{
			"role":    msg.Role(),
			"message": msg.Content(),
		})
	}
	return out
}
