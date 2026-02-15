package kafka

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/leet-gaming/match-making-api/pkg/infra/events/schemas"
)

func TestPlayerQueuedConsumer_handleMessage(t *testing.T) {

	t.Run("Valid PlayerQueued event is dispatched to handler", func(t *testing.T) {
		handlerCalled := false
		var receivedEnvelope *schemas.EventEnvelope
		var receivedPayload *schemas.PlayerQueuedPayload

		handler := func(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.PlayerQueuedPayload) error {
			handlerCalled = true
			receivedEnvelope = envelope
			receivedPayload = payload
			return nil
		}

		pqc := &PlayerQueuedConsumer{handler: handler}

		playerID := uuid.New().String()
		gameID := uuid.New().String()
		tenantID := uuid.New().String()
		clientID := uuid.New().String()
		resourceOwnerID := uuid.New().String()

		event := &schemas.MatchmakingEvent{
			Envelope: &schemas.EventEnvelope{
				Id:              uuid.New().String(),
				Type:            schemas.EventTypePlayerQueued,
				Source:          "replay-api",
				Specversion:     schemas.CloudEventsSpecVersion,
				Time:            timestamppb.Now(),
				Subject:         playerID,
				ResourceOwnerId: resourceOwnerID,
				DataschemaVersion: schemas.SchemaVersionV1,
			},
			Data: &schemas.MatchmakingEvent_PlayerQueued{
				PlayerQueued: &schemas.PlayerQueuedPayload{
					PlayerId: playerID,
					GameId:   gameID,
					Region:   "us-east-1",
					TenantId: tenantID,
					ClientId: clientID,
				},
			},
		}

		value, err := protojson.Marshal(event)
		assert.NoError(t, err)

		msg := &kafkago.Message{
			Topic: TopicMatchmakingCommands,
			Value: value,
		}

		err = pqc.handleMessage(context.Background(), msg)

		assert.NoError(t, err)
		assert.True(t, handlerCalled, "handler should have been called")
		assert.Equal(t, resourceOwnerID, receivedEnvelope.GetResourceOwnerId())
		assert.Equal(t, playerID, receivedPayload.GetPlayerId())
		assert.Equal(t, gameID, receivedPayload.GetGameId())
		assert.Equal(t, "us-east-1", receivedPayload.GetRegion())
	})

	t.Run("Malformed message is skipped (not retried)", func(t *testing.T) {
		handlerCalled := false
		handler := func(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.PlayerQueuedPayload) error {
			handlerCalled = true
			return nil
		}

		pqc := &PlayerQueuedConsumer{handler: handler}

		msg := &kafkago.Message{
			Topic: TopicMatchmakingCommands,
			Value: []byte("this is not valid protojson"),
		}

		err := pqc.handleMessage(context.Background(), msg)

		assert.NoError(t, err, "malformed messages should return nil to skip")
		assert.False(t, handlerCalled, "handler should NOT be called for malformed messages")
	})

	t.Run("Event with nil envelope is skipped", func(t *testing.T) {
		handlerCalled := false
		handler := func(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.PlayerQueuedPayload) error {
			handlerCalled = true
			return nil
		}

		pqc := &PlayerQueuedConsumer{handler: handler}

		// Serialize a MatchmakingEvent with no envelope
		event := &schemas.MatchmakingEvent{}
		value, err := protojson.Marshal(event)
		assert.NoError(t, err)

		msg := &kafkago.Message{
			Topic: TopicMatchmakingCommands,
			Value: value,
		}

		err = pqc.handleMessage(context.Background(), msg)

		assert.NoError(t, err, "events with nil envelope should be skipped")
		assert.False(t, handlerCalled)
	})

	t.Run("Event with missing resource_owner_id is skipped", func(t *testing.T) {
		handlerCalled := false
		handler := func(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.PlayerQueuedPayload) error {
			handlerCalled = true
			return nil
		}

		pqc := &PlayerQueuedConsumer{handler: handler}

		event := &schemas.MatchmakingEvent{
			Envelope: &schemas.EventEnvelope{
				Id:              uuid.New().String(),
				Type:            schemas.EventTypePlayerQueued,
				Source:          "replay-api",
				ResourceOwnerId: "", // empty
			},
			Data: &schemas.MatchmakingEvent_PlayerQueued{
				PlayerQueued: &schemas.PlayerQueuedPayload{
					PlayerId: uuid.New().String(),
					GameId:   uuid.New().String(),
					Region:   "us-east-1",
					TenantId: uuid.New().String(),
					ClientId: uuid.New().String(),
				},
			},
		}

		value, err := protojson.Marshal(event)
		assert.NoError(t, err)

		msg := &kafkago.Message{
			Topic: TopicMatchmakingCommands,
			Value: value,
		}

		err = pqc.handleMessage(context.Background(), msg)

		assert.NoError(t, err, "events failing ownership validation should be skipped")
		assert.False(t, handlerCalled, "handler should NOT be called for invalid ownership")
	})

	t.Run("Event with missing tenant_id is skipped", func(t *testing.T) {
		handlerCalled := false
		handler := func(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.PlayerQueuedPayload) error {
			handlerCalled = true
			return nil
		}

		pqc := &PlayerQueuedConsumer{handler: handler}

		event := &schemas.MatchmakingEvent{
			Envelope: &schemas.EventEnvelope{
				Id:              uuid.New().String(),
				Type:            schemas.EventTypePlayerQueued,
				Source:          "replay-api",
				ResourceOwnerId: uuid.New().String(),
			},
			Data: &schemas.MatchmakingEvent_PlayerQueued{
				PlayerQueued: &schemas.PlayerQueuedPayload{
					PlayerId: uuid.New().String(),
					GameId:   uuid.New().String(),
					Region:   "us-east-1",
					TenantId: "", // empty
					ClientId: uuid.New().String(),
				},
			},
		}

		value, err := protojson.Marshal(event)
		assert.NoError(t, err)

		msg := &kafkago.Message{
			Topic: TopicMatchmakingCommands,
			Value: value,
		}

		err = pqc.handleMessage(context.Background(), msg)

		assert.NoError(t, err, "events failing ownership validation should be skipped")
		assert.False(t, handlerCalled)
	})

	t.Run("Event with missing player_id is skipped", func(t *testing.T) {
		handlerCalled := false
		handler := func(ctx context.Context, envelope *schemas.EventEnvelope, payload *schemas.PlayerQueuedPayload) error {
			handlerCalled = true
			return nil
		}

		pqc := &PlayerQueuedConsumer{handler: handler}

		event := &schemas.MatchmakingEvent{
			Envelope: &schemas.EventEnvelope{
				Id:              uuid.New().String(),
				Type:            schemas.EventTypePlayerQueued,
				Source:          "replay-api",
				ResourceOwnerId: uuid.New().String(),
			},
			Data: &schemas.MatchmakingEvent_PlayerQueued{
				PlayerQueued: &schemas.PlayerQueuedPayload{
					PlayerId: "", // empty
					GameId:   uuid.New().String(),
					Region:   "us-east-1",
					TenantId: uuid.New().String(),
					ClientId: uuid.New().String(),
				},
			},
		}

		value, err := protojson.Marshal(event)
		assert.NoError(t, err)

		msg := &kafkago.Message{
			Topic: TopicMatchmakingCommands,
			Value: value,
		}

		err = pqc.handleMessage(context.Background(), msg)

		assert.NoError(t, err, "events failing ownership validation should be skipped")
		assert.False(t, handlerCalled)
	})
}

