package pairing_out_ports

import (
	pairing_entities "github.com/psavelis/match-making-api/pkg/domain/pairing/entities"
	pairing_value_objects "github.com/psavelis/match-making-api/pkg/domain/pairing/value-objects"
)

type PoolReader interface {
	FindPool(c *pairing_value_objects.Criteria) *pairing_entities.Pool
}
