package usecases

import "math"

const (
	// EloAlgorithmVersion identifies the rating algorithm for audit.
	EloAlgorithmVersion = "elo-v1"
	// DefaultMMR is the starting rating for new players.
	DefaultMMR = 1000
	// EloKFactor is the maximum rating change per match.
	EloKFactor = 32
)

// EloCalculator computes rating changes using the Elo system.
type EloCalculator struct {
	KFactor int32
}

// NewEloCalculator creates an Elo calculator with optional K factor.
func NewEloCalculator(kFactor int32) *EloCalculator {
	if kFactor <= 0 {
		kFactor = EloKFactor
	}
	return &EloCalculator{KFactor: kFactor}
}

// ExpectedScore returns the expected score (0-1) for player A against player B.
// E_a = 1 / (1 + 10^((R_b - R_a)/400))
func (e *EloCalculator) ExpectedScore(ratingA, ratingB int32) float64 {
	diff := float64(ratingB-ratingA) / 400.0
	return 1.0 / (1.0 + math.Pow(10, diff))
}

// Delta computes the rating change for a player.
// delta = K * (actualScore - expectedScore)
func (e *EloCalculator) Delta(rating, opponentRating int32, actualScore float64) int32 {
	expected := e.ExpectedScore(rating, opponentRating)
	delta := float64(e.KFactor) * (actualScore - expected)
	return int32(delta)
}

