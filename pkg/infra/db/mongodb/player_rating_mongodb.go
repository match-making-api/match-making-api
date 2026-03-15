package mongodb

import (
	"context"
	"time"

	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	playerRatingsCollection  = "player_ratings"
	ratingAuditCollection    = "rating_audit"
)

// Ensure playerRatingRepository implements PlayerRatingRepository.
var _ pairing_out.PlayerRatingRepository = (*playerRatingRepository)(nil)

type playerRatingRepository struct {
	client *mongo.Client
	dbName string
}

// NewPlayerRatingRepository creates a MongoDB repository for player ratings.
func NewPlayerRatingRepository(client *mongo.Client, dbName string) pairing_out.PlayerRatingRepository {
	return &playerRatingRepository{client: client, dbName: dbName}
}

func (r *playerRatingRepository) ratingsColl() *mongo.Collection {
	return r.client.Database(r.dbName).Collection(playerRatingsCollection)
}

func (r *playerRatingRepository) auditColl() *mongo.Collection {
	return r.client.Database(r.dbName).Collection(ratingAuditCollection)
}

// ratingDocKey builds a composite key for player+game+tenant+client.
func ratingDocKey(playerID, gameID, tenantID, clientID string) string {
	return playerID + "|" + gameID + "|" + tenantID + "|" + clientID
}

// GetByPlayer fetches the rating for a player.
func (r *playerRatingRepository) GetByPlayer(ctx context.Context, playerID, gameID, tenantID, clientID string) (*entities.PlayerRating, error) {
	coll := r.ratingsColl()
	key := ratingDocKey(playerID, gameID, tenantID, clientID)
	filter := bson.M{"_id": key}

	var doc struct {
		ID               string `bson:"_id"`
		PlayerID         string `bson:"player_id"`
		GameID           string `bson:"game_id"`
		TenantID         string `bson:"tenant_id"`
		ClientID         string `bson:"client_id"`
		ResourceOwnerID  string `bson:"resource_owner_id"`
		MMR              int32  `bson:"mmr"`
		UpdatedAt       time.Time `bson:"updated_at"`
	}

	err := coll.FindOne(ctx, filter).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &entities.PlayerRating{
		PlayerID:        doc.PlayerID,
		GameID:          doc.GameID,
		TenantID:        doc.TenantID,
		ClientID:        doc.ClientID,
		ResourceOwnerID: doc.ResourceOwnerID,
		MMR:             doc.MMR,
		UpdatedAt:       doc.UpdatedAt,
	}, nil
}

// Save upserts the player's rating.
func (r *playerRatingRepository) Save(ctx context.Context, rating *entities.PlayerRating) error {
	coll := r.ratingsColl()
	key := ratingDocKey(rating.PlayerID, rating.GameID, rating.TenantID, rating.ClientID)
	rating.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": key}
	update := bson.M{
		"$set": bson.M{
			"_id":                   key,
			"player_id":             rating.PlayerID,
			"game_id":               rating.GameID,
			"tenant_id":             rating.TenantID,
			"client_id":             rating.ClientID,
			"resource_owner_id":     rating.ResourceOwnerID,
			"mmr":                   rating.MMR,
			"updated_at":            rating.UpdatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx, filter, update, opts)
	return err
}

// SaveAuditEntry persists a rating change for audit trail.
func (r *playerRatingRepository) SaveAuditEntry(ctx context.Context, entry *entities.RatingAuditEntry) error {
	coll := r.auditColl()
	entry.UpdatedAt = time.Now().UTC()

	doc := bson.M{
		"match_id":           entry.MatchID.String(),
		"player_id":          entry.PlayerID,
		"mmr_before":         entry.MMRBefore,
		"mmr_after":          entry.MMRAfter,
		"delta":              entry.Delta,
		"algorithm_version":  entry.AlgorithmVersion,
		"reason":             entry.Reason,
		"updated_by":         entry.UpdatedBy,
		"updated_at":         entry.UpdatedAt,
	}

	_, err := coll.InsertOne(ctx, doc)
	return err
}
