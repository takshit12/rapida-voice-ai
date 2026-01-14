// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_agent_mcp_tool

import (
	"context"

	internal_tool "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool/internal"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/types"
)

// Just a placeholder for MCP specific tool caller interface
type MCPCaller interface {

	// list all available tool callers
	List() ([]internal_tool.ToolCaller, error)

	// call the tools
	Call(ctx context.Context, tool internal_tool.ToolCaller, messageId string, args string, communication internal_type.Communication) (map[string]interface{}, []*types.Metric)
}
