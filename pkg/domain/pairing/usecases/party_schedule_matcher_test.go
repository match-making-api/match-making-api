package usecases_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	pairing_usecases "github.com/leet-gaming/match-making-api/pkg/domain/pairing/usecases"
	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
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
						{Start: now.Add(6 * time.Hour), End: now.Add(7 * time.Hour)},
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

	tests := []struct {
		name    string
		pids    []uuid.UUID
		qty     int
		matched []uuid.UUID
		want    bool
		wantErr bool
	}{
		{"No parties available", []uuid.UUID{}, 1, []uuid.UUID{}, false, true},
		{"Insufficient parties available", []uuid.UUID{uuid1}, 2, []uuid.UUID{}, false, true},
		// {"Exact parties available (overlapping timeframes)", []uuid.UUID{uuid1, uuid2}, 2, []uuid.UUID{}, true, false},
		{"More parties than needed, all available", []uuid.UUID{uuid1, uuid2, uuid3}, 2, []uuid.UUID{}, false, true},
		{"More parties than needed, some unavailable", []uuid.UUID{uuid2, uuid3}, 2, []uuid.UUID{}, false, true},
		{"No exact match, but partial matches possible", []uuid.UUID{uuid1, uuid3, uuid4}, 2, []uuid.UUID{uuid1, uuid4}, true, false},
		// {"Three parties required, all available", []uuid.UUID{uuid1, uuid2, uuid3}, 3, []uuid.UUID{}, true, false},
		// {"Four parties required, all available", []uuid.UUID{uuid1, uuid2, uuid3, uuid4}, 4, []uuid.UUID{}, true, false},
		// {"Five parties required, all available", []uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5}, 5, []uuid.UUID{}, true, false},
		{"Three parties required, some unavailable", []uuid.UUID{uuid2, uuid3, uuid4}, 3, []uuid.UUID{}, false, true},
		{"Three parties required, some unavailable, partial matches possible", []uuid.UUID{uuid1, uuid2, uuid3, uuid4}, 3, []uuid.UUID{uuid1, uuid4}, true, false},
		{"Three parties required, some unavailable, no partial matches possible", []uuid.UUID{uuid1, uuid2, uuid3, uuid4}, 3, []uuid.UUID{}, false, true},
		{"Three parties required, some unavailable, partial matches possible", []uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5}, 3, []uuid.UUID{uuid1, uuid4, uuid5}, true, false},
		{"Three parties required, some unavailable, no partial matches possible", []uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5}, 3, []uuid.UUID{}, false, true},
		{"Three parties required, some unavailable, partial matches possible", []uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5, uuid6}, 3, []uuid.UUID{uuid1, uuid4, uuid5}, true, false},
		{"Three parties required, some unavailable, no partial matches possible", []uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5, uuid6}, 3, []uuid.UUID{}, false, true},
		{"Three parties required, some unavailable, partial matches possible", []uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5, uuid6, uuid7}, 3, []uuid.UUID{uuid1, uuid4, uuid5}, true, false},
		{"Three parties required, some unavailable, no partial matches possible", []uuid.UUID{uuid1, uuid2, uuid3, uuid4, uuid5, uuid6, uuid7}, 3, []uuid.UUID{}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := &pairing_usecases.PartyScheduleMatcher{Schedules: schedules}
			success, matchedParties, err := pm.Execute(tt.pids, tt.qty, tt.matched)
			if (err != nil) != tt.wantErr {
				t.Errorf("PartyScheduleMatcher.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if success != tt.want {
				t.Errorf("PartyScheduleMatcher.Execute() = %v, want %v", success, tt.want)
			}
			// Additional assertions for matched parties
			if success && len(tt.matched) > 0 {
				for _, pid := range tt.matched {
					_, ok := schedules[pid]
					assert.True(t, ok, "Matched party not found in schedules")
				}
			}
			// Additional assertions for matched parties returned by pm.Execute
			if success && len(matchedParties) > 0 {
				for _, pid := range matchedParties {
					_, ok := schedules[pid]
					assert.True(t, ok, "Matched party not found in returned parties")
				}
			}
		})
	}
}
