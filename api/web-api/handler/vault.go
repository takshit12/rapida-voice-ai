package web_handler

import (
	"context"

	internal_connect "github.com/lexatic/web-backend/api/web-api/internal/connect"
	internal_connects "github.com/lexatic/web-backend/api/web-api/internal/connect"
	internal_service "github.com/lexatic/web-backend/api/web-api/internal/service"
	internal_vault_service "github.com/lexatic/web-backend/api/web-api/internal/service/vault"
	config "github.com/lexatic/web-backend/config"
	integration_client "github.com/lexatic/web-backend/pkg/clients/integration"
	commons "github.com/lexatic/web-backend/pkg/commons"
	"github.com/lexatic/web-backend/pkg/connectors"
	"github.com/lexatic/web-backend/pkg/types"
	"github.com/lexatic/web-backend/pkg/utils"
	web_api "github.com/lexatic/web-backend/protos/lexatic-backend"
)

type webVaultApi struct {
	cfg               *config.AppConfig
	logger            commons.Logger
	postgres          connectors.PostgresConnector
	redis             connectors.RedisConnector
	vaultService      internal_service.VaultService
	integrationClient integration_client.IntegrationServiceClient
	hubspotConnect    internal_connect.HubspotConnect
}

type webVaultRPCApi struct {
	webVaultApi
}

type webVaultGRPCApi struct {
	webVaultApi
}

func NewVaultRPC(config *config.AppConfig, oauthCfg *config.OAuthConfig, logger commons.Logger, postgres connectors.PostgresConnector, redis connectors.RedisConnector) *webVaultRPCApi {
	return &webVaultRPCApi{
		webVaultApi{
			cfg:               config,
			logger:            logger,
			postgres:          postgres,
			vaultService:      internal_vault_service.NewVaultService(logger, postgres),
			integrationClient: integration_client.NewIntegrationServiceClientGRPC(config, logger, redis),
			hubspotConnect:    internal_connects.NewHubspotConnect(config, oauthCfg, logger, postgres),
		},
	}
}

func NewVaultGRPC(config *config.AppConfig, oauthCfg *config.OAuthConfig, logger commons.Logger, postgres connectors.PostgresConnector, redis connectors.RedisConnector) web_api.VaultServiceServer {
	return &webVaultGRPCApi{
		webVaultApi{
			cfg:               config,
			logger:            logger,
			postgres:          postgres,
			redis:             redis,
			vaultService:      internal_vault_service.NewVaultService(logger, postgres),
			integrationClient: integration_client.NewIntegrationServiceClientGRPC(config, logger, redis),
			hubspotConnect:    internal_connects.NewHubspotConnect(config, oauthCfg, logger, postgres),
		},
	}
}

func (wVault *webVaultGRPCApi) CreateProviderCredential(ctx context.Context, irRequest *web_api.CreateProviderCredentialRequest) (*web_api.GetCredentialResponse, error) {
	wVault.logger.Debugf("CreateProviderCredential from grpc with requestPayload %v, %v", irRequest, ctx)
	iAuth, isAuthenticated := types.GetAuthPrincipleGPRC(ctx)
	if !isAuthenticated {
		wVault.logger.Errorf("CreateProviderCredential from grpc with unauthenticated request")
		return utils.AuthenticateError[web_api.GetCredentialResponse]()
	}
	// first verify the credentials if not verified then return to user and say its not good credentials

	// verified, err := wVault.integrationClient.VerifyCredential(ctx, iAuth,
	// 	irRequest.GetProviderName(),
	// 	&web_api.Credential{
	// 		Id:    1,
	// 		Value: irRequest.GetCredential(),
	// 	})

	// if err != nil {
	// 	wVault.logger.Errorf("verification of the credentials failed with err %v", err)
	// 	return utils.ErrorWithCode[web_api.CreateProviderCredentialResponse](200,
	// 		err,
	// 		"Unable to verify the credentials, please check the credential and try again.")
	// }

	// if !verified.GetSuccess() {
	// 	wVault.logger.Errorf("verification for the key is not valid with error %+v", verified)
	// 	return utils.ErrorWithCode[web_api.CreateProviderCredentialResponse](200,
	// 		errors.New("unable to verify credentials"),
	// 		"Unable to verify the credentials, please check the credential and try again.")
	// }
	//  @todo later will make verified and not verified credentials
	vlt, err := wVault.vaultService.CreateOrganizationProviderCredential(ctx, iAuth, irRequest.GetProviderId(), irRequest.GetName(), irRequest.GetCredential().AsMap())
	if err != nil {
		wVault.logger.Errorf("vaultService.Create from grpc with err %v", err)
		return utils.Error[web_api.GetCredentialResponse](
			err,
			"Unable to create provider credential, please try again")
	}

	out := &web_api.VaultCredential{}
	err = utils.Cast(vlt, out)
	if err != nil {
		wVault.logger.Errorf("unable to cast the provider credentials to proto %v", err)
	}
	return utils.Success[web_api.GetCredentialResponse](out)
}

