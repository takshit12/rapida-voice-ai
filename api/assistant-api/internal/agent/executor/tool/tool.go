// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_agent_executor_tool

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	internal_agent_executor "github.com/rapidaai/api/assistant-api/internal/agent/executor"
	internal_tool "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal"
	internal_tool_local "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal/local"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_adapter_telemetry "github.com/rapidaai/api/assistant-api/internal/telemetry"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"

	"github.com/rapidaai/protos"

	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/types"
	type_enums "github.com/rapidaai/pkg/types/enums"
	"github.com/rapidaai/pkg/utils"
)

type toolExecutor struct {
	logger                 commons.Logger
	tools                  map[string]internal_tool.ToolCaller
	availableToolFunctions []*protos.FunctionDefinition
}

func NewToolExecutor(
	logger commons.Logger,
) internal_agent_executor.ToolExecutor {
	return &toolExecutor{
		logger:                 logger,
		tools:                  make(map[string]internal_tool.ToolCaller, 0),
		availableToolFunctions: make([]*protos.FunctionDefinition, 0),
	}
}

func (executor *toolExecutor) GetLocalTool(logger commons.Logger, toolOpts *internal_assistant_entity.AssistantTool, communcation internal_type.Communication) (internal_tool.ToolCaller, error) {
	switch toolOpts.ExecutionMethod {
	case "knowledge_retrieval":
		return internal_tool_local.NewKnowledgeRetrievalToolCaller(logger, toolOpts, communcation)
	case "api_request":
		return internal_tool_local.NewApiRequestToolCaller(logger, toolOpts, communcation)
	case "endpoint":
		return internal_tool_local.NewEndpointToolCaller(logger, toolOpts, communcation)
	case "put_on_hold":
		return internal_tool_local.NewPutOnHoldToolCaller(logger, toolOpts, communcation)
	case "end_of_conversation":
		return internal_tool_local.NewEndOfConversationCaller(logger, toolOpts, communcation)
	default:
		return nil, errors.New("illegal tool action provided")
	}
}

func (executor *toolExecutor) Initialize(ctx context.Context, communication internal_type.Communication) error {
	ctx, span, _ := communication.Tracer().StartSpan(ctx, utils.AssistantToolConnectStage)
	defer span.EndSpan(ctx, utils.AssistantToolConnectStage)

	//
	start := time.Now()
	for _, tool := range communication.Assistant().AssistantTools {
		caller, err := executor.GetLocalTool(executor.logger, tool, communication)
		if err != nil {
			executor.logger.Errorf("error while initialize tool action %s", err)
			continue
		}
		span.AddAttributes(ctx, internal_adapter_telemetry.KV{K: caller.Name(), V: internal_adapter_telemetry.StringValue(caller.ExecutionMethod())})
		tlf, err := caller.Definition()
		if err != nil {
			executor.logger.Errorf("unable to generate tool definition %s", err)
			continue
		}
		//
		executor.tools[caller.Name()] = caller
		executor.availableToolFunctions = append(executor.availableToolFunctions, tlf)

	}
	executor.logger.Benchmark("ToolExecutor.Init", time.Since(start))
	return nil
}

func (executor *toolExecutor) GetFunctionDefinitions() []*protos.FunctionDefinition {
	return executor.availableToolFunctions
}

func (executor *toolExecutor) tool(messageId string, in, out map[string]interface{}, metrics []*types.Metric, communication internal_type.Communication) error {
	utils.Go(communication.Context(), func() {
		communication.CreateConversationToolLog(messageId, in, out, metrics)
	})
	return nil
}

func (executor *toolExecutor) execute(ctx context.Context, message internal_type.LLMPacket, call *protos.ToolCall, communication internal_type.Communication) internal_type.LLMToolPacket {
	ctx, span, _ := communication.Tracer().StartSpan(ctx, utils.AssistantToolExecuteStage, internal_adapter_telemetry.MessageKV(message.ContextID))
	defer span.EndSpan(ctx, utils.AssistantToolExecuteStage)

	start := time.Now()
	metrics := make([]*types.Metric, 0)

	funC, ok := executor.tools[call.GetFunction().GetName()]
	if !ok {
		return internal_type.LLMToolPacket{ContextID: message.ContextID,
			Action: protos.AssistantConversationAction_ACTION_UNSPECIFIED, Result: map[string]interface{}{
				"error":   "unable to find tool.",
				"success": false,
				"status":  "FAIL",
			}}
	}

	// should return multiple things
	span.AddAttributes(ctx, internal_adapter_telemetry.KV{K: "function", V: internal_adapter_telemetry.StringValue(call.GetFunction().GetName())}, internal_adapter_telemetry.KV{K: "argument", V: internal_adapter_telemetry.StringValue(call.GetFunction().GetArguments())})

	output := funC.Call(ctx, message, call.GetId(), call.GetFunction().GetArguments(), communication)
	metrics = append(metrics, types.NewTimeTakenMetric(time.Since(start)))

	//
	executor.tool(message.ContextID, map[string]interface{}{
		"id":        call.Id,
		"name":      call.GetFunction().GetName(),
		"arguments": call.GetFunction().GetArguments(),
	}, output.Result, metrics, communication)

	executor.Log(ctx, funC, communication, message.ContextID, type_enums.RECORD_COMPLETE, int64(time.Since(start)), map[string]interface{}{
		"id":        call.Id,
		"name":      call.GetFunction().GetName(),
		"arguments": call.GetFunction().GetArguments(),
	}, output.Result)

	return output
}

func (executor *toolExecutor) ExecuteAll(ctx context.Context, message internal_type.LLMPacket, calls []*protos.ToolCall, communication internal_type.Communication) ([]internal_type.Packet, []*types.Content) {
	contents := make([]internal_type.Packet, 0)
	result := make([]*types.Content, 0)
	var wg sync.WaitGroup
	//
	for _, xt := range calls {
		xtCopy := xt
		wg.Add(1) // Move this outside of the goroutine
		utils.Go(context.Background(), func() {
			defer wg.Done()
			cntn := executor.execute(ctx, message, xtCopy, communication)
			contents = append(contents, cntn)
			//
			bt, err := json.Marshal(cntn.Result)
			if err != nil {
				result = append(result, &types.Content{
					ContentType:   xt.GetId(),
					ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
					Content:       []byte("unable to parse the response."),
					Name:          xt.GetFunction().GetName(),
				})
				return
			}
			result = append(result, &types.Content{
				ContentType:   xt.GetId(),
				ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
				Content:       bt,
				Name:          xt.GetFunction().GetName(),
			})

		})
	}
	wg.Wait()
	return contents, result
}

func (executor *toolExecutor) Log(ctx context.Context, toolCaller internal_tool.ToolCaller, communication internal_type.Communication, assistantConversationMessageId string, recordStatus type_enums.RecordState, timeTaken int64, in, out map[string]interface{}) {
	utils.Go(ctx, func() {
		i, _ := json.Marshal(in)
		o, _ := json.Marshal(out)
		communication.CreateToolLog(toolCaller.Id(), assistantConversationMessageId, toolCaller.Name(), toolCaller.ExecutionMethod(), recordStatus, timeTaken, i, o)
	})
}
