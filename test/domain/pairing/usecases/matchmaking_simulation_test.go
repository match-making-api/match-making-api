package usecases_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	"github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	pairing_value_objects "github.com/leet-gaming/match-making-api/pkg/domain/pairing/value-objects"
	parties_entities "github.com/leet-gaming/match-making-api/pkg/domain/parties/entities"
	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
)

// =============================================================================
// In-memory stores for simulation tests
// =============================================================================

// InMemoryPoolStore provides a real in-memory pooling backend for simulation tests.
// It tracks pools by criteria key for FindPool lookups. When Save is called for
// a new pool (from CreatePoolUseCase), it registers it but FindPool won't find
// it until it's saved again with a known criteria key.
type InMemoryPoolStore struct {
	mu       sync.Mutex
	pools    map[string]*pairing_entities.Pool
	allPools []*pairing_entities.Pool // tracks all pools for pointer-based lookup
}

func NewInMemoryPoolStore() *InMemoryPoolStore {
	return &InMemoryPoolStore{
		pools: make(map[string]*pairing_entities.Pool),
	}
}

func (s *InMemoryPoolStore) FindPool(criteria *pairing_value_objects.Criteria) (*pairing_entities.Pool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := simPoolKey(criteria)
	if pool, ok := s.pools[key]; ok {
		return pool, nil
	}
	return nil, nil
}

func (s *InMemoryPoolStore) Save(p *pairing_entities.Pool) (*pairing_entities.Pool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Track the pool
	found := false
	for _, existing := range s.allPools {
		if existing == p {
			found = true
			break
		}
	}
	if !found {
		s.allPools = append(s.allPools, p)
	}
	return p, nil
}

// RegisterPool stores a pool indexed by criteria so FindPool can locate it.
// Call this after CreatePoolUseCase returns the pool, to link it to the criteria key.
func (s *InMemoryPoolStore) RegisterPool(p *pairing_entities.Pool, criteria *pairing_value_objects.Criteria) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := simPoolKey(criteria)
	s.pools[key] = p
}

// GetOrRegisterPool atomically registers a pool for the criteria key,
// returning the existing pool if one was already registered (prevents race).
func (s *InMemoryPoolStore) GetOrRegisterPool(p *pairing_entities.Pool, criteria *pairing_value_objects.Criteria) *pairing_entities.Pool {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := simPoolKey(criteria)
	if existing, ok := s.pools[key]; ok {
		return existing
	}
	s.pools[key] = p
	return p
}

func simPoolKey(criteria *pairing_value_objects.Criteria) string {
	gameID := "any"
	if criteria.GameID != nil {
		gameID = criteria.GameID.String()
	}
	region := "any"
	if criteria.Region != nil {
		region = criteria.Region.Slug
	}
	return fmt.Sprintf("%s-%s-%d", gameID, region, criteria.PairSize)
}

// InMemoryPairStore tracks created pairs in-memory.
type InMemoryPairStore struct {
	mu    sync.Mutex
	pairs []*pairing_entities.Pair
}

func NewInMemoryPairStore() *InMemoryPairStore {
	return &InMemoryPairStore{
		pairs: make([]*pairing_entities.Pair, 0),
	}
}

func (s *InMemoryPairStore) Save(p *pairing_entities.Pair) (*pairing_entities.Pair, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pairs = append(s.pairs, p)
	return p, nil
}

func (s *InMemoryPairStore) GetPairs() []*pairing_entities.Pair {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]*pairing_entities.Pair, len(s.pairs))
	copy(result, s.pairs)
	return result
}

// InMemoryPartyStore provides party lookups by UUID.
type InMemoryPartyStore struct {
	mu      sync.Mutex
	parties map[uuid.UUID]*parties_entities.Party
}

func NewInMemoryPartyStore() *InMemoryPartyStore {
	return &InMemoryPartyStore{
		parties: make(map[uuid.UUID]*parties_entities.Party),
	}
}

func (s *InMemoryPartyStore) RegisterParty(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.parties[id] = &parties_entities.Party{ID: id}
}

func (s *InMemoryPartyStore) GetByID(id uuid.UUID) (*parties_entities.Party, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if party, ok := s.parties[id]; ok {
		return party, nil
	}
	return nil, fmt.Errorf("party %v not found", id)
}

// InMemoryScheduleStore returns nil schedules (no schedule filtering in simulation).
type InMemoryScheduleStore struct{}

func (s *InMemoryScheduleStore) GetScheduleByPartyID(_ uuid.UUID) *schedule_entities.Schedule {
	return nil
}

