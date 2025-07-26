package in

import (
	"context"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/infra/iam"
)

type VerifyRIDKeyCommand interface {
	Exec(ctx context.Context, key uuid.UUID, operationId string) (*iam.ValidateRIDResponse, error)
}
