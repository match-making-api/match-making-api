package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/metrics"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/templates"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

const DefaultReadyCheckTimeoutSeconds = 60

// StartReadyCheckUseCase initiates a readiness confirmation round for all players in a lobby.
// It creates individual Commitment entities, publishes a Kafka event, and fans out notifications.
type StartReadyCheckUseCase struct {
	CommitmentWriter  pairing_out.CommitmentWriter
	EventPublisher    EventPublisherInterface
	NotificationBatch *SendBatchNotificationUseCase
}

// StartReadyCheckPayload contains the data needed to start a ready check.
type StartReadyCheckPayload struct {
	LobbyID        uuid.UUID
	PlayerIDs      []uuid.UUID
	GameName       string
	Tier           string
	PrizePool      string
	TimeoutSeconds int
	Metadata       map[string]interface{}
}

// Execute starts the ready check by creating commitments and notifying all players.
func (uc *StartReadyCheckUseCase) Execute(ctx context.Context, payload StartReadyCheckPayload) (*pairing_entities.LobbyCommitmentSummary, error) {
	resourceOwner := common.GetResourceOwner(ctx)

	timeoutSeconds := payload.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = DefaultReadyCheckTimeoutSeconds
	}

	if len(payload.PlayerIDs) < 2 {
		return nil, fmt.Errorf("ready check requires at least 2 players, got %d", len(payload.PlayerIDs))
	}

	// Create one commitment per player
	commitments := make([]*pairing_entities.Commitment, 0, len(payload.PlayerIDs))
	for _, playerID := range payload.PlayerIDs {
		metadata := map[string]interface{}{
			"game_name":  payload.GameName,
			"tier":       payload.Tier,
			"prize_pool": payload.PrizePool,
		}
		// Merge additional metadata
		for k, v := range payload.Metadata {
			metadata[k] = v
		}

		commitment := pairing_entities.NewCommitment(
			resourceOwner,
			payload.LobbyID,
			playerID,
			timeoutSeconds,
			metadata,
		)
		commitments = append(commitments, commitment)
	}

	// Persist all commitments
	savedCommitments, err := uc.CommitmentWriter.SaveBatch(ctx, commitments)
	if err != nil {
		return nil, fmt.Errorf("failed to save commitments: %w", err)
	}

	// Build summary
	summary := pairing_entities.NewLobbyCommitmentSummary(payload.LobbyID, savedCommitments)

	// Record metric
	metrics.Global.RecordReadyCheckStarted(payload.LobbyID.String(), len(payload.PlayerIDs))

	// Publish READY_CHECK_STARTED event to Kafka
	if uc.EventPublisher != nil {
		readyCheckEvent := &kafka.ReadyCheckEvent{
			LobbyID:   payload.LobbyID,
			EventType: kafka.EventTypeReadyCheckStarted,
			PlayerIDs: payload.PlayerIDs,
			Summary:   summary,
		}
		if err := uc.EventPublisher.PublishReadyCheckEvent(ctx, readyCheckEvent); err != nil {
			slog.ErrorContext(ctx, "failed to publish READY_CHECK_STARTED event", "error", err, "lobby_id", payload.LobbyID)
			// Non-blocking: continue even if event publish fails
		}
	}

	// Fan out notifications to all players
	if uc.NotificationBatch != nil {
		notifPayloads := make([]SendNotificationPayload, 0, len(payload.PlayerIDs))
		for _, playerID := range payload.PlayerIDs {
			// TODO: resolve player language from their preferences; default to "en"
			lang := "en"
			tmpl := templates.Render(pairing_entities.NotificationTypeReadyCheck, lang, map[string]string{
				"game_name": payload.GameName,
				"timeout":   strconv.Itoa(timeoutSeconds),
			})

			notifPayloads = append(notifPayloads, SendNotificationPayload{
				UserID:  playerID,
				Channel: pairing_entities.NotificationChannelInApp,
				Type:    pairing_entities.NotificationTypeReadyCheck,
				Title:   tmpl.Title,
				Message: tmpl.Message,
				Metadata: map[string]interface{}{
					"lobby_id":        payload.LobbyID.String(),
					"game_name":       payload.GameName,
					"tier":            payload.Tier,
					"prize_pool":      payload.PrizePool,
					"timeout_seconds": timeoutSeconds,
				},
			})
		}

		batchPayload := SendBatchNotificationPayload{Notifications: notifPayloads}
		if _, err := uc.NotificationBatch.Execute(ctx, batchPayload); err != nil {
			slog.ErrorContext(ctx, "failed to send ready check notifications", "error", err, "lobby_id", payload.LobbyID)
		}
	}

	slog.InfoContext(ctx, "Ready check started",
		"lobby_id", payload.LobbyID,
		"player_count", len(payload.PlayerIDs),
		"timeout_seconds", timeoutSeconds)

	return summary, nil
}
