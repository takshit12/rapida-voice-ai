package assistant_router

import (
	"github.com/gin-gonic/gin"
	assistantApi "github.com/rapidaai/api/assistant-api/api/assistant"
	assistantDeploymentApi "github.com/rapidaai/api/assistant-api/api/assistant-deployment"
	assistantTalkApi "github.com/rapidaai/api/assistant-api/api/talk"
	"github.com/rapidaai/api/assistant-api/config"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/connectors"
	workflow_api "github.com/rapidaai/protos"
	"google.golang.org/grpc"
)

func AssistantApiRoute(
	Cfg *config.AssistantConfig,
	S *grpc.Server,
	Logger commons.Logger,
	Postgres connectors.PostgresConnector,
	Redis connectors.RedisConnector,
	Opensearch connectors.OpenSearchConnector,
) {
	workflow_api.RegisterAssistantServiceServer(S,
		assistantApi.NewAssistantGRPCApi(Cfg,
			Logger,
			Postgres,
			Redis,
			Opensearch,
			Opensearch,
		))
}

func AssistantDeploymentApiRoute(Cfg *config.AssistantConfig,
	S *grpc.Server,
	Logger commons.Logger,
	Postgres connectors.PostgresConnector) {
	workflow_api.RegisterAssistantDeploymentServiceServer(S,
		assistantDeploymentApi.NewAssistantDeploymentGRPCApi(Cfg,
			Logger,
			Postgres,
		))
}

func AssistantConversationApiRoute(
	Cfg *config.AssistantConfig,
	S *grpc.Server,
	Logger commons.Logger,
	Postgres connectors.PostgresConnector,
	Redis connectors.RedisConnector,
	Opensearch connectors.OpenSearchConnector,
) {
	workflow_api.RegisterTalkServiceServer(S,
		assistantTalkApi.NewConversationGRPCApi(Cfg,
			Logger,
			Postgres,
			Redis,
			Opensearch,
			Opensearch,
		))
}

func TalkCallbackApiRoute(
	cfg *config.AssistantConfig, engine *gin.Engine, logger commons.Logger,
	postgres connectors.PostgresConnector,
	redis connectors.RedisConnector,
	opensearch connectors.OpenSearchConnector) {
	apiv1 := engine.Group("v1/talk")
	talkRpcApi := assistantTalkApi.NewConversationApi(cfg,
		logger,
		postgres,
		redis,
		opensearch,
		opensearch,
	)
	{
		// global
		apiv1.GET("/:telephony/event/:assistantId", talkRpcApi.UnviersalCallback)
		apiv1.POST("/:telephony/event/:assistantId", talkRpcApi.UnviersalCallback)

		// session event
		apiv1.GET("/:telephony/usr/event/:assistantId/:conversationId/:authorization/:x-auth-id/:x-project-id", talkRpcApi.Callback)
		apiv1.POST("/:telephony/usr/event/:assistantId/:conversationId/:authorization/:x-auth-id/:x-project-id", talkRpcApi.Callback)
		apiv1.GET("/:telephony/prj/event/:assistantId/:conversationId/:x-api-key", talkRpcApi.Callback)
		apiv1.POST("/:telephony/prj/event/:assistantId/:conversationId/:x-api-key", talkRpcApi.Callback)

		// twilio whatsapp
		apiv1.POST("/twilio/whatsapp/:assistantToken", talkRpcApi.WhatsappReciever)

		// vonage call
		apiv1.GET("/:telephony/call/:assistantId", talkRpcApi.CallReciever)
		apiv1.GET("/:telephony/usr/:assistantId/:identifier/:conversationId/:authorization/:x-auth-id/:x-project-id", talkRpcApi.CallTalker)
		apiv1.GET("/:telephony/prj/:assistantId/:identifier/:conversationId/:x-api-key", talkRpcApi.CallTalker)

	}
}
