package usecases

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
)

type PartyScheduleMatcher struct {
	Schedules map[uuid.UUID]schedule_entities.Schedule
}

func (uc *PartyScheduleMatcher) Execute(pids []uuid.UUID, qty int, matched []uuid.UUID) (bool, []uuid.UUID, error) {
	fmt.Printf("Executing with pids: %v, qty: %d, matched: %v\n", pids, qty, matched)
	if len(matched) == qty {
		return true, matched, nil
	}

	if len(pids) == 0 {
		return false, matched, fmt.Errorf("not enough parties available to match the required quantity")
	}

	for i, current := range pids {
		if schedule, ok := uc.Schedules[current]; ok {
			matchingParties := uc.findMatchingParties(pids, i, schedule)
			fmt.Printf("Matching parties for %v: %v\n", current, matchingParties)
			if len(matchingParties) >= qty-len(matched)-1 {
				return uc.matchParties(pids, i, qty, matched, current, matchingParties)
			}
		}
	}

	return false, matched, fmt.Errorf("unable to match the required quantity of parties")
}

func (pm *PartyScheduleMatcher) matchParties(pids []uuid.UUID, currentIndex, qty int, matched []uuid.UUID, current uuid.UUID, matchingParties []uuid.UUID) (bool, []uuid.UUID, error) {
	newMatched := append(matched, current)
	remainingPids := append(pids[:currentIndex], pids[currentIndex+1:]...)
	for _, match := range matchingParties {
		newMatched = append(newMatched, match)
		remainingPids = removeUUID(remainingPids, match)
		if len(newMatched) == qty {
			return true, newMatched, nil
		}
	}
	return pm.Execute(remainingPids, qty, newMatched)
}

func (pm *PartyScheduleMatcher) findMatchingParties(pids []uuid.UUID, currentIndex int, schedule schedule_entities.Schedule) []uuid.UUID {
	var matchingParties []uuid.UUID
	for i, otherPID := range pids {
		if i == currentIndex {
			continue
		}
		if otherSchedule, ok := pm.Schedules[otherPID]; ok {
			if AreAvailable(schedule, otherSchedule) {
				matchingParties = append(matchingParties, otherPID)
			}
		}
	}
	return matchingParties
}

func AreAvailable(schedule1 schedule_entities.Schedule, schedule2 schedule_entities.Schedule) bool {
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

func removeUUID(slice []uuid.UUID, id uuid.UUID) []uuid.UUID {
	for i, v := range slice {
		if v == id {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
