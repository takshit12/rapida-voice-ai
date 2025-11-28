package internal_vonage_telephony

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rapidaai/api/assistant-api/config"
	internal_streamers "github.com/rapidaai/api/assistant-api/internal/streamers"
	internal_telephony "github.com/rapidaai/api/assistant-api/internal/telephony"
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

func NewVonageTelephony(config *config.AssistantConfig, logger commons.Logger) (internal_telephony.Telephony, error) {
	return &vonageTelephony{
		logger: logger,
		appCfg: config,
	}, nil
}

func (tpc *vonageTelephony) Callback(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, assistantConversationId uint64) (string, map[string]interface{}, error) {
	body, err := c.GetRawData() // Extract raw request body
	if err != nil {
		tpc.logger.Errorf("failed to read event body with error %+v", err)
		return "unknown", nil, fmt.Errorf("not implimented")
	}
	tpc.logger.Debugf("event from exotel | body: %s", string(body))
	return "unknown", nil, fmt.Errorf("not implimented")

}

func (vt *vonageTelephony) Auth(
	vaultCredential *protos.VaultCredential,
	opts utils.Option) (vonage.Auth, error) {
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
		return nil, fmt.Errorf("illegal vault config application_id is not found")
	}
	return clientAuth, nil
}

func (vt *vonageTelephony) MakeCall(
	auth types.SimplePrinciple,
	toPhone string,
	fromPhone string,
	assistantId, assistantConversationId uint64,
	vaultCredential *protos.VaultCredential,
	opts utils.Option,
) (map[string]interface{}, error) {
	cAuth, err := vt.Auth(vaultCredential, opts)
	if err != nil {
		return nil, err
	}
	ct := vonage.NewVoiceClient(cAuth)

	connectAction := ncco.Ncco{}
	nccoConnect := ncco.ConnectAction{
		EventType: "synchronous",
		EventUrl:  []string{fmt.Sprintf("https://%s/%s", vt.appCfg.MediaHost, internal_telephony.GetEventPath("vonage", auth, assistantId, assistantConversationId))},
		Endpoint: []ncco.Endpoint{ncco.WebSocketEndpoint{
			Uri: fmt.Sprintf("wss://%s/%s",
				vt.appCfg.MediaHost,
				internal_telephony.GetAnswerPath("vonage", auth, assistantId, assistantConversationId, toPhone)),
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
		vt.logger.Errorf("error while calling vonage %+v", apiError)
		return nil, apiError
	}

	if vErr.Error != nil {
		vt.logger.Errorf("error while calling vonage %+v", vErr.Error)
		return nil, fmt.Errorf("unable to make call with vonage")
	}
	return map[string]interface{}{
		"uuid":              result.Uuid,
		"status":            result.Status,
		"direction":         result.Direction,
		"conversation_uuid": result.ConversationUuid,
	}, nil
}

func (tpc *vonageTelephony) ReceiveCall(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, clientNumber string, assistantConversationId uint64) error {
	return nil
}

func (tpc *vonageTelephony) Streamer(c *gin.Context, connection *websocket.Conn, assistantID uint64, assistantVersion string, assistantConversationID uint64) internal_streamers.Streamer {
	return NewVonageWebsocketStreamer(tpc.logger, connection, assistantID,
		assistantVersion,
		assistantConversationID)
}
