package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/metrics"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/templates"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// ConfirmReadinessUseCase handles a player confirming their readiness in a lobby.
// On confirmation, it checks if all players are ready and triggers match start.
type ConfirmReadinessUseCase struct {
	CommitmentWriter pairing_out.CommitmentWriter
	CommitmentReader pairing_out.CommitmentReader
	EventPublisher   EventPublisherInterface
	NotificationBatch *SendBatchNotificationUseCase
}

// ConfirmReadinessPayload contains the data for a player confirming readiness.
type ConfirmReadinessPayload struct {
	LobbyID  uuid.UUID
	PlayerID uuid.UUID
	Channel  pairing_entities.NotificationChannel // Which channel the confirmation came from
}

// ConfirmReadinessResult contains the result of a confirmation.
type ConfirmReadinessResult struct {
	Commitment *pairing_entities.Commitment
	Summary    *pairing_entities.LobbyCommitmentSummary
	AllReady   bool
}

// Execute processes a player's readiness confirmation.
func (uc *ConfirmReadinessUseCase) Execute(ctx context.Context, payload ConfirmReadinessPayload) (*ConfirmReadinessResult, error) {
	// Find the player's commitment for this lobby
	commitment, err := uc.CommitmentReader.FindByPlayerAndLobby(ctx, payload.PlayerID, payload.LobbyID)
	if err != nil {
		return nil, fmt.Errorf("commitment not found for player %s in lobby %s: %w", payload.PlayerID, payload.LobbyID, err)
	}

	// Confirm the commitment
	if err := commitment.Confirm(payload.Channel); err != nil {
		return nil, fmt.Errorf("failed to confirm commitment: %w", err)
	}

	// Persist the updated commitment
	updatedCommitment, err := uc.CommitmentWriter.Update(ctx, commitment)
	if err != nil {
		return nil, fmt.Errorf("failed to update commitment: %w", err)
	}

	// Fetch all commitments for the lobby to compute summary
	allCommitments, err := uc.CommitmentReader.FindByLobbyID(ctx, payload.LobbyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lobby commitments: %w", err)
	}

	summary := pairing_entities.NewLobbyCommitmentSummary(payload.LobbyID, allCommitments)
	allReady := summary.AllConfirmed()

	// Record metrics
	latency := time.Since(commitment.CreatedAt)
	metrics.Global.RecordReadinessConfirmed(payload.LobbyID.String(), payload.PlayerID.String(), latency)
	if allReady {
		metrics.Global.RecordAllPlayersReady(payload.LobbyID.String())
	}

	// Publish READINESS_CONFIRMED event
	if uc.EventPublisher != nil {
		event := &kafka.ReadyCheckEvent{
			LobbyID:   payload.LobbyID,
			PlayerID:  &payload.PlayerID,
			EventType: kafka.EventTypeReadinessConfirmed,
			Summary:   summary,
		}
		if err := uc.EventPublisher.PublishReadyCheckEvent(ctx, event); err != nil {
			slog.ErrorContext(ctx, "failed to publish READINESS_CONFIRMED event", "error", err)
		}

		// If all players are ready, publish ALL_PLAYERS_READY event
		if allReady {
			allReadyEvent := &kafka.ReadyCheckEvent{
				LobbyID:   payload.LobbyID,
				EventType: kafka.EventTypeAllPlayersReady,
				Summary:   summary,
			}
			if err := uc.EventPublisher.PublishReadyCheckEvent(ctx, allReadyEvent); err != nil {
				slog.ErrorContext(ctx, "failed to publish ALL_PLAYERS_READY event", "error", err)
			}
		}
	}

	// Send notification to other players about this confirmation
	if uc.NotificationBatch != nil {
		resourceOwner := common.GetResourceOwner(ctx)
		_ = resourceOwner // used implicitly by batch use case

		otherPlayerPayloads := make([]SendNotificationPayload, 0)
		for _, c := range allCommitments {
			if c.PlayerID != payload.PlayerID {
				// TODO: resolve player language from preferences
				lang := "en"
				tmpl := templates.Render(pairing_entities.NotificationTypeReadinessConfirmed, lang, map[string]string{
					"player_name": payload.PlayerID.String(), // TODO: resolve display name
					"confirmed":   strconv.Itoa(summary.ConfirmedCount),
					"total":       strconv.Itoa(summary.TotalPlayers),
				})

				otherPlayerPayloads = append(otherPlayerPayloads, SendNotificationPayload{
					UserID:  c.PlayerID,
					Channel: pairing_entities.NotificationChannelInApp,
					Type:    pairing_entities.NotificationTypeReadinessConfirmed,
					Title:   tmpl.Title,
					Message: tmpl.Message,
					Metadata: map[string]interface{}{
						"lobby_id":        payload.LobbyID.String(),
						"confirmed_count": summary.ConfirmedCount,
						"total_players":   summary.TotalPlayers,
					},
				})
			}
		}

		if len(otherPlayerPayloads) > 0 {
			batchPayload := SendBatchNotificationPayload{Notifications: otherPlayerPayloads}
			if _, err := uc.NotificationBatch.Execute(ctx, batchPayload); err != nil {
				slog.ErrorContext(ctx, "failed to send readiness update notifications", "error", err)
			}
		}

		// If all ready, notify everyone
		if allReady {
			allReadyPayloads := make([]SendNotificationPayload, 0, len(allCommitments))
			for _, c := range allCommitments {
				// TODO: resolve player language from preferences
				lang := "en"
				tmpl := templates.Render(pairing_entities.NotificationTypeAllPlayersReady, lang, map[string]string{
					"game_name": "match", // TODO: resolve game name from lobby metadata
				})

				allReadyPayloads = append(allReadyPayloads, SendNotificationPayload{
					UserID:  c.PlayerID,
					Channel: pairing_entities.NotificationChannelInApp,
					Type:    pairing_entities.NotificationTypeAllPlayersReady,
					Title:   tmpl.Title,
					Message: tmpl.Message,
					Metadata: map[string]interface{}{
						"lobby_id": payload.LobbyID.String(),
					},
				})
			}
			batchPayload := SendBatchNotificationPayload{Notifications: allReadyPayloads}
			if _, err := uc.NotificationBatch.Execute(ctx, batchPayload); err != nil {
				slog.ErrorContext(ctx, "failed to send all-ready notifications", "error", err)
			}
		}
	}

	slog.InfoContext(ctx, "Player confirmed readiness",
		"lobby_id", payload.LobbyID,
		"player_id", payload.PlayerID,
		"confirmed_count", summary.ConfirmedCount,
		"total_players", summary.TotalPlayers,
		"all_ready", allReady)

	return &ConfirmReadinessResult{
		Commitment: updatedCommitment,
		Summary:    summary,
		AllReady:   allReady,
	}, nil
}
