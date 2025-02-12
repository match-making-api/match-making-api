package domain

import (
	"context"

	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing"
	"github.com/leet-gaming/match-making-api/pkg/domain/schedules"
)

// Inject initializes and sets up the domain components of the application.
// It sequentially injects dependencies for lobbies, pairing, and schedules.
//
// Parameters:
//   - ctx: A context.Context that can be used to cancel the operation or pass deadlines.
//
// Returns:
//   - error: An error if any of the injection processes fail, nil otherwise.
func Inject(ctx context.Context) error {
	err := lobbies.Inject(ctx)

	if err != nil {
		return err
	}

	err = pairing.Inject(ctx)

	if err != nil {
		return err
	}

	err = schedules.Inject(ctx)

	return err
}
