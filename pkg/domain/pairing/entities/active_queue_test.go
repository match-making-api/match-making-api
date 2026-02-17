package entities_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
)

func TestInMemoryActiveQueueStore_RegisterAndGet(t *testing.T) {
	ctx := context.Background()
	store := entities.NewInMemoryActiveQueueStore()

	playerID := uuid.New()
	entry := &entities.ActiveQueueEntry{
		PlayerID:        playerID,
		GameID:          uuid.New(),
		RegionSlug:      "us-east-1",
		TenantID:        uuid.New().String(),
		ClientID:        uuid.New().String(),
		ResourceOwnerID: uuid.New().String(),
		Position:        3,
		JoinedAt:        time.Now().UTC(),
	}

	require.NoError(t, store.Register(ctx, entry))

	got, err := store.Get(ctx, playerID)
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, playerID, got.PlayerID)
	assert.Equal(t, "us-east-1", got.RegionSlug)
	assert.Equal(t, 3, got.Position)
}

func TestInMemoryActiveQueueStore_RegisterIdempotent(t *testing.T) {
	ctx := context.Background()
	store := entities.NewInMemoryActiveQueueStore()

	playerID := uuid.New()
	originalTime := time.Now().UTC().Add(-5 * time.Minute)

	require.NoError(t, store.Register(ctx, &entities.ActiveQueueEntry{
		PlayerID: playerID,
		Position: 5,
		JoinedAt: originalTime,
	}))

	require.NoError(t, store.Register(ctx, &entities.ActiveQueueEntry{
		PlayerID: playerID,
		Position: 2,
		JoinedAt: time.Now().UTC(),
	}))

	got, err := store.Get(ctx, playerID)
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, 2, got.Position)
	assert.Equal(t, originalTime, got.JoinedAt, "JoinedAt should be preserved on re-register")
}

func TestInMemoryActiveQueueStore_Remove(t *testing.T) {
	ctx := context.Background()
	store := entities.NewInMemoryActiveQueueStore()

	playerID := uuid.New()
	require.NoError(t, store.Register(ctx, &entities.ActiveQueueEntry{
		PlayerID: playerID,
		Position: 1,
		JoinedAt: time.Now().UTC(),
	}))

	count, _ := store.Count(ctx)
	assert.Equal(t, 1, count)

	require.NoError(t, store.Remove(ctx, playerID))
	count, _ = store.Count(ctx)
	assert.Equal(t, 0, count)

	got, err := store.Get(ctx, playerID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestInMemoryActiveQueueStore_RemoveNonExistent(t *testing.T) {
	ctx := context.Background()
	store := entities.NewInMemoryActiveQueueStore()
	require.NoError(t, store.Remove(ctx, uuid.New()))

	count, _ := store.Count(ctx)
	assert.Equal(t, 0, count)
}

func TestInMemoryActiveQueueStore_GetAll(t *testing.T) {
	ctx := context.Background()
	store := entities.NewInMemoryActiveQueueStore()

	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for i, id := range ids {
		require.NoError(t, store.Register(ctx, &entities.ActiveQueueEntry{
			PlayerID:   id,
			RegionSlug: "eu-west-1",
			Position:   i + 1,
			JoinedAt:   time.Now().UTC(),
		}))
	}

	all, err := store.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	// Verify returned copies don't share references
	all[0].Position = 999
	original, _ := store.Get(ctx, all[0].PlayerID)
	assert.NotEqual(t, 999, original.Position, "GetAll should return copies")
}

func TestInMemoryActiveQueueStore_UpdatePosition(t *testing.T) {
	ctx := context.Background()
	store := entities.NewInMemoryActiveQueueStore()

	playerID := uuid.New()
	require.NoError(t, store.Register(ctx, &entities.ActiveQueueEntry{
		PlayerID: playerID,
		Position: 10,
		JoinedAt: time.Now().UTC(),
	}))

	ok, err := store.UpdatePosition(ctx, playerID, 5)
	require.NoError(t, err)
	assert.True(t, ok)

	got, _ := store.Get(ctx, playerID)
	assert.Equal(t, 5, got.Position)
}

func TestInMemoryActiveQueueStore_UpdatePositionNotFound(t *testing.T) {
	ctx := context.Background()
	store := entities.NewInMemoryActiveQueueStore()

	ok, err := store.UpdatePosition(ctx, uuid.New(), 5)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestInMemoryActiveQueueStore_Count(t *testing.T) {
	ctx := context.Background()
	store := entities.NewInMemoryActiveQueueStore()

	count, _ := store.Count(ctx)
	assert.Equal(t, 0, count)

	require.NoError(t, store.Register(ctx, &entities.ActiveQueueEntry{
		PlayerID: uuid.New(),
		JoinedAt: time.Now().UTC(),
	}))
	count, _ = store.Count(ctx)
	assert.Equal(t, 1, count)

	require.NoError(t, store.Register(ctx, &entities.ActiveQueueEntry{
		PlayerID: uuid.New(),
		JoinedAt: time.Now().UTC(),
	}))
	count, _ = store.Count(ctx)
	assert.Equal(t, 2, count)
}
