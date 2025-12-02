package domain

import (
	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/game"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing"
	"github.com/leet-gaming/match-making-api/pkg/domain/schedules"
)

// Inject initializes and sets up the domain components of the application.
// It sequentially injects dependencies for lobbies, pairing, schedules, and games.
//
// Parameters:
//   - c: A container.Container that can be used to cancel the operation or pass deadlines.
//
// Returns:
//   - error: An error if any of the injection processes fail, nil otherwise.
func Inject(c container.Container) error {
	return common.InjectAll(c, game.Inject, lobbies.Inject, pairing.Inject, schedules.Inject)
}
