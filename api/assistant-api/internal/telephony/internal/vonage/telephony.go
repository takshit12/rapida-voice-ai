// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_vonage_telephony

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rapidaai/api/assistant-api/config"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_conversation_entity "github.com/rapidaai/api/assistant-api/internal/entity/conversations"
	internal_streamers "github.com/rapidaai/api/assistant-api/internal/streamers"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/types"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
	"github.com/vonage/vonage-go-sdk"
	"github.com/vonage/vonage-go-sdk/ncco"
)

type vonageTelephony struct {
	appCfg *config.AssistantConfig
	logger commons.Logger
}

func NewVonageTelephony(config *config.AssistantConfig, logger commons.Logger) (internal_type.Telephony, error) {
	return &vonageTelephony{
		logger: logger,
		appCfg: config,
	}, nil
}

func (tpc *vonageTelephony) CatchAllStatusCallback(ctx *gin.Context) ([]types.Telemetry, error) {
	return nil, nil
}
func (tpc *vonageTelephony) StatusCallback(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, assistantConversationId uint64) ([]types.Telemetry, error) {
	body, err := c.GetRawData() // Extract raw request body
	if err != nil {
		tpc.logger.Errorf("failed to read request body with error %+v", err)
		return nil, fmt.Errorf("failed to read request body")
	}

	// Parse the JSON body
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		tpc.logger.Errorf("failed to parse request body: %+v", err)
		return nil, fmt.Errorf("failed to parse request body")
	}

	// Extract status from payload
	status, ok := payload["status"].(string)
	if !ok {
		tpc.logger.Errorf("status not found or invalid in payload")
		return nil, fmt.Errorf("status not found in payload")
	}
	tpc.logger.Debugf("event processed | status: %s, payload: %+v", status, payload)
	return []types.Telemetry{types.NewMetric("STATUS", status, utils.Ptr("Status of conversation")), types.NewEvent(status, payload)}, nil
}

func (vt *vonageTelephony) OutboundCall(
	auth types.SimplePrinciple,
	toPhone string,
	fromPhone string,
	assistantId, assistantConversationId uint64,
	vaultCredential *protos.VaultCredential,
	opts utils.Option,
) ([]types.Telemetry, error) {
	mtds := []types.Telemetry{
		types.NewMetadata("telephony.toPhone", toPhone),
		types.NewMetadata("telephony.fromPhone", fromPhone),
		types.NewMetadata("telephony.provider", "vonage"),
	}

	cAuth, err := vt.Auth(vaultCredential)
	if err != nil {
		return append(mtds, types.NewMetadata("telephony.error", fmt.Sprintf("authentication error: %s", err.Error())), types.NewMetric("STATUS", "FAILED", utils.Ptr("Status of telephony api"))), err
	}
	ct := vonage.NewVoiceClient(cAuth)

	connectAction := ncco.Ncco{}
	nccoConnect := ncco.ConnectAction{
		EventType: "synchronous",
		EventUrl:  []string{fmt.Sprintf("https://%s/%s", vt.appCfg.PublicAssistantHost, internal_type.GetEventPath("vonage", auth, assistantId, assistantConversationId))},
		Endpoint: []ncco.Endpoint{ncco.WebSocketEndpoint{
			Uri: fmt.Sprintf("wss://%s/%s",
				vt.appCfg.PublicAssistantHost,
				internal_type.GetAnswerPath("vonage", auth, assistantId, assistantConversationId, toPhone)),
			ContentType: "audio/l16;rate=16000",
		}},
	}
	connectAction.AddAction(nccoConnect)
	result, vErr, apiError := ct.CreateCall(
		vonage.CreateCallOpts{
			From: vonage.CallFrom{Type: "phone", Number: fromPhone},
			To:   vonage.CallTo{Type: "phone", Number: toPhone},
			Ncco: connectAction,
		})

	if apiError != nil {
		return append(mtds, types.NewMetadata("telephony.error", fmt.Sprintf("API error: %s", apiError.Error())), types.NewMetric("STATUS", "FAILED", utils.Ptr("Status of telephony api"))), apiError
	}

	if vErr.Error != nil {
		return append(mtds, types.NewMetadata("telephony.error", fmt.Sprintf("Calling error: %v", vErr.Error)), types.NewMetric("STATUS", "FAILED", utils.Ptr("Status of telephony api"))), fmt.Errorf("failed to create call")
	}

	return append(mtds,
		types.NewMetadata("telephony.conversation_uuid", result.ConversationUuid),
		types.NewMetadata("telephony.uuid", result.Uuid),
		types.NewEvent(result.Status, result),
		types.NewMetric("STATUS", "SUCCESS", utils.Ptr("Status of telephony api"))), nil
}

func (vt *vonageTelephony) InboundCall(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, clientNumber string, assistantConversationId uint64) error {
	c.JSON(http.StatusOK, []gin.H{
		{
			"action":    "connect",
			"eventType": "synchronous",
			"eventUrl":  []string{fmt.Sprintf("https://%s/%s", vt.appCfg.PublicAssistantHost, internal_type.GetEventPath("vonage", auth, assistantId, assistantConversationId))},
			"endpoint": []gin.H{
				{
					"type": "websocket",
					"uri": fmt.Sprintf("wss://%s/%s",
						vt.appCfg.PublicAssistantHost,
						internal_type.GetAnswerPath("vonage", auth, assistantId, assistantConversationId, clientNumber)),
					"content-type": "audio/l16;rate=16000",
				},
			},
		},
	})
	return nil
}

func (tpc *vonageTelephony) Streamer(c *gin.Context, connection *websocket.Conn, assistant *internal_assistant_entity.Assistant, assistantConversation *internal_conversation_entity.AssistantConversation, vltC *protos.VaultCredential) internal_streamers.Streamer {
	return NewVonageWebsocketStreamer(tpc.logger, connection, assistant, assistantConversation, vltC)
}

func (tpc *vonageTelephony) ReceiveCall(c *gin.Context) (*string, []types.Telemetry, error) {
	queryParams := make(map[string]string)
	telemetry := []types.Telemetry{}
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	clientNumber, ok := queryParams["from"]
	if !ok || clientNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assistant ID"})
		return nil, telemetry, fmt.Errorf("missing or empty 'from' query parameter")
	}

	if v, ok := queryParams["conversation_uuid"]; ok && v != "" {
		telemetry = append(telemetry,
			types.NewMetadata("telephony.conversation_uuid", v),
		)
	}

	if v, ok := queryParams["uuid"]; ok && v != "" {
		telemetry = append(telemetry,
			types.NewMetadata("telephony.uuid", v),
		)
	}
	return utils.Ptr(clientNumber), append(telemetry, types.NewEvent("webhook", queryParams), types.NewMetric("STATUS", "SUCCESS", utils.Ptr("Status of telephony api"))), nil
}

func (tpc *vonageTelephony) Auth(vaultCredential *protos.VaultCredential) (vonage.Auth, error) {
	privateKey, ok := vaultCredential.GetValue().AsMap()["private_key"]
	if !ok {
		return nil, fmt.Errorf("illegal vault config privateKey is not found")
	}
	applicationId, ok := vaultCredential.GetValue().AsMap()["application_id"]
	if !ok {
		return nil, fmt.Errorf("illegal vault config application_id is not found")
	}
	clientAuth, err := vonage.CreateAuthFromAppPrivateKey(applicationId.(string), []byte(privateKey.(string)))
	if err != nil {
		return nil, err
	}
	return clientAuth, nil
}
