// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_type

import "github.com/rapidaai/protos"

// Packet represents a generic request packet handled by the adapter layer.
// Concrete packet types (e.g., FlushPacket, InterruptionPacket) are used to
// signal specific control actions within a given context.
type Packet interface {
	ContextId() string
}

// FlushPacket represents a request to flush or reset state associated
// with a specific context.
type FlushPacket struct {
	// ContextID identifies the context to be flushed.
	ContextID string
}

// ContextId returns the identifier of the context associated with this flush request.
func (f FlushPacket) ContextId() string {
	return f.ContextID
}

// InterruptionPacket represents a request to interrupt ongoing processing
// within a specific context.
type InterruptionPacket struct {
	// ContextID identifies the context to be interrupted.
	ContextID string
}

// ContextId returns the identifier of the context associated with this interruption request.
func (f InterruptionPacket) ContextId() string {
	return f.ContextID
}

type TextPacket struct {
	// ContextID identifies the context to be flushed.
	ContextID string

	Text string
}

// ContextId returns the identifier of the context associated with this interruption request.
func (f TextPacket) ContextId() string {
	return f.ContextID
}

// h
type KnowledgeRetriveOption struct {
	EmbeddingProviderCredential *protos.VaultCredential
	RetrievalMethod             string
	TopK                        uint32
	ScoreThreshold              float32
}

type KnowledgeContextResult struct {
	ID         string                 `json:"id"`
	DocumentID string                 `json:"document_id"`
	Metadata   map[string]interface{} `json:"metadata"`
	Content    string                 `json:"content"`
	Score      float64                `json:"score"`
}
