package pairing_usecases

import (
	"github.com/google/uuid"
	party_entities "github.com/psavelis/match-making-api/pkg/domain/parties/entities"
)

type FindPairUseCase struct{}

type FindPairRequest struct {
	PartyID uuid.UUID
	// Peer  *party_entities.Peer

}

func (uc *FindPairUseCase) Execute() error {
	var party *party_entities.Party

	return nil
}
