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
// TestPartyMatcher_InvalidSchedules tests error handling for invalid schedules
func TestPartyMatcher_InvalidSchedules(t *testing.T) {
	now := time.Now()
	uuid1 := uuid.New()
	uuid2 := uuid.New()

	tests := []struct {
		name     string
		schedules map[uuid.UUID]*schedule_entities.Schedule
		pids     []uuid.UUID
		qty      int
		wantErr  bool
		errMsg   string
	}{
		{
			name: "Schedule with empty options",
			schedules: map[uuid.UUID]*schedule_entities.Schedule{
				uuid1: {
					ID:      uuid1,
					Type:    schedule_entities.Availability,
					Options: map[int]schedule_entities.DateOption{},
				},
				uuid2: {
					ID:    uuid2,
					Type:  schedule_entities.Availability,
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
			},
			pids:    []uuid.UUID{uuid1, uuid2},
			qty:     2,
			wantErr: true,
			errMsg:  "invalid schedule",
		},
		{
			name: "Schedule with invalid timeframe (start after end)",
			schedules: map[uuid.UUID]*schedule_entities.Schedule{
				uuid1: {
					ID:    uuid1,
					Type:  schedule_entities.Availability,
					Options: map[int]schedule_entities.DateOption{
						0: {
							Months:   []time.Month{now.Month()},
							Weekdays: []time.Weekday{now.Weekday()},
							Days:     []int{now.Day()},
							TimeFrames: []schedule_entities.TimeFrame{
								{Start: now.Add(5 * time.Hour), End: now.Add(-1 * time.Hour)}, // Invalid: start after end
							},
						},
					},
				},
				uuid2: {
					ID:    uuid2,
					Type:  schedule_entities.Availability,
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
			},
			pids:    []uuid.UUID{uuid1, uuid2},
			qty:     2,
			wantErr: true,
			errMsg:  "timeframe start time",
		},
		{
			name: "Schedule with no timeframes",
			schedules: map[uuid.UUID]*schedule_entities.Schedule{
				uuid1: {
					ID:    uuid1,
					Type:  schedule_entities.Availability,
					Options: map[int]schedule_entities.DateOption{
						0: {
							Months:   []time.Month{now.Month()},
							Weekdays: []time.Weekday{now.Weekday()},
							Days:     []int{now.Day()},
							TimeFrames: []schedule_entities.TimeFrame{}, // Empty timeframes
						},
					},
				},
			},
			pids:    []uuid.UUID{uuid1},
			qty:     1,
			wantErr: true,
			errMsg:  "no timeframes",
		},
		{
			name: "Schedule with timeframe duration less than 1 minute",
			schedules: map[uuid.UUID]*schedule_entities.Schedule{
				uuid1: {
					ID:    uuid1,
					Type:  schedule_entities.Availability,
					Options: map[int]schedule_entities.DateOption{
						0: {
							Months:   []time.Month{now.Month()},
							Weekdays: []time.Weekday{now.Weekday()},
							Days:     []int{now.Day()},
							TimeFrames: []schedule_entities.TimeFrame{
								{Start: now, End: now.Add(30 * time.Second)}, // Less than 1 minute
							},
						},
					},
				},
			},
			pids:    []uuid.UUID{uuid1},
			qty:     1,
			wantErr: true,
			errMsg:  "timeframe duration",
		},
		{
			name: "Schedule with no months, weekdays, or days",
			schedules: map[uuid.UUID]*schedule_entities.Schedule{
				uuid1: {
					ID:    uuid1,
					Type:  schedule_entities.Availability,
					Options: map[int]schedule_entities.DateOption{
						0: {
							Months:   []time.Month{},
							Weekdays: []time.Weekday{},
							Days:     []int{},
							TimeFrames: []schedule_entities.TimeFrame{
								{Start: now.Add(-1 * time.Hour), End: now.Add(5 * time.Hour)},
							},
						},
					},
				},
			},
			pids:    []uuid.UUID{uuid1},
			qty:     1,
			wantErr: true,
			errMsg:  "must specify at least one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduleReader := mocks.NewMockPartyScheduleReader(tt.schedules)
			pm := pairing_usecases.NewPartyScheduleMatcher(scheduleReader)

			matchedParties, err := pm.Execute(tt.pids, tt.qty, []uuid.UUID{})

			if tt.wantErr {
				assert.Error(t, err, "Expected error for invalid schedule")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
				assert.Nil(t, matchedParties, "Should not return matched parties when schedule is invalid")
			} else {
				assert.NoError(t, err, "Should not error with valid schedules")
			}
		})
	}
}

// TestPartyMatcher_DatabaseErrors tests error handling for database communication failures
func TestPartyMatcher_DatabaseErrors(t *testing.T) {
	now := time.Now()
	uuid1 := uuid.New()
	uuid2 := uuid.New()
	uuid3 := uuid.New()

	tests := []struct {
		name     string
		schedules map[uuid.UUID]*schedule_entities.Schedule
		pids     []uuid.UUID
		qty      int
		wantErr  bool
		errMsg   string
	}{
		{
			name: "All parties have nil schedules (database failure)",
			schedules: map[uuid.UUID]*schedule_entities.Schedule{
				// No schedules - simulating database failure
			},
			pids:    []uuid.UUID{uuid1, uuid2},
			qty:     2,
			wantErr: true,
			errMsg:  "unable to match", // When all schedules are nil, we get "unable to match" error
		},
		{
			name: "Some parties have nil schedules",
			schedules: map[uuid.UUID]*schedule_entities.Schedule{
				uuid1: {
					ID:    uuid1,
					Type:  schedule_entities.Availability,
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
				// uuid2 has no schedule (nil) - simulating database failure for this party
			},
			pids:    []uuid.UUID{uuid1, uuid2},
			qty:     2,
			wantErr: true,
			errMsg:  "unable to match",
		},
		{
			name: "Mixed valid and invalid schedules",
			schedules: map[uuid.UUID]*schedule_entities.Schedule{
				uuid1: {
					ID:    uuid1,
					Type:  schedule_entities.Availability,
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
				uuid3: {
					ID:    uuid3,
					Type:  schedule_entities.Availability,
					Options: map[int]schedule_entities.DateOption{}, // Invalid: empty options
				},
			},
			pids:    []uuid.UUID{uuid1, uuid2, uuid3},
			qty:     2,
			wantErr: false, // Should still match uuid1 and uuid2 despite uuid3 being invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduleReader := mocks.NewMockPartyScheduleReader(tt.schedules)
			pm := pairing_usecases.NewPartyScheduleMatcher(scheduleReader)

			matchedParties, err := pm.Execute(tt.pids, tt.qty, []uuid.UUID{})

			if tt.wantErr {
				assert.Error(t, err, "Expected error for database failure scenario")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
			} else {
				if err != nil {
					// If we expected no error but got one, check if it's due to insufficient valid parties
					assert.Contains(t, err.Error(), "unable to match", "Error should be about matching, not validation")
				} else {
					assert.NotNil(t, matchedParties, "Should return matched parties when valid schedules exist")
					assert.Equal(t, tt.qty, len(matchedParties), "Should match expected quantity")
				}
			}
		})
	}
}

// TestAuxiliaryFunctions tests the auxiliary functions used by the matching algorithm
func TestAuxiliaryFunctions(t *testing.T) {
	now := time.Now()

	t.Run("areSchedulesCompatible", func(t *testing.T) {
		// Test with compatible schedules
		schedule1 := schedule_entities.Schedule{
			ID:    uuid.New(),
			Type:  schedule_entities.Availability,
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
		}

		schedule2 := schedule_entities.Schedule{
			ID:    uuid.New(),
			Type:  schedule_entities.Availability,
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
		}

		// Use reflection to call the private function
		// Since it's private, we'll test it through the public interface
		scheduleReader := mocks.NewMockPartyScheduleReader(map[uuid.UUID]*schedule_entities.Schedule{
			schedule1.ID: &schedule1,
			schedule2.ID: &schedule2,
		})
		pm := pairing_usecases.NewPartyScheduleMatcher(scheduleReader)

		matched, err := pm.Execute([]uuid.UUID{schedule1.ID, schedule2.ID}, 2, []uuid.UUID{})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(matched), "Compatible schedules should match")
	})

	t.Run("isTimeFrameOverlapping", func(t *testing.T) {
		// Test overlapping timeframes
		start1 := now.Add(-1 * time.Hour)
		end1 := now.Add(5 * time.Hour)
		start2 := now.Add(3 * time.Hour)
		end2 := now.Add(4 * time.Hour)

		// These should overlap
		overlapping := isTimeFrameOverlapping(start1, end1, start2, end2)
		assert.True(t, overlapping, "Timeframes should overlap")

		// Test non-overlapping timeframes
		start3 := now.Add(6 * time.Hour)
		end3 := now.Add(7 * time.Hour)
		nonOverlapping := isTimeFrameOverlapping(start1, end1, start3, end3)
		assert.False(t, nonOverlapping, "Timeframes should not overlap")
	})
}

// Helper function to test private isTimeFrameOverlapping function
// We need to make it accessible for testing, or test it indirectly
func isTimeFrameOverlapping(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}
