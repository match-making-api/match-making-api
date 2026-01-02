package services

import (
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/db/mongodb"
)

type RegionService interface {
	common.QueryService[entities.Region]
}

type regionService struct {
	common.BaseQueryService[entities.Region, mongodb.RegionRepository]
}

func NewRegionService(repository mongodb.RegionRepository) RegionService {
	queryableFields := common.GetQueryableFields(map[string]bool{
		"Name":        common.ALLOW,
		"Description": common.ALLOW,
		"Slug":        common.DENY,
	})

	readableFields := common.GetReadableFields()

	return &regionService{
		BaseQueryService: common.BaseQueryService[entities.Region, mongodb.RegionRepository]{
			Repository:      repository,
			QueryableFields: queryableFields,
			ReadableFields:  readableFields,
			MaxPageSize:     100,
			Audience:        common.TenantAudienceIDKey,
		},
	}
}