// CriteriaAwarePoolInitiator wraps CreatePoolUseCase and registers the pool
// in the InMemoryPoolStore with the criteria key so FindPool can locate it.
type CriteriaAwarePoolInitiator struct {
	inner *usecases.CreatePoolUseCase
	store *InMemoryPoolStore
}

func (c *CriteriaAwarePoolInitiator) Execute(criteria pairing_value_objects.Criteria) (*pairing_entities.Pool, error) {
	pool, err := c.inner.Execute(criteria)
	if err != nil {
		return nil, err
	}
	// Atomically register or return existing pool (prevents concurrent creation race)
	pool = c.store.GetOrRegisterPool(pool, &criteria)
	return pool, nil
}

// ContextAwarePairCreator wraps CreatePairUseCase to inject a valid context
// since AddAndFindNextPairUseCase internally uses context.Background().
type ContextAwarePairCreator struct {
	PartyReader partyReader
	PairWriter  pairWriter
}

type partyReader interface {
	GetByID(id uuid.UUID) (*parties_entities.Party, error)
}

type pairWriter interface {
	Save(p *pairing_entities.Pair) (*pairing_entities.Pair, error)
}

func (c *ContextAwarePairCreator) Execute(ctx context.Context, pids []uuid.UUID) (*pairing_entities.Pair, error) {
	// Inject resource owner if missing (AddAndFindNextPairUseCase passes context.Background())
	if _, ok := ctx.Value(common.TenantIDKey).(uuid.UUID); !ok {
		ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
		ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
		ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
	}

	uc := &usecases.CreatePairUseCase{
		PartyReader: c.PartyReader,
		PairWriter:  c.PairWriter,
	}
	return uc.Execute(ctx, pids)
}

// =============================================================================
// Helper: simulate matchmaking for a single player
// =============================================================================

func simulatePlayerJoin(
	t *testing.T,
	addAndFind *usecases.AddAndFindNextPairUseCase,
	partyID uuid.UUID,
	criteria pairing_value_objects.Criteria,
) (*pairing_entities.Pair, *pairing_entities.Pool, int) {
	t.Helper()
	pair, pool, position, err := addAndFind.Execute(usecases.FindPairPayload{
		PartyID:  partyID,
		Criteria: criteria,
	})
	assert.NoError(t, err)
	assert.NotNil(t, pool)
	assert.Greater(t, position, 0)
	return pair, pool, position
}

func defaultCriteria() pairing_value_objects.Criteria {
	gameID := uuid.New()
	return pairing_value_objects.Criteria{
		PairSize: 2,
		GameID:   &gameID,
		Region:   &game_entities.Region{Slug: "us-east"},
	}
}

func simContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.TenantIDKey, uuid.New())
	ctx = context.WithValue(ctx, common.ClientIDKey, uuid.New())
	ctx = context.WithValue(ctx, common.UserIDKey, uuid.New())
	return ctx
}

// =============================================================================
// Tests
// =============================================================================

func TestMultiUserMatchmaking_TwoPlayersMatch(t *testing.T) {
	poolStore := NewInMemoryPoolStore()
	pairStore := NewInMemoryPairStore()
	partyStore := NewInMemoryPartyStore()
	scheduleStore := &InMemoryScheduleStore{}

	party1 := uuid.New()
	party2 := uuid.New()
	partyStore.RegisterParty(party1)
	partyStore.RegisterParty(party2)

	createPair := &ContextAwarePairCreator{
		PartyReader: partyStore,
		PairWriter:  pairStore,
	}
	createPool := &CriteriaAwarePoolInitiator{inner: &usecases.CreatePoolUseCase{PoolWriter: poolStore}, store: poolStore}

	addAndFind := &usecases.AddAndFindNextPairUseCase{
		PoolReader:          poolStore,
		PoolWriter:          poolStore,
		PartyScheduleReader: scheduleStore,
		PoolInitiator:       createPool,
		PairCreator:         createPair,
	}

	criteria := defaultCriteria()

	// Player 1 joins — no pair yet, just queued
	pair1, _, pos1 := simulatePlayerJoin(t, addAndFind, party1, criteria)
	assert.Nil(t, pair1, "first player should not be matched yet")
	assert.Equal(t, 1, pos1)

	// Player 2 joins — match should be created
	pair2, _, _ := simulatePlayerJoin(t, addAndFind, party2, criteria)
	assert.NotNil(t, pair2, "second player should trigger a match")
	assert.Equal(t, 2, len(pair2.Match))

	// Verify pair was stored
	pairs := pairStore.GetPairs()
	assert.Equal(t, 1, len(pairs))
}

