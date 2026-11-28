package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultTrafficMigrationConfig(t *testing.T) {
	cfg := DefaultTrafficMigrationConfig()
	assert.Equal(t, 100, cfg.NewPathPercent)
	assert.Equal(t, TrafficStrategyPercent, cfg.Strategy)
	assert.True(t, cfg.IsFullRollout())
	assert.True(t, cfg.IsLegacyRoutingRetired())
	assert.False(t, cfg.IsFullRollback())
}

func TestTrafficMigrationConfigFromEnv(t *testing.T) {
	t.Setenv("MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT", "10")
	t.Setenv("MATCHMAKING_TRAFFIC_STRATEGY", "region")
	t.Setenv("MATCHMAKING_TRAFFIC_REGION_ALLOWLIST", "us-east, eu-west ")

	cfg := TrafficMigrationConfigFromEnv()
	assert.Equal(t, 10, cfg.NewPathPercent)
	assert.Equal(t, TrafficStrategyRegion, cfg.Strategy)
	assert.Equal(t, []string{"us-east", "eu-west"}, cfg.RegionAllowlist)
	assert.False(t, cfg.IsFullRollout())
	assert.False(t, cfg.IsFullRollback())

	_ = os.Unsetenv("MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT")
	_ = os.Unsetenv("MATCHMAKING_TRAFFIC_STRATEGY")
	_ = os.Unsetenv("MATCHMAKING_TRAFFIC_REGION_ALLOWLIST")
}

func TestTrafficMigrationConfigFromEnv_ClampAndRollback(t *testing.T) {
	t.Setenv("MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT", "0")
	cfg := TrafficMigrationConfigFromEnv()
	assert.Equal(t, 0, cfg.NewPathPercent)
	assert.True(t, cfg.IsFullRollback())

	t.Setenv("MATCHMAKING_TRAFFIC_NEW_PATH_PERCENT", "250")
	cfg = TrafficMigrationConfigFromEnv()
	assert.Equal(t, 100, cfg.NewPathPercent)
}
