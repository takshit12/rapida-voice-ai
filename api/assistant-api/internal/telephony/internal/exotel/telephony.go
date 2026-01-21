// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_exotel_telephony

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rapidaai/api/assistant-api/config"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_conversation_entity "github.com/rapidaai/api/assistant-api/internal/entity/conversations"
	internal_streamers "github.com/rapidaai/api/assistant-api/internal/streamers"
	internal_exotel "github.com/rapidaai/api/assistant-api/internal/telephony/internal/exotel/internal"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"

	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/types"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type exotelTelephony struct {
	logger commons.Logger
	appCfg *config.AssistantConfig
}

func NewExotelTelephony(config *config.AssistantConfig, logger commons.Logger) (internal_type.Telephony, error) {
	return &exotelTelephony{
		logger: logger,
		appCfg: config,
	}, nil
}

func (tpc *exotelTelephony) CatchAllStatusCallback(ctx *gin.Context) ([]types.Telemetry, error) {
	return nil, nil
}

// EventCallback implements [Telephony].
func (tpc *exotelTelephony) StatusCallback(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, assistantConversationId uint64) ([]types.Telemetry, error) {
	form, err := c.MultipartForm()
	if err != nil {
		tpc.logger.Errorf("failed to parse multipart form-data with error %+v", err)
		return nil, fmt.Errorf("failed to parse multipart form-data")
	}

	eventDetails := make(map[string]interface{})
	for key, values := range form.Value {
		if len(values) > 0 {
			eventDetails[key] = values[0] // Take only the first value for simplicity
		} else {
			eventDetails[key] = nil
		}
	}
	callStatus := eventDetails["Status"]
	return []types.Telemetry{types.NewMetric("STATUS", fmt.Sprintf("%v", callStatus), utils.Ptr("Status of call or update")), types.NewEvent(fmt.Sprintf("%v", callStatus), eventDetails)}, nil

}
func (tpc *exotelTelephony) ClientUrl(vaultCredential *protos.VaultCredential, opts utils.Option) (*string, error) {
	accountSid, ok := vaultCredential.GetValue().AsMap()["account_sid"]
	if !ok {
		return nil, fmt.Errorf("illegal vault config accountSid is not found")
	}
	clientId, ok := vaultCredential.GetValue().AsMap()["client_id"]
	if !ok {
		return nil, fmt.Errorf("illegal vault config client_id not found")
	}
	authToken, ok := vaultCredential.GetValue().AsMap()["client_secret"]
	if !ok {
		return nil, fmt.Errorf("illegal vault config")
	}
	return utils.Ptr(fmt.Sprintf("https://%s:%s@api.exotel.com/v1/Accounts/%s/Calls/connect.json",
		clientId.(string), authToken.(string), accountSid.(string))), nil

}

func (tpc *exotelTelephony) AppUrl(
	vaultCredential *protos.VaultCredential,
	opts utils.Option) (*string, error) {
	accountSid, ok := vaultCredential.GetValue().AsMap()["account_sid"]
	if !ok {
		return nil, fmt.Errorf("illegal vault config accountSid is not found")
	}
	app_id, err := opts.GetString("app_id")
	if err != nil {
		return nil, fmt.Errorf("illegal app_id option is not found")
	}
	return utils.Ptr(fmt.Sprintf("http://my.exotel.com/%s/exoml/start_voice/%s", accountSid.(string), app_id)), nil

}

