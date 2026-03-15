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

type mockPrizeAmountResolver struct {
	mock.Mock
}

func (m *mockPrizeAmountResolver) Resolve(ctx context.Context, req pairing_out.PrizeResolveRequest) ([]pairing_out.WinnerPrize, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]pairing_out.WinnerPrize), args.Error(1)
}

type mockPrizesDistributedStore struct {
	mock.Mock
}

func (m *mockPrizesDistributedStore) HasDistributed(ctx context.Context, matchID uuid.UUID) (bool, error) {
	args := m.Called(ctx, matchID)
	return args.Bool(0), args.Error(1)
}

func (m *mockPrizesDistributedStore) MarkDistributed(ctx context.Context, matchID uuid.UUID) error {
	args := m.Called(ctx, matchID)
	return args.Error(0)
}

type mockPrizePublisher struct {
	mock.Mock
}

func (m *mockPrizePublisher) PublishPrizeDistributedProto(ctx context.Context, event *schemas.MatchmakingEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestPrizeDistributionHandler_Handle_Idempotent(t *testing.T) {
	ctx := context.Background()
	mockResolver := new(mockPrizeAmountResolver)
	mockStore := new(mockPrizesDistributedStore)
	mockPub := new(mockPrizePublisher)
	handler := usecases.NewPrizeDistributionHandler(mockResolver, mockStore, mockPub)

	matchID := uuid.New()
	envelope := &schemas.EventEnvelope{Id: uuid.New().String(), ResourceOwnerId: "rid"}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		PlayerIds:       []string{"p1", "p2"},
		WinnerTeamId:    "p1",
		IsDraw:          false,
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId:  "rid",
	}

	mockStore.On("HasDistributed", ctx, matchID).Return(true, nil)

	err := handler.Handle(ctx, envelope, payload)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
	mockResolver.AssertNotCalled(t, "Resolve")
	mockPub.AssertNotCalled(t, "PublishPrizeDistributedProto")
}

func TestPrizeDistributionHandler_Handle_NoWinners_Draw(t *testing.T) {
	ctx := context.Background()
	mockResolver := new(mockPrizeAmountResolver)
	mockStore := new(mockPrizesDistributedStore)
	mockPub := new(mockPrizePublisher)
	handler := usecases.NewPrizeDistributionHandler(mockResolver, mockStore, mockPub)

	matchID := uuid.New()
	envelope := &schemas.EventEnvelope{ResourceOwnerId: "rid"}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		PlayerIds:       []string{"p1", "p2"},
		IsDraw:          true,
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId: "rid",
	}

	mockStore.On("HasDistributed", ctx, matchID).Return(false, nil)

	err := handler.Handle(ctx, envelope, payload)

	assert.NoError(t, err)
	mockResolver.AssertNotCalled(t, "Resolve")
	mockPub.AssertNotCalled(t, "PublishPrizeDistributedProto")
}

func TestPrizeDistributionHandler_Handle_NoWinners_Unknown(t *testing.T) {
	ctx := context.Background()
	mockResolver := new(mockPrizeAmountResolver)
	mockStore := new(mockPrizesDistributedStore)
	mockPub := new(mockPrizePublisher)
	handler := usecases.NewPrizeDistributionHandler(mockResolver, mockStore, mockPub)

	matchID := uuid.New()
	envelope := &schemas.EventEnvelope{ResourceOwnerId: "rid"}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		PlayerIds:       []string{"p1", "p2"},
		WinnerTeamId:    "team-x",
		IsDraw:          false,
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId: "rid",
	}

	mockStore.On("HasDistributed", ctx, matchID).Return(false, nil)

	err := handler.Handle(ctx, envelope, payload)

	assert.NoError(t, err)
	mockResolver.AssertNotCalled(t, "Resolve")
	mockPub.AssertNotCalled(t, "PublishPrizeDistributedProto")
}

func TestPrizeDistributionHandler_Handle_MissingResourceOwnership(t *testing.T) {
	ctx := context.Background()
	mockResolver := new(mockPrizeAmountResolver)
	mockStore := new(mockPrizesDistributedStore)
	mockPub := new(mockPrizePublisher)
	handler := usecases.NewPrizeDistributionHandler(mockResolver, mockStore, mockPub)

	matchID := uuid.New()
	envelope := &schemas.EventEnvelope{ResourceOwnerId: "rid"}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		PlayerIds:       []string{"p1", "p2"},
		WinnerTeamId:    "p1",
		IsDraw:          false,
		TenantId:        "",
		ClientId:        "c1",
		ResourceOwnerId:  "rid",
	}

	mockStore.On("HasDistributed", ctx, matchID).Return(false, nil)

	err := handler.Handle(ctx, envelope, payload)

	assert.NoError(t, err)
	mockResolver.AssertNotCalled(t, "Resolve")
	mockPub.AssertNotCalled(t, "PublishPrizeDistributedProto")
}

