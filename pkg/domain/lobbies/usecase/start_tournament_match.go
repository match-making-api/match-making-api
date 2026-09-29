package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
)

var (
	// ErrNotATournament is returned when the lobby is not a tournament.
	ErrNotATournament = errors.New("not a tournament lobby")
	// ErrTournamentClosed is returned when the lobby is not open.
	ErrTournamentClosed = errors.New("tournament closed")
	// ErrTournamentForbidden is returned when the caller tenant/client does not own the lobby.
	ErrTournamentForbidden = errors.New("tournament forbidden")
	// ErrNotEnoughPlayers is returned when a tournament match cannot be paired.
	ErrNotEnoughPlayers = errors.New("not enough players")
	// ErrMatchAlreadyStarted is returned when the lobby match is already in progress.
	ErrMatchAlreadyStarted = errors.New("match already started")
)

// TournamentLobbyStore loads and saves lobbies.
type TournamentLobbyStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Lobby, error)
	Update(ctx context.Context, lobby *entities.Lobby) error
}

// PrizePoolLockedNotice is published when a tournament match locks the pool.
type PrizePoolLockedNotice struct {
	PoolID          uuid.UUID
	LobbyID         uuid.UUID
	TotalAmount     int64
	Currency        string
	TenantID        string
	ClientID        string
	ResourceOwnerID string
}

// PrizePoolPublisher emits prize-pool lifecycle events. It does not transfer funds.
type PrizePoolPublisher interface {
	PublishPrizePoolLocked(ctx context.Context, notice PrizePoolLockedNotice) error
}

// MatchFormationHandoff asks match formation (2503) to allocate a server.
// MatchStarted stays an inbound event from the game server (ADR-001).
type MatchFormationHandoff interface {
	EnqueueAndBroadcast(ctx context.Context, matchID, gameID uuid.UUID, region string, playerIDs []uuid.UUID, tenantID, clientID, resourceOwnerID string) error
}

// StartTournamentMatchInput identifies the lobby and optional match id.
type StartTournamentMatchInput struct {
	LobbyID uuid.UUID
	MatchID uuid.UUID
	Caller  common.ResourceOwner
}

// StartTournamentMatchUseCase builds a single-elimination bracket, locks the prize pool, and hands the match to server allocation.
type StartTournamentMatchUseCase struct {
	store   TournamentLobbyStore
	prizes  PrizePoolPublisher
	handoff MatchFormationHandoff
}

// NewStartTournamentMatchUseCase builds the start-match use case. Publisher and handoff may be nil.
func NewStartTournamentMatchUseCase(store TournamentLobbyStore, prizes PrizePoolPublisher, handoff MatchFormationHandoff) *StartTournamentMatchUseCase {
	return &StartTournamentMatchUseCase{store: store, prizes: prizes, handoff: handoff}
}

// Execute updates bracket state and locks the prize pool when a tournament match starts.
func (u *StartTournamentMatchUseCase) Execute(ctx context.Context, in StartTournamentMatchInput) (*entities.Lobby, error) {
	if u == nil || u.store == nil {
		return nil, errors.New("start tournament store is nil")
	}
	if in.Caller.TenantID == uuid.Nil || in.Caller.ClientID == uuid.Nil {
		return nil, common.ErrMissingResourceOwnership
	}
	lobby, err := u.store.GetByID(ctx, in.LobbyID)
	if err != nil {
		return nil, err
	}
	if lobby == nil {
		return nil, errors.New("lobby not found")
	}
	if err := lobby.ValidateOwnership(); err != nil {
		return nil, err
	}
	if lobby.ResourceOwner.TenantID != in.Caller.TenantID || lobby.ResourceOwner.ClientID != in.Caller.ClientID {
		return nil, ErrTournamentForbidden
	}
	if lobby.Type != entities.LobbyTypeTournament {
		return nil, ErrNotATournament
	}
	if lobby.Status == entities.LobbyStatusStarting || lobby.Status == entities.LobbyStatusStarted {
		return lobby, ErrMatchAlreadyStarted
	}
	if lobby.Status != entities.LobbyStatusOpen {
		return nil, ErrTournamentClosed
	}

	players := seatedPlayers(lobby)
	if len(players) < 2 {
		return nil, ErrNotEnoughPlayers
	}
	if lobby.Bracket == nil {
		lobby.Bracket = entities.BuildSingleElimination(players)
	}
	matchID := in.MatchID
	if matchID == uuid.Nil {
		matchID = uuid.New()
	}
	started := false
	for i := range lobby.Bracket.Matches {
		m := &lobby.Bracket.Matches[i]
		if m.Status == entities.BracketMatchPending && len(m.PlayerIDs) == 2 {
			m.Status = entities.BracketMatchStarted
			m.MatchID = matchID.String()
			started = true
			break
		}
	}
	if !started {
		return nil, ErrMatchAlreadyStarted
	}

	newlyLocked := false
	if lobby.PrizePool != nil && lobby.PrizePool.Status != entities.PrizePoolStatusLocked {
		lobby.PrizePool.Status = entities.PrizePoolStatusLocked
		newlyLocked = true
	}
	lobby.Status = entities.LobbyStatusStarting
	lobby.UpdatedAt = time.Now().UTC()
	if err := u.store.Update(ctx, lobby); err != nil {
		return nil, err
	}

	if newlyLocked && u.prizes != nil {
		poolID := lobby.ID
		if lobby.PrizePool.PrizePoolID != "" {
			if parsed, err := uuid.Parse(lobby.PrizePool.PrizePoolID); err == nil {
				poolID = parsed
			}
		}
		amount := int64(0)
		if lobby.PrizePool.EntryFeeCents > 0 {
			amount = int64(lobby.PrizePool.EntryFeeCents) * int64(len(players))
		}
		if err := u.prizes.PublishPrizePoolLocked(ctx, PrizePoolLockedNotice{
			PoolID:          poolID,
			LobbyID:         lobby.ID,
			TotalAmount:     amount,
			Currency:        "USD",
			TenantID:        lobby.ResourceOwner.TenantID.String(),
			ClientID:        lobby.ResourceOwner.ClientID.String(),
			ResourceOwnerID: lobby.ResourceOwner.UserID.String(),
		}); err != nil {
			return lobby, err
		}
	}

	if u.handoff != nil {
		if gameID, err := uuid.Parse(lobby.GameID); err == nil {
			_ = u.handoff.EnqueueAndBroadcast(ctx, matchID, gameID, lobby.Region, players, lobby.ResourceOwner.TenantID.String(), lobby.ResourceOwner.ClientID.String(), lobby.ResourceOwner.UserID.String())
		}
	}
	return lobby, nil
}

func seatedPlayers(lobby *entities.Lobby) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(lobby.PlayerSlots))
	for _, slot := range lobby.PlayerSlots {
		if slot.PlayerID != nil && !slot.IsSpectator {
			out = append(out, *slot.PlayerID)
		}
	}
	return out
}
