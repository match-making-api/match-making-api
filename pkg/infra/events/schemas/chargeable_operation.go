package schemas

// ChargeableOperationRequestedPayload is published when match-making detects a
// paid operation (e.g. queue priority boost). Wallet API consumes and executes billing.
// Match-making does NOT deduct balances or update subscription usage.
//
// Proto contract (also in matchmaking_events.proto): field numbers reserved for future codegen.
type ChargeableOperationRequestedPayload struct {
	OperationType   string `json:"operation_type"`
	AmountCents     int64  `json:"amount_cents"`
	Currency        string `json:"currency"`
	PlayerID        string `json:"player_id"`
	ResourceOwnerID string `json:"resource_owner_id"`
	TenantID        string `json:"tenant_id"`
	ClientID        string `json:"client_id"`
	GameID          string `json:"game_id"`
	Region          string `json:"region"`
	CorrelationID   string `json:"correlation_id"`
	IdempotencyKey  string `json:"idempotency_key"`
	RequestedAtMs   int64  `json:"requested_at_epoch_ms"`
}

// Operation type constants for ChargeableOperationRequested.
const (
	OperationTypeQueuePriorityBoost = "queue_priority_boost"
	OperationTypeTournamentEntry    = "tournament_entry"
	OperationTypeLobbyCreationFee   = "lobby_creation_fee"
)

// ChargeableOperationEvent is the CloudEvents-shaped JSON published to Kafka
// until ChargeableOperationRequested is wired into the generated MatchmakingEvent oneof.
type ChargeableOperationEvent struct {
	ID                 string                              `json:"id"`
	Type               string                              `json:"type"`
	Source             string                              `json:"source"`
	SpecVersion        string                              `json:"specversion"`
	TimeUnixMs         int64                               `json:"time_unix_ms"`
	Subject            string                              `json:"subject"`
	ResourceOwnerID    string                              `json:"resource_owner_id"`
	CorrelationID      string                              `json:"correlation_id"`
	DataschemaVersion  int32                               `json:"dataschema_version"`
	ChargeableOperationRequested *ChargeableOperationRequestedPayload `json:"chargeable_operation_requested"`
}
