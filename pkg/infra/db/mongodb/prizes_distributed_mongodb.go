package mongodb

import (
	"context"

	"github.com/google/uuid"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const prizesDistributedCollection = "prizes_distributed_matches"

// Ensure prizesDistributedStore implements PrizesDistributedStore.
var _ pairing_out.PrizesDistributedStore = (*prizesDistributedStore)(nil)

type prizesDistributedStore struct {
	client *mongo.Client
	dbName string
}

// NewPrizesDistributedStore creates a MongoDB store for prize distribution idempotency.
func NewPrizesDistributedStore(client *mongo.Client, dbName string) pairing_out.PrizesDistributedStore {
	return &prizesDistributedStore{client: client, dbName: dbName}
}

func (s *prizesDistributedStore) collection() *mongo.Collection {
	return s.client.Database(s.dbName).Collection(prizesDistributedCollection)
}

// HasDistributed returns true if PrizeDistributed was already published for this match.
func (s *prizesDistributedStore) HasDistributed(ctx context.Context, matchID uuid.UUID) (bool, error) {
	coll := s.collection()
	filter := bson.M{"_id": matchID.String()}
	err := coll.FindOne(ctx, filter).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// MarkDistributed records that prize distribution was published for this match.
func (s *prizesDistributedStore) MarkDistributed(ctx context.Context, matchID uuid.UUID) error {
	coll := s.collection()
	doc := bson.M{"_id": matchID.String()}
	_, err := coll.InsertOne(ctx, doc)
	return err
}
