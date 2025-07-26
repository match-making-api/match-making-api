package usecases

import (
	"context"

	"github.com/golobby/container/v3"
	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/domain/iam/ports/in"
	"github.com/leet-gaming/match-making-api/pkg/infra/iam"
)

type VerifyRIDUseCase struct {
	Client iam.RIDServiceClient
}

func NewVerifyRIDUseCase(client iam.RIDServiceClient) in.VerifyRIDKeyCommand {
	return &VerifyRIDUseCase{
		Client: client,
	}
}

func InjectVerifyRID(c container.Container) error {
	c.Singleton(func(client iam.RIDServiceClient) (in.VerifyRIDKeyCommand, error) {
		return NewVerifyRIDUseCase(client), nil
	})

	return nil
}

func (usecase *VerifyRIDUseCase) Exec(ctx context.Context, id uuid.UUID, operationID string) (*iam.ValidateRIDResponse, error) {
	req := &iam.ValidateRIDRequest{
		RidToken:    id.String(),
		OperationId: operationID,
	}

	resp, err := usecase.Client.ValidateRID(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
