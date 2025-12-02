package infra

import (
	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/infra/billing"
	"github.com/leet-gaming/match-making-api/pkg/infra/db/mongodb"
	"github.com/leet-gaming/match-making-api/pkg/infra/iam"
	"github.com/leet-gaming/match-making-api/pkg/infra/ioc"
	"github.com/leet-gaming/match-making-api/pkg/infra/squad"
)

// Inject sets up dependency injection for the application.
// It initializes and registers various components in the provided container.
//
// Parameters:
//   - c: A container.Container instance used for dependency injection.
//
// Returns:
//   - error: An error if the injection process fails, nil otherwise.
func Inject(c container.Container) error {
	return common.InjectAll(c, ioc.InjectIoc, mongodb.InjectGameRepository, mongodb.InjectGameModeRepository, mongodb.InjectRegionRepository, squad.Inject, billing.Inject, iam.Inject)
}
