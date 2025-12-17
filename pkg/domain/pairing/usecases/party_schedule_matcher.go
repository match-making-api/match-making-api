package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	pairing_in "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/in"
	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
	schedules_in_ports "github.com/leet-gaming/match-making-api/pkg/domain/schedules/ports/in"
)

// PartyScheduleMatcher implements the recursive matching algorithm for parties based on schedule compatibility
type PartyScheduleMatcher struct {
	ScheduleReader schedules_in_ports.PartyScheduleReader
}

// NewPartyScheduleMatcher creates a new instance of PartyScheduleMatcher
func NewPartyScheduleMatcher(scheduleReader schedules_in_ports.PartyScheduleReader) pairing_in.PartyScheduleMatcher {
	return &PartyScheduleMatcher{
		ScheduleReader: scheduleReader,
	}
}

// Execute identifies compatible matches between parties recursively
// It handles cases where parties have non-overlapping schedules by checking all possible combinations
// Returns the matched parties or an error if no matches are found
func (pm *PartyScheduleMatcher) Execute(pids []uuid.UUID, qty int, matched []uuid.UUID) ([]uuid.UUID, error) {
	ctx := context.Background()
	if qty <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	// Base case: we have enough matched parties
	if len(matched) >= qty {
		return matched, nil
	}

	// Base case: no more parties available
	if len(pids) == 0 {
		return nil, fmt.Errorf("not enough parties available to match the required quantity: need %d, have %d", qty, len(matched))
	}

	// Recursive case: try to find matches starting with each available party
	for i, current := range pids {
		currentSchedule := pm.ScheduleReader.GetScheduleByPartyID(current)
		if currentSchedule == nil {
			slog.WarnContext(ctx, "party has no schedule, skipping", "party_id", current)
			continue
		}

		// Find parties with compatible schedules
		matchingParties := pm.findMatchingParties(ctx, pids, i, *currentSchedule)
		
		// Check if we have enough matching parties to complete the match
		needed := qty - len(matched) - 1
		if len(matchingParties) >= needed {
			// Try to build a match starting with this party
			result, err := pm.matchParties(ctx, pids, i, qty, matched, current, matchingParties)
			if err == nil {
				return result, nil
			}
			// If this combination didn't work, continue to next party
		}
	}

	// No valid match found
	return nil, fmt.Errorf("unable to match the required quantity of parties: need %d, have %d matched, %d available", qty, len(matched), len(pids))
}

// matchParties attempts to build a complete match starting with the current party
func (pm *PartyScheduleMatcher) matchParties(ctx context.Context, pids []uuid.UUID, currentIndex, qty int, matched []uuid.UUID, current uuid.UUID, matchingParties []uuid.UUID) ([]uuid.UUID, error) {
	// Add current party to matched list
	newMatched := append(matched, current)
	
	// Remove current party from available list
	remainingPids := append(pids[:currentIndex], pids[currentIndex+1:]...)
	
	// Try to add matching parties until we reach the required quantity
	for _, match := range matchingParties {
		// Check if this party is still available (might have been used in a recursive call)
		if !containsUUID(remainingPids, match) {
			continue
		}
		
		newMatched = append(newMatched, match)
		remainingPids = removeUUID(remainingPids, match)
		
		// If we have enough parties, return success
		if len(newMatched) >= qty {
			return newMatched[:qty], nil
		}
	}
	
		// Recursively try to find remaining parties
	return pm.Execute(remainingPids, qty, newMatched)
}

// findMatchingParties finds all parties with schedules compatible with the given schedule
func (pm *PartyScheduleMatcher) findMatchingParties(ctx context.Context, pids []uuid.UUID, currentIndex int, schedule schedule_entities.Schedule) []uuid.UUID {
	var matchingParties []uuid.UUID
	
	for i, otherPID := range pids {
		// Skip the current party
		if i == currentIndex {
			continue
		}
		
		// Get the other party's schedule
		otherSchedule := pm.ScheduleReader.GetScheduleByPartyID(otherPID)
		if otherSchedule == nil {
			slog.DebugContext(ctx, "party has no schedule, skipping compatibility check", "party_id", otherPID)
			continue
		}
		
		// Check if schedules are compatible (have overlapping availability)
		if areSchedulesCompatible(schedule, *otherSchedule) {
			matchingParties = append(matchingParties, otherPID)
		}
	}
	
	return matchingParties
}

// areSchedulesCompatible checks if two schedules have any overlapping availability
// This handles cases where parties have non-overlapping schedules by returning false
func areSchedulesCompatible(schedule1 schedule_entities.Schedule, schedule2 schedule_entities.Schedule) bool {
	// Check all combinations of date options from both schedules
	for _, option1 := range schedule1.Options {
		for _, option2 := range schedule2.Options {
			if hasMatchingAvailability(option1, option2) {
				return true
			}
		}
	}
	return false
}

func hasMatchingAvailability(option1 schedule_entities.DateOption, option2 schedule_entities.DateOption) bool {
	for _, day1 := range option1.Days {
		for _, weekday1 := range option1.Weekdays {
			for _, timeframe1 := range option1.TimeFrames {
				for _, day2 := range option2.Days {
					for _, weekday2 := range option2.Weekdays {
						for _, timeframe2 := range option2.TimeFrames {
							if isAvailableCombination(day1, weekday1, timeframe1, day2, weekday2, timeframe2) {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func isAvailableCombination(day1 int, weekday1 time.Weekday, timeframe1 schedule_entities.TimeFrame,
	day2 int, weekday2 time.Weekday, timeframe2 schedule_entities.TimeFrame) bool {
	if weekday1 != weekday2 && !(weekday1 == time.Sunday && weekday2 == time.Saturday) {
		return false
	}
	if day1 != 0 && day2 != 0 && day1 != day2 {
		return false
	}
	return isTimeFrameOverlapping(timeframe1.Start, timeframe1.End, timeframe2.Start, timeframe2.End)
}

func isTimeFrameOverlapping(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}

// removeUUID removes the first occurrence of the given UUID from the slice
func removeUUID(slice []uuid.UUID, id uuid.UUID) []uuid.UUID {
	for i, v := range slice {
		if v == id {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// containsUUID checks if the slice contains the given UUID
func containsUUID(slice []uuid.UUID, id uuid.UUID) bool {
	for _, v := range slice {
		if v == id {
			return true
		}
	}
	return false
}
