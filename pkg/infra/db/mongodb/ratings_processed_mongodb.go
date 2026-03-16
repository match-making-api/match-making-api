package mongodb

import (
	"context"

	"github.com/google/uuid"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const ratingsProcessedCollection = "ratings_processed_matches"

// Ensure ratingsProcessedStore implements RatingsProcessedStore.
var _ pairing_out.RatingsProcessedStore = (*ratingsProcessedStore)(nil)

type ratingsProcessedStore struct {
	client *mongo.Client
	dbName string
}

// NewRatingsProcessedStore creates a MongoDB store for ratings idempotency.
func NewRatingsProcessedStore(client *mongo.Client, dbName string) pairing_out.RatingsProcessedStore {
	return &ratingsProcessedStore{client: client, dbName: dbName}
}

func (s *ratingsProcessedStore) collection() *mongo.Collection {
	return s.client.Database(s.dbName).Collection(ratingsProcessedCollection)
}

// HasProcessed returns true if ratings have already been calculated for this match.
func (s *ratingsProcessedStore) HasProcessed(ctx context.Context, matchID uuid.UUID) (bool, error) {
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

// MarkProcessed records that ratings were processed for this match.
func (s *ratingsProcessedStore) MarkProcessed(ctx context.Context, matchID uuid.UUID) error {
	coll := s.collection()
	doc := bson.M{"_id": matchID.String()}
	_, err := coll.InsertOne(ctx, doc)
	return err
}
