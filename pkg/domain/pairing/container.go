package pairing

import (
	"github.com/golobby/container/v3"
	pairing_in "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	schedules_in_ports "github.com/leet-gaming/match-making-api/pkg/domain/schedules/ports/in"
)

// Inject initializes and registers pairing-related dependencies in the provided container.
//
// Parameters:
//   - container: A container.Container object used as a dependency injection container.
//
// Returns:
//   An error if any initialization or registration fails, otherwise nil.
func Inject(c container.Container) error {
	// Register PartyScheduleMatcher use case
	if err := c.Singleton(func(scheduleReader schedules_in_ports.PartyScheduleReader) (pairing_in.PartyScheduleMatcher, error) {
		return usecases.NewPartyScheduleMatcher(scheduleReader), nil
	}); err != nil {
		return err
	}

	return nil
}
