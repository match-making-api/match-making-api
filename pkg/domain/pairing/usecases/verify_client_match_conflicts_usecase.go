package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	pairing_entities "github.com/leet-gaming/match-making-api/pkg/domain/pairing/entities"
	pairing_out "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/out"
	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
	schedules_in_ports "github.com/leet-gaming/match-making-api/pkg/domain/schedules/ports/in"
)

// VerifyClientMatchConflictsUseCase verifies that a client's matches do not have scheduling conflicts
type VerifyClientMatchConflictsUseCase struct {
	PairReader         pairing_out.PairReader
	PairWriter         pairing_out.PairWriter
	PartyScheduleReader schedules_in_ports.PartyScheduleReader
	ConflictNotifier   ConflictNotifier
}

// ConflictNotifier is an interface for notifying about conflicts
type ConflictNotifier interface {
	NotifyConflict(ctx context.Context, partyID uuid.UUID, pairID uuid.UUID, reason string) error
}

// ConflictResult represents the result of a conflict check
type ConflictResult struct {
	HasConflict  bool
	ConflictingPairs []uuid.UUID
}

// Execute verifies all matches associated with a client (party) for scheduling conflicts
func (uc *VerifyClientMatchConflictsUseCase) Execute(ctx context.Context, partyID uuid.UUID) (*ConflictResult, error) {
	slog.InfoContext(ctx, "verifying match conflicts for party", "party_id", partyID)

	// Get client's availability schedule
	clientSchedule := uc.PartyScheduleReader.GetScheduleByPartyID(partyID)
	if clientSchedule == nil {
		slog.WarnContext(ctx, "party has no schedule, cannot verify conflicts", "party_id", partyID)
		return &ConflictResult{HasConflict: false, ConflictingPairs: []uuid.UUID{}}, nil
	}

	// Get all pairs (matches) for this party
	pairs, err := uc.PairReader.FindPairsByPartyID(ctx, partyID)
	if err != nil {
		return nil, fmt.Errorf("failed to find pairs for party %v: %w", partyID, err)
	}

	if len(pairs) == 0 {
		slog.DebugContext(ctx, "no matches found for party", "party_id", partyID)
		return &ConflictResult{HasConflict: false, ConflictingPairs: []uuid.UUID{}}, nil
	}

	slog.InfoContext(ctx, "found matches for party", "party_id", partyID, "match_count", len(pairs))

	// Check for conflicts between matches
	conflictingPairs := uc.detectConflicts(ctx, partyID, pairs, clientSchedule)

	if len(conflictingPairs) > 0 {
		// Flag conflicting pairs and notify
		for _, pairID := range conflictingPairs {
			err := uc.flagConflict(ctx, pairID, fmt.Sprintf("conflict detected with other matches for party %v", partyID))
			if err != nil {
				slog.ErrorContext(ctx, "failed to flag conflict", "pair_id", pairID, "error", err)
				continue
			}

			// Notify client and relevant parties
			if uc.ConflictNotifier != nil {
				if err := uc.ConflictNotifier.NotifyConflict(ctx, partyID, pairID, "match conflict detected with your availability"); err != nil {
					slog.ErrorContext(ctx, "failed to send conflict notification", "pair_id", pairID, "party_id", partyID, "error", err)
				}
			}

			slog.WarnContext(ctx, "conflict detected and flagged", "pair_id", pairID, "party_id", partyID)
		}

		return &ConflictResult{HasConflict: true, ConflictingPairs: conflictingPairs}, nil
	}

	slog.InfoContext(ctx, "no conflicts found for party", "party_id", partyID)
	return &ConflictResult{HasConflict: false, ConflictingPairs: []uuid.UUID{}}, nil
}

