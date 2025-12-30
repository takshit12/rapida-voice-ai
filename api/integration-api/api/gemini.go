// Rapida – Open Source Voice AI Orchestration Platform
// Copyright (C) 2023-2025 Prashant Srivastav <prashant@rapida.ai>
// Licensed under a modified GPL-2.0. See the LICENSE file for details.
package integration_api

import (
	"context"

	"github.com/gin-gonic/gin"
	config "github.com/rapidaai/api/integration-api/config"
	internal_callers "github.com/rapidaai/api/integration-api/internal/caller"
	internal_gemini_callers "github.com/rapidaai/api/integration-api/internal/caller/gemini"
	commons "github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/connectors"
	integration_api "github.com/rapidaai/protos"
)

type geminiIntegrationApi struct {
	integrationApi
}

type geminiIntegrationRPCApi struct {
	geminiIntegrationApi
}

type geminiIntegrationGRPCApi struct {
	geminiIntegrationApi
}

// Embedding implements protos.GoogleServiceServer.
func (googAi *geminiIntegrationGRPCApi) Embedding(c context.Context, irRequest *integration_api.EmbeddingRequest) (*integration_api.EmbeddingResponse, error) {
	return googAi.integrationApi.Embedding(
		c, irRequest,
		"GEMINI",
		internal_gemini_callers.NewEmbeddingCaller(googAi.logger, irRequest.GetCredential()),
	)
}

func NewGeminiRPC(config *config.IntegrationConfig, logger commons.Logger, postgres connectors.PostgresConnector) *geminiIntegrationRPCApi {
	return &geminiIntegrationRPCApi{
		geminiIntegrationApi{
			integrationApi: NewInegrationApi(config, logger, postgres),
		},
	}
}

func NewGeminiGRPC(config *config.IntegrationConfig, logger commons.Logger, postgres connectors.PostgresConnector) integration_api.GeminiServiceServer {
	return &geminiIntegrationGRPCApi{
		geminiIntegrationApi{
			integrationApi: NewInegrationApi(config, logger, postgres),
		},
	}
}

// all the rpc handler
func (geminiRPC *geminiIntegrationRPCApi) Generate(c *gin.Context) {
	geminiRPC.logger.Debugf("Generate from rpc with gin context %v", c)
}
func (geminiRPC *geminiIntegrationRPCApi) Chat(c *gin.Context) {
	geminiRPC.logger.Debugf("Chat from rpc with gin context %v", c)
}

// StreamChat implements protos.GoogleServiceServer.
func (geminiGRPc *geminiIntegrationGRPCApi) StreamChat(irRequest *integration_api.ChatRequest, stream integration_api.GeminiService_StreamChatServer) error {
	return geminiGRPc.integrationApi.StreamChat(
		irRequest,

		stream.Context(),
		"GEMINI",
		internal_gemini_callers.NewLargeLanguageCaller(geminiGRPc.logger, irRequest.GetCredential()),
		stream.Send,
	)
}

// all grpc handler
func (googAi *geminiIntegrationGRPCApi) Chat(c context.Context, irRequest *integration_api.ChatRequest) (*integration_api.ChatResponse, error) {
	return googAi.integrationApi.Chat(
		c, irRequest,
		"GEMINI",
		internal_gemini_callers.NewLargeLanguageCaller(googAi.logger, irRequest.GetCredential()),
	)
}

func (geminiGRPC *geminiIntegrationGRPCApi) VerifyCredential(c context.Context, irRequest *integration_api.VerifyCredentialRequest) (*integration_api.VerifyCredentialResponse, error) {
	geminiCaller := internal_gemini_callers.NewVerifyCredentialCaller(geminiGRPC.logger, irRequest.Credential)
	st, err := geminiCaller.CredentialVerifier(
		c,
		&internal_callers.CredentialVerifierOptions{},
	)
	if err != nil {
		geminiGRPC.logger.Errorf("verify credential response with error %v", err)
		return &integration_api.VerifyCredentialResponse{
			Code:         401,
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}
	return &integration_api.VerifyCredentialResponse{
		Code:     200,
		Success:  true,
		Response: st,
	}, nil
}
