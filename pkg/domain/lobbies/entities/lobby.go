package lobbies_entities

import "github.com/google/uuid"

type Lobby struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	ClientID uuid.UUID
	Region   Region
}
