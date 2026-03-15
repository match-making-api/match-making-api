package schemas

// Event type constants for matchmaking events.
// Use these when setting EventEnvelope.Type to ensure consistency
// between replay-api and match-making-api.
const (
	EventTypePlayerQueued         = "PlayerQueued"
	EventTypePlayerLeftQueue      = "PlayerLeftQueue"
	EventTypePlayerQueueConfirmed = "PlayerQueueConfirmed"
	EventTypeMatchCreated         = "MatchCreated"
	EventTypeMatchCompleted       = "MatchCompleted"
	EventTypeRatingsUpdated       = "RatingsUpdated"
	EventTypeServerAllocated        = "ServerAllocated"
	EventTypeMatchStarted           = "MatchStarted"
	EventTypeMatchResultsCalculated = "MatchResultsCalculated"
)

// CloudEventsSpecVersion is the CloudEvents specification version (e.g. "1.0").
const CloudEventsSpecVersion = "1.0"

// Schema version for event evolution (EventEnvelope.DataschemaVersion).
// Increment when making backward-incompatible changes.
const SchemaVersionV1 = 1
