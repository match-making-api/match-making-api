package usecases

import (
	"context"

	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
)

// NotificationSender defines the interface for sending notifications through different channels
type NotificationSender interface {
	// Send sends a notification through the specified channel
	Send(ctx context.Context, notification *pairing_entities.Notification) error
	
	// GetChannel returns the channel this sender handles
	GetChannel() pairing_entities.NotificationChannel
	
	// IsAvailable checks if the sender is available/configured
	IsAvailable(ctx context.Context) bool
}

// InAppNotificationSender handles in-app notifications
type InAppNotificationSender struct{}

func NewInAppNotificationSender() NotificationSender {
	return &InAppNotificationSender{}
}

func (s *InAppNotificationSender) GetChannel() pairing_entities.NotificationChannel {
	return pairing_entities.NotificationChannelInApp
}

func (s *InAppNotificationSender) IsAvailable(ctx context.Context) bool {
	// In-app notifications are always available
	return true
}

func (s *InAppNotificationSender) Send(ctx context.Context, notification *pairing_entities.Notification) error {
	// In-app notifications are stored in the database and retrieved by the user
	// The actual delivery is handled by the frontend polling or websocket connection
	// This implementation just marks it as sent since it's stored in DB
	notification.MarkAsSent()
	return nil
}

// EmailNotificationSender handles email notifications
type EmailNotificationSender struct{}

func NewEmailNotificationSender() NotificationSender {
	return &EmailNotificationSender{}
}

func (s *EmailNotificationSender) GetChannel() pairing_entities.NotificationChannel {
	return pairing_entities.NotificationChannelEmail
}

func (s *EmailNotificationSender) IsAvailable(ctx context.Context) bool {
	// TODO: Check if email service is configured
	// For now, return true as placeholder
	return true
}

func (s *EmailNotificationSender) Send(ctx context.Context, notification *pairing_entities.Notification) error {
	// TODO: Implement actual email sending logic
	// This should integrate with an email service (SMTP, SendGrid, SES, etc.)
	// For now, this is a placeholder
	notification.MarkAsSent()
	return nil
}

// SMSNotificationSender handles SMS notifications
type SMSNotificationSender struct{}

func NewSMSNotificationSender() NotificationSender {
	return &SMSNotificationSender{}
}

func (s *SMSNotificationSender) GetChannel() pairing_entities.NotificationChannel {
	return pairing_entities.NotificationChannelSMS
}

func (s *SMSNotificationSender) IsAvailable(ctx context.Context) bool {
	// TODO: Check if SMS service is configured
	// For now, return false as SMS is optional
	return false
}

func (s *SMSNotificationSender) Send(ctx context.Context, notification *pairing_entities.Notification) error {
	// TODO: Implement actual SMS sending logic
	// This should integrate with an SMS service (Twilio, AWS SNS, etc.)
	// For now, this is a placeholder
	notification.MarkAsSent()
	return nil
}

// NotificationSenderFactory creates the appropriate sender for a channel
type NotificationSenderFactory struct {
	senders map[pairing_entities.NotificationChannel]NotificationSender
}

func NewNotificationSenderFactory() *NotificationSenderFactory {
	factory := &NotificationSenderFactory{
		senders: make(map[pairing_entities.NotificationChannel]NotificationSender),
	}
	
	// Register default senders
	factory.senders[pairing_entities.NotificationChannelInApp] = NewInAppNotificationSender()
	factory.senders[pairing_entities.NotificationChannelEmail] = NewEmailNotificationSender()
	factory.senders[pairing_entities.NotificationChannelSMS] = NewSMSNotificationSender()
	
	return factory
}

func (f *NotificationSenderFactory) GetSender(channel pairing_entities.NotificationChannel) NotificationSender {
	return f.senders[channel]
}

func (f *NotificationSenderFactory) RegisterSender(channel pairing_entities.NotificationChannel, sender NotificationSender) {
	f.senders[channel] = sender
}
