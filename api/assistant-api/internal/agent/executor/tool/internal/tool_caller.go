// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_tool

import (
	"context"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"

	"github.com/rapidaai/protos"
)

// ToolCaller defines the contract for invoking a tool/function that can be
// executed by the agent runtime. Implementations encapsulate tool metadata,
// execution semantics, and request/response handling.
//
// A ToolCaller is responsible for:
//   - Exposing a unique identifier and human-readable name
//   - Providing a function definition consumable by the LLM/runtime
//   - Declaring the execution method (e.g., sync, async, streaming)
//   - Executing the tool call and returning response packets
type ToolCaller interface {
	// Id returns the unique identifier of the tool.
	Id() uint64

	// Name returns the human-readable name of the tool.
	Name() string

	// Definition returns the function definition describing the tool's
	// input parameters and behavior, or an error if the definition
	// cannot be constructed.
	Definition() (*protos.FunctionDefinition, error)

	// ExecutionMethod returns the execution strategy used by the tool
	// (for example, synchronous or asynchronous execution).
	ExecutionMethod() string

	// Call executes the tool with the given arguments and communication
	// context. It returns a slice of Packets representing the tool's
	// response(s) to be consumed by the agent runtime.
	Call(ctx context.Context, pkt internal_type.LLMPacket, toolId string, args string, communication internal_type.Communication) internal_type.LLMToolPacket
}
