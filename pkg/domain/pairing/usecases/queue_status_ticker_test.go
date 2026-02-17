package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestQueueStatusTicker_Disabled(t *testing.T) {
	store := pairing_entities.NewInMemoryActiveQueueStore()

	cfg := &usecases.QueueStatusTickerConfig{
		Interval:                 100 * time.Millisecond,
		EstimatedWaitPerPosition: 10 * time.Second,
		Enabled:                  false,
	}

	ticker := usecases.NewQueueStatusTicker(store, nil, nil, nil, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := ticker.Start(ctx)
	assert.NoError(t, err, "disabled ticker should return immediately without error")
}

func TestQueueStatusTicker_EmptyQueue(t *testing.T) {
	store := pairing_entities.NewInMemoryActiveQueueStore()
	mockRegionReader := &mocks.MockPortRegionReader{}
	mockPoolReader := &mocks.MockPoolReader{}

	cfg := &usecases.QueueStatusTickerConfig{
		Interval:                 50 * time.Millisecond,
		EstimatedWaitPerPosition: 10 * time.Second,
		Enabled:                  true,
	}

	ticker := usecases.NewQueueStatusTicker(store, mockPoolReader, mockRegionReader, nil, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err := ticker.Start(ctx)
	assert.NoError(t, err)

	mockRegionReader.AssertNotCalled(t, "Search", mock.Anything, mock.Anything)
}

func TestQueueStatusTicker_RefreshesPosition(t *testing.T) {
	ctx := context.Background()
	store := pairing_entities.NewInMemoryActiveQueueStore()
	mockRegionReader := &mocks.MockPortRegionReader{}
	mockPoolReader := &mocks.MockPoolReader{}

	playerID := uuid.New()
	gameID := uuid.New()
	regionSlug := "us-east-1"

	region := &game_entities.Region{
		Name: "US East",
		Slug: regionSlug,
	}
	region.ID = uuid.New()

	_ = store.Register(ctx, &pairing_entities.ActiveQueueEntry{
		PlayerID:        playerID,
		GameID:          gameID,
		RegionSlug:      regionSlug,
		TenantID:        uuid.New().String(),
		ClientID:        uuid.New().String(),
		ResourceOwnerID: uuid.New().String(),
		Position:        10,
		JoinedAt:        time.Now().UTC(),
	})

	mockRegionReader.On("Search", mock.Anything, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)

	pool := &pairing_entities.Pool{
		Parties: []uuid.UUID{uuid.New(), playerID, uuid.New()},
	}
	mockPoolReader.On("FindPool", mock.Anything).Return(pool, nil)

	cfg := &usecases.QueueStatusTickerConfig{
		Interval:                 50 * time.Millisecond,
		EstimatedWaitPerPosition: 10 * time.Second,
		Enabled:                  true,
	}

	ticker := usecases.NewQueueStatusTicker(store, mockPoolReader, mockRegionReader, nil, cfg)

	tickerCtx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()

	_ = ticker.Start(tickerCtx)

	mockRegionReader.AssertCalled(t, "Search", mock.Anything, map[string]interface{}{"slug": regionSlug})
	mockPoolReader.AssertCalled(t, "FindPool", mock.Anything)

	entry, _ := store.Get(ctx, playerID)
	assert.NotNil(t, entry)
	assert.Equal(t, 2, entry.Position, "Position should be refreshed from pool state (1-based)")
}

func TestQueueStatusTicker_RemovesPlayerNotInPool(t *testing.T) {
	ctx := context.Background()
	store := pairing_entities.NewInMemoryActiveQueueStore()
	mockRegionReader := &mocks.MockPortRegionReader{}
	mockPoolReader := &mocks.MockPoolReader{}

	playerID := uuid.New()
	gameID := uuid.New()
	regionSlug := "eu-west-1"

	region := &game_entities.Region{
		Name: "EU West",
		Slug: regionSlug,
	}
	region.ID = uuid.New()

	_ = store.Register(ctx, &pairing_entities.ActiveQueueEntry{
		PlayerID:   playerID,
		GameID:     gameID,
		RegionSlug: regionSlug,
		Position:   1,
		JoinedAt:   time.Now().UTC(),
	})

	mockRegionReader.On("Search", mock.Anything, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)

	pool := &pairing_entities.Pool{
		Parties: []uuid.UUID{uuid.New()},
	}
	mockPoolReader.On("FindPool", mock.Anything).Return(pool, nil)

	cfg := &usecases.QueueStatusTickerConfig{
		Interval:                 50 * time.Millisecond,
		EstimatedWaitPerPosition: 10 * time.Second,
		Enabled:                  true,
	}

	ticker := usecases.NewQueueStatusTicker(store, mockPoolReader, mockRegionReader, nil, cfg)

	tickerCtx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()

	_ = ticker.Start(tickerCtx)

	got, _ := store.Get(ctx, playerID)
	assert.Nil(t, got, "Player not in pool should be removed from active queue")

	count, _ := store.Count(ctx)
	assert.Equal(t, 0, count)
}

func TestQueueStatusTicker_PoolNotFound(t *testing.T) {
	ctx := context.Background()
	store := pairing_entities.NewInMemoryActiveQueueStore()
	mockRegionReader := &mocks.MockPortRegionReader{}
	mockPoolReader := &mocks.MockPoolReader{}

	playerID := uuid.New()
	gameID := uuid.New()
	regionSlug := "ap-south-1"

	region := &game_entities.Region{
		Name: "AP South",
		Slug: regionSlug,
	}
	region.ID = uuid.New()

	_ = store.Register(ctx, &pairing_entities.ActiveQueueEntry{
		PlayerID:   playerID,
		GameID:     gameID,
		RegionSlug: regionSlug,
		Position:   1,
		JoinedAt:   time.Now().UTC(),
	})

	mockRegionReader.On("Search", mock.Anything, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{region}, nil)
	mockPoolReader.On("FindPool", mock.Anything).Return((*pairing_entities.Pool)(nil), nil)

	cfg := &usecases.QueueStatusTickerConfig{
		Interval:                 50 * time.Millisecond,
		EstimatedWaitPerPosition: 10 * time.Second,
		Enabled:                  true,
	}

	ticker := usecases.NewQueueStatusTicker(store, mockPoolReader, mockRegionReader, nil, cfg)

	tickerCtx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()

	_ = ticker.Start(tickerCtx)

	got, _ := store.Get(ctx, playerID)
	assert.Nil(t, got, "Player should be removed when pool not found")
}

func TestQueueStatusTicker_RegionNotFound(t *testing.T) {
	ctx := context.Background()
	store := pairing_entities.NewInMemoryActiveQueueStore()
	mockRegionReader := &mocks.MockPortRegionReader{}
	mockPoolReader := &mocks.MockPoolReader{}

	playerID := uuid.New()
	gameID := uuid.New()
	regionSlug := "unknown-region"

	_ = store.Register(ctx, &pairing_entities.ActiveQueueEntry{
		PlayerID:   playerID,
		GameID:     gameID,
		RegionSlug: regionSlug,
		Position:   1,
		JoinedAt:   time.Now().UTC(),
	})

	mockRegionReader.On("Search", mock.Anything, map[string]interface{}{"slug": regionSlug}).Return([]*game_entities.Region{}, nil)

	cfg := &usecases.QueueStatusTickerConfig{
		Interval:                 50 * time.Millisecond,
		EstimatedWaitPerPosition: 10 * time.Second,
		Enabled:                  true,
	}

	ticker := usecases.NewQueueStatusTicker(store, mockPoolReader, mockRegionReader, nil, cfg)

	tickerCtx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()

	_ = ticker.Start(tickerCtx)

	got, _ := store.Get(ctx, playerID)
	assert.Nil(t, got, "Player should be removed when region not found")
	mockPoolReader.AssertNotCalled(t, "FindPool", mock.Anything)
}

func TestDefaultQueueStatusTickerConfig(t *testing.T) {
	cfg := usecases.DefaultQueueStatusTickerConfig()

	assert.Equal(t, 5*time.Second, cfg.Interval)
	assert.Equal(t, 10*time.Second, cfg.EstimatedWaitPerPosition)
	assert.True(t, cfg.Enabled)
}

func TestQueueStatusTicker_NilConfig(t *testing.T) {
	store := pairing_entities.NewInMemoryActiveQueueStore()

	ticker := usecases.NewQueueStatusTicker(store, nil, nil, nil, nil)
	assert.NotNil(t, ticker, "should use default config when nil is passed")
}