func TestPrizeDistributionHandler_Handle_NoPrizesResolved(t *testing.T) {
	ctx := context.Background()
	mockResolver := new(mockPrizeAmountResolver)
	mockStore := new(mockPrizesDistributedStore)
	mockPub := new(mockPrizePublisher)
	handler := usecases.NewPrizeDistributionHandler(mockResolver, mockStore, mockPub)

	matchID := uuid.New()
	envelope := &schemas.EventEnvelope{ResourceOwnerId: "rid"}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		PlayerIds:       []string{"p1", "p2"},
		WinnerTeamId:    "p1",
		IsDraw:          false,
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId:  "rid",
	}

	mockStore.On("HasDistributed", ctx, matchID).Return(false, nil)
	mockResolver.On("Resolve", ctx, mock.MatchedBy(func(r pairing_out.PrizeResolveRequest) bool {
		return r.MatchID == matchID.String() && len(r.WinnerPlayerIDs) == 1 && r.WinnerPlayerIDs[0] == "p1"
	})).Return([]pairing_out.WinnerPrize{}, nil)

	err := handler.Handle(ctx, envelope, payload)

	assert.NoError(t, err)
	mockResolver.AssertExpectations(t)
	mockPub.AssertNotCalled(t, "PublishPrizeDistributedProto")
}

func TestPrizeDistributionHandler_Handle_Success(t *testing.T) {
	ctx := context.Background()
	mockResolver := new(mockPrizeAmountResolver)
	mockStore := new(mockPrizesDistributedStore)
	mockPub := new(mockPrizePublisher)
	handler := usecases.NewPrizeDistributionHandler(mockResolver, mockStore, mockPub)

	matchID := uuid.New()
	envelope := &schemas.EventEnvelope{
		Id:              uuid.New().String(),
		ResourceOwnerId: "rid",
		CorrelationId:   "corr-1",
	}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         matchID.String(),
		PlayerIds:       []string{"p1", "p2"},
		WinnerTeamId:    "p1",
		IsDraw:          false,
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId: "rid",
		LobbyId:         strPtr("lobby-1"),
		PrizePoolId:     strPtr("pool-1"),
	}

	prizes := []pairing_out.WinnerPrize{
		{PlayerID: "p1", AmountCents: 1000, Currency: "USD"},
	}

	mockStore.On("HasDistributed", ctx, matchID).Return(false, nil)
	mockResolver.On("Resolve", ctx, mock.Anything).Return(prizes, nil)
	mockPub.On("PublishPrizeDistributedProto", ctx, mock.MatchedBy(func(e *schemas.MatchmakingEvent) bool {
		pl := e.GetPrizeDistributed()
		return pl != nil &&
			pl.MatchId == matchID.String() &&
			len(pl.WinnerDetails) == 1 &&
			pl.WinnerDetails[0].PlayerId == "p1" &&
			pl.WinnerDetails[0].AmountCents == 1000 &&
			pl.TenantId == "t1" &&
			pl.ResourceOwnerId == "rid" &&
			pl.LobbyId == "lobby-1" &&
			pl.PrizePoolId == "pool-1"
	})).Return(nil)
	mockStore.On("MarkDistributed", ctx, matchID).Return(nil)

	err := handler.Handle(ctx, envelope, payload)

	assert.NoError(t, err)
	mockResolver.AssertExpectations(t)
	mockStore.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}

func TestPrizeDistributionHandler_Handle_ResolverError(t *testing.T) {
	ctx := context.Background()
	mockResolver := new(mockPrizeAmountResolver)
	mockStore := new(mockPrizesDistributedStore)
	mockPub := new(mockPrizePublisher)
	handler := usecases.NewPrizeDistributionHandler(mockResolver, mockStore, mockPub)

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

	mockStore.On("HasDistributed", ctx, matchID).Return(false, nil)
	mockResolver.On("Resolve", ctx, mock.Anything).Return(nil, errors.New("resolver error"))

	err := handler.Handle(ctx, envelope, payload)

	assert.Error(t, err)
	mockPub.AssertNotCalled(t, "PublishPrizeDistributedProto")
}

