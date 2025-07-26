package mongodb

import (
	"reflect"

	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"go.mongodb.org/mongo-driver/mongo"
)

type GameRepository struct {
	MongoDBRepository[game_entities.Game]
}

func NewGameRepository(client *mongo.Client, dbName string, collectionName string) *GameRepository {
	repo := MongoDBRepository[game_entities.Game]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(game_entities.Game{}),
		BsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(game_entities.Game{}).Name(),
		QueryableFields:   make(map[string]bool),
	}

	repo.InitQueryableFields(map[string]FieldInfo{
		"ID":          {true, "_id"},
		"Name":        {true, "name"},
		"Description": {true, "description"},
	})

	return &GameRepository{repo}
}
