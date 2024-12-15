package pairing_usecases

import pairing_value_objects "github.com/psavelis/match-making-api/pkg/domain/pairing/value-objects"

type CreatePoolUseCase struct{}

type CreatePoolPayload struct {
	Criteria *pairing_value_objects.Criteria
}

func (uc *CreatePoolUseCase) Execute(p *CreatePoolPayload) error {
	

	return nil
}
