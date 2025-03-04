// question: many-to-many between party/peers and lobbies
// fact: shard of a lobby (according to latency, preferences/settings/schedule)
// fact: when pattern matched, fifo
package entities

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	lobbies_entities "github.com/leet-gaming/match-making-api/pkg/domain/lobbies/entities"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
)

type Pool struct {
	Parties []uuid.UUID `json:"party_ids" bson:"party_ids"` // alterar para PairRequest/ objeto que armazene data de entraada no pool
	Lobby   lobbies_entities.Lobby
	// MinimumDate *time.Time
	// MaximumDate *time.Time

	// Criteria pairing_value_objects.Criteria `json:"criteria" bson:"criteria"`

	PartySize uint8
	CreatedAt time.Time
	UpdatedAt time.Time
	mutex     *sync.Mutex `json:"-" bson:"-"`
	cond      *sync.Cond  `json:"-" bson:"-"`
}

func NewPool(mutex *sync.Mutex, cond *sync.Cond, c pairing_value_objects.Criteria) *Pool {
	return &Pool{
		mutex: mutex,
		cond:  cond,
		// Criteria: c,
	}
}

func (e *Pool) Join(pid uuid.UUID) int {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for i, p := range e.Parties {
		if p == pid {
			e.cond.Signal()
			return i + 1
		}
	}

	e.Parties = append(e.Parties, pid) // TODO: timestamp da entrada no pool

	position := len(e.Parties)
	e.cond.Signal()

	return position
}

func (e *Pool) Peek(qty int) []uuid.UUID {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for len(e.Parties) == 0 {
		e.cond.Wait() // test how does it behaves on a empty queue
	}

	if len(e.Parties) < qty {
		return nil
	}

	p := e.Parties[:qty]
	e.Parties = e.Parties[qty:]
	return p
}

func (e *Pool) Remove(partyID uuid.UUID) (int, error) {
	for i, pid := range e.Parties {
		if pid == partyID {
			e.Parties = append(e.Parties[:i], e.Parties[i+1:]...)
			return i + 1, nil
		}
	}

	return -1, fmt.Errorf("Pool.Remove: PartyID %v not in pool")
}

func (e *Pool) IsQueued(pid uuid.UUID) (int, bool) {
	for i, p := range e.Parties {
		if p == pid {
			return i, true
		}
	}

	return -1, false
}
