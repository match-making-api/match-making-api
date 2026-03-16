package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

type mockAnalyticsTrackedStore struct {
	mock.Mock
}

func (m *mockAnalyticsTrackedStore) HasTracked(ctx context.Context, matchID uuid.UUID) (bool, error) {
	args := m.Called(ctx, matchID)
	return args.Bool(0), args.Error(1)
}

func (m *mockAnalyticsTrackedStore) MarkTracked(ctx context.Context, matchID uuid.UUID) error {
	args := m.Called(ctx, matchID)
	return args.Error(0)
}

type mockAnalyticsPublisher struct {
	mock.Mock
}

func (m *mockAnalyticsPublisher) PublishAnalyticsTrackedProto(ctx context.Context, event *schemas.MatchmakingEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestAnalyticsTrackedHandler_Handle_Idempotent(t *testing.T) {
	ctx := context.Background()
	mockStore := new(mockAnalyticsTrackedStore)
	mockPub := new(mockAnalyticsPublisher)
	handler := usecases.NewAnalyticsTrackedHandler(mockStore, mockPub)

	matchID := uuid.New()
	envelope := &schemas.EventEnvelope{Id: uuid.New().String(), ResourceOwnerId: "rid"}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		PlayerIds:       []string{"p1", "p2"},
		WinnerTeamId:    "p1",
		IsDraw:          false,
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId: "rid",
	}

	mockStore.On("HasTracked", ctx, matchID).Return(true, nil)

	err := handler.Handle(ctx, envelope, payload)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockPub.AssertNotCalled(t, "PublishAnalyticsTrackedProto")
}

func TestAnalyticsTrackedHandler_Handle_MissingResourceOwnership(t *testing.T) {
	ctx := context.Background()
	mockStore := new(mockAnalyticsTrackedStore)
	mockPub := new(mockAnalyticsPublisher)
	handler := usecases.NewAnalyticsTrackedHandler(mockStore, mockPub)

	matchID := uuid.New()
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		PlayerIds:       []string{"p1", "p2"},
		WinnerTeamId:    "p1",
		IsDraw:          false,
		TenantId:        "", // missing
		ClientId:        "c1",
		ResourceOwnerId: "rid",
	}

	mockStore.On("HasTracked", ctx, matchID).Return(false, nil)

	err := handler.Handle(ctx, &schemas.EventEnvelope{}, payload)

	assert.NoError(t, err)
	mockPub.AssertNotCalled(t, "PublishAnalyticsTrackedProto")
	mockStore.AssertNumberOfCalls(t, "MarkTracked", 0)
}

func TestAnalyticsTrackedHandler_Handle_Success(t *testing.T) {
	ctx := context.Background()
	mockStore := new(mockAnalyticsTrackedStore)
	mockPub := new(mockAnalyticsPublisher)
	handler := usecases.NewAnalyticsTrackedHandler(mockStore, mockPub)

	matchID := uuid.New()
	envelope := &schemas.EventEnvelope{Id: uuid.New().String(), ResourceOwnerId: "rid"}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		PlayerIds:       []string{"p1", "p2"},
		WinnerTeamId:    "p1",
		IsDraw:          false,
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId: "rid",
	}

	mockStore.On("HasTracked", ctx, matchID).Return(false, nil)
	mockPub.On("PublishAnalyticsTrackedProto", ctx, mock.MatchedBy(func(e *schemas.MatchmakingEvent) bool {
		pl := e.GetAnalyticsTracked()
		return pl != nil &&
			pl.MatchId == matchID.String() &&
			len(pl.PlayerIds) == 2 &&
			pl.WinnerTeamId == "p1" &&
			pl.ResourceOwnerId == "rid"
	})).Return(nil)
	mockStore.On("MarkTracked", ctx, matchID).Return(nil)

	err := handler.Handle(ctx, envelope, payload)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}

func TestAnalyticsTrackedHandler_Handle_HasTrackedError(t *testing.T) {
	ctx := context.Background()
	mockStore := new(mockAnalyticsTrackedStore)
	mockPub := new(mockAnalyticsPublisher)
	handler := usecases.NewAnalyticsTrackedHandler(mockStore, mockPub)

	matchID := uuid.New()
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId: "rid",
	}

	mockStore.On("HasTracked", ctx, matchID).Return(false, errors.New("db error"))

	err := handler.Handle(ctx, &schemas.EventEnvelope{}, payload)

	assert.Error(t, err)
	mockPub.AssertNotCalled(t, "PublishAnalyticsTrackedProto")
}

func TestAnalyticsTrackedHandler_Handle_PublishError(t *testing.T) {
	ctx := context.Background()
	mockStore := new(mockAnalyticsTrackedStore)
	mockPub := new(mockAnalyticsPublisher)
	handler := usecases.NewAnalyticsTrackedHandler(mockStore, mockPub)

	matchID := uuid.New()
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		PlayerIds:       []string{"p1"},
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId: "rid",
	}

	mockStore.On("HasTracked", ctx, matchID).Return(false, nil)
	mockPub.On("PublishAnalyticsTrackedProto", ctx, mock.Anything).Return(errors.New("kafka error"))

	err := handler.Handle(ctx, &schemas.EventEnvelope{}, payload)

	assert.Error(t, err)
	mockStore.AssertNumberOfCalls(t, "MarkTracked", 0)
}

func TestAnalyticsTrackedHandler_Handle_InvalidMatchID(t *testing.T) {
	ctx := context.Background()
	mockStore := new(mockAnalyticsTrackedStore)
	mockPub := new(mockAnalyticsPublisher)
	handler := usecases.NewAnalyticsTrackedHandler(mockStore, mockPub)

	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         "not-a-uuid",
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId: "rid",
	}

	err := handler.Handle(ctx, &schemas.EventEnvelope{}, payload)

	assert.NoError(t, err) // handler returns nil on invalid UUID (skip)
	mockStore.AssertNotCalled(t, "HasTracked")
	mockPub.AssertNotCalled(t, "PublishAnalyticsTrackedProto")
}

var _ pairing_out.AnalyticsTrackedStore = (*mockAnalyticsTrackedStore)(nil)
