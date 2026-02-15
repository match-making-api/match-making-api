package mongodb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const LobbyCollectionName = "lobbies"

// LobbyRepository handles lobby persistence
type LobbyRepository struct {
	collection *mongo.Collection
	dbName     string
}

// NewLobbyRepository creates a new lobby repository
func NewLobbyRepository(client *mongo.Client, dbName string) *LobbyRepository {
	collection := client.Database(dbName).Collection(LobbyCollectionName)
	repo := &LobbyRepository{
		collection: collection,
		dbName:     dbName,
	}
	
	// Create indexes
	repo.ensureIndexes()
	
	return repo
}

// ensureIndexes creates the necessary indexes for efficient querying
func (r *LobbyRepository) ensureIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		// Status + created_at for browsing open lobbies
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_status_created"),
		},
		// Game + region + status for filtered browsing
		{
			Keys: bson.D{
				{Key: "game_id", Value: 1},
				{Key: "region", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_game_region_status"),
		},
		// Visibility + featured for homepage
		{
			Keys: bson.D{
				{Key: "visibility", Value: 1},
				{Key: "is_featured", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_visibility_featured"),
		},
		// Creator ID for user's lobbies
		{
			Keys: bson.D{
				{Key: "creator_id", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_creator_status"),
		},
		// Player slots for finding lobbies a player is in
		{
			Keys: bson.D{
				{Key: "player_slots.player_id", Value: 1},
			},
			Options: options.Index().SetName("idx_player_id"),
		},
		// TTL index for auto-expiring lobbies
		{
			Keys: bson.D{
				{Key: "expires_at", Value: 1},
			},
			Options: options.Index().SetName("idx_expires_at").SetExpireAfterSeconds(0),
		},
		// Text index for search
		{
			Keys: bson.D{
				{Key: "name", Value: "text"},
				{Key: "description", Value: "text"},
				{Key: "tags", Value: "text"},
			},
			Options: options.Index().SetName("idx_text_search"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Warn("Failed to create lobby indexes", "error", err)
	} else {
		slog.Info("Lobby indexes ensured")
	}
}

// Create inserts a new lobby
func (r *LobbyRepository) Create(ctx context.Context, lobby *entities.Lobby) error {
	lobby.CreatedAt = time.Now()
	lobby.UpdatedAt = time.Now()
	
	_, err := r.collection.InsertOne(ctx, lobby)
	if err != nil {
		return fmt.Errorf("failed to create lobby: %w", err)
	}
	return nil
}

// GetByID retrieves a lobby by ID
func (r *LobbyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Lobby, error) {
	var lobby entities.Lobby
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&lobby)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}
	return &lobby, nil
}

// Update updates an existing lobby
func (r *LobbyRepository) Update(ctx context.Context, lobby *entities.Lobby) error {
	lobby.UpdatedAt = time.Now()
	
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": lobby.ID}, lobby)
	if err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}
	return nil
}

// Delete removes a lobby
func (r *LobbyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete lobby: %w", err)
	}
	return nil
}

// LobbySearchParams holds search criteria
type LobbySearchParams struct {
	GameID     string
	GameMode   string
	Region     string
	Status     string
	Visibility string
	Type       string
	CreatorID  *uuid.UUID
	PlayerID   *uuid.UUID
	Featured   *bool
	TextSearch string
	MinPlayers *int
	MaxPlayers *int
	Skip       int
	Limit      int
}

// Search finds lobbies based on search parameters
func (r *LobbyRepository) Search(ctx context.Context, params LobbySearchParams) ([]*entities.Lobby, int64, error) {
	filter := bson.M{}
	
	// Always exclude private lobbies from public searches unless explicitly searching for them
	if params.Visibility != string(entities.LobbyVisibilityPrivate) && params.CreatorID == nil && params.PlayerID == nil {
		filter["visibility"] = bson.M{"$ne": entities.LobbyVisibilityPrivate}
	}
	
	if params.GameID != "" {
		filter["game_id"] = params.GameID
	}
	if params.GameMode != "" {
		filter["game_mode"] = params.GameMode
	}
	if params.Region != "" {
		filter["region"] = params.Region
	}
	if params.Status != "" {
		filter["status"] = params.Status
	}
	if params.Visibility != "" {
		filter["visibility"] = params.Visibility
	}
	if params.Type != "" {
		filter["type"] = params.Type
	}
	if params.CreatorID != nil {
		filter["creator_id"] = *params.CreatorID
	}
	if params.PlayerID != nil {
		filter["player_slots.player_id"] = *params.PlayerID
	}
	if params.Featured != nil {
		filter["is_featured"] = *params.Featured
	}
	if params.TextSearch != "" {
		filter["$text"] = bson.M{"$search": params.TextSearch}
	}
	
	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count lobbies: %w", err)
	}
	
	// Set default pagination
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	
	opts := options.Find().
		SetSkip(int64(params.Skip)).
		SetLimit(int64(params.Limit)).
		SetSort(bson.D{
			{Key: "is_featured", Value: -1},
			{Key: "created_at", Value: -1},
		})
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search lobbies: %w", err)
	}
	defer cursor.Close(ctx)
	
	var lobbies []*entities.Lobby
	if err := cursor.All(ctx, &lobbies); err != nil {
		return nil, 0, fmt.Errorf("failed to decode lobbies: %w", err)
	}
	
	return lobbies, total, nil
}