func (tpc *exotelTelephony) OutboundCall(
	auth types.SimplePrinciple,
	// customer number
	toPhone string,
	// exo number
	fromPhone string,
	assistantId, assistantConversationId uint64,
	vaultCredential *protos.VaultCredential,
	opts utils.Option) ([]types.Telemetry, error) {
	mtds := []types.Telemetry{
		types.NewMetadata("telephony.toPhone", toPhone),
		types.NewMetadata("telephony.fromPhone", fromPhone),
		types.NewMetadata("telephony.provider", "exotel"),
	}
	clientUrl, err := tpc.ClientUrl(vaultCredential, opts)
	if err != nil {
		return append(mtds, types.NewEvent("FAILED", "Failed to build url, check credentials"), &types.Metric{Name: "STATUS", Value: "FAILED", Description: "Status of telephony api"}), err
	}

	appUrl, err := tpc.AppUrl(vaultCredential, opts)
	if err != nil {
		return append(mtds, types.NewEvent("FAILED", "Failed to build app url"), &types.Metric{Name: "STATUS", Value: "FAILED", Description: "Status of telephony api"}), err
	}

	formData := url.Values{}
	formData.Set("From", toPhone)
	formData.Set("CallerId", fromPhone)
	formData.Set("To", fromPhone)
	formData.Set("Url", *appUrl)
	formData.Set("StatusCallback", fmt.Sprintf("https://%s/%s", tpc.appCfg.PublicAssistantHost, internal_type.GetEventPath("exotel", auth, assistantId, assistantConversationId)))
	// for exotel there is no way to set dynamic path so pass it as custom filed
	formData.Set("CustomField", internal_type.GetAnswerPath("exotel", auth, assistantId,
		assistantConversationId,
		toPhone,
	))

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", *clientUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return append(mtds, &types.Metric{Name: "STATUS", Value: "FAILED", Description: "Status of telephony api"}), err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return append(mtds, &types.Metric{Name: "STATUS", Value: "FAILED", Description: "Status of telephony api"}), err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return append(mtds, &types.Metric{Name: "STATUS", Value: "FAILED", Description: "Status of telephony api"}), err
	}
	if resp.StatusCode != http.StatusOK {
		tpc.logger.Errorf("Unexpected HTTP Status: %d, Response Body: %s\n", resp.StatusCode, string(bodyBytes))
		return append(mtds, types.NewEvent("Failed", string(bodyBytes)), &types.Metric{Name: "STATUS", Value: "FAILED", Description: "Status of telephony API HTTP error"}), fmt.Errorf("status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var jsonResponse internal_exotel.MakeCallResponse
	// Wrap the JSON decoding in a detailed error message
	if err := json.Unmarshal(bodyBytes, &jsonResponse); err != nil {
		return append(mtds, types.NewEvent(jsonResponse.Call.Status, "Failed to decode response"), &types.Metric{Name: "STATUS", Value: "FAILED", Description: "Status of transaction"}), err
	}
	return append(mtds, types.NewMetadata("telephony.uuid", jsonResponse.Call.Sid), types.NewEvent(jsonResponse.Call.Status, jsonResponse), &types.Metric{Name: "STATUS", Value: "SUCCESS", Description: "Status of telephony api"}), nil
}

func (tpc *exotelTelephony) InboundCall(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, clientNumber string, assistantConversationId uint64) error {
	response := map[string]string{
		"url": fmt.Sprintf("wss://%s/%s",
			tpc.appCfg.PublicAssistantHost,
			internal_type.GetAnswerPath("exotel", auth, assistantId, assistantConversationId, clientNumber)),
	}
	c.JSON(http.StatusOK, response)
	return nil
}

func (tpc *exotelTelephony) Streamer(c *gin.Context, connection *websocket.Conn, assistant *internal_assistant_entity.Assistant, conversation *internal_conversation_entity.AssistantConversation, vlt *protos.VaultCredential) internal_streamers.Streamer {
	return NewExotelWebsocketStreamer(tpc.logger, connection, assistant, conversation, vlt)
}

func (tpc *exotelTelephony) ReceiveCall(c *gin.Context) (*string, []types.Telemetry, error) {
	queryParams := make(map[string]string)
	telemetry := []types.Telemetry{}
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	socketUrl, ok := queryParams["CustomField"]
	if ok {
		response := map[string]string{"url": fmt.Sprintf("wss://%s/%s", tpc.appCfg.PublicAssistantHost, socketUrl)}
		c.JSON(http.StatusOK, response)
		return nil, telemetry, fmt.Errorf("outbound call triggered")
	}

	clientNumber, ok := queryParams["CallFrom"]
	if !ok || clientNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid caller"})
		return nil, telemetry, fmt.Errorf("missing or empty 'from' query parameter")
	}

	return utils.Ptr(clientNumber), telemetry, nil
}
