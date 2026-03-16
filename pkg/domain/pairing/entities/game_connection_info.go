package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/leet-gaming/match-making-api/pkg/common"
)

// GameConnectionInfo contains the details needed for players to connect
// to the game server after all players have confirmed readiness.
// Fields are nullable because different games have different connection mechanisms:
//   - FPS games (CS2): server IP + port + passcode
//   - Console games: QR code scan
//   - Mobile games: deep link URL
type GameConnectionInfo struct {
	common.BaseEntity
	MatchID      uuid.UUID  `json:"match_id" bson:"match_id"`
	LobbyID      uuid.UUID  `json:"lobby_id" bson:"lobby_id"`
	GameID       string     `json:"game_id" bson:"game_id"`
	Region       string     `json:"region" bson:"region"`
	ServerURL    *string    `json:"server_url,omitempty" bson:"server_url,omitempty"`         // Full connection URL (e.g., steam://connect/...)
	ServerIP     *string    `json:"server_ip,omitempty" bson:"server_ip,omitempty"`           // Direct server IP address
	Port         *int       `json:"port,omitempty" bson:"port,omitempty"`                     // Server port
	Passcode     *string    `json:"passcode,omitempty" bson:"passcode,omitempty"`             // Server password/passcode
	QRCodeData   *string    `json:"qr_code_data,omitempty" bson:"qr_code_data,omitempty"`     // QR code payload (URL or data)
	Instructions string     `json:"instructions" bson:"instructions"`                          // Game-specific connect instructions
	DeepLink     *string    `json:"deep_link,omitempty" bson:"deep_link,omitempty"`            // Mobile/app deep link
	ExpiresAt    *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`          // When connection info expires
}

// NewGameConnectionInfo creates a new game connection info value object.
func NewGameConnectionInfo(
	resourceOwner common.ResourceOwner,
	matchID uuid.UUID,
	lobbyID uuid.UUID,
	gameID string,
	region string,
	instructions string,
) *GameConnectionInfo {
	return &GameConnectionInfo{
		BaseEntity:   common.NewEntity(resourceOwner),
		MatchID:      matchID,
		LobbyID:      lobbyID,
		GameID:       gameID,
		Region:       region,
		Instructions: instructions,
	}
}

// WithServer sets the server connection details (IP-based games like CS2).
func (g *GameConnectionInfo) WithServer(serverURL, serverIP string, port int, passcode string) *GameConnectionInfo {
	g.ServerURL = &serverURL
	g.ServerIP = &serverIP
	g.Port = &port
	g.Passcode = &passcode
	return g
}

// WithQRCode sets the QR code data (console/mobile games).
func (g *GameConnectionInfo) WithQRCode(qrData string) *GameConnectionInfo {
	g.QRCodeData = &qrData
	return g
}

// WithDeepLink sets the deep link URL (mobile/app games).
func (g *GameConnectionInfo) WithDeepLink(deepLink string) *GameConnectionInfo {
	g.DeepLink = &deepLink
	return g
}

// WithExpiration sets an expiration time for the connection info.
func (g *GameConnectionInfo) WithExpiration(expiresAt time.Time) *GameConnectionInfo {
	g.ExpiresAt = &expiresAt
	return g
}

// IsExpired returns true if the connection info has expired.
func (g *GameConnectionInfo) IsExpired() bool {
	if g.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*g.ExpiresAt)
}