func TestPrizeDistributionHandler_Handle_HasDistributedError(t *testing.T) {
	ctx := context.Background()
	mockResolver := new(mockPrizeAmountResolver)
	mockStore := new(mockPrizesDistributedStore)
	mockPub := new(mockPrizePublisher)
	handler := usecases.NewPrizeDistributionHandler(mockResolver, mockStore, mockPub)

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

	mockStore.On("HasDistributed", ctx, matchID).Return(false, errors.New("db error"))

	err := handler.Handle(ctx, envelope, payload)

	assert.Error(t, err)
	mockResolver.AssertNotCalled(t, "Resolve")
	mockPub.AssertNotCalled(t, "PublishPrizeDistributedProto")
}

func TestPrizeDistributionHandler_Handle_PublishError(t *testing.T) {
	ctx := context.Background()
	mockResolver := new(mockPrizeAmountResolver)
	mockStore := new(mockPrizesDistributedStore)
	mockPub := new(mockPrizePublisher)
	handler := usecases.NewPrizeDistributionHandler(mockResolver, mockStore, mockPub)

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

	prizes := []pairing_out.WinnerPrize{
		{PlayerID: "p1", AmountCents: 1000, Currency: "USD"},
	}

	mockStore.On("HasDistributed", ctx, matchID).Return(false, nil)
	mockResolver.On("Resolve", ctx, mock.Anything).Return(prizes, nil)
	mockPub.On("PublishPrizeDistributedProto", ctx, mock.Anything).Return(errors.New("kafka error"))

	err := handler.Handle(ctx, envelope, payload)

	assert.Error(t, err)
	mockStore.AssertNotCalled(t, "MarkDistributed")
}

func TestPrizeDistributionHandler_Handle_InvalidMatchID(t *testing.T) {
	ctx := context.Background()
	mockResolver := new(mockPrizeAmountResolver)
	mockStore := new(mockPrizesDistributedStore)
	mockPub := new(mockPrizePublisher)
	handler := usecases.NewPrizeDistributionHandler(mockResolver, mockStore, mockPub)

	envelope := &schemas.EventEnvelope{ResourceOwnerId: "rid"}
	payload := &schemas.MatchResultsCalculatedPayload{
		MatchId:         "not-a-uuid",
		PlayerIds:       []string{"p1", "p2"},
		WinnerTeamId:    "p1",
		IsDraw:          false,
		TenantId:        "t1",
		ClientId:        "c1",
		ResourceOwnerId: "rid",
	}

	err := handler.Handle(ctx, envelope, payload)

	assert.NoError(t, err)
	mockStore.AssertNotCalled(t, "HasDistributed")
	mockResolver.AssertNotCalled(t, "Resolve")
	mockPub.AssertNotCalled(t, "PublishPrizeDistributedProto")
}

func TestPrizeDistributionHandler_Handle_MarkDistributedError(t *testing.T) {
	ctx := context.Background()
	matchID := uuid.New()
	mockResolver := new(mockPrizeAmountResolver)
	mockStore := new(mockPrizesDistributedStore)
	mockPub := new(mockPrizePublisher)
	handler := usecases.NewPrizeDistributionHandler(mockResolver, mockStore, mockPub)

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

	mockStore.On("HasDistributed", ctx, matchID).Return(false, nil)
	mockResolver.On("Resolve", ctx, mock.Anything).Return([]pairing_out.WinnerPrize{
		{PlayerID: "p1", AmountCents: 1000, Currency: "USD"},
	}, nil)
	mockPub.On("PublishPrizeDistributedProto", ctx, mock.MatchedBy(func(e *schemas.MatchmakingEvent) bool {
		pl := e.GetPrizeDistributed()
		return pl != nil && pl.MatchId == matchID.String() && len(pl.WinnerDetails) == 1
	})).Return(nil)
	mockStore.On("MarkDistributed", ctx, matchID).Return(errors.New("db error"))

	err := handler.Handle(ctx, envelope, payload)

	assert.Error(t, err)
	mockPub.AssertExpectations(t)
}

func TestNoOpPrizeAmountResolver_Resolve(t *testing.T) {
	ctx := context.Background()
	resolver := usecases.NewNoOpPrizeAmountResolver()
	req := pairing_out.PrizeResolveRequest{
		MatchID:         uuid.New().String(),
		WinnerPlayerIDs: []string{"p1"},
		TenantID:        "t1",
		ClientID:        "c1",
	}

	prizes, err := resolver.Resolve(ctx, req)

	assert.NoError(t, err)
	assert.Nil(t, prizes)
}

func strPtr(s string) *string {
	return &s
}

var _ pairing_out.PrizeAmountResolver = (*mockPrizeAmountResolver)(nil)
var _ pairing_out.PrizesDistributedStore = (*mockPrizesDistributedStore)(nil)
