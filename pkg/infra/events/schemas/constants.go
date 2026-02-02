package schemas

// Event type constants for matchmaking events.
// Use these when setting EventEnvelope.EventType to ensure consistency
// between replay-api and match-making-api.
const (
	EventTypePlayerQueued    = "PlayerQueued"
	EventTypeMatchCreated    = "MatchCreated"
	EventTypeMatchCompleted  = "MatchCompleted"
	EventTypeRatingsUpdated  = "RatingsUpdated"
)

// Schema version for event evolution.
// Increment when making backward-incompatible changes.
const SchemaVersionV1 = 1
