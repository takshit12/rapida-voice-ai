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

	internal_agent_executor "github.com/rapidaai/api/assistant-api/internal/agent/executor"
	internal_tool "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal"
	internal_tool_local "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal/local"
	mcp "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal/mcp"
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
	mcpClients             map[string]mcp.MCPClient    // serverUrl -> client
	mcpTools               map[string]*mcp.MCPToolInfo // toolName -> MCP tool info
	tools                  map[string]internal_tool.ToolCaller
	availableToolFunctions []*protos.FunctionDefinition
}

func NewToolExecutor(
	logger commons.Logger,
) internal_agent_executor.ToolExecutor {
	return &toolExecutor{
		logger:                 logger,
		mcpClients:             make(map[string]mcp.MCPClient),
		mcpTools:               make(map[string]*mcp.MCPToolInfo),
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

// initializeMCPTool connects to an MCP server and fetches all available tools
func (executor *toolExecutor) initializeMCPTool(ctx context.Context, tool *internal_assistant_entity.AssistantTool, communication internal_type.Communication, tracer internal_adapter_telemetry.Tracer[utils.RapidaStage]) error {
	opts := tool.GetOptions()
	// Get MCP server URL
	serverUrl, err := opts.GetString("mcp.server_url")
	if err != nil {
		return fmt.Errorf("mcp.server_url is required: %w", err)
	}

	// Get or create MCP client for this server
	client, exists := executor.mcpClients[serverUrl]
	if !exists {
		client = mcp.NewDefaultMCPClient(executor.logger, 30*time.Second)
		executor.mcpClients[serverUrl] = client
		executor.logger.Infof("Connected to MCP server: %s", serverUrl)
	}

	// Fetch all available tools from the MCP server
	tools, err := client.ListTools(ctx, serverUrl)
	if err != nil {
		return fmt.Errorf("failed to list tools from MCP server %s: %w", serverUrl, err)
	}

	executor.logger.Infof("Fetched %d tools from MCP server: %s", len(tools), serverUrl)

	// Register each MCP tool
	for _, mcpTool := range tools {
		// Create tool caller for this MCP tool
		caller := mcp.NewDynamicMCPToolCaller(
			executor.logger,
			client,
			serverUrl,
			mcpTool,
		)

		tracer.AddAttributes(ctx, internal_adapter_telemetry.KV{
			K: mcpTool.Name,
			V: internal_adapter_telemetry.StringValue("mcp"),
		})

		// Store the tool
		executor.tools[mcpTool.Name] = caller
		executor.mcpTools[mcpTool.Name] = &mcp.MCPToolInfo{
			ServerUrl:  serverUrl,
			Definition: mcpTool,
		}
		executor.availableToolFunctions = append(executor.availableToolFunctions, mcpTool)
	}

	return nil
}

func (executor *toolExecutor) Initialize(ctx context.Context, communication internal_type.Communication) error {
	ctx, span, _ := communication.Tracer().StartSpan(ctx, utils.AssistantToolConnectStage)
	defer span.EndSpan(ctx, utils.AssistantToolConnectStage)

	start := time.Now()

	// Initialize local tools and MCP connections
	for _, tool := range communication.Assistant().AssistantTools {
		if tool.ExecutionMethod == "mcp" {
			// Handle MCP tool - connect to server and fetch tools
			if err := executor.initializeMCPTool(ctx, tool, communication, span); err != nil {
				executor.logger.Errorf("error while initialize MCP tool: %s", err)
				continue
			}
		} else {
			// Handle local tool
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
			executor.tools[caller.Name()] = caller
			executor.availableToolFunctions = append(executor.availableToolFunctions, tlf)
		}
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
	ctx, span, _ := communication.Tracer().StartSpan(ctx, utils.AssistantToolExecuteStage, internal_adapter_telemetry.MessageKV(message.ContextId()))
	defer span.EndSpan(ctx, utils.AssistantToolExecuteStage)

	start := time.Now()
	metrics := make([]*types.Metric, 0)

	funC, ok := executor.tools[call.GetFunction().GetName()]
	if !ok {
		return internal_type.LLMToolPacket{ContextID: message.ContextId(),
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
	executor.tool(message.ContextId(), map[string]interface{}{
		"id":        call.Id,
		"name":      call.GetFunction().GetName(),
		"arguments": call.GetFunction().GetArguments(),
	}, output.Result, metrics, communication)

	executor.Log(ctx, funC, communication, message.ContextId(), type_enums.RECORD_COMPLETE, int64(time.Since(start)), map[string]interface{}{
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
