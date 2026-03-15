package usecases

import (
	"context"
	"log/slog"

	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
)

// WhatsAppNotificationSender handles WhatsApp notifications via Twilio WhatsApp Business API.
// This is a stub implementation. When Twilio is configured, it will use pre-approved
// WhatsApp templates and fall back to SMS on delivery failure.
type WhatsAppNotificationSender struct {
	// twilioClient *twilio.RestClient
	// fromNumber string (WhatsApp-enabled Twilio number)
}

// NewWhatsAppNotificationSender creates a WhatsApp notification sender.
func NewWhatsAppNotificationSender() NotificationSender {
	return &WhatsAppNotificationSender{}
}

func (s *WhatsAppNotificationSender) GetChannel() pairing_entities.NotificationChannel {
	return pairing_entities.NotificationChannelWhatsApp
}

func (s *WhatsAppNotificationSender) IsAvailable(ctx context.Context) bool {
	// WhatsApp sending requires Twilio configuration
	// Returns false until a real Twilio adapter is implemented
	return false
}

func (s *WhatsAppNotificationSender) Send(ctx context.Context, notification *pairing_entities.Notification) error {
	// NoOp: logs and marks as sent. Replace with real Twilio WhatsApp adapter when configured.
	// Production implementation:
	// 1. Look up user's verified phone number from profile
	// 2. Build WhatsApp template message (pre-approved by Meta)
	// 3. Send via Twilio: client.Api.CreateMessage(params)
	// 4. Handle delivery receipts via webhook
	// 5. Fall back to SMS if WhatsApp delivery fails

	slog.InfoContext(ctx, "WhatsApp notification skipped (no WhatsApp adapter configured)",
		"notification_type", notification.Type,
		"user_id", notification.UserID,
	)
	notification.MarkAsSent()
	return nil
}