func (wVault *webVaultGRPCApi) DeleteCredential(c context.Context, irRequest *web_api.DeleteCredentialRequest) (*web_api.GetCredentialResponse, error) {
	iAuth, isAuthenticated := types.GetAuthPrincipleGPRC(c)
	if !isAuthenticated {
		wVault.logger.Errorf("DeleteProviderCredential from grpc with unauthenticated request")
		return utils.AuthenticateError[web_api.GetCredentialResponse]()
	}

	vlt, err := wVault.vaultService.Delete(c, iAuth, irRequest.GetVaultId())
	if err != nil {
		wVault.logger.Errorf("vaultService.Delete from grpc with err %v", err)
		return &web_api.GetCredentialResponse{
			Code:    400,
			Success: false,
			Error: &web_api.Error{
				ErrorCode:    400,
				ErrorMessage: err.Error(),
				HumanMessage: "Unable to delete provider credential, please try again",
			}}, nil
	}
	out := &web_api.VaultCredential{}
	err = utils.Cast(vlt, out)
	if err != nil {
		wVault.logger.Errorf("unable to cast the provider credentials to proto %v", err)
	}
	return utils.Success[web_api.GetCredentialResponse](out)
}

func (wVault *webVaultGRPCApi) GetAllOrganizationCredential(c context.Context, irRequest *web_api.GetAllOrganizationCredentialRequest) (*web_api.GetAllOrganizationCredentialResponse, error) {
	iAuth, isAuthenticated := types.GetSimplePrincipleGRPC(c)
	if !isAuthenticated {
		wVault.logger.Errorf("GetAllOrganizationCredential from grpc with unauthenticated request")
		return utils.AuthenticateError[web_api.GetAllOrganizationCredentialResponse]()
	}
	cnt, vlts, err := wVault.vaultService.GetAllOrganizationCredential(c, iAuth, irRequest.GetCriterias(), irRequest.GetPaginate())
	if err != nil {
		wVault.logger.Errorf("vaultService.GetAll from grpc with err %v", err)
		return utils.Error[web_api.GetAllOrganizationCredentialResponse](
			err,
			"Unable to get provider credentials, please try again",
		)
	}

	out := make([]*web_api.VaultCredential, len(*vlts))
	err = utils.Cast(vlts, &out)
	if err != nil {
		wVault.logger.Errorf("unable to cast vault object to proto %v", err)
	}

	for _, c := range out {
		c.Value = nil
	}
	return utils.PaginatedSuccess[web_api.GetAllOrganizationCredentialResponse, []*web_api.VaultCredential](
		uint32(cnt),
		irRequest.GetPaginate().GetPage(),
		out)

}

/*
this is not good idea as these apis are opened to public
*/
func (wVault *webVaultGRPCApi) GetProviderCredential(ctx context.Context, request *web_api.GetProviderCredentialRequest) (*web_api.GetCredentialResponse, error) {
	wVault.logger.Debugf("GetProviderCredential from grpc with requestPayload %v, %v", request, ctx)
	iAuth, isAuthenticated := types.GetSimplePrincipleGRPC(ctx)
	if !isAuthenticated {
		wVault.logger.Errorf("GetAllProviderCredential from grpc with unauthenticated request")
		return utils.AuthenticateError[web_api.GetCredentialResponse]()
	}
	vlt, err := wVault.vaultService.GetProviderCredential(ctx, iAuth, request.GetProviderId())
	if err != nil {
		return utils.Error[web_api.GetCredentialResponse](
			err,
			"Unable to get provider credential, please try again",
		)
	}
	wVault.logger.Debugf("returing few things like %+v", vlt)
	var out web_api.VaultCredential
	err = utils.Cast(vlt, &out)
	if err != nil {
		wVault.logger.Errorf("unable to cast vault object to proto %v", err)
	}
	return utils.Success[web_api.GetCredentialResponse, *web_api.VaultCredential](&out)
}

