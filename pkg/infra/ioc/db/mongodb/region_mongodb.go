package mongodb

import (
	"reflect"

	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"go.mongodb.org/mongo-driver/mongo"
)

type RegionRepository struct {
	MongoDBRepository[game_entities.Region]
}

func NewRegionRepository(client *mongo.Client, dbName string, collectionName string) *RegionRepository {
	repo := MongoDBRepository[game_entities.Region]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(game_entities.Region{}),
		BsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(game_entities.Region{}).Name(),
		QueryableFields:   make(map[string]bool),
	}

	repo.InitQueryableFields(map[string]FieldInfo{
		"ID":          {true, "_id"},
		"Name":        {true, "name"},
		"Description": {true, "description"},
	})

	return &RegionRepository{repo}
}
