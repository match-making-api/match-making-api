package usecases

import (
	"context"
	"log/slog"

	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
)

// PushNotificationSender handles web and mobile push notifications via FCM.
// When a real FCM adapter is configured, it looks up user push tokens and
// sends via the Firebase Admin SDK. Until then, it acts as a stub.
type PushNotificationSender struct {
	pushTokenReader pairing_out.PushTokenReader
	// fcmClient would hold the Firebase messaging client when configured
	// fcmClient *messaging.Client
}

// NewPushNotificationSender creates a push notification sender.
// Pass nil for pushTokenReader to create a stub sender.
func NewPushNotificationSender(pushTokenReader pairing_out.PushTokenReader) NotificationSender {
	return &PushNotificationSender{
		pushTokenReader: pushTokenReader,
	}
}

func (s *PushNotificationSender) GetChannel() pairing_entities.NotificationChannel {
	return pairing_entities.NotificationChannelPush
}

func (s *PushNotificationSender) IsAvailable(ctx context.Context) bool {
	// Push notifications are available when push token reader is configured
	// In production, also check if FCM client is initialized
	return s.pushTokenReader != nil
}

func (s *PushNotificationSender) Send(ctx context.Context, notification *pairing_entities.Notification) error {
	if s.pushTokenReader == nil {
		slog.InfoContext(ctx, "push notification skipped (no push adapter configured)",
			"notification_type", notification.Type,
			"user_id", notification.UserID,
		)
		notification.MarkAsSent()
		return nil
	}

	// Fetch active push tokens for the user
	tokens, err := s.pushTokenReader.FindActiveByUserID(ctx, notification.UserID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch push tokens",
			"user_id", notification.UserID,
			"error", err)
		return err
	}

	if len(tokens) == 0 {
		slog.InfoContext(ctx, "no active push tokens found for user",
			"user_id", notification.UserID)
		notification.MarkAsSent() // Not a failure — user simply has no devices registered
		return nil
	}

	// TODO: When FCM is configured, send to all active tokens:
	// for _, token := range tokens {
	//     msg := &messaging.Message{
	//         Token: token.Token,
	//         Notification: &messaging.Notification{
	//             Title: notification.Title,
	//             Body:  notification.Message,
	//         },
	//         Data: flattenMetadata(notification.Metadata),
	//         WebpushConfig: &messaging.WebpushConfig{
	//             FCMOptions: &messaging.WebpushFCMOptions{
	//                 Link: buildDeepLink(notification),
	//             },
	//         },
	//     }
	//     _, err := s.fcmClient.Send(ctx, msg)
	//     if err != nil {
	//         slog.ErrorContext(ctx, "FCM send failed", "token", token.Token, "error", err)
	//         // Deactivate invalid tokens (messaging.IsRegistrationTokenNotRegistered)
	//     }
	//     token.MarkUsed()
	// }

	slog.InfoContext(ctx, "push notification sent (stub mode)",
		"user_id", notification.UserID,
		"token_count", len(tokens),
		"notification_type", notification.Type)

	notification.MarkAsSent()
	return nil
}
