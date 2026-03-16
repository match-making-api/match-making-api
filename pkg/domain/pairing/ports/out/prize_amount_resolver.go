package pairing_out

import (
	"context"
)

// WinnerPrize represents a winner's prize amount for distribution.
type WinnerPrize struct {
	PlayerID string
	AmountCents int64
	Currency string // e.g. "USD"
}

// PrizeAmountResolver resolves prize amounts for a match. Used by prize distribution handler.
// Returns nil or empty slice when match has no prize pool; wallet API will not receive event.
// Implementations may query lobby prize pool, replay-api, or config.
type PrizeAmountResolver interface {
	// Resolve returns prize amounts per winner for the given match.
	// matchID, winnerPlayerIDs: from MatchResultsCalculated.
	// lobbyID, prizePoolID: optional, from MatchResultsCalculated when available.
	// Returns nil or empty = no prizes, skip distribution event.
	Resolve(ctx context.Context, req PrizeResolveRequest) ([]WinnerPrize, error)
}

// PrizeResolveRequest holds parameters for prize resolution.
type PrizeResolveRequest struct {
	MatchID          string
	WinnerPlayerIDs  []string
	LobbyID          string
	PrizePoolID      string
	TenantID         string
	ClientID         string
}
