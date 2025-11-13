package internal_vault_service

import (
	"context"
	"fmt"

	internal_entity "github.com/rapidaai/api/web-api/internal/entity"
	internal_services "github.com/rapidaai/api/web-api/internal/service"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/connectors"
	gorm_models "github.com/rapidaai/pkg/models/gorm"
	gorm_types "github.com/rapidaai/pkg/models/gorm/types"
	"github.com/rapidaai/pkg/types"
	type_enums "github.com/rapidaai/pkg/types/enums"
	web_api "github.com/rapidaai/protos"
	"gorm.io/gorm/clause"
)

type vaultService struct {
	logger   commons.Logger
	postgres connectors.PostgresConnector
}

func NewVaultService(logger commons.Logger, postgres connectors.PostgresConnector) internal_services.VaultService {
	return &vaultService{
		logger:   logger,
		postgres: postgres,
	}
}

func (vS *vaultService) CreateOrganizationToolCredential(ctx context.Context,
	auth types.Principle,
	toolId uint64,
	name string, credential map[string]interface{}) (*internal_entity.Vault, error) {
	return vS.createCredential(ctx, *auth.GetUserId(), name, credential, gorm_types.VAULT_TYPE_TOOL,
		toolId, gorm_types.VAULT_LEVEL_ORGANIZATION, *auth.GetCurrentOrganizationId())
}

func (vS *vaultService) CreateOrganizationProviderCredential(ctx context.Context,
	auth types.SimplePrinciple,
	providerId uint64,
	name string, credential map[string]interface{}) (*internal_entity.Vault, error) {
	return vS.createCredential(ctx, *auth.GetUserId(), name, credential, gorm_types.VAULT_TYPE_PROVIDER,
		providerId, gorm_types.VAULT_LEVEL_ORGANIZATION, *auth.GetCurrentOrganizationId())

}

func (vS *vaultService) CreateUserToolCredential(ctx context.Context,
	auth types.Principle,
	toolId uint64,
	name string,
	credential map[string]interface{},
) (*internal_entity.Vault, error) {
	return vS.createCredential(ctx, *auth.GetUserId(), name,
		credential,
		gorm_types.VAULT_TYPE_TOOL,
		toolId,
		gorm_types.VAULT_LEVEL_USER,
		*auth.GetUserId())
}

func (vs *vaultService) createCredential(ctx context.Context,
	userId uint64,
	name string,
	credential map[string]interface{},
	vaultType gorm_types.VaultType,
	vaultTypeId uint64, level gorm_types.VaultLevel, levelId uint64) (*internal_entity.Vault, error) {
	db := vs.postgres.DB(ctx)
	vlt := &internal_entity.Vault{
		Mutable: gorm_models.Mutable{
			CreatedBy: userId,
		},
		Name:         name,
		VaultType:    vaultType,
		VaultTypeId:  vaultTypeId,
		Value:        credential,
		VaultLevel:   level,
		VaultLevelId: levelId,
	}

	tx := db.Save(vlt)
	if err := tx.Error; err != nil {
		vs.logger.Debugf("unable to create organization credentials for tool %v", err)
		return nil, err
	}
	return vlt, nil
}

func (vS *vaultService) Delete(ctx context.Context, auth types.Principle, vaultId uint64) (*internal_entity.Vault, error) {
	db := vS.postgres.DB(ctx)
	vlt := &internal_entity.Vault{
		Mutable: gorm_models.Mutable{
			Status:    type_enums.RECORD_ARCHIEVE,
			UpdatedBy: *auth.GetUserId(),
		},
	}
	tx := db.Where("id = ? and vault_level = ? AND vault_level_id = ?", vaultId, gorm_types.VAULT_LEVEL_ORGANIZATION, *auth.GetCurrentOrganizationId()).Clauses(clause.Returning{}).Updates(vlt)
	if err := tx.Error; err != nil {
		vS.logger.Debugf("unable to delete vault %v")
		return nil, err
	}
	return vlt, nil
}

func (vS *vaultService) GetAllOrganizationCredential(ctx context.Context, auth types.SimplePrinciple, criterias []*web_api.Criteria, paginate *web_api.Paginate) (int64, []*internal_entity.Vault, error) {
	db := vS.postgres.DB(ctx)
	var vaults []*internal_entity.Vault
	var cnt int64

	qry := db.Debug().Model(internal_entity.Vault{})
	qry.
		Where("vault_level = ? AND vault_level_id = ? AND status = ?", gorm_types.VAULT_LEVEL_ORGANIZATION, *auth.GetCurrentOrganizationId(), type_enums.RECORD_ACTIVE)
	for _, ct := range criterias {
		switch ct.GetLogic() {
		case "or":
			qry.Or(fmt.Sprintf("%s = ?", ct.GetKey()), ct.GetValue())
		case "like":
			qry.Where(fmt.Sprintf("%s %s ?", ct.GetKey(), ct.GetLogic()), fmt.Sprintf("%%%s%%", ct.GetValue()))
		default:
			qry.Where(fmt.Sprintf("%s %s ?", ct.GetKey(), ct.GetLogic()), ct.GetValue())
		}
	}
	tx := qry.
		Scopes(gorm_models.
			Paginate(gorm_models.
				NewPaginated(
					int(paginate.GetPage()),
					int(paginate.GetPageSize()),
					&cnt,
					qry))).
		Order(clause.OrderByColumn{
			Column: clause.Column{Name: "created_date"},
			Desc:   true,
		}).Find(&vaults)

	if tx.Error != nil {
		vS.logger.Debugf("unable to find any vault %v", *auth.GetCurrentOrganizationId())
		return cnt, nil, tx.Error
	}

	return cnt, vaults, nil
}

