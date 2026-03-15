package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

type mockMatchResultRepository struct {
	mock.Mock
}

func (m *mockMatchResultRepository) GetByMatchID(ctx context.Context, matchID uuid.UUID) (*entities.MatchResult, error) {
	args := m.Called(ctx, matchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.MatchResult), args.Error(1)
}

func (m *mockMatchResultRepository) Save(ctx context.Context, result *entities.MatchResult) (*entities.MatchResult, error) {
	args := m.Called(ctx, result)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.MatchResult), args.Error(1)
}

type mockMatchResultsPublisher struct {
	mock.Mock
}

func (m *mockMatchResultsPublisher) PublishMatchResultsCalculatedProto(ctx context.Context, event *schemas.MatchmakingEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestMatchCompletedHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - persists and publishes MatchResultsCalculated", func(t *testing.T) {
		mockRepo := new(mockMatchResultRepository)
		mockPub := new(mockMatchResultsPublisher)
		handler := usecases.NewMatchCompletedHandler(mockRepo, mockPub)

		player1 := uuid.New().String()
		player2 := uuid.New().String()
		matchID := uuid.New()
		resourceOwnerID := uuid.New().String()

		envelope := &schemas.EventEnvelope{
			Id:              uuid.New().String(),
			Type:            schemas.EventTypeMatchCompleted,
			Source:          "game-server",
			ResourceOwnerId: resourceOwnerID,
		}
		payload := &schemas.MatchCompletedPayload{
			MatchId:             matchID.String(),
			PlayerIds:           []string{player1, player2},
			WinnerTeamId:        "team-a",
			IsDraw:              false,
			CompletedAtEpochMs:  1700000000000,
			TenantId:            "tenant-1",
			ClientId:            "client-1",
		}

		mockRepo.On("GetByMatchID", ctx, matchID).Return(nil, nil)
		saved := &entities.MatchResult{
			MatchID:        matchID,
			PlayerIDs:      []string{player1, player2},
			WinnerTeamID:   "team-a",
			IsDraw:         false,
			CompletedAtMs:  1700000000000,
			TenantID:       "tenant-1",
			ClientID:       "client-1",
			ResourceOwnerID: resourceOwnerID,
		}
		mockRepo.On("Save", ctx, mock.MatchedBy(func(r *entities.MatchResult) bool {
			return r.MatchID == matchID && r.WinnerTeamID == "team-a"
		})).Return(saved, nil)
		mockPub.On("PublishMatchResultsCalculatedProto", ctx, mock.MatchedBy(func(e *schemas.MatchmakingEvent) bool {
			pl := e.GetMatchResultsCalculated()
			return pl != nil &&
				pl.MatchId == matchID.String() &&
				pl.WinnerTeamId == "team-a" &&
				pl.ResourceOwnerId == resourceOwnerID
		})).Return(nil)

		err := handler.Handle(ctx, envelope, payload)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("Idempotent - skips when result already exists", func(t *testing.T) {
		mockRepo := new(mockMatchResultRepository)
		mockPub := new(mockMatchResultsPublisher)
		handler := usecases.NewMatchCompletedHandler(mockRepo, mockPub)

		matchID := uuid.New()
		existing := &entities.MatchResult{
			MatchID: matchID,
			PlayerIDs: []string{"p1", "p2"},
		}

		envelope := &schemas.EventEnvelope{ResourceOwnerId: "rid"}
		payload := &schemas.MatchCompletedPayload{
			MatchId:             matchID.String(),
			PlayerIds:           []string{"p1", "p2"},
			CompletedAtEpochMs:  1700000000000,
			TenantId:            "t1",
			ClientId:            "c1",
		}

		mockRepo.On("GetByMatchID", ctx, matchID).Return(existing, nil)
		// Save and Publish must NOT be called

		err := handler.Handle(ctx, envelope, payload)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockPub.AssertNotCalled(t, "PublishMatchResultsCalculatedProto")
	})

	t.Run("Invalid match_id - returns nil without retry", func(t *testing.T) {
		mockRepo := new(mockMatchResultRepository)
		mockPub := new(mockMatchResultsPublisher)
		handler := usecases.NewMatchCompletedHandler(mockRepo, mockPub)

		envelope := &schemas.EventEnvelope{ResourceOwnerId: "rid"}
		payload := &schemas.MatchCompletedPayload{
			MatchId:   "not-a-uuid",
			TenantId:  "t1",
			ClientId:  "c1",
		}

		err := handler.Handle(ctx, envelope, payload)

		assert.NoError(t, err)
		mockRepo.AssertNotCalled(t, "GetByMatchID")
		mockPub.AssertNotCalled(t, "PublishMatchResultsCalculatedProto")
	})

	t.Run("Save error - returns error for retry", func(t *testing.T) {
		mockRepo := new(mockMatchResultRepository)
		mockPub := new(mockMatchResultsPublisher)
		handler := usecases.NewMatchCompletedHandler(mockRepo, mockPub)

		matchID := uuid.New()
		envelope := &schemas.EventEnvelope{ResourceOwnerId: "rid"}
		payload := &schemas.MatchCompletedPayload{
			MatchId:             matchID.String(),
			PlayerIds:           []string{"p1"},
			CompletedAtEpochMs:  1700000000000,
			TenantId:            "t1",
			ClientId:            "c1",
		}

		mockRepo.On("GetByMatchID", ctx, matchID).Return(nil, nil)
		mockRepo.On("Save", ctx, mock.Anything).Return(nil, errors.New("db error"))

		err := handler.Handle(ctx, envelope, payload)

		assert.Error(t, err)
		mockPub.AssertNotCalled(t, "PublishMatchResultsCalculatedProto")
	})

	t.Run("Publish error - returns error for retry", func(t *testing.T) {
		mockRepo := new(mockMatchResultRepository)
		mockPub := new(mockMatchResultsPublisher)
		handler := usecases.NewMatchCompletedHandler(mockRepo, mockPub)

		matchID := uuid.New()
		saved := &entities.MatchResult{
			MatchID: matchID,
			PlayerIDs: []string{"p1"},
		}

		envelope := &schemas.EventEnvelope{ResourceOwnerId: "rid"}
		payload := &schemas.MatchCompletedPayload{
			MatchId:             matchID.String(),
			PlayerIds:           []string{"p1"},
			CompletedAtEpochMs:  1700000000000,
			TenantId:            "t1",
			ClientId:            "c1",
		}

		mockRepo.On("GetByMatchID", ctx, matchID).Return(nil, nil)
		mockRepo.On("Save", ctx, mock.Anything).Return(saved, nil)
		mockPub.On("PublishMatchResultsCalculatedProto", ctx, mock.Anything).Return(errors.New("kafka error"))

		err := handler.Handle(ctx, envelope, payload)

		assert.Error(t, err)
	})
}

// Ensure mockMatchResultRepository satisfies the interface.
var _ pairing_out.MatchResultRepository = (*mockMatchResultRepository)(nil)
