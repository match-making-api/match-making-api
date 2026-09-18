package mapping

import (
	"strings"

	"github.com/google/uuid"
)

// QueueJoinMapping describes how a replay-api Session/PlayerQueued maps into
// match-making Pool + Party (Refs 2508-003).
//
// Rules:
//  1. Solo join: party_id defaults to player_id (one party per player).
//  2. Group join: party_id is shared across members (when provided by producer).
//  3. Pool shard key: game_id + region (+ tenant_id/client_id for isolation).
//  4. Ownership: tenant_id, client_id, resource_owner_id must match event envelope.
type QueueJoinMapping struct {
	PlayerID        uuid.UUID
	PartyID         uuid.UUID
	GameID          uuid.UUID
	RegionSlug      string
	TenantID        string
	ClientID        string
	ResourceOwnerID string
	Solo            bool
}

// ResolvePartyID returns the party UUID for pool membership.
// emptyOrInvalid partyIDString → use playerID (solo).
func ResolvePartyID(playerID uuid.UUID, partyIDString string) (partyID uuid.UUID, solo bool) {
	partyIDString = strings.TrimSpace(partyIDString)
	if partyIDString == "" {
		return playerID, true
	}
	parsed, err := uuid.Parse(partyIDString)
	if err != nil || parsed == uuid.Nil {
		return playerID, true
	}
	return parsed, parsed == playerID
}

// MapQueueJoin builds the mapping used by PlayerQueued / PlayerLeftQueue consumers.
func MapQueueJoin(
	playerID, gameID uuid.UUID,
	region, tenantID, clientID, resourceOwnerID, partyIDString string,
) QueueJoinMapping {
	partyID, solo := ResolvePartyID(playerID, partyIDString)
	return QueueJoinMapping{
		PlayerID:        playerID,
		PartyID:         partyID,
		GameID:          gameID,
		RegionSlug:      region,
		TenantID:        tenantID,
		ClientID:        clientID,
		ResourceOwnerID: resourceOwnerID,
		Solo:            solo,
	}
}
