package mongodb

import (
	"context"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const matchResultsCollection = "match_results"

// Ensure matchResultRepository implements MatchResultRepository.
var _ pairing_out.MatchResultRepository = (*matchResultRepository)(nil)

type matchResultRepository struct {
	client *mongo.Client
	dbName string
}

// NewMatchResultRepository creates a MongoDB repository for match results.
func NewMatchResultRepository(client *mongo.Client, dbName string) pairing_out.MatchResultRepository {
	return &matchResultRepository{client: client, dbName: dbName}
}

func (r *matchResultRepository) collection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection(matchResultsCollection)
}

// Save persists a match result. Idempotent: uses match_id as _id; if exists, returns existing.
func (r *matchResultRepository) Save(ctx context.Context, result *entities.MatchResult) (*entities.MatchResult, error) {
	if err := result.ValidateOwnership(); err != nil {
		return nil, err
	}
	coll := r.collection()
	matchIDStr := result.MatchID.String()

	// Use match_id as _id for idempotency
	filter := bson.M{"_id": matchIDStr}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":                    matchIDStr,
			"player_ids":             result.PlayerIDs,
			"winner_team_id":         result.WinnerTeamID,
			"is_draw":                result.IsDraw,
			"completed_at_epoch_ms":  result.CompletedAtMs,
			"calculated_at_epoch_ms": result.CalculatedAtMs,
			"tenant_id":              result.TenantID,
			"client_id":              result.ClientID,
			"resource_owner_id":      result.ResourceOwnerID,
			"resource_owner": bson.M{
				"tenant_id": result.ResourceOwner.TenantID,
				"client_id": result.ResourceOwner.ClientID,
				"group_id":  result.ResourceOwner.GroupID,
				"user_id":   result.ResourceOwner.UserID,
			},
			"source_event_id": result.SourceEventID,
			"calculated_at":   result.CalculatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	res, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	// If upsert inserted, return our result; else fetch existing
	if res.UpsertedCount > 0 {
		return result, nil
	}

	existing, err := r.GetByMatchID(ctx, result.MatchID)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

// GetByMatchID returns the result for a match, or nil if not found.
func (r *matchResultRepository) GetByMatchID(ctx context.Context, matchID uuid.UUID) (*entities.MatchResult, error) {
	coll := r.collection()
	filter := bson.M{"_id": matchID.String()}

	var doc struct {
		MatchID         string `bson:"_id"`
		PlayerIDs       []string
		WinnerTeamID    string
		IsDraw          bool
		CompletedAtMs   int64 `bson:"completed_at_epoch_ms"`
		CalculatedAtMs  int64 `bson:"calculated_at_epoch_ms"`
		TenantID        string
		ClientID        string
		ResourceOwnerID string `bson:"resource_owner_id"`
		SourceEventID   string `bson:"source_event_id"`
	}

	err := coll.FindOne(ctx, filter).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	out := &entities.MatchResult{
		MatchID:         matchID,
		PlayerIDs:       doc.PlayerIDs,
		WinnerTeamID:    doc.WinnerTeamID,
		IsDraw:          doc.IsDraw,
		CompletedAtMs:   doc.CompletedAtMs,
		CalculatedAtMs:  doc.CalculatedAtMs,
		TenantID:        doc.TenantID,
		ClientID:        doc.ClientID,
		ResourceOwnerID: doc.ResourceOwnerID,
		SourceEventID:   doc.SourceEventID,
	}
	out.EnsureResourceOwner()
	return out, nil
}
