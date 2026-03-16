package mongodb

import (
	"context"

	"github.com/google/uuid"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const analyticsTrackedCollection = "analytics_tracked_matches"

// Ensure analyticsTrackedStore implements AnalyticsTrackedStore.
var _ pairing_out.AnalyticsTrackedStore = (*analyticsTrackedStore)(nil)

type analyticsTrackedStore struct {
	client *mongo.Client
	dbName string
}

// NewAnalyticsTrackedStore creates a MongoDB store for analytics idempotency.
func NewAnalyticsTrackedStore(client *mongo.Client, dbName string) pairing_out.AnalyticsTrackedStore {
	return &analyticsTrackedStore{client: client, dbName: dbName}
}

func (s *analyticsTrackedStore) collection() *mongo.Collection {
	return s.client.Database(s.dbName).Collection(analyticsTrackedCollection)
}

// HasTracked returns true if AnalyticsTracked was already published for this match.
func (s *analyticsTrackedStore) HasTracked(ctx context.Context, matchID uuid.UUID) (bool, error) {
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

// MarkTracked records that analytics was published for this match.
func (s *analyticsTrackedStore) MarkTracked(ctx context.Context, matchID uuid.UUID) error {
	coll := s.collection()
	doc := bson.M{"_id": matchID.String()}
	_, err := coll.InsertOne(ctx, doc)
	return err
}
