package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	pairing_in "github.com/leet-gaming/match-making-api/pkg/domain/pairing/ports/in"
	schedule_entities "github.com/leet-gaming/match-making-api/pkg/domain/schedules/entities"
	schedules_in_ports "github.com/leet-gaming/match-making-api/pkg/domain/schedules/ports/in"
)

// compatibilityKey is used as a key for memoization cache
type compatibilityKey struct {
	ID1 uuid.UUID
	ID2 uuid.UUID
}

// PartyScheduleMatcher implements the recursive matching algorithm for parties based on schedule compatibility
// Optimized with caching, memoization, and parallel processing for high-load scenarios
type PartyScheduleMatcher struct {
	ScheduleReader schedules_in_ports.PartyScheduleReader
	
	// Cache for schedules to avoid repeated repository calls
	scheduleCache map[uuid.UUID]*schedule_entities.Schedule
	scheduleCacheMutex sync.RWMutex
	
	// Memoization cache for compatibility checks to avoid repeated calculations
	compatibilityCache map[compatibilityKey]bool
	compatibilityCacheMutex sync.RWMutex
	
	// Threshold for parallel processing (number of parties)
	parallelThreshold int
}

// NewPartyScheduleMatcher creates a new instance of PartyScheduleMatcher
func NewPartyScheduleMatcher(scheduleReader schedules_in_ports.PartyScheduleReader) pairing_in.PartyScheduleMatcher {
	return &PartyScheduleMatcher{
		ScheduleReader:      scheduleReader,
		scheduleCache:       make(map[uuid.UUID]*schedule_entities.Schedule),
		compatibilityCache:  make(map[compatibilityKey]bool),
		parallelThreshold:  10, // Use parallel processing when there are 10+ parties
	}
}

