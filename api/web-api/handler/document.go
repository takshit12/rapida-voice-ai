package web_handler

import (
	"context"
	"errors"

	"github.com/lexatic/web-backend/config"
	document_client "github.com/lexatic/web-backend/pkg/clients/document"
	"github.com/lexatic/web-backend/pkg/commons"
	"github.com/lexatic/web-backend/pkg/connectors"
	"github.com/lexatic/web-backend/pkg/types"
	"github.com/lexatic/web-backend/pkg/utils"
	knowledge_api "github.com/lexatic/web-backend/protos/lexatic-backend"
)

type indexerApi struct {
	cfg                  *config.AppConfig
	logger               commons.Logger
	postgres             connectors.PostgresConnector
	redis                connectors.RedisConnector
	indexerServiceClient document_client.IndexerServiceClient
}

type indexerGrpcApi struct {
	indexerApi
}

func NewDocumentGRPCApi(config *config.AppConfig, logger commons.Logger,
	postgres connectors.PostgresConnector,
	redis connectors.RedisConnector) knowledge_api.DocumentServiceServer {
	return &indexerGrpcApi{
		indexerApi{
			cfg:                  config,
			logger:               logger,
			postgres:             postgres,
			redis:                redis,
			indexerServiceClient: document_client.NewIndexerServiceClient(config, logger, redis),
		},
	}
}

func (iApi *indexerApi) IndexKnowledgeDocument(ctx context.Context, cer *knowledge_api.IndexKnowledgeDocumentRequest) (*knowledge_api.IndexKnowledgeDocumentResponse, error) {
	iApi.logger.Debugf("index document request %v, %v", cer, ctx)
	iAuth, isAuthenticated := types.GetSimplePrincipleGRPC(ctx)
	if !isAuthenticated || !iAuth.HasProject() {
		iApi.logger.Errorf("unauthenticated request for invoke")
		return utils.Error[knowledge_api.IndexKnowledgeDocumentResponse](
			errors.New("unauthenticated request for invoke"),
			"Please provider valid service credentials to perfom invoke, read docs @ docs.rapida.ai",
		)
	}

	return iApi.indexerServiceClient.IndexKnowledgeDocument(ctx, iAuth, cer)
}
