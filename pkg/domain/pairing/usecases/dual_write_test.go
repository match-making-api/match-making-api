package usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
	"github.com/leet-gaming/match-making-api/test/mocks"
)

func TestHandlePlayerQueuedProto_SkippedWhenDualWriteDisabled(t *testing.T) {
	mockAddAndFind := &MockAddAndFindNextPairUseCase{}
	mockEventPublisher := &MockEventPublisher{}
	mockRegionReader := &mocks.MockPortRegionReader{}

	consumer := usecases.NewMatchmakingEventConsumer(
		mockAddAndFind, mockEventPublisher, mockRegionReader, nil, nil, nil, nil,
	)
	consumer.SetDualWrite(config.DualWriteConfig{
		Enabled:       false,
		FailurePolicy: config.DualWriteFailLegacyOnlyAndAlert,
	})

	err := consumer.HandlePlayerQueuedProto(context.Background(),
		&schemas.EventEnvelope{Id: "e1", ResourceOwnerId: uuid.New().String()},
		&schemas.PlayerQueuedPayload{
			PlayerId: uuid.New().String(),
			GameId:   uuid.New().String(),
			Region:   "us-east",
		})
	assert.NoError(t, err)
	mockAddAndFind.AssertNotCalled(t, "Execute", mock.Anything)
	mockEventPublisher.AssertNotCalled(t, "PublishMatchCreatedProto", mock.Anything, mock.Anything)
}

func TestHandlePlayerLeftQueueProto_SkippedWhenDualWriteDisabled(t *testing.T) {
	mockAddAndFind := &MockAddAndFindNextPairUseCase{}
	mockEventPublisher := &MockEventPublisher{}
	mockRegionReader := &mocks.MockPortRegionReader{}
	mockPoolReader := &mocks.MockPoolReader{}

	consumer := usecases.NewMatchmakingEventConsumer(
		mockAddAndFind, mockEventPublisher, mockRegionReader, mockPoolReader, nil, nil, nil,
	)
	consumer.SetDualWrite(config.DualWriteConfig{
		Enabled:       false,
		FailurePolicy: config.DualWriteFailLegacyOnlyAndAlert,
	})

	err := consumer.HandlePlayerLeftQueueProto(context.Background(),
		&schemas.EventEnvelope{Id: "e2", ResourceOwnerId: uuid.New().String()},
		&schemas.PlayerLeftQueuePayload{
			PlayerId: uuid.New().String(),
			GameId:   uuid.New().String(),
			Region:   "us-east",
		})
	assert.NoError(t, err)
	mockPoolReader.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
}

func TestDualWriteEnabled_DefaultTrue(t *testing.T) {
	consumer := usecases.NewMatchmakingEventConsumer(nil, nil, nil, nil, nil, nil, nil)
	assert.True(t, consumer.DualWriteEnabled())
}
