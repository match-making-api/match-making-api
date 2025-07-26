package mongodb

import (
	"reflect"

	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	"go.mongodb.org/mongo-driver/mongo"
)

type GameModeRepository struct {
	MongoDBRepository[game_entities.GameMode]
}

func NewGameModeRepository(client *mongo.Client, dbName string, collectionName string) *GameModeRepository {
	repo := MongoDBRepository[game_entities.GameMode]{
		mongoClient:       client,
		dbName:            dbName,
		mappingCache:      make(map[string]CacheItem),
		entityModel:       reflect.TypeOf(game_entities.GameMode{}),
		BsonFieldMappings: make(map[string]string),
		collectionName:    collectionName,
		entityName:        reflect.TypeOf(game_entities.GameMode{}).Name(),
		QueryableFields:   make(map[string]bool),
	}

	repo.InitQueryableFields(map[string]FieldInfo{
		"ID":          {true, "_id"},
		"Name":        {true, "name"},
		"Description": {true, "description"},
	})

	return &GameModeRepository{repo}
}
