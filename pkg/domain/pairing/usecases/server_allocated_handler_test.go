package usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

type mockBroadcastPublisher struct {
	mock.Mock
}

func (m *mockBroadcastPublisher) PublishWebSocketBroadcast(ctx context.Context, event *kafka.WebSocketBroadcastEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestServerAllocatedHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - broadcasts MatchReady to players", func(t *testing.T) {
		mockPub := new(mockBroadcastPublisher)
		handler := usecases.NewServerAllocatedHandler(mockPub)

		player1 := uuid.New().String()
		player2 := uuid.New().String()
		matchID := uuid.New().String()

		envelope := &schemas.EventEnvelope{
			Id:              uuid.New().String(),
			Type:            schemas.EventTypeServerAllocated,
			Source:          "game-server",
			ResourceOwnerId: uuid.New().String(),
		}
		payload := &schemas.ServerAllocatedPayload{
			MatchId:         matchID,
			ServerId:        "gs-001",
			Region:          "us-east-1",
			ResourceOwnerId: uuid.New().String(),
			PlayerIds:       []string{player1, player2},
		}

		mockPub.On("PublishWebSocketBroadcast", ctx, mock.MatchedBy(func(e *kafka.WebSocketBroadcastEvent) bool {
			return e.Type == kafka.EventTypeMatchReady &&
				len(e.TargetIDs) == 2 &&
				e.LobbyID != nil &&
				e.LobbyID.String() == matchID
		})).Return(nil)

		err := handler.Handle(ctx, envelope, payload)

		assert.NoError(t, err)
		mockPub.AssertExpectations(t)
	})

	t.Run("Skips invalid player_id", func(t *testing.T) {
		mockPub := new(mockBroadcastPublisher)
		handler := usecases.NewServerAllocatedHandler(mockPub)

		validPlayer := uuid.New().String()
		envelope := &schemas.EventEnvelope{ResourceOwnerId: uuid.New().String()}
		payload := &schemas.ServerAllocatedPayload{
			MatchId:         uuid.New().String(),
			ServerId:        "gs-001",
			Region:          "us-east-1",
			ResourceOwnerId: uuid.New().String(),
			PlayerIds:       []string{validPlayer, "not-a-uuid"},
		}

		mockPub.On("PublishWebSocketBroadcast", ctx, mock.MatchedBy(func(e *kafka.WebSocketBroadcastEvent) bool {
			return len(e.TargetIDs) == 1 && e.TargetIDs[0].String() == validPlayer
		})).Return(nil)

		err := handler.Handle(ctx, envelope, payload)

		assert.NoError(t, err)
		mockPub.AssertExpectations(t)
	})
}
