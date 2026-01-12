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
	"fmt"
	"sync"
	"time"

	internal_adapter_requests "github.com/rapidaai/api/assistant-api/internal/adapters"
	internal_requests "github.com/rapidaai/api/assistant-api/internal/adapters"
	internal_agent_executor "github.com/rapidaai/api/assistant-api/internal/agent/executor"
	internal_tool "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal"
	internal_tool_local "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal/local"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_adapter_telemetry "github.com/rapidaai/api/assistant-api/internal/telemetry"

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

func (executor *toolExecutor) GetLocalTool(logger commons.Logger, toolOpts *internal_assistant_entity.AssistantTool, communcation internal_adapter_requests.Communication) (internal_tool.ToolCaller, error) {
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

func (executor *toolExecutor) Initialize(ctx context.Context, communication internal_requests.Communication) error {
	ctx, span, _ := communication.Tracer().StartSpan(ctx, utils.AssistantToolConnectStage)
	defer span.EndSpan(ctx, utils.AssistantToolConnectStage)

	start := time.Now()
	for _, tool := range communication.Assistant().AssistantTools {
		caller, err := executor.GetLocalTool(executor.logger, tool, communication)
		if err != nil {
			executor.logger.Errorf("error while initialize tool action %s", err)
			continue
		}
		span.AddAttributes(ctx,
			internal_adapter_telemetry.KV{
				K: caller.Name(),
				V: internal_adapter_telemetry.StringValue(caller.ExecutionMethod()),
			},
		)
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

func (executor *toolExecutor) tool(
	messageId string,
	in,
	out map[string]interface{},
	metrics []*types.Metric,
	communication internal_requests.Communication) error {
	utils.Go(communication.Context(), func() {
		communication.CreateConversationToolLog(messageId, in, out, metrics)
	})
	return nil
}

func (executor *toolExecutor) Execute(ctx context.Context, messageid string, call *protos.ToolCall, communication internal_requests.Communication) *types.Content {
	ctx, span, _ := communication.Tracer().StartSpan(
		ctx,
		utils.AssistantToolExecuteStage,
		internal_adapter_telemetry.MessageKV(messageid),
	)
	defer span.EndSpan(ctx, utils.AssistantToolExecuteStage)

	start := time.Now()
	funcName := call.GetFunction().GetName()
	arguments := call.GetFunction().GetArguments()

	funC, ok := executor.tools[funcName]
	if !ok {
		executor.logger.Errorf("unable to find the func for tools with name %v", funcName)
		return &types.Content{
			ContentType:   commons.TEXT_CONTENT.String(),
			ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
			Content:       []byte(fmt.Sprintf("Unable to find the function name with %s", funcName)),
			Name:          call.Id,
		}
	}

	// should return multiple things
	span.AddAttributes(ctx,
		internal_adapter_telemetry.KV{
			K: "function",
			V: internal_adapter_telemetry.StringValue(funcName),
		},
		internal_adapter_telemetry.KV{
			K: "argument",
			V: internal_adapter_telemetry.StringValue(arguments),
		})

	output, metrics := funC.Call(ctx, messageid, arguments, communication)
	executor.tool(messageid, map[string]interface{}{
		"name":      funcName,
		"arguments": arguments,
	}, output, metrics, communication)

	executor.Log(ctx, funC, communication, messageid, type_enums.RECORD_COMPLETE, int64(time.Since(start)), map[string]interface{}{
		"name":      funcName,
		"arguments": arguments,
	}, output)
	//
	bt, err := json.Marshal(output)
	if err != nil {
		executor.logger.Errorf("error while calling function %v", err)
		return &types.Content{
			ContentType:   commons.TEXT_CONTENT.String(),
			ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
			Content:       []byte("unable to parse the response."),
			Name:          call.Id,
		}
	}
	executor.logger.Benchmark("ToolExecutor.Execute", time.Since(start))
	return &types.Content{
		ContentType:   commons.TEXT_CONTENT.String(),
		ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
		Content:       bt,
		Name:          call.Id,
	}
}

func (executor *toolExecutor) ExecuteAll(
	ctx context.Context,
	messageid string,
	calls []*protos.ToolCall,
	communication internal_requests.Communication) []*types.Content {
	start := time.Now()
	contents := make([]*types.Content, len(calls))
	var wg sync.WaitGroup
	for idx, xt := range calls {
		xtCopy := xt
		wg.Add(1) // Move this outside of the goroutine
		utils.Go(context.Background(), func() {
			defer wg.Done()
			cntn := executor.Execute(ctx, messageid, xtCopy, communication)
			cntn.Name = xtCopy.GetFunction().GetName()
			cntn.ContentType = xtCopy.GetId()
			contents[idx] = cntn
		})
	}
	wg.Wait()
	executor.logger.Benchmark("ToolExecutor.ExecuteAll", time.Since(start))
	return contents
}

func (executor *toolExecutor) Log(
	ctx context.Context,
	toolCaller internal_tool.ToolCaller,
	communication internal_requests.Communication,
	assistantConversationMessageId string,
	recordStatus type_enums.RecordState,
	timeTaken int64,
	in, out map[string]interface{},
) {
	utils.Go(ctx, func() {
		i, _ := json.Marshal(in)
		o, _ := json.Marshal(out)
		communication.CreateToolLog(
			toolCaller.Id(),
			assistantConversationMessageId,
			toolCaller.Name(),
			toolCaller.ExecutionMethod(),
			recordStatus,
			timeTaken,
			i, o,
		)
	})
}
