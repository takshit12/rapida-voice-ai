// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_type

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_conversation_entity "github.com/rapidaai/api/assistant-api/internal/entity/conversations"
	internal_streamers "github.com/rapidaai/api/assistant-api/internal/streamers"
	"github.com/rapidaai/pkg/types"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

// any telephony integration must impliment this interface to provide consistent behaviour
type Telephony interface {
	// streamer
	Streamer(c *gin.Context, connection *websocket.Conn, assistant *internal_assistant_entity.Assistant, assistantConversation *internal_conversation_entity.AssistantConversation, vltC *protos.VaultCredential) internal_streamers.Streamer
	//  event callback for a conversation
	StatusCallback(ctx *gin.Context, auth types.SimplePrinciple, assistantId, assistantConversationId uint64) ([]types.Telemetry, error)
	// catch all event callback
	CatchAllStatusCallback(ctx *gin.Context) ([]types.Telemetry, error)

	//
	ReceiveCall(c *gin.Context) (client *string, telemetry []types.Telemetry, err error)
	OutboundCall(auth types.SimplePrinciple, toPhone string, fromPhone string, assistantId, assistantConversationId uint64, vaultCredential *protos.VaultCredential, opts utils.Option) ([]types.Telemetry, error)
	InboundCall(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, clientNumber string, assistantConversationId uint64) error
}

func GetAnswerPath(provider string, auth types.SimplePrinciple, assistantId uint64, assistantConversationId uint64, toPhone string) string {
	switch auth.Type() {
	case "project":
		return fmt.Sprintf("v1/talk/%s/prj/%d/%s/%d/%s",
			provider,
			assistantId,
			toPhone,
			assistantConversationId,
			auth.GetCurrentToken())
	default:
		return fmt.Sprintf("v1/talk/%s/usr/%d/%s/%d/%s/%d/%d",
			provider,
			assistantId,
			toPhone,
			assistantConversationId,
			auth.GetCurrentToken(),
			*auth.GetUserId(),
			*auth.GetCurrentProjectId())
	}
}

func GetEventPath(provider string, auth types.SimplePrinciple, assistantId, assistantConversationId uint64) string {
	switch auth.Type() {
	case "project":
		return fmt.Sprintf("v1/talk/%s/prj/event/%d/%d/%s",
			provider,
			assistantId,
			assistantConversationId,
			auth.GetCurrentToken())
	default:
		return fmt.Sprintf("v1/talk/%s/usr/event/%d/%d/%s/%d/%d",
			provider,
			assistantId,
			assistantConversationId,
			auth.GetCurrentToken(),
			*auth.GetUserId(),
			*auth.GetCurrentProjectId())
	}
}