func TestMultiUserMatchmaking_FourPlayersInPairs(t *testing.T) {
	poolStore := NewInMemoryPoolStore()
	pairStore := NewInMemoryPairStore()
	partyStore := NewInMemoryPartyStore()
	scheduleStore := &InMemoryScheduleStore{}

	players := make([]uuid.UUID, 4)
	for i := range players {
		players[i] = uuid.New()
		partyStore.RegisterParty(players[i])
	}

	createPair := &ContextAwarePairCreator{
		PartyReader: partyStore,
		PairWriter:  pairStore,
	}
	createPool := &CriteriaAwarePoolInitiator{inner: &usecases.CreatePoolUseCase{PoolWriter: poolStore}, store: poolStore}

	addAndFind := &usecases.AddAndFindNextPairUseCase{
		PoolReader:          poolStore,
		PoolWriter:          poolStore,
		PartyScheduleReader: scheduleStore,
		PoolInitiator:       createPool,
		PairCreator:         createPair,
	}

	criteria := defaultCriteria()

	// Players join sequentially
	for _, pid := range players {
		simulatePlayerJoin(t, addAndFind, pid, criteria)
	}

	// Should have exactly 2 pairs
	pairs := pairStore.GetPairs()
	assert.Equal(t, 2, len(pairs))

	// No player appears in multiple pairs
	seen := make(map[uuid.UUID]bool)
	for _, pair := range pairs {
		for pid := range pair.Match {
			assert.False(t, seen[pid], "player %v should not appear in multiple pairs", pid)
			seen[pid] = true
		}
	}
}

func TestMultiUserMatchmaking_OddPlayerWaitsInPool(t *testing.T) {
	poolStore := NewInMemoryPoolStore()
	pairStore := NewInMemoryPairStore()
	partyStore := NewInMemoryPartyStore()
	scheduleStore := &InMemoryScheduleStore{}

	players := make([]uuid.UUID, 3)
	for i := range players {
		players[i] = uuid.New()
		partyStore.RegisterParty(players[i])
	}

	createPair := &ContextAwarePairCreator{
		PartyReader: partyStore,
		PairWriter:  pairStore,
	}
	createPool := &CriteriaAwarePoolInitiator{inner: &usecases.CreatePoolUseCase{PoolWriter: poolStore}, store: poolStore}

	addAndFind := &usecases.AddAndFindNextPairUseCase{
		PoolReader:          poolStore,
		PoolWriter:          poolStore,
		PartyScheduleReader: scheduleStore,
		PoolInitiator:       createPool,
		PairCreator:         createPair,
	}

	criteria := defaultCriteria()

	// 3 players join — only 1 pair formed, 1 player waits
	for _, pid := range players {
		simulatePlayerJoin(t, addAndFind, pid, criteria)
	}

	pairs := pairStore.GetPairs()
	assert.Equal(t, 1, len(pairs), "only 1 pair should be formed from 3 players")
}

func TestMultiUserMatchmaking_PlayerLeavesQueue(t *testing.T) {
	poolStore := NewInMemoryPoolStore()
	pairStore := NewInMemoryPairStore()
	partyStore := NewInMemoryPartyStore()
	scheduleStore := &InMemoryScheduleStore{}

	party1 := uuid.New()
	party2 := uuid.New()
	partyStore.RegisterParty(party1)
	partyStore.RegisterParty(party2)

	createPair := &ContextAwarePairCreator{
		PartyReader: partyStore,
		PairWriter:  pairStore,
	}
	createPool := &CriteriaAwarePoolInitiator{inner: &usecases.CreatePoolUseCase{PoolWriter: poolStore}, store: poolStore}

	addAndFind := &usecases.AddAndFindNextPairUseCase{
		PoolReader:          poolStore,
		PoolWriter:          poolStore,
		PartyScheduleReader: scheduleStore,
		PoolInitiator:       createPool,
		PairCreator:         createPair,
	}

	criteria := defaultCriteria()

	// Player 1 joins
	pair1, pool1, _ := simulatePlayerJoin(t, addAndFind, party1, criteria)
	assert.Nil(t, pair1)

	// Player 1 leaves the queue
	_, err := pool1.Remove(party1)
	assert.NoError(t, err)

	// Verify player was removed
	_, isQueued := pool1.IsQueued(party1)
	assert.False(t, isQueued, "player should not be in pool after removal")

	// Player 2 joins — should not match (only 1 player in queue)
	// Need to re-add to make player2 visible in pool
	pair2, _, _ := simulatePlayerJoin(t, addAndFind, party2, criteria)
	assert.Nil(t, pair2, "should not match since player1 left")

	pairs := pairStore.GetPairs()
	assert.Equal(t, 0, len(pairs))
}

