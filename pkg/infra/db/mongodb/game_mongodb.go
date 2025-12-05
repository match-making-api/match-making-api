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

// GameWriter interface for writing game data
type GameWriter interface {
	Create(ctx context.Context, game *entities.Game) (*entities.Game, error)
	Update(ctx context.Context, game *entities.Game) (*entities.Game, error)
	Put(ctx context.Context, gameID uuid.UUID, game *entities.Game) (string, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// GameReader interface for reading game data
type GameReader interface {
	common.Searchable[entities.Game]
}

// GameRepository combines all game data operations
type GameRepository interface {
	GameWriter
	GameReader
}

type gameRepository struct {
	MongoDBRepository[entities.Game]
}

func NewGameRepository(client *mongo.Client, dbName string, collectionName string) GameRepository {
	repo := MongoDBRepository[entities.Game]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(entities.Game{}),
		BsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entities.Game{}).Name(),
		QueryableFields:   make(map[string]bool),
	}

	repo.InitQueryableFields(map[string]FieldInfo{
		"ID":          {true, "_id"},
		"Name":        {true, "name"},
		"Description": {true, "description"},
	})

	return &gameRepository{repo}
}

// Create implements GameRepository.  Creates a new game in the database.  Returns the created game if successful, otherwise returns an error.  Note: This function will automatically assign a new UUID to the
func (r *gameRepository) Create(ctx context.Context, game *entities.Game) (*entities.Game, error) {
	game.ID = uuid.New()

	_, err := r.collection.InsertOne(ctx, game)
	if err != nil {
		return nil, err
	}

	return game, nil
}

// Update implements GameRepository.  Updates the game in the database.  If the game does not exist, it will return an error.  Otherwise, it will update the game and return the updated game.  The game's ID must be provided.  The provided game will overwrite any existing fields in the database.  The game's ID field must be set to the ID of the game to be updated.  Returns the updated game if successful, otherwise returns an error.  Note: This
func (r *gameRepository) Update(ctx context.Context, game *entities.Game) (*entities.Game, error) {
	filter := bson.M{"_id": game.ID}
	update := bson.M{"$set": game}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return game, nil
}

// Delete implements GameRepository.
func (r *gameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// Compile implements GameRepository.
func (r *gameRepository) Compile(ctx context.Context, searchParams []common.SearchAggregation, resultOptions common.SearchResultOptions) (*common.Search, error) {
	panic("unimplemented")
}

// GetByID implements GameRepository.
func (r *gameRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Game, error) {
	filter := bson.M{"_id": id}

	var game entities.Game
	err := r.collection.FindOne(ctx, filter).Decode(&game)
	if err != nil {
		return nil, err
	}

	return &game, nil
}

// Put implements GameRepository.
func (r *gameRepository) Put(ctx context.Context, gameID uuid.UUID, game *entities.Game) (string, error) {
	panic("unimplemented")
}

// Search implements GameRepository and common.Searchable.
func (r *gameRepository) Search(ctx context.Context, s common.Search) ([]*entities.Game, error) {
	filter := bson.M{}

	// If there are search filters, apply them here
	// For now, return all games
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var games []*entities.Game
	if err = cursor.All(ctx, &games); err != nil {
		return nil, err
	}

	return games, nil
}