func (vS *vaultService) Get(ctx context.Context, auth types.SimplePrinciple, id uint64) (*internal_entity.Vault, error) {
	db := vS.postgres.DB(ctx)
	var vault internal_entity.Vault
	tx := db.Where("id = ? AND status = ? AND vault_level = ? AND vault_level_id = ?",
		id,
		type_enums.RECORD_ACTIVE.String(),
		string(gorm_types.VAULT_LEVEL_ORGANIZATION),
		*auth.GetCurrentOrganizationId(),
	).Last(&vault)
	if tx.Error != nil {
		vS.logger.Errorf("get credential error  %v", tx.Error)
		return nil, tx.Error
	}
	return &vault, nil
}

// gorm_types.VAULT_TYPE_PROVIDER,
// rapidaProviderId,
func (vS *vaultService) GetProviderCredential(ctx context.Context,
	auth types.SimplePrinciple,
	providerId uint64) (*internal_entity.Vault, error) {
	db := vS.postgres.DB(ctx)
	var vault internal_entity.Vault

	tx := db.Where("status = ? and vault_type = ? and vault_type_id = ? and vault_level = ? and vault_level_id = ?",
		type_enums.RECORD_ACTIVE.String(),
		string(gorm_types.VAULT_TYPE_PROVIDER),
		providerId,
		string(gorm_types.VAULT_LEVEL_ORGANIZATION),
		*auth.GetCurrentOrganizationId(),
	).Last(&vault)
	if tx.Error != nil {
		vS.logger.Errorf("get credential error  %v", tx.Error)
		return nil, tx.Error
	}
	return &vault, nil
}

func (vS *vaultService) GetUserToolCredential(ctx context.Context,
	auth types.SimplePrinciple,
	toolId uint64) (*internal_entity.Vault, error) {
	db := vS.postgres.DB(ctx)
	var vault internal_entity.Vault

	tx := db.Where("status = ? and vault_type = ? and vault_type_id = ? and vault_level = ? and vault_level_id = ?",
		type_enums.RECORD_ACTIVE.String(),
		string(gorm_types.VAULT_TYPE_TOOL),
		toolId,
		string(gorm_types.VAULT_LEVEL_USER),
		*auth.GetUserId(),
	).Last(&vault)
	if tx.Error != nil {
		vS.logger.Errorf("get credential error  %v", tx.Error)
		return nil, tx.Error
	}
	return &vault, nil
}

func (vS *vaultService) CreateRapidaProviderCredential(ctx context.Context,
	organizationId uint64) (*internal_entity.Vault, error) {

	rapidaProviderId := uint64(8298870085084815298)
	return vS.createCredential(
		ctx,
		99,
		"default-rapida-credentials",
		// this is mistral key
		// mistral
		// EOruCAkjnW8O6M3a0VSKXPKUzLraQ5Gv
		// stability
		// sk-1aVQO9ElyaXhxuROVMFeRqQgCPBL5WJQyMNL7wzxkR27kFSw
		// replicate
		// r8_FvPKVcfvtL3NifEKEUvi7q2uEpNYUsm3MUpMN
		map[string]interface{}{
			"1987967168347635712": map[string]string{
				"key": "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",
			},
			"5212367370329048775": map[string]string{
				"key": "pa-f2E7NuNmZrC9ADFKE6KjnyZgxUH8_xliK3C0CAKVG00",
			},
			"198796716894742120": map[string]string{
				"key": "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",
			},
			"198796716894742119": map[string]string{
				"key": "pskpIUoGzlLvzuxczSpLaJpxuFhZiWzy",
			},
			"198796716894742118": map[string]string{
				"key": "AIzaSyBI19ykmjj-wsA_hR3rt-UrPkFtwQQ3hhY",
			},
			"1987967168452493312": map[string]string{
				"key": "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",
			},
			"1987967168435716096": map[string]string{
				"key": "nHuteTe84dihnImlgpwiD7Tk9cmAHP1qxHocstf5",
			},
			"1987967168431521792": map[string]string{
				"key": "r8_FvPKVcfvtL3NifEKEUvi7q2uEpNYUsm3MUpMN",
			},
		},
		gorm_types.VAULT_TYPE_PROVIDER,
		rapidaProviderId,
		gorm_types.VAULT_LEVEL_ORGANIZATION,
		organizationId,
	)
}