// GetFeaturedLobbies retrieves featured lobbies for homepage
func (r *LobbyRepository) GetFeaturedLobbies(ctx context.Context, gameID string, limit int) ([]*entities.Lobby, error) {
	filter := bson.M{
		"status":      entities.LobbyStatusOpen,
		"visibility":  bson.M{"$in": []entities.LobbyVisibility{entities.LobbyVisibilityPublic, entities.LobbyVisibilityMatchmaking}},
	}
	
	if gameID != "" {
		filter["game_id"] = gameID
	}
	
	if limit <= 0 {
		limit = 8
	}
	
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{
			{Key: "is_featured", Value: -1},
			{Key: "player_slots", Value: -1}, // More players = more active
			{Key: "created_at", Value: -1},
		})
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured lobbies: %w", err)
	}
	defer cursor.Close(ctx)
	
	var lobbies []*entities.Lobby
	if err := cursor.All(ctx, &lobbies); err != nil {
		return nil, fmt.Errorf("failed to decode featured lobbies: %w", err)
	}
	
	return lobbies, nil
}

// GetLobbyStats returns statistics about active lobbies
func (r *LobbyRepository) GetLobbyStats(ctx context.Context, gameID string) (*common.LobbyStats, error) {
	matchStage := bson.D{{Key: "$match", Value: bson.M{"status": entities.LobbyStatusOpen}}}
	if gameID != "" {
		matchStage = bson.D{{Key: "$match", Value: bson.M{"status": entities.LobbyStatusOpen, "game_id": gameID}}}
	}
	
	pipeline := mongo.Pipeline{
		matchStage,
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.M{"$sum": 1}},
			{Key: "total_players", Value: bson.M{"$sum": bson.M{"$size": "$player_slots"}}},
			{Key: "by_game", Value: bson.M{"$push": "$game_id"}},
			{Key: "by_region", Value: bson.M{"$push": "$region"}},
			{Key: "by_mode", Value: bson.M{"$push": "$game_mode"}},
		}}},
	}
	
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby stats: %w", err)
	}
	defer cursor.Close(ctx)
	
	stats := &common.LobbyStats{
		ByGame:   make(map[string]int),
		ByRegion: make(map[string]int),
		ByMode:   make(map[string]int),
	}
	
	if cursor.Next(ctx) {
		var result struct {
			Total        int      `bson:"total"`
			TotalPlayers int      `bson:"total_players"`
			ByGame       []string `bson:"by_game"`
			ByRegion     []string `bson:"by_region"`
			ByMode       []string `bson:"by_mode"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		
		stats.TotalActiveLobbies = result.Total
		stats.TotalPlayersWaiting = result.TotalPlayers
		
		for _, g := range result.ByGame {
			stats.ByGame[g]++
		}
		for _, r := range result.ByRegion {
			stats.ByRegion[r]++
		}
		for _, m := range result.ByMode {
			stats.ByMode[m]++
		}
	}
	
	return stats, nil
}
