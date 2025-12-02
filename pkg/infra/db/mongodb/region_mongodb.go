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

// RegionWriter interface for writing region data
type RegionWriter interface {
	Create(ctx context.Context, game *entities.Region) (*entities.Region, error)
	Update(ctx context.Context, game *entities.Region) (*entities.Region, error)
	Put(ctx context.Context, gameID uuid.UUID, game *entities.Region) (string, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// RegionReader interface for reading region data
type RegionReader interface {
	common.Searchable[entities.Region]
}

// RegionRepository combines all region data operations
type RegionRepository interface {
	RegionWriter
	RegionReader
}

type regionRepository struct {
	MongoDBRepository[entities.Region]
}

func NewRegionRepository(client *mongo.Client, dbName string, collectionName string) RegionRepository {
	repo := MongoDBRepository[entities.Region]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(entities.Region{}),
		BsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(entities.Region{}).Name(),
		QueryableFields:   make(map[string]bool),
	}

	repo.InitQueryableFields(map[string]FieldInfo{
		"ID":          {true, "_id"},
		"Name":        {true, "name"},
		"Description": {true, "description"},
	})

	return &regionRepository{repo}
}

func (r *regionRepository) GetByGameID(ctx context.Context, gameID uuid.UUID) ([]*entities.Region, error) {
	filter := bson.M{"game_id": gameID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var regions []*entities.Region
	for cursor.Next(ctx) {
		var region entities.Region
		if err := cursor.Decode(&region); err != nil {
			return nil, err
		}
		regions = append(regions, &region)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return regions, nil
}

// Create implements GameRepository.
func (r *regionRepository) Create(ctx context.Context, game *entities.Region) (*entities.Region, error) {
	panic("unimplemented")
}

// Update implements GameRepository.
func (r *regionRepository) Update(ctx context.Context, game *entities.Region) (*entities.Region, error) {
	panic("unimplemented")
}

// Delete implements GameRepository.
func (r *regionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// Compile implements GameRepository.
func (r *regionRepository) Compile(ctx context.Context, searchParams []common.SearchAggregation, resultOptions common.SearchResultOptions) (*common.Search, error) {
	panic("unimplemented")
}

// GetByID implements GameRepository.
func (r *regionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Region, error) {
	panic("unimplemented")
}

// Put implements GameRepository.
func (r *regionRepository) Put(ctx context.Context, gameID uuid.UUID, game *entities.Region) (string, error) {
	panic("unimplemented")
}

// Search implements GameRepository.
func (r *regionRepository) Search(ctx context.Context, s common.Search) ([]*entities.Region, error) {
	panic("unimplemented")
}
