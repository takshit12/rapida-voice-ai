package assistant_talk_api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	internal_factory "github.com/rapidaai/api/assistant-api/internal/factory"
	telephony "github.com/rapidaai/api/assistant-api/internal/factory/telephony"
	internal_services "github.com/rapidaai/api/assistant-api/internal/services"
	"github.com/rapidaai/pkg/types"
	type_enums "github.com/rapidaai/pkg/types/enums"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

func (cApi *ConversationApi) Callback(c *gin.Context) {
	iAuth, isAuthenticated := types.GetAuthPrinciple(c)
	if !isAuthenticated {
		cApi.logger.Debugf("illegal unable to authenticate")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthenticated request"})
		return
	}
	assistantIdStr := c.Param("assistantId")
	assistantId, err := strconv.ParseUint(assistantIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assistant ID"})
		return
	}

	conversationIdStr := c.Param("conversationId")
	conversationId, err := strconv.ParseUint(conversationIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	tlp := c.Param("telephony")
	_telephony, err := telephony.GetTelephony(
		telephony.Telephony(tlp),
		cApi.cfg,
		cApi.logger,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telephony"})
		return
	}

	eventType, payload, err := _telephony.Callback(c, iAuth, assistantId, conversationId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event to process"})
		return
	}

	_, err = cApi.assistantConversationService.
		CreateConversationTelephonyEvent(c, iAuth, tlp, conversationId, eventType, payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to process telephony event callback."})
		return
	}
}

// CallReciever handles incoming calls for the given assistant.
// @Router /v1/call/:assistantId [post]
// @Summary Recieve call for given assistant
// @Produce json
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
func (cApi *ConversationApi) CallReciever(c *gin.Context) {
	iAuth, isAuthenticated := types.GetAuthPrinciple(c)
	if !isAuthenticated {
		cApi.logger.Debugf("illegal unable to authenticate")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthenticated request"})
		return
	}

	assistantIdStr := c.Param("assistantId")
	assistantId, err := strconv.ParseUint(assistantIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assistant ID"})
		return
	}

	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	clientNumber, ok := queryParams["Caller"]
	if !ok {
		cApi.logger.Debugf("Missing 'Caller' number in Twilio request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'Caller' number"})
		return
	}

	tlp := c.Param("telephony")
	_telephony, err := telephony.GetTelephony(
		telephony.Telephony(tlp),
		cApi.cfg,
		cApi.logger,
	)

	assistant, err := cApi.
		assistantService.
		Get(c,
			iAuth,
			assistantId,
			utils.
				GetVersionDefinition("latest"),
			&internal_services.
				GetAssistantOption{
				InjectPhoneDeployment: true,
			})
	if err != nil {
		cApi.logger.Debugf("illegal unable to find assistant %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to initiate talker"})
		return
	}

	conversation, err := cApi.
		assistantConversationService.
		CreateConversation(
			c,
			iAuth,
			internal_factory.Identifier(utils.PhoneCall, c, iAuth, clientNumber),
			assistant.Id,
			assistant.AssistantProviderId,
			type_enums.DIRECTION_INBOUND,
			utils.PhoneCall,
		)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to initiate talker"})
		return
	}

	cApi.
		assistantConversationService.
		ApplyConversationMetrics(
			c, iAuth, conversation.Id, []*types.Metric{types.NewStatusMetric(type_enums.RECORD_CONNECTED)},
		)
	err = _telephony.ReceiveCall(c, iAuth, assistant.Id, clientNumber, conversation.Id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to initiate talker"})
		return
	}

}

func (cApi *ConversationApi) CallTalker(c *gin.Context) {
	start := time.Now()
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	websocketConnection, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		cApi.logger.Errorf("illegal while upgrading vonage talker with error %v", err)
		return
	}
	cApi.logger.Benchmark("ConversationApi.VonageCallTalker.upgradeConnection", time.Since(start))

	auth, isAuthenticated := types.GetAuthPrinciple(c)
	if !isAuthenticated {
		cApi.logger.Errorf("illegal unable to authenticate")
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthenticated request"})
		return
	}

	cApi.logger.Benchmark("conversationapi.VonageCallTalker.GetAuthPrinciple", time.Since(start))
	// Extract the client source from the stream context
	assistantId, err := strconv.ParseUint(c.Param("assistantId"), 10, 64)
	if err != nil {
		cApi.logger.Errorf("Invalid assistantId: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assistantId"})
		return
	}

	conversationId, err := strconv.ParseUint(c.Param("conversationId"), 10, 64)
	if err != nil {
		cApi.logger.Errorf("Invalid conversationId: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversationId"})
		return
	}

	identifier := c.Param("identifier")
	cApi.logger.Debugf("starting a call talker for with %s and params %+v", identifier, c.Params)

	tlp := c.Param("telephony")
	_telephony, err := telephony.GetTelephony(
		telephony.Telephony(tlp),
		cApi.cfg,
		cApi.logger,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telephony"})
		return
	}

	talker, err := internal_factory.GetTalker(
		utils.PhoneCall,
		c,
		cApi.cfg,
		cApi.logger,
		cApi.postgres,
		cApi.opensearch,
		cApi.redis,
		cApi.storage,
		_telephony.Streamer(
			c,
			websocketConnection,
			assistantId, "latest", conversationId,
		),
	)
	if err != nil {
		cApi.logger.Errorf("illegal to get talker %v", err)
		return
	}
	cApi.logger.Benchmark("conversationapi.VonageCallTalker.GetTalker", time.Since(start))
	cidentifier := internal_factory.
		Identifier(utils.PhoneCall, c, auth, identifier)

	err = talker.Connect(
		c, auth, cidentifier, &protos.AssistantConversationConfiguration{
			AssistantConversationId: conversationId,
			Assistant: &protos.AssistantDefinition{
				AssistantId: assistantId,
				Version:     "latest",
			},
		},
	)
	if err != nil {
		cApi.logger.Errorf("illegal to connect with talker %v", err)
		return
	}
	talker.Talk(
		c,
		auth,
		cidentifier,
	)
}
