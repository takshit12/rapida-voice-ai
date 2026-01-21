// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_twilio_telephony

import (
	"fmt"
	"net/http"
	"net/url"

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
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type twilioTelephony struct {
	appCfg *config.AssistantConfig
	logger commons.Logger
}

func NewTwilioTelephony(config *config.AssistantConfig, logger commons.Logger) (internal_type.Telephony, error) {
	return &twilioTelephony{
		appCfg: config,
		logger: logger,
	}, nil
}

func (tpc *twilioTelephony) client(vaultCredential *protos.VaultCredential) (*twilio.RestClient, error) {
	clientParams, err := tpc.clientParam(vaultCredential)
	if err != nil {
		return nil, err
	}
	return twilio.NewRestClientWithParams(*clientParams), nil
}

func (tpc *twilioTelephony) clientParam(vaultCredential *protos.VaultCredential) (*twilio.ClientParams, error) {
	accountSid, ok := vaultCredential.GetValue().AsMap()["account_sid"]
	if !ok {
		return nil, fmt.Errorf("illegal vault config accountSid is not found")
	}
	authToken, ok := vaultCredential.GetValue().AsMap()["account_token"]
	if !ok {
		return nil, fmt.Errorf("illegal vault config account_token not found")
	}
	return &twilio.ClientParams{
		Username: accountSid.(string),
		Password: authToken.(string),
	}, nil
}

func (tpc *twilioTelephony) CatchAllStatusCallback(ctx *gin.Context) ([]types.Telemetry, error) {
	return nil, nil
}
func (tpc *twilioTelephony) StatusCallback(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, assistantConversationId uint64) ([]types.Telemetry, error) {
	body, err := c.GetRawData() // Extract raw request body
	if err != nil {
		tpc.logger.Errorf("failed to read event body with error %+v", err)
		return nil, fmt.Errorf("not implimented")
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		tpc.logger.Errorf("failed to parse body with error %+v", err)
		return nil, fmt.Errorf("failed to parse request body")
	}

	eventDetails := make(map[string]interface{})
	for key, value := range values {
		if len(value) > 0 {
			eventDetails[key] = value[0]
		} else {
			eventDetails[key] = nil
		}
	}

	callStatusOrStreamEvent := eventDetails["CallStatus"]
	if streamEvent, ok := eventDetails["StreamEvent"]; ok {
		callStatusOrStreamEvent = streamEvent
	}
	return []types.Telemetry{types.NewMetric("STATUS", fmt.Sprintf("%v", callStatusOrStreamEvent), utils.Ptr("Status of conversation")), types.NewEvent(fmt.Sprintf("%v", callStatusOrStreamEvent), eventDetails)}, nil

}

func (tpc *twilioTelephony) OutboundCall(auth types.SimplePrinciple, toPhone string, fromPhone string, assistantId, assistantConversationId uint64, vaultCredential *protos.VaultCredential, opts utils.Option) ([]types.Telemetry, error) {
	mtds := []types.Telemetry{
		types.NewMetadata("telephony.toPhone", toPhone),
		types.NewMetadata("telephony.fromPhone", fromPhone),
		types.NewMetadata("telephony.provider", "twilio"),
	}

	client, err := tpc.client(vaultCredential)
	if err != nil {
		return append(mtds, types.NewMetadata("telephony.error", fmt.Sprintf("authentication error: %s", err.Error())), types.NewMetric("STATUS", "FAILED", utils.Ptr("Status of telephony api"))), err
	}
	callParams := &openapi.CreateCallParams{}
	callParams.SetTo(toPhone)
	callParams.SetFrom(fromPhone)
	callParams.SetStatusCallback(
		fmt.Sprintf("https://%s/%s", tpc.appCfg.PublicAssistantHost, internal_type.GetEventPath("twilio", auth, assistantId, assistantConversationId)),
	)
	callParams.SetStatusCallbackEvent([]string{
		"initiated", "ringing", "answered", "completed",
	})
	callParams.SetStatusCallbackMethod("POST")
	callParams.SetTwiml(
		tpc.CreateTwinML(
			tpc.appCfg.PublicAssistantHost,
			fmt.Sprintf("%d__%d", assistantId, assistantConversationId),
			internal_type.GetAnswerPath("twilio", auth, assistantId,
				assistantConversationId,
				toPhone,
			),
			fmt.Sprintf("https://%s/%s", tpc.appCfg.PublicAssistantHost, internal_type.GetEventPath("twilio", auth, assistantId, assistantConversationId)),
			assistantId,
			toPhone),
	)
	resp, err := client.Api.CreateCall(callParams)
	if err != nil || resp.Status == nil || resp.Sid == nil {
		return append(mtds, types.NewMetadata("telephony.error", fmt.Sprintf("API error: %s", err.Error())), types.NewMetric("STATUS", "FAILED", utils.Ptr("Status of telephony api"))), err
	}

	return append(mtds,
		types.NewMetadata("telephony.uuid", *resp.Sid),
		types.NewEvent(*resp.Status, resp),
		types.NewMetric("STATUS", "SUCCESS", utils.Ptr("Status of telephony api"))), nil
}

func (tpc *twilioTelephony) CreateTwinML(mediaServer string, name, path string, callback string, assistantId uint64, clientNumber string) string {
	return fmt.Sprintf(`
	    <Response>
		 	<Connect>
	        	<Stream url="wss://%s/%s" name="%s" statusCallback="%s" statusCallbackEvent="initiated ringing answered completed">
					<Parameter name="assistant_id" value="%d"/>
					<Parameter name="client_number" value="%s"/>
				</Stream>
			</Connect>
	    </Response>
	`,
		mediaServer,
		path,
		name,
		callback,
		assistantId,
		clientNumber,
	)
}

func (tpc *twilioTelephony) InboundCall(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, clientNumber string, assistantConversationId uint64) error {
	c.Data(http.StatusOK, "text/xml", []byte(
		tpc.CreateTwinML(
			tpc.appCfg.PublicAssistantHost,
			fmt.Sprintf("%d__%d", assistantId, assistantConversationId),
			fmt.Sprintf("v1/talk/twilio/prj/%d/%s/%d/%s",
				assistantId,
				clientNumber, assistantConversationId, auth.GetCurrentToken()),
			fmt.Sprintf("https://%s/%s", tpc.appCfg.PublicAssistantHost, internal_type.GetEventPath("twilio", auth, assistantId, assistantConversationId)),
			assistantId, clientNumber),
	))
	return nil
}

func (tpc *twilioTelephony) Streamer(c *gin.Context, connection *websocket.Conn, assistant *internal_assistant_entity.Assistant, conversation *internal_conversation_entity.AssistantConversation, vlt *protos.VaultCredential) internal_streamers.Streamer {
	return NewTwilioWebsocketStreamer(tpc.logger, connection, assistant, conversation, vlt)
}

func (tpc *twilioTelephony) ReceiveCall(c *gin.Context) (*string, []types.Telemetry, error) {
	queryParams := make(map[string]string)
	telemetry := []types.Telemetry{}
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	clientNumber, ok := queryParams["From"]
	if !ok || clientNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assistant ID"})
		return nil, telemetry, fmt.Errorf("missing or empty 'from' query parameter")
	}

	if v, ok := queryParams["CallSid"]; ok && v != "" {
		telemetry = append(telemetry,
			types.NewMetadata("telephony.uuid", v),
		)
	}
	return utils.Ptr(clientNumber), append(telemetry, types.NewEvent("webhook", queryParams), types.NewMetric("STATUS", "SUCCESS", utils.Ptr("Status of telephony api"))), nil
}
