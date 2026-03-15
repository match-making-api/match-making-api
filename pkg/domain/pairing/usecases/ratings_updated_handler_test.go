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

type mockPlayerRatingRepository struct {
	mock.Mock
}

func (m *mockPlayerRatingRepository) GetByPlayer(ctx context.Context, playerID, gameID, tenantID, clientID string) (*entities.PlayerRating, error) {
	args := m.Called(ctx, playerID, gameID, tenantID, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PlayerRating), args.Error(1)
}

func (m *mockPlayerRatingRepository) Save(ctx context.Context, rating *entities.PlayerRating) error {
	args := m.Called(ctx, rating)
	return args.Error(0)
}

func (m *mockPlayerRatingRepository) SaveAuditEntry(ctx context.Context, entry *entities.RatingAuditEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

type mockRatingsProcessedStore struct {
	mock.Mock
}

func (m *mockRatingsProcessedStore) HasProcessed(ctx context.Context, matchID uuid.UUID) (bool, error) {
	args := m.Called(ctx, matchID)
	return args.Bool(0), args.Error(1)
}

func (m *mockRatingsProcessedStore) MarkProcessed(ctx context.Context, matchID uuid.UUID) error {
	args := m.Called(ctx, matchID)
	return args.Error(0)
}

type mockRatingsPublisher struct {
	mock.Mock
}

func (m *mockRatingsPublisher) PublishRatingsUpdatedProto(ctx context.Context, event *schemas.MatchmakingEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestRatingsUpdatedHandler_Handle_Idempotent(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockPlayerRatingRepository)
	mockStore := new(mockRatingsProcessedStore)
	mockPub := new(mockRatingsPublisher)
	handler := usecases.NewRatingsUpdatedHandler(mockRepo, mockStore, mockPub, nil)

	matchID := uuid.New()
	envelope := &schemas.EventEnvelope{
		Id:              uuid.New().String(),
		ResourceOwnerId: "rid",
	}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:             matchID.String(),
		PlayerIds:           []string{"p1", "p2"},
		WinnerTeamId:        "p1",
		IsDraw:              false,
		TenantId:            "t1",
		ClientId:            "c1",
		ResourceOwnerId:     "rid",
		CompletedAtEpochMs: 1700000000000,
		CalculatedAtEpochMs: 1700000001000,
	}

	mockStore.On("HasProcessed", ctx, matchID).Return(true, nil)

	err := handler.Handle(ctx, envelope, payload)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetByPlayer")
	mockPub.AssertNotCalled(t, "PublishRatingsUpdatedProto")
}

func TestRatingsUpdatedHandler_Handle_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockPlayerRatingRepository)
	mockStore := new(mockRatingsProcessedStore)
	mockPub := new(mockRatingsPublisher)
	handler := usecases.NewRatingsUpdatedHandler(mockRepo, mockStore, mockPub, nil)

	matchID := uuid.New()
	envelope := &schemas.EventEnvelope{
		Id:              uuid.New().String(),
		ResourceOwnerId: "rid",
	}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:             matchID.String(),
		PlayerIds:           []string{"p1", "p2"},
		WinnerTeamId:        "p1",
		IsDraw:              false,
		TenantId:            "t1",
		ClientId:            "c1",
		ResourceOwnerId:     "rid",
		CompletedAtEpochMs:  1700000000000,
		CalculatedAtEpochMs: 1700000001000,
	}

	mockStore.On("HasProcessed", ctx, matchID).Return(false, nil)
	mockRepo.On("GetByPlayer", ctx, "p1", "", "t1", "c1").Return(nil, nil)
	mockRepo.On("GetByPlayer", ctx, "p2", "", "t1", "c1").Return(nil, nil)
	mockRepo.On("Save", ctx, mock.MatchedBy(func(r *entities.PlayerRating) bool {
		return r.PlayerID == "p1" || r.PlayerID == "p2"
	})).Return(nil).Times(2)
	mockRepo.On("SaveAuditEntry", ctx, mock.Anything).Return(nil).Times(2)
	mockStore.On("MarkProcessed", ctx, matchID).Return(nil)
	mockPub.On("PublishRatingsUpdatedProto", ctx, mock.MatchedBy(func(e *schemas.MatchmakingEvent) bool {
		pl := e.GetRatingsUpdated()
		return pl != nil &&
			pl.MatchId == matchID.String() &&
			len(pl.Deltas) == 2 &&
			pl.ResourceOwnerId == "rid"
	})).Return(nil)

	err := handler.Handle(ctx, envelope, payload)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockStore.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}

func TestRatingsUpdatedHandler_Handle_SaveError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockPlayerRatingRepository)
	mockStore := new(mockRatingsProcessedStore)
	mockPub := new(mockRatingsPublisher)
	handler := usecases.NewRatingsUpdatedHandler(mockRepo, mockStore, mockPub, nil)

	matchID := uuid.New()
	envelope := &schemas.EventEnvelope{ResourceOwnerId: "rid"}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		PlayerIds:       []string{"p1", "p2"},
		WinnerTeamId:    "p1",
		IsDraw:          false,
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId: "rid",
	}

	mockStore.On("HasProcessed", ctx, matchID).Return(false, nil)
	mockRepo.On("GetByPlayer", ctx, "p1", "", "t1", "c1").Return(nil, nil)
	mockRepo.On("GetByPlayer", ctx, "p2", "", "t1", "c1").Return(nil, nil)
	mockRepo.On("Save", ctx, mock.Anything).Return(errors.New("db error"))

	err := handler.Handle(ctx, envelope, payload)

	assert.Error(t, err)
	mockPub.AssertNotCalled(t, "PublishRatingsUpdatedProto")
}

var _ pairing_out.PlayerRatingRepository = (*mockPlayerRatingRepository)(nil)
var _ pairing_out.RatingsProcessedStore = (*mockRatingsProcessedStore)(nil)
