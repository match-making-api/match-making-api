package usecases_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
)

func TestNoOpPrizeAmountResolver_Resolve_AlwaysReturnsNil(t *testing.T) {
	ctx := context.Background()
	resolver := usecases.NewNoOpPrizeAmountResolver()

	prizes, err := resolver.Resolve(ctx, pairing_out.PrizeResolveRequest{
		MatchID:         "m1",
		WinnerPlayerIDs: []string{"p1"},
		TenantID:        "t1",
		ClientID:        "c1",
	})

	assert.NoError(t, err)
	assert.Nil(t, prizes)
}
