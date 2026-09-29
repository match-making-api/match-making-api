package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
)

var (
	// ErrNotATournament is returned when the lobby is not a tournament.
	ErrNotATournament = errors.New("not a tournament lobby")
	// ErrTournamentClosed is returned when the lobby is not open.
	ErrTournamentClosed = errors.New("tournament closed")
	// ErrTournamentForbidden is returned when the caller tenant/client does not own the lobby.
	ErrTournamentForbidden = errors.New("tournament forbidden")
)

// TournamentLobbyStore loads and saves lobbies.
type TournamentLobbyStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Lobby, error)
	Update(ctx context.Context, lobby *entities.Lobby) error
}
