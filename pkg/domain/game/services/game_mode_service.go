package services

import (
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/db/mongodb"
)

type GameModeService interface {
	common.QueryService[entities.GameMode]
}

// GameModeService is a service for managing game modes.
type gameModeService struct {
	common.BaseQueryService[entities.GameMode, mongodb.GameModeRepository]
}

// NewGameModeService creates a new GameModeService with the provided reader.
func NewGameModeService(repository mongodb.GameModeRepository) GameModeService {
	queryableFields := common.GetQueryableFields(map[string]bool{
		"Name":        common.ALLOW,
		"Description": common.ALLOW,
	})

	readableFields := common.GetReadableFields()

	return &gameModeService{
		BaseQueryService: common.BaseQueryService[entities.GameMode, mongodb.GameModeRepository]{
			Repository:      repository,
			QueryableFields: queryableFields,
			ReadableFields:  readableFields,
			MaxPageSize:     100,
			Audience:        common.TenantAudienceIDKey,
		},
	}
}
