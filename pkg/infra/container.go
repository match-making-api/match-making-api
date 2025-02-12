package infra

import (
	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/infra/ioc/db"
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
	err := db.InjectMongoDB(c)

	return err

}