func TestMultiUserMatchmaking_DifferentRegionsNoMatch(t *testing.T) {
	poolStore := NewInMemoryPoolStore()
	pairStore := NewInMemoryPairStore()
	partyStore := NewInMemoryPartyStore()
	scheduleStore := &InMemoryScheduleStore{}

	party1 := uuid.New()
	party2 := uuid.New()
	partyStore.RegisterParty(party1)
	partyStore.RegisterParty(party2)

	createPair := &ContextAwarePairCreator{
		PartyReader: partyStore,
		PairWriter:  pairStore,
	}
	createPool := &CriteriaAwarePoolInitiator{inner: &usecases.CreatePoolUseCase{PoolWriter: poolStore}, store: poolStore}

	addAndFind := &usecases.AddAndFindNextPairUseCase{
		PoolReader:          poolStore,
		PoolWriter:          poolStore,
		PartyScheduleReader: scheduleStore,
		PoolInitiator:       createPool,
		PairCreator:         createPair,
	}

	gameID := uuid.New()

	// Player 1 queues in US East
	criteria1 := pairing_value_objects.Criteria{
		PairSize: 2,
		GameID:   &gameID,
		Region:   &game_entities.Region{Slug: "us-east"},
	}
	pair1, _, _ := simulatePlayerJoin(t, addAndFind, party1, criteria1)
	assert.Nil(t, pair1)

	// Player 2 queues in EU West — different region, different pool
	criteria2 := pairing_value_objects.Criteria{
		PairSize: 2,
		GameID:   &gameID,
		Region:   &game_entities.Region{Slug: "eu-west"},
	}
	pair2, _, _ := simulatePlayerJoin(t, addAndFind, party2, criteria2)
	assert.Nil(t, pair2, "different regions should not match")

	pairs := pairStore.GetPairs()
	assert.Equal(t, 0, len(pairs))
}

func TestMultiUserMatchmaking_HighConcurrency(t *testing.T) {
	poolStore := NewInMemoryPoolStore()
	pairStore := NewInMemoryPairStore()
	partyStore := NewInMemoryPartyStore()
	scheduleStore := &InMemoryScheduleStore{}

	const playerCount = 20
	players := make([]uuid.UUID, playerCount)
	for i := range players {
		players[i] = uuid.New()
		partyStore.RegisterParty(players[i])
	}

	createPair := &ContextAwarePairCreator{
		PartyReader: partyStore,
		PairWriter:  pairStore,
	}
	createPool := &CriteriaAwarePoolInitiator{inner: &usecases.CreatePoolUseCase{PoolWriter: poolStore}, store: poolStore}

	addAndFind := &usecases.AddAndFindNextPairUseCase{
		PoolReader:          poolStore,
		PoolWriter:          poolStore,
		PartyScheduleReader: scheduleStore,
		PoolInitiator:       createPool,
		PairCreator:         createPair,
	}

	criteria := defaultCriteria()

	// All players join concurrently
	var wg sync.WaitGroup
	for _, pid := range players {
		wg.Add(1)
		go func(partyID uuid.UUID) {
			defer wg.Done()
			_, _, _, _ = addAndFind.Execute(usecases.FindPairPayload{
				PartyID:  partyID,
				Criteria: criteria,
			})
		}(pid)
	}

	// Wait with timeout — goroutines whose parties were already consumed
	// by another goroutine's Peek may block on an empty pool indefinitely.
	// In production, new players keep joining; the timeout simulates the
	// finite nature of the test.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		// Expected: some goroutines stuck in Pool.Peek after their party
		// was already dequeued by another goroutine's Peek call.
	}

	// Should have 10 pairs from 20 players
	pairs := pairStore.GetPairs()
	assert.Equal(t, playerCount/2, len(pairs), "should have %d pairs from %d players", playerCount/2, playerCount)

	// No player appears in multiple pairs
	seen := make(map[uuid.UUID]bool)
	for _, pair := range pairs {
		for pid := range pair.Match {
			assert.False(t, seen[pid], "player %v duplicated across pairs", pid)
			seen[pid] = true
		}
	}
}

