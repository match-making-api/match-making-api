package usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/metrics"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/templates"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// DeclineReadinessUseCase handles a player declining readiness, which cancels the lobby.
type DeclineReadinessUseCase struct {
	CommitmentWriter  pairing_out.CommitmentWriter
	CommitmentReader  pairing_out.CommitmentReader
	EventPublisher    EventPublisherInterface
	NotificationBatch *SendBatchNotificationUseCase
}

// DeclineReadinessPayload contains the data for a player declining readiness.
type DeclineReadinessPayload struct {
	LobbyID  uuid.UUID
	PlayerID uuid.UUID
}

// Execute processes a player's readiness decline and triggers lobby cancellation.
func (uc *DeclineReadinessUseCase) Execute(ctx context.Context, payload DeclineReadinessPayload) (*pairing_entities.LobbyCommitmentSummary, error) {
	// Find the player's commitment
	commitment, err := uc.CommitmentReader.FindByPlayerAndLobby(ctx, payload.PlayerID, payload.LobbyID)
	if err != nil {
		return nil, fmt.Errorf("commitment not found for player %s in lobby %s: %w", payload.PlayerID, payload.LobbyID, err)
	}

	// Decline the commitment
	if err := commitment.Decline(); err != nil {
		return nil, fmt.Errorf("failed to decline commitment: %w", err)
	}

	// Persist the updated commitment
	if _, err := uc.CommitmentWriter.Update(ctx, commitment); err != nil {
		return nil, fmt.Errorf("failed to update commitment: %w", err)
	}

	// Fetch all commitments for summary
	allCommitments, err := uc.CommitmentReader.FindByLobbyID(ctx, payload.LobbyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lobby commitments: %w", err)
	}

	summary := pairing_entities.NewLobbyCommitmentSummary(payload.LobbyID, allCommitments)

	// Record metric
	metrics.Global.RecordReadinessDeclined(payload.LobbyID.String(), payload.PlayerID.String())

	// Publish READINESS_DECLINED event
	if uc.EventPublisher != nil {
		event := &kafka.ReadyCheckEvent{
			LobbyID:   payload.LobbyID,
			PlayerID:  &payload.PlayerID,
			EventType: kafka.EventTypeReadinessDeclined,
			Summary:   summary,
		}
		if err := uc.EventPublisher.PublishReadyCheckEvent(ctx, event); err != nil {
			slog.ErrorContext(ctx, "failed to publish READINESS_DECLINED event", "error", err)
		}
	}

	// Notify other players that the match was cancelled due to decline
	if uc.NotificationBatch != nil {
		cancelPayloads := make([]SendNotificationPayload, 0)
		for _, c := range allCommitments {
			if c.PlayerID != payload.PlayerID {
				// TODO: resolve player language from preferences
				lang := "en"
				tmpl := templates.Render(pairing_entities.NotificationTypeReadinessDeclined, lang, map[string]string{
					"player_name": payload.PlayerID.String(), // TODO: resolve display name
					"game_name":   "match",                  // TODO: resolve from lobby metadata
				})

				cancelPayloads = append(cancelPayloads, SendNotificationPayload{
					UserID:  c.PlayerID,
					Channel: pairing_entities.NotificationChannelInApp,
					Type:    pairing_entities.NotificationTypeReadinessDeclined,
					Title:   tmpl.Title,
					Message: tmpl.Message,
					Metadata: map[string]interface{}{
						"lobby_id": payload.LobbyID.String(),
					},
				})
			}
		}

		if len(cancelPayloads) > 0 {
			batchPayload := SendBatchNotificationPayload{Notifications: cancelPayloads}
			if _, err := uc.NotificationBatch.Execute(ctx, batchPayload); err != nil {
				slog.ErrorContext(ctx, "failed to send decline notifications", "error", err)
			}
		}
	}

	slog.InfoContext(ctx, "Player declined readiness",
		"lobby_id", payload.LobbyID,
		"player_id", payload.PlayerID)

	return summary, nil
}
