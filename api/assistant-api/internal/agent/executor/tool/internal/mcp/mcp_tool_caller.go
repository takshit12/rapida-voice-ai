// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	internal_tool "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

// mcpToolCaller implements the ToolCaller interface for Model Context Protocol (MCP) tools.
// It allows invoking external MCP-compliant tools through a standardized protocol.
type mcpToolCaller struct {
	logger      commons.Logger
	toolOptions *internal_assistant_entity.AssistantTool
	mcpClient   MCPClient
	serverUrl   string
	toolName    string
}

// MCPClient defines the interface for communicating with MCP servers
type MCPClient interface {
	// CallTool invokes an MCP tool with the given name and arguments
	CallTool(ctx context.Context, serverUrl string, toolName string, arguments map[string]interface{}) (*MCPToolResponse, error)
	// GetToolDefinition retrieves the definition of an MCP tool
	GetToolDefinition(ctx context.Context, serverUrl string, toolName string) (*protos.FunctionDefinition, error)
	// ListTools retrieves all available tools from the MCP server
	ListTools(ctx context.Context, serverUrl string) ([]*protos.FunctionDefinition, error)
}

// MCPToolResponse represents the response from an MCP tool invocation
type MCPToolResponse struct {
	Success bool                   `json:"success"`
	Result  interface{}            `json:"result,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// MCPToolInfo holds information about an MCP tool
type MCPToolInfo struct {
	ServerUrl  string
	Definition *protos.FunctionDefinition
}

// NewMCPToolCaller creates a new MCP tool caller instance
func NewMCPToolCaller(
	logger commons.Logger,
	toolOptions *internal_assistant_entity.AssistantTool,
	mcpClient MCPClient,
	communication internal_type.Communication,
) (internal_tool.ToolCaller, error) {
	opts := toolOptions.GetOptions()

	// Extract MCP server URL from tool options
	serverUrl, err := opts.GetString("mcp.server_url")
	if err != nil {
		return nil, fmt.Errorf("mcp.server_url is required for MCP tools: %v", err)
	}

	// Extract MCP tool name (may differ from assistant tool name)
	toolName, err := opts.GetString("mcp.tool_name")
	if err != nil {
		// Default to assistant tool name if not specified
		toolName = toolOptions.Name
	}

	return &mcpToolCaller{
		logger:      logger,
		toolOptions: toolOptions,
		mcpClient:   mcpClient,
		serverUrl:   serverUrl,
		toolName:    toolName,
	}, nil
}

// Id returns the unique identifier of the tool
func (m *mcpToolCaller) Id() uint64 {
	return m.toolOptions.Id
}

// Name returns the human-readable name of the tool
func (m *mcpToolCaller) Name() string {
	return m.toolOptions.Name
}

// ExecutionMethod returns the execution strategy (mcp)
func (m *mcpToolCaller) ExecutionMethod() string {
	return m.toolOptions.ExecutionMethod
}

// Definition returns the function definition for the MCP tool
func (m *mcpToolCaller) Definition() (*protos.FunctionDefinition, error) {
	// Try to get definition from MCP server if available
	if m.mcpClient != nil {
		ctx := context.Background()
		if def, err := m.mcpClient.GetToolDefinition(ctx, m.serverUrl, m.toolName); err == nil && def != nil {
			return def, nil
		}
		m.logger.Warnf("Failed to get MCP tool definition from server, falling back to local definition")
	}

	// Fallback to local definition from tool options
	definition := &protos.FunctionDefinition{
		Name:       m.toolOptions.Name,
		Parameters: &protos.FunctionParameter{},
	}
	if m.toolOptions.Description != nil && *m.toolOptions.Description != "" {
		definition.Description = *m.toolOptions.Description
	}
	if err := utils.Cast(m.toolOptions.Fields, definition.Parameters); err != nil {
		return nil, fmt.Errorf("failed to cast tool fields to function parameters: %w", err)
	}
	return definition, nil
}

// Call executes the MCP tool with the given arguments
func (m *mcpToolCaller) Call(
	ctx context.Context,
	pkt internal_type.LLMPacket,
	toolId string,
	args string,
	communication internal_type.Communication,
) internal_type.LLMToolPacket {
	m.logger.Debugf("Calling MCP tool: %s with args: %s", m.Name(), args)

	// Parse arguments from JSON string to map
	var arguments map[string]interface{}
	if err := json.Unmarshal([]byte(args), &arguments); err != nil {
		m.logger.Errorf("Failed to parse MCP tool arguments: %v", err)
		return internal_type.LLMToolPacket{
			Name:      m.Name(),
			ContextID: pkt.ContextId(),
			Action:    protos.AssistantConversationAction_MCP_TOOL_CALL,
			Result:    m.Result("Invalid arguments format", false),
		}
	}

	// Call the MCP server
	response, err := m.mcpClient.CallTool(ctx, m.serverUrl, m.toolName, arguments)
	if err != nil {
		m.logger.Errorf("MCP tool call failed: %v", err)
		return internal_type.LLMToolPacket{
			Name:      m.Name(),
			ContextID: pkt.ContextId(),
			Action:    protos.AssistantConversationAction_MCP_TOOL_CALL,
			Result:    m.Result(fmt.Sprintf("Tool execution failed: %v", err), false),
		}
	}

	// Check if the MCP call was successful
	if !response.Success {
		errorMsg := response.Error
		if errorMsg == "" {
			errorMsg = "Unknown error occurred"
		}
		return internal_type.LLMToolPacket{
			Name:      m.Name(),
			ContextID: pkt.ContextId(),
			Action:    protos.AssistantConversationAction_MCP_TOOL_CALL,
			Result:    m.Result(errorMsg, false),
		}
	}

	// Build result map from response
	result := make(map[string]interface{})
	result["success"] = true
	result["status"] = "SUCCESS"

	if response.Data != nil && len(response.Data) > 0 {
		result["data"] = response.Data
	} else if response.Result != nil {
		result["data"] = response.Result
	}

	return internal_type.LLMToolPacket{
		Name:      m.Name(),
		ContextID: pkt.ContextId(),
		Action:    protos.AssistantConversationAction_MCP_TOOL_CALL,
		Result:    result,
	}
}

// Result formats the tool execution result
func (m *mcpToolCaller) Result(msg string, success bool) map[string]interface{} {
	if success {
		return map[string]interface{}{
			"data":    msg,
			"success": true,
			"status":  "SUCCESS",
		}
	}
	return map[string]interface{}{
		"error":   msg,
		"success": false,
		"status":  "FAIL",
	}
}

// dynamicMCPToolCaller is used for tools discovered dynamically from MCP servers
type dynamicMCPToolCaller struct {
	logger     commons.Logger
	mcpClient  MCPClient
	serverUrl  string
	definition *protos.FunctionDefinition
}

// NewDynamicMCPToolCaller creates a tool caller for a dynamically discovered MCP tool
func NewDynamicMCPToolCaller(
	logger commons.Logger,
	mcpClient MCPClient,
	serverUrl string,
	definition *protos.FunctionDefinition,
) internal_tool.ToolCaller {
	return &dynamicMCPToolCaller{
		logger:     logger,
		mcpClient:  mcpClient,
		serverUrl:  serverUrl,
		definition: definition,
	}
}

// Id returns a generated ID (0 for dynamic tools)
func (d *dynamicMCPToolCaller) Id() uint64 {
	return 0
}

// Name returns the tool name
func (d *dynamicMCPToolCaller) Name() string {
	return d.definition.Name
}

// ExecutionMethod returns "mcp"
func (d *dynamicMCPToolCaller) ExecutionMethod() string {
	return "mcp"
}

// Definition returns the function definition
func (d *dynamicMCPToolCaller) Definition() (*protos.FunctionDefinition, error) {
	return d.definition, nil
}

// Call executes the MCP tool
func (d *dynamicMCPToolCaller) Call(
	ctx context.Context,
	pkt internal_type.LLMPacket,
	toolId string,
	args string,
	communication internal_type.Communication,
) internal_type.LLMToolPacket {
	d.logger.Debugf("Calling dynamic MCP tool: %s with args: %s", d.Name(), args)

	// Parse arguments from JSON string to map
	var arguments map[string]interface{}
	if err := json.Unmarshal([]byte(args), &arguments); err != nil {
		d.logger.Errorf("Failed to parse MCP tool arguments: %v", err)
		return internal_type.LLMToolPacket{
			Name:      d.Name(),
			ContextID: pkt.ContextId(),
			Action:    protos.AssistantConversationAction_MCP_TOOL_CALL,
			Result:    d.Result("Invalid arguments format", false),
		}
	}

	// Call the MCP server
	response, err := d.mcpClient.CallTool(ctx, d.serverUrl, d.Name(), arguments)
	if err != nil {
		d.logger.Errorf("MCP tool call failed: %v", err)
		return internal_type.LLMToolPacket{
			Name:      d.Name(),
			ContextID: pkt.ContextId(),
			Action:    protos.AssistantConversationAction_MCP_TOOL_CALL,
			Result:    d.Result(fmt.Sprintf("Tool execution failed: %v", err), false),
		}
	}

	// Check if the MCP call was successful
	if !response.Success {
		errorMsg := response.Error
		if errorMsg == "" {
			errorMsg = "Unknown error occurred"
		}
		return internal_type.LLMToolPacket{
			Name:      d.Name(),
			ContextID: pkt.ContextId(),
			Action:    protos.AssistantConversationAction_MCP_TOOL_CALL,
			Result:    d.Result(errorMsg, false),
		}
	}

	// Build result map from response
	result := make(map[string]interface{})
	result["success"] = true
	result["status"] = "SUCCESS"

	if response.Data != nil && len(response.Data) > 0 {
		result["data"] = response.Data
	} else if response.Result != nil {
		result["data"] = response.Result
	}

	return internal_type.LLMToolPacket{
		Name:      d.Name(),
		ContextID: pkt.ContextId(),
		Action:    protos.AssistantConversationAction_MCP_TOOL_CALL,
		Result:    result,
	}
}

// Result formats the tool execution result
func (d *dynamicMCPToolCaller) Result(msg string, success bool) map[string]interface{} {
	if success {
		return map[string]interface{}{
			"data":    msg,
			"success": true,
			"status":  "SUCCESS",
		}
	}
	return map[string]interface{}{
		"error":   msg,
		"success": false,
		"status":  "FAIL",
	}
}