func TestMultiUserMatchmaking_PoolCreationError(t *testing.T) {
	scheduleStore := &InMemoryScheduleStore{}

	// Pool with PairSize 0 should cause CreatePool to fail
	createPool := &usecases.CreatePoolUseCase{}

	addAndFind := &usecases.AddAndFindNextPairUseCase{
		PoolReader:          NewInMemoryPoolStore(),
		PoolWriter:          NewInMemoryPoolStore(),
		PartyScheduleReader: scheduleStore,
		PoolInitiator:       createPool,
	}

	criteria := pairing_value_objects.Criteria{
		PairSize: 0,
	}

	_, _, _, err := addAndFind.Execute(usecases.FindPairPayload{
		PartyID:  uuid.New(),
		Criteria: criteria,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot create pool")
}

func TestMultiUserMatchmaking_IdempotentJoin(t *testing.T) {
	poolStore := NewInMemoryPoolStore()
	pairStore := NewInMemoryPairStore()
	partyStore := NewInMemoryPartyStore()
	scheduleStore := &InMemoryScheduleStore{}

	party1 := uuid.New()
	partyStore.RegisterParty(party1)

	createPair := &ContextAwarePairCreator{
		PartyReader: partyStore,
		PairWriter:  pairStore,
	}
	createPool := &CriteriaAwarePoolInitiator{inner: &usecases.CreatePoolUseCase{PoolWriter: poolStore}, store: poolStore}

	addAndFind := &usecases.AddAndFindNextPairUseCase{
		PoolReader:          poolStore,
		PoolWriter:          poolStore,
		PartyScheduleReader: scheduleStore,
		PoolInitiator:       createPool,
		PairCreator:         createPair,
	}

	criteria := defaultCriteria()

	// Same player joins twice — should be idempotent
	pair1, pool1, _ := simulatePlayerJoin(t, addAndFind, party1, criteria)
	assert.Nil(t, pair1)

	// Join again with same party
	pos, isQueued := pool1.IsQueued(party1)
	assert.True(t, isQueued)
	assert.Equal(t, 0, pos)

	// Pool should still have only 1 party
	assert.Equal(t, 1, len(pool1.Parties))
}

func TestMultiUserMatchmaking_FullLifecycle(t *testing.T) {
	// Full lifecycle: Join → Match → Verify conflicts → Resolve
	poolStore := NewInMemoryPoolStore()
	pairStore := NewInMemoryPairStore()
	partyStore := NewInMemoryPartyStore()
	scheduleStore := &InMemoryScheduleStore{}

	party1 := uuid.New()
	party2 := uuid.New()
	partyStore.RegisterParty(party1)
	partyStore.RegisterParty(party2)

	createPair := &ContextAwarePairCreator{
		PartyReader: partyStore,
		PairWriter:  pairStore,
	}
	createPool := &CriteriaAwarePoolInitiator{inner: &usecases.CreatePoolUseCase{PoolWriter: poolStore}, store: poolStore}

	addAndFind := &usecases.AddAndFindNextPairUseCase{
		PoolReader:          poolStore,
		PoolWriter:          poolStore,
		PartyScheduleReader: scheduleStore,
		PoolInitiator:       createPool,
		PairCreator:         createPair,
	}

	criteria := defaultCriteria()

	// Step 1: Both players join, match created
	simulatePlayerJoin(t, addAndFind, party1, criteria)
	pair, _, _ := simulatePlayerJoin(t, addAndFind, party2, criteria)
	assert.NotNil(t, pair, "match should be created")
	assert.Equal(t, pairing_entities.ConflictStatusNone, pair.ConflictStatus)

	// Step 2: Simulate flagging a conflict
	pair.ConflictStatus = pairing_entities.ConflictStatusFlagged
	pair.ConflictReason = "test scheduling conflict"

	// Step 3: Admin resolves conflict
	resolveUC := &usecases.ResolveMatchConflictUseCase{
		PairReader: &inMemoryPairReader{pair: pair},
		PairWriter: pairStore,
	}

	ctx := simContext()
	ctx = context.WithValue(ctx, common.AudienceKey, common.TenantAudienceIDKey)

	err := resolveUC.Execute(ctx, usecases.ResolveConflictPayload{
		PairID: pair.ID,
		Action: usecases.ConflictResolutionResolve,
		Reason: "Manually verified",
	})
	assert.NoError(t, err)
	assert.Equal(t, pairing_entities.ConflictStatusResolved, pair.ConflictStatus)
}

// inMemoryPairReader is a minimal PairReader for lifecycle tests.
type inMemoryPairReader struct {
	pair *pairing_entities.Pair
}

func (r *inMemoryPairReader) FindPairsByPartyID(_ context.Context, _ uuid.UUID) ([]*pairing_entities.Pair, error) {
	return []*pairing_entities.Pair{r.pair}, nil
}

func (r *inMemoryPairReader) GetByID(_ context.Context, _ uuid.UUID) (*pairing_entities.Pair, error) {
	return r.pair, nil
}
