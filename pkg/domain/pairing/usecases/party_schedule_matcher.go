package pairing_usecases

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	schedule_entities "github.com/psavelis/match-making-api/pkg/domain/schedules/entities"
)

// Implement the PartyMatcher struct
type PartyMatcher struct {
	Schedules map[uuid.UUID]schedule_entities.Schedule
}

func (pm *PartyMatcher) Execute(pids []uuid.UUID, qty int, matched []uuid.UUID) (bool, []uuid.UUID, error) {
	if len(matched) == qty {
		return true, matched, nil
	}

	if len(pids) == 0 {
		return false, matched, fmt.Errorf("not enough parties available to match the required quantity")
	}

	for i := 0; i < len(pids); i++ {
		current := pids[i]
		if schedule, ok := pm.Schedules[current]; ok {
			var matchingParties []uuid.UUID
			for _, otherPID := range pids[:i] {
				if otherSchedule, ok := pm.Schedules[otherPID]; ok {
					if AreAvailable(schedule, otherSchedule) {
						matchingParties = append(matchingParties, otherPID)
					}
				}
			}
			for _, otherPID := range pids[i+1:] {
				if otherSchedule, ok := pm.Schedules[otherPID]; ok {
					if AreAvailable(schedule, otherSchedule) {
						matchingParties = append(matchingParties, otherPID)
					}
				}
			}

			if len(matchingParties) >= qty-len(matched) {
				newMatched := append(matched, current)
				remainingPids := append(append([]uuid.UUID{}, matchingParties...), pids[i+1:]...)
				return pm.Execute(remainingPids, qty, newMatched)
			}
		}
	}

	return false, matched, fmt.Errorf("unable to match the required quantity of parties")
}

func AreAvailable(schedule1 schedule_entities.Schedule, schedule2 schedule_entities.Schedule) bool {
	// Loop through each DateOption in schedule1
	for _, option1 := range schedule1.Options {
		// Loop through each DateOption in schedule2
		for _, option2 := range schedule2.Options {
			// Check if any combination of options allows for a match
			if hasMatchingAvailability(option1, option2) {
				return true
			}
		}
	}
	return false
}

func hasMatchingAvailability(option1 schedule_entities.DateOption, option2 schedule_entities.DateOption) bool {
	// Loop through all possible combinations of days and timeframes
	for _, day1 := range option1.Days {
		for _, weekday1 := range option1.Weekdays {
			for _, timeframe1 := range option1.TimeFrames {
				for _, day2 := range option2.Days {
					for _, weekday2 := range option2.Weekdays {
						for _, timeframe2 := range option2.TimeFrames {
							// Check if the schedules have overlapping availability for the combination
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
	// Check if weekdays and days match or allow for pairing
	// (This logic might need adjustments based on your specific requirements)
	if weekday1 != weekday2 && !(weekday1 == time.Sunday && weekday2 == time.Saturday) {
		return false
	}
	if day1 != 0 && day2 != 0 && day1 != day2 {
		return false
	}

	// Check for overlapping timeframes
	return isTimeFrameOverlapping(timeframe1.Start, timeframe1.End, timeframe2.Start, timeframe2.End)
}

func isTimeFrameOverlapping(start1, end1, start2, end2 time.Time) bool {
	// Check if timeframe1 starts before timeframe2 ends and timeframe2 starts before timeframe1 ends
	return start1.Before(end2) && start2.Before(end1)
}
