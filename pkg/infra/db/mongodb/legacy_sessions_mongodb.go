package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DefaultLegacySessionsCollection = "matchmaking_sessions"

// LegacySessionsRepository reads replay-api matchmaking_sessions for consistency checks.
type LegacySessionsRepository struct {
	collection *mongo.Collection
}

// NewLegacySessionsRepository creates a read-only adapter for legacy sessions.
func NewLegacySessionsRepository(client *mongo.Client, dbName, collectionName string) *LegacySessionsRepository {
	if collectionName == "" {
		collectionName = DefaultLegacySessionsCollection
	}
	return &LegacySessionsRepository{
		collection: client.Database(dbName).Collection(collectionName),
	}
}

type legacySessionDoc struct {
	PlayerID  any       `bson:"player_id"`
	Status    string    `bson:"status"`
	MatchID   any       `bson:"match_id,omitempty"`
	ExpiresAt time.Time `bson:"expires_at"`
}

// ListByStatuses returns non-expired sessions whose status is in statuses.
func (r *LegacySessionsRepository) ListByStatuses(ctx context.Context, statuses []string) ([]pairing_out.LegacySessionRef, error) {
	if r == nil || r.collection == nil {
		return nil, fmt.Errorf("legacy sessions repository not configured")
	}
	filter := bson.M{
		"status":     bson.M{"$in": statuses},
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	}
	opts := options.Find().
		SetProjection(bson.M{"player_id": 1, "status": 1, "match_id": 1, "expires_at": 1}).
		SetLimit(50_000)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("list legacy sessions: %w", err)
	}
	defer cursor.Close(ctx)

	var out []pairing_out.LegacySessionRef
	for cursor.Next(ctx) {
		var doc legacySessionDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode legacy session: %w", err)
		}
		pid, err := parseBSONUUID(doc.PlayerID)
		if err != nil {
			continue
		}
		ref := pairing_out.LegacySessionRef{
			PlayerID: pid,
			Status:   doc.Status,
		}
		if mid, err := parseBSONUUID(doc.MatchID); err == nil {
			ref.MatchID = &mid
		}
		out = append(out, ref)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func parseBSONUUID(v any) (uuid.UUID, error) {
	switch t := v.(type) {
	case nil:
		return uuid.Nil, fmt.Errorf("nil uuid")
	case string:
		return uuid.Parse(t)
	case uuid.UUID:
		return t, nil
	case primitive.Binary:
		if len(t.Data) == 16 {
			return uuid.FromBytes(t.Data)
		}
		return uuid.Nil, fmt.Errorf("binary uuid length %d", len(t.Data))
	case []byte:
		if len(t) == 16 {
			return uuid.FromBytes(t)
		}
		return uuid.Nil, fmt.Errorf("byte uuid length %d", len(t))
	default:
		return uuid.Nil, fmt.Errorf("unsupported uuid type %T", v)
	}
}

var _ pairing_out.LegacySessionReader = (*LegacySessionsRepository)(nil)
