package internal_twilio_telephony

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rapidaai/api/assistant-api/config"
	internal_streamers "github.com/rapidaai/api/assistant-api/internal/streamers"
	internal_telephony "github.com/rapidaai/api/assistant-api/internal/telephony"
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

func NewTwilioTelephony(
	config *config.AssistantConfig,
	logger commons.Logger) (internal_telephony.Telephony, error) {
	return &twilioTelephony{
		appCfg: config,
		logger: logger,
	}, nil
}

func (tpc *twilioTelephony) Callback(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, assistantConversationId uint64) (string, map[string]interface{}, error) {
	body, err := c.GetRawData() // Extract raw request body
	if err != nil {
		tpc.logger.Errorf("failed to read event body with error %+v", err)
		return "unknown", nil, fmt.Errorf("not implimented")
	}
	tpc.logger.Debugf("event from twilio | body: %s", string(body))
	return "unknown", nil, fmt.Errorf("not implimented")

}
func (tpc *twilioTelephony) TwilioClientParam(vaultCredential *protos.VaultCredential,
	opts utils.Option) (*twilio.ClientParams, error) {
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

func (tpc *twilioTelephony) TwilioClient(vaultCredential *protos.VaultCredential,
	opts utils.Option) (*twilio.RestClient, error) {
	clientParams, err := tpc.TwilioClientParam(vaultCredential, opts)
	if err != nil {
		return nil, err
	}
	return twilio.NewRestClientWithParams(*clientParams), nil
}

func (tpc *twilioTelephony) MakeCall(
	auth types.SimplePrinciple,
	toPhone string,
	fromPhone string,
	assistantId, sessionId uint64,
	vaultCredential *protos.VaultCredential,
	opts utils.Option,
) (map[string]interface{}, error) {
	client, err := tpc.TwilioClient(vaultCredential, opts)
	if err != nil {
		return nil, err
	}
	callParams := &openapi.CreateCallParams{}
	callParams.SetTo(toPhone)
	callParams.SetFrom(fromPhone)
	callParams.SetStatusCallbackEvent([]string{
		fmt.Sprintf("https://%s/%s", tpc.appCfg.MediaHost, internal_telephony.GetEventPath("twilio", auth, assistantId, sessionId)),
	})
	callParams.SetTwiml(
		tpc.CreateTwinML(
			tpc.appCfg.MediaHost,
			internal_telephony.GetAnswerPath("twilio", auth, assistantId,
				sessionId,
				toPhone,
			),
			assistantId,
			toPhone),
	)
	resp, err := client.Api.CreateCall(callParams)
	if err != nil {
		return nil, err
	}
	// Convert entire response to JSON, then to map
	jsonData, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("error marshaling response to JSON: %v", err)
	}

	var responseMap map[string]interface{}
	err = json.Unmarshal(jsonData, &responseMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON to map: %v", err)
	}
	return responseMap, nil
}

func (tpc *twilioTelephony) CreateTwinML(mediaServer string, path string, assistantId uint64, clientNumber string) string {

	return fmt.Sprintf(`
	    <Response>
		 	<Connect>
	        	<Stream url="wss://%s/%s">
					<Parameter name="assistant_id" value="%d"/>
					<Parameter name="client_number" value="%s"/>
				</Stream>
			</Connect>
	    </Response>
	`,
		mediaServer,
		path,
		assistantId,
		clientNumber,
	)
}

func (tpc *twilioTelephony) ReceiveCall(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, clientNumber string, assistantConversationId uint64) error {
	c.Data(http.StatusOK, "text/xml", []byte(
		tpc.CreateTwinML(
			tpc.appCfg.MediaHost,
			fmt.Sprintf("v1/talk/twilio/prj/%d/%s/%d/%s",
				assistantId,
				clientNumber, assistantConversationId, auth.GetCurrentToken()), assistantId, clientNumber),
	))
	return nil
}

func (tpc *twilioTelephony) Streamer(c *gin.Context, connection *websocket.Conn, assistantID uint64, assistantVersion string, assistantConversationID uint64) internal_streamers.Streamer {
	return NewTwilioWebsocketStreamer(tpc.logger, connection, assistantID,
		assistantVersion,
		assistantConversationID)
}
