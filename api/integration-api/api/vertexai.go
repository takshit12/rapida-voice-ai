// Rapida – Open Source Voice AI Orchestration Platform
// Copyright (C) 2023-2025 Prashant Srivastav <prashant@rapida.ai>
// Licensed under a modified GPL-2.0. See the LICENSE file for details.
package integration_api

import (
	"context"

	"github.com/gin-gonic/gin"
	config "github.com/rapidaai/api/integration-api/config"
	internal_callers "github.com/rapidaai/api/integration-api/internal/caller"
	internal_vertexai_callers "github.com/rapidaai/api/integration-api/internal/caller/vertexai"
	commons "github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/connectors"
	"github.com/rapidaai/protos"
)

type vertexaiIntegrationApi struct {
	integrationApi
}

type vertexaiIntegrationRPCApi struct {
	vertexaiIntegrationApi
}

type vertexaiIntegrationGRPCApi struct {
	vertexaiIntegrationApi
}

// Embedding implements protos.VertexaiServiceServer.
func (googAi *vertexaiIntegrationGRPCApi) Embedding(c context.Context, irRequest *protos.EmbeddingRequest) (*protos.EmbeddingResponse, error) {
	return googAi.integrationApi.Embedding(
		c, irRequest,
		"VERTEXAI",
		internal_vertexai_callers.NewEmbeddingCaller(googAi.logger, irRequest.GetCredential()),
	)
}

func NewVertexaiRPC(config *config.IntegrationConfig, logger commons.Logger, postgres connectors.PostgresConnector) *vertexaiIntegrationRPCApi {
	return &vertexaiIntegrationRPCApi{
		vertexaiIntegrationApi{
			integrationApi: NewInegrationApi(config, logger, postgres),
		},
	}
}

func NewVertexaiGRPC(config *config.IntegrationConfig, logger commons.Logger, postgres connectors.PostgresConnector) protos.VertexAiServiceServer {
	return &vertexaiIntegrationGRPCApi{
		vertexaiIntegrationApi{
			integrationApi: NewInegrationApi(config, logger, postgres),
		},
	}
}

// all the rpc handler
func (vertexaiRPC *vertexaiIntegrationRPCApi) Generate(c *gin.Context) {
	vertexaiRPC.logger.Debugf("Generate from rpc with gin context %v", c)
}
func (vertexaiRPC *vertexaiIntegrationRPCApi) Chat(c *gin.Context) {
	vertexaiRPC.logger.Debugf("Chat from rpc with gin context %v", c)
}

// StreamChat implements protos.VertexaiServiceServer.
func (vertexaiGRPc *vertexaiIntegrationGRPCApi) StreamChat(irRequest *protos.ChatRequest, stream protos.VertexAiService_StreamChatServer) error {
	return vertexaiGRPc.integrationApi.StreamChat(
		irRequest,
		stream.Context(),
		"VERTEXAI",
		internal_vertexai_callers.NewLargeLanguageCaller(vertexaiGRPc.logger, irRequest.GetCredential()),
		stream.Send,
	)
}

// all grpc handler
func (googAi *vertexaiIntegrationGRPCApi) Chat(c context.Context, irRequest *protos.ChatRequest) (*protos.ChatResponse, error) {
	return googAi.integrationApi.Chat(
		c, irRequest,
		"VERTEXAI",
		internal_vertexai_callers.NewLargeLanguageCaller(googAi.logger, irRequest.GetCredential()),
	)
}

func (vertexaiGRPC *vertexaiIntegrationGRPCApi) VerifyCredential(c context.Context, irRequest *protos.VerifyCredentialRequest) (*protos.VerifyCredentialResponse, error) {
	vertexaiCaller := internal_vertexai_callers.NewVerifyCredentialCaller(vertexaiGRPC.logger, irRequest.Credential)
	st, err := vertexaiCaller.CredentialVerifier(
		c,
		&internal_callers.CredentialVerifierOptions{},
	)
	if err != nil {
		vertexaiGRPC.logger.Errorf("verify credential response with error %v", err)
		return &protos.VerifyCredentialResponse{
			Code:         401,
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}
	return &protos.VerifyCredentialResponse{
		Code:     200,
		Success:  true,
		Response: st,
	}, nil
}
