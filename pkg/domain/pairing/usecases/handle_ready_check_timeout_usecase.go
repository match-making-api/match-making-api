package usecases

import (
	"context"
	"log/slog"

	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// HandleReadyCheckTimeoutUseCase discovers expired commitments and processes lobby cancellation.
// This is called by the ReadyCheckTimeoutWorker on a periodic schedule.
type HandleReadyCheckTimeoutUseCase struct {
	CommitmentWriter  pairing_out.CommitmentWriter
	CommitmentReader  pairing_out.CommitmentReader
	EventPublisher    EventPublisherInterface
	NotificationBatch *SendBatchNotificationUseCase
}

// Execute finds all expired pending commitments, marks them timed out,
// and publishes timeout events for each affected lobby.
func (uc *HandleReadyCheckTimeoutUseCase) Execute(ctx context.Context) (int, error) {
	// Find all pending commitments that have expired
	expiredCommitments, err := uc.CommitmentReader.FindPendingExpired(ctx)
	if err != nil {
		return 0, err
	}

	if len(expiredCommitments) == 0 {
		return 0, nil
	}

	// Group by lobby for batch processing
	lobbyCommitments := make(map[string][]*pairing_entities.Commitment)
	for _, c := range expiredCommitments {
		key := c.LobbyID.String()
		lobbyCommitments[key] = append(lobbyCommitments[key], c)
	}

	processedCount := 0

	for _, commitments := range lobbyCommitments {
		lobbyID := commitments[0].LobbyID

		// Mark each expired commitment as timed out
		for _, c := range commitments {
			if err := c.TimeOut(); err != nil {
				slog.ErrorContext(ctx, "failed to time out commitment",
					"commitment_id", c.ID,
					"error", err)
				continue
			}

			if _, err := uc.CommitmentWriter.Update(ctx, c); err != nil {
				slog.ErrorContext(ctx, "failed to persist timed out commitment",
					"commitment_id", c.ID,
					"error", err)
				continue
			}
			processedCount++
		}

		// Fetch all commitments for this lobby to build a complete summary
		allCommitments, err := uc.CommitmentReader.FindByLobbyID(ctx, lobbyID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to fetch lobby commitments for timeout",
				"lobby_id", lobbyID,
				"error", err)
			continue
		}

		summary := pairing_entities.NewLobbyCommitmentSummary(lobbyID, allCommitments)

		// Publish READY_CHECK_TIMEOUT event
		if uc.EventPublisher != nil {
			event := &kafka.ReadyCheckEvent{
				LobbyID:   lobbyID,
				EventType: kafka.EventTypeReadyCheckTimeout,
				Summary:   summary,
			}
			if err := uc.EventPublisher.PublishReadyCheckEvent(ctx, event); err != nil {
				slog.ErrorContext(ctx, "failed to publish READY_CHECK_TIMEOUT event",
					"lobby_id", lobbyID,
					"error", err)
			}
		}

		// Notify all players in the lobby about the timeout
		if uc.NotificationBatch != nil {
			timeoutPayloads := make([]SendNotificationPayload, 0, len(allCommitments))
			for _, c := range allCommitments {
				timeoutPayloads = append(timeoutPayloads, SendNotificationPayload{
					UserID:  c.PlayerID,
					Channel: pairing_entities.NotificationChannelInApp,
					Type:    pairing_entities.NotificationTypeReadyCheckTimeout,
					Title:   "Ready Check Expired",
					Message: "Not all players confirmed in time. Returning to queue...",
					Metadata: map[string]interface{}{
						"lobby_id":      lobbyID.String(),
						"expired_count": summary.ExpiredCount,
					},
				})
			}

			batchPayload := SendBatchNotificationPayload{Notifications: timeoutPayloads}
			if _, err := uc.NotificationBatch.Execute(ctx, batchPayload); err != nil {
				slog.ErrorContext(ctx, "failed to send timeout notifications",
					"lobby_id", lobbyID,
					"error", err)
			}
		}

		slog.InfoContext(ctx, "Ready check timed out",
			"lobby_id", lobbyID,
			"expired_count", len(commitments),
			"total_players", summary.TotalPlayers)
	}

	return processedCount, nil
}
