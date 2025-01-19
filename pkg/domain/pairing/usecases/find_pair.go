package usecases

import (
	"github.com/google/uuid"
)

type FindPairUseCase struct{}

type FindPairRequest struct {
	PartyID uuid.UUID
	// Peer  *party_entities.Peer

}

func (uc *FindPairUseCase) Execute() error {
	//var party *entities.Party

	return nil
}
