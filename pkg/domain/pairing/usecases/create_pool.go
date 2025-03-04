package usecases

import (
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
)

type CreatePoolUseCase struct{}

type CreatePoolPayload struct {
	Criteria *pairing_value_objects.Criteria
}

func (uc *CreatePoolUseCase) Execute(p *CreatePoolPayload) (*pairing_entities.Pool, error) {

	// definir estrategia de sharding (ie, daily pool, week range pool, onlinepool)

	return nil, nil
}