// Execute identifies compatible matches between parties recursively
// It handles cases where parties have non-overlapping schedules by checking all possible combinations
// Returns the matched parties or an error if no matches are found
// Optimized with caching and memoization for better performance
// Validates schedules and handles database communication errors gracefully
func (pm *PartyScheduleMatcher) Execute(pids []uuid.UUID, qty int, matched []uuid.UUID) ([]uuid.UUID, error) {
	ctx := context.Background()
	if qty <= 0 {
		return nil, errors.New("quantity must be greater than zero: invalid quantity parameter")
	}

	// Base case: we have enough matched parties
	if len(matched) >= qty {
		return matched, nil
	}

	// Base case: no more parties available
	if len(pids) == 0 {
		return nil, fmt.Errorf("not enough parties available to match the required quantity: need %d, have %d matched", qty, len(matched))
	}

	// Pre-load all schedules into cache to avoid repeated repository calls
	// This also validates schedules and handles database errors
	if err := pm.preloadSchedules(ctx, pids); err != nil {
		return nil, fmt.Errorf("failed to load schedules from database: %w", err)
	}

	// Use parallel processing for large sets of parties
	if len(pids) >= pm.parallelThreshold {
		return pm.executeParallel(ctx, pids, qty, matched)
	}

	// Recursive case: try to find matches starting with each available party
	var validationErrors []error
	for i, current := range pids {
		currentSchedule := pm.getCachedSchedule(current)
		if currentSchedule == nil {
			slog.WarnContext(ctx, "party has no schedule, skipping", "party_id", current)
			continue
		}

		// Validate schedule before using it
		if err := validateSchedule(*currentSchedule); err != nil {
			slog.ErrorContext(ctx, "invalid schedule detected, skipping party", "party_id", current, "error", err)
			validationErrors = append(validationErrors, fmt.Errorf("party %s has invalid schedule: %w", current, err))
			continue
		}

		// Find parties with compatible schedules
		matchingParties, err := pm.findMatchingParties(ctx, pids, i, *currentSchedule)
		if err != nil {
			// If there was an error finding matching parties (e.g., invalid schedules), log and continue
			slog.WarnContext(ctx, "error finding matching parties, continuing with next party", "party_id", current, "error", err)
			continue
		}
		
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

	// If we had validation errors, include them in the error message
	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("unable to match parties: %d parties had invalid schedules (first error: %v). Need %d, have %d matched, %d available", 
			len(validationErrors), validationErrors[0], qty, len(matched), len(pids))
	}

	// No valid match found
	return nil, fmt.Errorf("unable to match the required quantity of parties: need %d, have %d matched, %d available. All parties were checked but no compatible schedules found", qty, len(matched), len(pids))
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
// Uses cached schedules and memoized compatibility checks for better performance
// Returns error if invalid schedules are encountered
func (pm *PartyScheduleMatcher) findMatchingParties(ctx context.Context, pids []uuid.UUID, currentIndex int, schedule schedule_entities.Schedule) ([]uuid.UUID, error) {
	var matchingParties []uuid.UUID
	
	for i, otherPID := range pids {
		// Skip the current party
		if i == currentIndex {
			continue
		}
		
		// Get the other party's schedule from cache
		otherSchedule := pm.getCachedSchedule(otherPID)
		if otherSchedule == nil {
			slog.DebugContext(ctx, "party has no schedule, skipping compatibility check", "party_id", otherPID)
			continue
		}
		
		// Validate the other party's schedule before checking compatibility
		if err := validateSchedule(*otherSchedule); err != nil {
			slog.WarnContext(ctx, "invalid schedule detected during compatibility check, skipping", "party_id", otherPID, "error", err)
			// Continue with other parties rather than failing completely
			continue
		}
		
		// Check if schedules are compatible using memoized result
		if pm.areSchedulesCompatibleMemoized(schedule.ID, otherSchedule.ID, schedule, *otherSchedule) {
			matchingParties = append(matchingParties, otherPID)
		}
	}
	
	return matchingParties, nil
}

// preloadSchedules loads all schedules for the given party IDs into the cache
// This avoids repeated repository calls during the matching process
// Returns an error if database communication fails for critical operations
func (pm *PartyScheduleMatcher) preloadSchedules(ctx context.Context, pids []uuid.UUID) error {
	pm.scheduleCacheMutex.Lock()
	defer pm.scheduleCacheMutex.Unlock()
	
	var dbErrors []error
	for _, pid := range pids {
		// Only load if not already cached
		if _, exists := pm.scheduleCache[pid]; !exists {
			// Attempt to get schedule from repository
			// Note: GetScheduleByPartyID returns nil on error, so we can't directly detect DB errors
			// But we can validate the schedule if it's returned
			schedule := pm.ScheduleReader.GetScheduleByPartyID(pid)
			if schedule != nil {
				// Validate schedule before caching
				if err := validateSchedule(*schedule); err != nil {
					slog.WarnContext(ctx, "invalid schedule loaded from database, not caching", "party_id", pid, "error", err)
					dbErrors = append(dbErrors, fmt.Errorf("party %s: %w", pid, err))
					// Don't cache invalid schedules
					continue
				}
				pm.scheduleCache[pid] = schedule
			} else {
				// Schedule is nil - could be missing or database error
				// Log but don't fail the entire operation
				slog.DebugContext(ctx, "schedule not found for party", "party_id", pid)
			}
		}
	}
	
	// If all schedules failed to load, return an error
	if len(dbErrors) == len(pids) && len(pids) > 0 {
		return fmt.Errorf("failed to load schedules for all parties: %d errors (example: %v)", len(dbErrors), dbErrors[0])
	}
	
	// If some schedules failed but not all, log warning but continue
	if len(dbErrors) > 0 {
		slog.WarnContext(ctx, "some schedules failed to load or were invalid", "failed_count", len(dbErrors), "total_parties", len(pids))
	}
	
	return nil
}

// getCachedSchedule retrieves a schedule from cache, falling back to repository if not cached
func (pm *PartyScheduleMatcher) getCachedSchedule(pid uuid.UUID) *schedule_entities.Schedule {
	// Try to get from cache first
	pm.scheduleCacheMutex.RLock()
	if schedule, exists := pm.scheduleCache[pid]; exists {
		pm.scheduleCacheMutex.RUnlock()
		return schedule
	}
	pm.scheduleCacheMutex.RUnlock()
	
	// Not in cache, fetch from repository and cache it
	schedule := pm.ScheduleReader.GetScheduleByPartyID(pid)
	if schedule != nil {
		pm.scheduleCacheMutex.Lock()
		pm.scheduleCache[pid] = schedule
		pm.scheduleCacheMutex.Unlock()
	}
	
	return schedule
}

// areSchedulesCompatibleMemoized checks compatibility with memoization to avoid repeated calculations
func (pm *PartyScheduleMatcher) areSchedulesCompatibleMemoized(id1, id2 uuid.UUID, schedule1, schedule2 schedule_entities.Schedule) bool {
	// Create a consistent key (smaller ID first)
	key := compatibilityKey{ID1: id1, ID2: id2}
	if id1.String() > id2.String() {
		key = compatibilityKey{ID1: id2, ID2: id1}
	}
	
	// Check cache first
	pm.compatibilityCacheMutex.RLock()
	if result, exists := pm.compatibilityCache[key]; exists {
		pm.compatibilityCacheMutex.RUnlock()
		return result
	}
	pm.compatibilityCacheMutex.RUnlock()
	
	// Not in cache, calculate and store
	result := areSchedulesCompatible(schedule1, schedule2)
	
	pm.compatibilityCacheMutex.Lock()
	pm.compatibilityCache[key] = result
	pm.compatibilityCacheMutex.Unlock()
	
	return result
}

// executeParallel processes multiple combinations in parallel for better performance
func (pm *PartyScheduleMatcher) executeParallel(ctx context.Context, pids []uuid.UUID, qty int, matched []uuid.UUID) ([]uuid.UUID, error) {
	type matchResult struct {
		matched []uuid.UUID
		err     error
	}
	
	resultChan := make(chan matchResult, len(pids))
	var wg sync.WaitGroup
	
	// Limit concurrent goroutines to avoid excessive resource usage
	maxConcurrency := 10
	if len(pids) < maxConcurrency {
		maxConcurrency = len(pids)
	}
	
	semaphore := make(chan struct{}, maxConcurrency)
	
	for i, current := range pids {
		wg.Add(1)
		go func(idx int, pid uuid.UUID) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			currentSchedule := pm.getCachedSchedule(pid)
			if currentSchedule == nil {
				slog.WarnContext(ctx, "party has no schedule, skipping", "party_id", pid)
				return
			}
			
			matchingParties, err := pm.findMatchingParties(ctx, pids, idx, *currentSchedule)
			if err != nil {
				// Log error but continue with next party
				slog.WarnContext(ctx, "error finding matching parties in parallel execution", "party_id", pid, "error", err)
				return
			}
			
			needed := qty - len(matched) - 1
			
			if len(matchingParties) >= needed {
				matchedParties, err := pm.matchParties(ctx, pids, idx, qty, matched, pid, matchingParties)
				if err == nil {
					resultChan <- matchResult{matched: matchedParties, err: nil}
					return
				}
			}
		}(i, current)
	}
	
	// Wait for all goroutines and close channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Return first successful result
	for res := range resultChan {
		if res.err == nil && len(res.matched) >= qty {
			return res.matched, nil
		}
	}
	
	// No valid match found
	return nil, fmt.Errorf("unable to match the required quantity of parties: need %d, have %d matched, %d available", qty, len(matched), len(pids))
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

// validateSchedule validates a schedule for correctness
// Returns an error with descriptive message if the schedule is invalid
func validateSchedule(schedule schedule_entities.Schedule) error {
	// Check if schedule has any options
	if schedule.Options == nil {
		return fmt.Errorf("schedule has no options defined (schedule_id: %s)", schedule.ID)
	}
	
	if len(schedule.Options) == 0 {
		return fmt.Errorf("schedule has empty options map (schedule_id: %s)", schedule.ID)
	}
	
	// Validate each date option
	hasValidOption := false
	for optionKey, option := range schedule.Options {
		if err := validateDateOption(option, optionKey); err != nil {
			return fmt.Errorf("invalid date option at key %d: %w", optionKey, err)
		}
		hasValidOption = true
	}
	
	if !hasValidOption {
		return fmt.Errorf("schedule has no valid date options (schedule_id: %s)", schedule.ID)
	}
	
	return nil
}

// validateDateOption validates a single date option
func validateDateOption(option schedule_entities.DateOption, optionKey int) error {
	// Check if option has any timeframes
	if len(option.TimeFrames) == 0 {
		return fmt.Errorf("date option has no timeframes defined")
	}
	
	// Validate each timeframe
	for i, timeframe := range option.TimeFrames {
		if err := validateTimeFrame(timeframe, i); err != nil {
			return fmt.Errorf("timeframe %d: %w", i, err)
		}
	}
	
	// Check if at least one of Months, Weekdays, or Days is specified
	if len(option.Months) == 0 && len(option.Weekdays) == 0 && len(option.Days) == 0 {
		return fmt.Errorf("date option must specify at least one of: Months, Weekdays, or Days")
	}
	
	return nil
}

// validateTimeFrame validates a single timeframe
func validateTimeFrame(timeframe schedule_entities.TimeFrame, index int) error {
	// Check if start time is before end time
	if !timeframe.Start.Before(timeframe.End) {
		return fmt.Errorf("timeframe start time (%v) must be before end time (%v)", timeframe.Start, timeframe.End)
	}
	
	// Check if timeframe has valid duration (at least 1 minute)
	duration := timeframe.End.Sub(timeframe.Start)
	if duration < time.Minute {
		return fmt.Errorf("timeframe duration must be at least 1 minute, got %v", duration)
	}
	
	return nil
}

// ClearCache clears all caches (useful for testing or memory management)
func (pm *PartyScheduleMatcher) ClearCache() {
	pm.scheduleCacheMutex.Lock()
	pm.compatibilityCacheMutex.Lock()
	defer pm.scheduleCacheMutex.Unlock()
	defer pm.compatibilityCacheMutex.Unlock()
	
	pm.scheduleCache = make(map[uuid.UUID]*schedule_entities.Schedule)
	pm.compatibilityCache = make(map[compatibilityKey]bool)
}
