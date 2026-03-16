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

const gameConnectionInfoCollectionName = "game_connection_info"

// GameConnectionInfoRepository implements GameConnectionInfoWriter and GameConnectionInfoReader ports
type GameConnectionInfoRepository struct {
	collection *mongo.Collection
}

// NewGameConnectionInfoRepository creates a new game connection info repository with indexes
func NewGameConnectionInfoRepository(client *mongo.Client, dbName string) *GameConnectionInfoRepository {
	collection := client.Database(dbName).Collection(gameConnectionInfoCollectionName)
	repo := &GameConnectionInfoRepository{collection: collection}
	repo.ensureIndexes(context.Background())
	return repo
}

func (r *GameConnectionInfoRepository) ensureIndexes(ctx context.Context) {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "lobby_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_gci_lobby_id_unique"),
		},
		{
			Keys:    bson.D{{Key: "match_id", Value: 1}},
			Options: options.Index().SetName("idx_gci_match_id"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("Failed to create game_connection_info indexes", "error", err)
	}
}

// Save persists game connection info (upsert)
func (r *GameConnectionInfoRepository) Save(ctx context.Context, info *pairing_entities.GameConnectionInfo) (*pairing_entities.GameConnectionInfo, error) {
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": info.ID}
	update := bson.M{"$set": info}
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save game connection info: %w", err)
	}
	return info, nil
}

// FindByLobbyID retrieves game connection info for a lobby
func (r *GameConnectionInfoRepository) FindByLobbyID(ctx context.Context, lobbyID uuid.UUID) (*pairing_entities.GameConnectionInfo, error) {
	var info pairing_entities.GameConnectionInfo
	err := r.collection.FindOne(ctx, bson.M{"lobby_id": lobbyID}).Decode(&info)
	if err != nil {
		return nil, fmt.Errorf("game connection info not found for lobby %s: %w", lobbyID, err)
	}
	return &info, nil
}

// FindByMatchID retrieves game connection info for a match
func (r *GameConnectionInfoRepository) FindByMatchID(ctx context.Context, matchID uuid.UUID) (*pairing_entities.GameConnectionInfo, error) {
	var info pairing_entities.GameConnectionInfo
	err := r.collection.FindOne(ctx, bson.M{"match_id": matchID}).Decode(&info)
	if err != nil {
		return nil, fmt.Errorf("game connection info not found for match %s: %w", matchID, err)
	}
	return &info, nil
}
