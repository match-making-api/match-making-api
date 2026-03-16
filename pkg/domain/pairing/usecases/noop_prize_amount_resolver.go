package usecases

import (
	"context"

	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// NoOpPrizeAmountResolver is a PrizeAmountResolver that returns no prizes.
// Use when prize pool / distribution rules are not yet implemented or live in replay-api.
// Replace with a real implementation (e.g. LobbyPrizePoolResolver) when prize context is available.
type NoOpPrizeAmountResolver struct{}

// NewNoOpPrizeAmountResolver creates a resolver that always returns no prizes.
func NewNoOpPrizeAmountResolver() pairing_out.PrizeAmountResolver {
	return &NoOpPrizeAmountResolver{}
}

// Resolve always returns nil (no prizes). Match-making will not produce PrizeDistributed for any match.
func (*NoOpPrizeAmountResolver) Resolve(_ context.Context, _ pairing_out.PrizeResolveRequest) ([]pairing_out.WinnerPrize, error) {
	return nil, nil
}
