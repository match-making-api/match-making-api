package pairing_usecases

import (
	"fmt"

	"github.com/google/uuid"
	pairing_entities "github.com/psavelis/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/psavelis/match-making-api/pkg/domain/pairing/ports/out"
	parties_out "github.com/psavelis/match-making-api/pkg/domain/parties/ports/out"
)

type CreatePairUseCase struct {
	PartyReader parties_out.PartyReader
	PairWriter  pairing_out.PairWriter
}

func (uc *CreatePairUseCase) Execute(partyIDs []uuid.UUID) (*pairing_entities.Pair, error) {
	var err error
	pair := pairing_entities.NewPair(len(partyIDs))
	for _, partyID := range partyIDs {
		pair.Match[partyID], err = uc.PartyReader.GetByID(partyID)

		if err != nil {
			return nil, fmt.Errorf("CreatePairUseCase.Execute: unable to create pair. PartyID: %v not found (Error: %v)", partyID, err)
		}
	}

	pair, err = uc.PairWriter.Save(pair)

	if err != nil {
		return nil, fmt.Errorf("CreatePairUseCase.Execute: unable to create pair for PartyIDs %v, due to create error: %v", partyIDs, err)
	}

	return pair, nil
}
