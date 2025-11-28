package internal_telephony

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	internal_streamers "github.com/rapidaai/api/assistant-api/internal/streamers"
	"github.com/rapidaai/pkg/types"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

// any telephony integration must impliment this interface to provide consistent behaviour
type Telephony interface {
	// streamer
	Streamer(c *gin.Context, connection *websocket.Conn, assistantID uint64, assistantVersion string, assistantConversationID uint64) internal_streamers.Streamer

	// for creating call throght telephony
	MakeCall(
		auth types.SimplePrinciple,
		toPhone string,
		fromPhone string,
		assistantId, assistantConversationId uint64,
		vaultCredential *protos.VaultCredential,
		opts utils.Option,
	) (map[string]interface{}, error)

	// for handling event callback
	Callback(ctx *gin.Context, auth types.SimplePrinciple, assistantId, assistantConversationId uint64) (string, map[string]interface{}, error)

	//
	ReceiveCall(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, clientNumber string, assistantConversationId uint64) error
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
