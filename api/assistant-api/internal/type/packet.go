// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_type

import (
	"github.com/rapidaai/pkg/types"
	"github.com/rapidaai/protos"
)

// Packet represents a generic request packet handled by the adapter layer.
// Concrete packet types (e.g., FlushPacket, InterruptionPacket) are used to
// signal specific control actions within a given context.
type Packet interface {
	ContextId() string
}

// Wrapper for message packet
type MessagePacket interface {
	Packet
	Role() string
	Content() string
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

	//
	Source string

	// start of interruption
	StartAt float64

	// end of interruption
	EndAt float64
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

// MetricPacket represents a request to send metrics within a specific context.
type MetricPacket struct {
	// ContextID identifies the context to be flushed.
	ContextID string

	// Metrics holds the list of metrics to be sent within the specified context.
	Metrics []*types.Metric
}

func (f MetricPacket) ContextId() string {
	return f.ContextID
}

type LLMStreamPacket struct {

	// contextID identifies the context to be flushed.
	ContextID string

	// message
	Message *types.Message
}

func (f LLMStreamPacket) ContextId() string {
	return f.ContextID
}

type StaticPacket struct {
	// contextID identifies the context to be flushed.
	ContextID string

	// message
	Text string
}

func (f StaticPacket) ContextId() string {
	return f.ContextID
}

func (f StaticPacket) Content() string {
	return f.Text
}

func (f StaticPacket) Role() string {
	return "rapida"
}

// llm packate

type LLMPacket struct {
	// contextID identifies the context to be flushed.
	ContextID string

	// message
	Message *types.Message
}

func (f LLMPacket) Content() string {
	return f.Message.String()
}

func (f LLMPacket) Role() string {
	return "assistant"
}

func (f LLMPacket) ContextId() string {
	return f.ContextID
}

type LLMToolPacket struct {

	// contextID identifies the context to be flushed.
	ContextID string

	// action
	Action protos.AssistantConversationAction_ActionType

	// result
	Result map[string]interface{}
}

func (f LLMToolPacket) ContextId() string {
	return f.ContextID
}

type TextToSpeechPacket struct {

	// contextID identifies the context to be flushed.
	ContextID string

	// audio chunk
	AudioChunk []byte
}

func (f TextToSpeechPacket) ContextId() string {
	return f.ContextID
}

type TextToSpeechFlushPacket struct {
	// contextID identifies the context to be flushed.
	ContextID string
}

func (f TextToSpeechFlushPacket) ContextId() string {
	return f.ContextID
}

type UserTextPacket struct {
	// contextID identifies the context to be flushed.
	ContextID string

	// text
	Text string
}

func (f UserTextPacket) ContextId() string {
	return f.ContextID
}

func (f UserTextPacket) Content() string {
	return f.Text
}

func (f UserTextPacket) Role() string {
	return "user"
}

type EndOfSpeechPacket struct {
	// contextID identifies the context to be flushed.
	ContextID string

	Speech string
}

func (f EndOfSpeechPacket) ContextId() string {
	return f.ContextID
}

type SpeechToTextPacket struct {
	ContextID string

	// script
	Script string

	// confidence
	Confidence float64

	// language
	Language string

	// interim
	Interim bool
}

func (f SpeechToTextPacket) ContextId() string {
	return f.ContextID
}

//

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
