package mongodb

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GameModeWriter interface for writing game mode data
type GameModeWriter interface {
	Create(ctx context.Context, game *entities.GameMode) (*entities.GameMode, error)
	Update(ctx context.Context, game *entities.GameMode) (*entities.GameMode, error)
	Put(ctx context.Context, gameID uuid.UUID, game *entities.GameMode) (string, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// GameModeReader interface for reading game mode data
type GameModeReader interface {
	common.Searchable[entities.GameMode]
}

// GameModeRepository combines all game mode data operations
type GameModeRepository interface {
	GameModeWriter
	GameModeReader
}

type gameModeRepository struct {
	MongoDBRepository[entities.GameMode]
}

// NewGameModeRepository creates a new game mode repository
func NewGameModeRepository(client *mongo.Client, dbName string, collectionName string) GameModeRepository {
	repo := MongoDBRepository[entities.GameMode]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(entities.GameMode{}),
		BsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entities.GameMode{}).Name(),
		QueryableFields:   make(map[string]bool),
	}

	repo.InitQueryableFields(map[string]FieldInfo{
		"ID":          {true, "_id"},
		"Name":        {true, "name"},
		"Description": {true, "description"},
	})

	return &gameModeRepository{repo}
}

// GetByGameID retrieves game modes by game ID
func (r *gameModeRepository) GetByGameID(ctx context.Context, gameID uuid.UUID) ([]*entities.GameMode, error) {
	filter := bson.M{"game_id": gameID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var gameModes []*entities.GameMode
	if err = cursor.All(ctx, &gameModes); err != nil {
		return nil, err
	}

	return gameModes, nil
}

// Compile implements GameModeRepository.
func (r *gameModeRepository) Compile(ctx context.Context, searchParams []common.SearchAggregation, resultOptions common.SearchResultOptions) (*common.Search, error) {
	panic("unimplemented")
}

// Create implements GameModeRepository.
func (r *gameModeRepository) Create(ctx context.Context, gameMode *entities.GameMode) (*entities.GameMode, error) {
	gameMode.ID = uuid.New()

	_, err := r.collection.InsertOne(ctx, gameMode)
	if err != nil {
		return nil, err
	}

	return gameMode, nil
}

// Update implements GameModeRepository.
func (r *gameModeRepository) Update(ctx context.Context, gameMode *entities.GameMode) (*entities.GameMode, error) {
	filter := bson.M{"_id": gameMode.ID}
	update := bson.M{"$set": gameMode}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return gameMode, nil
}

// Delete implements GameModeRepository.
func (r *gameModeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// GetByID implements GameModeRepository.
func (r *gameModeRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.GameMode, error) {
	filter := bson.M{"_id": id}

	var gameMode entities.GameMode
	err := r.collection.FindOne(ctx, filter).Decode(&gameMode)
	if err != nil {
		return nil, err
	}

	return &gameMode, nil
}

// Search implements GameModeRepository.
func (r *gameModeRepository) Search(ctx context.Context, s common.Search) ([]*entities.GameMode, error) {
	filter := bson.M{}

	// If there are search filters, apply them here
	// For now, return all game modes
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var gameModes []*entities.GameMode
	if err = cursor.All(ctx, &gameModes); err != nil {
		return nil, err
	}

	return gameModes, nil
}

// Put implements GameModeRepository.
func (r *gameModeRepository) Put(ctx context.Context, gameID uuid.UUID, gameMode *entities.GameMode) (string, error) {
	panic("unimplemented")
}
