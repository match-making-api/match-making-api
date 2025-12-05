package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/db/mongodb"
)

type GameService interface {
	common.QueryService[entities.Game]
	Create(ctx context.Context, game *entities.Game) (*entities.Game, error)
	Update(ctx context.Context, game *entities.Game) (*entities.Game, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// GameService is a service for managing game data
type gameService struct {
	common.BaseQueryService[entities.Game, mongodb.GameRepository]
}

// NewGameService creates a new GameService with the provided reader.
func NewGameService(repository mongodb.GameRepository) GameService {
	queryableFields := common.GetQueryableFields(map[string]bool{
		"Name":        common.ALLOW,
		"Description": common.ALLOW,
	})

	readableFields := common.GetReadableFields()

	return &gameService{
		BaseQueryService: common.BaseQueryService[entities.Game, mongodb.GameRepository]{
			Repository:      repository,
			QueryableFields: queryableFields,
			ReadableFields:  readableFields,
			MaxPageSize:     100,
			Audience:        common.TenantAudienceIDKey,
		},
	}
}

func (s *gameService) Create(ctx context.Context, game *entities.Game) (*entities.Game, error) {
	game.ID = uuid.New()

	createdGame, err := s.Repository.Create(ctx, game)

	if err != nil {
		return nil, err
	}

	return createdGame, nil
}

func (s *gameService) Update(ctx context.Context, game *entities.Game) (*entities.Game, error) {
	existingGame, err := s.Repository.GetByID(ctx, game.ID)

	if err != nil {
		return nil, err
	}

	existingGame.Name = game.Name
	existingGame.Description = game.Description

	updatedGame, err := s.Repository.Update(ctx, existingGame)

	if err != nil {
		return nil, err
	}

	return updatedGame, nil
}

func (s *gameService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.Repository.Delete(ctx, id)
}