// detectConflicts checks for scheduling conflicts between matches and with client availability
// A conflict occurs when:
// 1. A match (pair) has parties with schedules incompatible with the client's availability
// 2. Two matches have parties with incompatible schedules (meaning they cannot happen simultaneously)
func (uc *VerifyClientMatchConflictsUseCase) detectConflicts(
	ctx context.Context,
	partyID uuid.UUID,
	pairs []*pairing_entities.Pair,
	clientSchedule *schedule_entities.Schedule,
) []uuid.UUID {
	var conflictingPairs []uuid.UUID
	conflictSet := make(map[uuid.UUID]bool)

	// Check each pair against client's availability
	for _, pair := range pairs {
		// Get the collective schedule requirement for this pair (intersection of all party schedules)
		pairSchedule := uc.getPairSchedule(ctx, pair)
		
		if pairSchedule == nil {
			// If we can't determine the pair schedule, skip conflict check for this pair
			continue
		}

		// Check if client's availability is compatible with the pair's schedule requirement
		// If not compatible, this match conflicts with client's availability
		if !areSchedulesCompatibleForConflict(*clientSchedule, *pairSchedule) {
			conflictSet[pair.ID] = true
			slog.WarnContext(ctx, "pair conflicts with client availability", 
				"pair_id", pair.ID, "party_id", partyID)
		}
	}

	// Check for conflicts between pairs (two matches that cannot happen simultaneously)
	for i, pair1 := range pairs {
		for j, pair2 := range pairs {
			if i >= j {
				continue // Avoid duplicate checks
			}

			// Skip if either pair is already flagged
			if conflictSet[pair1.ID] || conflictSet[pair2.ID] {
				continue
			}

			schedule1 := uc.getPairSchedule(ctx, pair1)
			schedule2 := uc.getPairSchedule(ctx, pair2)

			if schedule1 == nil || schedule2 == nil {
				continue
			}

			// If two pairs have incompatible schedules, they conflict
			// This means the client cannot participate in both matches
			if !areSchedulesCompatibleForConflict(*schedule1, *schedule2) {
				// Flag both pairs as conflicting
				conflictSet[pair1.ID] = true
				conflictSet[pair2.ID] = true
				slog.WarnContext(ctx, "pairs have conflicting schedules", 
					"pair1_id", pair1.ID, "pair2_id", pair2.ID, "party_id", partyID)
			}
		}
	}

	// Convert set to slice
	for pairID := range conflictSet {
		conflictingPairs = append(conflictingPairs, pairID)
	}

	return conflictingPairs
}

// getPairSchedule computes the intersection of all party schedules in a pair
// This represents the time slots when all parties in the pair are available
func (uc *VerifyClientMatchConflictsUseCase) getPairSchedule(ctx context.Context, pair *pairing_entities.Pair) *schedule_entities.Schedule {
	if len(pair.Match) == 0 {
		return nil
	}

	var schedules []*schedule_entities.Schedule
	for partyID := range pair.Match {
		schedule := uc.PartyScheduleReader.GetScheduleByPartyID(partyID)
		if schedule != nil {
			schedules = append(schedules, schedule)
		}
	}

	if len(schedules) == 0 {
		return nil
	}

	// Start with the first schedule
	result := schedules[0]

	// Intersect with all other schedules
	// For simplicity, we check if all schedules are compatible with each other
	// If they are, we can use any one of them as representative
	// If not all are compatible, the pair itself has an internal conflict
	for i := 1; i < len(schedules); i++ {
		if !areSchedulesCompatibleForConflict(*result, *schedules[i]) {
			// Internal conflict in the pair - parties have incompatible schedules
			// This is also a conflict we should flag
			return nil
		}
	}

	return result
}

// flagConflict marks a pair as having a conflict
func (uc *VerifyClientMatchConflictsUseCase) flagConflict(ctx context.Context, pairID uuid.UUID, reason string) error {
	pair, err := uc.PairReader.GetByID(ctx, pairID)
	if err != nil {
		return fmt.Errorf("failed to get pair %v: %w", pairID, err)
	}

	pair.ConflictStatus = pairing_entities.ConflictStatusFlagged
	pair.ConflictReason = reason

	_, err = uc.PairWriter.Save(pair)
	if err != nil {
		return fmt.Errorf("failed to save flagged pair %v: %w", pairID, err)
	}

	return nil
}

// areSchedulesCompatible checks if two schedules have compatible availability
// Returns true if schedules are compatible (have overlapping availability), false if they conflict
// This uses the same logic as party_schedule_matcher.go
func areSchedulesCompatibleForConflict(schedule1, schedule2 schedule_entities.Schedule) bool {
	// Check all combinations of date options from both schedules
	for _, option1 := range schedule1.Options {
		for _, option2 := range schedule2.Options {
			if hasMatchingAvailabilityForConflict(option1, option2) {
				return true
			}
		}
	}
	return false
}

func hasMatchingAvailabilityForConflict(option1, option2 schedule_entities.DateOption) bool {
	for _, day1 := range option1.Days {
		for _, weekday1 := range option1.Weekdays {
			for _, timeframe1 := range option1.TimeFrames {
				for _, day2 := range option2.Days {
					for _, weekday2 := range option2.Weekdays {
						for _, timeframe2 := range option2.TimeFrames {
							if isAvailableCombinationForConflict(day1, weekday1, timeframe1, day2, weekday2, timeframe2) {
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

func isAvailableCombinationForConflict(day1 int, weekday1 time.Weekday, timeframe1 schedule_entities.TimeFrame,
	day2 int, weekday2 time.Weekday, timeframe2 schedule_entities.TimeFrame) bool {
	if weekday1 != weekday2 && !(weekday1 == time.Sunday && weekday2 == time.Saturday) {
		return false
	}
	if day1 != 0 && day2 != 0 && day1 != day2 {
		return false
	}
	return isTimeFrameOverlappingForConflict(timeframe1.Start, timeframe1.End, timeframe2.Start, timeframe2.End)
}

func isTimeFrameOverlappingForConflict(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}
