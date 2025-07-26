package iam

import (
	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/iam/usecases"
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
	return common.InjectAll(c, usecases.InjectVerifyRID)
}
