package internal_service

import (
	"context"

	internal_entity "github.com/lexatic/web-backend/api/web-api/internal/entity"
	"github.com/lexatic/web-backend/pkg/types"
	web_api "github.com/lexatic/web-backend/protos/lexatic-backend"
)

type VaultService interface {
	CreateOrganizationToolCredential(ctx context.Context,
		auth types.Principle,
		toolId uint64,
		name string, credential map[string]interface{}) (*internal_entity.Vault, error)

	//
	CreateOrganizationProviderCredential(ctx context.Context,
		auth types.SimplePrinciple,
		providerId uint64,
		name string, credential map[string]interface{}) (*internal_entity.Vault, error)

	//
	CreateUserToolCredential(ctx context.Context,
		auth types.Principle,
		toolId uint64,
		name string,
		credential map[string]interface{},
	) (*internal_entity.Vault, error)
	//
	Get(ctx context.Context, auth types.SimplePrinciple, vltId uint64) (*internal_entity.Vault, error)
	GetProviderCredential(ctx context.Context,
		auth types.SimplePrinciple, providerId uint64) (*internal_entity.Vault, error)

	Delete(ctx context.Context, auth types.Principle, vaultId uint64) (*internal_entity.Vault, error)
	GetAllOrganizationCredential(ctx context.Context, auth types.SimplePrinciple, criterias []*web_api.Criteria, paginate *web_api.Paginate) (int64, *[]internal_entity.Vault, error)
	CreateRapidaProviderCredential(ctx context.Context, organizationId uint64) (*internal_entity.Vault, error)
}
