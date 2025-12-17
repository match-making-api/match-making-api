package usecases_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	pairing_usecases "github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
	"github.com/leet-gaming/match-making-api/test/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPartyMatcher_Execute(t *testing.T) {
	now := time.Now()

	// Define some test UUIDs
	uuid1 := uuid.New()
	uuid2 := uuid.New()
	uuid3 := uuid.New()
	uuid4 := uuid.New()
	uuid5 := uuid.New()
	uuid6 := uuid.New()
	uuid7 := uuid.New()

	// Define some schedules with overlapping and non-overlapping timeframes
	schedules := map[uuid.UUID]schedule_entities.Schedule{
		uuid1: {
			ID:    uuid1,
			Type:  schedule_entities.Availability,
			Party: nil,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Months:   []time.Month{now.Month()},
					Weekdays: []time.Weekday{now.Weekday()},
					Days:     []int{now.Day()},
					TimeFrames: []schedule_entities.TimeFrame{
						{Start: now.Add(-1 * time.Hour), End: now.Add(5 * time.Hour)},
					},
				},
			},
		},
		uuid2: {
			ID:    uuid2,
			Type:  schedule_entities.Availability,
			Party: nil,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Months:   []time.Month{now.Month()},
					Weekdays: []time.Weekday{now.Weekday()},
					Days:     []int{now.Day() + 1},
					TimeFrames: []schedule_entities.TimeFrame{
						{Start: now.Add(1 * time.Hour), End: now.Add(2 * time.Hour)},
					},
				},
			},
		},
		uuid3: {
			ID:    uuid3,
			Type:  schedule_entities.Availability,
			Party: nil,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Months:   []time.Month{now.Month()},
					Weekdays: []time.Weekday{now.Weekday()},
					Days:     []int{now.Day() + 2},
					TimeFrames: []schedule_entities.TimeFrame{
						{Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour)},
					},
				},
			},
		},
		uuid4: {
			ID:    uuid4,
			Type:  schedule_entities.Availability,
			Party: nil,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Months:   []time.Month{now.Month()},
					Weekdays: []time.Weekday{now.Weekday()},
					Days:     []int{now.Day()},
					TimeFrames: []schedule_entities.TimeFrame{
						{Start: now.Add(-2 * time.Hour), End: now.Add(-1 * time.Hour)},
					},
				},
			},
		},
		uuid5: {
			ID:    uuid5,
			Type:  schedule_entities.Availability,
			Party: nil,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Months:   []time.Month{now.Month()},
					Weekdays: []time.Weekday{now.Weekday()},
					Days:     []int{now.Day()},
					TimeFrames: []schedule_entities.TimeFrame{
						{Start: now.Add(3 * time.Hour), End: now.Add(4 * time.Hour)},
					},
				},
			},
		},
		uuid6: {
			ID:    uuid6,
			Type:  schedule_entities.Availability,
			Party: nil,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Months:   []time.Month{now.Month()},
					Weekdays: []time.Weekday{now.Weekday()},
					Days:     []int{now.Day()},
					TimeFrames: []schedule_entities.TimeFrame{
						{Start: now.Add(4 * time.Hour), End: now.Add(5 * time.Hour)},
					},
				},
			},
		},
		uuid7: {
			ID:    uuid7,
			Type:  schedule_entities.Availability,
			Party: nil,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Months:   []time.Month{now.Month()},
					Weekdays: []time.Weekday{now.Weekday()},
					Days:     []int{now.Day()},
					TimeFrames: []schedule_entities.TimeFrame{
						{Start: now.Add(8 * time.Hour), End: now.Add(9 * time.Hour)},
					},
				},
			},
		},
	}

	// Convert schedules map to pointer map for mock
	scheduleMap := make(map[uuid.UUID]*schedule_entities.Schedule)
	for k, v := range schedules {
		scheduleCopy := v
		scheduleMap[k] = &scheduleCopy
	}

	tests := []struct {
		name      string
		pids      []uuid.UUID
		qty       int
		matched   []uuid.UUID
		wantMatch bool
		wantErr   bool
	}{
		{
			name:      "No parties available",
			pids:      []uuid.UUID{},
			qty:       1,
			matched:   []uuid.UUID{},
			wantMatch: false,
			wantErr:   true,
		},
		{
			name:      "Insufficient parties available",
			pids:      []uuid.UUID{uuid1},
			qty:       2,
			matched:   []uuid.UUID{},
			wantMatch: false,
			wantErr:   true,
		},
		{
			name:      "Parties with overlapping schedules should match",
			pids:      []uuid.UUID{uuid1, uuid5},
			qty:       2,
			matched:   []uuid.UUID{},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:      "Parties with non-overlapping schedules should not match",
			pids:      []uuid.UUID{uuid2, uuid3},
			qty:       2,
			matched:   []uuid.UUID{},
			wantMatch: false,
			wantErr:   true,
		},
		{
			name:      "Three parties with compatible schedules",
			pids:      []uuid.UUID{uuid1, uuid5, uuid6},
			qty:       3,
			matched:   []uuid.UUID{},
			wantMatch: true,
			wantErr:   false,
		},
		{
			name:      "Already matched parties should return immediately",
			pids:      []uuid.UUID{uuid1, uuid2},
			qty:       2,
			matched:   []uuid.UUID{uuid1, uuid5},
			wantMatch: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduleReader := mocks.NewMockPartyScheduleReader(scheduleMap)
			pm := pairing_usecases.NewPartyScheduleMatcher(scheduleReader)
			
			matchedParties, err := pm.Execute(tt.pids, tt.qty, tt.matched)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("PartyScheduleMatcher.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantMatch {
				assert.NotNil(t, matchedParties, "Expected matched parties but got nil")
				assert.Equal(t, tt.qty, len(matchedParties), "Expected %d matched parties, got %d", tt.qty, len(matchedParties))
				
				// Verify all matched parties have schedules
				for _, pid := range matchedParties {
					schedule := scheduleReader.GetScheduleByPartyID(pid)
					assert.NotNil(t, schedule, "Matched party %v should have a schedule", pid)
				}
			} else {
				if !tt.wantErr {
					assert.Nil(t, matchedParties, "Expected no match but got matched parties")
				}
			}
		})
	}
}
// TestPartyMatcher_Cache verifies that schedule caching works correctly
func TestPartyMatcher_Cache(t *testing.T) {
	now := time.Now()
	uuid1 := uuid.New()
	uuid2 := uuid.New()

	schedules := map[uuid.UUID]schedule_entities.Schedule{
		uuid1: {
			ID:    uuid1,
			Type:  schedule_entities.Availability,
			Party: nil,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Months:   []time.Month{now.Month()},
					Weekdays: []time.Weekday{now.Weekday()},
					Days:     []int{now.Day()},
					TimeFrames: []schedule_entities.TimeFrame{
						{Start: now.Add(-1 * time.Hour), End: now.Add(5 * time.Hour)},
					},
				},
			},
		},
		uuid2: {
			ID:    uuid2,
			Type:  schedule_entities.Availability,
			Party: nil,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Months:   []time.Month{now.Month()},
					Weekdays: []time.Weekday{now.Weekday()},
					Days:     []int{now.Day()},
					TimeFrames: []schedule_entities.TimeFrame{
						{Start: now.Add(3 * time.Hour), End: now.Add(4 * time.Hour)},
					},
				},
			},
		},
	}

	scheduleMap := make(map[uuid.UUID]*schedule_entities.Schedule)
	for k, v := range schedules {
		scheduleCopy := v
		scheduleMap[k] = &scheduleCopy
	}

	scheduleReader := mocks.NewMockPartyScheduleReader(scheduleMap)
	pm := pairing_usecases.NewPartyScheduleMatcher(scheduleReader)

	// First execution should load schedules into cache
	matchedParties, err := pm.Execute([]uuid.UUID{uuid1, uuid2}, 2, []uuid.UUID{})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(matchedParties))

	// Clear cache and verify it can be cleared
	if matcher, ok := pm.(*pairing_usecases.PartyScheduleMatcher); ok {
		matcher.ClearCache()
	}

	// Execute again - should still work (will reload from repository)
	matchedParties2, err2 := pm.Execute([]uuid.UUID{uuid1, uuid2}, 2, []uuid.UUID{})
	assert.NoError(t, err2)
	assert.Equal(t, 2, len(matchedParties2))
}

