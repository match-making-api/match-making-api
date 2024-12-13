package lobbies_entities

import "github.com/google/uuid"

type Region struct {
	ID          uuid.UUID
	Slug        string
	Description string
}
