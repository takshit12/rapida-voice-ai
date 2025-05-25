package internal_provider_service

import (
	"context"
	"fmt"

	internal_entity "github.com/lexatic/web-backend/api/web-api/internal/entity"
	internal_service "github.com/lexatic/web-backend/api/web-api/internal/service"
	"github.com/lexatic/web-backend/pkg/commons"
	"github.com/lexatic/web-backend/pkg/connectors"
	type_enums "github.com/lexatic/web-backend/pkg/types/enums"
	web_api "github.com/lexatic/web-backend/protos/lexatic-backend"
)

type providerService struct {
	logger   commons.Logger
	postgres connectors.PostgresConnector
}

func (ps *providerService) GetAllModelProvider(c context.Context, criterias []*web_api.Criteria) ([]*internal_entity.Provider, error) {
	var providers []*internal_entity.Provider
	db := ps.postgres.DB(c)
	tx := db.Where("status = ?", type_enums.RECORD_ACTIVE).Find(&providers)
	if tx.Error != nil {
		ps.logger.Errorf("unable to get all the providers with error %v", tx.Error)
		return nil, tx.Error
	}
	return providers, nil
}

func (ps *providerService) GetModel(c context.Context, modelId uint64) (*internal_entity.ProviderModel, error) {
	db := ps.postgres.DB(c)
	var pv internal_entity.ProviderModel
	if err := db.
		Preload("Parameters").
		Preload("Parameters.Metadatas").
		Preload("Provider").
		Preload("Metadatas").
		Where("id = ?", modelId).Find(&pv).Error; err != nil {
		ps.logger.Errorf("unable to get all model with err %v", err)
		return nil, err
	}
	return &pv, nil
}

// getting all the models
func (ps *providerService) GetAllModel(c context.Context, criterias []*web_api.Criteria) ([]*internal_entity.ProviderModel, error) {
	db := ps.postgres.DB(c)
	var pv []*internal_entity.ProviderModel
	qry := db.Preload("Parameters").
		Preload("Parameters.Metadatas").
		Preload("Metadatas").
		Joins("inner join providers Provider on Provider.id = provider_models.provider_id").
		Preload("Provider")
	qry = qry.Joins("JOIN provider_model_endpoints ON provider_models.id = provider_model_endpoints.provider_model_id").
		Select("provider_models.*, provider_model_endpoints.endpoint")
	for _, ct := range criterias {
		qry.Where(fmt.Sprintf("%s = ?", ct.GetKey()), ct.GetValue())
	}

	if err := qry.
		Where("provider_models.status = ?", "ACTIVE").
		Order("provider_models.name desc").
		Find(&pv).
		Error; err != nil {
		ps.logger.Errorf("unable to get all the model")
		return nil, err
	}
	return pv, nil
}

// GetAllToolProvider implements internal_service.ProviderService.
func (ps *providerService) GetAllToolProvider(ctx context.Context, criterias []*web_api.Criteria) ([]*internal_entity.ToolProvider, error) {
	var providers []*internal_entity.ToolProvider
	db := ps.postgres.DB(ctx)
	tx := db.Find(&providers)
	if tx.Error != nil {
		ps.logger.Errorf("unable to get all the providers with error %v", tx.Error)
		return nil, tx.Error
	}
	return providers, nil
}

func NewProviderService(logger commons.Logger, postgres connectors.PostgresConnector) internal_service.ProviderService {
	return &providerService{
		logger:   logger,
		postgres: postgres,
	}
}
