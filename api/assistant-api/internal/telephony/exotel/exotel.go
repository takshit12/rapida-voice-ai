package internal_exotel_telephony

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rapidaai/api/assistant-api/config"
	internal_streamers "github.com/rapidaai/api/assistant-api/internal/streamers"
	internal_telephony "github.com/rapidaai/api/assistant-api/internal/telephony"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/types"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type exotelTelephony struct {
	logger commons.Logger
	appCfg *config.AssistantConfig
}

// EventCallback implements [Telephony].
func (tpc *exotelTelephony) Callback(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, assistantConversationId uint64) (string, map[string]interface{}, error) {
	body, err := c.GetRawData() // Extract raw request body
	if err != nil {
		tpc.logger.Errorf("failed to read event body with error %+v", err)
		return "unknown", nil, fmt.Errorf("not implimented")
	}
	tpc.logger.Debugf("event from exotel | body: %s", string(body))
	return "unknown", nil, fmt.Errorf("not implimented")

}

func NewExotelTelephony(config *config.AssistantConfig, logger commons.Logger) (internal_telephony.Telephony, error) {
	return &exotelTelephony{
		logger: logger,
		appCfg: config,
	}, nil
}

func (tpc *exotelTelephony) ClientUrl(
	vaultCredential *protos.VaultCredential,
	opts utils.Option) (*string, error) {
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

func (tpc *exotelTelephony) MakeCall(
	auth types.SimplePrinciple,
	toPhone string,
	fromPhone string,
	assistantId, sessionId uint64,
	vaultCredential *protos.VaultCredential,
	opts utils.Option) (map[string]interface{}, error) {
	clientUrl, err := tpc.ClientUrl(vaultCredential, opts)
	if err != nil {
		return nil, err
	}
	formData := url.Values{}
	formData.Set("From", toPhone)
	formData.Set("CallerId", fromPhone)
	formData.Set("Url", fmt.Sprintf("wss://%s/%s",
		tpc.appCfg.MediaHost,
		internal_telephony.GetAnswerPath("exotel", auth, assistantId, sessionId, toPhone)))
	formData.Set("StatusCallback", fmt.Sprintf("https://%s/%s", tpc.appCfg.MediaHost, internal_telephony.GetEventPath("exotel", auth, assistantId, sessionId)))
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", *clientUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var jsonResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		return nil, err
	}
	return jsonResponse, nil
}

func (tpc *exotelTelephony) ReceiveCall(c *gin.Context, auth types.SimplePrinciple, assistantId uint64, clientNumber string, assistantConversationId uint64) error {
	return nil
}

func (tpc *exotelTelephony) Streamer(c *gin.Context, connection *websocket.Conn, assistantID uint64, assistantVersion string, assistantConversationID uint64) internal_streamers.Streamer {
	return NewExotelWebsocketStreamer(tpc.logger, connection, assistantID,
		assistantVersion,
		assistantConversationID)
}
