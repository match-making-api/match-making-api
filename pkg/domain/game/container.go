package game

import (
	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/usecases"
)

// Inject initializes and sets up the game module within the given container.
//
// Parameters:
//   - c: A container.Container instance used for dependency injection.
//
// Returns:
//   - An error if the injection process encounters any issues, or nil if successful.
func Inject(c container.Container) error {
	return common.InjectAll(c,
		// Game usecases
		usecases.InjectCreateGame,
		usecases.InjectUpdateGame,
		usecases.InjectDeleteGame,
		usecases.InjectGetGameByID,
		usecases.InjectSearchGames,
		// GameMode usecases
		usecases.InjectCreateGameMode,
		usecases.InjectUpdateGameMode,
		usecases.InjectDeleteGameMode,
		usecases.InjectGetGameModeByID,
		usecases.InjectGetGameModes,
		usecases.InjectSearchGameModes,
		// Region usecases
		usecases.InjectCreateRegion,
		usecases.InjectUpdateRegion,
		usecases.InjectDeleteRegion,
		usecases.InjectGetRegionByID,
		usecases.InjectGetRegions,
		usecases.InjectSearchRegions,
	)
}
