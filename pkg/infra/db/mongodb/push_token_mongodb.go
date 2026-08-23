package mongodb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const pushTokenCollectionName = "push_tokens"

// PushTokenRepository implements PushTokenWriter and PushTokenReader ports
type PushTokenRepository struct {
	collection *mongo.Collection
}

// NewPushTokenRepository creates a new push token repository with indexes
func NewPushTokenRepository(client *mongo.Client, dbName string) *PushTokenRepository {
	collection := client.Database(dbName).Collection(pushTokenCollectionName)
	repo := &PushTokenRepository{collection: collection}
	repo.ensureIndexes(context.Background())
	return repo
}

func (r *PushTokenRepository) ensureIndexes(ctx context.Context) {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("idx_push_token_user_id"),
		},
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_push_token_unique"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("Failed to create push_token indexes", "error", err)
	}
	EnsureResourceOwnershipIndexes(ctx, r.collection, false)
}

// Save persists a push token (upsert)
func (r *PushTokenRepository) Save(ctx context.Context, token *pairing_entities.PushToken) (*pairing_entities.PushToken, error) {
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": token.ID}
	update := bson.M{"$set": token}
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save push token: %w", err)
	}
	return token, nil
}

// Deactivate marks a push token as inactive
func (r *PushTokenRepository) Deactivate(ctx context.Context, tokenID uuid.UUID) error {
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": tokenID},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	return err
}

// FindByUserID retrieves all active push tokens for a user
func (r *PushTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*pairing_entities.PushToken, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":   userID,
		"is_active": true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find push tokens for user %s: %w", userID, err)
	}
	defer cursor.Close(ctx)

	var tokens []*pairing_entities.PushToken
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, fmt.Errorf("failed to decode push tokens: %w", err)
	}
	return tokens, nil
}

// FindActiveByUserID retrieves all active push tokens for a user (alias for FindByUserID)
func (r *PushTokenRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*pairing_entities.PushToken, error) {
	return r.FindByUserID(ctx, userID)
}
