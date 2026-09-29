package entities

import "github.com/google/uuid"

const BracketTypeSingleElimination = "single_elimination"

const (
	BracketMatchPending = "pending"
	BracketMatchStarted = "started"
	BracketMatchBye     = "bye"
)

// BracketMatch is one slot in a single-elimination round.
type BracketMatch struct {
	Slot      int         `json:"slot" bson:"slot"`
	Round     int         `json:"round" bson:"round"`
	PlayerIDs []uuid.UUID `json:"player_ids" bson:"player_ids"`
	Status    string      `json:"status" bson:"status"`
	MatchID   string      `json:"match_id,omitempty" bson:"match_id,omitempty"`
}

// Bracket is the auditable tournament slate stored on the lobby.
type Bracket struct {
	Type    string         `json:"type" bson:"type"`
	Matches []BracketMatch `json:"matches" bson:"matches"`
}

// BuildSingleElimination pairs seated players into round 1.
// An odd player receives a bye.
func BuildSingleElimination(playerIDs []uuid.UUID) *Bracket {
	matches := make([]BracketMatch, 0, (len(playerIDs)+1)/2)
	slot := 1
	for i := 0; i < len(playerIDs); i += 2 {
		if i+1 >= len(playerIDs) {
			matches = append(matches, BracketMatch{
				Slot:      slot,
				Round:     1,
				PlayerIDs: []uuid.UUID{playerIDs[i]},
				Status:    BracketMatchBye,
			})
			break
		}
		matches = append(matches, BracketMatch{
			Slot:      slot,
			Round:     1,
			PlayerIDs: []uuid.UUID{playerIDs[i], playerIDs[i+1]},
			Status:    BracketMatchPending,
		})
		slot++
	}
	return &Bracket{Type: BracketTypeSingleElimination, Matches: matches}
}
