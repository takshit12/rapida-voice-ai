package web_api

import (
	"context"
	"errors"
	"io"

	assistant_client "github.com/lexatic/web-backend/pkg/clients/workflow"
	"github.com/lexatic/web-backend/pkg/utils"
	web_api "github.com/lexatic/web-backend/protos/lexatic-backend"

	config "github.com/lexatic/web-backend/config"
	commons "github.com/lexatic/web-backend/pkg/commons"
	"github.com/lexatic/web-backend/pkg/connectors"
	"github.com/lexatic/web-backend/pkg/types"
)

type webTalkApi struct {
	WebApi
	cfg                          *config.AppConfig
	logger                       commons.Logger
	postgres                     connectors.PostgresConnector
	redis                        connectors.RedisConnector
	assistantConversactionClient assistant_client.AssistantConversactionServiceClient
}

type webTalkGRPCApi struct {
	webTalkApi
}

func NewTalkGRPC(config *config.AppConfig, logger commons.Logger, postgres connectors.PostgresConnector, redis connectors.RedisConnector) web_api.TalkServiceServer {
	return &webTalkGRPCApi{
		webTalkApi{
			WebApi:                       NewWebApi(config, logger, postgres, redis),
			cfg:                          config,
			logger:                       logger,
			postgres:                     postgres,
			redis:                        redis,
			assistantConversactionClient: assistant_client.NewAssistantConversactionServiceClientGRPC(config, logger, redis),
		},
	}
}

//
//

// GetAllConversactionMessage implements lexatic_backend.AssistantConversactionServiceServer.
func (assistant *webTalkGRPCApi) GetAllConversactionMessage(ctx context.Context, iRequest *web_api.GetAllConversactionMessageRequest) (*web_api.GetAllConversactionMessageResponse, error) {
	assistant.logger.Debugf("GetAllConversactionMessage started")
	iAuth, isAuthenticated := types.GetAuthPrincipleGPRC(ctx)
	if !isAuthenticated {
		assistant.logger.Errorf("unauthenticated request for get actvities")
		return nil, errors.New("unauthenticated request")
	}

	_page, _assistant, err := assistant.assistantConversactionClient.GetAllConversactionMessage(ctx, iAuth, iRequest.GetAssistantId(), iRequest.GetAssistantConversactionId(), iRequest.GetCriterias(), iRequest.GetPaginate())
	if err != nil {
		return utils.Error[web_api.GetAllConversactionMessageResponse](
			err,
			"Unable to get your assistant, please try again in sometime.")
	}

	return utils.PaginatedSuccess[web_api.GetAllConversactionMessageResponse, []*web_api.AssistantConversactionMessage](
		_page.GetTotalItem(), _page.GetCurrentPage(),
		_assistant)

}

func (assistant *webTalkGRPCApi) CreateAssistantMessage(cer *web_api.CreateAssistantMessageRequest, stream web_api.TalkService_CreateAssistantMessageServer) error {
	assistant.logger.Debugf("CreateAssistantMessage started")
	c := stream.Context()
	iAuth, isAuthenticated := types.GetAuthPrincipleGPRC(c)
	if !isAuthenticated {
		assistant.logger.Errorf("unauthenticated request for get actvities")
		return errors.New("unauthenticated request")
	}
	out, err := assistant.assistantConversactionClient.CreateAssistantMessage(c, iAuth, cer)
	if err != nil {
		return err
	}

	// Channel to handle errors from the upstream stream
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		for {
			st, recvErr := out.Recv()
			if recvErr == io.EOF {
				return // End of upstream stream
			}
			if recvErr != nil {
				errCh <- recvErr
				return
			}
			// Forward message to downstream stream
			if err := stream.Send(st); err != nil {
				errCh <- err
				return
			}
		}
	}()

	// Wait for any errors from the upstream stream or the context cancellation
	select {
	case err := <-errCh:
		return err
	case <-c.Done():
		return c.Err()
	}

}

// GetAllAssistantConversaction implements lexatic_backend.AssistantConversactionServiceServer.
func (assistant *webTalkGRPCApi) GetAllAssistantConversaction(c context.Context, iRequest *web_api.GetAllAssistantConversactionRequest) (*web_api.GetAllAssistantConversactionResponse, error) {
	assistant.logger.Debugf("CreateAssistantMessage started")
	iAuth, isAuthenticated := types.GetAuthPrincipleGPRC(c)
	if !isAuthenticated {
		assistant.logger.Errorf("unauthenticated request for get actvities")
		return nil, errors.New("unauthenticated request")
	}

	_page, _assistantConvo, err := assistant.assistantConversactionClient.GetAllAssistantConversaction(c, iAuth,
		iRequest.GetAssistantId(),
		iRequest.GetCriterias(), iRequest.GetPaginate())
	if err != nil {
		return utils.Error[web_api.GetAllAssistantConversactionResponse](
			err,
			"Unable to get your assistant, please try again in sometime.")
	}

	return utils.PaginatedSuccess[web_api.GetAllAssistantConversactionResponse, []*web_api.AssistantConversaction](
		_page.GetTotalItem(), _page.GetCurrentPage(),
		_assistantConvo)
}
