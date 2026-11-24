package traffic

import (
	"fmt"
	"testing"

	"github.com/leet-gaming/match-making-api/pkg/infra/config"
	"github.com/stretchr/testify/assert"
)

func TestRouter_PercentZeroIsRollback(t *testing.T) {
	r := NewRouter(config.TrafficMigrationConfig{NewPathPercent: 0, Strategy: config.TrafficStrategyPercent})
	d := r.Decide("player-1", "us-east")
	assert.Equal(t, PathLegacy, d.Path)
	assert.Equal(t, "percent_zero", d.Reason)
}

func TestRouter_PercentFull(t *testing.T) {
	r := NewRouter(config.TrafficMigrationConfig{NewPathPercent: 100, Strategy: config.TrafficStrategyPercent})
	d := r.Decide("player-1", "us-east")
	assert.Equal(t, PathNew, d.Path)
}

func TestRouter_PercentSticky(t *testing.T) {
	r := NewRouter(config.TrafficMigrationConfig{NewPathPercent: 10, Strategy: config.TrafficStrategyPercent})
	first := r.Decide("sticky-player-xyz", "us-east")
	for i := 0; i < 20; i++ {
		assert.Equal(t, first.Path, r.Decide("sticky-player-xyz", "us-east").Path)
	}
}

func TestRouter_PercentDistribution(t *testing.T) {
	r := NewRouter(config.TrafficMigrationConfig{NewPathPercent: 25, Strategy: config.TrafficStrategyPercent})
	newCount := 0
	const n = 1000
	for i := 0; i < n; i++ {
		if r.decide(fmt.Sprintf("player-%d", i), "").Path == PathNew {
			newCount++
		}
	}
	// Rough band around 25% for FNV distribution (not exact).
	ratio := float64(newCount) / float64(n)
	assert.Greater(t, ratio, 0.10)
	assert.Less(t, ratio, 0.40)
}

func TestRouter_RegionAllowlist(t *testing.T) {
	r := NewRouter(config.TrafficMigrationConfig{
		NewPathPercent:  100,
		Strategy:        config.TrafficStrategyRegion,
		RegionAllowlist: []string{"us-east"},
	})
	assert.Equal(t, PathNew, r.Decide("p1", "US-East").Path)
	assert.Equal(t, PathLegacy, r.Decide("p1", "eu-west").Path)
}

func TestStickyInPercent_Boundaries(t *testing.T) {
	assert.False(t, stickyInPercent("any", 0))
	assert.True(t, stickyInPercent("any", 100))
}
