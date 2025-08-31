package web_handler

import (
	"context"

	internal_service "github.com/lexatic/web-backend/api/web-api/internal/service"
	internal_provider_service "github.com/lexatic/web-backend/api/web-api/internal/service/provider"
	config "github.com/lexatic/web-backend/config"
	commons "github.com/lexatic/web-backend/pkg/commons"
	"github.com/lexatic/web-backend/pkg/connectors"
	"github.com/lexatic/web-backend/pkg/utils"
	web_api "github.com/lexatic/web-backend/protos/lexatic-backend"
)

type webProviderApi struct {
	WebApi
	cfg             *config.AppConfig
	logger          commons.Logger
	postgres        connectors.PostgresConnector
	redis           connectors.RedisConnector
	providerService internal_service.ProviderService
}

type webProviderRPCApi struct {
	webProviderApi
}

type webProviderGRPCApi struct {
	webProviderApi
}

func NewProviderGRPC(config *config.AppConfig, logger commons.Logger, postgres connectors.PostgresConnector, redis connectors.RedisConnector) web_api.ProviderServiceServer {
	return &webProviderGRPCApi{
		webProviderApi{
			WebApi:          NewWebApi(config, logger, postgres, redis),
			cfg:             config,
			logger:          logger,
			postgres:        postgres,
			redis:           redis,
			providerService: internal_provider_service.NewProviderService(logger, postgres),
		},
	}
}

// GetAllProvider implements lexatic_backend.ProviderServiceServer.
func (w *webProviderGRPCApi) GetAllModelProvider(ctx context.Context, gat *web_api.GetAllModelProviderRequest) (*web_api.GetAllModelProviderResponse, error) {
	providers, err := w.providerService.GetAllModelProvider(ctx, gat.GetCriterias())
	if err != nil {
		return utils.Error[web_api.GetAllModelProviderResponse](
			err,
			"Unable to get tool providers, please try again in sometime.")
	}
	prds := []*web_api.Provider{}
	err = utils.Cast(providers, &prds)
	if err != nil {
		w.logger.Errorf("unable to cast tool provider to proto object %v", err)
	}
	return utils.PaginatedSuccess[web_api.GetAllModelProviderResponse, []*web_api.Provider](
		uint32(len(prds)), 0,
		prds)
}

func (w *webProviderGRPCApi) GetAllToolProvider(ctx context.Context, gat *web_api.GetAllToolProviderRequest) (*web_api.GetAllToolProviderResponse, error) {
	providers, err := w.providerService.GetAllToolProvider(ctx, gat.GetCriterias())
	if err != nil {
		return utils.Error[web_api.GetAllToolProviderResponse](
			err,
			"Unable to get tool providers, please try again in sometime.")
	}
	prds := []*web_api.ToolProvider{}
	err = utils.Cast(providers, &prds)
	if err != nil {
		w.logger.Errorf("unable to cast tool provider to proto object %v", err)
	}
	return utils.PaginatedSuccess[web_api.GetAllToolProviderResponse, []*web_api.ToolProvider](
		uint32(len(prds)), 1,
		prds)
}
