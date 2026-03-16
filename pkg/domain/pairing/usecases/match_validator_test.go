package usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	parties_entities "github.com/leet-gaming/match-making-api/pkg/domain/parties/entities"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

func TestMatchValidator_Validate(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - all rules pass", func(t *testing.T) {
		validator := usecases.NewMatchValidator()
		pair := pairing_entities.NewPair(2, common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New()})
		pair.Match = map[uuid.UUID]*parties_entities.Party{
			uuid.New(): {ID: uuid.New()},
			uuid.New(): {ID: uuid.New()},
		}

		params := usecases.MatchValidationParams{
			Pair:            pair,
			ResourceOwnerID: uuid.New().String(),
			TenantID:        uuid.New().String(),
			ClientID:        uuid.New().String(),
		}

		err := validator.Validate(ctx, params)
		assert.NoError(t, err)
	})

	t.Run("Failure - missing resource_owner_id", func(t *testing.T) {
		validator := usecases.NewMatchValidator()
		pair := pairing_entities.NewPair(1, common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New()})
		pair.Match = map[uuid.UUID]*parties_entities.Party{uuid.New(): {ID: uuid.New()}}

		params := usecases.MatchValidationParams{
			Pair:            pair,
			ResourceOwnerID: "",
			TenantID:        uuid.New().String(),
			ClientID:        uuid.New().String(),
		}

		err := validator.Validate(ctx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource_owner_id")
	})

	t.Run("Failure - missing tenant_id", func(t *testing.T) {
		validator := usecases.NewMatchValidator()
		pair := pairing_entities.NewPair(1, common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New()})
		pair.Match = map[uuid.UUID]*parties_entities.Party{uuid.New(): {ID: uuid.New()}}

		params := usecases.MatchValidationParams{
			Pair:            pair,
			ResourceOwnerID: uuid.New().String(),
			TenantID:        "",
			ClientID:        uuid.New().String(),
		}

		err := validator.Validate(ctx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tenant_id")
	})

	t.Run("Failure - missing client_id", func(t *testing.T) {
		validator := usecases.NewMatchValidator()
		pair := pairing_entities.NewPair(1, common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New()})
		pair.Match = map[uuid.UUID]*parties_entities.Party{uuid.New(): {ID: uuid.New()}}

		params := usecases.MatchValidationParams{
			Pair:            pair,
			ResourceOwnerID: uuid.New().String(),
			TenantID:        uuid.New().String(),
			ClientID:        "",
		}

		err := validator.Validate(ctx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client_id")
	})

	t.Run("Failure - nil pair", func(t *testing.T) {
		validator := usecases.NewMatchValidator()
		params := usecases.MatchValidationParams{
			Pair:            nil,
			ResourceOwnerID: uuid.New().String(),
			TenantID:        uuid.New().String(),
			ClientID:        uuid.New().String(),
		}

		err := validator.Validate(ctx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pair is nil")
	})

	t.Run("Failure - empty pair", func(t *testing.T) {
		validator := usecases.NewMatchValidator()
		pair := pairing_entities.NewPair(0, common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New()})
		pair.Match = map[uuid.UUID]*parties_entities.Party{}

		params := usecases.MatchValidationParams{
			Pair:            pair,
			ResourceOwnerID: uuid.New().String(),
			TenantID:        uuid.New().String(),
			ClientID:        uuid.New().String(),
		}

		err := validator.Validate(ctx, params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no players")
	})
}

func TestMatchValidator_Validate_WithEnvelopeAndPayload(t *testing.T) {
	ctx := context.Background()
	validator := usecases.NewMatchValidator()

	envelope := &schemas.EventEnvelope{
		ResourceOwnerId: uuid.New().String(),
	}
	payload := &schemas.PlayerQueuedPayload{
		TenantId: uuid.New().String(),
		ClientId: uuid.New().String(),
	}
	pair := pairing_entities.NewPair(1, common.ResourceOwner{TenantID: uuid.New(), ClientID: uuid.New()})
	pair.Match = map[uuid.UUID]*parties_entities.Party{uuid.New(): {ID: uuid.New()}}

	params := usecases.MatchValidationParams{
		Pair:            pair,
		Envelope:        envelope,
		Payload:         payload,
		ResourceOwnerID: envelope.GetResourceOwnerId(),
		TenantID:        payload.GetTenantId(),
		ClientID:        payload.GetClientId(),
	}

	err := validator.Validate(ctx, params)
	assert.NoError(t, err)
}
