// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_websocket

import "encoding/json"

// =============================================================================
// WebSocket Message Types - Similar to AgentKit gRPC pattern
// =============================================================================

// WSMessageType defines the type of message and what data structure to expect
type WSMessageType string

const (
	// Request types (client -> server)
	WSTypeConfiguration WSMessageType = "configuration" // Data: WSConfigurationData
	WSTypeUserMessage   WSMessageType = "user_message"  // Data: WSUserMessageData

	// Response types (server -> client)
	WSTypeAssistantMessage WSMessageType = "assistant_message" // Data: WSAssistantMessageData
	WSTypeStream           WSMessageType = "stream"            // Data: WSStreamData
	WSTypeInterruption     WSMessageType = "interruption"      // Data: WSInterruptionData
	WSTypeError            WSMessageType = "error"             // Data: WSErrorData

	// Control types (bidirectional)
	WSTypePing WSMessageType = "ping" // Data: nil
	WSTypePong WSMessageType = "pong" // Data: nil
)

// =============================================================================
// Request/Response Envelope
// =============================================================================

// WSRequest represents an outgoing WebSocket message with typed data
type WSRequest struct {
	Type      WSMessageType `json:"type"`
	Timestamp int64         `json:"timestamp"`
	Data      interface{}   `json:"data,omitempty"`
}

// WSResponse represents an incoming WebSocket message with typed data
type WSResponse struct {
	Type    WSMessageType   `json:"type"`
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *WSErrorData    `json:"error,omitempty"`
}

// =============================================================================
// Data Structures for each message type
// =============================================================================

// WSConfigurationData contains initial connection configuration
// Used with: WSTypeConfiguration
type WSConfigurationData struct {
	AssistantID         uint64                 `json:"assistant_id"`
	ConversationID      uint64                 `json:"conversation_id"`
	AssistantDefinition *WSAssistantDefinition `json:"assistant,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	Args                map[string]interface{} `json:"args,omitempty"`
	Options             map[string]interface{} `json:"options,omitempty"`
}

// WSAssistantDefinition contains assistant metadata
type WSAssistantDefinition struct {
	AssistantID uint64 `json:"assistant_id"`
	Name        string `json:"name,omitempty"`
}

// WSUserMessageData contains user message content
// Used with: WSTypeUserMessage
type WSUserMessageData struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Completed bool   `json:"completed"`
	Timestamp int64  `json:"timestamp"`
}

// WSAssistantMessageData contains assistant response content
// Used with: WSTypeAssistantMessage
type WSAssistantMessageData struct {
	ID        string                     `json:"id"`
	Message   *WSAssistantMessageContent `json:"message"`
	Completed bool                       `json:"completed"`
	Metrics   []*WSMetric                `json:"metrics,omitempty"`
}

// WSAssistantMessageContent represents the message content (text or audio)
type WSAssistantMessageContent struct {
	Type    string `json:"type"` // "text" or "audio"
	Content string `json:"content,omitempty"`
	Audio   []byte `json:"audio,omitempty"`
}

// WSStreamData contains streaming text chunk
// Used with: WSTypeStream
type WSStreamData struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Index   int    `json:"index,omitempty"`
}

// WSInterruptionData contains interruption signal
// Used with: WSTypeInterruption
type WSInterruptionData struct {
	ID      string  `json:"id,omitempty"`
	Source  string  `json:"source,omitempty"` // "word", "vad"
	StartAt float64 `json:"start_at,omitempty"`
	EndAt   float64 `json:"end_at,omitempty"`
}

// WSErrorData contains error information
// Used with: WSTypeError or in WSResponse.Error
type WSErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// WSMetric contains metric information
type WSMetric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit,omitempty"`
}
