package usecases

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEloCalculator_ExpectedScore(t *testing.T) {
	calc := NewEloCalculator(EloKFactor)

	// Equal ratings: expected 0.5
	assert.InDelta(t, 0.5, calc.ExpectedScore(1000, 1000), 0.001)

	// Higher rating vs lower: expected > 0.5
	assert.Greater(t, calc.ExpectedScore(1200, 1000), 0.5)

	// Lower rating vs higher: expected < 0.5
	assert.Less(t, calc.ExpectedScore(800, 1000), 0.5)
}

func TestEloCalculator_Delta(t *testing.T) {
	calc := NewEloCalculator(32)

	// Draw: both get ~0 delta (small rounding)
	delta0 := calc.Delta(1000, 1000, 0.5)
	delta1 := calc.Delta(1000, 1000, 0.5)
	assert.Equal(t, int32(0), delta0)
	assert.Equal(t, int32(0), delta1)

	// Win: winner gains, loser loses (symmetric)
	deltaWin := calc.Delta(1000, 1000, 1)
	deltaLose := calc.Delta(1000, 1000, 0)
	assert.Greater(t, deltaWin, int32(0))
	assert.Less(t, deltaLose, int32(0))
	assert.Equal(t, -deltaLose, deltaWin)
}
