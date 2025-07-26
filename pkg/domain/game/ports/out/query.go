package out

import (
	"github.com/leet-gaming/match-making-api/pkg/common"
	game_entities "github.com/leet-gaming/match-making-api/pkg/domain/game/entities"
)

// GameReader interface for reading game data
type GameReader interface {
	common.Searchable[game_entities.Game]
}

// GameRepository combines all game data operations
type GameRepository interface {
	GameWriter
	GameReader
}

// GameModeReader interface for reading game mode data
type GameModeReader interface {
	common.Searchable[game_entities.GameMode]
}

// GameModeRepository combines all game mode data operations
type GameModeRepository interface {
	GameModeWriter
	GameModeReader
}

// RegionReader interface for reading region data
type RegionReader interface {
	common.Searchable[game_entities.Region]
}

// RegionRepository combines all region data operations
type RegionRepository interface {
	RegionWriter
	RegionReader
}
