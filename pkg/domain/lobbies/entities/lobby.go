package lobbies_entities

import "github.com/google/uuid"

type Lobby struct {
	ID     uuid.UUID
	Game   Game
	Region Region
}
