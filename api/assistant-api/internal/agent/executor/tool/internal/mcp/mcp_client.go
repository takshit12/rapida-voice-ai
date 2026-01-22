// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/protos"
)

// DefaultMCPClient is the default implementation of MCPClient using HTTP/REST
type DefaultMCPClient struct {
	logger     commons.Logger
	httpClient *http.Client
	timeout    time.Duration
}

// MCPToolListResponse represents the response from listing tools
type MCPToolListResponse struct {
	Success bool                         `json:"success"`
	Tools   []*protos.FunctionDefinition `json:"tools"`
	Error   string                       `json:"error,omitempty"`
}

// MCPToolCallRequest represents the request payload for calling an MCP tool
type MCPToolCallRequest struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPToolDefinitionRequest represents the request to get tool definition
type MCPToolDefinitionRequest struct {
	ToolName string `json:"tool_name"`
}

// NewDefaultMCPClient creates a new default MCP client
func NewDefaultMCPClient(logger commons.Logger, timeout time.Duration) MCPClient {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &DefaultMCPClient{
		logger: logger,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// CallTool invokes an MCP tool via HTTP POST request
func (c *DefaultMCPClient) CallTool(
	ctx context.Context,
	serverUrl string,
	toolName string,
	arguments map[string]interface{},
) (*MCPToolResponse, error) {
	c.logger.Debugf("Calling MCP tool '%s' at server: %s", toolName, serverUrl)

	// Prepare request payload
	request := MCPToolCallRequest{
		ToolName:  toolName,
		Arguments: arguments,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MCP request: %w", err)
	}

	// Build the endpoint URL (assuming /tools/call endpoint)
	endpoint := fmt.Sprintf("%s/tools/call", serverUrl)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MCP HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return &MCPToolResponse{
			Success: false,
			Error:   fmt.Sprintf("MCP server returned status %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	// Parse response
	var response MCPToolResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// If parsing fails, try to return raw body as result
		c.logger.Warnf("Failed to parse MCP response as JSON, returning raw response")
		return &MCPToolResponse{
			Success: true,
			Result:  string(body),
		}, nil
	}

	return &response, nil
}

// GetToolDefinition retrieves the definition of an MCP tool
func (c *DefaultMCPClient) GetToolDefinition(
	ctx context.Context,
	serverUrl string,
	toolName string,
) (*protos.FunctionDefinition, error) {
	c.logger.Debugf("Getting MCP tool definition for '%s' from server: %s", toolName, serverUrl)

	// Prepare request payload
	request := MCPToolDefinitionRequest{
		ToolName: toolName,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal definition request: %w", err)
	}

	// Build the endpoint URL (assuming /tools/definition endpoint)
	endpoint := fmt.Sprintf("%s/tools/definition", serverUrl)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MCP HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var definition protos.FunctionDefinition
	if err := json.Unmarshal(body, &definition); err != nil {
		return nil, fmt.Errorf("failed to parse tool definition: %w", err)
	}

	return &definition, nil
}

// ListTools retrieves all available tools from the MCP server
func (c *DefaultMCPClient) ListTools(
	ctx context.Context,
	serverUrl string,
) ([]*protos.FunctionDefinition, error) {
	c.logger.Debugf("Listing all tools from MCP server: %s", serverUrl)

	// Build the endpoint URL (assuming /tools/list endpoint)
	endpoint := fmt.Sprintf("%s/tools/list", serverUrl)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MCP HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response MCPToolListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse tool list: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("MCP server error: %s", response.Error)
	}

	return response.Tools, nil
}
