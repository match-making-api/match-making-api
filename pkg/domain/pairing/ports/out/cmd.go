package pairing_out

import pairing_entities "github.com/psavelis/match-making-api/pkg/domain/pairing/entities"

type PoolWriter interface {
	Save(p *pairing_entities.Pool) (*pairing_entities.Pool, error)
}

type PairWriter interface {
	Save(p *pairing_entities.Pair) (*pairing_entities.Pair, error)
}