func (wVault *webVaultGRPCApi) CreateToolCredential(
	ctx context.Context,
	irRequest *web_api.CreateToolCredentialRequest) (*web_api.GetCredentialResponse, error) {
	wVault.logger.Debugf("CreateToolCredentialRequest from grpc with requestPayload %v, %v", irRequest, ctx)
	iAuth, isAuthenticated := types.GetAuthPrincipleGPRC(ctx)
	if !isAuthenticated {
		wVault.logger.Errorf("CreateToolCredentialRequest from grpc with unauthenticated request")
		return utils.AuthenticateError[web_api.GetCredentialResponse]()
	}

	vlt, err := wVault.vaultService.CreateOrganizationToolCredential(ctx,
		iAuth,
		irRequest.GetToolId(),
		irRequest.GetName(), irRequest.GetCredential().AsMap())
	if err != nil {
		wVault.logger.Errorf("vaultService.Create from grpc with err %v", err)
		return utils.Error[web_api.GetCredentialResponse](
			err,
			"Unable to create tool credential, please try again")
	}

	out := &web_api.VaultCredential{}
	err = utils.Cast(vlt, out)
	if err != nil {
		wVault.logger.Errorf("unable to cast the provider credentials to proto %v", err)
	}
	return utils.Success[web_api.GetCredentialResponse](out)
}

func (wVault *webVaultGRPCApi) GetOauth2Credential(ctx context.Context, request *web_api.GetCredentialRequest) (*web_api.GetCredentialResponse, error) {
	wVault.logger.Debugf("GetOauth2VaultCredential from grpc with requestPayload %v, %v", request, ctx)
	iAuth, isAuthenticated := types.GetSimplePrincipleGRPC(ctx)
	if !isAuthenticated {
		wVault.logger.Errorf("GetAllProviderCredential from grpc with unauthenticated request")
		return utils.AuthenticateError[web_api.GetCredentialResponse]()
	}
	vlt, err := wVault.vaultService.Get(
		ctx, iAuth, request.GetVaultId())

	if err != nil {
		wVault.logger.Errorf("unable to get tool credentials %v", err)
		return utils.Error[web_api.GetCredentialResponse](err, "Unable to get tool credential to get list of files.")
	}
	token, _, err := wVault.hubspotConnect.ToToken(vlt.Value)
	if err != nil {
		wVault.logger.Errorf("unable to get tool credentials %v", err)
		return utils.Error[web_api.GetCredentialResponse](err, "Unable to get tool credential to get list of files.")
	}
	newToken, err := wVault.hubspotConnect.RefreshToken(ctx, token)

	vlt.Value = newToken.Map()
	//
	var out web_api.VaultCredential
	err = utils.Cast(vlt, &out)
	if err != nil {
		wVault.logger.Errorf("unable to cast vault object to proto %v", err)
	}
	return utils.Success[web_api.GetCredentialResponse, *web_api.VaultCredential](&out)
}

func (wVault *webVaultGRPCApi) GetCredential(ctx context.Context, request *web_api.GetCredentialRequest) (*web_api.GetCredentialResponse, error) {
	wVault.logger.Debugf("GetProviderCredential from grpc with requestPayload %v, %v", request, ctx)
	iAuth, isAuthenticated := types.GetSimplePrincipleGRPC(ctx)
	if !isAuthenticated {
		wVault.logger.Errorf("GetCredentialRequest from grpc with unauthenticated request")
		return utils.AuthenticateError[web_api.GetCredentialResponse]()
	}
	//
	vlt, err := wVault.vaultService.Get(ctx, iAuth, request.GetVaultId())
	if err != nil {
		return utils.Error[web_api.GetCredentialResponse](
			err,
			"Unable to get vault credential, please try again",
		)
	}
	wVault.logger.Debugf("returing few things like %+v", vlt)
	var out web_api.VaultCredential
	err = utils.Cast(vlt, &out)
	if err != nil {
		wVault.logger.Errorf("unable to cast vault object to proto %v", err)
	}
	return utils.Success[web_api.GetCredentialResponse, *web_api.VaultCredential](&out)
}
