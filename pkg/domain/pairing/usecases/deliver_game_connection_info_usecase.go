package usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/metrics"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/templates"
	"github.com/leet-gaming/match-making-api/pkg/infra/kafka"
)

// DeliverGameConnectionInfoUseCase delivers game server connection details
// to all players after all readiness confirmations are received.
type DeliverGameConnectionInfoUseCase struct {
	GameConnectionInfoWriter pairing_out.GameConnectionInfoWriter
	CommitmentReader         pairing_out.CommitmentReader
	EventPublisher           EventPublisherInterface
	NotificationBatch        *SendBatchNotificationUseCase
}

// DeliverGameConnectionInfoPayload contains the game server connection details.
type DeliverGameConnectionInfoPayload struct {
	MatchID      uuid.UUID
	LobbyID      uuid.UUID
	GameID       string
	Region       string
	ServerURL    *string
	ServerIP     *string
	Port         *int
	Passcode     *string
	QRCodeData   *string
	Instructions string
	DeepLink     *string
}

// Execute saves the connection info and delivers it to all players via notifications.
func (uc *DeliverGameConnectionInfoUseCase) Execute(ctx context.Context, payload DeliverGameConnectionInfoPayload) (*pairing_entities.GameConnectionInfo, error) {
	resourceOwner := common.GetResourceOwner(ctx)

	// Create connection info entity
	connInfo := pairing_entities.NewGameConnectionInfo(
		resourceOwner,
		payload.MatchID,
		payload.LobbyID,
		payload.GameID,
		payload.Region,
		payload.Instructions,
	)

	// Apply optional fields
	if payload.ServerURL != nil || payload.ServerIP != nil {
		serverURL := ""
		serverIP := ""
		port := 0
		passcode := ""
		if payload.ServerURL != nil {
			serverURL = *payload.ServerURL
		}
		if payload.ServerIP != nil {
			serverIP = *payload.ServerIP
		}
		if payload.Port != nil {
			port = *payload.Port
		}
		if payload.Passcode != nil {
			passcode = *payload.Passcode
		}
		connInfo.WithServer(serverURL, serverIP, port, passcode)
	}

	if payload.QRCodeData != nil {
		connInfo.WithQRCode(*payload.QRCodeData)
	}

	if payload.DeepLink != nil {
		connInfo.WithDeepLink(*payload.DeepLink)
	}

	// Persist connection info
	savedConnInfo, err := uc.GameConnectionInfoWriter.Save(ctx, connInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to save game connection info: %w", err)
	}

	// Fetch all players for this lobby
	allCommitments, err := uc.CommitmentReader.FindByLobbyID(ctx, payload.LobbyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lobby commitments: %w", err)
	}

	// Build notification message with connection details
	serverAddress := payload.Instructions
	if payload.ServerURL != nil {
		serverAddress = *payload.ServerURL
	} else if payload.ServerIP != nil {
		serverAddress = *payload.ServerIP
	}

	// Build metadata for notification
	metadata := map[string]interface{}{
		"match_id":     payload.MatchID.String(),
		"lobby_id":     payload.LobbyID.String(),
		"game_id":      payload.GameID,
		"region":       payload.Region,
		"instructions": payload.Instructions,
	}
	if payload.ServerURL != nil {
		metadata["server_url"] = *payload.ServerURL
	}
	if payload.ServerIP != nil {
		metadata["server_ip"] = *payload.ServerIP
	}
	if payload.Port != nil {
		metadata["port"] = *payload.Port
	}
	if payload.Passcode != nil {
		metadata["passcode"] = *payload.Passcode
	}
	if payload.QRCodeData != nil {
		metadata["qr_code_data"] = *payload.QRCodeData
	}
	if payload.DeepLink != nil {
		metadata["deep_link"] = *payload.DeepLink
	}

	// Publish GAME_CONNECTION_DELIVERED event
	if uc.EventPublisher != nil {
		event := &kafka.ReadyCheckEvent{
			LobbyID:            payload.LobbyID,
			EventType:          kafka.EventTypeGameConnectionDelivered,
			GameConnectionInfo: savedConnInfo,
		}
		if err := uc.EventPublisher.PublishReadyCheckEvent(ctx, event); err != nil {
			slog.ErrorContext(ctx, "failed to publish GAME_CONNECTION_DELIVERED event", "error", err)
		}
	}

	// Deliver to all players via all channels
	if uc.NotificationBatch != nil {
		connPayloads := make([]SendNotificationPayload, 0, len(allCommitments))
		for _, c := range allCommitments {
			// TODO: resolve player language from preferences
			lang := "en"
			tmpl := templates.Render(pairing_entities.NotificationTypeGameConnectionInfo, lang, map[string]string{
				"game_name":      payload.GameID,
				"server_address": serverAddress,
			})

			connPayloads = append(connPayloads, SendNotificationPayload{
				UserID:   c.PlayerID,
				Channel:  pairing_entities.NotificationChannelInApp,
				Type:     pairing_entities.NotificationTypeGameConnectionInfo,
				Title:    tmpl.Title,
				Message:  tmpl.Message,
				Metadata: metadata,
			})
		}

		batchPayload := SendBatchNotificationPayload{Notifications: connPayloads}
		if _, err := uc.NotificationBatch.Execute(ctx, batchPayload); err != nil {
			slog.ErrorContext(ctx, "failed to send game connection notifications", "error", err)
		}
	}

	slog.InfoContext(ctx, "Game connection info delivered",
		"match_id", payload.MatchID,
		"lobby_id", payload.LobbyID,
		"player_count", len(allCommitments))

	// Record metric
	metrics.Global.RecordConnectionInfoDelivered(payload.LobbyID.String(), payload.MatchID.String(), len(allCommitments))

	return savedConnInfo, nil
}
