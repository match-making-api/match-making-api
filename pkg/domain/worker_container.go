package domain

import (
	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/game"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing"
	"github.com/leet-gaming/match-making-api/pkg/domain/schedules"
)

// InjectWorker sets up only the domain dependencies needed by background workers.
// It skips IAM and lobbies which are only needed by the REST API.
//
// Includes: game (regions), pairing (consumer, ticker, pools), schedules (party matcher).
func InjectWorker(c container.Container) error {
	return common.InjectAll(c, game.Inject, pairing.Inject, schedules.Inject)
}
