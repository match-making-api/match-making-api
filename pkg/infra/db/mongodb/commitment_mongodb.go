package mongodb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const commitmentCollectionName = "commitments"

// CommitmentRepository implements CommitmentWriter and CommitmentReader ports
type CommitmentRepository struct {
	collection *mongo.Collection
}

// NewCommitmentRepository creates a new commitment repository with indexes
func NewCommitmentRepository(client *mongo.Client, dbName string) *CommitmentRepository {
	collection := client.Database(dbName).Collection(commitmentCollectionName)
	repo := &CommitmentRepository{collection: collection}
	repo.ensureIndexes(context.Background())
	return repo
}

func (r *CommitmentRepository) ensureIndexes(ctx context.Context) {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "lobby_id", Value: 1},
				{Key: "player_id", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_commitment_lobby_player_unique"),
		},
		{
			Keys:    bson.D{{Key: "lobby_id", Value: 1}},
			Options: options.Index().SetName("idx_commitment_lobby_id"),
		},
		{
			Keys:    bson.D{{Key: "player_id", Value: 1}},
			Options: options.Index().SetName("idx_commitment_player_id"),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "expires_at", Value: 1},
			},
			Options: options.Index().SetName("idx_commitment_status_expires"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("Failed to create commitment indexes", "error", err)
	}
	EnsureResourceOwnershipIndexes(ctx, r.collection, false)
}

// Save persists a commitment (upsert)
func (r *CommitmentRepository) Save(ctx context.Context, commitment *pairing_entities.Commitment) (*pairing_entities.Commitment, error) {
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": commitment.ID}
	update := bson.M{"$set": commitment}
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save commitment: %w", err)
	}
	return commitment, nil
}

// SaveBatch persists multiple commitments in a single bulk write
func (r *CommitmentRepository) SaveBatch(ctx context.Context, commitments []*pairing_entities.Commitment) ([]*pairing_entities.Commitment, error) {
	if len(commitments) == 0 {
		return commitments, nil
	}

	var models []mongo.WriteModel
	for _, c := range commitments {
		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": c.ID}).
			SetUpdate(bson.M{"$set": c}).
			SetUpsert(true)
		models = append(models, model)
	}

	_, err := r.collection.BulkWrite(ctx, models)
	if err != nil {
		return nil, fmt.Errorf("failed to save batch commitments: %w", err)
	}
	return commitments, nil
}

// Update persists changes to an existing commitment
func (r *CommitmentRepository) Update(ctx context.Context, commitment *pairing_entities.Commitment) (*pairing_entities.Commitment, error) {
	commitment.UpdatedAt = time.Now()
	filter := bson.M{"_id": commitment.ID}
	update := bson.M{"$set": commitment}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update commitment: %w", err)
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("commitment not found: %s", commitment.ID)
	}
	return commitment, nil
}

// GetByID retrieves a commitment by its ID
func (r *CommitmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*pairing_entities.Commitment, error) {
	var commitment pairing_entities.Commitment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&commitment)
	if err != nil {
		return nil, fmt.Errorf("commitment not found: %w", err)
	}
	return &commitment, nil
}

// FindByLobbyID retrieves all commitments for a lobby
func (r *CommitmentRepository) FindByLobbyID(ctx context.Context, lobbyID uuid.UUID) ([]*pairing_entities.Commitment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"lobby_id": lobbyID})
	if err != nil {
		return nil, fmt.Errorf("failed to find commitments by lobby: %w", err)
	}
	defer cursor.Close(ctx)

	var commitments []*pairing_entities.Commitment
	if err := cursor.All(ctx, &commitments); err != nil {
		return nil, fmt.Errorf("failed to decode commitments: %w", err)
	}
	return commitments, nil
}

// FindByPlayerAndLobby retrieves a player's commitment for a specific lobby
func (r *CommitmentRepository) FindByPlayerAndLobby(ctx context.Context, playerID, lobbyID uuid.UUID) (*pairing_entities.Commitment, error) {
	var commitment pairing_entities.Commitment
	err := r.collection.FindOne(ctx, bson.M{
		"player_id": playerID,
		"lobby_id":  lobbyID,
	}).Decode(&commitment)
	if err != nil {
		return nil, fmt.Errorf("commitment not found for player %s in lobby %s: %w", playerID, lobbyID, err)
	}
	return &commitment, nil
}

// FindPendingExpired retrieves all pending commitments that have expired
func (r *CommitmentRepository) FindPendingExpired(ctx context.Context) ([]*pairing_entities.Commitment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"status":     pairing_entities.CommitmentStatusPending,
		"expires_at": bson.M{"$lte": time.Now()},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find expired commitments: %w", err)
	}
	defer cursor.Close(ctx)

	var commitments []*pairing_entities.Commitment
	if err := cursor.All(ctx, &commitments); err != nil {
		return nil, fmt.Errorf("failed to decode expired commitments: %w", err)
	}
	return commitments, nil
}