// BenchmarkPartyMatcher_SmallSet benchmarks the matcher with a small set of parties
func BenchmarkPartyMatcher_SmallSet(b *testing.B) {
	now := time.Now()
	parties := make([]uuid.UUID, 5)
	schedules := make(map[uuid.UUID]*schedule_entities.Schedule)

	for i := 0; i < 5; i++ {
		parties[i] = uuid.New()
		schedules[parties[i]] = &schedule_entities.Schedule{
			ID:    parties[i],
			Type:  schedule_entities.Availability,
			Party: nil,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Months:   []time.Month{now.Month()},
					Weekdays: []time.Weekday{now.Weekday()},
					Days:     []int{now.Day()},
					TimeFrames: []schedule_entities.TimeFrame{
						{Start: now.Add(time.Duration(i) * time.Hour), End: now.Add(time.Duration(i+2) * time.Hour)},
					},
				},
			},
		}
	}

	scheduleReader := mocks.NewMockPartyScheduleReader(schedules)
	pm := pairing_usecases.NewPartyScheduleMatcher(scheduleReader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if matcher, ok := pm.(*pairing_usecases.PartyScheduleMatcher); ok {
			matcher.ClearCache()
		}
		_, _ = pm.Execute(parties, 3, []uuid.UUID{})
	}
}

// BenchmarkPartyMatcher_LargeSet benchmarks the matcher with a large set of parties (triggers parallel processing)
func BenchmarkPartyMatcher_LargeSet(b *testing.B) {
	now := time.Now()
	parties := make([]uuid.UUID, 20)
	schedules := make(map[uuid.UUID]*schedule_entities.Schedule)

	for i := 0; i < 20; i++ {
		parties[i] = uuid.New()
		schedules[parties[i]] = &schedule_entities.Schedule{
			ID:    parties[i],
			Type:  schedule_entities.Availability,
			Party: nil,
			Options: map[int]schedule_entities.DateOption{
				0: {
					Months:   []time.Month{now.Month()},
					Weekdays: []time.Weekday{now.Weekday()},
					Days:     []int{now.Day()},
					TimeFrames: []schedule_entities.TimeFrame{
						{Start: now.Add(time.Duration(i) * time.Hour), End: now.Add(time.Duration(i+2) * time.Hour)},
					},
				},
			},
		}
	}

	scheduleReader := mocks.NewMockPartyScheduleReader(schedules)
	pm := pairing_usecases.NewPartyScheduleMatcher(scheduleReader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if matcher, ok := pm.(*pairing_usecases.PartyScheduleMatcher); ok {
			matcher.ClearCache()
		}
		_, _ = pm.Execute(parties, 5, []uuid.UUID{})
	}
}