func TestValidateResourceOwnership(t *testing.T) {
	t.Run("Valid ownership", func(t *testing.T) {
		envelope := &schemas.EventEnvelope{ResourceOwnerId: uuid.New().String()}
		payload := &schemas.PlayerQueuedPayload{
			PlayerId: uuid.New().String(),
			TenantId: uuid.New().String(),
			ClientId: uuid.New().String(),
		}
		err := validateResourceOwnership(envelope, payload)
		assert.NoError(t, err)
	})

	t.Run("Empty resource_owner_id", func(t *testing.T) {
		envelope := &schemas.EventEnvelope{ResourceOwnerId: ""}
		payload := &schemas.PlayerQueuedPayload{
			PlayerId: uuid.New().String(),
			TenantId: uuid.New().String(),
			ClientId: uuid.New().String(),
		}
		err := validateResourceOwnership(envelope, payload)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrResourceOwnershipInvalid)
	})

	t.Run("Empty tenant_id", func(t *testing.T) {
		envelope := &schemas.EventEnvelope{ResourceOwnerId: uuid.New().String()}
		payload := &schemas.PlayerQueuedPayload{
			PlayerId: uuid.New().String(),
			TenantId: "",
			ClientId: uuid.New().String(),
		}
		err := validateResourceOwnership(envelope, payload)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrResourceOwnershipInvalid)
	})

	t.Run("Empty client_id", func(t *testing.T) {
		envelope := &schemas.EventEnvelope{ResourceOwnerId: uuid.New().String()}
		payload := &schemas.PlayerQueuedPayload{
			PlayerId: uuid.New().String(),
			TenantId: uuid.New().String(),
			ClientId: "",
		}
		err := validateResourceOwnership(envelope, payload)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrResourceOwnershipInvalid)
	})

	t.Run("Empty player_id", func(t *testing.T) {
		envelope := &schemas.EventEnvelope{ResourceOwnerId: uuid.New().String()}
		payload := &schemas.PlayerQueuedPayload{
			PlayerId: "",
			TenantId: uuid.New().String(),
			ClientId: uuid.New().String(),
		}
		err := validateResourceOwnership(envelope, payload)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrResourceOwnershipInvalid)
	})
}
