package entities

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// GameMode represents a game mode in the game ecosystem
type GameMode struct {
	common.BaseEntity
	GameID      uuid.UUID `json:"game_id" bson:"game_id"`         // ID of the game the game mode belongs to
	Name        string    `json:"name" bson:"name"`               // Name of the game mode
	Description string    `json:"description" bson:"description"` // Description of the game mode
}

func NewSearchGameModeByGameID(ctx context.Context, gameID uuid.UUID) common.Search {
	params := []common.SearchAggregation{
		{
			Params: []common.SearchParameter{
				{
					ValueParams: []common.SearchableValue{
						{
							Field: "GameID",
							Values: []any{
								gameID,
							},
						},
					},
				},
			},
		},
	}

	visibility := common.SearchVisibilityOptions{
		RequestSource:    common.GetResourceOwner(ctx),
		IntendedAudience: common.ClientApplicationAudienceIDKey,
	}

	result := common.SearchResultOptions{}

	return common.Search{
		SearchParams:      params,
		ResultOptions:     result,
		VisibilityOptions: visibility,
	}
}
