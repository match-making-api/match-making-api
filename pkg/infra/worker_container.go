package infra

import (
	"github.com/golobby/container/v3"
	"github.com/leet-gaming/match-making-api/pkg/common"
	"github.com/leet-gaming/match-making-api/pkg/infra/db/mongodb"
	"github.com/leet-gaming/match-making-api/pkg/infra/ioc"
	"github.com/leet-gaming/match-making-api/pkg/infra/observability"
)

// InjectWorker sets up only the infrastructure needed by background workers
// (consumers and tickers). It skips IAM, billing, squad, and other
// HTTP-specific dependencies that workers don't need.
//
// Includes: MongoDB (regions, game repos), Kafka, Redis.
func InjectWorker(c container.Container) error {
	if err := observability.Init(); err != nil {
		return err
	}

	return common.InjectAll(c,
		ioc.InjectIoc,
		mongodb.InjectGameRepository,
		mongodb.InjectGameModeRepository,
		mongodb.InjectRegionRepository,
		mongodb.InjectMatchResultRepository,
		mongodb.InjectPlayerRatingRepository,
		mongodb.InjectRatingsProcessedStore,
		mongodb.InjectPrizesDistributedStore,
		mongodb.InjectAnalyticsTrackedStore,
		InjectKafka,
		InjectRedis,
	)
}
